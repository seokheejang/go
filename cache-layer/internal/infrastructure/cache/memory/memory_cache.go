// Package memory implements an in-process cache with TTL and automatic cleanup functionality.
// The janitor goroutine removes expired keys at specified purge intervals.
package memory

import (
	"context"
	"sync"
	"time"

	"github.com/go-gorm/caches/v4"
	"github.com/seokheejang/go/cache-layer/internal/domains/cache"
)

// cacheItem represents a cached value with its expiration time
type cacheItem struct {
	data      []byte
	expiresAt time.Time
}

// memoryCache implements cache.Service interface
type memoryCache struct {
	store         *sync.Map
	ttl           time.Duration // TTL for each key
	purgeInterval time.Duration // Interval for periodic cleanup
	stopJanitor   chan struct{} // Signal to stop janitor
	once          sync.Once
}

// NewInMemoryCache creates a new in-memory cache instance
// ttl: Time-to-live for cache entries (0 means no expiration)
// purgeInterval: Interval for periodic cleanup (defaults to 1 minute if 0)
func NewInMemoryCache(ttl, purgeInterval time.Duration) cache.Service {
	if purgeInterval == 0 {
		purgeInterval = time.Minute
	}

	mc := &memoryCache{
		store:         &sync.Map{},
		ttl:           ttl,
		purgeInterval: purgeInterval,
		stopJanitor:   make(chan struct{}),
	}

	go mc.janitor()

	return mc
}

// janitor periodically removes expired keys from the cache
func (c *memoryCache) janitor() {
	ticker := time.NewTicker(c.purgeInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			now := time.Now()
			c.store.Range(func(k, v any) bool {
				item := v.(cacheItem)
				if c.ttl > 0 && now.After(item.expiresAt) {
					c.store.Delete(k)
				}
				return true
			})
		case <-c.stopJanitor:
			return
		}
	}
}

// Close stops the background janitor goroutine
func (c *memoryCache) Close() { c.once.Do(func() { close(c.stopJanitor) }) }

// Get retrieves a value from the cache
func (c *memoryCache) Get(ctx context.Context, key string, q *caches.Query[any]) (*caches.Query[any], error) {
	if q == nil {
		q = &caches.Query[any]{}
	}

	val, ok := c.store.Load(key)
	if !ok {
		return nil, nil
	}

	item := val.(cacheItem)
	if c.ttl > 0 && time.Now().After(item.expiresAt) {
		c.store.Delete(key)
		return nil, nil
	}

	if err := q.Unmarshal(item.data); err != nil {
		return nil, err
	}
	return q, nil
}

// Store saves a value to the cache
func (c *memoryCache) Store(ctx context.Context, key string, val *caches.Query[any]) error {
	if val == nil {
		return nil
	}
	data, err := val.Marshal()
	if err != nil {
		return err
	}

	expires := time.Time{}
	if c.ttl > 0 {
		expires = time.Now().Add(c.ttl)
	}
	c.store.Store(key, cacheItem{data: data, expiresAt: expires})
	return nil
}

// Delete removes a single key from the cache
func (c *memoryCache) Delete(ctx context.Context, key string) error {
	c.store.Delete(key)
	return nil
}

// Invalidate clears all cache entries
func (c *memoryCache) Invalidate(ctx context.Context) error {
	c.store = &sync.Map{}
	return nil
}
