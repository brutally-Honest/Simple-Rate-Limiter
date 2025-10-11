package middleware

import (
	"fmt"
	"net"
	"net/http"

	"github.com/brutally-Honest/simple-rate-limiter/ratelimiter"
)

type RateLimitMiddleware struct {
	strategy ratelimiter.Strategy
}

func NewRateLimitMiddleware(strategy ratelimiter.Strategy) *RateLimitMiddleware {
	return &RateLimitMiddleware{
		strategy: strategy,
	}
}

func (rl *RateLimitMiddleware) Wrap(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip, _, _ := net.SplitHostPort(r.RemoteAddr)
		fmt.Println("IP", ip)
		if ip == "" {
			ip = r.RemoteAddr
		}
		allowed := rl.strategy.Allow(ip)
		if !allowed {
			fmt.Println("Rate Limit exceeded")
			http.Error(w, "Rate limit exceeded", http.StatusTooManyRequests)
			return
		}
		next.ServeHTTP(w, r)
	})
}
