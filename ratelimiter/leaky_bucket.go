package ratelimiter

import (
	"context"
	"time"
)

type lbContainer struct {
	size        int
	lastRequest time.Time
}

type request struct {
	ip       string
	response chan bool
}

type LeakyBucket struct {
	buckets   map[string]*lbContainer
	threshold int
	interval  time.Duration
	requests  chan request

	ctx    context.Context
	cancel context.CancelFunc
}

func (lb *LeakyBucket) run() {
	ticker := time.NewTicker(lb.interval)
	defer ticker.Stop()

	for {
		now := time.Now()
		select {
		// rate limit check requests
		case req := <-lb.requests:
			bucket, exists := lb.buckets[req.ip]
			if !exists {
				bucket = &lbContainer{size: 0, lastRequest: now}
				lb.buckets[req.ip] = bucket
			}

			// Check if there's space in the bucket
			if bucket.size < lb.threshold {
				bucket.size++
				bucket.lastRequest = now
				req.response <- true
			} else {
				req.response <- false
			}

		// Leak from all buckets
		case <-ticker.C:
			for ip, bucket := range lb.buckets {
				if bucket.size > 0 {
					bucket.size--
				}

				if time.Since(bucket.lastRequest) > 5*time.Minute {
					delete(lb.buckets, ip)
				}
			}

		case <-lb.ctx.Done():
			return
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

func (lb *LeakyBucket) Close() {
	lb.cancel()
}

func NewLeakyBucket(threshold int, interval time.Duration) *LeakyBucket {
	ctx, cancel := context.WithCancel(context.Background())
	lb := &LeakyBucket{
		threshold: threshold,
		interval:  interval,
		requests:  make(chan request, 100), // default buffer of 100
		buckets:   make(map[string]*lbContainer),
		ctx:       ctx,
		cancel:    cancel,
	}
	go lb.run()
	return lb
}
