package repository

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type HighlightRepository struct {
	db *gorm.DB
}

func NewHighlightRepository(db *gorm.DB) *HighlightRepository {
	return &HighlightRepository{db: db}
}

// StreakData holds current streak info for a user.
type StreakData struct {
	UserID          uuid.UUID
	CurrentWin      int
	CurrentLose     int
	CurrentUnbeaten int
}

// LongestStreakData holds all-time longest streaks.
type LongestStreakData struct {
	UserID      uuid.UUID
	LongestWin  int
	LongestLose int
}

// PointsMovement holds point deltas for a time window.
type PointsMovement struct {
	UserID       uuid.UUID
	PointsGained int
	PointsLost   int
	MatchCount   int
}

// RankData holds current and yesterday's rank for a user.
type RankData struct {
	UserID        uuid.UUID
	CurrentRank   int
	YesterdayRank int
	CurrentScore  int
}

// RecentFormData holds the last 10 decisive results per user (newest first).
type RecentFormData struct {
	UserID  uuid.UUID
	Results []bool // true = win, false = loss
}

// WeeklyActivityData holds weekly match count and consecutive active days.
type WeeklyActivityData struct {
	UserID                uuid.UUID
	MatchesThisWeek       int
	ConsecutiveActiveDays int
}

// TotalsData holds career totals per user.
type TotalsData struct {
	UserID       uuid.UUID
	TotalWins    int
	TotalLosses  int
	TotalMatches int
	CurrentScore int
}

// StreakBreakerData records a match today where the loser had a win streak broken.
type StreakBreakerData struct {
	BreakerID    uuid.UUID
	BreakerName  string
	VictimID     uuid.UUID
	VictimName   string
	StreakType   string
	StreakLength int
}

// GetCurrentStreaks returns current win/lose/unbeaten streak for every active user.
func (r *HighlightRepository) GetCurrentStreaks() (map[uuid.UUID]*StreakData, error) {
	type row struct {
		UserID      uuid.UUID `gorm:"column:user_id"`
		PointChange int       `gorm:"column:point_change"`
	}
	var rows []row
	err := r.db.Raw(`
		SELECT mp.user_id, mp.point_change
		FROM match_participants mp
		JOIN matches m ON m.id = mp.match_id
		JOIN users u ON u.id = mp.user_id
		WHERE u.is_active = true
		ORDER BY mp.user_id, m.match_date DESC, m.created_at DESC
	`).Scan(&rows).Error
	if err != nil {
		return nil, err
	}

	byUser := make(map[uuid.UUID][]int)
	for _, r := range rows {
		byUser[r.UserID] = append(byUser[r.UserID], r.PointChange)
	}

	result := make(map[uuid.UUID]*StreakData)
	for userID, changes := range byUser {
		sd := &StreakData{UserID: userID}
		for _, pc := range changes {
			if pc > 0 {
				sd.CurrentWin++
			} else {
				break
			}
		}
		for _, pc := range changes {
			if pc < 0 {
				sd.CurrentLose++
			} else {
				break
			}
		}
		for _, pc := range changes {
			if pc >= 0 {
				sd.CurrentUnbeaten++
			} else {
				break
			}
		}
		result[userID] = sd
	}
	return result, nil
}

// GetLongestStreaks returns all-time longest win/lose streak per active user.
func (r *HighlightRepository) GetLongestStreaks() (map[uuid.UUID]*LongestStreakData, error) {
	type row struct {
		UserID      uuid.UUID `gorm:"column:user_id"`
		PointChange int       `gorm:"column:point_change"`
	}
	var rows []row
	err := r.db.Raw(`
		SELECT mp.user_id, mp.point_change
		FROM match_participants mp
		JOIN matches m ON m.id = mp.match_id
		JOIN users u ON u.id = mp.user_id
		WHERE u.is_active = true
		ORDER BY mp.user_id, m.match_date ASC, m.created_at ASC
	`).Scan(&rows).Error
	if err != nil {
		return nil, err
	}

	byUser := make(map[uuid.UUID][]int)
	for _, r := range rows {
		byUser[r.UserID] = append(byUser[r.UserID], r.PointChange)
	}

	result := make(map[uuid.UUID]*LongestStreakData)
	for userID, changes := range byUser {
		ld := &LongestStreakData{UserID: userID}
		curWin, curLose := 0, 0
		for _, pc := range changes {
			if pc > 0 {
				curWin++
				curLose = 0
			} else if pc < 0 {
				curLose++
				curWin = 0
			} else {
				curWin = 0
				curLose = 0
			}
			if curWin > ld.LongestWin {
				ld.LongestWin = curWin
			}
			if curLose > ld.LongestLose {
				ld.LongestLose = curLose
			}
		}
		result[userID] = ld
	}
	return result, nil
}

// GetPointsMovementToday returns gains, losses, and match count for today per active user.
func (r *HighlightRepository) GetPointsMovementToday() (map[uuid.UUID]*PointsMovement, error) {
	type row struct {
		UserID       uuid.UUID `gorm:"column:user_id"`
		PointsGained int       `gorm:"column:points_gained"`
		PointsLost   int       `gorm:"column:points_lost"`
		MatchCount   int       `gorm:"column:match_count"`
	}
	var rows []row
	err := r.db.Raw(`
		SELECT mp.user_id,
		       COALESCE(SUM(mp.point_change) FILTER (WHERE mp.point_change > 0), 0)        AS points_gained,
		       COALESCE(SUM(ABS(mp.point_change)) FILTER (WHERE mp.point_change < 0), 0)   AS points_lost,
		       COUNT(*) FILTER (WHERE mp.point_change != 0)                                AS match_count
		FROM match_participants mp
		JOIN matches m ON m.id = mp.match_id
		JOIN users u ON u.id = mp.user_id
		WHERE u.is_active = true
		  AND m.match_date >= CURRENT_DATE AT TIME ZONE 'Asia/Ho_Chi_Minh'
		GROUP BY mp.user_id
	`).Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	result := make(map[uuid.UUID]*PointsMovement)
	for _, r := range rows {
		result[r.UserID] = &PointsMovement{
			UserID: r.UserID, PointsGained: r.PointsGained,
			PointsLost: r.PointsLost, MatchCount: r.MatchCount,
		}
	}
	return result, nil
}

// GetPointsMovementLastHour returns point deltas in the rolling last-60-minute window.
func (r *HighlightRepository) GetPointsMovementLastHour() (map[uuid.UUID]*PointsMovement, error) {
	type row struct {
		UserID       uuid.UUID `gorm:"column:user_id"`
		PointsGained int       `gorm:"column:points_gained"`
		PointsLost   int       `gorm:"column:points_lost"`
		MatchCount   int       `gorm:"column:match_count"`
	}
	var rows []row
	err := r.db.Raw(`
		SELECT mp.user_id,
		       COALESCE(SUM(mp.point_change) FILTER (WHERE mp.point_change > 0), 0)        AS points_gained,
		       COALESCE(SUM(ABS(mp.point_change)) FILTER (WHERE mp.point_change < 0), 0)   AS points_lost,
		       COUNT(*) FILTER (WHERE mp.point_change != 0)                                AS match_count
		FROM match_participants mp
		JOIN matches m ON m.id = mp.match_id
		JOIN users u ON u.id = mp.user_id
		WHERE u.is_active = true
		  AND m.match_date >= NOW() - INTERVAL '1 hour'
		GROUP BY mp.user_id
	`).Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	result := make(map[uuid.UUID]*PointsMovement)
	for _, r := range rows {
		result[r.UserID] = &PointsMovement{
			UserID: r.UserID, PointsGained: r.PointsGained,
			PointsLost: r.PointsLost, MatchCount: r.MatchCount,
		}
	}
	return result, nil
}

// GetRankSnapshot returns current rank and yesterday's effective rank per active user.
func (r *HighlightRepository) GetRankSnapshot() (map[uuid.UUID]*RankData, error) {
	type row struct {
		UserID        uuid.UUID `gorm:"column:user_id"`
		CurrentScore  int       `gorm:"column:current_score"`
		CurrentRank   int       `gorm:"column:current_rank"`
		YesterdayRank int       `gorm:"column:yesterday_rank"`
	}
	var rows []row
	err := r.db.Raw(`
		WITH today_deltas AS (
		    SELECT mp.user_id, SUM(mp.point_change) AS delta
		    FROM match_participants mp
		    JOIN matches m ON m.id = mp.match_id
		    WHERE m.match_date >= CURRENT_DATE AT TIME ZONE 'Asia/Ho_Chi_Minh'
		    GROUP BY mp.user_id
		),
		scores AS (
		    SELECT u.id AS user_id,
		           u.current_score,
		           u.current_score - COALESCE(td.delta, 0) AS yesterday_score
		    FROM users u
		    LEFT JOIN today_deltas td ON td.user_id = u.id
		    WHERE u.is_active = true
		)
		SELECT user_id, current_score,
		       RANK() OVER (ORDER BY current_score   DESC) AS current_rank,
		       RANK() OVER (ORDER BY yesterday_score DESC) AS yesterday_rank
		FROM scores
	`).Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	result := make(map[uuid.UUID]*RankData)
	for _, r := range rows {
		result[r.UserID] = &RankData{
			UserID: r.UserID, CurrentScore: r.CurrentScore,
			CurrentRank: r.CurrentRank, YesterdayRank: r.YesterdayRank,
		}
	}
	return result, nil
}

// GetRecentForm returns the last 10 decisive results per active user, newest first.
func (r *HighlightRepository) GetRecentForm() (map[uuid.UUID]*RecentFormData, error) {
	type row struct {
		UserID      uuid.UUID `gorm:"column:user_id"`
		PointChange int       `gorm:"column:point_change"`
	}
	var rows []row
	err := r.db.Raw(`
		SELECT user_id, point_change
		FROM (
		    SELECT mp.user_id, mp.point_change,
		           ROW_NUMBER() OVER (PARTITION BY mp.user_id ORDER BY m.match_date DESC, m.created_at DESC) AS rn
		    FROM match_participants mp
		    JOIN matches m ON m.id = mp.match_id
		    JOIN users u ON u.id = mp.user_id
		    WHERE u.is_active = true AND mp.point_change != 0
		) ranked
		WHERE rn <= 10
		ORDER BY user_id, rn
	`).Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	result := make(map[uuid.UUID]*RecentFormData)
	for _, r := range rows {
		if _, ok := result[r.UserID]; !ok {
			result[r.UserID] = &RecentFormData{UserID: r.UserID}
		}
		result[r.UserID].Results = append(result[r.UserID].Results, r.PointChange > 0)
	}
	return result, nil
}

// GetWeeklyActivity returns match count this week and consecutive active days per active user.
func (r *HighlightRepository) GetWeeklyActivity() (map[uuid.UUID]*WeeklyActivityData, error) {
	type weekRow struct {
		UserID          uuid.UUID `gorm:"column:user_id"`
		MatchesThisWeek int       `gorm:"column:matches_this_week"`
	}
	type dayRow struct {
		UserID   uuid.UUID `gorm:"column:user_id"`
		MatchDay time.Time `gorm:"column:match_day"`
	}

	var weekRows []weekRow
	err := r.db.Raw(`
		SELECT mp.user_id,
		       COUNT(*) FILTER (WHERE mp.point_change != 0) AS matches_this_week
		FROM match_participants mp
		JOIN matches m ON m.id = mp.match_id
		JOIN users u ON u.id = mp.user_id
		WHERE u.is_active = true
		  AND m.match_date >= date_trunc('week', NOW() AT TIME ZONE 'Asia/Ho_Chi_Minh') AT TIME ZONE 'Asia/Ho_Chi_Minh'
		GROUP BY mp.user_id
	`).Scan(&weekRows).Error
	if err != nil {
		return nil, err
	}

	var dayRows []dayRow
	err = r.db.Raw(`
		SELECT DISTINCT mp.user_id,
		       DATE_TRUNC('day', m.match_date AT TIME ZONE 'Asia/Ho_Chi_Minh') AS match_day
		FROM match_participants mp
		JOIN matches m ON m.id = mp.match_id
		JOIN users u ON u.id = mp.user_id
		WHERE u.is_active = true AND mp.point_change != 0
		  AND m.match_date >= NOW() - INTERVAL '30 days'
		ORDER BY mp.user_id, match_day DESC
	`).Scan(&dayRows).Error
	if err != nil {
		return nil, err
	}

	result := make(map[uuid.UUID]*WeeklyActivityData)
	for _, r := range weekRows {
		result[r.UserID] = &WeeklyActivityData{UserID: r.UserID, MatchesThisWeek: r.MatchesThisWeek}
	}

	byUser := make(map[uuid.UUID][]time.Time)
	for _, d := range dayRows {
		byUser[d.UserID] = append(byUser[d.UserID], d.MatchDay)
	}
	for userID, days := range byUser {
		if _, ok := result[userID]; !ok {
			result[userID] = &WeeklyActivityData{UserID: userID}
		}
		streak := 0
		for i, day := range days {
			if i == 0 {
				streak = 1
				continue
			}
			diff := days[i-1].Sub(day)
			if diff >= 20*time.Hour && diff <= 28*time.Hour {
				streak++
			} else {
				break
			}
		}
		result[userID].ConsecutiveActiveDays = streak
	}
	return result, nil
}

// GetTotals returns career totals per active user.
func (r *HighlightRepository) GetTotals() (map[uuid.UUID]*TotalsData, error) {
	type row struct {
		UserID       uuid.UUID `gorm:"column:user_id"`
		TotalWins    int       `gorm:"column:total_wins"`
		TotalLosses  int       `gorm:"column:total_losses"`
		TotalMatches int       `gorm:"column:total_matches"`
		CurrentScore int       `gorm:"column:current_score"`
	}
	var rows []row
	err := r.db.Raw(`
		SELECT u.id AS user_id, u.current_score,
		       COUNT(mp.id) FILTER (WHERE mp.point_change > 0)  AS total_wins,
		       COUNT(mp.id) FILTER (WHERE mp.point_change < 0)  AS total_losses,
		       COUNT(mp.id) FILTER (WHERE mp.point_change != 0) AS total_matches
		FROM users u
		LEFT JOIN match_participants mp ON mp.user_id = u.id
		WHERE u.is_active = true
		GROUP BY u.id
	`).Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	result := make(map[uuid.UUID]*TotalsData)
	for _, r := range rows {
		result[r.UserID] = &TotalsData{
			UserID: r.UserID, TotalWins: r.TotalWins,
			TotalLosses: r.TotalLosses, TotalMatches: r.TotalMatches,
			CurrentScore: r.CurrentScore,
		}
	}
	return result, nil
}

// GetStreakBreakers returns matches today where the loser had a win streak >= StreakBreakerMinLength.
func (r *HighlightRepository) GetStreakBreakers() ([]*StreakBreakerData, error) {
	type row struct {
		BreakerID    uuid.UUID `gorm:"column:breaker_id"`
		BreakerName  string    `gorm:"column:breaker_name"`
		VictimID     uuid.UUID `gorm:"column:victim_id"`
		VictimName   string    `gorm:"column:victim_name"`
		StreakLength  int       `gorm:"column:streak_length"`
	}
	var rows []row
	// Counts how many consecutive wins the loser had before each today-match they lost.
	err := r.db.Raw(`
		WITH today_losses AS (
		    SELECT mp.user_id AS victim_id, mp.match_id, m.match_date AS lost_at
		    FROM match_participants mp
		    JOIN matches m ON m.id = mp.match_id
		    WHERE mp.point_change < 0
		      AND m.match_date >= CURRENT_DATE AT TIME ZONE 'Asia/Ho_Chi_Minh'
		),
		today_wins AS (
		    SELECT mp.user_id AS breaker_id, mp.match_id
		    FROM match_participants mp
		    JOIN matches m ON m.id = mp.match_id
		    WHERE mp.point_change > 0
		      AND m.match_date >= CURRENT_DATE AT TIME ZONE 'Asia/Ho_Chi_Minh'
		),
		pre_streaks AS (
		    SELECT tl.victim_id, tl.match_id,
		           COUNT(*) AS streak_length
		    FROM today_losses tl
		    JOIN (
		        SELECT mp2.user_id, mp2.point_change,
		               ROW_NUMBER() OVER (PARTITION BY mp2.user_id ORDER BY m2.match_date DESC, m2.created_at DESC) AS rn,
		               m2.match_date
		        FROM match_participants mp2
		        JOIN matches m2 ON m2.id = mp2.match_id
		        WHERE mp2.point_change != 0
		    ) prev ON prev.user_id = tl.victim_id
		             AND prev.match_date < tl.lost_at
		             AND prev.point_change > 0
		    WHERE prev.rn <= 20
		    GROUP BY tl.victim_id, tl.match_id
		    HAVING COUNT(*) >= ?
		)
		SELECT tw.breaker_id, ub.name AS breaker_name,
		       ps.victim_id,  uv.name AS victim_name,
		       ps.streak_length
		FROM pre_streaks ps
		JOIN today_wins tw ON tw.match_id = ps.match_id
		JOIN users ub ON ub.id = tw.breaker_id
		JOIN users uv ON uv.id = ps.victim_id
	`, 3).Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	var result []*StreakBreakerData
	for _, r := range rows {
		result = append(result, &StreakBreakerData{
			BreakerID: r.BreakerID, BreakerName: r.BreakerName,
			VictimID: r.VictimID, VictimName: r.VictimName,
			StreakType: "win", StreakLength: r.StreakLength,
		})
	}
	return result, nil
}
