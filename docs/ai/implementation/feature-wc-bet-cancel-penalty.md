---
phase: implementation
title: WC Bet Cancel Penalty — Implementation Guide
description: Key patterns, gotchas, and code reference for implementing cancel penalty
---

# Implementation Guide

## Development Setup

No new dependencies needed. All tools already in use:
- `github.com/shopspring/decimal` — penalty arithmetic
- GORM AutoMigrate — column additions
- Existing DB transaction helpers

For the `wc_wallet_logs.admin_id` nullable change: AutoMigrate will NOT drop the NOT NULL constraint on an existing column. You must run a manual migration:

```sql
ALTER TABLE wc_wallet_logs ALTER COLUMN admin_id DROP NOT NULL;
```

Add this to `backend/internal/database/database.go` after `AutoMigrate()` or in a dedicated migration function.

---

## Code Structure

Files to modify:
```
backend/internal/model/
  wc_match.go          — add CancelledAt to WcBet, nullable AdminID to WcWalletLog
  wc_user.go           — add CancelPenaltyEnabled/Percent to WcConfig

backend/internal/repository/
  wc_repository.go     — SoftCancelBet, CreateWalletLogSystem, GetWalletTx, ListBetHistoryForUser
                          + audit pending-bet queries for cancelled_at guard

backend/internal/service/
  wc_service.go        — rewrite DeleteBet(), add GetBetHistory(), validate cancel_penalty_percent

backend/internal/api/
  wc_handler.go        — add GetBetHistory handler
  router.go            — wire GET /wc/bets/history

frontend/src/types/
  wc.ts                — WcConfig + WcBet interface updates

frontend/src/services/
  wcService.ts         — add getBetHistory()

frontend/src/locales/
  vi.json / en.json    — cancel penalty i18n keys

frontend/src/views/
  WcBettingView.vue    — history tab data source, cancel dialog with penalty warning

frontend/src/components/wc/
  WcBetHistoryList.vue — Đã hủy badge for cancelled bets
  WcAdminPanel.vue     — cancel penalty config controls
```

---

## Implementation Notes

### Pending-bet query audit (critical)

All queries that find "active pending" bets must be updated from:
```go
Where("result IS NULL")
```
to:
```go
Where("result IS NULL AND (cancelled_at IS NULL OR cancelled_at = '0001-01-01 00:00:00+00')")
```

Key locations to check in `wc_repository.go`:
- `ListBetsForSettlement` (settlement logic — must not settle cancelled bets)
- Any per-match open-bet query used in `GET /wc/matches/:id/bets`
- `ListPendingBetsForUser` (used in BlockUser — cancelled bets should not be voided again)

### Penalty arithmetic

```go
import "github.com/shopspring/decimal"

stakeD   := decimal.NewFromInt(int64(bet.Stake))
percentD := decimal.NewFromInt(int64(cfg.CancelPenaltyPercent))
penaltyD := stakeD.Mul(percentD).Div(decimal.NewFromInt(100)).Floor()
penalty  := penaltyD.IntPart() // guaranteed non-negative integer
```

### Balance cap at zero

```go
wallet, err := s.repo.GetWalletTx(tx, wcUserID)
balBefore := wallet.Balance
deduction := math.Min(float64(penalty), balBefore)
if deduction <= 0 {
    // nothing to deduct — skip wallet ops
    return nil
}
```

`math.Min` is fine here because we're comparing two non-negative values for a floor operation, not doing money arithmetic.

### WcWalletLog construction — nullable AdminID

All existing sites that create `WcWalletLog` use:
```go
// Old (must update)
AdminID: someAdminUUID,

// New — admin-initiated
adminID := someAdminUUID
log := &model.WcWalletLog{ AdminID: &adminID, ... }

// New — system-initiated (cancel penalty)
log := &model.WcWalletLog{ AdminID: nil, ... }
```

Grep for `WcWalletLog{` to find all construction sites.

### SoftCancelBet repository method

```go
func (r *WcRepository) SoftCancelBet(tx *gorm.DB, id, wcUserID uuid.UUID) error {
    db := r.db
    if tx != nil {
        db = tx
    }
    now := time.Now()
    result := db.Model(&model.WcBet{}).
        Where("id = ? AND wc_user_id = ?", id, wcUserID).
        Update("cancelled_at", now)
    if result.Error != nil {
        return result.Error
    }
    if result.RowsAffected == 0 {
        return errors.New("bet not found or not authorized")
    }
    return nil
}
```

### ListBetHistoryForUser repository method

```go
func (r *WcRepository) ListBetHistoryForUser(wcUserID uuid.UUID) ([]*model.WcBet, error) {
    var bets []*model.WcBet
    err := r.db.
        Where("wc_user_id = ? AND (result IS NOT NULL OR cancelled_at IS NOT NULL)", wcUserID).
        Order("created_at DESC").
        Find(&bets).Error
    return bets, err
}
```

### Frontend cancel flow

```ts
// In WcBettingView.vue or a composable
async function handleCancelBet(bet: WcBet) {
  const cfg = store.config
  if (cfg?.cancel_penalty_enabled) {
    const penalty = Math.floor(bet.stake * cfg.cancel_penalty_percent / 100)
    if (penalty > 0) {
      try {
        await ElMessageBox.confirm(
          t('wc.cancelPenaltyWarning', { penalty, percent: cfg.cancel_penalty_percent }),
          t('wc.cancelBetTitle'),
          { type: 'warning', confirmButtonText: t('common.confirm'), cancelButtonText: t('common.cancel') }
        )
      } catch {
        return // user dismissed dialog
      }
    }
  }
  await wcService.deleteBet(bet.id)
  await refreshBets()
}
```

---

## Integration Points

- **`WcService.GetConfig()`** — already exists; DeleteBet will call it at start
- **`s.repo.UpdateWalletBalance(tx, userID, delta)`** — already exists for positive delta; pass negative delta for deduction
- **WebSocket broadcast** in DeleteBet — keep existing `buildCancelActivityEvent` call; it fires after the transaction commits

---

## Error Handling

- `GetConfig` failure → return error immediately (don't proceed without knowing penalty config)
- DB transaction failure → all changes rolled back atomically (no partial cancel + no partial wallet deduction)
- Wallet not found → treat as 0 balance (no deduction); log warning
- `SoftCancelBet` 0 rows affected → `"bet not found or not authorized"` (same message as current hard-delete)

---

## Security Notes

- Service still checks `bet.WcUserID != wcUserID` before cancelling — ownership guard unchanged
- Service still checks `bet.Result != nil` — cannot cancel settled bets
- Service still checks `isBetLocked(m)` — cannot cancel when match bets are locked
- Admin cannot cancel user bets via this endpoint — admin has separate void endpoint
