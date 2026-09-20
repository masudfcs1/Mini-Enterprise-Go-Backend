package user

import (
	"strings"
	"time"

	"go-mini-setup/pkg/errors"
)

// CreateUserRequest defines the incoming payload for creating a new user.
type CreateUserRequest struct {
	Email    string  `json:"email"`
	Name     *string `json:"name,omitempty"`
	Password *string `json:"password,omitempty"`
}

// Validate validates the CreateUserRequest.
func (r *CreateUserRequest) Validate() error {
	r.Email = strings.TrimSpace(r.Email)
	if r.Email == "" {
		return errors.NewBadRequestError("email is required")
	}
	if !strings.Contains(r.Email, "@") || !strings.Contains(r.Email, ".") {
		return errors.NewBadRequestError("invalid email address format")
	}
	if r.Name != nil {
		trimmed := strings.TrimSpace(*r.Name)
		r.Name = &trimmed
	}
	return nil
}

// UpdateUserRequest defines the incoming payload for updating an existing user.
type UpdateUserRequest struct {
	Email *string `json:"email,omitempty"`
	Name  *string `json:"name,omitempty"`
}

// Validate validates the UpdateUserRequest.
func (r *UpdateUserRequest) Validate() error {
	if r.Email == nil && r.Name == nil {
		return errors.NewBadRequestError("at least one field (email or name) must be provided to update")
	}
	if r.Email != nil {
		trimmed := strings.TrimSpace(*r.Email)
		if trimmed == "" || !strings.Contains(trimmed, "@") || !strings.Contains(trimmed, ".") {
			return errors.NewBadRequestError("invalid email address format")
		}
		r.Email = &trimmed
	}
	if r.Name != nil {
		trimmed := strings.TrimSpace(*r.Name)
		r.Name = &trimmed
	}
	return nil
}

// UserResponse is the clean, sanitized public response representation of a User.
type UserResponse struct {
	ID        string    `json:"id"`
	Email     string    `json:"email"`
	Name      *string   `json:"name,omitempty"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

// ToUserResponse converts a database UserModel to a safe UserResponse.
func ToUserResponse(u *UserModel) UserResponse {
	var name *string
	if val, ok := u.Name(); ok {
		name = &val
	}
	return UserResponse{
		ID:        u.ID,
		Email:     u.Email,
		Name:      name,
		CreatedAt: u.CreatedAt,
		UpdatedAt: u.UpdatedAt,
	}
}

// ToUserResponseList converts a list of UserModels to UserResponses.
func ToUserResponseList(users []UserModel) []UserResponse {
	res := make([]UserResponse, len(users))
	for i := range users {
		res[i] = ToUserResponse(&users[i])
	}
	return res
}
