package main

import (
	"context"
	"errors"
	"fmt"
	"log"
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
)

func main() {
	// 1. Load Configuration
	cfg := config.LoadConfig()
	log.Printf("[Server] Starting application in %s mode on port :%s", cfg.Env, cfg.Port)

	// 2. Connect Database (Prisma Client Go)
	dbInstance, err := database.New()
	if err != nil {
		log.Fatalf("[Server] Database connection error: %v", err)
	}
	defer func() {
		if err := dbInstance.Close(); err != nil {
			log.Printf("[Server] Error closing database: %v", err)
		}
	}()

	// 3. Initialize Repositories
	userRepo := user.NewUserRepository(dbInstance.Client)
	authRepo := auth.NewAuthRepository(dbInstance.Client)

	// 4. Initialize Services
	userService := user.NewUserService(userRepo)
	authService := auth.NewAuthService(authRepo)

	// 5. Initialize Handlers
	userHandler := user.NewHandler(userService)
	authHandler := auth.NewHandler(authService)

	// 6. Build Global Router
	handler := router.NewRouter(&router.Handlers{
		User: userHandler,
		Auth: authHandler,
	})

	// 7. Configure HTTP Server
	server := &http.Server{
		Addr:         fmt.Sprintf(":%s", cfg.Port),
		Handler:      handler,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// 8. Run Server in Background Goroutine
	serverErrors := make(chan error, 1)
	go func() {
		log.Printf("[Server] HTTP server is listening on http://localhost:%s", cfg.Port)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			serverErrors <- err
		}
	}()

	// 9. Listen for Graceful Shutdown Signals
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)

	select {
	case err := <-serverErrors:
		log.Fatalf("[Server] Failed to start server: %v", err)

	case sig := <-shutdown:
		log.Printf("[Server] Shutdown signal received (%s), starting graceful shutdown...", sig)

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		if err := server.Shutdown(ctx); err != nil {
			log.Printf("[Server] Graceful shutdown failed, forcing server close: %v", err)
			_ = server.Close()
		}

		log.Println("[Server] Server stopped successfully")
	}
}
