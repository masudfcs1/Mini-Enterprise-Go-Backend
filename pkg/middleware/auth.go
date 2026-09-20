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
)

// RequireAuth returns a middleware that validates JWT Bearer tokens.
func RequireAuth(secret string) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				errors.HandleError(w, errors.NewUnauthorizedError("authorization header is required"))
				return
			}

			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) != 2 || !strings.EqualFold(parts[0], "bearer") {
				errors.HandleError(w, errors.NewUnauthorizedError("authorization header format must be Bearer <token>"))
				return
			}

			tokenString := parts[1]
			claims, err := jwt.ValidateToken(tokenString, secret)
			if err != nil {
				errors.HandleError(w, errors.NewUnauthorizedError(err.Error()))
				return
			}

			ctx := context.WithValue(r.Context(), UserIDKey, claims.UserID)
			ctx = context.WithValue(ctx, UserEmailKey, claims.Email)

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
