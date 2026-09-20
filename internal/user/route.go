package user

import (
	"github.com/go-chi/chi/v5"
	"go-mini-setup/pkg/middleware"
)

// RegisterRoutes registers user routes to the provided router with RBAC guards.
func RegisterRoutes(r chi.Router, h *Handler, jwtSecret, jwtIssuer, jwtAudience string) {
	r.Route("/users", func(r chi.Router) {
		r.Get("/", h.GetUsers)
		r.Post("/", h.CreateUser)
		r.Get("/{id}", h.GetUser)
		r.Patch("/{id}", h.UpdateUser)

		// Sensitive administrative action: Delete user requires ADMIN or SUPER_ADMIN role
		r.With(
			middleware.RequireAuth(jwtSecret, jwtIssuer, jwtAudience),
			middleware.RequireRole("ADMIN", "SUPER_ADMIN"),
		).Delete("/{id}", h.DeleteUser)
	})
}
