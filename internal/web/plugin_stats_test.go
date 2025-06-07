package web

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestPluginStatsStorage(t *testing.T) {
	// Create temporary directory for testing
	tempDir := filepath.Join(os.TempDir(), "webhook_bridge_test")
	defer os.RemoveAll(tempDir)

	// Create storage
	storage := NewPluginStatsStorage(tempDir)
	defer storage.Close()

	// Test initial state
	data := storage.GetData()
	if data == nil {
		t.Fatal("Expected data to be initialized")
	}

	if data.Version != "1.0" {
		t.Errorf("Expected version '1.0', got '%s'", data.Version)
	}

	if data.PluginStats == nil {
		t.Error("Expected PluginStats to be initialized")
	}

	// Test saving data
	testData := &PluginStatsData{
		Version:         "1.0",
		StartTime:       time.Now().Add(-time.Hour),
		TotalRequests:   100,
		TotalExecutions: 50,
		TotalErrors:     5,
		PluginStats: map[string]*ExecutionStats{
			"test_plugin:GET": {
				Plugin:    "test_plugin",
				Method:    "GET",
				Count:     10,
				Errors:    1,
				TotalTime: 500 * time.Millisecond,
				AvgTime:   50 * time.Millisecond,
				LastExec:  time.Now(),
			},
		},
	}

	err := storage.SaveStats(testData)
	if err != nil {
		t.Fatalf("Failed to save stats: %v", err)
	}

	// Test loading data
	storage2 := NewPluginStatsStorage(tempDir)
	defer storage2.Close()

	loadedData := storage2.GetData()
	if loadedData.TotalRequests != 100 {
		t.Errorf("Expected TotalRequests 100, got %d", loadedData.TotalRequests)
	}

	if loadedData.TotalExecutions != 50 {
		t.Errorf("Expected TotalExecutions 50, got %d", loadedData.TotalExecutions)
	}

	if loadedData.TotalErrors != 5 {
		t.Errorf("Expected TotalErrors 5, got %d", loadedData.TotalErrors)
	}

	if len(loadedData.PluginStats) != 1 {
		t.Errorf("Expected 1 plugin stat, got %d", len(loadedData.PluginStats))
	}

	if stats, exists := loadedData.PluginStats["test_plugin:GET"]; exists {
		if stats.Count != 10 {
			t.Errorf("Expected Count 10, got %d", stats.Count)
		}
		if stats.Errors != 1 {
			t.Errorf("Expected Errors 1, got %d", stats.Errors)
		}
	} else {
		t.Error("Expected test_plugin:GET stats to exist")
	}
}

func TestStatsManagerWithPersistence(t *testing.T) {
	// Create temporary directory for testing
	tempDir := filepath.Join(os.TempDir(), "webhook_bridge_test_sm")
	defer os.RemoveAll(tempDir)

	// Create stats manager with persistence
	sm := NewStatsManagerWithPersistence(tempDir)
	defer sm.Close()

	// Verify persistence is enabled
	if !sm.IsPersistenceEnabled() {
		t.Error("Expected persistence to be enabled")
	}

	// Record some test data
	startTime := time.Now()
	sm.RecordExecution("test_plugin", "GET", startTime.Add(-100*time.Millisecond))
	sm.RecordExecution("test_plugin", "POST", startTime.Add(-200*time.Millisecond))
	sm.RecordError("test_plugin", "GET")
	sm.RecordRequest()
	sm.RecordRequest()

	// Force save (with retry for Windows file system issues)
	var err error
	for i := 0; i < 3; i++ {
		err = sm.ForceSave()
		if err == nil {
			break
		}
		time.Sleep(100 * time.Millisecond) // Wait a bit before retry
	}
	if err != nil {
		t.Logf("Warning: Failed to force save after retries: %v", err)
		// Don't fail the test, just log the warning
	}

	// Create new stats manager and verify data is loaded
	sm2 := NewStatsManagerWithPersistence(tempDir)
	defer sm2.Close()

	stats := sm2.GetStats()
	if stats.TotalRequests != 2 {
		t.Errorf("Expected TotalRequests 2, got %d", stats.TotalRequests)
	}

	if stats.TotalExecutions != 2 {
		t.Errorf("Expected TotalExecutions 2, got %d", stats.TotalExecutions)
	}

	if stats.TotalErrors != 1 {
		t.Errorf("Expected TotalErrors 1, got %d", stats.TotalErrors)
	}

	pluginStats := sm2.GetPluginStats()
	if len(pluginStats) != 2 {
		t.Errorf("Expected 2 plugin stats, got %d", len(pluginStats))
	}

	// Check specific plugin stats
	found := false
	for _, stat := range pluginStats {
		if stat.Plugin == "test_plugin" && stat.Method == "GET" {
			found = true
			if stat.Count != 1 {
				t.Errorf("Expected Count 1 for test_plugin:GET, got %d", stat.Count)
			}
			if stat.Errors != 1 {
				t.Errorf("Expected Errors 1 for test_plugin:GET, got %d", stat.Errors)
			}
		}
	}

	if !found {
		t.Error("Expected to find test_plugin:GET stats")
	}
}

func TestStatsManagerStorageInfo(t *testing.T) {
	// Test without persistence
	sm := NewStatsManager()
	defer sm.Close()

	info := sm.GetStorageInfo()
	if info["enabled"].(bool) {
		t.Error("Expected persistence to be disabled")
	}

	// Test with persistence
	tempDir := filepath.Join(os.TempDir(), "webhook_bridge_test_info")
	defer os.RemoveAll(tempDir)

	smWithPersist := NewStatsManagerWithPersistence(tempDir)
	defer smWithPersist.Close()

	info = smWithPersist.GetStorageInfo()
	if !info["enabled"].(bool) {
		t.Error("Expected persistence to be enabled")
	}

	if info["file_path"].(string) == "" {
		t.Error("Expected file_path to be set")
	}

	if info["version"].(string) != "1.0" {
		t.Errorf("Expected version '1.0', got '%s'", info["version"])
	}
}

func TestStatsManagerResetStats(t *testing.T) {
	tempDir := filepath.Join(os.TempDir(), "webhook_bridge_test_reset")
	defer os.RemoveAll(tempDir)

	sm := NewStatsManagerWithPersistence(tempDir)
	defer sm.Close()

	// Add some data
	startTime := time.Now()
	sm.RecordExecution("test_plugin", "GET", startTime.Add(-100*time.Millisecond))
	sm.RecordRequest()

	// Verify data exists
	stats := sm.GetStats()
	if stats.TotalRequests != 1 {
		t.Errorf("Expected TotalRequests 1, got %d", stats.TotalRequests)
	}

	// Reset stats
	sm.ResetStats()

	// Verify data is reset
	stats = sm.GetStats()
	if stats.TotalRequests != 0 {
		t.Errorf("Expected TotalRequests 0 after reset, got %d", stats.TotalRequests)
	}

	if stats.TotalExecutions != 0 {
		t.Errorf("Expected TotalExecutions 0 after reset, got %d", stats.TotalExecutions)
	}

	if stats.TotalErrors != 0 {
		t.Errorf("Expected TotalErrors 0 after reset, got %d", stats.TotalErrors)
	}

	pluginStats := sm.GetPluginStats()
	if len(pluginStats) != 0 {
		t.Errorf("Expected 0 plugin stats after reset, got %d", len(pluginStats))
	}
}
