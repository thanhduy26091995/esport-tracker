package service

import (
	"errors"
	"fmt"
	"time"

	"github.com/duyb/esport-score-tracker/internal/cache"
	"github.com/duyb/esport-score-tracker/internal/model"
	"github.com/duyb/esport-score-tracker/internal/repository"
	"github.com/google/uuid"
	"golang.org/x/sync/singleflight"
	"gorm.io/gorm"
)

type MatchService struct {
	matchRepo         *repository.MatchRepository
	userRepo          *repository.UserRepository
	settlementService *SettlementService
	configService     *ConfigService
	tierService       *TierService
	db                *gorm.DB
	cache             cache.CacheStore
	group             singleflight.Group
}

func NewMatchService(matchRepo *repository.MatchRepository, userRepo *repository.UserRepository, settlementService *SettlementService, configService *ConfigService, tierService *TierService, db *gorm.DB, c cache.CacheStore) *MatchService {
	return &MatchService{
		matchRepo:         matchRepo,
		userRepo:          userRepo,
		settlementService: settlementService,
		configService:     configService,
		tierService:       tierService,
		db:                db,
		cache:             c,
	}
}

// matchVersionedKey builds the versioned cache key for paginated match queries.
// A version counter (esport:matches:version) is incremented on each write,
// making old page keys unreachable orphans that expire via their 2-min TTL.
func (s *MatchService) matchVersionedKey(limit, offset int, playerID *uuid.UUID) string {
	version, _ := s.cache.GetInt("esport:matches:version")
	pid := "all"
	if playerID != nil {
		pid = playerID.String()
	}
	return fmt.Sprintf("esport:matches:v%d:%d:%d:%s", version, limit, offset, pid)
}

func (s *MatchService) invalidateMatchCaches() {
	_, _ = s.cache.Incr("esport:matches:version")
	_ = s.cache.Delete("esport:users:leaderboard")
	_ = s.cache.Delete("esport:users:all")
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

// teamSizes maps match_type to required [team1Size, team2Size].
var teamSizes = map[string][2]int{
	"1v1": {1, 1},
	"2v2": {2, 2},
	"1v2": {1, 2},
}

// calcPointChange returns the point delta for one participant.
// For 1v2: solo (team1) wins → +2×base, loses → -2×base;
//
//	each duo (team2) member wins → +base, loses → -base.
//
// For symmetric types (1v1, 2v2): winner → +base, loser → -base.
func calcPointChange(matchType string, teamNumber, winnerTeam, base int) int {
	if winnerTeam == 0 {
		return 0
	}
	won := teamNumber == winnerTeam
	if matchType == "1v2" {
		if teamNumber == 1 {
			if won {
				return 2 * base
			}
			return -2 * base
		}
		if won {
			return base
		}
		return -base
	}
	if won {
		return base
	}
	return -base
}

// CreateMatch creates a new match with participants and updates user scores
func (s *MatchService) CreateMatch(req *CreateMatchRequest) (*model.Match, error) {
	// Validate match type
	sz, ok := teamSizes[req.MatchType]
	if !ok {
		return nil, errors.New("match_type must be '1v1', '2v2', or '1v2'")
	}

	// Validate team sizes
	if len(req.Team1) != sz[0] || len(req.Team2) != sz[1] {
		return nil, fmt.Errorf("for %s, team1 needs %d player(s) and team2 needs %d player(s)", req.MatchType, sz[0], sz[1])
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
		pointChange := calcPointChange(req.MatchType, 1, req.WinnerTeam, basePoints)

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
		pointChange := calcPointChange(req.MatchType, 2, req.WinnerTeam, basePoints)

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

	// Recalculate tier for all participants post-commit (non-fatal).
	_ = s.tierService.RecalculateForUsers(allPlayers)
	s.invalidateMatchCaches()
	return createdMatch, nil
}

// GetMatchByID returns a match by ID
func (s *MatchService) GetMatchByID(id uuid.UUID) (*model.Match, error) {
	return s.matchRepo.GetByID(id)
}

// GetAllMatches returns matches with pagination, served from versioned cache when possible.
func (s *MatchService) GetAllMatches(limit, offset int, playerID *uuid.UUID) ([]*model.Match, error) {
	key := s.matchVersionedKey(limit, offset, playerID)
	return cache.GetOrFetch(s.cache, &s.group, key, 2*time.Minute, func() ([]*model.Match, error) {
		return s.matchRepo.GetAllFiltered(limit, offset, playerID)
	})
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
	if err := tx.Commit().Error; err != nil {
		return err
	}

	// Recalculate tier for all participants post-commit (non-fatal).
	participantIDs := make([]uuid.UUID, len(match.Participants))
	for i, p := range match.Participants {
		participantIDs[i] = p.UserID
	}
	_ = s.tierService.RecalculateForUsers(participantIDs)
	s.invalidateMatchCaches()
	return nil
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
