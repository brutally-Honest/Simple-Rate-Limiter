package main

import (
	"log"
	"net/http"
	"os"
	"time"

	"github.com/brutally-Honest/simple-rate-limiter/middleware"
	"github.com/brutally-Honest/simple-rate-limiter/ratelimiter"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "1783"
	}

	strategy := ratelimiter.NewFixedWindow(10, 1*time.Second)
	// strategy := ratelimiter.NewSlidingWindow(10, 1*time.Second)
	// strategy := ratelimiter.NewSlidingWindowCounter(10, 1*time.Second)
	// strategy := ratelimiter.NewTokenBucket(10, 0.5)
	// strategy := ratelimiter.NewLeakyBucket(10, 10*time.Second)

	rateLimiter := middleware.NewRateLimitMiddleware(strategy)
	limitHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("API endpoint - Hit\n")
	})

	mux := http.NewServeMux()
	mux.Handle("/rate-limit", rateLimiter.Wrap(limitHandler))

	server := &http.Server{
		Addr:    ":" + port,
		Handler: mux,
	}
	log.Printf("Server running on %s", port)

	err := server.ListenAndServe()
	if err != nil {
		log.Fatalf("Server error: %v", err)
	}
}
