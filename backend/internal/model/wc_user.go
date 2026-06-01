package model

import (
	"time"

	"github.com/google/uuid"
)

type WcUser struct {
	ID           uuid.UUID  `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	Name         string     `gorm:"type:varchar(100);not null;uniqueIndex" json:"name"`
	PasswordHash string     `gorm:"type:varchar(255);not null" json:"-"`
	IsAdmin      bool       `gorm:"default:false" json:"is_admin"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
}

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
