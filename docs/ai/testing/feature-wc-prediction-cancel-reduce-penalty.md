---
phase: testing
title: WC Prediction Cancel & Reduce Penalty — Testing
description: Test cases for cancel/reduce penalty feature
---

# Testing Strategy

## Test Coverage Goals

- Unit test coverage: 100% of penalty calculation functions
- Integration tests: all critical paths via real PostgreSQL (TEST_DATABASE_URL)
- Integration tests skip gracefully if TEST_DATABASE_URL not set

---

## Unit Tests

### `computeCancelPenalty` — `backend/internal/service/wc_penalty_test.go`

- [x] Feature disabled → 0
- [x] Zero percent → 0
- [x] Zero stake → 0
- [x] 20% of 100 → 20.0
- [x] 20% of 7 → 1.4 (float, no floor)
- [x] 20% of 4 → 0.8
- [x] 50% of 1 → 0.5
- [x] 100% of 100 → 100.0

### `computeReducePenalty` — `backend/internal/service/wc_penalty_test.go`

- [x] maxPercent=0 (no limit) → 0 penalty
- [x] Increase stake → 0
- [x] Same stake → 0
- [x] Reduce within 50% free threshold → 0
- [x] Reduce exactly to allowedMin → 0
- [x] Reduce below allowedMin → excess 10, penalty 2.0
- [x] Excess 1pt → 0.2 (no floor)
- [x] 30% max, 10% penalty → excess 10, penalty 1.0
- [x] Small stake: exactly at allowedMin → 0
- [x] Small stake: 1pt excess → 0.2
- [x] Deep reduction: 50% penalty percent → penalty 20.0

---

## Integration Tests

**File:** `backend/internal/service/wc_prediction_penalty_test.go`

Requires: `TEST_DATABASE_URL` pointing to a PostgreSQL test database.

### SubmitPrediction

- [x] `TestPrediction_Submit_SetsOriginalPoints` — `original_points` = `points` on create

### DeletePrediction (Cancel Penalty)

- [x] `TestPrediction_Delete_WithPenalty_DeductsWallet` — penalty=2.0 deducted, `cancelled_at` set, `cancel_penalty` persisted
- [x] `TestPrediction_Delete_PenaltyDisabled_NoDeduction` — wallet unchanged when feature off
- [x] `TestPrediction_Delete_FloatPenalty_NotRounded` — 20% of 7 = 1.4, not 1

### Partial Unique Index

- [x] `TestPrediction_Cancel_ThenRepredict_Allowed` — cancel + re-predict same pick succeeds (partial index)

### ListPredictions

- [x] `TestPrediction_List_IncludesCancelled` — cancelled predictions appear in list with `cancelled_at` set

### UpdatePredictionPoints (Reduce Penalty)

- [x] `TestPrediction_Reduce_WithinFreeThreshold_NoPenalty` — reduce 100→60 (allowedMin=50) → 0 penalty
- [x] `TestPrediction_Reduce_ExceedsThreshold_PenaltyApplied` — reduce 100→40 → penalty=2.0, wallet deducted, `reduce_penalty` accumulated
- [x] `TestPrediction_Reduce_NullOriginalPoints_NoPenalty` — old data (NULL original_points) → graceful no-penalty on first reduce

---

## Edge Cases Verified

| Case | Covered by |
|------|-----------|
| Float penalty (no floor) | `TestComputeCancelPenalty` + `TestPrediction_Delete_FloatPenalty_NotRounded` |
| Feature disabled → no penalty | `TestPrediction_Delete_PenaltyDisabled_NoDeduction` |
| Old data NULL original_points → grace | `TestPrediction_Reduce_NullOriginalPoints_NoPenalty` |
| Cancel → re-predict same pick | `TestPrediction_Cancel_ThenRepredict_Allowed` |
| Cancelled predictions in list | `TestPrediction_List_IncludesCancelled` |
| Settlement excludes cancelled | Code review: `ListPredictionsForMatch` filter verified |
| Analytics excludes cancelled | Code review: 10 queries + `GetMyBetStats` filter verified |

---

## Deferred Tests (follow-up)

- [ ] `penalty cap at wallet balance` — wallet < penalty → deduct only up to balance
- [ ] `GetMyBetStats` excluded cancelled predictions — requires analytics DB fixtures
