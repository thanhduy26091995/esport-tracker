---
phase: testing
title: WC Bet Cancel Penalty — Testing
description: Unit tests for penalty math, config validation, history sort, and bet-lock logic
---

# Testing Plan — WC Bet Cancel Penalty

## Test Files

| File | Tests | Type |
|------|-------|------|
| `backend/internal/service/wc_penalty_test.go` | `TestComputeCancelPenalty`, `TestComputeReducePenalty` | Unit |
| `backend/internal/service/wc_service_test.go` | `TestIsBetLocked`, `TestValidatePenaltyConfig`, `TestSortBetHistoryNewestFirst` | Unit |

All tests run without a database. DB-dependent service integration tests are in `wc_integration_test.go` and require `WC_TEST_DSN` to be set.

## Run Commands

```bash
# Unit tests only (no DB required)
go test ./internal/service/... -run "TestComputeCancelPenalty|TestComputeReducePenalty|TestIsBetLocked|TestValidatePenaltyConfig|TestSortBetHistory"

# Full service suite (DB tests auto-skip without WC_TEST_DSN)
go test ./internal/service/... -count=1 -timeout 120s
```

## Test Coverage

### `computeCancelPenalty` — 11 cases

| Scenario | Expected |
|---|---|
| Feature disabled | 0 regardless of stake |
| Zero penalty percent | 0 |
| Zero stake | 0 |
| 20% of 100 | 20 |
| 20% of 10 | 2 |
| Floor: 20% of 7 = 1.4 | 1 |
| Floor: 20% of 4 = 0.8 | 0 |
| Floor: 50% of 1 = 0.5 | 0 |
| 100% of 100 | 100 |
| 10% of 100 | 10 |
| 5% of 100 | 5 |

### `computeReducePenalty` — 11 cases

Key invariant: **`maxPercent=0` means "no reduction limit"** (fixed bug — see below).

| Scenario | penalty | excess | allowedMin |
|---|---|---|---|
| maxPercent=0 (no limit) | 0 | 0 | 0 |
| Increase stake | 0 | 0 | 0 |
| Same stake | 0 | 0 | 0 |
| Reduce within 50% max | 0 | 0 | 50 |
| Reduce exactly to allowedMin | 0 | 0 | 50 |
| Reduce below min (excess 10) | 2 | 10 | 50 |
| Excess 1 (floor rounds to 0) | 0 | 1 | 50 |
| 30% max, 10% penalty | 1 | 10 | 70 |
| Small stake at allowedMin | 0 | 0 | 5 |
| Small stake, 1-pt excess | 0 | 1 | 5 |
| Deep reduction, 50% penalty | 20 | 40 | 50 |

### `isBetLocked` — 6 cases

- Terminal statuses (live / completed / cancelled) → locked
- Scheduled with no lock time → not locked
- Scheduled with future lock time → not locked
- Scheduled with past lock time → locked

### `validatePenaltyConfig` — 9 cases

- Valid: all zeros, all 100s, typical (20/50/20)
- Error on negative or >100 for each of the three percent fields

### `sortBetHistoryNewestFirst` — 5 cases

- Unsorted items → newest first after sort
- Already sorted → unchanged
- Single item → no change
- Empty slice → no panic
- Mixed regular/custom kinds → sorted by CreatedAt only

## Bug Fixed

**`computeReducePenalty` — `maxPercent=0` interpretation**

The design doc specifies `bet_reduce_max_percent=0` means "no reduction limit". The original
inline code computed `allowedMinStake = originalStake - 0 = originalStake`, so any reduction
triggered a penalty — the opposite of the intended behavior.

Fixed in `wc_penalty.go` with an explicit guard:

```go
if maxPercent == 0 || newStake >= originalStake {
    return 0, 0, 0
}
```

Both `UpdateBetStake` and `PreviewReduceStake` use this function and behave correctly.

## Deferred

- **End-to-end cancel flow** (soft-cancel + wallet atomicity): requires DB — use `wc_integration_test.go` pattern with `WC_TEST_DSN`.
- **Frontend penalty dialog**: manual browser testing; no component unit test infrastructure.
- **Admin panel save**: requires E2E.
