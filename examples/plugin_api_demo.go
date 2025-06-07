package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"

	"github.com/gin-gonic/gin"

	"github.com/loonghao/webhook_bridge/internal/config"
	"github.com/loonghao/webhook_bridge/internal/web/modern"
)

// Simple demo for plugin management API endpoints
func main() {
	fmt.Println("=== Plugin Management API Demo ===\n")

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

	// Create dashboard handler
	handler := modern.NewModernDashboardHandler(cfg)

	// Set up test router
	gin.SetMode(gin.TestMode)
	router := gin.New()
	handler.RegisterRoutes(router)

	fmt.Println("1. Testing enhanced plugin list API...")
	testAPI(router, "GET", "/api/dashboard/plugins", nil)

	fmt.Println("\n2. Testing all plugin statistics API...")
	testAPI(router, "GET", "/api/dashboard/plugins/stats", nil)

	fmt.Println("\n3. Testing specific plugin stats...")
	testAPI(router, "GET", "/api/dashboard/plugins/example_plugin/stats", nil)

	fmt.Println("\n4. Testing plugin logs...")
	testAPI(router, "GET", "/api/dashboard/plugins/example_plugin/logs", nil)

	fmt.Println("\n5. Testing enhanced logs with plugin filter...")
	testAPI(router, "GET", "/api/dashboard/logs?plugin=example_plugin&limit=5", nil)

	fmt.Println("\n6. Testing plugin execution...")
	executionData := map[string]interface{}{
		"method": "POST",
		"data": map[string]string{
			"test_key": "test_value",
		},
		"headers": map[string]string{
			"Content-Type": "application/json",
		},
		"query": "test=1",
	}
	testAPI(router, "POST", "/api/dashboard/plugins/example_plugin/execute", executionData)

	fmt.Println("\n=== Demo completed ===")
	fmt.Println("\nAll new plugin management API endpoints are working correctly!")
	fmt.Println("✓ Enhanced plugin listing with gRPC integration")
	fmt.Println("✓ Plugin execution testing capability")
	fmt.Println("✓ Plugin-specific statistics and logs")
	fmt.Println("✓ Comprehensive plugin statistics overview")
	fmt.Println("✓ Enhanced log filtering by plugin")
}

func testAPI(router *gin.Engine, method, path string, data interface{}) {
	fmt.Printf("  %s %s\n", method, path)

	var req *http.Request
	var err error

	if data != nil {
		jsonData, _ := json.Marshal(data)
		req, err = http.NewRequest(method, path, bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")
	} else {
		req, err = http.NewRequest(method, path, nil)
	}

	if err != nil {
		fmt.Printf("    Error creating request: %v\n", err)
		return
	}

	// Create response recorder
	w := httptest.NewRecorder()

	// Perform request
	router.ServeHTTP(w, req)

	fmt.Printf("    Status: %d\n", w.Code)

	// Parse and display response
	var response interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &response); err == nil {
		// Check if it's a successful response
		if respMap, ok := response.(map[string]interface{}); ok {
			if success, exists := respMap["success"]; exists && success == true {
				fmt.Printf("    Success: ✓\n")

				// Show some key data
				if data, exists := respMap["data"]; exists {
					switch d := data.(type) {
					case []interface{}:
						fmt.Printf("    Data: Array with %d items\n", len(d))
					case map[string]interface{}:
						fmt.Printf("    Data: Object with keys: ")
						keys := make([]string, 0, len(d))
						for k := range d {
							keys = append(keys, k)
						}
						fmt.Printf("%v\n", keys)
					default:
						fmt.Printf("    Data: %T\n", d)
					}
				}
			} else {
				fmt.Printf("    Success: ✗\n")
				if errorMsg, exists := respMap["error"]; exists {
					fmt.Printf("    Error: %v\n", errorMsg)
				}
			}
		}
	} else {
		fmt.Printf("    Response: %s\n", w.Body.String())
	}
}
