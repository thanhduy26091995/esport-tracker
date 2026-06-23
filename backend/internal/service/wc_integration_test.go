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

// openWcTestDB connects to TEST_DATABASE_URL and migrates all WC tables.
func openWcTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("TEST_DATABASE_URL not set — skipping WC integration tests")
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
		&model.WcScoreOdds{},
		&model.WcWallet{},
		&model.WcWalletLog{},
		&model.WcBet{},
		&model.WcSettlement{},
		&model.WcSettlementDetail{},
	))

	// Ensure wc_config row exists
	var cfg model.WcConfig
	if db.First(&cfg, 1).Error != nil {
		db.Create(&model.WcConfig{ID: 1, IsEnabled: true})
	}

	// Ensure WC_JWT_SECRET is set for auth tests
	if os.Getenv("WC_JWT_SECRET") == "" {
		os.Setenv("WC_JWT_SECRET", "test-secret-for-wc-integration-tests")
	}

	return db
}

// newWcServices builds the service pair wired to the given DB.
func newWcServices(db *gorm.DB) (*WcService, *WcAuthService) {
	wcRepo := repository.NewWcRepository(db)
	wcUserRepo := repository.NewWcUserRepository(db)
	return NewWcService(wcRepo, wcUserRepo, nil), NewWcAuthService(wcUserRepo, wcRepo)
}

// seedWcMatch inserts a match and registers cleanup.
func seedWcMatch(t *testing.T, db *gorm.DB, opts ...func(*model.WcMatch)) *model.WcMatch {
	t.Helper()
	future := time.Now().Add(2 * time.Hour)
	m := &model.WcMatch{
		ExternalID: uuid.NewString(),
		HomeTeam:   "France",
		AwayTeam:   "Brazil",
		MatchDate:  future,
		Stage:      model.WcStageGroup,
		Status:     model.WcStatusScheduled,
		BetsLockedAt: &future,
	}
	for _, o := range opts {
		o(m)
	}
	require.NoError(t, db.Create(m).Error)
	t.Cleanup(func() {
		db.Where("match_id = ?", m.ID).Delete(&model.WcScoreOdds{})
		db.Where("match_id = ?", m.ID).Delete(&model.WcBet{})
		db.Delete(m)
	})
	return m
}

// seedWcUserWithGoogle creates a WC user with a google_id already set + wallet.
func seedWcUserWithGoogle(t *testing.T, db *gorm.DB, svc *WcAuthService, name, googleID string) *model.WcUser {
	t.Helper()
	user := &model.WcUser{Name: name, GoogleID: &googleID}
	require.NoError(t, db.Create(user).Error, "seed google-linked user %s", name)
	require.NoError(t, svc.wcRepo.CreateWallet(db, user.ID))
	t.Cleanup(func() {
		db.Where("wc_user_id = ?", user.ID).Delete(&model.WcWalletLog{})
		db.Where("wc_user_id = ?", user.ID).Delete(&model.WcWallet{})
		db.Delete(user)
	})
	return user
}

// seedWcUser creates a WC user + wallet directly via DB (Register was removed).
func seedWcUser(t *testing.T, svc *WcAuthService, name, password string) *model.WcUser {
	t.Helper()
	db := svc.wcRepo.DB()
	hash, err := bcrypt.GenerateFromPassword([]byte(password), 4) // cost 4 for test speed
	require.NoError(t, err)
	hashStr := string(hash)
	user := &model.WcUser{Name: name, PasswordHash: &hashStr}
	require.NoError(t, db.Create(user).Error, "seed user %s", name)
	require.NoError(t, svc.wcRepo.CreateWallet(db, user.ID))
	t.Cleanup(func() {
		db.Where("wc_user_id = ?", user.ID).Delete(&model.WcWalletLog{})
		db.Where("wc_user_id = ?", user.ID).Delete(&model.WcBet{})
		db.Where("wc_user_id = ?", user.ID).Delete(&model.WcWallet{})
		db.Delete(user)
	})
	return user
}

// ─── Auth: Login ──────────────────────────────────────────────────────────────

func TestWcAuth_Login_Success(t *testing.T) {
	db := openWcTestDB(t)
	_, authSvc := newWcServices(db)

	seedWcUser(t, authSvc, "Carol_WC", "mypassword")

	resp, err := authSvc.Login("Carol_WC", "mypassword")
	require.NoError(t, err)
	assert.NotEmpty(t, resp.Token)
	assert.Equal(t, "Carol_WC", resp.Name)
	assert.False(t, resp.GoogleLinked)
}

func TestWcAuth_Login_WrongPassword(t *testing.T) {
	db := openWcTestDB(t)
	_, authSvc := newWcServices(db)

	seedWcUser(t, authSvc, "Dave_WC", "correct")

	_, err := authSvc.Login("Dave_WC", "wrong")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid name or password")
}

func TestWcAuth_Login_UnknownUser(t *testing.T) {
	db := openWcTestDB(t)
	_, authSvc := newWcServices(db)

	_, err := authSvc.Login("nobody_"+uuid.NewString()[:8], "pass")
	assert.Error(t, err)
}

// ─── Auth: JWT round-trip ─────────────────────────────────────────────────────

func TestWcAuth_Token_RoundTrip(t *testing.T) {
	db := openWcTestDB(t)
	_, authSvc := newWcServices(db)

	seedWcUser(t, authSvc, "Eve_WC", "secret")

	resp, err := authSvc.Login("Eve_WC", "secret")
	require.NoError(t, err)

	claims, err := authSvc.VerifyToken(resp.Token)
	require.NoError(t, err)
	assert.Equal(t, "Eve_WC", claims.Name)
	assert.False(t, claims.IsAdmin)
}

func TestWcAuth_Token_InvalidString(t *testing.T) {
	db := openWcTestDB(t)
	_, authSvc := newWcServices(db)

	_, err := authSvc.VerifyToken("not.a.valid.jwt")
	assert.Error(t, err)
}

// ─── Bet placement ────────────────────────────────────────────────────────────

func TestWcBet_PlaceHandicap_Accepted(t *testing.T) {
	db := openWcTestDB(t)
	svc, authSvc := newWcServices(db)

	user := seedWcUser(t, authSvc, "Grace_WC", "pass")
	hv := 0.5
	oh := 1.90
	oa := 1.95
	m := seedWcMatch(t, db, func(m *model.WcMatch) {
		m.HandicapTeam = model.WcTeamHome
		m.HandicapValue = &hv
		m.OddsHandicapHome = &oh
		m.OddsHandicapAway = &oa
	})

	home := model.WcTeamHome
	bet, err := svc.PlaceBet(user.ID, PlaceBetRequest{
		MatchID:   m.ID,
		BetType:   model.WcBetTypeHandicap,
		BetChoice: &home,
		Stake:     100,
	})
	require.NoError(t, err)
	assert.Equal(t, model.WcBetTypeHandicap, bet.BetType)
	assert.Equal(t, 100, bet.Stake)
	assert.InDelta(t, 1.90, bet.OddsSnapshot, 0.001)
}

func TestWcBet_PlaceExactScore_Accepted(t *testing.T) {
	db := openWcTestDB(t)
	svc, authSvc := newWcServices(db)

	user := seedWcUser(t, authSvc, "Hank_WC", "pass")
	m := seedWcMatch(t, db)

	// Add a scoreline
	so := &model.WcScoreOdds{MatchID: m.ID, HomeScore: 2, AwayScore: 1, Odds: 6.00}
	require.NoError(t, db.Create(so).Error)
	t.Cleanup(func() { db.Delete(so) })

	hs, as := 2, 1
	bet, err := svc.PlaceBet(user.ID, PlaceBetRequest{
		MatchID:            m.ID,
		BetType:            model.WcBetTypeExactScore,
		PredictedHomeScore: &hs,
		PredictedAwayScore: &as,
		Stake:              50,
	})
	require.NoError(t, err)
	assert.Equal(t, 50, bet.Stake)
	assert.InDelta(t, 6.00, bet.OddsSnapshot, 0.001)
}

func TestWcBet_DuplicateHandicapSide_Rejected(t *testing.T) {
	db := openWcTestDB(t)
	svc, authSvc := newWcServices(db)

	user := seedWcUser(t, authSvc, "Ivy_WC", "pass")
	hv := 0.5
	oh := 1.90
	oa := 1.95
	m := seedWcMatch(t, db, func(m *model.WcMatch) {
		m.HandicapTeam = model.WcTeamHome
		m.HandicapValue = &hv
		m.OddsHandicapHome = &oh
		m.OddsHandicapAway = &oa
	})

	home := model.WcTeamHome
	_, err := svc.PlaceBet(user.ID, PlaceBetRequest{MatchID: m.ID, BetType: model.WcBetTypeHandicap, BetChoice: &home, Stake: 100})
	require.NoError(t, err)

	_, err = svc.PlaceBet(user.ID, PlaceBetRequest{MatchID: m.ID, BetType: model.WcBetTypeHandicap, BetChoice: &home, Stake: 50})
	assert.Error(t, err, "same side handicap twice must fail")
}

func TestWcBet_DuplicateExactScoreline_Rejected(t *testing.T) {
	db := openWcTestDB(t)
	svc, authSvc := newWcServices(db)

	user := seedWcUser(t, authSvc, "Jack_WC", "pass")
	m := seedWcMatch(t, db)

	so := &model.WcScoreOdds{MatchID: m.ID, HomeScore: 1, AwayScore: 0, Odds: 5.00}
	require.NoError(t, db.Create(so).Error)
	t.Cleanup(func() { db.Delete(so) })

	hs, as := 1, 0
	_, err := svc.PlaceBet(user.ID, PlaceBetRequest{MatchID: m.ID, BetType: model.WcBetTypeExactScore, PredictedHomeScore: &hs, PredictedAwayScore: &as, Stake: 100})
	require.NoError(t, err)

	_, err = svc.PlaceBet(user.ID, PlaceBetRequest{MatchID: m.ID, BetType: model.WcBetTypeExactScore, PredictedHomeScore: &hs, PredictedAwayScore: &as, Stake: 50})
	assert.Error(t, err, "same scoreline twice must fail")
}

func TestWcBet_DifferentScorelines_SameMatch_Accepted(t *testing.T) {
	db := openWcTestDB(t)
	svc, authSvc := newWcServices(db)

	user := seedWcUser(t, authSvc, "Kate_WC", "pass")
	m := seedWcMatch(t, db)

	so1 := &model.WcScoreOdds{MatchID: m.ID, HomeScore: 1, AwayScore: 0, Odds: 5.00}
	so2 := &model.WcScoreOdds{MatchID: m.ID, HomeScore: 2, AwayScore: 1, Odds: 6.00}
	require.NoError(t, db.Create(so1).Error)
	require.NoError(t, db.Create(so2).Error)
	t.Cleanup(func() { db.Delete(so1); db.Delete(so2) })

	hs1, as1 := 1, 0
	_, err := svc.PlaceBet(user.ID, PlaceBetRequest{MatchID: m.ID, BetType: model.WcBetTypeExactScore, PredictedHomeScore: &hs1, PredictedAwayScore: &as1, Stake: 100})
	require.NoError(t, err)

	hs2, as2 := 2, 1
	_, err = svc.PlaceBet(user.ID, PlaceBetRequest{MatchID: m.ID, BetType: model.WcBetTypeExactScore, PredictedHomeScore: &hs2, PredictedAwayScore: &as2, Stake: 100})
	assert.NoError(t, err, "different scorelines on same match must be accepted")
}

func TestWcBet_LockedMatch_Rejected(t *testing.T) {
	db := openWcTestDB(t)
	svc, authSvc := newWcServices(db)

	user := seedWcUser(t, authSvc, "Leo_WC", "pass")
	past := time.Now().Add(-1 * time.Hour)
	m := seedWcMatch(t, db, func(m *model.WcMatch) {
		m.BetsLockedAt = &past // already past → locked
	})

	hv := 0.5
	oh := 1.90
	oa := 1.95
	db.Model(m).Updates(map[string]interface{}{
		"handicap_team":      model.WcTeamHome,
		"handicap_value":     hv,
		"odds_handicap_home": oh,
		"odds_handicap_away": oa,
	})

	home := model.WcTeamHome
	_, err := svc.PlaceBet(user.ID, PlaceBetRequest{MatchID: m.ID, BetType: model.WcBetTypeHandicap, BetChoice: &home, Stake: 100})
	assert.Error(t, err, "bet on locked match must be rejected")
	assert.Contains(t, err.Error(), "closed")
}

func TestWcBet_ZeroBalance_Accepted(t *testing.T) {
	db := openWcTestDB(t)
	svc, authSvc := newWcServices(db)

	user := seedWcUser(t, authSvc, "Mia_WC", "pass")
	hv := 0.5
	oh := 1.90
	oa := 1.95
	m := seedWcMatch(t, db, func(m *model.WcMatch) {
		m.HandicapTeam = model.WcTeamHome
		m.HandicapValue = &hv
		m.OddsHandicapHome = &oh
		m.OddsHandicapAway = &oa
	})

	// Wallet starts at 0 — bet must still be accepted
	wallet, _ := svc.GetWallet(user.ID)
	require.Equal(t, 0, wallet.Balance)

	home := model.WcTeamHome
	_, err := svc.PlaceBet(user.ID, PlaceBetRequest{MatchID: m.ID, BetType: model.WcBetTypeHandicap, BetChoice: &home, Stake: 500})
	assert.NoError(t, err, "bet with zero balance must be accepted (credit-based)")
}

// ─── Match settlement ─────────────────────────────────────────────────────────

// buildSettledMatch creates a match, places a bet, settles it, and returns the wallet balance.
func buildSettledMatch(t *testing.T, db *gorm.DB, svc *WcService, authSvc *WcAuthService,
	actualHome, actualAway int, betChoice string, stake int, handicap float64, handicapTeam string,
	oddsHome, oddsAway float64) (float64, error) {

	t.Helper()
	user := seedWcUser(t, authSvc, "Settler_"+uuid.NewString()[:6], "pass")

	future := time.Now().Add(2 * time.Hour)
	m := &model.WcMatch{
		ExternalID:       uuid.NewString(),
		HomeTeam:         "A",
		AwayTeam:         "B",
		MatchDate:        future,
		Stage:            model.WcStageGroup,
		Status:           model.WcStatusScheduled,
		BetsLockedAt:     &future,
		HandicapTeam:     handicapTeam,
		HandicapValue:    &handicap,
		OddsHandicapHome: &oddsHome,
		OddsHandicapAway: &oddsAway,
	}
	require.NoError(t, db.Create(m).Error)
	t.Cleanup(func() {
		db.Where("match_id = ?", m.ID).Delete(&model.WcBet{})
		db.Delete(m)
	})

	choice := betChoice
	_, err := svc.PlaceBet(user.ID, PlaceBetRequest{
		MatchID:   m.ID,
		BetType:   model.WcBetTypeHandicap,
		BetChoice: &choice,
		Stake:     stake,
	})
	require.NoError(t, err)

	// Set actual score
	require.NoError(t, db.Model(m).Updates(map[string]interface{}{
		"home_score": actualHome,
		"away_score": actualAway,
		"status":     model.WcStatusCompleted,
	}).Error)

	_, _, err = svc.SettleMatch(m.ID)
	if err != nil {
		return 0, err
	}

	wallet, err := svc.GetWallet(user.ID)
	require.NoError(t, err)
	return wallet.Balance, nil
}

func TestWcSettle_WinnerCredited(t *testing.T) {
	db := openWcTestDB(t)
	svc, authSvc := newWcServices(db)

	// Home gives 0.5 ball, user bets home (stake=100, odds=1.90)
	// Score 2-1 → adjusted_home=1.5 > 1 → home wins → payout=190
	// wallet = payout - stake = 190 - 100 = 90
	balance, err := buildSettledMatch(t, db, svc, authSvc, 2, 1, "home", 100, 0.5, "home", 1.90, 1.95)
	require.NoError(t, err)
	assert.Equal(t, 90, balance)
}

func TestWcSettle_LoserDebited(t *testing.T) {
	db := openWcTestDB(t)
	svc, authSvc := newWcServices(db)

	// Home gives 0.5 ball, user bets home (stake=100)
	// Score 1-1 → adjusted_home=0.5 < 1 → away wins → payout=0
	// wallet = 0 - 100 = -100
	balance, err := buildSettledMatch(t, db, svc, authSvc, 1, 1, "home", 100, 0.5, "home", 1.90, 1.95)
	require.NoError(t, err)
	assert.Equal(t, -100, balance)
}

func TestWcSettle_Push_StakeReturned(t *testing.T) {
	db := openWcTestDB(t)
	svc, authSvc := newWcServices(db)

	// Home gives 1.0, score 2-1 → adjusted_home=1 == away 1 → push → payout=stake=100
	// wallet = 100 - 100 = 0
	balance, err := buildSettledMatch(t, db, svc, authSvc, 2, 1, "home", 100, 1.0, "home", 1.85, 1.95)
	require.NoError(t, err)
	assert.Equal(t, 0, balance)
}

func TestWcSettle_Idempotent(t *testing.T) {
	db := openWcTestDB(t)
	svc, authSvc := newWcServices(db)

	user := seedWcUser(t, authSvc, "Idem_"+uuid.NewString()[:6], "pass")
	hv := 0.5
	oh := 1.90
	oa := 1.95
	future := time.Now().Add(2 * time.Hour)
	m := &model.WcMatch{
		ExternalID:       uuid.NewString(),
		HomeTeam:         "A", AwayTeam: "B",
		MatchDate:        future,
		Stage:            model.WcStageGroup,
		Status:           model.WcStatusScheduled,
		BetsLockedAt:     &future,
		HandicapTeam:     model.WcTeamHome,
		HandicapValue:    &hv,
		OddsHandicapHome: &oh,
		OddsHandicapAway: &oa,
	}
	require.NoError(t, db.Create(m).Error)
	t.Cleanup(func() {
		db.Where("match_id = ?", m.ID).Delete(&model.WcBet{})
		db.Delete(m)
	})

	home := model.WcTeamHome
	_, err := svc.PlaceBet(user.ID, PlaceBetRequest{MatchID: m.ID, BetType: model.WcBetTypeHandicap, BetChoice: &home, Stake: 100})
	require.NoError(t, err)

	require.NoError(t, db.Model(m).Updates(map[string]interface{}{"home_score": 2, "away_score": 1}).Error)

	_, _, err = svc.SettleMatch(m.ID)
	require.NoError(t, err)
	wallet1, _ := svc.GetWallet(user.ID)

	// Settle again — wallet must be unchanged
	_, _, err = svc.SettleMatch(m.ID)
	require.NoError(t, err)
	wallet2, _ := svc.GetWallet(user.ID)

	assert.Equal(t, wallet1.Balance, wallet2.Balance, "re-settle must be idempotent")
}

// ─── Tournament settlement ────────────────────────────────────────────────────

func TestWcTournamentSettlement_SnapshotAndReset(t *testing.T) {
	db := openWcTestDB(t)
	svc, authSvc := newWcServices(db)

	// Two users with known balances
	userA := seedWcUser(t, authSvc, "SettleA_"+uuid.NewString()[:6], "pass")
	userB := seedWcUser(t, authSvc, "SettleB_"+uuid.NewString()[:6], "pass")

	// Manually set balances via admin top-up
	require.NoError(t, svc.AdminTopUp(userA.ID, userA.ID, 500, "test"))
	require.NoError(t, svc.AdminTopUp(userB.ID, userB.ID, -200, "test"))

	settlement, err := svc.CreateSettlement(userA.ID, "Test Settlement", 1000, "unit test")
	require.NoError(t, err)
	assert.NotEqual(t, uuid.Nil, settlement.ID)

	// Both wallets must be reset to 0
	wA, _ := svc.GetWallet(userA.ID)
	wB, _ := svc.GetWallet(userB.ID)
	assert.Equal(t, 0, wA.Balance, "wallet A must be 0 after settlement")
	assert.Equal(t, 0, wB.Balance, "wallet B must be 0 after settlement")

	// Settlement history must be queryable
	detail, err := svc.GetSettlement(settlement.ID)
	require.NoError(t, err)
	assert.Len(t, detail.Details, 2)

	t.Cleanup(func() {
		db.Where("settlement_id = ?", settlement.ID).Delete(&model.WcSettlementDetail{})
		db.Delete(settlement)
		db.Where("wc_user_id IN ?", []uuid.UUID{userA.ID, userB.ID}).Delete(&model.WcWalletLog{})
	})
}

func TestWcTournamentSettlement_HistoryPreserved(t *testing.T) {
	db := openWcTestDB(t)
	svc, authSvc := newWcServices(db)

	user := seedWcUser(t, authSvc, "Hist_"+uuid.NewString()[:6], "pass")
	require.NoError(t, svc.AdminTopUp(user.ID, user.ID, 300, ""))

	s1, err := svc.CreateSettlement(user.ID, "Settlement 1", 1000, "")
	require.NoError(t, err)
	t.Cleanup(func() {
		db.Where("settlement_id = ?", s1.ID).Delete(&model.WcSettlementDetail{})
		db.Delete(s1)
		db.Where("wc_user_id = ?", user.ID).Delete(&model.WcWalletLog{})
	})

	// Topup again and create second settlement
	require.NoError(t, svc.AdminTopUp(user.ID, user.ID, 150, ""))
	s2, err := svc.CreateSettlement(user.ID, "Settlement 2", 1000, "")
	require.NoError(t, err)
	t.Cleanup(func() {
		db.Where("settlement_id = ?", s2.ID).Delete(&model.WcSettlementDetail{})
		db.Delete(s2)
	})

	list, err := svc.ListSettlements()
	require.NoError(t, err)

	ids := make([]uuid.UUID, len(list))
	for i, s := range list {
		ids[i] = s.ID
	}
	assert.Contains(t, ids, s1.ID, "first settlement must be in history")
	assert.Contains(t, ids, s2.ID, "second settlement must be in history")
}

// ─── Poisson: BulkUpsertScoreOdds ────────────────────────────────────────────

func TestPoisson_GetMatchWithOdds_EmptyBeforeUpsert_ReturnsEmptySlice(t *testing.T) {
	// Regression: before the nil→[] fix, GetMatchWithOdds returned null for score_odds
	// when no rows existed, causing the frontend to fall back to score_multipliers.
	db := openWcTestDB(t)
	wcRepo := repository.NewWcRepository(db)

	m := seedWcMatch(t, db)

	result, err := wcRepo.GetMatchWithOdds(m.ID)
	require.NoError(t, err)

	// Must be empty slice — not nil — so JSON serializes as [] not null.
	require.NotNil(t, result.ScoreMultipliers, "ScoreMultipliers must be initialized to empty slice, not nil")
	assert.Empty(t, result.ScoreMultipliers)
}

func TestPoisson_BulkUpsert_ThenGetMatchWithOdds_ReturnsOdds(t *testing.T) {
	// Critical path: after Poisson generation and BulkUpsert, GetMatchWithOdds
	// must return the saved score_odds so exact-score bets can be placed.
	db := openWcTestDB(t)
	wcRepo := repository.NewWcRepository(db)
	poissonSvc := NewPoissonService()

	m := seedWcMatch(t, db)

	_, dbOdds := poissonSvc.GenerateScoreOdds(PoissonInput{
		MatchID:     m.ID,
		HomeLambda:  1.5,
		AwayLambda:  1.2,
		HouseMargin: 0.05,
		MinProb:     0.01,
	})
	require.NotEmpty(t, dbOdds)

	require.NoError(t, wcRepo.BulkUpsertScoreMultipliers(dbOdds))

	result, err := wcRepo.GetMatchWithOdds(m.ID)
	require.NoError(t, err)
	assert.NotEmpty(t, result.ScoreMultipliers, "score_multipliers must be non-empty after BulkUpsert")
	assert.Equal(t, len(dbOdds), len(result.ScoreMultipliers))
}

func TestPoisson_BulkUpsert_Idempotent(t *testing.T) {
	// Upserting the same scorelines twice must not create duplicates or fail.
	db := openWcTestDB(t)
	wcRepo := repository.NewWcRepository(db)
	poissonSvc := NewPoissonService()

	m := seedWcMatch(t, db)

	input := PoissonInput{
		MatchID:     m.ID,
		HomeLambda:  1.5,
		AwayLambda:  1.2,
		HouseMargin: 0.05,
		MinProb:     0.01,
	}
	_, dbOdds := poissonSvc.GenerateScoreOdds(input)
	require.NotEmpty(t, dbOdds)

	require.NoError(t, wcRepo.BulkUpsertScoreMultipliers(dbOdds))
	require.NoError(t, wcRepo.BulkUpsertScoreMultipliers(dbOdds), "second upsert must not fail")

	listed, err := wcRepo.ListScoreMultipliers(m.ID)
	require.NoError(t, err)
	assert.Equal(t, len(dbOdds), len(listed), "duplicate upsert must not create extra rows")
}

func TestPoisson_BulkUpsert_UpdatesOddsOnConflict(t *testing.T) {
	// When the same (match_id, home_score, away_score) is upserted with a different
	// odds value, the stored odds must be updated to the new value.
	db := openWcTestDB(t)
	wcRepo := repository.NewWcRepository(db)

	m := seedWcMatch(t, db)

	first := []model.WcScoreMultiplier{
		{ID: uuid.New(), MatchID: m.ID, HomeScore: 1, AwayScore: 0, Multiplier: 5.00},
	}
	require.NoError(t, wcRepo.BulkUpsertScoreMultipliers(first))

	updated := []model.WcScoreMultiplier{
		{ID: uuid.New(), MatchID: m.ID, HomeScore: 1, AwayScore: 0, Multiplier: 6.50},
	}
	require.NoError(t, wcRepo.BulkUpsertScoreMultipliers(updated))

	result, err := wcRepo.GetMatchWithOdds(m.ID)
	require.NoError(t, err)
	require.Len(t, result.ScoreMultipliers, 1)
	assert.InDelta(t, 6.50, result.ScoreMultipliers[0].Multiplier, 0.001, "multiplier must be updated on conflict")
}

func TestPoisson_BulkUpsert_ExactScoreBetCanBePlaced(t *testing.T) {
	// End-to-end: generate → upsert → place exact score bet.
	// This is the flow that was broken before the nil→[] fix.
	db := openWcTestDB(t)
	wcRepo := repository.NewWcRepository(db)
	poissonSvc := NewPoissonService()
	svc, authSvc := newWcServices(db)

	user := seedWcUser(t, authSvc, "PoissonBet_"+uuid.NewString()[:6], "pass")
	m := seedWcMatch(t, db)

	_, dbOdds := poissonSvc.GenerateScoreOdds(PoissonInput{
		MatchID:     m.ID,
		HomeLambda:  1.5,
		AwayLambda:  1.2,
		HouseMargin: 0.05,
		MinProb:     0.01,
	})
	require.NotEmpty(t, dbOdds)
	require.NoError(t, wcRepo.BulkUpsertScoreMultipliers(dbOdds))

	// Pick the first generated scoreline to bet on.
	target := dbOdds[0]
	hs, as := target.HomeScore, target.AwayScore

	bet, err := svc.PlaceBet(user.ID, PlaceBetRequest{
		MatchID:            m.ID,
		BetType:            model.WcBetTypeExactScore,
		PredictedHomeScore: &hs,
		PredictedAwayScore: &as,
		Stake:              100,
	})
	require.NoError(t, err, "exact score bet must succeed after Poisson odds are saved")
	assert.Equal(t, model.WcBetTypeExactScore, bet.BetType)
	assert.InDelta(t, target.Multiplier, bet.OddsSnapshot, 0.001)
}
