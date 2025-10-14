package ratelimiter

import (
	"fmt"
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
		fmt.Println("Count", data.count, "Time", time.Now().Format("15:04:05.000"))
		return true
	}

	return false
}

func NewFixedWindow(limit int, window time.Duration) *FixedWindow {
	return &FixedWindow{
		limit:  limit,
		window: window,
		ips:    make(map[string]*fixedWindowData),
	}
}
