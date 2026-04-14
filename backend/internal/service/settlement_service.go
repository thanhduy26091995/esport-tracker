package service

import (
	"errors"
	"fmt"
	"time"

	"github.com/duyb/esport-score-tracker/internal/model"
	"github.com/duyb/esport-score-tracker/internal/repository"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type SettlementService struct {
	settlementRepo *repository.SettlementRepository
	userRepo       *repository.UserRepository
	matchRepo      *repository.MatchRepository
	fundService    *FundService
	configService  *ConfigService
	db             *gorm.DB
}

func NewSettlementService(
	settlementRepo *repository.SettlementRepository,
	userRepo *repository.UserRepository,
	matchRepo *repository.MatchRepository,
	fundService *FundService,
	configService *ConfigService,
	db *gorm.DB,
) *SettlementService {
	return &SettlementService{
		settlementRepo: settlementRepo,
		userRepo:       userRepo,
		matchRepo:      matchRepo,
		fundService:    fundService,
		configService:  configService,
		db:             db,
	}
}

// CheckAndTriggerSettlement checks if a user needs settlement and triggers it
func (s *SettlementService) CheckAndTriggerSettlement(userID uuid.UUID) error {
	// Get user
	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		return err
	}

	// Get debt threshold
	debtThreshold, err := s.configService.GetDebtThreshold()
	if err != nil {
		return err
	}

	// Check if settlement is needed
	if user.CurrentScore > debtThreshold {
		return nil // No settlement needed
	}

	// Trigger settlement
	return s.TriggerSettlement(userID)
}

// TriggerSettlement executes the settlement process for a debtor
func (s *SettlementService) TriggerSettlement(debtorID uuid.UUID) error {
	// Begin transaction
	tx := s.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Get debtor
	debtor, err := s.userRepo.GetByID(debtorID)
	if err != nil {
		tx.Rollback()
		return err
	}

	// Check if debtor actually has debt
	if debtor.CurrentScore >= 0 {
		tx.Rollback()
		return errors.New("user does not have debt")
	}

	// Get configuration
	pointToVND, err := s.configService.GetPointToVND()
	if err != nil {
		tx.Rollback()
		return err
	}

	fundSplitPercent, err := s.configService.GetFundSplitPercent()
	if err != nil {
		tx.Rollback()
		return err
	}

	// Calculate total money amount (debt is negative, so negate it)
	debtPoints := -debtor.CurrentScore
	totalMoneyAmount := debtPoints * pointToVND

	// Calculate fund and winner portions
	fundAmount := (totalMoneyAmount * fundSplitPercent) / 100
	winnerAmount := totalMoneyAmount - fundAmount

	// Get debtor's match history to find winners
	matches, err := s.matchRepo.GetByUserID(debtorID, 0)
	if err != nil {
		tx.Rollback()
		return err
	}

	// Find all users who won against this debtor
	winnerMap := make(map[uuid.UUID]int) // userID -> points won against debtor
	matchIDs := []uuid.UUID{}

	for _, match := range matches {
		if match.IsLocked {
			continue // Skip already locked matches
		}

		matchIDs = append(matchIDs, match.ID)

		var debtorTeam int
		var winnerTeam int

		// Find debtor's team and winner team
		for _, p := range match.Participants {
			if p.UserID == debtorID {
				debtorTeam = p.TeamNumber
			}
		}

		winnerTeam = match.WinnerTeam

		// If debtor lost this match, add winners
		if debtorTeam != 0 && debtorTeam != winnerTeam {
			for _, p := range match.Participants {
				if p.TeamNumber == winnerTeam && p.UserID != debtorID {
					winnerMap[p.UserID] = winnerMap[p.UserID] + 1
				}
			}
		}
	}

	// Calculate total winning points
	totalWinningPoints := 0
	for _, points := range winnerMap {
		totalWinningPoints += points
	}

	if totalWinningPoints == 0 {
		tx.Rollback()
		return errors.New("no winners found for settlement")
	}

	// Create settlement record
	settlement := &model.DebtSettlement{
		DebtorID:            debtorID,
		DebtAmount:          debtor.CurrentScore, // Negative value (e.g., -7)
		MoneyAmount:         totalMoneyAmount,
		FundAmount:          fundAmount,
		WinnerDistribution:  winnerAmount,
		SettlementDate:      time.Now(),
		OriginalDebtPoints:  debtPoints,
	}

	if err := tx.Create(settlement).Error; err != nil {
		tx.Rollback()
		return err
	}

	// Distribute to winners and create winner records
	for winnerID, points := range winnerMap {
		winnerShare := (winnerAmount * points) / totalWinningPoints
		pointsToDeduct := (winnerShare + pointToVND - 1) / pointToVND // Round up

		// Create settlement winner record
		settlementWinner := &model.SettlementWinner{
			SettlementID: settlement.ID,
			WinnerID:     winnerID,
			MoneyAmount:  winnerShare,
			PointsDeducted: pointsToDeduct,
		}

		if err := tx.Create(settlementWinner).Error; err != nil {
			tx.Rollback()
			return err
		}

		// Deduct points from winner
		if err := tx.Model(&model.User{}).
			Where("id = ?", winnerID).
			Update("current_score", gorm.Expr("current_score - ?", pointsToDeduct)).
			Error; err != nil {
			tx.Rollback()
			return err
		}
	}

	// Reset debtor score to 0
	if err := tx.Model(&model.User{}).
		Where("id = ?", debtorID).
		Update("current_score", 0).
		Error; err != nil {
		tx.Rollback()
		return err
	}

	// Deposit fund amount
	if fundAmount > 0 {
		description := fmt.Sprintf("Debt settlement from %s (50%% of %d VND)", debtor.Name, totalMoneyAmount)
		if err := s.fundService.CreateSettlementDeposit(fundAmount, description); err != nil {
			tx.Rollback()
			return err
		}
	}

	// Lock all related matches
	for _, matchID := range matchIDs {
		if err := s.matchRepo.Lock(matchID); err != nil {
			tx.Rollback()
			return err
		}
	}

	// Commit transaction
	return tx.Commit().Error
}

// GetAllSettlements returns all settlements
func (s *SettlementService) GetAllSettlements(limit, offset int) ([]*model.DebtSettlement, error) {
	return s.settlementRepo.GetAll(limit, offset)
}

// GetSettlementByID returns a settlement by ID
func (s *SettlementService) GetSettlementByID(id uuid.UUID) (*model.DebtSettlement, error) {
	return s.settlementRepo.GetByID(id)
}

// GetSettlementsByDebtorID returns settlements for a specific debtor
func (s *SettlementService) GetSettlementsByDebtorID(debtorID uuid.UUID, limit int) ([]*model.DebtSettlement, error) {
	return s.settlementRepo.GetByDebtorID(debtorID, limit)
}

// GetSettlementStats returns settlement statistics
func (s *SettlementService) GetSettlementStats() (map[string]interface{}, error) {
	total, err := s.settlementRepo.CountTotal()
	if err != nil {
		return nil, err
	}

	today, err := s.settlementRepo.CountToday()
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"total_settlements": total,
		"today_settlements": today,
	}, nil
}
