package user

import (
	"github.com/go-chi/chi/v5"
)

// RegisterRoutes registers user routes to the provided router.
func RegisterRoutes(r chi.Router, h *Handler) {
	r.Route("/users", func(r chi.Router) {
		r.Get("/", h.GetUsers)
		r.Post("/", h.CreateUser)
		r.Get("/{id}", h.GetUser)
		r.Patch("/{id}", h.UpdateUser)
		r.Delete("/{id}", h.DeleteUser)
	})
}
