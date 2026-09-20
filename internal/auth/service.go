package auth

import (
	"context"
	"time"

	"golang.org/x/crypto/bcrypt"

	"go-mini-setup/internal/user"
	"go-mini-setup/pkg/errors"
	"go-mini-setup/pkg/jwt"
)

// AuthService handles authentication and account logic.
type AuthService interface {
	Register(ctx context.Context, req RegisterRequest) (*AuthResponse, error)
	Login(ctx context.Context, req LoginRequest) (*AuthResponse, error)
	GetMe(ctx context.Context, userID string) (*MeResponse, error)
}

type authService struct {
	repo      AuthRepository
	jwtSecret string
	jwtTTL    time.Duration
}

// NewAuthService creates a new AuthService instance with JWT configuration.
func NewAuthService(repo AuthRepository, jwtSecret string, jwtTTL time.Duration) AuthService {
	return &authService{
		repo:      repo,
		jwtSecret: jwtSecret,
		jwtTTL:    jwtTTL,
	}
}

func (s *authService) Register(ctx context.Context, req RegisterRequest) (*AuthResponse, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}

	existing, _ := s.repo.FindByEmail(ctx, req.Email)
	if existing != nil {
		return nil, errors.NewConflictError("an account with this email already exists")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, errors.NewInternalError("failed to secure password")
	}

	createdUser, err := s.repo.CreateUser(ctx, req.Email, string(hashedPassword), req.Name)
	if err != nil {
		return nil, err
	}

	token, err := jwt.GenerateToken(createdUser.ID, createdUser.Email, s.jwtSecret, s.jwtTTL)
	if err != nil {
		return nil, errors.NewInternalError("failed to generate authentication token")
	}

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
		return nil, errors.NewUnauthorizedError("invalid email or password")
	}

	storedHash, ok := foundUser.Password()
	if !ok {
		return nil, errors.NewUnauthorizedError("invalid email or password")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(storedHash), []byte(req.Password)); err != nil {
		return nil, errors.NewUnauthorizedError("invalid email or password")
	}

	token, err := jwt.GenerateToken(foundUser.ID, foundUser.Email, s.jwtSecret, s.jwtTTL)
	if err != nil {
		return nil, errors.NewInternalError("failed to generate authentication token")
	}

	return &AuthResponse{
		Token: token,
		User:  user.ToUserResponse(foundUser),
	}, nil
}

func (s *authService) GetMe(ctx context.Context, userID string) (*MeResponse, error) {
	if userID == "" {
		return nil, errors.NewUnauthorizedError("unauthorized user context")
	}

	foundUser, err := s.repo.FindByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if foundUser == nil {
		return nil, errors.NewNotFoundError("user account not found")
	}

	return &MeResponse{
		User: user.ToUserResponse(foundUser),
	}, nil
}
