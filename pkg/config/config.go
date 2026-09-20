package config

import (
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

// Config holds all configuration for the application.
type Config struct {
	Port                      string
	DatabaseURL               string
	Env                       string
	JWTAccessSecret           string
	JWTRefreshSecret          string
	AuthTokenSecret           string
	CookieSecret              string
	JWTAccessExpires          time.Duration
	JWTRefreshExpires         time.Duration
	JWTRefreshRememberExpires time.Duration
	JWTIssuer                 string
	JWTAudience               string
	AuthTokenTransport        string // "cookie" or "bearer"
}

func parseDurationWithDays(str string, fallback time.Duration) time.Duration {
	str = strings.TrimSpace(str)
	if str == "" {
		return fallback
	}

	if strings.HasSuffix(str, "d") {
		daysStr := strings.TrimSuffix(str, "d")
		days, err := strconv.Atoi(daysStr)
		if err == nil {
			return time.Duration(days) * 24 * time.Hour
		}
	}

	d, err := time.ParseDuration(str)
	if err != nil {
		return fallback
	}
	return d
}

// LoadConfig reads configuration from environment variables and .env file.
func LoadConfig() *Config {
	err := godotenv.Load()
	if err != nil {
		log.Println("[Config] Notice: No .env file found or unable to load, using system environment variables")
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		databaseURL = "postgresql://postgres:postgres@localhost:5432/go_backend?sslmode=disable"
	}

	// Auto-append pgbouncer=true for Neon pooler connection strings if missing
	if strings.Contains(databaseURL, "-pooler") && !strings.Contains(databaseURL, "pgbouncer=true") {
		if strings.Contains(databaseURL, "?") {
			databaseURL += "&pgbouncer=true"
		} else {
			databaseURL += "?pgbouncer=true"
		}
	}

	env := os.Getenv("ENV")
	if env == "" {
		env = "development"
	}

	jwtAccessSecret := os.Getenv("JWT_ACCESS_SECRET")
	if jwtAccessSecret == "" {
		jwtAccessSecret = "jwt-access-secret-at-least-32-chars-long-2026!"
	}

	jwtRefreshSecret := os.Getenv("JWT_REFRESH_SECRET")
	if jwtRefreshSecret == "" {
		jwtRefreshSecret = "jwt-refresh-secret-at-least-32-chars-long-2026!"
	}

	authTokenSecret := os.Getenv("AUTH_TOKEN_SECRET")
	if authTokenSecret == "" {
		authTokenSecret = "auth-challenge-hmac-secret-at-least-32-chars!"
	}

	cookieSecret := os.Getenv("COOKIE_SECRET")
	if cookieSecret == "" {
		cookieSecret = "cookie-secret-at-least-32-chars-long-2026!"
	}

	jwtAccessExpires := parseDurationWithDays(os.Getenv("JWT_ACCESS_EXPIRES"), 15*time.Minute)
	jwtRefreshExpires := parseDurationWithDays(os.Getenv("JWT_REFRESH_EXPIRES"), 7*24*time.Hour)
	jwtRefreshRememberExpires := parseDurationWithDays(os.Getenv("JWT_REFRESH_REMEMBER_EXPIRES"), 30*24*time.Hour)

	jwtIssuer := os.Getenv("JWT_ISSUER")
	if jwtIssuer == "" {
		jwtIssuer = "enterprise-backend"
	}

	jwtAudience := os.Getenv("JWT_AUDIENCE")
	if jwtAudience == "" {
		jwtAudience = "enterprise-api"
	}

	authTokenTransport := strings.ToLower(os.Getenv("AUTH_TOKEN_TRANSPORT"))
	if authTokenTransport == "" {
		authTokenTransport = "cookie"
	}

	return &Config{
		Port:                      port,
		DatabaseURL:               databaseURL,
		Env:                       env,
		JWTAccessSecret:           jwtAccessSecret,
		JWTRefreshSecret:          jwtRefreshSecret,
		AuthTokenSecret:           authTokenSecret,
		CookieSecret:              cookieSecret,
		JWTAccessExpires:          jwtAccessExpires,
		JWTRefreshExpires:         jwtRefreshExpires,
		JWTRefreshRememberExpires: jwtRefreshRememberExpires,
		JWTIssuer:                 jwtIssuer,
		JWTAudience:               jwtAudience,
		AuthTokenTransport:        authTokenTransport,
	}
}
