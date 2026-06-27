package model

import "time"

// WcAnalyticsResponse is the full response for GET /api/v1/wc/analytics/world-cup-2026.
type WcAnalyticsResponse struct {
	// From DB — always returned
	MatchStats       WcTournamentMatchStats `json:"match_stats"`
	// From football-data.org — graceful degradation if unavailable
	TopScorers       []WcTournamentScorer   `json:"top_scorers"`
	ScorersUpdatedAt *time.Time             `json:"scorers_updated_at,omitempty"`
	// From openfootball — graceful degradation if unavailable
	GoalTiming        []WcGoalTimingBucket `json:"goal_timing,omitempty"`
	HalfTimeStats     *WcHalfTimeStats     `json:"half_time_stats,omitempty"`
	TeamStats         []WcTeamStat         `json:"team_stats,omitempty"`
	GoalsByGroup      []WcGroupGoals       `json:"goals_by_group,omitempty"`
	TopScoringMatches []WcMatchDetail      `json:"top_scoring_matches,omitempty"`
	VenueStats        []WcVenueStat        `json:"venue_stats,omitempty"`
}

// --- DB-sourced ---

type WcTournamentMatchStats struct {
	TotalMatches        int                    `json:"total_matches"`
	TotalGoals          int                    `json:"total_goals"`
	AvgGoalsPerMatch    float64                `json:"avg_goals_per_match"`
	HomeWins            int                    `json:"home_wins"`
	AwayWins            int                    `json:"away_wins"`
	Draws               int                    `json:"draws"`
	CleanSheets         int                    `json:"clean_sheets"`
	HighestScoringMatch *WcTournamentMatchResult `json:"highest_scoring_match,omitempty"`
	GoalsByStage        []WcStageGoalsStat     `json:"goals_by_stage"`
}

type WcTournamentMatchResult struct {
	HomeTeam   string `json:"home_team"`
	AwayTeam   string `json:"away_team"`
	HomeScore  int    `json:"home_score"`
	AwayScore  int    `json:"away_score"`
	Stage      string `json:"stage"`
	TotalGoals int    `json:"total_goals"`
}

type WcStageGoalsStat struct {
	Stage   string `json:"stage"`
	Matches int    `json:"matches"`
	Goals   int    `json:"goals"`
}

// --- football-data.org sourced ---

type WcTournamentScorer struct {
	Rank          int    `json:"rank"`
	PlayerName    string `json:"player_name"`
	TeamName      string `json:"team_name"`
	TeamCode      string `json:"team_code"`
	TeamCrest     string `json:"team_crest"`
	Goals         int    `json:"goals"`
	Assists       *int   `json:"assists"` // nullable — FD API sometimes omits
	PlayedMatches int    `json:"played_matches"`
}

// --- openfootball sourced ---

type WcGoalTimingBucket struct {
	Label string `json:"label"` // "1-15" | "16-30" | "31-45" | "45+" | "46-60" | "61-75" | "76-90" | "90+"
	Goals int    `json:"goals"`
}

type WcHalfTimeStats struct {
	FirstHalfGoals  int `json:"first_half_goals"`
	SecondHalfGoals int `json:"second_half_goals"`
	OwnGoals        int `json:"own_goals"`
	PenaltyGoals    int `json:"penalty_goals"`
	Comebacks       int `json:"comebacks"`  // trailed at HT, won at FT
	HeldLead        int `json:"held_lead"`  // led at HT, won at FT
}

type WcTeamStat struct {
	TeamName     string `json:"team_name"`
	GoalsFor     int    `json:"goals_for"`
	GoalsAgainst int    `json:"goals_against"`
	Matches      int    `json:"matches"`
}

type WcGroupGoals struct {
	Group   string `json:"group"`
	Matches int    `json:"matches"`
	Goals   int    `json:"goals"`
}

type WcMatchDetail struct {
	HomeTeam   string `json:"home_team"`
	AwayTeam   string `json:"away_team"`
	HomeScore  int    `json:"home_score"`
	AwayScore  int    `json:"away_score"`
	TotalGoals int    `json:"total_goals"`
	Group      string `json:"group,omitempty"`
	Round      string `json:"round"`
	Date       string `json:"date"`
	Venue      string `json:"venue"`
}

type WcVenueStat struct {
	Venue   string `json:"venue"`
	Matches int    `json:"matches"`
	Goals   int    `json:"goals"`
}
