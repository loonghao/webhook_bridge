package execution

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/loonghao/webhook_bridge/internal/storage"
)

// ExecutionTracker manages execution tracking and history
type ExecutionTracker struct {
	storage storage.ExecutionStorage
	config  *TrackerConfig
	metrics *TrackerMetrics
}

// TrackerConfig defines configuration for execution tracking
type TrackerConfig struct {
	Enabled                    bool          `yaml:"enabled"`
	TrackInput                 bool          `yaml:"track_input"`
	TrackOutput                bool          `yaml:"track_output"`
	TrackErrors                bool          `yaml:"track_errors"`
	MaxInputSize               int           `yaml:"max_input_size"`
	MaxOutputSize              int           `yaml:"max_output_size"`
	CleanupInterval            time.Duration `yaml:"cleanup_interval"`
	RetentionDays              int           `yaml:"retention_days"`
	MetricsAggregationInterval time.Duration `yaml:"metrics_aggregation_interval"`
}

// ExecutionRequest represents a request to start execution tracking
type ExecutionRequest struct {
	PluginName string                 `json:"plugin_name"`
	HTTPMethod string                 `json:"http_method"`
	Input      map[string]interface{} `json:"input"`
	UserAgent  string                 `json:"user_agent,omitempty"`
	RemoteIP   string                 `json:"remote_ip,omitempty"`
	Tags       map[string]string      `json:"tags,omitempty"`
	Metadata   map[string]interface{} `json:"metadata,omitempty"`
}

// ExecutionContext represents the context of an ongoing execution
type ExecutionContext struct {
	ID        string    `json:"id"`
	StartTime time.Time `json:"start_time"`
	TraceID   string    `json:"trace_id"`
}

// ExecutionResult represents the result of an execution
type ExecutionResult struct {
	Output map[string]interface{} `json:"output,omitempty"`
	Error  error                  `json:"error,omitempty"`
}

// NewExecutionTracker creates a new execution tracker
func NewExecutionTracker(storage storage.ExecutionStorage, config *TrackerConfig) *ExecutionTracker {
	return &ExecutionTracker{
		storage: storage,
		config:  config,
		metrics: NewTrackerMetrics(),
	}
}

// StartExecution begins tracking a new execution
func (et *ExecutionTracker) StartExecution(ctx context.Context, req *ExecutionRequest) (*ExecutionContext, error) {
	if !et.config.Enabled {
		return &ExecutionContext{ID: "disabled"}, nil
	}

	executionID := uuid.New().String()
	traceID := extractTraceID(ctx)

	// Prepare input data
	var inputData []byte
	if et.config.TrackInput && req.Input != nil {
		data, err := json.Marshal(req.Input)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal input data: %w", err)
		}

		if len(data) > et.config.MaxInputSize {
			inputData = []byte(fmt.Sprintf(`{"truncated": true, "original_size": %d, "message": "Input data too large"}`, len(data)))
		} else {
			inputData = data
		}
	}

	// Prepare tags
	var tagsData []byte
	if req.Tags != nil {
		data, err := json.Marshal(req.Tags)
		if err != nil {
			log.Printf("Failed to marshal tags: %v", err)
		} else {
			tagsData = data
		}
	}

	// Prepare metadata
	var metadataData []byte
	if req.Metadata != nil {
		data, err := json.Marshal(req.Metadata)
		if err != nil {
			log.Printf("Failed to marshal metadata: %v", err)
		} else {
			metadataData = data
		}
	}

	execution := &storage.ExecutionRecord{
		ID:         executionID,
		PluginName: req.PluginName,
		HTTPMethod: req.HTTPMethod,
		StartTime:  time.Now(),
		Status:     storage.StatusRunning,
		Input:      inputData,
		Attempts:   1,
		TraceID:    traceID,
		UserAgent:  req.UserAgent,
		RemoteIP:   req.RemoteIP,
		Tags:       tagsData,
		Metadata:   metadataData,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	if err := et.storage.SaveExecution(ctx, execution); err != nil {
		return nil, fmt.Errorf("failed to save execution record: %w", err)
	}

	et.metrics.IncrementStarted(req.PluginName)

	return &ExecutionContext{
		ID:        executionID,
		StartTime: execution.StartTime,
		TraceID:   traceID,
	}, nil
}

// CompleteExecution marks an execution as completed
func (et *ExecutionTracker) CompleteExecution(ctx context.Context, execCtx *ExecutionContext, result *ExecutionResult) error {
	if !et.config.Enabled || execCtx.ID == "disabled" {
		return nil
	}

	execution, err := et.storage.GetExecution(ctx, execCtx.ID)
	if err != nil {
		return fmt.Errorf("failed to get execution record: %w", err)
	}

	now := time.Now()
	execution.EndTime = &now
	execution.Duration = now.Sub(execution.StartTime).Nanoseconds()
	execution.UpdatedAt = now

	if result.Error != nil {
		execution.Status = storage.StatusFailed
		if et.config.TrackErrors {
			execution.Error = result.Error.Error()
			execution.ErrorType = classifyError(result.Error)
		}
		et.metrics.IncrementFailed(execution.PluginName)
	} else {
		execution.Status = storage.StatusCompleted
		if et.config.TrackOutput && result.Output != nil {
			data, err := json.Marshal(result.Output)
			if err == nil {
				if len(data) > et.config.MaxOutputSize {
					execution.Output = []byte(fmt.Sprintf(`{"truncated": true, "original_size": %d, "message": "Output data too large"}`, len(data)))
				} else {
					execution.Output = data
				}
			}
		}
		et.metrics.IncrementCompleted(execution.PluginName)
	}

	et.metrics.RecordDuration(execution.PluginName, time.Duration(execution.Duration))

	return et.storage.UpdateExecution(ctx, execution)
}

// GetExecutionHistory retrieves execution history with filtering
func (et *ExecutionTracker) GetExecutionHistory(ctx context.Context, filter *storage.ExecutionFilter) ([]*storage.ExecutionRecord, error) {
	if !et.config.Enabled {
		return []*storage.ExecutionRecord{}, nil
	}

	return et.storage.ListExecutions(ctx, filter)
}

// GetExecutionStats retrieves execution statistics
func (et *ExecutionTracker) GetExecutionStats(ctx context.Context, filter *storage.StatsFilter) (*storage.ExecutionStats, error) {
	if !et.config.Enabled {
		return &storage.ExecutionStats{}, nil
	}

	return et.storage.GetExecutionStats(ctx, filter)
}

// GetExecution retrieves a specific execution record
func (et *ExecutionTracker) GetExecution(ctx context.Context, id string) (*storage.ExecutionRecord, error) {
	if !et.config.Enabled {
		return nil, fmt.Errorf("execution tracking is disabled")
	}

	return et.storage.GetExecution(ctx, id)
}

// CleanupOldExecutions removes old execution records
func (et *ExecutionTracker) CleanupOldExecutions(ctx context.Context) error {
	if !et.config.Enabled {
		return nil
	}

	return et.storage.CleanupOldExecutions(ctx, et.config.RetentionDays)
}

// StartCleanupWorker starts a background worker for cleaning up old records
func (et *ExecutionTracker) StartCleanupWorker(ctx context.Context) {
	if !et.config.Enabled {
		return
	}

	ticker := time.NewTicker(et.config.CleanupInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := et.CleanupOldExecutions(ctx); err != nil {
				log.Printf("Failed to cleanup old executions: %v", err)
			}
		}
	}
}

// GetMetrics returns current metrics
func (et *ExecutionTracker) GetMetrics() *TrackerMetrics {
	return et.metrics
}

// GetStorageInfo returns storage information
func (et *ExecutionTracker) GetStorageInfo(ctx context.Context) (*storage.StorageInfo, error) {
	if !et.config.Enabled {
		return nil, fmt.Errorf("execution tracking is disabled")
	}

	return et.storage.GetStorageInfo(ctx)
}

// classifyError categorizes errors for better analysis
func classifyError(err error) string {
	errStr := err.Error()

	switch {
	case strings.Contains(errStr, "timeout"):
		return "timeout"
	case strings.Contains(errStr, "connection"):
		return "connection_error"
	case strings.Contains(errStr, "permission"):
		return "permission_denied"
	case strings.Contains(errStr, "validation"):
		return "validation_error"
	case strings.Contains(errStr, "not found"):
		return "not_found"
	case strings.Contains(errStr, "unauthorized"):
		return "unauthorized"
	case strings.Contains(errStr, "forbidden"):
		return "forbidden"
	default:
		return "unknown_error"
	}
}

// extractTraceID extracts trace ID from context or generates a new one
func extractTraceID(ctx context.Context) string {
	if traceID := ctx.Value("trace_id"); traceID != nil {
		if id, ok := traceID.(string); ok {
			return id
		}
	}
	return uuid.New().String()
}
