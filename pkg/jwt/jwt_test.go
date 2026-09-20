package jwt

import (
	"testing"
	"time"
)

func TestJWT_GenerateAndValidate(t *testing.T) {
	secret := "test-jwt-secret-key-32bytes-secure!"
	userID := "user-12345"
	email := "test@example.com"

	// 1. Valid token
	token, err := GenerateToken(userID, email, secret, 1*time.Hour)
	if err != nil {
		t.Fatalf("failed to generate token: %v", err)
	}
	if token == "" {
		t.Fatal("expected non-empty token string")
	}

	claims, err := ValidateToken(token, secret)
	if err != nil {
		t.Fatalf("failed to validate token: %v", err)
	}
	if claims.UserID != userID {
		t.Errorf("expected userID %s, got %s", userID, claims.UserID)
	}
	if claims.Email != email {
		t.Errorf("expected email %s, got %s", email, claims.Email)
	}

	// 2. Invalid secret
	_, err = ValidateToken(token, "wrong-secret-key-123456789012345")
	if err == nil {
		t.Fatal("expected validation to fail with wrong secret, got nil")
	}

	// 3. Expired token
	expiredToken, err := GenerateToken(userID, email, secret, -1*time.Minute)
	if err != nil {
		t.Fatalf("failed to generate expired token: %v", err)
	}

	_, err = ValidateToken(expiredToken, secret)
	if err != ErrExpiredToken {
		t.Errorf("expected ErrExpiredToken, got %v", err)
	}
}
