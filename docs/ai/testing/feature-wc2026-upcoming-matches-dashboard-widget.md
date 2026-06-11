---
phase: testing
title: WC2026 Upcoming Matches Dashboard Widget — Testing Strategy
description: Test scope, cases, and validation criteria
---

# Testing Strategy

## Scope

- **Backend unit:** `MatchFilter` date range applied correctly in `ListMatches` query.
- **Backend integration:** `GET /api/v1/wc/matches?status=scheduled&date_from=...&date_to=...` returns correct subset.
- **Frontend component:** `WcUpcomingWidget.vue` renders correctly for given props.
- **Frontend integration:** `DashboardView.vue` fetches and renders widget; hides when empty.

## Test Files

| File | Layer | Status | Coverage |
|---|---|---|---|
| `backend/internal/repository/wc_repository_test.go` | Repository | ✅ Written | 8 test cases — DateFrom/DateTo branches |
| Manual / curl | API | Deferred (server must run) | End-to-end filter + response shape |
| Frontend component | Component | Deferred (no Vitest setup) | Visual rendering |

## Unit Tests

### Backend — `ListMatches` with date range filter

All 8 tests are in `backend/internal/repository/wc_repository_test.go`. They skip if `TEST_DATABASE_URL` is not set.

| Test name | Scenario |
|---|---|
| `TestListMatches_DateRange_InWindowReturned` | Match inside window returned; before/after excluded |
| `TestListMatches_DateRange_BoundariesInclusive` | Match at exact `date_from` and `date_to` — both included |
| `TestListMatches_DateRange_EmptyFilter_ReturnsAll` | No date filter → all matches returned (backward compat) |
| `TestListMatches_DateRange_OnlyDateFrom_LowerBoundOnly` | Only `date_from` set → lower bound only |
| `TestListMatches_DateRange_OnlyDateTo_UpperBoundOnly` | Only `date_to` set → upper bound only |
| `TestListMatches_DateRange_CombinedWithStatusFilter` | `status=scheduled` + date range → AND behaviour |
| `TestListMatches_DateRange_LiveMatchLookback` | Live match 2h ago within 4h lookback → included; 5h ago → excluded |
| `TestListMatches_DateRange_SortedByMatchDateASC` | Results are sorted `match_date ASC` |

## Integration Tests

**Deferred — manual curl** (run once backend is started):

```bash
# 1. Matches in 48h window
curl "http://localhost:8080/api/v1/wc/matches?date_from=$(date -u +%Y-%m-%dT%H:%M:%SZ)&date_to=$(date -u -v+48H +%Y-%m-%dT%H:%M:%SZ)"

# 2. With status filter — only scheduled
curl "http://localhost:8080/api/v1/wc/matches?status=scheduled&date_from=...&date_to=..."

# 3. Empty window — expect []
curl "http://localhost:8080/api/v1/wc/matches?date_from=2030-01-01T00:00:00Z&date_to=2030-01-02T00:00:00Z"
```

## Test Data & Environments

- Seeds: minimal — insert 2-3 `wc_matches` rows with controlled `match_date` and `status` values.
- No auth token required for `GET /matches` (public endpoint).
- Local dev against PostgreSQL; same as existing WC tests.

## Execution

```bash
# Backend tests — requires TEST_DATABASE_URL (skips gracefully without it)
cd backend
go test ./internal/repository/... -run TestListMatches -v

# With real DB
TEST_DATABASE_URL="host=localhost port=5432 user=postgres password=secret dbname=esport_test sslmode=disable" \
  go test ./internal/repository/... -run TestListMatches -v
```

## Coverage & Quality Gates

| Branch | Test | Status |
|---|---|---|
| `DateFrom` and `DateTo` both set | `TestListMatches_DateRange_InWindowReturned` | ✅ |
| Boundary inclusive (at exact `date_from`) | `TestListMatches_DateRange_BoundariesInclusive` | ✅ |
| Boundary inclusive (at exact `date_to`) | `TestListMatches_DateRange_BoundariesInclusive` | ✅ |
| Both fields empty → no filter | `TestListMatches_DateRange_EmptyFilter_ReturnsAll` | ✅ |
| Only `DateFrom` set | `TestListMatches_DateRange_OnlyDateFrom_LowerBoundOnly` | ✅ |
| Only `DateTo` set | `TestListMatches_DateRange_OnlyDateTo_UpperBoundOnly` | ✅ |
| Combined with `Status` filter (AND) | `TestListMatches_DateRange_CombinedWithStatusFilter` | ✅ |
| Live match within 4h lookback | `TestListMatches_DateRange_LiveMatchLookback` | ✅ |
| Results sorted `match_date ASC` | `TestListMatches_DateRange_SortedByMatchDateASC` | ✅ |
| Dashboard widget hidden when empty | Manual smoke test | Deferred |
| Dashboard widget visible (≤5 matches) | Manual smoke test | Deferred |
| Dashboard widget shows "Xem thêm →" (>5 matches) | Manual smoke test | Deferred |

## Risks & Gaps

- No automated E2E/component tests for Vue — widget rendering is manual verification only.
- WC feature flag off (503) scenario: `wcPublicApi` has no interceptor so the silent-catch absorbs it; manual test only.
- The existing `date=YYYY-MM-DD` single-day filter is not regressed — its code path is unchanged and covered by the `EmptyFilter` test (which exercises the full `ListMatches` function without date_from/date_to).
