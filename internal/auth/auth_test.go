package auth

import (
	"context"
	stdErrors "errors"
	"testing"
	"time"

	"go-mini-setup/internal/database/db"
	"go-mini-setup/pkg/config"
	"go-mini-setup/pkg/errors"
	"go-mini-setup/pkg/jwt"
)

type MockAuthRepository struct {
	users         map[string]db.UserModel
	refreshTokens map[string]db.RefreshTokenModel
}

func NewMockAuthRepository() *MockAuthRepository {
	return &MockAuthRepository{
		users:         make(map[string]db.UserModel),
		refreshTokens: make(map[string]db.RefreshTokenModel),
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

func (m *MockAuthRepository) CreateUser(ctx context.Context, email, password string, name, role *string) (*db.UserModel, error) {
	assignedRole := db.RoleUser
	if role != nil {
		assignedRole = db.Role(*role)
	}

	u := db.UserModel{
		InnerUser: db.InnerUser{
			ID:        "auth-user-1",
			Email:     email,
			Password:  &password,
			Name:      name,
			Role:      assignedRole,
			CreatedAt: db.DateTime(time.Now()),
			UpdatedAt: db.DateTime(time.Now()),
		},
	}
	m.users[u.ID] = u
	return &u, nil
}

func (m *MockAuthRepository) CreateRefreshToken(ctx context.Context, userID, token string, expiresAt time.Time) (*db.RefreshTokenModel, error) {
	rt := db.RefreshTokenModel{
		InnerRefreshToken: db.InnerRefreshToken{
			ID:        "rt-123",
			Token:     token,
			UserID:    userID,
			ExpiresAt: db.DateTime(expiresAt),
			CreatedAt: db.DateTime(time.Now()),
		},
	}
	m.refreshTokens[token] = rt
	return &rt, nil
}

func (m *MockAuthRepository) FindRefreshToken(ctx context.Context, token string) (*db.RefreshTokenModel, error) {
	rt, ok := m.refreshTokens[token]
	if !ok {
		return nil, nil
	}
	return &rt, nil
}

func (m *MockAuthRepository) DeleteRefreshToken(ctx context.Context, token string) error {
	delete(m.refreshTokens, token)
	return nil
}

func (m *MockAuthRepository) DeleteUserRefreshTokens(ctx context.Context, userID string) error {
	for k, v := range m.refreshTokens {
		if v.UserID == userID {
			delete(m.refreshTokens, k)
		}
	}
	return nil
}

func TestAuthService_FullEnterpriseAuthFlow(t *testing.T) {
	repo := NewMockAuthRepository()
	cfg := &config.Config{
		JWTAccessSecret:           "test-access-secret-32-chars-long!",
		JWTRefreshSecret:          "test-refresh-secret-32-chars-long!",
		JWTIssuer:                 "enterprise-backend",
		JWTAudience:               "enterprise-api",
		JWTAccessExpires:          15 * time.Minute,
		JWTRefreshExpires:         7 * 24 * time.Hour,
		JWTRefreshRememberExpires: 30 * 24 * time.Hour,
		AuthTokenTransport:        "cookie",
	}

	service := NewAuthService(repo, cfg)
	ctx := context.Background()

	name := "Charlie"
	role := "ADMIN"
	regReq := RegisterRequest{
		Email:    "charlie@example.com",
		Password: "password123",
		Name:     &name,
		Role:     &role,
	}

	// 1. Register with Role
	regRes, err := service.Register(ctx, regReq)
	if err != nil {
		t.Fatalf("expected register success, got %v", err)
	}
	if regRes.AccessToken == "" || regRes.RefreshToken == "" {
		t.Fatal("expected both access and refresh tokens")
	}
	if regRes.User.Role != "ADMIN" {
		t.Errorf("expected role ADMIN, got %s", regRes.User.Role)
	}

	// 2. Validate Access Token claims
	accessClaims, err := jwt.ValidateAccessToken(regRes.AccessToken, cfg.JWTAccessSecret, cfg.JWTIssuer, cfg.JWTAudience)
	if err != nil {
		t.Fatalf("failed to validate access token: %v", err)
	}
	if accessClaims.UserID != regRes.User.ID || accessClaims.Role != "ADMIN" {
		t.Errorf("access token claims mismatch: %+v", accessClaims)
	}

	// 3. Login with RememberMe
	loginRes, err := service.Login(ctx, LoginRequest{
		Email:      "charlie@example.com",
		Password:   "password123",
		RememberMe: true,
	})
	if err != nil {
		t.Fatalf("expected login success, got %v", err)
	}
	if loginRes.AccessToken == "" || loginRes.RefreshToken == "" {
		t.Fatal("expected access and refresh tokens on login")
	}

	// 4. Token Rotation: Refresh Token
	refreshRes, err := service.RefreshToken(ctx, loginRes.RefreshToken)
	if err != nil {
		t.Fatalf("failed to refresh token: %v", err)
	}
	if refreshRes.AccessToken == "" || refreshRes.RefreshToken == "" {
		t.Fatal("expected rotated tokens")
	}
	if refreshRes.RefreshToken == loginRes.RefreshToken {
		t.Error("expected new rotated refresh token, got identical token")
	}

	// 5. Old refresh token should now be revoked
	_, err = service.RefreshToken(ctx, loginRes.RefreshToken)
	if err == nil {
		t.Fatal("expected error using revoked refresh token, got nil")
	}

	// 6. GetMe
	meRes, err := service.GetMe(ctx, regRes.User.ID)
	if err != nil {
		t.Fatalf("expected GetMe success, got %v", err)
	}
	if meRes.User.Email != "charlie@example.com" || meRes.User.Role != "ADMIN" {
		t.Errorf("me response mismatch: %+v", meRes)
	}

	// 7. Logout revokes token
	err = service.Logout(ctx, refreshRes.RefreshToken)
	if err != nil {
		t.Fatalf("logout failed: %v", err)
	}

	_, err = service.RefreshToken(ctx, refreshRes.RefreshToken)
	if err == nil {
		t.Fatal("expected error refreshing after logout, got nil")
	}

	// 8. Invalid password error check
	_, err = service.Login(ctx, LoginRequest{
		Email:    "charlie@example.com",
		Password: "wrongpassword",
	})
	if err == nil {
		t.Fatal("expected unauthorized on wrong password, got nil")
	}
	var appErr *errors.AppError
	if !stdErrors.As(err, &appErr) || appErr.Status != 401 {
		t.Errorf("expected 401 status, got %v", err)
	}
}
