package model

import (
	"time"

	"github.com/google/uuid"
)

type WcUser struct {
	ID           uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	Name         string    `gorm:"type:varchar(100);not null;uniqueIndex" json:"name"`
	PasswordHash *string   `gorm:"type:varchar(255)" json:"-"`
	GoogleID     *string   `gorm:"type:varchar(100);uniqueIndex" json:"-"`
	Email        *string   `gorm:"type:varchar(255)" json:"-"`
	AvatarURL    *string   `gorm:"type:varchar(500)" json:"avatar_url"`
	IsAdmin      bool      `gorm:"default:false" json:"is_admin"`
	IsBlocked    bool      `gorm:"not null;default:false" json:"is_blocked"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

func (u *WcUser) GoogleLinked() bool { return u.GoogleID != nil }

func (WcUser) TableName() string {
	return "wc_users"
}

type WcConfig struct {
	ID        int        `gorm:"primaryKey;autoIncrement:false" json:"id"`
	IsEnabled bool       `gorm:"default:false" json:"is_enabled"`
	UpdatedAt time.Time  `json:"updated_at"`
	UpdatedBy *uuid.UUID `gorm:"type:uuid" json:"updated_by,omitempty"`
}

func (WcConfig) TableName() string {
	return "wc_config"
}
