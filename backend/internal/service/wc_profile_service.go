package service

import (
	"fmt"
	"strings"

	"github.com/duyb/esport-score-tracker/internal/model"
	"github.com/duyb/esport-score-tracker/internal/repository"
	"github.com/google/uuid"
)

var ErrNameTaken = fmt.Errorf("name is already taken")

type WcProfileService struct {
	userRepo *repository.WcUserRepository
}

func NewWcProfileService(userRepo *repository.WcUserRepository) *WcProfileService {
	return &WcProfileService{userRepo: userRepo}
}

func (s *WcProfileService) GetProfile(userID uuid.UUID) (*model.WcUser, error) {
	return s.userRepo.GetByID(userID)
}

func (s *WcProfileService) UpdateProfile(userID uuid.UUID, name *string, avatarURL *string) (*model.WcUser, error) {
	updates := map[string]interface{}{}

	if name != nil {
		n := strings.TrimSpace(*name)
		if len(n) < 2 {
			return nil, fmt.Errorf("name must be at least 2 characters")
		}
		updates["name"] = n
	}
	if avatarURL != nil {
		updates["avatar_url"] = *avatarURL
	}

	if len(updates) == 0 {
		return s.userRepo.GetByID(userID)
	}

	result := s.userRepo.DB().Model(&model.WcUser{}).Where("id = ?", userID).Updates(updates)
	if result.Error != nil {
		if isUniqueViolation(result.Error) {
			return nil, ErrNameTaken
		}
		return nil, result.Error
	}
	return s.userRepo.GetByID(userID)
}
