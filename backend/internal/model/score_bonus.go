package model

import (
	"time"

	"github.com/google/uuid"
)

type ScoreBonus struct {
	ID          uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	UserID      uuid.UUID `gorm:"type:uuid;not null;index" json:"user_id"`
	Points      int       `gorm:"not null" json:"points"`
	Description string    `gorm:"type:text" json:"description"`
	RecordedBy  string    `gorm:"type:varchar(100)" json:"recorded_by"`
	BonusDate   time.Time `gorm:"default:now()" json:"bonus_date"`
	CreatedAt   time.Time `json:"created_at"`
	User        User      `gorm:"foreignKey:UserID" json:"user,omitempty"`
}

func (ScoreBonus) TableName() string {
	return "score_bonuses"
}
