package api

import (
	"github.com/gin-gonic/gin"

	"github.com/loonghao/webhook_bridge/internal/config"
)

// Router represents the unified API router
type Router struct {
	engine   *gin.Engine
	config   *config.Config
	handlers map[string]Handler
}

// Handler interface for all API handlers
type Handler interface {
	RegisterRoutes(group *gin.RouterGroup)
}

// NewRouter creates a new unified API router
func NewRouter(cfg *config.Config) *Router {
	// Set Gin mode based on config
	if cfg.Server.Mode == "debug" {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}

	engine := gin.New()

	// Apply global middleware
	engine.Use(RequestIDMiddleware())
	engine.Use(CORSMiddleware())
	engine.Use(LoggingMiddleware())
	engine.Use(ErrorHandlingMiddleware())
	engine.Use(ValidationMiddleware())

	return &Router{
		engine:   engine,
		config:   cfg,
		handlers: make(map[string]Handler),
	}
}

// RegisterHandler registers a handler with a specific prefix
func (r *Router) RegisterHandler(prefix string, handler Handler) {
	r.handlers[prefix] = handler

	// Create route group for this handler
	group := r.engine.Group(prefix)
	handler.RegisterRoutes(group)
}

// SetupRoutes sets up all API routes
func (r *Router) SetupRoutes() {
	// Health check endpoint (no prefix)
	r.engine.GET("/health", r.healthCheck)
	r.engine.GET("/ping", r.ping)

	// API version info
	r.engine.GET("/api", r.apiInfo)
	r.engine.GET("/api/v1", r.apiV1Info)
	r.engine.GET("/api/latest", r.apiLatestInfo)
}

// GetEngine returns the Gin engine
func (r *Router) GetEngine() *gin.Engine {
	return r.engine
}

// Health check endpoints
func (r *Router) healthCheck(c *gin.Context) {
	Success(c, gin.H{
		"status":    "healthy",
		"timestamp": gin.H{},
		"version":   "2.0.0-hybrid",
		"uptime":    "running",
	}, "Service is healthy")
}

func (r *Router) ping(c *gin.Context) {
	Success(c, gin.H{
		"message": "pong",
	})
}

// API info endpoints
func (r *Router) apiInfo(c *gin.Context) {
	Success(c, gin.H{
		"name":        "Webhook Bridge API",
		"version":     "2.0.0-hybrid",
		"description": "Unified API for webhook bridge management",
		"endpoints": gin.H{
			"v1":        "/api/v1",
			"latest":    "/api/latest",
			"dashboard": "/api/dashboard",
		},
	})
}

func (r *Router) apiV1Info(c *gin.Context) {
	Success(c, gin.H{
		"version":     "1.0.0",
		"status":      "stable",
		"description": "Version 1 API endpoints",
		"endpoints": gin.H{
			"plugins":  "/api/v1/plugins",
			"webhooks": "/api/v1/webhook",
		},
	})
}

func (r *Router) apiLatestInfo(c *gin.Context) {
	Success(c, gin.H{
		"version":     "2.0.0",
		"status":      "stable",
		"description": "Latest API endpoints (alias for v1)",
		"endpoints": gin.H{
			"plugins":  "/api/latest/plugins",
			"webhooks": "/api/latest/webhook",
		},
	})
}

// Route groups for different API versions and features
const (
	// API version prefixes
	APIv1Prefix     = "/api/v1"
	APILatestPrefix = "/api/latest"
	DashboardPrefix = "/api/dashboard"

	// Feature prefixes
	WebhookPrefix      = "/webhook"
	PluginPrefix       = "/plugins"
	SystemPrefix       = "/system"
	ConfigPrefix       = "/config"
	LogsPrefix         = "/logs"
	WorkersPrefix      = "/workers"
	InterpretersPrefix = "/interpreters"
	ConnectionPrefix   = "/connection"
)

// Common route patterns
type RoutePattern struct {
	Method  string
	Path    string
	Handler gin.HandlerFunc
}

// BuildRoutes helper function to build routes from patterns
func BuildRoutes(group *gin.RouterGroup, patterns []RoutePattern) {
	for _, pattern := range patterns {
		switch pattern.Method {
		case "GET":
			group.GET(pattern.Path, pattern.Handler)
		case "POST":
			group.POST(pattern.Path, pattern.Handler)
		case "PUT":
			group.PUT(pattern.Path, pattern.Handler)
		case "DELETE":
			group.DELETE(pattern.Path, pattern.Handler)
		case "PATCH":
			group.PATCH(pattern.Path, pattern.Handler)
		}
	}
}

// Validation helpers
func ValidateRequired(c *gin.Context, field string, value interface{}) bool {
	if value == nil || value == "" {
		BadRequest(c, "Validation failed", field+" is required")
		return false
	}
	return true
}

func ValidateJSON(c *gin.Context, target interface{}) bool {
	if err := c.ShouldBindJSON(target); err != nil {
		BadRequest(c, "Invalid JSON format", err.Error())
		return false
	}
	return true
}

func ValidateQuery(c *gin.Context, target interface{}) bool {
	if err := c.ShouldBindQuery(target); err != nil {
		BadRequest(c, "Invalid query parameters", err.Error())
		return false
	}
	return true
}

func ValidateURI(c *gin.Context, target interface{}) bool {
	if err := c.ShouldBindUri(target); err != nil {
		BadRequest(c, "Invalid URI parameters", err.Error())
		return false
	}
	return true
}
