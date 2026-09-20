package middleware

import (
	"fmt"
	"net/http"
	"strings"

	"go-mini-setup/pkg/errors"
)

// RequireRole creates an RBAC middleware checking if the authenticated user has one of the allowed roles.
func RequireRole(allowedRoles ...string) func(next http.Handler) http.Handler {
	allowedMap := make(map[string]struct{}, len(allowedRoles))
	for _, r := range allowedRoles {
		allowedMap[strings.ToUpper(r)] = struct{}{}
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			userRole, ok := GetUserRole(r.Context())
			if !ok || userRole == "" {
				errors.HandleError(w, errors.NewUnauthorizedError("unauthenticated: no role claim found in token context"))
				return
			}

			if _, allowed := allowedMap[strings.ToUpper(userRole)]; !allowed {
				errors.HandleError(w, errors.NewForbiddenError(fmt.Sprintf("forbidden: role '%s' lacks required permissions", userRole)))
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
