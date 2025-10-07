package main

import (
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/brutally-Honest/simple-rate-limiter/middleware"
	"github.com/joho/godotenv"
)

func main() {
	godotenv.Load()
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080" // fallback
	}
	portString := fmt.Sprintf(":%s", port)
	fmt.Printf("Server running on :%s\n", portString)

	mux := http.NewServeMux()

	apiHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("Hit", time.Now())
		fmt.Fprintf(w, "API endpoint - Hit at %s\n", time.Now().Format(time.RFC3339))
	})

	mux.Handle("/api", middleware.RateLimitMiddleware(apiHandler))

	http.ListenAndServe(portString, mux)
}
