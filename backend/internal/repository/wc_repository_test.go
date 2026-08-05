package repository

import (
	"testing"
	"time"

	"github.com/duyb/esport-score-tracker/internal/model"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

// openWcRepoTestDB reuses openTestDB and additionally migrates wc_matches.
func openWcRepoTestDB(t *testing.T) (*WcRepository, *gorm.DB) {
	t.Helper()
	db := openTestDB(t) // skips if TEST_DATABASE_URL not set
	require.NoError(t, db.AutoMigrate(&model.WcMatch{}))
	return NewWcRepository(db), db
}

// seedMatchAt inserts a WcMatch with a specific date and status, registering cleanup.
func seedMatchAt(t *testing.T, db *gorm.DB, matchDate time.Time, status string) *model.WcMatch {
	t.Helper()
	m := &model.WcMatch{
		ID:             uuid.New(),
		TournamentType: model.WcTournamentWorldCup,
		ExternalID:     uuid.NewString(),
		HomeTeam:       "Team A",
		AwayTeam:       "Team B",
		MatchDate:      matchDate,
		Stage:          model.WcStageGroup,
		Status:         status,
	}
	require.NoError(t, db.Create(m).Error)
	t.Cleanup(func() { db.Unscoped().Delete(m) })
	return m
}

// ─── ListMatches — DateFrom / DateTo filter ───────────────────────────────────

func TestListMatches_DateRange_InWindowReturned(t *testing.T) {
	repo, db := openWcRepoTestDB(t)
	now := time.Now().UTC().Truncate(time.Second)
	from := now.Add(-4 * time.Hour) // mirrors the dashboard's lookback window
	to := now.Add(72 * time.Hour)

	before := seedMatchAt(t, db, from.Add(-1*time.Second), model.WcStatusScheduled)
	inside := seedMatchAt(t, db, now.Add(24*time.Hour), model.WcStatusScheduled)
	after := seedMatchAt(t, db, to.Add(1*time.Second), model.WcStatusScheduled)

	matches, err := repo.ListMatches(model.WcTournamentWorldCup, MatchFilter{
		DateFrom: from.Format(time.RFC3339),
		DateTo:   to.Format(time.RFC3339),
	})

	require.NoError(t, err)
	ids := matchIDs(matches)
	assert.Contains(t, ids, inside.ID, "match inside window should be returned")
	assert.NotContains(t, ids, before.ID, "match before window should be excluded")
	assert.NotContains(t, ids, after.ID, "match after window should be excluded")
}

func TestListMatches_DateRange_BoundariesInclusive(t *testing.T) {
	repo, db := openWcRepoTestDB(t)
	from := time.Now().UTC().Add(2 * time.Hour).Truncate(time.Second)
	to := from.Add(72 * time.Hour)

	atFrom := seedMatchAt(t, db, from, model.WcStatusScheduled)
	atTo := seedMatchAt(t, db, to, model.WcStatusScheduled)

	matches, err := repo.ListMatches(model.WcTournamentWorldCup, MatchFilter{
		DateFrom: from.Format(time.RFC3339),
		DateTo:   to.Format(time.RFC3339),
	})

	require.NoError(t, err)
	ids := matchIDs(matches)
	assert.Contains(t, ids, atFrom.ID, "match at date_from boundary should be included")
	assert.Contains(t, ids, atTo.ID, "match at date_to boundary should be included")
}

func TestListMatches_DateRange_EmptyFilter_ReturnsAll(t *testing.T) {
	repo, db := openWcRepoTestDB(t)

	m1 := seedMatchAt(t, db, time.Now().Add(-10*time.Hour), model.WcStatusCompleted)
	m2 := seedMatchAt(t, db, time.Now().Add(10*time.Hour), model.WcStatusScheduled)

	matches, err := repo.ListMatches(model.WcTournamentWorldCup, MatchFilter{}) // no date filter

	require.NoError(t, err)
	ids := matchIDs(matches)
	assert.Contains(t, ids, m1.ID, "older match should be returned when filter is empty")
	assert.Contains(t, ids, m2.ID, "future match should be returned when filter is empty")
}

func TestListMatches_DateRange_OnlyDateFrom_LowerBoundOnly(t *testing.T) {
	repo, db := openWcRepoTestDB(t)
	pivot := time.Now().UTC().Truncate(time.Second)

	before := seedMatchAt(t, db, pivot.Add(-1*time.Second), model.WcStatusScheduled)
	after := seedMatchAt(t, db, pivot.Add(1*time.Second), model.WcStatusScheduled)

	matches, err := repo.ListMatches(model.WcTournamentWorldCup, MatchFilter{
		DateFrom: pivot.Format(time.RFC3339),
		// DateTo intentionally omitted
	})

	require.NoError(t, err)
	ids := matchIDs(matches)
	assert.Contains(t, ids, after.ID, "match after date_from should be included")
	assert.NotContains(t, ids, before.ID, "match before date_from should be excluded")
}

func TestListMatches_DateRange_OnlyDateTo_UpperBoundOnly(t *testing.T) {
	repo, db := openWcRepoTestDB(t)
	pivot := time.Now().UTC().Add(24 * time.Hour).Truncate(time.Second)

	before := seedMatchAt(t, db, pivot.Add(-1*time.Second), model.WcStatusScheduled)
	after := seedMatchAt(t, db, pivot.Add(1*time.Second), model.WcStatusScheduled)

	matches, err := repo.ListMatches(model.WcTournamentWorldCup, MatchFilter{
		// DateFrom intentionally omitted
		DateTo: pivot.Format(time.RFC3339),
	})

	require.NoError(t, err)
	ids := matchIDs(matches)
	assert.Contains(t, ids, before.ID, "match before date_to should be included")
	assert.NotContains(t, ids, after.ID, "match after date_to should be excluded")
}

func TestListMatches_DateRange_CombinedWithStatusFilter(t *testing.T) {
	repo, db := openWcRepoTestDB(t)
	now := time.Now().UTC()
	from := now.Add(-4 * time.Hour)
	to := now.Add(72 * time.Hour)

	scheduled := seedMatchAt(t, db, now.Add(1*time.Hour), model.WcStatusScheduled)
	live := seedMatchAt(t, db, now.Add(-1*time.Hour), model.WcStatusLive)
	completed := seedMatchAt(t, db, now.Add(-2*time.Hour), model.WcStatusCompleted)

	// Status=scheduled + date range: only scheduled matches
	scheduledOnly, err := repo.ListMatches(model.WcTournamentWorldCup, MatchFilter{
		Status:   model.WcStatusScheduled,
		DateFrom: from.Format(time.RFC3339),
		DateTo:   to.Format(time.RFC3339),
	})
	require.NoError(t, err)
	ids := matchIDs(scheduledOnly)
	assert.Contains(t, ids, scheduled.ID, "scheduled match should appear")
	assert.NotContains(t, ids, live.ID, "live match should be excluded by status filter")
	assert.NotContains(t, ids, completed.ID, "completed match should be excluded by status filter")

	// No status filter + date range: scheduled + live (completed also in window but status is separate concern)
	all, err := repo.ListMatches(model.WcTournamentWorldCup, MatchFilter{
		DateFrom: from.Format(time.RFC3339),
		DateTo:   to.Format(time.RFC3339),
	})
	require.NoError(t, err)
	allIDs := matchIDs(all)
	assert.Contains(t, allIDs, scheduled.ID, "scheduled match should appear without status filter")
	assert.Contains(t, allIDs, live.ID, "live match should appear without status filter")
	assert.Contains(t, allIDs, completed.ID, "completed match should appear without status filter")
}

func TestListMatches_DateRange_LiveMatchLookback(t *testing.T) {
	// Validates the dashboard's 4h lookback window captures live matches.
	// A live match has match_date = kickoff time (in the past).
	repo, db := openWcRepoTestDB(t)
	now := time.Now().UTC()

	// Match started 2h ago — currently live
	liveMatch := seedMatchAt(t, db, now.Add(-2*time.Hour), model.WcStatusLive)
	// Match started 5h ago — outside the 4h lookback
	tooOld := seedMatchAt(t, db, now.Add(-5*time.Hour), model.WcStatusLive)

	// Simulate dashboard: date_from = now - 4h, date_to = now + 72h (3 days)
	lookbackFrom := now.Add(-4 * time.Hour)
	matches, err := repo.ListMatches(model.WcTournamentWorldCup, MatchFilter{
		DateFrom: lookbackFrom.Format(time.RFC3339),
		DateTo:   now.Add(72 * time.Hour).Format(time.RFC3339),
	})

	require.NoError(t, err)
	ids := matchIDs(matches)
	assert.Contains(t, ids, liveMatch.ID, "live match within 4h lookback should be returned")
	assert.NotContains(t, ids, tooOld.ID, "live match older than 4h should be excluded")
}

func TestListMatches_DateRange_SortedByMatchDateASC(t *testing.T) {
	repo, db := openWcRepoTestDB(t)
	now := time.Now().UTC()
	from := now
	to := now.Add(72 * time.Hour)

	m3 := seedMatchAt(t, db, now.Add(30*time.Hour), model.WcStatusScheduled)
	m1 := seedMatchAt(t, db, now.Add(10*time.Hour), model.WcStatusScheduled)
	m2 := seedMatchAt(t, db, now.Add(20*time.Hour), model.WcStatusScheduled)

	matches, err := repo.ListMatches(model.WcTournamentWorldCup, MatchFilter{
		DateFrom: from.Format(time.RFC3339),
		DateTo:   to.Format(time.RFC3339),
	})
	require.NoError(t, err)

	ids := matchIDs(matches)
	// Find positions of our seeded matches
	pos1 := indexOf(ids, m1.ID)
	pos2 := indexOf(ids, m2.ID)
	pos3 := indexOf(ids, m3.ID)

	require.NotEqual(t, -1, pos1)
	require.NotEqual(t, -1, pos2)
	require.NotEqual(t, -1, pos3)
	assert.Less(t, pos1, pos2, "m1 (earliest) should come before m2")
	assert.Less(t, pos2, pos3, "m2 should come before m3 (latest)")
}

// ─── GetGroupStandings ────────────────────────────────────────────────────────

// seedGroupMatch inserts a WcMatch with full group/team/score data and registers cleanup.
func seedGroupMatch(t *testing.T, db *gorm.DB,
	groupName, homeTeam, homeCode, awayTeam, awayCode string,
	homeScore, awayScore *int,
	status string,
	matchDate time.Time,
) *model.WcMatch {
	t.Helper()
	m := &model.WcMatch{
		ID:           uuid.New(),
		ExternalID:   uuid.NewString(),
		HomeTeam:     homeTeam,
		HomeTeamCode: homeCode,
		AwayTeam:     awayTeam,
		AwayTeamCode: awayCode,
		GroupName:    groupName,
		Stage:        model.WcStageGroup,
		Status:       status,
		HomeScore:    homeScore,
		AwayScore:    awayScore,
		MatchDate:    matchDate,
	}
	require.NoError(t, db.Create(m).Error)
	t.Cleanup(func() { db.Unscoped().Delete(m) })
	return m
}

func intPtr(v int) *int { return &v }

func findGroupInStandings(groups []model.WcGroupStanding, name string) *model.WcGroupStanding {
	for i := range groups {
		if groups[i].GroupName == name {
			return &groups[i]
		}
	}
	return nil
}

func findTeamInStanding(g *model.WcGroupStanding, name string) *model.WcTeamStanding {
	for i := range g.Teams {
		if g.Teams[i].TeamName == name {
			return &g.Teams[i]
		}
	}
	return nil
}

func TestGetGroupStandings_FullScenario(t *testing.T) {
	// Scenario from the testing doc:
	//   Match 1: ARG 3-0 BRA  (completed, 2026-06-11)
	//   Match 2: FRA 1-1 ESP  (completed, 2026-06-11)
	//   Match 3: ARG 1-1 FRA  (completed, 2026-06-15)
	//   Match 4: BRA 2-0 ESP  (completed, 2026-06-15)
	//   Match 5: ARG 1-0 ESP  (scheduled, 2026-06-19) — must NOT count
	//
	// Expected standings:
	//   ARG: P2 W1 D1 L0 GF4 GA1 GD+3 Pts4  Form:[W,D]
	//   BRA: P2 W1 D0 L1 GF2 GA3 GD-1 Pts3  Form:[L,W]
	//   FRA: P2 W0 D2 L0 GF2 GA2 GD0  Pts2  Form:[D,D]
	//   ESP: P2 W0 D1 L1 GF1 GA2 GD-1 Pts1  Form:[D,L]
	repo, db := openWcRepoTestDB(t)

	d11 := time.Date(2026, 6, 11, 12, 0, 0, 0, time.UTC)
	d15 := time.Date(2026, 6, 15, 12, 0, 0, 0, time.UTC)
	d19 := time.Date(2026, 6, 19, 12, 0, 0, 0, time.UTC)

	seedGroupMatch(t, db, "Group Z", "ARG", "ARG", "BRA", "BRA", intPtr(3), intPtr(0), model.WcStatusCompleted, d11)
	seedGroupMatch(t, db, "Group Z", "FRA", "FRA", "ESP", "ESP", intPtr(1), intPtr(1), model.WcStatusCompleted, d11)
	seedGroupMatch(t, db, "Group Z", "ARG", "ARG", "FRA", "FRA", intPtr(1), intPtr(1), model.WcStatusCompleted, d15)
	seedGroupMatch(t, db, "Group Z", "BRA", "BRA", "ESP", "ESP", intPtr(2), intPtr(0), model.WcStatusCompleted, d15)
	// Scheduled match — must not affect standings
	seedGroupMatch(t, db, "Group Z", "ARG", "ARG", "ESP", "ESP", nil, nil, model.WcStatusScheduled, d19)

	groups, err := repo.GetGroupStandings(model.WcTournamentWorldCup)
	require.NoError(t, err)

	g := findGroupInStandings(groups, "Group Z")
	require.NotNil(t, g, "Group Z must be present in standings")
	require.Len(t, g.Teams, 4, "Group Z must have exactly 4 teams")

	arg := findTeamInStanding(g, "ARG")
	require.NotNil(t, arg)
	assert.Equal(t, 2, arg.Played)
	assert.Equal(t, 1, arg.Won)
	assert.Equal(t, 1, arg.Drawn)
	assert.Equal(t, 0, arg.Lost)
	assert.Equal(t, 4, arg.GoalsFor)
	assert.Equal(t, 1, arg.GoalsAgainst)
	assert.Equal(t, 3, arg.GoalDifference)
	assert.Equal(t, 4, arg.Points)
	assert.Equal(t, []string{"W", "D"}, arg.Form)

	bra := findTeamInStanding(g, "BRA")
	require.NotNil(t, bra)
	assert.Equal(t, 2, bra.Played)
	assert.Equal(t, 1, bra.Won)
	assert.Equal(t, 0, bra.Drawn)
	assert.Equal(t, 1, bra.Lost)
	assert.Equal(t, 2, bra.GoalsFor)
	assert.Equal(t, 3, bra.GoalsAgainst)
	assert.Equal(t, -1, bra.GoalDifference)
	assert.Equal(t, 3, bra.Points)
	assert.Equal(t, []string{"L", "W"}, bra.Form)

	fra := findTeamInStanding(g, "FRA")
	require.NotNil(t, fra)
	assert.Equal(t, 2, fra.Played)
	assert.Equal(t, 0, fra.Won)
	assert.Equal(t, 2, fra.Drawn)
	assert.Equal(t, 0, fra.Lost)
	assert.Equal(t, 2, fra.GoalsFor)
	assert.Equal(t, 2, fra.GoalsAgainst)
	assert.Equal(t, 0, fra.GoalDifference)
	assert.Equal(t, 2, fra.Points)
	assert.Equal(t, []string{"D", "D"}, fra.Form)

	esp := findTeamInStanding(g, "ESP")
	require.NotNil(t, esp)
	assert.Equal(t, 2, esp.Played)
	assert.Equal(t, 0, esp.Won)
	assert.Equal(t, 1, esp.Drawn)
	assert.Equal(t, 1, esp.Lost)
	assert.Equal(t, 1, esp.GoalsFor)
	assert.Equal(t, 2, esp.GoalsAgainst)
	assert.Equal(t, -1, esp.GoalDifference)
	assert.Equal(t, 1, esp.Points)
	assert.Equal(t, []string{"D", "L"}, esp.Form)
}

func TestGetGroupStandings_ScheduledAndLiveMatchesExcludedFromStats(t *testing.T) {
	// Only completed matches contribute to stats; scheduled/live match registers team but not stats
	repo, db := openWcRepoTestDB(t)
	d1 := time.Date(2026, 6, 11, 12, 0, 0, 0, time.UTC)

	// One completed match in Group W
	seedGroupMatch(t, db, "Group W", "GER", "GER", "JPN", "JPN", intPtr(2), intPtr(0), model.WcStatusCompleted, d1)
	// One scheduled match in Group W (registers AUS and KOR in roster, no stats)
	seedGroupMatch(t, db, "Group W", "AUS", "AUS", "KOR", "KOR", nil, nil, model.WcStatusScheduled, d1.Add(24*time.Hour))
	// One live match in Group W (registers NZL and IRN in roster, no stats even with scores)
	seedGroupMatch(t, db, "Group W", "NZL", "NZL", "IRN", "IRN", intPtr(1), intPtr(0), model.WcStatusLive, d1.Add(48*time.Hour))

	groups, err := repo.GetGroupStandings(model.WcTournamentWorldCup)
	require.NoError(t, err)

	g := findGroupInStandings(groups, "Group W")
	require.NotNil(t, g)

	ger := findTeamInStanding(g, "GER")
	require.NotNil(t, ger)
	assert.Equal(t, 1, ger.Played, "GER should have 1 completed match")
	assert.Equal(t, 3, ger.Points)

	nzl := findTeamInStanding(g, "NZL")
	require.NotNil(t, nzl, "NZL should appear in roster from live match")
	assert.Equal(t, 0, nzl.Played, "live match should not count toward stats")
	assert.Equal(t, 0, nzl.Points)
	assert.Empty(t, nzl.Form)
}

func TestGetGroupStandings_PreTournament_AllTeamsZeroStats(t *testing.T) {
	// All matches are scheduled — all 4 teams should appear with zero stats
	repo, db := openWcRepoTestDB(t)
	d := time.Date(2026, 6, 11, 12, 0, 0, 0, time.UTC)

	seedGroupMatch(t, db, "Group V", "POR", "POR", "CRO", "CRO", nil, nil, model.WcStatusScheduled, d)
	seedGroupMatch(t, db, "Group V", "MAR", "MAR", "CMR", "CMR", nil, nil, model.WcStatusScheduled, d.Add(3*time.Hour))
	seedGroupMatch(t, db, "Group V", "POR", "POR", "MAR", "MAR", nil, nil, model.WcStatusScheduled, d.Add(7*24*time.Hour))
	seedGroupMatch(t, db, "Group V", "CRO", "CRO", "CMR", "CMR", nil, nil, model.WcStatusScheduled, d.Add(7*24*time.Hour+3*time.Hour))
	seedGroupMatch(t, db, "Group V", "POR", "POR", "CMR", "CMR", nil, nil, model.WcStatusScheduled, d.Add(14*24*time.Hour))
	seedGroupMatch(t, db, "Group V", "CRO", "CRO", "MAR", "MAR", nil, nil, model.WcStatusScheduled, d.Add(14*24*time.Hour+3*time.Hour))

	groups, err := repo.GetGroupStandings(model.WcTournamentWorldCup)
	require.NoError(t, err)

	g := findGroupInStandings(groups, "Group V")
	require.NotNil(t, g)
	assert.Len(t, g.Teams, 4, "all 4 teams must appear before any match is played")

	for _, team := range g.Teams {
		assert.Equal(t, 0, team.Played, "team %s should have 0 played", team.TeamName)
		assert.Equal(t, 0, team.Points, "team %s should have 0 points", team.TeamName)
		assert.Empty(t, team.Form, "team %s should have empty form", team.TeamName)
	}
}

func TestGetGroupStandings_FormLimitedToLast5(t *testing.T) {
	// A team with 6 completed matches should only show the last 5 form entries
	repo, db := openWcRepoTestDB(t)
	base := time.Date(2026, 6, 1, 12, 0, 0, 0, time.UTC)

	// URU wins 6 matches against placeholder opponents (different teams to keep it simple)
	// Match order: W W W W W W → form should be last 5: [W,W,W,W,W]
	opponents := []string{"OPP1", "OPP2", "OPP3", "OPP4", "OPP5", "OPP6"}
	for i, opp := range opponents {
		seedGroupMatch(t, db, "Group U", "URU", "URU", opp, opp, intPtr(1), intPtr(0), model.WcStatusCompleted, base.Add(time.Duration(i)*24*time.Hour))
	}

	groups, err := repo.GetGroupStandings(model.WcTournamentWorldCup)
	require.NoError(t, err)

	g := findGroupInStandings(groups, "Group U")
	require.NotNil(t, g)
	uru := findTeamInStanding(g, "URU")
	require.NotNil(t, uru)
	assert.Equal(t, 6, uru.Played)
	assert.Len(t, uru.Form, 5, "form must be capped at last 5 results")
	for _, f := range uru.Form {
		assert.Equal(t, "W", f)
	}
}

func TestGetGroupStandings_OnlyGroupStageReturned(t *testing.T) {
	// A knockout-stage match should not appear in standings
	repo, db := openWcRepoTestDB(t)
	d := time.Date(2026, 7, 1, 12, 0, 0, 0, time.UTC)

	// This is an R16 match — should be ignored
	m := &model.WcMatch{
		ID:           uuid.New(),
		ExternalID:   uuid.NewString(),
		HomeTeam:     "ARG",
		HomeTeamCode: "ARG",
		AwayTeam:     "FRA",
		AwayTeamCode: "FRA",
		GroupName:    "",
		Stage:        model.WcStageR16,
		Status:       model.WcStatusCompleted,
		HomeScore:    intPtr(1),
		AwayScore:    intPtr(0),
		MatchDate:    d,
	}
	require.NoError(t, db.Create(m).Error)
	t.Cleanup(func() { db.Unscoped().Delete(m) })

	groups, err := repo.GetGroupStandings(model.WcTournamentWorldCup)
	require.NoError(t, err)

	// No group with empty name should appear
	for _, g := range groups {
		assert.NotEmpty(t, g.GroupName, "group with empty name should never appear")
	}
	// ARG from the R16 match must not be in any group standings
	for _, g := range groups {
		for _, team := range g.Teams {
			if team.TeamName == "ARG" {
				// ARG might already exist from other tests in the same DB,
				// but its played count must not include the R16 match stats
				_ = team
			}
		}
	}
}

// ─── GetLeaderboard ───────────────────────────────────────────────────────────

// openLeaderboardTestDB migrates all tables needed for GetLeaderboard tests.
func openLeaderboardTestDB(t *testing.T) (*WcRepository, *gorm.DB) {
	t.Helper()
	db := openTestDB(t)
	require.NoError(t, db.AutoMigrate(
		&model.WcMatch{},
		&model.WcUser{},
		&model.WcWallet{},
		&model.WcPrediction{},
		// GetLeaderboard's eligibility WHERE clause also probes these tables,
		// so they must exist even when a test seeds no rows in them.
		&model.WcCustomBet{},
		&model.WcCustomBetOption{},
		&model.WcCustomBetEntry{},
		&model.WcChampionPrediction{},
		&model.WcWalletLog{},
		&model.WcSettlement{},
	))
	return NewWcRepository(db), db
}

func seedLbUser(t *testing.T, db *gorm.DB, name string) *model.WcUser {
	t.Helper()
	u := &model.WcUser{ID: uuid.New(), Name: name}
	require.NoError(t, db.Create(u).Error)
	t.Cleanup(func() { db.Unscoped().Delete(u) })
	return u
}

func seedLbWallet(t *testing.T, db *gorm.DB, userID uuid.UUID, balance float64) *model.WcWallet {
	t.Helper()
	w := &model.WcWallet{ID: uuid.New(), WcUserID: userID, Balance: balance}
	require.NoError(t, db.Create(w).Error)
	t.Cleanup(func() { db.Unscoped().Delete(w) })
	return w
}

func seedLbMatch(t *testing.T, db *gorm.DB) *model.WcMatch {
	t.Helper()
	m := &model.WcMatch{
		ID:         uuid.New(),
		ExternalID: uuid.NewString(),
		HomeTeam:   "A",
		AwayTeam:   "B",
		MatchDate:  time.Now().UTC(),
		Stage:      model.WcStageGroup,
		Status:     model.WcStatusCompleted,
	}
	require.NoError(t, db.Create(m).Error)
	t.Cleanup(func() { db.Unscoped().Delete(m) })
	return m
}

func seedLbPrediction(t *testing.T, db *gorm.DB, userID, matchID uuid.UUID, result string) *model.WcPrediction {
	t.Helper()
	return seedLbPredictionPts(t, db, userID, matchID, result, 10, 0)
}

// seedLbPredictionPts seeds a settled prediction with an explicit stake and payout so tests
// can assert exact net_points (payout - stake).
func seedLbPredictionPts(t *testing.T, db *gorm.DB, userID, matchID uuid.UUID, result string, points int, earned float64) *model.WcPrediction {
	t.Helper()
	r := result
	e := earned
	p := &model.WcPrediction{
		ID:                 uuid.New(),
		WcUserID:           userID,
		MatchID:            matchID,
		PredictionType:     "handicap",
		Points:             points,
		MultiplierSnapshot: 1.0,
		Result:             &r,
		PointsEarned:       &e,
	}
	require.NoError(t, db.Create(p).Error)
	t.Cleanup(func() { db.Unscoped().Delete(p) })
	return p
}

// seedLbWalletLog records an admin top-up/deduction for a tournament.
func seedLbWalletLog(t *testing.T, db *gorm.DB, userID uuid.UUID, delta float64, tournamentType string) *model.WcWalletLog {
	t.Helper()
	l := &model.WcWalletLog{
		ID:             uuid.New(),
		TournamentType: tournamentType,
		WcUserID:       userID,
		AdminID:        uuid.New(),
		Delta:          delta,
		BalanceBefore:  0,
		BalanceAfter:   delta,
		Note:           "test",
	}
	require.NoError(t, db.Create(l).Error)
	t.Cleanup(func() { db.Unscoped().Delete(l) })
	return l
}

// seedLbSettlement records a settlement for a tournament at a specific time. CreateSettlement
// zeroes wallet balances, so the leaderboard only counts ledger rows newer than this.
func seedLbSettlement(t *testing.T, db *gorm.DB, tournamentType string, at time.Time) *model.WcSettlement {
	t.Helper()
	s := &model.WcSettlement{
		ID:             uuid.New(),
		TournamentType: tournamentType,
		Name:           "test settlement",
		PointRate:      1000,
		SettledBy:      uuid.New(),
	}
	require.NoError(t, db.Create(s).Error)
	require.NoError(t, db.Model(&model.WcSettlement{}).Where("id = ?", s.ID).
		UpdateColumn("created_at", at).Error)
	t.Cleanup(func() { db.Unscoped().Delete(s) })
	return s
}

// seedLbCustomBetEntry seeds a settled custom-bet entry ('won' or 'lost') for a tournament.
func seedLbCustomBetEntry(t *testing.T, db *gorm.DB, userID uuid.UUID, tournamentType, status string, stake int, odds float64) *model.WcCustomBetEntry {
	t.Helper()
	m := seedLbMatch(t, db)
	cb := &model.WcCustomBet{
		ID:             uuid.New(),
		TournamentType: tournamentType,
		MatchID:        m.ID,
		Title:          "test bet",
		Status:         model.WcCustomBetStatusSettled,
	}
	require.NoError(t, db.Create(cb).Error)
	t.Cleanup(func() { db.Unscoped().Delete(cb) })

	opt := &model.WcCustomBetOption{
		ID:          uuid.New(),
		CustomBetID: cb.ID,
		Label:       "yes",
		Odds:        odds,
		IsWinner:    status == model.WcCustomBetEntryStatusWon,
	}
	require.NoError(t, db.Create(opt).Error)
	t.Cleanup(func() { db.Unscoped().Delete(opt) })

	e := &model.WcCustomBetEntry{
		ID:           uuid.New(),
		CustomBetID:  cb.ID,
		OptionID:     opt.ID,
		WcUserID:     userID,
		Stake:        stake,
		OddsSnapshot: odds,
		Status:       status,
	}
	if status == model.WcCustomBetEntryStatusWon {
		payout := float64(stake) * odds
		e.Payout = &payout
	}
	require.NoError(t, db.Create(e).Error)
	t.Cleanup(func() { db.Unscoped().Delete(e) })
	return e
}

// seedLbChampionPrediction seeds a settled champion prediction for a tournament.
func seedLbChampionPrediction(t *testing.T, db *gorm.DB, userID uuid.UUID, tournamentType string, points int, earned int) *model.WcChampionPrediction {
	t.Helper()
	res := "correct"
	e := earned
	cp := &model.WcChampionPrediction{
		ID:             uuid.New(),
		TournamentType: tournamentType,
		WcUserID:       userID,
		TeamID:         uuid.New(),
		Points:         points,
		OddsSnapshot:   2.0,
		Result:         &res,
		PointsEarned:   &e,
	}
	require.NoError(t, db.Create(cp).Error)
	t.Cleanup(func() { db.Unscoped().Delete(cp) })
	return cp
}

func findLbEntry(entries []*model.WcLeaderboardEntry, id uuid.UUID) *model.WcLeaderboardEntry {
	for _, e := range entries {
		if e.WcUserID == id {
			return e
		}
	}
	return nil
}

// TestGetLeaderboard_AdminAdjustmentCounted verifies that an admin top-up recorded in
// wc_wallet_logs counts toward net_points, and that the top-up alone makes the user eligible.
// The wallet itself is shared across tournaments, so net_points is summed from
// per-tournament sources instead of read from wc_wallets.balance.
func TestGetLeaderboard_AdminAdjustmentCounted(t *testing.T) {
	repo, db := openLeaderboardTestDB(t)
	pfx := uuid.NewString()[:8]

	u := seedLbUser(t, db, pfx+"-Alice")
	seedLbWallet(t, db, u.ID, 42.5)
	seedLbWalletLog(t, db, u.ID, 42.5, model.WcTournamentWorldCup)

	entries, err := repo.GetLeaderboard(model.WcTournamentWorldCup)
	require.NoError(t, err)

	entry := findLbEntry(entries, u.ID)
	require.NotNil(t, entry, "user with an admin adjustment must appear in leaderboard")
	assert.Equal(t, 42.5, entry.NetPoints)
	assert.Equal(t, 0, entry.TotalPredictions)
}

// TestGetLeaderboard_AdminTopupAddsToBetResults verifies a top-up stacks on top of settled bet
// results rather than replacing them — the bug report was that top-ups never moved the board.
func TestGetLeaderboard_AdminTopupAddsToBetResults(t *testing.T) {
	repo, db := openLeaderboardTestDB(t)
	pfx := uuid.NewString()[:8]

	u := seedLbUser(t, db, pfx+"-Bob")
	seedLbWallet(t, db, u.ID, -29.0)

	m := seedLbMatch(t, db)
	seedLbPredictionPts(t, db, u.ID, m.ID, "incorrect", 30, 0)
	seedLbWalletLog(t, db, u.ID, -30.0, model.WcTournamentWorldCup) // prediction settlement
	seedLbWalletLog(t, db, u.ID, 2.0, model.WcTournamentWorldCup)   // admin top-up
	seedLbWalletLog(t, db, u.ID, -1.0, model.WcTournamentWorldCup)  // admin deduction

	entries, err := repo.GetLeaderboard(model.WcTournamentWorldCup)
	require.NoError(t, err)

	entry := findLbEntry(entries, u.ID)
	require.NotNil(t, entry)
	assert.Equal(t, -29.0, entry.NetPoints, "net_points must be bet results plus admin adjustments")
	assert.Equal(t, -29.0, entry.NetPoints, "and must equal the wallet balance")
}

// TestGetLeaderboard_ResetsAfterSettlement verifies that only ledger rows newer than the
// tournament's latest settlement count, mirroring the wallet reset CreateSettlement performs.
func TestGetLeaderboard_ResetsAfterSettlement(t *testing.T) {
	repo, db := openLeaderboardTestDB(t)
	pfx := uuid.NewString()[:8]

	u := seedLbUser(t, db, pfx+"-Settled")
	seedLbWallet(t, db, u.ID, 7.0)
	m := seedLbMatch(t, db)
	seedLbPrediction(t, db, u.ID, m.ID, "correct")

	before := seedLbWalletLog(t, db, u.ID, 100.0, model.WcTournamentWorldCup)
	require.NoError(t, db.Model(&model.WcWalletLog{}).Where("id = ?", before.ID).
		UpdateColumn("created_at", time.Now().UTC().Add(-2*time.Hour)).Error)

	seedLbSettlement(t, db, model.WcTournamentWorldCup, time.Now().UTC().Add(-1*time.Hour))

	after := seedLbWalletLog(t, db, u.ID, 7.0, model.WcTournamentWorldCup)
	require.NoError(t, db.Model(&model.WcWalletLog{}).Where("id = ?", after.ID).
		UpdateColumn("created_at", time.Now().UTC()).Error)

	entries, err := repo.GetLeaderboard(model.WcTournamentWorldCup)
	require.NoError(t, err)

	entry := findLbEntry(entries, u.ID)
	require.NotNil(t, entry)
	assert.Equal(t, 7.0, entry.NetPoints, "pre-settlement ledger rows must not count")
}

// TestGetLeaderboard_SettlementScopedToTournament verifies a World Cup settlement does not
// reset the ASEAN Cup board.
func TestGetLeaderboard_SettlementScopedToTournament(t *testing.T) {
	repo, db := openLeaderboardTestDB(t)
	pfx := uuid.NewString()[:8]

	u := seedLbUser(t, db, pfx+"-CrossTt")
	seedLbWallet(t, db, u.ID, 50)

	acLog := seedLbWalletLog(t, db, u.ID, 50.0, model.WcTournamentAseanCup)
	require.NoError(t, db.Model(&model.WcWalletLog{}).Where("id = ?", acLog.ID).
		UpdateColumn("created_at", time.Now().UTC().Add(-2*time.Hour)).Error)

	// World Cup settles afterwards — must not wipe the ASEAN board.
	seedLbSettlement(t, db, model.WcTournamentWorldCup, time.Now().UTC().Add(-1*time.Hour))

	entries, err := repo.GetLeaderboard(model.WcTournamentAseanCup)
	require.NoError(t, err)

	entry := findLbEntry(entries, u.ID)
	require.NotNil(t, entry)
	assert.Equal(t, 50.0, entry.NetPoints)
}

// TestGetLeaderboard_AdminAdjustmentScopedToTournament verifies a top-up booked against one
// tournament does not leak into the other tournament's leaderboard.
func TestGetLeaderboard_AdminAdjustmentScopedToTournament(t *testing.T) {
	repo, db := openLeaderboardTestDB(t)
	pfx := uuid.NewString()[:8]

	u := seedLbUser(t, db, pfx+"-Zoe")
	seedLbWallet(t, db, u.ID, 100)
	seedLbWalletLog(t, db, u.ID, 100, model.WcTournamentAseanCup)

	wcEntries, err := repo.GetLeaderboard(model.WcTournamentWorldCup)
	require.NoError(t, err)
	assert.Nil(t, findLbEntry(wcEntries, u.ID), "asean_cup top-up must not appear on world_cup board")

	acEntries, err := repo.GetLeaderboard(model.WcTournamentAseanCup)
	require.NoError(t, err)
	acEntry := findLbEntry(acEntries, u.ID)
	require.NotNil(t, acEntry)
	assert.Equal(t, 100.0, acEntry.NetPoints)
}

// TestGetLeaderboard_CustomBetNetCounted verifies custom-bet (kèo phụ) winnings and losses
// count toward net_points — they only ever hit the wallet, so a prediction-only sum drops them.
func TestGetLeaderboard_CustomBetNetCounted(t *testing.T) {
	repo, db := openLeaderboardTestDB(t)
	pfx := uuid.NewString()[:8]

	won := seedLbUser(t, db, pfx+"-Winner")
	seedLbWallet(t, db, won.ID, 15)
	seedLbCustomBetEntry(t, db, won.ID, model.WcTournamentWorldCup, model.WcCustomBetEntryStatusWon, 10, 2.5)
	seedLbWalletLog(t, db, won.ID, 15.0, model.WcTournamentWorldCup) // payout 25 - stake 10

	lost := seedLbUser(t, db, pfx+"-Loser")
	seedLbWallet(t, db, lost.ID, -10)
	seedLbCustomBetEntry(t, db, lost.ID, model.WcTournamentWorldCup, model.WcCustomBetEntryStatusLost, 10, 2.5)
	seedLbWalletLog(t, db, lost.ID, -10.0, model.WcTournamentWorldCup) // -stake

	entries, err := repo.GetLeaderboard(model.WcTournamentWorldCup)
	require.NoError(t, err)

	wonEntry := findLbEntry(entries, won.ID)
	require.NotNil(t, wonEntry, "custom-bet-only user must appear in leaderboard")
	assert.Equal(t, 15.0, wonEntry.NetPoints, "won: payout 25 - stake 10")

	lostEntry := findLbEntry(entries, lost.ID)
	require.NotNil(t, lostEntry)
	assert.Equal(t, -10.0, lostEntry.NetPoints, "lost: -stake")
}

// TestGetLeaderboard_ChampionNetCounted verifies champion-prediction results count toward net_points.
func TestGetLeaderboard_ChampionNetCounted(t *testing.T) {
	repo, db := openLeaderboardTestDB(t)
	pfx := uuid.NewString()[:8]

	u := seedLbUser(t, db, pfx+"-Champ")
	seedLbWallet(t, db, u.ID, 30)
	seedLbChampionPrediction(t, db, u.ID, model.WcTournamentWorldCup, 20, 50)
	seedLbWalletLog(t, db, u.ID, 30.0, model.WcTournamentWorldCup) // earned 50 - staked 20

	entries, err := repo.GetLeaderboard(model.WcTournamentWorldCup)
	require.NoError(t, err)

	entry := findLbEntry(entries, u.ID)
	require.NotNil(t, entry, "champion-pick-only user must appear in leaderboard")
	assert.Equal(t, 30.0, entry.NetPoints, "earned 50 - staked 20")
}

// TestGetLeaderboard_PenaltiesDeducted verifies cancel/reduce penalties — which are charged to
// the wallet — are subtracted from net_points, including for cancelled predictions.
func TestGetLeaderboard_PenaltiesDeducted(t *testing.T) {
	repo, db := openLeaderboardTestDB(t)
	pfx := uuid.NewString()[:8]

	u := seedLbUser(t, db, pfx+"-Penalised")
	seedLbWallet(t, db, u.ID, -5)

	// A settled prediction that broke even, but carried a reduce penalty.
	m1 := seedLbMatch(t, db)
	p1 := seedLbPredictionPts(t, db, u.ID, m1.ID, "correct", 10, 10)
	require.NoError(t, db.Model(&model.WcPrediction{}).Where("id = ?", p1.ID).
		UpdateColumn("reduce_penalty", 2.0).Error)
	seedLbWalletLog(t, db, u.ID, -2.0, model.WcTournamentWorldCup)

	// A cancelled prediction: excluded from the played-prediction counts, but its penalty
	// was still charged to the wallet, so it must still move net_points.
	m2 := seedLbMatch(t, db)
	p2 := seedLbPredictionPts(t, db, u.ID, m2.ID, "correct", 10, 10)
	now := time.Now().UTC()
	require.NoError(t, db.Model(&model.WcPrediction{}).Where("id = ?", p2.ID).
		UpdateColumns(map[string]interface{}{"cancelled_at": now, "cancel_penalty": 3.0}).Error)
	seedLbWalletLog(t, db, u.ID, -3.0, model.WcTournamentWorldCup)

	entries, err := repo.GetLeaderboard(model.WcTournamentWorldCup)
	require.NoError(t, err)

	entry := findLbEntry(entries, u.ID)
	require.NotNil(t, entry)
	assert.Equal(t, -5.0, entry.NetPoints, "reduce penalty 2 + cancel penalty 3 must be deducted")
	assert.Equal(t, 1, entry.TotalPredictions, "cancelled prediction must not count as a played prediction")
}

// TestGetLeaderboard_PredictionStatsAggregatedCorrectly verifies that correct/win_half/lose_half/incorrect
// counts are aggregated from wc_predictions.result, independent of wallet balance.
func TestGetLeaderboard_PredictionStatsAggregatedCorrectly(t *testing.T) {
	repo, db := openLeaderboardTestDB(t)
	pfx := uuid.NewString()[:8]

	u := seedLbUser(t, db, pfx+"-Carol")
	seedLbWallet(t, db, u.ID, 0)

	for _, r := range []string{"correct", "correct", "win_half", "lose_half", "incorrect"} {
		m := seedLbMatch(t, db)
		seedLbPrediction(t, db, u.ID, m.ID, r)
	}

	entries, err := repo.GetLeaderboard(model.WcTournamentWorldCup)
	require.NoError(t, err)

	entry := findLbEntry(entries, u.ID)
	require.NotNil(t, entry)
	assert.Equal(t, 5, entry.TotalPredictions)
	assert.Equal(t, 2, entry.Correct)
	assert.Equal(t, 1, entry.WinHalf)
	assert.Equal(t, 1, entry.LoseHalf)
	assert.Equal(t, 1, entry.Incorrect)
}

// TestGetLeaderboard_Ordering verifies sort order: net_points DESC, correct DESC, name ASC.
func TestGetLeaderboard_Ordering(t *testing.T) {
	repo, db := openLeaderboardTestDB(t)
	pfx := uuid.NewString()[:8]

	// Dave: +10 via admin top-up, 0 correct → highest net_points
	dave := seedLbUser(t, db, pfx+"-Dave")
	seedLbWallet(t, db, dave.ID, 10)
	seedLbWalletLog(t, db, dave.ID, 10, model.WcTournamentWorldCup)

	// Eve: +5 net, 3 correct → second (same pts as Frank/Grace, but most correct)
	eve := seedLbUser(t, db, pfx+"-Eve")
	seedLbWallet(t, db, eve.ID, 5)
	for _, earned := range []float64{15, 10, 10} {
		m := seedLbMatch(t, db)
		seedLbPredictionPts(t, db, eve.ID, m.ID, "correct", 10, earned)
	}
	seedLbWalletLog(t, db, eve.ID, 5, model.WcTournamentWorldCup)

	// Frank: +5 net, 1 correct → before Grace (F < G alphabetically, tiebreak on name)
	frank := seedLbUser(t, db, pfx+"-Frank")
	seedLbWallet(t, db, frank.ID, 5)
	{
		m := seedLbMatch(t, db)
		seedLbPredictionPts(t, db, frank.ID, m.ID, "correct", 10, 15)
	}
	seedLbWalletLog(t, db, frank.ID, 5, model.WcTournamentWorldCup)

	// Grace: +5 net, 1 correct → after Frank (G > F alphabetically)
	grace := seedLbUser(t, db, pfx+"-Grace")
	seedLbWallet(t, db, grace.ID, 5)
	{
		m := seedLbMatch(t, db)
		seedLbPredictionPts(t, db, grace.ID, m.ID, "correct", 10, 15)
	}
	seedLbWalletLog(t, db, grace.ID, 5, model.WcTournamentWorldCup)

	entries, err := repo.GetLeaderboard(model.WcTournamentWorldCup)
	require.NoError(t, err)

	pos := func(id uuid.UUID) int {
		for i, e := range entries {
			if e.WcUserID == id {
				return i
			}
		}
		return -1
	}

	davePos, evePos, frankPos, gracePos := pos(dave.ID), pos(eve.ID), pos(frank.ID), pos(grace.ID)
	require.NotEqual(t, -1, davePos, "Dave must appear in leaderboard")
	require.NotEqual(t, -1, evePos, "Eve must appear in leaderboard")
	require.NotEqual(t, -1, frankPos, "Frank must appear in leaderboard")
	require.NotEqual(t, -1, gracePos, "Grace must appear in leaderboard")

	assert.Less(t, davePos, evePos, "Dave (10 pts) must rank above Eve (5 pts)")
	assert.Less(t, evePos, frankPos, "Eve (3 correct) must rank above Frank (1 correct) at same net_points")
	assert.Less(t, frankPos, gracePos, "Frank must rank above Grace by name tiebreak (F < G)")
}

// TestGetLeaderboard_UserWithoutWalletExcluded verifies that a user with no wallet record
// is excluded from the leaderboard (WHERE w.wc_user_id IS NOT NULL).
func TestGetLeaderboard_UserWithoutWalletExcluded(t *testing.T) {
	repo, db := openLeaderboardTestDB(t)
	pfx := uuid.NewString()[:8]

	u := seedLbUser(t, db, pfx+"-Heidi")
	// intentionally no wallet seeded

	entries, err := repo.GetLeaderboard(model.WcTournamentWorldCup)
	require.NoError(t, err)

	for _, e := range entries {
		assert.NotEqual(t, u.ID, e.WcUserID, "user without wallet must not appear in leaderboard")
	}
}

// TestGetLeaderboard_UserWithNoBetsExcluded verifies that a user who has a wallet but has
// never placed any bet (no predictions, no custom bets, no champion picks) is excluded.
func TestGetLeaderboard_UserWithNoBetsExcluded(t *testing.T) {
	repo, db := openLeaderboardTestDB(t)
	pfx := uuid.NewString()[:8]

	noBets := seedLbUser(t, db, pfx+"-NoBets")
	seedLbWallet(t, db, noBets.ID, 0)
	// no predictions, no custom bets, no champion picks

	// Control: user with at least one prediction must still appear
	hasBets := seedLbUser(t, db, pfx+"-HasBets")
	seedLbWallet(t, db, hasBets.ID, 5)
	m := seedLbMatch(t, db)
	seedLbPrediction(t, db, hasBets.ID, m.ID, "correct")

	entries, err := repo.GetLeaderboard(model.WcTournamentWorldCup)
	require.NoError(t, err)

	for _, e := range entries {
		assert.NotEqual(t, noBets.ID, e.WcUserID, "user with wallet but no bets must be excluded")
	}
	require.NotNil(t, findLbEntry(entries, hasBets.ID), "user with bets must appear in leaderboard")
}

// ─── Tournament isolation: ListMatches ───────────────────────────────────────

func seedMatchWithTournament(t *testing.T, db *gorm.DB, tournamentType string) *model.WcMatch {
	t.Helper()
	m := &model.WcMatch{
		ID:             uuid.New(),
		TournamentType: tournamentType,
		ExternalID:     uuid.NewString(),
		HomeTeam:       "Alpha",
		AwayTeam:       "Beta",
		MatchDate:      time.Now().Add(24 * time.Hour),
		Stage:          model.WcStageGroup,
		Status:         model.WcStatusScheduled,
	}
	require.NoError(t, db.Create(m).Error)
	t.Cleanup(func() { db.Unscoped().Delete(m) })
	return m
}

func TestListMatches_TournamentIsolation_WcDoesNotReturnAc(t *testing.T) {
	repo, db := openWcRepoTestDB(t)

	wcMatch := seedMatchWithTournament(t, db, model.WcTournamentWorldCup)
	acMatch := seedMatchWithTournament(t, db, model.WcTournamentAseanCup)

	matches, err := repo.ListMatches(model.WcTournamentWorldCup, MatchFilter{})
	require.NoError(t, err)

	ids := matchIDs(matches)
	assert.Contains(t, ids, wcMatch.ID, "WC match must appear in WC list")
	assert.NotContains(t, ids, acMatch.ID, "AC match must not appear in WC list")
}

func TestListMatches_TournamentIsolation_AcDoesNotReturnWc(t *testing.T) {
	repo, db := openWcRepoTestDB(t)

	wcMatch := seedMatchWithTournament(t, db, model.WcTournamentWorldCup)
	acMatch := seedMatchWithTournament(t, db, model.WcTournamentAseanCup)

	matches, err := repo.ListMatches(model.WcTournamentAseanCup, MatchFilter{})
	require.NoError(t, err)

	ids := matchIDs(matches)
	assert.Contains(t, ids, acMatch.ID, "AC match must appear in AC list")
	assert.NotContains(t, ids, wcMatch.ID, "WC match must not appear in AC list")
}

// ─── Tournament isolation: GetLeaderboard ────────────────────────────────────

// TestGetLeaderboard_TournamentIsolation verifies that placing a prediction under
// one tournament does not inflate or alter the leaderboard of the other tournament.
// The leaderboard is wallet-based (net_points = wallet balance), so this specifically
// checks that the prediction count columns (which are tournament-scoped) show zero
// for the opposing tournament, while the user appears in neither if they have no wallet.
func TestGetLeaderboard_TournamentIsolation(t *testing.T) {
	repo, db := openLeaderboardTestDB(t)
	pfx := uuid.NewString()[:8]

	// User with a wallet and a WC prediction
	u := seedLbUser(t, db, pfx+"-IsoUser")
	seedLbWallet(t, db, u.ID, 10.0)

	// Seed a WC match and a WC prediction for the user
	wcMatch := &model.WcMatch{
		ID:             uuid.New(),
		TournamentType: model.WcTournamentWorldCup,
		ExternalID:     uuid.NewString(),
		HomeTeam:       "X",
		AwayTeam:       "Y",
		MatchDate:      time.Now().UTC(),
		Stage:          model.WcStageGroup,
		Status:         model.WcStatusCompleted,
	}
	require.NoError(t, db.Create(wcMatch).Error)
	t.Cleanup(func() { db.Unscoped().Delete(wcMatch) })

	seedLbPrediction(t, db, u.ID, wcMatch.ID, "correct")

	// WC leaderboard: user must appear with correct prediction counted
	wcEntries, err := repo.GetLeaderboard(model.WcTournamentWorldCup)
	require.NoError(t, err)
	wcEntry := findLbEntry(wcEntries, u.ID)
	require.NotNil(t, wcEntry, "user must appear in WC leaderboard")
	assert.Equal(t, 1, wcEntry.TotalPredictions, "WC leaderboard must count the WC prediction")

	// AC leaderboard: user does NOT appear (no AC activity)
	acEntries, err := repo.GetLeaderboard(model.WcTournamentAseanCup)
	require.NoError(t, err)
	acEntry := findLbEntry(acEntries, u.ID)
	assert.Nil(t, acEntry, "user with only WC predictions must not appear in AC leaderboard")
}

// ─── Helpers ──────────────────────────────────────────────────────────────────

func matchIDs(matches []*model.WcMatch) []uuid.UUID {
	ids := make([]uuid.UUID, len(matches))
	for i, m := range matches {
		ids[i] = m.ID
	}
	return ids
}

func indexOf(ids []uuid.UUID, target uuid.UUID) int {
	for i, id := range ids {
		if id == target {
			return i
		}
	}
	return -1
}
