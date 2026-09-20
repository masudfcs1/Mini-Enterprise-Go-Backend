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

// Handlers collects all module handlers and global configurations for injection into the router.
type Handlers struct {
	User            *user.Handler
	Auth            *auth.Handler
	JWTAccessSecret string
	JWTIssuer       string
	JWTAudience     string
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
	r.Use(middleware.GlobalRateLimiter())

	// Root Welcome / Directory Index
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		response.Success(w, http.StatusOK, "Welcome to Mini Enterprise Go Backend API", map[string]interface{}{
			"architecture": "Modular Layered (Chi + Prisma ORM + PostgreSQL)",
			"features": []string{
				"Dual-Token Auth (Access 15m + Refresh 7d/30d)",
				"Prisma RBAC (USER, ADMIN, SUPER_ADMIN)",
				"Cookie & Bearer Token Transports",
				"Pino-style Colorful Logger",
				"IP Rate Limiting",
				"PgBouncer Pooler Support",
			},
			"version": "2.0.0",
			"endpoints": map[string]string{
				"health":        "GET /health",
				"users":         "GET /api/v1/users",
				"create_user":   "POST /api/v1/users",
				"delete_user":   "DELETE /api/v1/users/{id} (Requires ADMIN / SUPER_ADMIN)",
				"auth_register": "POST /api/v1/auth/register",
				"auth_login":    "POST /api/v1/auth/login",
				"auth_refresh":  "POST /api/v1/auth/refresh",
				"auth_logout":   "POST /api/v1/auth/logout",
				"auth_me":       "GET /api/v1/auth/me (Bearer token or cookie required)",
			},
		})
	})

	// Base Health Check
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		response.Success(w, http.StatusOK, "Service is healthy", map[string]interface{}{
			"status":    "UP",
			"timestamp": time.Now().UTC(),
			"version":   "2.0.0",
		})
	})

	// API v1 Namespace
	r.Route("/api/v1", func(r chi.Router) {
		if h.Auth != nil {
			auth.RegisterRoutes(r, h.Auth, h.JWTAccessSecret, h.JWTIssuer, h.JWTAudience)
		}
		if h.User != nil {
			user.RegisterRoutes(r, h.User, h.JWTAccessSecret, h.JWTIssuer, h.JWTAudience)
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
