package middleware

import (
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
		ip, _, err := net.SplitHostPort(r.RemoteAddr)
		if err != nil || ip == "" {
			ip = r.RemoteAddr
		}
		allowed := rl.strategy.Allow(ip)
		if !allowed {
			http.Error(w, "Rate limit exceeded", http.StatusTooManyRequests)
			return
		}
		next.ServeHTTP(w, r)
	})
}
