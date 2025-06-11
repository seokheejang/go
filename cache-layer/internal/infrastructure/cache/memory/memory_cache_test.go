package memory

import (
	"context"
	"testing"
	"time"

	"github.com/go-gorm/caches/v4"
	"github.com/stretchr/testify/assert"
)

func TestNewInMemoryCache(t *testing.T) {
	tests := []struct {
		name          string
		ttl           time.Duration
		purgeInterval time.Duration
		expectedTTL   time.Duration
		expectedPurge time.Duration
	}{
		{
			name:          "default purge interval",
			ttl:           time.Hour,
			purgeInterval: 0,
			expectedTTL:   time.Hour,
			expectedPurge: time.Minute,
		},
		{
			name:          "custom purge interval",
			ttl:           time.Hour,
			purgeInterval: time.Second * 30,
			expectedTTL:   time.Hour,
			expectedPurge: time.Second * 30,
		},
		{
			name:          "no expiration",
			ttl:           0,
			purgeInterval: time.Minute,
			expectedTTL:   0,
			expectedPurge: time.Minute,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cache := NewInMemoryCache(tt.ttl, tt.purgeInterval).(*memoryCache)
			assert.Equal(t, tt.expectedTTL, cache.ttl)
			assert.Equal(t, tt.expectedPurge, cache.purgeInterval)
			assert.NotNil(t, cache.store)
			assert.NotNil(t, cache.stopJanitor)
		})
	}
}

func TestMemoryCache_GetAndStore(t *testing.T) {
	cache := NewInMemoryCache(time.Hour, time.Minute)
	ctx := context.Background()

	t.Run("store and get valid data", func(t *testing.T) {
		key := "test-key"
		query := &caches.Query[any]{
			Dest: map[string]interface{}{
				"name": "test",
				"age":  float64(25),
			},
		}

		err := cache.Store(ctx, key, query)
		assert.NoError(t, err)

		result, err := cache.Get(ctx, key, &caches.Query[any]{})
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, query.Dest, result.Dest)
	})

	t.Run("get non-existent key", func(t *testing.T) {
		result, err := cache.Get(ctx, "non-existent", &caches.Query[any]{})
		assert.NoError(t, err)
		assert.Nil(t, result)
	})

	t.Run("store nil value", func(t *testing.T) {
		err := cache.Store(ctx, "nil-key", nil)
		assert.NoError(t, err)
	})
}

func TestMemoryCache_Expiration(t *testing.T) {
	ttl := time.Millisecond * 100
	purge := time.Millisecond * 50
	cache := NewInMemoryCache(ttl, purge)
	ctx := context.Background()

	t.Run("expired data should not be retrieved", func(t *testing.T) {
		key := "expired-key"
		query := &caches.Query[any]{
			Dest: map[string]interface{}{
				"test": "data",
			},
		}

		err := cache.Store(ctx, key, query)
		assert.NoError(t, err)

		// Wait for expiration
		time.Sleep(ttl + purge)

		result, err := cache.Get(ctx, key, &caches.Query[any]{})
		assert.NoError(t, err)
		assert.Nil(t, result)
	})

	t.Run("data should be available before expiration", func(t *testing.T) {
		key := "not-expired-key"
		query := &caches.Query[any]{
			Dest: map[string]interface{}{
				"test": "data",
			},
		}

		err := cache.Store(ctx, key, query)
		assert.NoError(t, err)

		// Wait for half of TTL
		time.Sleep(ttl / 2)

		result, err := cache.Get(ctx, key, &caches.Query[any]{})
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, query.Dest, result.Dest)
	})

	t.Run("multiple keys with different expiration times", func(t *testing.T) {
		keys := []string{"key1", "key2", "key3"}
		queries := []*caches.Query[any]{
			{Dest: map[string]interface{}{"value": "data1"}},
			{Dest: map[string]interface{}{"value": "data2"}},
			{Dest: map[string]interface{}{"value": "data3"}},
		}

		// Store all keys
		for i, key := range keys {
			err := cache.Store(ctx, key, queries[i])
			assert.NoError(t, err)
		}

		// Wait for half of TTL
		time.Sleep(ttl / 2)

		// All keys should still be available
		for i, key := range keys {
			result, err := cache.Get(ctx, key, &caches.Query[any]{})
			assert.NoError(t, err)
			assert.NotNil(t, result)
			assert.Equal(t, queries[i].Dest, result.Dest)
		}

		// Wait for full expiration
		time.Sleep(ttl/2 + purge)

		// All keys should be expired
		for _, key := range keys {
			result, err := cache.Get(ctx, key, &caches.Query[any]{})
			assert.NoError(t, err)
			assert.Nil(t, result)
		}
	})

	t.Run("no expiration with ttl=0", func(t *testing.T) {
		noExpiryCache := NewInMemoryCache(0, purge)
		key := "no-expiry-key"
		query := &caches.Query[any]{
			Dest: map[string]interface{}{
				"test": "data",
			},
		}

		err := noExpiryCache.Store(ctx, key, query)
		assert.NoError(t, err)

		// Wait longer than normal TTL
		time.Sleep(ttl * 2)

		result, err := noExpiryCache.Get(ctx, key, &caches.Query[any]{})
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, query.Dest, result.Dest)
	})

	t.Run("concurrent access during expiration", func(t *testing.T) {
		key := "concurrent-key"
		query := &caches.Query[any]{
			Dest: map[string]interface{}{
				"test": "data",
			},
		}

		err := cache.Store(ctx, key, query)
		assert.NoError(t, err)

		// Start multiple goroutines to access the key
		done := make(chan bool)
		for i := 0; i < 5; i++ {
			go func() {
				for j := 0; j < 10; j++ {
					result, err := cache.Get(ctx, key, &caches.Query[any]{})
					if err != nil {
						t.Errorf("Unexpected error: %v", err)
					}
					if result != nil {
						assert.Equal(t, query.Dest, result.Dest)
					}
					time.Sleep(time.Millisecond * 10)
				}
				done <- true
			}()
		}

		// Wait for expiration
		time.Sleep(ttl + purge)

		// Wait for all goroutines to finish
		for i := 0; i < 5; i++ {
			<-done
		}

		// Verify key is expired
		result, err := cache.Get(ctx, key, &caches.Query[any]{})
		assert.NoError(t, err)
		assert.Nil(t, result)
	})
}

func TestMemoryCache_Delete(t *testing.T) {
	cache := NewInMemoryCache(time.Hour, time.Minute)
	ctx := context.Background()

	t.Run("delete existing key", func(t *testing.T) {
		key := "delete-key"
		query := &caches.Query[any]{
			Dest: map[string]interface{}{
				"test": "data",
			},
		}

		// Store the data first
		err := cache.Store(ctx, key, query)
		assert.NoError(t, err)

		// Verify data is stored
		result, err := cache.Get(ctx, key, &caches.Query[any]{})
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, query.Dest, result.Dest)

		// Delete the key
		err = cache.Delete(ctx, key)
		assert.NoError(t, err)

		// Verify data is deleted
		result, err = cache.Get(ctx, key, &caches.Query[any]{})
		assert.NoError(t, err)
		assert.Nil(t, result)
	})

	t.Run("delete non-existent key", func(t *testing.T) {
		err := cache.Delete(ctx, "non-existent")
		assert.NoError(t, err)
	})
}

func TestMemoryCache_Invalidate(t *testing.T) {
	cache := NewInMemoryCache(time.Hour, time.Minute)
	ctx := context.Background()

	t.Run("invalidate should clear all data", func(t *testing.T) {
		// Store multiple items
		keys := []string{"key1", "key2", "key3"}
		for _, key := range keys {
			query := &caches.Query[any]{
				Dest: map[string]interface{}{
					"key": key,
				},
			}
			err := cache.Store(ctx, key, query)
			assert.NoError(t, err)
		}

		// Invalidate cache
		err := cache.Invalidate(ctx)
		assert.NoError(t, err)

		// Verify all items are gone
		for _, key := range keys {
			result, err := cache.Get(ctx, key, &caches.Query[any]{})
			assert.NoError(t, err)
			assert.Nil(t, result)
		}
	})
}

func TestMemoryCache_Close(t *testing.T) {
	t.Run("close should stop janitor", func(t *testing.T) {
		cache := NewInMemoryCache(time.Hour, time.Millisecond*100).(*memoryCache)

		// Close the cache
		cache.Close()

		// Verify stopJanitor channel is closed
		select {
		case <-cache.stopJanitor:
			// Channel is closed, which is what we want
		default:
			t.Error("stopJanitor channel should be closed")
		}
	})
}
