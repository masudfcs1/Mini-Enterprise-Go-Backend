package auth

import (
	"encoding/json"
	"net/http"

	"go-mini-setup/pkg/config"
	"go-mini-setup/pkg/errors"
	"go-mini-setup/pkg/middleware"
	"go-mini-setup/pkg/response"
)

// Handler handles HTTP requests for auth endpoints.
type Handler struct {
	service AuthService
	cfg     *config.Config
}

// NewHandler creates a new auth HTTP Handler.
func NewHandler(service AuthService, cfg *config.Config) *Handler {
	return &Handler{
		service: service,
		cfg:     cfg,
	}
}

// setAuthCookies sets HttpOnly cookies for access and refresh tokens.
func (h *Handler) setAuthCookies(w http.ResponseWriter, authRes *AuthResponse) {
	if h.cfg.AuthTokenTransport != "cookie" {
		return
	}

	isSecure := h.cfg.Env == "production"

	// 1. Access Token Cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "access_token",
		Value:    authRes.AccessToken,
		Path:     "/",
		MaxAge:   int(h.cfg.JWTAccessExpires.Seconds()),
		HttpOnly: true,
		Secure:   isSecure,
		SameSite: http.SameSiteLaxMode,
	})

	// 2. Refresh Token Cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "refresh_token",
		Value:    authRes.RefreshToken,
		Path:     "/api/v1/auth",
		MaxAge:   int(h.cfg.JWTRefreshRememberExpires.Seconds()),
		HttpOnly: true,
		Secure:   isSecure,
		SameSite: http.SameSiteLaxMode,
	})
}

// clearAuthCookies expires and removes authentication cookies.
func (h *Handler) clearAuthCookies(w http.ResponseWriter) {
	isSecure := h.cfg.Env == "production"

	http.SetCookie(w, &http.Cookie{
		Name:     "access_token",
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   isSecure,
		SameSite: http.SameSiteLaxMode,
	})

	http.SetCookie(w, &http.Cookie{
		Name:     "refresh_token",
		Value:    "",
		Path:     "/api/v1/auth",
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   isSecure,
		SameSite: http.SameSiteLaxMode,
	})
}

// Register handles POST /api/v1/auth/register.
func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	var req RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errors.HandleError(w, errors.NewBadRequestError("invalid request body JSON"))
		return
	}

	authRes, err := h.service.Register(r.Context(), req)
	if err != nil {
		errors.HandleError(w, err)
		return
	}

	h.setAuthCookies(w, authRes)
	response.Success(w, http.StatusCreated, "User registered successfully", authRes)
}

// Login handles POST /api/v1/auth/login.
func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errors.HandleError(w, errors.NewBadRequestError("invalid request body JSON"))
		return
	}

	authRes, err := h.service.Login(r.Context(), req)
	if err != nil {
		errors.HandleError(w, err)
		return
	}

	h.setAuthCookies(w, authRes)
	response.Success(w, http.StatusOK, "Login successful", authRes)
}

// Refresh handles POST /api/v1/auth/refresh (supports cookie or request body).
func (h *Handler) Refresh(w http.ResponseWriter, r *http.Request) {
	tokenString := ""

	// 1. Try reading refresh_token from HttpOnly Cookie
	if cookie, err := r.Cookie("refresh_token"); err == nil && cookie.Value != "" {
		tokenString = cookie.Value
	}

	// 2. Fallback to JSON request body
	if tokenString == "" && r.Body != nil {
		var req RefreshTokenRequest
		_ = json.NewDecoder(r.Body).Decode(&req)
		tokenString = req.RefreshToken
	}

	if tokenString == "" {
		errors.HandleError(w, errors.NewUnauthorizedError("refresh token not provided in cookie or body"))
		return
	}

	authRes, err := h.service.RefreshToken(r.Context(), tokenString)
	if err != nil {
		errors.HandleError(w, err)
		return
	}

	h.setAuthCookies(w, authRes)
	response.Success(w, http.StatusOK, "Token refreshed successfully", authRes)
}

// Logout handles POST /api/v1/auth/logout.
func (h *Handler) Logout(w http.ResponseWriter, r *http.Request) {
	tokenString := ""

	if cookie, err := r.Cookie("refresh_token"); err == nil {
		tokenString = cookie.Value
	}

	if tokenString == "" && r.Body != nil {
		var req LogoutRequest
		_ = json.NewDecoder(r.Body).Decode(&req)
		tokenString = req.RefreshToken
	}

	_ = h.service.Logout(r.Context(), tokenString)
	h.clearAuthCookies(w)

	response.Success(w, http.StatusOK, "Logged out successfully", nil)
}

// GetMe handles GET /api/v1/auth/me (JWT protected).
func (h *Handler) GetMe(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok || userID == "" {
		errors.HandleError(w, errors.NewUnauthorizedError("unauthorized: missing or invalid user context"))
		return
	}

	meRes, err := h.service.GetMe(r.Context(), userID)
	if err != nil {
		errors.HandleError(w, err)
		return
	}

	response.Success(w, http.StatusOK, "Current user profile fetched", meRes)
}
