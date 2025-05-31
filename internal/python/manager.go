package python

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/loonghao/webhook_bridge/internal/config"
)

// InterpreterInfo contains information about a discovered Python interpreter
type InterpreterInfo struct {
	Path         string            `json:"path"`
	Version      string            `json:"version"`
	Strategy     string            `json:"strategy"`
	VenvPath     string            `json:"venv_path,omitempty"`
	IsVirtual    bool              `json:"is_virtual"`
	Capabilities map[string]bool   `json:"capabilities"`
	Environment  map[string]string `json:"environment"`
	DiscoveredAt time.Time         `json:"discovered_at"`
}

// Manager handles Python interpreter discovery and management
type Manager struct {
	config           *config.PythonConfig
	cachedInterpreter *InterpreterInfo
	logger           *log.Logger
}

// NewManager creates a new Python manager
func NewManager(cfg *config.PythonConfig) *Manager {
	return &Manager{
		config: cfg,
		logger: log.New(os.Stdout, "[Python Manager] ", log.LstdFlags),
	}
}

// GetInterpreterInfo discovers and returns detailed information about the Python interpreter
func (m *Manager) GetInterpreterInfo() (*InterpreterInfo, error) {
	// Return cached info if available and recent
	if m.cachedInterpreter != nil && time.Since(m.cachedInterpreter.DiscoveredAt) < 5*time.Minute {
		return m.cachedInterpreter, nil
	}

	// Discover interpreter
	interpreterPath, err := m.DiscoverInterpreter()
	if err != nil {
		return nil, err
	}

	// Get detailed information
	info, err := m.analyzeInterpreter(interpreterPath)
	if err != nil {
		return nil, fmt.Errorf("failed to analyze interpreter: %w", err)
	}

	// Cache the result
	m.cachedInterpreter = info

	return info, nil
}

// analyzeInterpreter analyzes a Python interpreter and returns detailed information
func (m *Manager) analyzeInterpreter(interpreterPath string) (*InterpreterInfo, error) {
	info := &InterpreterInfo{
		Path:         interpreterPath,
		Strategy:     "auto", // Use auto as default since Strategy field was renamed
		Capabilities: make(map[string]bool),
		DiscoveredAt: time.Now(),
	}

	// Get Python version
	version, err := m.getPythonVersion(interpreterPath)
	if err != nil {
		return nil, fmt.Errorf("failed to get Python version: %w", err)
	}
	info.Version = version

	// Check if it's a virtual environment
	info.IsVirtual, info.VenvPath = m.checkVirtualEnvironment(interpreterPath)

	// Test capabilities
	info.Capabilities = m.testCapabilities(interpreterPath)

	// Prepare environment
	env, err := m.PrepareEnvironment(interpreterPath)
	if err != nil {
		return nil, fmt.Errorf("failed to prepare environment: %w", err)
	}
	info.Environment = env

	return info, nil
}

// DiscoverInterpreter discovers the Python interpreter based on configuration
func (m *Manager) DiscoverInterpreter() (string, error) {
	// Use the new Interpreter field to determine strategy
	if m.config.Interpreter != "auto" && filepath.IsAbs(m.config.Interpreter) {
		return m.validateCustomPath()
	}

	// Default to auto discovery
	return m.autoDiscover()
}

// validateCustomPath validates the custom Python interpreter path
func (m *Manager) validateCustomPath() (string, error) {
	if m.config.Interpreter == "" || m.config.Interpreter == "auto" {
		return "", fmt.Errorf("custom Python path is not configured")
	}

	if _, err := os.Stat(m.config.Interpreter); err != nil {
		return "", fmt.Errorf("custom Python interpreter not found at %s: %w", m.config.Interpreter, err)
	}

	// Test if it's a valid Python interpreter
	if err := m.testInterpreter(m.config.Interpreter); err != nil {
		return "", fmt.Errorf("invalid Python interpreter at %s: %w", m.config.Interpreter, err)
	}

	return m.config.Interpreter, nil
}

// discoverUVInterpreter discovers Python interpreter using UV
func (m *Manager) discoverUVInterpreter() (string, error) {
	// Check if UV is available
	uvPath, err := exec.LookPath("uv")
	if err != nil {
		return "", fmt.Errorf("uv not found in PATH: %w", err)
	}

	// Determine project path
	projectPath := m.config.UV.ProjectPath
	if projectPath == "" {
		// Use current directory or look for pyproject.toml
		if _, err := os.Stat("pyproject.toml"); err == nil {
			projectPath = "."
		} else {
			return "", fmt.Errorf("no pyproject.toml found and no UV project path configured")
		}
	}

	// Get absolute project path
	absProjectPath, err := filepath.Abs(projectPath)
	if err != nil {
		return "", fmt.Errorf("failed to get absolute project path: %w", err)
	}

	// Create virtual environment if it doesn't exist
	venvName := m.config.UV.VenvName
	if venvName == "" {
		venvName = ".venv"
	}

	// For .venv, check if it exists directly in the project directory
	if venvName == ".venv" {
		venvPath := filepath.Join(absProjectPath, ".venv")
		if _, err := os.Stat(venvPath); err == nil {
			// .venv exists, get Python path directly
			return m.getPythonFromVenv(venvPath)
		} else {
			// Create .venv using uv
			return m.createAndGetUVVenv(uvPath, absProjectPath, venvName)
		}
	}

	// For named virtual environments, use UV's venv management
	return m.createAndGetUVVenv(uvPath, absProjectPath, venvName)
}

// getPythonFromVenv gets Python interpreter path from a virtual environment directory
func (m *Manager) getPythonFromVenv(venvPath string) (string, error) {
	var pythonPath string

	if runtime.GOOS == "windows" {
		pythonPath = filepath.Join(venvPath, "Scripts", "python.exe")
	} else {
		pythonPath = filepath.Join(venvPath, "bin", "python")
	}

	// Check if Python executable exists
	if _, err := os.Stat(pythonPath); err != nil {
		return "", fmt.Errorf("Python executable not found in venv: %s", pythonPath)
	}

	// Test the interpreter
	if err := m.testInterpreter(pythonPath); err != nil {
		return "", fmt.Errorf("invalid Python interpreter in venv: %w", err)
	}

	return pythonPath, nil
}

// createAndGetUVVenv creates a UV virtual environment and returns Python path
func (m *Manager) createAndGetUVVenv(uvPath, projectPath, venvName string) (string, error) {
	// Check if virtual environment exists using UV
	listCmd := exec.Command(uvPath, "venv", "list")
	listCmd.Dir = projectPath
	output, err := listCmd.Output()

	venvExists := err == nil && strings.Contains(string(output), venvName)

	if !venvExists {
		// Create virtual environment
		createCmd := exec.Command(uvPath, "venv", venvName)
		createCmd.Dir = projectPath
		if err := createCmd.Run(); err != nil {
			return "", fmt.Errorf("failed to create UV virtual environment '%s': %w", venvName, err)
		}
	}

	// Get Python interpreter path from UV
	pythonCmd := exec.Command(uvPath, "run", "--env", venvName, "python", "-c", "import sys; print(sys.executable)")
	pythonCmd.Dir = projectPath
	pythonPath, err := pythonCmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get Python path from UV venv '%s': %w", venvName, err)
	}

	interpreterPath := strings.TrimSpace(string(pythonPath))

	// Test the interpreter
	if err := m.testInterpreter(interpreterPath); err != nil {
		return "", fmt.Errorf("UV Python interpreter test failed for '%s': %w", venvName, err)
	}

	return interpreterPath, nil
}

// discoverFromPATH discovers Python interpreter from PATH
func (m *Manager) discoverFromPATH() (string, error) {
	// Try common Python executable names
	pythonNames := []string{"python3", "python"}
	if runtime.GOOS == "windows" {
		pythonNames = []string{"python.exe", "python3.exe"}
	}
	
	for _, name := range pythonNames {
		path, err := exec.LookPath(name)
		if err != nil {
			continue
		}
		
		// Test if it's a valid Python interpreter
		if err := m.testInterpreter(path); err != nil {
			continue
		}
		
		return path, nil
	}
	
	return "", fmt.Errorf("no valid Python interpreter found in PATH")
}

// autoDiscover automatically discovers Python interpreter using priority order
func (m *Manager) autoDiscover() (string, error) {
	// Priority order: custom -> uv -> path

	// 1. Try custom path if configured
	if m.config.Interpreter != "" && m.config.Interpreter != "auto" {
		if path, err := m.validateCustomPath(); err == nil {
			return path, nil
		}
	}
	
	// 2. Try UV if enabled
	if m.config.UV.Enabled {
		if path, err := m.discoverUVInterpreter(); err == nil {
			return path, nil
		}
	}
	
	// 3. Try PATH as fallback
	if path, err := m.discoverFromPATH(); err == nil {
		return path, nil
	}
	
	return "", fmt.Errorf("failed to discover Python interpreter using auto strategy")
}

// testInterpreter tests if the given path is a valid Python interpreter
func (m *Manager) testInterpreter(path string) error {
	cmd := exec.Command(path, "-c", "import sys; print(sys.version_info[:2])")
	output, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("failed to execute Python interpreter: %w", err)
	}
	
	version := strings.TrimSpace(string(output))
	if !strings.Contains(version, "(3,") {
		return fmt.Errorf("Python 3.x required, found: %s", version)
	}
	
	return nil
}

// GetPluginDirs returns the configured plugin directories
func (m *Manager) GetPluginDirs() []string {
	if len(m.config.PluginDirs) > 0 {
		return m.config.PluginDirs
	}
	
	// Default plugin directories
	return []string{
		"./plugins",
		"./webhook_bridge/plugins",
		"./example_plugins",
	}
}

// PrepareEnvironment prepares the Python environment for plugin execution
func (m *Manager) PrepareEnvironment(interpreterPath string) (map[string]string, error) {
	env := make(map[string]string)
	
	// Copy current environment
	for _, envVar := range os.Environ() {
		parts := strings.SplitN(envVar, "=", 2)
		if len(parts) == 2 {
			env[parts[0]] = parts[1]
		}
	}
	
	// Set Python path
	pythonDir := filepath.Dir(interpreterPath)
	if currentPath, exists := env["PATH"]; exists {
		env["PATH"] = pythonDir + string(os.PathListSeparator) + currentPath
	} else {
		env["PATH"] = pythonDir
	}
	
	// Add plugin directories to PYTHONPATH
	pluginDirs := m.GetPluginDirs()
	if len(pluginDirs) > 0 {
		pythonPath := strings.Join(pluginDirs, string(os.PathListSeparator))
		if currentPythonPath, exists := env["PYTHONPATH"]; exists {
			env["PYTHONPATH"] = pythonPath + string(os.PathListSeparator) + currentPythonPath
		} else {
			env["PYTHONPATH"] = pythonPath
		}
	}
	
	return env, nil
}

// getPythonVersion gets the Python version string
func (m *Manager) getPythonVersion(interpreterPath string) (string, error) {
	cmd := exec.Command(interpreterPath, "-c", "import sys; print(f'{sys.version_info.major}.{sys.version_info.minor}.{sys.version_info.micro}')")
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get Python version: %w", err)
	}
	return strings.TrimSpace(string(output)), nil
}

// checkVirtualEnvironment checks if the interpreter is in a virtual environment
func (m *Manager) checkVirtualEnvironment(interpreterPath string) (bool, string) {
	// Check for common virtual environment indicators
	venvPaths := []string{
		filepath.Join(filepath.Dir(interpreterPath), ".."),  // Standard venv structure
		filepath.Join(filepath.Dir(interpreterPath), "..", ".."), // Some venv structures
	}

	for _, venvPath := range venvPaths {
		absVenvPath, err := filepath.Abs(venvPath)
		if err != nil {
			continue
		}

		// Check for pyvenv.cfg (standard virtual environment marker)
		pyvenvCfg := filepath.Join(absVenvPath, "pyvenv.cfg")
		if _, err := os.Stat(pyvenvCfg); err == nil {
			return true, absVenvPath
		}

		// Check for .venv directory name
		if filepath.Base(absVenvPath) == ".venv" {
			return true, absVenvPath
		}
	}

	// Check VIRTUAL_ENV environment variable
	if venvPath := os.Getenv("VIRTUAL_ENV"); venvPath != "" {
		return true, venvPath
	}

	return false, ""
}

// testCapabilities tests what capabilities the Python interpreter has
func (m *Manager) testCapabilities(interpreterPath string) map[string]bool {
	capabilities := make(map[string]bool)

	// Test basic modules
	modules := []string{
		"sys", "os", "json", "urllib", "http",
		"grpc", "grpcio", "fastapi", "requests",
	}

	for _, module := range modules {
		cmd := exec.Command(interpreterPath, "-c", fmt.Sprintf("import %s", module))
		err := cmd.Run()
		capabilities[module] = err == nil
	}

	// Test specific features
	features := map[string]string{
		"async_support":    "import asyncio",
		"type_hints":       "from typing import Dict, List",
		"pathlib":          "from pathlib import Path",
		"dataclasses":      "from dataclasses import dataclass",
		"f_strings":        "x = 1; f'{x}'",
	}

	for feature, testCode := range features {
		cmd := exec.Command(interpreterPath, "-c", testCode)
		err := cmd.Run()
		capabilities[feature] = err == nil
	}

	return capabilities
}

// ValidateEnvironment validates that the Python environment is suitable for webhook execution
func (m *Manager) ValidateEnvironment() error {
	info, err := m.GetInterpreterInfo()
	if err != nil {
		return fmt.Errorf("failed to get interpreter info: %w", err)
	}

	m.logger.Printf("Validating Python environment: %s", info.Path)
	m.logger.Printf("Python version: %s", info.Version)
	m.logger.Printf("Strategy: %s", info.Strategy)
	m.logger.Printf("Virtual environment: %v", info.IsVirtual)

	// Check required capabilities
	requiredCapabilities := []string{"sys", "os", "json"}
	for _, capability := range requiredCapabilities {
		if !info.Capabilities[capability] {
			return fmt.Errorf("required capability '%s' not available", capability)
		}
	}

	// Check if grpc is available (required for the executor)
	if !info.Capabilities["grpc"] && !info.Capabilities["grpcio"] {
		m.logger.Printf("Warning: gRPC not available, may need to install grpcio")
	}

	// Test plugin directories
	pluginDirs := m.GetPluginDirs()
	for _, dir := range pluginDirs {
		if _, err := os.Stat(dir); err == nil {
			m.logger.Printf("Plugin directory found: %s", dir)
		}
	}

	m.logger.Printf("Python environment validation completed successfully")
	return nil
}

// InstallDependencies installs required dependencies using the appropriate package manager
func (m *Manager) InstallDependencies(packages []string) error {
	info, err := m.GetInterpreterInfo()
	if err != nil {
		return fmt.Errorf("failed to get interpreter info: %w", err)
	}

	m.logger.Printf("Installing dependencies: %v", packages)

	// Try UV first if available and we're in a UV environment
	if m.config.UV.Enabled && info.Strategy == "uv" {
		return m.installWithUV(packages)
	}

	// Fall back to pip
	return m.installWithPip(info.Path, packages)
}

// installWithUV installs packages using UV
func (m *Manager) installWithUV(packages []string) error {
	uvPath, err := exec.LookPath("uv")
	if err != nil {
		return fmt.Errorf("uv not found: %w", err)
	}

	args := append([]string{"pip", "install"}, packages...)
	cmd := exec.Command(uvPath, args...)

	if m.config.UV.ProjectPath != "" {
		cmd.Dir = m.config.UV.ProjectPath
	}

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to install with UV: %w\nOutput: %s", err, output)
	}

	m.logger.Printf("Successfully installed packages with UV: %v", packages)
	return nil
}

// installWithPip installs packages using pip
func (m *Manager) installWithPip(pythonPath string, packages []string) error {
	args := append([]string{"-m", "pip", "install"}, packages...)
	cmd := exec.Command(pythonPath, args...)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to install with pip: %w\nOutput: %s", err, output)
	}

	m.logger.Printf("Successfully installed packages with pip: %v", packages)
	return nil
}
