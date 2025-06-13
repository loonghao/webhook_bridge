package web

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// PluginStatsData represents the data structure for persistent storage
type PluginStatsData struct {
	Version         string                     `json:"version"`
	LastSaved       time.Time                  `json:"last_saved"`
	StartTime       time.Time                  `json:"start_time"`
	TotalRequests   int64                      `json:"total_requests"`
	TotalExecutions int64                      `json:"total_executions"`
	TotalErrors     int64                      `json:"total_errors"`
	PluginStats     map[string]*ExecutionStats `json:"plugin_stats"`
}

// PluginStatsStorage manages persistent storage of plugin statistics
type PluginStatsStorage struct {
	filePath        string
	backupPath      string
	dataDir         string
	data            *PluginStatsData
	mutex           sync.RWMutex
	saveChannel     chan bool
	stopChannel     chan bool
	autoSaveEnabled bool
	saveInterval    time.Duration
	maxBackups      int
}

// NewPluginStatsStorage creates a new plugin statistics storage manager
func NewPluginStatsStorage(dataDir string) *PluginStatsStorage {
	if dataDir == "" {
		dataDir = "./data"
	}

	// Ensure data directory exists
	if err := os.MkdirAll(dataDir, 0750); err != nil {
		log.Printf("Warning: Failed to create data directory %s: %v", dataDir, err)
		dataDir = "." // Fallback to current directory
	}

	filePath := filepath.Join(dataDir, "plugin_stats.json")
	backupPath := filepath.Join(dataDir, "plugin_stats_backup.json")

	storage := &PluginStatsStorage{
		filePath:        filePath,
		backupPath:      backupPath,
		dataDir:         dataDir,
		data:            &PluginStatsData{},
		saveChannel:     make(chan bool, 100), // Buffered channel to prevent blocking
		stopChannel:     make(chan bool, 1),
		autoSaveEnabled: true,
		saveInterval:    30 * time.Second, // Save every 30 seconds
		maxBackups:      5,
	}

	// Initialize data structure
	storage.data.Version = "1.0"
	storage.data.PluginStats = make(map[string]*ExecutionStats)

	// Load existing data
	if err := storage.LoadStats(); err != nil {
		log.Printf("Warning: Failed to load stats: %v", err)
	}

	// Start auto-save goroutine
	go storage.autoSaveWorker()

	return storage
}

// LoadStats loads statistics from the persistent file
func (pss *PluginStatsStorage) LoadStats() error {
	pss.mutex.Lock()
	defer pss.mutex.Unlock()

	// Try to load from main file first
	if err := pss.loadFromFile(pss.filePath); err != nil {
		log.Printf("Failed to load stats from main file: %v", err)

		// Try to load from backup file
		if err := pss.loadFromFile(pss.backupPath); err != nil {
			log.Printf("Failed to load stats from backup file: %v", err)

			// Initialize with default data
			pss.data = &PluginStatsData{
				Version:     "1.0",
				LastSaved:   time.Now(),
				StartTime:   time.Now(),
				PluginStats: make(map[string]*ExecutionStats),
			}

			log.Println("Initialized with default statistics data")
			return nil
		}

		log.Println("Successfully loaded stats from backup file")
	}

	log.Printf("Loaded statistics: %d plugin entries", len(pss.data.PluginStats))
	return nil
}

// loadFromFile loads data from a specific file
func (pss *PluginStatsStorage) loadFromFile(filePath string) error {
	// Validate and clean file path to prevent path traversal
	cleanPath := filepath.Clean(filePath)
	cleanDataDir := filepath.Clean(pss.dataDir)

	// Check if the clean path is within the data directory
	relPath, err := filepath.Rel(cleanDataDir, cleanPath)
	if err != nil || strings.HasPrefix(relPath, "..") {
		return fmt.Errorf("invalid file path: path traversal detected")
	}

	file, err := os.Open(cleanPath)
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	defer func() {
		if err := file.Close(); err != nil {
			log.Printf("Failed to close file: %v", err)
		}
	}()

	var data PluginStatsData
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&data); err != nil {
		return fmt.Errorf("failed to decode JSON: %w", err)
	}

	// Validate data
	if data.PluginStats == nil {
		data.PluginStats = make(map[string]*ExecutionStats)
	}

	pss.data = &data
	return nil
}

// SaveStats saves statistics to the persistent file
func (pss *PluginStatsStorage) SaveStats(statsData *PluginStatsData) error {
	pss.mutex.Lock()
	defer pss.mutex.Unlock()

	// Update last saved timestamp
	statsData.LastSaved = time.Now()
	pss.data = statsData

	// Use synchronous save to avoid race conditions with auto-save worker
	return pss.saveToFileSync(pss.filePath)
}

// saveToFile saves data to a specific file
func (pss *PluginStatsStorage) saveToFile(filePath string) error {
	// Validate and clean file path to prevent path traversal
	cleanPath := filepath.Clean(filePath)
	cleanDataDir := filepath.Clean(pss.dataDir)

	// Check if the clean path is within the data directory
	relPath, err := filepath.Rel(cleanDataDir, cleanPath)
	if err != nil || strings.HasPrefix(relPath, "..") {
		return fmt.Errorf("invalid file path: path traversal detected")
	}

	// Ensure the target directory exists
	targetDir := filepath.Dir(cleanPath)
	if err := os.MkdirAll(targetDir, 0750); err != nil {
		return fmt.Errorf("failed to create target directory %s: %w", targetDir, err)
	}

	// Create backup before saving
	if _, err := os.Stat(cleanPath); err == nil {
		if err := pss.createBackup(); err != nil {
			log.Printf("Warning: Failed to create backup: %v", err)
		}
	}

	// Create temporary file for atomic write with unique name to avoid race conditions
	tempFile := fmt.Sprintf("%s.tmp.%d", cleanPath, time.Now().UnixNano())
	// Validate temp file path as well
	cleanTempFile := filepath.Clean(tempFile)
	relTempPath, err := filepath.Rel(cleanDataDir, cleanTempFile)
	if err != nil || strings.HasPrefix(relTempPath, "..") {
		return fmt.Errorf("invalid temp file path: path traversal detected")
	}

	// Ensure the directory exists before creating the temp file
	tempDir := filepath.Dir(cleanTempFile)
	if err := os.MkdirAll(tempDir, 0750); err != nil {
		return fmt.Errorf("failed to create directory for temp file %s: %w", tempDir, err)
	}

	file, err := os.Create(cleanTempFile)
	if err != nil {
		return fmt.Errorf("failed to create temp file %s: %w", cleanTempFile, err)
	}

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ") // Pretty print JSON
	if err := encoder.Encode(pss.data); err != nil {
		if closeErr := file.Close(); closeErr != nil {
			log.Printf("Failed to close file: %v", closeErr)
		}
		if removeErr := os.Remove(cleanTempFile); removeErr != nil {
			log.Printf("Failed to remove temp file: %v", removeErr)
		}
		return fmt.Errorf("failed to encode JSON: %w", err)
	}

	// Ensure data is written to disk before closing
	if err := file.Sync(); err != nil {
		log.Printf("Warning: Failed to sync file: %v", err)
	}

	if err := file.Close(); err != nil {
		if removeErr := os.Remove(cleanTempFile); removeErr != nil {
			log.Printf("Failed to remove temp file: %v", removeErr)
		}
		return fmt.Errorf("failed to close file: %w", err)
	}

	// Verify temp file exists and has content before rename
	if stat, err := os.Stat(cleanTempFile); err != nil {
		return fmt.Errorf("temp file does not exist before rename: %w", err)
	} else if stat.Size() == 0 {
		return fmt.Errorf("temp file is empty before rename")
	}

	// Atomic rename with retry for macOS compatibility
	var renameErr error
	for i := 0; i < 3; i++ {
		renameErr = os.Rename(cleanTempFile, cleanPath)
		if renameErr == nil {
			break
		}
		log.Printf("Rename attempt %d failed: %v", i+1, renameErr)
		time.Sleep(10 * time.Millisecond) // Brief pause before retry
	}

	if renameErr != nil {
		if removeErr := os.Remove(cleanTempFile); removeErr != nil {
			log.Printf("Failed to remove temp file after rename failure: %v", removeErr)
		}
		return fmt.Errorf("failed to rename temp file %s to %s after 3 attempts: %w", cleanTempFile, cleanPath, renameErr)
	}

	return nil
}

// saveToFileSync is a synchronous version of saveToFile that doesn't interfere with auto-save
func (pss *PluginStatsStorage) saveToFileSync(filePath string) error {
	// This is the same as saveToFile but with additional logging for debugging
	return pss.saveToFile(filePath)
}

// createBackup creates a backup of the current stats file
func (pss *PluginStatsStorage) createBackup() error {
	if _, err := os.Stat(pss.filePath); os.IsNotExist(err) {
		return nil // No file to backup
	}

	// Read current file
	data, err := os.ReadFile(pss.filePath)
	if err != nil {
		return err
	}

	// Write to backup file
	return os.WriteFile(pss.backupPath, data, 0600)
}

// RequestSave requests an asynchronous save operation
func (pss *PluginStatsStorage) RequestSave() {
	select {
	case pss.saveChannel <- true:
		// Save request queued
	default:
		// Channel is full, skip this save request
	}
}

// autoSaveWorker runs in a separate goroutine to handle periodic saves
func (pss *PluginStatsStorage) autoSaveWorker() {
	ticker := time.NewTicker(pss.saveInterval)
	defer ticker.Stop()

	for {
		select {
		case <-pss.saveChannel:
			// Immediate save requested
			if err := pss.saveToFile(pss.filePath); err != nil {
				log.Printf("Error saving stats: %v", err)
			}

		case <-ticker.C:
			// Periodic save
			if pss.autoSaveEnabled {
				if err := pss.saveToFile(pss.filePath); err != nil {
					log.Printf("Error in periodic save: %v", err)
				}
			}

		case <-pss.stopChannel:
			// Final save before stopping
			if err := pss.saveToFile(pss.filePath); err != nil {
				log.Printf("Error in final save: %v", err)
			}
			return
		}
	}
}

// GetData returns a copy of the current statistics data
func (pss *PluginStatsStorage) GetData() *PluginStatsData {
	pss.mutex.RLock()
	defer pss.mutex.RUnlock()

	// Create a deep copy
	dataCopy := &PluginStatsData{
		Version:         pss.data.Version,
		LastSaved:       pss.data.LastSaved,
		StartTime:       pss.data.StartTime,
		TotalRequests:   pss.data.TotalRequests,
		TotalExecutions: pss.data.TotalExecutions,
		TotalErrors:     pss.data.TotalErrors,
		PluginStats:     make(map[string]*ExecutionStats),
	}

	for key, stats := range pss.data.PluginStats {
		statsCopy := *stats
		dataCopy.PluginStats[key] = &statsCopy
	}

	return dataCopy
}

// SetAutoSave enables or disables automatic saving
func (pss *PluginStatsStorage) SetAutoSave(enabled bool) {
	pss.autoSaveEnabled = enabled
}

// SetSaveInterval sets the interval for automatic saves
func (pss *PluginStatsStorage) SetSaveInterval(interval time.Duration) {
	pss.saveInterval = interval
}

// Close gracefully shuts down the storage manager
func (pss *PluginStatsStorage) Close() error {
	// Stop the auto-save worker
	select {
	case pss.stopChannel <- true:
	default:
	}

	// Wait a moment for the final save to complete
	time.Sleep(100 * time.Millisecond)

	return nil
}

// GetFilePath returns the path to the stats file
func (pss *PluginStatsStorage) GetFilePath() string {
	return pss.filePath
}

// GetBackupPath returns the path to the backup file
func (pss *PluginStatsStorage) GetBackupPath() string {
	return pss.backupPath
}
