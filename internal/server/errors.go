package server

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// ErrorResponse represents a standardized error response
type ErrorResponse struct {
	Error     string      `json:"error"`
	Message   string      `json:"message"`
	Details   interface{} `json:"details,omitempty"`
	Path      string      `json:"path"`
	Method    string      `json:"method"`
	RequestID string      `json:"request_id,omitempty"`
	Timestamp string      `json:"timestamp"`
}

// respondWithError sends a standardized error response
func (s *Server) respondWithError(c *gin.Context, statusCode int, errorType, message string, details interface{}) {
	response := ErrorResponse{
		Error:     errorType,
		Message:   message,
		Details:   details,
		Path:      c.Request.URL.Path,
		Method:    c.Request.Method,
		Timestamp: time.Now().UTC().Format(time.RFC3339),
	}

	// Add request ID if available
	if requestID, exists := c.Get("request_id"); exists {
		if id, ok := requestID.(string); ok {
			response.RequestID = id
		}
	}

	c.JSON(statusCode, response)
}

// Common error responses
func (s *Server) badRequest(c *gin.Context, message string, details interface{}) {
	s.respondWithError(c, http.StatusBadRequest, "Bad Request", message, details)
}

func (s *Server) notFound(c *gin.Context, message string) {
	s.respondWithError(c, http.StatusNotFound, "Not Found", message, nil)
}

func (s *Server) internalError(c *gin.Context, message string, details interface{}) {
	s.respondWithError(c, http.StatusInternalServerError, "Internal Server Error", message, details)
}

func (s *Server) serviceUnavailable(c *gin.Context, message string, details interface{}) {
	s.respondWithError(c, http.StatusServiceUnavailable, "Service Unavailable", message, details)
}

func (s *Server) timeout(c *gin.Context, message string) {
	s.respondWithError(c, http.StatusRequestTimeout, "Request Timeout", message, nil)
}

// Recovery middleware for panic handling
func (s *Server) recoveryMiddleware() gin.HandlerFunc {
	return gin.CustomRecovery(func(c *gin.Context, recovered interface{}) {
		s.internalError(c, "Internal server error occurred", gin.H{
			"panic": recovered,
		})
		c.Abort()
	})
}
