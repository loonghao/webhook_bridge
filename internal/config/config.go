package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"

	"github.com/loonghao/webhook_bridge/internal/utils"
)

// Config represents the application configuration
type Config struct {
	Server      ServerConfig      `yaml:"server"`
	Python      PythonConfig      `yaml:"python"`
	Executor    ExecutorConfig    `yaml:"executor"`
	Logging     LoggingConfig     `yaml:"logging"`
	Directories DirectoriesConfig `yaml:"directories"`
}

// ServerConfig represents HTTP server configuration
type ServerConfig struct {
	Host     string `yaml:"host" default:"0.0.0.0"`
	Port     int    `yaml:"port" default:"0"`         // 0 means auto-assign
	Mode     string `yaml:"mode" default:"debug"`     // debug, release
	AutoPort bool   `yaml:"auto_port" default:"true"` // Enable automatic port assignment
}

// PythonConfig represents Python interpreter configuration
type PythonConfig struct {
	// Python interpreter path - "auto" for auto-detection or absolute path
	Interpreter string `yaml:"interpreter" default:"auto"`

	// Auto-download UV if not available
	AutoDownloadUV bool `yaml:"auto_download_uv" default:"true"`

	// Virtual environment path
	VenvPath string `yaml:"venv_path" default:".venv"`

	// UV virtual environment settings
	UV UVConfig `yaml:"uv"`

	// Plugin directories
	PluginDirs []string `yaml:"plugin_dirs"`

	// Validation settings
	Validation ValidationConfig `yaml:"validation"`

	// Auto-install missing dependencies
	AutoInstall bool `yaml:"auto_install" default:"false"`

	// Required packages for webhook execution
	RequiredPackages []string `yaml:"required_packages"`
}

// UVConfig represents UV virtual environment configuration
type UVConfig struct {
	Enabled     bool   `yaml:"enabled" default:"true"`
	ProjectPath string `yaml:"project_path"` // Path to Python project with pyproject.toml
	VenvName    string `yaml:"venv_name" default:".venv"`
}

// ValidationConfig represents Python environment validation settings
type ValidationConfig struct {
	// Enable environment validation
	Enabled bool `yaml:"enabled" default:"true"`

	// Minimum Python version required
	MinPythonVersion string `yaml:"min_python_version" default:"3.8"`

	// Required capabilities
	RequiredCapabilities []string `yaml:"required_capabilities"`

	// Fail on missing optional dependencies
	StrictMode bool `yaml:"strict_mode" default:"false"`

	// Cache validation results (in minutes)
	CacheTimeout int `yaml:"cache_timeout" default:"5"`
}

// ExecutorConfig represents Python executor service configuration
type ExecutorConfig struct {
	Host     string `yaml:"host" default:"localhost"`
	Port     int    `yaml:"port" default:"0"`         // 0 means auto-assign
	Timeout  int    `yaml:"timeout" default:"30"`     // seconds
	AutoPort bool   `yaml:"auto_port" default:"true"` // Enable automatic port assignment
}

// LoggingConfig represents logging configuration
type LoggingConfig struct {
	Level    string `yaml:"level" default:"info"`    // debug, info, warn, error
	Format   string `yaml:"format" default:"text"`   // text, json
	File     string `yaml:"file" default:""`         // Log file path (empty = stdout only)
	MaxSize  int    `yaml:"max_size" default:"100"`  // Max log file size in MB
	MaxAge   int    `yaml:"max_age" default:"30"`    // Max age in days
	Compress bool   `yaml:"compress" default:"true"` // Compress old log files
}

// DirectoriesConfig represents directory configuration
type DirectoriesConfig struct {
	WorkingDir string `yaml:"working_dir" default:""`       // Working directory (empty = current dir)
	LogDir     string `yaml:"log_dir" default:"logs"`       // Log directory relative to working dir
	ConfigDir  string `yaml:"config_dir" default:""`        // Config directory (empty = working dir)
	PluginDir  string `yaml:"plugin_dir" default:"plugins"` // Plugin directory relative to working dir
	DataDir    string `yaml:"data_dir" default:"data"`      // Data directory relative to working dir
}

// Load loads configuration from file or environment variables
func Load() (*Config, error) {
	cfg := &Config{}

	// Set defaults
	cfg.setDefaults()

	// Try to load from config file
	configPath := getConfigPath()
	if configPath != "" {
		if err := cfg.loadFromFile(configPath); err != nil {
			return nil, fmt.Errorf("failed to load config from file %s: %w", configPath, err)
		}
	}

	// Override with environment variables
	cfg.loadFromEnv()

	return cfg, nil
}

// setDefaults sets default values for configuration
func (c *Config) setDefaults() {
	c.Server.Host = "0.0.0.0"
	c.Server.Port = 0 // Auto-assign
	c.Server.Mode = "debug"
	c.Server.AutoPort = true

	c.Python.Interpreter = "auto"
	c.Python.AutoDownloadUV = true
	c.Python.VenvPath = ".venv"
	c.Python.UV.Enabled = true
	c.Python.UV.VenvName = ".venv"
	c.Python.Validation.Enabled = true
	c.Python.Validation.MinPythonVersion = "3.8"
	c.Python.Validation.CacheTimeout = 5
	c.Python.AutoInstall = false
	c.Python.RequiredPackages = []string{"grpcio", "grpcio-tools"}

	c.Executor.Host = "localhost"
	c.Executor.Port = 0 // Auto-assign
	c.Executor.Timeout = 30
	c.Executor.AutoPort = true

	c.Logging.Level = "info"
	c.Logging.Format = "text"
	c.Logging.File = "logs/webhook-bridge.log"
	c.Logging.MaxSize = 100
	c.Logging.MaxAge = 30
	c.Logging.Compress = true

	// Directories defaults
	c.Directories.WorkingDir = ""
	c.Directories.LogDir = "logs"
	c.Directories.ConfigDir = ""
	c.Directories.PluginDir = "plugins"
	c.Directories.DataDir = "data"
}

// loadFromFile loads configuration from YAML file
func (c *Config) loadFromFile(path string) error {
	// Validate file path to prevent directory traversal attacks
	if err := validateConfigPath(path); err != nil {
		return fmt.Errorf("invalid config path: %w", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	return yaml.Unmarshal(data, c)
}

// validateConfigPath validates that the config file path is safe
func validateConfigPath(path string) error {
	// Clean the path to resolve any .. or . components
	cleanPath := filepath.Clean(path)

	// Check for directory traversal attempts
	if strings.Contains(cleanPath, "..") {
		return fmt.Errorf("directory traversal not allowed")
	}

	// Only allow specific file extensions
	ext := strings.ToLower(filepath.Ext(cleanPath))
	if ext != ".yaml" && ext != ".yml" {
		return fmt.Errorf("only .yaml and .yml files are allowed")
	}

	// Get absolute path for further validation
	absPath, err := filepath.Abs(cleanPath)
	if err != nil {
		return fmt.Errorf("failed to get absolute path: %w", err)
	}

	// Get current working directory
	wd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get working directory: %w", err)
	}

	// Allow files in current directory or user config directory
	wdAbs, _ := filepath.Abs(wd)
	if strings.HasPrefix(absPath, wdAbs) {
		return nil
	}

	// Allow files in user config directory
	if configDir, err := os.UserConfigDir(); err == nil {
		configDirAbs, _ := filepath.Abs(configDir)
		if strings.HasPrefix(absPath, configDirAbs) {
			return nil
		}
	}

	// Allow files in system temp directory (for testing)
	if tempDir := os.TempDir(); tempDir != "" {
		tempDirAbs, _ := filepath.Abs(tempDir)
		if strings.HasPrefix(absPath, tempDirAbs) {
			return nil
		}
	}

	return fmt.Errorf("config file must be in current directory, user config directory, or temp directory")
}

// loadFromEnv loads configuration from environment variables
func (c *Config) loadFromEnv() {
	if host := os.Getenv("WEBHOOK_BRIDGE_HOST"); host != "" {
		c.Server.Host = host
	}
	if portStr := os.Getenv("WEBHOOK_BRIDGE_PORT"); portStr != "" {
		if port, err := utils.ParsePort(portStr); err == nil {
			c.Server.Port = port
			c.Server.AutoPort = false // Disable auto-port when explicitly set
		}
	}
	if mode := os.Getenv("WEBHOOK_BRIDGE_MODE"); mode != "" {
		c.Server.Mode = mode
	}
	if pythonPath := os.Getenv("WEBHOOK_BRIDGE_PYTHON_PATH"); pythonPath != "" {
		c.Python.Interpreter = pythonPath
	}
	if executorPortStr := os.Getenv("WEBHOOK_BRIDGE_EXECUTOR_PORT"); executorPortStr != "" {
		if port, err := utils.ParsePort(executorPortStr); err == nil {
			c.Executor.Port = port
			c.Executor.AutoPort = false // Disable auto-port when explicitly set
		}
	}
	if executorHost := os.Getenv("WEBHOOK_BRIDGE_EXECUTOR_HOST"); executorHost != "" {
		c.Executor.Host = executorHost
	}
}

// getConfigPath returns the configuration file path
func getConfigPath() string {
	// Check for config file in order of preference
	paths := []string{
		"config.yaml",
		"config.yml",
		"webhook_bridge.yaml",
		"webhook_bridge.yml",
	}

	for _, path := range paths {
		if _, err := os.Stat(path); err == nil {
			return path
		}
	}

	// Check in user config directory
	if configDir, err := os.UserConfigDir(); err == nil {
		userConfigPath := filepath.Join(configDir, "webhook_bridge", "config.yaml")
		if _, err := os.Stat(userConfigPath); err == nil {
			return userConfigPath
		}
	}

	return ""
}

// AssignPorts assigns ports automatically if needed
func (c *Config) AssignPorts() error {
	// Assign server port if needed
	if c.Server.AutoPort && c.Server.Port == 0 {
		port, err := utils.GetPortWithFallback(8000) // Prefer 8000, fallback to any free port
		if err != nil {
			return fmt.Errorf("failed to assign server port: %w", err)
		}
		c.Server.Port = port
	}

	// Assign executor port if needed
	if c.Executor.AutoPort && c.Executor.Port == 0 {
		port, err := utils.GetPortWithFallback(50051) // Prefer 50051, fallback to any free port
		if err != nil {
			return fmt.Errorf("failed to assign executor port: %w", err)
		}
		c.Executor.Port = port
	}

	return nil
}

// GetServerAddress returns the full server address
func (c *Config) GetServerAddress() string {
	return fmt.Sprintf("%s:%d", c.Server.Host, c.Server.Port)
}

// GetExecutorAddress returns the full executor address
func (c *Config) GetExecutorAddress() string {
	return fmt.Sprintf("%s:%d", c.Executor.Host, c.Executor.Port)
}

// Validate validates the configuration
func (c *Config) Validate() error {
	if c.Server.Port < 0 || c.Server.Port > 65535 {
		return fmt.Errorf("invalid server port: %d", c.Server.Port)
	}

	if c.Executor.Port < 0 || c.Executor.Port > 65535 {
		return fmt.Errorf("invalid executor port: %d", c.Executor.Port)
	}

	if c.Executor.Timeout <= 0 {
		return fmt.Errorf("executor timeout must be positive, got: %d", c.Executor.Timeout)
	}

	return nil
}

// LoadFromFile loads configuration from a specific file
func LoadFromFile(path string) (*Config, error) {
	cfg := &Config{}
	cfg.setDefaults()

	if err := cfg.loadFromFile(path); err != nil {
		return nil, fmt.Errorf("failed to load config from file %s: %w", path, err)
	}

	// Override with environment variables
	cfg.loadFromEnv()

	return cfg, nil
}

// Default returns a configuration with default values
func Default() *Config {
	cfg := &Config{}
	cfg.setDefaults()
	cfg.loadFromEnv()
	return cfg
}
