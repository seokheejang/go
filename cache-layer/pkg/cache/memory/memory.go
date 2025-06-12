package memory

import (
	"context"
	"sync"
	"time"

	"github.com/seokheejang/go/cache-layer/pkg/cache"
)

type memoryCache struct {
	mu      sync.RWMutex
	entries map[string]*cache.Entry
	options *cache.Options
}

// New creates a new memory cache instance
func New(options *cache.Options) (cache.Cache, error) {
	if options == nil {
		options = &cache.Options{
			DefaultTTL: 2 * time.Second,
			MaxTTL:     5 * time.Second,
			MaxSize:    1000,
		}
	}

	return &memoryCache{
		entries: make(map[string]*cache.Entry),
		options: options,
	}, nil
}

func (c *memoryCache) Get(ctx context.Context, key string) (interface{}, error) {
	c.mu.RLock()
	entry, exists := c.entries[key]
	c.mu.RUnlock()

	if !exists {
		return nil, nil
	}

	// Check if the entry has expired
	if entry.IsExpired() {
		c.mu.Lock()
		delete(c.entries, key)
		c.mu.Unlock()
		return nil, nil
	}

	return entry.Value, nil
}

func (c *memoryCache) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	if ttl == 0 {
		ttl = c.options.DefaultTTL
	}
	if ttl > c.options.MaxTTL {
		ttl = c.options.MaxTTL
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	// Check if we need to evict entries
	if int64(len(c.entries)) >= c.options.MaxSize {
		c.evictOldest()
	}

	c.entries[key] = cache.NewEntry(key, value, ttl)
	return nil
}

func (c *memoryCache) Delete(ctx context.Context, key string) error {
	c.mu.Lock()
	delete(c.entries, key)
	c.mu.Unlock()
	return nil
}

func (c *memoryCache) Clear(ctx context.Context) error {
	c.mu.Lock()
	c.entries = make(map[string]*cache.Entry)
	c.mu.Unlock()
	return nil
}

func (c *memoryCache) Close() error {
	c.Clear(context.Background())
	return nil
}

// evictOldest removes the oldest entry from the cache
func (c *memoryCache) evictOldest() {
	var oldestKey string
	var oldestTime time.Time

	for key, entry := range c.entries {
		if oldestKey == "" || entry.Created.Before(oldestTime) {
			oldestKey = key
			oldestTime = entry.Created
		}
	}

	if oldestKey != "" {
		delete(c.entries, oldestKey)
	}
}
