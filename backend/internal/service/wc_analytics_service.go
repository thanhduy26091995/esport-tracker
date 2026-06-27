package service

import (
	"log"
	"time"

	gocache "github.com/patrickmn/go-cache"

	"github.com/duyb/esport-score-tracker/internal/model"
	"github.com/duyb/esport-score-tracker/internal/repository"
	"github.com/google/uuid"
)

type WcAnalyticsService struct {
	repo      *repository.WcAnalyticsRepository
	wcRepo    *repository.WcRepository
	cache     *gocache.Cache
	fdClient  *WcFootballDataClient
	ofClient  *WcOpenFootballClient
}

func NewWcAnalyticsService(
	repo *repository.WcAnalyticsRepository,
	wcRepo *repository.WcRepository,
	cache *gocache.Cache,
	fdClient *WcFootballDataClient,
	ofClient *WcOpenFootballClient,
) *WcAnalyticsService {
	return &WcAnalyticsService{
		repo:     repo,
		wcRepo:   wcRepo,
		cache:    cache,
		fdClient: fdClient,
		ofClient: ofClient,
	}
}

// --- Tournament Analytics ---

const (
	cacheKeyWC2026       = "wc:analytics:wc2026"
	cacheKeyScorers      = "wc:analytics:scorers"
	cacheKeyOpenFootball = "wc:analytics:openfootball"
	cacheTTLWC2026       = 5 * time.Minute
	cacheTTLScorers      = 30 * time.Minute
	cacheTTLOpenFootball = 15 * time.Minute
)

type scorersCacheEntry struct {
	scorers   []model.WcTournamentScorer
	fetchedAt time.Time
}

func (s *WcAnalyticsService) GetWorldCup2026Analytics() (*model.WcAnalyticsResponse, error) {
	// Tier 1: full response cache — absorbs burst traffic
	if cached, found := s.cache.Get(cacheKeyWC2026); found {
		r := cached.(model.WcAnalyticsResponse)
		return &r, nil
	}

	matchStats, err := s.wcRepo.GetCompletedMatchStats()
	if err != nil {
		return nil, err
	}
	resp := &model.WcAnalyticsResponse{MatchStats: *matchStats}

	// Tier 2: football-data.org scorers (goals + assists + team crest)
	if cached, found := s.cache.Get(cacheKeyScorers); found {
		entry := cached.(scorersCacheEntry)
		resp.TopScorers = entry.scorers
		t := entry.fetchedAt
		resp.ScorersUpdatedAt = &t
	} else {
		scorers, fetchedAt, err := s.fdClient.GetWCScorers(20)
		if err != nil {
			log.Printf("[wc-analytics] football-data.org unavailable: %v", err)
		} else {
			s.cache.Set(cacheKeyScorers, scorersCacheEntry{scorers, fetchedAt}, cacheTTLScorers)
			resp.TopScorers = scorers
			resp.ScorersUpdatedAt = &fetchedAt
		}
	}

	// Tier 3: openfootball goal events
	if cached, found := s.cache.Get(cacheKeyOpenFootball); found {
		ofData := cached.(*WcOpenFootballData)
		applyOpenFootballData(resp, ofData)
	} else {
		ofData, err := s.ofClient.GetWCData()
		if err != nil {
			log.Printf("[wc-analytics] openfootball unavailable: %v", err)
		} else {
			s.cache.Set(cacheKeyOpenFootball, ofData, cacheTTLOpenFootball)
			applyOpenFootballData(resp, ofData)
		}
	}

	s.cache.Set(cacheKeyWC2026, *resp, cacheTTLWC2026)
	return resp, nil
}

func applyOpenFootballData(resp *model.WcAnalyticsResponse, d *WcOpenFootballData) {
	resp.GoalTiming = d.GoalTiming
	resp.HalfTimeStats = &d.HalfTimeStats
	resp.TeamStats = d.TeamStats
	resp.GoalsByGroup = d.GoalsByGroup
	resp.TopScoringMatches = d.TopScoringMatches
	resp.VenueStats = d.VenueStats
}

// --- DTOs ---

type AnalyticsTimelinePoint struct {
	Period   string  `json:"period"`
	Wins     int     `json:"wins"`
	Losses   int     `json:"losses"`
	Accuracy float64 `json:"accuracy"`
}

type AnalyticsCompareMetrics struct {
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

type MyAnalyticsResponse struct {
	Accuracy           float64                  `json:"accuracy"`
	SettledMatches     int                      `json:"settled_matches"`
	Wins               int                      `json:"wins"`
	Losses             int                      `json:"losses"`
	PendingBets        int                      `json:"pending_bets"`
	ProfileLabel       *string                  `json:"profile_label"`
	CurrentWinStreak   int                      `json:"current_win_streak"`
	CurrentLoseStreak  int                      `json:"current_lose_streak"`
	LongestWinStreak   int                      `json:"longest_win_streak"`
	BetTypeDistribution map[string]int          `json:"bet_type_distribution"`
	FavoriteTeams      []repository.TeamCountRow      `json:"favorite_teams"`
	FavoriteScorelines []repository.ScorelineCountRow `json:"favorite_scorelines"`
	AccuracyTimeline   []AnalyticsTimelinePoint `json:"accuracy_timeline"`
	CompareMetrics     AnalyticsCompareMetrics  `json:"compare_metrics"`
}

type CommunityAnalyticsResponse struct {
	TotalBetsPlaced       int                           `json:"total_bets_placed"`
	ActiveUsers           int                           `json:"active_users"`
	AvgAccuracy           float64                       `json:"avg_accuracy"`
	PredictionDistribution map[string]int               `json:"prediction_distribution"`
	TrendingTeams         []repository.TeamCountRow     `json:"trending_teams"`
	TrendingScorelines    []repository.ScorelineCountRow `json:"trending_scorelines"`
	CommunityCompareMetrics AnalyticsCompareMetrics     `json:"community_compare_metrics"`
	TopPredictors         []TopPredictorEntry           `json:"top_predictors"`
}

type TopPredictorEntry struct {
	UserID         string   `json:"user_id"`
	Name           string   `json:"name"`
	AvatarURL      *string  `json:"avatar_url"`
	Accuracy       float64  `json:"accuracy"`
	SettledMatches int      `json:"settled_matches"`
}

type CompareAnalyticsResponse struct {
	Me        AnalyticsCompareMetrics `json:"me"`
	Community AnalyticsCompareMetrics `json:"community"`
	MyAccuracy        float64 `json:"my_accuracy"`
	CommunityAccuracy float64 `json:"community_accuracy"`
}

// --- Period helpers ---

func PeriodFromParam(param string, from, to string) repository.AnalyticsPeriod {
	now := time.Now()
	// Custom date range takes precedence
	if from != "" && to != "" {
		f, errF := time.Parse("2006-01-02", from)
		t, errT := time.Parse("2006-01-02", to)
		if errF == nil && errT == nil {
			return repository.AnalyticsPeriod{From: f, To: t.Add(24*time.Hour - time.Second)}
		}
	}
	switch param {
	case "today":
		start := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
		return repository.AnalyticsPeriod{From: start, To: now}
	case "7d":
		return repository.AnalyticsPeriod{From: now.AddDate(0, 0, -7), To: now}
	case "14d":
		return repository.AnalyticsPeriod{From: now.AddDate(0, 0, -14), To: now}
	default: // "30d"
		return repository.AnalyticsPeriod{From: now.AddDate(0, 0, -30), To: now}
	}
}

// --- Build responses ---

func (s *WcAnalyticsService) BuildMyResponse(userID uuid.UUID, period repository.AnalyticsPeriod) (*MyAnalyticsResponse, error) {
	// 1. Match-level results for accuracy + streak
	matchResults, err := s.repo.GetMyMatchResults(userID, period)
	if err != nil {
		return nil, err
	}

	wins, losses := 0, 0
	for _, m := range matchResults {
		if m.TotalPayout > m.TotalStake {
			wins++
		} else {
			losses++
		}
	}
	settledMatches := wins + losses
	accuracy := 0.0
	if settledMatches > 0 {
		accuracy = float64(wins) / float64(settledMatches)
	}

	// 2. Streak (all-time, not period-filtered — use all results)
	allResults, err := s.repo.GetMyMatchResults(userID, repository.AnalyticsPeriod{
		From: time.Time{},
		To:   time.Now(),
	})
	if err != nil {
		return nil, err
	}
	currentWin, currentLose, longestWin := computeStreaks(allResults)

	// 3. Timeline
	timelineRows, err := s.repo.GetMyAccuracyTimeline(userID, period)
	if err != nil {
		return nil, err
	}
	timeline := make([]AnalyticsTimelinePoint, 0, len(timelineRows))
	for _, r := range timelineRows {
		total := r.Wins + r.Losses
		acc := 0.0
		if total > 0 {
			acc = float64(r.Wins) / float64(total)
		}
		timeline = append(timeline, AnalyticsTimelinePoint{
			Period:   r.Period,
			Wins:     r.Wins,
			Losses:   r.Losses,
			Accuracy: acc,
		})
	}

	// 4. Bet stats
	stats, err := s.repo.GetMyBetStats(userID)
	if err != nil {
		return nil, err
	}

	// 5. Favorite teams + scorelines
	favoriteTeams, err := s.repo.GetMyFavoriteTeams(userID, 5)
	if err != nil {
		return nil, err
	}
	favoriteScorelines, err := s.repo.GetMyFavoriteScorelines(userID, 5)
	if err != nil {
		return nil, err
	}

	// 6. Avg goals
	goalsRow, err := s.repo.GetMyAvgGoals(userID)
	if err != nil {
		return nil, err
	}

	// 7. Compute metrics
	metrics := buildMyMetrics(stats, goalsRow)

	// 8. Profile label (requires ≥ 3 settled matches)
	var profileLabel *string
	if settledMatches >= 3 {
		label := classifyProfile(stats, goalsRow)
		profileLabel = &label
	}

	return &MyAnalyticsResponse{
		Accuracy:        accuracy,
		SettledMatches:  settledMatches,
		Wins:            wins,
		Losses:          losses,
		PendingBets:     stats.PendingBets,
		ProfileLabel:    profileLabel,
		CurrentWinStreak:  currentWin,
		CurrentLoseStreak: currentLose,
		LongestWinStreak:  longestWin,
		BetTypeDistribution: map[string]int{
			"handicap":    stats.HandicapBets,
			"exact_score": stats.ExactScoreBets,
			"over_under":  stats.OverUnderBets,
			"custom":      stats.CustomBets,
		},
		FavoriteTeams:      favoriteTeams,
		FavoriteScorelines: favoriteScorelines,
		AccuracyTimeline:   timeline,
		CompareMetrics:     metrics,
	}, nil
}

func (s *WcAnalyticsService) BuildCommunityResponse() (*CommunityAnalyticsResponse, error) {
	distribution, err := s.repo.GetCommunityBetDistribution()
	if err != nil {
		return nil, err
	}
	trendingTeams, err := s.repo.GetCommunityTrendingTeams(10)
	if err != nil {
		return nil, err
	}
	trendingScorelines, err := s.repo.GetCommunityTrendingScorelines(10)
	if err != nil {
		return nil, err
	}
	totalBets, activeUsers, err := s.repo.GetCommunityTotals()
	if err != nil {
		return nil, err
	}
	predictors, err := s.repo.GetTopPredictors(3)
	if err != nil {
		return nil, err
	}
	communityStats, err := s.repo.GetCommunityAvgStats()
	if err != nil {
		return nil, err
	}

	top := make([]TopPredictorEntry, 0, len(predictors))
	for _, p := range predictors {
		acc := 0.0
		if p.SettledMatches > 0 {
			acc = float64(p.Wins) / float64(p.SettledMatches)
		}
		top = append(top, TopPredictorEntry{
			UserID:         p.WcUserID,
			Name:           p.Name,
			AvatarURL:      p.AvatarURL,
			Accuracy:       acc,
			SettledMatches: p.SettledMatches,
		})
	}

	return &CommunityAnalyticsResponse{
		TotalBetsPlaced:        totalBets,
		ActiveUsers:            activeUsers,
		AvgAccuracy:            communityStats.AvgAccuracy,
		PredictionDistribution: distribution,
		TrendingTeams:          trendingTeams,
		TrendingScorelines:     trendingScorelines,
		CommunityCompareMetrics: communityMetrics(communityStats),
		TopPredictors:          top,
	}, nil
}

func (s *WcAnalyticsService) BuildCompareResponse(userID uuid.UUID) (*CompareAnalyticsResponse, error) {
	// Reuse my response (all-time, 30d default)
	period := repository.AnalyticsPeriod{From: time.Now().AddDate(0, 0, -30), To: time.Now()}
	my, err := s.BuildMyResponse(userID, period)
	if err != nil {
		return nil, err
	}

	communityStats, err := s.repo.GetCommunityAvgStats()
	if err != nil {
		return nil, err
	}

	return &CompareAnalyticsResponse{
		Me:                my.CompareMetrics,
		Community:         communityMetrics(communityStats),
		MyAccuracy:        my.Accuracy,
		CommunityAccuracy: communityStats.AvgAccuracy,
	}, nil
}

// --- Private helpers ---

func computeStreaks(results []repository.MatchResultRow) (currentWin, currentLose, longestWin int) {
	// results are ordered ASC — walk in reverse for "current" streak
	n := len(results)
	if n == 0 {
		return 0, 0, 0
	}

	// Current streaks: walk newest-first
	inWin, inLose := true, true
	for i := n - 1; i >= 0; i-- {
		isWin := results[i].TotalPayout > results[i].TotalStake
		if inWin {
			if isWin {
				currentWin++
			} else {
				inWin = false
			}
		}
		if inLose {
			if !isWin {
				currentLose++
			} else {
				inLose = false
			}
		}
		if !inWin && !inLose {
			break
		}
	}

	// Longest win streak: full pass
	streak := 0
	for _, r := range results {
		if r.TotalPayout > r.TotalStake {
			streak++
			if streak > longestWin {
				longestWin = streak
			}
		} else {
			streak = 0
		}
	}
	return currentWin, currentLose, longestWin
}

func buildMyMetrics(stats *repository.MyBetStatsRow, goalsRow *repository.AvgGoalsRow) AnalyticsCompareMetrics {
	var homeBias *float64
	if stats.HandicapBets > 0 {
		v := float64(stats.HomePicks) / float64(stats.HandicapBets)
		homeBias = &v
	}

	var avgGoals *float64
	if goalsRow.Count > 0 {
		v := float64(goalsRow.TotalGoals) / float64(goalsRow.Count)
		avgGoals = &v
	}

	exactScoreRate := 0.0
	if stats.TotalBets > 0 {
		exactScoreRate = float64(stats.ExactScoreBets) / float64(stats.TotalBets)
	}

	var underdogRate *float64
	if stats.HandicapBets > 0 {
		v := float64(stats.AwayPicks) / float64(stats.HandicapBets)
		underdogRate = &v
	}

	var overPref *float64
	if stats.OverUnderBets > 0 {
		v := float64(stats.OverPicks) / float64(stats.OverUnderBets)
		overPref = &v
	}

	var exactHit *float64
	if stats.ExactScoreSettled > 0 {
		v := float64(stats.ExactScoreWins) / float64(stats.ExactScoreSettled)
		exactHit = &v
	}

	betFreq := 0.0 // populated from outer context if needed; zero here

	lastMinRate := 0.0
	if stats.TotalBets > 0 {
		lastMinRate = float64(stats.LastMinuteBets) / float64(stats.TotalBets)
	}

	return AnalyticsCompareMetrics{
		HomeBias:           homeBias,
		AvgGoalsPredicted:  avgGoals,
		ExactScoreRate:     exactScoreRate,
		UnderdogRate:       underdogRate,
		AvgStake:           stats.AvgStake,
		OverPreferenceRate: overPref,
		ExactScoreHitRate:  exactHit,
		BetFrequency:       betFreq,
		LastMinuteRate:     lastMinRate,
	}
}

func communityMetrics(s *repository.CommunityAvgStats) AnalyticsCompareMetrics {
	return AnalyticsCompareMetrics{
		HomeBias:           s.HomeBias,
		AvgGoalsPredicted:  s.AvgGoalsPredicted,
		ExactScoreRate:     s.ExactScoreRate,
		UnderdogRate:       s.UnderdogRate,
		AvgStake:           s.AvgStake,
		OverPreferenceRate: s.OverPreferenceRate,
		ExactScoreHitRate:  s.ExactScoreHitRate,
		BetFrequency:       s.BetFrequency,
		LastMinuteRate:     s.LastMinuteRate,
	}
}

func classifyProfile(stats *repository.MyBetStatsRow, goalsRow *repository.AvgGoalsRow) string {
	if stats.HandicapBets > 0 {
		awayRate := float64(stats.AwayPicks) / float64(stats.HandicapBets)
		if awayRate > 0.60 {
			return "underdog_lover"
		}
	}
	if goalsRow.Count > 0 {
		avg := float64(goalsRow.TotalGoals) / float64(goalsRow.Count)
		if avg > 3.0 {
			return "goal_hunter"
		}
	}
	if stats.TotalBets > 0 {
		if float64(stats.ExactScoreBets)/float64(stats.TotalBets) > 0.50 {
			return "aggressive_predictor"
		}
		if float64(stats.HandicapBets)/float64(stats.TotalBets) > 0.60 {
			return "conservative_predictor"
		}
	}
	return "balanced_predictor"
}
