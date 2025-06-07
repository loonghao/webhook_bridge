package main

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/loonghao/webhook_bridge/internal/web"
)

// Example demonstrating the plugin statistics persistence features
func main() {
	fmt.Println("=== Plugin Statistics Persistence Demo ===\n")

	// Create a temporary data directory for this demo
	dataDir := filepath.Join(os.TempDir(), "webhook_bridge_demo")
	defer os.RemoveAll(dataDir)

	fmt.Printf("Using data directory: %s\n\n", dataDir)

	// 1. Create a stats manager with persistence
	fmt.Println("1. Creating StatsManager with persistence...")
	sm := web.NewStatsManagerWithPersistence(dataDir)
	defer sm.Close()

	// Check if persistence is enabled
	fmt.Printf("   Persistence enabled: %v\n", sm.IsPersistenceEnabled())

	// Get storage info
	storageInfo := sm.GetStorageInfo()
	fmt.Printf("   Storage file: %s\n", storageInfo["file_path"])
	fmt.Printf("   Backup file: %s\n", storageInfo["backup_path"])

	// 2. Record some plugin executions
	fmt.Println("\n2. Recording plugin executions...")

	plugins := []struct {
		name   string
		method string
		delay  time.Duration
		errors int
	}{
		{"github_webhook", "POST", 45 * time.Millisecond, 0},
		{"slack_notification", "POST", 120 * time.Millisecond, 1},
		{"email_sender", "POST", 80 * time.Millisecond, 0},
		{"github_webhook", "GET", 30 * time.Millisecond, 0},
		{"database_logger", "POST", 15 * time.Millisecond, 0},
	}

	for i, plugin := range plugins {
		startTime := time.Now().Add(-plugin.delay)

		// Record execution
		sm.RecordExecution(plugin.name, plugin.method, startTime)

		// Record errors if any
		for j := 0; j < plugin.errors; j++ {
			sm.RecordError(plugin.name, plugin.method)
		}

		// Record HTTP requests
		sm.RecordRequest()

		fmt.Printf("   Recorded: %s:%s (execution %d)\n", plugin.name, plugin.method, i+1)
	}

	// 3. Display current statistics
	fmt.Println("\n3. Current statistics:")
	stats := sm.GetStats()
	fmt.Printf("   Total requests: %d\n", stats.TotalRequests)
	fmt.Printf("   Total executions: %d\n", stats.TotalExecutions)
	fmt.Printf("   Total errors: %d\n", stats.TotalErrors)
	fmt.Printf("   Error rate: %.2f%%\n", sm.GetErrorRate())

	// Display plugin statistics
	fmt.Println("\n   Plugin statistics:")
	pluginStats := sm.GetPluginStats()
	for _, stat := range pluginStats {
		fmt.Printf("     %s:%s - Count: %d, Errors: %d, Avg Time: %v\n",
			stat.Plugin, stat.Method, stat.Count, stat.Errors, stat.AvgTime)
	}

	// 4. Force save to demonstrate persistence
	fmt.Println("\n4. Forcing save to disk...")
	err := sm.ForceSave()
	if err != nil {
		fmt.Printf("   Error saving: %v\n", err)
	} else {
		fmt.Println("   Successfully saved to disk")
	}

	// Show updated storage info
	storageInfo = sm.GetStorageInfo()
	if lastSaved, ok := storageInfo["last_saved"].(time.Time); ok {
		fmt.Printf("   Last saved: %s\n", lastSaved.Format("2006-01-02 15:04:05"))
	}

	// 5. Close the first stats manager
	fmt.Println("\n5. Closing first StatsManager...")
	sm.Close()

	// 6. Create a new stats manager to demonstrate data recovery
	fmt.Println("\n6. Creating new StatsManager to test data recovery...")
	sm2 := web.NewStatsManagerWithPersistence(dataDir)
	defer sm2.Close()

	// Verify data was loaded
	stats2 := sm2.GetStats()
	fmt.Printf("   Loaded - Total requests: %d\n", stats2.TotalRequests)
	fmt.Printf("   Loaded - Total executions: %d\n", stats2.TotalExecutions)
	fmt.Printf("   Loaded - Total errors: %d\n", stats2.TotalErrors)

	// Verify plugin statistics were loaded
	pluginStats2 := sm2.GetPluginStats()
	fmt.Printf("   Loaded - Plugin stats count: %d\n", len(pluginStats2))

	// 7. Add more data to demonstrate incremental updates
	fmt.Println("\n7. Adding more data to demonstrate incremental updates...")
	startTime := time.Now().Add(-50 * time.Millisecond)
	sm2.RecordExecution("new_plugin", "PUT", startTime)
	sm2.RecordRequest()

	// Show updated statistics
	stats3 := sm2.GetStats()
	fmt.Printf("   Updated - Total requests: %d\n", stats3.TotalRequests)
	fmt.Printf("   Updated - Total executions: %d\n", stats3.TotalExecutions)

	// 8. Demonstrate reset functionality
	fmt.Println("\n8. Demonstrating reset functionality...")
	fmt.Printf("   Before reset - Total executions: %d\n", sm2.GetStats().TotalExecutions)

	sm2.ResetStats()

	fmt.Printf("   After reset - Total executions: %d\n", sm2.GetStats().TotalExecutions)
	fmt.Printf("   After reset - Plugin stats count: %d\n", len(sm2.GetPluginStats()))

	// 9. Show final storage information
	fmt.Println("\n9. Final storage information:")
	finalInfo := sm2.GetStorageInfo()
	for key, value := range finalInfo {
		if key == "last_saved" {
			if lastSaved, ok := value.(time.Time); ok {
				fmt.Printf("   %s: %s\n", key, lastSaved.Format("2006-01-02 15:04:05"))
			}
		} else {
			fmt.Printf("   %s: %v\n", key, value)
		}
	}

	fmt.Println("\n=== Demo completed ===")
	fmt.Println("\nKey features demonstrated:")
	fmt.Println("✓ Automatic persistence of plugin statistics")
	fmt.Println("✓ Data recovery after restart")
	fmt.Println("✓ Asynchronous saving with manual force save")
	fmt.Println("✓ Backup and recovery mechanisms")
	fmt.Println("✓ Statistics reset functionality")
	fmt.Println("✓ Storage information and monitoring")
}
