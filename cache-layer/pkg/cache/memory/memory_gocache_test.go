package memory

import (
	"context"
	"testing"
	"time"

	"github.com/seokheejang/go/cache-layer/pkg/cache"
	"github.com/stretchr/testify/assert"
)

func TestGoCacheWrapper(t *testing.T) {
	// Set test options
	options := &cache.Options{
		DefaultTTL: 100 * time.Millisecond,
	}

	// Create cache instance
	cache := NewGoCacheWrapper(options)
	ctx := context.Background()

	t.Run("Basic Set/Get Operation Test", func(t *testing.T) {
		// Given
		key := "test-key"
		value := "test-value"

		// When
		err := cache.Set(ctx, key, value, 0)

		// Then
		assert.NoError(t, err)

		// When
		result, err := cache.Get(ctx, key)

		// Then
		assert.NoError(t, err)
		assert.Equal(t, value, result)
	})

	t.Run("TTL Expiration Test", func(t *testing.T) {
		// Given
		key := "ttl-test-key"
		value := "ttl-test-value"
		ttl := 50 * time.Millisecond

		// When
		err := cache.Set(ctx, key, value, ttl)
		assert.NoError(t, err)

		// Then - Value should exist before TTL expiration
		result, err := cache.Get(ctx, key)
		assert.NoError(t, err)
		assert.Equal(t, value, result)

		// When - Wait for TTL expiration
		time.Sleep(ttl + 10*time.Millisecond)

		// Then - Value should be nil after TTL expiration
		result, err = cache.Get(ctx, key)
		assert.NoError(t, err)
		assert.Nil(t, result)
	})

	t.Run("Cleanup Test", func(t *testing.T) {
		// Given
		key1 := "cleanup-test-key1"
		key2 := "cleanup-test-key2"
		value := "cleanup-test-value"
		shortTTL := 50 * time.Millisecond
		longTTL := 200 * time.Millisecond

		// When
		err := cache.Set(ctx, key1, value, shortTTL)
		assert.NoError(t, err)
		err = cache.Set(ctx, key2, value, longTTL)
		assert.NoError(t, err)

		// Then - Both values should exist initially
		result1, err := cache.Get(ctx, key1)
		assert.NoError(t, err)
		assert.Equal(t, value, result1)
		result2, err := cache.Get(ctx, key2)
		assert.NoError(t, err)
		assert.Equal(t, value, result2)

		// When - Wait for shortTTL expiration
		time.Sleep(shortTTL + 10*time.Millisecond)

		// Then - key1 should be expired, but key2 should still exist
		result1, err = cache.Get(ctx, key1)
		assert.NoError(t, err)
		assert.Nil(t, result1)
		result2, err = cache.Get(ctx, key2)
		assert.NoError(t, err)
		assert.Equal(t, value, result2)
	})

	t.Run("Delete Test", func(t *testing.T) {
		// Given
		key := "delete-test-key"
		value := "delete-test-value"

		// When
		err := cache.Set(ctx, key, value, 0)
		assert.NoError(t, err)

		// Then - Value should exist before deletion
		result, err := cache.Get(ctx, key)
		assert.NoError(t, err)
		assert.Equal(t, value, result)

		// When
		err = cache.Delete(ctx, key)
		assert.NoError(t, err)

		// Then - Value should be nil after deletion
		result, err = cache.Get(ctx, key)
		assert.NoError(t, err)
		assert.Nil(t, result)
	})

	t.Run("Clear Test", func(t *testing.T) {
		// Given
		key1 := "clear-test-key1"
		key2 := "clear-test-key2"
		value := "clear-test-value"

		// When
		err := cache.Set(ctx, key1, value, 0)
		assert.NoError(t, err)
		err = cache.Set(ctx, key2, value, 0)
		assert.NoError(t, err)

		// Then - Both values should exist initially
		result1, err := cache.Get(ctx, key1)
		assert.NoError(t, err)
		assert.Equal(t, value, result1)
		result2, err := cache.Get(ctx, key2)
		assert.NoError(t, err)
		assert.Equal(t, value, result2)

		// When
		err = cache.Clear(ctx)
		assert.NoError(t, err)

		// Then - All values should be nil after clear
		result1, err = cache.Get(ctx, key1)
		assert.NoError(t, err)
		assert.Nil(t, result1)
		result2, err = cache.Get(ctx, key2)
		assert.NoError(t, err)
		assert.Nil(t, result2)
	})
}
