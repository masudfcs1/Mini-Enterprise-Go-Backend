package auth

import (
	"github.com/go-chi/chi/v5"
)

// RegisterRoutes registers auth routes to the provided router.
func RegisterRoutes(r chi.Router, h *Handler) {
	r.Route("/auth", func(r chi.Router) {
		r.Post("/register", h.Register)
		r.Post("/login", h.Login)
	})
}
