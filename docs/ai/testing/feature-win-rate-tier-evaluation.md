---
phase: testing
title: Win Rate & Tier Evaluation — Testing Strategy
description: Test scope and validation criteria for win rate calculation and tier auto-evaluation
---

# Testing Strategy

## Scope

- **Unit tests:** `EvaluateTier` function — all threshold boundary cases
- **Unit tests:** `TierService.RecalculateForUsers` — mock repository interactions
- **Integration tests:** Win rate SQL query (repository layer) against real DB
- **Manual/visual:** Win rate column and tier badge display in UsersView and DashboardView

## Test Files

| File | Package/Layer | Coverage Target | Status |
|------|---------------|----------------|--------|
| `backend/internal/service/tier_service_test.go` | service | 100% of `EvaluateTier` + `RecalculateForUsers` + `RecalculateAllTiers` | ✅ 22 tests passing |
| `backend/internal/repository/user_repository_test.go` | repository | win rate SQL query correctness | ✅ 6 tests (skip when no `TEST_DATABASE_URL`) |

## Unit Tests

### `EvaluateTier` — all boundary cases

| Input (winRate, totalMatches) | Expected tier |
|-------------------------------|---------------|
| 0.65, 20 | `"pro"` |
| 0.60, 10 | `"pro"` (exactly at threshold) |
| 0.59, 10 | `"normal"` |
| 0.40, 10 | `"normal"` (exactly at threshold) |
| 0.39, 10 | `"noob"` |
| 0.00, 10 | `"noob"` |
| 1.00, 10 | `"pro"` |
| 0.80, 9  | `"normal"` (< 10 matches → default) |
| 0.80, 0  | `"normal"` (0 matches → default) |
| 0.00, 0  | `"normal"` |

### `TierService.RecalculateForUsers`

- Calls `UpdateTier` for each user ID with the correct tier string
- Does not call `UpdateTier` for empty ID list
- Returns error if `UpdateTier` fails, but processes remaining users

## Integration Tests

### Win rate SQL query

Setup: Insert 2 users, 10 match_participants records (6 wins / 4 losses for user A)

| Assertion | Expected |
|-----------|----------|
| `user_a.win_rate` | 0.60 |
| `user_a.total_matches` | 10 |
| `user_a.won_matches` | 6 |
| `user_b.win_rate` | 0.0 (no matches) |
| `user_b.total_matches` | 0 |

### Tier recalculation on match create

- Record a match; assert both participants have their `tier` column updated in DB
- Delete the match; assert tiers are re-evaluated

### Tier recalculation does not block match creation

- If `UpdateTier` fails, match is still returned successfully

## Test Data & Environments

- Repository integration tests require `TEST_DATABASE_URL` (Postgres DSN). Example:
  ```
  TEST_DATABASE_URL="host=localhost port=5432 user=postgres password=secret dbname=esport_test sslmode=disable"
  ```
- Each test seeds its own rows and registers `t.Cleanup` to remove them — no shared state.
- Unit tests have no DB dependency and run unconditionally.

## Execution

```bash
cd backend

# Unit tests (no DB required)
go test ./internal/service/... -v

# Integration tests (requires TEST_DATABASE_URL)
TEST_DATABASE_URL="host=localhost port=5432 user=postgres password=secret dbname=esport_test sslmode=disable" \
  go test ./internal/repository/... -v

# All backend tests
go test ./...
```

## Coverage & Quality Gates

- `EvaluateTier`: 100% line coverage (pure function, all branches testable)
- `TierService.RecalculateForUsers`: all happy path + error path covered
- Win rate query: result correctness validated by integration test
- Manual check: UsersView and DashboardView leaderboard both show win rate column and tier badges

## Risks & Gaps

- No automated frontend component tests for `WinRateBadge` — rely on manual visual check
- Tier recalculation on match delete not covered if delete is not yet integrated — add test when implemented
- `< 10 matches` display shows "—" win rate — verify this renders correctly and does not break table layout
