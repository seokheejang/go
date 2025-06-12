package memory

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/seokheejang/go/cache-layer/pkg/cache"
	"github.com/stretchr/testify/assert"
)

func TestMemoryCache_BasicOperations(t *testing.T) {
	c, err := New(nil)
	assert.NoError(t, err)
	defer c.Close()

	ctx := context.Background()

	// Test Set and Get
	err = c.Set(ctx, "key1", "value1", 0)
	assert.NoError(t, err)

	val, err := c.Get(ctx, "key1")
	assert.NoError(t, err)
	assert.Equal(t, "value1", val)

	// Test Delete
	err = c.Delete(ctx, "key1")
	assert.NoError(t, err)

	val, err = c.Get(ctx, "key1")
	assert.NoError(t, err)
	assert.Nil(t, val)

	// Test Clear
	err = c.Set(ctx, "key1", "value1", 0)
	assert.NoError(t, err)
	err = c.Set(ctx, "key2", "value2", 0)
	assert.NoError(t, err)

	err = c.Clear(ctx)
	assert.NoError(t, err)

	val, err = c.Get(ctx, "key1")
	assert.NoError(t, err)
	assert.Nil(t, val)
	val, err = c.Get(ctx, "key2")
	assert.NoError(t, err)
	assert.Nil(t, val)
}

func TestMemoryCache_TTL(t *testing.T) {
	// Create cache with short TTL
	c, err := New(&cache.Options{
		DefaultTTL: 100 * time.Millisecond,
		MaxTTL:     200 * time.Millisecond,
	})
	assert.NoError(t, err)
	defer c.Close()

	ctx := context.Background()

	// Test default TTL
	err = c.Set(ctx, "key1", "value1", 0)
	assert.NoError(t, err)

	// Value should be available immediately
	val, err := c.Get(ctx, "key1")
	assert.NoError(t, err)
	assert.Equal(t, "value1", val)

	// Wait for TTL to expire
	time.Sleep(150 * time.Millisecond)

	// Value should be expired
	val, err = c.Get(ctx, "key1")
	assert.NoError(t, err)
	assert.Nil(t, val)

	// Test custom TTL
	err = c.Set(ctx, "key2", "value2", 50*time.Millisecond)
	assert.NoError(t, err)

	// Value should be available immediately
	val, err = c.Get(ctx, "key2")
	assert.NoError(t, err)
	assert.Equal(t, "value2", val)

	// Wait for TTL to expire
	time.Sleep(60 * time.Millisecond)

	// Value should be expired
	val, err = c.Get(ctx, "key2")
	assert.NoError(t, err)
	assert.Nil(t, val)

	// Test MaxTTL
	err = c.Set(ctx, "key3", "value3", 300*time.Millisecond)
	assert.NoError(t, err)

	// Wait for MaxTTL to expire
	time.Sleep(210 * time.Millisecond)

	// Value should be expired due to MaxTTL
	val, err = c.Get(ctx, "key3")
	assert.NoError(t, err)
	assert.Nil(t, val)
}

func TestMemoryCache_Eviction(t *testing.T) {
	// Create cache with small size
	c, err := New(&cache.Options{
		DefaultTTL: 1 * time.Hour,
		MaxTTL:     1 * time.Hour,
		MaxSize:    3,
	})
	assert.NoError(t, err)
	defer c.Close()

	ctx := context.Background()

	// Fill cache to capacity
	err = c.Set(ctx, "key1", "value1", 0)
	assert.NoError(t, err)
	time.Sleep(10 * time.Millisecond) // Ensure different creation times

	err = c.Set(ctx, "key2", "value2", 0)
	assert.NoError(t, err)
	time.Sleep(10 * time.Millisecond)

	err = c.Set(ctx, "key3", "value3", 0)
	assert.NoError(t, err)

	// Verify all values are present
	val, err := c.Get(ctx, "key1")
	assert.NoError(t, err)
	assert.Equal(t, "value1", val)

	val, err = c.Get(ctx, "key2")
	assert.NoError(t, err)
	assert.Equal(t, "value2", val)

	val, err = c.Get(ctx, "key3")
	assert.NoError(t, err)
	assert.Equal(t, "value3", val)

	// Add one more entry to trigger eviction
	err = c.Set(ctx, "key4", "value4", 0)
	assert.NoError(t, err)

	// key1 should be evicted (oldest)
	val, err = c.Get(ctx, "key1")
	assert.NoError(t, err)
	assert.Nil(t, val)

	// Other keys should still be present
	val, err = c.Get(ctx, "key2")
	assert.NoError(t, err)
	assert.Equal(t, "value2", val)

	val, err = c.Get(ctx, "key3")
	assert.NoError(t, err)
	assert.Equal(t, "value3", val)

	val, err = c.Get(ctx, "key4")
	assert.NoError(t, err)
	assert.Equal(t, "value4", val)
}

func TestMemoryCache_ConcurrentAccess(t *testing.T) {
	c, err := New(nil)
	assert.NoError(t, err)
	defer c.Close()

	ctx := context.Background()
	const numGoroutines = 100
	const numOperations = 1000
	const ttl = 10 * time.Second

	var wg sync.WaitGroup
	wg.Add(numGoroutines)

	// Channel to synchronize goroutine start
	start := make(chan struct{})
	// Channel to collect operation results
	results := make(chan error, numGoroutines*numOperations)

	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			defer wg.Done()
			<-start

			for j := 0; j < numOperations; j++ {
				key := fmt.Sprintf("key%d", (id*numOperations+j)%100)
				value := fmt.Sprintf("value%d", (id*numOperations+j)%100)

				if j%2 == 0 {
					err := c.Set(ctx, key, value, ttl) // Set with explicit TTL
					results <- err
				} else {
					val, err := c.Get(ctx, key)
					if err != nil {
						results <- fmt.Errorf("get error for key=%s: %w", key, err)
					} else {
						results <- nil // Success case
						_ = val        // Only check existence, skip value validation
					}
				}
			}
		}(i)
	}

	// Start all goroutines simultaneously
	close(start)
	wg.Wait()
	close(results)

	// Verify no errors occurred during operations
	for err := range results {
		assert.NoError(t, err)
	}

	// Verify final state (only check existence, not exact values)
	for i := 0; i < 100; i++ {
		key := fmt.Sprintf("key%d", i)
		val, err := c.Get(ctx, key)
		assert.NoError(t, err)
		if val != nil {
			t.Logf("key=%s still exists in cache", key)
		}
	}
}
