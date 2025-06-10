package server

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/loonghao/webhook_bridge/internal/storage"
)

// handleGetExecutions handles execution history list requests
func (s *Server) handleGetExecutions(c *gin.Context) {
	if s.executionTracker == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"error": "Execution tracking not available",
		})
		return
	}

	// Parse query parameters
	filter := &storage.ExecutionFilter{
		PluginName: c.Query("plugin"),
		Status:     storage.ExecutionStatus(c.Query("status")),
		Limit:      parseIntQuery(c, "limit", 50),
		Offset:     parseIntQuery(c, "offset", 0),
	}

	if startTime := c.Query("start_time"); startTime != "" {
		if t, err := time.Parse(time.RFC3339, startTime); err == nil {
			filter.StartTime = t
		}
	}

	if endTime := c.Query("end_time"); endTime != "" {
		if t, err := time.Parse(time.RFC3339, endTime); err == nil {
			filter.EndTime = t
		}
	}

	executions, err := s.executionTracker.GetExecutionHistory(c.Request.Context(), filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": fmt.Sprintf("Failed to get execution history: %v", err),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"executions": executions,
		"total":      len(executions),
		"filter":     filter,
		"timestamp":  time.Now().UTC().Format(time.RFC3339),
	})
}

// handleGetExecution handles specific execution details requests
func (s *Server) handleGetExecution(c *gin.Context) {
	if s.storage == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"error": "Execution storage not available",
		})
		return
	}

	executionID := c.Param("id")
	execution, err := s.storage.GetExecution(c.Request.Context(), executionID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error":     fmt.Sprintf("Execution not found: %v", err),
			"timestamp": time.Now().UTC().Format(time.RFC3339),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"execution": execution,
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	})
}

// handleGetExecutionStats handles execution statistics requests
func (s *Server) handleGetExecutionStats(c *gin.Context) {
	if s.executionTracker == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"error": "Execution tracking not available",
		})
		return
	}

	filter := &storage.StatsFilter{
		PluginName: c.Query("plugin"),
		Days:       parseIntQuery(c, "days", 7),
	}

	stats, err := s.executionTracker.GetExecutionStats(c.Request.Context(), filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": fmt.Sprintf("Failed to get execution stats: %v", err),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"stats":     stats,
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	})
}

// handleGetPluginStats handles plugin-specific statistics requests
func (s *Server) handleGetPluginStats(c *gin.Context) {
	if s.executionTracker == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"error": "Execution tracking not available",
		})
		return
	}

	pluginName := c.Param("plugin")
	filter := &storage.StatsFilter{
		PluginName: pluginName,
		Days:       parseIntQuery(c, "days", 7),
	}

	stats, err := s.executionTracker.GetExecutionStats(c.Request.Context(), filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": fmt.Sprintf("Failed to get plugin stats: %v", err),
		})
		return
	}

	// Get real-time metrics from tracker
	var realtimeMetrics map[string]interface{}
	if metrics := s.executionTracker.GetMetrics(); metrics != nil {
		pluginStats := metrics.GetStats(pluginName)
		realtimeMetrics = map[string]interface{}{
			"total_executions":      pluginStats.TotalExecutions,
			"successful_executions": pluginStats.SuccessfulExecutions,
			"failed_executions":     pluginStats.FailedExecutions,
			"success_rate":          metrics.GetSuccessRate(pluginName),
			"avg_execution_time":    metrics.GetAverageExecutionTime(pluginName).String(),
			"min_duration":          pluginStats.MinDuration.String(),
			"max_duration":          pluginStats.MaxDuration.String(),
			"last_execution":        pluginStats.LastExecution,
			"error_types":           pluginStats.ErrorTypes,
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"plugin":           pluginName,
		"stats":            stats,
		"realtime_metrics": realtimeMetrics,
		"timestamp":        time.Now().UTC().Format(time.RFC3339),
	})
}

// handleCleanupExecutions handles cleanup requests
func (s *Server) handleCleanupExecutions(c *gin.Context) {
	if s.executionTracker == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"error": "Execution tracking not available",
		})
		return
	}

	err := s.executionTracker.CleanupOldExecutions(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": fmt.Sprintf("Failed to cleanup executions: %v", err),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":   "Cleanup completed successfully",
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	})
}

// handleGetStorageInfo handles storage information requests
func (s *Server) handleGetStorageInfo(c *gin.Context) {
	if s.executionTracker == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"error": "Execution tracking not available",
		})
		return
	}

	info, err := s.executionTracker.GetStorageInfo(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": fmt.Sprintf("Failed to get storage info: %v", err),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"storage_info": info,
		"timestamp":    time.Now().UTC().Format(time.RFC3339),
	})
}

// parseIntQuery parses an integer query parameter with a default value
func parseIntQuery(c *gin.Context, key string, defaultValue int) int {
	if value := c.Query(key); value != "" {
		if parsed, err := strconv.Atoi(value); err == nil {
			return parsed
		}
	}
	return defaultValue
}
