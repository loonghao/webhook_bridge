package handlers

import (
	"context"
	"encoding/json"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/loonghao/webhook_bridge/api/proto"
	"github.com/loonghao/webhook_bridge/internal/api"
	"github.com/loonghao/webhook_bridge/internal/config"
	"github.com/loonghao/webhook_bridge/internal/grpc"
	"github.com/loonghao/webhook_bridge/internal/service"
)

// WebhookHandler handles all webhook-related API endpoints
type WebhookHandler struct {
	config       *config.Config
	grpcClient   *grpc.Client
	workerPool   service.WorkerPool
	statsManager service.StatsManager
}

// NewWebhookHandler creates a new webhook handler
func NewWebhookHandler(
	cfg *config.Config,
	grpcClient *grpc.Client,
	workerPool service.WorkerPool,
	statsManager service.StatsManager,
) *WebhookHandler {
	return &WebhookHandler{
		config:       cfg,
		grpcClient:   grpcClient,
		workerPool:   workerPool,
		statsManager: statsManager,
	}
}

// RegisterRoutes registers all webhook routes
func (h *WebhookHandler) RegisterRoutes(group *gin.RouterGroup) {
	// Plugin management endpoints
	group.GET("/plugins", h.listPlugins)
	group.GET("/plugins/:plugin", h.getPluginInfo)

	// Webhook execution endpoints (support all HTTP methods)
	group.GET("/webhook/:plugin", h.executePlugin)
	group.POST("/webhook/:plugin", h.executePlugin)
	group.PUT("/webhook/:plugin", h.executePlugin)
	group.DELETE("/webhook/:plugin", h.executePlugin)
	group.PATCH("/webhook/:plugin", h.executePlugin)
}

// Plugin management endpoints
func (h *WebhookHandler) listPlugins(c *gin.Context) {
	if h.grpcClient == nil || !h.grpcClient.IsConnected() {
		api.ServiceUnavailable(c, "Python executor not available", "gRPC client not connected")
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	resp, err := h.grpcClient.ListPlugins(ctx, &proto.ListPluginsRequest{})
	if err != nil {
		api.InternalError(c, "Failed to retrieve plugins", err.Error())
		return
	}

	// Transform plugins for public API response
	publicPlugins := make([]map[string]interface{}, len(resp.Plugins))
	for i, plugin := range resp.Plugins {
		publicPlugins[i] = map[string]interface{}{
			"name":          plugin.Name,
			"description":   plugin.Description,
			"path":          plugin.Path,
			"available":     plugin.IsAvailable,
			"methods":       plugin.SupportedMethods,
			"last_modified": plugin.LastModified,
		}
	}

	api.Success(c, publicPlugins, "Plugins retrieved successfully")
}

func (h *WebhookHandler) getPluginInfo(c *gin.Context) {
	pluginName := c.Param("plugin")
	if !api.ValidateRequired(c, "plugin name", pluginName) {
		return
	}

	if h.grpcClient == nil || !h.grpcClient.IsConnected() {
		api.ServiceUnavailable(c, "Python executor not available", "gRPC client not connected")
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	req := &proto.GetPluginInfoRequest{
		PluginName: pluginName,
	}

	resp, err := h.grpcClient.GetPluginInfo(ctx, req)
	if err != nil {
		api.NotFound(c, "Plugin not found", err.Error())
		return
	}

	if !resp.Found {
		api.NotFound(c, "Plugin not found", "Plugin does not exist")
		return
	}

	// Transform plugin info for public API response
	publicPlugin := map[string]interface{}{
		"name":          resp.Plugin.Name,
		"description":   resp.Plugin.Description,
		"path":          resp.Plugin.Path,
		"methods":       resp.Plugin.SupportedMethods,
		"available":     resp.Plugin.IsAvailable,
		"last_modified": resp.Plugin.LastModified,
	}

	api.Success(c, publicPlugin, "Plugin information retrieved successfully")
}

// Webhook execution endpoint
func (h *WebhookHandler) executePlugin(c *gin.Context) {
	pluginName := c.Param("plugin")
	if !api.ValidateRequired(c, "plugin name", pluginName) {
		return
	}

	// Check if gRPC client is available
	if h.grpcClient == nil || !h.grpcClient.IsConnected() {
		api.ServiceUnavailable(c, "Python executor not available",
			"The Python executor service is not running or not connected. Please check the system status.")
		return
	}

	// Get request method
	method := c.Request.Method

	// Parse request data based on content type and method
	var requestData map[string]interface{}

	switch method {
	case "GET":
		// For GET requests, use query parameters
		requestData = make(map[string]interface{})
		for key, values := range c.Request.URL.Query() {
			if len(values) == 1 {
				requestData[key] = values[0]
			} else {
				requestData[key] = values
			}
		}
	case "POST", "PUT", "PATCH":
		// For POST/PUT/PATCH requests, try to parse JSON body
		if c.GetHeader("Content-Type") == "application/json" {
			if err := c.ShouldBindJSON(&requestData); err != nil {
				// If JSON parsing fails, treat as form data or raw body
				requestData = map[string]interface{}{
					"raw_body": c.Request.Body,
				}
			}
		} else {
			// Handle form data or other content types
			requestData = make(map[string]interface{})
			if err := c.Request.ParseForm(); err == nil {
				for key, values := range c.Request.Form {
					if len(values) == 1 {
						requestData[key] = values[0]
					} else {
						requestData[key] = values
					}
				}
			}
		}
	case "DELETE":
		// For DELETE requests, use query parameters
		requestData = make(map[string]interface{})
		for key, values := range c.Request.URL.Query() {
			if len(values) == 1 {
				requestData[key] = values[0]
			} else {
				requestData[key] = values
			}
		}
	default:
		requestData = make(map[string]interface{})
	}

	// Add request metadata
	requestData["_meta"] = map[string]interface{}{
		"method":     method,
		"headers":    c.Request.Header,
		"remote_ip":  c.ClientIP(),
		"user_agent": c.GetHeader("User-Agent"),
		"request_id": c.GetString("request_id"),
	}

	// Submit job to worker pool for async execution
	jobData := map[string]interface{}{
		"type":       "webhook_execution",
		"plugin":     pluginName,
		"method":     method,
		"data":       requestData,
		"timestamp":  c.GetString("timestamp"),
		"request_id": c.GetString("request_id"),
	}

	jobID, err := h.workerPool.SubmitJob(jobData)
	if err != nil {
		api.InternalError(c, "Failed to submit webhook job", err.Error())
		return
	}

	// For synchronous execution, execute the plugin
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Convert data to proto format
	dataMap := make(map[string]string)
	for k, v := range requestData {
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
		HttpMethod: method,
		Data:       dataMap,
	}

	result, err := h.grpcClient.ExecutePlugin(ctx, req)
	if err != nil {
		// Update job status as failed
		h.workerPool.UpdateJobStatus(jobID, "failed", err.Error())

		// Return appropriate error response
		if err.Error() == "plugin not found" {
			api.NotFound(c, "Plugin not found", "The specified plugin does not exist or is not available")
		} else {
			api.InternalError(c, "Plugin execution failed", err.Error())
		}
		return
	}

	// Update job status as completed
	h.workerPool.UpdateJobStatus(jobID, "completed", "")

	// Update statistics
	h.statsManager.RecordPluginExecution(pluginName, method, true, 0) // TODO: Add actual execution time

	// Return successful response
	response := map[string]interface{}{
		"job_id":    jobID,
		"plugin":    pluginName,
		"method":    method,
		"result":    result,
		"timestamp": c.GetString("timestamp"),
	}

	api.Success(c, response, "Webhook executed successfully")
}

// Helper functions for webhook processing

// validatePluginMethod checks if the plugin supports the given HTTP method
func (h *WebhookHandler) validatePluginMethod(pluginName, method string) error {
	// For now, assume all methods are supported
	// TODO: Implement proper method validation
	return nil
}

// processWebhookData processes and validates webhook request data
func (h *WebhookHandler) processWebhookData(c *gin.Context) (map[string]interface{}, error) {
	data := make(map[string]interface{})

	// Add common webhook metadata
	data["webhook_meta"] = map[string]interface{}{
		"timestamp":    c.GetString("timestamp"),
		"request_id":   c.GetString("request_id"),
		"method":       c.Request.Method,
		"path":         c.Request.URL.Path,
		"query":        c.Request.URL.RawQuery,
		"remote_addr":  c.Request.RemoteAddr,
		"user_agent":   c.GetHeader("User-Agent"),
		"content_type": c.GetHeader("Content-Type"),
	}

	return data, nil
}
