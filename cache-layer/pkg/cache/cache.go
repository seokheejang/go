package cache

import (
	"context"
	"time"
)

// Options represents the configuration options for a cache
type Options struct {
	// DefaultTTL is the default time-to-live for cache entries
	DefaultTTL time.Duration
	// MaxTTL is the maximum time-to-live for cache entries
	MaxTTL time.Duration
	// MaxSize is the maximum number of entries in the cache
	MaxSize int64
}

// Entry represents a single cache entry
type Entry struct {
	Key     string
	Value   interface{}
	TTL     time.Duration
	Created time.Time
}

// NewEntry creates a new cache entry
func NewEntry(key string, value interface{}, ttl time.Duration) *Entry {
	return &Entry{
		Key:     key,
		Value:   value,
		TTL:     ttl,
		Created: time.Now(),
	}
}

// IsExpired checks if the cache entry has expired
func (e *Entry) IsExpired() bool {
	return time.Since(e.Created) > e.TTL
}

// Cache defines the interface for cache operations
type Cache interface {
	// Get retrieves a value from the cache
	Get(ctx context.Context, key string) (interface{}, error)

	// Set stores a value in the cache
	Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error

	// Delete removes a value from the cache
	Delete(ctx context.Context, key string) error

	// Clear removes all values from the cache
	Clear(ctx context.Context) error

	// Close releases any resources used by the cache
	Close() error
}
