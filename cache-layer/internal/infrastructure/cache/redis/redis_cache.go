package redis

import (
	"context"
	"time"

	"github.com/go-gorm/caches/v4"
	"github.com/redis/go-redis/v9"
	"github.com/seokheejang/go/cache-layer/internal/domains/cache"
)

const (
	defaultTTL = 2 * time.Second
)

type redisCache struct {
	rdb *redis.Client
	ttl time.Duration
}

// NewRedisCache creates a new Redis cache with optional TTL.
// If ttl is not provided (nil), defaultTTL will be used.
func NewRedisCache(rdb *redis.Client, ttl *time.Duration) cache.Service {
	duration := defaultTTL
	if ttl != nil {
		duration = *ttl
	}

	return &redisCache{
		rdb: rdb,
		ttl: duration,
	}
}

func (c *redisCache) Get(ctx context.Context, key string, q *caches.Query[any]) (*caches.Query[any], error) {
	res, err := c.rdb.Get(ctx, key).Bytes()
	if err == redis.Nil {
		return nil, nil
	}

	if err != nil {
		return nil, nil
	}

	if err := q.Unmarshal(res); err != nil {
		return nil, err
	}

	return q, nil
}

func (c *redisCache) Store(ctx context.Context, key string, val *caches.Query[any]) error {
	res, err := val.Marshal()
	if err != nil {
		return err
	}

	return c.rdb.Set(ctx, key, res, c.ttl).Err()
}

func (c *redisCache) Invalidate(ctx context.Context) error {
	return c.rdb.FlushDB(ctx).Err()
}

func (c *redisCache) Delete(ctx context.Context, key string) error {
	return c.rdb.Del(ctx, key).Err()
}

func (c *redisCache) Close() {}
