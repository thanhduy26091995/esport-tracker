package repository

import (
	"time"

	"github.com/duyb/esport-score-tracker/internal/model"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type MatchRepository struct {
	db *gorm.DB
}

func NewMatchRepository(db *gorm.DB) *MatchRepository {
	return &MatchRepository{db: db}
}

// Create creates a new match with participants
func (r *MatchRepository) Create(match *model.Match) error {
	return r.db.Create(match).Error
}

// GetByID returns a match by ID with participants and user details
func (r *MatchRepository) GetByID(id uuid.UUID) (*model.Match, error) {
	var match model.Match
	err := r.db.Preload("Participants.User").First(&match, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &match, nil
}

// GetAll returns all matches with pagination, ordered by match_date DESC
func (r *MatchRepository) GetAll(limit, offset int) ([]*model.Match, error) {
	var matches []*model.Match
	query := r.db.Preload("Participants.User").Order("match_date DESC, created_at DESC")
	
	if limit > 0 {
		query = query.Limit(limit).Offset(offset)
	}
	
	err := query.Find(&matches).Error
	return matches, err
}

// GetRecent returns the most recent matches
func (r *MatchRepository) GetRecent(limit int) ([]*model.Match, error) {
	var matches []*model.Match
	err := r.db.Preload("Participants.User").
		Order("match_date DESC, created_at DESC").
		Limit(limit).
		Find(&matches).Error
	return matches, err
}

// GetByUserID returns all matches for a specific user
func (r *MatchRepository) GetByUserID(userID uuid.UUID, limit int) ([]*model.Match, error) {
	var matches []*model.Match
	
	query := r.db.Preload("Participants.User").
		Joins("JOIN match_participants ON match_participants.match_id = matches.id").
		Where("match_participants.user_id = ?", userID).
		Order("match_date DESC, created_at DESC")
	
	if limit > 0 {
		query = query.Limit(limit)
	}
	
	err := query.Find(&matches).Error
	return matches, err
}

// Update updates an existing match
func (r *MatchRepository) Update(match *model.Match) error {
	return r.db.Save(match).Error
}

// Delete deletes a match and its participants (cascade)
func (r *MatchRepository) Delete(id uuid.UUID) error {
	return r.db.Delete(&model.Match{}, "id = ?", id).Error
}

// CountTotal returns the total number of matches
func (r *MatchRepository) CountTotal() (int64, error) {
	var count int64
	err := r.db.Model(&model.Match{}).Count(&count).Error
	return count, err
}

// CountToday returns the number of matches today
func (r *MatchRepository) CountToday() (int64, error) {
	var count int64
	today := time.Now().Truncate(24 * time.Hour)
	tomorrow := today.Add(24 * time.Hour)
	
	err := r.db.Model(&model.Match{}).
		Where("match_date >= ? AND match_date < ?", today, tomorrow).
		Count(&count).Error
	return count, err
}

// Lock marks a match as locked (prevents editing)
func (r *MatchRepository) Lock(id uuid.UUID) error {
	return r.db.Model(&model.Match{}).
		Where("id = ?", id).
		Update("is_locked", true).Error
}
