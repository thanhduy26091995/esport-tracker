package repository

import (
	"github.com/duyb/esport-score-tracker/internal/model"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type ScoreBonusRepository struct {
	db *gorm.DB
}

func NewScoreBonusRepository(db *gorm.DB) *ScoreBonusRepository {
	return &ScoreBonusRepository{db: db}
}

func (r *ScoreBonusRepository) Create(bonus *model.ScoreBonus) error {
	return r.db.Create(bonus).Error
}

func (r *ScoreBonusRepository) GetAll(limit, offset int) ([]*model.ScoreBonus, error) {
	var bonuses []*model.ScoreBonus
	q := r.db.Preload("User").Order("bonus_date DESC")
	if limit > 0 {
		q = q.Limit(limit).Offset(offset)
	}
	return bonuses, q.Find(&bonuses).Error
}

func (r *ScoreBonusRepository) GetByID(id uuid.UUID) (*model.ScoreBonus, error) {
	var bonus model.ScoreBonus
	err := r.db.Preload("User").First(&bonus, "id = ?", id).Error
	return &bonus, err
}

func (r *ScoreBonusRepository) Delete(id uuid.UUID) error {
	return r.db.Delete(&model.ScoreBonus{}, "id = ?", id).Error
}
