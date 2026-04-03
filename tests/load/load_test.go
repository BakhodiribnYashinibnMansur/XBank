// HTTP Load Tests for XBank API.
// These tests measure throughput and latency under concurrent load.
//
// Requirements: Running XBank API server at XBANK_URL (default: http://localhost:3000)
//
// Run:
//   go test ./tests/load/ -v -run Load -count=1
//   go test ./tests/load/ -v -bench=. -benchtime=10s
package load

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

var baseURL string

func init() {
	baseURL = os.Getenv("XBANK_URL")
	if baseURL == "" {
		baseURL = "http://localhost:3000"
	}
}

func skipIfNoServer(t *testing.T) {
	t.Helper()
	resp, err := http.Get(baseURL + "/health")
	if err != nil {
		t.Skipf("XBank server not running at %s: %v", baseURL, err)
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Skipf("XBank server not healthy: %d", resp.StatusCode)
	}
}

// ── Throughput Tests ─────────────────────────────────

func TestLoad_HealthEndpoint(t *testing.T) {
	skipIfNoServer(t)

	concurrency := 50
	duration := 5 * time.Second

	result := runLoadTest(concurrency, duration, func() (*http.Response, error) {
		return http.Get(baseURL + "/health")
	})

	t.Logf("Health Endpoint Load Test:")
	t.Logf("  Duration:    %v", duration)
	t.Logf("  Concurrency: %d", concurrency)
	t.Logf("  Requests:    %d", result.total)
	t.Logf("  Successes:   %d", result.success)
	t.Logf("  Errors:      %d", result.errors)
	t.Logf("  RPS:         %.0f", result.rps)
	t.Logf("  Avg Latency: %v", result.avgLatency)

	if result.rps < 100 {
		t.Errorf("Health endpoint RPS too low: %.0f (expected >100)", result.rps)
	}
}

func TestLoad_ReadinessProbe(t *testing.T) {
	skipIfNoServer(t)

	concurrency := 20
	duration := 5 * time.Second

	result := runLoadTest(concurrency, duration, func() (*http.Response, error) {
		return http.Get(baseURL + "/health/ready")
	})

	t.Logf("Readiness Probe Load Test:")
	t.Logf("  Requests: %d | RPS: %.0f | Errors: %d | Avg: %v",
		result.total, result.rps, result.errors, result.avgLatency)

	if result.errors > 0 {
		t.Errorf("Readiness probe had %d errors under load", result.errors)
	}
}

func TestLoad_PublicEndpoints(t *testing.T) {
	skipIfNoServer(t)

	concurrency := 30
	duration := 5 * time.Second

	// Test exchange rates (public, no auth needed)
	result := runLoadTest(concurrency, duration, func() (*http.Response, error) {
		return http.Get(baseURL + "/api/v1/currencies/rates")
	})

	t.Logf("Public Endpoint (/currencies/rates) Load Test:")
	t.Logf("  Requests: %d | RPS: %.0f | Errors: %d | Avg: %v",
		result.total, result.rps, result.errors, result.avgLatency)
}

func TestLoad_LoginEndpoint(t *testing.T) {
	skipIfNoServer(t)

	concurrency := 10
	duration := 3 * time.Second

	body, _ := json.Marshal(map[string]string{
		"email":    "loadtest@xbank.test",
		"password": "WrongPassword123",
	})

	result := runLoadTest(concurrency, duration, func() (*http.Response, error) {
		return http.Post(baseURL+"/api/v1/auth/login", "application/json", bytes.NewReader(body))
	})

	t.Logf("Login Endpoint Load Test (expected failures — testing rate limit):")
	t.Logf("  Requests: %d | RPS: %.0f | Avg: %v",
		result.total, result.rps, result.avgLatency)
}

func TestLoad_Concurrent_HealthAndReady(t *testing.T) {
	skipIfNoServer(t)

	concurrency := 100
	duration := 5 * time.Second

	var healthOK, readyOK atomic.Int64

	var wg sync.WaitGroup
	done := make(chan struct{})
	time.AfterFunc(duration, func() { close(done) })

	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			endpoint := "/health"
			if id%2 == 0 {
				endpoint = "/health/ready"
			}
			for {
				select {
				case <-done:
					return
				default:
					resp, err := http.Get(baseURL + endpoint)
					if err == nil {
						io.Copy(io.Discard, resp.Body)
						resp.Body.Close()
						if resp.StatusCode == 200 {
							if endpoint == "/health" {
								healthOK.Add(1)
							} else {
								readyOK.Add(1)
							}
						}
					}
				}
			}
		}(i)
	}

	wg.Wait()

	t.Logf("Concurrent Health+Ready Test (100 goroutines, 5s):")
	t.Logf("  /health OK:  %d", healthOK.Load())
	t.Logf("  /ready OK:   %d", readyOK.Load())
	t.Logf("  Total:       %d", healthOK.Load()+readyOK.Load())
}

// ── Latency Distribution ─────────────────────────────

func TestLoad_LatencyDistribution(t *testing.T) {
	skipIfNoServer(t)

	n := 200
	latencies := make([]time.Duration, 0, n)
	var mu sync.Mutex

	var wg sync.WaitGroup
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			start := time.Now()
			resp, err := http.Get(baseURL + "/health")
			elapsed := time.Since(start)
			if err == nil {
				io.Copy(io.Discard, resp.Body)
				resp.Body.Close()
			}
			mu.Lock()
			latencies = append(latencies, elapsed)
			mu.Unlock()
		}()
	}
	wg.Wait()

	// Sort and compute percentiles
	sortDurations(latencies)

	p50 := latencies[len(latencies)*50/100]
	p90 := latencies[len(latencies)*90/100]
	p95 := latencies[len(latencies)*95/100]
	p99 := latencies[len(latencies)*99/100]

	t.Logf("Latency Distribution (%d requests):", n)
	t.Logf("  p50: %v", p50)
	t.Logf("  p90: %v", p90)
	t.Logf("  p95: %v", p95)
	t.Logf("  p99: %v", p99)

	if p95 > 500*time.Millisecond {
		t.Errorf("p95 latency too high: %v (expected <500ms)", p95)
	}
}

// ── Helper Functions ─────────────────────────────────

type loadResult struct {
	total      int64
	success    int64
	errors     int64
	rps        float64
	avgLatency time.Duration
}

func runLoadTest(concurrency int, duration time.Duration, reqFn func() (*http.Response, error)) loadResult {
	var total, success, errors atomic.Int64
	var totalLatency atomic.Int64

	done := make(chan struct{})
	time.AfterFunc(duration, func() { close(done) })

	var wg sync.WaitGroup
	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				select {
				case <-done:
					return
				default:
					start := time.Now()
					resp, err := reqFn()
					elapsed := time.Since(start)

					total.Add(1)
					totalLatency.Add(int64(elapsed))

					if err != nil {
						errors.Add(1)
						continue
					}
					io.Copy(io.Discard, resp.Body)
					resp.Body.Close()

					if resp.StatusCode >= 200 && resp.StatusCode < 500 {
						success.Add(1)
					} else {
						errors.Add(1)
					}
				}
			}
		}()
	}

	wg.Wait()

	t := total.Load()
	avg := time.Duration(0)
	rps := float64(0)
	if t > 0 {
		avg = time.Duration(totalLatency.Load() / t)
		rps = float64(t) / duration.Seconds()
	}

	return loadResult{
		total:      t,
		success:    success.Load(),
		errors:     errors.Load(),
		rps:        rps,
		avgLatency: avg,
	}
}

func sortDurations(d []time.Duration) {
	for i := 1; i < len(d); i++ {
		key := d[i]
		j := i - 1
		for j >= 0 && d[j] > key {
			d[j+1] = d[j]
			j--
		}
		d[j+1] = key
	}
}

// ── Benchmarks ───────────────────────────────────────

func BenchmarkHTTP_Health(b *testing.B) {
	resp, err := http.Get(baseURL + "/health")
	if err != nil {
		b.Skip("Server not running")
	}
	resp.Body.Close()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		resp, err := http.Get(baseURL + "/health")
		if err == nil {
			io.Copy(io.Discard, resp.Body)
			resp.Body.Close()
		}
	}
}

func BenchmarkHTTP_Readiness(b *testing.B) {
	resp, err := http.Get(baseURL + "/health/ready")
	if err != nil {
		b.Skip("Server not running")
	}
	resp.Body.Close()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		resp, err := http.Get(baseURL + "/health/ready")
		if err == nil {
			io.Copy(io.Discard, resp.Body)
			resp.Body.Close()
		}
	}
}

func BenchmarkHTTP_Health_Parallel(b *testing.B) {
	resp, err := http.Get(baseURL + "/health")
	if err != nil {
		b.Skip("Server not running")
	}
	resp.Body.Close()

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			resp, err := http.Get(baseURL + "/health")
			if err == nil {
				io.Copy(io.Discard, resp.Body)
				resp.Body.Close()
			}
		}
	})
}

func BenchmarkHTTP_Login_Parallel(b *testing.B) {
	resp, err := http.Get(baseURL + "/health")
	if err != nil {
		b.Skip("Server not running")
	}
	resp.Body.Close()

	body := []byte(`{"email":"bench@xbank.test","password":"Wrong123456"}`)

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			resp, err := http.Post(
				baseURL+"/api/v1/auth/login",
				"application/json",
				bytes.NewReader(body),
			)
			if err == nil {
				io.Copy(io.Discard, resp.Body)
				resp.Body.Close()
			}
		}
	})
}

func init() {
	// Suppress fmt import requirement
	_ = fmt.Sprintf
}
