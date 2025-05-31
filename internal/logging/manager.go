package logging

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"time"

	"gopkg.in/natefinch/lumberjack.v2"

	"github.com/loonghao/webhook_bridge/internal/config"
)

// Manager manages application logging
type Manager struct {
	config  *config.LoggingConfig
	dirMgr  *config.DirectoryManager
	logger  *log.Logger
	logFile *lumberjack.Logger
	verbose bool
}

// NewManager creates a new logging manager
func NewManager(cfg *config.LoggingConfig, dirMgr *config.DirectoryManager, verbose bool) *Manager {
	return &Manager{
		config:  cfg,
		dirMgr:  dirMgr,
		verbose: verbose,
	}
}

// Initialize initializes the logging system
func (lm *Manager) Initialize() error {
	if lm.verbose {
		fmt.Printf("📝 Initializing logging system...\n")
		fmt.Printf("📝 Log level: %s\n", lm.config.Level)
		fmt.Printf("📝 Log format: %s\n", lm.config.Format)
	}

	var writers []io.Writer

	// Always include stdout
	writers = append(writers, os.Stdout)

	// Add file logging if configured
	if lm.config.File != "" {
		logPath := lm.dirMgr.GetLogFilePath(lm.config.File)

		if lm.verbose {
			fmt.Printf("📝 Log file: %s\n", logPath)
		}

		// Ensure log directory exists
		logDir := filepath.Dir(logPath)
		if err := os.MkdirAll(logDir, 0755); err != nil {
			return fmt.Errorf("failed to create log directory: %w", err)
		}

		// Setup rotating log file
		lm.logFile = &lumberjack.Logger{
			Filename:  logPath,
			MaxSize:   lm.config.MaxSize, // MB
			MaxAge:    lm.config.MaxAge,  // days
			Compress:  lm.config.Compress,
			LocalTime: true,
		}

		writers = append(writers, lm.logFile)
	}

	// Create multi-writer
	multiWriter := io.MultiWriter(writers...)

	// Create logger with appropriate format
	if lm.config.Format == "json" {
		lm.logger = log.New(multiWriter, "", 0)
	} else {
		lm.logger = log.New(multiWriter, "", log.LstdFlags|log.Lshortfile)
	}

	// Set global logger
	log.SetOutput(multiWriter)
	if lm.config.Format == "json" {
		log.SetFlags(0)
	} else {
		log.SetFlags(log.LstdFlags | log.Lshortfile)
	}

	if lm.verbose {
		fmt.Printf("✅ Logging system initialized successfully\n")
	}

	return nil
}

// GetLogger returns the configured logger
func (lm *Manager) GetLogger() *log.Logger {
	return lm.logger
}

// Log logs a message with the specified level
func (lm *Manager) Log(level, message string, args ...interface{}) {
	if !lm.shouldLog(level) {
		return
	}

	timestamp := time.Now().Format(time.RFC3339)

	if lm.config.Format == "json" {
		lm.logJSON(level, message, timestamp, args...)
	} else {
		lm.logText(level, message, timestamp, args...)
	}
}

// Debug logs a debug message
func (lm *Manager) Debug(message string, args ...interface{}) {
	lm.Log("debug", message, args...)
}

// Info logs an info message
func (lm *Manager) Info(message string, args ...interface{}) {
	lm.Log("info", message, args...)
}

// Warn logs a warning message
func (lm *Manager) Warn(message string, args ...interface{}) {
	lm.Log("warn", message, args...)
}

// Error logs an error message
func (lm *Manager) Error(message string, args ...interface{}) {
	lm.Log("error", message, args...)
}

// shouldLog determines if a message should be logged based on level
func (lm *Manager) shouldLog(level string) bool {
	levels := map[string]int{
		"debug": 0,
		"info":  1,
		"warn":  2,
		"error": 3,
	}

	configLevel, exists := levels[lm.config.Level]
	if !exists {
		configLevel = 1 // Default to info
	}

	messageLevel, exists := levels[level]
	if !exists {
		messageLevel = 1 // Default to info
	}

	return messageLevel >= configLevel
}

// logJSON logs a message in JSON format
func (lm *Manager) logJSON(level, message, timestamp string, args ...interface{}) {
	formattedMessage := fmt.Sprintf(message, args...)

	jsonLog := fmt.Sprintf(`{"timestamp":"%s","level":"%s","message":"%s","service":"webhook-bridge"}`,
		timestamp, level, formattedMessage)

	lm.logger.Println(jsonLog)
}

// logText logs a message in text format
func (lm *Manager) logText(level, message, timestamp string, args ...interface{}) {
	formattedMessage := fmt.Sprintf(message, args...)

	levelUpper := ""
	switch level {
	case "debug":
		levelUpper = "DEBUG"
	case "info":
		levelUpper = "INFO"
	case "warn":
		levelUpper = "WARN"
	case "error":
		levelUpper = "ERROR"
	default:
		levelUpper = "INFO"
	}

	textLog := fmt.Sprintf("[%s] %s: %s", levelUpper, timestamp, formattedMessage)
	lm.logger.Println(textLog)
}

// Close closes the logging system
func (lm *Manager) Close() error {
	if lm.logFile != nil {
		return lm.logFile.Close()
	}
	return nil
}

// Rotate rotates the log file
func (lm *Manager) Rotate() error {
	if lm.logFile != nil {
		return lm.logFile.Rotate()
	}
	return nil
}

// GetLogSummary returns a summary of the logging configuration
func (lm *Manager) GetLogSummary() string {
	summary := fmt.Sprintf(`Logging Configuration:
📝 Level: %s
📝 Format: %s`,
		lm.config.Level,
		lm.config.Format,
	)

	if lm.config.File != "" {
		logPath := lm.dirMgr.GetLogFilePath(lm.config.File)
		summary += fmt.Sprintf(`
📝 File: %s
📝 Max Size: %d MB
📝 Max Age: %d days
📝 Compress: %t`,
			logPath,
			lm.config.MaxSize,
			lm.config.MaxAge,
			lm.config.Compress,
		)
	} else {
		summary += "\n📝 File: Console only"
	}

	return summary
}

// SetupLoggingEnvironment sets up the complete logging environment
func (lm *Manager) SetupLoggingEnvironment() error {
	// Initialize logging
	if err := lm.Initialize(); err != nil {
		return fmt.Errorf("failed to initialize logging: %w", err)
	}

	if lm.verbose {
		fmt.Printf("✅ Logging environment setup complete\n")
		fmt.Printf("%s\n", lm.GetLogSummary())
	}

	return nil
}

// LogStartup logs application startup information
func (lm *Manager) LogStartup(version, buildTime string) {
	lm.Info("=== Webhook Bridge Starting ===")
	lm.Info("Version: %s", version)
	lm.Info("Build Time: %s", buildTime)
	lm.Info("Working Directory: %s", lm.dirMgr.GetWorkingDir())
	lm.Info("Log Directory: %s", lm.dirMgr.GetLogDir())
	lm.Info("Plugin Directory: %s", lm.dirMgr.GetPluginDir())
}

// LogShutdown logs application shutdown information
func (lm *Manager) LogShutdown() {
	lm.Info("=== Webhook Bridge Shutting Down ===")
}
