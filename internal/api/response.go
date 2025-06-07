package api

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// StandardResponse represents the unified API response format
type StandardResponse struct {
	Success   bool        `json:"success"`
	Data      interface{} `json:"data,omitempty"`
	Error     *ErrorInfo  `json:"error,omitempty"`
	Message   string      `json:"message,omitempty"`
	Timestamp string      `json:"timestamp"`
	RequestID string      `json:"request_id,omitempty"`
}

// ErrorInfo represents detailed error information
type ErrorInfo struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Details string `json:"details,omitempty"`
}

// PaginationInfo represents pagination metadata
type PaginationInfo struct {
	Page       int `json:"page"`
	PageSize   int `json:"page_size"`
	Total      int `json:"total"`
	TotalPages int `json:"total_pages"`
}

// PaginatedResponse represents a paginated API response
type PaginatedResponse struct {
	StandardResponse
	Pagination *PaginationInfo `json:"pagination,omitempty"`
}

// Response helper functions

// Success returns a successful response
func Success(c *gin.Context, data interface{}, message ...string) {
	msg := ""
	if len(message) > 0 {
		msg = message[0]
	}

	response := StandardResponse{
		Success:   true,
		Data:      data,
		Message:   msg,
		Timestamp: time.Now().Format(time.RFC3339),
		RequestID: getRequestID(c),
	}

	c.JSON(http.StatusOK, response)
}

// Error returns an error response
func Error(c *gin.Context, code string, message string, details ...string) {
	detail := ""
	if len(details) > 0 {
		detail = details[0]
	}

	response := StandardResponse{
		Success: false,
		Error: &ErrorInfo{
			Code:    code,
			Message: message,
			Details: detail,
		},
		Timestamp: time.Now().Format(time.RFC3339),
		RequestID: getRequestID(c),
	}

	statusCode := getHTTPStatusFromErrorCode(code)
	c.JSON(statusCode, response)
}

// BadRequest returns a 400 Bad Request response
func BadRequest(c *gin.Context, message string, details ...string) {
	Error(c, "BAD_REQUEST", message, details...)
}

// NotFound returns a 404 Not Found response
func NotFound(c *gin.Context, message string, details ...string) {
	Error(c, "NOT_FOUND", message, details...)
}

// InternalError returns a 500 Internal Server Error response
func InternalError(c *gin.Context, message string, details ...string) {
	Error(c, "INTERNAL_ERROR", message, details...)
}

// ServiceUnavailable returns a 503 Service Unavailable response
func ServiceUnavailable(c *gin.Context, message string, details ...string) {
	Error(c, "SERVICE_UNAVAILABLE", message, details...)
}

// Paginated returns a paginated response
func Paginated(c *gin.Context, data interface{}, pagination *PaginationInfo, message ...string) {
	msg := ""
	if len(message) > 0 {
		msg = message[0]
	}

	response := PaginatedResponse{
		StandardResponse: StandardResponse{
			Success:   true,
			Data:      data,
			Message:   msg,
			Timestamp: time.Now().Format(time.RFC3339),
			RequestID: getRequestID(c),
		},
		Pagination: pagination,
	}

	c.JSON(http.StatusOK, response)
}

// Helper functions

func getRequestID(c *gin.Context) string {
	if requestID := c.GetHeader("X-Request-ID"); requestID != "" {
		return requestID
	}
	if requestID := c.GetString("request_id"); requestID != "" {
		return requestID
	}
	return ""
}

func getHTTPStatusFromErrorCode(code string) int {
	switch code {
	case "BAD_REQUEST", "INVALID_INPUT", "VALIDATION_ERROR":
		return http.StatusBadRequest
	case "UNAUTHORIZED":
		return http.StatusUnauthorized
	case "FORBIDDEN":
		return http.StatusForbidden
	case "NOT_FOUND":
		return http.StatusNotFound
	case "CONFLICT":
		return http.StatusConflict
	case "SERVICE_UNAVAILABLE", "EXECUTOR_UNAVAILABLE":
		return http.StatusServiceUnavailable
	case "TIMEOUT":
		return http.StatusRequestTimeout
	default:
		return http.StatusInternalServerError
	}
}

// Common error codes
const (
	ErrorCodeBadRequest          = "BAD_REQUEST"
	ErrorCodeNotFound            = "NOT_FOUND"
	ErrorCodeInternalError       = "INTERNAL_ERROR"
	ErrorCodeServiceUnavailable  = "SERVICE_UNAVAILABLE"
	ErrorCodeExecutorUnavailable = "EXECUTOR_UNAVAILABLE"
	ErrorCodeValidationError     = "VALIDATION_ERROR"
	ErrorCodeTimeout             = "TIMEOUT"
	ErrorCodeUnauthorized        = "UNAUTHORIZED"
	ErrorCodeForbidden           = "FORBIDDEN"
	ErrorCodeConflict            = "CONFLICT"
)
