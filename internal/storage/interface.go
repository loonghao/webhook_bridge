package storage

import (
	"context"
	"time"
)

// ExecutionStorage defines the interface for execution history storage
type ExecutionStorage interface {
	// Basic CRUD operations
	SaveExecution(ctx context.Context, execution *ExecutionRecord) error
	GetExecution(ctx context.Context, id string) (*ExecutionRecord, error)
	UpdateExecution(ctx context.Context, execution *ExecutionRecord) error
	DeleteExecution(ctx context.Context, id string) error

	// Query operations
	ListExecutions(ctx context.Context, filter *ExecutionFilter) ([]*ExecutionRecord, error)
	GetExecutionStats(ctx context.Context, filter *StatsFilter) (*ExecutionStats, error)

	// Maintenance operations
	CleanupOldExecutions(ctx context.Context, retentionDays int) error
	GetStorageInfo(ctx context.Context) (*StorageInfo, error)

	// Lifecycle management
	Initialize(ctx context.Context) error
	Close() error
	HealthCheck(ctx context.Context) error
}

// ExecutionRecord represents a single plugin execution record
type ExecutionRecord struct {
	ID         string          `json:"id" db:"id"`
	PluginName string          `json:"plugin_name" db:"plugin_name"`
	HTTPMethod string          `json:"http_method" db:"http_method"`
	StartTime  time.Time       `json:"start_time" db:"start_time"`
	EndTime    *time.Time      `json:"end_time,omitempty" db:"end_time"`
	Status     ExecutionStatus `json:"status" db:"status"`
	Input      []byte          `json:"input" db:"input"`             // JSON stored
	Output     []byte          `json:"output,omitempty" db:"output"` // JSON stored
	Error      string          `json:"error,omitempty" db:"error"`
	ErrorType  string          `json:"error_type,omitempty" db:"error_type"`
	Duration   int64           `json:"duration" db:"duration"` // nanoseconds
	Attempts   int             `json:"attempts" db:"attempts"`
	RetryCount int             `json:"retry_count" db:"retry_count"`
	TraceID    string          `json:"trace_id,omitempty" db:"trace_id"`
	UserAgent  string          `json:"user_agent,omitempty" db:"user_agent"`
	RemoteIP   string          `json:"remote_ip,omitempty" db:"remote_ip"`
	Tags       []byte          `json:"tags,omitempty" db:"tags"`         // JSON stored
	Metadata   []byte          `json:"metadata,omitempty" db:"metadata"` // JSON stored
	CreatedAt  time.Time       `json:"created_at" db:"created_at"`
	UpdatedAt  time.Time       `json:"updated_at" db:"updated_at"`
}

// ExecutionStatus represents the status of an execution
type ExecutionStatus string

const (
	StatusPending   ExecutionStatus = "pending"
	StatusRunning   ExecutionStatus = "running"
	StatusCompleted ExecutionStatus = "completed"
	StatusFailed    ExecutionStatus = "failed"
	StatusTimeout   ExecutionStatus = "timeout"
	StatusCanceled  ExecutionStatus = "canceled"
)

// ExecutionFilter defines filters for querying executions
type ExecutionFilter struct {
	PluginName string            `json:"plugin_name,omitempty"`
	Status     ExecutionStatus   `json:"status,omitempty"`
	StartTime  time.Time         `json:"start_time,omitempty"`
	EndTime    time.Time         `json:"end_time,omitempty"`
	Limit      int               `json:"limit,omitempty"`
	Offset     int               `json:"offset,omitempty"`
	TraceID    string            `json:"trace_id,omitempty"`
	ErrorType  string            `json:"error_type,omitempty"`
	Tags       map[string]string `json:"tags,omitempty"`
}

// StatsFilter defines filters for execution statistics
type StatsFilter struct {
	PluginName string `json:"plugin_name,omitempty"`
	Days       int    `json:"days,omitempty"`
	StartDate  string `json:"start_date,omitempty"`
	EndDate    string `json:"end_date,omitempty"`
}

// ExecutionStats represents aggregated execution statistics
type ExecutionStats struct {
	TotalExecutions      int64        `json:"total_executions"`
	SuccessfulExecutions int64        `json:"successful_executions"`
	FailedExecutions     int64        `json:"failed_executions"`
	TimeoutExecutions    int64        `json:"timeout_executions"`
	SuccessRate          float64      `json:"success_rate"`
	AvgDuration          int64        `json:"avg_duration"`
	MinDuration          int64        `json:"min_duration"`
	MaxDuration          int64        `json:"max_duration"`
	UniquePlugins        int64        `json:"unique_plugins"`
	DailyStats           []DailyStats `json:"daily_stats,omitempty"`
}

// DailyStats represents daily execution statistics
type DailyStats struct {
	Date                 string  `json:"date"`
	TotalExecutions      int64   `json:"total_executions"`
	SuccessfulExecutions int64   `json:"successful_executions"`
	FailedExecutions     int64   `json:"failed_executions"`
	SuccessRate          float64 `json:"success_rate"`
	AvgDuration          int64   `json:"avg_duration"`
}

// StorageInfo provides information about the storage backend
type StorageInfo struct {
	Type         string     `json:"type"`
	Location     string     `json:"location"`
	Size         int64      `json:"size"`
	TotalRecords int64      `json:"total_records"`
	OldestRecord *time.Time `json:"oldest_record,omitempty"`
	NewestRecord *time.Time `json:"newest_record,omitempty"`
	Health       string     `json:"health"`
}
