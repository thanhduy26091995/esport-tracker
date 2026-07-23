package model

import (
	"time"

	"github.com/google/uuid"
)

const (
	WcCustomBetStatusOpen    = "open"
	WcCustomBetStatusClosed  = "closed"
	WcCustomBetStatusSettled = "settled"
	WcCustomBetStatusVoid    = "void"
)

const (
	WcCustomBetEntryStatusPending = "pending"
	WcCustomBetEntryStatusWon     = "won"
	WcCustomBetEntryStatusLost    = "lost"
	WcCustomBetEntryStatusVoid    = "void"
)

type WcCustomBet struct {
	ID             uuid.UUID  `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	TournamentType string     `gorm:"type:varchar(20);not null;default:'world_cup';index:idx_wc_custom_bets_tournament_type" json:"tournament_type"`
	MatchID        uuid.UUID  `gorm:"type:uuid;not null;index" json:"match_id"`
	Title          string     `gorm:"type:varchar(300);not null" json:"title"`
	Line           *float64   `gorm:"type:numeric(6,2)" json:"line,omitempty"`
	Status         string     `gorm:"type:varchar(20);not null;default:'open'" json:"status"`
	CreatedBy      *uuid.UUID `gorm:"type:uuid" json:"created_by,omitempty"`
	CreatedAt      time.Time  `json:"created_at"`
	SettledAt      *time.Time `json:"settled_at,omitempty"`
	SettledBy      *uuid.UUID `gorm:"type:uuid" json:"settled_by,omitempty"`
}

func (WcCustomBet) TableName() string { return "wc_custom_bets" }

type WcCustomBetOption struct {
	ID           uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	CustomBetID  uuid.UUID `gorm:"type:uuid;not null;index" json:"custom_bet_id"`
	Label        string    `gorm:"type:varchar(200);not null" json:"label"`
	Odds         float64   `gorm:"type:numeric(6,2);not null" json:"odds"`
	IsWinner     bool      `gorm:"not null;default:false" json:"is_winner"`
	DisplayOrder int       `gorm:"not null;default:0" json:"display_order"`
}

func (WcCustomBetOption) TableName() string { return "wc_custom_bet_options" }

type WcCustomBetEntry struct {
	ID            uuid.UUID  `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	CustomBetID   uuid.UUID  `gorm:"type:uuid;not null;index" json:"custom_bet_id"`
	OptionID      uuid.UUID  `gorm:"type:uuid;not null" json:"option_id"`
	WcUserID      uuid.UUID  `gorm:"type:uuid;not null" json:"wc_user_id"`
	Stake         int        `gorm:"not null" json:"stake"`
	OriginalStake *int       `json:"original_stake,omitempty"`
	OddsSnapshot  float64    `gorm:"type:numeric(6,2);not null" json:"odds_snapshot"`
	Payout        *float64   `gorm:"type:numeric(10,2)" json:"payout,omitempty"`
	Status        string     `gorm:"type:varchar(20);not null;default:'pending'" json:"status"`
	CancelledAt   *time.Time `gorm:"type:timestamptz" json:"cancelled_at,omitempty"`
	CancelPenalty *float64   `gorm:"type:numeric(10,2)" json:"cancel_penalty,omitempty"`
	CreatedAt     time.Time  `json:"created_at"`
}

func (WcCustomBetEntry) TableName() string { return "wc_custom_bet_entries" }

type WcCustomBetEntryPublic struct {
	ID           uuid.UUID `json:"id"`
	WcUserID     uuid.UUID `json:"wc_user_id"`
	OptionID     uuid.UUID `json:"option_id"`
	OptionLabel  string    `json:"option_label"`
	Name         string    `json:"name"`
	AvatarURL    *string   `json:"avatar_url"`
	Stake        int       `json:"stake"`
	OddsSnapshot float64   `json:"odds_snapshot"`
	Status       string    `json:"status"`
	Payout       *float64  `json:"payout,omitempty"`
	CreatedAt    time.Time `json:"created_at"`
}

type WcCustomBetWithOptions struct {
	WcCustomBet
	Options    []WcCustomBetOption      `json:"options"`
	MyEntry    *WcCustomBetEntry        `json:"my_entry,omitempty"`
	EntryCount int                      `json:"entry_count"`
	Entries    []WcCustomBetEntryPublic `json:"entries"`
}

type WcCustomBetEntryHistory struct {
	WcCustomBetEntry
	MatchID     uuid.UUID `json:"match_id"`
	BetTitle    string    `json:"bet_title"`
	BetLine     *float64  `json:"bet_line,omitempty"`
	OptionLabel string    `json:"option_label"`
	HomeTeam    string    `json:"home_team"`
	AwayTeam    string    `json:"away_team"`
	MatchDate   time.Time `json:"match_date"`
}
