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

func sendRequest(client *http.Client, num int, phase string, startTime time.Time, stats *testStats) {
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

func runExploitTest() {
	client := &http.Client{Timeout: 5 * time.Second}
	stats := &testStats{}
	startTime := time.Now()

	// Phase 1: Start window with 1 request
	fmt.Println("Phase 1: Sending 1 request to start the window")
	sendRequest(client, 1, "Window Start", startTime, stats)

	// Phase 2: Hammer requests at boundary - should get 19 more (9 + 10), but non deterministic
	fmt.Println("Phase 2: Waiting for window boundary, then hammering requests at boundary")
	time.Sleep(995 * time.Millisecond)

	var wg sync.WaitGroup
	for i := 2; i <= 26; i++ {
		wg.Add(1)
		go func(num int) {
			defer wg.Done()
			sendRequest(client, num, "Boundary Hammer", startTime, stats)
		}(i)
	}
	wg.Wait()

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
}
