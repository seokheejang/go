package memory

import (
	"context"
	"time"

	gocache "github.com/patrickmn/go-cache"
	"github.com/seokheejang/go/cache-layer/pkg/cache"
)

type goCacheWrapper struct {
	cache *gocache.Cache
}

func NewGoCacheWrapper(options *cache.Options) cache.Cache {
	defaultTTL := options.DefaultTTL
	cleanupInterval := 5 * time.Second // static janitor interval

	return &goCacheWrapper{
		cache: gocache.New(defaultTTL, cleanupInterval),
	}
}

func (c *goCacheWrapper) Get(_ context.Context, key string) (interface{}, error) {
	val, found := c.cache.Get(key)
	if !found {
		return nil, cache.ErrNotFound
	}

	return val, nil
}

func (c *goCacheWrapper) Set(_ context.Context, key string, value interface{}, ttl time.Duration) error {
	if ttl == 0 {
		ttl = gocache.DefaultExpiration
	}

	c.cache.Set(key, value, ttl)

	return nil
}

func (c *goCacheWrapper) Delete(_ context.Context, key string) error {
	c.cache.Delete(key)

	return nil
}

func (c *goCacheWrapper) Clear(_ context.Context) error {
	c.cache.Flush()

	return nil
}

func (c *goCacheWrapper) Close() error {
	// go-cache doesn't require Close()
	return nil
}
