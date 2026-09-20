package user

import (
	"context"
	stdErrors "errors"
	"testing"
	"time"

	"go-mini-setup/internal/database/db"
	"go-mini-setup/pkg/errors"
)

// MockUserRepository implements UserRepository in memory for testing.
type MockUserRepository struct {
	users map[string]UserModel
}

func NewMockUserRepository() *MockUserRepository {
	return &MockUserRepository{
		users: make(map[string]UserModel),
	}
}

func (m *MockUserRepository) FindAll(ctx context.Context) ([]UserModel, error) {
	list := make([]UserModel, 0, len(m.users))
	for _, u := range m.users {
		list = append(list, u)
	}
	return list, nil
}

func (m *MockUserRepository) FindByID(ctx context.Context, id string) (*UserModel, error) {
	u, ok := m.users[id]
	if !ok {
		return nil, nil
	}
	return &u, nil
}

func (m *MockUserRepository) FindByEmail(ctx context.Context, email string) (*UserModel, error) {
	for _, u := range m.users {
		if u.Email == email {
			return &u, nil
		}
	}
	return nil, nil
}

func (m *MockUserRepository) Create(ctx context.Context, req CreateUserRequest) (*UserModel, error) {
	u := UserModel{
		InnerUser: db.InnerUser{
			ID:        "user-123",
			Email:     req.Email,
			Name:      req.Name,
			CreatedAt: db.DateTime(time.Now()),
			UpdatedAt: db.DateTime(time.Now()),
		},
	}
	m.users[u.ID] = u
	return &u, nil
}

func (m *MockUserRepository) Update(ctx context.Context, id string, req UpdateUserRequest) (*UserModel, error) {
	u, ok := m.users[id]
	if !ok {
		return nil, db.ErrNotFound
	}
	if req.Email != nil {
		u.InnerUser.Email = *req.Email
	}
	if req.Name != nil {
		u.InnerUser.Name = req.Name
	}
	m.users[id] = u
	return &u, nil
}

func (m *MockUserRepository) Delete(ctx context.Context, id string) (*UserModel, error) {
	u, ok := m.users[id]
	if !ok {
		return nil, db.ErrNotFound
	}
	delete(m.users, id)
	return &u, nil
}

func TestUserService_CreateUser(t *testing.T) {
	repo := NewMockUserRepository()
	service := NewUserService(repo)
	ctx := context.Background()

	name := "Alice"
	req := CreateUserRequest{
		Email: "alice@example.com",
		Name:  &name,
	}

	userRes, err := service.CreateUser(ctx, req)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if userRes.Email != "alice@example.com" {
		t.Errorf("expected email 'alice@example.com', got %s", userRes.Email)
	}

	// Conflict test
	_, err = service.CreateUser(ctx, req)
	if err == nil {
		t.Fatal("expected error on duplicate email, got nil")
	}
	var appErr *errors.AppError
	if !stdErrors.As(err, &appErr) || appErr.Status != 409 {
		t.Errorf("expected 409 conflict, got %v", err)
	}
}

func TestUserService_GetUserByID(t *testing.T) {
	repo := NewMockUserRepository()
	service := NewUserService(repo)
	ctx := context.Background()

	// Not found test
	_, err := service.GetUserByID(ctx, "non-existent")
	if err == nil {
		t.Fatal("expected error for non-existent user, got nil")
	}

	// Create and find
	name := "Bob"
	created, err := service.CreateUser(ctx, CreateUserRequest{
		Email: "bob@example.com",
		Name:  &name,
	})
	if err != nil {
		t.Fatalf("failed to create user: %v", err)
	}

	found, err := service.GetUserByID(ctx, created.ID)
	if err != nil {
		t.Fatalf("expected to find user, got %v", err)
	}
	if found.Email != "bob@example.com" {
		t.Errorf("expected email bob@example.com, got %s", found.Email)
	}
}

func TestCreateUserRequest_Validation(t *testing.T) {
	tests := []struct {
		name    string
		req     CreateUserRequest
		wantErr bool
	}{
		{
			name:    "valid request",
			req:     CreateUserRequest{Email: "test@domain.com"},
			wantErr: false,
		},
		{
			name:    "empty email",
			req:     CreateUserRequest{Email: ""},
			wantErr: true,
		},
		{
			name:    "invalid email format",
			req:     CreateUserRequest{Email: "invalid-email"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.req.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
