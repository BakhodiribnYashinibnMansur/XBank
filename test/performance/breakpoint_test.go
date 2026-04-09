package performance_test

import (
	"net/http"
	"testing"
	"time"
)

// TestBreakpoint_FindMaxConcurrency finds the concurrency level where the system breaks.
// Pattern: double concurrency each step until error rate exceeds threshold.
// Reports the last stable concurrency level.
//
// Run with: go test ./test/performance/ -v -run TestBreakpoint -timeout 300s
func TestBreakpoint_FindMaxConcurrency(t *testing.T) {
	skipIfNoServer(t)

	errorThreshold := 5.0 // percent
	stepDuration := 5 * time.Second
	concurrencyLevels := []int{10, 25, 50, 100, 200, 400, 800, 1000}

	t.Log("=== Breakpoint Test: Finding max concurrency ===")
	t.Logf("  Error threshold: %.0f%%", errorThreshold)
	t.Logf("  Step duration:   %v", stepDuration)

	var lastStable int
	var breakpointLevel int

	for _, c := range concurrencyLevels {
		r := runConcurrent(c, stepDuration, func() (*http.Response, error) {
			return http.Get(baseURL + "/health")
		})

		errRate := float64(r.errors) / float64(max(r.total, 1)) * 100
		t.Logf("  C=%-4d | RPS: %6.0f | p95: %8v | p99: %8v | err: %.1f%%",
			c, r.rps, r.p95, r.p99, errRate)

		if errRate > errorThreshold {
			breakpointLevel = c
			t.Logf("\n  >>> BREAKPOINT at concurrency=%d (error rate %.1f%% > %.0f%%)", c, errRate, errorThreshold)
			break
		}
		lastStable = c
	}

	t.Log("\n=== Breakpoint Summary ===")
	if breakpointLevel > 0 {
		t.Logf("  Last stable concurrency:  %d", lastStable)
		t.Logf("  Breakpoint concurrency:   %d", breakpointLevel)
	} else {
		t.Logf("  System sustained all levels up to %d concurrent connections", concurrencyLevels[len(concurrencyLevels)-1])
	}

	// System should at least handle 50 concurrent connections
	if lastStable < 50 {
		t.Errorf("system broke before 50 concurrent connections (last stable: %d)", lastStable)
	}
}

// TestBreakpoint_FindMaxRPS finds the maximum sustainable RPS.
// Pattern: increase concurrency, observe where RPS plateaus and errors appear.
func TestBreakpoint_FindMaxRPS(t *testing.T) {
	skipIfNoServer(t)

	stepDuration := 5 * time.Second
	concurrencyLevels := []int{10, 25, 50, 100, 200, 400}

	t.Log("=== Breakpoint Test: Finding max RPS ===")

	var maxRPS float64
	var maxRPSConcurrency int

	for _, c := range concurrencyLevels {
		r := runConcurrent(c, stepDuration, func() (*http.Response, error) {
			return http.Get(baseURL + "/health")
		})

		errRate := float64(r.errors) / float64(max(r.total, 1)) * 100
		t.Logf("  C=%-4d | RPS: %6.0f | p95: %8v | err: %.1f%%", c, r.rps, r.p95, errRate)

		if r.rps > maxRPS && errRate < 1 {
			maxRPS = r.rps
			maxRPSConcurrency = c
		}

		// Stop if error rate is high
		if errRate > 10 {
			t.Logf("  >>> Stopping: error rate %.1f%% too high", errRate)
			break
		}
	}

	t.Log("\n=== Max RPS Summary ===")
	t.Logf("  Max sustainable RPS: %.0f (at concurrency=%d)", maxRPS, maxRPSConcurrency)

	// System should achieve at least 500 RPS on health endpoint
	if maxRPS < 500 {
		t.Errorf("max RPS %.0f is below 500 threshold", maxRPS)
	}
}

// TestBreakpoint_AuthenticatedMaxConcurrency finds breakpoint for authenticated requests.
func TestBreakpoint_AuthenticatedMaxConcurrency(t *testing.T) {
	skipIfNoServer(t)

	token := registerAndLogin(t, "breakpoint@perf.test")

	stepDuration := 5 * time.Second
	concurrencyLevels := []int{10, 25, 50, 100, 200, 400}
	errorThreshold := 5.0

	t.Log("=== Authenticated Breakpoint Test ===")

	var lastStable int
	for _, c := range concurrencyLevels {
		r := runConcurrent(c, stepDuration, func() (*http.Response, error) {
			return getWithAuth(baseURL+"/api/v1/accounts/list", token)
		})

		errRate := float64(r.errors) / float64(max(r.total, 1)) * 100
		t.Logf("  C=%-4d | RPS: %6.0f | p95: %8v | err: %.1f%%", c, r.rps, r.p95, errRate)

		if errRate > errorThreshold {
			t.Logf("  >>> BREAKPOINT at concurrency=%d", c)
			break
		}
		lastStable = c
	}

	t.Logf("  Last stable: %d concurrent authenticated requests", lastStable)
}
