---
phase: testing
title: WC Betting Config Improvements — Testing Strategy
description: Test scope for configurable bet limits, handicap display, and consistent bet-type labels
---

# Testing Strategy

## Scope

- Unit: `wc_config` service methods for reading/writing `min_points`/`max_points`; bet validation logic.
- Integration: Admin config update endpoint + prediction placement with dynamic limits.
- Manual/visual: Handicap display in pending/history tabs; label consistency across all WC views.

## Test Files

| File | Package/Layer | Coverage Target |
|------|---------------|----------------|
| `backend/internal/service/wc_config_service_test.go` | Go service | Config CRUD, default values |
| `backend/internal/api/wc_handler_test.go` | Go API | Bet validation with dynamic limits |

## Unit Tests

### Config Service
- `GetBetLimits()` returns `(min=1, max=5)` when row has no override (defaults).
- `UpdateBetLimits(min, max)` persists and `GetBetLimits()` reflects new values.
- `UpdateBetLimits` rejects `min <= 0`, `max < min`.

### Bet Validation
- Stake equal to `min_points` → accepted.
- Stake equal to `max_points` → accepted.
- Stake below `min_points` → 400 error with message.
- Stake above `max_points` → 400 error with message.
- Config row missing → falls back to defaults (1, 5).

## Integration Tests

- `PUT /admin/wc/config` with `{"min_points":2,"max_points":8}` → 200, DB updated.
- `POST /wc/matches/:id/predict` with stake=1 after setting min=2 → 400.
- `POST /wc/matches/:id/predict` with stake=8 after setting max=8 → 201.
- `GET /wc/config` returns updated `min_points` / `max_points`.

## Manual / Visual Checklist

### Configurable Limits
- [ ] Admin panel shows current min/max values on load.
- [ ] Saving new values: success toast, form re-fetches values.
- [ ] Prediction form `:min` and `:max` update after config change (reload or reactive).
- [ ] Champion prediction form also respects new limits.

### Handicap Display in Predict Tabs
- [ ] "Đang chờ kết quả" tab shows handicap line for handicap predictions.
- [ ] "Lịch sử" tab shows handicap line for settled handicap predictions.
- [ ] Exact score and O/U predictions do NOT show a spurious handicap line.
- [ ] Handicap line is blank/hidden when `home_handicap` is null (match not yet configured).

### Label Consistency
- [ ] Prediction form tabs: "Kèo Handicap" / "Kèo tỉ số" / "Kèo tài xỉu".
- [ ] Pending predictions list: same labels per row.
- [ ] History list: same labels per row.
- [ ] Match detail prediction summary: same labels.
- [ ] No raw `"handicap"` / `"exact_score"` / `"over_under"` visible in UI.

## Test Data & Environments

- Seed: a `wc_config` row with id=1 present; two matches with `home_handicap` set and one without.
- Local dev DB.

## Execution

```bash
cd backend && go test ./internal/service/... ./internal/api/... -run Wc
```

## Coverage & Quality Gates

- All config service unit tests green.
- API validation tests green.
- Manual checklist fully ticked before PR merge.

## Risks & Gaps

- Frontend reactivity: if config is fetched once at app start, forms may not reflect admin changes without a page reload. Acceptable for V1 — note in implementation.
