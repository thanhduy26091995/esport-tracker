package cache_test

import (
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/duyb/esport-score-tracker/internal/cache"
	"golang.org/x/sync/singleflight"
)

// --- RedisCache tests (Task 4.1) ---

func newTestRedisCache(t *testing.T) (*cache.RedisCache, *miniredis.Miniredis) {
	t.Helper()
	mr := miniredis.RunT(t)
	c, err := cache.NewRedisCache("redis://" + mr.Addr())
	if err != nil {
		t.Fatalf("NewRedisCache: %v", err)
	}
	return c, mr
}

func TestRedisCache_GetSet(t *testing.T) {
	c, _ := newTestRedisCache(t)

	if _, ok := c.Get("missing"); ok {
		t.Fatal("expected miss on empty cache")
	}

	if err := c.Set("k", "hello", time.Minute); err != nil {
		t.Fatalf("Set: %v", err)
	}
	val, ok := c.Get("k")
	if !ok || val != "hello" {
		t.Fatalf("want 'hello', got %q ok=%v", val, ok)
	}
}

func TestRedisCache_Delete(t *testing.T) {
	c, _ := newTestRedisCache(t)

	_ = c.Set("k", "v", time.Minute)
	_ = c.Delete("k")

	if _, ok := c.Get("k"); ok {
		t.Fatal("expected miss after delete")
	}
}

func TestRedisCache_TTLExpiry(t *testing.T) {
	c, mr := newTestRedisCache(t)

	_ = c.Set("k", "v", 2*time.Second)
	mr.FastForward(3 * time.Second)

	if _, ok := c.Get("k"); ok {
		t.Fatal("expected miss after TTL expiry")
	}
}

func TestRedisCache_DeleteByPattern(t *testing.T) {
	c, _ := newTestRedisCache(t)

	_ = c.Set("prefix:a", "1", time.Minute)
	_ = c.Set("prefix:b", "2", time.Minute)
	_ = c.Set("other:c", "3", time.Minute)

	if err := c.DeleteByPattern("prefix:*"); err != nil {
		t.Fatalf("DeleteByPattern: %v", err)
	}

	if _, ok := c.Get("prefix:a"); ok {
		t.Fatal("prefix:a should be deleted")
	}
	if _, ok := c.Get("prefix:b"); ok {
		t.Fatal("prefix:b should be deleted")
	}
	if _, ok := c.Get("other:c"); !ok {
		t.Fatal("other:c should survive pattern delete")
	}
}

func TestRedisCache_IncrAndGetInt(t *testing.T) {
	c, _ := newTestRedisCache(t)

	n, err := c.GetInt("counter")
	if err != nil || n != 0 {
		t.Fatalf("GetInt on missing key: want 0 nil, got %d %v", n, err)
	}

	v1, _ := c.Incr("counter")
	v2, _ := c.Incr("counter")
	v3, _ := c.Incr("counter")

	if v1 != 1 || v2 != 2 || v3 != 3 {
		t.Fatalf("Incr sequence: want 1,2,3 got %d,%d,%d", v1, v2, v3)
	}

	got, _ := c.GetInt("counter")
	if got != 3 {
		t.Fatalf("GetInt after Incr: want 3 got %d", got)
	}
}

// --- GetOrFetch tests (Task 4.2) ---

func TestGetOrFetch_CacheHit(t *testing.T) {
	c, _ := newTestRedisCache(t)
	var group singleflight.Group

	// Pre-populate cache with a JSON string.
	_ = c.Set("key", `"cached"`, time.Minute)

	calls := 0
	val, err := cache.GetOrFetch(c, &group, "key", time.Minute, func() (string, error) {
		calls++
		return "from-db", nil
	})
	if err != nil || val != "cached" {
		t.Fatalf("want 'cached', got %q %v", val, err)
	}
	if calls != 0 {
		t.Fatal("fetch fn should not be called on cache hit")
	}
}

func TestGetOrFetch_CacheMissPopulates(t *testing.T) {
	c, _ := newTestRedisCache(t)
	var group singleflight.Group

	calls := 0
	val, err := cache.GetOrFetch(c, &group, "key", time.Minute, func() (string, error) {
		calls++
		return "from-db", nil
	})
	if err != nil || val != "from-db" {
		t.Fatalf("want 'from-db', got %q %v", val, err)
	}
	if calls != 1 {
		t.Fatalf("fetch fn should be called exactly once, got %d", calls)
	}

	// Second call should hit cache.
	val2, _ := cache.GetOrFetch(c, &group, "key", time.Minute, func() (string, error) {
		calls++
		return "from-db-again", nil
	})
	if val2 != "from-db" || calls != 1 {
		t.Fatalf("second call: want cached 'from-db', got %q; calls=%d", val2, calls)
	}
}

func TestGetOrFetch_SingleflightDeduplication(t *testing.T) {
	c, _ := newTestRedisCache(t)
	var group singleflight.Group

	var dbCalls atomic.Int32
	var wg sync.WaitGroup
	results := make([]string, 20)
	errs := make([]error, 20)

	// Simulate 20 concurrent goroutines all missing the same cache key.
	for i := range results {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			results[idx], errs[idx] = cache.GetOrFetch(c, &group, "shared-key", time.Minute, func() (string, error) {
				dbCalls.Add(1)
				time.Sleep(20 * time.Millisecond) // simulate DB latency
				return fmt.Sprintf("result-%d", dbCalls.Load()), nil
			})
		}(i)
	}
	wg.Wait()

	for i, err := range errs {
		if err != nil {
			t.Fatalf("goroutine %d error: %v", i, err)
		}
	}

	// singleflight collapses all concurrent callers into one DB call.
	if dbCalls.Load() != 1 {
		t.Fatalf("expected exactly 1 DB call via singleflight, got %d", dbCalls.Load())
	}

	// All goroutines must have received the same result.
	for i, r := range results {
		if r != results[0] {
			t.Fatalf("goroutine %d got %q, expected %q", i, r, results[0])
		}
	}
}

func TestGetOrFetch_FetchError(t *testing.T) {
	c, _ := newTestRedisCache(t)
	var group singleflight.Group

	_, err := cache.GetOrFetch(c, &group, "key", time.Minute, func() (string, error) {
		return "", fmt.Errorf("db down")
	})
	if err == nil || err.Error() != "db down" {
		t.Fatalf("expected 'db down' error, got %v", err)
	}

	// Cache must remain empty after a fetch error.
	if _, ok := c.Get("key"); ok {
		t.Fatal("cache must not be populated after fetch error")
	}
}
