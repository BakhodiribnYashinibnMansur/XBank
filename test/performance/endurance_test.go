package performance_test

import (
	"net/http"
	"testing"
	"time"
)

// TestEndurance_SustainedLoad runs moderate load for an extended period.
// Pattern: constant concurrency for a long duration (2 minutes).
// Detects memory leaks, connection pool exhaustion, and gradual degradation.
//
// Run with: go test ./test/performance/ -v -run TestEndurance -timeout 300s
func TestEndurance_SustainedLoad(t *testing.T) {
	skipIfNoServer(t)

	// Split into 6 intervals of 20s each to track degradation over time
	intervals := 6
	intervalDuration := 20 * time.Second
	concurrency := 50

	t.Logf("Endurance test: %d intervals × %v = %v total, concurrency=%d",
		intervals, intervalDuration, time.Duration(intervals)*intervalDuration, concurrency)

	var allResults []result
	for i := 0; i < intervals; i++ {
		r := runConcurrent(concurrency, intervalDuration, func() (*http.Response, error) {
			return http.Get(baseURL + "/health")
		})
		logResult(t, timeFmt(i, intervalDuration), r)
		allResults = append(allResults, r)
	}

	t.Log("\n=== Endurance Summary ===")
	first := allResults[0]
	last := allResults[len(allResults)-1]

	t.Logf("  First interval:  RPS=%.0f p95=%v errors=%d", first.rps, first.p95, first.errors)
	t.Logf("  Last interval:   RPS=%.0f p95=%v errors=%d", last.rps, last.p95, last.errors)

	// Check for degradation: last interval p95 should not be >3x of first
	if last.p95 > first.p95*3 && first.p95 > 0 {
		t.Errorf("latency degradation: first p95=%v, last p95=%v (>3x)", first.p95, last.p95)
	}

	// Error rate should stay below 1% across all intervals
	var totalReqs, totalErrors int64
	for _, r := range allResults {
		totalReqs += r.total
		totalErrors += r.errors
	}
	errRate := float64(totalErrors) / float64(max(totalReqs, 1)) * 100
	t.Logf("  Total requests:  %d", totalReqs)
	t.Logf("  Total errors:    %d (%.2f%%)", totalErrors, errRate)

	if errRate > 1 {
		t.Errorf("overall error rate %.2f%% exceeds 1%% threshold", errRate)
	}
}

// TestEndurance_AuthenticatedSustained runs sustained load on authenticated endpoints.
func TestEndurance_AuthenticatedSustained(t *testing.T) {
	skipIfNoServer(t)

	token := registerAndLogin(t, "endurance@perf.test")

	intervals := 4
	intervalDuration := 15 * time.Second
	concurrency := 30

	t.Logf("Authenticated endurance: %d × %v, concurrency=%d", intervals, intervalDuration, concurrency)

	var allResults []result
	for i := 0; i < intervals; i++ {
		r := runConcurrent(concurrency, intervalDuration, func() (*http.Response, error) {
			return getWithAuth(baseURL+"/api/v1/accounts/list", token)
		})
		logResult(t, timeFmt(i, intervalDuration), r)
		allResults = append(allResults, r)
	}

	// Check no significant degradation
	first := allResults[0]
	last := allResults[len(allResults)-1]

	if last.p95 > first.p95*3 && first.p95 > 0 {
		t.Errorf("authenticated latency degradation: first p95=%v, last p95=%v", first.p95, last.p95)
	}
}

func timeFmt(interval int, dur time.Duration) string {
	start := time.Duration(interval) * dur
	end := start + dur
	return start.String() + "-" + end.String()
}
