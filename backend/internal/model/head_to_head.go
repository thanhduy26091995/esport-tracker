package model

import (
	"time"

	"github.com/google/uuid"
)

// H2HRow is a lean row from the opposing-side head-to-head query.
// It carries only what the aggregates (total/wins/win-rate/form/streak) need.
type H2HRow struct {
	MatchID    uuid.UUID `gorm:"column:match_id"`
	MatchType  string    `gorm:"column:match_type"`
	WinnerTeam int       `gorm:"column:winner_team"`
	MatchDate  time.Time `gorm:"column:match_date"`
	P1Team     int       `gorm:"column:p1_team"`
}

// H2HPlayer is the public player summary shown in the head-to-head header.
type H2HPlayer struct {
	ID           uuid.UUID `json:"id"`
	Name         string    `json:"name"`
	AvatarURL    *string   `json:"avatar_url"`
	FavoriteClub *string   `json:"favorite_club"`
	Tier         string    `json:"tier"`
	IsActive     bool      `json:"is_active"` // false = soft-deleted; UI tags "đã nghỉ"
}

// H2HParticipant is one player's slot in a recent match lineup.
type H2HParticipant struct {
	UserID    uuid.UUID `json:"user_id"`
	Name      string    `json:"name"`
	AvatarURL *string   `json:"avatar_url"`
	Team      int       `json:"team"` // 1 or 2
}

// H2HMatch is one recent head-to-head encounter, with the full lineup for "A+C vs B+D".
type H2HMatch struct {
	MatchID      uuid.UUID        `json:"match_id"`
	MatchType    string           `json:"match_type"`
	MatchDate    time.Time        `json:"match_date"`
	WinnerTeam   int              `json:"winner_team"`
	Player1Team  int              `json:"player1_team"`
	Player1Won   bool             `json:"player1_won"`
	Participants []H2HParticipant `json:"participants"`
}

// H2HStreak describes the current run of consecutive wins for one player.
type H2HStreak struct {
	PlayerID *uuid.UUID `json:"player_id"` // whose current streak (nil when there are no matches)
	Count    int        `json:"count"`     // consecutive wins in the most-recent run
}

// HeadToHeadResponse is the full head-to-head payload, oriented to Player 1.
type HeadToHeadResponse struct {
	Player1        H2HPlayer  `json:"player1"`
	Player2        H2HPlayer  `json:"player2"`
	TotalMatches   int        `json:"total_matches"`
	Player1Wins    int        `json:"player1_wins"`
	Player2Wins    int        `json:"player2_wins"`
	Player1WinRate float64    `json:"player1_win_rate"` // 0..1, 0 when total = 0
	Player2WinRate float64    `json:"player2_win_rate"`
	CurrentStreak  H2HStreak  `json:"current_streak"`
	Form           []string   `json:"form"`           // "W"/"L" from P1 POV, most-recent first, max 10
	RecentMatches  []H2HMatch `json:"recent_matches"` // most-recent first, max 10
}
