package cache

import (
	"encoding/json"
	"time"

	"golang.org/x/sync/singleflight"
)

// CacheStore is the abstraction over any cache backend (Redis, go-cache, noop).
type CacheStore interface {
	Get(key string) (string, bool)
	Set(key string, value string, ttl time.Duration) error
	Delete(key string) error
	// DeleteByPattern removes all keys matching a glob pattern. General-purpose utility.
	// Uses SCAN+DEL on Redis; prefix scan on go-cache.
	DeleteByPattern(pattern string) error
	// GetInt retrieves an integer counter. Returns 0 on miss (no error).
	GetInt(key string) (int64, error)
	// Incr atomically increments an integer counter, creating it at 1 if absent. No TTL.
	Incr(key string) (int64, error)
}

// GetOrFetch implements Cache-Aside with singleflight stampede prevention.
// On cache hit: returns cached value immediately.
// On cache miss: one goroutine fetches from DB (via fetch fn); concurrent misses
// for the same key wait and share the single result — no duplicate DB queries.
func GetOrFetch[T any](
	store CacheStore,
	group *singleflight.Group,
	key string,
	ttl time.Duration,
	fetch func() (T, error),
) (T, error) {
	// 1. Cache hit path
	if raw, ok := store.Get(key); ok {
		var result T
		if err := json.Unmarshal([]byte(raw), &result); err == nil {
			return result, nil
		}
		// Corrupt entry — fall through to fetch
	}

	// 2. Cache miss — singleflight deduplicates concurrent goroutines on same key
	val, err, _ := group.Do(key, func() (interface{}, error) {
		result, err := fetch()
		if err != nil {
			return nil, err
		}
		// 3. Populate cache; ignore Set errors — cache is best-effort
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
