package auth

import (
	"github.com/go-chi/chi/v5"
	"go-mini-setup/pkg/middleware"
)

// RegisterRoutes registers auth routes to the provided router.
func RegisterRoutes(r chi.Router, h *Handler, jwtSecret string) {
	r.Route("/auth", func(r chi.Router) {
		// Strict rate limiting for login & register to prevent brute-force attacks
		r.Group(func(r chi.Router) {
			r.Use(middleware.StrictAuthRateLimiter())
			r.Post("/register", h.Register)
			r.Post("/login", h.Login)
		})

		// Protected endpoints requiring a valid JWT Bearer token
		r.Group(func(r chi.Router) {
			r.Use(middleware.RequireAuth(jwtSecret))
			r.Get("/me", h.GetMe)
		})
	})
}
