# WC Custom Bet (Kèo Phụ)

## Overview

Admin-defined proposition bets attached to a WC match. Admin creates a bet with a title, optional line value, and 2–10 options (each with its own odds). Players pick one option and stake points. Admin manually settles by selecting the winning option. Completely separate tables from `wc_bets` — custom bets have dynamic option counts which the fixed `wc_bets` schema cannot support.

## Tables

| Table | Purpose |
|-------|---------|
| `wc_custom_bets` | One row per proposition: match, title, line NUMERIC(6,2), status, created_by, settled_at, settled_by |
| `wc_custom_bet_options` | N options per bet: label, odds NUMERIC(6,2), is_winner, display_order |
| `wc_custom_bet_entries` | One entry per player per bet: stake, odds_snapshot, payout NUMERIC(10,2), status |

Key constraints on `wc_custom_bet_entries`:
- `UNIQUE (custom_bet_id, wc_user_id)` — one entry per player per bet (enforced at DB level via `idx_custom_bet_entry_dedup`)

## Status Flows

**Bet status:** `open` → `closed` → `settled` | `void`
- `open`: players can place and cancel entries
- `closed`: no new entries; admin can re-open or settle
- `settled`: winner chosen, payouts credited — immutable
- `void`: all stakes cancelled — immutable

**Entry status:** `pending` → `won` | `lost` | `void`

## Key Rules

- `odds_snapshot` is captured at entry time — admin changing option odds does not affect existing entries
- One entry per player per bet; player must cancel existing entry before changing option
- Cancel only allowed when bet is `open` and entry is `pending`
- Stake min/max enforced from `wc_config.min_points` / `max_points`
- **No balance check at placement** — stake is NOT deducted when placing an entry (same as `wc_bets` and `wc_predictions`). Wallet is updated only at settlement.

## Wallet Lifecycle

Custom bets follow the same deferred-deduction model as regular `wc_bets`:

| Event | Wallet change |
|-------|--------------|
| PlaceEntry | **None** — stake is recorded on the entry but not moved |
| CancelEntry | **None** — nothing was deducted, so nothing to refund |
| Settle (winner) | `+(payout - stake)` — net profit only |
| Settle (loser) | `-stake` |
| VoidBet | **None** — stake was never deducted |

This means users with 0 balance can place custom bets, identical to regular predictions.

## Payout Formula

```
payout = math.Round(stake × odds_snapshot × 100) / 100
```

Winner net gain: `payout - stake` = `stake × (odds - 1)`

## Settlement (Tất Toán)

Full `db.Transaction` wrapping:
1. Mark winning option: `is_winner = true`
2. For each pending entry:
   - Winner: calculate payout, `UpdateEntryResult(won, payout)`, credit `payout - stake` to wallet
   - Loser: `UpdateEntryResult(lost, 0)`, deduct stake from wallet
3. Update bet: `status = settled`, `settled_at`, `settled_by`

**Void** (also transactional): for each pending entry, `UpdateEntryResult(void, 0)`. Set bet `status = void`. **No wallet changes** (stake was never taken).

## PlaceEntry / CancelEntry Atomicity

Both wrapped in `db.Transaction`:
- **PlaceEntry:** `CreateEntry(tx)` only — no wallet deduction. Duplicate key on `idx_custom_bet_entry_dedup` returns user-facing error "bạn đã đặt cược cho kèo này rồi".
- **CancelEntry:** `DeleteEntry(tx)` only — no wallet refund needed.

## FinalizeMatch Integration

`WcService.FinalizeMatch` returns `FinalizeMatchResult` which includes `unsettled_custom_bets_count`. After regular bet settlement, the service queries `customBetRepo.CountUnsettledForMatch(matchID)`. If > 0, the frontend store shows an `ElMessage.warning` reminding admin to settle pending custom bets for that match.

Similarly, `FinalizeAllMatches` result includes `matches_with_unsettled_custom_bets int64`.

## View Models

### WcCustomBetWithOptions
Standard response shape for all custom bet list endpoints:
```go
type WcCustomBetWithOptions struct {
    WcCustomBet
    Options    []WcCustomBetOption      `json:"options"`
    MyEntry    *WcCustomBetEntry        `json:"my_entry,omitempty"`
    EntryCount int                      `json:"entry_count"`
    Entries    []WcCustomBetEntryPublic `json:"entries"`  // all entries with user info
}
```

`WcCustomBetEntryPublic` includes: id, wc_user_id, option_id, option_label, name, avatar_url, stake, odds_snapshot, status, payout.

`MyEntry` is derived from `Entries` in Go (no extra DB query).

### WcCustomBetEntryHistory
Used by `GET /custom-bet-entries` (my history). JOINs entries with bets + options + matches to provide bet_title, option_label, home_team, away_team, match_date.

## Batched Loading

`listWithOptionsAndEntries(bets)` loads options and entries for all bets in a match with **2 batch queries** (one `WHERE custom_bet_id IN (...)` for options, one for entries), then assembles in Go. Total: 3 DB round-trips per match regardless of how many custom bets exist.

## Integration with Predictions List

Custom bet entries are surfaced in two places that were originally predictions-only:

### GET /matches/:id/predictions (public match predictions panel)
`WcService.ListPredictionsForMatchPublic` merges regular predictions + `customBetRepo.ListCustomEntriesForMatchPublic(matchID)`. Custom entries are adapted to `WcPredictionPublic`:
- `prediction_type = "custom"`
- `prediction_choice = option_label`
- `points = stake`, `multiplier_snapshot = odds_snapshot`
- `result`: `won→correct`, `lost→incorrect`, `void→void`, `pending→null`
- `points_earned = payout - stake` (net profit, for direct display)
- `bet_title` set — frontend shows as orange chip before option label

### GET /predictions (user's own prediction history)
`WcService.ListPredictions` merges regular predictions + `customBetRepo.ListCustomEntriesForUserAsHistory(userID)`. Adapted to `WcPredictionWithMatch`:
- Same field mapping as above, except `points_earned = payout` (full payout, for `(points_earned - points)` display formula in history list)
- `predictions_open = false` — no edit/delete buttons in history list
- Pending entries show a **Huỷ** button in `WcPredictionHistoryList`
- Settled entries appear in "Lịch sử" automatically (result is truthy)

`WcPredictionWithMatch` and `WcPredictionPublic` both carry a `bet_title *string` field (nil for non-custom rows).

## Files

| Layer | File |
|-------|------|
| Models | `backend/internal/model/wc_custom_bet.go` |
| Repository | `backend/internal/repository/wc_custom_bet_repository.go` |
| Service | `backend/internal/service/wc_custom_bet_service.go` |
| Handler | `backend/internal/api/wc_custom_bet_handler.go` |
| Routes | `backend/internal/api/router.go` (wcAdmin + wcAuth groups) |
| Admin UI | `frontend/src/components/wc/WcAdminCustomBetPanel.vue` |
| Player card | `frontend/src/components/wc/WcCustomBetCard.vue` — options, my entry, all entries list, place/cancel |
| History list | `frontend/src/components/wc/WcCustomBetHistoryList.vue` — standalone custom history (used in WcBettingView) |
| Prediction history | `frontend/src/components/wc/WcPredictionHistoryList.vue` — handles `prediction_type === 'custom'` inline |
| Types | `frontend/src/types/wc.ts` |
| Service calls | `frontend/src/services/wcService.ts` |

## API Routes

### Admin (requires `WcAdminMiddleware`)
```
GET  /api/v1/wc/admin/matches/:id/custom-bets
POST /api/v1/wc/admin/matches/:id/custom-bets   body: { title, line?, options: [{label, odds, display_order}] }
PUT  /api/v1/wc/admin/custom-bets/:id           body: { title?, line?, status? }
POST /api/v1/wc/admin/custom-bets/:id/settle    body: { winning_option_id }
PUT  /api/v1/wc/admin/custom-bets/:id/void
```

### Player (requires `WcAuthMiddleware`)
```
GET    /api/v1/wc/matches/:id/custom-bets      → WcCustomBetWithOptions[] (includes my_entry, entries[])
POST   /api/v1/wc/custom-bets/:id/entry        body: { option_id, stake }
DELETE /api/v1/wc/custom-bet-entries/:id
GET    /api/v1/wc/custom-bet-entries           → WcCustomBetEntryHistory[] (user's own entries)
```

### Merged into existing endpoints
```
GET /api/v1/wc/matches/:id/predictions   → also includes custom entries as prediction_type="custom"
GET /api/v1/wc/predictions               → also includes custom entries, supports Huỷ from history tab
```

## Frontend Access Points

| Where | How custom bets are accessed |
|-------|------------------------------|
| `/world-cup/predict` — Dự đoán tab | "Kèo phụ" tab inside `WcPredictionForm` dialog |
| `/world-cup/predict` — Tất cả dự đoán panel | Merged into predictions list via `/matches/:id/predictions` |
| `/world-cup/predict` — Đang chờ / Lịch sử tabs | Merged into `store.predictions` via `GET /predictions`; cancel button shown for pending entries |
| `/world-cup/bet` — Betting tab | Inline "Kèo phụ" toggle below each match card → `WcCustomBetCard` |
| `/world-cup/bet` — Đang chờ / Lịch sử tabs | Separate `WcCustomBetHistoryList` sections below regular bet history |
