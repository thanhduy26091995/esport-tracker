---
phase: testing
title: API Performance Optimization — Testing Strategy
description: Test cases for caching correctness, pagination fix, pool config, and load validation
---

# Testing Strategy

## Test Coverage Goals

- Unit tests: 100% of new `cache.go` code; service-layer cache logic
- Integration tests: paginated GET /matches correctness; cache hit/miss behavior end-to-end
- Performance tests: before/after p95 latency comparison for the 3 critical endpoints
- Manual smoke tests: verify no UI regressions after all changes

## Unit Tests

### Cache Abstraction (`internal/cache/cache_test.go`)

- [ ] Test `Get` returns false on empty cache
- [ ] Test `Set` then `Get` returns stored value
- [ ] Test TTL expiry — after TTL, `Get` returns false
- [ ] Test `Delete` removes key
- [ ] Test `DeletePrefix` removes all keys with matching prefix, leaves others
- [ ] Test concurrent `Set`/`Get` does not panic (goroutine safety)

### WC Service — Leaderboard Cache (`wc_service_test.go`)

- [ ] First call hits repository (mock: verify repo called once)
- [ ] Second call within TTL returns cached value (mock: verify repo called once total)
- [ ] After `invalidateLeaderboard()`, next call hits repository again
- [ ] Repository error on cache miss → error propagated, nothing cached

### Match Handler — Pagination (`match_handler_test.go`)

- [ ] GET /matches?page=1&limit=20 returns exactly 20 records
- [ ] GET /matches?page=2&limit=20 returns next 20 (no overlap with page 1)
- [ ] GET /matches with total=5, limit=20 returns 5 records and correct `total` field
- [ ] Bonuses returned belong to the date range of the page's matches only
- [ ] Empty DB returns `{data: [], total: 0}`

## Integration Tests

- [ ] GET /wc/leaderboard returns 200 with cached data on second call within TTL
- [ ] POST to settle a bet invalidates leaderboard cache (next GET re-queries DB)
- [ ] GET /matches page 1 + page 2 together cover all records with no duplicates
- [ ] DB pool: 25 concurrent requests to /wc/leaderboard all succeed without timeout
- [ ] gzip: response includes `Content-Encoding: gzip` header for large payloads

## End-to-End Tests

- [ ] Open leaderboard page, check it loads in <1s (cached subsequent loads <200ms)
- [ ] Open match feed page 1 and page 2, verify correct match order and no duplicates
- [ ] Admin settles a match → leaderboard on next page load reflects new scores
- [ ] Chat page still loads correctly (pagination not broken by changes)

## Performance Testing

### Baseline Measurement (before changes)
Run against staging with current code:

```bash
# wrk load test: 20 concurrent users, 30 seconds
wrk -t4 -c20 -d30s http://localhost:8080/api/wc/leaderboard
wrk -t4 -c20 -d30s "http://localhost:8080/api/matches?page=1&limit=20"
wrk -t4 -c20 -d30s http://localhost:8080/api/wc/matches
```

Record: req/sec, latency p50/p95/p99, error count.

### Target After Changes

| Endpoint | p95 Before | p95 Target |
|----------|-----------|-----------|
| GET /wc/leaderboard | ~800ms | ≤300ms (cached) |
| GET /matches | scales with N records | constant regardless of total count |
| GET /wc/matches | ~400ms | ≤150ms (cached) |

### Post-change Load Test

Repeat same wrk commands. Confirm:
- [ ] p95 ≤ targets above
- [ ] 0 errors (no DB connection timeouts, no 500s)
- [ ] RSS growth ≤ 10MB (cache memory footprint)

## Test Data

- Minimum 200 matches and 500 wc_predictions in test DB to exercise pagination properly
- At least 10 users with varied prediction counts for leaderboard
- At least 1 custom bet with 5+ options to test batch insert

## Manual Testing

- [ ] Visit leaderboard page — loads correctly, scores accurate
- [ ] Place a new prediction — confirm it appears after cache TTL expires (or manual invalidation)
- [ ] Admin: settle a match — confirm leaderboard updates
- [ ] Navigate match feed page 1 → 2 → 3 — no duplicate matches
- [ ] Check browser network panel: responses have `Content-Encoding: gzip` header
- [ ] Check backend logs: no SQL statements for leaderboard on repeated hits (logger.Warn level shows only slow queries)

## Bug Tracking

- If paginated matches show duplicates → check ORDER BY stability (add secondary sort by `id`)
- If leaderboard stale after settlement → check invalidation call was added to all settlement paths
- If connection errors under load → reduce MaxOpenConns or check PostgreSQL max_connections setting
