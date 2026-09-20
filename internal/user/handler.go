package user

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"go-mini-setup/pkg/errors"
	"go-mini-setup/pkg/response"
)

// Handler handles HTTP requests for user resources.
type Handler struct {
	service UserService
}

// NewHandler creates a new user HTTP Handler.
func NewHandler(service UserService) *Handler {
	return &Handler{
		service: service,
	}
}

// GetUsers handles GET /api/v1/users.
func (h *Handler) GetUsers(w http.ResponseWriter, r *http.Request) {
	users, err := h.service.GetUsers(r.Context())
	if err != nil {
		errors.HandleError(w, err)
		return
	}

	response.Success(w, http.StatusOK, "Users fetched successfully", users)
}

// GetUser handles GET /api/v1/users/{id}.
func (h *Handler) GetUser(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	user, err := h.service.GetUserByID(r.Context(), id)
	if err != nil {
		errors.HandleError(w, err)
		return
	}

	response.Success(w, http.StatusOK, "User fetched successfully", user)
}

// CreateUser handles POST /api/v1/users.
func (h *Handler) CreateUser(w http.ResponseWriter, r *http.Request) {
	var req CreateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errors.HandleError(w, errors.NewBadRequestError("invalid request body JSON"))
		return
	}

	user, err := h.service.CreateUser(r.Context(), req)
	if err != nil {
		errors.HandleError(w, err)
		return
	}

	response.Success(w, http.StatusCreated, "User created successfully", user)
}

// UpdateUser handles PATCH /api/v1/users/{id}.
func (h *Handler) UpdateUser(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	var req UpdateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errors.HandleError(w, errors.NewBadRequestError("invalid request body JSON"))
		return
	}

	user, err := h.service.UpdateUser(r.Context(), id, req)
	if err != nil {
		errors.HandleError(w, err)
		return
	}

	response.Success(w, http.StatusOK, "User updated successfully", user)
}

// DeleteUser handles DELETE /api/v1/users/{id}.
func (h *Handler) DeleteUser(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	if err := h.service.DeleteUser(r.Context(), id); err != nil {
		errors.HandleError(w, err)
		return
	}

	response.Success(w, http.StatusOK, "User deleted successfully", nil)
}
