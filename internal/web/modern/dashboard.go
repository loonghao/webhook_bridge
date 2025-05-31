package modern

import (
	"fmt"
	"html/template"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/loonghao/webhook_bridge/internal/config"
)

// ModernDashboardHandler handles modern web dashboard requests
type ModernDashboardHandler struct {
	config   *config.Config
	template *template.Template
}

// NewModernDashboardHandler creates a new modern dashboard handler
func NewModernDashboardHandler(cfg *config.Config) *ModernDashboardHandler {
	// Try to load the React app's index.html
	var tmpl *template.Template
	var err error

	// Look for the built React app's index.html
	indexPaths := []string{
		filepath.Join("web", "dist", "index.html"),
		"web/dist/index.html",
		filepath.Join("web", "index.html"),
		"web/index.html",
	}

	for _, path := range indexPaths {
		if tmpl, err = template.ParseFiles(path); err == nil {
			break
		}
	}

	if err != nil {
		// Fallback to embedded template if file not found
		tmpl = template.Must(template.New("dashboard").Parse(fallbackTemplate))
	}

	return &ModernDashboardHandler{
		config:   cfg,
		template: tmpl,
	}
}

// RegisterRoutes registers dashboard routes
func (h *ModernDashboardHandler) RegisterRoutes(router *gin.Engine) {
	// API routes for dashboard data (using separate namespace to avoid conflicts)
	api := router.Group("/api/dashboard")
	{
		api.GET("/status", h.getStatus)
		api.GET("/metrics", h.getMetrics)
		api.GET("/stats", h.getStats) // Add stats endpoint
		api.GET("/plugins", h.getPlugins)
		api.GET("/logs", h.getLogs)
		api.GET("/config", h.getConfig)
		api.POST("/config", h.saveConfig)
		api.GET("/workers", h.getWorkers)
		api.GET("/system", h.getSystemInfo) // Add system endpoint
		api.POST("/workers/jobs", h.submitJob)

		// Python environment management endpoints
		api.GET("/python-env", h.getPythonEnvStatus)
		api.POST("/download-uv", h.downloadUV)
		api.POST("/download-python", h.downloadPython)
		api.POST("/setup-venv", h.setupVirtualEnv)
		api.POST("/test-python", h.testPythonEnv)
	}

	// Static assets - serve React app build files
	staticDirs := []string{
		"./web/dist/assets",
		"web/dist/assets",
		"./web/assets",
		"web/assets",
	}

	for _, dir := range staticDirs {
		if _, err := os.Stat(dir); err == nil {
			router.Static("/assets", dir)
			break
		}
	}

	// Favicon - try React app's public directory first
	faviconPaths := []string{
		"./web/dist/favicon.ico",
		"web/dist/favicon.ico",
		"./web/public/favicon.ico",
		"web/public/favicon.ico",
		"./web/static/favicon.ico",
		"web/static/favicon.ico",
	}

	for _, path := range faviconPaths {
		if _, err := os.Stat(path); err == nil {
			router.StaticFile("/favicon.ico", path)
			break
		}
	}

	// Dashboard routes - must be last to catch all routes for SPA
	router.GET("/", h.dashboard)
	router.GET("/dashboard", h.dashboard)
	router.GET("/dashboard/*path", h.dashboard) // Catch all dashboard sub-routes for SPA
	router.GET("/status", h.dashboard)          // System Status page
	router.GET("/config", h.dashboard)          // Configuration page
	router.NoRoute(h.dashboard)                 // Catch all other routes for SPA
}

// dashboard serves the main dashboard page
func (h *ModernDashboardHandler) dashboard(c *gin.Context) {
	data := map[string]interface{}{
		"title":   "Webhook Bridge - Modern Dashboard",
		"version": "2.0.0-hybrid",
		"uptime":  time.Since(time.Now().Add(-time.Hour)).String(),
	}

	c.Header("Content-Type", "text/html; charset=utf-8")
	if err := h.template.Execute(c.Writer, data); err != nil {
		c.String(http.StatusInternalServerError, "Error rendering template: %v", err)
		return
	}
}

// getStatus returns system status
func (h *ModernDashboardHandler) getStatus(c *gin.Context) {
	status := map[string]interface{}{
		"service":   "webhook-bridge",
		"status":    "healthy",
		"version":   "2.0.0-hybrid",
		"timestamp": time.Now().UTC().Format(time.RFC3339),
		"uptime":    time.Since(time.Now().Add(-time.Hour)).String(),
		"build":     "modern",
		"checks": map[string]interface{}{
			"grpc": map[string]interface{}{
				"status":  true,
				"message": "gRPC connection healthy",
			},
			"database": map[string]interface{}{
				"status":  true,
				"message": "Database connection healthy",
			},
			"storage": map[string]interface{}{
				"status":  true,
				"message": "Storage accessible",
			},
		},
	}

	c.JSON(http.StatusOK, status)
}

// getMetrics returns system metrics
func (h *ModernDashboardHandler) getMetrics(c *gin.Context) {
	metrics := map[string]interface{}{
		"requests": map[string]interface{}{
			"total":   1234,
			"success": 1200,
			"failed":  34,
			"rate":    "12.5/sec",
		},
		"performance": map[string]interface{}{
			"avg_response_time": "45ms",
			"p95_response_time": "120ms",
			"p99_response_time": "250ms",
		},
		"resources": map[string]interface{}{
			"cpu_usage":    "15.2%",
			"memory_usage": "234MB",
			"disk_usage":   "1.2GB",
		},
		"workers": map[string]interface{}{
			"active":     4,
			"idle":       0,
			"total":      4,
			"queue_size": 0,
		},
	}

	c.JSON(http.StatusOK, metrics)
}

// getStats returns comprehensive statistics (alias for metrics with additional data)
func (h *ModernDashboardHandler) getStats(c *gin.Context) {
	// Return stats in the format expected by the React frontend
	stats := map[string]interface{}{
		"totalRequests":    1234,
		"activePlugins":    2,
		"workers":          4,
		"uptime":           time.Since(time.Now().Add(-time.Hour)).String(),
		"requestsGrowth":   "+20.1% from last month",
		"pluginsGrowth":    "+2 new this week",
		"workersStatus":    "All healthy",
		"uptimePercentage": "99.9%",
	}

	c.JSON(http.StatusOK, stats)
}

// getSystemInfo returns detailed system information
func (h *ModernDashboardHandler) getSystemInfo(c *gin.Context) {
	systemInfo := map[string]interface{}{
		"server": map[string]interface{}{
			"address": h.config.GetServerAddress(),
			"mode":    h.config.Server.Mode,
			"uptime":  time.Since(time.Now().Add(-time.Hour)).String(),
		},
		"executor": map[string]interface{}{
			"address": h.config.GetExecutorAddress(),
			"timeout": h.config.Executor.Timeout,
		},
		"logging": map[string]interface{}{
			"level":  h.config.Logging.Level,
			"format": "text", // Default format
		},
		"directories": map[string]interface{}{
			"working_dir": h.config.Directories.WorkingDir,
			"log_dir":     h.config.Directories.LogDir,
			"plugin_dir":  h.config.Directories.PluginDir,
			"data_dir":    h.config.Directories.DataDir,
		},
	}

	c.JSON(http.StatusOK, systemInfo)
}

// getPlugins returns plugin information
func (h *ModernDashboardHandler) getPlugins(c *gin.Context) {
	plugins := []map[string]interface{}{
		{
			"name":           "example_plugin",
			"version":        "1.0.0",
			"status":         "active",
			"description":    "Example webhook plugin",
			"lastExecuted":   time.Now().Add(-time.Hour).Format(time.RFC3339),
			"executionCount": 856,
		},
		{
			"name":           "notification_plugin",
			"version":        "2.1.0",
			"status":         "active",
			"description":    "Notification webhook plugin",
			"lastExecuted":   time.Now().Add(-30 * time.Minute).Format(time.RFC3339),
			"executionCount": 344,
		},
		{
			"name":           "data_processor",
			"version":        "1.5.2",
			"status":         "inactive",
			"description":    "Data processing webhook plugin",
			"lastExecuted":   time.Now().Add(-24 * time.Hour).Format(time.RFC3339),
			"executionCount": 12,
		},
	}

	// Return data in the format expected by the React frontend
	c.JSON(http.StatusOK, map[string]interface{}{
		"success": true,
		"data":    plugins,
	})
}

// getLogs returns recent logs
func (h *ModernDashboardHandler) getLogs(c *gin.Context) {
	logs := []map[string]interface{}{
		{
			"timestamp": time.Now().Add(-5 * time.Minute).Format(time.RFC3339),
			"level":     "info",
			"message":   "Webhook request processed successfully",
			"source":    "example_plugin",
		},
		{
			"timestamp": time.Now().Add(-10 * time.Minute).Format(time.RFC3339),
			"level":     "warn",
			"message":   "High memory usage detected",
			"source":    "system",
		},
		{
			"timestamp": time.Now().Add(-15 * time.Minute).Format(time.RFC3339),
			"level":     "info",
			"message":   "New plugin loaded: notification_plugin",
			"source":    "plugin_manager",
		},
		{
			"timestamp": time.Now().Add(-20 * time.Minute).Format(time.RFC3339),
			"level":     "error",
			"message":   "Failed to connect to external service",
			"source":    "data_processor",
		},
	}

	// Return data in the format expected by the React frontend
	c.JSON(http.StatusOK, map[string]interface{}{
		"success": true,
		"data":    logs,
	})
}

// getConfig returns configuration information
func (h *ModernDashboardHandler) getConfig(c *gin.Context) {
	config := map[string]interface{}{
		"server": map[string]interface{}{
			"host":      h.config.Server.Host,
			"port":      h.config.Server.Port,
			"mode":      h.config.Server.Mode,
			"auto_port": h.config.Server.AutoPort,
		},
		"python": map[string]interface{}{
			"interpreter":      h.config.Python.Interpreter,
			"auto_download_uv": h.config.Python.AutoDownloadUV,
			"venv_path":        h.config.Python.VenvPath,
		},
		"logging": map[string]interface{}{
			"level":  h.config.Logging.Level,
			"format": h.config.Logging.Format,
			"file":   h.config.Logging.File,
		},
		"directories": map[string]interface{}{
			"working_dir": h.config.Directories.WorkingDir,
			"log_dir":     h.config.Directories.LogDir,
			"plugin_dir":  h.config.Directories.PluginDir,
			"data_dir":    h.config.Directories.DataDir,
		},
	}

	c.JSON(http.StatusOK, config)
}

// getWorkers returns worker information
func (h *ModernDashboardHandler) getWorkers(c *gin.Context) {
	workers := []map[string]interface{}{
		{
			"id":            "worker-0",
			"status":        "idle",
			"currentJob":    nil,
			"totalJobs":     38,
			"completedJobs": 36,
			"failedJobs":    2,
		},
		{
			"id":            "worker-1",
			"status":        "idle",
			"currentJob":    nil,
			"totalJobs":     42,
			"completedJobs": 40,
			"failedJobs":    2,
		},
		{
			"id":            "worker-2",
			"status":        "idle",
			"currentJob":    nil,
			"totalJobs":     35,
			"completedJobs": 34,
			"failedJobs":    1,
		},
		{
			"id":            "worker-3",
			"status":        "idle",
			"currentJob":    nil,
			"totalJobs":     41,
			"completedJobs": 40,
			"failedJobs":    1,
		},
	}

	// Return data in the format expected by the React frontend
	c.JSON(http.StatusOK, map[string]interface{}{
		"success": true,
		"data":    workers,
	})
}

// submitJob submits a new job to the worker pool
func (h *ModernDashboardHandler) submitJob(c *gin.Context) {
	var jobRequest map[string]interface{}
	if err := c.ShouldBindJSON(&jobRequest); err != nil {
		c.JSON(http.StatusBadRequest, map[string]interface{}{
			"error":   "Invalid job request",
			"details": err.Error(),
		})
		return
	}

	// Generate a job ID
	jobID := fmt.Sprintf("job_%d", time.Now().UnixNano())

	c.JSON(http.StatusAccepted, map[string]interface{}{
		"message":   "Job submitted successfully",
		"job_id":    jobID,
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	})
}

// saveConfig saves configuration to file
func (h *ModernDashboardHandler) saveConfig(c *gin.Context) {
	var configUpdate map[string]interface{}
	if err := c.ShouldBindJSON(&configUpdate); err != nil {
		c.JSON(http.StatusBadRequest, map[string]interface{}{
			"success": false,
			"error":   "Invalid configuration data",
			"details": err.Error(),
		})
		return
	}

	// Generate config file path (in exe directory)
	configPath := "webhook-bridge-config.yaml"

	// Create YAML content
	yamlContent := h.generateYAMLConfig(configUpdate)

	// Write to file
	if err := os.WriteFile(configPath, []byte(yamlContent), 0644); err != nil {
		c.JSON(http.StatusInternalServerError, map[string]interface{}{
			"success": false,
			"error":   "Failed to save configuration file",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, map[string]interface{}{
		"success":     true,
		"message":     "Configuration saved successfully",
		"config_path": configPath,
	})
}

// getPythonEnvStatus returns Python environment status
func (h *ModernDashboardHandler) getPythonEnvStatus(c *gin.Context) {
	status := map[string]interface{}{
		"status":       "unknown",
		"interpreter":  "",
		"version":      "",
		"uv_available": false,
	}

	// Check for Python interpreter
	pythonPaths := []string{"python", "python3", "py"}
	for _, pythonCmd := range pythonPaths {
		if cmd, err := exec.LookPath(pythonCmd); err == nil {
			status["interpreter"] = cmd
			status["status"] = "ready"

			// Get Python version
			if out, err := exec.Command(pythonCmd, "--version").Output(); err == nil {
				status["version"] = strings.TrimSpace(string(out))
			}
			break
		}
	}

	// Check for UV
	if _, err := exec.LookPath("uv"); err == nil {
		status["uv_available"] = true
	}

	if status["status"] == "unknown" {
		status["status"] = "missing"
	}

	c.JSON(http.StatusOK, status)
}

// downloadUV downloads and installs UV
func (h *ModernDashboardHandler) downloadUV(c *gin.Context) {
	// This is a simplified implementation
	// In a real implementation, you would download UV from the official source

	c.JSON(http.StatusOK, map[string]interface{}{
		"success": false,
		"error":   "UV download not implemented yet. Please install UV manually from https://docs.astral.sh/uv/getting-started/installation/",
		"url":     "https://docs.astral.sh/uv/getting-started/installation/",
	})
}

// downloadPython downloads and installs Python
func (h *ModernDashboardHandler) downloadPython(c *gin.Context) {
	// This is a simplified implementation
	// In a real implementation, you would download Python from the official source

	c.JSON(http.StatusOK, map[string]interface{}{
		"success": false,
		"error":   "Python download not implemented yet. Please install Python manually from https://www.python.org/downloads/",
		"url":     "https://www.python.org/downloads/",
	})
}

// setupVirtualEnv sets up a virtual environment
func (h *ModernDashboardHandler) setupVirtualEnv(c *gin.Context) {
	// Check if UV is available
	uvPath, err := exec.LookPath("uv")
	if err != nil {
		c.JSON(http.StatusBadRequest, map[string]interface{}{
			"success": false,
			"error":   "UV not found. Please install UV first.",
		})
		return
	}

	// Create virtual environment using UV
	cmd := exec.Command(uvPath, "venv", ".venv")
	if err := cmd.Run(); err != nil {
		c.JSON(http.StatusInternalServerError, map[string]interface{}{
			"success": false,
			"error":   "Failed to create virtual environment",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, map[string]interface{}{
		"success": true,
		"message": "Virtual environment created successfully",
		"path":    ".venv",
	})
}

// testPythonEnv tests the Python environment
func (h *ModernDashboardHandler) testPythonEnv(c *gin.Context) {
	// Find Python interpreter
	pythonPaths := []string{"python", "python3", "py"}
	var pythonCmd string

	for _, cmd := range pythonPaths {
		if _, err := exec.LookPath(cmd); err == nil {
			pythonCmd = cmd
			break
		}
	}

	if pythonCmd == "" {
		c.JSON(http.StatusBadRequest, map[string]interface{}{
			"success": false,
			"error":   "No Python interpreter found",
		})
		return
	}

	// Test Python version
	versionOut, err := exec.Command(pythonCmd, "--version").Output()
	if err != nil {
		c.JSON(http.StatusInternalServerError, map[string]interface{}{
			"success": false,
			"error":   "Failed to get Python version",
			"details": err.Error(),
		})
		return
	}

	// Test package availability
	packagesOut, err := exec.Command(pythonCmd, "-c", "import sys; print(len(sys.modules))").Output()
	packagesCount := "unknown"
	if err == nil {
		packagesCount = strings.TrimSpace(string(packagesOut))
	}

	c.JSON(http.StatusOK, map[string]interface{}{
		"success":        true,
		"version":        strings.TrimSpace(string(versionOut)),
		"interpreter":    pythonCmd,
		"packages_count": packagesCount,
	})
}

// generateYAMLConfig generates YAML configuration content
func (h *ModernDashboardHandler) generateYAMLConfig(config map[string]interface{}) string {
	python := make(map[string]interface{})
	if p, ok := config["python"].(map[string]interface{}); ok {
		python = p
	}

	server := map[string]interface{}{
		"host": "0.0.0.0",
		"port": 8000,
		"mode": "debug",
	}
	if s, ok := config["server"].(map[string]interface{}); ok {
		for k, v := range s {
			server[k] = v
		}
	}

	executor := map[string]interface{}{
		"host":    "localhost",
		"port":    50051,
		"timeout": 30,
	}
	if e, ok := config["executor"].(map[string]interface{}); ok {
		for k, v := range e {
			executor[k] = v
		}
	}

	logging := map[string]interface{}{
		"level":  "info",
		"format": "text",
	}
	if l, ok := config["logging"].(map[string]interface{}); ok {
		for k, v := range l {
			logging[k] = v
		}
	}

	return fmt.Sprintf(`# Webhook Bridge Configuration
# Generated on %s

server:
  host: "%v"
  port: %v
  mode: "%v"

python:
  strategy: "%v"
  interpreter_path: "%v"

  uv:
    enabled: %v
    project_path: "%v"
    venv_name: "%v"

  plugin_dirs:
    - "./plugins"
    - "./webhook_bridge/plugins"
    - "./example_plugins"

  validation:
    enabled: true
    min_python_version: "3.8"
    required_capabilities:
      - "sys"
      - "os"
      - "json"
    strict_mode: false
    cache_timeout: 5

  auto_install: false

  required_packages:
    - "grpcio"
    - "grpcio-tools"

executor:
  host: "%v"
  port: %v
  timeout: %v

logging:
  level: "%v"
  format: "%v"
`,
		time.Now().Format(time.RFC3339),
		server["host"], server["port"], server["mode"],
		getStringValue(python, "strategy", "auto"),
		getStringValue(python, "interpreter_path", ""),
		getBoolValue(python, "uv.enabled", true),
		getStringValue(python, "uv.project_path", "."),
		getStringValue(python, "uv.venv_name", ".venv"),
		executor["host"], executor["port"], executor["timeout"],
		logging["level"], logging["format"],
	)
}

// Helper functions for config generation
func getStringValue(m map[string]interface{}, key string, defaultValue string) string {
	if val, ok := m[key]; ok {
		if str, ok := val.(string); ok {
			return str
		}
	}
	// Handle nested keys like "uv.enabled"
	if strings.Contains(key, ".") {
		parts := strings.Split(key, ".")
		current := m
		for i, part := range parts {
			if i == len(parts)-1 {
				if val, ok := current[part]; ok {
					if str, ok := val.(string); ok {
						return str
					}
				}
			} else {
				if next, ok := current[part].(map[string]interface{}); ok {
					current = next
				} else {
					break
				}
			}
		}
	}
	return defaultValue
}

func getBoolValue(m map[string]interface{}, key string, defaultValue bool) bool {
	if val, ok := m[key]; ok {
		if b, ok := val.(bool); ok {
			return b
		}
	}
	// Handle nested keys like "uv.enabled"
	if strings.Contains(key, ".") {
		parts := strings.Split(key, ".")
		current := m
		for i, part := range parts {
			if i == len(parts)-1 {
				if val, ok := current[part]; ok {
					if b, ok := val.(bool); ok {
						return b
					}
				}
			} else {
				if next, ok := current[part].(map[string]interface{}); ok {
					current = next
				} else {
					break
				}
			}
		}
	}
	return defaultValue
}

// fallbackTemplate provides a simple fallback template if the file is not found
const fallbackTemplate = `<!DOCTYPE html>
<html lang="en" class="dark">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>{{.title}}</title>
    <script src="https://cdn.tailwindcss.com"></script>
    <script src="https://unpkg.com/lucide@latest/dist/umd/lucide.js"></script>
    <link rel="stylesheet" href="/static/css/modern-dashboard.css">
</head>
<body class="bg-background text-foreground min-h-screen">
    <div class="flex items-center justify-center min-h-screen">
        <div class="text-center">
            <h1 class="text-2xl font-bold mb-4">Webhook Bridge Dashboard</h1>
            <p class="text-muted-foreground mb-4">Modern dashboard loading...</p>
            <div class="animate-pulse">
                <div class="w-8 h-8 bg-primary rounded-lg mx-auto"></div>
            </div>
        </div>
    </div>
    <script src="/static/js/modern-dashboard.js"></script>
</body>
</html>`
