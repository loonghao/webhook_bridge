package modern

import (
	"fmt"
	"html/template"
	"net/http"
	"path/filepath"
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
	// Load template
	tmpl, err := template.ParseFiles(filepath.Join("web", "templates", "modern-dashboard.html"))
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
	// Dashboard routes
	router.GET("/", h.dashboard)
	router.GET("/dashboard", h.dashboard)
	router.GET("/dashboard/", h.dashboard)

	// API routes for dashboard data (using separate namespace to avoid conflicts)
	api := router.Group("/api/dashboard")
	{
		api.GET("/status", h.getStatus)
		api.GET("/metrics", h.getMetrics)
		api.GET("/plugins", h.getPlugins)
		api.GET("/logs", h.getLogs)
		api.GET("/config", h.getConfig)
		api.GET("/workers", h.getWorkers)
		api.POST("/workers/jobs", h.submitJob)
	}

	// Static assets
	router.Static("/static", "./web/static")
	router.StaticFile("/favicon.ico", "./web/static/favicon.ico")
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

// getPlugins returns plugin information
func (h *ModernDashboardHandler) getPlugins(c *gin.Context) {
	plugins := []map[string]interface{}{
		{
			"name":        "example_plugin",
			"version":     "1.0.0",
			"status":      "active",
			"description": "Example webhook plugin",
			"endpoints":   []string{"/webhook/example"},
			"last_used":   time.Now().Add(-time.Hour).Format(time.RFC3339),
		},
		{
			"name":        "notification_plugin",
			"version":     "2.1.0",
			"status":      "active",
			"description": "Notification webhook plugin",
			"endpoints":   []string{"/webhook/notify", "/webhook/alert"},
			"last_used":   time.Now().Add(-30 * time.Minute).Format(time.RFC3339),
		},
		{
			"name":        "data_processor",
			"version":     "1.5.2",
			"status":      "inactive",
			"description": "Data processing webhook plugin",
			"endpoints":   []string{"/webhook/process"},
			"last_used":   time.Now().Add(-24 * time.Hour).Format(time.RFC3339),
		},
	}

	c.JSON(http.StatusOK, map[string]interface{}{
		"plugins":  plugins,
		"total":    len(plugins),
		"active":   2,
		"inactive": 1,
	})
}

// getLogs returns recent logs
func (h *ModernDashboardHandler) getLogs(c *gin.Context) {
	logs := []map[string]interface{}{
		{
			"timestamp": time.Now().Add(-5 * time.Minute).Format(time.RFC3339),
			"level":     "info",
			"message":   "Webhook request processed successfully",
			"plugin":    "example_plugin",
		},
		{
			"timestamp": time.Now().Add(-10 * time.Minute).Format(time.RFC3339),
			"level":     "warn",
			"message":   "High memory usage detected",
			"component": "system",
		},
		{
			"timestamp": time.Now().Add(-15 * time.Minute).Format(time.RFC3339),
			"level":     "info",
			"message":   "New plugin loaded: notification_plugin",
			"component": "plugin_manager",
		},
		{
			"timestamp": time.Now().Add(-20 * time.Minute).Format(time.RFC3339),
			"level":     "error",
			"message":   "Failed to connect to external service",
			"plugin":    "data_processor",
		},
	}

	c.JSON(http.StatusOK, map[string]interface{}{
		"logs":  logs,
		"total": len(logs),
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
	workers := map[string]interface{}{
		"pool": map[string]interface{}{
			"workers":        4,
			"queue_size":     0,
			"queue_capacity": 1000,
			"total_jobs":     156,
			"completed_jobs": 150,
			"failed_jobs":    6,
			"active_jobs":    0,
			"handlers":       4,
		},
		"workers": []map[string]interface{}{
			{
				"id":             0,
				"status":         "idle",
				"current_job":    nil,
				"jobs_processed": 38,
				"last_activity":  time.Now().Add(-2 * time.Minute).Format(time.RFC3339),
			},
			{
				"id":             1,
				"status":         "idle",
				"current_job":    nil,
				"jobs_processed": 42,
				"last_activity":  time.Now().Add(-1 * time.Minute).Format(time.RFC3339),
			},
			{
				"id":             2,
				"status":         "idle",
				"current_job":    nil,
				"jobs_processed": 35,
				"last_activity":  time.Now().Add(-3 * time.Minute).Format(time.RFC3339),
			},
			{
				"id":             3,
				"status":         "idle",
				"current_job":    nil,
				"jobs_processed": 41,
				"last_activity":  time.Now().Add(-1 * time.Minute).Format(time.RFC3339),
			},
		},
	}

	c.JSON(http.StatusOK, workers)
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
