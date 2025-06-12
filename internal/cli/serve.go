package cli

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/spf13/cobra"

	"github.com/loonghao/webhook_bridge/internal/config"
	"github.com/loonghao/webhook_bridge/internal/server"
)

// NewServeCommand creates the serve command for standalone operation
func NewServeCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "serve",
		Short: "Start standalone server (no external dependencies)",
		Long:  "Start the webhook bridge server in standalone mode without requiring separate binaries",
		RunE:  runServe,
	}

	cmd.Flags().StringP("env", "e", "dev", "Environment (dev, prod)")
	cmd.Flags().String("port", "0", "Server port (0 = auto-assign)")
	cmd.Flags().String("executor-port", "0", "Python executor port (0 = auto-assign)")
	cmd.Flags().String("config", "", "Configuration file path")

	return cmd
}

func runServe(cmd *cobra.Command, args []string) error {
	verbose, _ := cmd.Flags().GetBool("verbose")
	env, _ := cmd.Flags().GetString("env")
	port, _ := cmd.Flags().GetString("port")
	executorPort, _ := cmd.Flags().GetString("executor-port")
	configPath, _ := cmd.Flags().GetString("config")

	if verbose {
		fmt.Printf("🚀 Starting webhook bridge server in standalone mode...\n")
		fmt.Printf("📊 Environment: %s\n", env)
		fmt.Printf("🌐 Server port: %s\n", port)
		fmt.Printf("🐍 Executor port: %s\n", executorPort)
	}

	// Load configuration
	cfg, err := loadConfiguration(configPath, env, port, executorPort, verbose)
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	// Assign ports automatically if needed
	if err := cfg.AssignPorts(); err != nil {
		return fmt.Errorf("failed to assign ports: %w", err)
	}

	// Validate configuration
	if err := cfg.Validate(); err != nil {
		return fmt.Errorf("invalid configuration: %w", err)
	}

	if verbose {
		fmt.Printf("📊 Server will start on: %s\n", cfg.GetServerAddress())
		fmt.Printf("🐍 Executor expected at: %s\n", cfg.GetExecutorAddress())
	}

	// Create Gin router
	if cfg.Server.Mode == "release" {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.New()

	// Create server
	srv := server.New(cfg)

	// Start gRPC connection (this will try to connect to Python executor)
	if err := srv.Start(); err != nil {
		if verbose {
			fmt.Printf("⚠️  Warning: Failed to connect to Python executor: %v\n", err)
			fmt.Printf("🔧 Server will start in API-only mode. Python plugins will not be available.\n")
		}
	}
	defer srv.Stop()

	// Setup routes
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
		fmt.Printf("🚀 Webhook bridge server started successfully on %s\n", httpServer.Addr)
		fmt.Printf("🌐 Dashboard: http://localhost:%d/dashboard\n", cfg.Server.Port)
		fmt.Printf("📊 API: http://localhost:%d/api/v1\n", cfg.Server.Port)
		fmt.Printf("❤️  Health: http://localhost:%d/health\n", cfg.Server.Port)

		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	if verbose {
		fmt.Println("\n🛑 Shutting down server...")
	}

	// Give outstanding requests a deadline for completion
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := httpServer.Shutdown(ctx); err != nil {
		return fmt.Errorf("server forced to shutdown: %w", err)
	}

	fmt.Println("✅ Server exited gracefully")
	return nil
}

func loadConfiguration(configPath, env, port, executorPort string, verbose bool) (*config.Config, error) {
	var cfg *config.Config
	var err error

	// Try to load from specified config file
	if configPath != "" {
		if verbose {
			fmt.Printf("📝 Loading configuration from: %s\n", configPath)
		}
		cfg, err = config.LoadFromFile(configPath)
		if err != nil {
			return nil, fmt.Errorf("failed to load config from %s: %w", configPath, err)
		}
	} else {
		// Try to load default configuration
		if verbose {
			fmt.Printf("📝 Loading default configuration...\n")
		}
		cfg, err = config.Load()
		if err != nil {
			// If no config file exists, create a default one
			if verbose {
				fmt.Printf("📝 Creating default configuration...\n")
			}
			cfg = config.Default()
		}
	}

	// Override with command line parameters
	if port != "" && port != "0" {
		if verbose {
			fmt.Printf("🔧 Overriding server port: %s\n", port)
		}
		if portInt, err := strconv.Atoi(port); err == nil {
			cfg.Server.Port = portInt
			cfg.Server.AutoPort = false
		}
	} else if port == "0" {
		if verbose {
			fmt.Printf("🔧 Enabling auto-port assignment for server\n")
		}
		cfg.Server.Port = 0
		cfg.Server.AutoPort = true
	}

	if executorPort != "" && executorPort != "0" {
		if verbose {
			fmt.Printf("🔧 Overriding executor port: %s\n", executorPort)
		}
		if portInt, err := strconv.Atoi(executorPort); err == nil {
			cfg.Executor.Port = portInt
			cfg.Executor.AutoPort = false
		}
	} else if executorPort == "0" {
		if verbose {
			fmt.Printf("🔧 Enabling auto-port assignment for executor\n")
		}
		cfg.Executor.Port = 0
		cfg.Executor.AutoPort = true
	}

	// Set environment-specific defaults
	if env == "prod" {
		cfg.Server.Mode = "release"
		cfg.Logging.Level = "info"
	} else {
		cfg.Server.Mode = "debug"
		cfg.Logging.Level = "debug"
	}

	return cfg, nil
}
