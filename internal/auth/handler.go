package auth

import (
	"encoding/json"
	"net/http"

	"go-mini-setup/pkg/errors"
	"go-mini-setup/pkg/middleware"
	"go-mini-setup/pkg/response"
)

// Handler handles HTTP requests for auth endpoints.
type Handler struct {
	service AuthService
}

// NewHandler creates a new auth HTTP Handler.
func NewHandler(service AuthService) *Handler {
	return &Handler{
		service: service,
	}
}

// Register handles POST /api/v1/auth/register.
func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	var req RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errors.HandleError(w, errors.NewBadRequestError("invalid request body JSON"))
		return
	}

	authRes, err := h.service.Register(r.Context(), req)
	if err != nil {
		errors.HandleError(w, err)
		return
	}

	response.Success(w, http.StatusCreated, "User registered successfully", authRes)
}

// Login handles POST /api/v1/auth/login.
func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errors.HandleError(w, errors.NewBadRequestError("invalid request body JSON"))
		return
	}

	authRes, err := h.service.Login(r.Context(), req)
	if err != nil {
		errors.HandleError(w, err)
		return
	}

	response.Success(w, http.StatusOK, "Login successful", authRes)
}

// GetMe handles GET /api/v1/auth/me (JWT protected).
func (h *Handler) GetMe(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok || userID == "" {
		errors.HandleError(w, errors.NewUnauthorizedError("unauthorized: missing or invalid user context"))
		return
	}

	meRes, err := h.service.GetMe(r.Context(), userID)
	if err != nil {
		errors.HandleError(w, err)
		return
	}

	response.Success(w, http.StatusOK, "Current user profile fetched", meRes)
}
