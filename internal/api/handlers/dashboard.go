package handlers

import (
	"context"
	"encoding/json"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/loonghao/webhook_bridge/api/proto"
	"github.com/loonghao/webhook_bridge/internal/api"
	"github.com/loonghao/webhook_bridge/internal/config"
	"github.com/loonghao/webhook_bridge/internal/grpc"
	"github.com/loonghao/webhook_bridge/internal/service"
)

// DashboardHandler handles all dashboard-related API endpoints
type DashboardHandler struct {
	config        *config.Config
	grpcClient    *grpc.Client
	workerPool    service.WorkerPool
	logManager    service.LogManager
	statsManager  service.StatsManager
	connectionMgr service.ConnectionManager
}

// NewDashboardHandler creates a new dashboard handler
func NewDashboardHandler(
	cfg *config.Config,
	grpcClient *grpc.Client,
	workerPool service.WorkerPool,
	logManager service.LogManager,
	statsManager service.StatsManager,
	connectionMgr service.ConnectionManager,
) *DashboardHandler {
	return &DashboardHandler{
		config:        cfg,
		grpcClient:    grpcClient,
		workerPool:    workerPool,
		logManager:    logManager,
		statsManager:  statsManager,
		connectionMgr: connectionMgr,
	}
}

// RegisterRoutes registers all dashboard routes
func (h *DashboardHandler) RegisterRoutes(group *gin.RouterGroup) {
	// System endpoints
	group.GET("/status", h.getStatus)
	group.GET("/stats", h.getStats)
	group.GET("/metrics", h.getMetrics)
	group.GET("/system", h.getSystemInfo)

	// Plugin endpoints
	group.GET("/plugins", h.getPlugins)
	group.POST("/plugins/:name/execute", h.executePlugin)

	// Worker endpoints
	group.GET("/workers", h.getWorkers)
	group.POST("/workers/jobs", h.submitJob)

	// Log endpoints
	group.GET("/logs", h.getLogs)

	// Configuration endpoints
	group.GET("/config", h.getConfig)
	group.POST("/config", h.saveConfig)

	// Connection endpoints
	group.GET("/connection", h.getConnectionStatus)
	group.POST("/connection/reconnect", h.reconnectService)
	group.POST("/connection/test", h.testConnection)
}

// System endpoints
func (h *DashboardHandler) getStatus(c *gin.Context) {
	status := map[string]interface{}{
		"server_status":  "running",
		"grpc_connected": h.grpcClient != nil && h.grpcClient.IsConnected(),
		"worker_count":   h.workerPool.GetWorkerCount(),
		"active_workers": h.workerPool.GetActiveWorkerCount(),
		"total_jobs":     h.workerPool.GetTotalJobs(),
		"completed_jobs": h.workerPool.GetCompletedJobs(),
		"failed_jobs":    h.workerPool.GetFailedJobs(),
		"uptime":         h.workerPool.GetUptime(),
	}

	api.Success(c, status, "System status retrieved successfully")
}

func (h *DashboardHandler) getStats(c *gin.Context) {
	stats := map[string]interface{}{
		"total_requests":        h.statsManager.GetTotalRequests(),
		"successful_requests":   h.statsManager.GetSuccessfulRequests(),
		"failed_requests":       h.statsManager.GetFailedRequests(),
		"average_response_time": h.statsManager.GetAverageResponseTime(),
		"active_connections":    h.statsManager.GetActiveConnections(),
		"plugin_count":          h.statsManager.GetPluginCount(),
		"error_rate":            h.statsManager.GetErrorRate(),
	}

	api.Success(c, stats, "Dashboard statistics retrieved successfully")
}

func (h *DashboardHandler) getMetrics(c *gin.Context) {
	metrics := h.statsManager.GetMetrics()
	api.Success(c, metrics, "System metrics retrieved successfully")
}

func (h *DashboardHandler) getSystemInfo(c *gin.Context) {
	info := map[string]interface{}{
		"version":     "2.0.0-hybrid",
		"go_version":  "1.21+",
		"build_time":  "2025-06-03",
		"config_file": "webhook-bridge.yaml",
		"working_dir": h.config.Directories.WorkingDir,
		"log_dir":     h.config.Directories.LogDir,
		"plugin_dir":  h.config.Directories.PluginDir,
		"data_dir":    h.config.Directories.DataDir,
	}

	api.Success(c, info, "System information retrieved successfully")
}

// Plugin endpoints
func (h *DashboardHandler) getPlugins(c *gin.Context) {
	if h.grpcClient == nil || !h.grpcClient.IsConnected() {
		api.ServiceUnavailable(c, "Python executor not available", "gRPC client not connected")
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	resp, err := h.grpcClient.ListPlugins(ctx, &proto.ListPluginsRequest{})
	if err != nil {
		api.InternalError(c, "Failed to get plugins from executor", err.Error())
		return
	}

	api.Success(c, resp.Plugins, "Plugins retrieved successfully")
}

func (h *DashboardHandler) executePlugin(c *gin.Context) {
	pluginName := c.Param("name")
	if !api.ValidateRequired(c, "plugin name", pluginName) {
		return
	}

	var request struct {
		Method string                 `json:"method"`
		Data   map[string]interface{} `json:"data"`
	}

	if !api.ValidateJSON(c, &request) {
		return
	}

	if h.grpcClient == nil || !h.grpcClient.IsConnected() {
		api.ServiceUnavailable(c, "Python executor not available", "gRPC client not connected")
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Convert data to proto format
	dataMap := make(map[string]string)
	for k, v := range request.Data {
		if str, ok := v.(string); ok {
			dataMap[k] = str
		} else {
			// Convert non-string values to JSON
			if jsonBytes, err := json.Marshal(v); err == nil {
				dataMap[k] = string(jsonBytes)
			}
		}
	}

	req := &proto.ExecutePluginRequest{
		PluginName: pluginName,
		HttpMethod: request.Method,
		Data:       dataMap,
	}

	result, err := h.grpcClient.ExecutePlugin(ctx, req)
	if err != nil {
		api.InternalError(c, "Plugin execution failed", err.Error())
		return
	}

	api.Success(c, result, "Plugin executed successfully")
}

// Worker endpoints
func (h *DashboardHandler) getWorkers(c *gin.Context) {
	workers := h.workerPool.GetWorkerStats()
	api.Success(c, workers, "Worker information retrieved successfully")
}

func (h *DashboardHandler) submitJob(c *gin.Context) {
	var job map[string]interface{}
	if !api.ValidateJSON(c, &job) {
		return
	}

	jobID, err := h.workerPool.SubmitJob(job)
	if err != nil {
		api.InternalError(c, "Failed to submit job", err.Error())
		return
	}

	api.Success(c, map[string]interface{}{
		"job_id": jobID,
		"status": "submitted",
	}, "Job submitted successfully")
}

// Log endpoints
func (h *DashboardHandler) getLogs(c *gin.Context) {
	limitStr := c.DefaultQuery("limit", "50")
	level := c.DefaultQuery("level", "")

	limit, err := strconv.Atoi(limitStr)
	if err != nil {
		api.BadRequest(c, "Invalid limit parameter", "limit must be a valid integer")
		return
	}

	logs := h.logManager.GetLogs(limit, level)
	api.Success(c, logs, "Logs retrieved successfully")
}

// Configuration endpoints
func (h *DashboardHandler) getConfig(c *gin.Context) {
	config := map[string]interface{}{
		"server": map[string]interface{}{
			"host": h.config.Server.Host,
			"port": h.config.Server.Port,
			"mode": h.config.Server.Mode,
		},
		"logging": map[string]interface{}{
			"level":  h.config.Logging.Level,
			"format": h.config.Logging.Format,
		},
		"executor": map[string]interface{}{
			"host": h.config.Executor.Host,
			"port": h.config.Executor.Port,
		},
	}
	api.Success(c, config, "Configuration retrieved successfully")
}

func (h *DashboardHandler) saveConfig(c *gin.Context) {
	var newConfig map[string]interface{}
	if !api.ValidateJSON(c, &newConfig) {
		return
	}

	api.Success(c, map[string]interface{}{
		"status":  "saved",
		"message": "Configuration update not yet implemented",
	}, "Configuration saved successfully")
}

// Connection endpoints
func (h *DashboardHandler) getConnectionStatus(c *gin.Context) {
	status := h.connectionMgr.GetStatus()
	api.Success(c, status, "Connection status retrieved successfully")
}

func (h *DashboardHandler) reconnectService(c *gin.Context) {
	var request struct {
		InterpreterName string `json:"interpreter_name,omitempty"`
	}
	c.ShouldBindJSON(&request)

	err := h.connectionMgr.Reconnect(request.InterpreterName)
	if err != nil {
		api.InternalError(c, "Failed to reconnect service", err.Error())
		return
	}

	api.Success(c, map[string]interface{}{
		"status": "reconnecting",
	}, "Service reconnection initiated")
}

func (h *DashboardHandler) testConnection(c *gin.Context) {
	result, err := h.connectionMgr.TestConnection()
	if err != nil {
		api.InternalError(c, "Connection test failed", err.Error())
		return
	}

	api.Success(c, result, "Connection test completed")
}
