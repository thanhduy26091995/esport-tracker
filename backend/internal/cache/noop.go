package cache

import "time"

// NoopCache is a CacheStore that always misses on Get and ignores all writes.
// Used in unit tests that don't need cache behavior.
type NoopCache struct{}

func (n *NoopCache) Get(_ string) (string, bool)              { return "", false }
func (n *NoopCache) Set(_ string, _ string, _ time.Duration) error { return nil }
func (n *NoopCache) Delete(_ string) error                    { return nil }
func (n *NoopCache) DeleteByPattern(_ string) error           { return nil }
func (n *NoopCache) GetInt(_ string) (int64, error)           { return 0, nil }
func (n *NoopCache) Incr(_ string) (int64, error)             { return 1, nil }
