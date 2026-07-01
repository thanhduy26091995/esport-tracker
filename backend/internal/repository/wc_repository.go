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

func (r *WcRepository) UpdateBetLimits(min, max int, updatedBy *uuid.UUID) error {
	return r.db.Model(&model.WcConfig{}).
		Where("id = ?", 1).
		Updates(map[string]interface{}{
			"min_points": min,
			"max_points": max,
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
		Where("home_score IS NOT NULL AND away_score IS NOT NULL AND settled_at IS NULL AND status != ?", model.WcStatusLive).
		Order("match_date ASC").
		Find(&matches).Error
}

func (r *WcRepository) ListAllScoredMatches() ([]*model.WcMatch, error) {
	var matches []*model.WcMatch
	return matches, r.db.
		Where("home_score IS NOT NULL AND away_score IS NOT NULL AND status != ?", model.WcStatusLive).
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

// GetWalletTx reads the wallet within an active transaction so balance_before
// values in wallet logs are consistent with concurrent writes.
func (r *WcRepository) GetWalletTx(tx *gorm.DB, wcUserID uuid.UUID) (*model.WcWallet, error) {
	db := r.db
	if tx != nil {
		db = tx
	}
	var w model.WcWallet
	err := db.Where("wc_user_id = ?", wcUserID).First(&w).Error
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
		Order("CASE WHEN w.balance > 0 THEN 0 WHEN w.balance < 0 THEN 1 ELSE 2 END, w.balance DESC").
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
	if bet.OriginalPoints == nil {
		bet.OriginalPoints = &bet.Points
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

// SoftCancelPrediction sets cancelled_at and cancel_penalty without hard-deleting.
func (r *WcRepository) SoftCancelPrediction(tx *gorm.DB, id, wcUserID uuid.UUID, penalty float64) error {
	db := r.db
	if tx != nil {
		db = tx
	}
	now := time.Now()
	result := db.Model(&model.WcPrediction{}).
		Where("id = ? AND wc_user_id = ? AND cancelled_at IS NULL", id, wcUserID).
		Updates(map[string]interface{}{
			"cancelled_at":   now,
			"cancel_penalty": penalty,
		})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("prediction not found or already cancelled")
	}
	return nil
}

func (r *WcRepository) UpdatePredictionPoints(id uuid.UUID, points int) error {
	return r.db.Model(&model.WcPrediction{}).Where("id = ?", id).
		Updates(map[string]interface{}{
			"points":          points,
			"original_points": gorm.Expr("COALESCE(original_points, ?)", points),
		}).Error
}

func (r *WcRepository) UpdatePredictionPointsWithPenalty(tx *gorm.DB, id uuid.UUID, points int, penalty float64) error {
	db := r.db
	if tx != nil {
		db = tx
	}
	return db.Model(&model.WcPrediction{}).Where("id = ?", id).
		Updates(map[string]interface{}{
			"points":          points,
			"original_points": gorm.Expr("COALESCE(original_points, ?)", points),
			"reduce_penalty":  gorm.Expr("reduce_penalty + ?", penalty),
		}).Error
}

// BackfillOriginalPoints sets original_points = points for all predictions where original_points IS NULL.
// Returns the number of rows updated.
func (r *WcRepository) BackfillOriginalPoints() (int64, error) {
	result := r.db.Exec(`UPDATE wc_predictions SET original_points = points WHERE original_points IS NULL`)
	return result.RowsAffected, result.Error
}

func (r *WcRepository) ListPredictionsForMatch(matchID uuid.UUID) ([]*model.WcPrediction, error) {
	var bets []*model.WcPrediction
	err := r.db.Where("match_id = ? AND cancelled_at IS NULL", matchID).Find(&bets).Error
	return bets, err
}

func (r *WcRepository) ListPredictionsForMatchPublic(matchID uuid.UUID) ([]*model.WcPredictionPublic, error) {
	var bets []*model.WcPredictionPublic
	err := r.db.Table("wc_predictions b").
		Select("b.id, b.wc_user_id, u.name, u.avatar_url, b.prediction_type, b.prediction_choice, b.predicted_home_score, b.predicted_away_score, b.points, b.multiplier_snapshot, b.result, b.points_earned, b.created_at").
		Joins("JOIN wc_users u ON u.id = b.wc_user_id").
		Where("b.match_id = ? AND b.cancelled_at IS NULL", matchID).
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
			u.id                                AS wc_user_id,
			u.name,
			u.avatar_url,
			u.is_bot,
			COALESCE(w.balance, 0)              AS net_points,
			COALESCE(pred.total_predictions, 0) AS total_predictions,
			COALESCE(pred.correct, 0)           AS correct,
			COALESCE(pred.win_half, 0)          AS win_half,
			COALESCE(pred.lose_half, 0)         AS lose_half,
			COALESCE(pred.incorrect, 0)         AS incorrect
		FROM wc_users u
		LEFT JOIN wc_wallets w ON w.wc_user_id = u.id
		LEFT JOIN (
			SELECT
				b.wc_user_id,
				COUNT(b.id)                                            AS total_predictions,
				COUNT(b.id) FILTER (WHERE b.result = 'correct')       AS correct,
				COUNT(b.id) FILTER (WHERE b.result = 'win_half')      AS win_half,
				COUNT(b.id) FILTER (WHERE b.result = 'lose_half')     AS lose_half,
				COUNT(b.id) FILTER (WHERE b.result = 'incorrect')     AS incorrect
			FROM wc_predictions b
			WHERE b.result IS NOT NULL
			  AND b.cancelled_at IS NULL
			GROUP BY b.wc_user_id
		) pred ON pred.wc_user_id = u.id
		WHERE w.wc_user_id IS NOT NULL
		  AND (
		      EXISTS (SELECT 1 FROM wc_predictions       p  WHERE p.wc_user_id  = u.id)
		   OR EXISTS (SELECT 1 FROM wc_custom_bet_entries c  WHERE c.wc_user_id  = u.id)
		   OR EXISTS (SELECT 1 FROM wc_champion_predictions cp WHERE cp.wc_user_id = u.id)
		  )
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
	} else {
		updates["completed_at"] = nil
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
		Where("b.wc_user_id = ? AND b.cancelled_at IS NULL", wcUserID).
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
		Where("b.match_id = ? AND b.cancelled_at IS NULL", matchID).
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
	// Exclude void and cancelled bets.
	err := r.db.Where("match_id = ? AND (result IS NULL OR result != 'void') AND cancelled_at IS NULL", matchID).Find(&bets).Error
	return bets, err
}

// ListPendingBetsForUser returns all unsettled, non-cancelled bets for a user (used when blocking).
func (r *WcRepository) ListPendingBetsForUser(tx *gorm.DB, userID uuid.UUID) ([]*model.WcBet, error) {
	db := r.db
	if tx != nil {
		db = tx
	}
	var bets []*model.WcBet
	err := db.Where("wc_user_id = ? AND result IS NULL AND cancelled_at IS NULL", userID).Find(&bets).Error
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

func (r *WcRepository) UpdateBetStake(tx *gorm.DB, id uuid.UUID, stake int) error {
	db := r.db
	if tx != nil {
		db = tx
	}
	return db.Model(&model.WcBet{}).Where("id = ?", id).
		Updates(map[string]interface{}{
			"stake":          stake,
			"updated_at":     time.Now(),
			"original_stake": gorm.Expr("COALESCE(original_stake, stake)"),
		}).Error
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

// SoftCancelBet sets cancelled_at and cancel_penalty on a bet (instead of hard-deleting).
func (r *WcRepository) SoftCancelBet(tx *gorm.DB, id, wcUserID uuid.UUID, penalty float64) error {
	db := r.db
	if tx != nil {
		db = tx
	}
	now := time.Now()
	result := db.Model(&model.WcBet{}).
		Where("id = ? AND wc_user_id = ? AND cancelled_at IS NULL", id, wcUserID).
		Updates(map[string]interface{}{
			"cancelled_at":   now,
			"cancel_penalty": penalty,
			"updated_at":     now,
		})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("bet not found or not authorized")
	}
	return nil
}

// ListBetHistoryForUser returns settled + cancelled bets for a user, joined with match info.
func (r *WcRepository) ListBetHistoryForUser(wcUserID uuid.UUID) ([]*model.WcBetWithMatch, error) {
	var bets []*model.WcBetWithMatch
	err := r.db.Table("wc_bets b").
		Select(`b.*,
			m.home_team, m.away_team, m.match_date, m.status AS match_status, m.bets_locked_at,
			false AS betting_open`).
		Joins("JOIN wc_matches m ON m.id = b.match_id").
		Where("b.wc_user_id = ? AND (b.result IS NOT NULL OR b.cancelled_at IS NOT NULL)", wcUserID).
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

// UpdatePenaltyConfig updates cancel penalty and reduce stake penalty settings.
func (r *WcRepository) UpdatePenaltyConfig(cancelEnabled bool, cancelPercent, reduceMaxPercent, reducePenaltyPercent int, updatedBy *uuid.UUID) error {
	return r.db.Model(&model.WcConfig{}).
		Where("id = ?", 1).
		Updates(map[string]interface{}{
			"cancel_penalty_enabled":    cancelEnabled,
			"cancel_penalty_percent":    cancelPercent,
			"bet_reduce_max_percent":    reduceMaxPercent,
			"bet_reduce_penalty_percent": reducePenaltyPercent,
			"updated_by":                updatedBy,
			"updated_at":                time.Now(),
		}).Error
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

// GetGroupStandings computes group-stage standings from wc_matches.
// All group-stage matches (any status) are fetched to build the team roster.
// Stats (W/D/L, goals, form) are accumulated only from completed matches.
func (r *WcRepository) GetGroupStandings() ([]model.WcGroupStanding, error) {
	var matches []model.WcMatch
	if err := r.db.Where("stage = ?", model.WcStageGroup).
		Order("match_date ASC").
		Find(&matches).Error; err != nil {
		return nil, err
	}

	type teamAcc struct {
		code    string
		played  int
		won     int
		drawn   int
		lost    int
		gf      int
		ga      int
		results []string // all completed results in chronological order
	}

	// groupMap[groupName][teamName] → accumulator
	groupMap := map[string]map[string]*teamAcc{}

	for _, m := range matches {
		g := m.GroupName
		if g == "" {
			continue
		}
		if groupMap[g] == nil {
			groupMap[g] = map[string]*teamAcc{}
		}
		// Register team in roster regardless of match status
		if groupMap[g][m.HomeTeam] == nil {
			groupMap[g][m.HomeTeam] = &teamAcc{code: m.HomeTeamCode}
		}
		if groupMap[g][m.AwayTeam] == nil {
			groupMap[g][m.AwayTeam] = &teamAcc{code: m.AwayTeamCode}
		}

		// Only count stats for completed matches with valid scores
		if m.Status != model.WcStatusCompleted || m.HomeScore == nil || m.AwayScore == nil {
			continue
		}
		hs, as_ := *m.HomeScore, *m.AwayScore
		h := groupMap[g][m.HomeTeam]
		a := groupMap[g][m.AwayTeam]
		h.played++
		a.played++
		h.gf += hs
		h.ga += as_
		a.gf += as_
		a.ga += hs
		switch {
		case hs > as_:
			h.won++
			a.lost++
			h.results = append(h.results, "W")
			a.results = append(a.results, "L")
		case hs < as_:
			a.won++
			h.lost++
			h.results = append(h.results, "L")
			a.results = append(a.results, "W")
		default:
			h.drawn++
			a.drawn++
			h.results = append(h.results, "D")
			a.results = append(a.results, "D")
		}
	}

	// Convert map to sorted []WcGroupStanding
	// Sort groups alphabetically (A, B, C ... L)
	groupNames := make([]string, 0, len(groupMap))
	for g := range groupMap {
		groupNames = append(groupNames, g)
	}
	sortStrings(groupNames)

	standings := make([]model.WcGroupStanding, 0, len(groupNames))
	for _, gName := range groupNames {
		teams := groupMap[gName]
		teamNames := make([]string, 0, len(teams))
		for t := range teams {
			teamNames = append(teamNames, t)
		}
		sortStrings(teamNames)

		teamStandings := make([]model.WcTeamStanding, 0, len(teamNames))
		for _, tName := range teamNames {
			acc := teams[tName]
			// Form = last 5 results (slice is already chronological)
			form := acc.results
			if len(form) > 5 {
				form = form[len(form)-5:]
			}
			if form == nil {
				form = []string{}
			}
			pts := acc.won*3 + acc.drawn
			gd := acc.gf - acc.ga
			teamStandings = append(teamStandings, model.WcTeamStanding{
				TeamName:       tName,
				TeamCode:       acc.code,
				Played:         acc.played,
				Won:            acc.won,
				Drawn:          acc.drawn,
				Lost:           acc.lost,
				GoalsFor:       acc.gf,
				GoalsAgainst:   acc.ga,
				GoalDifference: gd,
				Points:         pts,
				Form:           form,
			})
		}
		standings = append(standings, model.WcGroupStanding{
			GroupName: gName,
			Teams:     teamStandings,
		})
	}
	return standings, nil
}

// sortStrings sorts a string slice in place (insertion sort — small N).
func sortStrings(s []string) {
	for i := 1; i < len(s); i++ {
		for j := i; j > 0 && s[j] < s[j-1]; j-- {
			s[j], s[j-1] = s[j-1], s[j]
		}
	}
}

// --- Tournament Analytics ---

// GetCompletedMatchStats returns aggregate match stats for the analytics page.
// Source of truth for total goals, H/A/D, clean sheets, goals by stage, and highest-scoring match.
func (r *WcRepository) GetCompletedMatchStats() (*model.WcTournamentMatchStats, error) {
	type aggRow struct {
		TotalMatches int
		TotalGoals   int
		HomeWins     int
		AwayWins     int
		Draws        int
		CleanSheets  int
	}
	var a aggRow
	err := r.db.Raw(`
		SELECT
			COUNT(*)                                                              AS total_matches,
			COALESCE(SUM(home_score + away_score), 0)                            AS total_goals,
			SUM(CASE WHEN home_score > away_score THEN 1 ELSE 0 END)             AS home_wins,
			SUM(CASE WHEN away_score > home_score THEN 1 ELSE 0 END)             AS away_wins,
			SUM(CASE WHEN home_score = away_score THEN 1 ELSE 0 END)             AS draws,
			SUM(CASE WHEN home_score = 0 OR away_score = 0 THEN 1 ELSE 0 END)   AS clean_sheets
		FROM wc_matches
		WHERE status = 'completed' AND home_score IS NOT NULL
	`).Scan(&a).Error
	if err != nil {
		return nil, err
	}

	type stageRow struct {
		Stage   string
		Matches int
		Goals   int
	}
	var stageRows []stageRow
	r.db.Raw(`
		SELECT stage, COUNT(*) AS matches, SUM(home_score + away_score) AS goals
		FROM wc_matches
		WHERE status = 'completed' AND home_score IS NOT NULL
		GROUP BY stage
		ORDER BY MIN(match_date)
	`).Scan(&stageRows)

	stageGoals := make([]model.WcStageGoalsStat, len(stageRows))
	for i, s := range stageRows {
		stageGoals[i] = model.WcStageGoalsStat{Stage: s.Stage, Matches: s.Matches, Goals: s.Goals}
	}

	var top model.WcMatch
	r.db.Where("status = 'completed' AND home_score IS NOT NULL").
		Order("(home_score + away_score) DESC").
		First(&top)

	var highest *model.WcTournamentMatchResult
	if top.ID != uuid.Nil {
		total := *top.HomeScore + *top.AwayScore
		highest = &model.WcTournamentMatchResult{
			HomeTeam:   top.HomeTeam,
			AwayTeam:   top.AwayTeam,
			HomeScore:  *top.HomeScore,
			AwayScore:  *top.AwayScore,
			Stage:      top.Stage,
			TotalGoals: total,
		}
	}

	avg := 0.0
	if a.TotalMatches > 0 {
		avg = math.Round(float64(a.TotalGoals)/float64(a.TotalMatches)*100) / 100
	}

	return &model.WcTournamentMatchStats{
		TotalMatches:        a.TotalMatches,
		TotalGoals:          a.TotalGoals,
		AvgGoalsPerMatch:    avg,
		HomeWins:            a.HomeWins,
		AwayWins:            a.AwayWins,
		Draws:               a.Draws,
		CleanSheets:         a.CleanSheets,
		HighestScoringMatch: highest,
		GoalsByStage:        stageGoals,
	}, nil
}
