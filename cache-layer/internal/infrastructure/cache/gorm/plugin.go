package gorm

import (
	"context"

	"github.com/go-gorm/caches/v4"
	"github.com/seokheejang/go/cache-layer/internal/domains/cache"
	"gorm.io/gorm"
)

// WithGormCache applies the cache plugin to the GORM DB instance
func WithGormCache(db *gorm.DB, cacheService cache.Service) error {
	// Create cache plugin with configuration
	cachePlugin := &caches.Caches{
		Conf: &caches.Config{
			Cacher: &gormCacher{
				cache: cacheService,
			},
			Easer: false,
		},
	}

	// Use the cache plugin
	return db.Use(cachePlugin)
}

// gormCacher implements the caches.Cacher interface
type gormCacher struct {
	cache cache.Service
}

// Get retrieves a value from the cache
func (c *gormCacher) Get(ctx context.Context, key string, q *caches.Query[any]) (*caches.Query[any], error) {
	if q == nil {
		q = &caches.Query[any]{}
	}
	return c.cache.Get(ctx, key, q)
}

// Store stores a value in the cache
func (c *gormCacher) Store(ctx context.Context, key string, val *caches.Query[any]) error {
	if val == nil {
		return nil
	}
	return c.cache.Store(ctx, key, val)
}

// Invalidate invalidates the cache
func (c *gormCacher) Invalidate(ctx context.Context) error {
	return c.cache.Invalidate(ctx)
}
