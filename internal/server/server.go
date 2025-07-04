package server

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"log"
	"net/http"
	"strings"
	"sync/atomic"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/loonghao/webhook_bridge/api/proto"
	"github.com/loonghao/webhook_bridge/internal/config"
	"github.com/loonghao/webhook_bridge/internal/execution"
	"github.com/loonghao/webhook_bridge/internal/grpc"
	"github.com/loonghao/webhook_bridge/internal/storage"
	"github.com/loonghao/webhook_bridge/internal/storage/sqlite"
	"github.com/loonghao/webhook_bridge/internal/web/modern"
	"github.com/loonghao/webhook_bridge/internal/worker"
	"github.com/loonghao/webhook_bridge/pkg/version"
)

// Server represents the webhook bridge server
type Server struct {
	config           *config.Config
	grpcClient       *grpc.Client
	modernDash       *modern.ModernDashboardHandler
	workerPool       *worker.Pool
	executionTracker *execution.ExecutionTracker
	storage          storage.ExecutionStorage

	// Metrics
	requestCount  int64
	errorCount    int64
	totalExecTime int64
	startTime     time.Time
}

// New creates a new server instance
func New(cfg *config.Config) *Server {
	// Initialize storage
	var executionStorage storage.ExecutionStorage

	switch cfg.Storage.Type {
	case "sqlite":
		executionStorage = sqlite.NewSQLiteStorage(&cfg.Storage.SQLite)
	default:
		log.Printf("⚠️  Unsupported storage type: %s", cfg.Storage.Type)
		log.Printf("🔄 Execution tracking will be disabled")
		executionStorage = nil
	}

	// Initialize execution tracker
	var tracker *execution.ExecutionTracker
	if executionStorage != nil {
		trackerConfig := &execution.TrackerConfig{
			Enabled:                    cfg.ExecutionTracking.Enabled,
			TrackInput:                 cfg.ExecutionTracking.TrackInput,
			TrackOutput:                cfg.ExecutionTracking.TrackOutput,
			TrackErrors:                cfg.ExecutionTracking.TrackErrors,
			MaxInputSize:               cfg.ExecutionTracking.MaxInputSize,
			MaxOutputSize:              cfg.ExecutionTracking.MaxOutputSize,
			CleanupInterval:            cfg.ExecutionTracking.CleanupInterval,
			RetentionDays:              cfg.Storage.SQLite.RetentionDays,
			MetricsAggregationInterval: cfg.ExecutionTracking.MetricsAggregationInterval,
		}

		tracker = execution.NewExecutionTracker(executionStorage, trackerConfig)
	}

	// Create gRPC client
	grpcClient := grpc.NewClient(&cfg.Executor)

	// Create dashboard handlers
	modernDash := modern.NewModernDashboardHandler(cfg)

	// Create worker pool
	workerPool := worker.NewPool(4) // Default 4 workers

	return &Server{
		config:           cfg,
		grpcClient:       grpcClient,
		modernDash:       modernDash,
		workerPool:       workerPool,
		executionTracker: tracker,
		storage:          executionStorage,
		startTime:        time.Now(),
	}
}

// Start initializes the gRPC connection and worker pool
func (s *Server) Start() error {
	// Initialize storage
	if s.storage != nil {
		ctx := context.Background()
		if err := s.storage.Initialize(ctx); err != nil {
			log.Printf("⚠️  Failed to initialize storage: %v", err)
			s.storage = nil
			s.executionTracker = nil
		} else {
			log.Printf("✅ Execution storage initialized successfully")

			// Start cleanup worker
			if s.executionTracker != nil {
				go s.executionTracker.StartCleanupWorker(ctx)
				log.Printf("✅ Execution cleanup worker started")
			}
		}
	}

	// Setup logging and statistics for gRPC client
	if s.modernDash != nil {
		logManager := s.modernDash.GetLogManager()
		statsManager := s.modernDash.GetStatsManager()

		if logManager != nil && statsManager != nil {
			grpc.SetupClientWithLoggingAndStats(s.grpcClient, logManager, statsManager)
			log.Printf("✅ Enhanced gRPC client with logging and statistics")
		}
	}

	// Connect to gRPC server
	err := s.grpcClient.Connect()

	// Start worker pool regardless of gRPC connection status
	if s.workerPool != nil {
		ctx := context.Background()
		s.workerPool.Start(ctx)

		// Register job handlers with execution tracking
		s.workerPool.RegisterHandler(worker.NewTrackedWebhookJobHandler(s.grpcClient, s.executionTracker))
		s.workerPool.RegisterHandler(worker.NewBatchJobHandler(s.grpcClient))
		s.workerPool.RegisterHandler(worker.NewScheduledJobHandler(s.grpcClient))
		s.workerPool.RegisterHandler(worker.NewHealthCheckJobHandler(s.grpcClient))
	}

	return err
}

// Stop closes the gRPC connection
func (s *Server) Stop() error {
	return s.grpcClient.Close()
}

// SetupRoutes configures the HTTP routes
func (s *Server) SetupRoutes(router *gin.Engine) {
	// Add custom middleware
	s.setupMiddleware(router)

	// Add LogManager middleware from ModernDashboardHandler
	if s.modernDash != nil && s.modernDash.GetLogManager() != nil {
		router.Use(s.modernDash.GetLogManager().LogMiddleware())
	}

	// Modern dashboard routes
	s.modernDash.RegisterRoutes(router)

	// Health check endpoint
	router.GET("/health", s.healthCheck)

	// Metrics endpoint
	router.GET("/metrics", s.metrics)

	// Worker management endpoints
	router.GET("/workers", s.getWorkerStats)
	router.POST("/workers/jobs", s.submitJob)

	// Execution tracking endpoints
	s.setupExecutionHistoryRoutes(router)

	// API v1 routes
	v1 := router.Group("/api/v1")
	v1.Use(s.apiMiddleware())
	{
		v1.GET("/plugins", s.listPlugins)
		v1.GET("/plugins/:plugin", s.getPluginInfo)
		v1.POST("/webhook/:plugin", s.executePlugin)
		v1.GET("/webhook/:plugin", s.executePlugin)
		v1.PUT("/webhook/:plugin", s.executePlugin)
		v1.DELETE("/webhook/:plugin", s.executePlugin)
	}

	// Latest API routes (alias for v1)
	latest := router.Group("/api/latest")
	latest.Use(s.apiMiddleware())
	{
		latest.GET("/plugins", s.listPlugins)
		latest.GET("/plugins/:plugin", s.getPluginInfo)
		latest.POST("/webhook/:plugin", s.executePlugin)
		latest.GET("/webhook/:plugin", s.executePlugin)
		latest.PUT("/webhook/:plugin", s.executePlugin)
		latest.DELETE("/webhook/:plugin", s.executePlugin)
	}

	// API documentation endpoint
	router.GET("/api", s.serveRoot)

	// Note: NoRoute handler is registered by ModernDashboardHandler to serve SPA routes
}

// setupMiddleware configures global middleware
func (s *Server) setupMiddleware(router *gin.Engine) {
	// Setup logging
	s.setupLogging()

	// Recovery middleware (must be first)
	router.Use(s.recoveryMiddleware())

	// Performance optimizations
	s.enablePerformanceOptimizations(router)

	// Custom logger middleware
	if s.config.Logging.Format == "json" {
		router.Use(s.structuredLogger())
	} else {
		router.Use(s.customLogger())
	}

	// CORS middleware
	router.Use(s.corsMiddleware())

	// Security headers middleware
	router.Use(s.securityHeadersMiddleware())

	// Request ID middleware
	router.Use(s.requestIDMiddleware())

	// Metrics middleware
	router.Use(s.metricsMiddleware())

	// Rate limiting middleware (if configured)
	if s.config.Server.Mode == "release" {
		router.Use(s.rateLimitMiddleware())
	}
}

// corsMiddleware handles CORS headers
func (s *Server) corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")

		// Check if origin is allowed
		allowedOrigin := ""
		for _, configOrigin := range s.config.Server.CORS.AllowedOrigins {
			if configOrigin == "*" || configOrigin == origin {
				allowedOrigin = origin
				break
			}
		}

		// Set CORS headers based on configuration
		if allowedOrigin != "" {
			c.Header("Access-Control-Allow-Origin", allowedOrigin)
		}

		if len(s.config.Server.CORS.AllowedMethods) > 0 {
			c.Header("Access-Control-Allow-Methods", strings.Join(s.config.Server.CORS.AllowedMethods, ", "))
		}

		if len(s.config.Server.CORS.AllowedHeaders) > 0 {
			c.Header("Access-Control-Allow-Headers", strings.Join(s.config.Server.CORS.AllowedHeaders, ", "))
		}

		if len(s.config.Server.CORS.ExposedHeaders) > 0 {
			c.Header("Access-Control-Expose-Headers", strings.Join(s.config.Server.CORS.ExposedHeaders, ", "))
		}

		if s.config.Server.CORS.AllowCredentials {
			c.Header("Access-Control-Allow-Credentials", "true")
		}

		if s.config.Server.CORS.MaxAge > 0 {
			c.Header("Access-Control-Max-Age", fmt.Sprintf("%d", s.config.Server.CORS.MaxAge))
		}

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}

// generateSecureRequestID generates a cryptographically secure request ID
func generateSecureRequestID() string {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		// Fallback to timestamp-based ID if crypto/rand fails
		return fmt.Sprintf("req_%d_%d", time.Now().UnixNano(), time.Now().Unix())
	}
	return fmt.Sprintf("req_%s", hex.EncodeToString(bytes))
}

// validatePluginName validates plugin name to prevent path traversal attacks
func validatePluginName(name string) error {
	if name == "" {
		return fmt.Errorf("plugin name cannot be empty")
	}

	// Check for path traversal attempts
	if strings.Contains(name, "..") || strings.Contains(name, "/") || strings.Contains(name, "\\") {
		return fmt.Errorf("plugin name contains invalid characters")
	}

	// Check length
	if len(name) > 100 {
		return fmt.Errorf("plugin name too long (max 100 characters)")
	}

	// Only allow alphanumeric characters, hyphens, and underscores
	for _, char := range name {
		if !((char >= 'a' && char <= 'z') ||
			(char >= 'A' && char <= 'Z') ||
			(char >= '0' && char <= '9') ||
			char == '-' || char == '_') {
			return fmt.Errorf("plugin name contains invalid character: %c", char)
		}
	}

	return nil
}

// securityHeadersMiddleware adds security headers to responses
func (s *Server) securityHeadersMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Prevent MIME type sniffing
		c.Header("X-Content-Type-Options", "nosniff")

		// Prevent clickjacking
		c.Header("X-Frame-Options", "DENY")

		// Enable XSS protection
		c.Header("X-XSS-Protection", "1; mode=block")

		// Prevent information disclosure
		c.Header("Server", "webhook-bridge")

		// Content Security Policy (basic)
		c.Header("Content-Security-Policy", "default-src 'self'; script-src 'self' 'unsafe-inline'; style-src 'self' 'unsafe-inline'")

		// Referrer Policy
		c.Header("Referrer-Policy", "strict-origin-when-cross-origin")

		c.Next()
	}
}

// requestIDMiddleware adds a unique request ID to each request
func (s *Server) requestIDMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := c.GetHeader("X-Request-ID")
		if requestID == "" {
			requestID = generateSecureRequestID()
		}

		c.Header("X-Request-ID", requestID)
		c.Set("request_id", requestID)
		c.Next()
	}
}

// metricsMiddleware tracks request metrics
func (s *Server) metricsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		atomic.AddInt64(&s.requestCount, 1)

		c.Next()

		// Track execution time
		duration := time.Since(start)
		atomic.AddInt64(&s.totalExecTime, duration.Nanoseconds())

		// Track errors
		if c.Writer.Status() >= 400 {
			atomic.AddInt64(&s.errorCount, 1)
		}

		// Add execution time header
		c.Header("X-Execution-Time", fmt.Sprintf("%.3fms", float64(duration.Nanoseconds())/1e6))
	}
}

// rateLimitMiddleware implements basic rate limiting
func (s *Server) rateLimitMiddleware() gin.HandlerFunc {
	// Simple in-memory rate limiter (for production, use Redis or similar)
	return func(c *gin.Context) {
		// TODO: Implement proper rate limiting
		c.Next()
	}
}

// apiMiddleware adds API-specific middleware
func (s *Server) apiMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Set API version header
		c.Header("X-API-Version", "v1")
		c.Header("Content-Type", "application/json")

		c.Next()
	}
}

// healthCheck handles health check requests
func (s *Server) healthCheck(c *gin.Context) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Check gRPC connection health
	grpcHealthy := true
	grpcMessage := "gRPC connection healthy"

	if s.grpcClient != nil {
		req := &proto.HealthCheckRequest{Service: "webhook-bridge"}
		_, err := s.grpcClient.HealthCheck(ctx, req)
		if err != nil {
			grpcHealthy = false
			grpcMessage = fmt.Sprintf("gRPC connection failed: %v", err)
		}
	} else {
		grpcHealthy = false
		grpcMessage = "gRPC client not initialized"
	}

	status := "healthy"
	httpStatus := http.StatusOK
	if !grpcHealthy {
		status = "degraded"
		httpStatus = http.StatusServiceUnavailable
	}

	versionInfo := version.Get()
	uptime := time.Since(s.startTime)

	c.JSON(httpStatus, gin.H{
		"status":    status,
		"service":   "webhook-bridge",
		"version":   versionInfo.Version,
		"build":     versionInfo.GitCommit,
		"uptime":    uptime.String(),
		"timestamp": time.Now().UTC().Format(time.RFC3339),
		"checks": gin.H{
			"grpc": gin.H{
				"status":  grpcHealthy,
				"message": grpcMessage,
			},
		},
	})
}

// metrics handles metrics requests
func (s *Server) metrics(c *gin.Context) {
	uptime := time.Since(s.startTime)
	requestCount := atomic.LoadInt64(&s.requestCount)
	errorCount := atomic.LoadInt64(&s.errorCount)
	totalExecTime := atomic.LoadInt64(&s.totalExecTime)

	var avgExecTime float64
	if requestCount > 0 {
		avgExecTime = float64(totalExecTime) / float64(requestCount) / 1e6 // Convert to milliseconds
	}

	errorRate := float64(0)
	if requestCount > 0 {
		errorRate = float64(errorCount) / float64(requestCount) * 100
	}

	c.JSON(http.StatusOK, gin.H{
		"service":        "webhook-bridge",
		"version":        version.Get().Version,
		"uptime":         uptime.String(),
		"uptime_seconds": uptime.Seconds(),
		"requests": gin.H{
			"total":        requestCount,
			"errors":       errorCount,
			"error_rate":   fmt.Sprintf("%.2f%%", errorRate),
			"success_rate": fmt.Sprintf("%.2f%%", 100-errorRate),
		},
		"performance": gin.H{
			"avg_response_time_ms": fmt.Sprintf("%.3f", avgExecTime),
			"total_exec_time_ms":   float64(totalExecTime) / 1e6,
		},
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	})
}

// serveRoot serves the root documentation page
func (s *Server) serveRoot(c *gin.Context) {
	versionInfo := version.Get()

	c.JSON(http.StatusOK, gin.H{
		"message": "Webhook Bridge API - Hybrid Architecture",
		"version": versionInfo.Version,
		"build":   versionInfo.GitCommit,
		"docs":    "/docs",
		"health":  "/health",
		"metrics": "/metrics",
		"api": gin.H{
			"v1":     "/api/v1",
			"latest": "/api/latest",
		},
		"endpoints": gin.H{
			"plugins": gin.H{
				"list":    "GET /api/v1/plugins",
				"info":    "GET /api/v1/plugins/{plugin}",
				"execute": "POST|GET|PUT|DELETE /api/v1/webhook/{plugin}",
			},
		},
		"architecture": gin.H{
			"frontend":      "Go HTTP Server (Gin)",
			"backend":       "Python Plugin Executor (gRPC)",
			"communication": "gRPC",
		},
	})
}

// handleNotFound handles 404 errors
func (s *Server) handleNotFound(c *gin.Context) {
	c.JSON(http.StatusNotFound, gin.H{
		"error":     "Not Found",
		"message":   "The requested endpoint does not exist",
		"path":      c.Request.URL.Path,
		"method":    c.Request.Method,
		"timestamp": time.Now().UTC().Format(time.RFC3339),
		"available_endpoints": gin.H{
			"health":  "GET /health",
			"metrics": "GET /metrics",
			"api":     "GET /api/v1/plugins",
			"docs":    "GET /",
		},
	})
}

// listPlugins handles plugin listing requests
func (s *Server) listPlugins(c *gin.Context) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Get filter from query parameter
	filter := c.Query("filter")

	req := &proto.ListPluginsRequest{
		Filter: filter,
	}

	resp, err := s.grpcClient.ListPlugins(ctx, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to list plugins",
			"details": err.Error(),
		})
		return
	}

	// Convert gRPC response to JSON
	plugins := make([]gin.H, len(resp.Plugins))
	for i, plugin := range resp.Plugins {
		plugins[i] = gin.H{
			"name":              plugin.Name,
			"path":              plugin.Path,
			"description":       plugin.Description,
			"supported_methods": plugin.SupportedMethods,
			"is_available":      plugin.IsAvailable,
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"plugins":     plugins,
		"total_count": resp.TotalCount,
		"filter":      filter,
		"timestamp":   time.Now().UTC().Format(time.RFC3339),
	})
}

// getPluginInfo handles plugin info requests
func (s *Server) getPluginInfo(c *gin.Context) {
	pluginName := c.Param("plugin")

	// Validate plugin name
	if err := validatePluginName(pluginName); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":     "Invalid plugin name",
			"details":   err.Error(),
			"plugin":    pluginName,
			"timestamp": time.Now().UTC().Format(time.RFC3339),
		})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	req := &proto.GetPluginInfoRequest{
		PluginName: pluginName,
	}

	resp, err := s.grpcClient.GetPluginInfo(ctx, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":     "Failed to get plugin info",
			"details":   err.Error(),
			"plugin":    pluginName,
			"timestamp": time.Now().UTC().Format(time.RFC3339),
		})
		return
	}

	if !resp.Found {
		c.JSON(http.StatusNotFound, gin.H{
			"error":     "Plugin not found",
			"plugin":    pluginName,
			"message":   fmt.Sprintf("Plugin '%s' does not exist", pluginName),
			"timestamp": time.Now().UTC().Format(time.RFC3339),
		})
		return
	}

	plugin := resp.Plugin
	c.JSON(http.StatusOK, gin.H{
		"name":              plugin.Name,
		"path":              plugin.Path,
		"description":       plugin.Description,
		"supported_methods": plugin.SupportedMethods,
		"is_available":      plugin.IsAvailable,
		"last_modified":     plugin.LastModified,
		"timestamp":         time.Now().UTC().Format(time.RFC3339),
	})
}

// executePlugin handles plugin execution requests
func (s *Server) executePlugin(c *gin.Context) {
	pluginName := c.Param("plugin")
	method := c.Request.Method

	// Validate plugin name
	if err := validatePluginName(pluginName); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":     "Invalid plugin name",
			"details":   err.Error(),
			"plugin":    pluginName,
			"method":    method,
			"timestamp": time.Now().UTC().Format(time.RFC3339),
		})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Get request data
	requestData := make(map[string]string)

	if method == "POST" || method == "PUT" {
		// Limit request body size to 10MB
		c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, 10*1024*1024)

		var jsonData map[string]interface{}
		if err := c.ShouldBindJSON(&jsonData); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "Invalid JSON payload",
				"details": err.Error(),
			})
			return
		}

		// Validate JSON data size
		if len(jsonData) > 1000 {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "Request payload too large",
				"details": "Maximum 1000 fields allowed",
			})
			return
		}

		// Convert to string map for gRPC
		for key, value := range jsonData {
			requestData[key] = fmt.Sprintf("%v", value)
		}
	} else {
		// For GET/DELETE, use query parameters
		for key, values := range c.Request.URL.Query() {
			if len(values) > 0 {
				requestData[key] = values[0]
			}
		}
	}

	// Prepare gRPC request
	req := &proto.ExecutePluginRequest{
		PluginName: pluginName,
		HttpMethod: method,
		Data:       requestData,
	}

	// Execute plugin via gRPC
	resp, err := s.grpcClient.ExecutePlugin(ctx, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to execute plugin",
			"details": err.Error(),
		})
		return
	}

	// Convert gRPC response to HTTP response
	statusCode := int(resp.StatusCode)
	if statusCode == 0 {
		statusCode = http.StatusOK
	}

	responseData := gin.H{
		"plugin":         pluginName,
		"method":         method,
		"status_code":    resp.StatusCode,
		"message":        resp.Message,
		"execution_time": fmt.Sprintf("%.3fms", resp.ExecutionTime*1000),
		"timestamp":      time.Now().UTC().Format(time.RFC3339),
	}

	// Add plugin data if available
	if len(resp.Data) > 0 {
		responseData["data"] = resp.Data
	}

	// Add error if present
	if resp.Error != "" {
		responseData["error"] = resp.Error
	}

	// Add request ID for tracing
	if requestID, exists := c.Get("request_id"); exists {
		responseData["request_id"] = requestID
	}

	c.JSON(statusCode, responseData)
}

// setupExecutionHistoryRoutes sets up execution history API routes
func (s *Server) setupExecutionHistoryRoutes(router *gin.Engine) {
	api := router.Group("/api/v1/executions")
	api.Use(s.apiMiddleware())

	// Get execution history list
	api.GET("", s.handleGetExecutions)

	// Get specific execution details
	api.GET("/:id", s.handleGetExecution)

	// Get execution statistics
	api.GET("/stats", s.handleGetExecutionStats)

	// Get plugin execution statistics
	api.GET("/stats/:plugin", s.handleGetPluginStats)

	// Cleanup old records
	api.DELETE("/cleanup", s.handleCleanupExecutions)

	// Get storage information
	api.GET("/storage/info", s.handleGetStorageInfo)
}

// HTTPServer wraps the HTTP server for service integration
type HTTPServer struct {
	Config *config.Config
	Router *gin.Engine
	server *http.Server
}

// Start starts the HTTP server
func (hs *HTTPServer) Start(ctx context.Context) error {
	hs.server = &http.Server{
		Addr:              hs.Config.GetServerAddress(),
		Handler:           hs.Router,
		ReadHeaderTimeout: 30 * time.Second, // Prevent Slowloris attacks
		ReadTimeout:       60 * time.Second,
		WriteTimeout:      60 * time.Second,
		IdleTimeout:       120 * time.Second,
	}

	// Start server in a goroutine
	go func() {
		if err := hs.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Printf("HTTP server error: %v", err)
		}
	}()

	// Wait for context cancellation
	<-ctx.Done()

	// Graceful shutdown
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	return hs.server.Shutdown(shutdownCtx)
}

// getWorkerStats returns worker pool statistics
func (s *Server) getWorkerStats(c *gin.Context) {
	if s.workerPool == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"error":   "Worker pool not available",
			"message": "Worker pool is not initialized",
		})
		return
	}

	stats := s.workerPool.GetStats()
	c.JSON(http.StatusOK, gin.H{
		"worker_pool": stats,
		"timestamp":   time.Now().UTC().Format(time.RFC3339),
	})
}

// submitJob submits a job to the worker pool
func (s *Server) submitJob(c *gin.Context) {
	if s.workerPool == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"error":   "Worker pool not available",
			"message": "Worker pool is not initialized",
		})
		return
	}

	var jobRequest struct {
		Type     string                 `json:"type" binding:"required"`
		Payload  map[string]interface{} `json:"payload" binding:"required"`
		Priority int                    `json:"priority"`
		MaxRetry int                    `json:"max_retry"`
	}

	if err := c.ShouldBindJSON(&jobRequest); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid job request",
			"details": err.Error(),
		})
		return
	}

	// Create job
	job := &worker.Job{
		Type:     jobRequest.Type,
		Payload:  jobRequest.Payload,
		Priority: jobRequest.Priority,
		MaxRetry: jobRequest.MaxRetry,
	}

	// Submit job
	if err := s.workerPool.Submit(job); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to submit job",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusAccepted, gin.H{
		"message":   "Job submitted successfully",
		"job_id":    job.ID,
		"job_type":  job.Type,
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	})
}
