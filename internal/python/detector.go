package python

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/loonghao/webhook_bridge/internal/config"
)

// DetectedInterpreter contains information about a detected Python interpreter
type DetectedInterpreter struct {
	Path        string
	Version     string
	IsVirtual   bool
	VenvPath    string
	UVManaged   bool
	Executable  bool
}

// DetectionResult contains the result of Python interpreter detection
type DetectionResult struct {
	Interpreter *DetectedInterpreter
	UVAvailable bool
	VenvExists  bool
	Error       error
}

// DetectPythonInterpreter intelligently detects the best available Python interpreter
func DetectPythonInterpreter(cfg *config.PythonConfig, verbose bool) (*DetectionResult, error) {
	result := &DetectionResult{}

	if verbose {
		fmt.Printf("🐍 Starting Python interpreter detection...\n")
	}

	// Step 1: Check if absolute path is configured
	if cfg.Interpreter != "auto" && filepath.IsAbs(cfg.Interpreter) {
		if verbose {
			fmt.Printf("📍 Using configured Python path: %s\n", cfg.Interpreter)
		}
		
		interpreter, err := validatePythonPath(cfg.Interpreter, verbose)
		if err != nil {
			result.Error = fmt.Errorf("configured Python path invalid: %w", err)
			return result, result.Error
		}
		
		result.Interpreter = interpreter
		return result, nil
	}

	// Step 2: Check for existing virtual environment
	venvPath := cfg.VenvPath
	if !filepath.IsAbs(venvPath) {
		venvPath = filepath.Join(".", venvPath)
	}
	
	if venvInterpreter := checkVirtualEnvironment(venvPath, verbose); venvInterpreter != nil {
		if verbose {
			fmt.Printf("✅ Found existing virtual environment: %s\n", venvPath)
		}
		result.Interpreter = venvInterpreter
		result.VenvExists = true
		return result, nil
	}

	// Step 3: Check system Python interpreters
	if systemInterpreter := checkSystemPython(verbose); systemInterpreter != nil {
		if verbose {
			fmt.Printf("✅ Found system Python: %s\n", systemInterpreter.Path)
		}
		result.Interpreter = systemInterpreter
		return result, nil
	}

	// Step 4: Check UV availability
	uvAvailable := checkUVAvailable(verbose)
	result.UVAvailable = uvAvailable

	if uvAvailable {
		if verbose {
			fmt.Printf("✅ UV is available, will use UV-managed Python\n")
		}
		
		// Try to create virtual environment with UV
		interpreter, err := createUVEnvironment(venvPath, cfg, verbose)
		if err != nil {
			result.Error = fmt.Errorf("failed to create UV environment: %w", err)
			return result, result.Error
		}
		
		result.Interpreter = interpreter
		result.VenvExists = true
		return result, nil
	}

	// Step 5: Auto-download UV if enabled
	if cfg.AutoDownloadUV {
		if verbose {
			fmt.Printf("📥 UV not found, attempting to download and install...\n")
		}
		
		if err := downloadAndInstallUV(verbose); err != nil {
			result.Error = fmt.Errorf("failed to download UV: %w", err)
			return result, result.Error
		}
		
		// Retry with UV
		interpreter, err := createUVEnvironment(venvPath, cfg, verbose)
		if err != nil {
			result.Error = fmt.Errorf("failed to create UV environment after installation: %w", err)
			return result, result.Error
		}
		
		result.Interpreter = interpreter
		result.UVAvailable = true
		result.VenvExists = true
		return result, nil
	}

	result.Error = fmt.Errorf("no suitable Python interpreter found")
	return result, result.Error
}

// validatePythonPath validates a given Python interpreter path
func validatePythonPath(pythonPath string, verbose bool) (*DetectedInterpreter, error) {
	if verbose {
		fmt.Printf("🔍 Validating Python path: %s\n", pythonPath)
	}

	// Check if file exists and is executable
	if _, err := os.Stat(pythonPath); err != nil {
		return nil, fmt.Errorf("Python interpreter not found: %s", pythonPath)
	}

	// Try to get version
	cmd := exec.Command(pythonPath, "--version")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to get Python version: %w", err)
	}

	version := strings.TrimSpace(string(output))
	if verbose {
		fmt.Printf("📋 Python version: %s\n", version)
	}

	return &DetectedInterpreter{
		Path:       pythonPath,
		Version:    version,
		Executable: true,
	}, nil
}

// checkVirtualEnvironment checks for existing virtual environment
func checkVirtualEnvironment(venvPath string, verbose bool) *DetectedInterpreter {
	if verbose {
		fmt.Printf("🔍 Checking for virtual environment: %s\n", venvPath)
	}

	var pythonExe string
	if runtime.GOOS == "windows" {
		pythonExe = filepath.Join(venvPath, "Scripts", "python.exe")
	} else {
		pythonExe = filepath.Join(venvPath, "bin", "python")
	}

	if _, err := os.Stat(pythonExe); err != nil {
		if verbose {
			fmt.Printf("❌ Virtual environment not found: %s\n", venvPath)
		}
		return nil
	}

	// Validate the interpreter
	interpreter, err := validatePythonPath(pythonExe, verbose)
	if err != nil {
		if verbose {
			fmt.Printf("❌ Virtual environment Python invalid: %v\n", err)
		}
		return nil
	}

	interpreter.IsVirtual = true
	interpreter.VenvPath = venvPath
	return interpreter
}

// checkSystemPython checks for system Python interpreters
func checkSystemPython(verbose bool) *DetectedInterpreter {
	if verbose {
		fmt.Printf("🔍 Checking system Python interpreters...\n")
	}

	// Try different Python commands in order of preference
	pythonCommands := []string{"python3", "python"}

	for _, cmd := range pythonCommands {
		if verbose {
			fmt.Printf("🔍 Trying: %s\n", cmd)
		}

		pythonPath, err := exec.LookPath(cmd)
		if err != nil {
			if verbose {
				fmt.Printf("❌ %s not found in PATH\n", cmd)
			}
			continue
		}

		interpreter, err := validatePythonPath(pythonPath, verbose)
		if err != nil {
			if verbose {
				fmt.Printf("❌ %s validation failed: %v\n", cmd, err)
			}
			continue
		}

		if verbose {
			fmt.Printf("✅ Found valid system Python: %s\n", pythonPath)
		}
		return interpreter
	}

	if verbose {
		fmt.Printf("❌ No valid system Python found\n")
	}
	return nil
}

// checkUVAvailable checks if UV tool is available
func checkUVAvailable(verbose bool) bool {
	if verbose {
		fmt.Printf("🔍 Checking UV availability...\n")
	}

	_, err := exec.LookPath("uv")
	available := err == nil

	if verbose {
		if available {
			fmt.Printf("✅ UV is available\n")
		} else {
			fmt.Printf("❌ UV not found in PATH\n")
		}
	}

	return available
}

// createUVEnvironment creates a virtual environment using UV
func createUVEnvironment(venvPath string, cfg *config.PythonConfig, verbose bool) (*DetectedInterpreter, error) {
	if verbose {
		fmt.Printf("🔨 Creating UV virtual environment: %s\n", venvPath)
	}

	// Create virtual environment
	cmd := exec.Command("uv", "venv", venvPath)
	if verbose {
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	}

	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("failed to create UV virtual environment: %w", err)
	}

	// Check the created environment
	interpreter := checkVirtualEnvironment(venvPath, verbose)
	if interpreter == nil {
		return nil, fmt.Errorf("created virtual environment is invalid")
	}

	interpreter.UVManaged = true

	// Install required packages if specified
	if len(cfg.RequiredPackages) > 0 && cfg.AutoInstall {
		if verbose {
			fmt.Printf("📦 Installing required packages...\n")
		}
		
		if err := installPackagesWithUV(venvPath, cfg.RequiredPackages, verbose); err != nil {
			if verbose {
				fmt.Printf("⚠️  Warning: Failed to install packages: %v\n", err)
			}
		}
	}

	return interpreter, nil
}

// downloadAndInstallUV downloads and installs UV
func downloadAndInstallUV(verbose bool) error {
	if verbose {
		fmt.Printf("📥 Downloading UV...\n")
	}

	var cmd *exec.Cmd
	
	switch runtime.GOOS {
	case "windows":
		// Use PowerShell to download and install UV on Windows
		script := `irm https://astral.sh/uv/install.ps1 | iex`
		cmd = exec.Command("powershell", "-Command", script)
	case "darwin", "linux":
		// Use curl to download and install UV on Unix-like systems
		script := `curl -LsSf https://astral.sh/uv/install.sh | sh`
		cmd = exec.Command("sh", "-c", script)
	default:
		return fmt.Errorf("unsupported operating system: %s", runtime.GOOS)
	}

	if verbose {
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	}

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to download and install UV: %w", err)
	}

	if verbose {
		fmt.Printf("✅ UV installed successfully\n")
	}

	return nil
}

// installPackagesWithUV installs packages using UV
func installPackagesWithUV(venvPath string, packages []string, verbose bool) error {
	if verbose {
		fmt.Printf("📦 Installing packages: %v\n", packages)
	}

	args := append([]string{"pip", "install"}, packages...)
	cmd := exec.Command("uv", args...)
	cmd.Env = append(os.Environ(), fmt.Sprintf("VIRTUAL_ENV=%s", venvPath))

	if verbose {
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	}

	return cmd.Run()
}
