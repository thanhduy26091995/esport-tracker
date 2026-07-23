package model

import (
	"time"

	"github.com/google/uuid"
)

// WcChampionTeam is a team that users can predict as champion.
type WcChampionTeam struct {
	ID             uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	TournamentType string    `gorm:"type:varchar(20);not null;default:'world_cup';uniqueIndex:idx_wc_champion_team_name_tournament;index:idx_wc_champion_teams_tournament_type" json:"tournament_type"`
	Name           string    `gorm:"type:varchar(100);not null;uniqueIndex:idx_wc_champion_team_name_tournament" json:"name"`
	Code           string    `gorm:"type:varchar(10);not null" json:"code"`
	FlagEmoji      string    `gorm:"type:varchar(10)" json:"flag_emoji"`
	Odds           float64   `gorm:"type:numeric(6,2);not null" json:"odds"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

func (WcChampionTeam) TableName() string { return "wc_champion_teams" }

// WcChampionConfig is the per-tournament state for champion prediction.
type WcChampionConfig struct {
	ID             int        `gorm:"primaryKey;autoIncrement:false" json:"id"`
	TournamentType string     `gorm:"type:varchar(20);not null;default:'world_cup';uniqueIndex" json:"tournament_type"`
	IsOpen         bool       `gorm:"default:false" json:"is_open"`
	WinnerID       *uuid.UUID `gorm:"type:uuid" json:"winner_id,omitempty"`
	SettledAt      *time.Time `json:"settled_at,omitempty"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
}

func (WcChampionConfig) TableName() string { return "wc_champion_config" }

// WcChampionPrediction stores champion predictions. Users may pick multiple teams (one per team).
type WcChampionPrediction struct {
	ID             uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	TournamentType string    `gorm:"type:varchar(20);not null;default:'world_cup';uniqueIndex:idx_wc_champion_pred_user_team;index" json:"tournament_type"`
	WcUserID       uuid.UUID `gorm:"type:uuid;not null;uniqueIndex:idx_wc_champion_pred_user_team" json:"wc_user_id"`
	TeamID         uuid.UUID `gorm:"type:uuid;not null;uniqueIndex:idx_wc_champion_pred_user_team" json:"team_id"`
	Points         int       `gorm:"not null" json:"points"`
	OddsSnapshot   float64   `gorm:"type:numeric(6,2);not null" json:"odds_snapshot"`
	Result         *string   `gorm:"type:varchar(20)" json:"result,omitempty"`
	PointsEarned   *int      `json:"points_earned,omitempty"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

func (WcChampionPrediction) TableName() string { return "wc_champion_predictions" }

// --- Response types ---

type WcChampionPredictionPublic struct {
	UserName        string    `json:"user_name"`
	WcUserID        uuid.UUID `json:"wc_user_id"`
	TeamName        string    `json:"team_name"`
	TeamCode        string    `json:"team_code"`
	FlagEmoji       string    `json:"flag_emoji"`
	Points          int       `json:"points"`
	OddsSnapshot    float64   `json:"odds_snapshot"`
	PayoutIfCorrect int       `json:"payout_if_correct"`
	Result          *string   `json:"result,omitempty"`
}

type WcChampionPredictionMine struct {
	ID              uuid.UUID  `json:"id"`
	TeamID          uuid.UUID  `json:"team_id"`
	TeamName        string     `json:"team_name"`
	TeamCode        string     `json:"team_code"`
	FlagEmoji       string     `json:"flag_emoji"`
	Points          int        `json:"points"`
	OddsSnapshot    float64    `json:"odds_snapshot"`
	PayoutIfCorrect int        `json:"payout_if_correct"`
	Result          *string    `json:"result,omitempty"`
	PointsEarned    *int       `json:"points_earned,omitempty"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
}

type WcChampionSettleResult struct {
	Winner             string `json:"winner"`
	SettledCount       int    `json:"settled_count"`
	SettledUserCount   int    `json:"settled_user_count"`
	CorrectCount       int    `json:"correct_count"`
	TotalPointsAwarded int    `json:"total_points_awarded"`
}

type WcChampionConfigPublic struct {
	IsOpen     bool                 `json:"is_open"`
	SettledAt  *time.Time           `json:"settled_at,omitempty"`
	WinnerTeam *WcChampionTeam      `json:"winner_team,omitempty"`
}
