package jwt

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var (
	ErrInvalidToken = errors.New("invalid or malformed token")
	ErrExpiredToken = errors.New("token has expired")
	ErrWrongType    = errors.New("unexpected token type")
	ErrBadAudience  = errors.New("token audience mismatch")
	ErrBadIssuer    = errors.New("token issuer mismatch")
)

const (
	TokenTypeAccess  = "access"
	TokenTypeRefresh = "refresh"
)

// Claims defines the payload carried by JWT tokens.
type Claims struct {
	UserID    string `json:"user_id"`
	Email     string `json:"email,omitempty"`
	Role      string `json:"role,omitempty"`
	TokenType string `json:"token_type"`
	jwt.RegisteredClaims
}

// GenerateAccessToken creates a short-lived access JWT token containing user identity and role.
func GenerateAccessToken(userID, email, role, secret, issuer, audience string, ttl time.Duration) (string, error) {
	now := time.Now()
	claims := Claims{
		UserID:    userID,
		Email:     email,
		Role:      role,
		TokenType: TokenTypeAccess,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(ttl)),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			Subject:   userID,
			Issuer:    issuer,
			Audience:  jwt.ClaimStrings{audience},
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}

// GenerateRefreshToken creates a long-lived refresh JWT token for stateful session renewal.
func GenerateRefreshToken(userID, secret, issuer, audience string, ttl time.Duration) (string, time.Time, error) {
	now := time.Now()
	expiresAt := now.Add(ttl)

	claims := Claims{
		UserID:    userID,
		TokenType: TokenTypeRefresh,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			Subject:   userID,
			Issuer:    issuer,
			Audience:  jwt.ClaimStrings{audience},
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(secret))
	if err != nil {
		return "", time.Time{}, fmt.Errorf("failed to sign refresh token: %w", err)
	}

	return tokenString, expiresAt, nil
}

// ValidateToken parses, validates and checks token type, issuer and audience.
func ValidateToken(tokenString, secret, expectedType, issuer, audience string) (*Claims, error) {
	parsedToken, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(secret), nil
	})

	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, ErrExpiredToken
		}
		return nil, ErrInvalidToken
	}

	claims, ok := parsedToken.Claims.(*Claims)
	if !ok || !parsedToken.Valid {
		return nil, ErrInvalidToken
	}

	if expectedType != "" && claims.TokenType != expectedType {
		return nil, ErrWrongType
	}

	if issuer != "" && claims.Issuer != issuer {
		return nil, ErrBadIssuer
	}

	return claims, nil
}

// ValidateAccessToken validates an incoming access token.
func ValidateAccessToken(tokenString, secret, issuer, audience string) (*Claims, error) {
	return ValidateToken(tokenString, secret, TokenTypeAccess, issuer, audience)
}

// ValidateRefreshToken validates an incoming refresh token.
func ValidateRefreshToken(tokenString, secret, issuer, audience string) (*Claims, error) {
	return ValidateToken(tokenString, secret, TokenTypeRefresh, issuer, audience)
}
