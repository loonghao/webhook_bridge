package sqlite

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	_ "github.com/mattn/go-sqlite3"

	"github.com/loonghao/webhook_bridge/internal/config"
	"github.com/loonghao/webhook_bridge/internal/storage"
)

// SQLiteStorage implements ExecutionStorage interface using SQLite
type SQLiteStorage struct {
	db     *sql.DB
	dbPath string
	config *config.SQLiteConfig
}

// NewSQLiteStorage creates a new SQLite storage instance
func NewSQLiteStorage(config *config.SQLiteConfig) *SQLiteStorage {
	return &SQLiteStorage{
		config: config,
		dbPath: config.DatabasePath,
	}
}

// Initialize initializes the SQLite database
func (s *SQLiteStorage) Initialize(ctx context.Context) error {
	// Ensure directory exists
	if err := os.MkdirAll(filepath.Dir(s.dbPath), 0755); err != nil {
		return fmt.Errorf("failed to create database directory: %w", err)
	}

	// Open database connection
	db, err := sql.Open("sqlite3", s.buildConnectionString())
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}

	s.db = db

	// Configure connection pool
	s.db.SetMaxOpenConns(s.config.MaxConnections)
	s.db.SetMaxIdleConns(s.config.MaxConnections / 2)
	s.db.SetConnMaxLifetime(time.Hour)

	// Run database migrations
	migrationManager := NewMigrationManager(s.db)
	if err := migrationManager.RunMigrations(ctx); err != nil {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	return nil
}

// buildConnectionString builds the SQLite connection string
func (s *SQLiteStorage) buildConnectionString() string {
	params := []string{s.dbPath}

	if s.config.EnableWAL {
		params = append(params, "?_journal_mode=WAL")
	}

	if s.config.EnableForeignKeys {
		if s.config.EnableWAL {
			params = append(params, "&_foreign_keys=on")
		} else {
			params = append(params, "?_foreign_keys=on")
		}
	}

	return strings.Join(params, "")
}

// SaveExecution saves an execution record
func (s *SQLiteStorage) SaveExecution(ctx context.Context, execution *storage.ExecutionRecord) error {
	query := `
	INSERT INTO executions (
		id, plugin_name, http_method, start_time, end_time, status,
		input, output, error, error_type, duration, attempts, retry_count,
		trace_id, user_agent, remote_ip, tags, metadata, created_at, updated_at
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	_, err := s.db.ExecContext(ctx, query,
		execution.ID,
		execution.PluginName,
		execution.HTTPMethod,
		execution.StartTime,
		execution.EndTime,
		execution.Status,
		execution.Input,
		execution.Output,
		execution.Error,
		execution.ErrorType,
		execution.Duration,
		execution.Attempts,
		execution.RetryCount,
		execution.TraceID,
		execution.UserAgent,
		execution.RemoteIP,
		execution.Tags,
		execution.Metadata,
		execution.CreatedAt,
		execution.UpdatedAt,
	)

	return err
}

// GetExecution retrieves an execution record by ID
func (s *SQLiteStorage) GetExecution(ctx context.Context, id string) (*storage.ExecutionRecord, error) {
	query := "SELECT * FROM executions WHERE id = ?"
	row := s.db.QueryRowContext(ctx, query, id)

	execution := &storage.ExecutionRecord{}
	err := row.Scan(
		&execution.ID,
		&execution.PluginName,
		&execution.HTTPMethod,
		&execution.StartTime,
		&execution.EndTime,
		&execution.Status,
		&execution.Input,
		&execution.Output,
		&execution.Error,
		&execution.ErrorType,
		&execution.Duration,
		&execution.Attempts,
		&execution.RetryCount,
		&execution.TraceID,
		&execution.UserAgent,
		&execution.RemoteIP,
		&execution.Tags,
		&execution.Metadata,
		&execution.CreatedAt,
		&execution.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	return execution, nil
}

// UpdateExecution updates an execution record
func (s *SQLiteStorage) UpdateExecution(ctx context.Context, execution *storage.ExecutionRecord) error {
	query := `
	UPDATE executions SET
		end_time = ?, status = ?, output = ?, error = ?, error_type = ?,
		duration = ?, attempts = ?, retry_count = ?, updated_at = ?
	WHERE id = ?
	`

	_, err := s.db.ExecContext(ctx, query,
		execution.EndTime,
		execution.Status,
		execution.Output,
		execution.Error,
		execution.ErrorType,
		execution.Duration,
		execution.Attempts,
		execution.RetryCount,
		execution.UpdatedAt,
		execution.ID,
	)

	return err
}

// DeleteExecution deletes an execution record
func (s *SQLiteStorage) DeleteExecution(ctx context.Context, id string) error {
	query := "DELETE FROM executions WHERE id = ?"
	_, err := s.db.ExecContext(ctx, query, id)
	return err
}

// HealthCheck checks if the database is accessible
func (s *SQLiteStorage) HealthCheck(ctx context.Context) error {
	return s.db.PingContext(ctx)
}

// ListExecutions retrieves execution records with filtering
func (s *SQLiteStorage) ListExecutions(ctx context.Context, filter *storage.ExecutionFilter) ([]*storage.ExecutionRecord, error) {
	query := "SELECT * FROM executions WHERE 1=1"
	args := []interface{}{}

	if filter.PluginName != "" {
		query += " AND plugin_name = ?"
		args = append(args, filter.PluginName)
	}

	if filter.Status != "" {
		query += " AND status = ?"
		args = append(args, filter.Status)
	}

	if !filter.StartTime.IsZero() {
		query += " AND start_time >= ?"
		args = append(args, filter.StartTime)
	}

	if !filter.EndTime.IsZero() {
		query += " AND start_time <= ?"
		args = append(args, filter.EndTime)
	}

	if filter.TraceID != "" {
		query += " AND trace_id = ?"
		args = append(args, filter.TraceID)
	}

	if filter.ErrorType != "" {
		query += " AND error_type = ?"
		args = append(args, filter.ErrorType)
	}

	query += " ORDER BY start_time DESC"

	if filter.Limit > 0 {
		query += " LIMIT ?"
		args = append(args, filter.Limit)
	}

	if filter.Offset > 0 {
		query += " OFFSET ?"
		args = append(args, filter.Offset)
	}

	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var executions []*storage.ExecutionRecord
	for rows.Next() {
		execution := &storage.ExecutionRecord{}
		err := rows.Scan(
			&execution.ID,
			&execution.PluginName,
			&execution.HTTPMethod,
			&execution.StartTime,
			&execution.EndTime,
			&execution.Status,
			&execution.Input,
			&execution.Output,
			&execution.Error,
			&execution.ErrorType,
			&execution.Duration,
			&execution.Attempts,
			&execution.RetryCount,
			&execution.TraceID,
			&execution.UserAgent,
			&execution.RemoteIP,
			&execution.Tags,
			&execution.Metadata,
			&execution.CreatedAt,
			&execution.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		executions = append(executions, execution)
	}

	return executions, nil
}

// GetExecutionStats retrieves execution statistics
func (s *SQLiteStorage) GetExecutionStats(ctx context.Context, filter *storage.StatsFilter) (*storage.ExecutionStats, error) {
	query := `
	SELECT
		COUNT(*) as total_executions,
		COUNT(CASE WHEN status = 'completed' THEN 1 END) as successful_executions,
		COUNT(CASE WHEN status = 'failed' THEN 1 END) as failed_executions,
		COUNT(CASE WHEN status = 'timeout' THEN 1 END) as timeout_executions,
		COALESCE(AVG(duration), 0) as avg_duration,
		COALESCE(MIN(duration), 0) as min_duration,
		COALESCE(MAX(duration), 0) as max_duration,
		COUNT(DISTINCT plugin_name) as unique_plugins
	FROM executions
	WHERE start_time >= datetime('now', '-' || ? || ' days')
	`

	args := []interface{}{filter.Days}

	if filter.PluginName != "" {
		query += " AND plugin_name = ?"
		args = append(args, filter.PluginName)
	}

	var stats storage.ExecutionStats
	err := s.db.QueryRowContext(ctx, query, args...).Scan(
		&stats.TotalExecutions,
		&stats.SuccessfulExecutions,
		&stats.FailedExecutions,
		&stats.TimeoutExecutions,
		&stats.AvgDuration,
		&stats.MinDuration,
		&stats.MaxDuration,
		&stats.UniquePlugins,
	)

	if err != nil {
		return nil, err
	}

	// Calculate success rate
	if stats.TotalExecutions > 0 {
		stats.SuccessRate = float64(stats.SuccessfulExecutions) / float64(stats.TotalExecutions) * 100
	}

	// Get daily statistics
	dailyStats, err := s.getDailyStats(ctx, filter)
	if err != nil {
		return nil, err
	}
	stats.DailyStats = dailyStats

	return &stats, nil
}

// getDailyStats retrieves daily execution statistics
func (s *SQLiteStorage) getDailyStats(ctx context.Context, filter *storage.StatsFilter) ([]storage.DailyStats, error) {
	query := `
	SELECT
		DATE(start_time) as date,
		COUNT(*) as total_executions,
		COUNT(CASE WHEN status = 'completed' THEN 1 END) as successful_executions,
		COUNT(CASE WHEN status = 'failed' THEN 1 END) as failed_executions,
		COALESCE(AVG(duration), 0) as avg_duration
	FROM executions
	WHERE start_time >= datetime('now', '-' || ? || ' days')
	`

	args := []interface{}{filter.Days}

	if filter.PluginName != "" {
		query += " AND plugin_name = ?"
		args = append(args, filter.PluginName)
	}

	query += " GROUP BY DATE(start_time) ORDER BY date DESC"

	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var dailyStats []storage.DailyStats
	for rows.Next() {
		var stat storage.DailyStats
		err := rows.Scan(
			&stat.Date,
			&stat.TotalExecutions,
			&stat.SuccessfulExecutions,
			&stat.FailedExecutions,
			&stat.AvgDuration,
		)
		if err != nil {
			return nil, err
		}

		if stat.TotalExecutions > 0 {
			stat.SuccessRate = float64(stat.SuccessfulExecutions) / float64(stat.TotalExecutions) * 100
		}

		dailyStats = append(dailyStats, stat)
	}

	return dailyStats, nil
}

// CleanupOldExecutions removes old execution records
func (s *SQLiteStorage) CleanupOldExecutions(ctx context.Context, retentionDays int) error {
	query := `
	DELETE FROM executions
	WHERE created_at < datetime('now', '-' || ? || ' days')
	`

	result, err := s.db.ExecContext(ctx, query, retentionDays)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected > 0 {
		log.Printf("🧹 Cleaned up %d old execution records", rowsAffected)
	}

	return nil
}

// GetStorageInfo returns information about the storage
func (s *SQLiteStorage) GetStorageInfo(ctx context.Context) (*storage.StorageInfo, error) {
	info := &storage.StorageInfo{
		Type:     "sqlite",
		Location: s.dbPath,
	}

	// Get database file size
	if stat, err := os.Stat(s.dbPath); err == nil {
		info.Size = stat.Size()
	}

	// Get record count
	var totalRecords int64
	err := s.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM executions").Scan(&totalRecords)
	if err != nil {
		return nil, err
	}
	info.TotalRecords = totalRecords

	// Get oldest and newest record times
	var oldestRecord, newestRecord sql.NullTime
	err = s.db.QueryRowContext(ctx,
		"SELECT MIN(created_at), MAX(created_at) FROM executions").Scan(&oldestRecord, &newestRecord)
	if err != nil {
		return nil, err
	}

	if oldestRecord.Valid {
		info.OldestRecord = &oldestRecord.Time
	}
	if newestRecord.Valid {
		info.NewestRecord = &newestRecord.Time
	}

	// Check health
	if err := s.HealthCheck(ctx); err != nil {
		info.Health = "unhealthy"
	} else {
		info.Health = "healthy"
	}

	return info, nil
}

// Close closes the database connection
func (s *SQLiteStorage) Close() error {
	if s.db != nil {
		return s.db.Close()
	}
	return nil
}
