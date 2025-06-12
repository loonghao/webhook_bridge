package worker

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/loonghao/webhook_bridge/api/proto"
	"github.com/loonghao/webhook_bridge/internal/execution"
	"github.com/loonghao/webhook_bridge/internal/grpc"
)

// TrackedWebhookJobHandler handles webhook jobs with execution tracking
type TrackedWebhookJobHandler struct {
	grpcClient *grpc.Client
	tracker    *execution.ExecutionTracker
}

// NewTrackedWebhookJobHandler creates a new tracked webhook job handler
func NewTrackedWebhookJobHandler(client *grpc.Client, tracker *execution.ExecutionTracker) *TrackedWebhookJobHandler {
	return &TrackedWebhookJobHandler{
		grpcClient: client,
		tracker:    tracker,
	}
}

// Handle processes a webhook job with execution tracking
func (h *TrackedWebhookJobHandler) Handle(ctx context.Context, job *Job) error {
	// Extract webhook parameters from job payload
	pluginName, ok := job.Payload["plugin"].(string)
	if !ok {
		return fmt.Errorf("missing or invalid plugin name")
	}

	method, ok := job.Payload["method"].(string)
	if !ok {
		method = "POST"
	}

	// Extract request data
	var inputData map[string]interface{}
	if data, exists := job.Payload["data"]; exists {
		if dataMap, ok := data.(map[string]interface{}); ok {
			inputData = dataMap
		}
	}

	// Extract additional metadata
	userAgent, _ := job.Payload["user_agent"].(string)
	remoteIP, _ := job.Payload["remote_ip"].(string)

	// Create execution request
	execReq := &execution.ExecutionRequest{
		PluginName: pluginName,
		HTTPMethod: method,
		Input:      inputData,
		UserAgent:  userAgent,
		RemoteIP:   remoteIP,
		Tags: map[string]string{
			"job_id":   job.ID,
			"job_type": "webhook",
		},
		Metadata: map[string]interface{}{
			"job_created_at": job.Created,
			"job_priority":   job.Priority,
		},
	}

	// Start execution tracking
	var execCtx *execution.ExecutionContext
	var err error

	if h.tracker != nil {
		execCtx, err = h.tracker.StartExecution(ctx, execReq)
		if err != nil {
			log.Printf("Failed to start execution tracking: %v", err)
			// Continue execution, don't fail because of tracking issues
		}
	}

	// Execute plugin
	result, execErr := h.executePlugin(ctx, pluginName, method, inputData)

	// Complete execution tracking
	if h.tracker != nil && execCtx != nil {
		execResult := &execution.ExecutionResult{
			Output: result,
			Error:  execErr,
		}

		if trackErr := h.tracker.CompleteExecution(ctx, execCtx, execResult); trackErr != nil {
			log.Printf("Failed to complete execution tracking: %v", trackErr)
		}
	}

	// Store result in job
	job.Result = result

	return execErr
}

// executePlugin executes the plugin via gRPC
func (h *TrackedWebhookJobHandler) executePlugin(ctx context.Context, pluginName, method string, inputData map[string]interface{}) (map[string]interface{}, error) {
	if h.grpcClient == nil {
		return nil, fmt.Errorf("gRPC client not available")
	}

	// Convert input data to string map for gRPC
	requestData := make(map[string]string)
	for k, v := range inputData {
		requestData[k] = fmt.Sprintf("%v", v)
	}

	// Call Python executor
	response, err := h.grpcClient.ExecutePlugin(ctx, &proto.ExecutePluginRequest{
		PluginName: pluginName,
		Data:       requestData,
		HttpMethod: method,
	})

	if err != nil {
		return nil, fmt.Errorf("plugin execution failed: %w", err)
	}

	// Convert response data to interface map
	result := make(map[string]interface{})
	for k, v := range response.Data {
		result[k] = v
	}

	// Add execution metadata
	result["_execution_time"] = response.ExecutionTime
	result["_status_code"] = response.StatusCode
	result["_message"] = response.Message

	if response.Error != "" {
		return result, fmt.Errorf("plugin error: %s", response.Error)
	}

	return result, nil
}

// Type returns the job type this handler processes
func (h *TrackedWebhookJobHandler) Type() string {
	return "webhook"
}

// CanHandle returns true if this handler can process the given job type
func (h *TrackedWebhookJobHandler) CanHandle(jobType string) bool {
	return jobType == "webhook"
}

// GetHandlerInfo returns information about this handler
func (h *TrackedWebhookJobHandler) GetHandlerInfo() HandlerInfo {
	return HandlerInfo{
		Name:        "TrackedWebhookJobHandler",
		Description: "Handles webhook plugin execution with execution tracking",
		JobTypes:    []string{"webhook"},
		Features: []string{
			"execution_tracking",
			"performance_metrics",
			"error_classification",
			"retry_support",
		},
	}
}

// GetMetrics returns handler-specific metrics
func (h *TrackedWebhookJobHandler) GetMetrics() map[string]interface{} {
	metrics := make(map[string]interface{})

	if h.tracker != nil {
		trackerMetrics := h.tracker.GetMetrics()
		if trackerMetrics != nil {
			allStats := trackerMetrics.GetAllStats()

			// Aggregate metrics across all plugins
			var totalExecutions, successfulExecutions, failedExecutions int64
			var totalDuration time.Duration

			for _, stats := range allStats {
				totalExecutions += stats.TotalExecutions
				successfulExecutions += stats.SuccessfulExecutions
				failedExecutions += stats.FailedExecutions
				totalDuration += stats.TotalDuration
			}

			metrics["total_executions"] = totalExecutions
			metrics["successful_executions"] = successfulExecutions
			metrics["failed_executions"] = failedExecutions
			metrics["success_rate"] = trackerMetrics.GetOverallSuccessRate()

			if totalExecutions > 0 {
				metrics["avg_execution_time_ms"] = float64(totalDuration.Nanoseconds()) / float64(totalExecutions) / 1e6
			}

			metrics["plugin_count"] = len(allStats)
			metrics["top_error_types"] = trackerMetrics.GetTopErrorTypes(5)
		}
	}

	metrics["handler_type"] = "tracked_webhook"
	metrics["tracking_enabled"] = h.tracker != nil
	metrics["grpc_available"] = h.grpcClient != nil

	return metrics
}

// Shutdown gracefully shuts down the handler
func (h *TrackedWebhookJobHandler) Shutdown(ctx context.Context) error {
	// No specific shutdown logic needed for this handler
	// The tracker cleanup is handled by the server
	return nil
}

// HandlerInfo provides information about a job handler
type HandlerInfo struct {
	Name        string   `json:"name"`
	Description string   `json:"description"`
	JobTypes    []string `json:"job_types"`
	Features    []string `json:"features"`
}
