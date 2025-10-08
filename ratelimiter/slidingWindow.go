package ratelimiter

import (
	"sync"
	"time"
)

type SlidingWindow struct {
	logs   map[string][]time.Time
	limit  int
	window time.Duration
	mu     sync.Mutex
}

func NewSlidingWindow(limit int, window time.Duration) *SlidingWindow {
	return &SlidingWindow{
		limit:  limit,
		window: window,
		logs:   make(map[string][]time.Time),
	}
}

func (sw *SlidingWindow) Allow(ip string) bool {
	sw.mu.Lock()
	defer sw.mu.Unlock()

	now := time.Now()
	currentWindow := now.Add(-sw.window)

	if logs, exists := sw.logs[ip]; exists {
		valid := []time.Time{}
		for _, ts := range logs {
			if ts.After(currentWindow) {
				valid = append(valid, ts)
			}
		}
		sw.logs[ip] = valid
	}

	if len(sw.logs[ip]) < sw.limit {
		sw.logs[ip] = append(sw.logs[ip], now)
		return true
	}

	return false

}

func (sw *SlidingWindow) Stats(ip string) int {
	return len(sw.logs[ip])
}
