package config

import (
	"log"
	"os"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

// Config holds all configuration for the application.
type Config struct {
	Port        string
	DatabaseURL string
	Env         string
	JWTSecret   string
	JWTTTL      time.Duration
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

	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		jwtSecret = "mini-enterprise-go-jwt-secret-key-32bytes-secure!"
	}

	return &Config{
		Port:        port,
		DatabaseURL: databaseURL,
		Env:         env,
		JWTSecret:   jwtSecret,
		JWTTTL:      24 * time.Hour,
	}
}
