package execution

import (
	"sync"
	"time"
)

// TrackerMetrics manages in-memory metrics for execution tracking
type TrackerMetrics struct {
	mutex sync.RWMutex
	stats map[string]*PluginStats
}

// PluginStats represents statistics for a specific plugin
type PluginStats struct {
	TotalExecutions      int64            `json:"total_executions"`
	SuccessfulExecutions int64            `json:"successful_executions"`
	FailedExecutions     int64            `json:"failed_executions"`
	TotalDuration        time.Duration    `json:"total_duration"`
	MinDuration          time.Duration    `json:"min_duration"`
	MaxDuration          time.Duration    `json:"max_duration"`
	LastExecution        time.Time        `json:"last_execution"`
	ErrorTypes           map[string]int64 `json:"error_types"`
}

// NewTrackerMetrics creates a new tracker metrics instance
func NewTrackerMetrics() *TrackerMetrics {
	return &TrackerMetrics{
		stats: make(map[string]*PluginStats),
	}
}

// IncrementStarted increments the started execution count for a plugin
func (tm *TrackerMetrics) IncrementStarted(pluginName string) {
	tm.mutex.Lock()
	defer tm.mutex.Unlock()

	if tm.stats[pluginName] == nil {
		tm.stats[pluginName] = &PluginStats{
			ErrorTypes: make(map[string]int64),
		}
	}

	tm.stats[pluginName].TotalExecutions++
	tm.stats[pluginName].LastExecution = time.Now()
}

// IncrementCompleted increments the completed execution count for a plugin
func (tm *TrackerMetrics) IncrementCompleted(pluginName string) {
	tm.mutex.Lock()
	defer tm.mutex.Unlock()

	if tm.stats[pluginName] != nil {
		tm.stats[pluginName].SuccessfulExecutions++
	}
}

// IncrementFailed increments the failed execution count for a plugin
func (tm *TrackerMetrics) IncrementFailed(pluginName string) {
	tm.mutex.Lock()
	defer tm.mutex.Unlock()

	if tm.stats[pluginName] != nil {
		tm.stats[pluginName].FailedExecutions++
	}
}

// IncrementError increments the error count for a specific error type
func (tm *TrackerMetrics) IncrementError(pluginName, errorType string) {
	tm.mutex.Lock()
	defer tm.mutex.Unlock()

	if tm.stats[pluginName] != nil {
		if tm.stats[pluginName].ErrorTypes == nil {
			tm.stats[pluginName].ErrorTypes = make(map[string]int64)
		}
		tm.stats[pluginName].ErrorTypes[errorType]++
	}
}

// RecordDuration records the execution duration for a plugin
func (tm *TrackerMetrics) RecordDuration(pluginName string, duration time.Duration) {
	tm.mutex.Lock()
	defer tm.mutex.Unlock()

	if tm.stats[pluginName] != nil {
		stats := tm.stats[pluginName]
		stats.TotalDuration += duration

		if stats.MinDuration == 0 || duration < stats.MinDuration {
			stats.MinDuration = duration
		}

		if duration > stats.MaxDuration {
			stats.MaxDuration = duration
		}
	}
}

// GetStats returns a copy of the statistics for a specific plugin
func (tm *TrackerMetrics) GetStats(pluginName string) *PluginStats {
	tm.mutex.RLock()
	defer tm.mutex.RUnlock()

	if stats, exists := tm.stats[pluginName]; exists {
		// Return a copy to avoid concurrent access issues
		errorTypes := make(map[string]int64)
		for k, v := range stats.ErrorTypes {
			errorTypes[k] = v
		}

		return &PluginStats{
			TotalExecutions:      stats.TotalExecutions,
			SuccessfulExecutions: stats.SuccessfulExecutions,
			FailedExecutions:     stats.FailedExecutions,
			TotalDuration:        stats.TotalDuration,
			MinDuration:          stats.MinDuration,
			MaxDuration:          stats.MaxDuration,
			LastExecution:        stats.LastExecution,
			ErrorTypes:           errorTypes,
		}
	}

	return &PluginStats{
		ErrorTypes: make(map[string]int64),
	}
}

// GetAllStats returns a copy of all plugin statistics
func (tm *TrackerMetrics) GetAllStats() map[string]*PluginStats {
	tm.mutex.RLock()
	defer tm.mutex.RUnlock()

	result := make(map[string]*PluginStats)
	for pluginName, stats := range tm.stats {
		errorTypes := make(map[string]int64)
		for k, v := range stats.ErrorTypes {
			errorTypes[k] = v
		}

		result[pluginName] = &PluginStats{
			TotalExecutions:      stats.TotalExecutions,
			SuccessfulExecutions: stats.SuccessfulExecutions,
			FailedExecutions:     stats.FailedExecutions,
			TotalDuration:        stats.TotalDuration,
			MinDuration:          stats.MinDuration,
			MaxDuration:          stats.MaxDuration,
			LastExecution:        stats.LastExecution,
			ErrorTypes:           errorTypes,
		}
	}

	return result
}

// GetAverageExecutionTime returns the average execution time for a plugin
func (tm *TrackerMetrics) GetAverageExecutionTime(pluginName string) time.Duration {
	tm.mutex.RLock()
	defer tm.mutex.RUnlock()

	if stats, exists := tm.stats[pluginName]; exists && stats.TotalExecutions > 0 {
		return stats.TotalDuration / time.Duration(stats.TotalExecutions)
	}

	return 0
}

// GetSuccessRate returns the success rate for a plugin as a percentage
func (tm *TrackerMetrics) GetSuccessRate(pluginName string) float64 {
	tm.mutex.RLock()
	defer tm.mutex.RUnlock()

	if stats, exists := tm.stats[pluginName]; exists && stats.TotalExecutions > 0 {
		return float64(stats.SuccessfulExecutions) / float64(stats.TotalExecutions) * 100
	}

	return 0
}

// GetTopErrorTypes returns the most common error types across all plugins
func (tm *TrackerMetrics) GetTopErrorTypes(limit int) map[string]int64 {
	tm.mutex.RLock()
	defer tm.mutex.RUnlock()

	errorCounts := make(map[string]int64)

	// Aggregate error types across all plugins
	for _, stats := range tm.stats {
		for errorType, count := range stats.ErrorTypes {
			errorCounts[errorType] += count
		}
	}

	// If we need to limit results, we would sort and take top N
	// For now, return all error types
	return errorCounts
}

// Reset clears all metrics
func (tm *TrackerMetrics) Reset() {
	tm.mutex.Lock()
	defer tm.mutex.Unlock()

	tm.stats = make(map[string]*PluginStats)
}

// ResetPlugin clears metrics for a specific plugin
func (tm *TrackerMetrics) ResetPlugin(pluginName string) {
	tm.mutex.Lock()
	defer tm.mutex.Unlock()

	delete(tm.stats, pluginName)
}

// GetPluginCount returns the number of plugins with recorded metrics
func (tm *TrackerMetrics) GetPluginCount() int {
	tm.mutex.RLock()
	defer tm.mutex.RUnlock()

	return len(tm.stats)
}

// GetTotalExecutions returns the total number of executions across all plugins
func (tm *TrackerMetrics) GetTotalExecutions() int64 {
	tm.mutex.RLock()
	defer tm.mutex.RUnlock()

	var total int64
	for _, stats := range tm.stats {
		total += stats.TotalExecutions
	}

	return total
}

// GetTotalSuccessfulExecutions returns the total number of successful executions
func (tm *TrackerMetrics) GetTotalSuccessfulExecutions() int64 {
	tm.mutex.RLock()
	defer tm.mutex.RUnlock()

	var total int64
	for _, stats := range tm.stats {
		total += stats.SuccessfulExecutions
	}

	return total
}

// GetTotalFailedExecutions returns the total number of failed executions
func (tm *TrackerMetrics) GetTotalFailedExecutions() int64 {
	tm.mutex.RLock()
	defer tm.mutex.RUnlock()

	var total int64
	for _, stats := range tm.stats {
		total += stats.FailedExecutions
	}

	return total
}

// GetOverallSuccessRate returns the overall success rate across all plugins
func (tm *TrackerMetrics) GetOverallSuccessRate() float64 {
	totalExecutions := tm.GetTotalExecutions()
	if totalExecutions == 0 {
		return 0
	}

	successfulExecutions := tm.GetTotalSuccessfulExecutions()
	return float64(successfulExecutions) / float64(totalExecutions) * 100
}
