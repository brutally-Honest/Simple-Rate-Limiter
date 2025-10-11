package main

import (
	"fmt"
	"net"
	"net/http"
	"os"
	"time"

	"github.com/brutally-Honest/simple-rate-limiter/middleware"
	"github.com/brutally-Honest/simple-rate-limiter/ratelimiter"
	"github.com/joho/godotenv"
)

func main() {
	godotenv.Load()
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	addr := fmt.Sprintf(":%s", port)
	fmt.Printf("Server running on :%s\n", addr)

	mux := http.NewServeMux()

	// strategy := ratelimiter.NewFixedWindow(10, 1*time.Second)
	// strategy := ratelimiter.NewSlidingWindow(10, 1*time.Second)
	// strategy := ratelimiter.NewSlidingWindowCounter(10, 1*time.Second)
	// strategy := ratelimiter.NewTokenBucket(10, 0.5)
	strategy := ratelimiter.NewLeakyBucket(10, 10*time.Second)

	rateLimiter := middleware.NewRateLimitMiddleware(strategy)
	limitHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("Hit", time.Now())
		fmt.Fprintf(w, "API endpoint - Hit at %s\n", time.Now().Format(time.RFC3339))
	})

	statsHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip, _, _ := net.SplitHostPort(r.RemoteAddr)
		if ip == "" {
			ip = r.RemoteAddr
		}
		count := strategy.Stats(ip)
		w.WriteHeader(http.StatusAccepted)
		fmt.Fprintf(w, "Requests from %s: %d\n", ip, count)
	})

	mux.Handle("/rate-limit", rateLimiter.Wrap(limitHandler))
	mux.Handle("/stats", statsHandler)

	http.ListenAndServe(addr, mux)
}
