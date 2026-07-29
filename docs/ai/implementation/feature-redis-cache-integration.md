---
phase: implementation
title: Redis Cache Integration — Implementation Guide
description: Step-by-step code patterns, file-by-file changes, and integration notes for Redis Cache-Aside integration
---

# Implementation Guide

## Development Setup

### Prerequisites

```bash
# Start Redis locally (Docker)
docker run -d -p 6379:6379 --name redis-dev redis:7-alpine

# Or install locally (Windows)
# Download from https://github.com/microsoftarchive/redis/releases
```

### Environment Variables

Add to `backend/.env`:

```env
REDIS_URL=redis://localhost:6379/0
```

If `REDIS_URL` is absent, the app uses go-cache in-memory fallback automatically.

### Dependencies

```bash
cd backend
go get github.com/redis/go-redis/v9
go get github.com/alicebob/miniredis/v2   # test-only
# golang.org/x/sync is likely already an indirect dep — verify with go.mod
go get golang.org/x/sync
```

## Code Structure

New directory: `backend/internal/cache/`

```
backend/internal/cache/
├── cache.go       # CacheStore interface + GetOrFetch generic helper
├── redis.go       # Redis implementation
├── gocache.go     # go-cache fallback (dev)
└── noop.go        # No-op (tests)
```

## Implementation Notes

### `cache.go` — Interface + Helper

```go
package cache

import (
    "encoding/json"
    "time"

    "golang.org/x/sync/singleflight"
)

// CacheStore is the abstraction over any cache backend.
type CacheStore interface {
    Get(key string) (string, bool)
    Set(key string, value string, ttl time.Duration) error
    Delete(key string) error
    DeleteByPattern(pattern string) error
}

// GetOrFetch implements Cache-Aside with singleflight stampede prevention.
// T must be JSON-serializable.
func GetOrFetch[T any](
    store CacheStore,
    group *singleflight.Group,
    key string,
    ttl time.Duration,
    fetch func() (T, error),
) (T, error) {
    // 1. Try cache first
    if raw, ok := store.Get(key); ok {
        var result T
        if err := json.Unmarshal([]byte(raw), &result); err == nil {
            return result, nil
        }
        // Corrupt cache entry — fall through to fetch
    }

    // 2. Cache miss — singleflight deduplicates concurrent misses
    val, err, _ := group.Do(key, func() (interface{}, error) {
        result, err := fetch()
        if err != nil {
            return nil, err
        }
        // 3. Populate cache
        if b, jsonErr := json.Marshal(result); jsonErr == nil {
            _ = store.Set(key, string(b), ttl)
        }
        return result, nil
    })
    if err != nil {
        var zero T
        return zero, err
    }
    return val.(T), nil
}
```

### `redis.go` — Redis Implementation

Key implementation notes:
- `ParseURL` handles `redis://`, `rediss://` (TLS), and `redis://:password@host:port/db` formats
- All errors in `Get` are treated as cache misses — never propagate Redis errors to callers
- `DeleteByPattern` uses `SCAN` (not `KEYS`) to avoid blocking Redis on large keysets

```go
package cache

import (
    "context"
    "time"

    "github.com/redis/go-redis/v9"
)

type RedisCache struct {
    client *redis.Client
    ctx    context.Context
}

func NewRedisCache(redisURL string) (*RedisCache, error) {
    opt, err := redis.ParseURL(redisURL)
    if err != nil {
        return nil, err
    }
    client := redis.NewClient(opt)
    ctx := context.Background()
    if err := client.Ping(ctx).Err(); err != nil {
        return nil, err
    }
    return &RedisCache{client: client, ctx: ctx}, nil
}

func (r *RedisCache) Get(key string) (string, bool) {
    val, err := r.client.Get(r.ctx, key).Result()
    if err != nil {
        return "", false // redis.Nil (miss) and network errors both become misses
    }
    return val, true
}

func (r *RedisCache) Set(key, value string, ttl time.Duration) error {
    return r.client.Set(r.ctx, key, value, ttl).Err()
}

func (r *RedisCache) Delete(key string) error {
    return r.client.Del(r.ctx, key).Err()
}

func (r *RedisCache) DeleteByPattern(pattern string) error {
    var cursor uint64
    for {
        keys, next, err := r.client.Scan(r.ctx, cursor, pattern, 100).Result()
        if err != nil {
            return err
        }
        if len(keys) > 0 {
            if err := r.client.Del(r.ctx, keys...).Err(); err != nil {
                return err
            }
        }
        cursor = next
        if cursor == 0 {
            break
        }
    }
    return nil
}
```

### `gocache.go` — Dev Fallback

```go
package cache

import (
    "time"

    gocachepkg "github.com/patrickmn/go-cache"
)

type GoCacheStore struct {
    c *gocachepkg.Cache
}

func NewGoCacheStore(defaultTTL, cleanupInterval time.Duration) *GoCacheStore {
    return &GoCacheStore{c: gocachepkg.New(defaultTTL, cleanupInterval)}
}

func (g *GoCacheStore) Get(key string) (string, bool) {
    if v, found := g.c.Get(key); found {
        if s, ok := v.(string); ok {
            return s, true
        }
    }
    return "", false
}

func (g *GoCacheStore) Set(key, value string, ttl time.Duration) error {
    g.c.Set(key, value, ttl)
    return nil
}

func (g *GoCacheStore) Delete(key string) error {
    g.c.Delete(key)
    return nil
}

func (g *GoCacheStore) DeleteByPattern(pattern string) error {
    // go-cache Items() returns all keys — scan and match prefix
    // Note: go-cache doesn't support glob; we strip trailing '*' and match prefix
    prefix := pattern
    if len(prefix) > 0 && prefix[len(prefix)-1] == '*' {
        prefix = prefix[:len(prefix)-1]
    }
    for k := range g.c.Items() {
        if len(prefix) == 0 || len(k) >= len(prefix) && k[:len(prefix)] == prefix {
            g.c.Delete(k)
        }
    }
    return nil
}
```

### Service Pattern — `UserService` Example

```go
type UserService struct {
    userRepo      *repository.UserRepository
    configService *ConfigService
    cache         cache.CacheStore
    group         singleflight.Group
}

func NewUserService(repo *repository.UserRepository, cfg *ConfigService, c cache.CacheStore) *UserService {
    return &UserService{userRepo: repo, configService: cfg, cache: c}
}

func (s *UserService) GetLeaderboard() ([]*model.User, error) {
    return cache.GetOrFetch(
        s.cache, &s.group,
        "esport:users:leaderboard",
        5*time.Minute,
        func() ([]*model.User, error) {
            return s.userRepo.GetLeaderboard()
        },
    )
}

func (s *UserService) CreateUser(req *CreateUserRequest) (*model.User, error) {
    user, err := s.userRepo.Create(req)
    if err != nil {
        return nil, err
    }
    // Invalidate after successful write
    _ = s.cache.Delete("esport:users:all")
    _ = s.cache.Delete("esport:users:leaderboard")
    return user, nil
}
```

### Matches Cache Key with Pagination

Because matches use pagination, the cache key includes page + limit + optional playerID:

```go
import "fmt"

func matchesCacheKey(page, limit int, playerID *uuid.UUID) string {
    pid := "all"
    if playerID != nil {
        pid = playerID.String()
    }
    return fmt.Sprintf("esport:matches:all:%d:%d:%s", page, limit, pid)
}
```

Invalidation uses pattern delete: `s.cache.DeleteByPattern("esport:matches:all:*")`

### Router Wiring (excerpt)

```go
// router.go

import "github.com/duyb/esport-score-tracker/internal/cache"

// ...inside SetupRouter():

var cacheStore cache.CacheStore
if redisURL := os.Getenv("REDIS_URL"); redisURL != "" {
    rc, err := cache.NewRedisCache(redisURL)
    if err != nil {
        log.Fatalf("Redis init failed: %v", err)
    }
    cacheStore = rc
    log.Println("Cache backend: Redis")
} else {
    cacheStore = cache.NewGoCacheStore(10*time.Minute, 5*time.Minute)
    log.Println("Cache backend: go-cache (dev)")
}

// Pass to services that use caching
userService  := service.NewUserService(userRepo, configService, cacheStore)
matchService := service.NewMatchService(matchRepo, userRepo, settlementService, configService, tierService, db, cacheStore)
// ...etc
```

## Integration Points

### Redis Key Namespace

All FC25 Tracker keys use two-level namespace: `{domain}:{entity}:{qualifier}`

- `esport:` — core esport tracker domain
- `wc:` — World Cup domain

This prevents accidental collisions if a shared Redis instance is used.

### go-redis vs go-cache API differences

| Operation | go-redis | go-cache |
|-----------|---------|---------|
| Get | `client.Get(ctx, key).Result()` → `(string, error)` | `c.Get(key)` → `(interface{}, bool)` |
| Set | `client.Set(ctx, key, val, ttl).Err()` | `c.Set(key, val, ttl)` |
| Delete | `client.Del(ctx, keys...).Err()` | `c.Delete(key)` |
| Pattern scan | `client.Scan(ctx, cursor, pattern, count).Result()` | Manual `Items()` iteration |

The `CacheStore` interface hides all of this — callers never see these differences.

## Error Handling

### Redis Errors → Cache Miss

```go
func (r *RedisCache) Get(key string) (string, bool) {
    val, err := r.client.Get(r.ctx, key).Result()
    if err != nil {
        // redis.Nil = key not found (normal miss)
        // other err = network/timeout error (treat as miss, don't surface)
        return "", false
    }
    return val, true
}
```

**Never let Redis errors fail an HTTP request.** The service will fall through to DB and the user gets their data normally.

### Failed Set After Fetch

If `Set` fails (e.g. Redis OOM, connection dropped), the fetch result is still returned to the caller. The next request will miss and refetch from DB — slightly more load but correct behavior.

```go
// In GetOrFetch:
if b, jsonErr := json.Marshal(result); jsonErr == nil {
    _ = store.Set(key, string(b), ttl)   // Intentional: ignore Set errors
}
return result, nil  // Always return the fetched value
```

## Performance Considerations

### Why JSON (not msgpack)?

- Human-readable: you can `redis-cli GET esport:users:leaderboard` and see the data during development
- No extra dependency
- Overhead: ~10–20% larger than msgpack at this data size (< 50 users, < 200 matches) — completely irrelevant

### singleflight Memory

`singleflight.Group` holds a map of in-flight calls. At this scale (< 10 concurrent users), the map holds at most a handful of entries at any time. No memory concern.

### `DeleteByPattern` Cost

`SCAN` with `COUNT 100` is O(keyspace_size / 100) iterations. With ~20 unique cache keys in total, this is effectively O(1). If keyspace grows to millions, revisit with versioned keys instead.

## Security Notes

- `REDIS_URL` must never be committed — add `*.env` to `.gitignore` (should already be there)
- Redis instance should not be publicly exposed — bind to localhost or use private network
- No sensitive user data (passwords, tokens) is stored in Redis — only aggregated query results
- JSON values in Redis are readable by anyone with Redis access — acceptable for this internal app
