package cli

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"

	"github.com/spf13/cobra"
)

// NewBuildCommand creates the build command
func NewBuildCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "build",
		Short: "Build all binaries",
		Long:  "Build Go binaries and prepare Python environment",
		RunE:  runBuild,
	}

	cmd.Flags().Bool("skip-python", false, "Skip Python environment setup")
	cmd.Flags().Bool("cross-platform", false, "Build for all platforms")

	return cmd
}

func runBuild(cmd *cobra.Command, args []string) error {
	verbose, _ := cmd.Flags().GetBool("verbose")
	skipPython, _ := cmd.Flags().GetBool("skip-python")
	crossPlatform, _ := cmd.Flags().GetBool("cross-platform")

	if verbose {
		fmt.Println("🔨 Starting build process...")
	}

	// Create build directory with secure permissions
	buildDir := "build"
	if err := os.MkdirAll(buildDir, 0750); err != nil {
		return fmt.Errorf("failed to create build directory: %w", err)
	}

	// Build Go binaries
	if err := buildGoBinaries(buildDir, crossPlatform, verbose); err != nil {
		return fmt.Errorf("failed to build Go binaries: %w", err)
	}

	// Setup Python environment
	if !skipPython {
		if err := setupPythonEnvironment(verbose); err != nil {
			return fmt.Errorf("failed to setup Python environment: %w", err)
		}
	}

	fmt.Println("✅ Build completed successfully!")
	return nil
}

func buildGoBinaries(buildDir string, crossPlatform bool, verbose bool) error {
	if verbose {
		fmt.Println("🔨 Building Go binaries...")
	}

	// Find Go executable
	goCmd := findGoCommand()
	if goCmd == "" {
		return fmt.Errorf("Go compiler not found. Please install Go or add it to PATH")
	}

	// Define build targets
	targets := []struct {
		name string
		path string
	}{
		{"webhook-bridge-server", "./cmd/server"},
		{"python-manager", "./cmd/python-manager"},
	}

	// Define platforms for cross-platform build
	platforms := []struct {
		goos   string
		goarch string
		suffix string
	}{
		{"windows", "amd64", ".exe"},
		{"linux", "amd64", ""},
		{"darwin", "amd64", ""},
	}

	if crossPlatform {
		// Build for all platforms
		for _, platform := range platforms {
			for _, target := range targets {
				outputName := target.name
				if platform.goos != "linux" {
					outputName += "-" + platform.goos
				}
				outputName += platform.suffix

				outputPath := filepath.Join(buildDir, outputName)

				if verbose {
					fmt.Printf("  Building %s for %s/%s...\n", target.name, platform.goos, platform.goarch)
				}

				cmd := exec.Command(goCmd, "build", "-o", outputPath, target.path)
				cmd.Env = append(os.Environ(),
					"GOOS="+platform.goos,
					"GOARCH="+platform.goarch,
				)

				if err := cmd.Run(); err != nil {
					return fmt.Errorf("failed to build %s for %s/%s: %w", target.name, platform.goos, platform.goarch, err)
				}
			}
		}
	} else {
		// Build for current platform only
		suffix := ""
		if runtime.GOOS == "windows" {
			suffix = ".exe"
		}

		for _, target := range targets {
			outputPath := filepath.Join(buildDir, target.name+suffix)

			if verbose {
				fmt.Printf("  Building %s...\n", target.name)
			}

			cmd := exec.Command(goCmd, "build", "-o", outputPath, target.path)
			if verbose {
				cmd.Stdout = os.Stdout
				cmd.Stderr = os.Stderr
			}
			if err := cmd.Run(); err != nil {
				return fmt.Errorf("failed to build %s: %w", target.name, err)
			}
		}
	}

	return nil
}

func setupPythonEnvironment(verbose bool) error {
	if verbose {
		fmt.Println("🐍 Setting up Python environment...")
	}

	// Check if Python is available
	pythonCmd := findPythonCommand()
	if pythonCmd == "" {
		return fmt.Errorf("Python not found. Please install Python 3.8 or later")
	}

	if verbose {
		fmt.Printf("  Found Python: %s\n", pythonCmd)
	}

	// Check if virtual environment exists
	venvDir := ".venv"
	if _, err := os.Stat(venvDir); os.IsNotExist(err) {
		if verbose {
			fmt.Println("  Creating Python virtual environment...")
		}

		// Try UV first, then fall back to venv
		if err := createVenvWithUV(verbose); err != nil {
			if verbose {
				fmt.Println("  UV not available, using standard venv...")
			}
			if err := createVenvWithPython(pythonCmd, verbose); err != nil {
				return fmt.Errorf("failed to create virtual environment: %w", err)
			}
		}
	}

	// Install dependencies
	if verbose {
		fmt.Println("  Installing Python dependencies...")
	}

	if err := installPythonDependencies(verbose); err != nil {
		return fmt.Errorf("failed to install Python dependencies: %w", err)
	}

	return nil
}

func findGoCommand() string {
	// Try common Go installation paths
	commands := []string{
		"go",
		"C:\\Program Files\\Go\\bin\\go.exe",
		"/usr/local/go/bin/go",
		"/usr/bin/go",
	}

	for _, cmd := range commands {
		if _, err := exec.LookPath(cmd); err == nil {
			return cmd
		}
		// Also try direct file existence check for absolute paths
		if filepath.IsAbs(cmd) {
			if _, err := os.Stat(cmd); err == nil {
				return cmd
			}
		}
	}
	return ""
}

func findPythonCommand() string {
	commands := []string{"python3", "python"}
	for _, cmd := range commands {
		if _, err := exec.LookPath(cmd); err == nil {
			return cmd
		}
	}
	return ""
}

func createVenvWithUV(verbose bool) error {
	cmd := exec.Command("uv", "venv", ".venv")
	if verbose {
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	}
	return cmd.Run()
}

func createVenvWithPython(pythonCmd string, verbose bool) error {
	cmd := exec.Command(pythonCmd, "-m", "venv", ".venv")
	if verbose {
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	}
	return cmd.Run()
}

func installPythonDependencies(verbose bool) error {
	// Determine the correct pip path
	var pipPath string
	if runtime.GOOS == "windows" {
		pipPath = filepath.Join(".venv", "Scripts", "python.exe")
	} else {
		pipPath = filepath.Join(".venv", "bin", "python")
	}

	// Check if UV is available for faster installation
	if _, err := exec.LookPath("uv"); err == nil {
		cmd := exec.Command("uv", "pip", "install", "-r", "requirements.txt")
		cmd.Dir = "."
		if verbose {
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
		}
		return cmd.Run()
	}

	// Fall back to regular pip
	cmd := exec.Command(pipPath, "-m", "pip", "install", "-r", "requirements.txt")
	if verbose {
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	}
	return cmd.Run()
}
