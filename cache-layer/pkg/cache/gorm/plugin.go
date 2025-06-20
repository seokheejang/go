package gorm

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"

	"github.com/go-gorm/caches/v4"
	"github.com/seokheejang/go/cache-layer/pkg/cache"
	"gorm.io/gorm"
)

// WithGormCache applies the cache plugin to the GORM DB instance
func WithGormCache(db *gorm.DB, cache cache.Cache) error {
	// Create cache plugin with configuration
	cachePlugin := &caches.Caches{
		Conf: &caches.Config{
			Cacher: &gormCacher{
				cache: cache,
			},
			Easer: false,
		},
	}

	// Use the cache plugin
	return db.Use(cachePlugin)
}

// gormCacher implements the caches.Cacher interface
type gormCacher struct {
	cache cache.Cache
}

// Get retrieves a value from the cache
func (c *gormCacher) Get(ctx context.Context, key string, q *caches.Query[any]) (*caches.Query[any], error) {
	if q == nil {
		q = &caches.Query[any]{}
	}

	value, err := c.cache.Get(ctx, key)

	if err == cache.ErrNotFound {
		return nil, nil // cache miss
	}

	if err != nil {
		return nil, err
	}

	if value == nil {
		return nil, nil
	}

	// Convert the cached value back to Query
	data, err := json.Marshal(value)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal cached value: %w", err)
	}

	if err := json.Unmarshal(data, q); err != nil {
		return nil, fmt.Errorf("failed to unmarshal cached value: %w", err)
	}

	if q.Dest == nil || reflect.ValueOf(q.Dest).Kind() != reflect.Ptr {
		return nil, fmt.Errorf("q.Dest must be a non-nil pointer for cache hydration")
	}

	return q, nil
}

// Store stores a value in the cache
func (c *gormCacher) Store(ctx context.Context, key string, val *caches.Query[any]) error {
	if val == nil {
		return nil
	}

	// Convert Query to a storable value
	data, err := json.Marshal(val)
	if err != nil {
		return fmt.Errorf("failed to marshal query: %w", err)
	}

	var value interface{}
	if err := json.Unmarshal(data, &value); err != nil {
		return fmt.Errorf("failed to unmarshal query: %w", err)
	}

	return c.cache.Set(ctx, key, value, 0) // TTL is managed by the cache implementation
}

// Invalidate invalidates the cache
func (c *gormCacher) Invalidate(ctx context.Context) error {
	return c.cache.Clear(ctx)
}
