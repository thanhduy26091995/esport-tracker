---
phase: planning
title: StatsAPI Odds Import — Task Breakdown
description: Implementation order, effort estimates, and dependencies
---

# Project Planning & Task Breakdown

## Milestones

- [ ] **M1:** Backend sync service + manual import endpoints (no cron yet)
- [ ] **M2:** Frontend import dialogs + admin panel integration
- [ ] **M3:** Background cron + sync log table + admin log view

---

## Task Breakdown

### Phase 1: Backend Foundation

- [ ] **T1.1** — Verify thestatsapi.com API docs (endpoints, response format for Asian handicap + exact score odds, rate limits)
  - *Effort: 1h* | *Blocker: must complete before T1.2*
- [ ] **T1.2** — Add `odds_synced_at` column migration to `wc_matches`
  - *Effort: 30min*
- [ ] **T1.3** — Create `wc_sync_logs` table migration
  - *Effort: 30min*
- [ ] **T1.4** — Implement `StatsApiSyncService` with `FetchOddsForMatch`, `ImportHandicapForMatch`, `ImportScoreOddsForMatch`
  - *Effort: 3-4h* | Depends on T1.1 for API response mapping
- [ ] **T1.5** — Implement `WcSyncHandler` with `ImportOdds` and `ImportScoreOdds` endpoints (preview_only flag)
  - *Effort: 1.5h* | Depends on T1.4
- [ ] **T1.6** — Wire up new routes in `cmd/server/main.go` (admin-only middleware)
  - *Effort: 30min*
- [ ] **T1.7** — Add `STATSAPI_KEY` and `STATSAPI_BASE_URL` to env config + docker-compose
  - *Effort: 30min*

### Phase 2: Frontend Integration

- [ ] **T2.1** — Add `importHandicapOdds`, `importScoreOdds`, `getSyncLogs` to `wcService.ts`
  - *Effort: 1h*
- [ ] **T2.2** — Add TypeScript types for `ImportOddsPreview`, `ImportScoreOddsPreview`, `WcSyncLog`
  - *Effort: 30min*
- [ ] **T2.3** — Create `WcImportOddsDialog.vue` (preview → confirm flow)
  - *Effort: 2h*
- [ ] **T2.4** — Create `WcImportScoreOddsDialog.vue` (preview table → confirm)
  - *Effort: 1.5h*
- [ ] **T2.5** — Add "Import kèo" + "Import tỉ số" buttons + `odds_synced_at` chip vào match card trong `WcAdminPanel.vue`
  - *Effort: 1h*

### Phase 3: Background Cron + Logging

- [ ] **T3.1** — Implement `SyncUpcomingMatches` trong `StatsApiSyncService` (query matches scheduled trong 24h, fill blank only)
  - *Effort: 2h*
- [ ] **T3.2** — Implement `StartCron` goroutine với `time.Ticker` + configurable interval
  - *Effort: 1h*
- [ ] **T3.3** — Implement `GetSyncLogs` endpoint + repository query
  - *Effort: 1h*
- [ ] **T3.4** — Thêm sync log view vào admin panel (accordion hoặc tab "Lịch sử sync")
  - *Effort: 1h*
- [ ] **T3.5** — E2E test: trigger cron manually, verify log + DB update
  - *Effort: 1h*

---

## Dependencies

```
T1.1 (verify API) → T1.4 (implement client)
T1.4 → T1.5 (handler)
T1.5 → T2.1 (frontend service)
T2.1 → T2.3, T2.4, T2.5
T1.4 → T3.1 → T3.2
T3.1 → T3.3 → T3.4
```

---

## Timeline & Estimates

| Phase | Effort | Notes |
|---|---|---|
| Phase 1 — Backend | ~7-8h | Blocked by T1.1 API verification |
| Phase 2 — Frontend | ~6h | Parallel sau khi T1.5 done |
| Phase 3 — Cron + Logs | ~5h | Sau M1 + M2 |
| **Total** | **~18-20h** | |

---

## Risks & Mitigation

| Risk | Probability | Impact | Mitigation |
|---|---|---|---|
| thestatsapi.com không có Asian handicap odds | High | Critical | Verify trước (T1.1). Nếu không có → đổi sang API khác (TheOddsAPI) |
| thestatsapi.com không có exact score odds | High | Medium | Scope lại — exact score odds vẫn nhập tay hoặc dùng API khác |
| ExternalID trong WcMatch không khớp với fixture ID của API | Medium | Medium | Thêm mapping table hoặc search by team names + date |
| Rate limit quá thấp làm cron bị block | Low | Medium | Tăng cron interval, cache response 30min |
| Preview data inaccurate (odds thay đổi giữa preview và confirm) | Low | Low | Hiển thị "fetched_at" timestamp để admin biết độ mới của data |

---

## Resources Needed

- API key thestatsapi.com (đã có)
- Xem API documentation để xác định endpoints + response format
- Env var `STATSAPI_KEY` được thêm vào production deploy
