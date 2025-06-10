package main

import (
	"context"
	"fmt"
	"time"

	"github.com/loonghao/webhook_bridge/api/proto"
	"github.com/loonghao/webhook_bridge/internal/config"
	"github.com/loonghao/webhook_bridge/internal/grpc"
	"github.com/loonghao/webhook_bridge/internal/web"
)

// Example demonstrating enhanced plugin execution with automatic logging and statistics
func main() {
	fmt.Println("=== Enhanced Plugin Execution Demo ===")
	fmt.Println()

	// Create configuration
	cfg := &config.ExecutorConfig{
		Host:    "localhost",
		Port:    50051,
		Timeout: 30,
	}

	// Create data directory for persistence
	dataDir := "demo_data"

	// Create log manager
	logManager := web.NewPersistentLogManager("demo_logs", 100)
	defer logManager.Close()

	// Create stats manager with persistence
	statsManager := web.NewStatsManagerWithPersistence(dataDir)
	defer statsManager.Close()

	// Create gRPC client
	grpcClient := grpc.NewClient(cfg)

	// Setup enhanced logging and statistics
	grpc.SetupClientWithLoggingAndStats(grpcClient, logManager, statsManager)
	fmt.Println("✅ Enhanced gRPC client with logging and statistics")

	// Try to connect (this will likely fail since no Python executor is running)
	fmt.Println("\n1. Attempting to connect to Python executor...")
	if err := grpcClient.Connect(); err != nil {
		fmt.Printf("⚠️  Connection failed (expected): %v\n", err)
		fmt.Println("   This is normal if no Python executor is running")
	} else {
		fmt.Println("✅ Connected to Python executor")
	}

	// Demonstrate plugin execution attempts (will show logging even on failure)
	fmt.Println("\n2. Demonstrating plugin execution with automatic logging...")

	testCases := []struct {
		pluginName string
		method     string
		data       map[string]string
	}{
		{
			pluginName: "github_webhook",
			method:     "POST",
			data: map[string]string{
				"repository": "webhook_bridge",
				"action":     "push",
				"branch":     "main",
			},
		},
		{
			pluginName: "slack_notification",
			method:     "POST",
			data: map[string]string{
				"channel": "#general",
				"message": "Deployment completed successfully",
			},
		},
	}

	for i, testCase := range testCases {
		fmt.Printf("\n   Test %d: Executing %s:%s\n", i+1, testCase.pluginName, testCase.method)

		// Create execution request
		req := &proto.ExecutePluginRequest{
			PluginName: testCase.pluginName,
			HttpMethod: testCase.method,
			Data:       testCase.data,
		}

		// Execute plugin (this will automatically log and record statistics)
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		resp, err := grpcClient.ExecutePlugin(ctx, req)
		cancel()

		if err != nil {
			fmt.Printf("     ❌ Execution failed: %v\n", err)
		} else {
			fmt.Printf("     ✅ Execution completed: status=%d, message=%s\n", resp.StatusCode, resp.Message)
		}

		// Small delay between executions
		time.Sleep(100 * time.Millisecond)
	}

	// Wait a moment for async logging to complete
	time.Sleep(500 * time.Millisecond)

	// Display collected statistics
	fmt.Println("\n3. Collected Statistics:")
	stats := statsManager.GetStats()
	fmt.Printf("   Total Requests: %d\n", stats.TotalRequests)
	fmt.Printf("   Total Executions: %d\n", stats.TotalExecutions)
	fmt.Printf("   Total Errors: %d\n", stats.TotalErrors)
	fmt.Printf("   Error Rate: %.2f%%\n", statsManager.GetErrorRate())

	// Display plugin-specific statistics
	pluginStats := statsManager.GetPluginStats()
	if len(pluginStats) > 0 {
		fmt.Println("\n   Plugin Statistics:")
		for _, stat := range pluginStats {
			fmt.Printf("     %s:%s - Count: %d, Errors: %d, Avg Time: %v\n",
				stat.Plugin, stat.Method, stat.Count, stat.Errors, stat.AvgTime)
		}
	}

	// Display recent logs
	fmt.Println("\n4. Recent Logs:")
	logs := logManager.GetLogs("", 10) // Get last 10 logs
	for _, logEntry := range logs {
		fmt.Printf("   [%s] %s: %s\n",
			logEntry.Timestamp.Format("15:04:05"),
			logEntry.Level,
			logEntry.Message)
		if logEntry.PluginName != "" {
			fmt.Printf("     Plugin: %s\n", logEntry.PluginName)
		}
	}

	// Display plugin-specific logs
	fmt.Println("\n5. Plugin-specific Logs (github_webhook):")
	githubLogs := logManager.GetLogsWithFilters("", "github_webhook", 5)
	for _, logEntry := range githubLogs {
		fmt.Printf("   [%s] %s: %s\n",
			logEntry.Timestamp.Format("15:04:05"),
			logEntry.Level,
			logEntry.Message)
	}

	// Show storage information
	fmt.Println("\n6. Storage Information:")
	storageInfo := statsManager.GetStorageInfo()
	for key, value := range storageInfo {
		if key == "last_saved" {
			if lastSaved, ok := value.(time.Time); ok {
				fmt.Printf("   %s: %s\n", key, lastSaved.Format("2006-01-02 15:04:05"))
			}
		} else {
			fmt.Printf("   %s: %v\n", key, value)
		}
	}

	// Close connections
	grpcClient.Close()

	fmt.Println("\n=== Demo completed ===")
	fmt.Println("\nKey features demonstrated:")
	fmt.Println("✅ Automatic logging of plugin execution start/end")
	fmt.Println("✅ Automatic statistics recording for each execution")
	fmt.Println("✅ Error tracking and logging")
	fmt.Println("✅ Plugin-specific log filtering")
	fmt.Println("✅ Persistent storage of statistics")
	fmt.Println("✅ Non-intrusive integration with existing execution flow")
	fmt.Println("\nAll plugin executions are now automatically tracked!")
}
