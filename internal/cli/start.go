package cli

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"runtime"
	"syscall"
	"time"

	"github.com/spf13/cobra"

	"github.com/loonghao/webhook_bridge/internal/config"
	"github.com/loonghao/webhook_bridge/internal/python"
)

// NewStartCommand creates the start command
func NewStartCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "start",
		Short: "Start the webhook bridge service",
		Long:  "Start both the Go HTTP server and Python executor service",
		RunE:  runStart,
	}

	cmd.Flags().StringP("env", "e", "dev", "Environment (dev, prod)")
	cmd.Flags().BoolP("daemon", "d", false, "Run as daemon (background)")
	cmd.Flags().Bool("build", false, "Build before starting (default: auto-detect)")
	cmd.Flags().Bool("force-build", false, "Force build even if binaries exist")
	cmd.Flags().StringP("config", "c", "", "Configuration file path")
	cmd.Flags().String("server-port", "", "Override server port")
	cmd.Flags().String("executor-port", "", "Override executor port")

	return cmd
}

func runStart(cmd *cobra.Command, args []string) error {
	verbose, _ := cmd.Flags().GetBool("verbose")
	env, _ := cmd.Flags().GetString("env")
	daemon, _ := cmd.Flags().GetBool("daemon")
	build, _ := cmd.Flags().GetBool("build")
	forceBuild, _ := cmd.Flags().GetBool("force-build")
	configPath, _ := cmd.Flags().GetString("config")
	serverPort, _ := cmd.Flags().GetString("server-port")
	executorPort, _ := cmd.Flags().GetString("executor-port")

	if verbose {
		fmt.Printf("🚀 Starting webhook bridge in %s mode...\n", env)
	}

	// Step 1: Intelligent configuration loading
	workingDir, _ := os.Getwd()
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

	// Override with command line parameters
	if serverPort != "" {
		if port, err := parsePort(serverPort); err == nil {
			cfg.Server.Port = port
			cfg.Server.AutoPort = false
		}
	}
	if executorPort != "" {
		if port, err := parsePort(executorPort); err == nil {
			cfg.Executor.Port = port
			cfg.Executor.AutoPort = false
		}
	}

	// Set environment-specific defaults
	if env == "prod" {
		cfg.Server.Mode = "release"
		cfg.Logging.Level = "info"
	}

	// Setup configuration environment
	if err := configManager.SetupConfigEnvironment(cfg); err != nil {
		return fmt.Errorf("failed to setup configuration environment: %w", err)
	}

	if verbose {
		fmt.Printf("\n%s\n\n", configManager.GetConfigSummary(cfg))
	}

	// Step 2: Intelligent Python interpreter detection
	pythonResult, err := python.DetectPythonInterpreter(&cfg.Python, verbose)
	if err != nil {
		return fmt.Errorf("Python interpreter detection failed: %w", err)
	}

	if verbose {
		fmt.Printf("🐍 Python interpreter detected: %s\n", pythonResult.Interpreter.Path)
		fmt.Printf("📋 Python version: %s\n", pythonResult.Interpreter.Version)
		if pythonResult.Interpreter.IsVirtual {
			fmt.Printf("🏠 Virtual environment: %s\n", pythonResult.Interpreter.VenvPath)
		}
		if pythonResult.UVAvailable {
			fmt.Printf("⚡ UV available: Yes\n")
		}
		fmt.Printf("\n")
	}

	// Step 3: Determine if we need to build
	needsBuild := forceBuild || build || shouldAutoBuild(verbose)

	if needsBuild {
		if verbose {
			fmt.Println("🔨 Building...")
		}
		buildCmd := NewBuildCommand()
		if err := buildCmd.RunE(buildCmd, []string{}); err != nil {
			if forceBuild || build {
				return fmt.Errorf("build failed: %w", err)
			}
			// If auto-build fails, try to continue with existing binaries
			if verbose {
				fmt.Printf("⚠️  Build failed, attempting to use existing binaries...\n")
			}
		}
	}

	if daemon {
		return runAsDaemon(cfg, pythonResult, verbose)
	}

	// Step 4: Start services
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start Python executor with detected interpreter
	pythonCmd, err := startPythonExecutorWithConfig(ctx, cfg, pythonResult, verbose)
	if err != nil {
		return fmt.Errorf("failed to start Python executor: %w", err)
	}
	defer func() {
		if pythonCmd.Process != nil {
			pythonCmd.Process.Kill()
		}
	}()

	// Wait for Python executor to start
	time.Sleep(2 * time.Second)

	// Start Go server
	goCmd, err := startGoServerWithConfig(ctx, cfg, verbose)
	if err != nil {
		return fmt.Errorf("failed to start Go server: %w", err)
	}
	defer func() {
		if goCmd.Process != nil {
			goCmd.Process.Kill()
		}
	}()

	if verbose {
		fmt.Printf("✅ Webhook bridge started successfully!\n")
		fmt.Printf("🌐 Server: http://localhost:%d\n", cfg.Server.Port)
		fmt.Printf("🐍 Executor: localhost:%d\n", cfg.Executor.Port)
		fmt.Printf("📊 Health: http://localhost:%d/health\n", cfg.Server.Port)
		fmt.Printf("🎛️  Dashboard: http://localhost:%d/dashboard\n", cfg.Server.Port)
	}

	// Wait for interrupt signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	fmt.Println("\n🛑 Shutting down...")
	return nil
}

// shouldAutoBuild determines if we should automatically build based on missing binaries
func shouldAutoBuild(verbose bool) bool {
	// Check if required binaries exist
	requiredBinaries := []string{
		"bin/webhook-bridge-server",
		"bin/webhook-bridge-server.exe", // Windows
		"bin/python-manager",
		"bin/python-manager.exe", // Windows
	}

	// Check if any of the binaries exist
	hasAnyBinary := false
	for _, binary := range requiredBinaries {
		if _, err := os.Stat(binary); err == nil {
			hasAnyBinary = true
			break
		}
	}

	// Check if Python environment exists
	hasPythonEnv := false
	if _, err := os.Stat(".venv"); err == nil {
		hasPythonEnv = true
	}

	shouldBuild := !hasAnyBinary || !hasPythonEnv

	if verbose && shouldBuild {
		if !hasAnyBinary {
			fmt.Println("📦 No existing binaries found, will build automatically")
		}
		if !hasPythonEnv {
			fmt.Println("🐍 No Python environment found, will setup automatically")
		}
	}

	return shouldBuild
}

func setupConfiguration(env string, verbose bool) error {
	configFile := "config.yaml"

	// Check if config file exists
	if _, err := os.Stat(configFile); os.IsNotExist(err) {
		if verbose {
			fmt.Printf("📝 Creating configuration for %s environment...\n", env)
		}

		var sourceConfig string
		switch env {
		case "prod":
			sourceConfig = "config.prod.yaml"
		case "dev":
			sourceConfig = "config.dev.yaml"
		default:
			sourceConfig = "config.example.yaml"
		}

		// Copy configuration file
		if err := copyFile(sourceConfig, configFile); err != nil {
			return fmt.Errorf("failed to copy configuration: %w", err)
		}
	}

	return nil
}

func startPythonExecutor(ctx context.Context, port string, verbose bool) (*exec.Cmd, error) {
	// Find Python executable in virtual environment
	var pythonPath string
	if runtime.GOOS == "windows" {
		pythonPath = filepath.Join(".venv", "Scripts", "python.exe")
	} else {
		pythonPath = filepath.Join(".venv", "bin", "python")
	}

	// Check if Python executable exists
	if _, err := os.Stat(pythonPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("Python virtual environment not found. Run 'webhook-bridge build' first")
	}

	// Prepare command arguments
	args := []string{"python_executor/main.py", "--plugin-dirs", "example_plugins"}
	if port != "" {
		args = append(args, "--port", port)
	}

	// Create command
	cmd := exec.CommandContext(ctx, pythonPath, args...)
	cmd.Dir = "."

	if verbose {
		fmt.Println("🐍 Starting Python executor...")
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	}

	// Start the process
	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("failed to start Python executor: %w", err)
	}

	return cmd, nil
}

func startGoServer(ctx context.Context, serverPort, executorPort string, verbose bool) (*exec.Cmd, error) {
	// Find Go server executable
	var serverPath string
	if runtime.GOOS == "windows" {
		serverPath = filepath.Join("build", "webhook-bridge-server.exe")
	} else {
		serverPath = filepath.Join("build", "webhook-bridge-server")
	}

	// Check if server executable exists
	if _, err := os.Stat(serverPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("Go server executable not found. Run 'webhook-bridge build' first")
	}

	// Create command
	cmd := exec.CommandContext(ctx, serverPath)
	cmd.Dir = "."

	// Set environment variables for port overrides
	env := os.Environ()
	if serverPort != "" {
		env = append(env, "WEBHOOK_BRIDGE_PORT="+serverPort)
	}
	if executorPort != "" {
		env = append(env, "WEBHOOK_BRIDGE_EXECUTOR_PORT="+executorPort)
	}
	cmd.Env = env

	if verbose {
		fmt.Println("🚀 Starting Go server...")
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	}

	// Start the process
	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("failed to start Go server: %w", err)
	}

	return cmd, nil
}

func copyFile(src, dst string) error {
	sourceFile, err := os.ReadFile(src)
	if err != nil {
		return err
	}

	return os.WriteFile(dst, sourceFile, 0600)
}

// parsePort parses a port string to integer
func parsePort(portStr string) (int, error) {
	if portStr == "" {
		return 0, fmt.Errorf("empty port string")
	}

	port := 0
	if _, err := fmt.Sscanf(portStr, "%d", &port); err != nil {
		return 0, fmt.Errorf("invalid port: %s", portStr)
	}

	if port < 1 || port > 65535 {
		return 0, fmt.Errorf("port out of range: %d", port)
	}

	return port, nil
}

// startPythonExecutorWithConfig starts Python executor using detected interpreter and config
func startPythonExecutorWithConfig(ctx context.Context, cfg *config.Config, pythonResult *python.DetectionResult, verbose bool) (*exec.Cmd, error) {
	if verbose {
		fmt.Println("🐍 Starting Python executor with detected interpreter...")
	}

	// Use detected Python interpreter
	pythonPath := pythonResult.Interpreter.Path

	// Prepare command arguments
	args := []string{"python_executor/main.py", "--plugin-dirs", "example_plugins"}
	args = append(args, "--port", fmt.Sprintf("%d", cfg.Executor.Port))

	// Create command
	cmd := exec.CommandContext(ctx, pythonPath, args...)
	cmd.Dir = "."

	// Set environment for virtual environment if needed
	if pythonResult.Interpreter.IsVirtual {
		env := os.Environ()
		env = append(env, fmt.Sprintf("VIRTUAL_ENV=%s", pythonResult.Interpreter.VenvPath))
		cmd.Env = env
	}

	if verbose {
		fmt.Printf("🐍 Python path: %s\n", pythonPath)
		fmt.Printf("🔧 Executor port: %d\n", cfg.Executor.Port)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	}

	// Start the process
	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("failed to start Python executor: %w", err)
	}

	return cmd, nil
}

// startGoServerWithConfig starts Go server using configuration
func startGoServerWithConfig(ctx context.Context, cfg *config.Config, verbose bool) (*exec.Cmd, error) {
	if verbose {
		fmt.Println("🚀 Starting Go server...")
	}

	// Find Go server executable
	var serverPath string
	if runtime.GOOS == "windows" {
		serverPath = filepath.Join("build", "webhook-bridge-server.exe")
	} else {
		serverPath = filepath.Join("build", "webhook-bridge-server")
	}

	// Check if server executable exists
	if _, err := os.Stat(serverPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("Go server executable not found. Run 'webhook-bridge build' first")
	}

	// Create command
	cmd := exec.CommandContext(ctx, serverPath)
	cmd.Dir = "."

	// Set environment variables from configuration
	env := os.Environ()
	env = append(env, fmt.Sprintf("WEBHOOK_BRIDGE_PORT=%d", cfg.Server.Port))
	env = append(env, fmt.Sprintf("WEBHOOK_BRIDGE_EXECUTOR_PORT=%d", cfg.Executor.Port))
	env = append(env, fmt.Sprintf("WEBHOOK_BRIDGE_MODE=%s", cfg.Server.Mode))
	cmd.Env = env

	if verbose {
		fmt.Printf("🌐 Server port: %d\n", cfg.Server.Port)
		fmt.Printf("🔧 Executor port: %d\n", cfg.Executor.Port)
		fmt.Printf("📊 Mode: %s\n", cfg.Server.Mode)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	}

	// Start the process
	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("failed to start Go server: %w", err)
	}

	return cmd, nil
}

// runAsDaemon runs the services in daemon mode
func runAsDaemon(cfg *config.Config, pythonResult *python.DetectionResult, verbose bool) error {
	if verbose {
		fmt.Println("🔄 Starting services in daemon mode...")
	}

	ctx := context.Background()

	// Start Python executor
	pythonCmd, err := startPythonExecutorWithConfig(ctx, cfg, pythonResult, verbose)
	if err != nil {
		return fmt.Errorf("failed to start Python executor: %w", err)
	}

	// Wait for Python executor to start
	time.Sleep(2 * time.Second)

	// Start Go server
	goCmd, err := startGoServerWithConfig(ctx, cfg, verbose)
	if err != nil {
		pythonCmd.Process.Kill()
		return fmt.Errorf("failed to start Go server: %w", err)
	}

	fmt.Println("✅ Services started in daemon mode")
	fmt.Printf("📊 Server PID: %d\n", goCmd.Process.Pid)
	fmt.Printf("🐍 Python Executor PID: %d\n", pythonCmd.Process.Pid)
	fmt.Printf("🌐 Server: http://localhost:%d\n", cfg.Server.Port)
	fmt.Printf("📊 Health: http://localhost:%d/health\n", cfg.Server.Port)

	return nil
}
