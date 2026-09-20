package auth

import (
	"context"
	"time"

	"golang.org/x/crypto/bcrypt"

	"go-mini-setup/internal/database/db"
	"go-mini-setup/internal/user"
	"go-mini-setup/pkg/config"
	"go-mini-setup/pkg/errors"
	"go-mini-setup/pkg/jwt"
)

// AuthService handles authentication, tokens, and account operations.
type AuthService interface {
	Register(ctx context.Context, req RegisterRequest) (*AuthResponse, error)
	Login(ctx context.Context, req LoginRequest) (*AuthResponse, error)
	RefreshToken(ctx context.Context, refreshTokenStr string) (*AuthResponse, error)
	Logout(ctx context.Context, refreshTokenStr string) error
	GetMe(ctx context.Context, userID string) (*MeResponse, error)
}

type authService struct {
	repo AuthRepository
	cfg  *config.Config
}

// NewAuthService creates a new AuthService instance with full enterprise configuration.
func NewAuthService(repo AuthRepository, cfg *config.Config) AuthService {
	return &authService{
		repo: repo,
		cfg:  cfg,
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

	createdUser, err := s.repo.CreateUser(ctx, req.Email, string(hashedPassword), req.Name, req.Role)
	if err != nil {
		return nil, err
	}

	userRole := string(createdUser.Role)
	if userRole == "" {
		userRole = string(db.RoleUser)
	}

	// 1. Generate Access Token (15m)
	accessToken, err := jwt.GenerateAccessToken(
		createdUser.ID,
		createdUser.Email,
		userRole,
		s.cfg.JWTAccessSecret,
		s.cfg.JWTIssuer,
		s.cfg.JWTAudience,
		s.cfg.JWTAccessExpires,
	)
	if err != nil {
		return nil, errors.NewInternalError("failed to generate access token")
	}

	// 2. Generate Refresh Token (7d)
	refreshToken, expiresAt, err := jwt.GenerateRefreshToken(
		createdUser.ID,
		s.cfg.JWTRefreshSecret,
		s.cfg.JWTIssuer,
		s.cfg.JWTAudience,
		s.cfg.JWTRefreshExpires,
	)
	if err != nil {
		return nil, errors.NewInternalError("failed to generate refresh token")
	}

	// 3. Persist Refresh Token in Database for State Tracking & Revocation
	if _, err := s.repo.CreateRefreshToken(ctx, createdUser.ID, refreshToken, expiresAt); err != nil {
		return nil, errors.NewInternalError("failed to persist refresh token session")
	}

	return &AuthResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		TokenType:    "Bearer",
		ExpiresIn:    int64(s.cfg.JWTAccessExpires.Seconds()),
		User:         user.ToUserResponse(createdUser),
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

	userRole := string(foundUser.Role)
	if userRole == "" {
		userRole = string(db.RoleUser)
	}

	// Determine Refresh Token TTL based on RememberMe
	refreshTTL := s.cfg.JWTRefreshExpires
	if req.RememberMe {
		refreshTTL = s.cfg.JWTRefreshRememberExpires
	}

	// 1. Generate Access Token (15m)
	accessToken, err := jwt.GenerateAccessToken(
		foundUser.ID,
		foundUser.Email,
		userRole,
		s.cfg.JWTAccessSecret,
		s.cfg.JWTIssuer,
		s.cfg.JWTAudience,
		s.cfg.JWTAccessExpires,
	)
	if err != nil {
		return nil, errors.NewInternalError("failed to generate access token")
	}

	// 2. Generate Refresh Token
	refreshToken, expiresAt, err := jwt.GenerateRefreshToken(
		foundUser.ID,
		s.cfg.JWTRefreshSecret,
		s.cfg.JWTIssuer,
		s.cfg.JWTAudience,
		refreshTTL,
	)
	if err != nil {
		return nil, errors.NewInternalError("failed to generate refresh token")
	}

	// 3. Persist Refresh Token
	if _, err := s.repo.CreateRefreshToken(ctx, foundUser.ID, refreshToken, expiresAt); err != nil {
		return nil, errors.NewInternalError("failed to persist refresh token session")
	}

	return &AuthResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		TokenType:    "Bearer",
		ExpiresIn:    int64(s.cfg.JWTAccessExpires.Seconds()),
		User:         user.ToUserResponse(foundUser),
	}, nil
}

func (s *authService) RefreshToken(ctx context.Context, refreshTokenStr string) (*AuthResponse, error) {
	if refreshTokenStr == "" {
		return nil, errors.NewUnauthorizedError("refresh token is required")
	}

	// 1. Cryptographically validate the Refresh Token signature and type
	claims, err := jwt.ValidateRefreshToken(refreshTokenStr, s.cfg.JWTRefreshSecret, s.cfg.JWTIssuer, s.cfg.JWTAudience)
	if err != nil {
		return nil, errors.NewUnauthorizedError("invalid or expired refresh token")
	}

	// 2. Check Database record to verify token has not been revoked
	tokenRecord, err := s.repo.FindRefreshToken(ctx, refreshTokenStr)
	if err != nil || tokenRecord == nil {
		return nil, errors.NewUnauthorizedError("refresh token session has been revoked or is invalid")
	}

	if time.Now().After(time.Time(tokenRecord.ExpiresAt)) {
		_ = s.repo.DeleteRefreshToken(ctx, refreshTokenStr)
		return nil, errors.NewUnauthorizedError("refresh token session has expired")
	}

	// 3. Find User
	foundUser, err := s.repo.FindByID(ctx, claims.UserID)
	if err != nil || foundUser == nil {
		return nil, errors.NewUnauthorizedError("user account no longer exists")
	}

	// 4. Token Rotation: Revoke old refresh token
	_ = s.repo.DeleteRefreshToken(ctx, refreshTokenStr)

	userRole := string(foundUser.Role)
	if userRole == "" {
		userRole = string(db.RoleUser)
	}

	// 5. Generate new Access Token
	newAccessToken, err := jwt.GenerateAccessToken(
		foundUser.ID,
		foundUser.Email,
		userRole,
		s.cfg.JWTAccessSecret,
		s.cfg.JWTIssuer,
		s.cfg.JWTAudience,
		s.cfg.JWTAccessExpires,
	)
	if err != nil {
		return nil, errors.NewInternalError("failed to generate new access token")
	}

	// 6. Generate new Refresh Token
	newRefreshToken, expiresAt, err := jwt.GenerateRefreshToken(
		foundUser.ID,
		s.cfg.JWTRefreshSecret,
		s.cfg.JWTIssuer,
		s.cfg.JWTAudience,
		s.cfg.JWTRefreshExpires,
	)
	if err != nil {
		return nil, errors.NewInternalError("failed to generate new refresh token")
	}

	// 7. Store new rotated refresh token
	if _, err := s.repo.CreateRefreshToken(ctx, foundUser.ID, newRefreshToken, expiresAt); err != nil {
		return nil, errors.NewInternalError("failed to persist rotated refresh token")
	}

	return &AuthResponse{
		AccessToken:  newAccessToken,
		RefreshToken: newRefreshToken,
		TokenType:    "Bearer",
		ExpiresIn:    int64(s.cfg.JWTAccessExpires.Seconds()),
		User:         user.ToUserResponse(foundUser),
	}, nil
}

func (s *authService) Logout(ctx context.Context, refreshTokenStr string) error {
	if refreshTokenStr != "" {
		_ = s.repo.DeleteRefreshToken(ctx, refreshTokenStr)
	}
	return nil
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
