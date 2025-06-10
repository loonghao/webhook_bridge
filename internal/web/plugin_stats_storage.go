package web

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
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
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		log.Printf("Warning: Failed to create data directory %s: %v", dataDir, err)
		dataDir = "." // Fallback to current directory
	}

	filePath := filepath.Join(dataDir, "plugin_stats.json")
	backupPath := filepath.Join(dataDir, "plugin_stats_backup.json")

	storage := &PluginStatsStorage{
		filePath:        filePath,
		backupPath:      backupPath,
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
	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

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

	return pss.saveToFile(pss.filePath)
}

// saveToFile saves data to a specific file
func (pss *PluginStatsStorage) saveToFile(filePath string) error {
	// Create backup before saving
	if _, err := os.Stat(filePath); err == nil {
		if err := pss.createBackup(); err != nil {
			log.Printf("Warning: Failed to create backup: %v", err)
		}
	}

	// Create temporary file for atomic write
	tempFile := filePath + ".tmp"
	file, err := os.Create(tempFile)
	if err != nil {
		return fmt.Errorf("failed to create temp file: %w", err)
	}

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ") // Pretty print JSON
	if err := encoder.Encode(pss.data); err != nil {
		file.Close()
		os.Remove(tempFile)
		return fmt.Errorf("failed to encode JSON: %w", err)
	}

	file.Close()

	// Atomic rename
	if err := os.Rename(tempFile, filePath); err != nil {
		os.Remove(tempFile)
		return fmt.Errorf("failed to rename temp file: %w", err)
	}

	return nil
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
	return os.WriteFile(pss.backupPath, data, 0644)
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
