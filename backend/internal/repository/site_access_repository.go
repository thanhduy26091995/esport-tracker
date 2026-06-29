package repository

import (
	"github.com/duyb/esport-score-tracker/internal/model"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type SiteAccessRepository struct {
	db *gorm.DB
}

func NewSiteAccessRepository(db *gorm.DB) *SiteAccessRepository {
	return &SiteAccessRepository{db: db}
}

func (r *SiteAccessRepository) Get() (*model.SiteAccessConfig, error) {
	var cfg model.SiteAccessConfig
	return &cfg, r.db.First(&cfg, 1).Error
}

func (r *SiteAccessRepository) Save(question, answerHash string, enabled bool) error {
	return r.db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "id"}},
		DoUpdates: clause.AssignmentColumns([]string{"question", "answer_hash", "enabled", "updated_at"}),
	}).Create(&model.SiteAccessConfig{
		ID:         1,
		Question:   question,
		AnswerHash: answerHash,
		Enabled:    enabled,
	}).Error
}
