package repository

import (
	"errors"
	"math"
	"time"

	"github.com/duyb/esport-score-tracker/internal/model"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type WcRepository struct {
	db *gorm.DB
}

func NewWcRepository(db *gorm.DB) *WcRepository {
	return &WcRepository{db: db}
}

func (r *WcRepository) DB() *gorm.DB {
	return r.db
}

// MatchFilter is used by ListMatches.
type MatchFilter struct {
	Status   string
	Stage    string
	Group    string
	Date     string // "YYYY-MM-DD" — filter by match date (local date)
	DateFrom string // ISO8601 UTC — match_date >= DateFrom
	DateTo   string // ISO8601 UTC — match_date <= DateTo
}

// --- Config ---

func (r *WcRepository) GetConfig() (*model.WcConfig, error) {
	var cfg model.WcConfig
	err := r.db.First(&cfg, 1).Error
	if err != nil {
		return nil, err
	}
	return &cfg, nil
}

func (r *WcRepository) UpdateConfig(isEnabled bool, updatedBy *uuid.UUID) error {
	return r.db.Model(&model.WcConfig{}).
		Where("id = ?", 1).
		Updates(map[string]interface{}{
			"is_enabled": isEnabled,
			"updated_by": updatedBy,
			"updated_at": time.Now(),
		}).Error
}

// --- Matches ---

func (r *WcRepository) UpsertMatches(matches []model.WcMatch) error {
	return r.db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "external_id"}},
		DoUpdates: clause.AssignmentColumns([]string{
			"home_team", "away_team", "home_team_code", "away_team_code",
			"match_date", "group_name", "stage", "venue",
			"home_score", "away_score", "status",
			"predictions_locked_at", "updated_at",
		}),
	}).Create(&matches).Error
}

func (r *WcRepository) ListMatches(f MatchFilter) ([]*model.WcMatch, error) {
	q := r.db.Order("match_date ASC")
	if f.Status != "" {
		q = q.Where("status = ?", f.Status)
	}
	if f.Stage != "" {
		q = q.Where("stage = ?", f.Stage)
	}
	if f.Group != "" {
		q = q.Where("group_name = ?", f.Group)
	}
	if f.Date != "" {
		q = q.Where("DATE(match_date AT TIME ZONE 'UTC') = ?", f.Date)
	}
	if f.DateFrom != "" {
		q = q.Where("match_date >= ?", f.DateFrom)
	}
	if f.DateTo != "" {
		q = q.Where("match_date <= ?", f.DateTo)
	}
	var matches []*model.WcMatch
	return matches, q.Find(&matches).Error
}

func (r *WcRepository) ListUnfinalizedScoredMatches() ([]*model.WcMatch, error) {
	var matches []*model.WcMatch
	return matches, r.db.
		Where("home_score IS NOT NULL AND away_score IS NOT NULL AND settled_at IS NULL").
		Order("match_date ASC").
		Find(&matches).Error
}

func (r *WcRepository) ListAllScoredMatches() ([]*model.WcMatch, error) {
	var matches []*model.WcMatch
	return matches, r.db.
		Where("home_score IS NOT NULL AND away_score IS NOT NULL").
		Order("match_date ASC").
		Find(&matches).Error
}

func (r *WcRepository) GetMatch(id uuid.UUID) (*model.WcMatch, error) {
	var m model.WcMatch
	err := r.db.First(&m, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &m, nil
}

func (r *WcRepository) GetMatchWithOdds(id uuid.UUID) (*model.WcMatchWithOdds, error) {
	var m model.WcMatchWithOdds
	m.ScoreMultipliers = []model.WcScoreMultiplier{}
	err := r.db.Model(&model.WcMatch{}).
		Where("wc_matches.id = ?", id).
		First(&m.WcMatch).Error
	if err != nil {
		return nil, err
	}
	err = r.db.Where("match_id = ?", id).
		Order("home_score ASC, away_score ASC").
		Find(&m.ScoreMultipliers).Error
	return &m, err
}

func (r *WcRepository) UpdateMatch(id uuid.UUID, fields map[string]interface{}) error {
	return r.db.Model(&model.WcMatch{}).
		Where("id = ?", id).
		Updates(fields).Error
}

func (r *WcRepository) LockMatch(id uuid.UUID) error {
	now := time.Now()
	return r.db.Model(&model.WcMatch{}).
		Where("id = ?", id).
		Update("predictions_locked_at", now).Error
}

// --- Score multipliers ---

func (r *WcRepository) CreateScoreMultiplier(so *model.WcScoreMultiplier) error {
	return r.db.Create(so).Error
}

func (r *WcRepository) UpdateScoreMultiplier(id uuid.UUID, multiplier float64) error {
	return r.db.Model(&model.WcScoreMultiplier{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{"multiplier": multiplier, "updated_at": time.Now()}).Error
}

func (r *WcRepository) DeleteScoreMultiplier(id uuid.UUID) error {
	return r.db.Delete(&model.WcScoreMultiplier{}, "id = ?", id).Error
}

func (r *WcRepository) GetScoreMultiplier(matchID uuid.UUID, homeScore, awayScore int) (*model.WcScoreMultiplier, error) {
	var so model.WcScoreMultiplier
	err := r.db.Where("match_id = ? AND home_score = ? AND away_score = ?", matchID, homeScore, awayScore).
		First(&so).Error
	if err != nil {
		return nil, err
	}
	return &so, nil
}

func (r *WcRepository) GetScoreMultiplierByID(id uuid.UUID) (*model.WcScoreMultiplier, error) {
	var so model.WcScoreMultiplier
	err := r.db.First(&so, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &so, nil
}

func (r *WcRepository) BulkUpsertScoreMultipliers(multipliers []model.WcScoreMultiplier) error {
	return r.db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "match_id"}, {Name: "home_score"}, {Name: "away_score"}},
		DoUpdates: clause.AssignmentColumns([]string{"multiplier", "updated_at"}),
	}).Create(&multipliers).Error
}

func (r *WcRepository) ListScoreMultipliers(matchID uuid.UUID) ([]*model.WcScoreMultiplier, error) {
	var odds []*model.WcScoreMultiplier
	err := r.db.Where("match_id = ?", matchID).
		Order("home_score ASC, away_score ASC").
		Find(&odds).Error
	return odds, err
}

// --- Wallet ---

func (r *WcRepository) CreateWallet(tx *gorm.DB, wcUserID uuid.UUID) error {
	return tx.Create(&model.WcWallet{WcUserID: wcUserID, Balance: 0}).Error
}

func (r *WcRepository) GetWallet(wcUserID uuid.UUID) (*model.WcWallet, error) {
	var w model.WcWallet
	err := r.db.Where("wc_user_id = ?", wcUserID).First(&w).Error
	if err != nil {
		return nil, err
	}
	return &w, nil
}

func (r *WcRepository) GetAllWallets() ([]*model.WcWalletWithUser, error) {
	var rows []*model.WcWalletWithUser
	err := r.db.Table("wc_wallets w").
		Select("w.*, u.name").
		Joins("JOIN wc_users u ON u.id = w.wc_user_id").
		Order("w.balance DESC").
		Scan(&rows).Error
	return rows, err
}

func (r *WcRepository) UpdateWalletBalance(tx *gorm.DB, wcUserID uuid.UUID, delta float64) error {
	db := r.db
	if tx != nil {
		db = tx
	}
	return db.Model(&model.WcWallet{}).
		Where("wc_user_id = ?", wcUserID).
		UpdateColumn("balance", gorm.Expr("balance + ?", delta)).Error
}

func (r *WcRepository) LogWalletChange(tx *gorm.DB, log *model.WcWalletLog) error {
	db := r.db
	if tx != nil {
		db = tx
	}
	return db.Create(log).Error
}

func (r *WcRepository) GetWalletLogs(wcUserID uuid.UUID) ([]*model.WcWalletLog, error) {
	var logs []*model.WcWalletLog
	err := r.db.Where("wc_user_id = ?", wcUserID).
		Order("created_at DESC").
		Find(&logs).Error
	return logs, err
}

func (r *WcRepository) ResetAllWallets(tx *gorm.DB) error {
	return tx.Model(&model.WcWallet{}).
		Where("1 = 1").
		Update("balance", 0).Error
}

// --- Predictions ---

func (r *WcRepository) CreatePrediction(tx *gorm.DB, bet *model.WcPrediction) error {
	db := r.db
	if tx != nil {
		db = tx
	}
	return db.Create(bet).Error
}

func (r *WcRepository) ListPredictions(wcUserID uuid.UUID) ([]*model.WcPredictionWithMatch, error) {
	var bets []*model.WcPredictionWithMatch
	err := r.db.Table("wc_predictions b").
		Select("b.*, m.home_team, m.away_team, m.match_date, m.status AS match_status, m.predictions_open, m.predictions_locked_at").
		Joins("JOIN wc_matches m ON m.id = b.match_id").
		Where("b.wc_user_id = ?", wcUserID).
		Order("b.created_at DESC").
		Scan(&bets).Error
	return bets, err
}

func (r *WcRepository) GetPredictionByID(id uuid.UUID) (*model.WcPrediction, error) {
	var bet model.WcPrediction
	err := r.db.Where("id = ?", id).First(&bet).Error
	return &bet, err
}

func (r *WcRepository) DeletePrediction(id uuid.UUID) error {
	return r.db.Delete(&model.WcPrediction{}, "id = ?", id).Error
}

func (r *WcRepository) UpdatePredictionPoints(id uuid.UUID, points int) error {
	return r.db.Model(&model.WcPrediction{}).Where("id = ?", id).Update("points", points).Error
}

func (r *WcRepository) ListPredictionsForMatch(matchID uuid.UUID) ([]*model.WcPrediction, error) {
	var bets []*model.WcPrediction
	err := r.db.Where("match_id = ?", matchID).Find(&bets).Error
	return bets, err
}

func (r *WcRepository) ListPredictionsForMatchPublic(matchID uuid.UUID) ([]*model.WcPredictionPublic, error) {
	var bets []*model.WcPredictionPublic
	err := r.db.Table("wc_predictions b").
		Select("b.id, b.wc_user_id, u.name, u.avatar_url, b.prediction_type, b.prediction_choice, b.predicted_home_score, b.predicted_away_score, b.points, b.multiplier_snapshot, b.result, b.points_earned, b.created_at").
		Joins("JOIN wc_users u ON u.id = b.wc_user_id").
		Where("b.match_id = ?", matchID).
		Order("b.created_at ASC").
		Scan(&bets).Error
	return bets, err
}

func (r *WcRepository) UpdatePredictionResult(tx *gorm.DB, betID uuid.UUID, result string, pointsEarned float64) error {
	db := r.db
	if tx != nil {
		db = tx
	}
	return db.Model(&model.WcPrediction{}).
		Where("id = ?", betID).
		Updates(map[string]interface{}{"result": result, "points_earned": pointsEarned}).Error
}

func (r *WcRepository) GetLeaderboard() ([]*model.WcLeaderboardEntry, error) {
	rows := make([]*model.WcLeaderboardEntry, 0)
	err := r.db.Raw(`
		SELECT
			u.id         AS wc_user_id,
			u.name,
			u.avatar_url,
			COALESCE(SUM(COALESCE(b.points_earned, 0) - b.points) FILTER (WHERE b.result IS NOT NULL), 0) AS net_points,
			COUNT(b.id) FILTER (WHERE b.result IS NOT NULL)              AS total_predictions,
			COUNT(b.id) FILTER (WHERE b.result = 'correct')              AS correct,
			COUNT(b.id) FILTER (WHERE b.result = 'win_half')             AS win_half,
			COUNT(b.id) FILTER (WHERE b.result = 'lose_half')            AS lose_half,
			COUNT(b.id) FILTER (WHERE b.result = 'incorrect')            AS incorrect
		FROM wc_users u
		LEFT JOIN wc_predictions b ON b.wc_user_id = u.id
		GROUP BY u.id, u.name, u.avatar_url
		HAVING COUNT(b.id) > 0
		ORDER BY net_points DESC, correct DESC, u.name ASC
	`).Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	for i := range rows {
		rows[i].Rank = i + 1
	}
	return rows, nil
}

// --- Settlement ---

func (r *WcRepository) CreateSettlement(tx *gorm.DB, s *model.WcSettlement, details []*model.WcSettlementDetail) error {
	if err := tx.Create(s).Error; err != nil {
		return err
	}
	for i := range details {
		details[i].SettlementID = s.ID
	}
	return tx.Create(&details).Error
}

func (r *WcRepository) ListSettlements() ([]*model.WcSettlement, error) {
	var list []*model.WcSettlement
	err := r.db.Order("created_at DESC").Find(&list).Error
	return list, err
}

func (r *WcRepository) GetSettlement(id uuid.UUID) (*model.WcSettlementWithDetails, error) {
	var s model.WcSettlement
	if err := r.db.First(&s, "id = ?", id).Error; err != nil {
		return nil, err
	}
	var details []*model.WcSettlementDetailWithUser
	err := r.db.Table("wc_settlement_details d").
		Select("d.*, u.name").
		Joins("JOIN wc_users u ON u.id = d.wc_user_id").
		Where("d.settlement_id = ?", id).
		Order("d.final_balance DESC").
		Scan(&details).Error
	if err != nil {
		return nil, err
	}
	return &model.WcSettlementWithDetails{WcSettlement: s, Details: details}, nil
}

func (r *WcRepository) UpdateSettlementDetailStatus(settlementID, wcUserID uuid.UUID, status, doneNote string) error {
	updates := map[string]interface{}{
		"status":    status,
		"done_note": doneNote,
	}
	if status == model.WcSettlementStatusDone {
		now := time.Now()
		updates["completed_at"] = &now
	}
	return r.db.Model(&model.WcSettlementDetail{}).
		Where("settlement_id = ? AND wc_user_id = ?", settlementID, wcUserID).
		Updates(updates).Error
}

// --- Score odds (for betting system) ---

func (r *WcRepository) GetScoreOdds(matchID uuid.UUID, homeScore, awayScore int) (*model.WcScoreOdds, error) {
	var so model.WcScoreOdds
	err := r.db.Where("match_id = ? AND home_score = ? AND away_score = ?", matchID, homeScore, awayScore).
		First(&so).Error
	if err != nil {
		return nil, err
	}
	return &so, nil
}

func (r *WcRepository) CreateScoreOdds(so *model.WcScoreOdds) error {
	return r.db.Create(so).Error
}

func (r *WcRepository) UpdateScoreOdds(id uuid.UUID, odds float64) error {
	return r.db.Model(&model.WcScoreOdds{}).Where("id = ?", id).
		Updates(map[string]interface{}{"odds": odds, "updated_at": time.Now()}).Error
}

func (r *WcRepository) DeleteScoreOdds(id uuid.UUID) error {
	return r.db.Delete(&model.WcScoreOdds{}, "id = ?", id).Error
}

func (r *WcRepository) BulkUpsertScoreOdds(odds []model.WcScoreOdds) error {
	return r.db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "match_id"}, {Name: "home_score"}, {Name: "away_score"}},
		DoUpdates: clause.AssignmentColumns([]string{"odds", "updated_at"}),
	}).Create(&odds).Error
}

func (r *WcRepository) ListScoreOdds(matchID uuid.UUID) ([]*model.WcScoreOdds, error) {
	odds := []*model.WcScoreOdds{}
	err := r.db.Where("match_id = ?", matchID).
		Order("home_score ASC, away_score ASC").
		Find(&odds).Error
	return odds, err
}

// --- Bets ---

func (r *WcRepository) CreateBet(bet *model.WcBet) error {
	return r.db.Create(bet).Error
}

func (r *WcRepository) GetBet(id uuid.UUID) (*model.WcBet, error) {
	var bet model.WcBet
	err := r.db.First(&bet, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &bet, nil
}

func (r *WcRepository) ListBets(wcUserID uuid.UUID) ([]*model.WcBetWithMatch, error) {
	var bets []*model.WcBetWithMatch
	err := r.db.Table("wc_bets b").
		Select(`b.*,
			m.home_team, m.away_team, m.match_date, m.status AS match_status, m.bets_locked_at,
			(m.bets_locked_at IS NULL OR m.bets_locked_at > NOW()) AND m.status NOT IN ('completed','cancelled') AS betting_open`).
		Joins("JOIN wc_matches m ON m.id = b.match_id").
		Where("b.wc_user_id = ?", wcUserID).
		Order("b.created_at DESC").
		Scan(&bets).Error
	if err != nil {
		return nil, err
	}
	if bets == nil {
		bets = []*model.WcBetWithMatch{}
	}
	return bets, nil
}

func (r *WcRepository) ListBetsForMatch(matchID uuid.UUID) ([]*model.WcBetPublic, error) {
	var bets []*model.WcBetPublic
	err := r.db.Table("wc_bets b").
		Select("b.id, b.wc_user_id, u.name, u.avatar_url, b.bet_type, b.bet_choice, b.stake, b.odds_snapshot, b.predicted_home_score, b.predicted_away_score, b.result, b.payout, b.created_at").
		Joins("JOIN wc_users u ON u.id = b.wc_user_id").
		Where("b.match_id = ?", matchID).
		Order("b.created_at ASC").
		Scan(&bets).Error
	if err != nil {
		return nil, err
	}
	if bets == nil {
		bets = []*model.WcBetPublic{}
	}
	return bets, nil
}

func (r *WcRepository) ListBetsForSettlement(matchID uuid.UUID) ([]*model.WcBet, error) {
	var bets []*model.WcBet
	err := r.db.Where("match_id = ?", matchID).Find(&bets).Error
	return bets, err
}

// ListPendingBetsForUser returns all unsettled bets for a user (used when blocking).
func (r *WcRepository) ListPendingBetsForUser(tx *gorm.DB, userID uuid.UUID) ([]*model.WcBet, error) {
	db := r.db
	if tx != nil {
		db = tx
	}
	var bets []*model.WcBet
	err := db.Where("wc_user_id = ? AND result IS NULL", userID).Find(&bets).Error
	return bets, err
}

// VoidBet sets result='void' and payout=stake for a bet (used when blocking a user).
func (r *WcRepository) VoidBet(tx *gorm.DB, betID uuid.UUID, stake int) error {
	db := r.db
	if tx != nil {
		db = tx
	}
	return db.Model(&model.WcBet{}).Where("id = ?", betID).
		Updates(map[string]interface{}{
			"result": "void",
			"payout": float64(stake),
		}).Error
}

func (r *WcRepository) UpdateBetStake(id uuid.UUID, stake int) error {
	return r.db.Model(&model.WcBet{}).Where("id = ?", id).
		Updates(map[string]interface{}{"stake": stake, "updated_at": time.Now()}).Error
}

func (r *WcRepository) UpdateBetResult(tx *gorm.DB, id uuid.UUID, result string, payout float64) error {
	db := r.db
	if tx != nil {
		db = tx
	}
	return db.Model(&model.WcBet{}).Where("id = ?", id).
		Updates(map[string]interface{}{"result": result, "payout": payout}).Error
}

func (r *WcRepository) DeleteBet(id, wcUserID uuid.UUID) error {
	result := r.db.Where("id = ? AND wc_user_id = ?", id, wcUserID).Delete(&model.WcBet{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("bet not found or not authorized")
	}
	return nil
}

// ListAllMatches returns all wc_matches (used by setup-mapping).
func (r *WcRepository) ListAllMatches() ([]*model.WcMatch, error) {
	var matches []*model.WcMatch
	err := r.db.Order("match_date ASC").Find(&matches).Error
	return matches, err
}

// ListUpcomingMatchesWithStatsapiID returns scheduled matches within the next 48h that have a statsapi_fixture_id.
func (r *WcRepository) ListUpcomingMatchesWithStatsapiID() ([]*model.WcMatch, error) {
	var matches []*model.WcMatch
	cutoff := time.Now().Add(48 * time.Hour)
	err := r.db.
		Where("status = ? AND match_date <= ? AND statsapi_fixture_id IS NOT NULL", "scheduled", cutoff).
		Order("match_date ASC").
		Find(&matches).Error
	return matches, err
}

// CreateSyncLog writes a sync log entry.
func (r *WcRepository) CreateSyncLog(log *model.WcSyncLog) error {
	return r.db.Create(log).Error
}

// GetSyncLogs returns the last 20 sync log entries (most recent first).
func (r *WcRepository) GetSyncLogs() ([]*model.WcSyncLog, error) {
	var logs []*model.WcSyncLog
	err := r.db.Order("created_at DESC").Limit(20).Find(&logs).Error
	return logs, err
}

// GetHousePnL aggregates bet data to compute house profit/loss.
func (r *WcRepository) GetHousePnL() (*model.HousePnLResponse, error) {
	type aggregate struct {
		TotalStakeSettled  float64
		TotalPayoutSettled float64
		SettledBetCount    int
		TotalStakeVoid     float64
		TotalStakePending  float64
		PendingBetCount    int
	}
	var agg aggregate
	err := r.db.Raw(`
		SELECT
			COALESCE(SUM(stake) FILTER (WHERE result IS NOT NULL AND result != 'void'), 0) AS total_stake_settled,
			COALESCE(SUM(payout) FILTER (WHERE result IS NOT NULL AND result != 'void'), 0) AS total_payout_settled,
			COALESCE(COUNT(*) FILTER (WHERE result IS NOT NULL AND result != 'void'), 0) AS settled_bet_count,
			COALESCE(SUM(stake) FILTER (WHERE result = 'void'), 0) AS total_stake_void,
			COALESCE(SUM(stake) FILTER (WHERE result IS NULL), 0) AS total_stake_pending,
			COALESCE(COUNT(*) FILTER (WHERE result IS NULL), 0) AS pending_bet_count
		FROM wc_bets
	`).Scan(&agg).Error
	if err != nil {
		return nil, err
	}

	type matchRow struct {
		MatchID   string
		HomeTeam  string
		AwayTeam  string
		MatchDate string
		Stage     string
		Stake     float64
		Payout    float64
		BetCount  int
	}
	var rows []matchRow
	err = r.db.Raw(`
		SELECT
			b.match_id::text AS match_id,
			m.home_team,
			m.away_team,
			m.match_date::text AS match_date,
			m.stage,
			SUM(b.stake)  AS stake,
			SUM(b.payout) AS payout,
			COUNT(b.id)   AS bet_count
		FROM wc_bets b
		JOIN wc_matches m ON m.id = b.match_id
		WHERE b.result IS NOT NULL AND b.result != 'void'
		GROUP BY b.match_id, m.home_team, m.away_team, m.match_date, m.stage
		ORDER BY (SUM(b.stake) - SUM(b.payout)) ASC
	`).Scan(&rows).Error
	if err != nil {
		return nil, err
	}

	breakdown := make([]model.HousePnLMatch, 0, len(rows))
	for _, row := range rows {
		breakdown = append(breakdown, model.HousePnLMatch{
			MatchID:  row.MatchID,
			HomeTeam: row.HomeTeam,
			AwayTeam: row.AwayTeam,
			MatchDate: row.MatchDate,
			Stage:    row.Stage,
			Stake:    row.Stake,
			Payout:   row.Payout,
			Profit:   row.Stake - row.Payout,
			BetCount: row.BetCount,
		})
	}

	return &model.HousePnLResponse{
		TotalStakeSettled:  agg.TotalStakeSettled,
		TotalPayoutSettled: agg.TotalPayoutSettled,
		HouseProfit:        agg.TotalStakeSettled - agg.TotalPayoutSettled,
		TotalStakeVoid:     agg.TotalStakeVoid,
		TotalStakePending:  agg.TotalStakePending,
		PendingBetCount:    agg.PendingBetCount,
		SettledBetCount:    agg.SettledBetCount,
		MatchBreakdown:     breakdown,
		GeneratedAt:        time.Now().Format(time.RFC3339),
	}, nil
}

// CountLiveMatches returns the number of matches currently in 'live' status.
func (r *WcRepository) CountLiveMatches() (int, error) {
	var count int64
	err := r.db.Model(&model.WcMatch{}).Where("status = ?", model.WcStatusLive).Count(&count).Error
	return int(count), err
}

// PreviewSettlement reads all wallet balances and computes direction + amount for each user.
// Nothing is written to DB.
func (r *WcRepository) PreviewSettlement(pointRate float64) ([]*model.WcSettlementPreviewRow, error) {
	wallets, err := r.GetAllWallets()
	if err != nil {
		return nil, err
	}
	rows := make([]*model.WcSettlementPreviewRow, 0, len(wallets))
	for _, w := range wallets {
		dir := model.WcDirectionEven
		if w.Balance > 0 {
			dir = model.WcDirectionPay
		} else if w.Balance < 0 {
			dir = model.WcDirectionCollect
		}
		rows = append(rows, &model.WcSettlementPreviewRow{
			WcUserID:  w.WcUserID,
			Name:      w.Name,
			Balance:   w.Balance,
			Direction: dir,
			Amount:    math.Abs(w.Balance) * pointRate,
		})
	}
	return rows, nil
}
