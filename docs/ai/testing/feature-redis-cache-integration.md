---
phase: testing
title: Redis Cache Integration — Testing Strategy
description: Unit and integration tests for CacheStore, GetOrFetch helper, singleflight behavior, and cache invalidation
---

# Testing Strategy

## Test Coverage Goals

- **100% coverage** of `cache/` package (interface, Redis impl, go-cache impl, noop)
- **Singleflight behavior** verified with concurrent goroutines — this is the hardest-to-see behavior, must be explicit
- **Cache invalidation** verified per write operation — delete correct keys, no extra keys deleted
- **Redis error resilience** — errors in Get/Set/Delete do not propagate to callers
- **TTL expiry** — verify keys expire and trigger re-fetch (using miniredis clock control)

## Unit Tests

### `cache/redis_test.go` — using `miniredis`

```bash
go get github.com/alicebob/miniredis/v2
```

- [ ] **Get miss**: key absent → `("", false)` returned
- [ ] **Get hit**: key set → `(value, true)` returned
- [ ] **Get after TTL expiry**: set key, advance miniredis clock past TTL, get → `("", false)`
- [ ] **Set stores value with TTL**: set → get → correct value; verify TTL via `miniredis.TTL(key)`
- [ ] **Delete removes key**: set → delete → get → miss
- [ ] **DeleteByPattern**: set 3 keys with prefix, call `DeleteByPattern("prefix:*")` → all 3 gone, unrelated key untouched
- [ ] **Get on Redis error**: close miniredis server → `Get` returns `("", false)`, no panic
- [ ] **Set on Redis error**: closed server → `Set` returns error (caller ignores it)
- [ ] **DeleteByPattern with empty result**: no matching keys → returns nil (no error)

```go
// Example test structure
func TestRedisCache_GetMiss(t *testing.T) {
    mr := miniredis.RunT(t)
    rc, _ := NewRedisCache("redis://" + mr.Addr())
    
    val, ok := rc.Get("nonexistent-key")
    
    assert.False(t, ok)
    assert.Empty(t, val)
}

func TestRedisCache_TTLExpiry(t *testing.T) {
    mr := miniredis.RunT(t)
    rc, _ := NewRedisCache("redis://" + mr.Addr())
    
    rc.Set("key", "value", 1*time.Minute)
    mr.FastForward(2 * time.Minute)   // advance miniredis clock
    
    _, ok := rc.Get("key")
    assert.False(t, ok)
}
```

### `cache/cache_test.go` — `GetOrFetch` helper

- [ ] **First call invokes fetch fn**: call `GetOrFetch` with empty cache → fetch fn called once → value returned
- [ ] **Second call hits cache**: call twice → fetch fn called exactly once total
- [ ] **Cache populated after first call**: after `GetOrFetch`, `store.Get(key)` returns the value
- [ ] **Fetch error propagates**: fetch fn returns error → `GetOrFetch` returns same error, nothing cached
- [ ] **Corrupt cache entry falls through to fetch**: manually set invalid JSON → `GetOrFetch` calls fetch fn and overwrites

### `cache/cache_test.go` — Singleflight behavior

This test verifies the core stampede prevention:

```go
func TestGetOrFetch_Singleflight(t *testing.T) {
    store := NewNoopCache()
    group := &singleflight.Group{}
    callCount := atomic.Int32{}
    
    fetch := func() (string, error) {
        callCount.Add(1)
        time.Sleep(50 * time.Millisecond)  // simulate DB latency
        return "result", nil
    }
    
    // Launch 10 goroutines simultaneously — all miss the noop cache
    var wg sync.WaitGroup
    results := make([]string, 10)
    for i := range results {
        wg.Add(1)
        go func(i int) {
            defer wg.Done()
            v, _ := GetOrFetch(store, group, "key", time.Minute, fetch)
            results[i] = v
        }(i)
    }
    wg.Wait()
    
    // fetch fn must have been called exactly once despite 10 concurrent misses
    assert.Equal(t, int32(1), callCount.Load())
    // all goroutines got the same result
    for _, r := range results {
        assert.Equal(t, "result", r)
    }
}
```

- [ ] **Concurrent misses**: 10 goroutines, all miss, fetch fn called exactly once
- [ ] **All callers receive same value**: all 10 goroutines get `"result"`
- [ ] **Second wave after first completes**: after group settles, second wave calls fetch fn again (singleflight is not a persistent cache)

### Service Unit Tests — Invalidation

For each service that invalidates cache, add a test with a spy/fake `CacheStore`:

```go
type SpyCacheStore struct {
    NoopCache
    deletedKeys []string
}

func (s *SpyCacheStore) Delete(key string) error {
    s.deletedKeys = append(s.deletedKeys, key)
    return nil
}

func (s *SpyCacheStore) DeleteByPattern(pattern string) error {
    s.deletedKeys = append(s.deletedKeys, pattern)
    return nil
}
```

- [ ] **`UserService.CreateUser` invalidates**: `esport:users:all`, `esport:users:leaderboard`
- [ ] **`UserService.UpdateUser` invalidates**: `esport:users:all`, `esport:users:leaderboard`, `esport:users:payment-ranking`
- [ ] **`UserService.DeleteUser` invalidates**: `esport:users:all`, `esport:users:leaderboard`
- [ ] **`MatchService.CreateMatch` invalidates**: `esport:matches:all:*` pattern, `esport:users:leaderboard`, `esport:users:all`
- [ ] **`MatchService.DeleteMatch` invalidates**: same keys as CreateMatch
- [ ] **`ConfigService.UpdateConfig` invalidates**: `esport:config`
- [ ] **WcService settle/sync invalidates**: `wc:leaderboard` and/or `wc:matches:all`

## Integration Tests

- [ ] **Full round-trip**: start miniredis → `NewRedisCache` → `GetOrFetch` → verify key in miniredis
- [ ] **Pattern delete across real Redis**: set 5 paginated match keys → `DeleteByPattern("esport:matches:all:*")` → all 5 gone
- [ ] **Fallback on Redis down**: configure Redis URL to unreachable address → service still returns data from DB (no 5xx)

## End-to-End Tests

No automated E2E tests for the cache layer — it's transparent to the API. Manual verification is sufficient:

- [ ] Start server with `REDIS_URL` set → confirm startup log says "Cache backend: Redis"
- [ ] `GET /api/v1/users/leaderboard` twice → first response slow (DB), second fast (Redis)
- [ ] `redis-cli GET esport:users:leaderboard` → see cached JSON
- [ ] `POST /api/v1/matches` → `redis-cli KEYS esport:*` → leaderboard key gone
- [ ] Next `GET /api/v1/users/leaderboard` → key repopulated in Redis

- [ ] Start server without `REDIS_URL` → confirm startup log says "Cache backend: go-cache"
- [ ] Same requests work correctly (proving go-cache fallback functional)

## Test Data

- Unit tests use `miniredis` (in-process Redis for tests — no external server needed)
- Service tests use `NoopCache` (or `SpyCacheStore` for invalidation tests) — no Redis required
- No test seed data changes needed

## Test Reporting & Coverage

Run with:
```bash
cd backend
go test ./internal/cache/... -v -count=1
go test ./internal/service/... -v -count=1 -run TestCache
go test ./... -cover
```

Target: `internal/cache/` at 100% coverage.

## Manual Testing

**Happy path:**
1. `docker run -d -p 6379:6379 redis:7-alpine`
2. Set `REDIS_URL=redis://localhost:6379/0` in `.env`
3. `go run cmd/server/main.go` — confirm "Cache backend: Redis" in logs
4. Open the leaderboard page — network tab should show fast subsequent loads
5. Record a match → check `redis-cli KEYS *` → leaderboard key missing (invalidated)
6. Reload leaderboard → key repopulated

**Fallback path:**
1. Remove `REDIS_URL` from `.env`
2. Restart server — confirm "Cache backend: go-cache" in logs
3. All features work normally

**Redis disconnect:**
1. Start with Redis running
2. `docker stop redis-dev` while server is running
3. Make a request — should succeed (falls back to DB), no 500 error

## Performance Testing

Informal benchmark only (no tooling needed):
- Before: open browser DevTools, record a `GET /api/v1/users/leaderboard` — note response time (should be 50–500ms)
- After: second request for same endpoint — should be < 10ms (Redis round-trip)
- Observed speedup documents the cache working as intended

## Bug Tracking

Known risks to watch for:
- **Wrong key deleted on write**: check the invalidation map in design doc and cross-reference with spy tests
- **singleflight key reuse across types**: if two different structs use the same cache key string, singleflight `Do` result would be cast to the wrong type → panic. Mitigate by unique key namespacing and test coverage.
- **Nil pointer in service constructors**: after adding `cache CacheStore` field, ensure no existing test constructs a service with `nil` cache. All existing tests must pass `NoopCache`.
