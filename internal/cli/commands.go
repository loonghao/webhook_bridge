package cli

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/spf13/cobra"
)

// NewStopCommand creates the stop command
func NewStopCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "stop",
		Short: "Stop the webhook bridge service",
		RunE:  runStop,
	}
}

func runStop(cmd *cobra.Command, args []string) error {
	verbose, _ := cmd.Flags().GetBool("verbose")

	if verbose {
		fmt.Println("🛑 Stopping webhook bridge services...")
	}

	stopped := 0

	// Stop Go server
	if err := stopProcessByName("webhook-bridge-server", verbose); err == nil {
		stopped++
	}

	// Stop Python executor
	if err := stopProcessByName("python", verbose); err == nil {
		stopped++
	}

	if stopped > 0 {
		fmt.Printf("✅ Stopped %d service(s)\n", stopped)
	} else {
		fmt.Println("ℹ️  No running services found")
	}

	return nil
}

// NewStatusCommand creates the status command
func NewStatusCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "status",
		Short: "Show service status",
		RunE:  runStatus,
	}
}

func runStatus(cmd *cobra.Command, args []string) error {
	verbose, _ := cmd.Flags().GetBool("verbose")

	fmt.Println("📊 Webhook Bridge Service Status")
	fmt.Println("================================")

	// Check Go server
	if pid := findProcessByName("webhook-bridge-server"); pid > 0 {
		fmt.Printf("🚀 Go Server: ✅ Running (PID: %d)\n", pid)
	} else {
		fmt.Println("🚀 Go Server: ❌ Not running")
	}

	// Check Python executor
	if pid := findProcessByName("python"); pid > 0 {
		fmt.Printf("🐍 Python Executor: ✅ Running (PID: %d)\n", pid)
	} else {
		fmt.Println("🐍 Python Executor: ❌ Not running")
	}

	// Check build status
	fmt.Println("\n🔨 Build Status:")
	checkBuildStatus(verbose)

	// Check configuration
	fmt.Println("\n📝 Configuration:")
	checkConfiguration(verbose)

	return nil
}

// NewCleanCommand creates the clean command
func NewCleanCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "clean",
		Short: "Clean build artifacts and temporary files",
		RunE:  runClean,
	}
}

func runClean(cmd *cobra.Command, args []string) error {
	verbose, _ := cmd.Flags().GetBool("verbose")

	if verbose {
		fmt.Println("🧹 Cleaning build artifacts...")
	}

	// Directories to clean
	dirs := []string{"build", "dist", "bin"}
	for _, dir := range dirs {
		if err := os.RemoveAll(dir); err != nil && verbose {
			fmt.Printf("  Warning: failed to remove %s: %v\n", dir, err)
		} else if verbose {
			fmt.Printf("  Removed %s/\n", dir)
		}
	}

	// Files to clean
	patterns := []string{"*.log", "*.pid", "coverage.out", "coverage.html"}
	for _, pattern := range patterns {
		matches, _ := filepath.Glob(pattern)
		for _, match := range matches {
			if err := os.Remove(match); err != nil && verbose {
				fmt.Printf("  Warning: failed to remove %s: %v\n", match, err)
			} else if verbose {
				fmt.Printf("  Removed %s\n", match)
			}
		}
	}

	// Clean Go cache
	if verbose {
		fmt.Println("  Cleaning Go cache...")
	}
	if err := exec.Command("go", "clean", "-cache").Run(); err != nil && verbose {
		fmt.Printf("  ⚠️  Warning: Failed to clean Go cache: %v\n", err)
	}

	fmt.Println("✅ Clean completed")
	return nil
}

// NewTestCommand creates the test command
func NewTestCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "test",
		Short: "Run tests",
		RunE:  runTest,
	}

	cmd.Flags().Bool("go", true, "Run Go tests")
	cmd.Flags().Bool("python", true, "Run Python tests")
	cmd.Flags().Bool("integration", false, "Run integration tests")
	cmd.Flags().Bool("coverage", false, "Generate coverage report")

	return cmd
}

func runTest(cmd *cobra.Command, args []string) error {
	verbose, _ := cmd.Flags().GetBool("verbose")
	runGo, _ := cmd.Flags().GetBool("go")
	runPython, _ := cmd.Flags().GetBool("python")
	runIntegration, _ := cmd.Flags().GetBool("integration")
	coverage, _ := cmd.Flags().GetBool("coverage")

	if verbose {
		fmt.Println("🧪 Running tests...")
	}

	passed := 0
	total := 0

	// Run Go tests
	if runGo {
		total++
		if verbose {
			fmt.Println("  Running Go tests...")
		}

		var goCmd *exec.Cmd
		if coverage {
			goCmd = exec.Command("go", "test", "-coverprofile=coverage.out", "./...")
		} else {
			goCmd = exec.Command("go", "test", "./...")
		}

		if verbose {
			goCmd.Stdout = os.Stdout
			goCmd.Stderr = os.Stderr
		}

		if err := goCmd.Run(); err == nil {
			passed++
			if verbose {
				fmt.Println("  ✅ Go tests passed")
			}
		} else if verbose {
			fmt.Println("  ❌ Go tests failed")
		}

		// Generate coverage report
		if coverage {
			if err := exec.Command("go", "tool", "cover", "-html=coverage.out", "-o", "coverage.html").Run(); err != nil {
				if verbose {
					fmt.Printf("  ⚠️  Warning: Failed to generate coverage report: %v\n", err)
				}
			} else if verbose {
				fmt.Println("  📊 Coverage report generated: coverage.html")
			}
		}
	}

	// Run Python tests
	if runPython {
		total++
		if verbose {
			fmt.Println("  Running Python tests...")
		}

		var pythonPath string
		if runtime.GOOS == "windows" {
			pythonPath = filepath.Join(".venv", "Scripts", "python.exe")
		} else {
			pythonPath = filepath.Join(".venv", "bin", "python")
		}

		pythonCmd := exec.Command(pythonPath, "-m", "pytest", "tests/", "-v")
		if verbose {
			pythonCmd.Stdout = os.Stdout
			pythonCmd.Stderr = os.Stderr
		}

		if err := pythonCmd.Run(); err == nil {
			passed++
			if verbose {
				fmt.Println("  ✅ Python tests passed")
			}
		} else if verbose {
			fmt.Println("  ❌ Python tests failed")
		}
	}

	// Run integration tests
	if runIntegration {
		total++
		if verbose {
			fmt.Println("  Running integration tests...")
		}

		var pythonPath string
		if runtime.GOOS == "windows" {
			pythonPath = filepath.Join(".venv", "Scripts", "python.exe")
		} else {
			pythonPath = filepath.Join(".venv", "bin", "python")
		}

		integrationCmd := exec.Command(pythonPath, "test_go_python_integration.py")
		if verbose {
			integrationCmd.Stdout = os.Stdout
			integrationCmd.Stderr = os.Stderr
		}

		if err := integrationCmd.Run(); err == nil {
			passed++
			if verbose {
				fmt.Println("  ✅ Integration tests passed")
			}
		} else if verbose {
			fmt.Println("  ❌ Integration tests failed")
		}
	}

	fmt.Printf("📊 Test Results: %d/%d passed\n", passed, total)
	if passed == total {
		fmt.Println("🎉 All tests passed!")
	}

	return nil
}

// NewDashboardCommand creates the dashboard command
func NewDashboardCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "dashboard",
		Short: "Open web dashboard",
		Long:  "Start the service and open the web dashboard for monitoring and management",
		RunE:  runDashboard,
	}

	cmd.Flags().StringP("env", "e", "dev", "Environment (dev, prod)")
	cmd.Flags().String("port", "", "Override server port")
	cmd.Flags().Bool("no-browser", false, "Don't open browser automatically")

	return cmd
}

// NewConfigCommand creates the config command
func NewConfigCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config",
		Short: "Manage configuration",
		RunE:  runConfig,
	}

	cmd.Flags().String("env", "", "Set environment (dev, prod)")
	cmd.Flags().Bool("show", false, "Show current configuration")

	return cmd
}

func runDashboard(cmd *cobra.Command, args []string) error {
	verbose, _ := cmd.Flags().GetBool("verbose")
	env, _ := cmd.Flags().GetString("env")
	port, _ := cmd.Flags().GetString("port")
	noBrowser, _ := cmd.Flags().GetBool("no-browser")

	if verbose {
		fmt.Printf("🌐 Starting webhook bridge with web dashboard...\n")
	}

	// Build first
	if verbose {
		fmt.Println("🔨 Building...")
	}
	buildCmd := NewBuildCommand()
	if err := buildCmd.RunE(buildCmd, []string{}); err != nil {
		return fmt.Errorf("build failed: %w", err)
	}

	// Setup configuration
	if err := setupConfiguration(env, verbose); err != nil {
		return fmt.Errorf("failed to setup configuration: %w", err)
	}

	// Start services in background
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start Python executor
	pythonCmd, err := startPythonExecutor(ctx, "", verbose)
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
	goCmd, err := startGoServer(ctx, port, "", verbose)
	if err != nil {
		return fmt.Errorf("failed to start Go server: %w", err)
	}
	defer func() {
		if goCmd.Process != nil {
			goCmd.Process.Kill()
		}
	}()

	// Wait for server to start
	time.Sleep(3 * time.Second)

	// Determine dashboard URL
	dashboardPort := "8000" // default
	if port != "" {
		dashboardPort = port
	}
	dashboardURL := fmt.Sprintf("http://localhost:%s/dashboard", dashboardPort)

	fmt.Printf("🌐 Web Dashboard available at: %s\n", dashboardURL)
	fmt.Printf("📊 API endpoints available at: http://localhost:%s/api/v1\n", dashboardPort)

	// Open browser if requested
	if !noBrowser {
		if verbose {
			fmt.Printf("🌐 Opening dashboard in browser...\n")
		}
		openBrowser(dashboardURL)
	}

	fmt.Println("📊 Dashboard is running. Press Ctrl+C to stop...")

	// Wait for interrupt signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	fmt.Println("\n🛑 Stopping dashboard...")
	return nil
}

func runConfig(cmd *cobra.Command, args []string) error {
	env, _ := cmd.Flags().GetString("env")
	show, _ := cmd.Flags().GetBool("show")

	if env != "" {
		return setupConfiguration(env, true)
	}

	if show {
		if content, err := os.ReadFile("config.yaml"); err == nil {
			fmt.Println("📝 Current Configuration:")
			fmt.Println("========================")
			fmt.Print(string(content))
		} else {
			fmt.Println("❌ No configuration file found")
		}
	}

	return nil
}

// NewDeployCommand creates the deploy command
func NewDeployCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "deploy",
		Short: "Deploy the application",
		RunE:  runDeploy,
	}

	cmd.Flags().StringP("env", "e", "dev", "Environment (dev, prod)")
	cmd.Flags().Bool("skip-tests", false, "Skip running tests")
	cmd.Flags().Bool("docker", false, "Deploy using Docker")

	return cmd
}

func runDeploy(cmd *cobra.Command, args []string) error {
	verbose, _ := cmd.Flags().GetBool("verbose")
	env, _ := cmd.Flags().GetString("env")
	skipTests, _ := cmd.Flags().GetBool("skip-tests")
	docker, _ := cmd.Flags().GetBool("docker")

	if verbose {
		fmt.Printf("🚀 Deploying for %s environment...\n", env)
	}

	if docker {
		return deployWithDocker(env, verbose)
	}

	// Standard deployment
	steps := []struct {
		name string
		fn   func() error
	}{
		{"Clean", func() error { return runClean(cmd, []string{}) }},
		{"Build", func() error { return runBuild(cmd, []string{}) }},
	}

	if !skipTests {
		steps = append(steps, struct {
			name string
			fn   func() error
		}{"Test", func() error { return runTest(cmd, []string{}) }})
	}

	for _, step := range steps {
		if verbose {
			fmt.Printf("  %s...\n", step.name)
		}
		if err := step.fn(); err != nil {
			return fmt.Errorf("%s failed: %w", step.name, err)
		}
	}

	fmt.Println("✅ Deployment completed!")
	return nil
}

// Helper functions

func stopProcessByName(name string, verbose bool) error {
	if runtime.GOOS == "windows" {
		cmd := exec.Command("taskkill", "/F", "/IM", name+".exe")
		return cmd.Run()
	} else {
		cmd := exec.Command("pkill", "-f", name)
		return cmd.Run()
	}
}

func findProcessByName(name string) int {
	if runtime.GOOS == "windows" {
		cmd := exec.Command("tasklist", "/FI", "IMAGENAME eq "+name+".exe", "/FO", "CSV", "/NH")
		output, err := cmd.Output()
		if err != nil {
			return 0
		}
		lines := strings.Split(string(output), "\n")
		if len(lines) > 0 && strings.Contains(lines[0], name) {
			parts := strings.Split(lines[0], ",")
			if len(parts) > 1 {
				pidStr := strings.Trim(parts[1], "\"")
				if pid, err := strconv.Atoi(pidStr); err == nil {
					return pid
				}
			}
		}
	} else {
		cmd := exec.Command("pgrep", "-f", name)
		output, err := cmd.Output()
		if err != nil {
			return 0
		}
		pidStr := strings.TrimSpace(string(output))
		if pid, err := strconv.Atoi(pidStr); err == nil {
			return pid
		}
	}
	return 0
}

func checkBuildStatus(verbose bool) {
	var serverPath, managerPath string
	if runtime.GOOS == "windows" {
		serverPath = filepath.Join("build", "webhook-bridge-server.exe")
		managerPath = filepath.Join("build", "python-manager.exe")
	} else {
		serverPath = filepath.Join("build", "webhook-bridge-server")
		managerPath = filepath.Join("build", "python-manager")
	}

	if _, err := os.Stat(serverPath); err == nil {
		fmt.Println("  🚀 Go Server: ✅ Built")
	} else {
		fmt.Println("  🚀 Go Server: ❌ Not built")
	}

	if _, err := os.Stat(managerPath); err == nil {
		fmt.Println("  🔧 Python Manager: ✅ Built")
	} else {
		fmt.Println("  🔧 Python Manager: ❌ Not built")
	}

	if _, err := os.Stat(".venv"); err == nil {
		fmt.Println("  🐍 Python Environment: ✅ Ready")
	} else {
		fmt.Println("  🐍 Python Environment: ❌ Not setup")
	}
}

func checkConfiguration(verbose bool) {
	if _, err := os.Stat("config.yaml"); err == nil {
		fmt.Println("  📝 config.yaml: ✅ Present")
	} else {
		fmt.Println("  📝 config.yaml: ❌ Missing")
	}

	configs := []string{"config.dev.yaml", "config.prod.yaml", "config.example.yaml"}
	for _, config := range configs {
		if _, err := os.Stat(config); err == nil {
			fmt.Printf("  📝 %s: ✅ Present\n", config)
		} else {
			fmt.Printf("  📝 %s: ❌ Missing\n", config)
		}
	}
}

func deployWithDocker(env string, verbose bool) error {
	if verbose {
		fmt.Println("🐳 Deploying with Docker...")
	}

	// Build Docker image
	cmd := exec.Command("docker", "build", "-t", "webhook-bridge:latest", ".")
	if verbose {
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	}

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("Docker build failed: %w", err)
	}

	if verbose {
		fmt.Println("✅ Docker image built successfully")
	}

	return nil
}

func openBrowser(url string) error {
	var cmd *exec.Cmd

	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", url)
	case "darwin":
		cmd = exec.Command("open", url)
	case "linux":
		cmd = exec.Command("xdg-open", url)
	default:
		return fmt.Errorf("unsupported platform")
	}

	return cmd.Start()
}
