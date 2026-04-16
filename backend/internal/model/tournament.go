package model

import (
	"time"

	"github.com/google/uuid"
)

type Tournament struct {
	ID           uuid.UUID               `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	Name         string                  `gorm:"type:varchar(200);not null" json:"name"`
	MatchType    string                  `gorm:"type:varchar(10);not null" json:"match_type"` // "1v1" | "2v2"
	Status       string                  `gorm:"type:varchar(20);default:'active'" json:"status"` // active | completed
	AffectsScore bool                    `gorm:"default:true" json:"affects_score"`
	EntryFee     int                     `gorm:"default:0" json:"entry_fee"`
	CreatedAt    time.Time               `json:"created_at"`
	UpdatedAt    time.Time               `json:"updated_at"`
	Participants []TournamentParticipant `gorm:"foreignKey:TournamentID;constraint:OnDelete:CASCADE" json:"participants,omitempty"`
	Matches      []TournamentMatch       `gorm:"foreignKey:TournamentID;constraint:OnDelete:CASCADE" json:"matches,omitempty"`
}

func (Tournament) TableName() string { return "tournaments" }

type TournamentParticipant struct {
	ID                   uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	TournamentID         uuid.UUID `gorm:"type:uuid;not null" json:"tournament_id"`
	UserID               uuid.UUID `gorm:"type:uuid;not null" json:"user_id"`
	TierSnapshot         string    `gorm:"type:varchar(10);default:'normal'" json:"tier_snapshot"`
	HandicapRateSnapshot float64   `gorm:"default:0.0" json:"handicap_rate_snapshot"`
	User                 User      `gorm:"foreignKey:UserID" json:"user,omitempty"`
}

func (TournamentParticipant) TableName() string { return "tournament_participants" }

type TournamentMatch struct {
	ID             uuid.UUID  `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	TournamentID   uuid.UUID  `gorm:"type:uuid;not null" json:"tournament_id"`
	Round          int        `gorm:"not null" json:"round"`
	MatchOrder     int        `gorm:"not null" json:"match_order"`
	Team1Player1ID uuid.UUID  `gorm:"type:uuid;not null" json:"team1_player1_id"`
	Team1Player2ID *uuid.UUID `gorm:"type:uuid" json:"team1_player2_id,omitempty"`
	Team2Player1ID uuid.UUID  `gorm:"type:uuid;not null" json:"team2_player1_id"`
	Team2Player2ID *uuid.UUID `gorm:"type:uuid" json:"team2_player2_id,omitempty"`
	HandicapTeam1  float64    `gorm:"default:0.0" json:"handicap_team1"`
	HandicapTeam2  float64    `gorm:"default:0.0" json:"handicap_team2"`
	Status         string     `gorm:"type:varchar(20);default:'pending'" json:"status"` // pending | completed
	ActualScore1   *int       `json:"actual_score1,omitempty"`
	ActualScore2   *int       `json:"actual_score2,omitempty"`
	EffectiveWinner int       `gorm:"default:0" json:"effective_winner"` // 0=draw/pending, 1=team1, 2=team2
	MatchID        *uuid.UUID `gorm:"type:uuid" json:"match_id,omitempty"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
	// Preloaded relations
	Team1Player1 *User `gorm:"-" json:"team1_player1,omitempty"`
	Team1Player2 *User `gorm:"-" json:"team1_player2,omitempty"`
	Team2Player1 *User `gorm:"-" json:"team2_player1,omitempty"`
	Team2Player2 *User `gorm:"-" json:"team2_player2,omitempty"`
}

func (TournamentMatch) TableName() string { return "tournament_matches" }
