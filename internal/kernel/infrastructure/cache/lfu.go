package cache

import (
	"sync"
	"time"
)

// LFU implements a Least Frequently Used cache with optional TTL.
// On eviction, the entry with the lowest access frequency is removed.
// Thread-safe via sync.Mutex.
type LFU[K comparable, V any] struct {
	mu       sync.Mutex
	capacity int
	items    map[K]*lfuEntry[K, V]
	minFreq  int
	freqs    map[int]map[K]struct{} // frequency → set of keys
}

type lfuEntry[K comparable, V any] struct {
	key       K
	value     V
	freq      int
	expiresAt time.Time
}

// NewLFU creates an LFU cache with the given maximum capacity.
func NewLFU[K comparable, V any](capacity int) *LFU[K, V] {
	if capacity <= 0 {
		capacity = 256
	}
	return &LFU[K, V]{
		capacity: capacity,
		items:    make(map[K]*lfuEntry[K, V], capacity),
		freqs:    make(map[int]map[K]struct{}),
	}
}

func (c *LFU[K, V]) Set(key K, value V, ttl time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if entry, ok := c.items[key]; ok {
		entry.value = value
		if ttl > 0 {
			entry.expiresAt = time.Now().Add(ttl)
		}
		c.incrementFreq(entry)
		return
	}

	if len(c.items) >= c.capacity {
		c.evict()
	}

	entry := &lfuEntry[K, V]{key: key, value: value, freq: 1}
	if ttl > 0 {
		entry.expiresAt = time.Now().Add(ttl)
	}

	c.items[key] = entry
	if c.freqs[1] == nil {
		c.freqs[1] = make(map[K]struct{})
	}
	c.freqs[1][key] = struct{}{}
	c.minFreq = 1
}

func (c *LFU[K, V]) Get(key K) (V, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	entry, ok := c.items[key]
	if !ok {
		var zero V
		return zero, false
	}

	if !entry.expiresAt.IsZero() && time.Now().After(entry.expiresAt) {
		c.removeEntry(entry)
		var zero V
		return zero, false
	}

	c.incrementFreq(entry)
	return entry.value, true
}

func (c *LFU[K, V]) Remove(key K) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if entry, ok := c.items[key]; ok {
		c.removeEntry(entry)
	}
}

func (c *LFU[K, V]) Purge() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.items = make(map[K]*lfuEntry[K, V], c.capacity)
	c.freqs = make(map[int]map[K]struct{})
	c.minFreq = 0
}

func (c *LFU[K, V]) Len() int {
	c.mu.Lock()
	defer c.mu.Unlock()
	return len(c.items)
}

func (c *LFU[K, V]) incrementFreq(entry *lfuEntry[K, V]) {
	oldFreq := entry.freq
	delete(c.freqs[oldFreq], entry.key)
	if len(c.freqs[oldFreq]) == 0 {
		delete(c.freqs, oldFreq)
		if c.minFreq == oldFreq {
			c.minFreq++
		}
	}

	entry.freq++
	if c.freqs[entry.freq] == nil {
		c.freqs[entry.freq] = make(map[K]struct{})
	}
	c.freqs[entry.freq][entry.key] = struct{}{}
}

func (c *LFU[K, V]) evict() {
	keys := c.freqs[c.minFreq]
	for k := range keys {
		c.removeEntry(c.items[k])
		break // remove one
	}
}

func (c *LFU[K, V]) removeEntry(entry *lfuEntry[K, V]) {
	delete(c.freqs[entry.freq], entry.key)
	if len(c.freqs[entry.freq]) == 0 {
		delete(c.freqs, entry.freq)
	}
	delete(c.items, entry.key)
}
