package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go-mini-setup/internal/auth"
	"go-mini-setup/internal/database"
	"go-mini-setup/internal/router"
	"go-mini-setup/internal/user"
	"go-mini-setup/pkg/config"
	"go-mini-setup/pkg/logger"
)

func main() {
	// 1. Load Configuration
	cfg := config.LoadConfig()

	// 2. Initialize Pino-Style Structured & Colorful Logger
	logger.InitLogger(cfg.Env)

	// 3. Connect Database (Prisma Client Go with PgBouncer Support)
	dbInstance, err := database.New()
	if err != nil {
		logger.Log.Fatal().Err(err).Msg("[Database] Failed to connect to database via Prisma")
	}
	defer func() {
		if err := dbInstance.Close(); err != nil {
			logger.Log.Error().Err(err).Msg("[Database] Error closing database connection")
		}
	}()

	// 4. Print Vibrant Startup Banner
	dbTarget := "PostgreSQL (PgBouncer Enabled)"
	logger.PrintBanner(cfg.Port, cfg.Env, dbTarget)

	// 5. Initialize Repositories
	userRepo := user.NewUserRepository(dbInstance.Client)
	authRepo := auth.NewAuthRepository(dbInstance.Client)

	// 6. Initialize Services
	userService := user.NewUserService(userRepo)
	authService := auth.NewAuthService(authRepo, cfg.JWTSecret, cfg.JWTTTL)

	// 7. Initialize Handlers
	userHandler := user.NewHandler(userService)
	authHandler := auth.NewHandler(authService)

	// 8. Build Global Router with Middlewares
	handler := router.NewRouter(&router.Handlers{
		User:      userHandler,
		Auth:      authHandler,
		JWTSecret: cfg.JWTSecret,
	})

	// 9. Configure HTTP Server
	server := &http.Server{
		Addr:         fmt.Sprintf(":%s", cfg.Port),
		Handler:      handler,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// 10. Run Server in Background Goroutine
	serverErrors := make(chan error, 1)
	go func() {
		logger.Log.Info().Str("port", cfg.Port).Msg("[Server] HTTP server ready to accept incoming requests")
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			serverErrors <- err
		}
	}()

	// 11. Listen for Graceful Shutdown Signals (SIGINT, SIGTERM)
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)

	select {
	case err := <-serverErrors:
		logger.Log.Fatal().Err(err).Msg("[Server] Fatal error encountered while starting server")

	case sig := <-shutdown:
		logger.Log.Info().Str("signal", sig.String()).Msg("[Server] Shutdown signal received, gracefully terminating...")

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		if err := server.Shutdown(ctx); err != nil {
			logger.Log.Error().Err(err).Msg("[Server] Graceful shutdown timed out, forcing shutdown")
			_ = server.Close()
		}

		logger.Log.Info().Msg("[Server] Server stopped successfully. Goodbye!")
	}
}
