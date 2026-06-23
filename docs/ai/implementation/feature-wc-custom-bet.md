---
phase: implementation
title: WC Custom Bet — Implementation Guide
description: Implementation notes for custom proposition bets (kèo phụ)
---

# Implementation Guide

## Development Setup

No new dependencies. Uses existing Go/GORM, PostgreSQL, Vue 3, Pinia, Element Plus.

```bash
cd backend && go run cmd/server/main.go
cd frontend && npm run dev
```

## Code Structure

```
backend/internal/
  model/wc_custom_bet.go                 ← NEW: 3 models + view model
  repository/wc_custom_bet_repository.go ← NEW
  service/wc_custom_bet_service.go       ← NEW
  api/wc_custom_bet_handler.go           ← NEW
  api/router.go                          ← wire 7 new routes
  database/database.go                   ← add 3 models to AutoMigrate

frontend/src/
  types/wc.ts                            ← add 3 new interfaces
  services/wcService.ts                  ← add 8 new API methods
  utils/wcBetType.ts                     ← register 'custom' type
  locales/vi.json                        ← add betTypeCustom key
  components/wc/
    WcCustomBetCard.vue                  ← NEW: player-facing
    WcAdminCustomBetPanel.vue            ← NEW: admin create/settle/void
```

## Implementation Notes

### Phase 1: Backend Models

**`wc_custom_bet.go`** — key points:
- `WcCustomBetEntry` has `UNIQUE (custom_bet_id, wc_user_id)` enforced at DB level
- `OddsSnapshot` is `NUMERIC(6,2)` stored at bet time — never updated after creation
- `Payout` is `*float64` nullable — set only on settlement

**`database.go`** — add to AutoMigrate:
```go
&model.WcCustomBet{},
&model.WcCustomBetOption{},
&model.WcCustomBetEntry{},
```

### Phase 1: Repository

**`ListForMatchWithMyEntry`** — most complex query; join options + left join on user's entry:
```go
func (r *WcCustomBetRepository) ListForMatchWithMyEntry(matchID, userID uuid.UUID) ([]*WcCustomBetWithOptions, error) {
    var bets []model.WcCustomBet
    r.db.Where("match_id = ?", matchID).Order("created_at ASC").Find(&bets)
    // For each bet, load options + find user's entry
    // Return assembled WcCustomBetWithOptions slice
}
```

**`PlaceEntry`** — do NOT deduct wallet here; service handles wallet:
```go
func (r *WcCustomBetRepository) CreateEntry(entry *model.WcCustomBetEntry) error {
    return r.db.Create(entry).Error
    // DB UNIQUE constraint handles double-bet protection
}
```

### Phase 1: Service — Settlement Transaction

```go
func (s *WcCustomBetService) Settle(betID, winningOptionID, adminID uuid.UUID) error {
    return s.db.Transaction(func(tx *gorm.DB) error {
        // 1. Lock and load bet
        var bet model.WcCustomBet
        tx.Set("gorm:query_option", "FOR UPDATE").First(&bet, betID)
        if bet.Status != "open" && bet.Status != "closed" {
            return fmt.Errorf("bet is already settled or void")
        }

        // 2. Mark winning option
        tx.Model(&model.WcCustomBetOption{}).
            Where("id = ?", winningOptionID).
            Update("is_winner", true)

        // 3. Load all entries
        var entries []model.WcCustomBetEntry
        tx.Where("custom_bet_id = ? AND status = 'pending'", betID).Find(&entries)

        // 4. Settle each entry
        now := time.Now()
        for _, entry := range entries {
            if entry.OptionID == winningOptionID {
                payout := math.Round(float64(entry.Stake)*entry.OddsSnapshot*100) / 100
                tx.Model(&entry).Updates(map[string]any{"status": "won", "payout": payout})
                // Credit wallet
                s.walletRepo.Credit(tx, entry.WcUserID, payout, "custom_bet_win", entry.ID.String())
            } else {
                tx.Model(&entry).Updates(map[string]any{"status": "lost", "payout": 0})
            }
        }

        // 5. Update bet status
        tx.Model(&bet).Updates(map[string]any{
            "status": "settled", "settled_at": &now, "settled_by": adminID,
        })
        return nil
    })
}
```

### Phase 1: Service — PlaceEntry

```go
func (s *WcCustomBetService) PlaceEntry(betID, userID, optionID uuid.UUID, stake int) error {
    // Validate limits
    cfg, _ := s.wcRepo.GetConfig()
    if stake < cfg.MinPoints || stake > cfg.MaxPoints {
        return fmt.Errorf("điểm cược phải từ %d đến %d", cfg.MinPoints, cfg.MaxPoints)
    }
    // Validate bet is open
    bet, _ := s.repo.GetByID(betID)
    if bet.Status != "open" {
        return fmt.Errorf("kèo đã đóng")
    }
    // Validate option belongs to bet
    option := findOption(bet.Options, optionID)
    if option == nil {
        return fmt.Errorf("lựa chọn không hợp lệ")
    }
    // Deduct wallet
    s.walletRepo.Debit(userID, stake, "custom_bet_stake", betID.String())
    // Create entry with odds snapshot
    entry := &model.WcCustomBetEntry{
        CustomBetID:  betID,
        OptionID:     optionID,
        WcUserID:     userID,
        Stake:        stake,
        OddsSnapshot: option.Odds,
        Status:       "pending",
    }
    return s.repo.CreateEntry(entry)
    // UNIQUE constraint returns error if player already has an entry
}
```

### Phase 2: Admin Component — `WcAdminCustomBetPanel.vue`

Key UI states:
1. **Create mode:** Title input + dynamic options list (add/remove rows, label + odds per row) + Submit
2. **List mode:** Each existing bet shows title, status badge, entries count, action buttons (Settle / Close / Void)
3. **Settle dialog:** Dropdown/radio to select winning option → confirm button

### Phase 3: Player Component — `WcCustomBetCard.vue`

Key UI states:
1. **Open + no entry:** Options list with odds chips, stake input, "Đặt cược" button
2. **Open + has entry:** Show my entry (chosen option + stake); "Huỷ" button to cancel
3. **Closed + pending:** Show my entry; "Đang chờ kết quả"
4. **Settled:** Highlight winning option; show my result (Thắng +X điểm / Thua)
5. **Void:** Show "Kèo đã huỷ — đã hoàn tiền"

### Phase 4: Label

In `wcBetType.ts`:
```ts
export const WC_BET_TYPES = ['handicap', 'exact_score', 'over_under', 'corner', 'custom'] as const
export const BET_TYPE_I18N_KEYS = {
  // existing...
  custom: 'wc.betTypeCustom',
}
```

In `vi.json`:
```json
"betTypeCustom": "Kèo phụ"
```

## Integration Points

- `WcCustomBetService.PlaceEntry` uses the same `walletRepo.Debit` pattern as `WcService.PlaceBet`
- `WcCustomBetService.Settle` uses the same `walletRepo.Credit` pattern as `WcService.settle`
- All wallet changes write to `wc_wallet_logs` — custom bets appear in the same audit trail
- House P&L: sum(stake) − sum(payout) across `wc_custom_bet_entries` for a match

## Error Handling

- `UNIQUE (custom_bet_id, wc_user_id)` violation → return "Bạn đã đặt cược cho kèo này rồi"
- Settlement on already-settled bet → 400 "Kèo đã tất toán"
- Cancel after bet closed → 400 "Kèo đã đóng, không thể huỷ"
- Wallet insufficient balance → 400 "Số dư không đủ"

## Security Notes

- All admin endpoints behind `WcAdminMiddleware`
- All player endpoints behind `WcAuthMiddleware` + `WcFeatureMiddleware`
- `CancelEntry` validates `entry.WcUserID == callerID` before deleting
