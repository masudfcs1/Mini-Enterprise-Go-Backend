package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

// Config holds all configuration for the application.
type Config struct {
	Port        string
	DatabaseURL string
	Env         string
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

	env := os.Getenv("ENV")
	if env == "" {
		env = "development"
	}

	return &Config{
		Port:        port,
		DatabaseURL: databaseURL,
		Env:         env,
	}
}
