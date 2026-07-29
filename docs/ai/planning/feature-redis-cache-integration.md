---
phase: planning
title: Redis Cache Integration — Planning
description: Task breakdown, dependencies, effort estimates, and implementation order
---

# Project Planning & Task Breakdown

## Milestones

- [ ] **M1: Cache Infrastructure** — `CacheStore` interface + Redis + fallbacks wired in router
- [ ] **M2: Core Esport Cache** — Users, Leaderboard, Matches, Config, Fund cached with invalidation
- [ ] **M3: WC Cache** — WC Leaderboard, WC Matches, WC Config, migrate analytics cache
- [ ] **M4: Tests & Docs** — Unit tests with miniredis, integration smoke test, update knowledge docs

## Task Breakdown

### Phase 1: Cache Infrastructure

- [ ] **1.1** Add `go-redis/redis/v9` and `golang.org/x/sync` to `go.mod`
  - Run `go get github.com/redis/go-redis/v9`
  - `golang.org/x/sync` is likely already indirect; verify
  - Effort: 10 min

- [ ] **1.2** Create `backend/internal/cache/cache.go`
  - Define `CacheStore` interface: `Get`, `Set`, `Delete`, `DeleteByPattern`, `GetInt`, `Incr`
  - `GetInt` / `Incr` are needed for the versioned match cache key counter
  - Implement generic `GetOrFetch[T]` helper using singleflight + JSON
  - Effort: 30 min

- [ ] **1.3** Create `backend/internal/cache/redis.go`
  - `NewRedisCache(redisURL string)` — parse URL, Ping on init
  - Implement all `CacheStore` methods
  - `DeleteByPattern` uses `SCAN + DEL` loop (cursor-based, 100 keys per page)
  - Effort: 45 min

- [ ] **1.4** Create `backend/internal/cache/gocache.go`
  - Wrap `github.com/patrickmn/go-cache` as `CacheStore`
  - `DeleteByPattern` uses an `Items()` scan + prefix match
  - Effort: 20 min

- [ ] **1.5** Create `backend/internal/cache/noop.go`
  - `Get` always returns `("", false)`; `Set`/`Delete`/`DeleteByPattern` are no-ops
  - Used in unit tests that don't want cache side effects
  - Effort: 10 min

- [ ] **1.6** Wire cache into `router.go`
  - Read `REDIS_URL` env; if set → `NewRedisCache`, else → `NewGoCacheStore`
  - Log which backend is active on startup
  - Effort: 15 min

### Phase 2: Core Esport Cache

- [ ] **2.1** Update `UserService` constructor to accept `CacheStore`
  - Add `cache CacheStore` + `group singleflight.Group` fields
  - Wrap `GetAll` → cache key `esport:users:all`, TTL 10 min
  - Wrap `GetLeaderboard` → cache key `esport:users:leaderboard`, TTL 5 min
  - Wrap `GetPaymentRanking` → cache key `esport:users:payment-ranking`, TTL 10 min
  - Effort: 45 min

- [ ] **2.2** Add cache invalidation to `UserService` write paths
  - `CreateUser` → delete `esport:users:all`, `esport:users:leaderboard`
  - `UpdateUser` → delete `esport:users:all`, `esport:users:leaderboard`, `esport:users:payment-ranking`
  - `DeleteUser` → delete `esport:users:all`, `esport:users:leaderboard`
  - Effort: 20 min

- [ ] **2.3** Update `MatchService` constructor to accept `CacheStore`
  - Wrap `GetAllMatches` → versioned key `esport:matches:v{N}:{page}:{limit}:{playerID}`, TTL 2 min
  - Read path: `GetInt("esport:matches:version")` → build versioned key → `GetOrFetch`
  - Effort: 30 min

- [ ] **2.4** Add cache invalidation to `MatchService` write paths
  - `CreateMatch` → `Incr("esport:matches:version")`, delete `esport:users:leaderboard`, `esport:users:all`
  - `DeleteMatch` → same: `Incr` version + delete leaderboard + users
  - No `DeleteByPattern` needed — version bump orphans old keys; they expire via TTL
  - Effort: 20 min

- [ ] **2.4a** Add `ScoreBonusService` cache invalidation
  - Add `cache CacheStore` field to `ScoreBonusService`
  - `CreateBonus` → delete `esport:users:leaderboard`, `esport:users:all`
  - `DeleteBonus` → same keys
  - Effort: 15 min

- [ ] **2.5** Update `ConfigService` constructor to accept `CacheStore`
  - Wrap `GetConfig` → cache key `esport:config`, TTL 30 min
  - `UpdateConfig` → delete `esport:config`
  - Effort: 25 min

- [ ] **2.6** Update `FundService` constructor to accept `CacheStore`
  - Wrap `GetFundSummary` → cache key `esport:fund:totals`, TTL 15 min
  - `SettlementService.CreateSettlement` → delete `esport:fund:totals`, `esport:users:leaderboard`, `esport:users:payment-ranking`
  - Note: `SettlementService` needs access to cache; add as field or call via `FundService` invalidation hook
  - Effort: 30 min

### Phase 3: WC Cache

- [ ] **3.1** Update `WcService` constructor to accept `CacheStore`
  - Wrap `GetLeaderboard(tournamentType)` → cache key `wc:leaderboard:{tournament_type}`, TTL 5 min
  - Wrap `GetAllMatches(tournamentType)` → cache key `wc:matches:all:{tournament_type}`, TTL 5 min
  - All keys scoped by `tournament_type` so WC2026 and ASEAN Cup caches are independent
  - Effort: 40 min

- [ ] **3.2** Add cache invalidation to WC write paths
  - WC `SettleBets(tournamentType)` → delete `wc:leaderboard:{tournament_type}`
  - WC cron sync → delete `wc:matches:all:{tournament_type}` (update `cron/wc_sync.go`)
  - WC admin sync → delete `wc:matches:all:{tournament_type}`
  - WC `UpdateConfig(tournamentType)` → delete `wc:config:{tournament_type}`
  - Effort: 25 min

- [ ] **3.3** Cache WC config read
  - Wherever `wc_config` is read → cache key `wc:config:{tournament_type}`, TTL 15 min
  - Effort: 20 min

- [ ] **3.4** Migrate `analyticsCache` from raw go-cache to `CacheStore`
  - `WcAnalyticsService` currently holds a `*gocache.Cache` directly
  - Change constructor to accept `CacheStore` instead
  - Replace `gocache.Set/Get` calls with `CacheStore.Set/Get`
  - Remove direct `gocache` import from `router.go` for analytics
  - Effort: 30 min

### Phase 4: Tests & Documentation

- [ ] **4.1** Unit tests for `RedisCache` using `miniredis`
  - `go get github.com/alicebob/miniredis/v2`
  - Test: `Get` miss, `Get` hit, `Set` with TTL, `Delete`, `DeleteByPattern`
  - Test: Redis error returns miss (not panic)
  - Test: TTL expiry (advance miniredis clock)
  - Effort: 60 min

- [ ] **4.2** Unit tests for `GetOrFetch` helper
  - Test: first call → fetch fn invoked, value cached
  - Test: second call → fetch fn NOT invoked (cache hit)
  - Test: concurrent calls while in-flight → fetch fn invoked only once (singleflight)
  - Use `NoopCache` + call counter to verify singleflight behavior
  - Effort: 45 min

- [ ] **4.3** Update service unit tests
  - Existing service tests that use `NewUserService(...)` etc. need to pass a `NoopCache`
  - Ensure existing test coverage doesn't break
  - Effort: 30 min

- [ ] **4.4** Update `docs/ai/knowledge/backend-patterns.md`
  - Add "Cache Pattern" section describing Cache-Aside, singleflight, and invalidation conventions
  - Effort: 15 min

- [ ] **4.5** Update `CLAUDE.md` design doc table with link to this feature
  - Effort: 5 min

## Dependencies

```
1.1 → 1.2 → 1.3, 1.4, 1.5
1.3, 1.4, 1.5 → 1.6
1.6 → 2.1 → 2.2
      2.3 → 2.4
      2.5
      2.6
1.6 → 3.1 → 3.2
      3.3
      3.4
All Phase 2+3 → 4.1, 4.2, 4.3
```

External: `go-redis/v9` package available; `miniredis` for tests; Redis server for local dev (Docker or local install).

## Timeline & Estimates

| Phase | Tasks | Estimated Effort |
|-------|-------|-----------------|
| Phase 1: Infrastructure | 1.1–1.6 | ~2 hours |
| Phase 2: Core Esport Cache | 2.1–2.6 | ~2.5 hours |
| Phase 3: WC Cache | 3.1–3.4 | ~2 hours |
| Phase 4: Tests & Docs | 4.1–4.5 | ~2.5 hours |
| **Total** | | **~9 hours** |

Suggested order: complete M1 → M2 → M4 (partial tests) → M3 → M4 (finalize).  
M3 is independent of M2 and can be done in any order after M1.

## Risks & Mitigation

| Risk | Likelihood | Impact | Mitigation |
|------|-----------|--------|-----------|
| `DeleteByPattern` misses keys due to wrong glob syntax | Medium | Low (stale cache, not failure) | Test pattern delete explicitly with miniredis |
| singleflight key collision between services | Low | Medium | Each service uses its own `singleflight.Group`; keys are namespaced (`esport:` vs `wc:`) |
| Redis unavailable in prod → all requests fail | Low | High | Redis error in `Get/Set` treated as cache miss; service falls back to DB |
| Missed invalidation → stale data visible to users | Medium | Low-Medium | TTL is the backstop; worst case users see stale data for at most the key's TTL |
| go-cache fallback behavior differs from Redis | Low | Low | Both implement `CacheStore`; test both implementations |
| Existing tests break due to new constructor params | Medium | Low | Pass `NoopCache` to all service constructors in tests |

## Resources Needed

- **Redis**: Docker (`docker run -p 6379:6379 redis:7`) or Redis Cloud free tier for prod
- **Go packages**: `github.com/redis/go-redis/v9`, `golang.org/x/sync`, `github.com/alicebob/miniredis/v2` (test only)
- **Environment variable**: `REDIS_URL=redis://localhost:6379/0` added to `.env`
- **Knowledge**: `docs/ai/knowledge/backend-patterns.md` (existing DI patterns to follow)
