package model

import (
	"time"

	"github.com/google/uuid"
)

// Highlight sections
const (
	SectionTrending    = "trending"
	SectionDailyRecap  = "daily_recap"
	SectionCompetitive = "competitive"
	SectionSocial      = "social"
)

// Highlight types — Category 1: Streaks
const (
	TypeStreakWin       = "streak_win"
	TypeStreakLose      = "streak_lose"
	TypeStreakUnbeaten  = "streak_unbeaten"
	TypeStreakBrokenWin = "streak_broken_win"
	TypeStreakBrokenLose = "streak_broken_lose"
)

// Highlight types — Category 2: Rank / Point Movement
const (
	TypePointsGainedToday   = "points_gained_today"
	TypePointsLostToday     = "points_lost_today"
	TypeRankClimbed         = "rank_climbed"
	TypeRankDropped         = "rank_dropped"
	TypeFastestClimberToday = "fastest_climber_today"
	TypeBiggestCollapse     = "biggest_collapse"
)

// Highlight types — Category 3: Recent Form
const (
	TypeFormHot    = "form_hot"
	TypeFormCold   = "form_cold"
	TypeFormStable = "form_stable"
)

// Highlight types — Category 4: Activity
const (
	TypeMostActiveToday  = "most_active_today"
	TypeMarathon         = "marathon"
	TypeActiveStreakDays = "active_streak_days"
)

// Highlight types — Category 5: Hot / Cold (uses streaks + form data)
const (
	TypeHotPlayer  = "hot_player"
	TypeColdPlayer = "cold_player"
)

// Highlight types — Category 6: Fast Climb / Collapse
const (
	TypeFastClimbHour    = "fast_climb_hour"
	TypeFastCollapseHour = "fast_collapse_hour"
)

// Highlight types — Category 7: Milestones
const (
	TypeMilestoneWins    = "milestone_wins"
	TypeMilestonePoints  = "milestone_points"
	TypeMilestoneMatches = "milestone_matches"
	TypeMilestoneTop10   = "milestone_top10"
)

// Highlight types — Category 8: Social / Fun
const (
	TypeSocialCooking  = "social_cooking"
	TypeSocialTilt     = "social_tilt"
	TypeSocialTryhard  = "social_tryhard"
	TypeSocialGrinder  = "social_grinder"
	TypeSocialSneaky   = "social_sneaky"
)

// Milestone thresholds
var MilestoneWins = []int{100, 500, 1000}
var MilestonePoints = []int{1000, 2000, 5000}
var MilestoneMatches = []int{500, 1000}

// Rule thresholds — all named constants for easy tuning
const (
	HotPlayerMinWinRate    = 0.80
	HotPlayerMinMatchesToday = 5
	ColdPlayerMinLoseStreak  = 5

	StreakHighPriority   = 10 // streak length for priority 95
	StreakMedPriority    = 5  // streak length for priority 80
	StreakLowPriority    = 3  // streak length for priority 60

	FormHotMinWinRate    = 0.80 // 8/10
	FormColdMaxWinRate   = 0.40 // 4/10 or worse

	MarathonMinMatches    = 20
	TryhardMinMatches     = 10
	GrinderMinMatches     = 15
	SneakyMaxMatches      = 5
	SneakyMinRankClimb    = 2

	PointsHighGain = 100
	PointsMedGain  = 50

	StreakBreakerMinLength = 3 // minimum streak to be considered "broken"
)

// Highlight is a single computed highlight card.
// value semantics: win rates as 0.80, counts/points as float64(n).
type Highlight struct {
	PlayerID    uuid.UUID `json:"player_id"`
	PlayerName  string    `json:"player_name"`
	SecondName  string    `json:"second_name,omitempty"` // for streak_broken types (victim)
	Type        string    `json:"type"`
	Section     string    `json:"section"`
	Emoji       string    `json:"emoji"`
	Message     string    `json:"message"`
	Value       float64   `json:"value"`
	Priority    int       `json:"priority"`
}

// HighlightsResponse is the top-level API response, pre-grouped by section.
type HighlightsResponse struct {
	Trending    []Highlight `json:"trending"`
	DailyRecap  []Highlight `json:"daily_recap"`
	Competitive []Highlight `json:"competitive"`
	Social      []Highlight `json:"social"`
	GeneratedAt time.Time   `json:"generated_at"`
}
