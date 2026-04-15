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
	_, err = s.TriggerSettlement(userID, nil)
	return err
}

// TriggerSettlement executes the settlement process for a debtor
// If winnerIDs is provided, use manual winner selection instead of match history
func (s *SettlementService) TriggerSettlement(debtorID uuid.UUID, winnerIDs []uuid.UUID) (*model.DebtSettlement, error) {
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
		return nil, err
	}

	// Check if debtor actually has debt
	if debtor.CurrentScore >= 0 {
		tx.Rollback()
		return nil, errors.New("user does not have debt")
	}

	// Get configuration
	pointToVND, err := s.configService.GetPointToVND()
	if err != nil {
		tx.Rollback()
		return nil, err
	}

	fundSplitPercent, err := s.configService.GetFundSplitPercent()
	if err != nil {
		tx.Rollback()
		return nil, err
	}

	// Calculate total money amount (debt is negative, so negate it)
	debtPoints := -debtor.CurrentScore
	totalMoneyAmount := debtPoints * pointToVND

	// Calculate fund and winner portions
	fundAmount := (totalMoneyAmount * fundSplitPercent) / 100
	winnerAmount := totalMoneyAmount - fundAmount

	var winnerMap map[uuid.UUID]int // userID -> points
	var matchIDs []uuid.UUID

	// Determine winners: manual selection or match history
	if len(winnerIDs) > 0 {
		// Manual winner selection
		winnerMap = make(map[uuid.UUID]int)
		
		// Validate all winners exist and have positive scores
		for _, winnerID := range winnerIDs {
			if winnerID == debtorID {
				tx.Rollback()
				return nil, errors.New("debtor cannot be a winner")
			}
			
			winner, err := s.userRepo.GetByID(winnerID)
			if err != nil {
				tx.Rollback()
				return nil, errors.New("invalid winner ID")
			}
			
			if winner.CurrentScore <= 0 {
				tx.Rollback()
				return nil, errors.New("winners must have positive scores")
			}
			
			// Equal weight for manual winners
			winnerMap[winnerID] = 1
		}
	} else {
		// Get debtor's match history to find winners
		matches, err := s.matchRepo.GetByUserID(debtorID, 0)
		if err != nil {
			tx.Rollback()
			return nil, err
		}

		// Find all users who won against this debtor
		winnerMap = make(map[uuid.UUID]int) // userID -> points won against debtor

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
	}

	// Calculate total winning points
	totalWinningPoints := 0
	for _, points := range winnerMap {
		totalWinningPoints += points
	}

	if totalWinningPoints == 0 {
		tx.Rollback()
		return nil, errors.New("no winners found for settlement")
	}

	// Create settlement record
	settlement := &model.DebtSettlement{
		DebtorID:            debtorID,
		DebtAmount:          debtor.CurrentScore, // Negative value (e.g., -7)
		MoneyAmount:         totalMoneyAmount,
		ToFund:              float64(fundAmount),         // Legacy column
		ToWinners:           float64(winnerAmount),       // Legacy column
		FundAmount:          fundAmount,
		WinnerDistribution:  winnerAmount,
		SettlementDate:      time.Now(),
		OriginalDebtPoints:  debtPoints,
	}

	if err := tx.Create(settlement).Error; err != nil {
		tx.Rollback()
		return nil, err
	}

	// Distribute to winners and create winner records
	totalPointsDeducted := 0
	for winnerID, points := range winnerMap {
		winnerShare := (winnerAmount * points) / totalWinningPoints
		// Winners lose points proportional to TOTAL debt, not just their share
		pointsToDeduct := (debtPoints * points) / totalWinningPoints

		// Create settlement winner record
		settlementWinner := &model.SettlementWinner{
			SettlementID: settlement.ID,
			WinnerID:     winnerID,
			MoneyAmount:  winnerShare,
			PointsDeducted: pointsToDeduct,
		}

		if err := tx.Create(settlementWinner).Error; err != nil {
			tx.Rollback()
			return nil, err
		}

		// Deduct points from winner
		if err := tx.Model(&model.User{}).
			Where("id = ?", winnerID).
			Update("current_score", gorm.Expr("current_score - ?", pointsToDeduct)).
			Error; err != nil {
			tx.Rollback()
			return nil, err
		}

		totalPointsDeducted += pointsToDeduct
	}

	// Clear debtor's debt (add back the debt points to make score 0)
	if err := tx.Model(&model.User{}).
		Where("id = ?", debtorID).
		Update("current_score", gorm.Expr("current_score + ?", debtPoints)).
		Error; err != nil {
		tx.Rollback()
		return nil, err
	}

	// Deposit fund amount
	if fundAmount > 0 {
		description := fmt.Sprintf("Settlement: %d%% fund share from %s's debt (%d VND)", fundSplitPercent, debtor.Name, totalMoneyAmount)
		if err := s.fundService.CreateSettlementDeposit(fundAmount, description); err != nil {
			tx.Rollback()
			return nil, err
		}
	}

	// Lock all related matches (only for automatic mode)
	if len(winnerIDs) == 0 {
		for _, matchID := range matchIDs {
			if err := s.matchRepo.Lock(matchID); err != nil {
				tx.Rollback()
				return nil, err
			}
		}
	}

	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		return nil, err
	}

	// Fetch and return the complete settlement with relations
	return s.settlementRepo.GetByID(settlement.ID)
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
