package ratelimiter

import (
	"context"
	"sync"
	"time"
)

type swcwindow struct {
	prevCount    int
	currCount    int
	currentStart time.Time
}

type SlidingWindowCounter struct {
	mu      sync.Mutex
	limit   int
	window  time.Duration
	windows map[string]*swcwindow

	ctx    context.Context
	cancel context.CancelFunc
}

func (swc *SlidingWindowCounter) Allow(ip string) bool {
	swc.mu.Lock()
	defer swc.mu.Unlock()

	now := time.Now()

	data, exists := swc.windows[ip]
	if !exists {
		swc.windows[ip] = &swcwindow{
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

func (swc *SlidingWindowCounter) cleanAll() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-swc.ctx.Done():
			return
		case <-ticker.C:
			swc.clean()
		}
	}
}

func (swc *SlidingWindowCounter) clean() {
	swc.mu.Lock()
	defer swc.mu.Unlock()

	for ip, data := range swc.windows {
		if time.Since(data.currentStart) >= swc.window {
			delete(swc.windows, ip)
		}
	}
}

func (swc *SlidingWindowCounter) Close() {
	swc.cancel()
}

func NewSlidingWindowCounter(limit int, windowDuration time.Duration) *SlidingWindowCounter {
	ctx, cancel := context.WithCancel(context.Background())
	swc := &SlidingWindowCounter{
		limit:   limit,
		window:  windowDuration,
		windows: make(map[string]*swcwindow),
		ctx:     ctx,
		cancel:  cancel,
	}
	go swc.cleanAll()
	return swc
}
