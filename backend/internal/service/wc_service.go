package service

import (
	"errors"
	"fmt"
	"math"
	"time"

	"github.com/duyb/esport-score-tracker/internal/model"
	"github.com/duyb/esport-score-tracker/internal/repository"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type WcService struct {
	repo     *repository.WcRepository
	userRepo *repository.WcUserRepository
	football *footballClient
}

func NewWcService(repo *repository.WcRepository, userRepo *repository.WcUserRepository) *WcService {
	return &WcService{repo: repo, userRepo: userRepo, football: newFootballClient()}
}

// --- Config ---

func (s *WcService) GetConfig() (*model.WcConfig, error) {
	return s.repo.GetConfig()
}

func (s *WcService) SetConfig(isEnabled bool, updatedBy uuid.UUID) error {
	return s.repo.UpdateConfig(isEnabled, &updatedBy)
}

// --- Sync ---

func (s *WcService) SyncMatches() (int, error) {
	matches, err := s.football.FetchWCMatches()
	if err != nil {
		return 0, err
	}
	if len(matches) == 0 {
		return 0, nil
	}
	if err := s.repo.UpsertMatches(matches); err != nil {
		return 0, fmt.Errorf("failed to upsert matches: %w", err)
	}
	return len(matches), nil
}

// --- Matches ---

func (s *WcService) ListMatches(f repository.MatchFilter) ([]*model.WcMatch, error) {
	return s.repo.ListMatches(f)
}

func (s *WcService) GetMatch(id uuid.UUID) (*model.WcMatch, error) {
	return s.repo.GetMatch(id)
}

func (s *WcService) GetMatchWithOdds(id uuid.UUID) (*model.WcMatchWithOdds, error) {
	return s.repo.GetMatchWithOdds(id)
}

func (s *WcService) UpdateMatch(id uuid.UUID, fields map[string]interface{}) error {
	return s.repo.UpdateMatch(id, fields)
}

func (s *WcService) OpenMatch(id uuid.UUID) error {
	return s.repo.UpdateMatch(id, map[string]interface{}{"predictions_open": true})
}

func (s *WcService) CloseMatch(id uuid.UUID) error {
	return s.repo.UpdateMatch(id, map[string]interface{}{"predictions_open": false})
}

// --- Score multipliers ---

func (s *WcService) AddScoreMultiplier(matchID uuid.UUID, homeScore, awayScore int, odds float64) (*model.WcScoreMultiplier, error) {
	m, err := s.repo.GetMatch(matchID)
	if err != nil {
		return nil, fmt.Errorf("match not found")
	}
	if isFinished(m) {
		return nil, fmt.Errorf("match is already completed or cancelled — cannot add score odds")
	}
	so := &model.WcScoreMultiplier{
		MatchID:   matchID,
		HomeScore: homeScore,
		AwayScore: awayScore,
		Multiplier: odds,
	}
	return so, s.repo.CreateScoreMultiplier(so)
}

func (s *WcService) UpdateScoreMultiplier(id uuid.UUID, multiplier float64) error {
	return s.repo.UpdateScoreMultiplier(id, multiplier)
}

func (s *WcService) DeleteScoreMultiplier(id uuid.UUID) error {
	so, err := s.repo.GetScoreMultiplierByID(id)
	if err != nil {
		return fmt.Errorf("score odds not found")
	}
	m, err := s.repo.GetMatch(so.MatchID)
	if err != nil {
		return err
	}
	if isFinished(m) {
		return fmt.Errorf("match is already completed or cancelled — cannot delete score odds")
	}
	return s.repo.DeleteScoreMultiplier(id)
}

func (s *WcService) ListScoreMultipliers(matchID uuid.UUID) ([]*model.WcScoreMultiplier, error) {
	return s.repo.ListScoreMultipliers(matchID)
}

// --- Predictions ---

type SubmitPredictionRequest struct {
	MatchID            uuid.UUID
	PredictionType     string
	PredictionChoice   *string // handicap: 'home'|'away'; nil for exact_score
	PredictedHomeScore *int    // exact_score only
	PredictedAwayScore *int    // exact_score only
	Points             int
}

func (s *WcService) SubmitPrediction(wcUserID uuid.UUID, req SubmitPredictionRequest) (*model.WcPrediction, error) {
	m, err := s.repo.GetMatch(req.MatchID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("match not found")
		}
		return nil, err
	}

	if isLocked(m) {
		return nil, fmt.Errorf("predictions are closed for this match")
	}

	if req.Points <= 0 {
		return nil, fmt.Errorf("points must be greater than 0")
	}
	if req.Points > 5 {
		return nil, fmt.Errorf("points must not exceed 5 per prediction")
	}

	var multiplierSnapshot float64
	var handicapSnapshot *float64
	var handicapTeamSnapshot *string

	switch req.PredictionType {
	case model.WcPredictionTypeHandicap:
		if req.PredictionChoice == nil || (*req.PredictionChoice != model.WcTeamHome && *req.PredictionChoice != model.WcTeamAway) {
			return nil, fmt.Errorf("prediction_choice must be 'home' or 'away' for handicap predictions")
		}
		if m.HandicapValue == nil || m.OddsHandicapHome == nil || m.OddsHandicapAway == nil {
			return nil, fmt.Errorf("handicap odds not set for this match")
		}
		if *req.PredictionChoice == model.WcTeamHome {
			multiplierSnapshot = *m.OddsHandicapHome
		} else {
			multiplierSnapshot = *m.OddsHandicapAway
		}
		handicapSnapshot = m.HandicapValue
		handicapTeamSnapshot = &m.HandicapTeam

	case model.WcPredictionTypeExactScore:
		if req.PredictedHomeScore == nil || req.PredictedAwayScore == nil {
			return nil, fmt.Errorf("predicted_home_score and predicted_away_score are required")
		}
		so, err := s.repo.GetScoreMultiplier(req.MatchID, *req.PredictedHomeScore, *req.PredictedAwayScore)
		if err != nil {
			return nil, fmt.Errorf("scoreline %d:%d is not available for this match",
				*req.PredictedHomeScore, *req.PredictedAwayScore)
		}
		multiplierSnapshot = so.Multiplier

	default:
		return nil, fmt.Errorf("prediction_type must be 'handicap' or 'exact_score'")
	}

	bet := &model.WcPrediction{
		WcUserID:             wcUserID,
		MatchID:              req.MatchID,
		PredictionType:       req.PredictionType,
		PredictionChoice:     req.PredictionChoice,
		PredictedHomeScore:   req.PredictedHomeScore,
		PredictedAwayScore:   req.PredictedAwayScore,
		Points:               req.Points,
		MultiplierSnapshot:   multiplierSnapshot,
		HandicapSnapshot:     handicapSnapshot,
		HandicapTeamSnapshot: handicapTeamSnapshot,
	}

	if err := s.repo.CreatePrediction(nil, bet); err != nil {
		return nil, fmt.Errorf("failed to submit prediction (may be duplicate): %w", err)
	}
	return bet, nil
}

func (s *WcService) ListPredictions(wcUserID uuid.UUID) ([]*model.WcPredictionWithMatch, error) {
	return s.repo.ListPredictions(wcUserID)
}

func (s *WcService) DeletePrediction(wcUserID, betID uuid.UUID) error {
	bet, err := s.repo.GetPredictionByID(betID)
	if err != nil {
		return fmt.Errorf("prediction not found")
	}
	if bet.WcUserID != wcUserID {
		return fmt.Errorf("unauthorized")
	}
	if bet.Result != nil {
		return fmt.Errorf("cannot delete a finalized prediction")
	}
	m, err := s.repo.GetMatch(bet.MatchID)
	if err != nil {
		return fmt.Errorf("match not found")
	}
	if isLocked(m) {
		return fmt.Errorf("cannot modify prediction: match is locked")
	}
	return s.repo.DeletePrediction(betID)
}

func (s *WcService) UpdatePredictionPoints(wcUserID, betID uuid.UUID, points int) error {
	if points <= 0 {
		return fmt.Errorf("points must be greater than 0")
	}
	if points > 5 {
		return fmt.Errorf("points must not exceed 5 per prediction")
	}
	bet, err := s.repo.GetPredictionByID(betID)
	if err != nil {
		return fmt.Errorf("prediction not found")
	}
	if bet.WcUserID != wcUserID {
		return fmt.Errorf("unauthorized")
	}
	if bet.Result != nil {
		return fmt.Errorf("cannot modify a finalized prediction")
	}
	m, err := s.repo.GetMatch(bet.MatchID)
	if err != nil {
		return fmt.Errorf("match not found")
	}
	if isLocked(m) {
		return fmt.Errorf("cannot modify prediction: match is locked")
	}
	return s.repo.UpdatePredictionPoints(betID, points)
}

func (s *WcService) ListPredictionsForMatchPublic(matchID uuid.UUID) ([]*model.WcPredictionPublic, error) {
	return s.repo.ListPredictionsForMatchPublic(matchID)
}

func (s *WcService) GetWallet(wcUserID uuid.UUID) (*model.WcWallet, error) {
	w, err := s.repo.GetWallet(wcUserID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			if createErr := s.repo.CreateWallet(s.repo.DB(), wcUserID); createErr != nil {
				return nil, createErr
			}
			return s.repo.GetWallet(wcUserID)
		}
		return nil, err
	}
	return w, nil
}

func (s *WcService) GetLeaderboard() ([]*model.WcLeaderboardEntry, error) {
	return s.repo.GetLeaderboard()
}

// --- Match settlement ---

// FinalizeMatch evaluates all predictions on a match, credits/debits wallets, and marks settled_at.
// Idempotent: re-settling reverses previous payouts before re-applying.
func (s *WcService) FinalizeMatch(matchID uuid.UUID) (int, int, error) {
	m, err := s.repo.GetMatch(matchID)
	if err != nil {
		return 0, 0, fmt.Errorf("match not found")
	}
	if m.HomeScore == nil || m.AwayScore == nil {
		return 0, 0, fmt.Errorf("match score not set — cannot settle")
	}

	bets, err := s.repo.ListPredictionsForMatch(matchID)
	if err != nil {
		return 0, 0, err
	}

	db := s.repo.DB()
	var totalPointsEarned int
	processed := 0

	err = db.Transaction(func(tx *gorm.DB) error {
		for _, bet := range bets {
			// Reverse previous payout if re-settling
			if bet.Result != nil && bet.PointsEarned != nil && *bet.PointsEarned > 0 {
				prevDelta := float64(-(*bet.PointsEarned))
				if err := s.repo.UpdateWalletBalance(tx, bet.WcUserID, prevDelta); err != nil {
					return err
				}
			}

			var result string
			var pointsEarned int

			switch bet.PredictionType {
			case model.WcPredictionTypeHandicap:
				result, pointsEarned = evaluateHandicapPrediction(bet, *m.HomeScore, *m.AwayScore)
			case model.WcPredictionTypeExactScore:
				result, pointsEarned = evaluateExactScorePrediction(bet, *m.HomeScore, *m.AwayScore)
			}

			if err := s.repo.UpdatePredictionResult(tx, bet.ID, result, pointsEarned); err != nil {
				return err
			}

			// Credit wallet with pointsEarned (0 for losses)
			if pointsEarned > 0 {
				if err := s.repo.UpdateWalletBalance(tx, bet.WcUserID, float64(pointsEarned)); err != nil {
					return err
				}
			}
			// Deduct points from wallet (happens on bet placement for losses only conceptually;
			// here we track net: pointsEarned - points)
			if err := s.repo.UpdateWalletBalance(tx, bet.WcUserID, float64(-bet.Points)); err != nil {
				return err
			}

			totalPointsEarned += pointsEarned
			processed++
		}

		// Mark match as settled
		now := time.Now()
		return s.repo.UpdateMatch(matchID, map[string]interface{}{
			"settled_at": now,
			"status":     model.WcStatusCompleted,
		})
	})

	return processed, totalPointsEarned, err
}

// --- Tournament settlement ---

func (s *WcService) PreviewSettlement(pointRate float64) ([]*model.WcSettlementPreviewRow, error) {
	return s.repo.PreviewSettlement(pointRate)
}

func (s *WcService) CreateSettlement(adminID uuid.UUID, name string, pointRate float64, note string) (*model.WcSettlement, error) {
	wallets, err := s.repo.GetAllWallets()
	if err != nil {
		return nil, err
	}

	settlement := &model.WcSettlement{
		Name:      name,
		PointRate: pointRate,
		SettledBy: adminID,
		Note:      note,
	}

	details := make([]*model.WcSettlementDetail, 0, len(wallets))
	for _, w := range wallets {
		dir := model.WcDirectionEven
		if w.Balance > 0 {
			dir = model.WcDirectionPay
		} else if w.Balance < 0 {
			dir = model.WcDirectionCollect
		}
		details = append(details, &model.WcSettlementDetail{
			WcUserID:     w.WcUserID,
			FinalBalance: w.Balance,
			Amount:       math.Abs(w.Balance) * pointRate,
			Direction:    dir,
			Status:       model.WcSettlementStatusPending,
		})
	}

	db := s.repo.DB()
	err = db.Transaction(func(tx *gorm.DB) error {
		if err := s.repo.CreateSettlement(tx, settlement, details); err != nil {
			return err
		}
		return s.repo.ResetAllWallets(tx)
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create settlement: %w", err)
	}
	return settlement, nil
}

func (s *WcService) GetHousePnL() (*model.HousePnLResponse, error) {
	return s.repo.GetHousePnL()
}

func (s *WcService) ListSettlements() ([]*model.WcSettlement, error) {
	return s.repo.ListSettlements()
}

func (s *WcService) GetSettlement(id uuid.UUID) (*model.WcSettlementWithDetails, error) {
	return s.repo.GetSettlement(id)
}

func (s *WcService) MarkSettlementDone(settlementID, wcUserID uuid.UUID, doneNote string) error {
	return s.repo.UpdateSettlementDetailStatus(settlementID, wcUserID, model.WcSettlementStatusDone, doneNote)
}

// --- Wallet admin ops ---

func (s *WcService) AdminTopUp(adminID, wcUserID uuid.UUID, delta int, note string) error {
	if delta == 0 {
		return fmt.Errorf("delta cannot be 0")
	}
	wallet, err := s.repo.GetWallet(wcUserID)
	if err != nil {
		return fmt.Errorf("wallet not found for user")
	}
	balanceBefore := wallet.Balance
	deltaF := float64(delta)
	balanceAfter := balanceBefore + deltaF

	db := s.repo.DB()
	return db.Transaction(func(tx *gorm.DB) error {
		if err := s.repo.UpdateWalletBalance(tx, wcUserID, deltaF); err != nil {
			return err
		}
		return s.repo.LogWalletChange(tx, &model.WcWalletLog{
			WcUserID:      wcUserID,
			AdminID:       adminID,
			Delta:         deltaF,
			BalanceBefore: balanceBefore,
			BalanceAfter:  balanceAfter,
			Note:          note,
		})
	})
}

func (s *WcService) GetWalletLogs(wcUserID uuid.UUID) ([]*model.WcWalletLog, error) {
	return s.repo.GetWalletLogs(wcUserID)
}

func (s *WcService) GetAllWallets() ([]*model.WcWalletWithUser, error) {
	return s.repo.GetAllWallets()
}

// --- User management ---

func (s *WcService) GetAllUsers() ([]*model.WcUser, error) {
	return s.userRepo.GetAll()
}

func (s *WcService) SetAdminRole(wcUserID uuid.UUID, isAdmin bool) error {
	return s.userRepo.SetAdminRole(wcUserID, isAdmin)
}

// --- Bets ---

type PlaceBetRequest struct {
	MatchID            uuid.UUID
	BetType            string
	BetChoice          *string
	Stake              int
	PredictedHomeScore *int
	PredictedAwayScore *int
}

func (s *WcService) PlaceBet(wcUserID uuid.UUID, req PlaceBetRequest) (*model.WcBet, error) {
	m, err := s.repo.GetMatch(req.MatchID)
	if err != nil {
		return nil, fmt.Errorf("match not found")
	}
	if isBetLocked(m) {
		return nil, fmt.Errorf("betting is closed for this match")
	}
	if req.Stake <= 0 {
		return nil, fmt.Errorf("stake must be greater than 0")
	}
	if req.Stake > 5 {
		return nil, fmt.Errorf("stake must not exceed 5 per bet")
	}

	var oddsSnapshot float64
	var handicapSnapshot *float64
	var handicapTeamSnapshot *string

	switch req.BetType {
	case model.WcBetTypeHandicap:
		if req.BetChoice == nil || (*req.BetChoice != model.WcTeamHome && *req.BetChoice != model.WcTeamAway) {
			return nil, fmt.Errorf("bet_choice must be 'home' or 'away'")
		}
		if m.HandicapValue == nil || m.OddsHandicapHome == nil || m.OddsHandicapAway == nil {
			return nil, fmt.Errorf("handicap odds not configured for this match")
		}
		if *req.BetChoice == model.WcTeamHome {
			oddsSnapshot = *m.OddsHandicapHome
		} else {
			oddsSnapshot = *m.OddsHandicapAway
		}
		handicapSnapshot = m.HandicapValue
		handicapTeamSnapshot = &m.HandicapTeam
	case model.WcBetTypeExactScore:
		if req.PredictedHomeScore == nil || req.PredictedAwayScore == nil {
			return nil, fmt.Errorf("predicted scores required for exact score bet")
		}
		so, err := s.repo.GetScoreOdds(req.MatchID, *req.PredictedHomeScore, *req.PredictedAwayScore)
		if err != nil {
			return nil, fmt.Errorf("scoreline %d:%d is not available for this match", *req.PredictedHomeScore, *req.PredictedAwayScore)
		}
		oddsSnapshot = so.Odds
	default:
		return nil, fmt.Errorf("bet_type must be 'handicap' or 'exact_score'")
	}

	bet := &model.WcBet{
		WcUserID:             wcUserID,
		MatchID:              req.MatchID,
		BetType:              req.BetType,
		BetChoice:            req.BetChoice,
		Stake:                req.Stake,
		OddsSnapshot:         oddsSnapshot,
		HandicapSnapshot:     handicapSnapshot,
		HandicapTeamSnapshot: handicapTeamSnapshot,
		PredictedHomeScore:   req.PredictedHomeScore,
		PredictedAwayScore:   req.PredictedAwayScore,
	}
	if err := s.repo.CreateBet(bet); err != nil {
		return nil, fmt.Errorf("failed to place bet (may be duplicate): %w", err)
	}
	return bet, nil
}

func (s *WcService) ListBets(wcUserID uuid.UUID) ([]*model.WcBetWithMatch, error) {
	return s.repo.ListBets(wcUserID)
}

func (s *WcService) ListBetsForMatch(matchID uuid.UUID) ([]*model.WcBetPublic, error) {
	return s.repo.ListBetsForMatch(matchID)
}

func (s *WcService) UpdateBetStake(wcUserID, betID uuid.UUID, stake int) error {
	if stake <= 0 {
		return fmt.Errorf("stake must be greater than 0")
	}
	if stake > 5 {
		return fmt.Errorf("stake must not exceed 5 per bet")
	}
	bet, err := s.repo.GetBet(betID)
	if err != nil {
		return fmt.Errorf("bet not found")
	}
	if bet.WcUserID != wcUserID {
		return fmt.Errorf("unauthorized")
	}
	if bet.Result != nil {
		return fmt.Errorf("cannot modify a settled bet")
	}
	m, err := s.repo.GetMatch(bet.MatchID)
	if err != nil {
		return fmt.Errorf("match not found")
	}
	if isBetLocked(m) {
		return fmt.Errorf("betting is closed for this match")
	}
	return s.repo.UpdateBetStake(betID, stake)
}

func (s *WcService) DeleteBet(wcUserID, betID uuid.UUID) error {
	bet, err := s.repo.GetBet(betID)
	if err != nil {
		return fmt.Errorf("bet not found")
	}
	if bet.WcUserID != wcUserID {
		return fmt.Errorf("unauthorized")
	}
	if bet.Result != nil {
		return fmt.Errorf("cannot delete a settled bet")
	}
	m, err := s.repo.GetMatch(bet.MatchID)
	if err != nil {
		return fmt.Errorf("match not found")
	}
	if isBetLocked(m) {
		return fmt.Errorf("betting is closed for this match")
	}
	return s.repo.DeleteBet(betID, wcUserID)
}

// SettleMatch evaluates all bets on a match and updates wallets. Idempotent.
func (s *WcService) SettleMatch(matchID uuid.UUID) (int, float64, error) {
	m, err := s.repo.GetMatch(matchID)
	if err != nil {
		return 0, 0, fmt.Errorf("match not found")
	}
	if m.Status == model.WcStatusCancelled {
		return 0, 0, fmt.Errorf("cannot settle a cancelled match")
	}
	if m.HomeScore == nil || m.AwayScore == nil {
		return 0, 0, fmt.Errorf("match score not set — cannot settle")
	}
	if m.BetsLockedAt == nil || time.Now().Before(*m.BetsLockedAt) {
		return 0, 0, fmt.Errorf("bets are not locked yet — cannot settle")
	}

	bets, err := s.repo.ListBetsForSettlement(matchID)
	if err != nil {
		return 0, 0, err
	}

	db := s.repo.DB()
	var totalPayout float64
	processed := 0

	err = db.Transaction(func(tx *gorm.DB) error {
		for _, bet := range bets {
			// Reverse previous settlement for idempotency
			if bet.Result != nil && bet.Payout != nil {
				prevNet := *bet.Payout - float64(bet.Stake)
				if err := s.repo.UpdateWalletBalance(tx, bet.WcUserID, -prevNet); err != nil {
					return err
				}
			}

			var result string
			var payout float64
			switch bet.BetType {
			case model.WcBetTypeHandicap:
				result, payout = evaluateHandicapBet(bet, *m.HomeScore, *m.AwayScore)
			case model.WcBetTypeExactScore:
				result, payout = evaluateExactScoreBet(bet, *m.HomeScore, *m.AwayScore)
			}

			if err := s.repo.UpdateBetResult(tx, bet.ID, result, payout); err != nil {
				return err
			}

			netChange := payout - float64(bet.Stake)
			if err := s.repo.UpdateWalletBalance(tx, bet.WcUserID, netChange); err != nil {
				return err
			}

			totalPayout += payout
			processed++
		}
		return nil
	})

	return processed, totalPayout, err
}

// --- Score odds (admin) ---

func (s *WcService) AddScoreOdds(matchID uuid.UUID, homeScore, awayScore int, odds float64) (*model.WcScoreOdds, error) {
	m, err := s.repo.GetMatch(matchID)
	if err != nil {
		return nil, fmt.Errorf("match not found")
	}
	if isFinished(m) {
		return nil, fmt.Errorf("match is already finished")
	}
	so := &model.WcScoreOdds{MatchID: matchID, HomeScore: homeScore, AwayScore: awayScore, Odds: odds}
	return so, s.repo.CreateScoreOdds(so)
}

func (s *WcService) UpdateScoreOdds(id uuid.UUID, odds float64) error {
	return s.repo.UpdateScoreOdds(id, odds)
}

func (s *WcService) DeleteScoreOdds(id uuid.UUID) error {
	return s.repo.DeleteScoreOdds(id)
}

func (s *WcService) ListScoreOdds(matchID uuid.UUID) ([]*model.WcScoreOdds, error) {
	return s.repo.ListScoreOdds(matchID)
}

// --- Helpers ---

func isBetLocked(m *model.WcMatch) bool {
	if m.Status == model.WcStatusCompleted || m.Status == model.WcStatusCancelled {
		return true
	}
	if m.BetsLockedAt != nil && time.Now().After(*m.BetsLockedAt) {
		return true
	}
	return false
}

// --- Helpers ---

func isLocked(m *model.WcMatch) bool {
	if m.Status == model.WcStatusLive || m.Status == model.WcStatusCompleted || m.Status == model.WcStatusCancelled {
		return true
	}
	if !m.PredictionsOpen {
		return true
	}
	if m.PredictionsLockedAt != nil && time.Now().After(*m.PredictionsLockedAt) {
		return true
	}
	return false
}

// isFinished checks only terminal statuses — used for admin score-odds management.
// Unlike isLocked, it allows edits on live/locked matches.
func isFinished(m *model.WcMatch) bool {
	return m.Status == model.WcStatusCompleted || m.Status == model.WcStatusCancelled
}

// isQuarterHandicap returns true for .25/.75 handicaps which use split-bet (Asian quarter-ball) rules.
func isQuarterHandicap(h float64) bool {
	frac := math.Mod(math.Abs(h), 0.5)
	return math.Abs(frac-0.25) < 0.001
}

// evalSubHandicap evaluates a single handicap line, returning "win", "lose", or "push".
func evalSubHandicap(homeScore, awayScore int, h float64, handicapTeam, betChoice string) string {
	var adjustedHome float64
	if handicapTeam == model.WcTeamHome {
		adjustedHome = float64(homeScore) - h
	} else {
		adjustedHome = float64(homeScore) + h
	}
	adjustedAway := float64(awayScore)
	if adjustedHome > adjustedAway {
		if betChoice == model.WcTeamHome {
			return "win"
		}
		return "lose"
	}
	if adjustedHome < adjustedAway {
		if betChoice == model.WcTeamAway {
			return "win"
		}
		return "lose"
	}
	return "push"
}

// evaluateHandicapPrediction applies Asian handicap rules for the prediction system.
// Quarter handicaps (e.g. 1.25, 0.75) split the stake across two lines.
func evaluateHandicapPrediction(bet *model.WcPrediction, homeScore, awayScore int) (string, int) {
	if bet.HandicapSnapshot == nil || bet.HandicapTeamSnapshot == nil || bet.PredictionChoice == nil {
		return model.WcResultIncorrect, 0
	}

	h := *bet.HandicapSnapshot
	hTeam := *bet.HandicapTeamSnapshot
	choice := *bet.PredictionChoice

	if isQuarterHandicap(h) {
		r1 := evalSubHandicap(homeScore, awayScore, h-0.25, hTeam, choice)
		r2 := evalSubHandicap(homeScore, awayScore, h+0.25, hTeam, choice)
		half := float64(bet.Points) / 2
		switch {
		case r1 == "win" && r2 == "win":
			return model.WcResultCorrect, int(math.Floor(float64(bet.Points) * bet.MultiplierSnapshot))
		case r1 == "lose" && r2 == "lose":
			return model.WcResultIncorrect, 0
		case (r1 == "win" || r2 == "win"):
			// one win + one push
			return model.WcResultWinHalf, int(math.Floor(half*bet.MultiplierSnapshot)) + int(math.Floor(half))
		default:
			// one push + one lose
			return model.WcResultLoseHalf, int(math.Floor(half))
		}
	}

	var adjustedHome float64
	if hTeam == model.WcTeamHome {
		adjustedHome = float64(homeScore) - h
	} else {
		adjustedHome = float64(homeScore) + h
	}
	adjustedAway := float64(awayScore)

	var winner string
	if adjustedHome > adjustedAway {
		winner = model.WcTeamHome
	} else if adjustedHome < adjustedAway {
		winner = model.WcTeamAway
	} else {
		return model.WcResultVoid, bet.Points
	}

	if winner == choice {
		return model.WcResultCorrect, int(math.Floor(float64(bet.Points) * bet.MultiplierSnapshot))
	}
	return model.WcResultIncorrect, 0
}

// evaluateExactScorePrediction checks exact score for the prediction system.
func evaluateExactScorePrediction(bet *model.WcPrediction, homeScore, awayScore int) (string, int) {
	if bet.PredictedHomeScore == nil || bet.PredictedAwayScore == nil {
		return model.WcResultIncorrect, 0
	}
	if *bet.PredictedHomeScore == homeScore && *bet.PredictedAwayScore == awayScore {
		pointsEarned := int(math.Floor(float64(bet.Points) * bet.MultiplierSnapshot))
		return model.WcResultCorrect, pointsEarned
	}
	return model.WcResultIncorrect, 0
}

// evaluateHandicapBet applies Asian handicap rules for the betting system.
// Quarter handicaps (e.g. 1.25, 0.75) split the stake across two lines.
func evaluateHandicapBet(bet *model.WcBet, homeScore, awayScore int) (string, float64) {
	if bet.HandicapSnapshot == nil || bet.HandicapTeamSnapshot == nil || bet.BetChoice == nil {
		return model.WcResultLose, 0
	}

	h := *bet.HandicapSnapshot
	hTeam := *bet.HandicapTeamSnapshot
	choice := *bet.BetChoice

	if isQuarterHandicap(h) {
		r1 := evalSubHandicap(homeScore, awayScore, h-0.25, hTeam, choice)
		r2 := evalSubHandicap(homeScore, awayScore, h+0.25, hTeam, choice)
		half := float64(bet.Stake) / 2
		switch {
		case r1 == "win" && r2 == "win":
			return model.WcResultWin, math.Round(float64(bet.Stake)*bet.OddsSnapshot*100) / 100
		case r1 == "lose" && r2 == "lose":
			return model.WcResultLose, 0
		case (r1 == "win" || r2 == "win"):
			// one win + one push: payout winning half + refund pushing half
			return model.WcResultWinHalf, math.Round((half*bet.OddsSnapshot+half)*100) / 100
		default:
			// one push + one lose: refund pushing half only
			return model.WcResultLoseHalf, math.Round(half*100) / 100
		}
	}

	var adjustedHome float64
	if hTeam == model.WcTeamHome {
		adjustedHome = float64(homeScore) - h
	} else {
		adjustedHome = float64(homeScore) + h
	}
	adjustedAway := float64(awayScore)

	var winner string
	if adjustedHome > adjustedAway {
		winner = model.WcTeamHome
	} else if adjustedHome < adjustedAway {
		winner = model.WcTeamAway
	} else {
		return model.WcResultPush, float64(bet.Stake)
	}

	if winner == choice {
		return model.WcResultWin, math.Round(float64(bet.Stake)*bet.OddsSnapshot*100) / 100
	}
	return model.WcResultLose, 0
}

// evaluateExactScoreBet checks exact score for the betting system.
func evaluateExactScoreBet(bet *model.WcBet, homeScore, awayScore int) (string, float64) {
	if bet.PredictedHomeScore == nil || bet.PredictedAwayScore == nil {
		return model.WcResultLose, 0
	}
	if *bet.PredictedHomeScore == homeScore && *bet.PredictedAwayScore == awayScore {
		payout := math.Round(float64(bet.Stake)*bet.OddsSnapshot*100) / 100
		return model.WcResultWin, payout
	}
	return model.WcResultLose, 0
}
