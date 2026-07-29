package cache

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

// RedisCache is the production CacheStore backed by a Redis server.
type RedisCache struct {
	client *redis.Client
	ctx    context.Context
}

// NewRedisCache parses redisURL, connects, and pings Redis.
// Returns an error if the URL is invalid or Redis is unreachable — caller decides fallback.
func NewRedisCache(redisURL string) (*RedisCache, error) {
	opt, err := redis.ParseURL(redisURL)
	if err != nil {
		return nil, err
	}
	client := redis.NewClient(opt)
	ctx := context.Background()
	if err := client.Ping(ctx).Err(); err != nil {
		client.Close()
		return nil, err
	}
	return &RedisCache{client: client, ctx: ctx}, nil
}

func (r *RedisCache) Get(key string) (string, bool) {
	val, err := r.client.Get(r.ctx, key).Result()
	if err != nil {
		// redis.Nil = normal miss; any other error = treat as miss (never surface to callers)
		return "", false
	}
	return val, true
}

func (r *RedisCache) Set(key, value string, ttl time.Duration) error {
	return r.client.Set(r.ctx, key, value, ttl).Err()
}

func (r *RedisCache) Delete(key string) error {
	return r.client.Del(r.ctx, key).Err()
}

// DeleteByPattern removes all keys matching a Redis glob pattern using cursor-based SCAN.
// Uses SCAN (not KEYS) to avoid blocking Redis on large keysets.
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

func (r *RedisCache) GetInt(key string) (int64, error) {
	val, err := r.client.Get(r.ctx, key).Int64()
	if err == redis.Nil {
		return 0, nil
	}
	return val, err
}

// Incr atomically increments a counter key. Creates it at 1 if absent. No TTL is set.
func (r *RedisCache) Incr(key string) (int64, error) {
	return r.client.Incr(r.ctx, key).Result()
}
