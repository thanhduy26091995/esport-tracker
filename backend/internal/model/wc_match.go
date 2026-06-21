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

// Prediction type constants
const (
	WcPredictionTypeHandicap   = "handicap"
	WcPredictionTypeExactScore = "exact_score"
	WcPredictionTypeOverUnder  = "over_under"
)

// Over/Under choice constants
const (
	WcChoiceOver  = "over"
	WcChoiceUnder = "under"
)

// Prediction result constants
const (
	WcResultCorrect   = "correct"
	WcResultIncorrect = "incorrect"
	WcResultVoid      = "void"
)

// Bet type constants
const (
	WcBetTypeHandicap   = "handicap"
	WcBetTypeExactScore = "exact_score"
	WcBetTypeOverUnder  = "over_under"
)

// Bet result constants
const (
	WcResultWin      = "win"
	WcResultLose     = "lose"
	WcResultPush     = "push"
	WcResultWinHalf  = "win_half"  // split handicap: win one sub-bet, push the other
	WcResultLoseHalf = "lose_half" // split handicap: lose one sub-bet, push the other
)

// Settlement direction constants
const (
	WcDirectionPay     = "pay"
	WcDirectionCollect = "collect"
	WcDirectionEven    = "even"
)

// Settlement detail status
const (
	WcSettlementStatusPending = "pending"
	WcSettlementStatusDone    = "done"
)

type WcMatch struct {
	ID                  uuid.UUID  `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	ExternalID          string     `gorm:"type:varchar(64);uniqueIndex" json:"external_id"`
	HomeTeam            string     `gorm:"type:varchar(100);not null" json:"home_team"`
	AwayTeam            string     `gorm:"type:varchar(100);not null" json:"away_team"`
	HomeTeamCode        string     `gorm:"type:char(3)" json:"home_team_code"`
	AwayTeamCode        string     `gorm:"type:char(3)" json:"away_team_code"`
	MatchDate           time.Time  `gorm:"not null" json:"match_date"`
	GroupName           string     `gorm:"type:varchar(30)" json:"group_name"`
	Stage               string     `gorm:"type:varchar(30);not null;default:'group'" json:"stage"`
	Venue               string     `gorm:"type:varchar(100)" json:"venue"`
	HomeScore           *int       `json:"home_score"`
	AwayScore           *int       `json:"away_score"`
	Status              string     `gorm:"type:varchar(20);not null;default:'scheduled'" json:"status"`
	HandicapTeam        string     `gorm:"type:varchar(5)" json:"handicap_team"`
	HandicapValue       *float64   `gorm:"type:numeric(5,2)" json:"handicap_value"`
	OddsHandicapHome    *float64   `gorm:"type:numeric(5,2)" json:"odds_handicap_home"`
	OddsHandicapAway    *float64   `gorm:"type:numeric(5,2)" json:"odds_handicap_away"`
	PredictionsOpen     bool       `gorm:"not null;default:false" json:"predictions_open"`
	PredictionsLockedAt *time.Time `json:"predictions_locked_at"`
	BetsLockedAt        *time.Time `json:"bets_locked_at"`
	SettledAt           *time.Time `json:"settled_at"`
	// StatsAPI integration fields
	StatsapiFixtureID *string    `gorm:"type:varchar(64);uniqueIndex" json:"statsapi_fixture_id"`
	OULine            *float64   `gorm:"type:numeric(4,1)" json:"ou_line"`
	OddsOver          *float64   `gorm:"type:numeric(5,2)" json:"odds_over"`
	OddsUnder         *float64   `gorm:"type:numeric(5,2)" json:"odds_under"`
	OUSyncedAt        *time.Time `json:"ou_synced_at"`
	OddsSyncedAt      *time.Time `json:"odds_synced_at"`
	PoissonSyncedAt   *time.Time `json:"poisson_synced_at"`
	CreatedAt         time.Time  `json:"created_at"`
	UpdatedAt         time.Time  `json:"updated_at"`
}

func (WcMatch) TableName() string {
	return "wc_matches"
}

type WcScoreMultiplier struct {
	ID        uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	MatchID   uuid.UUID `gorm:"type:uuid;not null;uniqueIndex:idx_score_multiplier_scoreline" json:"match_id"`
	HomeScore int       `gorm:"not null;uniqueIndex:idx_score_multiplier_scoreline" json:"home_score"`
	AwayScore int       `gorm:"not null;uniqueIndex:idx_score_multiplier_scoreline" json:"away_score"`
	Multiplier float64  `gorm:"type:numeric(5,2);not null" json:"multiplier"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (WcScoreMultiplier) TableName() string {
	return "wc_score_multipliers"
}

type WcWallet struct {
	ID        uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	WcUserID  uuid.UUID `gorm:"type:uuid;not null;uniqueIndex" json:"wc_user_id"`
	Balance   float64   `gorm:"type:numeric(10,2);default:0" json:"balance"`
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
	Delta         float64   `gorm:"type:numeric(10,2);not null" json:"delta"`
	BalanceBefore float64   `gorm:"type:numeric(10,2);not null" json:"balance_before"`
	BalanceAfter  float64   `gorm:"type:numeric(10,2);not null" json:"balance_after"`
	Note          string    `gorm:"type:varchar(255)" json:"note"`
	CreatedAt     time.Time `json:"created_at"`
}

func (WcWalletLog) TableName() string {
	return "wc_wallet_logs"
}

type WcPrediction struct {
	ID uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`

	WcUserID       uuid.UUID `gorm:"type:uuid;not null;index;uniqueIndex:idx_prediction_hc_dedup;uniqueIndex:idx_prediction_es_dedup" json:"wc_user_id"`
	MatchID        uuid.UUID `gorm:"type:uuid;not null;index;uniqueIndex:idx_prediction_hc_dedup;uniqueIndex:idx_prediction_es_dedup" json:"match_id"`
	PredictionType string    `gorm:"type:varchar(15);not null;uniqueIndex:idx_prediction_hc_dedup" json:"prediction_type"`
	Points         int       `gorm:"not null" json:"points"`
	MultiplierSnapshot float64 `gorm:"type:numeric(5,2);not null" json:"multiplier_snapshot"`

	// Handicap prediction fields (nullable for exact_score predictions)
	PredictionChoice     *string  `gorm:"type:varchar(5);uniqueIndex:idx_prediction_hc_dedup" json:"prediction_choice,omitempty"`
	HandicapSnapshot     *float64 `gorm:"type:numeric(4,1)" json:"handicap_snapshot,omitempty"`
	HandicapTeamSnapshot *string  `gorm:"type:varchar(5)" json:"handicap_team_snapshot,omitempty"`

	// Exact score prediction fields (nullable for handicap predictions)
	PredictedHomeScore *int `gorm:"uniqueIndex:idx_prediction_es_dedup" json:"predicted_home_score,omitempty"`
	PredictedAwayScore *int `gorm:"uniqueIndex:idx_prediction_es_dedup" json:"predicted_away_score,omitempty"`

	// Result
	Result       *string  `gorm:"type:varchar(10)" json:"result,omitempty"`
	PointsEarned *float64 `gorm:"type:numeric(10,2)" json:"points_earned,omitempty"`
	CreatedAt    time.Time `json:"created_at"`
}

func (WcPrediction) TableName() string {
	return "wc_predictions"
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
	FinalBalance float64    `gorm:"type:numeric(10,2);not null" json:"final_balance"`
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

// WcScoreOdds stores money odds for exact-score bets (separate from prediction multipliers).
type WcScoreOdds struct {
	ID        uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	MatchID   uuid.UUID `gorm:"type:uuid;not null;uniqueIndex:idx_score_odds_scoreline" json:"match_id"`
	HomeScore int       `gorm:"not null;uniqueIndex:idx_score_odds_scoreline" json:"home_score"`
	AwayScore int       `gorm:"not null;uniqueIndex:idx_score_odds_scoreline" json:"away_score"`
	Odds      float64   `gorm:"type:numeric(5,2);not null" json:"odds"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (WcScoreOdds) TableName() string { return "wc_score_odds" }

// WcBet stores a real-money bet placed by a user on a match.
type WcBet struct {
	ID                   uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	WcUserID             uuid.UUID `gorm:"type:uuid;not null;index;uniqueIndex:idx_bet_hc_dedup;uniqueIndex:idx_bet_es_dedup" json:"wc_user_id"`
	MatchID              uuid.UUID `gorm:"type:uuid;not null;index;uniqueIndex:idx_bet_hc_dedup;uniqueIndex:idx_bet_es_dedup" json:"match_id"`
	BetType              string    `gorm:"type:varchar(15);not null;uniqueIndex:idx_bet_hc_dedup" json:"bet_type"`
	BetChoice            *string   `gorm:"type:varchar(5);uniqueIndex:idx_bet_hc_dedup" json:"bet_choice,omitempty"`
	Stake                int       `gorm:"not null" json:"stake"`
	OddsSnapshot         float64   `gorm:"type:numeric(5,2);not null" json:"odds_snapshot"`
	HandicapSnapshot     *float64  `gorm:"type:numeric(5,2)" json:"handicap_snapshot,omitempty"`
	HandicapTeamSnapshot *string   `gorm:"type:varchar(5)" json:"handicap_team_snapshot,omitempty"`
	PredictedHomeScore   *int      `gorm:"uniqueIndex:idx_bet_es_dedup" json:"predicted_home_score,omitempty"`
	PredictedAwayScore   *int      `gorm:"uniqueIndex:idx_bet_es_dedup" json:"predicted_away_score,omitempty"`
	Result               *string   `gorm:"type:varchar(10)" json:"result,omitempty"`
	Payout               *float64  `gorm:"type:numeric(10,2)" json:"payout,omitempty"`
	CreatedAt            time.Time `json:"created_at"`
	UpdatedAt            time.Time `json:"updated_at"`
}

func (WcBet) TableName() string { return "wc_bets" }

// WcBetWithMatch is used by ListBets — includes match context and live lock status.
type WcBetWithMatch struct {
	WcBet
	HomeTeam    string     `json:"home_team"`
	AwayTeam    string     `json:"away_team"`
	MatchDate   time.Time  `json:"match_date"`
	MatchStatus string     `json:"match_status"`
	BettingOpen bool       `json:"betting_open"` // computed in SQL, not a DB column
	BetsLockedAt *time.Time `json:"bets_locked_at"`
}

// WcBetPublic is used by GetMatchBets — shows bets on a match to all users.
type WcBetPublic struct {
	ID                 uuid.UUID `json:"id"`
	WcUserID           uuid.UUID `json:"wc_user_id"`
	Name               string    `json:"name"`
	AvatarURL          *string   `json:"avatar_url"`
	BetType            string    `json:"bet_type"`
	BetChoice          *string   `json:"bet_choice,omitempty"`
	Stake              int       `json:"stake"`
	OddsSnapshot       float64   `json:"odds_snapshot"`
	PredictedHomeScore *int      `json:"predicted_home_score,omitempty"`
	PredictedAwayScore *int      `json:"predicted_away_score,omitempty"`
	Result             *string   `json:"result,omitempty"`
	Payout             *float64  `json:"payout,omitempty"`
	CreatedAt          time.Time `json:"created_at"`
}

// WcMatchWithOdds is returned by GET /matches/:id.
type WcMatchWithOdds struct {
	WcMatch
	ScoreMultipliers []WcScoreMultiplier `json:"score_multipliers"`
}

// WcLeaderboardEntry is used by GET /leaderboard.
type WcLeaderboardEntry struct {
	Rank             int       `json:"rank"`
	WcUserID         uuid.UUID `json:"wc_user_id"`
	Name             string    `json:"name"`
	AvatarURL        *string   `json:"avatar_url"`
	NetPoints        float64   `json:"net_points"`
	TotalPredictions int       `json:"total_predictions"`
	Correct          int       `json:"correct"`
	WinHalf          int       `json:"win_half"`
	LoseHalf         int       `json:"lose_half"`
	Incorrect        int       `json:"incorrect"`
}

// WcWalletWithUser is used by GetAllWallets (admin panel + settlement preview).
type WcWalletWithUser struct {
	WcWallet
	Name string `json:"name"`
}

// WcPredictionWithMatch is used by ListPredictions (user prediction history).
type WcPredictionWithMatch struct {
	WcPrediction
	HomeTeam            string     `json:"home_team"`
	AwayTeam            string     `json:"away_team"`
	MatchDate           time.Time  `json:"match_date"`
	MatchStatus         string     `json:"match_status"`
	PredictionsOpen     bool       `json:"predictions_open"`
	PredictionsLockedAt *time.Time `json:"predictions_locked_at"`
}

// WcPredictionPublic is used by ListPredictionsForMatchPublic — visible to everyone.
type WcPredictionPublic struct {
	ID                 uuid.UUID `json:"id"`
	WcUserID           uuid.UUID `json:"wc_user_id"`
	Name               string    `json:"name"`
	AvatarURL          *string   `json:"avatar_url"`
	PredictionType     string    `json:"prediction_type"`
	PredictionChoice   *string   `json:"prediction_choice,omitempty"`
	PredictedHomeScore *int      `json:"predicted_home_score,omitempty"`
	PredictedAwayScore *int      `json:"predicted_away_score,omitempty"`
	Points             int       `json:"points"`
	MultiplierSnapshot float64   `json:"multiplier_snapshot"`
	Result             *string   `json:"result,omitempty"`
	PointsEarned       *float64  `json:"points_earned,omitempty"`
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
	Name string `json:"user_name"`
}

// WcSyncLog records each sync operation (manual or cron).
type WcSyncLog struct {
	ID             uuid.UUID  `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	Trigger        string     `gorm:"type:varchar(20);not null" json:"trigger"`
	SyncType       string     `gorm:"type:varchar(20);not null" json:"sync_type"`
	TriggeredBy    *uuid.UUID `gorm:"type:uuid" json:"triggered_by"`
	MatchesUpdated int        `gorm:"not null;default:0" json:"matches_updated"`
	MatchesFailed  int        `gorm:"not null;default:0" json:"matches_failed"`
	ErrorDetail    *string    `gorm:"type:text" json:"error_detail"`
	CreatedAt      time.Time  `json:"created_at"`
}

func (WcSyncLog) TableName() string { return "wc_sync_logs" }

// HousePnLResponse is returned by GET /admin/house-pnl.
type HousePnLResponse struct {
	TotalStakeSettled  float64        `json:"total_stake_settled"`
	TotalPayoutSettled float64        `json:"total_payout_settled"`
	HouseProfit        float64        `json:"house_profit"`
	TotalStakeVoid     float64        `json:"total_stake_void"`
	TotalStakePending  float64        `json:"total_stake_pending"`
	PendingBetCount    int            `json:"pending_bet_count"`
	SettledBetCount    int            `json:"settled_bet_count"`
	MatchBreakdown     []HousePnLMatch `json:"match_breakdown"`
	GeneratedAt        string         `json:"generated_at"`
}

type HousePnLMatch struct {
	MatchID  string  `json:"match_id"`
	HomeTeam string  `json:"home_team"`
	AwayTeam string  `json:"away_team"`
	MatchDate string `json:"match_date"`
	Stage    string  `json:"stage"`
	Stake    float64 `json:"stake"`
	Payout   float64 `json:"payout"`
	Profit   float64 `json:"profit"`
	BetCount int     `json:"bet_count"`
}

// WcSettlementPreviewRow is returned by GET /admin/settlements/preview.
type WcSettlementPreviewRow struct {
	WcUserID  uuid.UUID `json:"wc_user_id"`
	Name      string    `json:"user_name"`
	Balance   float64   `json:"balance"`
	Direction string    `json:"direction"`
	Amount    float64   `json:"amount"`
}

// FinalizePreviewRow is one prediction row in the preview response.
type FinalizePreviewRow struct {
	WcUserID        uuid.UUID `json:"wc_user_id"`
	UserName        string    `json:"user_name"`
	PredictionType  string    `json:"prediction_type"`  // handicap | exact_score | over_under
	Points          int       `json:"points"`           // stake
	Multiplier      float64   `json:"multiplier"`       // MultiplierSnapshot
	NewResult       string    `json:"new_result"`       // correct | incorrect | void | win_half | lose_half
	NewPointsEarned float64   `json:"new_points_earned"`
	NetDelta        float64   `json:"net_delta"` // new_points_earned - float64(points)
}

// FinalizePreviewMatch is the per-match section of a preview.
type FinalizePreviewMatch struct {
	MatchID        string               `json:"match_id"`
	HomeTeam       string               `json:"home_team"`
	AwayTeam       string               `json:"away_team"`
	HomeScore      int                  `json:"home_score"`
	AwayScore      int                  `json:"away_score"`
	Stage          string               `json:"stage"`
	AlreadySettled bool                 `json:"already_settled"`
	Predictions    []FinalizePreviewRow `json:"predictions"`
}

// FinalizePreviewHouse is the admin summary section of a preview.
type FinalizePreviewHouse struct {
	TotalStaked     float64 `json:"total_staked"`
	TotalPaidOut    float64 `json:"total_paid_out"`
	HouseNet        float64 `json:"house_net"`
	PredictionCount int     `json:"prediction_count"`
	MatchCount      int     `json:"match_count"`
}

// FinalizePreviewResult is returned by all preview endpoints.
type FinalizePreviewResult struct {
	Matches      []FinalizePreviewMatch `json:"matches"`
	HouseSummary FinalizePreviewHouse   `json:"house_summary"`
}
