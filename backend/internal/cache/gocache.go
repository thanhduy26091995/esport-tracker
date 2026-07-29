package cache

import (
	"sync"
	"sync/atomic"
	"time"

	gocachepkg "github.com/patrickmn/go-cache"
)

// GoCacheStore is the dev/fallback CacheStore backed by in-process go-cache.
// Used when REDIS_URL is unset or Redis is unreachable at startup.
type GoCacheStore struct {
	c        *gocachepkg.Cache
	counters sync.Map // key → *atomic.Int64 for GetInt/Incr
}

// NewGoCacheStore creates a GoCacheStore with the given default TTL and cleanup interval.
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

// DeleteByPattern deletes all keys whose prefix matches pattern (strips trailing '*').
func (g *GoCacheStore) DeleteByPattern(pattern string) error {
	prefix := pattern
	if len(prefix) > 0 && prefix[len(prefix)-1] == '*' {
		prefix = prefix[:len(prefix)-1]
	}
	for k := range g.c.Items() {
		if len(prefix) == 0 || (len(k) >= len(prefix) && k[:len(prefix)] == prefix) {
			g.c.Delete(k)
		}
	}
	return nil
}

func (g *GoCacheStore) GetInt(key string) (int64, error) {
	if v, loaded := g.counters.Load(key); loaded {
		return v.(*atomic.Int64).Load(), nil
	}
	return 0, nil
}

func (g *GoCacheStore) Incr(key string) (int64, error) {
	v, _ := g.counters.LoadOrStore(key, &atomic.Int64{})
	return v.(*atomic.Int64).Add(1), nil
}
