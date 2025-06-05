package web

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// TestNewStatsManager tests the creation of a new stats manager
func TestNewStatsManager(t *testing.T) {
	sm := NewStatsManager()

	assert.NotNil(t, sm)
	assert.NotNil(t, sm.pluginStats)
	assert.False(t, sm.persistEnabled)
	assert.WithinDuration(t, time.Now(), sm.startTime, time.Second)
}

// TestRecordExecution tests recording plugin executions
func TestRecordExecution(t *testing.T) {
	sm := NewStatsManager()

	startTime := time.Now().Add(-100 * time.Millisecond)
	sm.RecordExecution("test_plugin", "POST", startTime)

	stats := sm.GetStats()
	assert.Equal(t, int64(1), stats.TotalExecutions)

	pluginStats := sm.GetPluginStats()
	assert.Len(t, pluginStats, 1)

	key := "test_plugin:POST"
	assert.Contains(t, pluginStats, key)
	assert.Equal(t, int64(1), pluginStats[key].Count)
	assert.Equal(t, "test_plugin", pluginStats[key].Plugin)
	assert.Equal(t, "POST", pluginStats[key].Method)
}

// TestRecordError tests recording execution errors
func TestRecordError(t *testing.T) {
	sm := NewStatsManager()

	// First record an execution
	sm.RecordExecution("test_plugin", "POST", time.Now())

	// Then record an error
	sm.RecordError("test_plugin", "POST")

	stats := sm.GetStats()
	assert.Equal(t, int64(1), stats.TotalErrors)

	pluginStats := sm.GetPluginStats()
	key := "test_plugin:POST"
	assert.Equal(t, int64(1), pluginStats[key].Errors)
}

// TestRecordRequest tests recording HTTP requests
func TestRecordRequest(t *testing.T) {
	sm := NewStatsManager()

	sm.RecordRequest()
	sm.RecordRequest()

	stats := sm.GetStats()
	assert.Equal(t, int64(2), stats.TotalRequests)
}

// TestGetStats tests getting system statistics
func TestGetStats(t *testing.T) {
	sm := NewStatsManager()

	// Record some data
	sm.RecordExecution("plugin1", "GET", time.Now())
	sm.RecordExecution("plugin2", "POST", time.Now())
	sm.RecordError("plugin1", "GET")
	sm.RecordRequest()

	stats := sm.GetStats()

	assert.Equal(t, int64(2), stats.TotalExecutions)
	assert.Equal(t, int64(1), stats.TotalErrors)
	assert.Equal(t, int64(1), stats.TotalRequests)
	assert.Greater(t, stats.MemoryUsage, uint64(0))
	assert.Greater(t, stats.Goroutines, 0)
	assert.GreaterOrEqual(t, stats.CPUUsage, float64(0))
}

// TestGetPluginStats tests getting plugin-specific statistics
func TestGetPluginStats(t *testing.T) {
	sm := NewStatsManager()

	// Record executions for different plugins
	sm.RecordExecution("plugin1", "GET", time.Now().Add(-50*time.Millisecond))
	sm.RecordExecution("plugin1", "POST", time.Now().Add(-30*time.Millisecond))
	sm.RecordExecution("plugin2", "GET", time.Now().Add(-20*time.Millisecond))

	pluginStats := sm.GetPluginStats()

	assert.Len(t, pluginStats, 3)
	assert.Contains(t, pluginStats, "plugin1:GET")
	assert.Contains(t, pluginStats, "plugin1:POST")
	assert.Contains(t, pluginStats, "plugin2:GET")

	// Check individual stats
	plugin1Get := pluginStats["plugin1:GET"]
	assert.Equal(t, "plugin1", plugin1Get.Plugin)
	assert.Equal(t, "GET", plugin1Get.Method)
	assert.Equal(t, int64(1), plugin1Get.Count)
	assert.Greater(t, plugin1Get.TotalTime, time.Duration(0))
}

// TestGetTopPlugins tests getting top plugins by execution count
func TestGetTopPlugins(t *testing.T) {
	sm := NewStatsManager()

	// Record different numbers of executions
	for i := 0; i < 5; i++ {
		sm.RecordExecution("plugin1", "GET", time.Now())
	}
	for i := 0; i < 3; i++ {
		sm.RecordExecution("plugin2", "POST", time.Now())
	}
	for i := 0; i < 7; i++ {
		sm.RecordExecution("plugin3", "PUT", time.Now())
	}

	topPlugins := sm.GetTopPlugins(2)

	assert.Len(t, topPlugins, 2)
	// Should be sorted by count (descending)
	assert.Equal(t, int64(7), topPlugins[0].Count) // plugin3
	assert.Equal(t, int64(5), topPlugins[1].Count) // plugin1
}

// TestGetErrorRate tests error rate calculation
func TestGetErrorRate(t *testing.T) {
	sm := NewStatsManager()

	// Test with no executions
	assert.Equal(t, float64(0), sm.GetErrorRate())

	// Record 10 executions and 2 errors
	for i := 0; i < 10; i++ {
		sm.RecordExecution("test_plugin", "GET", time.Now())
	}
	sm.RecordError("test_plugin", "GET")
	sm.RecordError("test_plugin", "GET")

	errorRate := sm.GetErrorRate()
	assert.Equal(t, float64(20), errorRate) // 2/10 * 100 = 20%
}

// TestGetRequestsPerSecond tests requests per second calculation
func TestGetRequestsPerSecond(t *testing.T) {
	sm := NewStatsManager()

	// Record some requests
	for i := 0; i < 5; i++ {
		sm.RecordRequest()
	}

	// Wait a bit to ensure uptime > 0
	time.Sleep(10 * time.Millisecond)

	rps := sm.GetRequestsPerSecond()
	assert.Greater(t, rps, float64(0))
}

// TestGetExecutionsPerSecond tests executions per second calculation
func TestGetExecutionsPerSecond(t *testing.T) {
	sm := NewStatsManager()

	// Record some executions
	for i := 0; i < 3; i++ {
		sm.RecordExecution("test_plugin", "GET", time.Now())
	}

	// Wait a bit to ensure uptime > 0
	time.Sleep(10 * time.Millisecond)

	eps := sm.GetExecutionsPerSecond()
	assert.Greater(t, eps, float64(0))
}

// TestGetUptime tests uptime calculation
func TestGetUptime(t *testing.T) {
	sm := NewStatsManager()

	// Wait a bit
	time.Sleep(10 * time.Millisecond)

	uptime := sm.GetUptime()
	assert.Greater(t, uptime, time.Duration(0))
	assert.Less(t, uptime, time.Second) // Should be less than a second
}

// TestGetUptimeString tests uptime string formatting
func TestGetUptimeString(t *testing.T) {
	sm := NewStatsManager()

	// Wait a bit
	time.Sleep(10 * time.Millisecond)

	uptimeStr := sm.GetUptimeString()
	assert.NotEmpty(t, uptimeStr)
	assert.Contains(t, uptimeStr, "s") // Should contain seconds
}

// TestReset tests resetting statistics
func TestReset(t *testing.T) {
	sm := NewStatsManager()

	// Record some data
	sm.RecordExecution("test_plugin", "GET", time.Now())
	sm.RecordError("test_plugin", "GET")
	sm.RecordRequest()

	// Verify data exists
	stats := sm.GetStats()
	assert.Greater(t, stats.TotalExecutions, int64(0))
	assert.Greater(t, stats.TotalErrors, int64(0))
	assert.Greater(t, stats.TotalRequests, int64(0))

	// Reset
	sm.Reset()

	// Verify data is cleared
	stats = sm.GetStats()
	assert.Equal(t, int64(0), stats.TotalExecutions)
	assert.Equal(t, int64(0), stats.TotalErrors)
	assert.Equal(t, int64(0), stats.TotalRequests)

	pluginStats := sm.GetPluginStats()
	assert.Len(t, pluginStats, 0)
}

// TestGetDetailedStats tests getting detailed statistics
func TestGetDetailedStats(t *testing.T) {
	sm := NewStatsManager()

	// Record some data
	sm.RecordExecution("test_plugin", "GET", time.Now())
	sm.RecordRequest()

	detailedStats := sm.GetDetailedStats()

	assert.Contains(t, detailedStats, "system")
	assert.Contains(t, detailedStats, "plugins")
	assert.Contains(t, detailedStats, "top_plugins")
	assert.Contains(t, detailedStats, "timestamp")

	systemStats := detailedStats["system"].(map[string]interface{})
	assert.Contains(t, systemStats, "uptime")
	assert.Contains(t, systemStats, "total_requests")
	assert.Contains(t, systemStats, "total_executions")
	assert.Contains(t, systemStats, "error_rate")
}

// TestConcurrentAccess tests concurrent access to stats manager
func TestConcurrentAccess(t *testing.T) {
	sm := NewStatsManager()

	done := make(chan bool, 20)

	// Start multiple goroutines doing different operations
	for i := 0; i < 5; i++ {
		go func() {
			defer func() { done <- true }()
			sm.RecordExecution("test_plugin", "GET", time.Now())
		}()

		go func() {
			defer func() { done <- true }()
			sm.RecordError("test_plugin", "GET")
		}()

		go func() {
			defer func() { done <- true }()
			sm.RecordRequest()
		}()

		go func() {
			defer func() { done <- true }()
			sm.GetStats()
		}()
	}

	// Wait for all goroutines to complete
	for i := 0; i < 20; i++ {
		<-done
	}

	// If we reach here without panic, the test passes
	assert.True(t, true)
}

// TestAverageTimeCalculation tests average execution time calculation
func TestAverageTimeCalculation(t *testing.T) {
	sm := NewStatsManager()

	// Record executions with known durations
	sm.RecordExecution("test_plugin", "GET", time.Now().Add(-100*time.Millisecond))
	sm.RecordExecution("test_plugin", "GET", time.Now().Add(-200*time.Millisecond))

	pluginStats := sm.GetPluginStats()
	stats := pluginStats["test_plugin:GET"]

	assert.Equal(t, int64(2), stats.Count)
	assert.Greater(t, stats.TotalTime, time.Duration(0))
	assert.Greater(t, stats.AvgTime, time.Duration(0))

	// Average should be TotalTime / Count
	expectedAvg := stats.TotalTime / time.Duration(stats.Count)
	assert.Equal(t, expectedAvg, stats.AvgTime)
}
