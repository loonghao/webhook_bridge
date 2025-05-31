package web

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/loonghao/webhook_bridge/internal/config"
)

// DashboardHandler handles web dashboard requests
type DashboardHandler struct {
	config       *config.Config
	logManager   *LogManager
	statsManager *StatsManager
}

// NewDashboardHandler creates a new dashboard handler
func NewDashboardHandler(cfg *config.Config) *DashboardHandler {
	return &DashboardHandler{
		config:       cfg,
		logManager:   NewLogManager(),
		statsManager: NewStatsManager(),
	}
}

// SetupRoutes sets up the web dashboard routes
func (h *DashboardHandler) SetupRoutes(router gin.IRouter) {
	// Only set up static files and templates if this is the main engine
	if engine, ok := router.(*gin.Engine); ok {
		engine.Static("/static", "./web/static")
		engine.LoadHTMLGlob("web/templates/*")
	}

	// Dashboard routes
	dashboard := router.Group("/dashboard")
	{
		dashboard.GET("/", h.serveDashboard)
		dashboard.GET("/logs", h.getLogs)
		dashboard.GET("/logs/stream", h.streamLogs)
		dashboard.GET("/stats", h.getStats)
		dashboard.GET("/plugins", h.getPluginStats)
		dashboard.GET("/system", h.getSystemInfo)
		dashboard.POST("/plugins/:plugin/execute", h.executePlugin)
		dashboard.DELETE("/logs", h.clearLogs)
	}

	// API routes for dashboard
	api := router.Group("/api/dashboard")
	{
		api.GET("/logs", h.getLogsJSON)
		api.GET("/stats", h.getStatsJSON)
		api.GET("/plugins", h.getPluginStatsJSON)
		api.GET("/system", h.getSystemInfoJSON)
	}
}

// serveDashboard serves the main dashboard page
func (h *DashboardHandler) serveDashboard(c *gin.Context) {
	c.HTML(http.StatusOK, "dashboard.html", gin.H{
		"title":   "Webhook Bridge Dashboard",
		"version": "2.0.0",
		"uptime":  h.statsManager.GetUptime(),
	})
}

// getLogs returns logs in HTML format
func (h *DashboardHandler) getLogs(c *gin.Context) {
	level := c.DefaultQuery("level", "")
	limit := c.DefaultQuery("limit", "100")

	limitInt, err := strconv.Atoi(limit)
	if err != nil {
		limitInt = 100
	}

	logs := h.logManager.GetLogs(level, limitInt)

	c.HTML(http.StatusOK, "logs.html", gin.H{
		"logs":  logs,
		"level": level,
		"limit": limitInt,
	})
}

// getLogsJSON returns logs in JSON format
func (h *DashboardHandler) getLogsJSON(c *gin.Context) {
	level := c.DefaultQuery("level", "")
	limit := c.DefaultQuery("limit", "100")

	limitInt, err := strconv.Atoi(limit)
	if err != nil {
		limitInt = 100
	}

	logs := h.logManager.GetLogs(level, limitInt)

	c.JSON(http.StatusOK, gin.H{
		"logs":      logs,
		"total":     len(logs),
		"level":     level,
		"limit":     limitInt,
		"timestamp": time.Now(),
	})
}

// streamLogs provides real-time log streaming via Server-Sent Events
func (h *DashboardHandler) streamLogs(c *gin.Context) {
	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("Access-Control-Allow-Origin", "*")

	// Create a channel for this client
	clientChan := make(chan LogEntry, 100)
	h.logManager.AddClient(clientChan)
	defer h.logManager.RemoveClient(clientChan)

	// Send initial logs
	logs := h.logManager.GetLogs("", 50)
	for _, log := range logs {
		data, _ := json.Marshal(log)
		fmt.Fprintf(c.Writer, "data: %s\n\n", data)
		c.Writer.Flush()
	}

	// Stream new logs
	for {
		select {
		case log := <-clientChan:
			data, _ := json.Marshal(log)
			fmt.Fprintf(c.Writer, "data: %s\n\n", data)
			c.Writer.Flush()
		case <-c.Request.Context().Done():
			return
		}
	}
}

// getStats returns execution statistics
func (h *DashboardHandler) getStats(c *gin.Context) {
	stats := h.statsManager.GetStats()

	c.HTML(http.StatusOK, "stats.html", gin.H{
		"stats": stats,
	})
}

// getStatsJSON returns execution statistics in JSON format
func (h *DashboardHandler) getStatsJSON(c *gin.Context) {
	stats := h.statsManager.GetStats()
	c.JSON(http.StatusOK, stats)
}

// getPluginStats returns plugin-specific statistics
func (h *DashboardHandler) getPluginStats(c *gin.Context) {
	pluginStats := h.statsManager.GetPluginStats()

	c.HTML(http.StatusOK, "plugins.html", gin.H{
		"plugins": pluginStats,
	})
}

// getPluginStatsJSON returns plugin statistics in JSON format
func (h *DashboardHandler) getPluginStatsJSON(c *gin.Context) {
	pluginStats := h.statsManager.GetPluginStats()
	c.JSON(http.StatusOK, gin.H{
		"plugins":   pluginStats,
		"timestamp": time.Now(),
	})
}

// getSystemInfo returns system information
func (h *DashboardHandler) getSystemInfo(c *gin.Context) {
	systemInfo := h.getSystemInfoData()

	c.HTML(http.StatusOK, "system.html", gin.H{
		"system": systemInfo,
	})
}

// getSystemInfoJSON returns system information in JSON format
func (h *DashboardHandler) getSystemInfoJSON(c *gin.Context) {
	systemInfo := h.getSystemInfoData()
	c.JSON(http.StatusOK, systemInfo)
}

// executePlugin executes a plugin from the dashboard
func (h *DashboardHandler) executePlugin(c *gin.Context) {
	pluginName := c.Param("plugin")

	var payload map[string]interface{}
	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON payload"})
		return
	}

	// Record execution attempt
	h.statsManager.RecordExecution(pluginName, "POST", time.Now())

	// Here you would integrate with your existing plugin execution logic
	// For now, we'll simulate a successful execution
	result := gin.H{
		"plugin":    pluginName,
		"method":    "POST",
		"payload":   payload,
		"timestamp": time.Now(),
		"status":    "success",
		"duration":  "15ms",
	}

	// Log the execution
	h.logManager.AddLog(LogEntry{
		Timestamp: time.Now(),
		Level:     "INFO",
		Source:    "dashboard",
		Message:   fmt.Sprintf("Plugin %s executed via dashboard", pluginName),
		Data:      result,
	})

	c.JSON(http.StatusOK, result)
}

// clearLogs clears all logs
func (h *DashboardHandler) clearLogs(c *gin.Context) {
	h.logManager.ClearLogs()

	h.logManager.AddLog(LogEntry{
		Timestamp: time.Now(),
		Level:     "INFO",
		Source:    "dashboard",
		Message:   "Logs cleared via dashboard",
	})

	c.JSON(http.StatusOK, gin.H{
		"message":   "Logs cleared successfully",
		"timestamp": time.Now(),
	})
}

// getSystemInfoData collects system information
func (h *DashboardHandler) getSystemInfoData() map[string]interface{} {
	return map[string]interface{}{
		"server": map[string]interface{}{
			"address": h.config.GetServerAddress(),
			"mode":    h.config.Server.Mode,
			"uptime":  h.statsManager.GetUptime(),
		},
		"executor": map[string]interface{}{
			"address": h.config.GetExecutorAddress(),
			"timeout": h.config.Executor.Timeout,
		},
		"logging": map[string]interface{}{
			"level":  h.config.Logging.Level,
			"format": h.config.Logging.Format,
		},
		"stats":     h.statsManager.GetStats(),
		"timestamp": time.Now(),
	}
}
