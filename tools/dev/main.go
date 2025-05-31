// Package main provides development tools for webhook-bridge
// Similar to cargo commands for Rust
package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

func main() {
	if len(os.Args) < 2 {
		showHelp()
		return
	}

	command := os.Args[1]
	args := os.Args[2:]

	// Handle dashboard subcommands
	if command == "dashboard" {
		handleDashboardCommand(args)
		return
	}

	switch command {
	case "help", "-h", "--help":
		showHelp()
	case "dev-setup":
		devSetup()
	case "deps":
		installDeps()
	case "proto":
		generateProto()
	case "build":
		buildProject(args)
	case "start":
		runDev()
	case "start-dev":
		startDevEnvironment()
	case "clean":
		cleanProject()
	case "test":
		runTests(args)
	case "test-go":
		runGoTests()
	case "test-python":
		runPythonTests()
	case "test-race":
		runRaceTests()
	case "test-coverage":
		runCoverageTests()
	case "lint":
		runLint()
	case "format":
		formatCode()
	case "verify":
		verifyProject()
	case "version":
		showVersion()
	case "doctor":
		checkProject()
	case "install":
		installDeps()
	case "release":
		createRelease()
	case "release-snapshot":
		createSnapshotRelease()
	case "release-dry-run":
		dryRunRelease()
	default:
		fmt.Printf("❌ Unknown command: %s\n", command)
		showHelp()
		os.Exit(1)
	}
}

func handleDashboardCommand(args []string) {
	if len(args) == 0 {
		showDashboardHelp()
		return
	}

	dashboard := &DashboardCommands{}
	subcommand := args[0]
	subargs := args[1:]

	switch subcommand {
	case "build":
		if err := dashboard.Build(subargs); err != nil {
			fmt.Printf("❌ Dashboard build failed: %v\n", err)
			os.Exit(1)
		}
	case "dev":
		if err := dashboard.Dev(subargs); err != nil {
			fmt.Printf("❌ Dashboard dev failed: %v\n", err)
			os.Exit(1)
		}
	case "install":
		if err := dashboard.Install(subargs); err != nil {
			fmt.Printf("❌ Dashboard install failed: %v\n", err)
			os.Exit(1)
		}
	case "lint":
		if err := dashboard.Lint(subargs); err != nil {
			fmt.Printf("❌ Dashboard lint failed: %v\n", err)
			os.Exit(1)
		}
	case "type-check":
		if err := dashboard.TypeCheck(subargs); err != nil {
			fmt.Printf("❌ Dashboard type check failed: %v\n", err)
			os.Exit(1)
		}
	case "clean":
		if err := dashboard.Clean(subargs); err != nil {
			fmt.Printf("❌ Dashboard clean failed: %v\n", err)
			os.Exit(1)
		}
	case "serve":
		if err := dashboard.Serve(subargs); err != nil {
			fmt.Printf("❌ Dashboard serve failed: %v\n", err)
			os.Exit(1)
		}
	case "help", "-h", "--help":
		showDashboardHelp()
	default:
		fmt.Printf("❌ Unknown dashboard command: %s\n", subcommand)
		showDashboardHelp()
		os.Exit(1)
	}
}

func showHelp() {
	fmt.Println(`🚀 webhook-bridge development tool

USAGE:
    go run dev.go <COMMAND> [args...]

DEVELOPMENT:
    dev-setup    Setup development environment
    deps         Install all dependencies
    proto        Generate protobuf files
    build        Build the project binaries
    start        Start development environment (manual)
    start-dev    Start integrated dev environment (auto)
    clean        Clean build artifacts

DASHBOARD:
    dashboard    Dashboard development commands
                 Use 'go run dev.go dashboard help' for more info

TESTING:
    test         Run all tests
    test-go      Run Go tests only
    test-python  Run Python tests only
    test-race    Run tests with race detection
    test-coverage Run tests with coverage

CODE QUALITY:
    lint         Run linters
    format       Format code
    verify       Run all verification checks

RELEASE:
    release      Create a release with GoReleaser
    release-snapshot Create a snapshot release
    release-dry-run  Dry run release process

UTILITIES:
    version      Show version information
    doctor       Check development environment health
    install      Install development dependencies

EXAMPLES:
    go run dev.go dev-setup
    go run dev.go build
    go run dev.go test
    go run dev.go dashboard build
    go run dev.go dashboard dev
    go run dev.go release-snapshot`)
}

func showDashboardHelp() {
	fmt.Println(`🎛️ Dashboard Development Commands

USAGE:
    go run dev.go dashboard <COMMAND> [args...]

COMMANDS:
    build        Build TypeScript dashboard
                 --watch, -w     Watch mode for development
                 --production    Production build with minification
                 --clean, -c     Clean before building

    dev          Start dashboard development mode
                 (TypeScript watch + Go server)

    install      Install dashboard dependencies (npm install)

    lint         Run TypeScript linting
                 --fix, -f       Auto-fix linting issues

    type-check   Run TypeScript type checking

    clean        Clean dashboard build artifacts and node_modules

    serve        Serve dashboard for development
                 --port, -p      Specify port (default: 8080)

    help         Show this help message

EXAMPLES:
    go run dev.go dashboard build
    go run dev.go dashboard build --watch
    go run dev.go dashboard build --production
    go run dev.go dashboard dev
    go run dev.go dashboard lint --fix
    go run dev.go dashboard serve --port 3000

DEVELOPMENT WORKFLOW:
    1. go run dev.go dashboard install    # Install dependencies
    2. go run dev.go dashboard dev        # Start development mode
    3. Open http://localhost:8000         # View dashboard

PRODUCTION BUILD:
    go run dev.go dashboard build --production`)
}

func generateProto() {
	fmt.Println("🔧 Generating protobuf files...")

	// Ensure api/proto directory exists with secure permissions
	if err := os.MkdirAll("api/proto", 0750); err != nil {
		fmt.Printf("Error creating directory: %v\n", err)
		os.Exit(1)
	}

	// Generate Go protobuf files
	cmd := exec.Command("protoc",
		"--go_out=.", "--go_opt=paths=source_relative",
		"--go-grpc_out=.", "--go-grpc_opt=paths=source_relative",
		"api/proto/webhook.proto")

	if err := cmd.Run(); err != nil {
		fmt.Printf("Error generating Go protobuf: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("✅ Protobuf files generated successfully")
}

func buildProject(args []string) {
	fmt.Println("🔨 Building project...")

	// Ensure protobuf files exist
	if !fileExists("api/proto/webhook.pb.go") {
		fmt.Println("Protobuf files not found, generating...")
		generateProto()
	}

	// Build main CLI
	buildBinary("./cmd/webhook-bridge", "webhook-bridge")

	// Build server
	buildBinary("./cmd/server", "webhook-bridge-server")

	// Build python manager
	buildBinary("./cmd/python-manager", "python-manager")

	fmt.Println("✅ Build completed successfully")
}

func buildBinary(source, name string) {
	var output string
	if runtime.GOOS == "windows" {
		output = name + ".exe"
	} else {
		output = name
	}

	cmd := exec.Command("go", "build", "-o", output, source)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		fmt.Printf("Error building %s: %v\n", name, err)
		os.Exit(1)
	}

	fmt.Printf("✅ Built %s\n", output)
}

func runTests(args []string) {
	fmt.Println("🧪 Running tests...")

	testArgs := []string{"test", "-v", "./..."}
	if len(args) > 0 {
		testArgs = append(testArgs, args...)
	}

	cmd := exec.Command("go", testArgs...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		fmt.Printf("Tests failed: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("✅ All tests passed")
}

func runLint() {
	fmt.Println("🔍 Running linters...")

	// Run golangci-lint
	cmd := exec.Command("golangci-lint", "run")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		fmt.Printf("Linting failed: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("✅ Linting passed")
}

func cleanProject() {
	fmt.Println("🧹 Cleaning project...")

	// Remove build artifacts
	patterns := []string{
		"*.exe", "webhook-bridge-server", "python-manager",
		"coverage.out", "coverage.html", "*.log", "*.pid",
	}

	for _, pattern := range patterns {
		matches, _ := filepath.Glob(pattern)
		for _, match := range matches {
			os.Remove(match)
			fmt.Printf("Removed %s\n", match)
		}
	}

	// Remove build directories
	dirs := []string{"build", "dist"}
	for _, dir := range dirs {
		if dirExists(dir) {
			os.RemoveAll(dir)
			fmt.Printf("Removed %s/\n", dir)
		}
	}

	fmt.Println("✅ Clean completed")
}

func runDev() {
	fmt.Println("🚀 Starting development environment...")

	// Build first
	buildProject(nil)

	fmt.Println("Development environment ready!")
	fmt.Println("Run the following commands in separate terminals:")
	fmt.Println("  ./webhook-bridge-server")
	fmt.Println("  python python_executor/main.py")
}

func startDevEnvironment() {
	fmt.Println("🚀 Starting integrated development environment...")

	// Import and call the start dev function
	// Note: This would require importing the startdev package
	// For now, we'll use a simpler approach
	fmt.Println("Starting development servers...")

	// Build first
	buildProject(nil)

	fmt.Println("✅ Development environment ready!")
	fmt.Println("Run the following commands in separate terminals:")
	fmt.Println("  ./webhook-bridge-server")
	fmt.Println("  python python_executor/main.py")
}

func installDeps() {
	fmt.Println("📦 Installing dependencies...")

	// Go dependencies
	cmd := exec.Command("go", "mod", "download")
	if err := cmd.Run(); err != nil {
		fmt.Printf("Error downloading Go dependencies: %v\n", err)
		os.Exit(1)
	}

	// Install protobuf tools
	tools := []string{
		"google.golang.org/protobuf/cmd/protoc-gen-go@latest",
		"google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest",
	}

	for _, tool := range tools {
		cmd := exec.Command("go", "install", tool)
		if err := cmd.Run(); err != nil {
			fmt.Printf("Error installing %s: %v\n", tool, err)
		}
	}

	fmt.Println("✅ Dependencies installed")
}

func checkProject() {
	fmt.Println("🔍 Checking project health...")

	checks := []struct {
		name string
		fn   func() bool
	}{
		{"Go modules", checkGoMod},
		{"Protobuf files", checkProtobuf},
		{"Required tools", checkTools},
	}

	allPassed := true
	for _, check := range checks {
		if check.fn() {
			fmt.Printf("✅ %s\n", check.name)
		} else {
			fmt.Printf("❌ %s\n", check.name)
			allPassed = false
		}
	}

	if allPassed {
		fmt.Println("✅ Project health check passed")
	} else {
		fmt.Println("❌ Project health check failed")
		os.Exit(1)
	}
}

func checkGoMod() bool {
	return fileExists("go.mod")
}

func checkProtobuf() bool {
	return fileExists("api/proto/webhook.proto")
}

func checkTools() bool {
	tools := []string{"protoc", "go"}
	for _, tool := range tools {
		if _, err := exec.LookPath(tool); err != nil {
			return false
		}
	}
	return true
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func dirExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.IsDir()
}

func devSetup() {
	fmt.Println("🚀 Setting up development environment...")
	installDeps()
	generateProto()
	fmt.Println("✅ Development environment setup complete!")
}

func runGoTests() {
	fmt.Println("🧪 Running Go tests...")
	cmd := exec.Command("go", "test", "-v", "./...")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		fmt.Printf("Go tests failed: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("✅ Go tests passed")
}

func runPythonTests() {
	fmt.Println("🐍 Running Python tests...")

	// Try different Python commands
	pythonCmds := []string{"python", "python3", "py"}
	var pythonCmd string

	for _, cmd := range pythonCmds {
		if _, err := exec.LookPath(cmd); err == nil {
			pythonCmd = cmd
			break
		}
	}

	if pythonCmd == "" {
		fmt.Println("⚠️ Python not found, skipping Python tests")
		return
	}

	cmd := exec.Command(pythonCmd, "-m", "pytest", "tests/")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		fmt.Printf("Python tests failed: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("✅ Python tests passed")
}

func runRaceTests() {
	fmt.Println("🏃 Running Go tests with race detection...")
	cmd := exec.Command("go", "test", "-race", "-v", "./...")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		fmt.Printf("Race tests failed: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("✅ Race tests passed")
}

func runCoverageTests() {
	fmt.Println("📊 Running tests with coverage...")

	// Run tests with coverage
	cmd := exec.Command("go", "test", "-coverprofile=coverage.out", "./...")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		fmt.Printf("Coverage tests failed: %v\n", err)
		os.Exit(1)
	}

	// Generate HTML coverage report
	cmd = exec.Command("go", "tool", "cover", "-html=coverage.out", "-o", "coverage.html")
	if err := cmd.Run(); err != nil {
		fmt.Printf("Failed to generate coverage report: %v\n", err)
	} else {
		fmt.Println("📈 Coverage report generated: coverage.html")
	}

	fmt.Println("✅ Coverage tests completed")
}

func formatCode() {
	fmt.Println("🎨 Formatting code...")

	// Format Go code
	cmd := exec.Command("go", "fmt", "./...")
	if err := cmd.Run(); err != nil {
		fmt.Printf("Go formatting failed: %v\n", err)
		os.Exit(1)
	}

	// Format Python code if available
	if _, err := exec.LookPath("ruff"); err == nil {
		cmd = exec.Command("ruff", "format", ".")
		if err := cmd.Run(); err != nil {
			fmt.Printf("Python formatting failed: %v\n", err)
		}
	}

	fmt.Println("✅ Code formatting completed")
}

func verifyProject() {
	fmt.Println("🔍 Running verification checks...")

	checks := []struct {
		name string
		fn   func() bool
	}{
		{"Go formatting", verifyGoFmt},
		{"Go vet", verifyGoVet},
		{"Go modules", verifyGoMod},
		{"Protobuf files", checkProtobuf},
	}

	allPassed := true
	for _, check := range checks {
		fmt.Printf("Checking %s... ", check.name)
		if check.fn() {
			fmt.Println("✅")
		} else {
			fmt.Println("❌")
			allPassed = false
		}
	}

	if allPassed {
		fmt.Println("✅ All verification checks passed")
	} else {
		fmt.Println("❌ Some verification checks failed")
		os.Exit(1)
	}
}

func verifyGoFmt() bool {
	cmd := exec.Command("gofmt", "-l", ".")
	output, err := cmd.Output()
	return err == nil && len(output) == 0
}

func verifyGoVet() bool {
	cmd := exec.Command("go", "vet", "./...")
	return cmd.Run() == nil
}

func verifyGoMod() bool {
	cmd := exec.Command("go", "mod", "verify")
	return cmd.Run() == nil
}

func showVersion() {
	fmt.Println("📋 Version Information")
	fmt.Println("Project: webhook-bridge")

	// Get version from git
	if version := getGitVersion(); version != "" {
		fmt.Printf("Version: %s\n", version)
	}

	// Get Go version
	if output, err := exec.Command("go", "version").Output(); err == nil {
		fmt.Printf("Go: %s", string(output))
	}

	// Get build info
	fmt.Printf("OS: %s\n", runtime.GOOS)
	fmt.Printf("Arch: %s\n", runtime.GOARCH)
}

func getGitVersion() string {
	cmd := exec.Command("git", "describe", "--tags", "--always", "--dirty")
	output, err := cmd.Output()
	if err != nil {
		return "dev"
	}
	return strings.TrimSpace(string(output))
}

func createRelease() {
	fmt.Println("🚀 Creating release with GoReleaser...")

	// Check if goreleaser is available
	if _, err := exec.LookPath("goreleaser"); err != nil {
		fmt.Println("❌ GoReleaser not found. Installing...")
		installDeps()
	}

	// Check if we're on a tag
	cmd := exec.Command("git", "tag", "--points-at", "HEAD")
	output, err := cmd.Output()
	if err != nil || len(strings.TrimSpace(string(output))) == 0 {
		fmt.Println("❌ No tag found at HEAD. Please create a tag first:")
		fmt.Println("  git tag v1.0.0")
		fmt.Println("  git push origin v1.0.0")
		os.Exit(1)
	}

	// Run goreleaser
	cmd = exec.Command("goreleaser", "release", "--clean")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		fmt.Printf("Release failed: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("✅ Release completed successfully!")
}

func createSnapshotRelease() {
	fmt.Println("📸 Creating snapshot release...")

	// Check if goreleaser is available
	if _, err := exec.LookPath("goreleaser"); err != nil {
		fmt.Println("❌ GoReleaser not found. Installing...")
		installDeps()
	}

	// Run goreleaser snapshot
	cmd := exec.Command("goreleaser", "release", "--snapshot", "--clean")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		fmt.Printf("Snapshot release failed: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("✅ Snapshot release completed!")
	fmt.Println("📁 Check the dist/ directory for artifacts")
}

func dryRunRelease() {
	fmt.Println("🧪 Running dry-run release...")

	// Check if goreleaser is available
	if _, err := exec.LookPath("goreleaser"); err != nil {
		fmt.Println("❌ GoReleaser not found. Installing...")
		installDeps()
	}

	// Run goreleaser dry-run
	cmd := exec.Command("goreleaser", "release", "--skip=publish", "--clean")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		fmt.Printf("Dry-run release failed: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("✅ Dry-run release completed!")
	fmt.Println("📁 Check the dist/ directory for artifacts")
}

// DashboardCommands handles dashboard-related development tasks
type DashboardCommands struct{}

// Build builds the TypeScript dashboard
func (d *DashboardCommands) Build(args []string) error {
	var watch, production, clean bool

	for _, arg := range args {
		switch arg {
		case "--watch", "-w":
			watch = true
		case "--production", "--prod", "-p":
			production = true
		case "--clean", "-c":
			clean = true
		}
	}

	webDir := filepath.Join("web")
	if _, err := os.Stat(filepath.Join(webDir, "tsconfig.json")); os.IsNotExist(err) {
		return fmt.Errorf("❌ Error: Must run from project root directory (tsconfig.json not found)")
	}

	// Check for Node.js
	if !commandExists("node") {
		return fmt.Errorf("❌ Error: Node.js is not installed or not in PATH\nPlease install Node.js from https://nodejs.org/")
	}

	// Check for npm
	if !commandExists("npm") {
		return fmt.Errorf("❌ Error: npm is not installed or not in PATH")
	}

	// Change to web directory
	originalDir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current directory: %w", err)
	}
	defer os.Chdir(originalDir)

	if err := os.Chdir(webDir); err != nil {
		return fmt.Errorf("failed to change to web directory: %w", err)
	}

	// Clean if requested
	if clean {
		printColored("🧹 Cleaning build directory...", ColorYellow)
		distDir := filepath.Join("static", "js", "dist")
		if _, err := os.Stat(distDir); err == nil {
			if err := os.RemoveAll(distDir); err != nil {
				return fmt.Errorf("failed to clean dist directory: %w", err)
			}
		}
		printColored("✅ Clean completed", ColorGreen)
	}

	// Check if node_modules exists
	if _, err := os.Stat("node_modules"); os.IsNotExist(err) {
		printColored("📦 Installing dependencies...", ColorYellow)
		if err := runCommand("npm", "install"); err != nil {
			return fmt.Errorf("❌ Failed to install dependencies: %w", err)
		}
		printColored("✅ Dependencies installed", ColorGreen)
	}

	// Create dist directory if it doesn't exist
	distDir := filepath.Join("static", "js", "dist")
	if err := os.MkdirAll(distDir, 0755); err != nil {
		return fmt.Errorf("failed to create dist directory: %w", err)
	}

	// Build based on mode
	if watch {
		printColored("👀 Starting TypeScript watch mode...", ColorYellow)
		printColored("Press Ctrl+C to stop", ColorYellow)
		return runCommand("npm", "run", "build:watch")
	} else if production {
		printColored("🏗️ Building for production...", ColorYellow)
		if err := runCommand("npm", "run", "build:prod"); err != nil {
			return fmt.Errorf("❌ Production build failed: %w", err)
		}
		printColored("✅ Production build completed", ColorGreen)
		printColored("📁 Output: web/static/js/dist/", ColorGreen)
	} else {
		printColored("🏗️ Building TypeScript dashboard...", ColorYellow)
		if err := runCommand("npm", "run", "build"); err != nil {
			return fmt.Errorf("❌ Build failed: %w", err)
		}
		printColored("✅ Build completed successfully", ColorGreen)
		printColored("📁 Output: web/static/js/dist/", ColorGreen)
	}

	printColored("🎉 Dashboard build process completed!", ColorGreen)
	return nil
}

// Dev starts the dashboard in development mode
func (d *DashboardCommands) Dev(args []string) error {
	printColored("🚀 Starting dashboard development mode...", ColorYellow)

	// Start TypeScript watch mode in background
	go func() {
		if err := d.Build([]string{"--watch"}); err != nil {
			printColored(fmt.Sprintf("TypeScript watch failed: %v", err), ColorRed)
		}
	}()

	// Start the Go server
	printColored("🌐 Starting Go server...", ColorYellow)
	return runCommand("go", "run", "cmd/server/main.go")
}

// Install installs dashboard dependencies
func (d *DashboardCommands) Install(args []string) error {
	webDir := filepath.Join("web")
	originalDir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current directory: %w", err)
	}
	defer os.Chdir(originalDir)

	if err := os.Chdir(webDir); err != nil {
		return fmt.Errorf("failed to change to web directory: %w", err)
	}

	printColored("📦 Installing dashboard dependencies...", ColorYellow)
	if err := runCommand("npm", "install"); err != nil {
		return fmt.Errorf("failed to install dependencies: %w", err)
	}

	printColored("✅ Dashboard dependencies installed", ColorGreen)
	return nil
}

// Lint runs TypeScript linting
func (d *DashboardCommands) Lint(args []string) error {
	var fix bool
	for _, arg := range args {
		if arg == "--fix" || arg == "-f" {
			fix = true
		}
	}

	webDir := filepath.Join("web")
	originalDir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current directory: %w", err)
	}
	defer os.Chdir(originalDir)

	if err := os.Chdir(webDir); err != nil {
		return fmt.Errorf("failed to change to web directory: %w", err)
	}

	printColored("🔍 Running TypeScript linting...", ColorYellow)

	var cmd string
	if fix {
		cmd = "lint:fix"
	} else {
		cmd = "lint"
	}

	if err := runCommand("npm", "run", cmd); err != nil {
		return fmt.Errorf("linting failed: %w", err)
	}

	printColored("✅ Linting completed", ColorGreen)
	return nil
}

// TypeCheck runs TypeScript type checking
func (d *DashboardCommands) TypeCheck(args []string) error {
	webDir := filepath.Join("web")
	originalDir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current directory: %w", err)
	}
	defer os.Chdir(originalDir)

	if err := os.Chdir(webDir); err != nil {
		return fmt.Errorf("failed to change to web directory: %w", err)
	}

	printColored("🔍 Running TypeScript type checking...", ColorYellow)
	if err := runCommand("npm", "run", "type-check"); err != nil {
		return fmt.Errorf("type checking failed: %w", err)
	}

	printColored("✅ Type checking passed", ColorGreen)
	return nil
}

// Clean cleans dashboard build artifacts
func (d *DashboardCommands) Clean(args []string) error {
	printColored("🧹 Cleaning dashboard build artifacts...", ColorYellow)

	distDir := filepath.Join("web", "static", "js", "dist")
	if _, err := os.Stat(distDir); err == nil {
		if err := os.RemoveAll(distDir); err != nil {
			return fmt.Errorf("failed to clean dist directory: %w", err)
		}
	}

	nodeModulesDir := filepath.Join("web", "node_modules")
	if _, err := os.Stat(nodeModulesDir); err == nil {
		printColored("🗑️ Removing node_modules...", ColorYellow)
		if err := os.RemoveAll(nodeModulesDir); err != nil {
			return fmt.Errorf("failed to clean node_modules: %w", err)
		}
	}

	printColored("✅ Dashboard cleaned", ColorGreen)
	return nil
}

// Serve serves the dashboard for development
func (d *DashboardCommands) Serve(args []string) error {
	port := "8080"
	for i, arg := range args {
		if arg == "--port" || arg == "-p" {
			if i+1 < len(args) {
				port = args[i+1]
			}
		}
	}

	webDir := filepath.Join("web")
	originalDir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current directory: %w", err)
	}
	defer os.Chdir(originalDir)

	if err := os.Chdir(webDir); err != nil {
		return fmt.Errorf("failed to change to web directory: %w", err)
	}

	printColored(fmt.Sprintf("🌐 Serving dashboard on http://localhost:%s", port), ColorGreen)

	// Try different static servers
	if commandExists("python3") {
		return runCommand("python3", "-m", "http.server", port)
	} else if commandExists("python") {
		return runCommand("python", "-m", "http.server", port)
	} else if commandExists("npx") {
		return runCommand("npx", "serve", "-p", port, ".")
	} else {
		return fmt.Errorf("no suitable static server found (python3, python, or npx required)")
	}
}

// commandExists checks if a command exists in PATH
func commandExists(cmd string) bool {
	_, err := exec.LookPath(cmd)
	return err == nil
}

// runCommand runs a command and prints output
func runCommand(name string, args ...string) error {
	cmd := exec.Command(name, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	// Set environment for cross-platform compatibility
	cmd.Env = os.Environ()
	if runtime.GOOS == "windows" {
		cmd.Env = append(cmd.Env, "FORCE_COLOR=1")
	}

	return cmd.Run()
}

// Color constants
const (
	ColorReset  = "\033[0m"
	ColorRed    = "\033[31m"
	ColorGreen  = "\033[32m"
	ColorYellow = "\033[33m"
	ColorBlue   = "\033[34m"
	ColorPurple = "\033[35m"
	ColorCyan   = "\033[36m"
	ColorWhite  = "\033[37m"
)

// printColored prints colored text
func printColored(text, color string) {
	if runtime.GOOS == "windows" {
		// Windows might not support ANSI colors in all terminals
		fmt.Println(text)
	} else {
		fmt.Printf("%s%s%s\n", color, text, ColorReset)
	}
}
