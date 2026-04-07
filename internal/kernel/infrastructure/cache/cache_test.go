package cache

import (
	"testing"
	"time"
)

// ── LRU Tests ────────────────────────────

func TestLRU_SetAndGet(t *testing.T) {
	c := NewLRU[string, int](3)
	c.Set("a", 1, 0)
	c.Set("b", 2, 0)

	v, ok := c.Get("a")
	if !ok || v != 1 {
		t.Errorf("expected 1, got %d", v)
	}
}

func TestLRU_Eviction(t *testing.T) {
	c := NewLRU[string, int](2)
	c.Set("a", 1, 0)
	c.Set("b", 2, 0)
	c.Set("c", 3, 0) // evicts "a"

	if _, ok := c.Get("a"); ok {
		t.Error("a should be evicted")
	}
	if v, ok := c.Get("b"); !ok || v != 2 {
		t.Error("b should exist")
	}
}

func TestLRU_AccessUpdatesOrder(t *testing.T) {
	c := NewLRU[string, int](2)
	c.Set("a", 1, 0)
	c.Set("b", 2, 0)
	c.Get("a")        // access a → moves to front
	c.Set("c", 3, 0)  // evicts "b" (least recent)

	if _, ok := c.Get("a"); !ok {
		t.Error("a should survive eviction after access")
	}
	if _, ok := c.Get("b"); ok {
		t.Error("b should be evicted")
	}
}

func TestLRU_TTL(t *testing.T) {
	c := NewLRU[string, int](10)
	c.Set("x", 42, 50*time.Millisecond)

	v, ok := c.Get("x")
	if !ok || v != 42 {
		t.Error("should find before TTL")
	}

	time.Sleep(60 * time.Millisecond)

	_, ok = c.Get("x")
	if ok {
		t.Error("should expire after TTL")
	}
}

func TestLRU_Remove(t *testing.T) {
	c := NewLRU[string, int](10)
	c.Set("a", 1, 0)
	c.Remove("a")

	if _, ok := c.Get("a"); ok {
		t.Error("should be removed")
	}
}

func TestLRU_Purge(t *testing.T) {
	c := NewLRU[string, int](10)
	c.Set("a", 1, 0)
	c.Set("b", 2, 0)
	c.Purge()

	if c.Len() != 0 {
		t.Errorf("expected 0 after purge, got %d", c.Len())
	}
}

// ── LFU Tests ────────────────────────────

func TestLFU_SetAndGet(t *testing.T) {
	c := NewLFU[string, int](3)
	c.Set("a", 1, 0)

	v, ok := c.Get("a")
	if !ok || v != 1 {
		t.Errorf("expected 1, got %d", v)
	}
}

func TestLFU_EvictsLeastFrequent(t *testing.T) {
	c := NewLFU[string, int](2)
	c.Set("a", 1, 0)
	c.Set("b", 2, 0)
	c.Get("a") // freq(a)=2, freq(b)=1
	c.Set("c", 3, 0) // evicts "b" (lowest freq)

	if _, ok := c.Get("a"); !ok {
		t.Error("a should survive (higher freq)")
	}
	if _, ok := c.Get("b"); ok {
		t.Error("b should be evicted (lowest freq)")
	}
}

func TestLFU_TTL(t *testing.T) {
	c := NewLFU[string, int](10)
	c.Set("x", 42, 50*time.Millisecond)

	time.Sleep(60 * time.Millisecond)

	_, ok := c.Get("x")
	if ok {
		t.Error("should expire after TTL")
	}
}

func TestLFU_Purge(t *testing.T) {
	c := NewLFU[string, int](10)
	c.Set("a", 1, 0)
	c.Set("b", 2, 0)
	c.Purge()

	if c.Len() != 0 {
		t.Errorf("expected 0 after purge, got %d", c.Len())
	}
}
