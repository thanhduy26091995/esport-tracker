package service

import (
	"fmt"
	"math"

	"github.com/duyb/esport-score-tracker/internal/model"
	"github.com/duyb/esport-score-tracker/internal/repository"
	"github.com/duyb/esport-score-tracker/internal/ws"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type WcChampionService struct {
	repo     *repository.WcChampionRepository
	wcRepo   *repository.WcRepository
	userRepo *repository.WcUserRepository
	hub      ws.HubBroadcaster
}

func NewWcChampionService(repo *repository.WcChampionRepository, wcRepo *repository.WcRepository, userRepo *repository.WcUserRepository, hub ws.HubBroadcaster) *WcChampionService {
	return &WcChampionService{repo: repo, wcRepo: wcRepo, userRepo: userRepo, hub: hub}
}

// --- Config ---

func (s *WcChampionService) GetPublicConfig(tournamentType string) (*model.WcChampionConfigPublic, error) {
	cfg, err := s.repo.GetConfig(tournamentType)
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

func (s *WcChampionService) UpdateConfig(tournamentType string, isOpen bool) error {
	return s.repo.UpdateConfig(tournamentType, isOpen)
}

// --- Teams ---

func (s *WcChampionService) ListTeams(tournamentType string) ([]*model.WcChampionTeam, error) {
	return s.repo.ListTeams(tournamentType)
}

func (s *WcChampionService) CreateTeam(tournamentType, name, code, flagEmoji string, odds float64) (*model.WcChampionTeam, error) {
	t := &model.WcChampionTeam{
		TournamentType: tournamentType,
		Name:           name,
		Code:           code,
		FlagEmoji:      flagEmoji,
		Odds:           odds,
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

func (s *WcChampionService) GetAllPredictions(tournamentType string) ([]*model.WcChampionPredictionPublic, error) {
	return s.repo.GetAllPredictions(tournamentType)
}

func (s *WcChampionService) GetMyPrediction(tournamentType string, wcUserID uuid.UUID) (*model.WcChampionPredictionMine, error) {
	return s.repo.GetMyPrediction(tournamentType, wcUserID)
}

func (s *WcChampionService) GetMyPredictions(tournamentType string, wcUserID uuid.UUID) ([]*model.WcChampionPredictionMine, error) {
	return s.repo.GetMyPredictions(tournamentType, wcUserID)
}

func (s *WcChampionService) PlaceOrUpdatePrediction(tournamentType string, wcUserID, teamID uuid.UUID, points int) (*model.WcChampionPredictionMine, error) {
	user, err := s.userRepo.GetByID(wcUserID)
	if err != nil {
		return nil, fmt.Errorf("user not found")
	}
	if user.IsBlocked {
		return nil, fmt.Errorf("user is blocked from placing predictions")
	}
	cfg, err := s.repo.GetConfig(tournamentType)
	if err != nil {
		return nil, fmt.Errorf("champion config not found")
	}
	if !cfg.IsOpen {
		return nil, fmt.Errorf("champion prediction window is closed")
	}
	betCfg, err := s.wcRepo.GetConfig(tournamentType)
	if err != nil {
		return nil, fmt.Errorf("failed to load config")
	}
	if points < betCfg.MinPoints || points > betCfg.MaxPoints {
		return nil, fmt.Errorf("điểm cược phải từ %d đến %d", betCfg.MinPoints, betCfg.MaxPoints)
	}
	team, err := s.repo.GetTeam(teamID)
	if err != nil {
		return nil, fmt.Errorf("team not found")
	}
	pred := &model.WcChampionPrediction{
		TournamentType: tournamentType,
		WcUserID:       wcUserID,
		TeamID:         teamID,
		Points:         points,
		OddsSnapshot:   team.Odds,
	}
	if err := s.repo.CreatePrediction(pred); err != nil {
		return nil, fmt.Errorf("already predicted this team — each team can only be picked once")
	}

	if s.hub != nil {
		s.hub.Broadcast(ws.ActivityEvent{
			Type:      "bet_placed",
			UserID:    wcUserID.String(),
			UserName:  user.Name,
			BetType:   "champion",
			Selection: team.Name,
			Stake:     points,
		})
	}

	return s.repo.GetMyPrediction(tournamentType, wcUserID)
}

func (s *WcChampionService) DeletePrediction(tournamentType string, wcUserID uuid.UUID) error {
	cfg, err := s.repo.GetConfig(tournamentType)
	if err != nil {
		return fmt.Errorf("champion config not found")
	}
	if !cfg.IsOpen {
		return fmt.Errorf("champion prediction window is closed — cannot delete")
	}
	return s.repo.DeletePrediction(tournamentType, wcUserID)
}

func (s *WcChampionService) DeletePredictionByID(tournamentType string, wcUserID, predID uuid.UUID) error {
	cfg, err := s.repo.GetConfig(tournamentType)
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
func (s *WcChampionService) SettleChampion(tournamentType string, adminID, winnerTeamID uuid.UUID) (*model.WcChampionSettleResult, error) {
	cfg, err := s.repo.GetConfig(tournamentType)
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

	preds, err := s.repo.ListPredictionsForSettle(tournamentType)
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
				pointsEarned = int(math.Round(float64(p.Points) * p.OddsSnapshot))
				delta = float64(pointsEarned) - float64(p.Points) // net profit (deferred-deduction model)
				result.CorrectCount++
				result.TotalPointsAwarded += pointsEarned
			} else {
				delta = -float64(p.Points)
			}

			if err := s.repo.SettlePrediction(tx, p.ID, resStr, pointsEarned); err != nil {
				return err
			}

			wallet, err := s.wcRepo.GetWalletTx(tx, p.WcUserID)
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
		return s.repo.MarkSettled(tournamentType, winnerTeamID)
	})
	if txErr != nil {
		return nil, txErr
	}
	result.SettledUserCount = len(settledUsers)
	return result, nil
}
