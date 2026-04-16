package service

import (
	"errors"
	"fmt"
	"time"

	"github.com/duyb/esport-score-tracker/internal/model"
	"github.com/duyb/esport-score-tracker/internal/repository"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type MatchService struct {
	matchRepo         *repository.MatchRepository
	userRepo          *repository.UserRepository
	settlementService *SettlementService
	configService     *ConfigService
	db                *gorm.DB
}

func NewMatchService(matchRepo *repository.MatchRepository, userRepo *repository.UserRepository, settlementService *SettlementService, configService *ConfigService, db *gorm.DB) *MatchService {
	return &MatchService{
		matchRepo:         matchRepo,
		userRepo:          userRepo,
		settlementService: settlementService,
		configService:     configService,
		db:                db,
	}
}

// CreateMatchRequest represents the request to create a match
type CreateMatchRequest struct {
	MatchType         string      `json:"match_type" binding:"required"` // "1v1" or "2v2"
	Team1             []uuid.UUID `json:"team1" binding:"required"`
	Team2             []uuid.UUID `json:"team2" binding:"required"`
	WinnerTeam        int         `json:"winner_team"` // 0 = draw, 1 or 2
	MatchDate         *time.Time  `json:"match_date,omitempty"`
	PointsPerWin      int         `json:"points_per_win,omitempty"` // 0 = use config default
	TournamentMatchID *uuid.UUID  `json:"tournament_match_id,omitempty"`
}

// CreateMatch creates a new match with participants and updates user scores
func (s *MatchService) CreateMatch(req *CreateMatchRequest) (*model.Match, error) {
	// Validate match type
	if req.MatchType != "1v1" && req.MatchType != "2v2" {
		return nil, errors.New("match_type must be '1v1' or '2v2'")
	}

	// Validate team sizes
	expectedSize := 1
	if req.MatchType == "2v2" {
		expectedSize = 2
	}
	if len(req.Team1) != expectedSize || len(req.Team2) != expectedSize {
		return nil, fmt.Errorf("each team must have exactly %d player(s) for %s", expectedSize, req.MatchType)
	}

	// Validate winner team
	if req.WinnerTeam != 0 && req.WinnerTeam != 1 && req.WinnerTeam != 2 {
		return nil, errors.New("winner_team must be 0, 1, or 2")
	}

	// Validate no duplicate players
	allPlayers := append(req.Team1, req.Team2...)
	seen := make(map[uuid.UUID]bool)
	for _, playerID := range allPlayers {
		if seen[playerID] {
			return nil, errors.New("duplicate player found in match")
		}
		seen[playerID] = true
	}

	// Validate all users exist
	for _, playerID := range allPlayers {
		_, err := s.userRepo.GetByID(playerID)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, fmt.Errorf("user with ID %s not found", playerID)
			}
			return nil, err
		}
	}

	// Set match date
	matchDate := time.Now()
	if req.MatchDate != nil {
		matchDate = *req.MatchDate
	}

	// Determine points per win (request value takes precedence over config)
	basePoints := req.PointsPerWin
	if basePoints <= 0 {
		basePoints, _ = s.configService.GetPointsPerWin()
		if basePoints <= 0 {
			basePoints = 1
		}
	}

	// Begin transaction
	tx := s.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Create match
	match := &model.Match{
		MatchType:         req.MatchType,
		WinnerTeam:        req.WinnerTeam,
		MatchDate:         matchDate,
		IsLocked:          false,
		TournamentMatchID: req.TournamentMatchID,
	}

	if err := tx.Create(match).Error; err != nil {
		tx.Rollback()
		return nil, err
	}

	// Create participants and update scores
	// Team 1 participants
	for _, userID := range req.Team1 {
		pointChange := 0
		if req.WinnerTeam != 0 {
			if req.WinnerTeam == 1 {
				pointChange = basePoints
			} else {
				pointChange = -basePoints
			}
		}

		participant := &model.MatchParticipant{
			MatchID:     match.ID,
			UserID:      userID,
			TeamNumber:  1,
			PointChange: pointChange,
		}

		if err := tx.Create(participant).Error; err != nil {
			tx.Rollback()
			return nil, err
		}

		// Update user score
		if err := tx.Model(&model.User{}).
			Where("id = ?", userID).
			Update("current_score", gorm.Expr("current_score + ?", pointChange)).
			Error; err != nil {
			tx.Rollback()
			return nil, err
		}
	}

	// Team 2 participants
	for _, userID := range req.Team2 {
		pointChange := 0
		if req.WinnerTeam != 0 {
			if req.WinnerTeam == 2 {
				pointChange = basePoints
			} else {
				pointChange = -basePoints
			}
		}

		participant := &model.MatchParticipant{
			MatchID:     match.ID,
			UserID:      userID,
			TeamNumber:  2,
			PointChange: pointChange,
		}

		if err := tx.Create(participant).Error; err != nil {
			tx.Rollback()
			return nil, err
		}

		// Update user score
		if err := tx.Model(&model.User{}).
			Where("id = ?", userID).
			Update("current_score", gorm.Expr("current_score + ?", pointChange)).
			Error; err != nil {
			tx.Rollback()
			return nil, err
		}
	}

	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		return nil, err
	}

	// Fetch the created match with participants
	createdMatch, err := s.matchRepo.GetByID(match.ID)
	if err != nil {
		return nil, err
	}

	// Check if any participant needs settlement (auto-trigger)
	// Skip when WinnerTeam=0 (draw/no score change) — no scores changed, no settlement possible
	if req.WinnerTeam != 0 && s.settlementService != nil {
		autoSettlement, _ := s.configService.GetAutoSettlement()
		if autoSettlement {
			allPlayers := append(req.Team1, req.Team2...)
			for _, playerID := range allPlayers {
				_ = s.settlementService.CheckAndTriggerSettlement(playerID)
			}
		}
	}

	return createdMatch, nil
}

// GetMatchByID returns a match by ID
func (s *MatchService) GetMatchByID(id uuid.UUID) (*model.Match, error) {
	return s.matchRepo.GetByID(id)
}

// GetAllMatches returns all matches with pagination
func (s *MatchService) GetAllMatches(limit, offset int) ([]*model.Match, error) {
	return s.matchRepo.GetAll(limit, offset)
}

// GetRecentMatches returns recent matches
func (s *MatchService) GetRecentMatches(limit int) ([]*model.Match, error) {
	if limit <= 0 {
		limit = 10
	}
	return s.matchRepo.GetRecent(limit)
}

// GetMatchesByUserID returns all matches for a user
func (s *MatchService) GetMatchesByUserID(userID uuid.UUID, limit int) ([]*model.Match, error) {
	return s.matchRepo.GetByUserID(userID, limit)
}

// DeleteMatch deletes a match and reverts the score changes
func (s *MatchService) DeleteMatch(id uuid.UUID) error {
	// Get the match first
	match, err := s.matchRepo.GetByID(id)
	if err != nil {
		return err
	}

	// Check if match is locked
	if match.IsLocked {
		return errors.New("cannot delete a locked match")
	}

	// Begin transaction
	tx := s.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Revert score changes for all participants
	for _, participant := range match.Participants {
		// Subtract the point change to revert it
		if err := tx.Model(&model.User{}).
			Where("id = ?", participant.UserID).
			Update("current_score", gorm.Expr("current_score - ?", participant.PointChange)).
			Error; err != nil {
			tx.Rollback()
			return err
		}
	}

	// Delete the match (cascades to participants)
	if err := tx.Delete(&model.Match{}, "id = ?", id).Error; err != nil {
		tx.Rollback()
		return err
	}

	// Commit transaction
	return tx.Commit().Error
}

// GetMatchStats returns statistics about matches
func (s *MatchService) GetMatchStats() (map[string]interface{}, error) {
	totalMatches, err := s.matchRepo.CountTotal()
	if err != nil {
		return nil, err
	}

	todayMatches, err := s.matchRepo.CountToday()
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"total": totalMatches,
		"today": todayMatches,
	}, nil
}
