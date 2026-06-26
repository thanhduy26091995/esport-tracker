---
phase: planning
title: API Performance Optimization — Project Planning
description: Task breakdown, dependencies, and effort estimates for API performance work
---

# Project Planning & Task Breakdown

## Milestones

- [ ] **M1: Quick wins** — Pool config + DB logger + gzip (1 hour, zero risk, deploy immediately)
- [ ] **M2: Pagination fix** — Fix GET /matches app-level pagination (2–3 hours, high impact)
- [ ] **M3: Caching layer** — go-cache + leaderboard/WC matches caching (3–4 hours, highest impact)
- [ ] **M4: DB indexes** — Migration with 4 indexes (30 min, low risk)
- [ ] **M5: Cleanup** — Batch insert, code cleanup, load test (1–2 hours)

## Task Breakdown

### Phase 1: Quick Wins (no behavior change, safe to ship immediately)

- [ ] **1.1** Configure DB connection pool in `database.go`
  - SetMaxOpenConns(25), SetMaxIdleConns(5), SetConnMaxLifetime(5m), SetConnMaxIdleTime(30s)
  - File: `backend/internal/database/database.go`
  - Effort: 15 min

- [ ] **1.2** Change DB logger from `logger.Info` to `logger.Warn` in `database.go`
  - Eliminates verbose SQL log noise in production
  - File: `backend/internal/database/database.go`
  - Effort: 5 min

- [ ] **1.3** Add gzip compression middleware to router
  - `go get github.com/gin-contrib/gzip`
  - Add `router.Use(gzip.Gzip(gzip.DefaultCompression))` after CORS
  - File: `backend/internal/api/router.go`
  - Effort: 20 min

### Phase 2: Pagination Fix (high impact, requires care)

- [ ] **2.1** Fix `match_repository.go` GetAll — ensure DB-level LIMIT/OFFSET for matches
  - Verify current GetAll already passes limit/offset to SQL (it does per line 35)
  - File: `backend/internal/repository/match_repository.go`
  - Effort: 30 min

- [ ] **2.2** Fix `match_handler.go` GetAllMatches — remove app-level "fetch all" pattern
  - Replace `GetAllMatches(0, 0)` with paginated call
  - Scope bonus fetch to page date range instead of all bonuses
  - Maintain same response shape: `{data, total, page, limit}`
  - File: `backend/internal/api/match_handler.go`
  - Effort: 2 hours

- [ ] **2.3** Test pagination fix
  - Verify page 1, page 2, last page return correct non-overlapping records
  - Verify total count is correct
  - Effort: 30 min

### Phase 3: Caching Layer (highest impact)

- [ ] **3.1** Create cache abstraction `internal/cache/cache.go`
  - `CacheStore` interface: `Get(key) (any, bool)`, `Set(key, val, ttl)`, `Delete(key)`, `DeletePrefix(prefix)`
  - `GoCacheStore` implementation using `github.com/patrickmn/go-cache`
  - `go get github.com/patrickmn/go-cache`
  - File: `backend/internal/cache/cache.go`
  - Effort: 45 min

- [ ] **3.2** Wire `CacheStore` into DI in `router.go`
  - Instantiate `cache.NewGoCacheStore(5*time.Minute, 10*time.Minute)` (default TTL, cleanup interval)
  - Pass into WC service constructor
  - File: `backend/internal/api/router.go`
  - Effort: 20 min

- [ ] **3.3** Cache GET /wc/leaderboard in `wc_service.go`
  - Key: `wc:leaderboard`, TTL: 2 min
  - Invalidate on: `SettleMatch`, `RecalculateAll`, `SettleCustomBet`, `SettleChampionPrediction`
  - File: `backend/internal/service/wc_service.go`
  - Effort: 1 hour

- [ ] **3.4** Cache GET /wc/matches in `wc_service.go`
  - Key: `wc:matches:all`, TTL: 5 min
  - Invalidate on: `SyncMatches`, `AdminUpdateMatch`
  - File: `backend/internal/service/wc_service.go`
  - Effort: 45 min

- [ ] **3.5** Cache GET /wc/config in `wc_service.go`
  - Key: `wc:config`, TTL: 10 min
  - Invalidate on: `UpdateConfig`
  - Effort: 30 min

### Phase 4: DB Indexes

- [ ] **4.1** Write migration file with 4 indexes
  - `idx_match_participants_match_id` ON match_participants(match_id)
  - `idx_wc_bets_result` ON wc_bets(result)
  - `idx_wc_predictions_result` ON wc_predictions(result)
  - `idx_wc_chat_messages_created_at` ON wc_chat_messages(created_at DESC)
  - All use `IF NOT EXISTS` — safe to replay
  - File: `backend/migrations/YYYYMMDD_add_performance_indexes.sql` (or Go migration)
  - Effort: 30 min

### Phase 5: Cleanup & Batch Insert

- [ ] **5.1** Fix custom bet options loop insert → batch insert
  - Replace `for range options { tx.Create(&option) }` with `tx.Create(&options)`
  - File: `backend/internal/repository/wc_custom_bet_repository.go`
  - Effort: 20 min

- [ ] **5.2** Load test with wrk or k6 (20 concurrent users, 30s)
  - Target endpoints: `/wc/leaderboard`, `/matches`, `/wc/matches`
  - Record p50/p95/p99 before and after each phase
  - Effort: 1 hour

## Dependencies

```
1.1, 1.2, 1.3  →  independent, run in any order
2.1            →  prerequisite for 2.2
2.2            →  prerequisite for 2.3
3.1            →  prerequisite for 3.2, 3.3, 3.4, 3.5
3.2            →  prerequisite for 3.3, 3.4, 3.5
4.1            →  independent (DB migration, run anytime)
5.1            →  independent
5.2            →  should run after all phases complete
```

## Timeline & Estimates

| Phase | Tasks | Effort | Notes |
|-------|-------|--------|-------|
| Phase 1: Quick wins | 1.1–1.3 | ~40 min | Zero risk, deploy same day |
| Phase 2: Pagination fix | 2.1–2.3 | ~3 hours | Requires careful testing |
| Phase 3: Caching | 3.1–3.5 | ~4 hours | Highest ROI |
| Phase 4: Indexes | 4.1 | ~30 min | Run during low-traffic window |
| Phase 5: Cleanup | 5.1–5.2 | ~1.5 hours | Nice-to-have |
| **Total** | | **~10 hours** | Spread over 2–3 sessions |

## Risks & Mitigation

| Risk | Likelihood | Impact | Mitigation |
|------|-----------|--------|-----------|
| Cache serves stale leaderboard during active betting | Medium | Low | TTL 2 min acceptable; add manual invalidation endpoint for admin |
| Pagination fix breaks existing frontend | Medium | High | Keep same response shape; test with real data before deploy |
| go-cache adds memory overhead | Low | Low | Max ~5MB for current data sizes; monitor RSS |
| DB index migration locks table | Low | Medium | `CREATE INDEX CONCURRENTLY` if on live PostgreSQL (or use `IF NOT EXISTS` + run during maintenance) |
| Pool config too aggressive | Low | Low | Start with MaxOpenConns=25; tune down if DB reports too many connections |

## Resources Needed

- `github.com/patrickmn/go-cache` — in-memory TTL cache
- `github.com/gin-contrib/gzip` — gzip response middleware
- No new infrastructure required
- Load testing tool: wrk or k6 (optional but recommended)
