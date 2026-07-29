---
phase: requirements
title: Redis Cache Integration — Requirements
description: Integrate Redis as shared cache layer for FC25 Tracker to improve performance and teach Cache-Aside pattern, stampede prevention, and TTL strategy
---

# Requirements & Problem Understanding

## Problem Statement

Every HTTP request in the FC25 Tracker currently hits PostgreSQL directly. The app's data mutation pattern is highly asymmetric:

- **Record Match (POST /matches)** — the only operation that changes core esport data during the day
- **All GET endpoints** (leaderboard, users, config, fund, matches feed) — read-only, data changes rarely or only during admin operations

This means > 95% of read requests are fetching data that hasn't changed since the last request. Without a cache layer, each request pays full DB query cost for identical results, which is unnecessary and wasteful.

**Secondary problem:** This is a learning project for Redis fundamentals. The implementation must demonstrate real Redis patterns: Cache-Aside, singleflight (Cache Stampede prevention), explicit TTL design, and cache invalidation on write.

## Goals & Objectives

### Primary Goals
- Introduce Redis as the shared cache layer with `go-redis/redis` client
- Implement **Cache-Aside pattern** across key read paths (check cache → miss → DB → populate cache)
- Prevent **Cache Stampede** using `golang.org/x/sync/singleflight` — when a key expires and N goroutines miss simultaneously, only one goes to DB
- Design intentional **TTL strategy** per cache key based on actual data change frequency
- Implement **cache invalidation on write** — when a record is created/updated/deleted, the relevant cache keys are deleted so the next read fetches fresh data

### Secondary Goals
- Replace the existing ad-hoc `go-cache` usage (`analyticsCache` in router.go) with the new `CacheStore` abstraction
- Keep cache logic in the **service layer** — repositories stay pure DB access, handlers stay pure HTTP parsing
- Make the `CacheStore` interface concrete enough to be testable with a mock

### Non-Goals
- **No write-through caching** — too complex for the learning goal; write-invalidate is simpler and sufficient
- **No Redis Cluster / Sentinel** — single-node Redis, this is a friend-group app
- **No frontend changes** — cache is invisible to the frontend; all endpoints keep the same response shape
- **No Redis persistence (AOF/RDB)** — cache is ephemeral; a Redis restart is fine (TTL-based repopulation handles it)
- **No cache for WC settlement/bet writes** — those are infrequent admin operations; no need to add complexity

## User Stories & Use Cases

**As a developer learning Redis**, I want to:
- Trace through a cache miss: request → Redis miss → DB query → populate Redis → return → next request hits Redis
- See a concrete singleflight implementation so I understand how it prevents duplicate DB queries on key expiry under concurrency
- Read TTL decisions explained with reasoning per data type, not just arbitrary numbers
- Observe cache invalidation: record a match → relevant cache keys deleted → next leaderboard read is fresh

**As a user of the app**, I want:
- The player leaderboard to load faster (no DB aggregation query on every page load)
- The match feed to respond quickly (no full-table scan on each navigation)
- Config and fund totals to be instant (set-and-forget data, almost never changes)

### Key Workflows

1. **Cache hit path**: `GET /api/v1/users/leaderboard` → service calls `cache.Get("esport:users:leaderboard")` → hit → return JSON directly, no DB
2. **Cache miss path**: same request, key expired → singleflight barrier → one goroutine calls DB → populates key with TTL → all waiting goroutines get same result
3. **Write invalidation path**: `POST /api/v1/matches` → `CreateMatch()` succeeds → service calls `cache.Delete("esport:users:leaderboard")`, `cache.Delete("esport:matches:all:*")` → next read repopulates from DB

## Success Criteria

- [ ] If `REDIS_URL` is set but Redis is unreachable at startup, server logs a warning and automatically falls back to go-cache — it does not crash
- [ ] `CacheStore` interface defined with `Get`, `Set`, `Delete`, `DeleteByPattern` operations
- [ ] Redis implementation of `CacheStore` fully tested with unit tests using a mock or `miniredis`
- [ ] Cache-Aside pattern applied to: users list, leaderboard, payment ranking, matches feed, config, fund totals, WC leaderboard, WC matches list, WC config
- [ ] WC cache keys are scoped by `tournament_type` (e.g. `wc:leaderboard:wc2026`, `wc:leaderboard:asean_cup`) so each tournament's cache is independently managed — a write to one tournament does not evict the other
- [ ] Singleflight applied to every cache-miss path (not just some) — consistent pattern across all services
- [ ] Every write operation (create/delete match, create/delete score bonus, update config, settle, etc.) invalidates the relevant cache keys — score bonus writes invalidate `esport:users:leaderboard` and `esport:users:all` since bonuses directly affect player rankings
- [ ] TTL documented per key in a central table in the design doc
- [ ] On Redis failure (connection lost mid-request), the service falls back to DB transparently — cache errors must never cause 5xx responses
- [ ] Repeated GET requests to cached endpoints (leaderboard, users, matches feed) complete in < 5ms, verifiable in browser DevTools Network tab (vs. observed 50–300ms on direct DB queries)

## Constraints & Assumptions

### Technical Constraints
- Go 1.21+, Gin, GORM — existing stack unchanged
- Redis 7.x (Docker locally, Redis Cloud / Railway free tier on prod)
- `go-redis/redis/v9` — the de-facto standard Go Redis client
- `golang.org/x/sync/singleflight` — stdlib-adjacent, no extra dependency risk
- Cache logic lives in the **service layer** only (per existing backend patterns in `docs/ai/knowledge/backend-patterns.md`)

### Business Constraints
- Zero downtime migration: app must work without Redis until `REDIS_URL` is set (dev mode falls back to no-op cache or go-cache stub)
- No breaking API changes

### Assumptions
- Redis is single-node (no clustering complexity)
- Data volumes are small (friend group), so full-list caching (not partial/cursor) is appropriate
- A cache miss on every Redis restart is acceptable — the app keeps working, just slower until keys warm up

## Questions & Open Items

- **Dev fallback**: Should `REDIS_URL` being unset use the existing `go-cache` as fallback, or use a no-op cache and always hit DB? → Recommend: `go-cache` fallback in dev, Redis required in prod (controlled by env var presence)
- **Pattern delete vs version suffix**: ~~SCAN+DEL approach~~ — **Decision: use versioned keys** for paginated match cache. A `esport:matches:version` counter is stored in Redis and incremented on each match write. Read keys include the current version (`esport:matches:v{N}:all:{page}:{limit}:{playerID}`), making old pages orphans that expire via TTL. No scan needed; old keys self-clean. Adds one extra Redis GET per read (to fetch version), but no scan overhead and teaches a real-world cache-busting technique.
- **Analytics cache**: ~~Unconfirmed~~ — **Decision: in scope.** The existing `analyticsCache` (raw `*gocache.Cache` in `router.go`) will be migrated to use the `CacheStore` abstraction, removing the direct `gocache` import from `router.go` and unifying all cache usage behind one interface.
