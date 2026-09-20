package auth

import (
	"strings"

	"go-mini-setup/internal/user"
	"go-mini-setup/pkg/errors"
)

// RegisterRequest defines the incoming payload for account registration.
type RegisterRequest struct {
	Email    string  `json:"email"`
	Password string  `json:"password"`
	Name     *string `json:"name,omitempty"`
}

func (r *RegisterRequest) Validate() error {
	r.Email = strings.TrimSpace(r.Email)
	if r.Email == "" {
		return errors.NewBadRequestError("email is required")
	}
	if !strings.Contains(r.Email, "@") || !strings.Contains(r.Email, ".") {
		return errors.NewBadRequestError("invalid email address format")
	}
	if len(r.Password) < 6 {
		return errors.NewBadRequestError("password must be at least 6 characters")
	}
	if r.Name != nil {
		trimmed := strings.TrimSpace(*r.Name)
		r.Name = &trimmed
	}
	return nil
}

// LoginRequest defines the incoming payload for authentication.
type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (r *LoginRequest) Validate() error {
	r.Email = strings.TrimSpace(r.Email)
	if r.Email == "" {
		return errors.NewBadRequestError("email is required")
	}
	if r.Password == "" {
		return errors.NewBadRequestError("password is required")
	}
	return nil
}

// AuthResponse defines the returned authentication token and user data.
type AuthResponse struct {
	Token string            `json:"token"`
	User  user.UserResponse `json:"user"`
}
