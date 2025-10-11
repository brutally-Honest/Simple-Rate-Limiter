package ratelimiter

import (
	"sync"
	"time"
)

type container struct {
	size int
}

type LeakyBucket struct {
	mu        sync.Mutex
	buckets   map[string]*container
	threshold int
	interval  time.Duration
	requests  chan string
}

// central go routine to leak tokens and handle requests
func (lb *LeakyBucket) run() {
	ticker := time.NewTicker(lb.interval)
	for {
		select {
		case ip := <-lb.requests:
			lb.handleRequest(ip)

		case <-ticker.C:
			lb.leakAll()
		}
	}
}

func (lb *LeakyBucket) handleRequest(ip string) {
	lb.mu.Lock()
	defer lb.mu.Unlock()

	b, ok := lb.buckets[ip]
	if !ok {
		b = &container{}
		lb.buckets[ip] = b
	}

	if b.size < lb.threshold {
		b.size++
	}
}

func (lb *LeakyBucket) leakAll() {
	lb.mu.Lock()
	defer lb.mu.Unlock()

	for _, b := range lb.buckets {
		if b.size > 0 {
			b.size--
		}
	}
}

func NewLeakyBucket(threshold int, interval time.Duration) *LeakyBucket {
	lb := &LeakyBucket{
		interval:  interval,
		threshold: threshold,
		requests:  make(chan string, 1000),
		buckets:   make(map[string]*container),
	}
	go lb.run()
	return lb
}

func (lb *LeakyBucket) Allow(ip string) bool {
	select {
	case lb.requests <- ip:
		return true
	default:
		return false
	}
}

func (lb *LeakyBucket) Stats(ip string) int {

	lb.mu.Lock()
	defer lb.mu.Unlock()

	b, ok := lb.buckets[ip]
	if ok {
		return b.size
	}
	return 0
}
