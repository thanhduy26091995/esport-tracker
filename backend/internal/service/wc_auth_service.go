package service

import (
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
	"gorm.io/gorm"
)

type WcAuthService struct {
	userRepo *repository.WcUserRepository
	wcRepo   *repository.WcRepository
}

func NewWcAuthService(userRepo *repository.WcUserRepository, wcRepo *repository.WcRepository) *WcAuthService {
	return &WcAuthService{userRepo: userRepo, wcRepo: wcRepo}
}

type WcClaims struct {
	WcUserID uuid.UUID `json:"wc_user_id"`
	Name     string    `json:"name"`
	IsAdmin  bool      `json:"is_admin"`
	jwt.RegisteredClaims
}

func (s *WcAuthService) Register(name, password string) (string, *model.WcUser, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return "", nil, fmt.Errorf("name cannot be empty")
	}
	if len(name) < 2 {
		return "", nil, fmt.Errorf("name must be at least 2 characters")
	}

	_, err := s.userRepo.GetByName(name)
	if err == nil {
		return "", nil, fmt.Errorf("name '%s' is already taken", name)
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return "", nil, err
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), 12)
	if err != nil {
		return "", nil, fmt.Errorf("failed to hash password: %w", err)
	}

	user := &model.WcUser{
		Name:         name,
		PasswordHash: string(hash),
		IsAdmin:      false,
	}

	// Create user + wallet in one transaction
	db := s.wcRepo.DB()
	err = db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(user).Error; err != nil {
			return err
		}
		return s.wcRepo.CreateWallet(tx, user.ID)
	})
	if err != nil {
		return "", nil, fmt.Errorf("failed to register: %w", err)
	}

	token, err := s.signToken(user)
	if err != nil {
		return "", nil, err
	}
	return token, user, nil
}

func (s *WcAuthService) Login(name, password string) (string, *model.WcUser, error) {
	user, err := s.userRepo.GetByName(strings.TrimSpace(name))
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return "", nil, fmt.Errorf("invalid name or password")
		}
		return "", nil, err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return "", nil, fmt.Errorf("invalid name or password")
	}

	token, err := s.signToken(user)
	if err != nil {
		return "", nil, err
	}
	return token, user, nil
}

func (s *WcAuthService) ResetPassword(name string) error {
	user, err := s.userRepo.GetByName(strings.TrimSpace(name))
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("user '%s' not found", name)
		}
		return err
	}

	newPassword := user.Name + "_@123"
	hash, err := bcrypt.GenerateFromPassword([]byte(newPassword), 12)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	return s.wcRepo.DB().Model(&model.WcUser{}).
		Where("id = ?", user.ID).
		Update("password_hash", string(hash)).Error
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
