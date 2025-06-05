package web

import (
	"fmt"
	"runtime"
	"sync"
	"time"
)

// ExecutionStats represents execution statistics
type ExecutionStats struct {
	Plugin    string        `json:"plugin"`
	Method    string        `json:"method"`
	Count     int64         `json:"count"`
	LastExec  time.Time     `json:"last_execution"`
	AvgTime   time.Duration `json:"average_time"`
	TotalTime time.Duration `json:"total_time"`
	Errors    int64         `json:"errors"`
}

// SystemStats represents system statistics
type SystemStats struct {
	Uptime          time.Duration `json:"uptime"`
	TotalRequests   int64         `json:"total_requests"`
	TotalExecutions int64         `json:"total_executions"`
	TotalErrors     int64         `json:"total_errors"`
	MemoryUsage     uint64        `json:"memory_usage"`
	Goroutines      int           `json:"goroutines"`
	CPUUsage        float64       `json:"cpu_usage"`
}

// StatsManager manages execution and system statistics
type StatsManager struct {
	startTime       time.Time
	pluginStats     map[string]*ExecutionStats
	totalRequests   int64
	totalExecutions int64
	totalErrors     int64
	mutex           sync.RWMutex
	storage         *PluginStatsStorage // Persistent storage
	persistEnabled  bool                // Whether persistence is enabled
}

// NewStatsManager creates a new statistics manager
func NewStatsManager() *StatsManager {
	return &StatsManager{
		startTime:      time.Now(),
		pluginStats:    make(map[string]*ExecutionStats),
		persistEnabled: false, // Disabled by default
	}
}

// NewStatsManagerWithPersistence creates a new statistics manager with persistence
func NewStatsManagerWithPersistence(dataDir string) *StatsManager {
	storage := NewPluginStatsStorage(dataDir)

	sm := &StatsManager{
		startTime:      time.Now(),
		pluginStats:    make(map[string]*ExecutionStats),
		storage:        storage,
		persistEnabled: true,
	}

	// Load existing data
	sm.LoadStats()

	return sm
}

// RecordExecution records a plugin execution
func (sm *StatsManager) RecordExecution(plugin, method string, startTime time.Time) {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()

	duration := time.Since(startTime)
	key := plugin + ":" + method

	stats, exists := sm.pluginStats[key]
	if !exists {
		stats = &ExecutionStats{
			Plugin: plugin,
			Method: method,
		}
		sm.pluginStats[key] = stats
	}

	stats.Count++
	stats.LastExec = time.Now()
	stats.TotalTime += duration

	// Calculate average time
	stats.AvgTime = stats.TotalTime / time.Duration(stats.Count)

	sm.totalExecutions++

	// Request async save if persistence is enabled
	sm.requestSave()
}

// RecordError records an execution error
func (sm *StatsManager) RecordError(plugin, method string) {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()

	key := plugin + ":" + method
	if stats, exists := sm.pluginStats[key]; exists {
		stats.Errors++
	}

	sm.totalErrors++

	// Request async save if persistence is enabled
	sm.requestSave()
}

// RecordRequest records an HTTP request
func (sm *StatsManager) RecordRequest() {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()
	sm.totalRequests++
}

// GetStats returns overall system statistics
func (sm *StatsManager) GetStats() SystemStats {
	sm.mutex.RLock()
	defer sm.mutex.RUnlock()

	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	return SystemStats{
		Uptime:          time.Since(sm.startTime),
		TotalRequests:   sm.totalRequests,
		TotalExecutions: sm.totalExecutions,
		TotalErrors:     sm.totalErrors,
		MemoryUsage:     m.Alloc,
		Goroutines:      runtime.NumGoroutine(),
		CPUUsage:        sm.getCPUUsage(),
	}
}

// GetPluginStats returns plugin-specific statistics
func (sm *StatsManager) GetPluginStats() map[string]*ExecutionStats {
	sm.mutex.RLock()
	defer sm.mutex.RUnlock()

	// Create a copy to avoid race conditions
	result := make(map[string]*ExecutionStats)
	for key, stats := range sm.pluginStats {
		statsCopy := *stats
		result[key] = &statsCopy
	}

	return result
}

// GetTopPlugins returns the most frequently used plugins
func (sm *StatsManager) GetTopPlugins(limit int) []*ExecutionStats {
	sm.mutex.RLock()
	defer sm.mutex.RUnlock()

	var stats []*ExecutionStats
	for _, stat := range sm.pluginStats {
		statCopy := *stat
		stats = append(stats, &statCopy)
	}

	// Sort by execution count (descending)
	for i := 0; i < len(stats)-1; i++ {
		for j := i + 1; j < len(stats); j++ {
			if stats[i].Count < stats[j].Count {
				stats[i], stats[j] = stats[j], stats[i]
			}
		}
	}

	if limit > 0 && len(stats) > limit {
		stats = stats[:limit]
	}

	return stats
}

// GetUptime returns the service uptime
func (sm *StatsManager) GetUptime() time.Duration {
	return time.Since(sm.startTime)
}

// GetUptimeString returns the uptime as a formatted string
func (sm *StatsManager) GetUptimeString() string {
	uptime := sm.GetUptime()
	days := int(uptime.Hours()) / 24
	hours := int(uptime.Hours()) % 24
	minutes := int(uptime.Minutes()) % 60
	seconds := int(uptime.Seconds()) % 60

	if days > 0 {
		return fmt.Sprintf("%dd %dh %dm %ds", days, hours, minutes, seconds)
	} else if hours > 0 {
		return fmt.Sprintf("%dh %dm %ds", hours, minutes, seconds)
	} else if minutes > 0 {
		return fmt.Sprintf("%dm %ds", minutes, seconds)
	} else {
		return fmt.Sprintf("%ds", seconds)
	}
}

// GetErrorRate returns the overall error rate
func (sm *StatsManager) GetErrorRate() float64 {
	sm.mutex.RLock()
	defer sm.mutex.RUnlock()

	if sm.totalExecutions == 0 {
		return 0.0
	}

	return float64(sm.totalErrors) / float64(sm.totalExecutions) * 100
}

// GetRequestsPerSecond returns the average requests per second
func (sm *StatsManager) GetRequestsPerSecond() float64 {
	sm.mutex.RLock()
	defer sm.mutex.RUnlock()

	uptime := time.Since(sm.startTime).Seconds()
	if uptime == 0 {
		return 0.0
	}

	return float64(sm.totalRequests) / uptime
}

// GetExecutionsPerSecond returns the average executions per second
func (sm *StatsManager) GetExecutionsPerSecond() float64 {
	sm.mutex.RLock()
	defer sm.mutex.RUnlock()

	uptime := time.Since(sm.startTime).Seconds()
	if uptime == 0 {
		return 0.0
	}

	return float64(sm.totalExecutions) / uptime
}

// Reset resets all statistics
func (sm *StatsManager) Reset() {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()

	sm.startTime = time.Now()
	sm.pluginStats = make(map[string]*ExecutionStats)
	sm.totalRequests = 0
	sm.totalExecutions = 0
	sm.totalErrors = 0
}

// getCPUUsage returns a simple CPU usage estimate
// Note: This is a simplified implementation
func (sm *StatsManager) getCPUUsage() float64 {
	// This is a placeholder implementation
	// In a real application, you might want to use a more sophisticated
	// CPU monitoring approach
	return float64(runtime.NumGoroutine()) / 100.0
}

// GetDetailedStats returns detailed statistics for dashboard
func (sm *StatsManager) GetDetailedStats() map[string]interface{} {
	stats := sm.GetStats()
	pluginStats := sm.GetPluginStats()
	topPlugins := sm.GetTopPlugins(5)

	return map[string]interface{}{
		"system": map[string]interface{}{
			"uptime":             sm.GetUptimeString(),
			"total_requests":     stats.TotalRequests,
			"total_executions":   stats.TotalExecutions,
			"total_errors":       stats.TotalErrors,
			"error_rate":         sm.GetErrorRate(),
			"requests_per_sec":   sm.GetRequestsPerSecond(),
			"executions_per_sec": sm.GetExecutionsPerSecond(),
			"memory_usage_mb":    float64(stats.MemoryUsage) / 1024 / 1024,
			"goroutines":         stats.Goroutines,
			"cpu_usage":          stats.CPUUsage,
		},
		"plugins":     pluginStats,
		"top_plugins": topPlugins,
		"timestamp":   time.Now(),
	}
}

// Persistence Methods

// LoadStats loads statistics from persistent storage
func (sm *StatsManager) LoadStats() error {
	if !sm.persistEnabled || sm.storage == nil {
		return nil
	}

	data := sm.storage.GetData()
	if data == nil {
		return fmt.Errorf("no data available in storage")
	}

	sm.mutex.Lock()
	defer sm.mutex.Unlock()

	// Restore start time if available
	if !data.StartTime.IsZero() {
		sm.startTime = data.StartTime
	}

	// Restore counters
	sm.totalRequests = data.TotalRequests
	sm.totalExecutions = data.TotalExecutions
	sm.totalErrors = data.TotalErrors

	// Restore plugin statistics
	sm.pluginStats = make(map[string]*ExecutionStats)
	for key, stats := range data.PluginStats {
		statsCopy := *stats
		sm.pluginStats[key] = &statsCopy
	}

	return nil
}

// SaveStats saves current statistics to persistent storage
func (sm *StatsManager) SaveStats() error {
	if !sm.persistEnabled || sm.storage == nil {
		return nil
	}

	sm.mutex.RLock()
	defer sm.mutex.RUnlock()

	data := &PluginStatsData{
		Version:         "1.0",
		LastSaved:       time.Now(),
		StartTime:       sm.startTime,
		TotalRequests:   sm.totalRequests,
		TotalExecutions: sm.totalExecutions,
		TotalErrors:     sm.totalErrors,
		PluginStats:     make(map[string]*ExecutionStats),
	}

	// Copy plugin statistics
	for key, stats := range sm.pluginStats {
		statsCopy := *stats
		data.PluginStats[key] = &statsCopy
	}

	return sm.storage.SaveStats(data)
}

// requestSave requests an asynchronous save operation
func (sm *StatsManager) requestSave() {
	if sm.persistEnabled && sm.storage != nil {
		sm.storage.RequestSave()
	}
}

// EnablePersistence enables or disables persistence
func (sm *StatsManager) EnablePersistence(enabled bool) {
	sm.persistEnabled = enabled
}

// IsPersistenceEnabled returns whether persistence is enabled
func (sm *StatsManager) IsPersistenceEnabled() bool {
	return sm.persistEnabled
}

// GetStorageInfo returns information about the storage
func (sm *StatsManager) GetStorageInfo() map[string]interface{} {
	if !sm.persistEnabled || sm.storage == nil {
		return map[string]interface{}{
			"enabled":    false,
			"file_path":  "",
			"last_saved": nil,
		}
	}

	data := sm.storage.GetData()
	return map[string]interface{}{
		"enabled":     true,
		"file_path":   sm.storage.GetFilePath(),
		"backup_path": sm.storage.GetBackupPath(),
		"last_saved":  data.LastSaved,
		"version":     data.Version,
	}
}

// ForceSave forces an immediate save operation
func (sm *StatsManager) ForceSave() error {
	if !sm.persistEnabled || sm.storage == nil {
		return fmt.Errorf("persistence is not enabled")
	}
	return sm.SaveStats()
}

// ResetStats resets all statistics (but preserves start time)
func (sm *StatsManager) ResetStats() {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()

	sm.pluginStats = make(map[string]*ExecutionStats)
	sm.totalRequests = 0
	sm.totalExecutions = 0
	sm.totalErrors = 0

	// Request save to persist the reset
	sm.requestSave()
}

// Close gracefully shuts down the stats manager
func (sm *StatsManager) Close() error {
	if sm.persistEnabled && sm.storage != nil {
		// Force a final save
		if err := sm.SaveStats(); err != nil {
			return fmt.Errorf("failed to save stats during shutdown: %w", err)
		}

		// Close the storage
		return sm.storage.Close()
	}
	return nil
}
