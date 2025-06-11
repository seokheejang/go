package cache

import (
	"context"

	"github.com/go-gorm/caches/v4"
)

type Repository interface {
	Get(ctx context.Context, key string, query *caches.Query[any]) (*caches.Query[any], error)
	Store(ctx context.Context, key string, val *caches.Query[any]) error
	Delete(ctx context.Context, key string) error
	Invalidate(ctx context.Context) error
	Close()
}
