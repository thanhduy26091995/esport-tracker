package repository

import (
	"github.com/duyb/esport-score-tracker/internal/model"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type UserRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) *UserRepository {
	return &UserRepository{db: db}
}

// GetAll returns all active users ordered by current_score DESC (leaderboard)
func (r *UserRepository) GetAll() ([]*model.User, error) {
	var users []*model.User
	err := r.db.Where("is_active = ?", true).
		Order("current_score DESC, name ASC").
		Find(&users).Error
	return users, err
}

// GetByID returns a user by ID
func (r *UserRepository) GetByID(id uuid.UUID) (*model.User, error) {
	var user model.User
	err := r.db.Where("id = ? AND is_active = ?", id, true).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// GetByName returns a user by name (for duplicate checking)
func (r *UserRepository) GetByName(name string) (*model.User, error) {
	var user model.User
	err := r.db.Where("name = ? AND is_active = ?", name, true).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// Create creates a new user
func (r *UserRepository) Create(user *model.User) error {
	return r.db.Create(user).Error
}

// Update updates an existing user
func (r *UserRepository) Update(user *model.User) error {
	return r.db.Save(user).Error
}

// SoftDelete marks a user as inactive (soft delete)
func (r *UserRepository) SoftDelete(id uuid.UUID) error {
	return r.db.Model(&model.User{}).
		Where("id = ?", id).
		Update("is_active", false).Error
}

// GetLeaderboard returns users with their ranking
func (r *UserRepository) GetLeaderboard(limit int) ([]*model.User, error) {
	var users []*model.User
	query := r.db.Where("is_active = ?", true).
		Order("current_score DESC, name ASC")
	
	if limit > 0 {
		query = query.Limit(limit)
	}
	
	err := query.Find(&users).Error
	return users, err
}

// UpdateScore updates a user's current score
func (r *UserRepository) UpdateScore(id uuid.UUID, scoreChange int) error {
	return r.db.Model(&model.User{}).
		Where("id = ?", id).
		UpdateColumn("current_score", gorm.Expr("current_score + ?", scoreChange)).Error
}
