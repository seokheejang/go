package redis

import (
	"context"
	"encoding/json"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/seokheejang/go/cache-layer/pkg/cache"
)

type redisCache struct {
	client  *redis.Client
	options *cache.Options
}

// New creates a new Redis cache instance
func New(client *redis.Client, options *cache.Options) (cache.Cache, error) {
	if options == nil {
		options = &cache.Options{
			DefaultTTL: 2 * time.Second,
			MaxTTL:     5 * time.Second,
			MaxSize:    0, // Redis handles size limits internally
		}
	}

	return &redisCache{
		client:  client,
		options: options,
	}, nil
}

func (c *redisCache) Get(ctx context.Context, key string) (interface{}, error) {
	data, err := c.client.Get(ctx, key).Bytes()
	if err == redis.Nil {
		return nil, cache.ErrNotFound
	}
	if err != nil {
		return nil, err
	}

	var value interface{}
	if err := json.Unmarshal(data, &value); err != nil {
		return nil, err
	}

	return value, nil
}

func (c *redisCache) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	if ttl == 0 {
		ttl = c.options.DefaultTTL
	}
	if ttl > c.options.MaxTTL {
		ttl = c.options.MaxTTL
	}

	data, err := json.Marshal(value)
	if err != nil {
		return err
	}

	return c.client.Set(ctx, key, data, ttl).Err()
}

func (c *redisCache) Delete(ctx context.Context, key string) error {
	return c.client.Del(ctx, key).Err()
}

func (c *redisCache) Clear(ctx context.Context) error {
	return c.client.FlushDB(ctx).Err()
}

func (c *redisCache) Close() error {
	return c.client.Close()
}
