package auth

import (
	"context"
	stdErrors "errors"
	"testing"
	"time"

	"go-mini-setup/internal/database/db"
	"go-mini-setup/pkg/errors"
	"go-mini-setup/pkg/jwt"
)

type MockAuthRepository struct {
	users map[string]db.UserModel
}

func NewMockAuthRepository() *MockAuthRepository {
	return &MockAuthRepository{
		users: make(map[string]db.UserModel),
	}
}

func (m *MockAuthRepository) FindByEmail(ctx context.Context, email string) (*db.UserModel, error) {
	for _, u := range m.users {
		if u.Email == email {
			return &u, nil
		}
	}
	return nil, nil
}

func (m *MockAuthRepository) FindByID(ctx context.Context, id string) (*db.UserModel, error) {
	u, ok := m.users[id]
	if !ok {
		return nil, nil
	}
	return &u, nil
}

func (m *MockAuthRepository) CreateUser(ctx context.Context, email, password string, name *string) (*db.UserModel, error) {
	u := db.UserModel{
		InnerUser: db.InnerUser{
			ID:        "auth-user-1",
			Email:     email,
			Password:  &password,
			Name:      name,
			CreatedAt: db.DateTime(time.Now()),
			UpdatedAt: db.DateTime(time.Now()),
		},
	}
	m.users[u.ID] = u
	return &u, nil
}

func TestAuthService_RegisterAndLogin(t *testing.T) {
	repo := NewMockAuthRepository()
	secret := "test-secret-key-32bytes-secure!"
	service := NewAuthService(repo, secret, 1*time.Hour)
	ctx := context.Background()

	name := "Charlie"
	regReq := RegisterRequest{
		Email:    "charlie@example.com",
		Password: "password123",
		Name:     &name,
	}

	// Register
	regRes, err := service.Register(ctx, regReq)
	if err != nil {
		t.Fatalf("expected register success, got %v", err)
	}
	if regRes.Token == "" {
		t.Fatal("expected non-empty token")
	}
	if regRes.User.Email != "charlie@example.com" {
		t.Errorf("expected email charlie@example.com, got %s", regRes.User.Email)
	}

	// Validate JWT signature and claims
	claims, err := jwt.ValidateToken(regRes.Token, secret)
	if err != nil {
		t.Fatalf("expected valid JWT token, got error: %v", err)
	}
	if claims.UserID != regRes.User.ID {
		t.Errorf("expected claims UserID %s, got %s", regRes.User.ID, claims.UserID)
	}
	if claims.Email != regRes.User.Email {
		t.Errorf("expected claims Email %s, got %s", regRes.User.Email, claims.Email)
	}

	// Conflict Register
	_, err = service.Register(ctx, regReq)
	if err == nil {
		t.Fatal("expected conflict on duplicate email, got nil")
	}

	// Login Success
	loginRes, err := service.Login(ctx, LoginRequest{
		Email:    "charlie@example.com",
		Password: "password123",
	})
	if err != nil {
		t.Fatalf("expected login success, got %v", err)
	}
	if loginRes.Token == "" {
		t.Fatal("expected non-empty auth token on login")
	}

	// GetMe Success
	meRes, err := service.GetMe(ctx, regRes.User.ID)
	if err != nil {
		t.Fatalf("expected GetMe success, got %v", err)
	}
	if meRes.User.Email != "charlie@example.com" {
		t.Errorf("expected email charlie@example.com, got %s", meRes.User.Email)
	}

	// Login Invalid Password
	_, err = service.Login(ctx, LoginRequest{
		Email:    "charlie@example.com",
		Password: "wrongpassword",
	})
	if err == nil {
		t.Fatal("expected unauthorized error for wrong password, got nil")
	}
	var appErr *errors.AppError
	if !stdErrors.As(err, &appErr) || appErr.Status != 401 {
		t.Errorf("expected 401 status, got %v", err)
	}
}
