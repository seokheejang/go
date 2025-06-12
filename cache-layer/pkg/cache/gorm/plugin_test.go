package gorm

import (
	"context"
	"testing"
	"time"

	"github.com/go-gorm/caches/v4"
	"github.com/seokheejang/go/cache-layer/pkg/cache"
	"github.com/seokheejang/go/cache-layer/pkg/cache/memory"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// TestModel is a simple model for testing
type TestModel struct {
	ID   uint `gorm:"primarykey"`
	Name string
}

func setupTestDB(t *testing.T) (*gorm.DB, *gormCacher) {
	// Create in-memory SQLite database
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	assert.NoError(t, err)

	// Create test table
	err = db.AutoMigrate(&TestModel{})
	assert.NoError(t, err)

	// Create memory cache
	memCache := memory.NewGoCacheWrapper(&cache.Options{
		DefaultTTL: 100 * time.Millisecond,
	})

	// Create gormCacher
	cacher := &gormCacher{
		cache: memCache,
	}

	return db, cacher
}

func TestWithGormCache(t *testing.T) {
	db, cacher := setupTestDB(t)
	ctx := context.Background()

	// Apply cache plugin
	err := WithGormCache(db, cacher.cache)
	assert.NoError(t, err)

	t.Run("Cache Store and Get Test", func(t *testing.T) {
		// Given
		model := &TestModel{Name: "test"}
		query := &caches.Query[any]{
			Dest: model,
		}
		key := "test-key"

		// When - Store in cache
		err := cacher.Store(ctx, key, query)
		assert.NoError(t, err)

		// Then - Get from cache
		resultQuery := &caches.Query[any]{
			Dest: &TestModel{}, // Initialize with a new pointer
		}
		result, err := cacher.Get(ctx, key, resultQuery)
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.NotNil(t, result.Dest)

		// Type assertion and value check
		resultModel, ok := result.Dest.(*TestModel)
		assert.True(t, ok, "Failed to type assert result to *TestModel")
		assert.Equal(t, model.Name, resultModel.Name)
	})

	t.Run("Cache Invalidation Test", func(t *testing.T) {
		// Given
		model := &TestModel{Name: "test"}
		query := &caches.Query[any]{
			Dest: model,
		}
		key := "test-key"

		// When - Store in cache
		err := cacher.Store(ctx, key, query)
		assert.NoError(t, err)

		// When - Invalidate cache
		err = cacher.Invalidate(ctx)
		assert.NoError(t, err)

		// Then - Get should return nil
		resultQuery := &caches.Query[any]{
			Dest: &TestModel{}, // Initialize with a new pointer
		}
		result, err := cacher.Get(ctx, key, resultQuery)
		assert.NoError(t, err)
		assert.Nil(t, result)
	})

	t.Run("Nil Query Test", func(t *testing.T) {
		// Given
		key := "nil-test-key"

		// When - Store nil query
		err := cacher.Store(ctx, key, nil)
		assert.NoError(t, err)

		// Then - Get should return nil
		resultQuery := &caches.Query[any]{
			Dest: &TestModel{}, // Initialize with a new pointer
		}
		result, err := cacher.Get(ctx, key, resultQuery)
		assert.NoError(t, err)
		assert.Nil(t, result)
	})

	t.Run("Invalid Destination Test", func(t *testing.T) {
		// Given
		model := TestModel{Name: "test"} // Not a pointer
		query := &caches.Query[any]{
			Dest: model,
		}
		key := "invalid-dest-key"

		// When - Store in cache
		err := cacher.Store(ctx, key, query)
		assert.NoError(t, err)

		// Then - Get should return error
		resultQuery := &caches.Query[any]{
			Dest: TestModel{}, // Not a pointer
		}
		result, err := cacher.Get(ctx, key, resultQuery)
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "must be a non-nil pointer")
	})

	t.Run("Database Integration Test", func(t *testing.T) {
		// Given
		model := &TestModel{Name: "test-db"}

		// When - Create record
		err := db.Create(model).Error
		assert.NoError(t, err)

		// When - Query with cache
		var result TestModel
		err = db.First(&result, model.ID).Error
		assert.NoError(t, err)

		// Then - Verify result
		assert.Equal(t, model.Name, result.Name)

		// When - Query again (should use cache)
		var cachedResult TestModel
		err = db.First(&cachedResult, model.ID).Error
		assert.NoError(t, err)

		// Then - Verify cached result
		assert.Equal(t, model.Name, cachedResult.Name)
	})
}
