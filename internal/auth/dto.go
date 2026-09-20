package auth

import (
	"strings"

	"go-mini-setup/internal/database/db"
	"go-mini-setup/internal/user"
	"go-mini-setup/pkg/errors"
)

// RegisterRequest defines the incoming payload for account registration.
type RegisterRequest struct {
	Email    string  `json:"email"`
	Password string  `json:"password"`
	Name     *string `json:"name,omitempty"`
	Role     *string `json:"role,omitempty"`
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
	if r.Role != nil {
		roleUpper := strings.ToUpper(strings.TrimSpace(*r.Role))
		if roleUpper != string(db.RoleUser) && roleUpper != string(db.RoleAdmin) && roleUpper != string(db.RoleSuperAdmin) {
			return errors.NewBadRequestError("role must be USER, ADMIN, or SUPER_ADMIN")
		}
		r.Role = &roleUpper
	}
	return nil
}

// LoginRequest defines the incoming payload for authentication.
type LoginRequest struct {
	Email      string `json:"email"`
	Password   string `json:"password"`
	RememberMe bool   `json:"remember_me,omitempty"`
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

// RefreshTokenRequest defines payload for token refresh via request body (if not using cookie).
type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token,omitempty"`
}

// LogoutRequest defines payload for logout revocation via request body (if not using cookie).
type LogoutRequest struct {
	RefreshToken string `json:"refresh_token,omitempty"`
}

// AuthResponse defines the returned authentication tokens and user data.
type AuthResponse struct {
	AccessToken  string            `json:"access_token"`
	RefreshToken string            `json:"refresh_token,omitempty"`
	TokenType    string            `json:"token_type"`
	ExpiresIn    int64             `json:"expires_in"` // access token lifetime in seconds
	User         user.UserResponse `json:"user"`
}

// MeResponse defines the user profile returned by /auth/me.
type MeResponse struct {
	User user.UserResponse `json:"user"`
}
