package repository

import (
	"fmt"
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
	MatchID     string
	MatchDate   time.Time
	TotalPayout float64
	TotalStake  float64
}

// TimelineRow holds wins/losses for one day bucket.
type TimelineRow struct {
	Period string
	Wins   int
	Losses int
}

// TeamCountRow holds a team name and bet count.
type TeamCountRow struct {
	Team     string `json:"team"`
	BetCount int    `json:"bet_count"`
}

// ScorelineCountRow holds a scoreline string and count.
type ScorelineCountRow struct {
	Scoreline string `json:"scoreline"`
	Count     int    `json:"count"`
}

// BetTypeDistribution holds counts per bet type.
type BetTypeDistribution struct {
	Handicap   int
	ExactScore int
	OverUnder  int
	Custom     int
}

// MyBetStatsRow holds raw aggregate stats for one user across all bet sources.
type MyBetStatsRow struct {
	TotalBets         int
	PendingBets       int
	HandicapBets      int
	ExactScoreBets    int
	OverUnderBets     int
	CustomBets        int
	HomePicks         int
	AwayPicks         int
	OverPicks         int
	UnderPicks        int
	TotalStake        int
	AvgStake          float64
	LastMinuteBets    int
	ExactScoreWins    int
	ExactScoreSettled int
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

// settledBetsCTEFor returns the settled_bets CTE filtered by tournamentType.
// It UNIONs wc_bets + wc_predictions + wc_custom_bet_entries for a single user.
// Parameters: userID, userID, userID (one per branch).
// tt must be an internal constant ('world_cup' or 'asean_cup').
func settledBetsCTEFor(tt string) string {
	return fmt.Sprintf(`
	settled_bets AS (
		SELECT b.match_id, b.stake::numeric AS stake, COALESCE(b.payout, 0) AS payout
		FROM wc_bets b
		WHERE b.wc_user_id = ? AND b.result IS NOT NULL AND b.tournament_type = '%s'

		UNION ALL

		SELECT p.match_id, p.points::numeric AS stake, COALESCE(p.points_earned, 0) AS payout
		FROM wc_predictions p
		WHERE p.wc_user_id = ? AND p.result IS NOT NULL AND p.result != 'void' AND p.cancelled_at IS NULL AND p.tournament_type = '%s'

		UNION ALL

		SELECT cb.match_id, e.stake::numeric AS stake, COALESCE(e.payout, 0) AS payout
		FROM wc_custom_bet_entries e
		JOIN wc_custom_bets cb ON cb.id = e.custom_bet_id
		WHERE e.wc_user_id = ? AND e.status IN ('won', 'lost') AND cb.tournament_type = '%s'
	)
`, tt, tt, tt)
}

// GetMyMatchResults returns per-match net payout vs stake for a user within a period.
func (r *WcAnalyticsRepository) GetMyMatchResults(userID uuid.UUID, p AnalyticsPeriod, tournamentType string) ([]MatchResultRow, error) {
	var rows []MatchResultRow
	err := r.db.Raw(`
		WITH `+settledBetsCTEFor(tournamentType)+`
		SELECT
			s.match_id::text AS match_id,
			m.match_date,
			SUM(s.payout)    AS total_payout,
			SUM(s.stake)     AS total_stake
		FROM settled_bets s
		JOIN wc_matches m ON m.id = s.match_id
		WHERE m.match_date >= ? AND m.match_date <= ?
		GROUP BY s.match_id, m.match_date
		ORDER BY m.match_date ASC
	`, userID, userID, userID, p.From, p.To).Scan(&rows).Error
	return rows, err
}

// GetMyAccuracyTimeline returns wins/losses bucketed by day within the period.
func (r *WcAnalyticsRepository) GetMyAccuracyTimeline(userID uuid.UUID, p AnalyticsPeriod, tournamentType string) ([]TimelineRow, error) {
	var rows []TimelineRow
	err := r.db.Raw(`
		WITH `+settledBetsCTEFor(tournamentType)+`,
		match_results AS (
			SELECT
				DATE(m.match_date) AS period,
				SUM(s.payout)      AS total_payout,
				SUM(s.stake)       AS total_stake
			FROM settled_bets s
			JOIN wc_matches m ON m.id = s.match_id
			WHERE m.match_date >= ? AND m.match_date <= ?
			GROUP BY s.match_id, DATE(m.match_date)
		)
		SELECT
			period::text,
			COUNT(*) FILTER (WHERE total_payout > total_stake) AS wins,
			COUNT(*) FILTER (WHERE total_payout <= total_stake) AS losses
		FROM match_results
		GROUP BY period
		ORDER BY period ASC
	`, userID, userID, userID, p.From, p.To).Scan(&rows).Error
	return rows, err
}

// GetMyBetStats returns aggregate betting stats for a user across all bet sources (no period filter).
func (r *WcAnalyticsRepository) GetMyBetStats(userID uuid.UUID, tournamentType string) (*MyBetStatsRow, error) {
	var row MyBetStatsRow
	err := r.db.Raw(`
		WITH unified AS (
			SELECT
				b.match_id,
				b.bet_type                                                                     AS bet_type,
				b.bet_choice                                                                   AS choice,
				b.stake::numeric                                                               AS stake,
				(b.result IS NULL)                                                             AS is_pending,
				(b.bet_type = 'exact_score' AND b.result IS NOT NULL
				 AND COALESCE(b.payout, 0) > b.stake::numeric)                                AS is_exact_win,
				(b.bet_type = 'exact_score' AND b.result IS NOT NULL)                         AS is_exact_settled,
				b.created_at
			FROM wc_bets b
			WHERE b.wc_user_id = ? AND b.tournament_type = ?

			UNION ALL

			SELECT
				p.match_id,
				p.prediction_type                                                              AS bet_type,
				p.prediction_choice                                                            AS choice,
				p.points::numeric                                                              AS stake,
				(p.result IS NULL)                                                             AS is_pending,
				(p.prediction_type = 'exact_score' AND p.result = 'correct')                  AS is_exact_win,
				(p.prediction_type = 'exact_score' AND p.result IS NOT NULL
				 AND p.result != 'void')                                                       AS is_exact_settled,
				p.created_at
			FROM wc_predictions p
			WHERE p.wc_user_id = ?
			  AND p.cancelled_at IS NULL
			  AND (p.result IS NULL OR p.result != 'void')
			  AND p.tournament_type = ?

			UNION ALL

			SELECT
				cb.match_id,
				'custom'                                                                       AS bet_type,
				NULL                                                                           AS choice,
				e.stake::numeric                                                               AS stake,
				(e.status = 'pending')                                                        AS is_pending,
				false                                                                          AS is_exact_win,
				false                                                                          AS is_exact_settled,
				e.created_at
			FROM wc_custom_bet_entries e
			JOIN wc_custom_bets cb ON cb.id = e.custom_bet_id
			WHERE e.wc_user_id = ?
			  AND e.status != 'void'
			  AND cb.tournament_type = ?
		)
		SELECT
			COUNT(*)                                                                           AS total_bets,
			COUNT(*) FILTER (WHERE u.is_pending)                                               AS pending_bets,
			COUNT(*) FILTER (WHERE u.bet_type = 'handicap')                                    AS handicap_bets,
			COUNT(*) FILTER (WHERE u.bet_type = 'exact_score')                                 AS exact_score_bets,
			COUNT(*) FILTER (WHERE u.bet_type = 'over_under')                                  AS over_under_bets,
			COUNT(*) FILTER (WHERE u.bet_type = 'custom')                                      AS custom_bets,
			COUNT(*) FILTER (WHERE u.bet_type = 'handicap' AND u.choice = 'home')              AS home_picks,
			COUNT(*) FILTER (WHERE u.bet_type = 'handicap' AND u.choice = 'away')              AS away_picks,
			COUNT(*) FILTER (WHERE u.bet_type = 'over_under' AND u.choice = 'over')            AS over_picks,
			COUNT(*) FILTER (WHERE u.bet_type = 'over_under' AND u.choice = 'under')           AS under_picks,
			COALESCE(SUM(u.stake), 0)                                                          AS total_stake,
			COALESCE(AVG(u.stake), 0)                                                          AS avg_stake,
			COUNT(*) FILTER (
				WHERE u.created_at >= m.match_date - INTERVAL '2 hours'
				  AND u.created_at <= m.match_date
			)                                                                                  AS last_minute_bets,
			COUNT(*) FILTER (WHERE u.is_exact_win)                                             AS exact_score_wins,
			COUNT(*) FILTER (WHERE u.is_exact_settled)                                         AS exact_score_settled
		FROM unified u
		JOIN wc_matches m ON m.id = u.match_id
	`, userID, tournamentType, userID, tournamentType, userID, tournamentType).Scan(&row).Error
	return &row, err
}

// GetMyFavoriteTeams returns top-N teams from handicap bets/predictions for a user.
func (r *WcAnalyticsRepository) GetMyFavoriteTeams(userID uuid.UUID, limit int, tournamentType string) ([]TeamCountRow, error) {
	var rows []TeamCountRow
	err := r.db.Raw(`
		SELECT
			CASE WHEN src.choice = 'home' THEN m.home_team ELSE m.away_team END AS team,
			COUNT(*) AS bet_count
		FROM (
			SELECT b.match_id, b.bet_choice AS choice
			FROM wc_bets b
			WHERE b.wc_user_id = ?
			  AND b.bet_type = 'handicap'
			  AND b.bet_choice IN ('home', 'away')
			  AND b.tournament_type = ?

			UNION ALL

			SELECT p.match_id, p.prediction_choice AS choice
			FROM wc_predictions p
			WHERE p.wc_user_id = ?
			  AND p.prediction_type = 'handicap'
			  AND p.prediction_choice IN ('home', 'away')
			  AND (p.result IS NULL OR p.result != 'void')
			  AND p.cancelled_at IS NULL
			  AND p.tournament_type = ?
		) src
		JOIN wc_matches m ON m.id = src.match_id
		GROUP BY team
		ORDER BY bet_count DESC
		LIMIT ?
	`, userID, tournamentType, userID, tournamentType, limit).Scan(&rows).Error
	return rows, err
}

// GetMyFavoriteScorelines returns top-N scorelines from exact_score bets/predictions.
func (r *WcAnalyticsRepository) GetMyFavoriteScorelines(userID uuid.UUID, limit int, tournamentType string) ([]ScorelineCountRow, error) {
	var rows []ScorelineCountRow
	err := r.db.Raw(`
		SELECT
			CONCAT(src.home_score, '-', src.away_score) AS scoreline,
			COUNT(*) AS count
		FROM (
			SELECT b.predicted_home_score AS home_score, b.predicted_away_score AS away_score
			FROM wc_bets b
			WHERE b.wc_user_id = ?
			  AND b.bet_type = 'exact_score'
			  AND b.predicted_home_score IS NOT NULL
			  AND b.tournament_type = ?

			UNION ALL

			SELECT p.predicted_home_score AS home_score, p.predicted_away_score AS away_score
			FROM wc_predictions p
			WHERE p.wc_user_id = ?
			  AND p.prediction_type = 'exact_score'
			  AND p.predicted_home_score IS NOT NULL
			  AND (p.result IS NULL OR p.result != 'void')
			  AND p.cancelled_at IS NULL
			  AND p.tournament_type = ?
		) src
		GROUP BY scoreline
		ORDER BY count DESC
		LIMIT ?
	`, userID, tournamentType, userID, tournamentType, limit).Scan(&rows).Error
	return rows, err
}

// GetMyAvgGoals returns total goals and count from exact_score bets/predictions for avg calculation.
func (r *WcAnalyticsRepository) GetMyAvgGoals(userID uuid.UUID, tournamentType string) (*AvgGoalsRow, error) {
	var row AvgGoalsRow
	err := r.db.Raw(`
		SELECT
			COALESCE(SUM(src.home_score + src.away_score), 0) AS total_goals,
			COUNT(*) AS count
		FROM (
			SELECT b.predicted_home_score AS home_score, b.predicted_away_score AS away_score
			FROM wc_bets b
			WHERE b.wc_user_id = ?
			  AND b.bet_type = 'exact_score'
			  AND b.predicted_home_score IS NOT NULL
			  AND b.predicted_away_score IS NOT NULL
			  AND b.tournament_type = ?

			UNION ALL

			SELECT p.predicted_home_score AS home_score, p.predicted_away_score AS away_score
			FROM wc_predictions p
			WHERE p.wc_user_id = ?
			  AND p.prediction_type = 'exact_score'
			  AND p.predicted_home_score IS NOT NULL
			  AND p.predicted_away_score IS NOT NULL
			  AND (p.result IS NULL OR p.result != 'void')
			  AND p.cancelled_at IS NULL
			  AND p.tournament_type = ?
		) src
	`, userID, tournamentType, userID, tournamentType).Scan(&row).Error
	return &row, err
}

// GetCommunityBetDistribution returns home/away/other prediction split across all bet sources.
func (r *WcAnalyticsRepository) GetCommunityBetDistribution(tournamentType string) (map[string]int, error) {
	type bucketRow struct {
		Bucket string
		Cnt    int
	}
	var rows []bucketRow
	err := r.db.Raw(`
		SELECT
			CASE
				WHEN bet_type = 'handicap' AND choice = 'home' THEN 'home'
				WHEN bet_type = 'handicap' AND choice = 'away' THEN 'away'
				ELSE 'other'
			END AS bucket,
			COUNT(*) AS cnt
		FROM (
			SELECT b.bet_type, b.bet_choice AS choice
			FROM wc_bets b
			WHERE b.tournament_type = ?

			UNION ALL

			SELECT p.prediction_type AS bet_type, p.prediction_choice AS choice
			FROM wc_predictions p
			WHERE (p.result IS NULL OR p.result != 'void')
			  AND p.cancelled_at IS NULL
			  AND p.tournament_type = ?

			UNION ALL

			SELECT 'custom' AS bet_type, NULL AS choice
			FROM wc_custom_bet_entries e
			JOIN wc_custom_bets cb ON cb.id = e.custom_bet_id
			WHERE e.status != 'void'
			  AND cb.tournament_type = ?
		) src
		GROUP BY bucket
	`, tournamentType, tournamentType, tournamentType).Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	result := map[string]int{"home": 0, "away": 0, "other": 0}
	for _, r := range rows {
		result[r.Bucket] = r.Cnt
	}
	return result, nil
}

// GetCommunityTrendingTeams returns top-N teams by handicap bet count in the last 7 days.
func (r *WcAnalyticsRepository) GetCommunityTrendingTeams(limit int, tournamentType string) ([]TeamCountRow, error) {
	var rows []TeamCountRow
	err := r.db.Raw(`
		SELECT
			CASE WHEN src.choice = 'home' THEN m.home_team ELSE m.away_team END AS team,
			COUNT(*) AS bet_count
		FROM (
			SELECT b.match_id, b.bet_choice AS choice
			FROM wc_bets b
			WHERE b.bet_type = 'handicap'
			  AND b.bet_choice IN ('home', 'away')
			  AND b.created_at > NOW() - INTERVAL '7 days'
			  AND b.tournament_type = ?

			UNION ALL

			SELECT p.match_id, p.prediction_choice AS choice
			FROM wc_predictions p
			WHERE p.prediction_type = 'handicap'
			  AND p.prediction_choice IN ('home', 'away')
			  AND (p.result IS NULL OR p.result != 'void')
			  AND p.cancelled_at IS NULL
			  AND p.created_at > NOW() - INTERVAL '7 days'
			  AND p.tournament_type = ?
		) src
		JOIN wc_matches m ON m.id = src.match_id
		GROUP BY team
		ORDER BY bet_count DESC
		LIMIT ?
	`, tournamentType, tournamentType, limit).Scan(&rows).Error
	return rows, err
}

// GetCommunityTrendingScorelines returns top-N scorelines by pick count across bets/predictions.
func (r *WcAnalyticsRepository) GetCommunityTrendingScorelines(limit int, tournamentType string) ([]ScorelineCountRow, error) {
	var rows []ScorelineCountRow
	err := r.db.Raw(`
		SELECT
			CONCAT(src.home_score, '-', src.away_score) AS scoreline,
			COUNT(*) AS count
		FROM (
			SELECT b.predicted_home_score AS home_score, b.predicted_away_score AS away_score
			FROM wc_bets b
			WHERE b.bet_type = 'exact_score'
			  AND b.predicted_home_score IS NOT NULL
			  AND b.tournament_type = ?

			UNION ALL

			SELECT p.predicted_home_score AS home_score, p.predicted_away_score AS away_score
			FROM wc_predictions p
			WHERE p.prediction_type = 'exact_score'
			  AND p.predicted_home_score IS NOT NULL
			  AND (p.result IS NULL OR p.result != 'void')
			  AND p.cancelled_at IS NULL
			  AND p.tournament_type = ?
		) src
		GROUP BY scoreline
		ORDER BY count DESC
		LIMIT ?
	`, tournamentType, tournamentType, limit).Scan(&rows).Error
	return rows, err
}

// GetCommunityTotals returns total bets placed and distinct active user count across all sources.
func (r *WcAnalyticsRepository) GetCommunityTotals(tournamentType string) (totalBets int, activeUsers int, err error) {
	type totalsRow struct {
		TotalBets   int
		ActiveUsers int
	}
	var row totalsRow
	err = r.db.Raw(`
		SELECT
			COUNT(*)                   AS total_bets,
			COUNT(DISTINCT wc_user_id) AS active_users
		FROM (
			SELECT b.id, b.wc_user_id FROM wc_bets b
			WHERE b.tournament_type = ?

			UNION ALL

			SELECT p.id, p.wc_user_id FROM wc_predictions p
			WHERE (p.result IS NULL OR p.result != 'void')
			  AND p.cancelled_at IS NULL
			  AND p.tournament_type = ?

			UNION ALL

			SELECT e.id, e.wc_user_id FROM wc_custom_bet_entries e
			JOIN wc_custom_bets cb ON cb.id = e.custom_bet_id
			WHERE e.status != 'void'
			  AND cb.tournament_type = ?
		) src
	`, tournamentType, tournamentType, tournamentType).Scan(&row).Error
	return row.TotalBets, row.ActiveUsers, err
}

// GetTopPredictors returns users ranked by accuracy (settled matches), min minMatches.
func (r *WcAnalyticsRepository) GetTopPredictors(minMatches int, tournamentType string) ([]PredictorRow, error) {
	var rows []PredictorRow
	err := r.db.Raw(`
		WITH all_settled AS (
			SELECT b.wc_user_id, b.match_id, b.stake::numeric AS stake, COALESCE(b.payout, 0) AS payout
			FROM wc_bets b
			WHERE b.result IS NOT NULL AND b.tournament_type = ?

			UNION ALL

			SELECT p.wc_user_id, p.match_id, p.points::numeric AS stake, COALESCE(p.points_earned, 0) AS payout
			FROM wc_predictions p
			WHERE p.result IS NOT NULL AND p.result != 'void' AND p.cancelled_at IS NULL AND p.tournament_type = ?

			UNION ALL

			SELECT e.wc_user_id, cb.match_id, e.stake::numeric AS stake, COALESCE(e.payout, 0) AS payout
			FROM wc_custom_bet_entries e
			JOIN wc_custom_bets cb ON cb.id = e.custom_bet_id
			WHERE e.status IN ('won', 'lost') AND cb.tournament_type = ?
		),
		match_results AS (
			SELECT
				wc_user_id,
				match_id,
				SUM(payout) AS total_payout,
				SUM(stake)  AS total_stake
			FROM all_settled
			GROUP BY wc_user_id, match_id
		),
		user_stats AS (
			SELECT
				wc_user_id,
				COUNT(*)                                             AS settled_matches,
				COUNT(*) FILTER (WHERE total_payout > total_stake)  AS wins,
				COUNT(*) FILTER (WHERE total_payout <= total_stake) AS losses
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
	`, tournamentType, tournamentType, tournamentType, minMatches).Scan(&rows).Error
	return rows, err
}

// GetCommunityAvgStats returns community-wide averages for the compare table.
func (r *WcAnalyticsRepository) GetCommunityAvgStats(tournamentType string) (*CommunityAvgStats, error) {
	var row CommunityAvgStats
	err := r.db.Raw(`
		WITH all_bets AS (
			SELECT
				b.wc_user_id,
				b.match_id,
				b.bet_type,
				b.bet_choice                                                                   AS choice,
				b.stake::numeric                                                               AS stake,
				(b.result IS NOT NULL)                                                         AS is_settled,
				COALESCE(b.payout, 0)                                                          AS payout,
				(b.bet_type = 'exact_score' AND b.result IS NOT NULL
				 AND COALESCE(b.payout, 0) > b.stake::numeric)                                AS is_exact_win,
				(b.bet_type = 'exact_score' AND b.result IS NOT NULL)                         AS is_exact_settled,
				b.predicted_home_score,
				b.predicted_away_score,
				b.created_at
			FROM wc_bets b
			WHERE b.tournament_type = ?

			UNION ALL

			SELECT
				p.wc_user_id,
				p.match_id,
				p.prediction_type                                                              AS bet_type,
				p.prediction_choice                                                            AS choice,
				p.points::numeric                                                              AS stake,
				(p.result IS NOT NULL AND p.result != 'void')                                 AS is_settled,
				COALESCE(p.points_earned, 0)                                                   AS payout,
				(p.prediction_type = 'exact_score' AND p.result = 'correct')                  AS is_exact_win,
				(p.prediction_type = 'exact_score' AND p.result IS NOT NULL
				 AND p.result != 'void')                                                       AS is_exact_settled,
				p.predicted_home_score,
				p.predicted_away_score,
				p.created_at
			FROM wc_predictions p
			WHERE (p.result IS NULL OR p.result != 'void')
			  AND p.cancelled_at IS NULL
			  AND p.tournament_type = ?

			UNION ALL

			SELECT
				e.wc_user_id,
				cb.match_id,
				'custom'                                                                       AS bet_type,
				NULL                                                                           AS choice,
				e.stake::numeric                                                               AS stake,
				(e.status IN ('won', 'lost'))                                                  AS is_settled,
				COALESCE(e.payout, 0)                                                          AS payout,
				false                                                                          AS is_exact_win,
				false                                                                          AS is_exact_settled,
				NULL::int                                                                      AS predicted_home_score,
				NULL::int                                                                      AS predicted_away_score,
				e.created_at
			FROM wc_custom_bet_entries e
			JOIN wc_custom_bets cb ON cb.id = e.custom_bet_id
			WHERE e.status != 'void'
			  AND cb.tournament_type = ?
		),
		match_results AS (
			SELECT
				wc_user_id,
				match_id,
				SUM(payout) AS total_payout,
				SUM(stake)  AS total_stake
			FROM all_bets
			WHERE is_settled
			GROUP BY wc_user_id, match_id
		),
		user_accuracy AS (
			SELECT AVG(CASE WHEN total_payout > total_stake THEN 1.0 ELSE 0.0 END) AS avg_acc
			FROM match_results
		),
		stats AS (
			SELECT
				COUNT(*)                                                                       AS total_bets,
				COUNT(*) FILTER (WHERE bet_type = 'handicap' AND choice = 'home')             AS home_picks,
				COUNT(*) FILTER (WHERE bet_type = 'handicap')                                 AS handicap_bets,
				COUNT(*) FILTER (WHERE bet_type = 'exact_score')                              AS exact_bets,
				COUNT(*) FILTER (WHERE bet_type = 'over_under' AND choice = 'over')           AS over_picks,
				COUNT(*) FILTER (WHERE bet_type = 'over_under')                               AS ou_bets,
				COUNT(*) FILTER (WHERE bet_type = 'handicap' AND choice = 'away')             AS away_picks,
				COALESCE(AVG(stake), 0)                                                       AS avg_stake,
				COUNT(*) FILTER (WHERE is_exact_win)                                          AS exact_wins,
				COUNT(*) FILTER (WHERE is_exact_settled)                                      AS exact_settled
			FROM all_bets
		),
		avg_goals AS (
			SELECT
				COALESCE(SUM(predicted_home_score + predicted_away_score), 0) AS total_goals,
				COUNT(*) AS count
			FROM all_bets
			WHERE bet_type = 'exact_score'
			  AND predicted_home_score IS NOT NULL
		),
		last_minute AS (
			SELECT
				COUNT(*) FILTER (
					WHERE ab.created_at >= m.match_date - INTERVAL '2 hours'
					  AND ab.created_at <= m.match_date
				)        AS lm_bets,
				COUNT(*) AS total
			FROM all_bets ab
			JOIN wc_matches m ON m.id = ab.match_id
		),
		distinct_users AS (
			SELECT COUNT(DISTINCT wc_user_id) AS n FROM all_bets
		),
		completed_matches AS (
			SELECT COUNT(DISTINCT match_id) AS n FROM match_results
		)
		SELECT
			COALESCE(ua.avg_acc, 0)                                                            AS avg_accuracy,
			CASE WHEN s.handicap_bets > 0 THEN s.home_picks::float / s.handicap_bets ELSE NULL END AS home_bias,
			CASE WHEN ag.count > 0 THEN ag.total_goals::float / ag.count ELSE NULL END         AS avg_goals_predicted,
			CASE WHEN s.total_bets > 0 THEN s.exact_bets::float / s.total_bets ELSE 0 END     AS exact_score_rate,
			CASE WHEN s.handicap_bets > 0 THEN s.away_picks::float / s.handicap_bets ELSE NULL END AS underdog_rate,
			s.avg_stake,
			CASE WHEN s.ou_bets > 0 THEN s.over_picks::float / s.ou_bets ELSE NULL END        AS over_preference_rate,
			CASE WHEN s.exact_settled > 0 THEN s.exact_wins::float / s.exact_settled ELSE NULL END AS exact_score_hit_rate,
			CASE WHEN cm.n > 0 THEN s.total_bets::float / NULLIF(du.n, 0) / cm.n ELSE 0 END  AS bet_frequency,
			CASE WHEN lm.total > 0 THEN lm.lm_bets::float / lm.total ELSE 0 END               AS last_minute_rate
		FROM stats s, user_accuracy ua, last_minute lm, avg_goals ag, distinct_users du, completed_matches cm
	`, tournamentType, tournamentType, tournamentType).Scan(&row).Error
	return &row, err
}

// CommunityAvgStats holds community-wide compare metrics.
type CommunityAvgStats struct {
	AvgAccuracy        float64  `json:"avg_accuracy"`
	HomeBias           *float64 `json:"home_bias"`
	AvgGoalsPredicted  *float64 `json:"avg_goals_predicted"`
	ExactScoreRate     float64  `json:"exact_score_rate"`
	UnderdogRate       *float64 `json:"underdog_rate"`
	AvgStake           float64  `json:"avg_stake"`
	OverPreferenceRate *float64 `json:"over_preference_rate"`
	ExactScoreHitRate  *float64 `json:"exact_score_hit_rate"`
	BetFrequency       float64  `json:"bet_frequency"`
	LastMinuteRate     float64  `json:"last_minute_rate"`
}
