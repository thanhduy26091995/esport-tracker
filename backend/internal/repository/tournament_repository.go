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
	// Explicitly select columns to update, including zero-value int (effective_winner=0 for draw)
	// and nullable pointer (match_id=NULL when reverting). GORM's Updates() skips zero values
	// by default, so we must name each column explicitly.
	return r.db.Model(m).Select(
		"actual_score1", "actual_score2", "effective_winner", "status", "match_id",
	).Updates(map[string]interface{}{
		"actual_score1":    m.ActualScore1,
		"actual_score2":    m.ActualScore2,
		"effective_winner": m.EffectiveWinner,
		"status":           m.Status,
		"match_id":         m.MatchID,
	}).Error
}

func (r *TournamentRepository) GetMatchesByTournamentID(tournamentID uuid.UUID) ([]*model.TournamentMatch, error) {
	var matches []*model.TournamentMatch
	err := r.db.Where("tournament_id = ?", tournamentID).
		Order("round ASC, match_order ASC").
		Find(&matches).Error
	return matches, err
}
