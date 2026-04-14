package model

import (
	"time"

	"github.com/google/uuid"
)

type Match struct {
	ID           uuid.UUID          `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	MatchType    string             `gorm:"type:varchar(10);not null" json:"match_type"` // "1v1" or "2v2"
	WinnerTeam   int                `gorm:"not null" json:"winner_team"`                  // 1 or 2
	MatchDate    time.Time          `gorm:"default:now()" json:"match_date"`
	RecordedBy   string             `gorm:"type:varchar(100)" json:"recorded_by"`
	CreatedAt    time.Time          `json:"created_at"`
	IsLocked     bool               `gorm:"default:false" json:"is_locked"`
	Participants []MatchParticipant `gorm:"foreignKey:MatchID;constraint:OnDelete:CASCADE" json:"participants,omitempty"`
}

func (Match) TableName() string {
	return "matches"
}

type MatchParticipant struct {
	ID          uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	MatchID     uuid.UUID `gorm:"type:uuid;not null" json:"match_id"`
	UserID      uuid.UUID `gorm:"type:uuid;not null" json:"user_id"`
	TeamNumber  int       `gorm:"not null" json:"team_number"`  // 1 or 2
	PointChange int       `gorm:"not null" json:"point_change"` // +1 or -1
	User        User      `gorm:"foreignKey:UserID" json:"user,omitempty"`
}

func (MatchParticipant) TableName() string {
	return "match_participants"
}
