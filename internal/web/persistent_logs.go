package web

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// PersistentLogManager manages application logs with file persistence
type PersistentLogManager struct {
	logs    []LogEntry
	clients []chan LogEntry
	mutex   sync.RWMutex
	nextID  int64
	maxLogs int
	logFile string
	logDir  string
	enabled bool
}

// NewPersistentLogManager creates a new persistent log manager
func NewPersistentLogManager(logDir string, maxLogs int) *PersistentLogManager {
	if maxLogs <= 0 {
		maxLogs = 1000
	}

	// Ensure log directory exists
	if logDir == "" {
		logDir = "./logs"
	}

	if err := os.MkdirAll(logDir, 0755); err != nil {
		log.Printf("Warning: Failed to create log directory %s: %v", logDir, err)
		logDir = "." // Fallback to current directory
	}

	logFile := filepath.Join(logDir, "webhook-bridge-dashboard.jsonl")

	lm := &PersistentLogManager{
		logs:    make([]LogEntry, 0),
		clients: make([]chan LogEntry, 0),
		maxLogs: maxLogs,
		logFile: logFile,
		logDir:  logDir,
		enabled: true,
	}

	// Load existing logs from file
	lm.loadLogsFromFile()

	// Add some initial logs if none exist
	if len(lm.logs) == 0 {
		lm.addInitialLogs()
	}

	return lm
}

// loadLogsFromFile loads logs from the persistent file
func (lm *PersistentLogManager) loadLogsFromFile() {
	file, err := os.Open(lm.logFile)
	if err != nil {
		if !os.IsNotExist(err) {
			log.Printf("Warning: Failed to open log file %s: %v", lm.logFile, err)
		}
		return
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	var loadedLogs []LogEntry
	maxID := int64(0)

	for scanner.Scan() {
		var entry LogEntry
		if err := json.Unmarshal(scanner.Bytes(), &entry); err != nil {
			log.Printf("Warning: Failed to parse log entry: %v", err)
			continue
		}
		loadedLogs = append(loadedLogs, entry)
		if entry.ID > maxID {
			maxID = entry.ID
		}
	}

	if err := scanner.Err(); err != nil {
		log.Printf("Warning: Error reading log file: %v", err)
	}

	// Keep only the most recent logs
	if len(loadedLogs) > lm.maxLogs {
		loadedLogs = loadedLogs[len(loadedLogs)-lm.maxLogs:]
	}

	lm.logs = loadedLogs
	lm.nextID = maxID

	log.Printf("Loaded %d logs from %s", len(lm.logs), lm.logFile)
}

// saveLogToFile appends a log entry to the persistent file
func (lm *PersistentLogManager) saveLogToFile(entry LogEntry) {
	if !lm.enabled {
		return
	}

	file, err := os.OpenFile(lm.logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		log.Printf("Warning: Failed to open log file for writing: %v", err)
		return
	}
	defer file.Close()

	data, err := json.Marshal(entry)
	if err != nil {
		log.Printf("Warning: Failed to marshal log entry: %v", err)
		return
	}

	if _, err := file.Write(append(data, '\n')); err != nil {
		log.Printf("Warning: Failed to write log entry: %v", err)
	}
}

// addInitialLogs adds some sample logs for demonstration
func (lm *PersistentLogManager) addInitialLogs() {
	initialLogs := []LogEntry{
		{
			Timestamp: time.Now().Add(-30 * time.Minute),
			Level:     "INFO",
			Source:    "system",
			Message:   "Webhook Bridge service started",
			Data: map[string]interface{}{
				"version": "2.0.0-hybrid",
				"port":    8000,
			},
		},
		{
			Timestamp: time.Now().Add(-25 * time.Minute),
			Level:     "INFO",
			Source:    "plugin_manager",
			Message:   "Plugin directory scanned",
			Data: map[string]interface{}{
				"plugins_found": 3,
				"directory":     "./plugins",
			},
		},
		{
			Timestamp: time.Now().Add(-20 * time.Minute),
			Level:     "INFO",
			Source:    "grpc",
			Message:   "gRPC connection established",
			Data: map[string]interface{}{
				"address": "localhost:50051",
			},
		},
		{
			Timestamp: time.Now().Add(-15 * time.Minute),
			Level:     "WARN",
			Source:    "system",
			Message:   "High memory usage detected",
			Data: map[string]interface{}{
				"memory_usage": "85%",
				"threshold":    "80%",
			},
		},
		{
			Timestamp: time.Now().Add(-10 * time.Minute),
			Level:     "INFO",
			Source:    "webhook",
			Message:   "Webhook request processed successfully",
			Data: map[string]interface{}{
				"plugin":         "example_plugin",
				"execution_time": "45ms",
				"status":         "success",
			},
		},
	}

	for _, entry := range initialLogs {
		lm.AddLog(entry)
	}
}

// AddLog adds a new log entry with persistence
func (lm *PersistentLogManager) AddLog(entry LogEntry) {
	lm.mutex.Lock()
	defer lm.mutex.Unlock()

	// Assign ID and ensure timestamp
	lm.nextID++
	entry.ID = lm.nextID
	if entry.Timestamp.IsZero() {
		entry.Timestamp = time.Now()
	}

	// Add to memory
	lm.logs = append(lm.logs, entry)

	// Trim logs if exceeding max
	if len(lm.logs) > lm.maxLogs {
		lm.logs = lm.logs[len(lm.logs)-lm.maxLogs:]
	}

	// Save to file
	go lm.saveLogToFile(entry)

	// Broadcast to all clients
	for _, client := range lm.clients {
		select {
		case client <- entry:
		default:
			// Client channel is full, skip
		}
	}
}

// GetLogs returns logs filtered by level and limited by count
func (lm *PersistentLogManager) GetLogs(level string, limit int) []LogEntry {
	return lm.GetLogsWithFilters(level, "", limit)
}

// GetLogsWithFilters returns logs filtered by level, plugin name and limited by count
func (lm *PersistentLogManager) GetLogsWithFilters(level string, pluginName string, limit int) []LogEntry {
	lm.mutex.RLock()
	defer lm.mutex.RUnlock()

	var filtered []LogEntry

	// Filter by level and plugin name if specified
	for _, log := range lm.logs {
		// Check level filter
		if level != "" && level != "all" && log.Level != level {
			continue
		}

		// Check plugin name filter
		if pluginName != "" && log.PluginName != pluginName {
			continue
		}

		filtered = append(filtered, log)
	}

	// Apply limit
	if limit > 0 && len(filtered) > limit {
		filtered = filtered[len(filtered)-limit:]
	}

	// Reverse to show newest first
	for i, j := 0, len(filtered)-1; i < j; i, j = i+1, j-1 {
		filtered[i], filtered[j] = filtered[j], filtered[i]
	}

	return filtered
}

// GetLogsByPlugin returns logs for a specific plugin
func (lm *PersistentLogManager) GetLogsByPlugin(pluginName string, limit int) []LogEntry {
	return lm.GetLogsWithFilters("", pluginName, limit)
}

// GetAvailablePlugins returns a list of all plugin names that have logs
func (lm *PersistentLogManager) GetAvailablePlugins() []string {
	lm.mutex.RLock()
	defer lm.mutex.RUnlock()

	pluginSet := make(map[string]bool)
	for _, log := range lm.logs {
		if log.PluginName != "" {
			pluginSet[log.PluginName] = true
		}
	}

	var plugins []string
	for plugin := range pluginSet {
		plugins = append(plugins, plugin)
	}

	return plugins
}

// AddClient adds a client for real-time log streaming
func (lm *PersistentLogManager) AddClient(client chan LogEntry) {
	lm.mutex.Lock()
	defer lm.mutex.Unlock()
	lm.clients = append(lm.clients, client)
}

// RemoveClient removes a client from log streaming
func (lm *PersistentLogManager) RemoveClient(client chan LogEntry) {
	lm.mutex.Lock()
	defer lm.mutex.Unlock()

	for i, c := range lm.clients {
		if c == client {
			lm.clients = append(lm.clients[:i], lm.clients[i+1:]...)
			close(client)
			break
		}
	}
}

// ClearLogs clears all logs from memory and file
func (lm *PersistentLogManager) ClearLogs() {
	lm.mutex.Lock()
	defer lm.mutex.Unlock()

	lm.logs = make([]LogEntry, 0)

	// Clear the log file
	if err := os.Truncate(lm.logFile, 0); err != nil {
		log.Printf("Warning: Failed to clear log file: %v", err)
	}
}

// GetLogCount returns the total number of logs
func (lm *PersistentLogManager) GetLogCount() int {
	lm.mutex.RLock()
	defer lm.mutex.RUnlock()
	return len(lm.logs)
}

// GetLogStats returns statistics about logs
func (lm *PersistentLogManager) GetLogStats() map[string]interface{} {
	lm.mutex.RLock()
	defer lm.mutex.RUnlock()

	stats := map[string]interface{}{
		"total":   len(lm.logs),
		"levels":  make(map[string]int),
		"sources": make(map[string]int),
		"plugins": make(map[string]int),
	}

	levels := stats["levels"].(map[string]int)
	sources := stats["sources"].(map[string]int)
	plugins := stats["plugins"].(map[string]int)

	for _, log := range lm.logs {
		levels[log.Level]++
		sources[log.Source]++
		if log.PluginName != "" {
			plugins[log.PluginName]++
		}
	}

	return stats
}

// LogMiddleware creates a Gin middleware for logging requests
func (lm *PersistentLogManager) LogMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		raw := c.Request.URL.RawQuery

		// Skip logging for static assets and frequent endpoints
		if shouldSkipLogging(path) {
			c.Next()
			return
		}

		// Process request
		c.Next()

		// Log after request
		end := time.Now()
		latency := end.Sub(start)

		if raw != "" {
			path = path + "?" + raw
		}

		entry := LogEntry{
			Timestamp: end,
			Level:     "INFO",
			Source:    "http",
			Message:   fmt.Sprintf("%s %s", c.Request.Method, path),
			Data: map[string]interface{}{
				"method":     c.Request.Method,
				"path":       path,
				"status":     c.Writer.Status(),
				"latency":    latency.String(),
				"ip":         c.ClientIP(),
				"user_agent": c.Request.UserAgent(),
			},
		}

		// Set level based on status code
		if c.Writer.Status() >= 500 {
			entry.Level = "ERROR"
		} else if c.Writer.Status() >= 400 {
			entry.Level = "WARN"
		}

		lm.AddLog(entry)
	}
}

// shouldSkipLogging determines if a path should be skipped from logging
func shouldSkipLogging(path string) bool {
	skipPaths := []string{
		"/assets/",
		// "/favicon.ico", // Temporarily enable favicon logging for debugging
		"/api/dashboard/logs/stream", // Skip WebSocket connections
	}

	for _, skipPath := range skipPaths {
		if len(path) >= len(skipPath) && path[:len(skipPath)] == skipPath {
			return true
		}
	}

	return false
}

// AddTestLog adds a test log entry for demonstration
func (lm *PersistentLogManager) AddTestLog() {
	entry := LogEntry{
		Timestamp: time.Now(),
		Level:     "INFO",
		Source:    "test",
		Message:   "Test log entry added",
		Data: map[string]interface{}{
			"test": true,
		},
	}
	lm.AddLog(entry)
}

// Close closes the log manager and cleans up resources
func (lm *PersistentLogManager) Close() {
	lm.mutex.Lock()
	defer lm.mutex.Unlock()

	// Close all client channels
	for _, client := range lm.clients {
		close(client)
	}
	lm.clients = nil
	lm.enabled = false
}
