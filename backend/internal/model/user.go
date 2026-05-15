package model

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID           uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	Name         string    `gorm:"type:varchar(100);not null;uniqueIndex:idx_users_active_name,where:is_active = true" json:"name"`
	CurrentScore int       `gorm:"default:0" json:"current_score"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
	IsActive     bool      `gorm:"default:true" json:"is_active"`
	Tier         string    `gorm:"type:varchar(10);default:'normal'" json:"tier"`
	HandicapRate float64   `gorm:"default:0.0" json:"handicap_rate"`
}

func (User) TableName() string {
	return "users"
}

type UserWithStats struct {
	User
	WinRate      float64 `gorm:"column:win_rate" json:"win_rate"`
	TotalMatches int     `gorm:"column:total_matches" json:"total_matches"`
	WonMatches   int     `gorm:"column:won_matches" json:"won_matches"`
}

type UserWithPaymentTotal struct {
	User
	TotalPaid       int     `json:"total_paid"`
	TotalDebtPoints int     `json:"total_debt_points"`
	WinRate         float64 `gorm:"column:win_rate" json:"win_rate"`
	TotalMatches    int     `gorm:"column:total_matches" json:"total_matches"`
	WonMatches      int     `gorm:"column:won_matches" json:"won_matches"`
}
