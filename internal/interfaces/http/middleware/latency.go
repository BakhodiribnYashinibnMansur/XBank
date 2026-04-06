package middleware

import (
	"sort"
	"sync"
	"time"

	"github.com/gofiber/fiber/v2"
)

// LatencyTracker tracks request latencies and computes percentiles in real-time.
// Access via GetLatencyStats() for monitoring dashboards.
type LatencyTracker struct {
	mu       sync.Mutex
	samples  []time.Duration
	maxSize  int
}

var globalTracker = &LatencyTracker{maxSize: 10000}

// LatencyTrackerMiddleware records request latencies for percentile computation.
func LatencyTrackerMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		start := time.Now()
		err := c.Next()
		globalTracker.record(time.Since(start))
		return err
	}
}

func (t *LatencyTracker) record(d time.Duration) {
	t.mu.Lock()
	defer t.mu.Unlock()

	t.samples = append(t.samples, d)
	if len(t.samples) > t.maxSize {
		// Keep last half when buffer full
		t.samples = t.samples[t.maxSize/2:]
	}
}

// LatencyStats holds computed percentile values.
type LatencyStats struct {
	P50     time.Duration `json:"p50"`
	P90     time.Duration `json:"p90"`
	P95     time.Duration `json:"p95"`
	P99     time.Duration `json:"p99"`
	Samples int           `json:"samples"`
}

// GetLatencyStats returns current percentile statistics.
func GetLatencyStats() LatencyStats {
	globalTracker.mu.Lock()
	if len(globalTracker.samples) == 0 {
		globalTracker.mu.Unlock()
		return LatencyStats{}
	}
	sorted := make([]time.Duration, len(globalTracker.samples))
	copy(sorted, globalTracker.samples)
	globalTracker.mu.Unlock()

	sort.Slice(sorted, func(i, j int) bool { return sorted[i] < sorted[j] })
	n := len(sorted)

	return LatencyStats{
		P50:     sorted[n*50/100],
		P90:     sorted[n*90/100],
		P95:     sorted[n*95/100],
		P99:     sorted[n*99/100],
		Samples: n,
	}
}
