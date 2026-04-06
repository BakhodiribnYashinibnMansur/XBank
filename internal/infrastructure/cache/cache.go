// Package cache provides thread-safe in-memory cache implementations
// with various eviction strategies. All caches implement the Cache interface.
package cache

import "time"

// Cache is the generic interface for all in-memory caches.
type Cache[K comparable, V any] interface {
	// Set adds or updates a key-value pair with optional TTL.
	// Zero TTL means no expiration.
	Set(key K, value V, ttl time.Duration)
	// Get retrieves a value by key. Returns false if not found or expired.
	Get(key K) (V, bool)
	// Remove deletes a key from the cache.
	Remove(key K)
	// Purge clears all entries.
	Purge()
	// Len returns the number of entries in the cache.
	Len() int
}
