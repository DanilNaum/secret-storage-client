// Package cache provides a generic TTL-based caching system.
// It implements a thread-safe cache that can store any type of data
// with configurable time-to-live (TTL) values for automatic expiration.
// This package uses Go generics to provide type safety while maintaining flexibility.
package cache

import (
	"sync"
	"time"
)

// CacheEntry represents a cached item with metadata.
// It stores the cached data along with timing information
// such as when it was cached and its time-to-live duration.
// The generic type T allows this cache to store any type of data.
type CacheEntry[T any] struct {
	// Record is the cached data of type T.
	Record *T
	// CachedAt is the timestamp when the item was cached.
	CachedAt time.Time
	// TTL is the time-to-live duration for this cache entry.
	// After this duration, the entry is considered expired.
	TTL time.Duration
}

// IsExpired checks if the cache entry has expired.
// Returns true if the time since caching exceeds the TTL duration.
// Expired entries should be refreshed or removed from cache.
func (e *CacheEntry[T]) IsExpired() bool {
	return time.Since(e.CachedAt) > e.TTL
}

// Cache manages caching of items with TTL support using Go generics.
// It provides thread-safe operations for storing and retrieving data
// with automatic expiration based on configurable TTL values.
// The generic type T allows the cache to store any type of data while maintaining type safety.
type Cache[T any] struct {
	// entries stores the cached items indexed by string keys.
	entries map[string]*CacheEntry[T]
	// mutex provides thread-safe access to the cache.
	mutex sync.RWMutex
	// defaultTTL is the default time-to-live for new cache entries.
	defaultTTL time.Duration
}

// NewCache creates a new generic cache with the specified default TTL.
// The defaultTTL parameter sets how long items remain valid in the cache
// before they are considered expired and need to be refreshed.
// 
// Example usage:
//   recordCache := NewCache[models.Record](5 * time.Minute)
//   userCache := NewCache[User](1 * time.Hour)
func NewCache[T any](defaultTTL time.Duration) *Cache[T] {
	return &Cache[T]{
		entries:    make(map[string]*CacheEntry[T]),
		defaultTTL: defaultTTL,
	}
}

// Get retrieves an item from cache if it exists and is not expired.
// Returns the cached item, whether it was found, and when it was cached.
// If the item is expired, it returns false for found but still provides the cached time.
// This allows callers to distinguish between "not found" and "expired" states.
//
// The method is thread-safe and uses read locks for optimal performance.
func (c *Cache[T]) Get(id string) (*T, bool, time.Time) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	entry, exists := c.entries[id]
	if !exists {
		return nil, false, time.Time{}
	}

	if entry.IsExpired() {
		return nil, false, entry.CachedAt
	}

	return entry.Record, true, entry.CachedAt
}

// Set stores an item in cache with the default TTL.
// The item will be cached with the default TTL specified during cache creation.
// This is a convenience method that uses the cache's default TTL setting.
//
// The method is thread-safe and uses write locks to ensure data consistency.
func (c *Cache[T]) Set(id string, record *T) {
	c.SetWithTTL(id, record, c.defaultTTL)
}

// SetWithTTL stores an item in cache with a custom TTL.
// This allows for per-item TTL customization, useful for data
// that may need different caching strategies based on content or usage patterns.
// For example, frequently accessed items might have longer TTLs,
// while sensitive data might have shorter TTLs.
//
// The method is thread-safe and uses write locks to ensure data consistency.
func (c *Cache[T]) SetWithTTL(id string, record *T, ttl time.Duration) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	c.entries[id] = &CacheEntry[T]{
		Record:   record,
		CachedAt: time.Now(),
		TTL:      ttl,
	}
}

// Remove removes an item from cache by its ID.
// Used when an item is deleted from the source or when manual cache invalidation is needed.
// This operation is idempotent - removing a non-existent item has no effect.
//
// The method is thread-safe and uses write locks to ensure data consistency.
func (c *Cache[T]) Remove(id string) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	delete(c.entries, id)
}

// Clear removes all entries from cache.
// This is useful for cache invalidation scenarios, memory cleanup,
// or when switching contexts that require a fresh cache state.
// After calling Clear, the cache will be empty and ready for new entries.
//
// The method is thread-safe and uses write locks to ensure data consistency.
func (c *Cache[T]) Clear() {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	c.entries = make(map[string]*CacheEntry[T])
}

// GetCachedAt returns the cache timestamp for an item.
// Returns the time when the item was cached and whether it exists in cache.
// This is useful for displaying cache status information in user interfaces
// or for implementing custom cache management logic.
//
// The method is thread-safe and uses read locks for optimal performance.
func (c *Cache[T]) GetCachedAt(id string) (time.Time, bool) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	entry, exists := c.entries[id]
	if !exists {
		return time.Time{}, false
	}

	return entry.CachedAt, true
}

// IsExpired checks if a cached item is expired.
// Returns true if the item doesn't exist in cache or if it has exceeded its TTL.
// This method is useful for cache validation and for determining
// whether cached data should be refreshed from the source.
//
// The method is thread-safe and uses read locks for optimal performance.
func (c *Cache[T]) IsExpired(id string) bool {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	entry, exists := c.entries[id]
	if !exists {
		return true
	}

	return entry.IsExpired()
}