package model

import "time"

type SiteAccessConfig struct {
	ID         int       `gorm:"primaryKey;autoIncrement" json:"id"`
	Question   string    `gorm:"type:text;not null;default:''" json:"question"`
	AnswerHash string    `gorm:"type:varchar(64);not null;default:''" json:"-"`
	Enabled    bool      `gorm:"not null;default:false" json:"enabled"`
	UpdatedAt  time.Time `json:"updated_at"`
}

func (SiteAccessConfig) TableName() string { return "site_access_config" }
