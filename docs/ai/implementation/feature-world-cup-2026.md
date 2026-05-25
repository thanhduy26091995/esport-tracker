---
phase: implementation
title: World Cup 2026 — Implementation Guide
description: Technical notes, patterns, and gotchas for the WC2026 tracker + betting feature
---

# Implementation Guide

## Development Setup

1. Register a free account at https://www.football-data.org/ → get API key
2. Add to backend `.env`:
   ```
   FOOTBALL_DATA_API_KEY=your_key_here
   ```
3. Run the DB migration (Phase 1.1) before starting any service work
4. WC2026 competition code in football-data.org API = `WC`

---

## Code Structure

```
backend/internal/
  model/wc_match.go
  repository/wc_repository.go
  service/
    wc_service.go
    wc_football_client.go
  api/wc_handler.go

frontend/src/
  types/wc.ts
  services/wcService.ts
  stores/wcStore.ts
  views/
    WcScheduleView.vue
    WcBettingView.vue
  components/wc/
    WcMatchCard.vue
    WcGroupFilter.vue
    WcBetForm.vue
    WcBetHistoryList.vue
    WcLeaderboard.vue
```

---

## Implementation Notes

### football-data.org response mapping

The API returns matches under `/v4/competitions/WC/matches`. Key fields:

```json
{
  "id": 123456,
  "utcDate": "2026-06-11T18:00:00Z",
  "status": "SCHEDULED",         // SCHEDULED | LIVE | IN_PLAY | FINISHED | CANCELLED
  "stage": "GROUP_STAGE",        // GROUP_STAGE | LAST_32 | LAST_16 | QUARTER_FINALS | SEMI_FINALS | FINAL | THIRD_PLACE
  "group": "GROUP_A",
  "homeTeam": { "name": "Brazil", "tla": "BRA" },
  "awayTeam": { "name": "Argentina", "tla": "ARG" },
  "score": {
    "fullTime": { "home": null, "away": null }
  },
  "venue": "SoFi Stadium"
}
```

Status mapping: `SCHEDULED` → `scheduled`, `IN_PLAY`/`LIVE` → `live`, `FINISHED` → `completed`, `CANCELLED` → `cancelled`

Stage mapping: `GROUP_STAGE` → `group`, `LAST_32` → `r32`, `LAST_16` → `r16`, `QUARTER_FINALS` → `qf`, `SEMI_FINALS` → `sf`, `FINAL` → `final`, `THIRD_PLACE` → `third_place`

### Settlement idempotency

Re-running `SettleMatch(matchID)` must be safe (admin may correct a wrong score):

```
BEGIN;
  -- 1. Reverse all previous payouts for this match
  SELECT bets WHERE match_id = X AND result IS NOT NULL → previous settled bets
  FOR each bet: UPDATE wc_wallets SET balance = balance - bet.payout WHERE user_id = bet.user_id
  UPDATE wc_bets SET result = NULL, payout = NULL WHERE match_id = X

  -- 2. Re-evaluate all bets with fresh score
  FOR each bet: evaluate → result, payout
  UPDATE wc_bets SET result = ..., payout = ... WHERE id = ...
  UPDATE wc_wallets SET balance = balance + payout WHERE user_id = ...
  
  -- 3. Mark settled
  UPDATE wc_matches SET settled_at = NOW() WHERE id = X
COMMIT;
```

### Handicap evaluation (Go pseudo-code)

```go
func evaluateHandicapBet(bet WcBet, homeScore, awayScore int) (string, int) {
    hv := float64(bet.HandicapSnapshot)
    var adjHome float64
    if bet... // use handicap_team and value from bet snapshot
    // If handicap_team = "home": home gives goals → adjHome = homeScore - hv
    // If handicap_team = "away": away gives goals → adjHome = homeScore + hv

    var winner string
    switch {
    case adjHome > float64(awayScore):
        winner = "home"
    case adjHome < float64(awayScore):
        winner = "away"
    default:
        // push (only with whole-number handicap)
        return "push", bet.Stake
    }

    if bet.BetChoice == winner {
        return "win", int(float64(bet.Stake) * float64(bet.OddsSnapshot))
    }
    return "lose", 0
}
```

### Wallet transaction safety

Always wrap wallet deduction + bet creation in a single DB transaction. No balance check — negative balance is allowed.

```go
tx := db.Begin()
// SELECT wc_wallets WHERE user_id = ? FOR UPDATE  (wallet always exists — created at registration)
// duplicate check: handicap → unique on (user_id, match_id, bet_type, bet_choice)
//                  exact_score → unique on (user_id, match_id, predicted_home_score, predicted_away_score)
// UPDATE wc_wallets SET balance = balance - stake WHERE user_id = ?  (no lower bound)
// INSERT INTO wc_bets ...
tx.Commit()
```

### Settlement transaction

All steps run in a single transaction — no partial state possible:

```go
tx := db.Begin()
// 1. Load all wallets with user name: SELECT wc_wallets JOIN users
// 2. INSERT INTO wc_settlements (name, point_rate, settled_by, note)
// 3. For each wallet:
//      direction = "pay" if balance > 0, "collect" if balance < 0, "even" if = 0
//      amount = abs(balance) * point_rate
//      INSERT INTO wc_settlement_details (settlement_id, user_id, final_balance, amount, direction)
// 4. UPDATE wc_wallets SET balance = 0, updated_at = NOW()
tx.Commit()
```

Re-running settlement on an already-reset wallet (all balances = 0) will produce a valid but empty settlement — admin should avoid this, but it is harmless.

### Settlement preview (no DB write)

```go
func PreviewSettlement(pointRate float64) []SettlementRow {
    wallets := repo.ListAllWallets()
    rows := make([]SettlementRow, len(wallets))
    for i, w := range wallets {
        direction := directionFor(w.Balance)
        rows[i] = SettlementRow{
            UserID:       w.UserID,
            Name:         w.Name,
            FinalBalance: w.Balance,
            Amount:       math.Abs(float64(w.Balance)) * pointRate,
            Direction:    direction,
        }
    }
    return rows
}
```

### Exact score evaluation (Go pseudo-code)

```go
func evaluateExactScoreBet(bet WcBet, homeScore, awayScore int) (string, int) {
    if bet.PredictedHomeScore == homeScore && bet.PredictedAwayScore == awayScore {
        return "win", int(float64(bet.Stake) * float64(bet.OddsSnapshot))
    }
    return "lose", 0
}
```

### UI Language — Points only, no gambling terms

All frontend text must use fun/game language. Backend field names (`bet`, `stake`, `payout`) are internal only — never shown directly in the UI.

| Backend term | UI display (Vietnamese) |
|---|---|
| Place bet | Tham gia dự đoán |
| Bet / bet_type | Dự đoán |
| Stake | Điểm tham gia |
| Payout | Điểm nhận được |
| Wallet | Ví điểm WC |
| Win (result) | Dự đoán đúng ✓ |
| Lose (result) | Dự đoán sai ✗ |
| Push | Hoà chấp |
| Bet history | Lịch sử dự đoán |
| Leaderboard | Bảng xếp hạng dự đoán |
| Handicap bet | Dự đoán có chấp |
| Exact score bet | Dự đoán tỉ số |
| Odds | Hệ số nhân |
| Settle match | Tất toán điểm |
| Top-up wallet | Cộng điểm |
| Net profit | Điểm thưởng tích lũy |
| Settlement preview | Xem trước tất toán |
| Create settlement | Tạo tất toán |
| Settlement history | Lịch sử tất toán |
| Direction: pay | Admin chi (user thắng) |
| Direction: collect | Admin thu (user thua) |
| Direction: even | Hoà vốn |
| Mark done | Đánh dấu đã xong |
| point_rate | Tỉ lệ quy đổi (điểm/VND) |

### Frontend form — live preview

While user types stake amount, show:
- For handicap tab: pick Home or Away button (one active per side); `Điểm nhận được nếu đúng = điểm tham gia × hệ số` + handicap string (e.g., "France -1.5 × 1.90"). User can activate both sides independently.
- For exact score tab: grid of scoreline cards — user can select **multiple** cards; each selected card expands to show a "Điểm tham gia" input + `Điểm nhận được = X × hệ số` preview. Submitting sends one `POST /bets` per selected scoreline sequentially; partial failure surfaced without rolling back already-placed predictions.

### Admin workflow

1. **Before tournament**: Admin clicks "Đồng bộ lịch thi đấu" → fixtures imported. All wallets start at 0 automatically — no init step needed.
2. **Before each match**: Admin opens match → enters hệ số chấp + hệ số nhân → saves (predictions now open for users)
3. **During tournament** (optional): Admin syncs to update scores
4. **After match**: Admin confirms final score → clicks "Tất toán trận" → system evaluates all predictions and credits/debits points (balance can go negative for losers)
5. **End of tournament (or any time)**: Admin opens "Tất toán giải" tab → sets tỉ lệ quy đổi → previews who owes/is owed → confirms → system snapshots, saves history, resets all wallets to 0 → admin manually collects/pays each person and marks them done

---

## Integration Points

- `router.go`: Add WC routes after existing v1 routes; admin middleware reuse
- `main.go` / `setup.go`: Initialize `WcRepository`, `WcFootballClient`, `WcService`, `WcHandler`
- `nav` sidebar in `MainLayout.vue`: Add "World Cup" nav item with trophy/football icon
- Router: Add `/world-cup` and `/world-cup/bet` in Vue router

---

## Error Handling

| Error | HTTP | Message |
|---|---|---|
| Match already locked | 422 | "Trận đấu đã bắt đầu, không thể tham gia dự đoán" |
| Duplicate bet | 409 | "Bạn đã dự đoán loại này cho trận này rồi" |
| Match not settled yet | 422 | "Trận đấu chưa có kết quả" |
| Football API error | 503 | "Không thể đồng bộ dữ liệu, thử lại sau" |
| Settlement with no registered users | 422 | "Chưa có người dùng nào để tất toán" |

---

## Performance Considerations

- `ListMatches` query: single JOIN with no N+1; add `match_date` index (already in migration)
- Bet placement: SELECT FOR UPDATE on wallet row to prevent race; single transaction
- Leaderboard: `ORDER BY balance DESC` on `wc_wallets JOIN users` — no pagination needed for a small team
- Schedule page: fetch all matches once on mount, filter client-side (no backend re-fetch on filter change)

---

## Security Notes

- Admin endpoints (`/admin/*`) must check `isAdmin` middleware — same pattern as existing config/settlement admin routes
- Bet placement validates `userID` from JWT — users cannot bet on behalf of others
- `stake` validation: must be > 0 and ≤ current wallet balance
- All bets on a match are always visible to everyone (transparent fun; no money involved — points only)
