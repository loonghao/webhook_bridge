package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/loonghao/webhook_bridge/internal/config"
	"github.com/loonghao/webhook_bridge/internal/server"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Assign ports automatically if needed
	if err := cfg.AssignPorts(); err != nil {
		log.Fatalf("Failed to assign ports: %v", err)
	}

	// Validate configuration
	if err := cfg.Validate(); err != nil {
		log.Fatalf("Invalid configuration: %v", err)
	}

	// Log the assigned ports
	log.Printf("Server will start on: %s", cfg.GetServerAddress())
	log.Printf("Executor expected at: %s", cfg.GetExecutorAddress())

	// Create Gin router
	if cfg.Server.Mode == "release" {
		gin.SetMode(gin.ReleaseMode)
	}

	// Create router without default middleware (we'll add our own)
	router := gin.New()

	// Setup server
	srv := server.New(cfg)

	// Start gRPC connection (non-fatal if it fails)
	if err := srv.Start(); err != nil {
		log.Printf("⚠️  Failed to start gRPC connection: %v", err)
		log.Printf("🔧 Server will start in API-only mode (limited functionality)")
	}
	defer srv.Stop()

	srv.SetupRoutes(router)

	// Create HTTP server with security timeouts
	httpServer := &http.Server{
		Addr:              cfg.GetServerAddress(),
		Handler:           router,
		ReadHeaderTimeout: 30 * time.Second, // Prevent Slowloris attacks
		ReadTimeout:       60 * time.Second,
		WriteTimeout:      60 * time.Second,
		IdleTimeout:       120 * time.Second,
	}

	// Start server in a goroutine
	go func() {
		log.Printf("🚀 Webhook bridge server started successfully on %s", httpServer.Addr)
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")

	// Give outstanding requests a deadline for completion
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := httpServer.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exited")
}
