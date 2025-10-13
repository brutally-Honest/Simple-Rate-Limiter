package main

import (
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"
)

const (
	exploitURL = "http://localhost:1783/rate-limit"
	exploitIP  = "203.0.113.42"
)

type testStats struct {
	allowed  int
	rejected int
	mu       sync.Mutex
}

func (s *testStats) recordAllowed() {
	s.mu.Lock()
	s.allowed++
	s.mu.Unlock()
}

func (s *testStats) recordRejected() {
	s.mu.Lock()
	s.rejected++
	s.mu.Unlock()
}

func main() {
	fmt.Println("=== Fixed Window Rate Limiter Boundary Exploitation Test ===")
	fmt.Println("Configuration: 10 requests per second limit")
	fmt.Println("Target endpoint:", exploitURL)
	fmt.Println()

	runExploitTest()
}

func runExploitTest() {
	client := &http.Client{Timeout: 5 * time.Second}
	stats := &testStats{}
	startTime := time.Now()

	sendRequest := func(num int, phase string) {
		req, _ := http.NewRequest("GET", exploitURL, nil)
		req.Header.Set("X-Forwarded-For", exploitIP)

		resp, err := client.Do(req)
		if err != nil {
			fmt.Printf("  ✗ Request %d error: %v\n", num, err)
			stats.recordRejected()
			return
		}
		defer resp.Body.Close()
		io.Copy(io.Discard, resp.Body)

		elapsed := time.Since(startTime).Milliseconds()

		if resp.StatusCode == http.StatusOK {
			fmt.Printf("  ✓ Request %d allowed (200) at T+%dms [%s]\n", num, elapsed, phase)
			stats.recordAllowed()
		} else {
			fmt.Printf("  ✗ Request %d rejected (429) at T+%dms [%s]\n", num, elapsed, phase)
			stats.recordRejected()
		}
	}

	// Phase 1: Start window with 1 request
	fmt.Println("Phase 1: Sending 1 request to start the window")
	sendRequest(1, "Phase 1")
	fmt.Printf("  Total allowed: %d, rejected: %d\n\n", stats.allowed, stats.rejected)

	// Phase 2: Send 9 requests near window end concurrently
	fmt.Println("Phase 2: Waiting 950ms, then sending 9 requests at window end")
	time.Sleep(950 * time.Millisecond)

	var wg sync.WaitGroup
	for i := 2; i <= 10; i++ {
		wg.Add(1)
		go func(num int) {
			defer wg.Done()
			sendRequest(num, "Phase 2")
		}(i)
	}
	wg.Wait()
	fmt.Printf("  Total allowed: %d, rejected: %d\n\n", stats.allowed, stats.rejected)

	// Phase 3: Send 10 requests in new window concurrently
	fmt.Println("Phase 3: Waiting 100ms for window reset, then sending 10 requests")
	time.Sleep(100 * time.Millisecond)

	for i := 11; i <= 30; i++ {
		wg.Add(1)
		go func(num int) {
			defer wg.Done()
			sendRequest(num, "Phase 3")
		}(i)
	}
	wg.Wait()
	fmt.Printf("  Total allowed: %d, rejected: %d\n\n", stats.allowed, stats.rejected)

	// Print results
	printExploitResults(stats, time.Since(startTime))
}

func printExploitResults(stats *testStats, duration time.Duration) {
	fmt.Println("=== RESULTS ===")
	fmt.Printf("Total requests sent: %d\n", stats.allowed+stats.rejected)
	fmt.Printf("Requests allowed: %d\n", stats.allowed)
	fmt.Printf("Requests rejected: %d\n", stats.rejected)
	fmt.Printf("Time span: %dms (~%.1f seconds)\n", duration.Milliseconds(), duration.Seconds())
	fmt.Printf("Configured limit: 10 requests/second\n")
	fmt.Printf("Actual throughput: %.1f requests/second\n\n", float64(stats.allowed)/duration.Seconds())

	switch {
	case stats.allowed == 20:
		fmt.Println("✓ EXPLOITATION SUCCESSFUL: 2x limit (20 requests) achieved at adjacent window boundaries")
	case stats.allowed >= 19:
		fmt.Printf("⚠ Near successful: %d requests allowed (expected 20)\n", stats.allowed)
	default:
		fmt.Printf("✗ Exploitation failed: only %d requests allowed\n", stats.allowed)
	}
}
