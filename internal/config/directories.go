package config

import (
	"fmt"
	"os"
	"path/filepath"
)

// DirectoryManager manages application directories
type DirectoryManager struct {
	config     *DirectoriesConfig
	workingDir string
	verbose    bool
}

// NewDirectoryManager creates a new directory manager
func NewDirectoryManager(cfg *DirectoriesConfig, workingDir string, verbose bool) *DirectoryManager {
	if workingDir == "" {
		workingDir, _ = os.Getwd()
	}

	return &DirectoryManager{
		config:     cfg,
		workingDir: workingDir,
		verbose:    verbose,
	}
}

// Initialize initializes all required directories
func (dm *DirectoryManager) Initialize() error {
	if dm.verbose {
		fmt.Printf("📁 Initializing directories...\n")
		fmt.Printf("📁 Working directory: %s\n", dm.workingDir)
	}

	// Create all required directories
	dirs := []struct {
		name string
		path string
	}{
		{"logs", dm.GetLogDir()},
		{"plugins", dm.GetPluginDir()},
		{"data", dm.GetDataDir()},
	}

	for _, dir := range dirs {
		if err := dm.ensureDirectory(dir.path, dir.name); err != nil {
			return fmt.Errorf("failed to create %s directory: %w", dir.name, err)
		}
	}

	if dm.verbose {
		fmt.Printf("✅ All directories initialized successfully\n")
	}

	return nil
}

// GetWorkingDir returns the working directory
func (dm *DirectoryManager) GetWorkingDir() string {
	if dm.config.WorkingDir != "" && filepath.IsAbs(dm.config.WorkingDir) {
		return dm.config.WorkingDir
	}
	return dm.workingDir
}

// GetLogDir returns the log directory path
func (dm *DirectoryManager) GetLogDir() string {
	if filepath.IsAbs(dm.config.LogDir) {
		return dm.config.LogDir
	}
	return filepath.Join(dm.GetWorkingDir(), dm.config.LogDir)
}

// GetConfigDir returns the config directory path
func (dm *DirectoryManager) GetConfigDir() string {
	if dm.config.ConfigDir != "" {
		if filepath.IsAbs(dm.config.ConfigDir) {
			return dm.config.ConfigDir
		}
		return filepath.Join(dm.GetWorkingDir(), dm.config.ConfigDir)
	}
	return dm.GetWorkingDir()
}

// GetPluginDir returns the plugin directory path
func (dm *DirectoryManager) GetPluginDir() string {
	if filepath.IsAbs(dm.config.PluginDir) {
		return dm.config.PluginDir
	}
	return filepath.Join(dm.GetWorkingDir(), dm.config.PluginDir)
}

// GetDataDir returns the data directory path
func (dm *DirectoryManager) GetDataDir() string {
	if filepath.IsAbs(dm.config.DataDir) {
		return dm.config.DataDir
	}
	return filepath.Join(dm.GetWorkingDir(), dm.config.DataDir)
}

// GetLogFilePath returns the full log file path
func (dm *DirectoryManager) GetLogFilePath(logFile string) string {
	if logFile == "" {
		return ""
	}

	if filepath.IsAbs(logFile) {
		return logFile
	}

	// If relative path, make it relative to working directory
	return filepath.Join(dm.GetWorkingDir(), logFile)
}

// ensureDirectory creates a directory if it doesn't exist
func (dm *DirectoryManager) ensureDirectory(path, name string) error {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		if dm.verbose {
			fmt.Printf("📁 Creating %s directory: %s\n", name, path)
		}

		if err := os.MkdirAll(path, 0750); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", path, err)
		}
	} else if err != nil {
		return fmt.Errorf("failed to check directory %s: %w", path, err)
	} else if dm.verbose {
		fmt.Printf("✅ %s directory exists: %s\n", name, path)
	}

	return nil
}

// GetDirectorySummary returns a summary of all directories
func (dm *DirectoryManager) GetDirectorySummary() string {
	return fmt.Sprintf(`Directory Configuration:
📁 Working Directory: %s
📁 Log Directory: %s
📁 Config Directory: %s
📁 Plugin Directory: %s
📁 Data Directory: %s`,
		dm.GetWorkingDir(),
		dm.GetLogDir(),
		dm.GetConfigDir(),
		dm.GetPluginDir(),
		dm.GetDataDir(),
	)
}

// ValidateDirectories validates that all directories are accessible
func (dm *DirectoryManager) ValidateDirectories() error {
	dirs := map[string]string{
		"working": dm.GetWorkingDir(),
		"log":     dm.GetLogDir(),
		"config":  dm.GetConfigDir(),
		"plugin":  dm.GetPluginDir(),
		"data":    dm.GetDataDir(),
	}

	for name, path := range dirs {
		if err := dm.validateDirectory(path, name); err != nil {
			return err
		}
	}

	return nil
}

// validateDirectory validates a single directory
func (dm *DirectoryManager) validateDirectory(path, name string) error {
	// Check if directory exists
	info, err := os.Stat(path)
	if err != nil {
		return fmt.Errorf("%s directory does not exist: %s", name, path)
	}

	// Check if it's actually a directory
	if !info.IsDir() {
		return fmt.Errorf("%s path is not a directory: %s", name, path)
	}

	// Check if we can write to it
	testFile := filepath.Join(path, ".webhook_bridge_test")
	if err := os.WriteFile(testFile, []byte("test"), 0600); err != nil {
		return fmt.Errorf("cannot write to %s directory: %s", name, path)
	}
	os.Remove(testFile) // Clean up

	if dm.verbose {
		fmt.Printf("✅ %s directory validated: %s\n", name, path)
	}

	return nil
}

// SetupDirectoryEnvironment sets up the complete directory environment
func (dm *DirectoryManager) SetupDirectoryEnvironment() error {
	// Initialize directories
	if err := dm.Initialize(); err != nil {
		return fmt.Errorf("failed to initialize directories: %w", err)
	}

	// Validate directories
	if err := dm.ValidateDirectories(); err != nil {
		return fmt.Errorf("directory validation failed: %w", err)
	}

	if dm.verbose {
		fmt.Printf("✅ Directory environment setup complete\n")
		fmt.Printf("%s\n", dm.GetDirectorySummary())
	}

	return nil
}

// CleanupOldLogs cleans up old log files based on configuration
func (dm *DirectoryManager) CleanupOldLogs(maxAge int) error {
	logDir := dm.GetLogDir()

	if dm.verbose {
		fmt.Printf("🧹 Cleaning up logs older than %d days in: %s\n", maxAge, logDir)
	}

	// Implementation for log cleanup would go here
	// This is a placeholder for now

	return nil
}
