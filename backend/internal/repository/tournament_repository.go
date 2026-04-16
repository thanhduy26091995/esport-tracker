package repository

import (
	"github.com/duyb/esport-score-tracker/internal/model"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type TournamentRepository struct {
	db *gorm.DB
}

func NewTournamentRepository(db *gorm.DB) *TournamentRepository {
	return &TournamentRepository{db: db}
}

func (r *TournamentRepository) Create(t *model.Tournament) error {
	return r.db.Create(t).Error
}

func (r *TournamentRepository) GetAll() ([]*model.Tournament, error) {
	var tournaments []*model.Tournament
	err := r.db.
		Preload("Participants").
		Order("created_at DESC").
		Find(&tournaments).Error
	return tournaments, err
}

func (r *TournamentRepository) GetByID(id uuid.UUID) (*model.Tournament, error) {
	var t model.Tournament
	err := r.db.
		Preload("Participants").
		Preload("Participants.User").
		Preload("Matches").
		First(&t, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &t, nil
}

func (r *TournamentRepository) Update(t *model.Tournament) error {
	return r.db.Save(t).Error
}

func (r *TournamentRepository) Delete(id uuid.UUID) error {
	return r.db.Delete(&model.Tournament{}, "id = ?", id).Error
}

func (r *TournamentRepository) GetMatch(id uuid.UUID) (*model.TournamentMatch, error) {
	var m model.TournamentMatch
	err := r.db.First(&m, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &m, nil
}

func (r *TournamentRepository) SaveMatch(m *model.TournamentMatch) error {
	// Select("*") forces GORM to update all fields including zero values and nil pointers
	return r.db.Model(m).Select("*").Updates(m).Error
}

func (r *TournamentRepository) GetMatchesByTournamentID(tournamentID uuid.UUID) ([]*model.TournamentMatch, error) {
	var matches []*model.TournamentMatch
	err := r.db.Where("tournament_id = ?", tournamentID).
		Order("round ASC, match_order ASC").
		Find(&matches).Error
	return matches, err
}
