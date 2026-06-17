---
phase: testing
title: StatsAPI Odds Import — Testing Strategy
description: Test scope for thestatsapi.com integration
---

# Testing Strategy

## Scope

- **Unit:** StatsApiSyncService — response parsing, field mapping, overwrite policy logic
- **Integration:** Import endpoints (mock HTTP client) + DB upsert behavior
- **Manual:** Preview dialog UX, confirm flow, admin panel integration

## Test Files

| File | Layer | Coverage Target |
|---|---|---|
| `internal/service/statsapi_sync_service_test.go` | Service | Response parsing, overwrite logic |
| `internal/api/wc_sync_handler_test.go` | Handler | Preview vs confirm, auth |

## Unit Tests

- `FetchOddsForMatch`: parse valid response → correct struct
- `FetchOddsForMatch`: API 404 → return error, not panic
- `FetchOddsForMatch`: timeout → return error within 11s
- `ImportHandicapForMatch(overwrite=false)`: skip if `handicap_value != null`
- `ImportHandicapForMatch(overwrite=true)`: always write even if field not null
- `SyncUpcomingMatches`: only processes matches with null handicap AND match_date ≤ NOW()+24h

## Integration Tests

- POST `import-odds` preview_only=true → returns preview without writing to DB
- POST `import-odds` preview_only=false → writes handicap fields + sets `odds_synced_at`
- POST `import-odds` twice → idempotent, no duplicate logs
- POST `import-score-odds` → upserts, no duplicates for same scoreline
- GET `sync-logs` → returns logs sorted by created_at desc

## Test Data & Environments

- Mock thestatsapi.com HTTP responses using `httptest.NewServer` hoặc interface injection.
- Seed một `wc_match` với `external_id` matching mock fixture ID.

## Execution

```bash
cd backend
go test ./internal/service/... -run TestStatsApi
go test ./internal/api/... -run TestWcSync
```

## Risks & Gaps

- Không có real integration test với thestatsapi.com (dùng mock) — verify manually với live API key.
- Exact score odds format phụ thuộc API response → cần update test sau T1.1.
