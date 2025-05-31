package config

import (
	"fmt"
	"os"
	"path/filepath"
)

// ConfigManager handles intelligent configuration loading and management
type ConfigManager struct {
	workingDir   string
	explicitPath string
	verbose      bool
	dirManager   *DirectoryManager
}

// NewConfigManager creates a new configuration manager
func NewConfigManager(workingDir, explicitPath string, verbose bool) *ConfigManager {
	if workingDir == "" {
		workingDir, _ = os.Getwd()
	}

	return &ConfigManager{
		workingDir:   workingDir,
		explicitPath: explicitPath,
		verbose:      verbose,
	}
}

// LoadConfig intelligently loads configuration based on the requirements
func (cm *ConfigManager) LoadConfig() (*Config, error) {
	if cm.verbose {
		fmt.Printf("📝 Loading configuration...\n")
		fmt.Printf("📁 Working directory: %s\n", cm.workingDir)
	}

	// Case 1: Explicit config path specified
	if cm.explicitPath != "" {
		if cm.verbose {
			fmt.Printf("📋 Using explicit config path: %s\n", cm.explicitPath)
		}
		return cm.loadFromExplicitPath()
	}

	// Case 2: Look for config files in working directory
	configPath := cm.findConfigInWorkingDir()
	if configPath != "" {
		if cm.verbose {
			fmt.Printf("📋 Found config file: %s\n", configPath)
		}
		return cm.loadFromPath(configPath)
	}

	// Case 3: No config file found, create default
	if cm.verbose {
		fmt.Printf("📋 No config file found, using default configuration\n")
	}
	return cm.createDefaultConfig()
}

// loadFromExplicitPath loads config from explicitly specified path
func (cm *ConfigManager) loadFromExplicitPath() (*Config, error) {
	if _, err := os.Stat(cm.explicitPath); err != nil {
		return nil, fmt.Errorf("specified config file not found: %s", cm.explicitPath)
	}

	cfg, err := LoadFromFile(cm.explicitPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load config from %s: %w", cm.explicitPath, err)
	}

	if cm.verbose {
		fmt.Printf("✅ Configuration loaded from: %s\n", cm.explicitPath)
	}

	return cfg, nil
}

// findConfigInWorkingDir searches for config files in the working directory
func (cm *ConfigManager) findConfigInWorkingDir() string {
	// Configuration file names in order of preference
	configNames := []string{
		"config.yaml",
		"config.yml",
		"webhook-bridge.yaml",
		"webhook-bridge.yml",
		"webhook_bridge.yaml",
		"webhook_bridge.yml",
	}

	for _, name := range configNames {
		configPath := filepath.Join(cm.workingDir, name)
		if _, err := os.Stat(configPath); err == nil {
			if cm.verbose {
				fmt.Printf("🔍 Found config file: %s\n", configPath)
			}
			return configPath
		}
	}

	if cm.verbose {
		fmt.Printf("🔍 No config file found in working directory\n")
	}
	return ""
}

// loadFromPath loads configuration from a specific path
func (cm *ConfigManager) loadFromPath(path string) (*Config, error) {
	cfg, err := LoadFromFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to load config from %s: %w", path, err)
	}

	if cm.verbose {
		fmt.Printf("✅ Configuration loaded from: %s\n", path)
	}

	return cfg, nil
}

// createDefaultConfig creates a default configuration
func (cm *ConfigManager) createDefaultConfig() (*Config, error) {
	cfg := Default()

	if cm.verbose {
		fmt.Printf("✅ Default configuration created\n")
		fmt.Printf("🐍 Python interpreter: %s\n", cfg.Python.Interpreter)
		fmt.Printf("🌐 Server mode: %s\n", cfg.Server.Mode)
	}

	return cfg, nil
}

// SaveDefaultConfig saves a default configuration file to the working directory
func (cm *ConfigManager) SaveDefaultConfig() error {
	configPath := filepath.Join(cm.workingDir, "config.yaml")

	// Check if config already exists
	if _, err := os.Stat(configPath); err == nil {
		if cm.verbose {
			fmt.Printf("📋 Config file already exists: %s\n", configPath)
		}
		return nil
	}

	// Create YAML content
	yamlContent := `# Webhook Bridge Configuration
server:
  host: "0.0.0.0"
  port: 0  # 0 = auto-assign port
  mode: "debug"  # debug, release
  auto_port: true

python:
  interpreter: "auto"  # "auto" or absolute path like "/usr/bin/python3"
  auto_download_uv: true
  venv_path: ".venv"
  uv:
    enabled: true
    venv_name: ".venv"
  validation:
    enabled: true
    min_python_version: "3.8"
    cache_timeout: 5
  auto_install: false
  required_packages:
    - "grpcio"
    - "grpcio-tools"

executor:
  host: "localhost"
  port: 0  # 0 = auto-assign port
  timeout: 30
  auto_port: true

logging:
  level: "info"  # debug, info, warn, error
  format: "text"  # text, json
  file: "logs/webhook-bridge.log"  # Log file path (empty = console only)
  max_size: 100  # Max log file size in MB
  max_age: 30    # Max age in days
  compress: true # Compress old log files

directories:
  working_dir: ""  # Working directory (empty = current dir)
  log_dir: "logs"  # Log directory relative to working dir
  config_dir: ""   # Config directory (empty = working dir)
  plugin_dir: "plugins"  # Plugin directory relative to working dir
  data_dir: "data"       # Data directory relative to working dir
`

	if err := os.WriteFile(configPath, []byte(yamlContent), 0644); err != nil {
		return fmt.Errorf("failed to save default config: %w", err)
	}

	if cm.verbose {
		fmt.Printf("✅ Default config saved to: %s\n", configPath)
	}

	return nil
}

// GetConfigSummary returns a summary of the loaded configuration
func (cm *ConfigManager) GetConfigSummary(cfg *Config) string {
	summary := fmt.Sprintf(`Configuration Summary:
📁 Working Directory: %s
🐍 Python Interpreter: %s
🌐 Server: %s (mode: %s)
🔧 Executor: %s (timeout: %ds)
📊 Logging: %s (%s format)`,
		cm.workingDir,
		cfg.Python.Interpreter,
		cfg.GetServerAddress(),
		cfg.Server.Mode,
		cfg.GetExecutorAddress(),
		cfg.Executor.Timeout,
		cfg.Logging.Level,
		cfg.Logging.Format,
	)

	if cm.explicitPath != "" {
		summary = fmt.Sprintf("📋 Config File: %s\n%s", cm.explicitPath, summary)
	} else if configPath := cm.findConfigInWorkingDir(); configPath != "" {
		summary = fmt.Sprintf("📋 Config File: %s\n%s", configPath, summary)
	} else {
		summary = fmt.Sprintf("📋 Config File: Default (no file found)\n%s", summary)
	}

	return summary
}

// ValidateWorkingDirectory validates that the working directory is suitable
func (cm *ConfigManager) ValidateWorkingDirectory() error {
	// Check if working directory exists
	if _, err := os.Stat(cm.workingDir); err != nil {
		return fmt.Errorf("working directory does not exist: %s", cm.workingDir)
	}

	// Check if we can write to the working directory
	testFile := filepath.Join(cm.workingDir, ".webhook_bridge_test")
	if err := os.WriteFile(testFile, []byte("test"), 0644); err != nil {
		return fmt.Errorf("cannot write to working directory: %s", cm.workingDir)
	}
	os.Remove(testFile) // Clean up

	if cm.verbose {
		fmt.Printf("✅ Working directory validated: %s\n", cm.workingDir)
	}

	return nil
}

// SetupConfigEnvironment sets up the configuration environment
func (cm *ConfigManager) SetupConfigEnvironment(cfg *Config) error {
	// Setup directory manager
	cm.dirManager = NewDirectoryManager(&cfg.Directories, cm.workingDir, cm.verbose)

	// Initialize directories
	if err := cm.dirManager.SetupDirectoryEnvironment(); err != nil {
		return fmt.Errorf("failed to setup directories: %w", err)
	}

	// Assign ports if needed
	if err := cfg.AssignPorts(); err != nil {
		return fmt.Errorf("failed to assign ports: %w", err)
	}

	// Validate configuration
	if err := cfg.Validate(); err != nil {
		return fmt.Errorf("configuration validation failed: %w", err)
	}

	if cm.verbose {
		fmt.Printf("✅ Configuration environment setup complete\n")
		fmt.Printf("🌐 Server will start on: %s\n", cfg.GetServerAddress())
		fmt.Printf("🔧 Executor will start on: %s\n", cfg.GetExecutorAddress())
	}

	return nil
}

// GetDirectoryManager returns the directory manager
func (cm *ConfigManager) GetDirectoryManager() *DirectoryManager {
	return cm.dirManager
}
