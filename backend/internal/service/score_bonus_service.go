package service

import (
	"errors"
	"fmt"
	"time"

	"github.com/duyb/esport-score-tracker/internal/cache"
	"github.com/duyb/esport-score-tracker/internal/model"
	"github.com/duyb/esport-score-tracker/internal/repository"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type ScoreBonusService struct {
	repo        *repository.ScoreBonusRepository
	userRepo    *repository.UserRepository
	tierService *TierService
	db          *gorm.DB
	cache       cache.CacheStore
}

func NewScoreBonusService(
	repo *repository.ScoreBonusRepository,
	userRepo *repository.UserRepository,
	tierService *TierService,
	db *gorm.DB,
	c cache.CacheStore,
) *ScoreBonusService {
	return &ScoreBonusService{repo: repo, userRepo: userRepo, tierService: tierService, db: db, cache: c}
}

func (s *ScoreBonusService) invalidateScoreCaches() {
	_ = s.cache.Delete("esport:users:leaderboard")
	_ = s.cache.Delete("esport:users:all")
}

type CreateScoreBonusRequest struct {
	UserID      uuid.UUID  `json:"user_id" binding:"required"`
	Points      int        `json:"points" binding:"required"`
	Description string     `json:"description"`
	BonusDate   *time.Time `json:"bonus_date,omitempty"`
}

func (s *ScoreBonusService) CreateBonus(req *CreateScoreBonusRequest) (*model.ScoreBonus, error) {
	if req.Points <= 0 {
		return nil, errors.New("points must be positive")
	}

	if _, err := s.userRepo.GetByID(req.UserID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("user with ID %s not found", req.UserID)
		}
		return nil, err
	}

	bonusDate := time.Now()
	if req.BonusDate != nil {
		bonusDate = *req.BonusDate
	}

	tx := s.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	bonus := &model.ScoreBonus{
		UserID:      req.UserID,
		Points:      req.Points,
		Description: req.Description,
		BonusDate:   bonusDate,
	}
	if err := tx.Create(bonus).Error; err != nil {
		tx.Rollback()
		return nil, err
	}

	if err := tx.Model(&model.User{}).
		Where("id = ?", req.UserID).
		Update("current_score", gorm.Expr("current_score + ?", req.Points)).
		Error; err != nil {
		tx.Rollback()
		return nil, err
	}

	if err := tx.Commit().Error; err != nil {
		return nil, err
	}

	_ = s.tierService.RecalculateForUsers([]uuid.UUID{req.UserID})
	s.invalidateScoreCaches()
	return s.repo.GetByID(bonus.ID)
}

func (s *ScoreBonusService) DeleteBonus(id uuid.UUID) error {
	bonus, err := s.repo.GetByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("score bonus not found")
		}
		return err
	}

	tx := s.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	if err := tx.Model(&model.User{}).
		Where("id = ?", bonus.UserID).
		Update("current_score", gorm.Expr("current_score - ?", bonus.Points)).
		Error; err != nil {
		tx.Rollback()
		return err
	}

	if err := tx.Delete(&model.ScoreBonus{}, "id = ?", id).Error; err != nil {
		tx.Rollback()
		return err
	}

	if err := tx.Commit().Error; err != nil {
		return err
	}

	_ = s.tierService.RecalculateForUsers([]uuid.UUID{bonus.UserID})
	s.invalidateScoreCaches()
	return nil
}

func (s *ScoreBonusService) GetAll(limit, offset int) ([]*model.ScoreBonus, error) {
	return s.repo.GetAll(limit, offset)
}
