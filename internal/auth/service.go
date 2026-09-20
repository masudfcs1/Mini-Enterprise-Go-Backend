package auth

import (
	"context"
	"crypto/rand"
	"encoding/hex"

	"go-mini-setup/internal/user"
	"go-mini-setup/pkg/errors"
)

// AuthService handles authentication logic.
type AuthService interface {
	Register(ctx context.Context, req RegisterRequest) (*AuthResponse, error)
	Login(ctx context.Context, req LoginRequest) (*AuthResponse, error)
}

type authService struct {
	repo AuthRepository
}

// NewAuthService creates a new AuthService instance.
func NewAuthService(repo AuthRepository) AuthService {
	return &authService{
		repo: repo,
	}
}

func generateToken() string {
	bytes := make([]byte, 24)
	_, _ = rand.Read(bytes)
	return hex.EncodeToString(bytes)
}

func (s *authService) Register(ctx context.Context, req RegisterRequest) (*AuthResponse, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}

	existing, _ := s.repo.FindByEmail(ctx, req.Email)
	if existing != nil {
		return nil, errors.NewConflictError("an account with this email already exists")
	}

	createdUser, err := s.repo.CreateUser(ctx, req.Email, req.Password, req.Name)
	if err != nil {
		return nil, err
	}

	token := generateToken()
	return &AuthResponse{
		Token: token,
		User:  user.ToUserResponse(createdUser),
	}, nil
}

func (s *authService) Login(ctx context.Context, req LoginRequest) (*AuthResponse, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}

	foundUser, err := s.repo.FindByEmail(ctx, req.Email)
	if err != nil {
		return nil, err
	}
	if foundUser == nil {
		return nil, errors.NewUnauthorizedError("invalid credentials")
	}

	if pwd, ok := foundUser.Password(); !ok || pwd != req.Password {
		return nil, errors.NewUnauthorizedError("invalid credentials")
	}

	token := generateToken()
	return &AuthResponse{
		Token: token,
		User:  user.ToUserResponse(foundUser),
	}, nil
}
