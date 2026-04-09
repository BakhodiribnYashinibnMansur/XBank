package performance_test

import (
	"net/http"
	"testing"
	"time"
)

// TestStress_HealthEndpoint gradually increases load to find performance limits.
// Pattern: ramp up concurrency in stages, measure degradation.
func TestStress_HealthEndpoint(t *testing.T) {
	skipIfNoServer(t)

	phases := []phase{
		{concurrency: 10, duration: 5 * time.Second},
		{concurrency: 50, duration: 5 * time.Second},
		{concurrency: 100, duration: 5 * time.Second},
		{concurrency: 200, duration: 5 * time.Second},
		{concurrency: 500, duration: 5 * time.Second},
	}

	results := runPhased(t, phases, func() (*http.Response, error) {
		return http.Get(baseURL + "/health")
	})

	t.Log("\n=== Stress Test Summary ===")
	for i, r := range results {
		errRate := float64(r.errors) / float64(max(r.total, 1)) * 100
		t.Logf("  C=%-3d | RPS: %6.0f | p95: %8v | p99: %8v | err: %.1f%%",
			phases[i].concurrency, r.rps, r.p95, r.p99, errRate)
	}

	// Final phase error rate should be < 5%
	last := results[len(results)-1]
	errRate := float64(last.errors) / float64(max(last.total, 1)) * 100
	if errRate > 5 {
		t.Errorf("error rate at max concurrency: %.1f%% (threshold: 5%%)", errRate)
	}
}

// TestStress_AuthenticatedEndpoint stresses a protected endpoint.
func TestStress_AuthenticatedEndpoint(t *testing.T) {
	skipIfNoServer(t)

	token := registerAndLogin(t, "stress-auth@perf.test")

	phases := []phase{
		{concurrency: 10, duration: 5 * time.Second},
		{concurrency: 50, duration: 5 * time.Second},
		{concurrency: 100, duration: 5 * time.Second},
	}

	results := runPhased(t, phases, func() (*http.Response, error) {
		return getWithAuth(baseURL+"/api/v1/accounts/list", token)
	})

	t.Log("\n=== Authenticated Stress Summary ===")
	for i, r := range results {
		t.Logf("  C=%-3d | RPS: %6.0f | p95: %8v | err: %d",
			phases[i].concurrency, r.rps, r.p95, r.errors)
	}
}

// TestStress_LoginEndpoint stresses login with rate limiting.
func TestStress_LoginEndpoint(t *testing.T) {
	skipIfNoServer(t)

	phases := []phase{
		{concurrency: 5, duration: 3 * time.Second},
		{concurrency: 20, duration: 3 * time.Second},
		{concurrency: 50, duration: 3 * time.Second},
	}

	results := runPhased(t, phases, func() (*http.Response, error) {
		return postJSON(baseURL+"/api/v1/auth/login", map[string]string{
			"email": "stress-login@perf.test", "password": "Wrong123!",
		})
	})

	t.Log("\n=== Login Stress Summary (rate-limited) ===")
	for i, r := range results {
		t.Logf("  C=%-3d | RPS: %6.0f | p95: %8v | 429s expected",
			phases[i].concurrency, r.rps, r.p95)
	}
}

func max(a, b int64) int64 {
	if a > b {
		return a
	}
	return b
}
