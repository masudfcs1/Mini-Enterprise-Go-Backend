package router

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"

	"go-mini-setup/internal/auth"
	"go-mini-setup/internal/user"
	"go-mini-setup/pkg/middleware"
	"go-mini-setup/pkg/response"
)

// Handlers collects all module handlers for injection into the router.
type Handlers struct {
	User *user.Handler
	Auth *auth.Handler
}

// NewRouter constructs the global application router with all middlewares and mounted modules.
func NewRouter(h *Handlers) http.Handler {
	r := chi.NewRouter()

	// Global Core Middlewares
	r.Use(chimiddleware.RequestID)
	r.Use(chimiddleware.RealIP)
	r.Use(middleware.RequestLogger())
	r.Use(middleware.Recoverer())
	r.Use(middleware.CORS())
	r.Use(middleware.Timeout(60 * time.Second))

	// Base Health Check
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		response.Success(w, http.StatusOK, "Service is healthy", map[string]interface{}{
			"status":    "UP",
			"timestamp": time.Now().UTC(),
			"version":   "1.0.0",
		})
	})

	// API v1 Namespace
	r.Route("/api/v1", func(r chi.Router) {
		if h.Auth != nil {
			auth.RegisterRoutes(r, h.Auth)
		}
		if h.User != nil {
			user.RegisterRoutes(r, h.User)
		}
	})

	// Custom 404 Not Found JSON Handler
	r.NotFound(func(w http.ResponseWriter, r *http.Request) {
		response.Error(w, http.StatusNotFound, "Route not found")
	})

	// Custom 405 Method Not Allowed JSON Handler
	r.MethodNotAllowed(func(w http.ResponseWriter, r *http.Request) {
		response.Error(w, http.StatusMethodNotAllowed, "Method not allowed")
	})

	return r
}
