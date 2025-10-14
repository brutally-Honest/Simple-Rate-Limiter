package ratelimiter

import (
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

func NewTokenBucket(capacity int, refillRate float64) *TokenBucket {
	return &TokenBucket{
		capacity:   capacity,
		refillRate: refillRate,
		buckets:    make(map[string]*tBucket),
	}
}
