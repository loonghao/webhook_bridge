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
	"github.com/loonghao/webhook_bridge/internal/logging"
	"github.com/loonghao/webhook_bridge/internal/server"
)

func main() {
	// Get current working directory
	workingDir, err := os.Getwd()
	if err != nil {
		log.Fatalf("Failed to get working directory: %v", err)
	}

	// Create configuration manager
	configManager := config.NewConfigManager(workingDir, "", true) // verbose=true for server

	// Validate working directory
	if err := configManager.ValidateWorkingDirectory(); err != nil {
		log.Fatalf("Working directory validation failed: %v", err)
	}

	// Load configuration
	cfg, err := configManager.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Setup configuration environment (creates directories, logs, etc.)
	if err := configManager.SetupConfigEnvironment(cfg); err != nil {
		log.Fatalf("Failed to setup configuration environment: %v", err)
	}

	// Setup logging system
	dirManager := configManager.GetDirectoryManager()
	if dirManager != nil {
		logManager := logging.NewManager(&cfg.Logging, dirManager, true)
		if err := logManager.SetupLoggingEnvironment(); err != nil {
			log.Printf("Warning: Failed to setup logging: %v", err)
		} else {
			// Log startup information
			logManager.LogStartup("2.0.0-hybrid", time.Now().Format(time.RFC3339))
		}
		defer logManager.Close()
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
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)

	log.Printf("🚀 Webhook bridge server started successfully on %s", cfg.GetServerAddress())
	log.Println("Press Ctrl+C to stop the server")

	<-quit
	log.Println("🛑 Shutting down server...")

	// Give outstanding requests a deadline for completion
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := httpServer.Shutdown(ctx); err != nil {
		log.Printf("⚠️  Server forced to shutdown: %v", err)
	} else {
		log.Println("✅ Server exited gracefully")
	}
}
