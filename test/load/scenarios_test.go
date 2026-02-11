// Package load provides load testing utilities for the Stratavore HTTP API.
//
// Run a specific scenario:
//
//	go test ./test/load/... -run TestConcurrentLaunches -v -load-addr http://localhost:50049
//
// Run all scenarios:
//
//	go test ./test/load/... -v -load-addr http://localhost:50049
//
// Environment / flags:
//
//	-load-addr      Base URL of the stratavored HTTP API (default http://localhost:50049)
//	-load-runners   Target runner count for runner-count scenarios (default 100)
//	-load-events    Target events/s for event throughput scenarios (default 1000)
package load

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// ─── Flags (registered once; parsed by TestMain or individual tests) ─────────

var (
	daemonAddr   = flag.String("load-addr", "http://localhost:50049", "Base URL of stratavored HTTP API")
	targetRunners = flag.Int("load-runners", 100, "Concurrent runner count for load tests")
	targetEvents  = flag.Int("load-events", 1000, "Target events/s for throughput tests")
)

// ─── Helpers ─────────────────────────────────────────────────────────────────

func baseURL() string { return *daemonAddr }

func postJSON(path string, body interface{}) (*http.Response, error) {
	data, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}
	resp, err := http.Post(baseURL()+path, "application/json", bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("POST %s: %w", path, err)
	}
	return resp, nil
}

func drainClose(resp *http.Response) {
	if resp != nil {
		io.Copy(io.Discard, resp.Body)
		resp.Body.Close()
	}
}

// Result captures timing for a single request.
type Result struct {
	Latency    time.Duration
	StatusCode int
	Err        error
}

// Stats summarises a slice of Results.
type Stats struct {
	Total     int
	Success   int
	Errors    int
	MinMS     float64
	MaxMS     float64
	MeanMS    float64
	P50MS     float64
	P95MS     float64
	P99MS     float64
	ErrorRate float64 // 0–1
}

func summarise(results []Result) Stats {
	s := Stats{Total: len(results)}
	if s.Total == 0 {
		return s
	}

	latencies := make([]float64, 0, s.Total)
	for _, r := range results {
		ms := float64(r.Latency.Milliseconds())
		latencies = append(latencies, ms)
		if r.Err != nil || r.StatusCode >= 400 {
			s.Errors++
		} else {
			s.Success++
		}
	}

	// Sort (insertion sort is fine for <10k entries)
	for i := 1; i < len(latencies); i++ {
		for j := i; j > 0 && latencies[j] < latencies[j-1]; j-- {
			latencies[j], latencies[j-1] = latencies[j-1], latencies[j]
		}
	}

	s.MinMS = latencies[0]
	s.MaxMS = latencies[len(latencies)-1]
	var total float64
	for _, l := range latencies {
		total += l
	}
	s.MeanMS = total / float64(len(latencies))
	s.P50MS = latencies[len(latencies)*50/100]
	s.P95MS = latencies[len(latencies)*95/100]
	s.P99MS = latencies[len(latencies)*99/100]
	s.ErrorRate = float64(s.Errors) / float64(s.Total)
	return s
}

func logStats(t *testing.T, label string, s Stats) {
	t.Helper()
	t.Logf("[%s] total=%d ok=%d errors=%d error_rate=%.1f%%",
		label, s.Total, s.Success, s.Errors, s.ErrorRate*100)
	t.Logf("[%s] latency ms: min=%.1f mean=%.1f p50=%.1f p95=%.1f p99=%.1f max=%.1f",
		label, s.MinMS, s.MeanMS, s.P50MS, s.P95MS, s.P99MS, s.MaxMS)
}

// checkDaemon pings /api/v1/health and skips the test if the daemon is not up.
func checkDaemon(t *testing.T) {
	t.Helper()
	resp, err := http.Get(baseURL() + "/api/v1/health")
	if err != nil {
		t.Skipf("stratavored not reachable at %s: %v", baseURL(), err)
	}
	drainClose(resp)
	if resp.StatusCode != http.StatusOK {
		t.Skipf("stratavored health check failed: status %d", resp.StatusCode)
	}
}

// ─── Scenario 1: Concurrent Runner Launches ──────────────────────────────────

// TestConcurrentLaunches fires N simultaneous LaunchRunner requests and
// asserts that the daemon handles them without errors and p99 latency stays
// under 5 seconds.
func TestConcurrentLaunches(t *testing.T) {
	checkDaemon(t)

	concurrency := *targetRunners
	t.Logf("launching %d runners concurrently", concurrency)

	results := make([]Result, concurrency)
	var wg sync.WaitGroup

	for i := 0; i < concurrency; i++ {
		i := i
		wg.Add(1)
		go func() {
			defer wg.Done()
			start := time.Now()
			resp, err := postJSON("/api/v1/runners/launch", map[string]interface{}{
				"project_name":      fmt.Sprintf("load-test-project-%d", i%10),
				"project_path":      "/tmp/load-test",
				"runtime_type":      "process",
				"conversation_mode": "new",
				"flags":             []string{"--dangerously-skip-permissions"},
			})
			results[i] = Result{
				Latency:    time.Since(start),
				StatusCode: statusOf(resp),
				Err:        err,
			}
			drainClose(resp)
		}()
	}
	wg.Wait()

	s := summarise(results)
	logStats(t, "concurrent_launches", s)

	if s.ErrorRate > 0.01 {
		t.Errorf("error rate %.1f%% exceeds 1%% threshold", s.ErrorRate*100)
	}
	if s.P99MS > 5000 {
		t.Errorf("p99 latency %.1fms exceeds 5000ms threshold", s.P99MS)
	}
}

// ─── Scenario 2: Heartbeat Throughput ────────────────────────────────────────

// TestHeartbeatThroughput sends heartbeats at the target events/s rate for
// 30 seconds and measures sustained throughput.
func TestHeartbeatThroughput(t *testing.T) {
	checkDaemon(t)

	targetEPS := *targetEvents
	duration := 30 * time.Second
	t.Logf("heartbeat throughput: %d events/s for %s", targetEPS, duration)

	var (
		sent    int64
		errored int64
		results []Result
		mu      sync.Mutex
	)

	// Determine worker count (each worker sends 1 req at a time)
	workers := targetEPS / 10 // ~10 req/s per worker
	if workers < 1 {
		workers = 1
	}

	ctx := make(chan struct{})
	var wg sync.WaitGroup
	start := time.Now()

	// Limiter: allow targetEPS requests per second globally
	ticker := time.NewTicker(time.Second / time.Duration(targetEPS))
	defer ticker.Stop()

	for w := 0; w < workers; w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				select {
				case <-ctx:
					return
				case <-ticker.C:
					t0 := time.Now()
					resp, err := postJSON("/api/v1/heartbeat", map[string]interface{}{
						"runner_id":     fmt.Sprintf("load-runner-%d", rand.Intn(1000)),
						"status":        "running",
						"cpu_percent":   rand.Float64() * 100,
						"memory_mb":     rand.Int63n(512),
						"agent_version": "1.4.0",
					})
					r := Result{
						Latency:    time.Since(t0),
						StatusCode: statusOf(resp),
						Err:        err,
					}
					drainClose(resp)
					if err != nil || r.StatusCode >= 400 {
						atomic.AddInt64(&errored, 1)
					}
					atomic.AddInt64(&sent, 1)
					mu.Lock()
					results = append(results, r)
					mu.Unlock()
				}
			}
		}()
	}

	time.Sleep(duration)
	close(ctx)
	wg.Wait()

	elapsed := time.Since(start).Seconds()
	actualEPS := float64(atomic.LoadInt64(&sent)) / elapsed
	errRate := float64(atomic.LoadInt64(&errored)) / float64(atomic.LoadInt64(&sent))

	t.Logf("sent=%d elapsed=%.1fs actual_eps=%.1f error_rate=%.2f%%",
		sent, elapsed, actualEPS, errRate*100)

	mu.Lock()
	s := summarise(results)
	mu.Unlock()
	logStats(t, "heartbeat_throughput", s)

	if errRate > 0.01 {
		t.Errorf("error rate %.2f%% exceeds 1%% threshold", errRate*100)
	}
	if s.P95MS > 500 {
		t.Errorf("p95 heartbeat latency %.1fms exceeds 500ms threshold", s.P95MS)
	}
}

// ─── Scenario 3: Status Endpoint Under Load ──────────────────────────────────

// TestStatusUnderLoad queries /api/v1/status repeatedly from multiple
// goroutines to measure read-heavy performance.
func TestStatusUnderLoad(t *testing.T) {
	checkDaemon(t)

	concurrency := 50
	requestsEach := 100
	total := concurrency * requestsEach
	t.Logf("status under load: %d goroutines × %d requests = %d total", concurrency, requestsEach, total)

	results := make([]Result, total)
	var idx int64
	var wg sync.WaitGroup

	for w := 0; w < concurrency; w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < requestsEach; j++ {
				t0 := time.Now()
				resp, err := http.Get(baseURL() + "/api/v1/status")
				i := int(atomic.AddInt64(&idx, 1)) - 1
				results[i] = Result{
					Latency:    time.Since(t0),
					StatusCode: statusOf(resp),
					Err:        err,
				}
				drainClose(resp)
			}
		}()
	}
	wg.Wait()

	s := summarise(results)
	logStats(t, "status_under_load", s)

	if s.ErrorRate > 0.001 {
		t.Errorf("error rate %.2f%% exceeds 0.1%% threshold", s.ErrorRate*100)
	}
	if s.P99MS > 1000 {
		t.Errorf("p99 latency %.1fms exceeds 1000ms threshold", s.P99MS)
	}
}

// ─── Scenario 4: Mixed Workload ───────────────────────────────────────────────

// TestMixedWorkload simulates realistic concurrent API traffic: launches,
// heartbeats, status checks, and list operations.
func TestMixedWorkload(t *testing.T) {
	checkDaemon(t)

	duration := 20 * time.Second
	workers := 20
	t.Logf("mixed workload: %d workers for %s", workers, duration)

	type op struct {
		name string
		fn   func() Result
	}

	ops := []op{
		{"status", func() Result {
			t0 := time.Now()
			resp, err := http.Get(baseURL() + "/api/v1/status")
			r := Result{Latency: time.Since(t0), StatusCode: statusOf(resp), Err: err}
			drainClose(resp)
			return r
		}},
		{"list_runners", func() Result {
			t0 := time.Now()
			resp, err := http.Get(baseURL() + "/api/v1/runners/list")
			r := Result{Latency: time.Since(t0), StatusCode: statusOf(resp), Err: err}
			drainClose(resp)
			return r
		}},
		{"list_projects", func() Result {
			t0 := time.Now()
			resp, err := http.Get(baseURL() + "/api/v1/projects/list")
			r := Result{Latency: time.Since(t0), StatusCode: statusOf(resp), Err: err}
			drainClose(resp)
			return r
		}},
		{"heartbeat", func() Result {
			t0 := time.Now()
			resp, err := postJSON("/api/v1/heartbeat", map[string]interface{}{
				"runner_id":     fmt.Sprintf("mix-runner-%d", rand.Intn(50)),
				"status":        "running",
				"agent_version": "1.4.0",
			})
			r := Result{Latency: time.Since(t0), StatusCode: statusOf(resp), Err: err}
			drainClose(resp)
			return r
		}},
	}

	allResults := make([]Result, 0, workers*200)
	var mu sync.Mutex
	var wg sync.WaitGroup
	done := make(chan struct{})

	for w := 0; w < workers; w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				select {
				case <-done:
					return
				default:
					o := ops[rand.Intn(len(ops))]
					r := o.fn()
					mu.Lock()
					allResults = append(allResults, r)
					mu.Unlock()
					time.Sleep(time.Duration(rand.Intn(50)) * time.Millisecond)
				}
			}
		}()
	}

	time.Sleep(duration)
	close(done)
	wg.Wait()

	s := summarise(allResults)
	logStats(t, "mixed_workload", s)

	if s.ErrorRate > 0.02 {
		t.Errorf("error rate %.1f%% exceeds 2%% threshold", s.ErrorRate*100)
	}
}

// statusOf safely extracts the status code from an *http.Response that may be nil.
func statusOf(resp *http.Response) int {
	if resp == nil {
		return 0
	}
	return resp.StatusCode
}
