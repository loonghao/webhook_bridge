package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/loonghao/webhook_bridge/internal/config"
	"github.com/loonghao/webhook_bridge/internal/web/modern"
)

// Example demonstrating the new plugin management API endpoints
func main() {
	fmt.Println("=== Plugin Management API Demo ===")
	fmt.Println()

	// Create a test configuration
	cfg := &config.Config{
		Server: config.ServerConfig{
			Host: "localhost",
			Port: 8080,
			Mode: "debug",
		},
		Executor: config.ExecutorConfig{
			Host:    "localhost",
			Port:    50051,
			Timeout: 30,
		},
		Directories: config.DirectoriesConfig{
			WorkingDir: ".",
			LogDir:     "logs",
			PluginDir:  "plugins",
			DataDir:    "data",
		},
		Logging: config.LoggingConfig{
			Level:  "info",
			Format: "text",
		},
	}

	// Create dashboard handler with plugin management capabilities
	handler := modern.NewModernDashboardHandler(cfg)

	// Set up Gin router
	gin.SetMode(gin.TestMode)
	router := gin.New()
	handler.RegisterRoutes(router)

	// Start server in background
	go func() {
		fmt.Printf("Starting test server on http://localhost:8080\n")
		if err := router.Run(":8080"); err != nil {
			fmt.Printf("⚠️ Server failed to start: %v\n", err)
		}
	}()

	// Wait for server to start
	time.Sleep(2 * time.Second)

	// Test the new plugin management API endpoints
	baseURL := "http://localhost:8080/api/dashboard"

	fmt.Println("1. Testing plugin list API...")
	testGetRequest(baseURL + "/plugins")

	fmt.Println("\n2. Testing plugin statistics API...")
	testGetRequest(baseURL + "/plugins/stats")

	fmt.Println("\n3. Testing specific plugin stats...")
	testGetRequest(baseURL + "/plugins/example_plugin/stats")

	fmt.Println("\n4. Testing plugin logs...")
	testGetRequest(baseURL + "/plugins/example_plugin/logs")

	fmt.Println("\n5. Testing plugin execution...")
	testPluginExecution(baseURL + "/plugins/example_plugin/execute")

	fmt.Println("\n6. Testing logs with plugin filter...")
	testGetRequest(baseURL + "/logs?plugin=example_plugin&limit=10")

	fmt.Println("\n=== Demo completed ===")
	fmt.Println("\nNew API endpoints demonstrated:")
	fmt.Println("✅ GET /api/dashboard/plugins - Enhanced with real gRPC integration")
	fmt.Println("✅ POST /api/dashboard/plugins/:name/execute - Manual plugin testing")
	fmt.Println("✅ GET /api/dashboard/plugins/:name/stats - Plugin-specific statistics")
	fmt.Println("✅ GET /api/dashboard/plugins/:name/logs - Plugin-specific logs")
	fmt.Println("✅ GET /api/dashboard/plugins/stats - All plugin statistics")
	fmt.Println("✅ GET /api/dashboard/logs?plugin=name - Enhanced logs with plugin filtering")
}

func testGetRequest(url string) {
	fmt.Printf("  GET %s\n", url)

	// Validate URL to prevent SSRF attacks
	if !strings.HasPrefix(url, "http://localhost:8080/") {
		fmt.Printf("    Error: Invalid URL - only localhost:8080 allowed\n")
		return
	}

	// Create HTTP client with timeout for security
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	req, err := http.NewRequest("GET", "http://localhost:8080"+strings.TrimPrefix(url, "http://localhost:8080"), nil)
	if err != nil {
		fmt.Printf("    Error creating request: %v\n", err)
		return
	}

	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("    Error: %v\n", err)
		return
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			fmt.Printf("    Warning: Failed to close response body: %v\n", err)
		}
	}()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("    Error reading response: %v\n", err)
		return
	}

	fmt.Printf("    Status: %d\n", resp.StatusCode)

	// Pretty print JSON response
	var jsonData any
	if err := json.Unmarshal(body, &jsonData); err == nil {
		prettyJSON, _ := json.MarshalIndent(jsonData, "    ", "  ")
		fmt.Printf("    Response: %s\n", string(prettyJSON))
	} else {
		fmt.Printf("    Response: %s\n", string(body))
	}
}

func testPluginExecution(url string) {
	fmt.Printf("  POST %s\n", url)

	// Validate URL to prevent SSRF attacks
	if !strings.HasPrefix(url, "http://localhost:8080/") {
		fmt.Printf("    Error: Invalid URL - only localhost:8080 allowed\n")
		return
	}

	// Create test execution request
	requestData := map[string]any{
		"method": "POST",
		"data": map[string]string{
			"test_key": "test_value",
			"message":  "Hello from API test",
		},
		"headers": map[string]string{
			"Content-Type": "application/json",
			"X-Test":       "true",
		},
		"query": "test=1&debug=true",
	}

	jsonData, err := json.Marshal(requestData)
	if err != nil {
		fmt.Printf("    Error marshaling request: %v\n", err)
		return
	}

	// Create HTTP client with timeout for security
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	req, err := http.NewRequest("POST", "http://localhost:8080"+strings.TrimPrefix(url, "http://localhost:8080"), bytes.NewBuffer(jsonData))
	if err != nil {
		fmt.Printf("    Error creating request: %v\n", err)
		return
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("    Error: %v\n", err)
		return
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			fmt.Printf("    Warning: Failed to close response body: %v\n", err)
		}
	}()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("    Error reading response: %v\n", err)
		return
	}

	fmt.Printf("    Status: %d\n", resp.StatusCode)

	// Pretty print JSON response
	var jsonResponse any
	if err := json.Unmarshal(body, &jsonResponse); err == nil {
		prettyJSON, _ := json.MarshalIndent(jsonResponse, "    ", "  ")
		fmt.Printf("    Response: %s\n", string(prettyJSON))
	} else {
		fmt.Printf("    Response: %s\n", string(body))
	}
}
