package model

import (
	"time"

	"github.com/google/uuid"
)

// Match status constants
const (
	WcStatusScheduled = "scheduled"
	WcStatusLive      = "live"
	WcStatusCompleted = "completed"
	WcStatusCancelled = "cancelled"
)

// Stage constants
const (
	WcStageGroup      = "group"
	WcStageR32        = "r32"
	WcStageR16        = "r16"
	WcStageQF         = "qf"
	WcStageSF         = "sf"
	WcStageFinal      = "final"
	WcStageThirdPlace = "third_place"
)

// Handicap team constants
const (
	WcTeamHome = "home"
	WcTeamAway = "away"
)

// Bet type constants
const (
	WcBetTypeHandicap   = "handicap"
	WcBetTypeExactScore = "exact_score"
)

// Bet result constants
const (
	WcResultWin  = "win"
	WcResultLose = "lose"
	WcResultPush = "push"
)

// Settlement direction constants
const (
	WcDirectionPay     = "pay"     // admin pays user (user won overall)
	WcDirectionCollect = "collect" // admin collects from user (user lost overall)
	WcDirectionEven    = "even"
)

// Settlement detail status
const (
	WcSettlementStatusPending = "pending"
	WcSettlementStatusDone    = "done"
)

type WcMatch struct {
	ID               uuid.UUID  `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	ExternalID       string     `gorm:"type:varchar(64);uniqueIndex" json:"external_id"`
	HomeTeam         string     `gorm:"type:varchar(100);not null" json:"home_team"`
	AwayTeam         string     `gorm:"type:varchar(100);not null" json:"away_team"`
	HomeTeamCode     string     `gorm:"type:char(3)" json:"home_team_code"`
	AwayTeamCode     string     `gorm:"type:char(3)" json:"away_team_code"`
	MatchDate        time.Time  `gorm:"not null" json:"match_date"`
	GroupName        string     `gorm:"type:varchar(30)" json:"group_name"`
	Stage            string     `gorm:"type:varchar(30);not null;default:'group'" json:"stage"`
	Venue            string     `gorm:"type:varchar(100)" json:"venue"`
	HomeScore        *int       `json:"home_score"`
	AwayScore        *int       `json:"away_score"`
	Status           string     `gorm:"type:varchar(20);not null;default:'scheduled'" json:"status"`
	HandicapTeam     string     `gorm:"type:varchar(5)" json:"handicap_team"`
	HandicapValue    *float64   `gorm:"type:numeric(4,1)" json:"handicap_value"`
	OddsHandicapHome *float64   `gorm:"type:numeric(5,2)" json:"odds_handicap_home"`
	OddsHandicapAway *float64   `gorm:"type:numeric(5,2)" json:"odds_handicap_away"`
	BetsLockedAt     *time.Time `json:"bets_locked_at"`
	SettledAt        *time.Time `json:"settled_at"`
	CreatedAt        time.Time  `json:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at"`
}

func (WcMatch) TableName() string {
	return "wc_matches"
}

type WcScoreOdds struct {
	ID        uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	MatchID   uuid.UUID `gorm:"type:uuid;not null;uniqueIndex:idx_score_odds_scoreline" json:"match_id"`
	HomeScore int       `gorm:"not null;uniqueIndex:idx_score_odds_scoreline" json:"home_score"`
	AwayScore int       `gorm:"not null;uniqueIndex:idx_score_odds_scoreline" json:"away_score"`
	Odds      float64   `gorm:"type:numeric(5,2);not null" json:"odds"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (WcScoreOdds) TableName() string {
	return "wc_score_odds"
}

type WcWallet struct {
	ID        uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	WcUserID  uuid.UUID `gorm:"type:uuid;not null;uniqueIndex" json:"wc_user_id"`
	Balance   int       `gorm:"default:0" json:"balance"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (WcWallet) TableName() string {
	return "wc_wallets"
}

type WcWalletLog struct {
	ID            uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	WcUserID      uuid.UUID `gorm:"type:uuid;not null;index" json:"wc_user_id"`
	AdminID       uuid.UUID `gorm:"type:uuid;not null" json:"admin_id"`
	Delta         int       `gorm:"not null" json:"delta"`
	BalanceBefore int       `gorm:"not null" json:"balance_before"`
	BalanceAfter  int       `gorm:"not null" json:"balance_after"`
	Note          string    `gorm:"type:varchar(255)" json:"note"`
	CreatedAt     time.Time `json:"created_at"`
}

func (WcWalletLog) TableName() string {
	return "wc_wallet_logs"
}

type WcBet struct {
	ID uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`

	// Both bet types
	WcUserID  uuid.UUID `gorm:"type:uuid;not null;index;uniqueIndex:idx_bet_hc_dedup;uniqueIndex:idx_bet_es_dedup" json:"wc_user_id"`
	MatchID   uuid.UUID `gorm:"type:uuid;not null;index;uniqueIndex:idx_bet_hc_dedup;uniqueIndex:idx_bet_es_dedup" json:"match_id"`
	BetType   string    `gorm:"type:varchar(15);not null;uniqueIndex:idx_bet_hc_dedup" json:"bet_type"`
	Stake     int       `gorm:"not null" json:"stake"`
	OddsSnapshot float64 `gorm:"type:numeric(5,2);not null" json:"odds_snapshot"`

	// Handicap bet fields (nullable for exact_score bets)
	BetChoice            *string  `gorm:"type:varchar(5);uniqueIndex:idx_bet_hc_dedup" json:"bet_choice,omitempty"`
	HandicapSnapshot     *float64 `gorm:"type:numeric(4,1)" json:"handicap_snapshot,omitempty"`
	HandicapTeamSnapshot *string  `gorm:"type:varchar(5)" json:"handicap_team_snapshot,omitempty"`

	// Exact score bet fields (nullable for handicap bets)
	PredictedHomeScore *int `gorm:"uniqueIndex:idx_bet_es_dedup" json:"predicted_home_score,omitempty"`
	PredictedAwayScore *int `gorm:"uniqueIndex:idx_bet_es_dedup" json:"predicted_away_score,omitempty"`

	// Settlement
	Result    *string `gorm:"type:varchar(10)" json:"result,omitempty"`
	Payout    *int    `json:"payout,omitempty"`
	CreatedAt time.Time `json:"created_at"`
}

func (WcBet) TableName() string {
	return "wc_bets"
}

type WcSettlement struct {
	ID        uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	Name      string    `gorm:"type:varchar(100);not null" json:"name"`
	PointRate float64   `gorm:"type:numeric(10,2);not null" json:"point_rate"`
	SettledBy uuid.UUID `gorm:"type:uuid;not null" json:"settled_by"`
	Note      string    `gorm:"type:varchar(255)" json:"note"`
	CreatedAt time.Time `json:"created_at"`
}

func (WcSettlement) TableName() string {
	return "wc_settlements"
}

type WcSettlementDetail struct {
	ID           uuid.UUID  `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	SettlementID uuid.UUID  `gorm:"type:uuid;not null;index;uniqueIndex:idx_settlement_user" json:"settlement_id"`
	WcUserID     uuid.UUID  `gorm:"type:uuid;not null;uniqueIndex:idx_settlement_user" json:"wc_user_id"`
	FinalBalance int        `gorm:"not null" json:"final_balance"`
	Amount       float64    `gorm:"type:numeric(12,2);not null" json:"amount"`
	Direction    string     `gorm:"type:varchar(10);not null" json:"direction"`
	Status       string     `gorm:"type:varchar(20);not null;default:'pending'" json:"status"`
	CompletedAt  *time.Time `json:"completed_at,omitempty"`
	DoneNote     string     `gorm:"type:varchar(255)" json:"done_note"`
	CreatedAt    time.Time  `json:"created_at"`
}

func (WcSettlementDetail) TableName() string {
	return "wc_settlement_details"
}

// WcMatchWithOdds is returned by GET /matches/:id.
type WcMatchWithOdds struct {
	WcMatch
	ScoreOdds []WcScoreOdds `json:"score_odds"`
}

// WcLeaderboardEntry is used by GET /leaderboard.
type WcLeaderboardEntry struct {
	Rank      int       `json:"rank"`
	WcUserID  uuid.UUID `json:"wc_user_id"`
	Name      string    `json:"name"`
	NetProfit int       `json:"net_profit"`
	TotalBets int       `json:"total_bets"`
	Wins      int       `json:"wins"`
}

// WcWalletWithUser is used by GetAllWallets (admin panel + settlement preview).
type WcWalletWithUser struct {
	WcWallet
	Name string `json:"name"`
}

// WcBetWithMatch is used by ListBets (user bet history).
type WcBetWithMatch struct {
	WcBet
	HomeTeam     string     `json:"home_team"`
	AwayTeam     string     `json:"away_team"`
	MatchDate    time.Time  `json:"match_date"`
	MatchStatus  string     `json:"match_status"`
	BetsLockedAt *time.Time `json:"bets_locked_at"`
}

// WcBetPublic is used by ListBetsForMatchPublic — visible to everyone.
type WcBetPublic struct {
	ID                 uuid.UUID `json:"id"`
	WcUserID           uuid.UUID `json:"wc_user_id"`
	Name               string    `json:"name"`
	BetType            string    `json:"bet_type"`
	BetChoice          *string   `json:"bet_choice,omitempty"`
	PredictedHomeScore *int      `json:"predicted_home_score,omitempty"`
	PredictedAwayScore *int      `json:"predicted_away_score,omitempty"`
	Stake              int       `json:"stake"`
	OddsSnapshot       float64   `json:"odds_snapshot"`
	Result             *string   `json:"result,omitempty"`
	Payout             *int      `json:"payout,omitempty"`
	CreatedAt          time.Time `json:"created_at"`
}

// WcSettlementWithDetails is returned by GET /admin/settlements/:id.
type WcSettlementWithDetails struct {
	WcSettlement
	Details []*WcSettlementDetailWithUser `json:"details"`
}

// WcSettlementDetailWithUser extends WcSettlementDetail with the user's name.
type WcSettlementDetailWithUser struct {
	WcSettlementDetail
	Name string `json:"name"`
}

// WcSettlementPreviewRow is returned by GET /admin/settlements/preview.
type WcSettlementPreviewRow struct {
	WcUserID  uuid.UUID `json:"wc_user_id"`
	Name      string    `json:"name"`
	Balance   int       `json:"balance"`
	Direction string    `json:"direction"`
	Amount    float64   `json:"amount"`
}
