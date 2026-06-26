---
phase: implementation
title: API Performance Optimization — Implementation Guide
description: Concrete code changes, patterns, and file-by-file notes for the performance work
---

# Implementation Guide

## Development Setup

No new infrastructure required. Dependencies to add:

```bash
cd backend
go get github.com/patrickmn/go-cache
go get github.com/gin-contrib/gzip
```

## Code Structure

New file:
```
backend/internal/cache/
  cache.go        ← CacheStore interface + GoCacheStore implementation
```

Modified files:
```
backend/internal/database/database.go
backend/internal/api/router.go
backend/internal/api/match_handler.go
backend/internal/repository/match_repository.go
backend/internal/service/wc_service.go
backend/internal/repository/wc_custom_bet_repository.go
backend/migrations/YYYYMMDD_add_performance_indexes.sql
```

## Implementation Notes

### 1. Cache Abstraction (`internal/cache/cache.go`)

```go
package cache

import (
    "time"
    gocache "github.com/patrickmn/go-cache"
)

type Store interface {
    Get(key string) (any, bool)
    Set(key string, value any, ttl time.Duration)
    Delete(key string)
    DeletePrefix(prefix string)
}

type GoCacheStore struct {
    c *gocache.Cache
}

func NewGoCacheStore(defaultTTL, cleanupInterval time.Duration) *GoCacheStore {
    return &GoCacheStore{c: gocache.New(defaultTTL, cleanupInterval)}
}

func (s *GoCacheStore) Get(key string) (any, bool) {
    return s.c.Get(key)
}

func (s *GoCacheStore) Set(key string, value any, ttl time.Duration) {
    s.c.Set(key, value, ttl)
}

func (s *GoCacheStore) Delete(key string) {
    s.c.Delete(key)
}

func (s *GoCacheStore) DeletePrefix(prefix string) {
    for k := range s.c.Items() {
        if len(k) >= len(prefix) && k[:len(prefix)] == prefix {
            s.c.Delete(k)
        }
    }
}
```

### 2. DB Connection Pool (`internal/database/database.go`)

After `gorm.Open(...)`, add:

```go
sqlDB, err := db.DB()
if err != nil {
    return nil, fmt.Errorf("get underlying sql.DB: %w", err)
}
sqlDB.SetMaxOpenConns(25)
sqlDB.SetMaxIdleConns(5)
sqlDB.SetConnMaxLifetime(5 * time.Minute)
sqlDB.SetConnMaxIdleTime(30 * time.Second)
```

Change logger level from `logger.Info` to `logger.Warn`:
```go
gorm.Open(postgres.Open(dsn), &gorm.Config{
    Logger: logger.Default.LogMode(logger.Warn),  // was logger.Info
})
```

### 3. Gzip Middleware (`internal/api/router.go`)

```go
import "github.com/gin-contrib/gzip"

// Add after cors.New(...)
router.Use(gzip.Gzip(gzip.DefaultCompression))
```

### 4. Fix Match Handler Pagination (`internal/api/match_handler.go`)

**Current anti-pattern** (lines ~106-145):
```go
matches, _ := matchService.GetAllMatches(0, 0)   // fetches ALL
bonuses, _ := bonusService.GetAll(0, 0)           // fetches ALL
feed := buildFeed(matches, bonuses)
total := len(feed)
feed = feed[start:end]                            // paginates in memory
```

**Fix pattern:**
```go
// 1. Get total count first
total, _ := matchService.CountAllMatches()

// 2. Fetch only the page of matches from DB
matches, _ := matchService.GetAllMatches(limit, offset)

// 3. Scope bonuses to the date range of returned matches only
var startDate, endDate time.Time
// derive from matches[0].PlayedAt and matches[len-1].PlayedAt
bonuses, _ := bonusService.GetByDateRange(startDate, endDate)

// 4. Build feed from scoped data
feed := buildFeed(matches, bonuses)
```

This requires adding `CountAllMatches()` to the service and ensuring `GetAllMatches(limit, offset)` is already DB-paginated (it is — `match_repository.go` line 35 applies LIMIT/OFFSET).

### 5. Leaderboard Cache (`internal/service/wc_service.go`)

```go
const leaderboardCacheTTL = 2 * time.Minute

func (s *WCService) GetLeaderboard() ([]LeaderboardEntry, error) {
    if v, ok := s.cache.Get("wc:leaderboard"); ok {
        return v.([]LeaderboardEntry), nil
    }
    entries, err := s.repo.GetLeaderboard()
    if err != nil {
        return nil, err
    }
    s.cache.Set("wc:leaderboard", entries, leaderboardCacheTTL)
    return entries, nil
}

// Call this after any settlement operation
func (s *WCService) invalidateLeaderboard() {
    s.cache.Delete("wc:leaderboard")
}
```

Call `s.invalidateLeaderboard()` at the end of: `SettleMatch`, `RecalculateAll`, `SettleCustomBet`, `SettleChampionPrediction`, `AdminSettleAll`.

### 6. WC Matches Cache

```go
const wcMatchesCacheTTL = 5 * time.Minute

func (s *WCService) GetAllMatches() ([]WCMatch, error) {
    if v, ok := s.cache.Get("wc:matches:all"); ok {
        return v.([]WCMatch), nil
    }
    matches, err := s.repo.GetAllMatches()
    if err != nil {
        return nil, err
    }
    s.cache.Set("wc:matches:all", matches, wcMatchesCacheTTL)
    return matches, nil
}
```

Invalidate on `SyncMatches` and any admin match update.

### 7. Batch Insert Custom Bet Options

In `wc_custom_bet_repository.go`, replace the loop:

```go
// Before:
for _, opt := range bet.Options {
    if err := tx.Create(&opt).Error; err != nil {
        return err
    }
}

// After:
if err := tx.Create(&bet.Options).Error; err != nil {
    return err
}
```

### 8. DB Index Migration

```sql
-- backend/migrations/20260626_add_performance_indexes.sql

CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_match_participants_match_id
    ON match_participants(match_id);

CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_wc_bets_result
    ON wc_bets(result);

CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_wc_predictions_result
    ON wc_predictions(result);

CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_wc_chat_messages_created_at
    ON wc_chat_messages(created_at DESC);
```

Note: `CONCURRENTLY` avoids table lock on live DB. Not supported inside transactions.

## Integration Points

- `CacheStore` is injected into `WCService` via constructor: `NewWCService(repo, cache)`
- `router.go` instantiates `cache.NewGoCacheStore(5*time.Minute, 10*time.Minute)` and passes it when constructing `WCService`
- No changes to repository layer interfaces — cache lives in service layer only

## Error Handling

- Cache miss → always fall through to DB, never return an error for cache miss
- Cache store itself never returns errors (go-cache panics on nil store, not on miss)
- If DB fails on cache miss, return error normally — do NOT cache error responses

## Performance Considerations

- **go-cache GC**: Items expire lazily; cleanup goroutine runs every 10 minutes
- **Memory estimate**: Leaderboard ~50 entries × ~200 bytes = ~10KB per cached item — negligible
- **Pool sizing**: MaxOpenConns=25 is conservative for a PostgreSQL instance; adjust up if needed
- **`CONCURRENTLY` index**: Runs without table lock but takes longer; safe for production

## Security Notes

- Cache keys are internal strings — no user input in cache keys (no cache poisoning risk)
- Cache holds non-sensitive aggregate data only (leaderboard rankings, match data)
- No auth tokens or personal data cached
