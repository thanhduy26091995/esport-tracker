package service

import (
	"testing"
	"time"

	"github.com/duyb/esport-score-tracker/internal/cache"
	"github.com/duyb/esport-score-tracker/internal/model"
	"github.com/duyb/esport-score-tracker/internal/repository"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

// ─── Test helpers ─────────────────────────────────────────────────────────────

func openCustomBetTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db := openWcTestDB(t)
	require.NoError(t, db.AutoMigrate(
		&model.WcCustomBet{},
		&model.WcCustomBetOption{},
		&model.WcCustomBetEntry{},
	))
	return db
}

func newCustomBetService(db *gorm.DB) (*WcCustomBetService, *WcAuthService) {
	wcRepo := repository.NewWcRepository(db)
	wcUserRepo := repository.NewWcUserRepository(db)
	customBetRepo := repository.NewWcCustomBetRepository(db)
	authSvc := NewWcAuthService(wcUserRepo, wcRepo)
	return NewWcCustomBetService(customBetRepo, wcRepo, wcUserRepo, nil, cache.NewGoCacheStore(time.Minute, time.Minute)), authSvc
}

// seedCustomBetUser creates a user + wallet with a given starting balance.
func seedCustomBetUser(t *testing.T, db *gorm.DB, authSvc *WcAuthService, name string, balance float64) *model.WcUser {
	t.Helper()
	user := seedWcUser(t, authSvc, name+"_"+uuid.NewString()[:6], "pass")
	require.NoError(t, db.Model(&model.WcWallet{}).
		Where("wc_user_id = ?", user.ID).
		Update("balance", balance).Error)
	return user
}

// seedCustomBetMatch creates a WcMatch and registers cleanup.
func seedCustomBetMatch(t *testing.T, db *gorm.DB) *model.WcMatch {
	t.Helper()
	return seedWcMatch(t, db)
}

// seedCustomBet creates a bet with two default options.
func seedCustomBet(t *testing.T, svc *WcCustomBetService, matchID, adminID uuid.UUID, status string) (*model.WcCustomBetWithOptions, []model.WcCustomBetOption) {
	t.Helper()
	opts := []CreateCustomBetOption{
		{Label: "Có", Odds: 1.8, DisplayOrder: 0},
		{Label: "Không", Odds: 2.0, DisplayOrder: 1},
	}
	result, err := svc.CreateCustomBet(matchID, adminID, "Test kèo "+uuid.NewString()[:6], nil, opts)
	require.NoError(t, err)
	if status != model.WcCustomBetStatusOpen {
		require.NoError(t, svc.repo.UpdateBet(nil, result.ID, map[string]interface{}{"status": status}))
		result.Status = status
	}
	t.Cleanup(func() {
		svc.repo.GetOptions(result.ID) // touch to ensure FK cascade
		svc.wcRepo.DB().Where("custom_bet_id = ?", result.ID).Delete(&model.WcCustomBetEntry{})
		svc.wcRepo.DB().Where("custom_bet_id = ?", result.ID).Delete(&model.WcCustomBetOption{})
		svc.wcRepo.DB().Delete(&model.WcCustomBet{}, result.ID)
	})
	return result, result.Options
}

func walletBalance(t *testing.T, db *gorm.DB, userID uuid.UUID) float64 {
	t.Helper()
	var w model.WcWallet
	require.NoError(t, db.Where("wc_user_id = ?", userID).First(&w).Error)
	return w.Balance
}

// ─── CreateCustomBet ──────────────────────────────────────────────────────────

func TestWcCustomBet_Create_HappyPath(t *testing.T) {
	db := openCustomBetTestDB(t)
	svc, authSvc := newCustomBetService(db)
	admin := seedCustomBetUser(t, db, authSvc, "Admin", 0)
	match := seedCustomBetMatch(t, db)

	opts := []CreateCustomBetOption{
		{Label: "Có", Odds: 1.8, DisplayOrder: 0},
		{Label: "Không", Odds: 2.0, DisplayOrder: 1},
		{Label: "Không rõ", Odds: 3.5, DisplayOrder: 2},
	}
	result, err := svc.CreateCustomBet(match.ID, admin.ID, "Có 6 phạt góc?", nil, opts)
	require.NoError(t, err)
	t.Cleanup(func() {
		db.Where("custom_bet_id = ?", result.ID).Delete(&model.WcCustomBetOption{})
		db.Delete(&model.WcCustomBet{}, result.ID)
	})

	assert.Equal(t, "Có 6 phạt góc?", result.Title)
	assert.Equal(t, model.WcCustomBetStatusOpen, result.Status)
	assert.Equal(t, match.ID, result.MatchID)
	assert.Len(t, result.Options, 3)
	assert.Equal(t, "Có", result.Options[0].Label)
	assert.InDelta(t, 1.8, result.Options[0].Odds, 0.001)
}

func TestWcCustomBet_Create_TooFewOptions(t *testing.T) {
	db := openCustomBetTestDB(t)
	svc, authSvc := newCustomBetService(db)
	admin := seedCustomBetUser(t, db, authSvc, "Admin", 0)
	match := seedCustomBetMatch(t, db)

	_, err := svc.CreateCustomBet(match.ID, admin.ID, "title", nil, []CreateCustomBetOption{
		{Label: "only one", Odds: 1.9, DisplayOrder: 0},
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "2")
}

func TestWcCustomBet_Create_TooManyOptions(t *testing.T) {
	db := openCustomBetTestDB(t)
	svc, authSvc := newCustomBetService(db)
	admin := seedCustomBetUser(t, db, authSvc, "Admin", 0)
	match := seedCustomBetMatch(t, db)

	opts := make([]CreateCustomBetOption, 11)
	for i := range opts {
		opts[i] = CreateCustomBetOption{Label: "opt", Odds: 1.5, DisplayOrder: i}
	}
	_, err := svc.CreateCustomBet(match.ID, admin.ID, "title", nil, opts)
	require.Error(t, err)
}

func TestWcCustomBet_Create_EmptyLabel(t *testing.T) {
	db := openCustomBetTestDB(t)
	svc, authSvc := newCustomBetService(db)
	admin := seedCustomBetUser(t, db, authSvc, "Admin", 0)
	match := seedCustomBetMatch(t, db)

	_, err := svc.CreateCustomBet(match.ID, admin.ID, "title", nil, []CreateCustomBetOption{
		{Label: "", Odds: 1.8, DisplayOrder: 0},
		{Label: "B", Odds: 2.0, DisplayOrder: 1},
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "label")
}

func TestWcCustomBet_Create_ZeroOdds(t *testing.T) {
	db := openCustomBetTestDB(t)
	svc, authSvc := newCustomBetService(db)
	admin := seedCustomBetUser(t, db, authSvc, "Admin", 0)
	match := seedCustomBetMatch(t, db)

	_, err := svc.CreateCustomBet(match.ID, admin.ID, "title", nil, []CreateCustomBetOption{
		{Label: "A", Odds: 0, DisplayOrder: 0},
		{Label: "B", Odds: 2.0, DisplayOrder: 1},
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "odds")
}

// ─── PlaceEntry ───────────────────────────────────────────────────────────────

func TestWcCustomBet_PlaceEntry_HappyPath(t *testing.T) {
	db := openCustomBetTestDB(t)
	svc, authSvc := newCustomBetService(db)
	admin := seedCustomBetUser(t, db, authSvc, "Admin", 0)
	player := seedCustomBetUser(t, db, authSvc, "Player", 10)
	match := seedCustomBetMatch(t, db)
	bet, opts := seedCustomBet(t, svc, match.ID, admin.ID, model.WcCustomBetStatusOpen)

	err := svc.PlaceEntry(bet.ID, player.ID, opts[0].ID, 3, "test-user")
	require.NoError(t, err)

	// Wallet unchanged at placement (deferred-deduction model)
	assert.InDelta(t, 10.0, walletBalance(t, db, player.ID), 0.001)

	// Entry exists with correct odds snapshot
	var entry model.WcCustomBetEntry
	require.NoError(t, db.Where("custom_bet_id = ? AND wc_user_id = ?", bet.ID, player.ID).First(&entry).Error)
	assert.Equal(t, opts[0].ID, entry.OptionID)
	assert.Equal(t, 3, entry.Stake)
	assert.InDelta(t, opts[0].Odds, entry.OddsSnapshot, 0.001)
	assert.Equal(t, model.WcCustomBetEntryStatusPending, entry.Status)
}

func TestWcCustomBet_PlaceEntry_BetClosed(t *testing.T) {
	db := openCustomBetTestDB(t)
	svc, authSvc := newCustomBetService(db)
	admin := seedCustomBetUser(t, db, authSvc, "Admin", 0)
	player := seedCustomBetUser(t, db, authSvc, "Player", 10)
	match := seedCustomBetMatch(t, db)
	bet, opts := seedCustomBet(t, svc, match.ID, admin.ID, model.WcCustomBetStatusClosed)

	err := svc.PlaceEntry(bet.ID, player.ID, opts[0].ID, 3, "test-user")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "đóng")
}

func TestWcCustomBet_PlaceEntry_OptionFromDifferentBet(t *testing.T) {
	db := openCustomBetTestDB(t)
	svc, authSvc := newCustomBetService(db)
	admin := seedCustomBetUser(t, db, authSvc, "Admin", 0)
	player := seedCustomBetUser(t, db, authSvc, "Player", 10)
	match := seedCustomBetMatch(t, db)
	bet1, _ := seedCustomBet(t, svc, match.ID, admin.ID, model.WcCustomBetStatusOpen)
	_, opts2 := seedCustomBet(t, svc, match.ID, admin.ID, model.WcCustomBetStatusOpen)

	// Try to place on bet1 using an option from bet2
	err := svc.PlaceEntry(bet1.ID, player.ID, opts2[0].ID, 3, "test-user")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "hợp lệ")
}

func TestWcCustomBet_PlaceEntry_StakeBelowMin(t *testing.T) {
	db := openCustomBetTestDB(t)
	svc, authSvc := newCustomBetService(db)
	admin := seedCustomBetUser(t, db, authSvc, "Admin", 0)
	player := seedCustomBetUser(t, db, authSvc, "Player", 10)
	match := seedCustomBetMatch(t, db)
	bet, opts := seedCustomBet(t, svc, match.ID, admin.ID, model.WcCustomBetStatusOpen)

	// min_points defaults to 1; stake 0 is below
	err := svc.PlaceEntry(bet.ID, player.ID, opts[0].ID, 0, "test-user")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "điểm cược")
}

func TestWcCustomBet_PlaceEntry_ZeroBalance(t *testing.T) {
	// Deferred-deduction model: no upfront balance check — 0-balance users can place.
	db := openCustomBetTestDB(t)
	svc, authSvc := newCustomBetService(db)
	admin := seedCustomBetUser(t, db, authSvc, "Admin", 0)
	player := seedCustomBetUser(t, db, authSvc, "Player", 0)
	match := seedCustomBetMatch(t, db)
	bet, opts := seedCustomBet(t, svc, match.ID, admin.ID, model.WcCustomBetStatusOpen)

	err := svc.PlaceEntry(bet.ID, player.ID, opts[0].ID, 3, "test-user")
	require.NoError(t, err)
	// Balance still 0 — nothing deducted at placement
	assert.InDelta(t, 0.0, walletBalance(t, db, player.ID), 0.001)
}

func TestWcCustomBet_PlaceEntry_DuplicateEntry(t *testing.T) {
	db := openCustomBetTestDB(t)
	svc, authSvc := newCustomBetService(db)
	admin := seedCustomBetUser(t, db, authSvc, "Admin", 0)
	player := seedCustomBetUser(t, db, authSvc, "Player", 20)
	match := seedCustomBetMatch(t, db)
	bet, opts := seedCustomBet(t, svc, match.ID, admin.ID, model.WcCustomBetStatusOpen)

	require.NoError(t, svc.PlaceEntry(bet.ID, player.ID, opts[0].ID, 2, "test-user"))
	err := svc.PlaceEntry(bet.ID, player.ID, opts[1].ID, 2, "test-user") // same bet, different option
	require.Error(t, err)
	// Wallet unchanged throughout (no deduction at placement)
	assert.InDelta(t, 20.0, walletBalance(t, db, player.ID), 0.001)
}

// ─── CancelEntry ──────────────────────────────────────────────────────────────

func TestWcCustomBet_CancelEntry_HappyPath(t *testing.T) {
	db := openCustomBetTestDB(t)
	svc, authSvc := newCustomBetService(db)
	admin := seedCustomBetUser(t, db, authSvc, "Admin", 0)
	player := seedCustomBetUser(t, db, authSvc, "Player", 10)
	match := seedCustomBetMatch(t, db)
	bet, opts := seedCustomBet(t, svc, match.ID, admin.ID, model.WcCustomBetStatusOpen)

	require.NoError(t, svc.PlaceEntry(bet.ID, player.ID, opts[0].ID, 4, "test-user"))
	// Wallet unchanged at placement (deferred-deduction model)
	assert.InDelta(t, 10.0, walletBalance(t, db, player.ID), 0.001)

	var entry model.WcCustomBetEntry
	require.NoError(t, db.Where("custom_bet_id = ? AND wc_user_id = ?", bet.ID, player.ID).First(&entry).Error)

	require.NoError(t, svc.CancelEntry(entry.ID, player.ID))
	// Wallet still unchanged — cancel needs no refund (stake was never taken)
	assert.InDelta(t, 10.0, walletBalance(t, db, player.ID), 0.001)
	// Entry deleted
	assert.Error(t, db.First(&entry, entry.ID).Error)
}

func TestWcCustomBet_CancelEntry_WrongUser(t *testing.T) {
	db := openCustomBetTestDB(t)
	svc, authSvc := newCustomBetService(db)
	admin := seedCustomBetUser(t, db, authSvc, "Admin", 0)
	player1 := seedCustomBetUser(t, db, authSvc, "P1", 10)
	player2 := seedCustomBetUser(t, db, authSvc, "P2", 10)
	match := seedCustomBetMatch(t, db)
	bet, opts := seedCustomBet(t, svc, match.ID, admin.ID, model.WcCustomBetStatusOpen)

	require.NoError(t, svc.PlaceEntry(bet.ID, player1.ID, opts[0].ID, 3, "test-user"))
	var entry model.WcCustomBetEntry
	require.NoError(t, db.Where("custom_bet_id = ? AND wc_user_id = ?", bet.ID, player1.ID).First(&entry).Error)

	err := svc.CancelEntry(entry.ID, player2.ID)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "quyền")
}

func TestWcCustomBet_CancelEntry_BetClosed(t *testing.T) {
	db := openCustomBetTestDB(t)
	svc, authSvc := newCustomBetService(db)
	admin := seedCustomBetUser(t, db, authSvc, "Admin", 0)
	player := seedCustomBetUser(t, db, authSvc, "Player", 10)
	match := seedCustomBetMatch(t, db)
	bet, opts := seedCustomBet(t, svc, match.ID, admin.ID, model.WcCustomBetStatusOpen)

	require.NoError(t, svc.PlaceEntry(bet.ID, player.ID, opts[0].ID, 3, "test-user"))
	// Close the bet
	require.NoError(t, svc.repo.UpdateBet(nil, bet.ID, map[string]interface{}{"status": model.WcCustomBetStatusClosed}))

	var entry model.WcCustomBetEntry
	require.NoError(t, db.Where("custom_bet_id = ? AND wc_user_id = ?", bet.ID, player.ID).First(&entry).Error)

	err := svc.CancelEntry(entry.ID, player.ID)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "đóng")
}

// ─── Settle ───────────────────────────────────────────────────────────────────

func TestWcCustomBet_Settle_HappyPath(t *testing.T) {
	db := openCustomBetTestDB(t)
	svc, authSvc := newCustomBetService(db)
	admin := seedCustomBetUser(t, db, authSvc, "Admin", 0)
	winner := seedCustomBetUser(t, db, authSvc, "Winner", 10)
	loser := seedCustomBetUser(t, db, authSvc, "Loser", 10)
	match := seedCustomBetMatch(t, db)
	bet, opts := seedCustomBet(t, svc, match.ID, admin.ID, model.WcCustomBetStatusOpen)

	// winner bets on opts[0] (Có, odds 1.8), loser bets on opts[1] (Không, odds 2.0)
	require.NoError(t, svc.PlaceEntry(bet.ID, winner.ID, opts[0].ID, 5, "test-user"))
	require.NoError(t, svc.PlaceEntry(bet.ID, loser.ID, opts[1].ID, 3, "test-user"))

	require.NoError(t, svc.Settle(bet.ID, opts[0].ID, admin.ID))

	// Winner: payout = round(5 * 1.8 * 100)/100 = 9.00; net credit = payout - stake = 4; 10 + 4 = 14
	assert.InDelta(t, 14.0, walletBalance(t, db, winner.ID), 0.001)
	// Loser: stake deducted at settlement; 10 - 3 = 7
	assert.InDelta(t, 7.0, walletBalance(t, db, loser.ID), 0.001)

	// Check entries
	var winEntry, loseEntry model.WcCustomBetEntry
	require.NoError(t, db.Where("custom_bet_id = ? AND wc_user_id = ?", bet.ID, winner.ID).First(&winEntry).Error)
	require.NoError(t, db.Where("custom_bet_id = ? AND wc_user_id = ?", bet.ID, loser.ID).First(&loseEntry).Error)
	assert.Equal(t, model.WcCustomBetEntryStatusWon, winEntry.Status)
	assert.InDelta(t, 9.0, *winEntry.Payout, 0.001)
	assert.Equal(t, model.WcCustomBetEntryStatusLost, loseEntry.Status)

	// Bet status
	var updatedBet model.WcCustomBet
	require.NoError(t, db.First(&updatedBet, bet.ID).Error)
	assert.Equal(t, model.WcCustomBetStatusSettled, updatedBet.Status)
	assert.NotNil(t, updatedBet.SettledAt)
	assert.Equal(t, admin.ID, *updatedBet.SettledBy)

	// Winning option flagged
	var winOpt model.WcCustomBetOption
	require.NoError(t, db.First(&winOpt, opts[0].ID).Error)
	assert.True(t, winOpt.IsWinner)
}

func TestWcCustomBet_Settle_PayoutRounding(t *testing.T) {
	// 3 * 1.8 = 5.4 — verify no floating point drift
	db := openCustomBetTestDB(t)
	svc, authSvc := newCustomBetService(db)
	admin := seedCustomBetUser(t, db, authSvc, "Admin", 0)
	player := seedCustomBetUser(t, db, authSvc, "Player", 10)
	match := seedCustomBetMatch(t, db)
	bet, opts := seedCustomBet(t, svc, match.ID, admin.ID, model.WcCustomBetStatusOpen)

	require.NoError(t, svc.PlaceEntry(bet.ID, player.ID, opts[0].ID, 3, "test-user"))
	require.NoError(t, svc.Settle(bet.ID, opts[0].ID, admin.ID))

	// payout = round(3 * 1.8 * 100)/100 = 5.4; net credit = 5.4 - 3 = 2.4; 10 + 2.4 = 12.4
	assert.InDelta(t, 12.4, walletBalance(t, db, player.ID), 0.001)
}

func TestWcCustomBet_Settle_AlreadySettled(t *testing.T) {
	db := openCustomBetTestDB(t)
	svc, authSvc := newCustomBetService(db)
	admin := seedCustomBetUser(t, db, authSvc, "Admin", 0)
	match := seedCustomBetMatch(t, db)
	bet, opts := seedCustomBet(t, svc, match.ID, admin.ID, model.WcCustomBetStatusOpen)

	require.NoError(t, svc.Settle(bet.ID, opts[0].ID, admin.ID))
	err := svc.Settle(bet.ID, opts[0].ID, admin.ID)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "tất toán")
}

func TestWcCustomBet_Settle_WinningOptionFromDifferentBet(t *testing.T) {
	db := openCustomBetTestDB(t)
	svc, authSvc := newCustomBetService(db)
	admin := seedCustomBetUser(t, db, authSvc, "Admin", 0)
	match := seedCustomBetMatch(t, db)
	bet1, _ := seedCustomBet(t, svc, match.ID, admin.ID, model.WcCustomBetStatusOpen)
	_, opts2 := seedCustomBet(t, svc, match.ID, admin.ID, model.WcCustomBetStatusOpen)

	err := svc.Settle(bet1.ID, opts2[0].ID, admin.ID)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "hợp lệ")
}

// ─── VoidBet ─────────────────────────────────────────────────────────────────

func TestWcCustomBet_Void_HappyPath(t *testing.T) {
	db := openCustomBetTestDB(t)
	svc, authSvc := newCustomBetService(db)
	admin := seedCustomBetUser(t, db, authSvc, "Admin", 0)
	p1 := seedCustomBetUser(t, db, authSvc, "P1", 10)
	p2 := seedCustomBetUser(t, db, authSvc, "P2", 10)
	match := seedCustomBetMatch(t, db)
	bet, opts := seedCustomBet(t, svc, match.ID, admin.ID, model.WcCustomBetStatusOpen)

	require.NoError(t, svc.PlaceEntry(bet.ID, p1.ID, opts[0].ID, 4, "test-user"))
	require.NoError(t, svc.PlaceEntry(bet.ID, p2.ID, opts[1].ID, 3, "test-user"))

	require.NoError(t, svc.VoidBet(bet.ID))

	// Wallet unchanged — stake was never deducted at placement, void makes no wallet changes
	assert.InDelta(t, 10.0, walletBalance(t, db, p1.ID), 0.001)
	assert.InDelta(t, 10.0, walletBalance(t, db, p2.ID), 0.001)

	// Entry statuses
	var e1, e2 model.WcCustomBetEntry
	require.NoError(t, db.Where("wc_user_id = ? AND custom_bet_id = ?", p1.ID, bet.ID).First(&e1).Error)
	require.NoError(t, db.Where("wc_user_id = ? AND custom_bet_id = ?", p2.ID, bet.ID).First(&e2).Error)
	assert.Equal(t, model.WcCustomBetEntryStatusVoid, e1.Status)
	assert.Equal(t, model.WcCustomBetEntryStatusVoid, e2.Status)

	// Bet status
	var updatedBet model.WcCustomBet
	require.NoError(t, db.First(&updatedBet, bet.ID).Error)
	assert.Equal(t, model.WcCustomBetStatusVoid, updatedBet.Status)
}

func TestWcCustomBet_Void_AlreadyVoid(t *testing.T) {
	db := openCustomBetTestDB(t)
	svc, authSvc := newCustomBetService(db)
	admin := seedCustomBetUser(t, db, authSvc, "Admin", 0)
	match := seedCustomBetMatch(t, db)
	bet, _ := seedCustomBet(t, svc, match.ID, admin.ID, model.WcCustomBetStatusOpen)

	require.NoError(t, svc.VoidBet(bet.ID))
	err := svc.VoidBet(bet.ID)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "huỷ rồi")
}

func TestWcCustomBet_Void_AlreadySettled(t *testing.T) {
	db := openCustomBetTestDB(t)
	svc, authSvc := newCustomBetService(db)
	admin := seedCustomBetUser(t, db, authSvc, "Admin", 0)
	match := seedCustomBetMatch(t, db)
	bet, opts := seedCustomBet(t, svc, match.ID, admin.ID, model.WcCustomBetStatusOpen)

	require.NoError(t, svc.Settle(bet.ID, opts[0].ID, admin.ID))
	err := svc.VoidBet(bet.ID)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "tất toán")
}
