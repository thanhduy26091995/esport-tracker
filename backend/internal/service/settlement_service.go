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

type WinnerAllocation struct {
	ID             uuid.UUID
	PointsToDeduct int
}

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
	_, err = s.TriggerSettlement(userID, nil) // nil = auto mode (match history)
	return err
}

// TriggerSettlement executes the settlement process for a debtor.
// If winners is non-empty, uses the provided per-winner point allocations instead of match history.
func (s *SettlementService) TriggerSettlement(debtorID uuid.UUID, winners []WinnerAllocation) (*model.DebtSettlement, error) {
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

	debtPoints := -debtor.CurrentScore

	var matchIDs []uuid.UUID
	var settledPoints int // actual points collected from winners

	if len(winners) > 0 {
		// Manual mode: validate winners and sum allocated points
		for _, w := range winners {
			if w.ID == debtorID {
				tx.Rollback()
				return nil, errors.New("debtor cannot be a winner")
			}

			winner, err := s.userRepo.GetByID(w.ID)
			if err != nil {
				tx.Rollback()
				return nil, errors.New("invalid winner ID")
			}

			if winner.CurrentScore <= 0 {
				tx.Rollback()
				return nil, errors.New("winners must have positive scores")
			}

			if w.PointsToDeduct > winner.CurrentScore {
				tx.Rollback()
				return nil, errors.New("points to deduct exceeds winner score")
			}

			if w.PointsToDeduct > debtPoints-settledPoints {
				tx.Rollback()
				return nil, errors.New("total allocated points exceeds debt")
			}

			settledPoints += w.PointsToDeduct
		}

		if settledPoints == 0 {
			tx.Rollback()
			return nil, errors.New("no winners found for settlement")
		}
	} else {
		// Auto mode: derive winners from match history
		matches, err := s.matchRepo.GetByUserID(debtorID, 0)
		if err != nil {
			tx.Rollback()
			return nil, err
		}

		winnerMap := make(map[uuid.UUID]int)

		for _, match := range matches {
			if match.IsLocked {
				continue
			}

			matchIDs = append(matchIDs, match.ID)

			var debtorTeam int
			for _, p := range match.Participants {
				if p.UserID == debtorID {
					debtorTeam = p.TeamNumber
				}
			}

			winnerTeam := match.WinnerTeam
			if debtorTeam != 0 && debtorTeam != winnerTeam {
				for _, p := range match.Participants {
					if p.TeamNumber == winnerTeam && p.UserID != debtorID {
						winnerMap[p.UserID]++
					}
				}
			}
		}

		totalWinningPoints := 0
		for _, pts := range winnerMap {
			totalWinningPoints += pts
		}
		if totalWinningPoints == 0 {
			tx.Rollback()
			return nil, errors.New("no winners found for settlement")
		}

		for _, pts := range winnerMap {
			settledPoints += (debtPoints * pts) / totalWinningPoints
		}
	}

	// All money calculations are based on settled points, not full debt
	actualMoneyAmount := settledPoints * pointToVND
	fundAmount := (actualMoneyAmount * fundSplitPercent) / 100
	winnerAmount := actualMoneyAmount - fundAmount

	// Create settlement record
	settlement := &model.DebtSettlement{
		DebtorID:           debtorID,
		DebtAmount:         debtor.CurrentScore,
		MoneyAmount:        actualMoneyAmount,
		ToFund:             float64(fundAmount),
		ToWinners:          float64(winnerAmount),
		FundAmount:         fundAmount,
		WinnerDistribution: winnerAmount,
		SettlementDate:     time.Now(),
		OriginalDebtPoints: debtPoints,
	}

	if err := tx.Create(settlement).Error; err != nil {
		tx.Rollback()
		return nil, err
	}

	// Distribute to winners
	if len(winners) > 0 {
		for _, w := range winners {
			winnerShare := (winnerAmount * w.PointsToDeduct) / settledPoints

			settlementWinner := &model.SettlementWinner{
				SettlementID:   settlement.ID,
				WinnerID:       w.ID,
				MoneyAmount:    winnerShare,
				PointsDeducted: w.PointsToDeduct,
			}
			if err := tx.Create(settlementWinner).Error; err != nil {
				tx.Rollback()
				return nil, err
			}

			if err := tx.Model(&model.User{}).
				Where("id = ?", w.ID).
				Update("current_score", gorm.Expr("current_score - ?", w.PointsToDeduct)).
				Error; err != nil {
				tx.Rollback()
				return nil, err
			}
		}
	} else {
		winnerMap := make(map[uuid.UUID]int)
		totalWinningPoints := 0
		matches, _ := s.matchRepo.GetByUserID(debtorID, 0)
		for _, match := range matches {
			if match.IsLocked {
				continue
			}
			var debtorTeam int
			for _, p := range match.Participants {
				if p.UserID == debtorID {
					debtorTeam = p.TeamNumber
				}
			}
			winnerTeam := match.WinnerTeam
			if debtorTeam != 0 && debtorTeam != winnerTeam {
				for _, p := range match.Participants {
					if p.TeamNumber == winnerTeam && p.UserID != debtorID {
						winnerMap[p.UserID]++
					}
				}
			}
		}
		for _, pts := range winnerMap {
			totalWinningPoints += pts
		}

		for winnerID, pts := range winnerMap {
			pointsToDeduct := (debtPoints * pts) / totalWinningPoints
			winnerShare := (winnerAmount * pts) / totalWinningPoints

			settlementWinner := &model.SettlementWinner{
				SettlementID:   settlement.ID,
				WinnerID:       winnerID,
				MoneyAmount:    winnerShare,
				PointsDeducted: pointsToDeduct,
			}
			if err := tx.Create(settlementWinner).Error; err != nil {
				tx.Rollback()
				return nil, err
			}

			if err := tx.Model(&model.User{}).
				Where("id = ?", winnerID).
				Update("current_score", gorm.Expr("current_score - ?", pointsToDeduct)).
				Error; err != nil {
				tx.Rollback()
				return nil, err
			}
		}

		for _, matchID := range matchIDs {
			if err := s.matchRepo.Lock(matchID); err != nil {
				tx.Rollback()
				return nil, err
			}
		}
	}

	// Debtor recovers only the settled points (partial settlement leaves remaining debt)
	if err := tx.Model(&model.User{}).
		Where("id = ?", debtorID).
		Update("current_score", gorm.Expr("current_score + ?", settledPoints)).
		Error; err != nil {
		tx.Rollback()
		return nil, err
	}

	// Deposit fund amount
	if fundAmount > 0 {
		description := fmt.Sprintf("Settlement: %d%% fund share from %s's debt (%d VND)", fundSplitPercent, debtor.Name, actualMoneyAmount)
		if err := s.fundService.CreateSettlementDeposit(fundAmount, description); err != nil {
			tx.Rollback()
			return nil, err
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

// GetFundContributors returns ranked fund contribution totals per user
func (s *SettlementService) GetFundContributors() ([]*model.FundContributor, error) {
	rows, err := s.settlementRepo.GetFundContributors()
	if err != nil {
		return nil, err
	}
	for i, r := range rows {
		r.Rank = i + 1
	}
	return rows, nil
}
