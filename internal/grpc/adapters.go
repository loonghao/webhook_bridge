package grpc

import (
	"time"

	"github.com/loonghao/webhook_bridge/internal/web"
)

// LogManagerAdapter adapts web.PersistentLogManager to grpc.LogManager interface
type LogManagerAdapter struct {
	logManager *web.PersistentLogManager
}

// NewLogManagerAdapter creates a new log manager adapter
func NewLogManagerAdapter(logManager *web.PersistentLogManager) *LogManagerAdapter {
	return &LogManagerAdapter{
		logManager: logManager,
	}
}

// AddLog implements the LogManager interface
func (a *LogManagerAdapter) AddLog(entry LogEntry) {
	if a.logManager == nil {
		return
	}

	// Convert grpc.LogEntry to web.LogEntry
	webLogEntry := web.LogEntry{
		Timestamp:  entry.Timestamp,
		Level:      entry.Level,
		Source:     entry.Source,
		Message:    entry.Message,
		PluginName: entry.PluginName,
		Data:       entry.Data,
	}

	a.logManager.AddLog(webLogEntry)
}

// StatsManagerAdapter adapts web.StatsManager to grpc.StatsManager interface
type StatsManagerAdapter struct {
	statsManager *web.StatsManager
}

// NewStatsManagerAdapter creates a new stats manager adapter
func NewStatsManagerAdapter(statsManager *web.StatsManager) *StatsManagerAdapter {
	return &StatsManagerAdapter{
		statsManager: statsManager,
	}
}

// RecordExecution implements the StatsManager interface
func (a *StatsManagerAdapter) RecordExecution(plugin, method string, startTime time.Time) {
	if a.statsManager == nil {
		return
	}

	a.statsManager.RecordExecution(plugin, method, startTime)
}

// RecordError implements the StatsManager interface
func (a *StatsManagerAdapter) RecordError(plugin, method string) {
	if a.statsManager == nil {
		return
	}

	a.statsManager.RecordError(plugin, method)
}

// SetupClientWithLoggingAndStats configures a gRPC client with logging and statistics
func SetupClientWithLoggingAndStats(client *Client, logManager *web.PersistentLogManager, statsManager *web.StatsManager) {
	if logManager != nil {
		logAdapter := NewLogManagerAdapter(logManager)
		client.SetLogManager(logAdapter)
	}

	if statsManager != nil {
		statsAdapter := NewStatsManagerAdapter(statsManager)
		client.SetStatsManager(statsAdapter)
	}
}
