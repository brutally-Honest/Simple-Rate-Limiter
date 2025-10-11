package ratelimiter

import (
	"sync"
	"time"
)

type container struct {
	size int
}

type request struct {
	ip       string
	response chan bool
}

type LeakyBucket struct {
	mu        sync.Mutex
	buckets   map[string]*container
	threshold int
	interval  time.Duration
	requests  chan request
}

func NewLeakyBucket(threshold int, interval time.Duration) *LeakyBucket {
	lb := &LeakyBucket{
		threshold: threshold,
		interval:  interval,
		requests:  make(chan request, 100),
		buckets:   make(map[string]*container),
	}
	go lb.run()
	return lb
}

// central go routine
func (lb *LeakyBucket) run() {
	ticker := time.NewTicker(lb.interval)
	defer ticker.Stop()

	for {
		select {
		// rate limit check requests
		case req := <-lb.requests:
			bucket, exists := lb.buckets[req.ip]
			if !exists {
				bucket = &container{size: 0}
				lb.buckets[req.ip] = bucket
			}

			// Check if there's space in the bucket
			if bucket.size < lb.threshold {
				bucket.size++
				req.response <- true
			} else {
				req.response <- false
			}

		// Leak from all buckets
		case <-ticker.C:
			for _, bucket := range lb.buckets {
				if bucket.size > 0 {
					bucket.size--
				}
			}
		}
	}
}

func (lb *LeakyBucket) Allow(ip string) bool {
	response := make(chan bool, 1)

	lb.requests <- request{
		ip:       ip,
		response: response,
	}

	return <-response
}

func (lb *LeakyBucket) Stats(ip string) int {
	lb.mu.Lock()
	defer lb.mu.Unlock()

	b, ok := lb.buckets[ip]
	if !ok {
		return 0
	}
	return b.size
}
