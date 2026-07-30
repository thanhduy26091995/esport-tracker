package service

import (
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/duyb/esport-score-tracker/internal/cache"
	"github.com/duyb/esport-score-tracker/internal/model"
	"github.com/duyb/esport-score-tracker/internal/repository"
	"github.com/google/uuid"
	"golang.org/x/sync/singleflight"
	"gorm.io/gorm"
)

// Sentinel errors for head-to-head so the handler can map correct HTTP codes.
var (
	ErrSamePlayer     = errors.New("cannot compare a player with themselves")
	ErrPlayerNotFound = errors.New("player not found")
)

type UserService struct {
	repo      *repository.UserRepository
	matchRepo *repository.MatchRepository
	configSvc *ConfigService
	cache     cache.CacheStore
	group     singleflight.Group
}

func NewUserService(repo *repository.UserRepository, matchRepo *repository.MatchRepository, configSvc *ConfigService, c cache.CacheStore) *UserService {
	return &UserService{repo: repo, matchRepo: matchRepo, configSvc: configSvc, cache: c}
}

func (s *UserService) minMatchesForTier() int {
	if s.configSvc != nil {
		if v, err := s.configSvc.GetMinMatchesForTier(); err == nil {
			return v
		}
	}
	return defaultMinMatches
}

func (s *UserService) winRateThresholds() (pro float64, normal float64) {
	pro, normal = defaultProThreshold, defaultNormalThreshold
	if s.configSvc != nil {
		if v, err := s.configSvc.GetProWinRateThreshold(); err == nil {
			pro = v
		}
		if v, err := s.configSvc.GetNormalWinRateThreshold(); err == nil {
			normal = v
		}
	}
	return
}

// GetAll returns all active users with tier and win rate computed against the live config threshold.
func (s *UserService) GetAll() ([]*model.UserWithStats, error) {
	return cache.GetOrFetch(s.cache, &s.group, "esport:users:all", 10*time.Minute, func() ([]*model.UserWithStats, error) {
		users, err := s.repo.GetAll()
		if err != nil {
			return nil, err
		}
		minMatches := s.minMatchesForTier()
		proThres, normalThres := s.winRateThresholds()
		for _, u := range users {
			if u.TotalMatches < minMatches {
				u.WinRate = 0
				u.Tier = TierNormal
			} else {
				u.Tier = EvaluateTier(u.WinRate, u.TotalMatches, minMatches, proThres, normalThres)
			}
		}
		return users, nil
	})
}

// GetByID returns a user with computed win rate stats.
func (s *UserService) GetByID(id uuid.UUID) (*model.UserWithStats, error) {
	user, err := s.repo.GetByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("user not found")
		}
		return nil, err
	}
	return user, nil
}

func (s *UserService) invalidateUserCaches() {
	_ = s.cache.Delete("esport:users:all")
	_ = s.cache.Delete("esport:users:all:with-inactive")
	_ = s.cache.DeleteByPattern("esport:users:leaderboard:*")
	_ = s.cache.Delete("esport:users:payment-ranking")
}

// GetAllIncludingInactive returns all users (active + soft-deleted) with tier/win rate,
// active first. Used by the head-to-head player picker so historical matchups are selectable.
func (s *UserService) GetAllIncludingInactive() ([]*model.UserWithStats, error) {
	return cache.GetOrFetch(s.cache, &s.group, "esport:users:all:with-inactive", 10*time.Minute, func() ([]*model.UserWithStats, error) {
		users, err := s.repo.GetAllIncludingInactive()
		if err != nil {
			return nil, err
		}
		minMatches := s.minMatchesForTier()
		proThres, normalThres := s.winRateThresholds()
		for _, u := range users {
			if u.TotalMatches < minMatches {
				u.WinRate = 0
				u.Tier = TierNormal
			} else {
				u.Tier = EvaluateTier(u.WinRate, u.TotalMatches, minMatches, proThres, normalThres)
			}
		}
		return users, nil
	})
}

// GetHeadToHead returns the opponents-only head-to-head record between two players,
// oriented to player1. Aggregates over all match types (1v1/2v2/1v2); teammate
// encounters are excluded by the repository join. Soft-deleted players are accepted.
func (s *UserService) GetHeadToHead(p1, p2 uuid.UUID) (*model.HeadToHeadResponse, error) {
	if p1 == p2 {
		return nil, ErrSamePlayer
	}

	version, _ := s.cache.GetInt("esport:matches:version")
	key := fmt.Sprintf("esport:h2h:v%d:%s:%s", version, p1, p2)
	return cache.GetOrFetch(s.cache, &s.group, key, 5*time.Minute, func() (*model.HeadToHeadResponse, error) {
		user1, err := s.repo.GetByIDIncludingInactive(p1)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, ErrPlayerNotFound
			}
			return nil, err
		}
		user2, err := s.repo.GetByIDIncludingInactive(p2)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, ErrPlayerNotFound
			}
			return nil, err
		}

		rows, err := s.matchRepo.GetHeadToHeadMatches(p1, p2)
		if err != nil {
			return nil, err
		}

		// Lineups: only the top-N most-recent matches need full participant detail.
		recentRows := rows
		if len(recentRows) > h2hListLimit {
			recentRows = recentRows[:h2hListLimit]
		}
		lineups, err := s.h2hLineups(recentRows)
		if err != nil {
			return nil, err
		}

		return aggregateHeadToHead(user1, user2, rows, lineups), nil
	})
}

const h2hListLimit = 10

// aggregateHeadToHead builds the P1-oriented response from the (most-recent-first) rows.
// Pure: totals/form/streak are computed over the full history; recent matches (with
// lineups) are capped to h2hListLimit. Player2Wins is the complement (no draws exist).
func aggregateHeadToHead(user1, user2 *model.User, rows []model.H2HRow, lineups map[uuid.UUID][]model.H2HParticipant) *model.HeadToHeadResponse {
	resp := &model.HeadToHeadResponse{
		Player1:       toH2HPlayer(user1),
		Player2:       toH2HPlayer(user2),
		TotalMatches:  len(rows),
		Form:          make([]string, 0, h2hListLimit),
		RecentMatches: make([]model.H2HMatch, 0, h2hListLimit),
	}

	for _, r := range rows {
		p1Won := r.P1Team == r.WinnerTeam
		if p1Won {
			resp.Player1Wins++
		}
		if len(resp.Form) < h2hListLimit {
			resp.Form = append(resp.Form, winLossLabel(p1Won))
		}
	}
	resp.Player2Wins = resp.TotalMatches - resp.Player1Wins
	if resp.TotalMatches > 0 {
		resp.Player1WinRate = float64(resp.Player1Wins) / float64(resp.TotalMatches)
		resp.Player2WinRate = float64(resp.Player2Wins) / float64(resp.TotalMatches)
		resp.CurrentStreak = computeH2HStreak(rows, user1.ID, user2.ID)
	}

	limit := len(rows)
	if limit > h2hListLimit {
		limit = h2hListLimit
	}
	for _, r := range rows[:limit] {
		resp.RecentMatches = append(resp.RecentMatches, model.H2HMatch{
			MatchID:      r.MatchID,
			MatchType:    r.MatchType,
			MatchDate:    r.MatchDate,
			WinnerTeam:   r.WinnerTeam,
			Player1Team:  r.P1Team,
			Player1Won:   r.P1Team == r.WinnerTeam,
			Participants: lineups[r.MatchID],
		})
	}
	return resp
}

func winLossLabel(won bool) string {
	if won {
		return "W"
	}
	return "L"
}

func toH2HPlayer(u *model.User) model.H2HPlayer {
	return model.H2HPlayer{
		ID:           u.ID,
		Name:         u.Name,
		AvatarURL:    u.AvatarURL,
		FavoriteClub: u.FavoriteClub,
		Tier:         u.Tier,
		IsActive:     u.IsActive,
	}
}

// computeH2HStreak returns the current run of consecutive wins for whichever player
// won the most-recent encounter. Assumes rows are most-recent first and non-empty.
func computeH2HStreak(rows []model.H2HRow, p1, p2 uuid.UUID) model.H2HStreak {
	firstP1Won := rows[0].P1Team == rows[0].WinnerTeam
	count := 0
	for _, r := range rows {
		if (r.P1Team == r.WinnerTeam) == firstP1Won {
			count++
		} else {
			break
		}
	}
	winner := p2
	if firstP1Won {
		winner = p1
	}
	return model.H2HStreak{PlayerID: &winner, Count: count}
}

// h2hLineups fetches full participant lineups for the given rows, keyed by match ID.
func (s *UserService) h2hLineups(rows []model.H2HRow) (map[uuid.UUID][]model.H2HParticipant, error) {
	lineups := make(map[uuid.UUID][]model.H2HParticipant, len(rows))
	if len(rows) == 0 {
		return lineups, nil
	}
	ids := make([]uuid.UUID, 0, len(rows))
	for _, r := range rows {
		ids = append(ids, r.MatchID)
	}
	matches, err := s.matchRepo.GetMatchesWithParticipants(ids)
	if err != nil {
		return nil, err
	}
	for _, m := range matches {
		parts := make([]model.H2HParticipant, 0, len(m.Participants))
		for _, p := range m.Participants {
			parts = append(parts, model.H2HParticipant{
				UserID:    p.UserID,
				Name:      p.User.Name,
				AvatarURL: p.User.AvatarURL,
				Team:      p.TeamNumber,
			})
		}
		lineups[m.ID] = parts
	}
	return lineups, nil
}

// CreateUser creates a new user with validation
func (s *UserService) CreateUser(name string, tier string, handicapRate float64) (*model.User, error) {
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

	// Apply defaults and validate optional fields
	if tier == "" {
		tier = "normal"
	}
	if tier != "pro" && tier != "normal" && tier != "noob" {
		return nil, fmt.Errorf("tier must be one of: pro, normal, noob")
	}
	if handicapRate < 0 || handicapRate > 5 {
		return nil, fmt.Errorf("handicap_rate must be between 0 and 5")
	}

	// Create user
	user := &model.User{
		Name:         name,
		CurrentScore: 0,
		IsActive:     true,
		Tier:         tier,
		HandicapRate: handicapRate,
	}

	if err := s.repo.Create(user); err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}
	s.invalidateUserCaches()
	return user, nil
}

// UpdateUser updates a user's name, tier, and handicap_rate with validation
func (s *UserService) UpdateUser(id uuid.UUID, name string, tier string, handicapRate float64) (*model.UserWithStats, error) {
	// Get existing user (includes win rate stats)
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
		if tier != "pro" && tier != "normal" && tier != "noob" {
			return nil, fmt.Errorf("tier must be one of: pro, normal, noob")
		}
		user.Tier = tier
	}

	user.HandicapRate = handicapRate

	if err := s.repo.Update(&user.User); err != nil {
		return nil, fmt.Errorf("failed to update user: %w", err)
	}
	s.invalidateUserCaches()
	return user, nil
}

var validClubSlugs = map[string]bool{
	// Premier League
	"man-city": true, "liverpool": true, "man-utd": true, "chelsea": true,
	"arsenal": true, "spurs": true, "newcastle": true, "aston-villa": true,
	"west-ham": true, "everton": true,
	// La Liga
	"real-madrid": true, "barcelona": true, "atletico": true, "sevilla": true,
	"betis": true, "valencia": true, "villarreal": true,
	// Bundesliga
	"bayern": true, "dortmund": true, "rb-leipzig": true, "leverkusen": true,
	"frankfurt": true, "gladbach": true,
	// Serie A
	"juventus": true, "inter": true, "ac-milan": true, "napoli": true,
	"roma": true, "lazio": true, "atalanta": true, "fiorentina": true,
	// Ligue 1
	"psg": true, "marseille": true, "lyon": true, "monaco": true, "lille": true,
	// Others
	"porto": true, "benfica": true, "ajax": true, "flamengo": true,
	// Default
	"none": true,
}

var allowedAvatarMIME = map[string]string{
	"image/jpeg": "jpg",
	"image/png":  "png",
	"image/gif":  "gif",
	"image/webp": "webp",
}

const avatarDir = "uploads/avatars"

// UploadAvatar validates and stores an avatar file, returning the public URL.
func (s *UserService) UploadAvatar(userID uuid.UUID, file multipart.File, header *multipart.FileHeader) (string, error) {
	if header.Size > 5<<20 {
		return "", fmt.Errorf("file too large (max 5 MB)")
	}

	buf := make([]byte, 512)
	n, _ := file.Read(buf)
	mimeType := http.DetectContentType(buf[:n])
	ext, ok := allowedAvatarMIME[mimeType]
	if !ok {
		return "", fmt.Errorf("unsupported file type: %s", mimeType)
	}

	if _, err := file.Seek(0, io.SeekStart); err != nil {
		return "", fmt.Errorf("failed to read file")
	}

	if err := os.MkdirAll(avatarDir, 0755); err != nil {
		return "", fmt.Errorf("failed to prepare storage")
	}

	filename := uuid.New().String() + "." + ext
	dst := filepath.Join(avatarDir, filename)

	out, err := os.Create(dst)
	if err != nil {
		return "", fmt.Errorf("failed to save file")
	}
	defer out.Close()
	if _, err := io.Copy(out, file); err != nil {
		os.Remove(dst)
		return "", fmt.Errorf("failed to write file")
	}

	// Delete previous avatar file
	existing, _ := s.repo.GetByID(userID)
	if existing != nil && existing.AvatarURL != nil {
		oldPath := strings.TrimPrefix(*existing.AvatarURL, "/")
		os.Remove(oldPath)
	}

	avatarURL := "/" + dst
	if err := s.repo.UpdateAvatarURL(userID, avatarURL); err != nil {
		os.Remove(dst)
		return "", fmt.Errorf("failed to persist avatar URL")
	}
	_ = s.cache.Delete("esport:users:all")
	return avatarURL, nil
}

// SetAvatarURL sets a remote URL as the avatar without uploading a file.
func (s *UserService) SetAvatarURL(userID uuid.UUID, avatarURL string) error {
	if _, err := s.repo.GetByID(userID); err != nil {
		return fmt.Errorf("user not found")
	}
	if err := s.repo.UpdateAvatarURL(userID, avatarURL); err != nil {
		return err
	}
	_ = s.cache.Delete("esport:users:all")
	return nil
}

// DeleteAvatar removes the avatar file and clears the URL.
func (s *UserService) DeleteAvatar(userID uuid.UUID) error {
	existing, err := s.repo.GetByID(userID)
	if err != nil {
		return fmt.Errorf("user not found")
	}
	if existing.AvatarURL != nil {
		oldPath := strings.TrimPrefix(*existing.AvatarURL, "/")
		os.Remove(oldPath)
	}
	if err := s.repo.ClearAvatarURL(userID); err != nil {
		return err
	}
	_ = s.cache.Delete("esport:users:all")
	return nil
}

// UpdateClub sets the favorite club for a user. Empty string clears it.
func (s *UserService) UpdateClub(userID uuid.UUID, club string) error {
	if club != "" && !validClubSlugs[club] {
		return fmt.Errorf("unknown club slug: %s", club)
	}
	_, err := s.repo.GetByID(userID)
	if err != nil {
		return fmt.Errorf("user not found")
	}
	if err := s.repo.UpdateFavoriteClub(userID, club); err != nil {
		return err
	}
	_ = s.cache.Delete("esport:users:all")
	return nil
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
	s.invalidateUserCaches()
	return nil
}

// GetLeaderboard returns the leaderboard with tier and win rate computed against the live config threshold.
func (s *UserService) GetLeaderboard(limit int) ([]*model.UserWithStats, error) {
	key := fmt.Sprintf("esport:users:leaderboard:%d", limit)
	return cache.GetOrFetch(s.cache, &s.group, key, 5*time.Minute, func() ([]*model.UserWithStats, error) {
		users, err := s.repo.GetLeaderboard(limit)
		if err != nil {
			return nil, err
		}
		minMatches := s.minMatchesForTier()
		proThres, normalThres := s.winRateThresholds()
		for _, u := range users {
			if u.TotalMatches < minMatches {
				u.WinRate = 0
				u.Tier = TierNormal
			} else {
				u.Tier = EvaluateTier(u.WinRate, u.TotalMatches, minMatches, proThres, normalThres)
			}
		}
		return users, nil
	})
}

// GetPaymentRanking returns active users sorted by total historical settlement money paid DESC,
// with tier and win rate computed against the live config threshold.
func (s *UserService) GetPaymentRanking() ([]*model.UserWithPaymentTotal, error) {
	return cache.GetOrFetch(s.cache, &s.group, "esport:users:payment-ranking", 10*time.Minute, func() ([]*model.UserWithPaymentTotal, error) {
		users, err := s.repo.GetPaymentRanking()
		if err != nil {
			return nil, err
		}
		minMatches := s.minMatchesForTier()
		proThres, normalThres := s.winRateThresholds()
		for _, u := range users {
			if u.TotalMatches < minMatches {
				u.WinRate = 0
				u.Tier = TierNormal
			} else {
				u.Tier = EvaluateTier(u.WinRate, u.TotalMatches, minMatches, proThres, normalThres)
			}
		}
		return users, nil
	})
}
