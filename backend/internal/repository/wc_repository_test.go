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
		ID:         uuid.New(),
		ExternalID: uuid.NewString(),
		HomeTeam:   "Team A",
		AwayTeam:   "Team B",
		MatchDate:  matchDate,
		Stage:      model.WcStageGroup,
		Status:     status,
	}
	require.NoError(t, db.Create(m).Error)
	t.Cleanup(func() { db.Unscoped().Delete(m) })
	return m
}

// ─── ListMatches — DateFrom / DateTo filter ───────────────────────────────────

func TestListMatches_DateRange_InWindowReturned(t *testing.T) {
	repo, db := openWcRepoTestDB(t)
	now := time.Now().UTC().Truncate(time.Second)
	from := now.Add(-4 * time.Hour)  // mirrors the dashboard's lookback window
	to := now.Add(72 * time.Hour)

	before := seedMatchAt(t, db, from.Add(-1*time.Second), model.WcStatusScheduled)
	inside := seedMatchAt(t, db, now.Add(24*time.Hour), model.WcStatusScheduled)
	after := seedMatchAt(t, db, to.Add(1*time.Second), model.WcStatusScheduled)

	matches, err := repo.ListMatches(MatchFilter{
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

	matches, err := repo.ListMatches(MatchFilter{
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

	matches, err := repo.ListMatches(MatchFilter{}) // no date filter

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

	matches, err := repo.ListMatches(MatchFilter{
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

	matches, err := repo.ListMatches(MatchFilter{
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
	scheduledOnly, err := repo.ListMatches(MatchFilter{
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
	all, err := repo.ListMatches(MatchFilter{
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
	matches, err := repo.ListMatches(MatchFilter{
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

	matches, err := repo.ListMatches(MatchFilter{
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
