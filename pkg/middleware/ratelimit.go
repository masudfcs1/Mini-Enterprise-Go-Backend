package middleware

import (
	"net/http"
	"time"

	"github.com/go-chi/httprate"
	"go-mini-setup/pkg/response"
)

// RateLimiter returns an IP-based rate limiting middleware with custom JSON 429 response.
func RateLimiter(requests int, window time.Duration) func(next http.Handler) http.Handler {
	return httprate.Limit(
		requests,
		window,
		httprate.WithKeyFuncs(httprate.KeyByIP),
		httprate.WithLimitHandler(func(w http.ResponseWriter, r *http.Request) {
			response.Error(w, http.StatusTooManyRequests, "Too many requests. Please wait and try again later.")
		}),
	)
}

// GlobalRateLimiter configures 100 requests per minute across standard endpoints.
func GlobalRateLimiter() func(next http.Handler) http.Handler {
	return RateLimiter(100, 1*time.Minute)
}

// StrictAuthRateLimiter configures 10 requests per minute for auth endpoints to prevent brute-force attacks.
func StrictAuthRateLimiter() func(next http.Handler) http.Handler {
	return RateLimiter(10, 1*time.Minute)
}
