package service

import (
	"fmt"

	"github.com/duyb/esport-score-tracker/internal/model"
	"github.com/duyb/esport-score-tracker/internal/repository"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type WcChampionService struct {
	repo     *repository.WcChampionRepository
	wcRepo   *repository.WcRepository
	userRepo *repository.WcUserRepository
}

func NewWcChampionService(repo *repository.WcChampionRepository, wcRepo *repository.WcRepository, userRepo *repository.WcUserRepository) *WcChampionService {
	return &WcChampionService{repo: repo, wcRepo: wcRepo, userRepo: userRepo}
}

// --- Config ---

func (s *WcChampionService) GetPublicConfig() (*model.WcChampionConfigPublic, error) {
	cfg, err := s.repo.GetConfig()
	if err != nil {
		return nil, err
	}
	pub := &model.WcChampionConfigPublic{
		IsOpen:    cfg.IsOpen,
		SettledAt: cfg.SettledAt,
	}
	if cfg.WinnerID != nil {
		if team, err := s.repo.GetTeam(*cfg.WinnerID); err == nil {
			pub.WinnerTeam = team
		}
	}
	return pub, nil
}

func (s *WcChampionService) UpdateConfig(isOpen bool) error {
	return s.repo.UpdateConfig(isOpen)
}

// --- Teams ---

func (s *WcChampionService) ListTeams() ([]*model.WcChampionTeam, error) {
	return s.repo.ListTeams()
}

func (s *WcChampionService) CreateTeam(name, code, flagEmoji string, odds float64) (*model.WcChampionTeam, error) {
	t := &model.WcChampionTeam{
		Name:      name,
		Code:      code,
		FlagEmoji: flagEmoji,
		Odds:      odds,
	}
	return t, s.repo.CreateTeam(t)
}

func (s *WcChampionService) UpdateTeamOdds(teamID uuid.UUID, odds float64) error {
	if odds <= 1 {
		return fmt.Errorf("odds must be greater than 1")
	}
	return s.repo.UpdateTeamOdds(teamID, odds)
}

// --- Predictions ---

func (s *WcChampionService) GetAllPredictions() ([]*model.WcChampionPredictionPublic, error) {
	return s.repo.GetAllPredictions()
}

func (s *WcChampionService) GetMyPrediction(wcUserID uuid.UUID) (*model.WcChampionPredictionMine, error) {
	return s.repo.GetMyPrediction(wcUserID)
}

func (s *WcChampionService) GetMyPredictions(wcUserID uuid.UUID) ([]*model.WcChampionPredictionMine, error) {
	return s.repo.GetMyPredictions(wcUserID)
}

func (s *WcChampionService) PlaceOrUpdatePrediction(wcUserID, teamID uuid.UUID, points int) (*model.WcChampionPredictionMine, error) {
	user, err := s.userRepo.GetByID(wcUserID)
	if err != nil {
		return nil, fmt.Errorf("user not found")
	}
	if user.IsBlocked {
		return nil, fmt.Errorf("user is blocked from placing predictions")
	}
	cfg, err := s.repo.GetConfig()
	if err != nil {
		return nil, fmt.Errorf("champion config not found")
	}
	if !cfg.IsOpen {
		return nil, fmt.Errorf("champion prediction window is closed")
	}
	if points < 1 || points > 5 {
		return nil, fmt.Errorf("points must be between 1 and 5")
	}
	team, err := s.repo.GetTeam(teamID)
	if err != nil {
		return nil, fmt.Errorf("team not found")
	}
	pred := &model.WcChampionPrediction{
		WcUserID:     wcUserID,
		TeamID:       teamID,
		Points:       points,
		OddsSnapshot: team.Odds,
	}
	if err := s.repo.CreatePrediction(pred); err != nil {
		return nil, fmt.Errorf("already predicted this team — each team can only be picked once")
	}
	return s.repo.GetMyPrediction(wcUserID)
}

func (s *WcChampionService) DeletePrediction(wcUserID uuid.UUID) error {
	cfg, err := s.repo.GetConfig()
	if err != nil {
		return fmt.Errorf("champion config not found")
	}
	if !cfg.IsOpen {
		return fmt.Errorf("champion prediction window is closed — cannot delete")
	}
	return s.repo.DeletePrediction(wcUserID)
}

func (s *WcChampionService) DeletePredictionByID(wcUserID, predID uuid.UUID) error {
	cfg, err := s.repo.GetConfig()
	if err != nil {
		return fmt.Errorf("champion config not found")
	}
	if !cfg.IsOpen {
		return fmt.Errorf("champion prediction window is closed — cannot delete")
	}
	if err := s.repo.DeletePredictionByID(predID, wcUserID); err != nil {
		if err == gorm.ErrRecordNotFound {
			return fmt.Errorf("prediction not found")
		}
		return err
	}
	return nil
}

// SettleChampion declares the winner and settles all predictions.
// Winners receive payout; losers are deducted their wagered points.
// Idempotent: returns error if already settled.
func (s *WcChampionService) SettleChampion(adminID, winnerTeamID uuid.UUID) (*model.WcChampionSettleResult, error) {
	cfg, err := s.repo.GetConfig()
	if err != nil {
		return nil, fmt.Errorf("champion config not found")
	}
	if cfg.SettledAt != nil {
		return nil, fmt.Errorf("champion has already been settled")
	}
	if cfg.IsOpen {
		return nil, fmt.Errorf("close the prediction window before settling")
	}
	winnerTeam, err := s.repo.GetTeam(winnerTeamID)
	if err != nil {
		return nil, fmt.Errorf("winner team not found")
	}

	preds, err := s.repo.ListPredictionsForSettle()
	if err != nil {
		return nil, err
	}

	result := &model.WcChampionSettleResult{Winner: winnerTeam.Name, SettledCount: len(preds)}
	settledUsers := make(map[uuid.UUID]struct{})

	db := s.repo.DB()
	txErr := db.Transaction(func(tx *gorm.DB) error {
		for _, p := range preds {
			isCorrect := p.TeamID == winnerTeamID
			resStr := model.WcResultIncorrect
			pointsEarned := 0
			var delta float64

			if isCorrect {
				resStr = model.WcResultCorrect
				pointsEarned = int(float64(p.Points) * p.OddsSnapshot)
				delta = float64(pointsEarned)
				result.CorrectCount++
				result.TotalPointsAwarded += pointsEarned
			} else {
				delta = -float64(p.Points)
			}

			if err := s.repo.SettlePrediction(tx, p.ID, resStr, pointsEarned); err != nil {
				return err
			}

			wallet, err := s.wcRepo.GetWallet(p.WcUserID)
			if err != nil {
				return fmt.Errorf("wallet not found for user %v", p.WcUserID)
			}
			balanceBefore := wallet.Balance
			if err := s.wcRepo.UpdateWalletBalance(tx, p.WcUserID, delta); err != nil {
				return err
			}
			note := "champion settle — incorrect"
			if isCorrect {
				note = "champion settle — correct"
			}
			if err := s.wcRepo.LogWalletChange(tx, &model.WcWalletLog{
				WcUserID:      p.WcUserID,
				AdminID:       adminID,
				Delta:         delta,
				BalanceBefore: balanceBefore,
				BalanceAfter:  balanceBefore + delta,
				Note:          note,
			}); err != nil {
				return err
			}

			settledUsers[p.WcUserID] = struct{}{}
		}
		return s.repo.MarkSettled(winnerTeamID)
	})
	if txErr != nil {
		return nil, txErr
	}
	result.SettledUserCount = len(settledUsers)
	return result, nil
}
