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
	return s.repo.UpdateMatch(id, map[string]interface{}{"betting_open": true})
}

func (s *WcService) CloseMatch(id uuid.UUID) error {
	return s.repo.UpdateMatch(id, map[string]interface{}{"betting_open": false})
}

// --- Score odds ---

func (s *WcService) AddScoreOdds(matchID uuid.UUID, homeScore, awayScore int, odds float64) (*model.WcScoreOdds, error) {
	m, err := s.repo.GetMatch(matchID)
	if err != nil {
		return nil, fmt.Errorf("match not found")
	}
	if isFinished(m) {
		return nil, fmt.Errorf("match is already completed or cancelled — cannot add score odds")
	}
	so := &model.WcScoreOdds{
		MatchID:   matchID,
		HomeScore: homeScore,
		AwayScore: awayScore,
		Odds:      odds,
	}
	return so, s.repo.CreateScoreOdds(so)
}

func (s *WcService) UpdateScoreOdds(id uuid.UUID, odds float64) error {
	return s.repo.UpdateScoreOdds(id, odds)
}

func (s *WcService) DeleteScoreOdds(id uuid.UUID) error {
	so, err := s.repo.GetScoreOddsByID(id)
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
	return s.repo.DeleteScoreOdds(id)
}

func (s *WcService) ListScoreOdds(matchID uuid.UUID) ([]*model.WcScoreOdds, error) {
	return s.repo.ListScoreOdds(matchID)
}

// --- Betting ---

type PlaceBetRequest struct {
	MatchID            uuid.UUID
	BetType            string
	BetChoice          *string // handicap: 'home'|'away'; nil for exact_score
	PredictedHomeScore *int    // exact_score only
	PredictedAwayScore *int    // exact_score only
	Stake              int
}

func (s *WcService) PlaceBet(wcUserID uuid.UUID, req PlaceBetRequest) (*model.WcBet, error) {
	m, err := s.repo.GetMatch(req.MatchID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("match not found")
		}
		return nil, err
	}

	if isLocked(m) {
		return nil, fmt.Errorf("bets are closed for this match")
	}

	if req.Stake <= 0 {
		return nil, fmt.Errorf("stake must be greater than 0")
	}

	var oddsSnapshot float64
	var handicapSnapshot *float64
	var handicapTeamSnapshot *string

	switch req.BetType {
	case model.WcBetTypeHandicap:
		if req.BetChoice == nil || (*req.BetChoice != model.WcTeamHome && *req.BetChoice != model.WcTeamAway) {
			return nil, fmt.Errorf("bet_choice must be 'home' or 'away' for handicap bets")
		}
		if m.HandicapValue == nil || m.OddsHandicapHome == nil || m.OddsHandicapAway == nil {
			return nil, fmt.Errorf("handicap odds not set for this match")
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
			return nil, fmt.Errorf("predicted_home_score and predicted_away_score are required")
		}
		so, err := s.repo.GetScoreOdds(req.MatchID, *req.PredictedHomeScore, *req.PredictedAwayScore)
		if err != nil {
			return nil, fmt.Errorf("scoreline %d:%d is not available for betting on this match",
				*req.PredictedHomeScore, *req.PredictedAwayScore)
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
		PredictedHomeScore:   req.PredictedHomeScore,
		PredictedAwayScore:   req.PredictedAwayScore,
		Stake:                req.Stake,
		OddsSnapshot:         oddsSnapshot,
		HandicapSnapshot:     handicapSnapshot,
		HandicapTeamSnapshot: handicapTeamSnapshot,
	}

	if err := s.repo.CreateBet(nil, bet); err != nil {
		return nil, fmt.Errorf("failed to place bet (may be duplicate): %w", err)
	}
	return bet, nil
}

func (s *WcService) ListBets(wcUserID uuid.UUID) ([]*model.WcBetWithMatch, error) {
	return s.repo.ListBets(wcUserID)
}

func (s *WcService) DeleteBet(wcUserID, betID uuid.UUID) error {
	bet, err := s.repo.GetBetByID(betID)
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
	if isLocked(m) {
		return fmt.Errorf("cannot modify bet: match is locked")
	}
	return s.repo.DeleteBet(betID)
}

func (s *WcService) UpdateBetStake(wcUserID, betID uuid.UUID, stake int) error {
	if stake <= 0 {
		return fmt.Errorf("stake must be greater than 0")
	}
	bet, err := s.repo.GetBetByID(betID)
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
	if isLocked(m) {
		return fmt.Errorf("cannot modify bet: match is locked")
	}
	return s.repo.UpdateBetStake(betID, stake)
}

func (s *WcService) ListBetsForMatchPublic(matchID uuid.UUID) ([]*model.WcBetPublic, error) {
	return s.repo.ListBetsForMatchPublic(matchID)
}

func (s *WcService) GetWallet(wcUserID uuid.UUID) (*model.WcWallet, error) {
	return s.repo.GetWallet(wcUserID)
}

func (s *WcService) GetLeaderboard() ([]*model.WcLeaderboardEntry, error) {
	return s.repo.GetLeaderboard()
}

// --- Match settlement ---

// SettleMatch evaluates all bets on a match, credits/debits wallets, and marks settled_at.
// Idempotent: re-settling reverses previous payouts before re-applying.
func (s *WcService) SettleMatch(matchID uuid.UUID) (int, int, error) {
	m, err := s.repo.GetMatch(matchID)
	if err != nil {
		return 0, 0, fmt.Errorf("match not found")
	}
	if m.HomeScore == nil || m.AwayScore == nil {
		return 0, 0, fmt.Errorf("match score not set — cannot settle")
	}

	bets, err := s.repo.ListBetsForMatch(matchID)
	if err != nil {
		return 0, 0, err
	}

	db := s.repo.DB()
	var totalPayout int
	processed := 0

	err = db.Transaction(func(tx *gorm.DB) error {
		for _, bet := range bets {
			// Reverse previous payout if re-settling
			if bet.Result != nil && bet.Payout != nil && *bet.Payout > 0 {
				prevDelta := -(*bet.Payout)
				if err := s.repo.UpdateWalletBalance(tx, bet.WcUserID, prevDelta); err != nil {
					return err
				}
			}

			var result string
			var payout int

			switch bet.BetType {
			case model.WcBetTypeHandicap:
				result, payout = evaluateHandicapBet(bet, *m.HomeScore, *m.AwayScore)
			case model.WcBetTypeExactScore:
				result, payout = evaluateExactScoreBet(bet, *m.HomeScore, *m.AwayScore)
			}

			if err := s.repo.UpdateBetResult(tx, bet.ID, result, payout); err != nil {
				return err
			}

			// Credit wallet with payout (0 for losses)
			if payout > 0 {
				if err := s.repo.UpdateWalletBalance(tx, bet.WcUserID, payout); err != nil {
					return err
				}
			}
			// Deduct stake from wallet (happens on bet placement for losses only conceptually;
			// here we track net: payout - stake)
			if err := s.repo.UpdateWalletBalance(tx, bet.WcUserID, -bet.Stake); err != nil {
				return err
			}

			totalPayout += payout
			processed++
		}

		// Mark match as settled
		now := time.Now()
		return s.repo.UpdateMatch(matchID, map[string]interface{}{
			"settled_at": now,
			"status":     model.WcStatusCompleted,
		})
	})

	return processed, totalPayout, err
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
			Amount:       math.Abs(float64(w.Balance)) * pointRate,
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
	balanceAfter := balanceBefore + delta

	db := s.repo.DB()
	return db.Transaction(func(tx *gorm.DB) error {
		if err := s.repo.UpdateWalletBalance(tx, wcUserID, delta); err != nil {
			return err
		}
		return s.repo.LogWalletChange(tx, &model.WcWalletLog{
			WcUserID:      wcUserID,
			AdminID:       adminID,
			Delta:         delta,
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

// --- Helpers ---

func isLocked(m *model.WcMatch) bool {
	if m.Status == model.WcStatusLive || m.Status == model.WcStatusCompleted || m.Status == model.WcStatusCancelled {
		return true
	}
	if !m.BettingOpen {
		return true
	}
	if m.BetsLockedAt != nil && time.Now().After(*m.BetsLockedAt) {
		return true
	}
	return false
}

// isFinished checks only terminal statuses — used for admin score-odds management.
// Unlike isLocked, it allows edits on live/locked matches.
func isFinished(m *model.WcMatch) bool {
	return m.Status == model.WcStatusCompleted || m.Status == model.WcStatusCancelled
}

// evaluateHandicapBet applies Asian handicap rules to determine result and payout.
func evaluateHandicapBet(bet *model.WcBet, homeScore, awayScore int) (string, int) {
	if bet.HandicapSnapshot == nil || bet.HandicapTeamSnapshot == nil || bet.BetChoice == nil {
		return model.WcResultLose, 0
	}

	h := *bet.HandicapSnapshot
	// Adjust home score: if handicap team is 'home', home gives goals (subtract); if 'away', home receives goals (add).
	var adjustedHome float64
	if *bet.HandicapTeamSnapshot == model.WcTeamHome {
		adjustedHome = float64(homeScore) - h
	} else {
		adjustedHome = float64(homeScore) + h
	}
	adjustedAway := float64(awayScore)

	// Determine handicap winner
	var handicapWinner string
	if adjustedHome > adjustedAway {
		handicapWinner = model.WcTeamHome
	} else if adjustedHome < adjustedAway {
		handicapWinner = model.WcTeamAway
	} else {
		// Push: adjusted scores equal (only possible with whole-number handicap)
		return model.WcResultPush, bet.Stake
	}

	if handicapWinner == *bet.BetChoice {
		payout := int(math.Floor(float64(bet.Stake) * bet.OddsSnapshot))
		return model.WcResultWin, payout
	}
	return model.WcResultLose, 0
}

// evaluateExactScoreBet checks if the predicted score matches the actual score.
func evaluateExactScoreBet(bet *model.WcBet, homeScore, awayScore int) (string, int) {
	if bet.PredictedHomeScore == nil || bet.PredictedAwayScore == nil {
		return model.WcResultLose, 0
	}
	if *bet.PredictedHomeScore == homeScore && *bet.PredictedAwayScore == awayScore {
		payout := int(math.Floor(float64(bet.Stake) * bet.OddsSnapshot))
		return model.WcResultWin, payout
	}
	return model.WcResultLose, 0
}
