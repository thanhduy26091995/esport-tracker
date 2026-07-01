package service

import (
	"os"
	"testing"
	"time"

	"github.com/duyb/esport-score-tracker/internal/model"
	"github.com/duyb/esport-score-tracker/internal/repository"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// openPredictionTestDB connects to TEST_DATABASE_URL and migrates all WC tables including WcPrediction.
func openPredictionTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("TEST_DATABASE_URL not set — skipping WC prediction penalty integration tests")
	}
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		t.Skipf("cannot connect to test DB (%v) — skipping", err)
	}
	require.NoError(t, db.AutoMigrate(
		&model.WcUser{},
		&model.WcConfig{},
		&model.WcMatch{},
		&model.WcScoreMultiplier{},
		&model.WcScoreOdds{},
		&model.WcWallet{},
		&model.WcWalletLog{},
		&model.WcPrediction{},
	))

	// Ensure wc_config row with penalty defaults
	var cfg model.WcConfig
	if db.First(&cfg, 1).Error != nil {
		db.Create(&model.WcConfig{ID: 1, IsEnabled: true, MinPoints: 1, MaxPoints: 100})
	}

	return db
}

func newPredictionService(db *gorm.DB) *WcService {
	wcRepo := repository.NewWcRepository(db)
	wcUserRepo := repository.NewWcUserRepository(db)
	return NewWcService(wcRepo, wcUserRepo, nil, nil)
}

// seedPredUser creates a WC user with a seeded wallet balance.
func seedPredUser(t *testing.T, db *gorm.DB, name string, balance float64) *model.WcUser {
	t.Helper()
	hash, _ := bcrypt.GenerateFromPassword([]byte("pass"), 4)
	hashStr := string(hash)
	user := &model.WcUser{Name: name + "_" + uuid.NewString()[:6], PasswordHash: &hashStr}
	require.NoError(t, db.Create(user).Error)

	wcRepo := repository.NewWcRepository(db)
	require.NoError(t, wcRepo.CreateWallet(db, user.ID))
	if balance > 0 {
		require.NoError(t, wcRepo.UpdateWalletBalance(db, user.ID, balance))
	}

	t.Cleanup(func() {
		db.Where("wc_user_id = ?", user.ID).Delete(&model.WcWalletLog{})
		db.Where("wc_user_id = ?", user.ID).Delete(&model.WcPrediction{})
		db.Where("wc_user_id = ?", user.ID).Delete(&model.WcWallet{})
		db.Delete(user)
	})
	return user
}

// seedHandicapMatch creates a WcMatch with handicap odds and a future lock time.
func seedHandicapMatch(t *testing.T, db *gorm.DB) *model.WcMatch {
	t.Helper()
	future := time.Now().Add(2 * time.Hour)
	hv := 0.5
	oh, oa := 1.9, 1.95
	m := &model.WcMatch{
		ExternalID:   uuid.NewString(),
		HomeTeam:     "France",
		AwayTeam:     "Brazil",
		MatchDate:    future,
		Stage:        model.WcStageGroup,
		Status:       model.WcStatusScheduled,
		BetsLockedAt: &future,
		HandicapTeam:     model.WcTeamHome,
		HandicapValue:    &hv,
		OddsHandicapHome: &oh,
		OddsHandicapAway: &oa,
	}
	require.NoError(t, db.Create(m).Error)
	t.Cleanup(func() {
		db.Where("match_id = ?", m.ID).Delete(&model.WcPrediction{})
		db.Delete(m)
	})
	return m
}

// setCancelPenaltyConfig updates wc_config penalty settings in DB.
func setCancelPenaltyConfig(t *testing.T, db *gorm.DB, enabled bool, cancelPct, reduceMax, reducePenPct int) {
	t.Helper()
	require.NoError(t, db.Model(&model.WcConfig{}).Where("id = 1").Updates(map[string]interface{}{
		"cancel_penalty_enabled":    enabled,
		"cancel_penalty_percent":    cancelPct,
		"bet_reduce_max_percent":    reduceMax,
		"bet_reduce_penalty_percent": reducePenPct,
	}).Error)
	t.Cleanup(func() {
		db.Model(&model.WcConfig{}).Where("id = 1").Updates(map[string]interface{}{
			"cancel_penalty_enabled":    false,
			"cancel_penalty_percent":    20,
			"bet_reduce_max_percent":    50,
			"bet_reduce_penalty_percent": 20,
		})
	})
}

// ─── SubmitPrediction: original_points ───────────────────────────────────────

func TestPrediction_Submit_SetsOriginalPoints(t *testing.T) {
	db := openPredictionTestDB(t)
	svc := newPredictionService(db)

	user := seedPredUser(t, db, "Alice", 50)
	m := seedHandicapMatch(t, db)
	choice := model.WcTeamHome

	pred, err := svc.SubmitPrediction(user.ID, SubmitPredictionRequest{
		MatchID:          m.ID,
		PredictionType:   model.WcPredictionTypeHandicap,
		PredictionChoice: &choice,
		Points:           10,
	})
	require.NoError(t, err)
	require.NotNil(t, pred.OriginalPoints)
	assert.Equal(t, 10, *pred.OriginalPoints)
}

// ─── DeletePrediction: cancel penalty ────────────────────────────────────────

func TestPrediction_Delete_WithPenalty_DeductsWallet(t *testing.T) {
	db := openPredictionTestDB(t)
	svc := newPredictionService(db)
	setCancelPenaltyConfig(t, db, true, 20, 50, 20)

	user := seedPredUser(t, db, "Bob", 100)
	m := seedHandicapMatch(t, db)
	choice := model.WcTeamHome

	pred, err := svc.SubmitPrediction(user.ID, SubmitPredictionRequest{
		MatchID:          m.ID,
		PredictionType:   model.WcPredictionTypeHandicap,
		PredictionChoice: &choice,
		Points:           10,
	})
	require.NoError(t, err)

	// 20% of 10 = 2.0 penalty
	penalty, err := svc.DeletePrediction(user.ID, pred.ID)
	require.NoError(t, err)
	assert.Equal(t, 2.0, penalty)

	// wallet: 100 + 0 (wallet not charged on submit) - 2.0 = 98.0
	wcRepo := repository.NewWcRepository(db)
	wallet, err := wcRepo.GetWallet(user.ID)
	require.NoError(t, err)
	assert.Equal(t, 98.0, wallet.Balance)

	// prediction should be soft-cancelled
	var p model.WcPrediction
	require.NoError(t, db.First(&p, pred.ID).Error)
	assert.NotNil(t, p.CancelledAt)
	assert.NotNil(t, p.CancelPenalty)
	assert.Equal(t, 2.0, *p.CancelPenalty)
}

func TestPrediction_Delete_PenaltyDisabled_NoDeduction(t *testing.T) {
	db := openPredictionTestDB(t)
	svc := newPredictionService(db)
	setCancelPenaltyConfig(t, db, false, 20, 50, 20) // disabled

	user := seedPredUser(t, db, "Carol", 100)
	m := seedHandicapMatch(t, db)
	choice := model.WcTeamAway

	pred, err := svc.SubmitPrediction(user.ID, SubmitPredictionRequest{
		MatchID:          m.ID,
		PredictionType:   model.WcPredictionTypeHandicap,
		PredictionChoice: &choice,
		Points:           10,
	})
	require.NoError(t, err)

	penalty, err := svc.DeletePrediction(user.ID, pred.ID)
	require.NoError(t, err)
	assert.Equal(t, 0.0, penalty)

	wcRepo := repository.NewWcRepository(db)
	wallet, err := wcRepo.GetWallet(user.ID)
	require.NoError(t, err)
	assert.Equal(t, 100.0, wallet.Balance) // unchanged
}

func TestPrediction_Delete_FloatPenalty_NotRounded(t *testing.T) {
	db := openPredictionTestDB(t)
	svc := newPredictionService(db)
	setCancelPenaltyConfig(t, db, true, 20, 50, 20)

	user := seedPredUser(t, db, "Dave", 100)
	m := seedHandicapMatch(t, db)
	choice := model.WcTeamHome

	pred, err := svc.SubmitPrediction(user.ID, SubmitPredictionRequest{
		MatchID:          m.ID,
		PredictionType:   model.WcPredictionTypeHandicap,
		PredictionChoice: &choice,
		Points:           7, // 20% of 7 = 1.4 (not floor to 1)
	})
	require.NoError(t, err)

	penalty, err := svc.DeletePrediction(user.ID, pred.ID)
	require.NoError(t, err)
	assert.Equal(t, 1.4, penalty)

	wcRepo := repository.NewWcRepository(db)
	wallet, err := wcRepo.GetWallet(user.ID)
	require.NoError(t, err)
	assert.Equal(t, 98.6, wallet.Balance)
}

// ─── Partial unique index: cancel → re-predict ───────────────────────────────

func TestPrediction_Cancel_ThenRepredict_Allowed(t *testing.T) {
	db := openPredictionTestDB(t)
	svc := newPredictionService(db)
	setCancelPenaltyConfig(t, db, false, 0, 50, 20)

	user := seedPredUser(t, db, "Eve", 100)
	m := seedHandicapMatch(t, db)
	choice := model.WcTeamHome

	// First prediction
	pred1, err := svc.SubmitPrediction(user.ID, SubmitPredictionRequest{
		MatchID:          m.ID,
		PredictionType:   model.WcPredictionTypeHandicap,
		PredictionChoice: &choice,
		Points:           5,
	})
	require.NoError(t, err)

	// Cancel it
	_, err = svc.DeletePrediction(user.ID, pred1.ID)
	require.NoError(t, err)

	// Re-predict with same pick — partial index should allow it
	pred2, err := svc.SubmitPrediction(user.ID, SubmitPredictionRequest{
		MatchID:          m.ID,
		PredictionType:   model.WcPredictionTypeHandicap,
		PredictionChoice: &choice,
		Points:           5,
	})
	require.NoError(t, err)
	assert.NotEqual(t, pred1.ID, pred2.ID)
}

// ─── ListPredictions: returns cancelled ──────────────────────────────────────

func TestPrediction_List_IncludesCancelled(t *testing.T) {
	db := openPredictionTestDB(t)
	svc := newPredictionService(db)
	setCancelPenaltyConfig(t, db, false, 0, 50, 20)

	user := seedPredUser(t, db, "Frank", 100)
	m := seedHandicapMatch(t, db)
	choice := model.WcTeamHome

	pred, err := svc.SubmitPrediction(user.ID, SubmitPredictionRequest{
		MatchID:          m.ID,
		PredictionType:   model.WcPredictionTypeHandicap,
		PredictionChoice: &choice,
		Points:           5,
	})
	require.NoError(t, err)

	_, err = svc.DeletePrediction(user.ID, pred.ID)
	require.NoError(t, err)

	preds, err := svc.ListPredictions(user.ID)
	require.NoError(t, err)

	found := false
	for _, p := range preds {
		if p.ID == pred.ID {
			found = true
			assert.NotNil(t, p.CancelledAt)
		}
	}
	assert.True(t, found, "cancelled prediction should appear in ListPredictions")
}

// ─── UpdatePredictionPoints: reduce penalty ──────────────────────────────────

func TestPrediction_Reduce_WithinFreeThreshold_NoPenalty(t *testing.T) {
	db := openPredictionTestDB(t)
	svc := newPredictionService(db)
	setCancelPenaltyConfig(t, db, true, 20, 50, 20) // max 50% free

	user := seedPredUser(t, db, "Grace", 100)
	m := seedHandicapMatch(t, db)
	choice := model.WcTeamHome

	pred, err := svc.SubmitPrediction(user.ID, SubmitPredictionRequest{
		MatchID:          m.ID,
		PredictionType:   model.WcPredictionTypeHandicap,
		PredictionChoice: &choice,
		Points:           100,
	})
	require.NoError(t, err)

	// Reduce 100 → 60 (within 50% = allowedMin 50, 60 >= 50 → free)
	penalty, err := svc.UpdatePredictionPoints(user.ID, pred.ID, 60)
	require.NoError(t, err)
	assert.Equal(t, 0.0, penalty)

	wcRepo := repository.NewWcRepository(db)
	wallet, err := wcRepo.GetWallet(user.ID)
	require.NoError(t, err)
	assert.Equal(t, 100.0, wallet.Balance)
}

func TestPrediction_Reduce_ExceedsThreshold_PenaltyApplied(t *testing.T) {
	db := openPredictionTestDB(t)
	svc := newPredictionService(db)
	setCancelPenaltyConfig(t, db, true, 20, 50, 20) // max 50%, penalty 20%

	user := seedPredUser(t, db, "Heidi", 100)
	m := seedHandicapMatch(t, db)
	choice := model.WcTeamHome

	pred, err := svc.SubmitPrediction(user.ID, SubmitPredictionRequest{
		MatchID:          m.ID,
		PredictionType:   model.WcPredictionTypeHandicap,
		PredictionChoice: &choice,
		Points:           100,
	})
	require.NoError(t, err)

	// Reduce 100 → 40 (excess = 50-40 = 10, penalty = 10 * 20% = 2.0)
	penalty, err := svc.UpdatePredictionPoints(user.ID, pred.ID, 40)
	require.NoError(t, err)
	assert.Equal(t, 2.0, penalty)

	wcRepo := repository.NewWcRepository(db)
	wallet, err := wcRepo.GetWallet(user.ID)
	require.NoError(t, err)
	assert.Equal(t, 98.0, wallet.Balance)

	// Old prediction should be soft-cancelled with cancel_penalty storing the reduce penalty
	var oldPred model.WcPrediction
	require.NoError(t, db.First(&oldPred, pred.ID).Error)
	assert.NotNil(t, oldPred.CancelledAt, "old prediction should be soft-cancelled")
	require.NotNil(t, oldPred.CancelPenalty)
	assert.Equal(t, 2.0, *oldPred.CancelPenalty)

	// A new prediction should exist with the reduced points
	var newPreds []model.WcPrediction
	require.NoError(t, db.Where("match_id = ? AND wc_user_id = ? AND cancelled_at IS NULL", m.ID, user.ID).Find(&newPreds).Error)
	require.Len(t, newPreds, 1, "one active prediction should exist after reduce")
	assert.Equal(t, 40, newPreds[0].Points)
	require.NotNil(t, newPreds[0].OriginalPoints)
	assert.Equal(t, 40, *newPreds[0].OriginalPoints)
}

func TestPrediction_Reduce_NullOriginalPoints_NoPenalty(t *testing.T) {
	db := openPredictionTestDB(t)
	svc := newPredictionService(db)
	setCancelPenaltyConfig(t, db, true, 20, 50, 20)

	user := seedPredUser(t, db, "Ivan", 100)
	m := seedHandicapMatch(t, db)
	choice := model.WcTeamHome

	pred, err := svc.SubmitPrediction(user.ID, SubmitPredictionRequest{
		MatchID:          m.ID,
		PredictionType:   model.WcPredictionTypeHandicap,
		PredictionChoice: &choice,
		Points:           100,
	})
	require.NoError(t, err)

	// Simulate old data: clear original_points
	require.NoError(t, db.Model(&model.WcPrediction{}).Where("id = ?", pred.ID).
		Update("original_points", nil).Error)

	// Reduce below threshold — no penalty because original_points is NULL (grace for old data)
	penalty, err := svc.UpdatePredictionPoints(user.ID, pred.ID, 10)
	require.NoError(t, err)
	assert.Equal(t, 0.0, penalty)

	wcRepo := repository.NewWcRepository(db)
	wallet, err := wcRepo.GetWallet(user.ID)
	require.NoError(t, err)
	assert.Equal(t, 100.0, wallet.Balance) // no deduction
}
