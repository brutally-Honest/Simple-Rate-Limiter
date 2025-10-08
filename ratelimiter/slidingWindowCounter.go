package ratelimiter

import (
	"sync"
	"time"
)

type window struct {
	prevCount    int
	currCount    int
	currentStart time.Time
}

type SlidindWindowCounter struct {
	mu      sync.Mutex
	limit   int
	window  time.Duration
	windows map[string]*window
}

func NewSlidingWindowCounter(limit int, windowDuration time.Duration) *SlidindWindowCounter {
	return &SlidindWindowCounter{
		limit:   limit,
		window:  windowDuration,
		windows: make(map[string]*window),
	}
}

func (swc *SlidindWindowCounter) Allow(ip string) bool {
	swc.mu.Lock()
	defer swc.mu.Unlock()

	now := time.Now()

	data, exists := swc.windows[ip]
	if !exists {
		swc.windows[ip] = &window{
			currCount:    1,
			prevCount:    0,
			currentStart: now,
		}
		return true
	}

	elapsed := time.Since(data.currentStart)
	if elapsed >= swc.window {
		data.prevCount = data.currCount
		data.currCount = 1
		data.currentStart = now
		return true
	}

	// Calculate weighted count from previous + current window
	prevWeight := float64(swc.window-elapsed) / float64(swc.window)
	estimatedCount := int(float64(data.prevCount)*prevWeight) + data.currCount

	if estimatedCount < swc.limit {
		data.currCount++
		return true
	}

	return false
}

func (swc *SlidindWindowCounter) Stats(ip string) int {
	data, exists := swc.windows[ip]
	if !exists {
		return 1
	}
	elapsed := time.Since(data.currentStart)
	if elapsed >= swc.window {
		return 1
	}
	prevWeight := float64(swc.window-elapsed) / float64(swc.window)
	estimatedCount := int(float64(data.prevCount)*prevWeight) + data.currCount
	return estimatedCount
}
