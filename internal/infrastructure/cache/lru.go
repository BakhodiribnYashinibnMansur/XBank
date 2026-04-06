package cache

import (
	"container/list"
	"sync"
	"time"
)

// LRU implements a Least Recently Used cache with optional TTL per entry.
// Thread-safe via sync.RWMutex.
type LRU[K comparable, V any] struct {
	mu       sync.RWMutex
	capacity int
	items    map[K]*list.Element
	order    *list.List // front = most recent, back = least recent
}

type lruEntry[K comparable, V any] struct {
	key       K
	value     V
	expiresAt time.Time // zero = no expiry
}

// NewLRU creates an LRU cache with the given maximum capacity.
func NewLRU[K comparable, V any](capacity int) *LRU[K, V] {
	if capacity <= 0 {
		capacity = 256
	}
	return &LRU[K, V]{
		capacity: capacity,
		items:    make(map[K]*list.Element, capacity),
		order:    list.New(),
	}
}

func (c *LRU[K, V]) Set(key K, value V, ttl time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()

	entry := lruEntry[K, V]{key: key, value: value}
	if ttl > 0 {
		entry.expiresAt = time.Now().Add(ttl)
	}

	if el, ok := c.items[key]; ok {
		el.Value = entry
		c.order.MoveToFront(el)
		return
	}

	if c.order.Len() >= c.capacity {
		c.evict()
	}

	el := c.order.PushFront(entry)
	c.items[key] = el
}

func (c *LRU[K, V]) Get(key K) (V, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	el, ok := c.items[key]
	if !ok {
		var zero V
		return zero, false
	}

	entry := el.Value.(lruEntry[K, V])

	// Check TTL
	if !entry.expiresAt.IsZero() && time.Now().After(entry.expiresAt) {
		c.removeElement(el)
		var zero V
		return zero, false
	}

	c.order.MoveToFront(el)
	return entry.value, true
}

func (c *LRU[K, V]) Remove(key K) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if el, ok := c.items[key]; ok {
		c.removeElement(el)
	}
}

func (c *LRU[K, V]) Purge() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.items = make(map[K]*list.Element, c.capacity)
	c.order.Init()
}

func (c *LRU[K, V]) Len() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.order.Len()
}

func (c *LRU[K, V]) evict() {
	el := c.order.Back()
	if el != nil {
		c.removeElement(el)
	}
}

func (c *LRU[K, V]) removeElement(el *list.Element) {
	entry := el.Value.(lruEntry[K, V])
	delete(c.items, entry.key)
	c.order.Remove(el)
}
