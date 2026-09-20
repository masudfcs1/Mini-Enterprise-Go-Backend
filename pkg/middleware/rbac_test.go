package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRBAC_RequireRole(t *testing.T) {
	adminOnlyHandler := RequireRole("ADMIN", "SUPER_ADMIN")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("admin access granted"))
	}))

	// 1. Missing context -> 401
	req1 := httptest.NewRequest(http.MethodGet, "/admin", nil)
	w1 := httptest.NewRecorder()
	adminOnlyHandler.ServeHTTP(w1, req1)
	if w1.Code != http.StatusUnauthorized {
		t.Errorf("expected status 401, got %d", w1.Code)
	}

	// 2. Role USER -> 403 Forbidden
	req2 := httptest.NewRequest(http.MethodGet, "/admin", nil)
	ctx2 := context.WithValue(req2.Context(), UserRoleKey, "USER")
	w2 := httptest.NewRecorder()
	adminOnlyHandler.ServeHTTP(w2, req2.WithContext(ctx2))
	if w2.Code != http.StatusForbidden {
		t.Errorf("expected status 403, got %d", w2.Code)
	}

	// 3. Role ADMIN -> 200 OK
	req3 := httptest.NewRequest(http.MethodGet, "/admin", nil)
	ctx3 := context.WithValue(req3.Context(), UserRoleKey, "ADMIN")
	w3 := httptest.NewRecorder()
	adminOnlyHandler.ServeHTTP(w3, req3.WithContext(ctx3))
	if w3.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w3.Code)
	}
}
