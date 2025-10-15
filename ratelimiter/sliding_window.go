package ratelimiter

import (
	"context"
	"sync"
	"time"
)

type SlidingWindow struct {
	logs   map[string][]time.Time
	limit  int
	window time.Duration
	mu     sync.Mutex

	ctx    context.Context
	cancel context.CancelFunc
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

func (sw *SlidingWindow) clean() {
	sw.mu.Lock()
	defer sw.mu.Unlock()

	for ip, data := range sw.logs {
		if len(data) > 0 && time.Since(data[0]) >= sw.window {
			delete(sw.logs, ip)
		}
	}
}

func (sw *SlidingWindow) cleanAll() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-sw.ctx.Done():
			return
		case <-ticker.C:
			sw.clean()
		}
	}
}

func (sw *SlidingWindow) Close() {
	sw.cancel()
}

func NewSlidingWindow(limit int, window time.Duration) *SlidingWindow {
	ctx, cancel := context.WithCancel(context.Background())
	sw := &SlidingWindow{
		limit:  limit,
		window: window,
		logs:   make(map[string][]time.Time),
		ctx:    ctx,
		cancel: cancel,
	}
	go sw.cleanAll()
	return sw
}
