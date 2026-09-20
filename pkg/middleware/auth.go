package middleware

import (
	"context"
	"net/http"
	"strings"

	"go-mini-setup/pkg/errors"
	"go-mini-setup/pkg/jwt"
)

type contextKey string

const (
	UserIDKey    contextKey = "userID"
	UserEmailKey contextKey = "userEmail"
	UserRoleKey  contextKey = "userRole"
)

// RequireAuth returns a middleware that validates JWT access tokens from Authorization Header or Cookie.
func RequireAuth(secret, issuer, audience string) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			tokenString := ""

			// 1. Try Authorization Header
			authHeader := r.Header.Get("Authorization")
			if authHeader != "" {
				parts := strings.SplitN(authHeader, " ", 2)
				if len(parts) == 2 && strings.EqualFold(parts[0], "bearer") {
					tokenString = strings.TrimSpace(parts[1])
				}
			}

			// 2. Fallback to access_token cookie
			if tokenString == "" {
				if cookie, err := r.Cookie("access_token"); err == nil && cookie.Value != "" {
					tokenString = cookie.Value
				}
			}

			if tokenString == "" {
				errors.HandleError(w, errors.NewUnauthorizedError("missing authentication token in header or cookie"))
				return
			}

			claims, err := jwt.ValidateAccessToken(tokenString, secret, issuer, audience)
			if err != nil {
				errors.HandleError(w, errors.NewUnauthorizedError(err.Error()))
				return
			}

			ctx := context.WithValue(r.Context(), UserIDKey, claims.UserID)
			ctx = context.WithValue(ctx, UserEmailKey, claims.Email)
			ctx = context.WithValue(ctx, UserRoleKey, claims.Role)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// GetUserID extracts the authenticated user ID from context.
func GetUserID(ctx context.Context) (string, bool) {
	id, ok := ctx.Value(UserIDKey).(string)
	return id, ok
}

// GetUserEmail extracts the authenticated user email from context.
func GetUserEmail(ctx context.Context) (string, bool) {
	email, ok := ctx.Value(UserEmailKey).(string)
	return email, ok
}

// GetUserRole extracts the authenticated user role from context.
func GetUserRole(ctx context.Context) (string, bool) {
	role, ok := ctx.Value(UserRoleKey).(string)
	return role, ok
}
