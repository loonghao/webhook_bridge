package sqlite

import (
	"context"
	"database/sql"
	"fmt"
	"log"
)

// Migration represents a database migration
type Migration struct {
	Version     int
	Description string
	SQL         string
}

// migrations contains all database migrations in order
var migrations = []Migration{
	{
		Version:     1,
		Description: "Create initial execution tables",
		SQL: `
		CREATE TABLE IF NOT EXISTS executions (
			id TEXT PRIMARY KEY,
			plugin_name TEXT NOT NULL,
			http_method TEXT NOT NULL,
			start_time DATETIME NOT NULL,
			end_time DATETIME,
			status TEXT NOT NULL,
			input TEXT,
			output TEXT,
			error TEXT,
			error_type TEXT,
			duration INTEGER DEFAULT 0,
			attempts INTEGER DEFAULT 1,
			retry_count INTEGER DEFAULT 0,
			trace_id TEXT,
			user_agent TEXT,
			remote_ip TEXT,
			tags TEXT,
			metadata TEXT,
			created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
		);
		
		CREATE TABLE IF NOT EXISTS execution_attempts (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			execution_id TEXT NOT NULL,
			attempt_number INTEGER NOT NULL,
			start_time DATETIME NOT NULL,
			end_time DATETIME,
			status TEXT NOT NULL,
			error TEXT,
			duration INTEGER DEFAULT 0,
			created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (execution_id) REFERENCES executions(id) ON DELETE CASCADE
		);
		
		CREATE TABLE IF NOT EXISTS execution_metrics (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			date DATE NOT NULL,
			plugin_name TEXT NOT NULL,
			total_executions INTEGER DEFAULT 0,
			successful_executions INTEGER DEFAULT 0,
			failed_executions INTEGER DEFAULT 0,
			avg_duration INTEGER DEFAULT 0,
			max_duration INTEGER DEFAULT 0,
			min_duration INTEGER DEFAULT 0,
			created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			UNIQUE(date, plugin_name)
		);
		
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version INTEGER PRIMARY KEY,
			applied_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
		);
		`,
	},
	{
		Version:     2,
		Description: "Add execution indexes for performance",
		SQL: `
		CREATE INDEX IF NOT EXISTS idx_executions_plugin_name ON executions(plugin_name);
		CREATE INDEX IF NOT EXISTS idx_executions_status ON executions(status);
		CREATE INDEX IF NOT EXISTS idx_executions_start_time ON executions(start_time);
		CREATE INDEX IF NOT EXISTS idx_executions_created_at ON executions(created_at);
		CREATE INDEX IF NOT EXISTS idx_execution_attempts_execution_id ON execution_attempts(execution_id);
		CREATE INDEX IF NOT EXISTS idx_execution_metrics_date ON execution_metrics(date);
		CREATE INDEX IF NOT EXISTS idx_execution_metrics_plugin_name ON execution_metrics(plugin_name);
		`,
	},
	{
		Version:     3,
		Description: "Add execution tags and metadata support",
		SQL: `
		CREATE TABLE IF NOT EXISTS execution_tags (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			execution_id TEXT NOT NULL,
			tag_name TEXT NOT NULL,
			tag_value TEXT,
			created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (execution_id) REFERENCES executions(id) ON DELETE CASCADE
		);
		
		CREATE INDEX IF NOT EXISTS idx_execution_tags_execution_id ON execution_tags(execution_id);
		CREATE INDEX IF NOT EXISTS idx_execution_tags_name ON execution_tags(tag_name);
		`,
	},
}

// MigrationManager handles database migrations
type MigrationManager struct {
	db *sql.DB
}

// NewMigrationManager creates a new migration manager
func NewMigrationManager(db *sql.DB) *MigrationManager {
	return &MigrationManager{db: db}
}

// RunMigrations executes all pending migrations
func (mm *MigrationManager) RunMigrations(ctx context.Context) error {
	// Ensure migrations table exists
	if err := mm.ensureMigrationsTable(ctx); err != nil {
		return fmt.Errorf("failed to create migrations table: %w", err)
	}

	// Get current version
	currentVersion, err := mm.getCurrentVersion(ctx)
	if err != nil {
		return fmt.Errorf("failed to get current version: %w", err)
	}

	// Execute pending migrations
	for _, migration := range migrations {
		if migration.Version > currentVersion {
			if err := mm.runMigration(ctx, migration); err != nil {
				return fmt.Errorf("failed to run migration %d: %w", migration.Version, err)
			}
			log.Printf("✅ Applied migration %d: %s", migration.Version, migration.Description)
		}
	}

	return nil
}

// ensureMigrationsTable creates the migrations table if it doesn't exist
func (mm *MigrationManager) ensureMigrationsTable(ctx context.Context) error {
	query := `
	CREATE TABLE IF NOT EXISTS schema_migrations (
		version INTEGER PRIMARY KEY,
		applied_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
	);
	`
	_, err := mm.db.ExecContext(ctx, query)
	return err
}

// getCurrentVersion returns the current migration version
func (mm *MigrationManager) getCurrentVersion(ctx context.Context) (int, error) {
	query := "SELECT COALESCE(MAX(version), 0) FROM schema_migrations"
	var version int
	err := mm.db.QueryRowContext(ctx, query).Scan(&version)
	return version, err
}

// runMigration executes a single migration in a transaction
func (mm *MigrationManager) runMigration(ctx context.Context, migration Migration) error {
	tx, err := mm.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Execute migration SQL
	if _, err := tx.ExecContext(ctx, migration.SQL); err != nil {
		return err
	}

	// Record migration version
	if _, err := tx.ExecContext(ctx, "INSERT INTO schema_migrations (version) VALUES (?)", migration.Version); err != nil {
		return err
	}

	return tx.Commit()
}

// GetAppliedMigrations returns a list of applied migration versions
func (mm *MigrationManager) GetAppliedMigrations(ctx context.Context) ([]int, error) {
	query := "SELECT version FROM schema_migrations ORDER BY version"
	rows, err := mm.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var versions []int
	for rows.Next() {
		var version int
		if err := rows.Scan(&version); err != nil {
			return nil, err
		}
		versions = append(versions, version)
	}

	return versions, nil
}

// GetPendingMigrations returns a list of pending migration versions
func (mm *MigrationManager) GetPendingMigrations(ctx context.Context) ([]Migration, error) {
	currentVersion, err := mm.getCurrentVersion(ctx)
	if err != nil {
		return nil, err
	}

	var pending []Migration
	for _, migration := range migrations {
		if migration.Version > currentVersion {
			pending = append(pending, migration)
		}
	}

	return pending, nil
}
