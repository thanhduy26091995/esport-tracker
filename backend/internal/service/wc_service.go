package service

import (
	"errors"
	"fmt"
	"math"
	"time"

	"github.com/duyb/esport-score-tracker/internal/model"
	"github.com/duyb/esport-score-tracker/internal/repository"
	"github.com/duyb/esport-score-tracker/internal/ws"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type WcService struct {
	repo          *repository.WcRepository
	userRepo      *repository.WcUserRepository
	customBetRepo *repository.WcCustomBetRepository
	football      *footballClient
	hub           ws.HubBroadcaster // nil-safe: no broadcast when nil
}

func NewWcService(repo *repository.WcRepository, userRepo *repository.WcUserRepository, customBetRepo *repository.WcCustomBetRepository, hub ws.HubBroadcaster) *WcService {
	return &WcService{repo: repo, userRepo: userRepo, customBetRepo: customBetRepo, football: newFootballClient(), hub: hub}
}

// --- Config ---

func (s *WcService) GetConfig() (*model.WcConfig, error) {
	return s.repo.GetConfig()
}

func (s *WcService) SetConfig(isEnabled bool, updatedBy uuid.UUID) error {
	return s.repo.UpdateConfig(isEnabled, &updatedBy)
}

func (s *WcService) SetBetLimits(min, max int, updatedBy uuid.UUID) error {
	if min < 1 {
		return fmt.Errorf("min_points phải >= 1")
	}
	if max < min {
		return fmt.Errorf("max_points phải >= min_points")
	}
	return s.repo.UpdateBetLimits(min, max, &updatedBy)
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

// MatchScheduleSummary is returned by GetMatchScheduleSummary for cron scheduling decisions.
type MatchScheduleSummary struct {
	LiveCount int
}

// GetMatchScheduleSummary returns a lightweight summary used by the sync cron to decide interval.
func (s *WcService) GetMatchScheduleSummary() (*MatchScheduleSummary, error) {
	liveCount, err := s.repo.CountLiveMatches()
	if err != nil {
		return nil, err
	}
	return &MatchScheduleSummary{LiveCount: liveCount}, nil
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

	user, err := s.userRepo.GetByID(wcUserID)
	if err != nil {
		return nil, fmt.Errorf("user not found")
	}
	if user.IsBlocked {
		return nil, fmt.Errorf("user is blocked from placing predictions")
	}

	betCfg, err := s.repo.GetConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to load config")
	}
	if req.Points < betCfg.MinPoints || req.Points > betCfg.MaxPoints {
		return nil, fmt.Errorf("điểm cược phải từ %d đến %d", betCfg.MinPoints, betCfg.MaxPoints)
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

	case model.WcPredictionTypeOverUnder:
		if req.PredictionChoice == nil || (*req.PredictionChoice != model.WcChoiceOver && *req.PredictionChoice != model.WcChoiceUnder) {
			return nil, fmt.Errorf("prediction_choice must be 'over' or 'under'")
		}
		if m.OULine == nil || m.OddsOver == nil || m.OddsUnder == nil {
			return nil, fmt.Errorf("O/U odds not set for this match")
		}
		if *req.PredictionChoice == model.WcChoiceOver {
			multiplierSnapshot = *m.OddsOver
		} else {
			multiplierSnapshot = *m.OddsUnder
		}
		handicapSnapshot = m.OULine

	default:
		return nil, fmt.Errorf("prediction_type must be 'handicap', 'exact_score', or 'over_under'")
	}

	bet := &model.WcPrediction{
		WcUserID:             wcUserID,
		MatchID:              req.MatchID,
		PredictionType:       req.PredictionType,
		PredictionChoice:     req.PredictionChoice,
		PredictedHomeScore:   req.PredictedHomeScore,
		PredictedAwayScore:   req.PredictedAwayScore,
		Points:               req.Points,
		OriginalPoints:       &req.Points,
		MultiplierSnapshot:   multiplierSnapshot,
		HandicapSnapshot:     handicapSnapshot,
		HandicapTeamSnapshot: handicapTeamSnapshot,
	}

	if err := s.repo.CreatePrediction(nil, bet); err != nil {
		return nil, fmt.Errorf("failed to submit prediction (may be duplicate): %w", err)
	}

	if s.hub != nil {
		event := buildPredictionActivityEvent(wcUserID.String(), user.Name, req, m)
		s.hub.Broadcast(event)
	}

	return bet, nil
}

func (s *WcService) ListPredictions(wcUserID uuid.UUID) ([]*model.WcPredictionWithMatch, error) {
	predictions, err := s.repo.ListPredictions(wcUserID)
	if err != nil {
		return nil, err
	}
	if s.customBetRepo != nil {
		customEntries, err := s.customBetRepo.ListCustomEntriesForUserAsHistory(wcUserID)
		if err == nil {
			predictions = append(predictions, customEntries...)
		}
	}
	return predictions, nil
}

func (s *WcService) DeletePrediction(wcUserID, betID uuid.UUID) (penaltyApplied float64, err error) {
	bet, err := s.repo.GetPredictionByID(betID)
	if err != nil {
		return 0, fmt.Errorf("prediction not found")
	}
	if bet.WcUserID != wcUserID {
		return 0, fmt.Errorf("unauthorized")
	}
	if bet.Result != nil {
		return 0, fmt.Errorf("cannot delete a finalized prediction")
	}
	if bet.CancelledAt != nil {
		return 0, fmt.Errorf("prediction already cancelled")
	}
	m, err := s.repo.GetMatch(bet.MatchID)
	if err != nil {
		return 0, fmt.Errorf("match not found")
	}
	if isLocked(m) {
		return 0, fmt.Errorf("cannot modify prediction: match is locked")
	}

	cfg, err := s.repo.GetConfig()
	if err != nil {
		return 0, fmt.Errorf("failed to load config")
	}

	penalty := computeCancelPenalty(bet.Points, cfg.CancelPenaltyPercent, cfg.CancelPenaltyEnabled)

	db := s.repo.DB()
	err = db.Transaction(func(tx *gorm.DB) error {
		if err := s.repo.SoftCancelPrediction(tx, betID, wcUserID, penalty); err != nil {
			return err
		}
		if penalty > 0 {
			wallet, err := s.repo.GetWalletTx(tx, wcUserID)
			if err != nil {
				return fmt.Errorf("wallet not found")
			}
			balanceBefore := wallet.Balance
			if err := s.repo.UpdateWalletBalance(tx, wcUserID, -penalty); err != nil {
				return err
			}
			return s.repo.LogWalletChange(tx, &model.WcWalletLog{
				WcUserID:      wcUserID,
				AdminID:       uuid.Nil,
				Delta:         -penalty,
				BalanceBefore: balanceBefore,
				BalanceAfter:  balanceBefore - penalty,
				Note:          fmt.Sprintf("prediction cancel penalty — %d%%", cfg.CancelPenaltyPercent),
			})
		}
		return nil
	})
	if err != nil {
		return 0, err
	}

	if s.hub != nil {
		user, _ := s.userRepo.GetByID(wcUserID)
		userName := ""
		if user != nil {
			userName = user.Name
		}
		s.hub.Broadcast(buildCancelActivityEvent(wcUserID.String(), userName, bet.PredictionType, m.HomeTeam, m.AwayTeam, bet.MatchID.String()))
	}
	return penalty, nil
}

func (s *WcService) UpdatePredictionPoints(wcUserID, betID uuid.UUID, points int) (penaltyApplied float64, err error) {
	cfg, err := s.repo.GetConfig()
	if err != nil {
		return 0, fmt.Errorf("failed to load config")
	}
	if points < cfg.MinPoints || points > cfg.MaxPoints {
		return 0, fmt.Errorf("điểm cược phải từ %d đến %d", cfg.MinPoints, cfg.MaxPoints)
	}
	bet, err := s.repo.GetPredictionByID(betID)
	if err != nil {
		return 0, fmt.Errorf("prediction not found")
	}
	if bet.WcUserID != wcUserID {
		return 0, fmt.Errorf("unauthorized")
	}
	if bet.Result != nil {
		return 0, fmt.Errorf("cannot modify a finalized prediction")
	}
	if bet.CancelledAt != nil {
		return 0, fmt.Errorf("prediction is cancelled")
	}
	m, err := s.repo.GetMatch(bet.MatchID)
	if err != nil {
		return 0, fmt.Errorf("match not found")
	}
	if isLocked(m) {
		return 0, fmt.Errorf("cannot modify prediction: match is locked")
	}

	var penalty float64
	if points < bet.Points && bet.OriginalPoints != nil {
		penalty, _, _ = computeReducePenalty(*bet.OriginalPoints, points, cfg.BetReduceMaxPercent, cfg.BetReducePenaltyPercent)
	}

	if penalty > 0 {
		db := s.repo.DB()
		if err := db.Transaction(func(tx *gorm.DB) error {
			// Soft-cancel old prediction, storing penalty as cancel_penalty
			if err := s.repo.SoftCancelPrediction(tx, betID, wcUserID, penalty); err != nil {
				return err
			}
			// Deduct penalty from wallet
			wallet, err := s.repo.GetWalletTx(tx, wcUserID)
			if err != nil {
				return fmt.Errorf("wallet not found")
			}
			balanceBefore := wallet.Balance
			if err := s.repo.UpdateWalletBalance(tx, wcUserID, -penalty); err != nil {
				return err
			}
			if err := s.repo.LogWalletChange(tx, &model.WcWalletLog{
				WcUserID:      wcUserID,
				AdminID:       uuid.Nil,
				Delta:         -penalty,
				BalanceBefore: balanceBefore,
				BalanceAfter:  balanceBefore - penalty,
				Note:          fmt.Sprintf("prediction reduce penalty — points %d→%d", bet.Points, points),
			}); err != nil {
				return err
			}
			// Create new prediction with reduced points
			newBet := &model.WcPrediction{
				WcUserID:             bet.WcUserID,
				MatchID:              bet.MatchID,
				PredictionType:       bet.PredictionType,
				PredictionChoice:     bet.PredictionChoice,
				PredictedHomeScore:   bet.PredictedHomeScore,
				PredictedAwayScore:   bet.PredictedAwayScore,
				Points:               points,
				OriginalPoints:       &points,
				MultiplierSnapshot:   bet.MultiplierSnapshot,
				HandicapSnapshot:     bet.HandicapSnapshot,
				HandicapTeamSnapshot: bet.HandicapTeamSnapshot,
			}
			return s.repo.CreatePrediction(tx, newBet)
		}); err != nil {
			return 0, err
		}
		return penalty, nil
	}
	return 0, s.repo.UpdatePredictionPoints(betID, points)
}

// PreviewReducePredictionPoints computes the penalty for reducing a prediction's points (dry-run).
func (s *WcService) PreviewReducePredictionPoints(wcUserID, betID uuid.UUID, newPoints int) (*ReduceStakePreview, error) {
	bet, err := s.repo.GetPredictionByID(betID)
	if err != nil {
		return nil, fmt.Errorf("prediction not found")
	}
	if bet.WcUserID != wcUserID {
		return nil, fmt.Errorf("unauthorized")
	}
	if bet.Result != nil {
		return nil, fmt.Errorf("cannot modify a finalized prediction")
	}
	if bet.CancelledAt != nil {
		return nil, fmt.Errorf("prediction is cancelled")
	}

	if newPoints >= bet.Points || bet.OriginalPoints == nil {
		return &ReduceStakePreview{}, nil
	}

	cfg, err := s.repo.GetConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to load config")
	}

	penalty, excessReduction, allowedMinStake := computeReducePenalty(*bet.OriginalPoints, newPoints, cfg.BetReduceMaxPercent, cfg.BetReducePenaltyPercent)
	return &ReduceStakePreview{
		Penalty:         penalty,
		ExcessReduction: excessReduction,
		AllowedMinStake: allowedMinStake,
	}, nil
}

func (s *WcService) ListPredictionsForMatchPublic(matchID uuid.UUID) ([]*model.WcPredictionPublic, error) {
	predictions, err := s.repo.ListPredictionsForMatchPublic(matchID)
	if err != nil {
		return nil, err
	}
	if s.customBetRepo != nil {
		customEntries, err := s.customBetRepo.ListCustomEntriesForMatchPublic(matchID)
		if err == nil {
			predictions = append(predictions, customEntries...)
		}
	}
	return predictions, nil
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

type FinalizeMatchResult struct {
	PredictionsProcessed     int     `json:"predictions_processed"`
	TotalPointsAwarded       float64 `json:"total_points_awarded"`
	UnsettledCustomBetsCount int64   `json:"unsettled_custom_bets_count"`
}

// FinalizeMatch evaluates all predictions on a match, credits/debits wallets, and marks settled_at.
// Returns an error if the match has already been finalized.
func (s *WcService) FinalizeMatch(matchID uuid.UUID) (*FinalizeMatchResult, error) {
	m, err := s.repo.GetMatch(matchID)
	if err != nil {
		return nil, fmt.Errorf("match not found")
	}
	if m.SettledAt != nil {
		return nil, fmt.Errorf("match already finalized at %s", m.SettledAt.Format("2006-01-02 15:04:05"))
	}
	if m.HomeScore == nil || m.AwayScore == nil {
		return nil, fmt.Errorf("match score not set — cannot settle")
	}

	bets, err := s.repo.ListPredictionsForMatch(matchID)
	if err != nil {
		return nil, err
	}

	db := s.repo.DB()
	res := &FinalizeMatchResult{}

	err = db.Transaction(func(tx *gorm.DB) error {
		for _, bet := range bets {
			var result string
			var pointsEarned float64

			switch bet.PredictionType {
			case model.WcPredictionTypeHandicap:
				result, pointsEarned = evaluateHandicapPrediction(bet, *m.HomeScore, *m.AwayScore)
			case model.WcPredictionTypeExactScore:
				result, pointsEarned = evaluateExactScorePrediction(bet, *m.HomeScore, *m.AwayScore)
			case model.WcPredictionTypeOverUnder:
				result, pointsEarned = evaluateOverUnderPrediction(bet, *m.HomeScore, *m.AwayScore)
			}

			if err := s.repo.UpdatePredictionResult(tx, bet.ID, result, pointsEarned); err != nil {
				return err
			}

			w, err := s.repo.GetWalletTx(tx, bet.WcUserID)
			if err != nil {
				return err
			}
			balanceBefore := w.Balance

			if pointsEarned > 0 {
				if err := s.repo.UpdateWalletBalance(tx, bet.WcUserID, pointsEarned); err != nil {
					return err
				}
			}
			if err := s.repo.UpdateWalletBalance(tx, bet.WcUserID, float64(-bet.Points)); err != nil {
				return err
			}

			netChange := pointsEarned - float64(bet.Points)
			if err := s.repo.LogWalletChange(tx, &model.WcWalletLog{
				WcUserID:      bet.WcUserID,
				AdminID:       uuid.Nil,
				Delta:         netChange,
				BalanceBefore: balanceBefore,
				BalanceAfter:  balanceBefore + netChange,
				Note:          "prediction settle — " + result,
			}); err != nil {
				return err
			}

			res.TotalPointsAwarded += pointsEarned
			res.PredictionsProcessed++
		}

		// Mark match as settled
		now := time.Now()
		return s.repo.UpdateMatch(matchID, map[string]interface{}{
			"settled_at": now,
			"status":     model.WcStatusCompleted,
		})
	})
	if err != nil {
		return nil, err
	}

	// Check for unsettled custom bets on this match
	if s.customBetRepo != nil {
		res.UnsettledCustomBetsCount, _ = s.customBetRepo.CountUnsettledForMatch(matchID)
	}

	return res, nil
}

// FinalizeAllResult summarises a bulk-finalize run.
type FinalizeAllResult struct {
	Processed                      int     `json:"processed"`
	Skipped                        int     `json:"skipped"`
	TotalPointsAwarded             float64 `json:"total_points_awarded"`
	MatchesWithUnsettledCustomBets int64   `json:"matches_with_unsettled_custom_bets"`
}

// FinalizeAllMatches finalizes every scored-but-not-yet-settled match in one call.
// Matches already finalized are counted as skipped (not an error).
func (s *WcService) FinalizeAllMatches() (*FinalizeAllResult, error) {
	matches, err := s.repo.ListUnfinalizedScoredMatches()
	if err != nil {
		return nil, err
	}
	result := &FinalizeAllResult{}
	for _, m := range matches {
		r, err := s.FinalizeMatch(m.ID)
		if err != nil {
			result.Skipped++
			continue
		}
		result.Processed++
		result.TotalPointsAwarded += r.TotalPointsAwarded
	}
	// Count all matches (not just finalized ones) that still have unsettled custom bets
	if s.customBetRepo != nil {
		result.MatchesWithUnsettledCustomBets, _ = s.customBetRepo.CountMatchesWithUnsettled()
	}
	return result, nil
}

// RefinalizeAllMatches re-calculates points_earned for every scored match (including already settled).
// For already-settled predictions, it reverses the old wallet change before applying the correct value.
// Use this to fix data from before fractional points_earned was supported.
func (s *WcService) RefinalizeAllMatches() (*FinalizeAllResult, error) {
	matches, err := s.repo.ListAllScoredMatches()
	if err != nil {
		return nil, err
	}
	result := &FinalizeAllResult{}
	db := s.repo.DB()

	for _, m := range matches {
		if m.HomeScore == nil || m.AwayScore == nil {
			result.Skipped++
			continue
		}
		bets, err := s.repo.ListPredictionsForMatch(m.ID)
		if err != nil {
			result.Skipped++
			continue
		}

		var totalPointsEarned float64
		err = db.Transaction(func(tx *gorm.DB) error {
			for _, bet := range bets {
				// Determine new result first; skip unknown types without touching wallet
				var res string
				var pointsEarned float64
				switch bet.PredictionType {
				case model.WcPredictionTypeHandicap:
					res, pointsEarned = evaluateHandicapPrediction(bet, *m.HomeScore, *m.AwayScore)
				case model.WcPredictionTypeExactScore:
					res, pointsEarned = evaluateExactScorePrediction(bet, *m.HomeScore, *m.AwayScore)
				case model.WcPredictionTypeOverUnder:
					res, pointsEarned = evaluateOverUnderPrediction(bet, *m.HomeScore, *m.AwayScore)
				default:
					continue
				}

				w, err := s.repo.GetWalletTx(tx, bet.WcUserID)
				if err != nil {
					return err
				}
				balanceBefore := w.Balance

				// Reverse old wallet change if already settled.
				// Treat NULL points_earned as 0 (covers old data that stored NULL for losses).
				var prevNet float64
				if bet.Result != nil {
					var prevEarned float64
					if bet.PointsEarned != nil {
						prevEarned = *bet.PointsEarned
					}
					prevNet = prevEarned - float64(bet.Points)
					if err := s.repo.UpdateWalletBalance(tx, bet.WcUserID, -prevNet); err != nil {
						return err
					}
				}

				if err := s.repo.UpdatePredictionResult(tx, bet.ID, res, pointsEarned); err != nil {
					return err
				}
				if pointsEarned > 0 {
					if err := s.repo.UpdateWalletBalance(tx, bet.WcUserID, pointsEarned); err != nil {
						return err
					}
				}
				if err := s.repo.UpdateWalletBalance(tx, bet.WcUserID, float64(-bet.Points)); err != nil {
					return err
				}

				newNet := pointsEarned - float64(bet.Points)
				effectiveDelta := newNet - prevNet
				noteStr := "prediction settle — " + res
				if bet.Result != nil {
					noteStr = "prediction refinalize — " + res
				}
				if err := s.repo.LogWalletChange(tx, &model.WcWalletLog{
					WcUserID:      bet.WcUserID,
					AdminID:       uuid.Nil,
					Delta:         effectiveDelta,
					BalanceBefore: balanceBefore,
					BalanceAfter:  balanceBefore + effectiveDelta,
					Note:          noteStr,
				}); err != nil {
					return err
				}

				totalPointsEarned += pointsEarned
			}

			now := time.Now()
			return s.repo.UpdateMatch(m.ID, map[string]interface{}{
				"settled_at": now,
				"status":     model.WcStatusCompleted,
			})
		})

		if err != nil {
			result.Skipped++
			continue
		}
		result.Processed++
		result.TotalPointsAwarded += totalPointsEarned
	}
	return result, nil
}

// --- Preview (dry-run) ---

func buildPreviewRow(bet *model.WcPrediction, homeScore, awayScore int) model.FinalizePreviewRow {
	var result string
	var pointsEarned float64
	switch bet.PredictionType {
	case model.WcPredictionTypeHandicap:
		result, pointsEarned = evaluateHandicapPrediction(bet, homeScore, awayScore)
	case model.WcPredictionTypeExactScore:
		result, pointsEarned = evaluateExactScorePrediction(bet, homeScore, awayScore)
	case model.WcPredictionTypeOverUnder:
		result, pointsEarned = evaluateOverUnderPrediction(bet, homeScore, awayScore)
	default:
		result = "void"
	}
	return model.FinalizePreviewRow{
		WcUserID:        bet.WcUserID,
		UserName:        "",      // populated by caller
		PredictionType:  bet.PredictionType,
		Points:          bet.Points,
		Multiplier:      bet.MultiplierSnapshot,
		NewResult:       result,
		NewPointsEarned: pointsEarned,
		NetDelta:        pointsEarned - float64(bet.Points),
	}
}

func buildPreviewResult(matches []*model.WcMatch, getPredictions func(uuid.UUID) ([]*model.WcPrediction, error), getUserName func(uuid.UUID) string, excludeSettled bool) (*model.FinalizePreviewResult, error) {
	result := &model.FinalizePreviewResult{
		Matches: []model.FinalizePreviewMatch{},
	}
	for _, m := range matches {
		if m.HomeScore == nil || m.AwayScore == nil {
			continue
		}
		bets, err := getPredictions(m.ID)
		if err != nil {
			return nil, err
		}
		pm := model.FinalizePreviewMatch{
			MatchID:        m.ID.String(),
			HomeTeam:       m.HomeTeam,
			AwayTeam:       m.AwayTeam,
			HomeScore:      *m.HomeScore,
			AwayScore:      *m.AwayScore,
			Stage:          m.Stage,
			AlreadySettled: m.SettledAt != nil,
			Predictions:    []model.FinalizePreviewRow{},
		}
		countInSummary := !excludeSettled || m.SettledAt == nil
		for _, bet := range bets {
			row := buildPreviewRow(bet, *m.HomeScore, *m.AwayScore)
			row.UserName = getUserName(bet.WcUserID)
			pm.Predictions = append(pm.Predictions, row)
			if countInSummary {
				result.HouseSummary.TotalStaked += float64(bet.Points)
				result.HouseSummary.TotalPaidOut += row.NewPointsEarned
				result.HouseSummary.PredictionCount++
			}
		}
		result.Matches = append(result.Matches, pm)
		if countInSummary {
			result.HouseSummary.MatchCount++
		}
	}
	result.HouseSummary.HouseNet = result.HouseSummary.TotalStaked - result.HouseSummary.TotalPaidOut
	return result, nil
}

// PreviewFinalizeMatch computes what FinalizeMatch would do for a single match (read-only).
func (s *WcService) PreviewFinalizeMatch(matchID uuid.UUID) (*model.FinalizePreviewResult, error) {
	m, err := s.repo.GetMatch(matchID)
	if err != nil {
		return nil, fmt.Errorf("match not found")
	}
	if m.HomeScore == nil || m.AwayScore == nil {
		return nil, fmt.Errorf("match score not set — cannot preview")
	}
	users, _ := s.userRepo.GetAll()
	nameMap := make(map[uuid.UUID]string, len(users))
	for _, u := range users {
		nameMap[u.ID] = u.Name
	}
	return buildPreviewResult(
		[]*model.WcMatch{m},
		func(id uuid.UUID) ([]*model.WcPrediction, error) {
			return s.repo.ListPredictionsForMatch(id)
		},
		func(id uuid.UUID) string { return nameMap[id] },
		true,
	)
}

// PreviewFinalizeAll computes what FinalizeAllMatches would do (read-only).
func (s *WcService) PreviewFinalizeAll() (*model.FinalizePreviewResult, error) {
	matches, err := s.repo.ListUnfinalizedScoredMatches()
	if err != nil {
		return nil, err
	}
	users, _ := s.userRepo.GetAll()
	nameMap := make(map[uuid.UUID]string, len(users))
	for _, u := range users {
		nameMap[u.ID] = u.Name
	}
	return buildPreviewResult(
		matches,
		func(id uuid.UUID) ([]*model.WcPrediction, error) {
			return s.repo.ListPredictionsForMatch(id)
		},
		func(id uuid.UUID) string { return nameMap[id] },
		true,
	)
}

// PreviewRefinalizeAll computes what RefinalizeAllMatches would do (read-only).
func (s *WcService) PreviewRefinalizeAll() (*model.FinalizePreviewResult, error) {
	matches, err := s.repo.ListAllScoredMatches()
	if err != nil {
		return nil, err
	}
	users, _ := s.userRepo.GetAll()
	nameMap := make(map[uuid.UUID]string, len(users))
	for _, u := range users {
		nameMap[u.ID] = u.Name
	}
	return buildPreviewResult(
		matches,
		func(id uuid.UUID) ([]*model.WcPrediction, error) {
			return s.repo.ListPredictionsForMatch(id)
		},
		func(id uuid.UUID) string { return nameMap[id] },
		false,
	)
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

func (s *WcService) MarkSettlementDone(settlementID, wcUserID uuid.UUID, status, doneNote string) error {
	if status != model.WcSettlementStatusDone && status != model.WcSettlementStatusPending {
		return fmt.Errorf("status must be 'done' or 'pending'")
	}
	return s.repo.UpdateSettlementDetailStatus(settlementID, wcUserID, status, doneNote)
}

// --- Wallet admin ops ---

func (s *WcService) AdminTopUp(adminID, wcUserID uuid.UUID, delta int, note string) error {
	if delta == 0 {
		return fmt.Errorf("delta cannot be 0")
	}
	deltaF := float64(delta)
	db := s.repo.DB()
	return db.Transaction(func(tx *gorm.DB) error {
		wallet, err := s.repo.GetWalletTx(tx, wcUserID)
		if err != nil {
			return fmt.Errorf("wallet not found for user")
		}
		balanceBefore := wallet.Balance
		if err := s.repo.UpdateWalletBalance(tx, wcUserID, deltaF); err != nil {
			return err
		}
		return s.repo.LogWalletChange(tx, &model.WcWalletLog{
			WcUserID:      wcUserID,
			AdminID:       adminID,
			Delta:         deltaF,
			BalanceBefore: balanceBefore,
			BalanceAfter:  balanceBefore + deltaF,
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

func (s *WcService) SetUserBot(wcUserID uuid.UUID, isBot bool) error {
	return s.userRepo.SetBot(wcUserID, isBot)
}

// --- User block/unblock ---

// BlockUser blocks targetID: voids all pending bets (no wallet change — deferred deduction means
// stake is never taken at placement), then sets is_blocked=true. Returns the count of bets voided.
func (s *WcService) BlockUser(adminID, targetID uuid.UUID) (int, error) {
	if adminID == targetID {
		return 0, fmt.Errorf("cannot block yourself")
	}
	db := s.repo.DB()
	voidedCount := 0
	err := db.Transaction(func(tx *gorm.DB) error {
		pendingBets, err := s.repo.ListPendingBetsForUser(tx, targetID)
		if err != nil {
			return err
		}
		for _, bet := range pendingBets {
			if err := s.repo.VoidBet(tx, bet.ID, bet.Stake); err != nil {
				return err
			}
			// No wallet change: wc_bets uses deferred deduction — stake is never taken at
			// placement, so there is nothing to refund when voiding on block.
			voidedCount++
		}
		return s.userRepo.SetBlockedTx(tx, targetID, true)
	})
	return voidedCount, err
}

func (s *WcService) UnblockUser(targetID uuid.UUID) error {
	return s.userRepo.SetBlocked(targetID, false)
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
	user, err := s.userRepo.GetByID(wcUserID)
	if err != nil {
		return nil, fmt.Errorf("user not found")
	}
	if user.IsBlocked {
		return nil, fmt.Errorf("user is blocked from placing bets")
	}
	betCfg, err := s.repo.GetConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to load config")
	}
	if req.Stake < betCfg.MinPoints || req.Stake > betCfg.MaxPoints {
		return nil, fmt.Errorf("điểm cược phải từ %d đến %d", betCfg.MinPoints, betCfg.MaxPoints)
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
		so, err := s.repo.GetScoreMultiplier(req.MatchID, *req.PredictedHomeScore, *req.PredictedAwayScore)
		if err != nil {
			return nil, fmt.Errorf("scoreline %d:%d is not available for this match", *req.PredictedHomeScore, *req.PredictedAwayScore)
		}
		oddsSnapshot = so.Multiplier
	case model.WcBetTypeOverUnder:
		if req.BetChoice == nil || (*req.BetChoice != model.WcChoiceOver && *req.BetChoice != model.WcChoiceUnder) {
			return nil, fmt.Errorf("bet_choice must be 'over' or 'under'")
		}
		if m.OULine == nil || m.OddsOver == nil || m.OddsUnder == nil {
			return nil, fmt.Errorf("O/U odds not configured for this match")
		}
		if *req.BetChoice == model.WcChoiceOver {
			oddsSnapshot = *m.OddsOver
		} else {
			oddsSnapshot = *m.OddsUnder
		}
		handicapSnapshot = m.OULine
	default:
		return nil, fmt.Errorf("bet_type must be 'handicap', 'exact_score', or 'over_under'")
	}

	originalStake := req.Stake
	bet := &model.WcBet{
		WcUserID:             wcUserID,
		MatchID:              req.MatchID,
		BetType:              req.BetType,
		BetChoice:            req.BetChoice,
		Stake:                req.Stake,
		OriginalStake:        &originalStake,
		OddsSnapshot:         oddsSnapshot,
		HandicapSnapshot:     handicapSnapshot,
		HandicapTeamSnapshot: handicapTeamSnapshot,
		PredictedHomeScore:   req.PredictedHomeScore,
		PredictedAwayScore:   req.PredictedAwayScore,
	}
	if err := s.repo.CreateBet(bet); err != nil {
		return nil, fmt.Errorf("failed to place bet (may be duplicate): %w", err)
	}

	if s.hub != nil {
		event := buildActivityEvent(wcUserID.String(), user.Name, req, m)
		s.hub.Broadcast(event)
	}

	return bet, nil
}

func (s *WcService) ListBets(wcUserID uuid.UUID) ([]*model.WcBetWithMatch, error) {
	return s.repo.ListBets(wcUserID)
}

func (s *WcService) ListBetsForMatch(matchID uuid.UUID) ([]*model.WcBetPublic, error) {
	return s.repo.ListBetsForMatch(matchID)
}

// ReduceStakePreview is returned by PreviewReduceStake (dry-run before confirm).
type ReduceStakePreview struct {
	Penalty         float64 `json:"penalty"`
	ExcessReduction int     `json:"excess_reduction"`
	AllowedMinStake int     `json:"allowed_min_stake"`
}

func (s *WcService) UpdateBetStake(wcUserID, betID uuid.UUID, stake int) error {
	cfg, err := s.repo.GetConfig()
	if err != nil {
		return fmt.Errorf("failed to load config")
	}
	if stake < cfg.MinPoints || stake > cfg.MaxPoints {
		return fmt.Errorf("điểm cược phải từ %d đến %d", cfg.MinPoints, cfg.MaxPoints)
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
	if bet.CancelledAt != nil {
		return fmt.Errorf("bet is cancelled")
	}
	m, err := s.repo.GetMatch(bet.MatchID)
	if err != nil {
		return fmt.Errorf("match not found")
	}
	if isBetLocked(m) {
		return fmt.Errorf("betting is closed for this match")
	}

	var penalty float64
	if stake < bet.Stake && bet.OriginalStake != nil {
		penalty, _, _ = computeReducePenalty(*bet.OriginalStake, stake, cfg.BetReduceMaxPercent, cfg.BetReducePenaltyPercent)
	}

	if penalty > 0 {
		db := s.repo.DB()
		return db.Transaction(func(tx *gorm.DB) error {
			if err := s.repo.UpdateBetStake(tx, betID, stake); err != nil {
				return err
			}
			wallet, err := s.repo.GetWalletTx(tx, wcUserID)
			if err != nil {
				return fmt.Errorf("wallet not found")
			}
			balanceBefore := wallet.Balance
			if err := s.repo.UpdateWalletBalance(tx, wcUserID, -penalty); err != nil {
				return err
			}
			return s.repo.LogWalletChange(tx, &model.WcWalletLog{
				WcUserID:      wcUserID,
				AdminID:       uuid.Nil,
				Delta:         -penalty,
				BalanceBefore: balanceBefore,
				BalanceAfter:  balanceBefore - penalty,
				Note:          fmt.Sprintf("bet reduce penalty — stake %d→%d", bet.Stake, stake),
			})
		})
	}
	return s.repo.UpdateBetStake(nil, betID, stake)
}

// PreviewReduceStake computes the penalty for reducing a bet's stake (dry-run, no DB writes).
func (s *WcService) PreviewReduceStake(wcUserID, betID uuid.UUID, newStake int) (*ReduceStakePreview, error) {
	bet, err := s.repo.GetBet(betID)
	if err != nil {
		return nil, fmt.Errorf("bet not found")
	}
	if bet.WcUserID != wcUserID {
		return nil, fmt.Errorf("unauthorized")
	}
	if bet.Result != nil {
		return nil, fmt.Errorf("cannot modify a settled bet")
	}
	if bet.CancelledAt != nil {
		return nil, fmt.Errorf("bet is cancelled")
	}

	if newStake >= bet.Stake || bet.OriginalStake == nil {
		return &ReduceStakePreview{}, nil
	}

	cfg, err := s.repo.GetConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to load config")
	}

	penalty, excessReduction, allowedMinStake := computeReducePenalty(*bet.OriginalStake, newStake, cfg.BetReduceMaxPercent, cfg.BetReducePenaltyPercent)
	return &ReduceStakePreview{
		Penalty:         penalty,
		ExcessReduction: excessReduction,
		AllowedMinStake: allowedMinStake,
	}, nil
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
	if bet.CancelledAt != nil {
		return fmt.Errorf("bet already cancelled")
	}
	m, err := s.repo.GetMatch(bet.MatchID)
	if err != nil {
		return fmt.Errorf("match not found")
	}
	if isBetLocked(m) {
		return fmt.Errorf("betting is closed for this match")
	}

	cfg, err := s.repo.GetConfig()
	if err != nil {
		return fmt.Errorf("failed to load config")
	}

	penalty := computeCancelPenalty(bet.Stake, cfg.CancelPenaltyPercent, cfg.CancelPenaltyEnabled)

	db := s.repo.DB()
	err = db.Transaction(func(tx *gorm.DB) error {
		if err := s.repo.SoftCancelBet(tx, betID, wcUserID, penalty); err != nil {
			return err
		}
		if penalty > 0 {
			wallet, err := s.repo.GetWalletTx(tx, wcUserID)
			if err != nil {
				return fmt.Errorf("wallet not found")
			}
			balanceBefore := wallet.Balance
			if err := s.repo.UpdateWalletBalance(tx, wcUserID, -penalty); err != nil {
				return err
			}
			return s.repo.LogWalletChange(tx, &model.WcWalletLog{
				WcUserID:      wcUserID,
				AdminID:       uuid.Nil,
				Delta:         -penalty,
				BalanceBefore: balanceBefore,
				BalanceAfter:  balanceBefore - penalty,
				Note:          fmt.Sprintf("bet cancel penalty — %d%%", cfg.CancelPenaltyPercent),
			})
		}
		return nil
	})
	if err != nil {
		return err
	}

	if s.hub != nil {
		user, _ := s.userRepo.GetByID(wcUserID)
		userName := ""
		if user != nil {
			userName = user.Name
		}
		s.hub.Broadcast(buildCancelActivityEvent(wcUserID.String(), userName, bet.BetType, m.HomeTeam, m.AwayTeam, bet.MatchID.String()))
	}
	return nil
}

// GetBetHistory returns the combined settled + cancelled history for a user
// (regular bets + custom bet entries), sorted chronologically newest-first.
func (s *WcService) GetBetHistory(wcUserID uuid.UUID) ([]model.BetHistoryItem, error) {
	regularBets, err := s.repo.ListBetHistoryForUser(wcUserID)
	if err != nil {
		return nil, err
	}

	var items []model.BetHistoryItem

	for _, b := range regularBets {
		betType := b.BetType
		item := model.BetHistoryItem{
			ID:                 b.ID.String(),
			Kind:               "regular",
			MatchID:            b.MatchID.String(),
			HomeTeam:           b.HomeTeam,
			AwayTeam:           b.AwayTeam,
			MatchDate:          b.MatchDate,
			BetType:            &betType,
			BetChoice:          b.BetChoice,
			Stake:              b.Stake,
			OriginalStake:      b.OriginalStake,
			OddsSnapshot:       b.OddsSnapshot,
			PredictedHomeScore: b.PredictedHomeScore,
			PredictedAwayScore: b.PredictedAwayScore,
			Result:             b.Result,
			Payout:             b.Payout,
			CancelledAt:        b.CancelledAt,
			CancelPenalty:      b.CancelPenalty,
			CreatedAt:          b.CreatedAt,
		}
		items = append(items, item)
	}

	if s.customBetRepo != nil {
		customEntries, err := s.customBetRepo.ListCancelledOrSettledEntriesForUser(wcUserID)
		if err == nil {
			for _, e := range customEntries {
				betTitle := e.BetTitle
				optLabel := e.OptionLabel
				var result *string
				switch e.Status {
				case "won":
					r := "win"
					result = &r
				case "lost":
					r := "lose"
					result = &r
				case "void":
					r := "void"
					result = &r
				}
				item := model.BetHistoryItem{
					ID:            e.ID.String(),
					Kind:          "custom",
					MatchID:       e.MatchID.String(),
					HomeTeam:      e.HomeTeam,
					AwayTeam:      e.AwayTeam,
					MatchDate:     e.MatchDate,
					BetTitle:      &betTitle,
					OptionLabel:   &optLabel,
					Stake:         e.Stake,
					OriginalStake: e.OriginalStake,
					OddsSnapshot:  e.OddsSnapshot,
					Result:        result,
					Payout:        e.Payout,
					CancelledAt:   e.CancelledAt,
					CancelPenalty: e.CancelPenalty,
					CreatedAt:     e.CreatedAt,
				}
				items = append(items, item)
			}
		}
	}

	sortBetHistoryNewestFirst(items)

	if items == nil {
		items = []model.BetHistoryItem{}
	}
	return items, nil
}

// sortBetHistoryNewestFirst sorts BetHistoryItems in-place with newest CreatedAt first.
func sortBetHistoryNewestFirst(items []model.BetHistoryItem) {
	for i := 1; i < len(items); i++ {
		for j := i; j > 0 && items[j].CreatedAt.After(items[j-1].CreatedAt); j-- {
			items[j], items[j-1] = items[j-1], items[j]
		}
	}
}

// SetPenaltyConfig updates the cancel + reduce-stake penalty settings.
func (s *WcService) SetPenaltyConfig(cancelEnabled bool, cancelPercent, reduceMaxPercent, reducePenaltyPercent int, updatedBy uuid.UUID) error {
	if err := validatePenaltyConfig(cancelPercent, reduceMaxPercent, reducePenaltyPercent); err != nil {
		return err
	}
	return s.repo.UpdatePenaltyConfig(cancelEnabled, cancelPercent, reduceMaxPercent, reducePenaltyPercent, &updatedBy)
}

// BackfillOriginalPoints sets original_points = points for predictions where original_points IS NULL.
// Safe to call multiple times (idempotent). Returns number of rows updated.
func (s *WcService) BackfillOriginalPoints() (int64, error) {
	return s.repo.BackfillOriginalPoints()
}

// validatePenaltyConfig checks that all three percentage values are in [0, 100].
func validatePenaltyConfig(cancelPercent, reduceMaxPercent, reducePenaltyPercent int) error {
	if cancelPercent < 0 || cancelPercent > 100 {
		return fmt.Errorf("cancel_penalty_percent phải từ 0 đến 100")
	}
	if reduceMaxPercent < 0 || reduceMaxPercent > 100 {
		return fmt.Errorf("bet_reduce_max_percent phải từ 0 đến 100")
	}
	if reducePenaltyPercent < 0 || reducePenaltyPercent > 100 {
		return fmt.Errorf("bet_reduce_penalty_percent phải từ 0 đến 100")
	}
	return nil
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
			w, err := s.repo.GetWalletTx(tx, bet.WcUserID)
			if err != nil {
				return err
			}
			balanceBefore := w.Balance

			// Reverse previous settlement for idempotency
			var prevNet float64
			if bet.Result != nil && bet.Payout != nil {
				prevNet = *bet.Payout - float64(bet.Stake)
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
			case model.WcBetTypeOverUnder:
				result, payout = evaluateOverUnderBet(bet, *m.HomeScore, *m.AwayScore)
			}

			if err := s.repo.UpdateBetResult(tx, bet.ID, result, payout); err != nil {
				return err
			}

			netChange := payout - float64(bet.Stake)
			if err := s.repo.UpdateWalletBalance(tx, bet.WcUserID, netChange); err != nil {
				return err
			}

			effectiveDelta := netChange - prevNet
			noteStr := "bet settle — " + result
			if bet.Result != nil {
				noteStr = "bet re-settle — " + result
			}
			if err := s.repo.LogWalletChange(tx, &model.WcWalletLog{
				WcUserID:      bet.WcUserID,
				AdminID:       uuid.Nil,
				Delta:         effectiveDelta,
				BalanceBefore: balanceBefore,
				BalanceAfter:  balanceBefore + effectiveDelta,
				Note:          noteStr,
			}); err != nil {
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

// --- Helpers ---

func isBetLocked(m *model.WcMatch) bool {
	if m.Status == model.WcStatusLive || m.Status == model.WcStatusCompleted || m.Status == model.WcStatusCancelled {
		return true
	}
	if m.BetsLockedAt != nil && time.Now().After(*m.BetsLockedAt) {
		return true
	}
	return false
}

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

// round2 rounds a float64 to 2 decimal places.
func round2(v float64) float64 {
	return math.Round(v*100) / 100
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
func evaluateHandicapPrediction(bet *model.WcPrediction, homeScore, awayScore int) (string, float64) {
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
			return model.WcResultCorrect, round2(float64(bet.Points) * bet.MultiplierSnapshot)
		case r1 == "lose" && r2 == "lose":
			return model.WcResultIncorrect, 0
		case (r1 == "win" || r2 == "win"):
			// one win + one push
			return model.WcResultWinHalf, round2(half*bet.MultiplierSnapshot + half)
		default:
			// one push + one lose
			return model.WcResultLoseHalf, round2(half)
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
		return model.WcResultVoid, float64(bet.Points)
	}

	if winner == choice {
		return model.WcResultCorrect, round2(float64(bet.Points) * bet.MultiplierSnapshot)
	}
	return model.WcResultIncorrect, 0
}

// evaluateExactScorePrediction checks exact score for the prediction system.
func evaluateExactScorePrediction(bet *model.WcPrediction, homeScore, awayScore int) (string, float64) {
	if bet.PredictedHomeScore == nil || bet.PredictedAwayScore == nil {
		return model.WcResultIncorrect, 0
	}
	if *bet.PredictedHomeScore == homeScore && *bet.PredictedAwayScore == awayScore {
		return model.WcResultCorrect, round2(float64(bet.Points) * bet.MultiplierSnapshot)
	}
	return model.WcResultIncorrect, 0
}

// evaluateOverUnderPrediction checks total goals vs the O/U line.
// HandicapSnapshot stores the OU line; PredictionChoice is "over" or "under".
// Exact hit on the line = void (stake refunded).
func evaluateOverUnderPrediction(bet *model.WcPrediction, homeScore, awayScore int) (string, float64) {
	if bet.PredictionChoice == nil || bet.HandicapSnapshot == nil {
		return model.WcResultIncorrect, 0
	}
	total := float64(homeScore + awayScore)
	line := *bet.HandicapSnapshot
	choice := *bet.PredictionChoice

	if isQuarterHandicap(line) {
		r1 := evalSubOverUnder(total, line-0.25, choice)
		r2 := evalSubOverUnder(total, line+0.25, choice)
		half := float64(bet.Points) / 2
		switch {
		case r1 == "win" && r2 == "win":
			return model.WcResultCorrect, round2(float64(bet.Points) * bet.MultiplierSnapshot)
		case r1 == "lose" && r2 == "lose":
			return model.WcResultIncorrect, 0
		case (r1 == "win" || r2 == "win"):
			// one win + one push
			return model.WcResultWinHalf, round2(half*bet.MultiplierSnapshot + half)
		default:
			// one push + one lose
			return model.WcResultLoseHalf, round2(half)
		}
	}

	if total == line {
		return model.WcResultVoid, float64(bet.Points)
	}
	if (total > line && choice == model.WcChoiceOver) || (total < line && choice == model.WcChoiceUnder) {
		return model.WcResultCorrect, round2(float64(bet.Points) * bet.MultiplierSnapshot)
	}
	return model.WcResultIncorrect, 0
}

// evalSubOverUnder evaluates a single O/U line, returning "win", "lose", or "push".
func evalSubOverUnder(total, line float64, choice string) string {
	if total == line {
		return "push"
	}
	if (total > line && choice == model.WcChoiceOver) || (total < line && choice == model.WcChoiceUnder) {
		return "win"
	}
	return "lose"
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

// evaluateOverUnderBet checks total goals vs the O/U line for the betting system.
// HandicapSnapshot stores the OU line; BetChoice is "over" or "under".
// Exact hit on the line = push (stake refunded).
func evaluateOverUnderBet(bet *model.WcBet, homeScore, awayScore int) (string, float64) {
	if bet.BetChoice == nil || bet.HandicapSnapshot == nil {
		return model.WcResultLose, 0
	}
	total := float64(homeScore + awayScore)
	line := *bet.HandicapSnapshot
	choice := *bet.BetChoice

	if isQuarterHandicap(line) {
		r1 := evalSubOverUnder(total, line-0.25, choice)
		r2 := evalSubOverUnder(total, line+0.25, choice)
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

	if total == line {
		return model.WcResultPush, float64(bet.Stake)
	}
	if (total > line && choice == model.WcChoiceOver) || (total < line && choice == model.WcChoiceUnder) {
		return model.WcResultWin, math.Round(float64(bet.Stake)*bet.OddsSnapshot*100) / 100
	}
	return model.WcResultLose, 0
}

// GetGroupStandings computes sorted group-stage standings from match data.
// Teams are sorted: points DESC → goal_difference DESC → goals_for DESC → team_name ASC.
func (s *WcService) GetGroupStandings() (*model.WcStandingsResponse, error) {
	groups, err := s.repo.GetGroupStandings()
	if err != nil {
		return nil, err
	}
	for i := range groups {
		sortTeamStandings(groups[i].Teams)
	}
	return &model.WcStandingsResponse{Groups: groups}, nil
}

func sortTeamStandings(teams []model.WcTeamStanding) {
	n := len(teams)
	for i := 1; i < n; i++ {
		for j := i; j > 0; j-- {
			a, b := teams[j-1], teams[j]
			if standingLess(b, a) {
				teams[j-1], teams[j] = teams[j], teams[j-1]
			} else {
				break
			}
		}
	}
}

// standingLess returns true if a ranks higher than b.
func standingLess(a, b model.WcTeamStanding) bool {
	if a.Points != b.Points {
		return a.Points > b.Points
	}
	if a.GoalDifference != b.GoalDifference {
		return a.GoalDifference > b.GoalDifference
	}
	if a.GoalsFor != b.GoalsFor {
		return a.GoalsFor > b.GoalsFor
	}
	return a.TeamName < b.TeamName
}
