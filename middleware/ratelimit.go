package middleware

import (
	"fmt"
	"net/http"
)

func RateLimitMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("Middleware hit", r.Method, r.RemoteAddr)
		next.ServeHTTP(w, r)
	})
}
