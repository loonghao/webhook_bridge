package server

import (
	"fmt"
	"io"
	"os"
	"time"

	"github.com/gin-gonic/gin"
)

// LoggingConfig represents logging configuration
type LoggingConfig struct {
	Level  string
	Format string
	Output io.Writer
}

// setupLogging configures logging for the server
func (s *Server) setupLogging() {
	// Configure Gin logging based on mode
	if s.config.Server.Mode == "release" {
		gin.SetMode(gin.ReleaseMode)
		// In production, you might want to use a proper logging library
		gin.DefaultWriter = os.Stdout
	} else {
		gin.SetMode(gin.DebugMode)
		gin.DefaultWriter = os.Stdout
	}
}

// customLogger creates a custom logger middleware
func (s *Server) customLogger() gin.HandlerFunc {
	return gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {
		return fmt.Sprintf("[%s] %s %s %s %d %s %s %s\n",
			param.TimeStamp.Format("2006-01-02 15:04:05"),
			param.ClientIP,
			param.Method,
			param.Path,
			param.StatusCode,
			param.Latency,
			param.Request.UserAgent(),
			param.ErrorMessage,
		)
	})
}

// structuredLogger creates a structured logger for JSON output
func (s *Server) structuredLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		raw := c.Request.URL.RawQuery

		// Process request
		c.Next()

		// Calculate latency
		latency := time.Since(start)

		// Get request ID
		requestID := ""
		if id, exists := c.Get("request_id"); exists {
			if reqID, ok := id.(string); ok {
				requestID = reqID
			}
		}

		// Log in structured format (JSON-like)
		logData := map[string]interface{}{
			"timestamp":   start.Format(time.RFC3339),
			"request_id":  requestID,
			"method":      c.Request.Method,
			"path":        path,
			"query":       raw,
			"status_code": c.Writer.Status(),
			"latency_ms":  float64(latency.Nanoseconds()) / 1e6,
			"client_ip":   c.ClientIP(),
			"user_agent":  c.Request.UserAgent(),
			"body_size":   c.Writer.Size(),
		}

		// Add error if present
		if len(c.Errors) > 0 {
			logData["errors"] = c.Errors.String()
		}

		// In a real application, you would use a proper structured logger like logrus or zap
		fmt.Printf("LOG: %+v\n", logData)
	}
}

// logRequest logs incoming requests with details
func (s *Server) logRequest(c *gin.Context, message string, details map[string]interface{}) {
	requestID := ""
	if id, exists := c.Get("request_id"); exists {
		if reqID, ok := id.(string); ok {
			requestID = reqID
		}
	}

	logEntry := map[string]interface{}{
		"timestamp":  time.Now().Format(time.RFC3339),
		"request_id": requestID,
		"message":    message,
		"method":     c.Request.Method,
		"path":       c.Request.URL.Path,
		"client_ip":  c.ClientIP(),
	}

	// Merge additional details
	for k, v := range details {
		logEntry[k] = v
	}

	fmt.Printf("REQUEST: %+v\n", logEntry)
}

// logError logs errors with context
func (s *Server) logError(c *gin.Context, err error, message string, details map[string]interface{}) {
	requestID := ""
	if id, exists := c.Get("request_id"); exists {
		if reqID, ok := id.(string); ok {
			requestID = reqID
		}
	}

	logEntry := map[string]interface{}{
		"timestamp":  time.Now().Format(time.RFC3339),
		"request_id": requestID,
		"level":      "error",
		"message":    message,
		"error":      err.Error(),
		"method":     c.Request.Method,
		"path":       c.Request.URL.Path,
		"client_ip":  c.ClientIP(),
	}

	// Merge additional details
	for k, v := range details {
		logEntry[k] = v
	}

	fmt.Printf("ERROR: %+v\n", logEntry)
}
