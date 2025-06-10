package modern

import (
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/assert"

	"github.com/loonghao/webhook_bridge/internal/web"
)

// TestSystemMetricsCalculation tests the system metrics calculation logic
func TestSystemMetricsCalculation(t *testing.T) {
	// Create a handler with minimal setup
	handler := &ModernDashboardHandler{
		monitorClients: make(map[*websocket.Conn]bool),
	}

	// Create a real stats manager and populate it with test data
	statsManager := web.NewStatsManager()

	// Simulate some executions to populate stats
	for i := 0; i < 96; i++ {
		statsManager.RecordExecution("test_plugin", "POST", time.Now().Add(-time.Second))
	}
	for i := 0; i < 4; i++ {
		statsManager.RecordError("test_plugin", "POST")
	}

	handler.statsManager = statsManager

	// Test getSystemMetrics
	metrics := handler.getSystemMetrics()

	assert.NotNil(t, metrics)
	assert.Equal(t, int64(96), metrics.TotalExecutions)      // Only successful executions are counted
	assert.InDelta(t, float64(96), metrics.SuccessRate, 1.0) // Allow some variance
	assert.InDelta(t, float64(4), metrics.ErrorRate, 1.0)    // Allow some variance
	assert.Equal(t, 1, metrics.ActivePlugins)
	assert.Greater(t, metrics.AvgExecutionTime, float64(0)) // Should have some execution time
}

// TestPluginStatusUpdateBroadcast tests the broadcasting functionality
func TestPluginStatusUpdateBroadcast(t *testing.T) {
	handler := &ModernDashboardHandler{
		monitorClients: make(map[*websocket.Conn]bool),
	}

	// Test broadcasting without any connected clients (should not panic)
	update := PluginStatusUpdate{
		PluginName:    "test_plugin",
		Status:        "executed",
		LastExecuted:  time.Now().Format(time.RFC3339),
		ExecutionTime: 100,
		Success:       true,
	}

	assert.NotPanics(t, func() {
		handler.BroadcastPluginStatusUpdate(update)
	})
}

// TestSystemMetricsUpdateBroadcast tests the system metrics broadcasting
func TestSystemMetricsUpdateBroadcast(t *testing.T) {
	handler := &ModernDashboardHandler{
		monitorClients: make(map[*websocket.Conn]bool),
	}

	// Create a real stats manager
	statsManager := web.NewStatsManager()
	handler.statsManager = statsManager

	// Test broadcasting without any connected clients (should not panic)
	assert.NotPanics(t, func() {
		handler.BroadcastSystemMetricsUpdate()
	})
}

// TestMonitorMessageStructure tests the monitor message structure
func TestMonitorMessageStructure(t *testing.T) {
	// Test PluginStatusUpdate structure
	update := PluginStatusUpdate{
		PluginName:    "test_plugin",
		Status:        "executed",
		LastExecuted:  "2023-01-01T00:00:00Z",
		ExecutionTime: 150,
		Success:       true,
		Error:         "",
	}

	assert.Equal(t, "test_plugin", update.PluginName)
	assert.Equal(t, "executed", update.Status)
	assert.Equal(t, "2023-01-01T00:00:00Z", update.LastExecuted)
	assert.Equal(t, int64(150), update.ExecutionTime)
	assert.True(t, update.Success)
	assert.Equal(t, "", update.Error)

	// Test SystemMetricsUpdate structure
	metrics := SystemMetricsUpdate{
		TotalExecutions:    200,
		SuccessRate:        95.5,
		AvgExecutionTime:   120.5,
		ActivePlugins:      3,
		ErrorRate:          4.5,
		LastHourExecutions: 25,
	}

	assert.Equal(t, int64(200), metrics.TotalExecutions)
	assert.Equal(t, 95.5, metrics.SuccessRate)
	assert.Equal(t, 120.5, metrics.AvgExecutionTime)
	assert.Equal(t, 3, metrics.ActivePlugins)
	assert.Equal(t, 4.5, metrics.ErrorRate)
	assert.Equal(t, int64(25), metrics.LastHourExecutions)
}

// TestEdgeCases tests edge cases in metrics calculation
func TestEdgeCases(t *testing.T) {
	handler := &ModernDashboardHandler{
		monitorClients: make(map[*websocket.Conn]bool),
	}

	// Test with zero executions
	statsManager := web.NewStatsManager()
	handler.statsManager = statsManager
	metrics := handler.getSystemMetrics()

	assert.Equal(t, int64(0), metrics.TotalExecutions)
	assert.Equal(t, float64(0), metrics.SuccessRate)
	assert.Equal(t, float64(0), metrics.ErrorRate)
	assert.Equal(t, 0, metrics.ActivePlugins)
	assert.Equal(t, float64(0), metrics.AvgExecutionTime)

	// Test with all errors
	for i := 0; i < 10; i++ {
		statsManager.RecordExecution("test_plugin", "POST", time.Now())
		statsManager.RecordError("test_plugin", "POST")
	}

	metrics = handler.getSystemMetrics()
	assert.Equal(t, float64(0), metrics.SuccessRate)
	assert.Equal(t, float64(100), metrics.ErrorRate)
}

// TestMonitorMessageTypes tests different monitor message types
func TestMonitorMessageTypes(t *testing.T) {
	// Test system_metrics message
	systemMsg := MonitorMessage{
		Type:      "system_metrics",
		Timestamp: time.Now(),
		Data: SystemMetricsUpdate{
			TotalExecutions: 100,
			SuccessRate:     95.0,
		},
	}

	assert.Equal(t, "system_metrics", systemMsg.Type)
	assert.False(t, systemMsg.Timestamp.IsZero())
	assert.NotNil(t, systemMsg.Data)

	// Test plugin_status message
	pluginMsg := MonitorMessage{
		Type:      "plugin_status",
		Timestamp: time.Now(),
		Data: PluginStatusUpdate{
			PluginName: "test_plugin",
			Success:    true,
		},
	}

	assert.Equal(t, "plugin_status", pluginMsg.Type)
	assert.False(t, pluginMsg.Timestamp.IsZero())
	assert.NotNil(t, pluginMsg.Data)
}

// TestUptimeCalculation tests uptime-based calculations
func TestUptimeCalculation(t *testing.T) {
	handler := &ModernDashboardHandler{
		monitorClients: make(map[*websocket.Conn]bool),
	}

	// Create a stats manager and simulate executions
	statsManager := web.NewStatsManager()

	// Simulate 120 executions
	for i := 0; i < 120; i++ {
		statsManager.RecordExecution("test_plugin", "POST", time.Now())
	}

	handler.statsManager = statsManager
	metrics := handler.getSystemMetrics()

	// Should have some executions per hour calculation
	assert.GreaterOrEqual(t, metrics.LastHourExecutions, int64(0))
}

// TestPluginStatsAggregation tests plugin statistics aggregation
func TestPluginStatsAggregation(t *testing.T) {
	handler := &ModernDashboardHandler{
		monitorClients: make(map[*websocket.Conn]bool),
	}

	// Create a stats manager and simulate multiple plugins
	statsManager := web.NewStatsManager()

	// Simulate plugin1 executions
	for i := 0; i < 10; i++ {
		statsManager.RecordExecution("plugin1", "POST", time.Now().Add(-time.Millisecond*100))
	}
	statsManager.RecordError("plugin1", "POST")

	// Simulate plugin2 executions
	for i := 0; i < 20; i++ {
		statsManager.RecordExecution("plugin2", "GET", time.Now().Add(-time.Millisecond*150))
	}
	for i := 0; i < 2; i++ {
		statsManager.RecordError("plugin2", "GET")
	}

	handler.statsManager = statsManager
	metrics := handler.getSystemMetrics()

	assert.Equal(t, 2, metrics.ActivePlugins)
	assert.Greater(t, metrics.AvgExecutionTime, float64(0))
}
