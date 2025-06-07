package python

import (
	"fmt"
	"os/exec"
	"path/filepath"
	"sync"
	"time"

	"github.com/loonghao/webhook_bridge/internal/config"
)

// InterpreterManager manages multiple Python interpreters
type InterpreterManager struct {
	config       *config.PythonConfig
	interpreters map[string]*ManagedInterpreter
	active       string
	mutex        sync.RWMutex
}

// ManagedInterpreter represents a managed Python interpreter instance
type ManagedInterpreter struct {
	Config    *config.InterpreterConfig
	Detected  *DetectedInterpreter
	Status    InterpreterStatus
	LastCheck time.Time
	Error     error
}

// InterpreterStatus represents the status of an interpreter
type InterpreterStatus int

const (
	StatusUnknown InterpreterStatus = iota
	StatusValidating
	StatusReady
	StatusError
	StatusUnavailable
)

func (s InterpreterStatus) String() string {
	switch s {
	case StatusValidating:
		return "validating"
	case StatusReady:
		return "ready"
	case StatusError:
		return "error"
	case StatusUnavailable:
		return "unavailable"
	default:
		return "unknown"
	}
}

// NewInterpreterManager creates a new interpreter manager
func NewInterpreterManager(cfg *config.PythonConfig) *InterpreterManager {
	return &InterpreterManager{
		config:       cfg,
		interpreters: make(map[string]*ManagedInterpreter),
		active:       cfg.ActiveInterpreter,
	}
}

// AddInterpreter adds a new interpreter configuration
func (im *InterpreterManager) AddInterpreter(name string, cfg *config.InterpreterConfig) error {
	im.mutex.Lock()
	defer im.mutex.Unlock()

	if _, exists := im.interpreters[name]; exists {
		return fmt.Errorf("interpreter %s already exists", name)
	}

	managed := &ManagedInterpreter{
		Config: cfg,
		Status: StatusUnknown,
	}

	im.interpreters[name] = managed

	// Update config
	if im.config.Interpreters == nil {
		im.config.Interpreters = make(map[string]config.InterpreterConfig)
	}
	im.config.Interpreters[name] = *cfg

	return nil
}

// RemoveInterpreter removes an interpreter configuration
func (im *InterpreterManager) RemoveInterpreter(name string) error {
	im.mutex.Lock()
	defer im.mutex.Unlock()

	if name == im.active {
		return fmt.Errorf("cannot remove active interpreter %s", name)
	}

	delete(im.interpreters, name)
	if im.config.Interpreters != nil {
		delete(im.config.Interpreters, name)
	}

	return nil
}

// SetActiveInterpreter sets the active interpreter
func (im *InterpreterManager) SetActiveInterpreter(name string) error {
	im.mutex.Lock()
	defer im.mutex.Unlock()

	if _, exists := im.interpreters[name]; !exists {
		return fmt.Errorf("interpreter %s not found", name)
	}

	im.active = name
	im.config.ActiveInterpreter = name
	return nil
}

// GetActiveInterpreter returns the currently active interpreter
func (im *InterpreterManager) GetActiveInterpreter() (*ManagedInterpreter, error) {
	im.mutex.RLock()
	defer im.mutex.RUnlock()

	if im.active == "" {
		return nil, fmt.Errorf("no active interpreter set")
	}

	managed, exists := im.interpreters[im.active]
	if !exists {
		return nil, fmt.Errorf("active interpreter %s not found", im.active)
	}

	return managed, nil
}

// ListInterpreters returns all configured interpreters
func (im *InterpreterManager) ListInterpreters() map[string]*ManagedInterpreter {
	im.mutex.RLock()
	defer im.mutex.RUnlock()

	result := make(map[string]*ManagedInterpreter)
	for name, managed := range im.interpreters {
		result[name] = managed
	}
	return result
}

// ValidateInterpreter validates a specific interpreter
func (im *InterpreterManager) ValidateInterpreter(name string) error {
	im.mutex.Lock()
	managed, exists := im.interpreters[name]
	if !exists {
		im.mutex.Unlock()
		return fmt.Errorf("interpreter %s not found", name)
	}

	managed.Status = StatusValidating
	managed.LastCheck = time.Now()
	im.mutex.Unlock()

	// Perform validation outside of lock
	detected, err := im.validateInterpreterPath(managed.Config.Path)

	im.mutex.Lock()
	defer im.mutex.Unlock()

	if err != nil {
		managed.Status = StatusError
		managed.Error = err
		managed.Config.ValidationError = err.Error()
		managed.Config.Validated = false
	} else {
		managed.Status = StatusReady
		managed.Detected = detected
		managed.Error = nil
		managed.Config.ValidationError = ""
		managed.Config.Validated = true
		managed.Config.LastValidated = time.Now().Format(time.RFC3339)
	}

	return err
}

// ValidateAllInterpreters validates all configured interpreters
func (im *InterpreterManager) ValidateAllInterpreters() map[string]error {
	interpreters := im.ListInterpreters()
	results := make(map[string]error)

	for name := range interpreters {
		results[name] = im.ValidateInterpreter(name)
	}

	return results
}

// validateInterpreterPath validates a Python interpreter path
func (im *InterpreterManager) validateInterpreterPath(path string) (*DetectedInterpreter, error) {
	if path == "" {
		return nil, fmt.Errorf("interpreter path is empty")
	}

	// Check if path is absolute
	if !filepath.IsAbs(path) {
		// Try to find in PATH
		fullPath, err := exec.LookPath(path)
		if err != nil {
			return nil, fmt.Errorf("interpreter not found in PATH: %s", path)
		}
		path = fullPath
	}

	// Validate the interpreter
	return validatePythonPath(path, false)
}

// LoadFromConfig loads interpreters from configuration
func (im *InterpreterManager) LoadFromConfig() error {
	im.mutex.Lock()
	defer im.mutex.Unlock()

	if im.config.Interpreters == nil {
		return nil
	}

	for name, cfg := range im.config.Interpreters {
		cfgCopy := cfg // Create a copy to avoid pointer issues
		managed := &ManagedInterpreter{
			Config: &cfgCopy,
			Status: StatusUnknown,
		}
		im.interpreters[name] = managed
	}

	return nil
}

// GetInterpreterInfo returns detailed information about an interpreter
func (im *InterpreterManager) GetInterpreterInfo(name string) (map[string]interface{}, error) {
	im.mutex.RLock()
	managed, exists := im.interpreters[name]
	if !exists {
		im.mutex.RUnlock()
		return nil, fmt.Errorf("interpreter %s not found", name)
	}
	im.mutex.RUnlock()

	info := map[string]interface{}{
		"name":              name,
		"path":              managed.Config.Path,
		"status":            managed.Status.String(),
		"validated":         managed.Config.Validated,
		"last_validated":    managed.Config.LastValidated,
		"validation_error":  managed.Config.ValidationError,
		"last_check":        managed.LastCheck.Format(time.RFC3339),
		"use_uv":            managed.Config.UseUV,
		"venv_path":         managed.Config.VenvPath,
		"required_packages": managed.Config.RequiredPackages,
		"environment":       managed.Config.Environment,
	}

	if managed.Detected != nil {
		info["version"] = managed.Detected.Version
		info["executable"] = managed.Detected.Path
		info["is_virtual"] = managed.Detected.IsVirtual
		info["uv_managed"] = managed.Detected.UVManaged
	}

	if managed.Error != nil {
		info["error"] = managed.Error.Error()
	}

	return info, nil
}

// AutoDiscoverInterpreters automatically discovers available Python interpreters
func (im *InterpreterManager) AutoDiscoverInterpreters() ([]config.InterpreterConfig, error) {
	var discovered []config.InterpreterConfig

	// Common Python executable names
	pythonCommands := []string{"python3", "python", "python3.11", "python3.10", "python3.9", "python3.8"}

	for _, cmd := range pythonCommands {
		path, err := exec.LookPath(cmd)
		if err != nil {
			continue
		}

		detected, err := validatePythonPath(path, false)
		if err != nil {
			continue
		}

		cfg := config.InterpreterConfig{
			Name:          fmt.Sprintf("Python %s (%s)", detected.Version, cmd),
			Path:          path,
			Validated:     true,
			LastValidated: time.Now().Format(time.RFC3339),
		}

		discovered = append(discovered, cfg)
	}

	return discovered, nil
}
