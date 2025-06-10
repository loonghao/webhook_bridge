package worker

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/loonghao/webhook_bridge/api/proto"
	"github.com/loonghao/webhook_bridge/internal/grpc"
)

// Job type constants
const (
	jobTypeWebhook = "webhook"
)

// WebhookJobHandler handles webhook execution jobs
type WebhookJobHandler struct {
	grpcClient *grpc.Client
}

// NewWebhookJobHandler creates a new webhook job handler
func NewWebhookJobHandler(grpcClient *grpc.Client) *WebhookJobHandler {
	return &WebhookJobHandler{
		grpcClient: grpcClient,
	}
}

// Type returns the job type this handler processes
func (h *WebhookJobHandler) Type() string {
	return jobTypeWebhook
}

// Handle processes a webhook job
func (h *WebhookJobHandler) Handle(ctx context.Context, job *Job) error {
	log.Printf("Processing webhook job %s", job.ID)

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
	var requestData map[string]string
	if data, exists := job.Payload["data"]; exists {
		if dataMap, ok := data.(map[string]interface{}); ok {
			requestData = make(map[string]string)
			for k, v := range dataMap {
				requestData[k] = fmt.Sprintf("%v", v)
			}
		}
	}

	// Create gRPC request
	req := &proto.ExecutePluginRequest{
		PluginName: pluginName,
		HttpMethod: method,
		Data:       requestData,
	}

	// Execute plugin via gRPC
	resp, err := h.grpcClient.ExecutePlugin(ctx, req)
	if err != nil {
		return fmt.Errorf("failed to execute plugin %s: %w", pluginName, err)
	}

	// Store result
	job.Result = map[string]interface{}{
		"plugin":         pluginName,
		"method":         method,
		"status_code":    resp.StatusCode,
		"message":        resp.Message,
		"execution_time": resp.ExecutionTime,
		"data":           resp.Data,
		"error":          resp.Error,
	}

	log.Printf("Webhook job %s completed: plugin=%s, status=%d", job.ID, pluginName, resp.StatusCode)
	return nil
}

// BatchJobHandler handles batch processing jobs
type BatchJobHandler struct {
	grpcClient *grpc.Client
}

// NewBatchJobHandler creates a new batch job handler
func NewBatchJobHandler(grpcClient *grpc.Client) *BatchJobHandler {
	return &BatchJobHandler{
		grpcClient: grpcClient,
	}
}

// Type returns the job type this handler processes
func (h *BatchJobHandler) Type() string {
	return "batch"
}

// Handle processes a batch job
func (h *BatchJobHandler) Handle(ctx context.Context, job *Job) error {
	log.Printf("Processing batch job %s", job.ID)

	// Extract batch parameters
	items, ok := job.Payload["items"].([]interface{})
	if !ok {
		return fmt.Errorf("missing or invalid batch items")
	}

	results := make([]map[string]interface{}, 0, len(items))
	var errors []string

	// Process each item in the batch
	for i, item := range items {
		itemMap, ok := item.(map[string]interface{})
		if !ok {
			errors = append(errors, fmt.Sprintf("item %d: invalid format", i))
			continue
		}

		pluginName, ok := itemMap["plugin"].(string)
		if !ok {
			errors = append(errors, fmt.Sprintf("item %d: missing plugin name", i))
			continue
		}

		method, ok := itemMap["method"].(string)
		if !ok {
			method = "POST"
		}

		// Extract request data
		var requestData map[string]string
		if data, exists := itemMap["data"]; exists {
			if dataMap, ok := data.(map[string]interface{}); ok {
				requestData = make(map[string]string)
				for k, v := range dataMap {
					requestData[k] = fmt.Sprintf("%v", v)
				}
			}
		}

		// Create gRPC request
		req := &proto.ExecutePluginRequest{
			PluginName: pluginName,
			HttpMethod: method,
			Data:       requestData,
		}

		// Execute plugin
		resp, err := h.grpcClient.ExecutePlugin(ctx, req)
		if err != nil {
			errors = append(errors, fmt.Sprintf("item %d: %v", i, err))
			continue
		}

		// Add result
		results = append(results, map[string]interface{}{
			"index":          i,
			"plugin":         pluginName,
			"method":         method,
			"status_code":    resp.StatusCode,
			"message":        resp.Message,
			"execution_time": resp.ExecutionTime,
			"data":           resp.Data,
			"error":          resp.Error,
		})
	}

	// Store batch result
	job.Result = map[string]interface{}{
		"total_items":      len(items),
		"successful_items": len(results),
		"failed_items":     len(errors),
		"results":          results,
		"errors":           errors,
	}

	if len(errors) > 0 {
		log.Printf("Batch job %s completed with %d errors out of %d items", job.ID, len(errors), len(items))
	} else {
		log.Printf("Batch job %s completed successfully: %d items processed", job.ID, len(items))
	}

	return nil
}

// ScheduledJobHandler handles scheduled/delayed jobs
type ScheduledJobHandler struct {
	grpcClient *grpc.Client
}

// NewScheduledJobHandler creates a new scheduled job handler
func NewScheduledJobHandler(grpcClient *grpc.Client) *ScheduledJobHandler {
	return &ScheduledJobHandler{
		grpcClient: grpcClient,
	}
}

// Type returns the job type this handler processes
func (h *ScheduledJobHandler) Type() string {
	return "scheduled"
}

// Handle processes a scheduled job
func (h *ScheduledJobHandler) Handle(ctx context.Context, job *Job) error {
	log.Printf("Processing scheduled job %s", job.ID)

	// Check if it's time to execute
	executeAt, ok := job.Payload["execute_at"].(string)
	if ok {
		executeTime, err := time.Parse(time.RFC3339, executeAt)
		if err != nil {
			return fmt.Errorf("invalid execute_at time format: %w", err)
		}

		// Wait until execution time
		if time.Now().Before(executeTime) {
			waitDuration := executeTime.Sub(time.Now())
			log.Printf("Scheduled job %s waiting %v until execution", job.ID, waitDuration)

			select {
			case <-time.After(waitDuration):
				// Continue with execution
			case <-ctx.Done():
				return ctx.Err()
			}
		}
	}

	// Extract the actual job to execute
	actualJob, ok := job.Payload["job"].(map[string]interface{})
	if !ok {
		return fmt.Errorf("missing or invalid job payload")
	}

	// Determine job type
	jobType, ok := actualJob["type"].(string)
	if !ok {
		jobType = jobTypeWebhook
	}

	// Create a new job for the actual execution
	executionJob := &Job{
		ID:       fmt.Sprintf("%s_exec", job.ID),
		Type:     jobType,
		Payload:  actualJob,
		Priority: job.Priority,
		Created:  time.Now(),
	}

	// Execute based on type
	switch jobType {
	case jobTypeWebhook:
		handler := NewWebhookJobHandler(h.grpcClient)
		if err := handler.Handle(ctx, executionJob); err != nil {
			return err
		}
	case "batch":
		handler := NewBatchJobHandler(h.grpcClient)
		if err := handler.Handle(ctx, executionJob); err != nil {
			return err
		}
	default:
		return fmt.Errorf("unsupported scheduled job type: %s", jobType)
	}

	// Store result
	job.Result = executionJob.Result

	log.Printf("Scheduled job %s executed successfully", job.ID)
	return nil
}

// HealthCheckJobHandler handles health check jobs
type HealthCheckJobHandler struct {
	grpcClient *grpc.Client
}

// NewHealthCheckJobHandler creates a new health check job handler
func NewHealthCheckJobHandler(grpcClient *grpc.Client) *HealthCheckJobHandler {
	return &HealthCheckJobHandler{
		grpcClient: grpcClient,
	}
}

// Type returns the job type this handler processes
func (h *HealthCheckJobHandler) Type() string {
	return "health_check"
}

// Handle processes a health check job
func (h *HealthCheckJobHandler) Handle(ctx context.Context, job *Job) error {
	log.Printf("Processing health check job %s", job.ID)

	// Perform health check
	req := &proto.HealthCheckRequest{
		Service: "webhook-bridge",
	}

	resp, err := h.grpcClient.HealthCheck(ctx, req)
	if err != nil {
		job.Result = map[string]interface{}{
			"status":     "unhealthy",
			"error":      err.Error(),
			"checked_at": time.Now(),
		}
		return fmt.Errorf("health check failed: %w", err)
	}

	// Store result
	job.Result = map[string]interface{}{
		"status":     resp.Status,
		"message":    resp.Message,
		"checked_at": time.Now(),
		"details":    resp.Details,
	}

	log.Printf("Health check job %s completed: status=%s", job.ID, resp.Status)
	return nil
}

// CreateJobFromJSON creates a job from JSON payload
func CreateJobFromJSON(jobType string, jsonPayload []byte) (*Job, error) {
	var payload map[string]interface{}
	if err := json.Unmarshal(jsonPayload, &payload); err != nil {
		return nil, fmt.Errorf("failed to parse JSON payload: %w", err)
	}

	job := &Job{
		Type:     jobType,
		Payload:  payload,
		Priority: 0,
		MaxRetry: 3,
		Created:  time.Now(),
	}

	// Extract priority if specified
	if priority, ok := payload["priority"].(float64); ok {
		job.Priority = int(priority)
	}

	// Extract max retry if specified
	if maxRetry, ok := payload["max_retry"].(float64); ok {
		job.MaxRetry = int(maxRetry)
	}

	return job, nil
}
