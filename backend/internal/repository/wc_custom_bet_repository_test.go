package repository

import (
	"testing"

	"github.com/duyb/esport-score-tracker/internal/model"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

// openCustomBetTestDB migrates the tables needed for custom-bet repo tests.
func openCustomBetTestDB(t *testing.T) (*WcCustomBetRepository, *gorm.DB) {
	t.Helper()
	db := openTestDB(t) // skips if TEST_DATABASE_URL not set
	require.NoError(t, db.AutoMigrate(&model.WcCustomBet{}))
	return NewWcCustomBetRepository(db), db
}

// seedCustomBet inserts a wc_custom_bets row with the given match and status.
func seedCustomBet(t *testing.T, db *gorm.DB, matchID uuid.UUID, status string) *model.WcCustomBet {
	t.Helper()
	b := &model.WcCustomBet{
		ID:      uuid.New(),
		MatchID: matchID,
		Title:   "kèo phụ test",
		Status:  status,
	}
	require.NoError(t, db.Create(b).Error)
	t.Cleanup(func() { db.Unscoped().Delete(b) })
	return b
}

// ─── CountOpenByMatchIDs ────────────────────────────────────────────────────────

func TestCountOpenByMatchIDs_CountsOnlyOpen(t *testing.T) {
	repo, db := openCustomBetTestDB(t)
	match := uuid.New()

	// 2 open, plus one of each other status — only the 2 open should count.
	seedCustomBet(t, db, match, model.WcCustomBetStatusOpen)
	seedCustomBet(t, db, match, model.WcCustomBetStatusOpen)
	seedCustomBet(t, db, match, model.WcCustomBetStatusClosed)
	seedCustomBet(t, db, match, model.WcCustomBetStatusSettled)
	seedCustomBet(t, db, match, model.WcCustomBetStatusVoid)

	counts, err := repo.CountOpenByMatchIDs([]uuid.UUID{match})
	require.NoError(t, err)
	assert.Equal(t, 2, counts[match], "only open custom bets should be counted")
}

func TestCountOpenByMatchIDs_SeparatesByMatch(t *testing.T) {
	repo, db := openCustomBetTestDB(t)
	matchA := uuid.New()
	matchB := uuid.New()

	seedCustomBet(t, db, matchA, model.WcCustomBetStatusOpen)
	seedCustomBet(t, db, matchB, model.WcCustomBetStatusOpen)
	seedCustomBet(t, db, matchB, model.WcCustomBetStatusOpen)

	counts, err := repo.CountOpenByMatchIDs([]uuid.UUID{matchA, matchB})
	require.NoError(t, err)
	assert.Equal(t, 1, counts[matchA])
	assert.Equal(t, 2, counts[matchB])
}

func TestCountOpenByMatchIDs_MatchWithNoOpenOmitted(t *testing.T) {
	repo, db := openCustomBetTestDB(t)
	match := uuid.New()

	// Only a settled bet — the match should be absent from the result map (→ 0).
	seedCustomBet(t, db, match, model.WcCustomBetStatusSettled)

	counts, err := repo.CountOpenByMatchIDs([]uuid.UUID{match})
	require.NoError(t, err)
	_, ok := counts[match]
	assert.False(t, ok, "match with no open bets should be omitted from the map")
	assert.Equal(t, 0, counts[match], "absent key reads as zero")
}

func TestCountOpenByMatchIDs_EmptyInput(t *testing.T) {
	repo, _ := openCustomBetTestDB(t)

	counts, err := repo.CountOpenByMatchIDs(nil)
	require.NoError(t, err)
	assert.Empty(t, counts, "empty id slice returns an empty map")
}
