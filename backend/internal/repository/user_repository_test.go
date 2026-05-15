package repository

import (
	"os"
	"testing"

	"github.com/duyb/esport-score-tracker/internal/model"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// openTestDB connects using TEST_DATABASE_URL (Postgres DSN).
// Example: TEST_DATABASE_URL="host=localhost port=5432 user=postgres password=secret dbname=esport_test sslmode=disable"
func openTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("TEST_DATABASE_URL not set — skipping repository integration tests")
	}
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	require.NoError(t, err, "open test DB")
	require.NoError(t, db.AutoMigrate(&model.User{}, &model.Match{}, &model.MatchParticipant{}))
	return db
}

// seedUser inserts a user and registers cleanup.
func seedUser(t *testing.T, db *gorm.DB, name string) *model.User {
	t.Helper()
	u := &model.User{
		ID:       uuid.New(),
		Name:     name,
		Tier:     "normal",
		IsActive: true,
	}
	require.NoError(t, db.Create(u).Error)
	t.Cleanup(func() { db.Unscoped().Delete(u) })
	return u
}

// seedParticipant inserts a match_participant row with a synthetic match ID.
func seedParticipant(t *testing.T, db *gorm.DB, userID uuid.UUID, pointChange int) {
	t.Helper()
	mp := &model.MatchParticipant{
		ID:          uuid.New(),
		MatchID:     uuid.New(), // synthetic — FK not enforced in test schema
		UserID:      userID,
		TeamNumber:  1,
		PointChange: pointChange,
	}
	require.NoError(t, db.Create(mp).Error)
	t.Cleanup(func() { db.Unscoped().Delete(mp) })
}

// ─── GetWinRatesBatch ─────────────────────────────────────────────────────────

func TestGetWinRatesBatch_BasicWinRate(t *testing.T) {
	db := openTestDB(t)
	repo := NewUserRepository(db)

	// User A: 10 matches, 6 wins, 4 losses → 60% win rate
	userA := seedUser(t, db, "PlayerA_WinRate")
	for i := 0; i < 6; i++ {
		seedParticipant(t, db, userA.ID, 1) // win
	}
	for i := 0; i < 4; i++ {
		seedParticipant(t, db, userA.ID, -1) // loss
	}

	result, err := repo.GetWinRatesBatch([]uuid.UUID{userA.ID})
	require.NoError(t, err)

	got, ok := result[userA.ID]
	require.True(t, ok, "user A must be in result")
	assert.Equal(t, 10, got.TotalMatches)
	assert.Equal(t, 6, got.WonMatches)
	assert.InDelta(t, 0.60, got.WinRate, 0.001)
}

func TestGetWinRatesBatch_DrawsExcluded(t *testing.T) {
	db := openTestDB(t)
	repo := NewUserRepository(db)

	// 6 wins + 4 draws: draws (point_change=0) must not count
	// Expected: total_matches=6, won_matches=6, win_rate=1.0
	userB := seedUser(t, db, "PlayerB_Draws")
	for i := 0; i < 6; i++ {
		seedParticipant(t, db, userB.ID, 1) // win
	}
	for i := 0; i < 4; i++ {
		seedParticipant(t, db, userB.ID, 0) // draw — must be excluded
	}

	result, err := repo.GetWinRatesBatch([]uuid.UUID{userB.ID})
	require.NoError(t, err)

	got, ok := result[userB.ID]
	require.True(t, ok)
	assert.Equal(t, 6, got.TotalMatches, "draws must not count toward total_matches")
	assert.Equal(t, 6, got.WonMatches)
	assert.InDelta(t, 1.0, got.WinRate, 0.001)
}

func TestGetWinRatesBatch_NoMatches(t *testing.T) {
	db := openTestDB(t)
	repo := NewUserRepository(db)

	userC := seedUser(t, db, "PlayerC_NoMatches")

	result, err := repo.GetWinRatesBatch([]uuid.UUID{userC.ID})
	require.NoError(t, err)

	got, ok := result[userC.ID]
	require.True(t, ok)
	assert.Equal(t, 0, got.TotalMatches)
	assert.Equal(t, 0, got.WonMatches)
	assert.InDelta(t, 0.0, got.WinRate, 0.001)
}

func TestGetWinRatesBatch_MultipleUsers(t *testing.T) {
	db := openTestDB(t)
	repo := NewUserRepository(db)

	// Pro: 12 wins / 20 matches = 60%
	pro := seedUser(t, db, "PlayerPro_Multi")
	for i := 0; i < 12; i++ {
		seedParticipant(t, db, pro.ID, 1)
	}
	for i := 0; i < 8; i++ {
		seedParticipant(t, db, pro.ID, -1)
	}

	// Noob: 2 wins / 10 matches = 20%
	noob := seedUser(t, db, "PlayerNoob_Multi")
	for i := 0; i < 2; i++ {
		seedParticipant(t, db, noob.ID, 1)
	}
	for i := 0; i < 8; i++ {
		seedParticipant(t, db, noob.ID, -1)
	}

	result, err := repo.GetWinRatesBatch([]uuid.UUID{pro.ID, noob.ID})
	require.NoError(t, err)
	assert.Len(t, result, 2)

	assert.Equal(t, 20, result[pro.ID].TotalMatches)
	assert.InDelta(t, 0.60, result[pro.ID].WinRate, 0.001)

	assert.Equal(t, 10, result[noob.ID].TotalMatches)
	assert.InDelta(t, 0.20, result[noob.ID].WinRate, 0.001)
}

func TestGetWinRatesBatch_EmptyIDs(t *testing.T) {
	db := openTestDB(t)
	repo := NewUserRepository(db)

	result, err := repo.GetWinRatesBatch([]uuid.UUID{})
	require.NoError(t, err)
	assert.Empty(t, result)
}

// ─── UpdateTier ───────────────────────────────────────────────────────────────

func TestUpdateTier_PersistsTierValue(t *testing.T) {
	db := openTestDB(t)
	repo := NewUserRepository(db)

	user := seedUser(t, db, "PlayerTierUpdate")

	require.NoError(t, repo.UpdateTier(user.ID, "pro"))

	var updated model.User
	require.NoError(t, db.First(&updated, "id = ?", user.ID).Error)
	assert.Equal(t, "pro", updated.Tier)
}
