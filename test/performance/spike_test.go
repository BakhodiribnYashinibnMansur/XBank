package performance_test

import (
	"net/http"
	"testing"
	"time"
)

// TestSpike_SuddenTrafficBurst simulates a sudden traffic spike.
// Pattern: low baseline → sudden spike → back to baseline.
// Validates that the system recovers after the spike.
func TestSpike_SuddenTrafficBurst(t *testing.T) {
	skipIfNoServer(t)

	phases := []phase{
		{concurrency: 10, duration: 5 * time.Second},  // baseline
		{concurrency: 500, duration: 5 * time.Second}, // spike
		{concurrency: 10, duration: 5 * time.Second},  // recovery
	}

	results := runPhased(t, phases, func() (*http.Response, error) {
		return http.Get(baseURL + "/health")
	})

	baseline := results[0]
	spike := results[1]
	recovery := results[2]

	t.Log("\n=== Spike Test Summary ===")
	t.Logf("  Baseline:  RPS=%.0f p95=%v errors=%d", baseline.rps, baseline.p95, baseline.errors)
	t.Logf("  Spike:     RPS=%.0f p95=%v errors=%d", spike.rps, spike.p95, spike.errors)
	t.Logf("  Recovery:  RPS=%.0f p95=%v errors=%d", recovery.rps, recovery.p95, recovery.errors)

	// Recovery should have 0 errors (system fully recovered)
	if recovery.errors > 0 {
		t.Errorf("system did not recover: %d errors in recovery phase", recovery.errors)
	}

	// Recovery RPS should be at least 50% of baseline
	if recovery.rps < baseline.rps*0.5 {
		t.Errorf("recovery RPS (%.0f) is less than 50%% of baseline (%.0f)", recovery.rps, baseline.rps)
	}
}

// TestSpike_AuthenticatedBurst simulates a spike on authenticated endpoints.
func TestSpike_AuthenticatedBurst(t *testing.T) {
	skipIfNoServer(t)

	token := registerAndLogin(t, "spike-auth@perf.test")

	phases := []phase{
		{concurrency: 5, duration: 3 * time.Second},   // baseline
		{concurrency: 200, duration: 5 * time.Second},  // spike
		{concurrency: 5, duration: 3 * time.Second},   // recovery
	}

	results := runPhased(t, phases, func() (*http.Response, error) {
		return getWithAuth(baseURL+"/api/v1/accounts/list", token)
	})

	t.Log("\n=== Authenticated Spike Summary ===")
	for i, r := range results {
		labels := []string{"Baseline", "Spike", "Recovery"}
		t.Logf("  %-9s: RPS=%.0f p95=%v errors=%d", labels[i], r.rps, r.p95, r.errors)
	}

	// Recovery phase should recover
	if results[2].errors > 0 {
		t.Errorf("authenticated endpoint did not recover: %d errors", results[2].errors)
	}
}
