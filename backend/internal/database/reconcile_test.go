package database

import (
	"os"
	"testing"
	"time"

	"github.com/duyb/esport-score-tracker/internal/model"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// openReconcileTestDB migrates the tables ReconcileWalletLedgerSQL touches.
func openReconcileTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("TEST_DATABASE_URL not set — skipping ledger reconciliation tests")
	}
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	require.NoError(t, err, "open test DB")
	require.NoError(t, db.AutoMigrate(
		&model.WcUser{},
		&model.WcWallet{},
		&model.WcWalletLog{},
		&model.WcSettlement{},
	))
	return db
}

func seedRecUser(t *testing.T, db *gorm.DB) *model.WcUser {
	t.Helper()
	u := &model.WcUser{ID: uuid.New(), Name: "Rec-" + uuid.NewString()[:8]}
	require.NoError(t, db.Create(u).Error)
	t.Cleanup(func() { db.Unscoped().Delete(u) })
	return u
}

func seedRecWallet(t *testing.T, db *gorm.DB, userID uuid.UUID, balance float64) {
	t.Helper()
	w := &model.WcWallet{ID: uuid.New(), WcUserID: userID, Balance: balance}
	require.NoError(t, db.Create(w).Error)
	t.Cleanup(func() { db.Unscoped().Delete(w) })
}

func seedRecLog(t *testing.T, db *gorm.DB, userID uuid.UUID, delta float64, tt string) {
	t.Helper()
	l := &model.WcWalletLog{
		ID: uuid.New(), TournamentType: tt, WcUserID: userID,
		AdminID: uuid.New(), Delta: delta, Note: "seed",
	}
	require.NoError(t, db.Create(l).Error)
	t.Cleanup(func() { db.Unscoped().Delete(l) })
}

// ledgerTotal mirrors the leaderboard's net_points expression for world_cup.
func ledgerTotal(t *testing.T, db *gorm.DB, userID uuid.UUID) float64 {
	t.Helper()
	var total *float64
	require.NoError(t, db.Raw(`
		SELECT SUM(wl.delta) FROM wc_wallet_logs wl
		WHERE wl.wc_user_id = ? AND wl.tournament_type = 'world_cup'
		  AND wl.created_at > COALESCE(
				(SELECT MAX(s.created_at) FROM wc_settlements s WHERE s.tournament_type = 'world_cup'),
				TIMESTAMPTZ '-infinity')`, userID).Scan(&total).Error)
	if total == nil {
		return 0
	}
	return *total
}

// cleanupReconcileRows removes rows the migration inserts so tests stay independent.
func cleanupReconcileRows(t *testing.T, db *gorm.DB) {
	t.Helper()
	t.Cleanup(func() {
		db.Unscoped().Where("note = ?", "leaderboard ledger reconciliation").Delete(&model.WcWalletLog{})
	})
}

// TestReconcile_LedgerGapClosed is the guarantee that no member's current points change:
// a wallet whose ledger is incomplete gets a balancing row so the two match exactly.
func TestReconcile_LedgerGapClosed(t *testing.T) {
	db := openReconcileTestDB(t)
	cleanupReconcileRows(t, db)

	u := seedRecUser(t, db)
	seedRecWallet(t, db, u.ID, 100)
	seedRecLog(t, db, u.ID, 30, model.WcTournamentWorldCup) // ledger only knows 30 of the 100

	require.NoError(t, db.Exec(ReconcileWalletLedgerSQL).Error)

	assert.Equal(t, 100.0, ledgerTotal(t, db, u.ID),
		"ledger must reproduce the wallet balance so current standings are preserved")
}

// TestReconcile_WalletWithNoLedgerAtAll covers users whose whole balance predates logging.
func TestReconcile_WalletWithNoLedgerAtAll(t *testing.T) {
	db := openReconcileTestDB(t)
	cleanupReconcileRows(t, db)

	u := seedRecUser(t, db)
	seedRecWallet(t, db, u.ID, -42.5)

	require.NoError(t, db.Exec(ReconcileWalletLedgerSQL).Error)

	assert.Equal(t, -42.5, ledgerTotal(t, db, u.ID))
}

// TestReconcile_AlreadyConsistent_NoRowInserted verifies a complete ledger is left alone.
func TestReconcile_AlreadyConsistent_NoRowInserted(t *testing.T) {
	db := openReconcileTestDB(t)
	cleanupReconcileRows(t, db)

	u := seedRecUser(t, db)
	seedRecWallet(t, db, u.ID, 55)
	seedRecLog(t, db, u.ID, 55, model.WcTournamentWorldCup)

	require.NoError(t, db.Exec(ReconcileWalletLedgerSQL).Error)

	var n int64
	require.NoError(t, db.Model(&model.WcWalletLog{}).
		Where("wc_user_id = ? AND note = ?", u.ID, "leaderboard ledger reconciliation").
		Count(&n).Error)
	assert.Zero(t, n, "a consistent ledger must not get a balancing row")
	assert.Equal(t, 55.0, ledgerTotal(t, db, u.ID))
}

// TestReconcile_Idempotent verifies a second run is a no-op — the migration list re-runs on
// every boot, and a second balancing row would double every member's points.
func TestReconcile_Idempotent(t *testing.T) {
	db := openReconcileTestDB(t)
	cleanupReconcileRows(t, db)

	u := seedRecUser(t, db)
	seedRecWallet(t, db, u.ID, 70)
	seedRecLog(t, db, u.ID, 20, model.WcTournamentWorldCup)

	require.NoError(t, db.Exec(ReconcileWalletLedgerSQL).Error)
	first := ledgerTotal(t, db, u.ID)
	require.NoError(t, db.Exec(ReconcileWalletLedgerSQL).Error)

	assert.Equal(t, 70.0, first)
	assert.Equal(t, first, ledgerTotal(t, db, u.ID), "re-running must not change any total")
}

// seedRecSettlement records a settlement at a specific time.
func seedRecSettlement(t *testing.T, db *gorm.DB, tt string, at time.Time) *model.WcSettlement {
	t.Helper()
	s := &model.WcSettlement{
		ID: uuid.New(), TournamentType: tt,
		Name: "rec-" + uuid.NewString()[:6], PointRate: 1000, SettledBy: uuid.New(),
	}
	require.NoError(t, db.Create(s).Error)
	require.NoError(t, db.Model(&model.WcSettlement{}).Where("id = ?", s.ID).
		UpdateColumn("created_at", at).Error)
	t.Cleanup(func() { db.Unscoped().Delete(s) })
	return s
}

// seedRecLogAt seeds a wallet log at a specific time.
func seedRecLogAt(t *testing.T, db *gorm.DB, userID uuid.UUID, delta float64, tt string, at time.Time) *model.WcWalletLog {
	t.Helper()
	l := &model.WcWalletLog{
		ID: uuid.New(), TournamentType: tt, WcUserID: userID,
		AdminID: uuid.New(), Delta: delta, Note: "seed",
	}
	require.NoError(t, db.Create(l).Error)
	require.NoError(t, db.Model(&model.WcWalletLog{}).Where("id = ?", l.ID).
		UpdateColumn("created_at", at).Error)
	t.Cleanup(func() { db.Unscoped().Delete(l) })
	return l
}

func logTournament(t *testing.T, db *gorm.DB, id uuid.UUID) string {
	t.Helper()
	var got string
	require.NoError(t, db.Model(&model.WcWalletLog{}).Where("id = ?", id).
		Select("tournament_type").Scan(&got).Error)
	return got
}

// TestBackfillAsean_SplitsAtLatestSettlement verifies post-settlement rows move to asean_cup
// while the World Cup era is left alone.
func TestBackfillAsean_SplitsAtLatestSettlement(t *testing.T) {
	db := openReconcileTestDB(t)
	cleanupReconcileRows(t, db)

	u := seedRecUser(t, db)
	seedRecWallet(t, db, u.ID, 5)

	base := time.Now().UTC()
	before := seedRecLogAt(t, db, u.ID, 100, model.WcTournamentWorldCup, base.Add(-3*time.Hour))
	seedRecSettlement(t, db, model.WcTournamentAseanCup, base.Add(-2*time.Hour))
	after := seedRecLogAt(t, db, u.ID, 5, model.WcTournamentWorldCup, base.Add(-1*time.Hour))

	require.NoError(t, db.Exec(BackfillAseanWalletLogsSQL).Error)

	assert.Equal(t, model.WcTournamentWorldCup, logTournament(t, db, before.ID),
		"pre-settlement rows stay on world_cup")
	assert.Equal(t, model.WcTournamentAseanCup, logTournament(t, db, after.ID),
		"post-settlement rows move to asean_cup")
}

// TestBackfillAsean_Idempotent is the important one: the migration list re-runs on every boot,
// and a second pass after a future settlement would silently re-tag fresh ASEAN history.
func TestBackfillAsean_Idempotent(t *testing.T) {
	db := openReconcileTestDB(t)
	cleanupReconcileRows(t, db)

	u := seedRecUser(t, db)
	seedRecWallet(t, db, u.ID, 5)

	base := time.Now().UTC()
	wc := seedRecLogAt(t, db, u.ID, 100, model.WcTournamentWorldCup, base.Add(-3*time.Hour))
	seedRecSettlement(t, db, model.WcTournamentAseanCup, base.Add(-2*time.Hour))
	ac := seedRecLogAt(t, db, u.ID, 5, model.WcTournamentWorldCup, base.Add(-1*time.Hour))

	require.NoError(t, db.Exec(BackfillAseanWalletLogsSQL).Error)

	// A later settlement plus new world_cup-tagged activity must survive a second pass.
	seedRecSettlement(t, db, model.WcTournamentWorldCup, base.Add(-30*time.Minute))
	fresh := seedRecLogAt(t, db, u.ID, 7, model.WcTournamentWorldCup, base.Add(-10*time.Minute))

	require.NoError(t, db.Exec(BackfillAseanWalletLogsSQL).Error)

	assert.Equal(t, model.WcTournamentWorldCup, logTournament(t, db, wc.ID))
	assert.Equal(t, model.WcTournamentAseanCup, logTournament(t, db, ac.ID))
	assert.Equal(t, model.WcTournamentWorldCup, logTournament(t, db, fresh.ID),
		"a second pass must not re-tag activity recorded after the backfill")
}

// TestReconcile_CountsOnlyPostSettlementLedger verifies the balancing row accounts for the
// settlement reset: pre-settlement rows are excluded from both the gap and the result.
func TestReconcile_CountsOnlyPostSettlementLedger(t *testing.T) {
	db := openReconcileTestDB(t)
	cleanupReconcileRows(t, db)

	u := seedRecUser(t, db)
	seedRecWallet(t, db, u.ID, 12)

	old := &model.WcWalletLog{
		ID: uuid.New(), TournamentType: model.WcTournamentWorldCup, WcUserID: u.ID,
		AdminID: uuid.New(), Delta: 999, Note: "seed-old",
	}
	require.NoError(t, db.Create(old).Error)
	require.NoError(t, db.Model(&model.WcWalletLog{}).Where("id = ?", old.ID).
		UpdateColumn("created_at", time.Now().UTC().Add(-2*time.Hour)).Error)
	t.Cleanup(func() { db.Unscoped().Delete(old) })

	s := &model.WcSettlement{
		ID: uuid.New(), TournamentType: model.WcTournamentWorldCup,
		Name: "rec settlement", PointRate: 1000, SettledBy: uuid.New(),
	}
	require.NoError(t, db.Create(s).Error)
	require.NoError(t, db.Model(&model.WcSettlement{}).Where("id = ?", s.ID).
		UpdateColumn("created_at", time.Now().UTC().Add(-1*time.Hour)).Error)
	t.Cleanup(func() { db.Unscoped().Delete(s) })

	require.NoError(t, db.Exec(ReconcileWalletLedgerSQL).Error)

	assert.Equal(t, 12.0, ledgerTotal(t, db, u.ID),
		"the pre-settlement 999 must be ignored; only the current balance is reproduced")
}
