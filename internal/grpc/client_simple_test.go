package grpc

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/loonghao/webhook_bridge/api/proto"
	"github.com/loonghao/webhook_bridge/internal/config"
)

// TestNewClient tests the creation of a new gRPC client
func TestNewClient(t *testing.T) {
	cfg := &config.ExecutorConfig{
		Host:    "localhost",
		Port:    50051,
		Timeout: 30,
	}

	client := NewClient(cfg)

	assert.NotNil(t, client)
	assert.False(t, client.IsConnected())
}

// TestClientConnectionState tests the connection state management
func TestClientConnectionState(t *testing.T) {
	cfg := &config.ExecutorConfig{
		Host:    "localhost",
		Port:    50051,
		Timeout: 30,
	}
	client := NewClient(cfg)

	// Initially not connected
	assert.False(t, client.IsConnected())

	// Test connection attempt (will fail since no server is running)
	err := client.Connect()
	assert.Error(t, err)
	assert.False(t, client.IsConnected())
}

// TestClientClose tests the close functionality
func TestClientClose(t *testing.T) {
	cfg := &config.ExecutorConfig{
		Host:    "localhost",
		Port:    50051,
		Timeout: 30,
	}
	client := NewClient(cfg)

	// Test close without connection (should not panic)
	assert.NotPanics(t, func() {
		err := client.Close()
		assert.NoError(t, err)
	})
}

// TestExecutePluginRequest tests the plugin execution request structure
func TestExecutePluginRequest(t *testing.T) {
	req := &proto.ExecutePluginRequest{
		PluginName: "test_plugin",
		HttpMethod: "POST",
		Data:       map[string]string{"key": "value"},
		Headers: map[string]string{
			"Content-Type": "application/json",
		},
	}

	assert.Equal(t, "test_plugin", req.PluginName)
	assert.Equal(t, "POST", req.HttpMethod)
	assert.NotNil(t, req.Data)
	assert.Equal(t, "application/json", req.Headers["Content-Type"])
}

// TestClientWithTimeout tests client operations with timeout
func TestClientWithTimeout(t *testing.T) {
	cfg := &config.ExecutorConfig{
		Host:    "localhost",
		Port:    50051,
		Timeout: 1, // 1 second timeout
	}
	client := NewClient(cfg)

	// Test with context timeout
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	// This should timeout quickly since no server is running
	req := &proto.ExecutePluginRequest{
		PluginName: "test_plugin",
		HttpMethod: "GET",
	}

	_, err := client.ExecutePlugin(ctx, req)
	assert.Error(t, err)
}

// TestClientErrorHandling tests error handling scenarios
func TestClientErrorHandling(t *testing.T) {
	cfg := &config.ExecutorConfig{
		Host:    "invalid-host",
		Port:    99999,
		Timeout: 1,
	}
	client := NewClient(cfg)

	// Test connection to invalid host
	err := client.Connect()
	assert.Error(t, err)

	// Test operations on disconnected client
	ctx := context.Background()

	_, err = client.ExecutePlugin(ctx, &proto.ExecutePluginRequest{})
	assert.Error(t, err)

	_, err = client.ListPlugins(ctx, &proto.ListPluginsRequest{})
	assert.Error(t, err)

	_, err = client.GetPluginInfo(ctx, &proto.GetPluginInfoRequest{})
	assert.Error(t, err)
}

// TestClientConcurrentAccess tests concurrent access to client
func TestClientConcurrentAccess(t *testing.T) {
	cfg := &config.ExecutorConfig{
		Host:    "localhost",
		Port:    50051,
		Timeout: 30,
	}
	client := NewClient(cfg)

	// Test concurrent IsConnected calls (should not panic)
	done := make(chan bool, 10)

	for i := 0; i < 10; i++ {
		go func() {
			defer func() { done <- true }()
			client.IsConnected()
		}()
	}

	// Wait for all goroutines to complete
	for i := 0; i < 10; i++ {
		<-done
	}

	// If we reach here without panic, the test passes
	assert.True(t, true)
}

// TestLogEntry tests the log entry structure
func TestLogEntry(t *testing.T) {
	entry := LogEntry{
		Timestamp:  time.Now(),
		Level:      "INFO",
		Source:     "test",
		Message:    "Test message",
		PluginName: "test_plugin",
		Data: map[string]interface{}{
			"key": "value",
		},
	}

	assert.Equal(t, "INFO", entry.Level)
	assert.Equal(t, "test", entry.Source)
	assert.Equal(t, "Test message", entry.Message)
	assert.Equal(t, "test_plugin", entry.PluginName)
	assert.NotNil(t, entry.Data)
}

// TestClientSetManagers tests setting log and stats managers
func TestClientSetManagers(t *testing.T) {
	cfg := &config.ExecutorConfig{
		Host:    "localhost",
		Port:    50051,
		Timeout: 30,
	}
	client := NewClient(cfg)

	// Create mock managers
	logManager := &MockLogManager{}
	statsManager := &MockStatsManager{}

	// Test setting managers (should not panic)
	assert.NotPanics(t, func() {
		client.SetLogManager(logManager)
		client.SetStatsManager(statsManager)
	})
}

// MockLogManager is a simple mock for testing
type MockLogManager struct{}

func (m *MockLogManager) AddLog(entry LogEntry) {
	// Mock implementation - do nothing
}

// MockStatsManager is a simple mock for testing
type MockStatsManager struct{}

func (m *MockStatsManager) RecordExecution(plugin, method string, startTime time.Time) {
	// Mock implementation - do nothing
}

func (m *MockStatsManager) RecordError(plugin, method string) {
	// Mock implementation - do nothing
}

// TestIsConnectionError tests the connection error detection
func TestIsConnectionError(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name:     "nil error",
			err:      nil,
			expected: false,
		},
		{
			name:     "connection refused",
			err:      assert.AnError,
			expected: false, // Since assert.AnError doesn't contain connection keywords
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isConnectionError(tt.err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestClientReconnectLogic tests the reconnection logic
func TestClientReconnectLogic(t *testing.T) {
	cfg := &config.ExecutorConfig{
		Host:    "localhost",
		Port:    50051,
		Timeout: 1,
	}
	client := NewClient(cfg)

	// Test reconnect without initial connection
	err := client.Reconnect()
	assert.Error(t, err)
	assert.False(t, client.IsConnected())

	// Test multiple reconnect attempts (should eventually fail)
	for i := 0; i < 5; i++ {
		err = client.Reconnect()
		assert.Error(t, err)
	}
}

// TestProtoStructures tests proto message structures
func TestProtoStructures(t *testing.T) {
	// Test ExecutePluginRequest
	execReq := &proto.ExecutePluginRequest{
		PluginName: "test",
		HttpMethod: "POST",
	}
	assert.Equal(t, "test", execReq.PluginName)
	assert.Equal(t, "POST", execReq.HttpMethod)

	// Test ListPluginsRequest
	listReq := &proto.ListPluginsRequest{}
	assert.NotNil(t, listReq)

	// Test GetPluginInfoRequest
	infoReq := &proto.GetPluginInfoRequest{
		PluginName: "test",
	}
	assert.Equal(t, "test", infoReq.PluginName)

	// Test HealthCheckRequest
	healthReq := &proto.HealthCheckRequest{}
	assert.NotNil(t, healthReq)
}
