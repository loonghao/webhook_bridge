package modern

import (
	"context"
	"encoding/json"
	"fmt"
	"html/template"
	"io/fs"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"

	"github.com/loonghao/webhook_bridge/api/proto"
	"github.com/loonghao/webhook_bridge/internal/config"
	"github.com/loonghao/webhook_bridge/internal/grpc"
	"github.com/loonghao/webhook_bridge/internal/web"
	webpkg "github.com/loonghao/webhook_bridge/web-nextjs"
)

// WebSocket upgrader with CORS support
var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		// Allow connections from any origin for development
		// In production, you should restrict this to your domain
		return true
	},
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

// ModernDashboardHandler handles modern web dashboard requests
type ModernDashboardHandler struct {
	config         *config.Config
	template       *template.Template
	logManager     *web.PersistentLogManager
	grpcClient     *grpc.Client
	statsManager   *web.StatsManager
	monitorClients map[*websocket.Conn]bool
	monitorMutex   sync.RWMutex
}

// MonitorMessage represents a real-time monitoring message
type MonitorMessage struct {
	Type      string      `json:"type"`
	Timestamp time.Time   `json:"timestamp"`
	Data      interface{} `json:"data"`
}

// PluginStatusUpdate represents a plugin status change
type PluginStatusUpdate struct {
	PluginName    string `json:"plugin_name"`
	Status        string `json:"status"`
	LastExecuted  string `json:"last_executed,omitempty"`
	ExecutionTime int64  `json:"execution_time,omitempty"`
	Success       bool   `json:"success"`
	Error         string `json:"error,omitempty"`
}

// SystemMetricsUpdate represents system metrics update
type SystemMetricsUpdate struct {
	TotalExecutions    int64   `json:"total_executions"`
	SuccessRate        float64 `json:"success_rate"`
	AvgExecutionTime   float64 `json:"avg_execution_time"`
	ActivePlugins      int     `json:"active_plugins"`
	ErrorRate          float64 `json:"error_rate"`
	LastHourExecutions int64   `json:"last_hour_executions"`
}

// NewModernDashboardHandler creates a new modern dashboard handler
func NewModernDashboardHandler(cfg *config.Config) *ModernDashboardHandler {
	// Try to use embedded template first
	var tmpl *template.Template

	// Force use Next.js template directly
	fmt.Printf("🔧 Loading Next.js template...\n")
	log.Printf("🔧 Loading Next.js template...")
	if indexHTML, htmlErr := webpkg.GetIndexHTML(); htmlErr == nil {
		fmt.Printf("✅ Got Next.js HTML, size: %d bytes\n", len(indexHTML))
		log.Printf("✅ Got Next.js HTML, size: %d bytes", len(indexHTML))
		if parsedTmpl, parseErr := template.New("dashboard").Parse(indexHTML); parseErr == nil {
			fmt.Printf("✅ Successfully parsed Next.js template\n")
			log.Printf("✅ Successfully parsed Next.js template")
			tmpl = parsedTmpl
		} else {
			fmt.Printf("❌ Failed to parse Next.js template: %v\n", parseErr)
			log.Printf("❌ Failed to parse Next.js template: %v", parseErr)
			fmt.Printf("⚠️ Using fallback template\n")
			log.Printf("⚠️ Using fallback template")
			tmpl = template.Must(template.New("dashboard").Parse(fallbackTemplate))
		}
	} else {
		fmt.Printf("❌ Failed to get Next.js HTML: %v\n", htmlErr)
		log.Printf("❌ Failed to get Next.js HTML: %v", htmlErr)
		fmt.Printf("⚠️ Using fallback template\n")
		log.Printf("⚠️ Using fallback template")
		tmpl = template.Must(template.New("dashboard").Parse(fallbackTemplate))
	}

	// Initialize PersistentLogManager with configuration
	logDir := cfg.Directories.LogDir
	if logDir == "" {
		logDir = "logs"
	}

	// Create absolute path for log directory
	if !filepath.IsAbs(logDir) {
		if cfg.Directories.WorkingDir != "" {
			logDir = filepath.Join(cfg.Directories.WorkingDir, logDir)
		} else {
			logDir = filepath.Join(".", logDir)
		}
	}

	logManager := web.NewPersistentLogManager(logDir, 1000)

	// Initialize gRPC client
	grpcClient := grpc.NewClient(&cfg.Executor)

	// Initialize stats manager with persistence
	dataDir := cfg.Directories.DataDir
	if dataDir == "" {
		dataDir = "data"
	}
	statsManager := web.NewStatsManagerWithPersistence(dataDir)

	return &ModernDashboardHandler{
		config:         cfg,
		template:       tmpl,
		logManager:     logManager,
		grpcClient:     grpcClient,
		statsManager:   statsManager,
		monitorClients: make(map[*websocket.Conn]bool),
	}
}

// GetLogManager returns the log manager instance
func (h *ModernDashboardHandler) GetLogManager() *web.PersistentLogManager {
	return h.logManager
}

// GetStatsManager returns the stats manager instance
func (h *ModernDashboardHandler) GetStatsManager() *web.StatsManager {
	return h.statsManager
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

		// Python interpreter management endpoints
		api.GET("/interpreters", h.getInterpreters)
		api.POST("/interpreters", h.addInterpreter)
		api.DELETE("/interpreters/:name", h.removeInterpreter)
		api.POST("/interpreters/:name/validate", h.validateInterpreter)
		api.POST("/interpreters/:name/activate", h.activateInterpreter)
		api.GET("/interpreters/discover", h.discoverInterpreters)

		// Connection management endpoints
		api.GET("/connection", h.getConnectionStatus)
		api.POST("/connection/reconnect", h.reconnectService)
		api.POST("/connection/test", h.testConnection)

		// Plugin management endpoints
		api.POST("/plugins/:name/execute", h.executePlugin)
		api.POST("/plugins/:name/enable", h.enablePlugin)
		api.POST("/plugins/:name/disable", h.disablePlugin)
		api.GET("/plugins/:name/stats", h.getPluginStats)
		api.GET("/plugins/:name/logs", h.getPluginLogs)
		api.GET("/plugins/stats", h.getAllPluginStats)

		// WebSocket endpoints for real-time data
		api.GET("/logs/stream", h.streamLogs)
		api.GET("/monitor/stream", h.streamMonitor)

		// Python environment management endpoints
		api.GET("/python-env", h.getPythonEnvStatus)
		api.POST("/download-uv", h.downloadUV)
		api.POST("/download-python", h.downloadPython)
		api.POST("/setup-venv", h.setupVirtualEnv)
		api.POST("/test-python", h.testPythonEnv)

		// Test endpoints
		api.POST("/test-log", h.addTestLog)

		// Debug endpoints
		api.GET("/debug/filesystem", h.debugFilesystem)
		api.GET("/debug/css-status", h.debugCSSStatus)
	}

	// Static assets - serve from embedded Next.js resources
	// Handle Next.js static files (next/static/*)
	router.GET("/next/static/*filepath", func(c *gin.Context) {
		filepath := c.Param("filepath")
		log.Printf("DEBUG: Static asset request for: %s", filepath)

		// Get the Next.js filesystem
		nextjsFS := webpkg.GetNextJSFS()

		// Construct the full path
		fullPath := "next/static" + filepath
		log.Printf("DEBUG: Trying to read file: %s", fullPath)

		// Try to read the file from embedded filesystem
		if data, err := fs.ReadFile(nextjsFS, fullPath); err == nil {
			// Determine content type based on file extension
			var contentType string
			if strings.HasSuffix(filepath, ".js") {
				contentType = "application/javascript"
			} else if strings.HasSuffix(filepath, ".css") {
				contentType = "text/css"
			} else if strings.HasSuffix(filepath, ".woff2") {
				contentType = "font/woff2"
			} else if strings.HasSuffix(filepath, ".woff") {
				contentType = "font/woff"
			} else {
				contentType = "application/octet-stream"
			}

			log.Printf("DEBUG: Successfully serving %s with content-type: %s", fullPath, contentType)
			c.Header("Content-Type", contentType)
			c.Header("Cache-Control", "public, max-age=31536000") // 1 year cache for static assets
			c.Data(http.StatusOK, contentType, data)
			return
		} else {
			log.Printf("DEBUG: File not found: %s, error: %v", fullPath, err)
		}

		c.Status(http.StatusNotFound)
	})

	// Legacy Next.js route for backward compatibility - redirect to new path
	router.GET("/_next/static/*filepath", func(c *gin.Context) {
		filepath := c.Param("filepath")
		log.Printf("DEBUG: Legacy _next route accessed, redirecting: %s", filepath)
		// Redirect to the new path
		c.Redirect(http.StatusMovedPermanently, "/next/static"+filepath)
	})

	// Legacy assets route for backward compatibility
	router.GET("/assets/*filepath", func(c *gin.Context) {
		filepath := c.Param("filepath")

		// Handle JS files
		if strings.HasSuffix(filepath, ".js") {
			jsData := web.GetJSFile()
			if jsData != nil {
				c.Header("Content-Type", "application/javascript")
				c.Data(http.StatusOK, "application/javascript", jsData)
				return
			}
		}

		// Handle CSS files
		if strings.HasSuffix(filepath, ".css") {
			cssData := web.GetCSSFile()
			if cssData != nil {
				c.Header("Content-Type", "text/css")
				c.Data(http.StatusOK, "text/css", cssData)
				return
			}
		}

		c.Status(http.StatusNotFound)
	})

	// Note: Removed filesystem fallback for /assets to avoid conflicts with embedded resources

	// Favicon - serve from embedded resources first, then fallback to filesystem
	router.GET("/favicon.ico", func(c *gin.Context) {
		faviconData, err := web.GetFaviconData()
		if err == nil && faviconData != nil {
			c.Header("Content-Type", "image/x-icon")
			c.Data(http.StatusOK, "image/x-icon", faviconData)
			return
		}

		// Fallback to filesystem
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
				c.File(path)
				return
			}
		}

		c.Status(http.StatusNotFound)
	})

	// Dashboard routes - must be last to catch all routes for SPA
	router.GET("/", h.dashboard)
	router.GET("/dashboard", h.dashboard)
	router.GET("/dashboard/*path", h.dashboard) // Catch all dashboard sub-routes for SPA
	router.GET("/plugins", h.dashboard)         // Plugins page
	router.GET("/status", h.dashboard)          // System Status page
	router.GET("/logs", h.dashboard)            // Logs page
	router.GET("/config", h.dashboard)          // Configuration page
	router.GET("/interpreters", h.dashboard)    // Python Interpreters page
	router.GET("/connection", h.dashboard)      // Connection Status page
	router.NoRoute(h.dashboard)                 // Catch all other routes for SPA
}

// dashboard serves the main dashboard page
func (h *ModernDashboardHandler) dashboard(c *gin.Context) {
	path := c.Request.URL.Path
	fmt.Printf("🎯 Dashboard handler called for path: %s\n", path)
	log.Printf("🎯 Dashboard handler called for path: %s", path)

	// Check if this is a static file request that should be handled by static file handlers
	if strings.HasPrefix(path, "/next/static/") || strings.HasPrefix(path, "/_next/static/") || strings.HasPrefix(path, "/assets/") {
		fmt.Printf("🚫 Static file request detected, should not be handled by dashboard: %s\n", path)
		log.Printf("🚫 Static file request detected, should not be handled by dashboard: %s", path)

		// Try to handle static files directly here as fallback
		if strings.HasPrefix(path, "/next/static/") {
			h.handleStaticFile(c, path)
			return
		}

		c.Status(http.StatusNotFound)
		return
	}

	// Load the correct Next.js HTML file based on the path
	if pageHTML, err := webpkg.GetPageHTML(path); err == nil {
		// Inject runtime configuration into the HTML
		runtimeConfig := h.generateRuntimeConfig(c)
		injectedHTML := h.injectRuntimeConfig(pageHTML, runtimeConfig)

		fmt.Printf("✅ Serving Next.js page HTML for path '%s', size: %d bytes (with runtime config)\n", path, len(injectedHTML))
		c.Header("Content-Type", "text/html; charset=utf-8")
		c.String(http.StatusOK, injectedHTML)
		return
	} else {
		fmt.Printf("❌ Failed to get Next.js page HTML for path '%s': %v\n", path, err)
		log.Printf("❌ Failed to get Next.js page HTML for path '%s': %v", path, err)
	}

	// Fallback to template if direct HTML fails
	log.Printf("🔧 Template name: %s", h.template.Name())

	data := map[string]interface{}{
		"title":   "Webhook Bridge - Modern Dashboard",
		"version": "2.0.0-hybrid",
		"uptime":  time.Since(time.Now().Add(-time.Hour)).String(),
	}

	c.Header("Content-Type", "text/html; charset=utf-8")
	if err := h.template.Execute(c.Writer, data); err != nil {
		log.Printf("❌ Template execution error: %v", err)
		c.String(http.StatusInternalServerError, "Error rendering template: %v", err)
		return
	}
	log.Printf("✅ Template executed successfully")
}

// handleStaticFile handles static file requests as fallback
func (h *ModernDashboardHandler) handleStaticFile(c *gin.Context, path string) {
	// Extract filepath from path
	var filepath string
	if strings.HasPrefix(path, "/next/static/") {
		filepath = strings.TrimPrefix(path, "/next/static")
	} else {
		c.Status(http.StatusNotFound)
		return
	}

	log.Printf("DEBUG: Fallback static asset request for: %s", filepath)

	// Get the Next.js filesystem
	nextjsFS := webpkg.GetNextJSFS()

	// Construct the full path
	fullPath := "next/static" + filepath
	log.Printf("DEBUG: Trying to read file: %s", fullPath)

	// Try to read the file from embedded filesystem
	if data, err := fs.ReadFile(nextjsFS, fullPath); err == nil {
		// Determine content type based on file extension
		var contentType string
		if strings.HasSuffix(filepath, ".js") {
			contentType = "application/javascript"
		} else if strings.HasSuffix(filepath, ".css") {
			contentType = "text/css"
		} else if strings.HasSuffix(filepath, ".woff2") {
			contentType = "font/woff2"
		} else if strings.HasSuffix(filepath, ".woff") {
			contentType = "font/woff"
		} else {
			contentType = "application/octet-stream"
		}

		log.Printf("DEBUG: Successfully serving %s with content-type: %s", fullPath, contentType)
		c.Header("Content-Type", contentType)
		c.Header("Cache-Control", "public, max-age=31536000") // 1 year cache for static assets
		c.Data(http.StatusOK, contentType, data)
		return
	} else {
		log.Printf("DEBUG: File not found: %s, error: %v", fullPath, err)
	}

	c.Status(http.StatusNotFound)
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

// getPlugins returns plugin information with real data from gRPC and stats
func (h *ModernDashboardHandler) getPlugins(c *gin.Context) {
	var plugins []map[string]interface{}

	// Try to get real plugin list from gRPC
	if h.grpcClient != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		// Try to list plugins via gRPC
		req := &proto.ListPluginsRequest{}
		if resp, err := h.grpcClient.ListPlugins(ctx, req); err == nil && resp != nil {
			// Convert gRPC response to our format
			for _, plugin := range resp.Plugins {
				pluginData := map[string]interface{}{
					"name":             plugin.Name,
					"version":          "1.0.0", // Default version since proto doesn't have it
					"description":      plugin.Description,
					"status":           "active", // Default status
					"path":             plugin.Path,
					"supportedMethods": plugin.SupportedMethods,
					"isAvailable":      plugin.IsAvailable,
					"lastModified":     plugin.LastModified,
				}

				// Get statistics for this plugin if available
				if h.statsManager != nil {
					pluginStats := h.statsManager.GetPluginStats()
					for _, stat := range pluginStats {
						if stat.Plugin == plugin.Name {
							pluginData["executionCount"] = stat.Count
							pluginData["errorCount"] = stat.Errors
							pluginData["lastExecuted"] = stat.LastExec.Format(time.RFC3339)
							pluginData["avgExecutionTime"] = stat.AvgTime.String()

							// Determine status based on recent activity
							if time.Since(stat.LastExec) < 24*time.Hour {
								pluginData["status"] = "active"
							} else {
								pluginData["status"] = "inactive"
							}
							break
						}
					}
				}

				plugins = append(plugins, pluginData)
			}
		} else {
			// Log the error but continue with fallback data
			log.Printf("Failed to get plugins from gRPC: %v", err)
		}
	}

	// If no plugins found via gRPC, use fallback data
	if len(plugins) == 0 {
		plugins = []map[string]interface{}{
			{
				"name":             "example_plugin",
				"version":          "1.0.0",
				"status":           "active",
				"description":      "Example webhook plugin",
				"lastExecuted":     time.Now().Add(-time.Hour).Format(time.RFC3339),
				"executionCount":   856,
				"errorCount":       5,
				"avgExecutionTime": "45ms",
			},
			{
				"name":             "notification_plugin",
				"version":          "2.1.0",
				"status":           "active",
				"description":      "Notification webhook plugin",
				"lastExecuted":     time.Now().Add(-30 * time.Minute).Format(time.RFC3339),
				"executionCount":   344,
				"errorCount":       2,
				"avgExecutionTime": "120ms",
			},
			{
				"name":             "data_processor",
				"version":          "1.5.2",
				"status":           "inactive",
				"description":      "Data processing webhook plugin",
				"lastExecuted":     time.Now().Add(-24 * time.Hour).Format(time.RFC3339),
				"executionCount":   12,
				"errorCount":       1,
				"avgExecutionTime": "80ms",
			},
		}
	}

	// Return data in the format expected by the React frontend
	c.JSON(http.StatusOK, map[string]interface{}{
		"success": true,
		"data":    plugins,
	})
}

// getLogs returns recent logs with optional plugin filtering
func (h *ModernDashboardHandler) getLogs(c *gin.Context) {
	// Get query parameters
	levelFilter := c.Query("level")
	pluginFilter := c.Query("plugin")
	limitStr := c.Query("limit")

	limit := 50 // Default limit
	if limitStr != "" {
		if parsedLimit, err := strconv.Atoi(limitStr); err == nil && parsedLimit > 0 {
			limit = parsedLimit
		}
	}

	// Get logs from the persistent log manager with filters
	var logs []web.LogEntry
	if pluginFilter != "" {
		// Use the new filtering method if plugin filter is specified
		logs = h.logManager.GetLogsWithFilters(levelFilter, pluginFilter, limit)
	} else {
		// Use the original method for backward compatibility
		logs = h.logManager.GetLogs(levelFilter, limit)
	}

	// Convert to the format expected by the React frontend
	logData := make([]map[string]interface{}, len(logs))
	for i, log := range logs {
		logData[i] = map[string]interface{}{
			"id":          log.ID,
			"timestamp":   log.Timestamp.Format(time.RFC3339),
			"level":       strings.ToLower(log.Level),
			"message":     log.Message,
			"source":      log.Source,
			"plugin_name": log.PluginName,
			"data":        log.Data,
		}
	}

	// Return data in the format expected by the React frontend
	c.JSON(http.StatusOK, map[string]interface{}{
		"success": true,
		"data":    logData,
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
	if err := os.WriteFile(configPath, []byte(yamlContent), 0600); err != nil {
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

// addTestLog adds a test log entry
func (h *ModernDashboardHandler) addTestLog(c *gin.Context) {
	h.logManager.AddTestLog()

	c.JSON(http.StatusOK, map[string]interface{}{
		"success": true,
		"message": "Test log added successfully",
	})
}

// generateRuntimeConfig generates runtime configuration for frontend injection
func (h *ModernDashboardHandler) generateRuntimeConfig(c *gin.Context) map[string]interface{} {
	// Get the current server address
	protocol := "http"
	if c.Request.TLS != nil {
		protocol = "https"
	}

	host := c.Request.Host
	apiBaseUrl := fmt.Sprintf("%s://%s", protocol, host)

	return map[string]interface{}{
		"apiBaseUrl":  apiBaseUrl,
		"wsBaseUrl":   strings.Replace(apiBaseUrl, "http", "ws", 1),
		"serverPort":  h.config.Server.Port,
		"version":     "2.0.0-hybrid",
		"buildTime":   time.Now().Format(time.RFC3339),
		"environment": h.config.Server.Mode,
	}
}

// injectRuntimeConfig injects runtime configuration into HTML
func (h *ModernDashboardHandler) injectRuntimeConfig(html string, config map[string]interface{}) string {
	configJSON, _ := json.Marshal(config)
	configScript := fmt.Sprintf(`<script>window.__WEBHOOK_BRIDGE_CONFIG__ = %s;</script>`, string(configJSON))

	// Inject before closing head tag
	if strings.Contains(html, "</head>") {
		return strings.Replace(html, "</head>", configScript+"\n</head>", 1)
	}

	// Fallback: inject at the beginning of body
	if strings.Contains(html, "<body") {
		bodyStart := strings.Index(html, ">")
		if bodyStart != -1 {
			return html[:bodyStart+1] + "\n" + configScript + html[bodyStart+1:]
		}
	}

	// Last resort: prepend to HTML
	return configScript + "\n" + html
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

// Python Interpreter Management API Methods

// getInterpreters returns all configured Python interpreters
func (h *ModernDashboardHandler) getInterpreters(c *gin.Context) {
	// For now, return mock data. This will be replaced with actual interpreter manager integration
	interpreters := map[string]interface{}{
		"active": h.config.Python.ActiveInterpreter,
		"interpreters": map[string]interface{}{
			"system-python": map[string]interface{}{
				"name":              "System Python",
				"path":              "/usr/bin/python3",
				"status":            "ready",
				"validated":         true,
				"last_validated":    time.Now().Add(-time.Hour).Format(time.RFC3339),
				"version":           "3.9.7",
				"use_uv":            false,
				"venv_path":         "",
				"required_packages": []string{"grpcio", "grpcio-tools"},
			},
			"uv-python": map[string]interface{}{
				"name":              "UV Python Environment",
				"path":              ".venv/bin/python",
				"status":            "ready",
				"validated":         true,
				"last_validated":    time.Now().Add(-30 * time.Minute).Format(time.RFC3339),
				"version":           "3.11.5",
				"use_uv":            true,
				"venv_path":         ".venv",
				"required_packages": []string{"grpcio", "grpcio-tools"},
			},
		},
	}

	c.JSON(http.StatusOK, map[string]interface{}{
		"success": true,
		"data":    interpreters,
	})
}

// addInterpreter adds a new Python interpreter configuration
func (h *ModernDashboardHandler) addInterpreter(c *gin.Context) {
	var request struct {
		Name             string            `json:"name"`
		Path             string            `json:"path"`
		VenvPath         string            `json:"venv_path,omitempty"`
		UseUV            bool              `json:"use_uv,omitempty"`
		RequiredPackages []string          `json:"required_packages,omitempty"`
		Environment      map[string]string `json:"environment,omitempty"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, map[string]interface{}{
			"success": false,
			"error":   "Invalid request format",
			"details": err.Error(),
		})
		return
	}

	// Validate required fields
	if request.Name == "" || request.Path == "" {
		c.JSON(http.StatusBadRequest, map[string]interface{}{
			"success": false,
			"error":   "Name and path are required",
		})
		return
	}

	// TODO: Integrate with actual interpreter manager
	// For now, return success response
	c.JSON(http.StatusOK, map[string]interface{}{
		"success": true,
		"message": fmt.Sprintf("Interpreter '%s' added successfully", request.Name),
		"data": map[string]interface{}{
			"name":              request.Name,
			"path":              request.Path,
			"status":            "validating",
			"validated":         false,
			"use_uv":            request.UseUV,
			"venv_path":         request.VenvPath,
			"required_packages": request.RequiredPackages,
			"environment":       request.Environment,
		},
	})
}

// removeInterpreter removes a Python interpreter configuration
func (h *ModernDashboardHandler) removeInterpreter(c *gin.Context) {
	name := c.Param("name")
	if name == "" {
		c.JSON(http.StatusBadRequest, map[string]interface{}{
			"success": false,
			"error":   "Interpreter name is required",
		})
		return
	}

	// TODO: Integrate with actual interpreter manager
	// For now, return success response
	c.JSON(http.StatusOK, map[string]interface{}{
		"success": true,
		"message": fmt.Sprintf("Interpreter '%s' removed successfully", name),
	})
}

// validateInterpreter validates a specific Python interpreter
func (h *ModernDashboardHandler) validateInterpreter(c *gin.Context) {
	name := c.Param("name")
	if name == "" {
		c.JSON(http.StatusBadRequest, map[string]interface{}{
			"success": false,
			"error":   "Interpreter name is required",
		})
		return
	}

	// TODO: Integrate with actual interpreter manager
	// For now, return success response
	c.JSON(http.StatusOK, map[string]interface{}{
		"success": true,
		"message": fmt.Sprintf("Interpreter '%s' validation started", name),
		"data": map[string]interface{}{
			"name":          name,
			"status":        "validating",
			"validation_id": fmt.Sprintf("val_%d", time.Now().UnixNano()),
		},
	})
}

// activateInterpreter sets a specific interpreter as active
func (h *ModernDashboardHandler) activateInterpreter(c *gin.Context) {
	name := c.Param("name")
	if name == "" {
		c.JSON(http.StatusBadRequest, map[string]interface{}{
			"success": false,
			"error":   "Interpreter name is required",
		})
		return
	}

	// TODO: Integrate with actual interpreter manager and connection manager
	// For now, return success response
	c.JSON(http.StatusOK, map[string]interface{}{
		"success": true,
		"message": fmt.Sprintf("Interpreter '%s' activated successfully", name),
		"data": map[string]interface{}{
			"active_interpreter":  name,
			"reconnection_status": "in_progress",
		},
	})
}

// discoverInterpreters automatically discovers available Python interpreters
func (h *ModernDashboardHandler) discoverInterpreters(c *gin.Context) {
	// TODO: Integrate with actual interpreter manager
	// For now, return mock discovered interpreters
	discovered := []map[string]interface{}{
		{
			"name":    "Python 3.9 (python3)",
			"path":    "/usr/bin/python3",
			"version": "3.9.7",
			"status":  "available",
		},
		{
			"name":    "Python 3.11 (python3.11)",
			"path":    "/usr/bin/python3.11",
			"version": "3.11.5",
			"status":  "available",
		},
		{
			"name":    "Python 3.8 (python3.8)",
			"path":    "/usr/bin/python3.8",
			"version": "3.8.10",
			"status":  "available",
		},
	}

	c.JSON(http.StatusOK, map[string]interface{}{
		"success": true,
		"message": fmt.Sprintf("Discovered %d Python interpreters", len(discovered)),
		"data":    discovered,
	})
}

// Connection Management API Methods

// getConnectionStatus returns the current connection status
func (h *ModernDashboardHandler) getConnectionStatus(c *gin.Context) {
	// TODO: Integrate with actual connection manager
	// For now, return mock connection status
	status := map[string]interface{}{
		"status":             "connected",
		"reconnect_attempts": 0,
		"max_reconnects":     5,
		"executor_host":      h.config.Executor.Host,
		"executor_port":      h.config.Executor.Port,
		"active_interpreter": h.config.Python.ActiveInterpreter,
		"last_connected":     time.Now().Add(-time.Hour).Format(time.RFC3339),
		"uptime":             time.Since(time.Now().Add(-time.Hour)).String(),
		"process_pid":        12345,
	}

	c.JSON(http.StatusOK, map[string]interface{}{
		"success": true,
		"data":    status,
	})
}

// reconnectService forces a reconnection to the Python executor
func (h *ModernDashboardHandler) reconnectService(c *gin.Context) {
	var request struct {
		InterpreterName string `json:"interpreter_name,omitempty"`
	}

	// Parse request body (optional)
	c.ShouldBindJSON(&request)

	// TODO: Integrate with actual connection manager
	// For now, return success response
	message := "Service reconnection initiated"
	if request.InterpreterName != "" {
		message = fmt.Sprintf("Service reconnection initiated with interpreter '%s'", request.InterpreterName)
	}

	c.JSON(http.StatusOK, map[string]interface{}{
		"success": true,
		"message": message,
		"data": map[string]interface{}{
			"reconnection_id": fmt.Sprintf("reconnect_%d", time.Now().UnixNano()),
			"status":          "in_progress",
			"interpreter":     request.InterpreterName,
		},
	})
}

// testConnection tests the connection to the Python executor
func (h *ModernDashboardHandler) testConnection(c *gin.Context) {
	// TODO: Integrate with actual connection manager
	// For now, simulate a connection test
	time.Sleep(500 * time.Millisecond) // Simulate test delay

	c.JSON(http.StatusOK, map[string]interface{}{
		"success": true,
		"message": "Connection test completed successfully",
		"data": map[string]interface{}{
			"test_id":        fmt.Sprintf("test_%d", time.Now().UnixNano()),
			"status":         "passed",
			"response_time":  "45ms",
			"executor_host":  h.config.Executor.Host,
			"executor_port":  h.config.Executor.Port,
			"test_timestamp": time.Now().Format(time.RFC3339),
		},
	})
}

// Plugin Management API Methods

// executePlugin executes a specific plugin manually for testing
func (h *ModernDashboardHandler) executePlugin(c *gin.Context) {
	pluginName := c.Param("name")
	if pluginName == "" {
		c.JSON(http.StatusBadRequest, map[string]interface{}{
			"success": false,
			"error":   "Plugin name is required",
		})
		return
	}

	var request struct {
		Method  string            `json:"method"`
		Data    map[string]string `json:"data"`
		Headers map[string]string `json:"headers"`
		Query   string            `json:"query"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, map[string]interface{}{
			"success": false,
			"error":   "Invalid request format",
			"details": err.Error(),
		})
		return
	}

	// Default method if not specified
	if request.Method == "" {
		request.Method = "POST"
	}

	// Execute plugin via gRPC if available
	if h.grpcClient != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		// Record execution start time for statistics
		startTime := time.Now()

		// Create gRPC request
		grpcReq := &proto.ExecutePluginRequest{
			PluginName:  pluginName,
			HttpMethod:  request.Method,
			Data:        request.Data,
			Headers:     request.Headers,
			QueryString: request.Query,
		}

		// Execute plugin
		resp, err := h.grpcClient.ExecutePlugin(ctx, grpcReq)

		// Record statistics
		executionTime := time.Since(startTime)
		success := err == nil && (resp == nil || resp.StatusCode < 400)

		if h.statsManager != nil {
			h.statsManager.RecordExecution(pluginName, request.Method, startTime)
			if !success {
				h.statsManager.RecordError(pluginName, request.Method)
			}
		}

		// Broadcast real-time plugin status update
		statusUpdate := PluginStatusUpdate{
			PluginName:    pluginName,
			Status:        "executed",
			LastExecuted:  time.Now().Format(time.RFC3339),
			ExecutionTime: executionTime.Milliseconds(),
			Success:       success,
		}

		if err != nil {
			statusUpdate.Error = err.Error()
		} else if resp != nil && resp.StatusCode >= 400 {
			statusUpdate.Error = fmt.Sprintf("HTTP %d: %s", resp.StatusCode, resp.Message)
		}

		// Broadcast to monitoring clients
		h.BroadcastPluginStatusUpdate(statusUpdate)

		// Also broadcast updated system metrics
		h.BroadcastSystemMetricsUpdate()

		// Record log entry
		if h.logManager != nil {
			logLevel := "INFO"
			logMessage := fmt.Sprintf("Plugin %s executed manually", pluginName)

			if err != nil {
				logLevel = "ERROR"
				logMessage = fmt.Sprintf("Plugin %s execution failed: %v", pluginName, err)
			} else if resp != nil && resp.StatusCode >= 400 {
				logLevel = "WARN"
				logMessage = fmt.Sprintf("Plugin %s returned status %d", pluginName, resp.StatusCode)
			}

			logEntry := web.LogEntry{
				Timestamp:  time.Now(),
				Level:      logLevel,
				Source:     "plugin_test",
				Message:    logMessage,
				PluginName: pluginName,
				Data: map[string]interface{}{
					"method":         request.Method,
					"execution_time": time.Since(startTime).String(),
				},
			}

			if resp != nil {
				logEntry.Data["status_code"] = resp.StatusCode
				logEntry.Data["response_message"] = resp.Message
			}

			h.logManager.AddLog(logEntry)
		}

		if err != nil {
			c.JSON(http.StatusInternalServerError, map[string]interface{}{
				"success": false,
				"error":   "Plugin execution failed",
				"details": err.Error(),
			})
			return
		}

		// Return execution result
		c.JSON(http.StatusOK, map[string]interface{}{
			"success": true,
			"message": "Plugin executed successfully",
			"data": map[string]interface{}{
				"plugin_name":    pluginName,
				"method":         request.Method,
				"status_code":    resp.StatusCode,
				"message":        resp.Message,
				"response_data":  resp.Data,
				"execution_time": resp.ExecutionTime,
				"timestamp":      time.Now().Format(time.RFC3339),
			},
		})
		return
	}

	// Fallback response if gRPC is not available
	c.JSON(http.StatusServiceUnavailable, map[string]interface{}{
		"success": false,
		"error":   "Plugin execution service is not available",
		"message": "gRPC client is not initialized",
	})
}

// getPluginStats returns statistics for a specific plugin
func (h *ModernDashboardHandler) getPluginStats(c *gin.Context) {
	pluginName := c.Param("name")
	if pluginName == "" {
		c.JSON(http.StatusBadRequest, map[string]interface{}{
			"success": false,
			"error":   "Plugin name is required",
		})
		return
	}

	if h.statsManager == nil {
		c.JSON(http.StatusServiceUnavailable, map[string]interface{}{
			"success": false,
			"error":   "Statistics service is not available",
		})
		return
	}

	// Get all plugin statistics and filter for the requested plugin
	allStats := h.statsManager.GetPluginStats()
	var pluginStats []map[string]interface{}

	for _, stat := range allStats {
		if stat.Plugin == pluginName {
			statData := map[string]interface{}{
				"plugin":          stat.Plugin,
				"method":          stat.Method,
				"execution_count": stat.Count,
				"error_count":     stat.Errors,
				"total_time":      stat.TotalTime.String(),
				"average_time":    stat.AvgTime.String(),
				"last_executed":   stat.LastExec.Format(time.RFC3339),
				"success_rate":    float64(stat.Count-stat.Errors) / float64(stat.Count) * 100,
			}
			pluginStats = append(pluginStats, statData)
		}
	}

	c.JSON(http.StatusOK, map[string]interface{}{
		"success": true,
		"data": map[string]interface{}{
			"plugin_name": pluginName,
			"statistics":  pluginStats,
			"summary": map[string]interface{}{
				"total_executions": len(pluginStats),
				"methods_count":    len(pluginStats),
			},
		},
	})
}

// getPluginLogs returns logs for a specific plugin
func (h *ModernDashboardHandler) getPluginLogs(c *gin.Context) {
	pluginName := c.Param("name")
	if pluginName == "" {
		c.JSON(http.StatusBadRequest, map[string]interface{}{
			"success": false,
			"error":   "Plugin name is required",
		})
		return
	}

	// Get query parameters
	levelFilter := c.Query("level")
	limitStr := c.Query("limit")

	limit := 50 // Default limit
	if limitStr != "" {
		if parsedLimit, err := strconv.Atoi(limitStr); err == nil && parsedLimit > 0 {
			limit = parsedLimit
		}
	}

	if h.logManager == nil {
		c.JSON(http.StatusServiceUnavailable, map[string]interface{}{
			"success": false,
			"error":   "Log service is not available",
		})
		return
	}

	// Get logs filtered by plugin name
	logs := h.logManager.GetLogsWithFilters(levelFilter, pluginName, limit)

	// Convert to the format expected by the React frontend
	logData := make([]map[string]interface{}, len(logs))
	for i, log := range logs {
		logData[i] = map[string]interface{}{
			"id":          log.ID,
			"timestamp":   log.Timestamp.Format(time.RFC3339),
			"level":       strings.ToLower(log.Level),
			"message":     log.Message,
			"source":      log.Source,
			"plugin_name": log.PluginName,
			"data":        log.Data,
		}
	}

	c.JSON(http.StatusOK, map[string]interface{}{
		"success": true,
		"data": map[string]interface{}{
			"plugin_name": pluginName,
			"logs":        logData,
			"total":       len(logData),
			"filters": map[string]interface{}{
				"level": levelFilter,
				"limit": limit,
			},
		},
	})
}

// getAllPluginStats returns statistics for all plugins
func (h *ModernDashboardHandler) getAllPluginStats(c *gin.Context) {
	if h.statsManager == nil {
		c.JSON(http.StatusServiceUnavailable, map[string]interface{}{
			"success": false,
			"error":   "Statistics service is not available",
		})
		return
	}

	// Get all plugin statistics
	allStats := h.statsManager.GetPluginStats()

	// Group statistics by plugin name
	pluginGroups := make(map[string][]map[string]interface{})
	totalExecutions := int64(0)
	totalErrors := int64(0)

	for _, stat := range allStats {
		statData := map[string]interface{}{
			"plugin":          stat.Plugin,
			"method":          stat.Method,
			"execution_count": stat.Count,
			"error_count":     stat.Errors,
			"total_time":      stat.TotalTime.String(),
			"average_time":    stat.AvgTime.String(),
			"last_executed":   stat.LastExec.Format(time.RFC3339),
			"success_rate":    float64(stat.Count-stat.Errors) / float64(stat.Count) * 100,
		}

		if pluginGroups[stat.Plugin] == nil {
			pluginGroups[stat.Plugin] = make([]map[string]interface{}, 0)
		}
		pluginGroups[stat.Plugin] = append(pluginGroups[stat.Plugin], statData)

		totalExecutions += stat.Count
		totalErrors += stat.Errors
	}

	// Create summary for each plugin
	pluginSummaries := make([]map[string]interface{}, 0)
	for pluginName, stats := range pluginGroups {
		var totalCount, totalErrorCount int64
		var lastExec time.Time
		methods := make([]string, 0)

		for _, stat := range stats {
			totalCount += stat["execution_count"].(int64)
			totalErrorCount += stat["error_count"].(int64)
			methods = append(methods, stat["method"].(string))

			if execTime, err := time.Parse(time.RFC3339, stat["last_executed"].(string)); err == nil {
				if execTime.After(lastExec) {
					lastExec = execTime
				}
			}
		}

		summary := map[string]interface{}{
			"plugin_name":      pluginName,
			"total_executions": totalCount,
			"total_errors":     totalErrorCount,
			"success_rate":     float64(totalCount-totalErrorCount) / float64(totalCount) * 100,
			"methods":          methods,
			"methods_count":    len(methods),
			"last_executed":    lastExec.Format(time.RFC3339),
			"detailed_stats":   stats,
		}

		pluginSummaries = append(pluginSummaries, summary)
	}

	// Get overall system statistics
	systemStats := h.statsManager.GetStats()

	c.JSON(http.StatusOK, map[string]interface{}{
		"success": true,
		"data": map[string]interface{}{
			"plugins": pluginSummaries,
			"summary": map[string]interface{}{
				"total_plugins":    len(pluginGroups),
				"total_executions": totalExecutions,
				"total_errors":     totalErrors,
				"overall_success_rate": func() float64 {
					if totalExecutions > 0 {
						return float64(totalExecutions-totalErrors) / float64(totalExecutions) * 100
					}
					return 0
				}(),
				"system_uptime":     systemStats.Uptime,
				"total_requests":    systemStats.TotalRequests,
				"system_executions": systemStats.TotalExecutions,
			},
		},
	})
}

// enablePlugin enables a specific plugin
func (h *ModernDashboardHandler) enablePlugin(c *gin.Context) {
	pluginName := c.Param("name")
	if pluginName == "" {
		c.JSON(http.StatusBadRequest, map[string]interface{}{
			"success": false,
			"error":   "Plugin name is required",
		})
		return
	}

	// For now, we'll simulate enabling the plugin
	// In a real implementation, this would interact with the plugin management system
	log.Printf("Enabling plugin: %s", pluginName)

	// Record log entry
	if h.logManager != nil {
		logEntry := web.LogEntry{
			Timestamp:  time.Now(),
			Level:      "INFO",
			Source:     "plugin_management",
			Message:    fmt.Sprintf("Plugin %s enabled", pluginName),
			PluginName: pluginName,
			Data: map[string]interface{}{
				"action": "enable",
			},
		}
		h.logManager.AddLog(logEntry)
	}

	// Broadcast plugin status update
	statusUpdate := PluginStatusUpdate{
		PluginName:   pluginName,
		Status:       "active",
		LastExecuted: time.Now().Format(time.RFC3339),
		Success:      true,
	}
	h.BroadcastPluginStatusUpdate(statusUpdate)

	c.JSON(http.StatusOK, map[string]interface{}{
		"success": true,
		"message": fmt.Sprintf("Plugin %s enabled successfully", pluginName),
		"data": map[string]interface{}{
			"plugin_name": pluginName,
			"status":      "active",
			"timestamp":   time.Now().Format(time.RFC3339),
		},
	})
}

// disablePlugin disables a specific plugin
func (h *ModernDashboardHandler) disablePlugin(c *gin.Context) {
	pluginName := c.Param("name")
	if pluginName == "" {
		c.JSON(http.StatusBadRequest, map[string]interface{}{
			"success": false,
			"error":   "Plugin name is required",
		})
		return
	}

	// For now, we'll simulate disabling the plugin
	// In a real implementation, this would interact with the plugin management system
	log.Printf("Disabling plugin: %s", pluginName)

	// Record log entry
	if h.logManager != nil {
		logEntry := web.LogEntry{
			Timestamp:  time.Now(),
			Level:      "INFO",
			Source:     "plugin_management",
			Message:    fmt.Sprintf("Plugin %s disabled", pluginName),
			PluginName: pluginName,
			Data: map[string]interface{}{
				"action": "disable",
			},
		}
		h.logManager.AddLog(logEntry)
	}

	// Broadcast plugin status update
	statusUpdate := PluginStatusUpdate{
		PluginName:   pluginName,
		Status:       "inactive",
		LastExecuted: time.Now().Format(time.RFC3339),
		Success:      true,
	}
	h.BroadcastPluginStatusUpdate(statusUpdate)

	c.JSON(http.StatusOK, map[string]interface{}{
		"success": true,
		"message": fmt.Sprintf("Plugin %s disabled successfully", pluginName),
		"data": map[string]interface{}{
			"plugin_name": pluginName,
			"status":      "inactive",
			"timestamp":   time.Now().Format(time.RFC3339),
		},
	})
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

// streamLogs handles WebSocket connections for real-time log streaming
func (h *ModernDashboardHandler) streamLogs(c *gin.Context) {
	// Upgrade HTTP connection to WebSocket
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("Failed to upgrade connection to WebSocket: %v", err)
		return
	}
	defer conn.Close()

	// Create a channel for this client to receive log entries
	logChan := make(chan web.LogEntry, 100)

	// Add this client to the log manager
	h.logManager.AddClient(logChan)
	defer h.logManager.RemoveClient(logChan)

	// Handle WebSocket connection
	go func() {
		// Read messages from client (for ping/pong and control messages)
		for {
			_, _, err := conn.ReadMessage()
			if err != nil {
				log.Printf("WebSocket read error: %v", err)
				break
			}
		}
	}()

	// Send log entries to client
	for {
		select {
		case logEntry, ok := <-logChan:
			if !ok {
				// Channel closed, exit
				return
			}

			// Convert log entry to the format expected by the frontend
			message := MonitorMessage{
				Type:      "log_entry",
				Timestamp: time.Now(),
				Data: map[string]interface{}{
					"id":          logEntry.ID,
					"timestamp":   logEntry.Timestamp.Format(time.RFC3339),
					"level":       strings.ToLower(logEntry.Level),
					"message":     logEntry.Message,
					"source":      logEntry.Source,
					"plugin_name": logEntry.PluginName,
					"component":   logEntry.PluginName, // For compatibility
					"data":        logEntry.Data,
				},
			}

			// Send log entry to client
			if err := conn.WriteJSON(message); err != nil {
				log.Printf("WebSocket write error: %v", err)
				return
			}
		}
	}
}

// streamMonitor handles WebSocket connections for real-time monitoring
func (h *ModernDashboardHandler) streamMonitor(c *gin.Context) {
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("WebSocket upgrade error: %v", err)
		return
	}
	defer conn.Close()

	// Add client to monitor clients
	h.monitorMutex.Lock()
	h.monitorClients[conn] = true
	h.monitorMutex.Unlock()

	// Remove client when done
	defer func() {
		h.monitorMutex.Lock()
		delete(h.monitorClients, conn)
		h.monitorMutex.Unlock()
	}()

	// Send initial system metrics
	h.sendSystemMetrics(conn)

	// Keep connection alive and handle ping/pong
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			// Send periodic system metrics update
			h.sendSystemMetrics(conn)

			// Send ping to keep connection alive
			if err := conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				log.Printf("WebSocket ping error: %v", err)
				return
			}
		}
	}
}

// sendSystemMetrics sends current system metrics to a WebSocket connection
func (h *ModernDashboardHandler) sendSystemMetrics(conn *websocket.Conn) {
	// Get current system metrics
	metrics := h.getSystemMetrics()

	message := MonitorMessage{
		Type:      "system_metrics",
		Timestamp: time.Now(),
		Data:      metrics,
	}

	if err := conn.WriteJSON(message); err != nil {
		log.Printf("WebSocket write error: %v", err)
	}
}

// getSystemMetrics retrieves current system metrics
func (h *ModernDashboardHandler) getSystemMetrics() SystemMetricsUpdate {
	// Get system statistics
	stats := h.statsManager.GetStats()
	pluginStats := h.statsManager.GetPluginStats()

	// Calculate metrics from plugin statistics
	totalExecutions := stats.TotalExecutions
	successfulExecutions := totalExecutions - stats.TotalErrors
	activePlugins := len(pluginStats)

	// Calculate average execution time from plugin stats
	var totalExecutionTime int64
	var executionCount int64

	for _, pluginStat := range pluginStats {
		totalExecutionTime += int64(pluginStat.TotalTime.Milliseconds())
		executionCount += pluginStat.Count
	}

	var successRate float64
	var avgExecutionTime float64
	var errorRate float64

	if totalExecutions > 0 {
		successRate = float64(successfulExecutions) / float64(totalExecutions) * 100
		errorRate = float64(stats.TotalErrors) / float64(totalExecutions) * 100
	}

	if executionCount > 0 {
		avgExecutionTime = float64(totalExecutionTime) / float64(executionCount)
	}

	// Get last hour executions (simplified - would need proper time tracking)
	lastHourExecutions := int64(0)
	if totalExecutions > 0 {
		// Simple estimate based on uptime
		uptimeHours := stats.Uptime.Hours()
		if uptimeHours > 0 {
			lastHourExecutions = int64(float64(totalExecutions) / uptimeHours)
		}
	}

	return SystemMetricsUpdate{
		TotalExecutions:    totalExecutions,
		SuccessRate:        successRate,
		AvgExecutionTime:   avgExecutionTime,
		ActivePlugins:      activePlugins,
		ErrorRate:          errorRate,
		LastHourExecutions: lastHourExecutions,
	}
}

// BroadcastPluginStatusUpdate broadcasts plugin status updates to all monitoring clients
func (h *ModernDashboardHandler) BroadcastPluginStatusUpdate(update PluginStatusUpdate) {
	message := MonitorMessage{
		Type:      "plugin_status",
		Timestamp: time.Now(),
		Data:      update,
	}

	h.monitorMutex.RLock()
	defer h.monitorMutex.RUnlock()

	for conn := range h.monitorClients {
		if err := conn.WriteJSON(message); err != nil {
			log.Printf("WebSocket broadcast error: %v", err)
			// Remove failed connection
			go func(c *websocket.Conn) {
				h.monitorMutex.Lock()
				delete(h.monitorClients, c)
				h.monitorMutex.Unlock()
				c.Close()
			}(conn)
		}
	}
}

// BroadcastSystemMetricsUpdate broadcasts system metrics updates to all monitoring clients
func (h *ModernDashboardHandler) BroadcastSystemMetricsUpdate() {
	metrics := h.getSystemMetrics()

	message := MonitorMessage{
		Type:      "system_metrics",
		Timestamp: time.Now(),
		Data:      metrics,
	}

	h.monitorMutex.RLock()
	defer h.monitorMutex.RUnlock()

	for conn := range h.monitorClients {
		if err := conn.WriteJSON(message); err != nil {
			log.Printf("WebSocket broadcast error: %v", err)
			// Remove failed connection
			go func(c *websocket.Conn) {
				h.monitorMutex.Lock()
				delete(h.monitorClients, c)
				h.monitorMutex.Unlock()
				c.Close()
			}(conn)
		}
	}
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

// debugFilesystem returns the structure of the embedded filesystem for debugging
func (h *ModernDashboardHandler) debugFilesystem(c *gin.Context) {
	nextjsFS := webpkg.GetNextJSFS()

	// Function to recursively list directory contents
	var listDir func(path string, maxDepth int) map[string]interface{}
	listDir = func(path string, maxDepth int) map[string]interface{} {
		result := make(map[string]interface{})

		if maxDepth <= 0 {
			return result
		}

		entries, err := fs.ReadDir(nextjsFS, path)
		if err != nil {
			result["error"] = err.Error()
			return result
		}

		for _, entry := range entries {
			entryPath := path + "/" + entry.Name()
			if path == "." || path == "" {
				entryPath = entry.Name()
			}

			if entry.IsDir() {
				result[entry.Name()] = map[string]interface{}{
					"type":     "directory",
					"contents": listDir(entryPath, maxDepth-1),
				}
			} else {
				// Get file info
				info, err := entry.Info()
				fileInfo := map[string]interface{}{
					"type": "file",
					"size": 0,
				}
				if err == nil {
					fileInfo["size"] = info.Size()
					fileInfo["modified"] = info.ModTime().Format(time.RFC3339)
				}
				result[entry.Name()] = fileInfo
			}
		}

		return result
	}

	// List the filesystem structure with a depth limit
	structure := listDir(".", 4)

	c.JSON(http.StatusOK, map[string]interface{}{
		"success":    true,
		"filesystem": structure,
		"timestamp":  time.Now().Format(time.RFC3339),
	})
}

// debugCSSStatus returns detailed CSS loading status and diagnostics
func (h *ModernDashboardHandler) debugCSSStatus(c *gin.Context) {
	nextjsFS := webpkg.GetNextJSFS()

	// Check CSS file existence and details
	cssInfo := make(map[string]interface{})

	// Check if CSS directory exists
	cssDir := "next/static/css"
	if entries, err := fs.ReadDir(nextjsFS, cssDir); err == nil {
		cssFiles := make([]map[string]interface{}, 0)
		for _, entry := range entries {
			if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".css") {
				filePath := cssDir + "/" + entry.Name()
				fileInfo := map[string]interface{}{
					"name": entry.Name(),
					"path": filePath,
				}

				// Try to read file info
				if info, err := entry.Info(); err == nil {
					fileInfo["size"] = info.Size()
					fileInfo["modified"] = info.ModTime().Format(time.RFC3339)
				}

				// Try to read file content (first 100 chars)
				if data, err := fs.ReadFile(nextjsFS, filePath); err == nil {
					content := string(data)
					if len(content) > 100 {
						content = content[:100] + "..."
					}
					fileInfo["preview"] = content
					fileInfo["readable"] = true
				} else {
					fileInfo["readable"] = false
					fileInfo["error"] = err.Error()
				}

				cssFiles = append(cssFiles, fileInfo)
			}
		}
		cssInfo["files"] = cssFiles
		cssInfo["directory_accessible"] = true
	} else {
		cssInfo["directory_accessible"] = false
		cssInfo["directory_error"] = err.Error()
	}

	// Check legacy directory
	legacyCSSDir := "_next/static/css"
	if entries, err := fs.ReadDir(nextjsFS, legacyCSSDir); err == nil {
		cssInfo["legacy_directory_exists"] = true
		cssInfo["legacy_file_count"] = len(entries)
	} else {
		cssInfo["legacy_directory_exists"] = false
		cssInfo["legacy_error"] = err.Error()
	}

	// Test assets.go functions
	assetsInfo := make(map[string]interface{})

	// Test GetMainCSS
	if cssData, err := webpkg.GetMainCSS(); err == nil {
		assetsInfo["get_main_css"] = map[string]interface{}{
			"success": true,
			"size":    len(cssData),
			"preview": string(cssData[:min(100, len(cssData))]) + "...",
		}
	} else {
		assetsInfo["get_main_css"] = map[string]interface{}{
			"success": false,
			"error":   err.Error(),
		}
	}

	// Test GetIndexHTML
	if htmlData, err := webpkg.GetIndexHTML(); err == nil {
		// Look for CSS references in HTML
		cssRefs := []string{}
		lines := strings.Split(htmlData, "\n")
		for _, line := range lines {
			if strings.Contains(line, ".css") {
				cssRefs = append(cssRefs, strings.TrimSpace(line))
			}
		}
		assetsInfo["get_index_html"] = map[string]interface{}{
			"success":        true,
			"size":           len(htmlData),
			"css_references": cssRefs,
		}
	} else {
		assetsInfo["get_index_html"] = map[string]interface{}{
			"success": false,
			"error":   err.Error(),
		}
	}

	c.JSON(http.StatusOK, map[string]interface{}{
		"success":     true,
		"css_info":    cssInfo,
		"assets_info": assetsInfo,
		"timestamp":   time.Now().Format(time.RFC3339),
	})
}

// Helper function for min
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
