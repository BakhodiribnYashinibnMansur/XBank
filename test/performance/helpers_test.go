package performance_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"os"
	"sort"
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
}

// --- Load runner ---

type phase struct {
	concurrency int
	duration    time.Duration
}

type result struct {
	total      int64
	success    int64
	errors     int64
	rps        float64
	avgLatency time.Duration
	p50, p90, p95, p99 time.Duration
	maxLatency time.Duration
	latencies  []time.Duration
}

// runPhased executes multiple phases sequentially, returning per-phase results.
func runPhased(t *testing.T, phases []phase, reqFn func() (*http.Response, error)) []result {
	t.Helper()
	var results []result
	for i, p := range phases {
		t.Logf("Phase %d: concurrency=%d duration=%v", i+1, p.concurrency, p.duration)
		r := runConcurrent(p.concurrency, p.duration, reqFn)
		logResult(t, fmt.Sprintf("Phase %d", i+1), r)
		results = append(results, r)
	}
	return results
}

// runConcurrent runs reqFn with given concurrency for given duration, collecting metrics.
func runConcurrent(concurrency int, duration time.Duration, reqFn func() (*http.Response, error)) result {
	var total, success, errors atomic.Int64
	var mu sync.Mutex
	var latencies []time.Duration

	done := make(chan struct{})
	time.AfterFunc(duration, func() { close(done) })

	var wg sync.WaitGroup
	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			var local []time.Duration
			for {
				select {
				case <-done:
					mu.Lock()
					latencies = append(latencies, local...)
					mu.Unlock()
					return
				default:
					start := time.Now()
					resp, err := reqFn()
					elapsed := time.Since(start)

					total.Add(1)
					local = append(local, elapsed)

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
	r := result{
		total:   t,
		success: success.Load(),
		errors:  errors.Load(),
	}

	if t > 0 {
		sort.Slice(latencies, func(i, j int) bool { return latencies[i] < latencies[j] })
		r.latencies = latencies
		r.rps = float64(t) / duration.Seconds()
		r.avgLatency = avgDuration(latencies)
		r.p50 = percentile(latencies, 50)
		r.p90 = percentile(latencies, 90)
		r.p95 = percentile(latencies, 95)
		r.p99 = percentile(latencies, 99)
		r.maxLatency = latencies[len(latencies)-1]
	}

	return r
}

func logResult(t *testing.T, label string, r result) {
	t.Helper()
	errRate := float64(0)
	if r.total > 0 {
		errRate = float64(r.errors) / float64(r.total) * 100
	}
	t.Logf("  %s: requests=%d success=%d errors=%d(%.1f%%) rps=%.0f avg=%v p50=%v p90=%v p95=%v p99=%v max=%v",
		label, r.total, r.success, r.errors, errRate, r.rps, r.avgLatency,
		r.p50, r.p90, r.p95, r.p99, r.maxLatency)
}

func percentile(sorted []time.Duration, p int) time.Duration {
	if len(sorted) == 0 {
		return 0
	}
	idx := int(math.Ceil(float64(p)/100*float64(len(sorted)))) - 1
	if idx < 0 {
		idx = 0
	}
	if idx >= len(sorted) {
		idx = len(sorted) - 1
	}
	return sorted[idx]
}

func avgDuration(d []time.Duration) time.Duration {
	if len(d) == 0 {
		return 0
	}
	var sum int64
	for _, v := range d {
		sum += int64(v)
	}
	return time.Duration(sum / int64(len(d)))
}

// --- HTTP helpers ---

func postJSON(url string, body interface{}) (*http.Response, error) {
	b, _ := json.Marshal(body)
	return http.Post(url, "application/json", bytes.NewReader(b))
}

// registerAndLogin creates a user and returns the access token.
func registerAndLogin(t *testing.T, email string) string {
	t.Helper()

	body, _ := json.Marshal(map[string]string{
		"email": email, "password": "PerfTest123!", "first_name": "Perf",
	})
	resp, err := http.Post(baseURL+"/api/v1/auth/register", "application/json", bytes.NewReader(body))
	if err != nil {
		t.Fatalf("register: %v", err)
	}
	resp.Body.Close()

	body, _ = json.Marshal(map[string]string{"email": email, "password": "PerfTest123!"})
	resp, err = http.Post(baseURL+"/api/v1/auth/login", "application/json", bytes.NewReader(body))
	if err != nil {
		t.Fatalf("login: %v", err)
	}
	defer resp.Body.Close()

	var authResp struct {
		Data struct {
			AccessToken string `json:"access_token"`
		} `json:"data"`
	}
	json.NewDecoder(resp.Body).Decode(&authResp)
	if authResp.Data.AccessToken == "" {
		t.Fatal("no access token from login")
	}
	return authResp.Data.AccessToken
}

func getWithAuth(url, token string) (*http.Response, error) {
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("Authorization", "Bearer "+token)
	return http.DefaultClient.Do(req)
}

func init() {
	_ = fmt.Sprintf
	_ = os.Getenv
}
