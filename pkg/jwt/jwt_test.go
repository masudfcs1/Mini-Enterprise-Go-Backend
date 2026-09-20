package jwt

import (
	"testing"
	"time"
)

func TestJWT_AccessAndRefreshToken(t *testing.T) {
	accessSecret := "test-access-secret-32-chars-long!"
	refreshSecret := "test-refresh-secret-32-chars-long!"
	issuer := "enterprise-backend"
	audience := "enterprise-api"
	userID := "user-12345"
	email := "test@example.com"
	role := "ADMIN"

	// 1. Valid Access Token
	accessToken, err := GenerateAccessToken(userID, email, role, accessSecret, issuer, audience, 15*time.Minute)
	if err != nil {
		t.Fatalf("failed to generate access token: %v", err)
	}

	claims, err := ValidateAccessToken(accessToken, accessSecret, issuer, audience)
	if err != nil {
		t.Fatalf("failed to validate access token: %v", err)
	}
	if claims.UserID != userID || claims.Email != email || claims.Role != role {
		t.Errorf("claims mismatch: %+v", claims)
	}
	if claims.TokenType != TokenTypeAccess {
		t.Errorf("expected access token type, got %s", claims.TokenType)
	}

	// 2. Valid Refresh Token
	refreshToken, _, err := GenerateRefreshToken(userID, refreshSecret, issuer, audience, 7*24*time.Hour)
	if err != nil {
		t.Fatalf("failed to generate refresh token: %v", err)
	}

	refreshClaims, err := ValidateRefreshToken(refreshToken, refreshSecret, issuer, audience)
	if err != nil {
		t.Fatalf("failed to validate refresh token: %v", err)
	}
	if refreshClaims.UserID != userID || refreshClaims.TokenType != TokenTypeRefresh {
		t.Errorf("refresh claims mismatch: %+v", refreshClaims)
	}

	// 3. Wrong Token Type: passing access token to ValidateRefreshToken should error
	_, err = ValidateRefreshToken(accessToken, accessSecret, issuer, audience)
	if err != ErrWrongType {
		t.Errorf("expected ErrWrongType, got %v", err)
	}

	// 4. Bad Secret
	_, err = ValidateAccessToken(accessToken, "wrong-secret-key-32-characters-1234", issuer, audience)
	if err == nil {
		t.Fatal("expected error on wrong secret, got nil")
	}

	// 5. Expired Token
	expiredToken, err := GenerateAccessToken(userID, email, role, accessSecret, issuer, audience, -1*time.Minute)
	if err != nil {
		t.Fatalf("failed to generate expired token: %v", err)
	}

	_, err = ValidateAccessToken(expiredToken, accessSecret, issuer, audience)
	if err != ErrExpiredToken {
		t.Errorf("expected ErrExpiredToken, got %v", err)
	}
}
