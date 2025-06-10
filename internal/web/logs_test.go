package web

import (
	"testing"
	"time"
)

func TestLogEntryWithPluginName(t *testing.T) {
	// Test that LogEntry can handle plugin names
	entry := LogEntry{
		ID:         1,
		Timestamp:  time.Now(),
		Level:      "INFO",
		Source:     "plugin",
		Message:    "Plugin executed successfully",
		PluginName: "test_plugin",
		Data: map[string]interface{}{
			"execution_time": "45ms",
		},
	}

	// Verify all fields are set correctly
	if entry.ID != 1 {
		t.Errorf("Expected ID 1, got %d", entry.ID)
	}
	if entry.Level != "INFO" {
		t.Errorf("Expected level 'INFO', got '%s'", entry.Level)
	}
	if entry.Source != "plugin" {
		t.Errorf("Expected source 'plugin', got '%s'", entry.Source)
	}
	if entry.Message != "Plugin executed successfully" {
		t.Errorf("Expected message 'Plugin executed successfully', got '%s'", entry.Message)
	}
	if entry.PluginName != "test_plugin" {
		t.Errorf("Expected plugin name 'test_plugin', got '%s'", entry.PluginName)
	}
	if entry.Timestamp.IsZero() {
		t.Error("Expected timestamp to be set")
	}
	if entry.Data["execution_time"] != "45ms" {
		t.Errorf("Expected execution_time '45ms', got '%v'", entry.Data["execution_time"])
	}
}

func TestLogManagerPluginFiltering(t *testing.T) {
	lm := NewLogManager()

	// Add test logs with different plugins
	testLogs := []LogEntry{
		{
			Level:      "INFO",
			Source:     "plugin",
			Message:    "Plugin A executed",
			PluginName: "plugin_a",
		},
		{
			Level:      "ERROR",
			Source:     "plugin",
			Message:    "Plugin B failed",
			PluginName: "plugin_b",
		},
		{
			Level:      "INFO",
			Source:     "system",
			Message:    "System message",
			PluginName: "", // No plugin
		},
		{
			Level:      "WARN",
			Source:     "plugin",
			Message:    "Plugin A warning",
			PluginName: "plugin_a",
		},
	}

	for _, log := range testLogs {
		lm.AddLog(log)
	}

	// Test filtering by plugin name
	pluginALogs := lm.GetLogsByPlugin("plugin_a", 0)
	if len(pluginALogs) != 2 {
		t.Errorf("Expected 2 logs for plugin_a, got %d", len(pluginALogs))
	}

	pluginBLogs := lm.GetLogsByPlugin("plugin_b", 0)
	if len(pluginBLogs) != 1 {
		t.Errorf("Expected 1 log for plugin_b, got %d", len(pluginBLogs))
	}

	// Test filtering by level and plugin
	errorLogs := lm.GetLogsWithFilters("ERROR", "plugin_b", 0)
	if len(errorLogs) != 1 {
		t.Errorf("Expected 1 ERROR log for plugin_b, got %d", len(errorLogs))
	}

	// Test getting available plugins
	plugins := lm.GetAvailablePlugins()
	expectedPlugins := map[string]bool{
		"plugin_a": true,
		"plugin_b": true,
	}

	if len(plugins) != 2 {
		t.Errorf("Expected 2 plugins, got %d", len(plugins))
	}

	for _, plugin := range plugins {
		if !expectedPlugins[plugin] {
			t.Errorf("Unexpected plugin: %s", plugin)
		}
	}
}

func TestLogStatsWithPlugins(t *testing.T) {
	lm := NewLogManager()

	// Add test logs
	testLogs := []LogEntry{
		{Level: "INFO", Source: "plugin", PluginName: "plugin_a"},
		{Level: "ERROR", Source: "plugin", PluginName: "plugin_a"},
		{Level: "INFO", Source: "plugin", PluginName: "plugin_b"},
		{Level: "INFO", Source: "system", PluginName: ""},
	}

	for _, log := range testLogs {
		lm.AddLog(log)
	}

	stats := lm.GetLogStats()

	// Check total
	if stats["total"] != 4 {
		t.Errorf("Expected total 4, got %v", stats["total"])
	}

	// Check plugin stats
	plugins := stats["plugins"].(map[string]int)
	if plugins["plugin_a"] != 2 {
		t.Errorf("Expected 2 logs for plugin_a, got %d", plugins["plugin_a"])
	}
	if plugins["plugin_b"] != 1 {
		t.Errorf("Expected 1 log for plugin_b, got %d", plugins["plugin_b"])
	}
}
