package main

import (
	"fmt"

	"github.com/loonghao/webhook_bridge/internal/web"
)

// Example demonstrating the new plugin logging features
func main() {
	// Create a new log manager
	logManager := web.NewLogManager()

	// Add some sample plugin logs
	sampleLogs := []web.LogEntry{
		{
			Level:      "INFO",
			Source:     "plugin",
			Message:    "GitHub webhook plugin executed successfully",
			PluginName: "github_webhook",
			Data: map[string]any{
				"execution_time": "45ms",
				"status":         "success",
				"repository":     "webhook_bridge",
			},
		},
		{
			Level:      "ERROR",
			Source:     "plugin",
			Message:    "Slack notification plugin failed",
			PluginName: "slack_notification",
			Data: map[string]any{
				"execution_time": "120ms",
				"error":          "connection timeout",
				"retry_count":    3,
			},
		},
		{
			Level:      "WARN",
			Source:     "plugin",
			Message:    "GitHub webhook plugin rate limit warning",
			PluginName: "github_webhook",
			Data: map[string]any{
				"execution_time": "30ms",
				"rate_limit":     "approaching",
				"remaining":      10,
			},
		},
		{
			Level:      "INFO",
			Source:     "system",
			Message:    "System health check completed",
			PluginName: "", // No plugin associated
			Data: map[string]any{
				"memory_usage": "65%",
				"cpu_usage":    "23%",
			},
		},
	}

	// Add logs to the manager
	for _, log := range sampleLogs {
		logManager.AddLog(log)
	}

	fmt.Println("=== Plugin Logging Features Demo ===")
	fmt.Println()

	// 1. Get all logs
	fmt.Println("1. All logs:")
	allLogs := logManager.GetLogs("", 0)
	for _, log := range allLogs {
		fmt.Printf("   [%s] %s: %s (Plugin: %s)\n",
			log.Level, log.Source, log.Message, log.PluginName)
	}

	// 2. Filter by plugin name
	fmt.Println("\n2. GitHub webhook plugin logs:")
	githubLogs := logManager.GetLogsByPlugin("github_webhook", 0)
	for _, log := range githubLogs {
		fmt.Printf("   [%s] %s: %s\n", log.Level, log.Source, log.Message)
	}

	// 3. Filter by level and plugin
	fmt.Println("\n3. ERROR logs for slack_notification plugin:")
	errorLogs := logManager.GetLogsWithFilters("ERROR", "slack_notification", 0)
	for _, log := range errorLogs {
		fmt.Printf("   [%s] %s: %s\n", log.Level, log.Source, log.Message)
	}

	// 4. Get available plugins
	fmt.Println("\n4. Available plugins:")
	plugins := logManager.GetAvailablePlugins()
	for _, plugin := range plugins {
		fmt.Printf("   - %s\n", plugin)
	}

	// 5. Get log statistics
	fmt.Println("\n5. Log statistics:")
	stats := logManager.GetLogStats()
	fmt.Printf("   Total logs: %v\n", stats["total"])

	if levels, ok := stats["levels"].(map[string]int); ok {
		fmt.Println("   By level:")
		for level, count := range levels {
			fmt.Printf("     %s: %d\n", level, count)
		}
	}

	if plugins, ok := stats["plugins"].(map[string]int); ok {
		fmt.Println("   By plugin:")
		for plugin, count := range plugins {
			fmt.Printf("     %s: %d\n", plugin, count)
		}
	}

	fmt.Println("\n=== Demo completed ===")
}
