package service

import (
	"errors"
	"fmt"
	"strings"

	"github.com/duyb/esport-score-tracker/internal/model"
	"github.com/duyb/esport-score-tracker/internal/repository"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type UserService struct {
	repo *repository.UserRepository
}

func NewUserService(repo *repository.UserRepository) *UserService {
	return &UserService{repo: repo}
}

// GetAll returns all active users (leaderboard)
func (s *UserService) GetAll() ([]*model.User, error) {
	return s.repo.GetAll()
}

// GetByID returns a user by ID
func (s *UserService) GetByID(id uuid.UUID) (*model.User, error) {
	user, err := s.repo.GetByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("user not found")
		}
		return nil, err
	}
	return user, nil
}

// CreateUser creates a new user with validation
func (s *UserService) CreateUser(name string) (*model.User, error) {
	// Validate name
	name = strings.TrimSpace(name)
	if name == "" {
		return nil, fmt.Errorf("name cannot be empty")
	}
	if len(name) < 2 {
		return nil, fmt.Errorf("name must be at least 2 characters")
	}
	if len(name) > 100 {
		return nil, fmt.Errorf("name cannot exceed 100 characters")
	}

	// Check for duplicate name
	existing, err := s.repo.GetByName(name)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}
	if existing != nil {
		return nil, fmt.Errorf("user with name '%s' already exists", name)
	}

	// Create user
	user := &model.User{
		Name:         name,
		CurrentScore: 0,
		IsActive:     true,
	}

	if err := s.repo.Create(user); err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	return user, nil
}

// UpdateUser updates a user's name, tier, and handicap_rate with validation
func (s *UserService) UpdateUser(id uuid.UUID, name string, tier string, handicapRate float64) (*model.User, error) {
	// Get existing user
	user, err := s.GetByID(id)
	if err != nil {
		return nil, err
	}

	// Validate and update name if provided
	name = strings.TrimSpace(name)
	if name != "" {
		if len(name) < 2 {
			return nil, fmt.Errorf("name must be at least 2 characters")
		}
		if len(name) > 100 {
			return nil, fmt.Errorf("name cannot exceed 100 characters")
		}

		// Check for duplicate name (excluding current user)
		existing, err := s.repo.GetByName(name)
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, err
		}
		if existing != nil && existing.ID != id {
			return nil, fmt.Errorf("user with name '%s' already exists", name)
		}

		user.Name = name
	}

	// Validate and update tier if provided
	if tier != "" {
		if tier != "pro" && tier != "normal" && tier != "noop" {
			return nil, fmt.Errorf("tier must be one of: pro, normal, noop")
		}
		user.Tier = tier
	}

	user.HandicapRate = handicapRate

	if err := s.repo.Update(user); err != nil {
		return nil, fmt.Errorf("failed to update user: %w", err)
	}

	return user, nil
}

// DeleteUser soft deletes a user
func (s *UserService) DeleteUser(id uuid.UUID) error {
	// Check if user exists
	_, err := s.GetByID(id)
	if err != nil {
		return err
	}

	// Soft delete
	if err := s.repo.SoftDelete(id); err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}

	return nil
}

// GetLeaderboard returns the leaderboard with optional limit
func (s *UserService) GetLeaderboard(limit int) ([]*model.User, error) {
	return s.repo.GetLeaderboard(limit)
}
