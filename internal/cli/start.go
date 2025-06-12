package cli

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/spf13/cobra"

	"github.com/loonghao/webhook_bridge/internal/config"
	"github.com/loonghao/webhook_bridge/internal/logging"
	"github.com/loonghao/webhook_bridge/internal/python"
	"github.com/loonghao/webhook_bridge/internal/server"
)

// NewStartCommand creates the start command
func NewStartCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "start",
		Short: "Start the webhook bridge service",
		Long: `Start all services (frontend, backend, Python executor) in unified mode.
This command starts both Python executor and Go server in a unified process.

Features:
- Automatic Python executor startup
- Integrated Go server with gRPC client
- Unified process management
- Graceful shutdown handling`,
		RunE: runStart,
	}

	cmd.Flags().String("port", "", "Server port (default from config)")
	cmd.Flags().String("host", "", "Server host (default from config)")
	cmd.Flags().String("config", "", "Configuration file path")
	cmd.Flags().String("mode", "", "Server mode (debug, release)")
	cmd.Flags().String("log-level", "", "Log level (debug, info, warn, error)")
	cmd.Flags().Bool("no-python", false, "Skip Python executor startup (API-only mode)")
	cmd.Flags().Bool("stagewise", false, "Enable stagewise debugging mode")

	return cmd
}

func runStart(cmd *cobra.Command, args []string) error {
	verbose, _ := cmd.Flags().GetBool("verbose")
	port, _ := cmd.Flags().GetString("port")
	host, _ := cmd.Flags().GetString("host")
	configPath, _ := cmd.Flags().GetString("config")
	mode, _ := cmd.Flags().GetString("mode")
	logLevel, _ := cmd.Flags().GetString("log-level")
	noPython, _ := cmd.Flags().GetBool("no-python")
	stagewise, _ := cmd.Flags().GetBool("stagewise")

	if verbose {
		fmt.Printf("🚀 Starting webhook bridge service...\n")
		fmt.Printf("📊 Configuration: %s\n", configPath)
		fmt.Printf("🌐 Host: %s\n", host)
		fmt.Printf("🌐 Port: %s\n", port)
		fmt.Printf("🔧 Mode: %s\n", mode)
		fmt.Printf("📝 Log Level: %s\n", logLevel)
		fmt.Printf("🐍 Python Executor: %v\n", !noPython)
		if stagewise {
			fmt.Printf("🎭 Stagewise Debug: enabled\n")
		}
	}

	return runUnifiedService(verbose, port, host, configPath, mode, logLevel, noPython, stagewise)
}

// runUnifiedService runs the unified webhook bridge service
func runUnifiedService(verbose bool, port, host, configPath, mode, logLevel string, noPython, stagewise bool) error {
	// Get current working directory
	workingDir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get working directory: %w", err)
	}

	// Create configuration manager
	configManager := config.NewConfigManager(workingDir, configPath, verbose)

	// Validate working directory
	if err := configManager.ValidateWorkingDirectory(); err != nil {
		return fmt.Errorf("working directory validation failed: %w", err)
	}

	// Load configuration
	cfg, err := configManager.LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	// Override configuration with command line flags
	if host != "" {
		cfg.Server.Host = host
	}
	if port != "" {
		if portInt, parseErr := parsePort(port); parseErr == nil {
			cfg.Server.Port = portInt
		} else {
			return fmt.Errorf("invalid port: %s", port)
		}
	}
	if mode != "" {
		cfg.Server.Mode = mode
	}
	if logLevel != "" {
		cfg.Logging.Level = logLevel
	}

	return startUnifiedServices(cfg, configManager, noPython, stagewise, verbose)
}

// startUnifiedServices starts all services in unified mode
func startUnifiedServices(cfg *config.Config, configManager *config.ConfigManager, noPython, stagewise, verbose bool) error {
	// Setup configuration environment
	if err := configManager.SetupConfigEnvironment(cfg); err != nil {
		return fmt.Errorf("failed to setup configuration environment: %w", err)
	}

	// Setup logging system
	dirManager := configManager.GetDirectoryManager()
	var logManager *logging.Manager
	if dirManager != nil {
		logManager = logging.NewManager(&cfg.Logging, dirManager, verbose)
		if err := logManager.SetupLoggingEnvironment(); err != nil {
			log.Printf("Warning: Failed to setup logging: %v", err)
		} else {
			logManager.LogStartup("2.0.0-unified", time.Now().Format(time.RFC3339))
		}
		defer func() {
			if logManager != nil {
				logManager.Close()
			}
		}()
	}

	// Context for managing services
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var pythonCmd *exec.Cmd

	// Step 1: Start Python executor (if enabled)
	if !noPython {
		pythonCmd = startPythonExecutorService(ctx, cfg, verbose)
	}

	// Step 1.5: Build frontend with stagewise if enabled
	if stagewise {
		if err := buildFrontendWithStagewise(verbose); err != nil {
			if verbose {
				fmt.Printf("⚠️  Frontend build failed: %v\n", err)
				fmt.Printf("🔧 Continuing without frontend build\n")
			}
		}
	}

	// Step 2: Start Go server (integrated)
	return startGoServerService(cfg, pythonCmd, verbose)
}

// startPythonExecutorService starts Python executor as a service component
func startPythonExecutorService(ctx context.Context, cfg *config.Config, verbose bool) *exec.Cmd {
	if verbose {
		fmt.Printf("🐍 Step 1: Starting Python executor service...\n")
	}

	// Detect Python interpreter
	pythonResult, err := python.DetectPythonInterpreter(&cfg.Python, verbose)
	if err != nil {
		if verbose {
			fmt.Printf("⚠️  Python detection failed: %v\n", err)
			fmt.Printf("🔧 Continuing without Python executor (API-only mode)\n")
		}
		return nil
	}

	pythonCmd, err := startPythonExecutorUnified(ctx, cfg, pythonResult, verbose)
	if err != nil {
		if verbose {
			fmt.Printf("⚠️  Failed to start Python executor: %v\n", err)
			fmt.Printf("🔧 Continuing without Python executor (API-only mode)\n")
		}
		return nil
	}

	// Wait for Python executor to initialize
	if verbose {
		fmt.Printf("⏳ Waiting for Python executor to initialize...\n")
	}
	time.Sleep(3 * time.Second)

	if verbose {
		fmt.Printf("✅ Python executor started on port %d\n", cfg.Executor.Port)
	}

	return pythonCmd
}

// startGoServerService starts Go server as an integrated service
func startGoServerService(cfg *config.Config, pythonCmd *exec.Cmd, verbose bool) error {
	if verbose {
		fmt.Printf("🚀 Step 2: Starting Go server (integrated)...\n")
	}

	// Create Gin router
	if cfg.Server.Mode == "release" {
		gin.SetMode(gin.ReleaseMode)
	}
	router := gin.New()

	// Create server
	srv := server.New(cfg)

	// Start gRPC connection
	if err := srv.Start(); err != nil {
		if verbose {
			fmt.Printf("⚠️  gRPC connection failed: %v\n", err)
			if pythonCmd == nil {
				fmt.Printf("🔧 This is expected in API-only mode\n")
			} else {
				fmt.Printf("🔧 Server will start with limited functionality\n")
			}
		}
	} else {
		if verbose {
			fmt.Printf("✅ gRPC connection established\n")
		}
	}
	defer srv.Stop()

	// Setup routes
	srv.SetupRoutes(router)

	// Create HTTP server
	httpServer := &http.Server{
		Addr:              cfg.GetServerAddress(),
		Handler:           router,
		ReadHeaderTimeout: 30 * time.Second,
		ReadTimeout:       60 * time.Second,
		WriteTimeout:      60 * time.Second,
		IdleTimeout:       120 * time.Second,
	}

	return runHTTPServer(httpServer, cfg, pythonCmd, verbose)
}

// runHTTPServer runs the HTTP server with graceful shutdown
func runHTTPServer(httpServer *http.Server, cfg *config.Config, pythonCmd *exec.Cmd, verbose bool) error {
	// Start HTTP server in goroutine
	go func() {
		fmt.Printf("🚀 Webhook bridge service started!\n")
		fmt.Printf("🌐 Server: http://localhost:%d/\n", cfg.Server.Port)
		fmt.Printf("📊 Dashboard: http://localhost:%d/dashboard\n", cfg.Server.Port)
		fmt.Printf("📊 API: http://localhost:%d/api/v1\n", cfg.Server.Port)
		fmt.Printf("❤️  Health: http://localhost:%d/health\n", cfg.Server.Port)
		if pythonCmd != nil {
			fmt.Printf("🐍 Python Executor: localhost:%d (gRPC)\n", cfg.Executor.Port)
		}
		fmt.Printf("💡 Press Ctrl+C to stop all services\n")

		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("HTTP server error: %v", err)
		}
	}()

	// Setup cleanup for Python process
	if pythonCmd != nil {
		defer func() {
			if pythonCmd.Process != nil {
				if verbose {
					fmt.Printf("🛑 Stopping Python executor...\n")
				}
				if err := pythonCmd.Process.Kill(); err != nil && verbose {
					fmt.Printf("⚠️  Failed to kill Python process: %v\n", err)
				}
			}
		}()
	}

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-quit

	if verbose {
		fmt.Printf("\n🛑 Shutting down service...\n")
	}

	// Graceful shutdown
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()

	if err := httpServer.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("server forced to shutdown: %w", err)
	}

	fmt.Printf("✅ Service stopped gracefully\n")
	return nil
}

// startPythonExecutorUnified starts Python executor as a service component
func startPythonExecutorUnified(ctx context.Context, cfg *config.Config, pythonResult *python.DetectionResult, verbose bool) (*exec.Cmd, error) {
	pythonPath := pythonResult.Interpreter.Path

	// Prepare command arguments
	args := []string{"python_executor/main.py"}
	args = append(args, "--host", cfg.Executor.Host)
	args = append(args, "--port", fmt.Sprintf("%d", cfg.Executor.Port))

	// Add plugin directories
	if len(cfg.Python.PluginDirs) > 0 {
		for _, dir := range cfg.Python.PluginDirs {
			args = append(args, "--plugin-dirs", dir)
		}
	} else {
		args = append(args, "--plugin-dirs", "example_plugins")
	}

	// Create command
	cmd := exec.CommandContext(ctx, pythonPath, args...)
	cmd.Dir = "."

	// Set environment for virtual environment if needed
	env := os.Environ()
	if pythonResult.Interpreter.IsVirtual {
		env = append(env, fmt.Sprintf("VIRTUAL_ENV=%s", pythonResult.Interpreter.VenvPath))
	}
	cmd.Env = env

	// Setup output handling
	if verbose {
		fmt.Printf("🐍 Python path: %s\n", pythonPath)
		fmt.Printf("🔧 Executor address: %s:%d\n", cfg.Executor.Host, cfg.Executor.Port)
		fmt.Printf("📁 Plugin directories: %v\n", cfg.Python.PluginDirs)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	}

	// Start the process
	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("failed to start Python executor: %w", err)
	}

	return cmd, nil
}

// buildFrontendWithStagewise builds the frontend with stagewise debugging enabled
func buildFrontendWithStagewise(verbose bool) error {
	// Check if we're in the correct directory
	if _, err := os.Stat("web-nextjs"); err != nil {
		return fmt.Errorf("web-nextjs directory not found")
	}

	// Check if npm is available
	if _, err := exec.LookPath("npm"); err != nil {
		return fmt.Errorf("npm not found in PATH")
	}

	if verbose {
		fmt.Printf("📦 Building frontend with stagewise debug enabled...\n")
	}

	// Set environment variables for stagewise build
	env := os.Environ()
	env = append(env, "NEXT_PUBLIC_ENABLE_STAGEWISE=true")
	env = append(env, "NEXT_PUBLIC_DEBUG_MODE=true")

	// Run npm run build:debug in web-nextjs directory
	cmd := exec.Command("npm", "run", "build:debug")
	cmd.Dir = "web-nextjs"
	cmd.Env = env

	if verbose {
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		fmt.Printf("🔧 Running: npm run build:debug in web-nextjs directory\n")
		fmt.Printf("🌍 Environment: NEXT_PUBLIC_ENABLE_STAGEWISE=true, NEXT_PUBLIC_DEBUG_MODE=true\n")
	}

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("npm run build:debug failed: %w", err)
	}

	if verbose {
		fmt.Printf("✅ Frontend stagewise build completed successfully\n")
	}

	return nil
}
