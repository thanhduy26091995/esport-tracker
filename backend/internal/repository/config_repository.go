package repository

import (
	"github.com/duyb/esport-score-tracker/internal/model"
	"gorm.io/gorm"
)

type ConfigRepository struct {
	db *gorm.DB
}

func NewConfigRepository(db *gorm.DB) *ConfigRepository {
	return &ConfigRepository{db: db}
}

// GetByKey returns a config value by key
func (r *ConfigRepository) GetByKey(key string) (*model.Config, error) {
	var config model.Config
	err := r.db.Where("key = ?", key).First(&config).Error
	if err != nil {
		return nil, err
	}
	return &config, nil
}

// GetAll returns all config entries
func (r *ConfigRepository) GetAll() ([]*model.Config, error) {
	var configs []*model.Config
	err := r.db.Find(&configs).Error
	return configs, err
}

// Update updates a config value
func (r *ConfigRepository) Update(config *model.Config) error {
	return r.db.Save(config).Error
}

// UpdateByKey updates a config value by key
func (r *ConfigRepository) UpdateByKey(key, value string) error {
	return r.db.Model(&model.Config{}).
		Where("key = ?", key).
		Update("value", value).Error
}
