package service

import (
	"errors"
	"strconv"

	"github.com/duyb/esport-score-tracker/internal/model"
	"github.com/duyb/esport-score-tracker/internal/repository"
)

type ConfigService struct {
	configRepo *repository.ConfigRepository
}

func NewConfigService(configRepo *repository.ConfigRepository) *ConfigService {
	return &ConfigService{configRepo: configRepo}
}

// GetAllConfig returns all configuration entries
func (s *ConfigService) GetAllConfig() ([]*model.Config, error) {
	return s.configRepo.GetAll()
}

// GetConfigByKey returns a specific config value
func (s *ConfigService) GetConfigByKey(key string) (*model.Config, error) {
	return s.configRepo.GetByKey(key)
}

// UpdateConfig updates a config value with validation
func (s *ConfigService) UpdateConfig(key, value string) error {
	// Validate based on key type
	switch key {
	case "debt_threshold":
		// Must be negative integer
		val, err := strconv.Atoi(value)
		if err != nil {
			return errors.New("debt_threshold must be an integer")
		}
		if val > 0 {
			return errors.New("debt_threshold must be negative or zero")
		}
	case "point_to_vnd":
		// Must be positive integer
		val, err := strconv.Atoi(value)
		if err != nil {
			return errors.New("point_to_vnd must be an integer")
		}
		if val <= 0 {
			return errors.New("point_to_vnd must be positive")
		}
	case "fund_split_percent":
		// Must be between 0 and 100
		val, err := strconv.Atoi(value)
		if err != nil {
			return errors.New("fund_split_percent must be an integer")
		}
		if val < 0 || val > 100 {
			return errors.New("fund_split_percent must be between 0 and 100")
		}
	case "points_per_win":
		val, err := strconv.Atoi(value)
		if err != nil {
			return errors.New("points_per_win must be an integer")
		}
		if val <= 0 {
			return errors.New("points_per_win must be greater than 0")
		}
	case "auto_settlement":
		if value != "true" && value != "false" {
			return errors.New("auto_settlement must be true or false")
		}
	default:
		return errors.New("invalid config key")
	}

	return s.configRepo.UpdateByKey(key, value)
}

// UpdateAllConfig validates and updates all provided key-value config pairs atomically
func (s *ConfigService) UpdateAllConfig(updates map[string]string) ([]*model.Config, error) {
	for key, value := range updates {
		if err := s.UpdateConfig(key, value); err != nil {
			return nil, err
		}
	}
	return s.configRepo.GetAll()
}

// GetDebtThreshold returns the debt threshold as an integer
func (s *ConfigService) GetDebtThreshold() (int, error) {
	config, err := s.configRepo.GetByKey("debt_threshold")
	if err != nil {
		return -6, err // default value
	}
	val, err := strconv.Atoi(config.Value)
	if err != nil {
		return -6, err
	}
	return val, nil
}

// GetPointToVND returns the point to VND conversion rate
func (s *ConfigService) GetPointToVND() (int, error) {
	config, err := s.configRepo.GetByKey("point_to_vnd")
	if err != nil {
		return 22000, err // default value
	}
	val, err := strconv.Atoi(config.Value)
	if err != nil {
		return 22000, err
	}
	return val, nil
}

// GetFundSplitPercent returns the fund split percentage
func (s *ConfigService) GetFundSplitPercent() (int, error) {
	config, err := s.configRepo.GetByKey("fund_split_percent")
	if err != nil {
		return 50, err // default value
	}
	val, err := strconv.Atoi(config.Value)
	if err != nil {
		return 50, err
	}
	return val, nil
}

// GetPointsPerWin returns points awarded per win (default: 1)
func (s *ConfigService) GetPointsPerWin() (int, error) {
	config, err := s.configRepo.GetByKey("points_per_win")
	if err != nil {
		return 1, err // default value
	}
	val, err := strconv.Atoi(config.Value)
	if err != nil {
		return 1, err
	}
	return val, nil
}

// GetAutoSettlement returns whether automatic settlement is enabled (default: false)
func (s *ConfigService) GetAutoSettlement() (bool, error) {
	config, err := s.configRepo.GetByKey("auto_settlement")
	if err != nil {
		return false, err // default: manual
	}
	return config.Value == "true", nil
}
