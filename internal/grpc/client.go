package grpc

import (
	"context"
	"fmt"
	"strings"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/loonghao/webhook_bridge/api/proto"
	"github.com/loonghao/webhook_bridge/internal/config"
)

// Client represents a gRPC client for communicating with Python executor
type Client struct {
	config       *config.ExecutorConfig
	conn         *grpc.ClientConn
	client       proto.WebhookExecutorClient
	timeout      time.Duration
	connected    bool
	retryCount   int
	maxRetries   int
	logManager   LogManager   // Interface for logging
	statsManager StatsManager // Interface for statistics
}

// LogManager interface for plugin execution logging
type LogManager interface {
	AddLog(entry LogEntry)
}

// StatsManager interface for plugin execution statistics
type StatsManager interface {
	RecordExecution(plugin, method string, startTime time.Time)
	RecordError(plugin, method string)
}

// LogEntry represents a log entry (simplified interface)
type LogEntry struct {
	Timestamp  time.Time
	Level      string
	Source     string
	Message    string
	PluginName string
	Data       map[string]interface{}
}

// NewClient creates a new gRPC client
func NewClient(cfg *config.ExecutorConfig) *Client {
	return &Client{
		config:     cfg,
		timeout:    time.Duration(cfg.Timeout) * time.Second,
		connected:  false,
		retryCount: 0,
		maxRetries: 3,
	}
}

// SetLogManager sets the log manager for plugin execution logging
func (c *Client) SetLogManager(logManager LogManager) {
	c.logManager = logManager
}

// SetStatsManager sets the stats manager for plugin execution statistics
func (c *Client) SetStatsManager(statsManager StatsManager) {
	c.statsManager = statsManager
}

// Connect establishes connection to the Python executor service
func (c *Client) Connect() error {
	address := fmt.Sprintf("%s:%d", c.config.Host, c.config.Port)

	// Add connection options for better reliability
	opts := []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(),
		grpc.WithTimeout(10 * time.Second),
	}

	conn, err := grpc.Dial(address, opts...)
	if err != nil {
		return fmt.Errorf("failed to connect to Python executor at %s: %w", address, err)
	}

	c.conn = conn
	c.client = proto.NewWebhookExecutorClient(conn)
	c.connected = true
	c.retryCount = 0

	return nil
}

// Close closes the gRPC connection
func (c *Client) Close() error {
	if c.conn != nil {
		c.connected = false
		return c.conn.Close()
	}
	return nil
}

// IsConnected returns whether the client is connected
func (c *Client) IsConnected() bool {
	return c.connected && c.conn != nil
}

// Reconnect attempts to reconnect to the Python executor
func (c *Client) Reconnect() error {
	if c.retryCount >= c.maxRetries {
		return fmt.Errorf("max reconnection attempts (%d) exceeded", c.maxRetries)
	}

	// Close existing connection
	if c.conn != nil {
		c.conn.Close()
	}

	c.retryCount++

	// Wait before retry
	time.Sleep(time.Duration(c.retryCount) * time.Second)

	return c.Connect()
}

// ExecutePlugin executes a plugin via gRPC with retry logic and logging/stats
func (c *Client) ExecutePlugin(ctx context.Context, req *proto.ExecutePluginRequest) (*proto.ExecutePluginResponse, error) {
	// Record execution start time for statistics
	startTime := time.Now()
	pluginName := req.PluginName
	method := req.HttpMethod

	// Log execution start
	if c.logManager != nil {
		c.logManager.AddLog(LogEntry{
			Timestamp:  startTime,
			Level:      "INFO",
			Source:     "grpc_client",
			Message:    fmt.Sprintf("Starting plugin execution: %s [%s]", pluginName, method),
			PluginName: pluginName,
			Data: map[string]interface{}{
				"method":      method,
				"plugin_name": pluginName,
				"request_id":  fmt.Sprintf("%d", startTime.UnixNano()),
			},
		})
	}

	// Execute plugin with retry logic
	resp, err := c.executeWithRetry(func() (interface{}, error) {
		if c.client == nil {
			return nil, fmt.Errorf("gRPC client not connected")
		}

		// Create context with timeout
		ctx, cancel := context.WithTimeout(ctx, c.timeout)
		defer cancel()

		return c.client.ExecutePlugin(ctx, req)
	})

	// Record execution statistics
	if c.statsManager != nil {
		c.statsManager.RecordExecution(pluginName, method, startTime)
		if err != nil || (resp != nil && resp.StatusCode >= 400) {
			c.statsManager.RecordError(pluginName, method)
		}
	}

	// Log execution result
	if c.logManager != nil {
		executionTime := time.Since(startTime)
		logLevel := "INFO"
		logMessage := fmt.Sprintf("Plugin execution completed: %s [%s] in %v", pluginName, method, executionTime)

		logData := map[string]interface{}{
			"method":         method,
			"plugin_name":    pluginName,
			"execution_time": executionTime.String(),
			"request_id":     fmt.Sprintf("%d", startTime.UnixNano()),
		}

		if err != nil {
			logLevel = "ERROR"
			logMessage = fmt.Sprintf("Plugin execution failed: %s [%s] - %v", pluginName, method, err)
			logData["error"] = err.Error()
		} else if resp != nil {
			logData["status_code"] = resp.StatusCode
			logData["response_message"] = resp.Message

			if resp.StatusCode >= 400 {
				logLevel = "WARN"
				logMessage = fmt.Sprintf("Plugin execution returned error status: %s [%s] - %d", pluginName, method, resp.StatusCode)
			}
		}

		c.logManager.AddLog(LogEntry{
			Timestamp:  time.Now(),
			Level:      logLevel,
			Source:     "grpc_client",
			Message:    logMessage,
			PluginName: pluginName,
			Data:       logData,
		})
	}

	return resp, err
}

// executeWithRetry executes a gRPC call with automatic retry on connection failures
func (c *Client) executeWithRetry(fn func() (interface{}, error)) (*proto.ExecutePluginResponse, error) {
	for attempt := 0; attempt <= c.maxRetries; attempt++ {
		result, err := fn()
		if err == nil {
			if resp, ok := result.(*proto.ExecutePluginResponse); ok {
				return resp, nil
			}
			return nil, fmt.Errorf("unexpected response type")
		}

		// Check if it's a connection error
		if isConnectionError(err) && attempt < c.maxRetries {
			// Try to reconnect
			if reconnectErr := c.Reconnect(); reconnectErr != nil {
				return nil, fmt.Errorf("failed to reconnect: %w", reconnectErr)
			}
			continue
		}

		return nil, err
	}

	return nil, fmt.Errorf("max retry attempts exceeded")
}

// isConnectionError checks if the error is a connection-related error
func isConnectionError(err error) bool {
	if err == nil {
		return false
	}

	errStr := err.Error()
	connectionErrors := []string{
		"connection refused",
		"connection reset",
		"connection closed",
		"no such host",
		"network is unreachable",
		"timeout",
		"context deadline exceeded",
	}

	for _, connErr := range connectionErrors {
		if strings.Contains(strings.ToLower(errStr), strings.ToLower(connErr)) {
			return true
		}
	}

	return false
}

// ListPlugins lists available plugins via gRPC
func (c *Client) ListPlugins(ctx context.Context, req *proto.ListPluginsRequest) (*proto.ListPluginsResponse, error) {
	if c.client == nil {
		return nil, fmt.Errorf("gRPC client not connected")
	}

	// Create context with timeout
	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	return c.client.ListPlugins(ctx, req)
}

// GetPluginInfo gets plugin information via gRPC
func (c *Client) GetPluginInfo(ctx context.Context, req *proto.GetPluginInfoRequest) (*proto.GetPluginInfoResponse, error) {
	if c.client == nil {
		return nil, fmt.Errorf("gRPC client not connected")
	}

	// Create context with timeout
	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	return c.client.GetPluginInfo(ctx, req)
}

// HealthCheck performs health check via gRPC
func (c *Client) HealthCheck(ctx context.Context, req *proto.HealthCheckRequest) (*proto.HealthCheckResponse, error) {
	if c.client == nil {
		return nil, fmt.Errorf("gRPC client not connected")
	}

	// Create context with timeout
	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	return c.client.HealthCheck(ctx, req)
}
