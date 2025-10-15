package ratelimiter

import (
	"context"
	"sync"
	"time"
)

type fixedWindowData struct {
	count       int
	windowStart time.Time
}

type FixedWindow struct {
	limit  int
	window time.Duration
	ips    map[string]*fixedWindowData
	mu     sync.Mutex

	ctx    context.Context
	cancel context.CancelFunc
}

func (f *FixedWindow) Allow(ip string) bool {
	f.mu.Lock()
	defer f.mu.Unlock()

	data, exists := f.ips[ip]
	if !exists || time.Since(data.windowStart) >= f.window {
		f.ips[ip] = &fixedWindowData{
			count:       1,
			windowStart: time.Now(),
		}
		return true
	}

	if data.count < f.limit {
		data.count++
		return true
	}

	return false
}

// clear ips
func (f *FixedWindow) clean() {
	f.mu.Lock()
	defer f.mu.Unlock()

	for ip, data := range f.ips {
		if time.Since(data.windowStart) >= f.window {
			delete(f.ips, ip)
		}
	}
}

func (f *FixedWindow) cleanAll() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-f.ctx.Done():
			return
		case <-ticker.C:
			f.clean()
		}
	}
}

// Graceful shutdown
func (f *FixedWindow) Close() {
	f.cancel()
}

func NewFixedWindow(limit int, window time.Duration) *FixedWindow {
	ctx, cancel := context.WithCancel(context.Background())
	fw := &FixedWindow{
		limit:  limit,
		window: window,
		ips:    make(map[string]*fixedWindowData),
		ctx:    ctx,
		cancel: cancel,
	}
	go fw.cleanAll()
	return fw
}
