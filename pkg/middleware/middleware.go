package middleware

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"go-mini-setup/pkg/response"
)

// RequestLogger returns a Chi-compatible middleware for logging requests with timing.
func RequestLogger() func(next http.Handler) http.Handler {
	return middleware.Logger
}

// Recoverer returns a panic recovery middleware that catches panics and returns 500 JSON.
func Recoverer() func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if rvr := recover(); rvr != nil {
					response.Error(w, http.StatusInternalServerError, "An unexpected server error occurred")
				}
			}()
			next.ServeHTTP(w, r)
		})
	}
}

// CORS configures permissive and modern CORS headers.
func CORS() func(next http.Handler) http.Handler {
	return cors.Handler(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: false,
		MaxAge:           300,
	})
}

// JSONContentType sets default response Content-Type to application/json.
func JSONContentType(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		next.ServeHTTP(w, r)
	})
}

// Timeout sets request context timeout.
func Timeout(timeout time.Duration) func(next http.Handler) http.Handler {
	return middleware.Timeout(timeout)
}
