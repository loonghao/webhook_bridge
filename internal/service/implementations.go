package service

import (
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"

	"github.com/loonghao/webhook_bridge/internal/config"
	"github.com/loonghao/webhook_bridge/internal/grpc"
)

// Job represents a work job
type Job struct {
	ID        string                 `json:"id"`
	Type      string                 `json:"type"`
	Data      map[string]interface{} `json:"data"`
	Status    string                 `json:"status"`
	CreatedAt time.Time              `json:"created_at"`
	UpdatedAt time.Time              `json:"updated_at"`
	Error     string                 `json:"error,omitempty"`
	Result    interface{}            `json:"result,omitempty"`
}

// WorkerStats represents worker statistics
type WorkerStats struct {
	ID            string `json:"id"`
	Status        string `json:"status"`
	CurrentJob    string `json:"current_job,omitempty"`
	JobsCompleted int64  `json:"jobs_completed"`
	JobsFailed    int64  `json:"jobs_failed"`
	Uptime        string `json:"uptime"`
	LastActivity  string `json:"last_activity"`
}

// JobStatus constants
const (
	JobStatusPending   = "pending"
	JobStatusRunning   = "running"
	JobStatusCompleted = "completed"
	JobStatusFailed    = "failed"
	JobStatusCancelled = "cancelled"
)

// Worker status constants
const (
	WorkerStatusIdle    = "idle"
	WorkerStatusBusy    = "busy"
	WorkerStatusStopped = "stopped"
	WorkerStatusError   = "error"
)

// Connection status constants
const (
	ConnectionStatusConnected    = "connected"
	ConnectionStatusDisconnected = "disconnected"
	ConnectionStatusConnecting   = "connecting"
	ConnectionStatusError        = "error"
)

// Log level constants
const (
	LogLevelDebug = "debug"
	LogLevelInfo  = "info"
	LogLevelWarn  = "warn"
	LogLevelError = "error"
)

// LogEntry represents a log entry
type LogEntry struct {
	Timestamp string `json:"timestamp"`
	Level     string `json:"level"`
	Message   string `json:"message"`
	Source    string `json:"source,omitempty"`
	Plugin    string `json:"plugin,omitempty"`
	Details   string `json:"details,omitempty"`
}

// ConnectionStatus represents connection status
type ConnectionStatus struct {
	Status                string   `json:"status"`
	ExecutorHost          string   `json:"executor_host"`
	ExecutorPort          int      `json:"executor_port"`
	LastConnected         *string  `json:"last_connected,omitempty"`
	ErrorMessage          *string  `json:"error_message,omitempty"`
	PythonVersion         *string  `json:"python_version,omitempty"`
	AvailableInterpreters []string `json:"available_interpreters,omitempty"`
	Uptime                string   `json:"uptime"`
	ConnectionCount       int      `json:"connection_count"`
}

// LogManager interface for log management
type LogManager interface {
	GetLogs(limit int, level string) []LogEntry
	AddLog(entry LogEntry)
	StreamLogs() <-chan LogEntry
	SetLevel(level string)
	GetLevel() string
}

// StatsManager interface for statistics management
type StatsManager interface {
	GetTotalRequests() int64
	GetSuccessfulRequests() int64
	GetFailedRequests() int64
	GetAverageResponseTime() float64
	GetActiveConnections() int
	GetPluginCount() int
	GetErrorRate() float64
	GetMetrics() map[string]interface{}
	RecordPluginExecution(plugin, method string, success bool, duration time.Duration)
	RecordRequest(success bool, duration time.Duration)
}

// ConnectionManager interface for connection management
type ConnectionManager interface {
	GetStatus() ConnectionStatus
	Reconnect(interpreterName string) error
	TestConnection() (map[string]interface{}, error)
	IsConnected() bool
	GetLastError() error
}

// WorkerPool interface for worker management
type WorkerPool interface {
	Start()
	Stop()
	SubmitJob(jobData map[string]interface{}) (string, error)
	UpdateJobStatus(jobID, status, errorMsg string)
	GetWorkerCount() int
	GetActiveWorkerCount() int
	GetTotalJobs() int64
	GetCompletedJobs() int64
	GetFailedJobs() int64
	GetUptime() string
	GetWorkerStats() []WorkerStats
}

// BasicLogManager implements LogManager interface
type BasicLogManager struct {
	logs    []LogEntry
	level   string
	maxLogs int
	mutex   sync.RWMutex
	stream  chan LogEntry
}

// NewLogManager creates a new log manager
func NewLogManager(logFile, level string) LogManager {
	return &BasicLogManager{
		logs:    make([]LogEntry, 0),
		level:   level,
		maxLogs: 1000,
		stream:  make(chan LogEntry, 100),
	}
}

func (lm *BasicLogManager) GetLogs(limit int, level string) []LogEntry {
	lm.mutex.RLock()
	defer lm.mutex.RUnlock()

	filtered := make([]LogEntry, 0)
	for i := len(lm.logs) - 1; i >= 0 && len(filtered) < limit; i-- {
		log := lm.logs[i]
		if level == "" || log.Level == level {
			filtered = append(filtered, log)
		}
	}

	return filtered
}

func (lm *BasicLogManager) AddLog(entry LogEntry) {
	lm.mutex.Lock()
	defer lm.mutex.Unlock()

	if entry.Timestamp == "" {
		entry.Timestamp = time.Now().Format(time.RFC3339)
	}

	lm.logs = append(lm.logs, entry)

	// Keep only the last maxLogs entries
	if len(lm.logs) > lm.maxLogs {
		lm.logs = lm.logs[len(lm.logs)-lm.maxLogs:]
	}

	// Send to stream
	select {
	case lm.stream <- entry:
	default:
		// Stream is full, skip
	}
}

func (lm *BasicLogManager) StreamLogs() <-chan LogEntry {
	return lm.stream
}

func (lm *BasicLogManager) SetLevel(level string) {
	lm.mutex.Lock()
	defer lm.mutex.Unlock()
	lm.level = level
}

func (lm *BasicLogManager) GetLevel() string {
	lm.mutex.RLock()
	defer lm.mutex.RUnlock()
	return lm.level
}

// BasicStatsManager implements StatsManager interface
type BasicStatsManager struct {
	totalRequests      int64
	successfulRequests int64
	failedRequests     int64
	totalResponseTime  time.Duration
	activeConnections  int
	pluginCount        int
	pluginStats        map[string]*PluginStats
	mutex              sync.RWMutex
	startTime          time.Time
}

type PluginStats struct {
	ExecutionCount int64
	ErrorCount     int64
	TotalDuration  time.Duration
	LastExecution  time.Time
}

// NewStatsManager creates a new stats manager
func NewStatsManager() StatsManager {
	return &BasicStatsManager{
		pluginStats: make(map[string]*PluginStats),
		startTime:   time.Now(),
	}
}

func (sm *BasicStatsManager) GetTotalRequests() int64 {
	sm.mutex.RLock()
	defer sm.mutex.RUnlock()
	return sm.totalRequests
}

func (sm *BasicStatsManager) GetSuccessfulRequests() int64 {
	sm.mutex.RLock()
	defer sm.mutex.RUnlock()
	return sm.successfulRequests
}

func (sm *BasicStatsManager) GetFailedRequests() int64 {
	sm.mutex.RLock()
	defer sm.mutex.RUnlock()
	return sm.failedRequests
}

func (sm *BasicStatsManager) GetAverageResponseTime() float64 {
	sm.mutex.RLock()
	defer sm.mutex.RUnlock()
	if sm.totalRequests == 0 {
		return 0
	}
	return float64(sm.totalResponseTime.Milliseconds()) / float64(sm.totalRequests)
}

func (sm *BasicStatsManager) GetActiveConnections() int {
	sm.mutex.RLock()
	defer sm.mutex.RUnlock()
	return sm.activeConnections
}

func (sm *BasicStatsManager) GetPluginCount() int {
	sm.mutex.RLock()
	defer sm.mutex.RUnlock()
	return sm.pluginCount
}

func (sm *BasicStatsManager) GetErrorRate() float64 {
	sm.mutex.RLock()
	defer sm.mutex.RUnlock()
	if sm.totalRequests == 0 {
		return 0
	}
	return float64(sm.failedRequests) / float64(sm.totalRequests) * 100
}

func (sm *BasicStatsManager) GetMetrics() map[string]interface{} {
	sm.mutex.RLock()
	defer sm.mutex.RUnlock()

	return map[string]interface{}{
		"total_requests":        sm.totalRequests,
		"successful_requests":   sm.successfulRequests,
		"failed_requests":       sm.failedRequests,
		"average_response_time": sm.GetAverageResponseTime(),
		"active_connections":    sm.activeConnections,
		"plugin_count":          sm.pluginCount,
		"error_rate":            sm.GetErrorRate(),
		"uptime":                time.Since(sm.startTime).String(),
		"plugin_stats":          sm.pluginStats,
	}
}

func (sm *BasicStatsManager) RecordPluginExecution(plugin, method string, success bool, duration time.Duration) {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()

	key := fmt.Sprintf("%s:%s", plugin, method)
	if sm.pluginStats[key] == nil {
		sm.pluginStats[key] = &PluginStats{}
	}

	stats := sm.pluginStats[key]
	stats.ExecutionCount++
	stats.TotalDuration += duration
	stats.LastExecution = time.Now()

	if !success {
		stats.ErrorCount++
	}
}

func (sm *BasicStatsManager) RecordRequest(success bool, duration time.Duration) {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()

	sm.totalRequests++
	sm.totalResponseTime += duration

	if success {
		sm.successfulRequests++
	} else {
		sm.failedRequests++
	}
}

// BasicConnectionManager implements ConnectionManager interface
type BasicConnectionManager struct {
	grpcClient *grpc.Client
	config     *config.Config
	status     ConnectionStatus
	mutex      sync.RWMutex
	startTime  time.Time
	lastError  error
}

// NewConnectionManager creates a new connection manager
func NewConnectionManager(grpcClient *grpc.Client, cfg *config.Config) ConnectionManager {
	status := ConnectionStatus{
		ExecutorHost: cfg.Executor.Host,
		ExecutorPort: cfg.Executor.Port,
		Uptime:       "0s",
	}

	if grpcClient != nil && grpcClient.IsConnected() {
		status.Status = ConnectionStatusConnected
		now := time.Now().Format(time.RFC3339)
		status.LastConnected = &now
	} else {
		status.Status = ConnectionStatusDisconnected
	}

	return &BasicConnectionManager{
		grpcClient: grpcClient,
		config:     cfg,
		status:     status,
		startTime:  time.Now(),
	}
}

func (cm *BasicConnectionManager) GetStatus() ConnectionStatus {
	cm.mutex.RLock()
	defer cm.mutex.RUnlock()

	// Update uptime
	cm.status.Uptime = time.Since(cm.startTime).String()

	// Update connection status
	if cm.grpcClient != nil && cm.grpcClient.IsConnected() {
		cm.status.Status = ConnectionStatusConnected
		if cm.status.LastConnected == nil {
			now := time.Now().Format(time.RFC3339)
			cm.status.LastConnected = &now
		}
	} else {
		cm.status.Status = ConnectionStatusDisconnected
		if cm.lastError != nil {
			errMsg := cm.lastError.Error()
			cm.status.ErrorMessage = &errMsg
		}
	}

	return cm.status
}

func (cm *BasicConnectionManager) Reconnect(interpreterName string) error {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()

	cm.status.Status = ConnectionStatusConnecting

	// TODO: Implement actual reconnection logic
	// For now, just simulate reconnection
	time.Sleep(100 * time.Millisecond)

	if cm.grpcClient != nil {
		// Try to reconnect
		err := cm.grpcClient.Reconnect()
		if err != nil {
			cm.lastError = err
			cm.status.Status = ConnectionStatusError
			errMsg := err.Error()
			cm.status.ErrorMessage = &errMsg
			return err
		}

		cm.status.Status = ConnectionStatusConnected
		now := time.Now().Format(time.RFC3339)
		cm.status.LastConnected = &now
		cm.status.ErrorMessage = nil
	}

	return nil
}

func (cm *BasicConnectionManager) TestConnection() (map[string]interface{}, error) {
	cm.mutex.RLock()
	defer cm.mutex.RUnlock()

	result := map[string]interface{}{
		"timestamp": time.Now().Format(time.RFC3339),
		"host":      cm.config.Executor.Host,
		"port":      cm.config.Executor.Port,
	}

	if cm.grpcClient != nil && cm.grpcClient.IsConnected() {
		result["status"] = "success"
		result["message"] = "Connection test successful"
		result["response_time"] = "< 1ms"
		return result, nil
	}

	result["status"] = "failed"
	result["message"] = "gRPC client not connected"
	if cm.lastError != nil {
		result["error"] = cm.lastError.Error()
	}

	return result, fmt.Errorf("connection test failed")
}

func (cm *BasicConnectionManager) IsConnected() bool {
	cm.mutex.RLock()
	defer cm.mutex.RUnlock()
	return cm.grpcClient != nil && cm.grpcClient.IsConnected()
}

func (cm *BasicConnectionManager) GetLastError() error {
	cm.mutex.RLock()
	defer cm.mutex.RUnlock()
	return cm.lastError
}

// BasicWorkerPool implements WorkerPool interface
type BasicWorkerPool struct {
	workers       []*Worker
	jobQueue      chan *Job
	jobs          map[string]*Job
	workerCount   int
	totalJobs     int64
	completedJobs int64
	failedJobs    int64
	startTime     time.Time
	mutex         sync.RWMutex
	stopChan      chan struct{}
	running       bool
}

// Worker represents a worker
type Worker struct {
	ID            string
	Status        string
	CurrentJob    *Job
	JobsCompleted int64
	JobsFailed    int64
	StartTime     time.Time
	LastActivity  time.Time
	stopChan      chan struct{}
}

// NewWorkerPool creates a new worker pool
func NewWorkerPool(workerCount int) WorkerPool {
	return &BasicWorkerPool{
		workers:     make([]*Worker, 0, workerCount),
		jobQueue:    make(chan *Job, workerCount*10),
		jobs:        make(map[string]*Job),
		workerCount: workerCount,
		startTime:   time.Now(),
		stopChan:    make(chan struct{}),
	}
}

func (wp *BasicWorkerPool) Start() {
	wp.mutex.Lock()
	defer wp.mutex.Unlock()

	if wp.running {
		return
	}

	wp.running = true

	// Start workers
	for i := 0; i < wp.workerCount; i++ {
		worker := &Worker{
			ID:           fmt.Sprintf("worker-%d", i+1),
			Status:       WorkerStatusIdle,
			StartTime:    time.Now(),
			LastActivity: time.Now(),
			stopChan:     make(chan struct{}),
		}
		wp.workers = append(wp.workers, worker)
		go wp.runWorker(worker)
	}
}

func (wp *BasicWorkerPool) Stop() {
	wp.mutex.Lock()
	defer wp.mutex.Unlock()

	if !wp.running {
		return
	}

	wp.running = false
	close(wp.stopChan)

	// Stop all workers
	for _, worker := range wp.workers {
		close(worker.stopChan)
		worker.Status = WorkerStatusStopped
	}
}

func (wp *BasicWorkerPool) SubmitJob(jobData map[string]interface{}) (string, error) {
	wp.mutex.Lock()
	defer wp.mutex.Unlock()

	if !wp.running {
		return "", fmt.Errorf("worker pool is not running")
	}

	job := &Job{
		ID:        uuid.New().String(),
		Type:      fmt.Sprintf("%v", jobData["type"]),
		Data:      jobData,
		Status:    JobStatusPending,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	wp.jobs[job.ID] = job
	wp.totalJobs++

	// Submit to job queue
	select {
	case wp.jobQueue <- job:
		return job.ID, nil
	default:
		return "", fmt.Errorf("job queue is full")
	}
}

func (wp *BasicWorkerPool) UpdateJobStatus(jobID, status, errorMsg string) {
	wp.mutex.Lock()
	defer wp.mutex.Unlock()

	job, exists := wp.jobs[jobID]
	if !exists {
		return
	}

	job.Status = status
	job.UpdatedAt = time.Now()

	if errorMsg != "" {
		job.Error = errorMsg
	}

	if status == JobStatusCompleted {
		wp.completedJobs++
	} else if status == JobStatusFailed {
		wp.failedJobs++
	}
}

func (wp *BasicWorkerPool) GetWorkerCount() int {
	wp.mutex.RLock()
	defer wp.mutex.RUnlock()
	return wp.workerCount
}

func (wp *BasicWorkerPool) GetActiveWorkerCount() int {
	wp.mutex.RLock()
	defer wp.mutex.RUnlock()

	active := 0
	for _, worker := range wp.workers {
		if worker.Status == WorkerStatusBusy {
			active++
		}
	}
	return active
}

func (wp *BasicWorkerPool) GetTotalJobs() int64 {
	wp.mutex.RLock()
	defer wp.mutex.RUnlock()
	return wp.totalJobs
}

func (wp *BasicWorkerPool) GetCompletedJobs() int64 {
	wp.mutex.RLock()
	defer wp.mutex.RUnlock()
	return wp.completedJobs
}

func (wp *BasicWorkerPool) GetFailedJobs() int64 {
	wp.mutex.RLock()
	defer wp.mutex.RUnlock()
	return wp.failedJobs
}

func (wp *BasicWorkerPool) GetUptime() string {
	wp.mutex.RLock()
	defer wp.mutex.RUnlock()
	return time.Since(wp.startTime).String()
}

func (wp *BasicWorkerPool) GetWorkerStats() []WorkerStats {
	wp.mutex.RLock()
	defer wp.mutex.RUnlock()

	stats := make([]WorkerStats, len(wp.workers))
	for i, worker := range wp.workers {
		currentJob := ""
		if worker.CurrentJob != nil {
			currentJob = worker.CurrentJob.ID
		}

		stats[i] = WorkerStats{
			ID:            worker.ID,
			Status:        worker.Status,
			CurrentJob:    currentJob,
			JobsCompleted: worker.JobsCompleted,
			JobsFailed:    worker.JobsFailed,
			Uptime:        time.Since(worker.StartTime).String(),
			LastActivity:  worker.LastActivity.Format(time.RFC3339),
		}
	}

	return stats
}

func (wp *BasicWorkerPool) runWorker(worker *Worker) {
	for {
		select {
		case <-worker.stopChan:
			return
		case job := <-wp.jobQueue:
			wp.processJob(worker, job)
		}
	}
}

func (wp *BasicWorkerPool) processJob(worker *Worker, job *Job) {
	worker.Status = WorkerStatusBusy
	worker.CurrentJob = job
	worker.LastActivity = time.Now()

	job.Status = JobStatusRunning
	job.UpdatedAt = time.Now()

	// Simulate job processing
	time.Sleep(100 * time.Millisecond)

	// Mark job as completed
	job.Status = JobStatusCompleted
	job.UpdatedAt = time.Now()
	job.Result = map[string]interface{}{
		"message": "Job processed successfully",
		"worker":  worker.ID,
	}

	worker.JobsCompleted++
	worker.CurrentJob = nil
	worker.Status = WorkerStatusIdle
	worker.LastActivity = time.Now()

	wp.mutex.Lock()
	wp.completedJobs++
	wp.mutex.Unlock()
}
