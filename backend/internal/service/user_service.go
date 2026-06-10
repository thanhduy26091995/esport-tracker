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

	"github.com/duyb/esport-score-tracker/internal/model"
	"github.com/duyb/esport-score-tracker/internal/repository"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type UserService struct {
	repo      *repository.UserRepository
	configSvc *ConfigService
}

func NewUserService(repo *repository.UserRepository, configSvc *ConfigService) *UserService {
	return &UserService{repo: repo, configSvc: configSvc}
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
	return avatarURL, nil
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
	return s.repo.ClearAvatarURL(userID)
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
	return s.repo.UpdateFavoriteClub(userID, club)
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

// GetLeaderboard returns the leaderboard with tier and win rate computed against the live config threshold.
func (s *UserService) GetLeaderboard(limit int) ([]*model.UserWithStats, error) {
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
}

// GetPaymentRanking returns active users sorted by total historical settlement money paid DESC,
// with tier and win rate computed against the live config threshold.
func (s *UserService) GetPaymentRanking() ([]*model.UserWithPaymentTotal, error) {
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
}
