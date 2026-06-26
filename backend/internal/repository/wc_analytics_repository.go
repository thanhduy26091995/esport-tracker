package repository

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type WcAnalyticsRepository struct {
	db *gorm.DB
}

func NewWcAnalyticsRepository(db *gorm.DB) *WcAnalyticsRepository {
	return &WcAnalyticsRepository{db: db}
}

// AnalyticsPeriod holds the date range for filtering.
type AnalyticsPeriod struct {
	From time.Time
	To   time.Time
}

// MatchResultRow is used to compute per-match net result.
type MatchResultRow struct {
	MatchID      string
	MatchDate    time.Time
	TotalPayout  float64
	TotalStake   float64
}

// TimelineRow holds wins/losses for one day bucket.
type TimelineRow struct {
	Period string
	Wins   int
	Losses int
}

// TeamCountRow holds a team name and bet count.
type TeamCountRow struct {
	Team     string
	BetCount int
}

// ScorelineCountRow holds a scoreline string and count.
type ScorelineCountRow struct {
	Scoreline string
	Count     int
}

// BetTypeDistribution holds counts per bet type.
type BetTypeDistribution struct {
	Handicap   int
	ExactScore int
	OverUnder  int
}

// MyBetStatsRow holds raw aggregate stats for one user.
type MyBetStatsRow struct {
	TotalBets          int
	PendingBets        int
	HandicapBets       int
	ExactScoreBets     int
	OverUnderBets      int
	HomePicks          int  // handicap bets where bet_choice='home'
	AwayPicks          int  // handicap bets where bet_choice='away'
	OverPicks          int  // ou bets where bet_choice='over'
	UnderPicks         int  // ou bets where bet_choice='under'
	TotalStake         int
	AvgStake           float64
	LastMinuteBets     int  // placed < 2h before match_date
	ExactScoreWins     int  // settled exact_score bets that won
	ExactScoreSettled  int
}

// PredictorRow holds top predictor data.
type PredictorRow struct {
	WcUserID       string
	Name           string
	AvatarURL      *string
	Wins           int
	Losses         int
	SettledMatches int
}

// AvgGoalsRow holds parsed scoreline goal data.
type AvgGoalsRow struct {
	TotalGoals int
	Count      int
}

// GetMyMatchResults returns all settled non-void match-level results for a user within a period.
// Used for accuracy, streak, and timeline computation.
func (r *WcAnalyticsRepository) GetMyMatchResults(userID uuid.UUID, p AnalyticsPeriod) ([]MatchResultRow, error) {
	var rows []MatchResultRow
	err := r.db.Raw(`
		SELECT
			b.match_id::text AS match_id,
			m.match_date,
			COALESCE(SUM(b.payout), 0) AS total_payout,
			SUM(b.stake)               AS total_stake
		FROM wc_bets b
		JOIN wc_matches m ON m.id = b.match_id
		WHERE b.wc_user_id = ?
		  AND b.result IS NOT NULL
		  AND b.result != 'void'
		  AND m.match_date >= ?
		  AND m.match_date <= ?
		GROUP BY b.match_id, m.match_date
		ORDER BY m.match_date ASC
	`, userID, p.From, p.To).Scan(&rows).Error
	return rows, err
}

// GetMyAccuracyTimeline returns wins/losses bucketed by day within the period.
func (r *WcAnalyticsRepository) GetMyAccuracyTimeline(userID uuid.UUID, p AnalyticsPeriod) ([]TimelineRow, error) {
	var rows []TimelineRow
	err := r.db.Raw(`
		WITH match_results AS (
			SELECT
				DATE(m.match_date) AS period,
				COALESCE(SUM(b.payout), 0) AS total_payout,
				SUM(b.stake)               AS total_stake
			FROM wc_bets b
			JOIN wc_matches m ON m.id = b.match_id
			WHERE b.wc_user_id = ?
			  AND b.result IS NOT NULL
			  AND b.result != 'void'
			  AND m.match_date >= ?
			  AND m.match_date <= ?
			GROUP BY b.match_id, DATE(m.match_date)
		)
		SELECT
			period::text,
			COUNT(*) FILTER (WHERE total_payout > total_stake) AS wins,
			COUNT(*) FILTER (WHERE total_payout <= total_stake) AS losses
		FROM match_results
		GROUP BY period
		ORDER BY period ASC
	`, userID, p.From, p.To).Scan(&rows).Error
	return rows, err
}

// GetMyBetStats returns aggregate betting stats for a user (all-time, not period-filtered).
func (r *WcAnalyticsRepository) GetMyBetStats(userID uuid.UUID) (*MyBetStatsRow, error) {
	var row MyBetStatsRow
	err := r.db.Raw(`
		SELECT
			COUNT(*)                                                             AS total_bets,
			COUNT(*) FILTER (WHERE b.result IS NULL)                            AS pending_bets,
			COUNT(*) FILTER (WHERE b.bet_type = 'handicap')                     AS handicap_bets,
			COUNT(*) FILTER (WHERE b.bet_type = 'exact_score')                  AS exact_score_bets,
			COUNT(*) FILTER (WHERE b.bet_type = 'over_under')                   AS over_under_bets,
			COUNT(*) FILTER (WHERE b.bet_type = 'handicap' AND b.bet_choice = 'home') AS home_picks,
			COUNT(*) FILTER (WHERE b.bet_type = 'handicap' AND b.bet_choice = 'away') AS away_picks,
			COUNT(*) FILTER (WHERE b.bet_type = 'over_under' AND b.bet_choice = 'over')  AS over_picks,
			COUNT(*) FILTER (WHERE b.bet_type = 'over_under' AND b.bet_choice = 'under') AS under_picks,
			COALESCE(SUM(b.stake), 0)                                            AS total_stake,
			COALESCE(AVG(b.stake), 0)                                            AS avg_stake,
			COUNT(*) FILTER (WHERE b.created_at >= m.match_date - INTERVAL '2 hours'
			                   AND b.created_at <= m.match_date)                 AS last_minute_bets,
			COUNT(*) FILTER (WHERE b.bet_type = 'exact_score'
			                   AND b.result IS NOT NULL AND b.result != 'void'
			                   AND COALESCE(b.payout, 0) > b.stake)              AS exact_score_wins,
			COUNT(*) FILTER (WHERE b.bet_type = 'exact_score'
			                   AND b.result IS NOT NULL AND b.result != 'void')  AS exact_score_settled
		FROM wc_bets b
		JOIN wc_matches m ON m.id = b.match_id
		WHERE b.wc_user_id = ?
	`, userID).Scan(&row).Error
	return &row, err
}

// GetMyFavoriteTeams returns top-N teams from handicap bets for a user.
func (r *WcAnalyticsRepository) GetMyFavoriteTeams(userID uuid.UUID, limit int) ([]TeamCountRow, error) {
	var rows []TeamCountRow
	err := r.db.Raw(`
		SELECT
			CASE WHEN b.bet_choice = 'home' THEN m.home_team ELSE m.away_team END AS team,
			COUNT(*) AS bet_count
		FROM wc_bets b
		JOIN wc_matches m ON m.id = b.match_id
		WHERE b.wc_user_id = ?
		  AND b.bet_type = 'handicap'
		  AND b.bet_choice IN ('home', 'away')
		GROUP BY team
		ORDER BY bet_count DESC
		LIMIT ?
	`, userID, limit).Scan(&rows).Error
	return rows, err
}

// GetMyFavoriteScorelines returns top-N scorelines from exact_score bets.
func (r *WcAnalyticsRepository) GetMyFavoriteScorelines(userID uuid.UUID, limit int) ([]ScorelineCountRow, error) {
	var rows []ScorelineCountRow
	err := r.db.Raw(`
		SELECT
			CONCAT(b.predicted_home_score, '-', b.predicted_away_score) AS scoreline,
			COUNT(*) AS count
		FROM wc_bets b
		WHERE b.wc_user_id = ?
		  AND b.bet_type = 'exact_score'
		  AND b.predicted_home_score IS NOT NULL
		GROUP BY scoreline
		ORDER BY count DESC
		LIMIT ?
	`, userID, limit).Scan(&rows).Error
	return rows, err
}

// GetMyAvgGoals returns total goals and count from exact_score bets for avg calculation.
func (r *WcAnalyticsRepository) GetMyAvgGoals(userID uuid.UUID) (*AvgGoalsRow, error) {
	var row AvgGoalsRow
	err := r.db.Raw(`
		SELECT
			COALESCE(SUM(b.predicted_home_score + b.predicted_away_score), 0) AS total_goals,
			COUNT(*) AS count
		FROM wc_bets b
		WHERE b.wc_user_id = ?
		  AND b.bet_type = 'exact_score'
		  AND b.predicted_home_score IS NOT NULL
		  AND b.predicted_away_score IS NOT NULL
	`, userID).Scan(&row).Error
	return &row, err
}

// GetCommunityBetDistribution returns home/away/other prediction split (all non-void bets).
func (r *WcAnalyticsRepository) GetCommunityBetDistribution() (map[string]int, error) {
	type bucketRow struct {
		Bucket string
		Cnt    int
	}
	var rows []bucketRow
	err := r.db.Raw(`
		SELECT
			CASE
				WHEN b.bet_type = 'handicap' AND b.bet_choice = 'home' THEN 'home'
				WHEN b.bet_type = 'handicap' AND b.bet_choice = 'away' THEN 'away'
				ELSE 'other'
			END AS bucket,
			COUNT(*) AS cnt
		FROM wc_bets b
		WHERE b.result IS NULL OR b.result != 'void'
		GROUP BY bucket
	`).Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	result := map[string]int{"home": 0, "away": 0, "other": 0}
	for _, r := range rows {
		result[r.Bucket] = r.Cnt
	}
	return result, nil
}

// GetCommunityTrendingTeams returns top-N teams by bet count in the last 7 days (non-void handicap bets).
func (r *WcAnalyticsRepository) GetCommunityTrendingTeams(limit int) ([]TeamCountRow, error) {
	var rows []TeamCountRow
	err := r.db.Raw(`
		SELECT
			CASE WHEN b.bet_choice = 'home' THEN m.home_team ELSE m.away_team END AS team,
			COUNT(*) AS bet_count
		FROM wc_bets b
		JOIN wc_matches m ON m.id = b.match_id
		WHERE b.bet_type = 'handicap'
		  AND b.bet_choice IN ('home', 'away')
		  AND b.result != 'void'
		  AND b.created_at > NOW() - INTERVAL '7 days'
		GROUP BY team
		ORDER BY bet_count DESC
		LIMIT ?
	`, limit).Scan(&rows).Error
	return rows, err
}

// GetCommunityTrendingScorelines returns top-N scorelines by pick count (all-time).
func (r *WcAnalyticsRepository) GetCommunityTrendingScorelines(limit int) ([]ScorelineCountRow, error) {
	var rows []ScorelineCountRow
	err := r.db.Raw(`
		SELECT
			CONCAT(b.predicted_home_score, '-', b.predicted_away_score) AS scoreline,
			COUNT(*) AS count
		FROM wc_bets b
		WHERE b.bet_type = 'exact_score'
		  AND b.predicted_home_score IS NOT NULL
		  AND (b.result IS NULL OR b.result != 'void')
		GROUP BY scoreline
		ORDER BY count DESC
		LIMIT ?
	`, limit).Scan(&rows).Error
	return rows, err
}

// GetCommunityTotals returns total bets placed and distinct active user count.
func (r *WcAnalyticsRepository) GetCommunityTotals() (totalBets int, activeUsers int, err error) {
	type totalsRow struct {
		TotalBets   int
		ActiveUsers int
	}
	var row totalsRow
	err = r.db.Raw(`
		SELECT
			COUNT(*)                    AS total_bets,
			COUNT(DISTINCT wc_user_id)  AS active_users
		FROM wc_bets
		WHERE result IS NULL OR result != 'void'
	`).Scan(&row).Error
	return row.TotalBets, row.ActiveUsers, err
}

// GetTopPredictors returns users ranked by accuracy (settled matches), min minMatches.
func (r *WcAnalyticsRepository) GetTopPredictors(minMatches int) ([]PredictorRow, error) {
	var rows []PredictorRow
	err := r.db.Raw(`
		WITH match_results AS (
			SELECT
				b.wc_user_id,
				b.match_id,
				COALESCE(SUM(b.payout), 0) AS total_payout,
				SUM(b.stake)               AS total_stake
			FROM wc_bets b
			WHERE b.result IS NOT NULL AND b.result != 'void'
			GROUP BY b.wc_user_id, b.match_id
		),
		user_stats AS (
			SELECT
				wc_user_id,
				COUNT(*)                                              AS settled_matches,
				COUNT(*) FILTER (WHERE total_payout > total_stake)    AS wins,
				COUNT(*) FILTER (WHERE total_payout <= total_stake)   AS losses
			FROM match_results
			GROUP BY wc_user_id
			HAVING COUNT(*) >= ?
		)
		SELECT
			u.id::text AS wc_user_id,
			u.name,
			u.avatar_url,
			s.wins,
			s.losses,
			s.settled_matches
		FROM user_stats s
		JOIN wc_users u ON u.id = s.wc_user_id
		ORDER BY (s.wins::float / NULLIF(s.settled_matches, 0)) DESC, s.settled_matches DESC
		LIMIT 20
	`, minMatches).Scan(&rows).Error
	return rows, err
}

// GetCommunityAvgStats returns community-wide averages for the compare table.
func (r *WcAnalyticsRepository) GetCommunityAvgStats() (*CommunityAvgStats, error) {
	var row CommunityAvgStats
	err := r.db.Raw(`
		WITH match_results AS (
			SELECT
				b.wc_user_id,
				b.match_id,
				COALESCE(SUM(b.payout), 0) AS total_payout,
				SUM(b.stake)               AS total_stake
			FROM wc_bets b
			WHERE b.result IS NOT NULL AND b.result != 'void'
			GROUP BY b.wc_user_id, b.match_id
		),
		user_accuracy AS (
			SELECT AVG(CASE WHEN total_payout > total_stake THEN 1.0 ELSE 0.0 END) AS avg_acc
			FROM match_results
		),
		community_bets AS (
			SELECT
				COUNT(*)                                                                   AS total_bets,
				COUNT(*) FILTER (WHERE bet_type = 'handicap' AND bet_choice = 'home')     AS home_picks,
				COUNT(*) FILTER (WHERE bet_type = 'handicap')                             AS handicap_bets,
				COUNT(*) FILTER (WHERE bet_type = 'exact_score')                          AS exact_bets,
				COUNT(*) FILTER (WHERE bet_type = 'over_under' AND bet_choice = 'over')   AS over_picks,
				COUNT(*) FILTER (WHERE bet_type = 'over_under')                           AS ou_bets,
				COUNT(*) FILTER (WHERE bet_type = 'handicap' AND bet_choice = 'away')     AS away_picks,
				COALESCE(AVG(stake), 0)                                                   AS avg_stake,
				COUNT(*) FILTER (WHERE bet_type = 'exact_score'
				                   AND result IS NOT NULL AND result != 'void'
				                   AND COALESCE(payout, 0) > stake)                       AS exact_wins,
				COUNT(*) FILTER (WHERE bet_type = 'exact_score'
				                   AND result IS NOT NULL AND result != 'void')            AS exact_settled
			FROM wc_bets b
			WHERE result IS NULL OR result != 'void'
		),
		last_minute AS (
			SELECT
				COUNT(*) FILTER (WHERE b.created_at >= m.match_date - INTERVAL '2 hours'
				                   AND b.created_at <= m.match_date)  AS lm_bets,
				COUNT(*)                                               AS total
			FROM wc_bets b
			JOIN wc_matches m ON m.id = b.match_id
			WHERE b.result IS NULL OR b.result != 'void'
		),
		avg_goals AS (
			SELECT
				COALESCE(SUM(predicted_home_score + predicted_away_score), 0) AS total_goals,
				COUNT(*) AS count
			FROM wc_bets
			WHERE bet_type = 'exact_score'
			  AND predicted_home_score IS NOT NULL
			  AND (result IS NULL OR result != 'void')
		),
		distinct_users AS (
			SELECT COUNT(DISTINCT wc_user_id) AS n FROM wc_bets WHERE result IS NULL OR result != 'void'
		),
		completed_matches AS (
			SELECT COUNT(DISTINCT b.match_id) AS n
			FROM wc_bets b
			WHERE b.result IS NOT NULL AND b.result != 'void'
		)
		SELECT
			COALESCE(ua.avg_acc, 0)                                                        AS avg_accuracy,
			CASE WHEN cb.handicap_bets > 0 THEN cb.home_picks::float / cb.handicap_bets ELSE NULL END AS home_bias,
			CASE WHEN ag.count > 0 THEN ag.total_goals::float / ag.count ELSE NULL END     AS avg_goals_predicted,
			CASE WHEN cb.total_bets > 0 THEN cb.exact_bets::float / cb.total_bets ELSE 0 END AS exact_score_rate,
			CASE WHEN cb.handicap_bets > 0 THEN cb.away_picks::float / cb.handicap_bets ELSE NULL END AS underdog_rate,
			cb.avg_stake,
			CASE WHEN cb.ou_bets > 0 THEN cb.over_picks::float / cb.ou_bets ELSE NULL END  AS over_preference_rate,
			CASE WHEN cb.exact_settled > 0 THEN cb.exact_wins::float / cb.exact_settled ELSE NULL END AS exact_score_hit_rate,
			CASE WHEN cm.n > 0 THEN cb.total_bets::float / NULLIF(du.n, 0) / cm.n ELSE 0 END AS bet_frequency,
			CASE WHEN lm.total > 0 THEN lm.lm_bets::float / lm.total ELSE 0 END           AS last_minute_rate
		FROM community_bets cb, user_accuracy ua, last_minute lm, avg_goals ag, distinct_users du, completed_matches cm
	`).Scan(&row).Error
	return &row, err
}

// CommunityAvgStats holds community-wide compare metrics.
type CommunityAvgStats struct {
	AvgAccuracy         float64  `json:"avg_accuracy"`
	HomeBias            *float64 `json:"home_bias"`
	AvgGoalsPredicted   *float64 `json:"avg_goals_predicted"`
	ExactScoreRate      float64  `json:"exact_score_rate"`
	UnderdogRate        *float64 `json:"underdog_rate"`
	AvgStake            float64  `json:"avg_stake"`
	OverPreferenceRate  *float64 `json:"over_preference_rate"`
	ExactScoreHitRate   *float64 `json:"exact_score_hit_rate"`
	BetFrequency        float64  `json:"bet_frequency"`
	LastMinuteRate      float64  `json:"last_minute_rate"`
}
