package web

import (
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// LogEntry represents a single log entry
type LogEntry struct {
	ID         int64                  `json:"id"`
	Timestamp  time.Time              `json:"timestamp"`
	Level      string                 `json:"level"`
	Source     string                 `json:"source"`
	Message    string                 `json:"message"`
	PluginName string                 `json:"plugin_name,omitempty"` // Plugin name for plugin-related logs
	Data       map[string]interface{} `json:"data,omitempty"`
}

// LogManager manages application logs for the dashboard
type LogManager struct {
	logs    []LogEntry
	clients []chan LogEntry
	mutex   sync.RWMutex
	nextID  int64
	maxLogs int
}

// NewLogManager creates a new log manager
func NewLogManager() *LogManager {
	return &LogManager{
		logs:    make([]LogEntry, 0),
		clients: make([]chan LogEntry, 0),
		maxLogs: 1000, // Keep last 1000 logs
	}
}

// AddLog adds a new log entry
func (lm *LogManager) AddLog(entry LogEntry) {
	lm.mutex.Lock()
	defer lm.mutex.Unlock()

	// Assign ID and ensure timestamp
	lm.nextID++
	entry.ID = lm.nextID
	if entry.Timestamp.IsZero() {
		entry.Timestamp = time.Now()
	}

	// Add to logs
	lm.logs = append(lm.logs, entry)

	// Trim logs if exceeding max
	if len(lm.logs) > lm.maxLogs {
		lm.logs = lm.logs[len(lm.logs)-lm.maxLogs:]
	}

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
func (lm *LogManager) GetLogs(level string, limit int) []LogEntry {
	return lm.GetLogsWithFilters(level, "", limit)
}

// GetLogsWithFilters returns logs filtered by level, plugin name and limited by count
func (lm *LogManager) GetLogsWithFilters(level string, pluginName string, limit int) []LogEntry {
	lm.mutex.RLock()
	defer lm.mutex.RUnlock()

	var filtered []LogEntry

	// Filter by level and plugin name if specified
	for _, log := range lm.logs {
		// Check level filter
		if level != "" && log.Level != level {
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
func (lm *LogManager) GetLogsByPlugin(pluginName string, limit int) []LogEntry {
	return lm.GetLogsWithFilters("", pluginName, limit)
}

// GetAvailablePlugins returns a list of all plugin names that have logs
func (lm *LogManager) GetAvailablePlugins() []string {
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
func (lm *LogManager) AddClient(client chan LogEntry) {
	lm.mutex.Lock()
	defer lm.mutex.Unlock()
	lm.clients = append(lm.clients, client)
}

// RemoveClient removes a client from log streaming
func (lm *LogManager) RemoveClient(client chan LogEntry) {
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

// ClearLogs clears all logs
func (lm *LogManager) ClearLogs() {
	lm.mutex.Lock()
	defer lm.mutex.Unlock()
	lm.logs = make([]LogEntry, 0)
}

// GetLogCount returns the total number of logs
func (lm *LogManager) GetLogCount() int {
	lm.mutex.RLock()
	defer lm.mutex.RUnlock()
	return len(lm.logs)
}

// GetLogsByTimeRange returns logs within a time range
func (lm *LogManager) GetLogsByTimeRange(start, end time.Time) []LogEntry {
	lm.mutex.RLock()
	defer lm.mutex.RUnlock()

	var filtered []LogEntry
	for _, log := range lm.logs {
		if log.Timestamp.After(start) && log.Timestamp.Before(end) {
			filtered = append(filtered, log)
		}
	}

	return filtered
}

// GetLogStats returns statistics about logs
func (lm *LogManager) GetLogStats() map[string]interface{} {
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
func (lm *LogManager) LogMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		raw := c.Request.URL.RawQuery

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
			Message:   "HTTP Request",
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
		if c.Writer.Status() >= 400 {
			entry.Level = "ERROR"
		} else if c.Writer.Status() >= 300 {
			entry.Level = "WARN"
		}

		lm.AddLog(entry)
	}
}
