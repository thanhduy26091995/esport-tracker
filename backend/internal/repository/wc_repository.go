package repository

import (
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
	Status string
	Stage  string
	Group  string
	Date   string // "YYYY-MM-DD" — filter by match date (local date)
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
	var matches []*model.WcMatch
	return matches, q.Find(&matches).Error
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

func (r *WcRepository) UpdateWalletBalance(tx *gorm.DB, wcUserID uuid.UUID, delta int) error {
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
		Select("b.id, b.wc_user_id, u.name, b.prediction_type, b.prediction_choice, b.predicted_home_score, b.predicted_away_score, b.points, b.multiplier_snapshot, b.result, b.points_earned, b.created_at").
		Joins("JOIN wc_users u ON u.id = b.wc_user_id").
		Where("b.match_id = ?", matchID).
		Order("b.created_at ASC").
		Scan(&bets).Error
	return bets, err
}

func (r *WcRepository) UpdatePredictionResult(tx *gorm.DB, betID uuid.UUID, result string, pointsEarned int) error {
	db := r.db
	if tx != nil {
		db = tx
	}
	return db.Model(&model.WcPrediction{}).
		Where("id = ?", betID).
		Updates(map[string]interface{}{"result": result, "points_earned": pointsEarned}).Error
}

func (r *WcRepository) GetLeaderboard() ([]*model.WcLeaderboardEntry, error) {
	var rows []*model.WcLeaderboardEntry
	err := r.db.Raw(`
		SELECT
			u.id   AS wc_user_id,
			u.name,
			COALESCE(SUM(b.points_earned - b.points) FILTER (WHERE b.result IS NOT NULL), 0) AS net_points,
			COUNT(b.id) FILTER (WHERE b.result IS NOT NULL)                                   AS total_predictions,
			COUNT(b.id) FILTER (WHERE b.result = 'correct')                                   AS correct
		FROM wc_users u
		LEFT JOIN wc_predictions b ON b.wc_user_id = u.id
		GROUP BY u.id, u.name
		ORDER BY net_points DESC, u.name ASC
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
			Amount:    math.Abs(float64(w.Balance)) * pointRate,
		})
	}
	return rows, nil
}
