package ratelimiter

import (
	"context"
	"sync"
	"time"
)

type tBucket struct {
	lastRefill time.Time
	tokens     float64
}

type TokenBucket struct {
	mu         sync.Mutex
	capacity   int
	refillRate float64
	buckets    map[string]*tBucket

	ctx    context.Context
	cancel context.CancelFunc
}

// Push based
func (tb *TokenBucket) Allow(ip string) bool {
	tb.mu.Lock()
	defer tb.mu.Unlock()

	now := time.Now()
	data, exists := tb.buckets[ip]
	if !exists {
		data = &tBucket{
			tokens:     float64(tb.capacity),
			lastRefill: now,
		}
		tb.buckets[ip] = data
	}

	elapsed := now.Sub(data.lastRefill).Seconds()
	data.tokens += elapsed * tb.refillRate
	if data.tokens > float64(tb.capacity) {
		data.tokens = float64(tb.capacity)
	}
	data.lastRefill = now

	if data.tokens >= 1 {
		data.tokens -= 1
		return true
	}
	return false
}
func (tb *TokenBucket) clean() {
	tb.mu.Lock()
	defer tb.mu.Unlock()

	for ip, data := range tb.buckets {
		if time.Since(data.lastRefill) >= 5*time.Minute {
			delete(tb.buckets, ip)
		}
	}
}

func (tb *TokenBucket) cleanAll() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-tb.ctx.Done():
			return
		case <-ticker.C:
			tb.clean()
		}
	}
}

func (tb *TokenBucket) Close() {
	tb.cancel()
}

func NewTokenBucket(capacity int, refillRate float64) *TokenBucket {
	ctx, cancel := context.WithCancel(context.Background())
	tb := &TokenBucket{
		capacity:   capacity,
		refillRate: refillRate,
		buckets:    make(map[string]*tBucket),
		ctx:        ctx,
		cancel:     cancel,
	}
	go tb.cleanAll()
	return tb
}
