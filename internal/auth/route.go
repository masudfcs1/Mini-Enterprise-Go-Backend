package auth

import (
	"github.com/go-chi/chi/v5"
	"go-mini-setup/pkg/middleware"
)

// RegisterRoutes registers auth routes to the provided router.
func RegisterRoutes(r chi.Router, h *Handler, jwtSecret, jwtIssuer, jwtAudience string) {
	r.Route("/auth", func(r chi.Router) {
		// Strict rate limiting for sensitive credential & token endpoints
		r.Group(func(r chi.Router) {
			r.Use(middleware.StrictAuthRateLimiter())
			r.Post("/register", h.Register)
			r.Post("/login", h.Login)
			r.Post("/refresh", h.Refresh)
		})

		// Logout endpoint
		r.Post("/logout", h.Logout)

		// Protected endpoints requiring a valid JWT Access token
		r.Group(func(r chi.Router) {
			r.Use(middleware.RequireAuth(jwtSecret, jwtIssuer, jwtAudience))
			r.Get("/me", h.GetMe)
		})
	})
}
