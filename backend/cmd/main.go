package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/ayush/supportiq/internal/config"
	"github.com/ayush/supportiq/internal/database"
	"github.com/ayush/supportiq/internal/routes"
	"github.com/ayush/supportiq/internal/utils"
	"github.com/gin-gonic/gin"
)

func main() {
	// Load configuration from environment
	cfg, err := config.Load()
	if err != nil {
		utils.Logger.Fatalf("Failed to load configuration: %v", err)
	}

	// Set Gin mode before creating the router
	if cfg.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	// Establish database connection
	db, err := database.Connect(cfg)
	if err != nil {
		utils.Logger.Fatalf("Failed to connect to database: %v", err)
	}

	// Create a context that is cancelled when the process receives a shutdown signal.
	// Background goroutines (SLA monitor, analytics scheduler, email workers) use this
	// context so they stop cleanly instead of leaking after SIGTERM.
	serverCtx, stopServer := context.WithCancel(context.Background())

	// Build router with all routes and middleware
	router := routes.SetupRouter(db, cfg, serverCtx)

	srv := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server in a goroutine so it does not block the shutdown listener
	go func() {
		utils.Logger.Infof("Server listening on port %s (env: %s)", cfg.Port, cfg.Environment)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			utils.Logger.Fatalf("Server failed: %v", err)
		}
	}()

	// Block until OS interrupt or termination signal is received
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	utils.Logger.Info("Shutdown signal received, draining connections...")

	// Cancel the server context so all background goroutines exit.
	stopServer()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		utils.Logger.Fatalf("Forced shutdown due to error: %v", err)
	}

	utils.Logger.Info("Server exited cleanly")
}
