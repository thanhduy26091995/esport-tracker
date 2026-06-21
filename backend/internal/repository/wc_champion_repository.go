package repository

import (
	"time"

	"github.com/duyb/esport-score-tracker/internal/model"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type WcChampionRepository struct {
	db *gorm.DB
}

func NewWcChampionRepository(db *gorm.DB) *WcChampionRepository {
	return &WcChampionRepository{db: db}
}

func (r *WcChampionRepository) DB() *gorm.DB { return r.db }

// --- Config ---

func (r *WcChampionRepository) GetConfig() (*model.WcChampionConfig, error) {
	var cfg model.WcChampionConfig
	err := r.db.First(&cfg, 1).Error
	return &cfg, err
}

func (r *WcChampionRepository) UpdateConfig(isOpen bool) error {
	return r.db.Model(&model.WcChampionConfig{}).
		Where("id = 1").
		Updates(map[string]any{"is_open": isOpen, "updated_at": time.Now()}).Error
}

func (r *WcChampionRepository) MarkSettled(winnerID uuid.UUID) error {
	now := time.Now()
	return r.db.Model(&model.WcChampionConfig{}).
		Where("id = 1").
		Updates(map[string]any{
			"winner_id":  winnerID,
			"settled_at": now,
			"updated_at": now,
		}).Error
}

// --- Teams ---

func (r *WcChampionRepository) ListTeams() ([]*model.WcChampionTeam, error) {
	var teams []*model.WcChampionTeam
	err := r.db.Order("odds ASC, name ASC").Find(&teams).Error
	return teams, err
}

func (r *WcChampionRepository) GetTeam(id uuid.UUID) (*model.WcChampionTeam, error) {
	var t model.WcChampionTeam
	err := r.db.First(&t, "id = ?", id).Error
	return &t, err
}

func (r *WcChampionRepository) CreateTeam(t *model.WcChampionTeam) error {
	return r.db.Create(t).Error
}

func (r *WcChampionRepository) UpdateTeamOdds(id uuid.UUID, odds float64) error {
	return r.db.Model(&model.WcChampionTeam{}).
		Where("id = ?", id).
		Updates(map[string]any{"odds": odds, "updated_at": time.Now()}).Error
}

// --- Predictions ---

func (r *WcChampionRepository) GetMyPrediction(wcUserID uuid.UUID) (*model.WcChampionPredictionMine, error) {
	var row model.WcChampionPredictionMine
	err := r.db.Table("wc_champion_predictions p").
		Select("p.id, p.team_id, t.name AS team_name, t.code AS team_code, t.flag_emoji, p.points, p.odds_snapshot, p.result, p.points_earned, p.created_at, p.updated_at").
		Joins("JOIN wc_champion_teams t ON t.id = p.team_id").
		Where("p.wc_user_id = ?", wcUserID).
		Scan(&row).Error
	if err != nil {
		return nil, err
	}
	if row.ID == uuid.Nil {
		return nil, gorm.ErrRecordNotFound
	}
	row.PayoutIfCorrect = int(float64(row.Points) * row.OddsSnapshot)
	return &row, nil
}

func (r *WcChampionRepository) GetMyPredictions(wcUserID uuid.UUID) ([]*model.WcChampionPredictionMine, error) {
	var rows []*model.WcChampionPredictionMine
	err := r.db.Table("wc_champion_predictions p").
		Select("p.id, p.team_id, t.name AS team_name, t.code AS team_code, t.flag_emoji, p.points, p.odds_snapshot, p.result, p.points_earned, p.created_at, p.updated_at").
		Joins("JOIN wc_champion_teams t ON t.id = p.team_id").
		Where("p.wc_user_id = ?", wcUserID).
		Order("p.created_at ASC").
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	for _, row := range rows {
		row.PayoutIfCorrect = int(float64(row.Points) * row.OddsSnapshot)
	}
	return rows, nil
}

func (r *WcChampionRepository) CreatePrediction(p *model.WcChampionPrediction) error {
	return r.db.Create(p).Error
}

func (r *WcChampionRepository) DeletePredictionByID(id, wcUserID uuid.UUID) error {
	res := r.db.Where("id = ? AND wc_user_id = ?", id, wcUserID).Delete(&model.WcChampionPrediction{})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func (r *WcChampionRepository) GetAllPredictions() ([]*model.WcChampionPredictionPublic, error) {
	rows := make([]*model.WcChampionPredictionPublic, 0)
	err := r.db.Table("wc_champion_predictions p").
		Select("u.name AS user_name, p.wc_user_id, t.name AS team_name, t.code AS team_code, t.flag_emoji, p.points, p.odds_snapshot, p.result").
		Joins("JOIN wc_champion_teams t ON t.id = p.team_id").
		Joins("JOIN wc_users u ON u.id = p.wc_user_id").
		Order("p.created_at ASC").
		Scan(&rows).Error
	for _, r := range rows {
		r.PayoutIfCorrect = int(float64(r.Points) * r.OddsSnapshot)
	}
	return rows, err
}

func (r *WcChampionRepository) UpsertPrediction(p *model.WcChampionPrediction) error {
	// Update if exists, insert if not
	var existing model.WcChampionPrediction
	err := r.db.Where("wc_user_id = ?", p.WcUserID).First(&existing).Error
	if err == gorm.ErrRecordNotFound {
		return r.db.Create(p).Error
	}
	if err != nil {
		return err
	}
	return r.db.Model(&existing).Updates(map[string]any{
		"team_id":       p.TeamID,
		"points":        p.Points,
		"odds_snapshot": p.OddsSnapshot,
		"result":        nil,
		"points_earned": nil,
		"updated_at":    time.Now(),
	}).Error
}

func (r *WcChampionRepository) DeletePrediction(wcUserID uuid.UUID) error {
	return r.db.Where("wc_user_id = ?", wcUserID).Delete(&model.WcChampionPrediction{}).Error
}

// ListPredictionsForSettle returns all predictions for settlement processing.
func (r *WcChampionRepository) ListPredictionsForSettle() ([]*model.WcChampionPrediction, error) {
	var preds []*model.WcChampionPrediction
	err := r.db.Find(&preds).Error
	return preds, err
}

// SettlePrediction marks result + points_earned for one prediction row.
func (r *WcChampionRepository) SettlePrediction(tx *gorm.DB, id uuid.UUID, result string, pointsEarned int) error {
	db := r.db
	if tx != nil {
		db = tx
	}
	return db.Model(&model.WcChampionPrediction{}).
		Where("id = ?", id).
		Updates(map[string]any{
			"result":        result,
			"points_earned": pointsEarned,
			"updated_at":    time.Now(),
		}).Error
}
