---
phase: implementation
title: StatsAPI Odds Import — Implementation Guide
description: Technical implementation notes and patterns
---

# Implementation Guide

## Development Setup

1. Lấy API key và base URL từ thestatsapi.com dashboard
2. Set env vars: `STATSAPI_KEY=<key>` và `STATSAPI_BASE_URL=https://api.thestatsapi.com/v1`
3. Chạy migration để thêm `odds_synced_at` vào `wc_matches` và tạo `wc_sync_logs`

## Code Structure

```
backend/internal/
  service/
    statsapi_sync_service.go   ← NEW: HTTP client + sync logic
  api/
    wc_sync_handler.go         ← NEW: import-odds, import-score-odds, sync-logs endpoints
  model/
    wc_sync_log.go             ← NEW: WcSyncLog struct
  repository/
    wc_repository.go           ← ADD: UpsertHandicap, UpsertScoreOdds, CreateSyncLog, ListSyncLogs
  migrations/
    XXXX_add_odds_synced_at.sql
    XXXX_create_wc_sync_logs.sql

frontend/src/
  services/
    wcService.ts               ← ADD: importHandicapOdds, importScoreOdds, getSyncLogs
  types/
    wc.ts                      ← ADD: ImportOddsPreview, ImportScoreOddsPreview, WcSyncLog
  components/wc/
    WcImportOddsDialog.vue     ← NEW
    WcImportScoreOddsDialog.vue ← NEW
    WcAdminPanel.vue           ← EDIT: add buttons + synced_at chip
```

## Implementation Notes

### StatsApiSyncService — HTTP Client

```go
func (s *StatsApiSyncService) FetchOddsForMatch(externalID string) (*StatsApiFixtureOdds, error) {
    url := fmt.Sprintf("%s/odds?fixture=%s", s.baseURL, externalID)
    req, _ := http.NewRequest("GET", url, nil)
    req.Header.Set("X-Auth-Token", s.apiKey)  // header name TBD from API docs
    req.Header.Set("Accept", "application/json")

    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()
    req = req.WithContext(ctx)
    // ... unmarshal response
}
```

### Upsert Pattern cho WcScoreOdds

```sql
INSERT INTO wc_score_odds (match_id, home_score, away_score, odds)
VALUES ($1, $2, $3, $4)
ON CONFLICT (match_id, home_score, away_score)
DO UPDATE SET odds = EXCLUDED.odds, updated_at = NOW();
```

Unique index đã có: `idx_score_odds_scoreline (match_id, home_score, away_score)` — safe to upsert.

### Cron Logic (fill blank only)

```go
func (s *StatsApiSyncService) SyncUpcomingMatches() (updated, failed int, err error) {
    matches, _ := s.repo.ListMatchesForCron() // status=scheduled, match_date <= NOW()+24h, handicap_value IS NULL
    for _, m := range matches {
        time.Sleep(1 * time.Second) // rate limit buffer
        if err := s.ImportHandicapForMatch(m.ID, false); err != nil {
            failed++
        } else {
            updated++
        }
    }
    return
}
```

### Frontend Preview Dialog Pattern

```vue
<!-- WcImportOddsDialog.vue — simplified flow -->
async function fetchPreview() {
  preview.value = await wcService.importHandicapOdds(props.matchId, true)
}
async function confirm() {
  await wcService.importHandicapOdds(props.matchId, false)
  emit('imported')
}
```

## Integration Points

- **Routes:** Thêm vào `cmd/server/main.go` dưới admin middleware group:
  ```go
  adminGroup.POST("/matches/:id/import-odds", syncHandler.ImportOdds)
  adminGroup.POST("/matches/:id/import-score-odds", syncHandler.ImportScoreOdds)
  adminGroup.GET("/sync-logs", syncHandler.GetSyncLogs)
  ```
- **Cron start:** Gọi `syncSvc.StartCron(30 * time.Minute)` trong `main()` sau khi DB connected.

## Error Handling

- API 404/no data → trả về `{ "error": "no odds available for this match" }` với HTTP 404.
- API timeout → trả về `{ "error": "stats api timeout" }` với HTTP 502.
- Cron errors → log + continue, không panic.

## Security Notes

- `STATSAPI_KEY` không được log, không được expose trong bất kỳ response nào.
- Import endpoints đều dùng `WcAdminMiddleware` — chỉ admin mới gọi được.
