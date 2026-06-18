package service

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/duyb/esport-score-tracker/internal/model"
	"github.com/duyb/esport-score-tracker/internal/repository"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"google.golang.org/api/idtoken"
	"gorm.io/gorm"
)

var (
	ErrInvalidGoogleToken  = errors.New("invalid or expired Google token")
	ErrGoogleAlreadyLinked = errors.New("this Google account is already linked to another player")
	ErrAlreadyLinked       = errors.New("this account already has a Google account linked")
)

// googleVerifier is a function that validates a Google ID token and returns its payload.
// The default is idtoken.Validate; tests inject a fake to avoid real network calls.
type googleVerifier func(ctx context.Context, idToken, audience string) (*idtoken.Payload, error)

type WcAuthService struct {
	userRepo     *repository.WcUserRepository
	wcRepo       *repository.WcRepository
	verifyGoogle googleVerifier
}

func NewWcAuthService(userRepo *repository.WcUserRepository, wcRepo *repository.WcRepository) *WcAuthService {
	return &WcAuthService{
		userRepo:     userRepo,
		wcRepo:       wcRepo,
		verifyGoogle: idtoken.Validate,
	}
}

// withVerifier returns a shallow copy of the service with the given token verifier.
// Intended for tests only.
func (s *WcAuthService) withVerifier(v googleVerifier) *WcAuthService {
	cp := *s
	cp.verifyGoogle = v
	return &cp
}

type WcClaims struct {
	WcUserID uuid.UUID `json:"wc_user_id"`
	Name     string    `json:"name"`
	IsAdmin  bool      `json:"is_admin"`
	jwt.RegisteredClaims
}

type WcAuthResponse struct {
	Token        string    `json:"token"`
	UserID       uuid.UUID `json:"user_id"`
	Name         string    `json:"name"`
	AvatarURL    *string   `json:"avatar_url"`
	IsAdmin      bool      `json:"is_admin"`
	GoogleLinked bool      `json:"google_linked"`
}

func (s *WcAuthService) Login(name, password string) (*WcAuthResponse, error) {
	user, err := s.userRepo.GetByName(strings.TrimSpace(name))
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("invalid name or password")
		}
		return nil, err
	}

	if user.PasswordHash == nil {
		return nil, fmt.Errorf("invalid name or password")
	}
	if err := bcrypt.CompareHashAndPassword([]byte(*user.PasswordHash), []byte(password)); err != nil {
		return nil, fmt.Errorf("invalid name or password")
	}

	token, err := s.signToken(user)
	if err != nil {
		return nil, err
	}
	return &WcAuthResponse{
		Token:        token,
		UserID:       user.ID,
		Name:         user.Name,
		AvatarURL:    user.AvatarURL,
		IsAdmin:      user.IsAdmin,
		GoogleLinked: user.GoogleLinked(),
	}, nil
}

// GoogleLoginOrCreate logs in an existing Google-linked account or creates a new one.
func (s *WcAuthService) GoogleLoginOrCreate(ctx context.Context, idTokenStr string) (*WcAuthResponse, error) {
	payload, err := s.verifyGoogleToken(ctx, idTokenStr)
	if err != nil {
		return nil, ErrInvalidGoogleToken
	}

	googleEmail, _ := payload.Claims["email"].(string)

	user, err := s.userRepo.GetByGoogleID(payload.Subject)
	if err == nil {
		// Silently persist email if not yet stored
		if googleEmail != "" && user.Email == nil {
			s.wcRepo.DB().Model(user).Update("email", googleEmail)
		}
		token, err := s.signToken(user)
		if err != nil {
			return nil, err
		}
		return &WcAuthResponse{
			Token:        token,
			UserID:       user.ID,
			Name:         user.Name,
			AvatarURL:    user.AvatarURL,
			IsAdmin:      user.IsAdmin,
			GoogleLinked: true,
		}, nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	// Auto-create new account
	googleName, _ := payload.Claims["name"].(string)
	googlePic, _ := payload.Claims["picture"].(string)
	if googleName == "" {
		googleName = "Player"
	}
	name := s.uniqueName(googleName)

	var emailPtr *string
	if googleEmail != "" {
		emailPtr = &googleEmail
	}
	newUser := &model.WcUser{
		GoogleID:  &payload.Subject,
		Name:      name,
		Email:     emailPtr,
		AvatarURL: &googlePic,
	}
	db := s.wcRepo.DB()
	err = db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(newUser).Error; err != nil {
			return err
		}
		return s.wcRepo.CreateWallet(tx, newUser.ID)
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create account: %w", err)
	}

	token, err := s.signToken(newUser)
	if err != nil {
		return nil, err
	}
	return &WcAuthResponse{
		Token:        token,
		UserID:       newUser.ID,
		Name:         newUser.Name,
		AvatarURL:    newUser.AvatarURL,
		IsAdmin:      newUser.IsAdmin,
		GoogleLinked: true,
	}, nil
}

// LinkGoogleToAccount links a Google identity to an already-authenticated WC account.
func (s *WcAuthService) LinkGoogleToAccount(ctx context.Context, userID uuid.UUID, idTokenStr string) (*string, error) {
	payload, err := s.verifyGoogleToken(ctx, idTokenStr)
	if err != nil {
		return nil, ErrInvalidGoogleToken
	}

	avatarURL, _ := payload.Claims["picture"].(string)
	email, _ := payload.Claims["email"].(string)
	updates := map[string]interface{}{
		"google_id":  payload.Subject,
		"avatar_url": avatarURL,
	}
	if email != "" {
		updates["email"] = email
	}
	result := s.wcRepo.DB().Model(&model.WcUser{}).
		Where("id = ? AND google_id IS NULL", userID).
		Updates(updates)
	if result.Error != nil {
		if isUniqueViolation(result.Error) {
			return nil, ErrGoogleAlreadyLinked
		}
		return nil, result.Error
	}
	if result.RowsAffected == 0 {
		return nil, ErrAlreadyLinked
	}
	return &avatarURL, nil
}

func (s *WcAuthService) VerifyToken(tokenStr string) (*WcClaims, error) {
	secret := os.Getenv("WC_JWT_SECRET")
	token, err := jwt.ParseWithClaims(tokenStr, &WcClaims{}, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method")
		}
		return []byte(secret), nil
	})
	if err != nil {
		return nil, err
	}
	claims, ok := token.Claims.(*WcClaims)
	if !ok || !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}
	return claims, nil
}

func (s *WcAuthService) verifyGoogleToken(ctx context.Context, idTokenStr string) (*idtoken.Payload, error) {
	clientID := os.Getenv("GOOGLE_CLIENT_ID")
	return s.verifyGoogle(ctx, idTokenStr, clientID)
}

func (s *WcAuthService) signToken(user *model.WcUser) (string, error) {
	secret := os.Getenv("WC_JWT_SECRET")
	claims := WcClaims{
		WcUserID: user.ID,
		Name:     user.Name,
		IsAdmin:  user.IsAdmin,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(7 * 24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}

// uniqueName returns name if available, otherwise appends a numeric suffix.
func (s *WcAuthService) uniqueName(base string) string {
	name := base
	for i := 1; i <= 99; i++ {
		if _, err := s.userRepo.GetByName(name); errors.Is(err, gorm.ErrRecordNotFound) {
			return name
		}
		name = fmt.Sprintf("%s%d", base, i)
	}
	return fmt.Sprintf("%s_%s", base, uuid.New().String()[:4])
}

func isUniqueViolation(err error) bool {
	return err != nil && strings.Contains(err.Error(), "unique")
}
