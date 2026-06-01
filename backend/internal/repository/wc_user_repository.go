package repository

import (
	"github.com/duyb/esport-score-tracker/internal/model"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type WcUserRepository struct {
	db *gorm.DB
}

func NewWcUserRepository(db *gorm.DB) *WcUserRepository {
	return &WcUserRepository{db: db}
}

func (r *WcUserRepository) Create(user *model.WcUser) error {
	return r.db.Create(user).Error
}

func (r *WcUserRepository) GetByName(name string) (*model.WcUser, error) {
	var user model.WcUser
	err := r.db.Where("name = ?", name).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *WcUserRepository) GetByID(id uuid.UUID) (*model.WcUser, error) {
	var user model.WcUser
	err := r.db.First(&user, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *WcUserRepository) GetAll() ([]*model.WcUser, error) {
	var users []*model.WcUser
	err := r.db.Order("name ASC").Find(&users).Error
	return users, err
}

func (r *WcUserRepository) SetAdminRole(id uuid.UUID, isAdmin bool) error {
	return r.db.Model(&model.WcUser{}).
		Where("id = ?", id).
		Update("is_admin", isAdmin).Error
}
