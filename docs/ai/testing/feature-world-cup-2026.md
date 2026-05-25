---
phase: testing
title: World Cup 2026 — Testing Strategy
description: Test scope, critical paths, and settlement edge cases
---

# Testing Strategy

## Scope

| Layer | What to test |
|---|---|
| Unit | Settlement helpers (`evaluateHandicapBet`, `evaluateExactScoreBet`), bet validation |
| Integration | PlaceBet transaction, SettleMatch transaction, idempotency |
| E2E / manual | Sync → set handicap/exact-odds → place both bet types → settle → verify wallets |

---

## Test Files

| File | Package | Coverage Target |
|---|---|---|
| `backend/internal/service/wc_service_test.go` | `service` | ≥ 90% of settlement logic |
| `backend/internal/repository/wc_repository_test.go` | `repository` | ≥ 80% (real DB, integration) |

---

## Unit Tests — Settlement Logic

### `evaluateHandicapBet`

| Scenario | Input | Expected |
|---|---|---|
| Home wins with handicap (half-ball) | BRA 2–0 ARG, home gives -1.5, bet on home, odds 1.90 | win, payout = floor(100 × 1.90) = 190 |
| Home loses with handicap | BRA 1–0 ARG, home gives -1.5, bet on home | lose, payout = 0 |
| Away wins with handicap | BRA 0–0 ARG, away gives -0.5, bet on away | win (away effectively +0.5) |
| Push (whole-number handicap, exact) | BRA 2–1 ARG, home gives -1.0, bet on home | push, payout = stake |
| Home gives 0 (level) | BRA 1–0 ARG, handicap = 0, bet on home | win |
| Away bet wins | BRA 0–2 ARG, home gives -1.5, bet on away | win |

### `evaluateExactScoreBet`

| Scenario | Input | Expected |
|---|---|---|
| Exact match | actual 2–1, predicted 2–1, odds_snapshot 5.00, stake 100 | win, payout = 500 |
| Home score wrong by 1 | actual 2–1, predicted 3–1 | lose, payout = 0 |
| Away score wrong by 1 | actual 2–1, predicted 2–0 | lose, payout = 0 |
| Both wrong | actual 2–1, predicted 0–0 | lose, payout = 0 |
| 0–0 draw correct | actual 0–0, predicted 0–0, odds_snapshot 8.00, stake 50 | win, payout = 400 |
| High-odds scoreline | actual 3–2, predicted 3–2, odds 12.00, stake 100 | win, payout = 1200 |

### `PlaceBet` — exact score validation

| Scenario | Expected |
|---|---|
| Scoreline exists in `wc_score_odds` | 201, wallet debited, bet inserted |
| Scoreline NOT in `wc_score_odds` | 422 "Tỉ số này không có trong danh sách cược" |
| `wc_score_odds` list is empty for match | 422, no exact score bets accepted |
| Same scoreline bet again same match | 409, wallet unchanged |
| Different scoreline same match | 201, second bet inserted, wallet debited again |
| Handicap + exact score on same match | Both accepted (201 each), wallet debited twice |

### `PlaceBet` validation

| Scenario | Expected |
|---|---|
| Stake > wallet balance | 422, wallet unchanged |
| Match already locked (past `bets_locked_at`) | 422, no bet created |
| Duplicate handicap same side same match | 409, wallet unchanged |
| Duplicate exact scoreline same match | 409, wallet unchanged |
| Handicap + exact score on same match | Both allowed, no duplicate error |
| Valid bet | 201, wallet debited, bet inserted |
| Stake = 0 | 422 validation error |

---

## Integration Tests — Repository

- `UpsertMatches`: insert 3 matches, upsert same external IDs with updated scores → only 3 rows, scores updated
- `CreateScoreOdds` + `GetScoreOdds`: insert 1–0 @ 5.00; query (1,0) → returns 5.00; query (2,0) → not found
- `UpdateScoreOdds`: change 1–0 from 5.00 → 6.50; re-query → 6.50
- `DeleteScoreOdds`: delete 1–0; query (1,0) → not found
- `GetOrCreateWallet`: call twice for same user → same record, balance unchanged
- `UpdateWalletBalance`: delta = +50, then -30 → final balance = initial + 20
- `CreateBet` + `ListBets`: insert bet, query by user → bet appears with correct match join
- `ListBetsForMatch`: user places handicap + 3 exact score bets on match A, 1 bet on match B → `ListBetsForMatch(A)` returns 4
- Multiple exact score bets same match: insert 1:0 and 2:1 for same user/match → both inserted, distinct rows; attempt 1:0 again → unique constraint error

---

## Integration Tests — Service

### PlaceBet transaction

```
Setup: wallet balance = 200
Action: PlaceBet(stake=150)
Assert: wallet.balance = 50, bet.result = nil
```

### PlaceBet — multiple exact score bets same match

```
Setup: wallet balance = 300
Action: PlaceBet(exact_score 1:0, stake=100), PlaceBet(exact_score 2:1, stake=100)
Assert: wallet.balance = 100, two bet rows for same match/user, both result = nil
Action: PlaceBet(exact_score 1:0, stake=50)  -- duplicate scoreline
Assert: 409, wallet.balance unchanged = 100
```

### SettleMatch — handicap win

```
Setup: match 2–1 (home wins); bet: handicap home -0.5, stake 100, odds 1.90
Action: SettleMatch(matchID)
Assert: bet.result = "win", bet.payout = 190, wallet.balance = initial - 100 + 190
```

### SettleMatch — exact score win

```
Setup: match 2–1; exact score bet predicted 2–1, odds 5.00, stake 100
Action: SettleMatch(matchID)
Assert: bet.result = "win", bet.payout = 500, wallet.balance = initial - 100 + 500
```

### SettleMatch — exact score lose

```
Setup: match 2–1; exact score bet predicted 1–0, stake 100
Action: SettleMatch(matchID)
Assert: bet.result = "lose", bet.payout = 0, wallet.balance = initial - 100
```

### SettleMatch — multiple exact score bets, one wins

```
Setup: match 2–1; user has three bets:
  - exact score 2:1, odds 6.00, stake 100  → win
  - exact score 1:0, odds 5.00, stake 50   → lose
  - handicap home -0.5, odds 1.90, stake 100 → win
Action: SettleMatch(matchID)
Assert:
  - bet1 result = "win",  payout = 600
  - bet2 result = "lose", payout = 0
  - bet3 result = "win",  payout = 190
  - wallet.balance = initial - 250 + 600 + 190 = initial + 540
```

### SettleMatch — idempotency

```
Setup: match settled once (result = 2–1)
Action: admin corrects score to 1–2, SettleMatch again
Assert: loser bet correctly becomes "lose"; winner bet correctly becomes "lose"; wallets corrected
Assert: running SettleMatch a third time with same score → wallets unchanged
```

### SettleMatch — push

```
Setup: match 1–0; handicap home -1.0; stake 100
Action: SettleMatch → adjusted home = 1 - 1 = 0, away = 0 → push
Assert: bet.result = "push", bet.payout = 100, wallet.balance = initial (unchanged net)
```

---

## Test Data & Environments

- Requires real DB (same pattern as existing repo integration tests: `DB_*` env vars)
- Match fixtures: create in test setup with `wc_matches` insert; no external API calls in tests
- Wallet: create via `wc_wallets` insert with known balance

---

## Execution

```bash
# Unit + integration tests for WC service
go test ./internal/service/... -run TestWc -v

# Repository integration tests (requires DB)
go test ./internal/repository/... -run TestWc -v

# Coverage
go test ./internal/service/... -cover -coverprofile=wc.out
go tool cover -func=wc.out | grep wc_service
```

---

## Coverage & Quality Gates

- `evaluateHandicapBet`: 100% — all branches including push
- `evaluateExactScoreBet`: 100% — win and lose paths
- `PlaceBet` exact score path: scoreline-not-found rejection must be covered
- `SettleMatch` idempotency path: must be covered
- PlaceBet validation paths (locked, insufficient, duplicate): all must be covered

---

## Risks & Gaps

- **football-data.org field names may change** for WC2026 before the tournament — test the sync function with a recorded API fixture (saved JSON file) rather than live calls
- **Time zone handling**: `match_date` from API is UTC; `bets_locked_at` should also be stored UTC; frontend displays in local time
- **Re-settle race condition**: if two admins trigger settle simultaneously — mitigated by DB transaction but worth noting
