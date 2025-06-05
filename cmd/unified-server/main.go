package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/loonghao/webhook_bridge/internal/api"
	"github.com/loonghao/webhook_bridge/internal/api/handlers"
	"github.com/loonghao/webhook_bridge/internal/config"
	"github.com/loonghao/webhook_bridge/internal/grpc"
	"github.com/loonghao/webhook_bridge/internal/service"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Initialize services
	services, err := initializeServices(cfg)
	if err != nil {
		log.Fatalf("Failed to initialize services: %v", err)
	}
	defer services.cleanup()

	// Create unified API router
	router := api.NewRouter(cfg)

	// Setup all routes
	setupRoutes(router, cfg, services)

	// Create HTTP server
	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", cfg.Server.Port),
		Handler: router.GetEngine(),
	}

	// Start server in a goroutine
	go func() {
		log.Printf("🚀 Unified Webhook Bridge Server starting on port %d", cfg.Server.Port)
		log.Printf("📊 Dashboard: http://localhost:%d", cfg.Server.Port)
		log.Printf("🔗 API: http://localhost:%d/api", cfg.Server.Port)
		log.Printf("📖 API Docs: http://localhost:%d/api/v1", cfg.Server.Port)

		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("🛑 Shutting down server...")

	// Graceful shutdown with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Printf("Server forced to shutdown: %v", err)
	} else {
		log.Println("✅ Server exited gracefully")
	}
}

// Services holds all initialized services
type Services struct {
	grpcClient    *grpc.Client
	workerPool    service.WorkerPool
	logManager    service.LogManager
	statsManager  service.StatsManager
	connectionMgr service.ConnectionManager
}

func (s *Services) cleanup() {
	if s.grpcClient != nil {
		s.grpcClient.Close()
	}
	if s.workerPool != nil {
		s.workerPool.Stop()
	}
}

// initializeServices initializes all required services
func initializeServices(cfg *config.Config) (*Services, error) {
	services := &Services{}

	// Initialize gRPC client
	grpcClient := grpc.NewClient(&cfg.Executor)
	if grpcClient == nil {
		log.Printf("⚠️  Failed to initialize gRPC client")
		log.Printf("🔄 Server will continue without Python executor connection")
	} else {
		// Attempt to connect to Python executor
		if err := grpcClient.Connect(); err != nil {
			log.Printf("⚠️  Failed to connect to Python executor: %v", err)
			log.Printf("🔄 Server will continue without Python executor connection")
			grpcClient = nil
		} else {
			services.grpcClient = grpcClient
			log.Printf("✅ gRPC client connected to %s:%d", cfg.Executor.Host, cfg.Executor.Port)
		}
	}

	// Initialize worker pool
	workerCount := 4 // Default worker count
	workerPool := service.NewWorkerPool(workerCount)
	workerPool.Start()
	services.workerPool = workerPool
	log.Printf("✅ Worker pool started with %d workers", workerCount)

	// Initialize log manager
	logManager := service.NewLogManager(cfg.Logging.File, cfg.Logging.Level)
	services.logManager = logManager
	log.Printf("✅ Log manager initialized")

	// Initialize stats manager
	statsManager := service.NewStatsManager()
	services.statsManager = statsManager
	log.Printf("✅ Stats manager initialized")

	// Initialize connection manager
	connectionMgr := service.NewConnectionManager(grpcClient, cfg)
	services.connectionMgr = connectionMgr
	log.Printf("✅ Connection manager initialized")

	return services, nil
}

// setupRoutes configures all API routes and handlers
func setupRoutes(router *api.Router, cfg *config.Config, services *Services) {
	// Setup basic routes
	router.SetupRoutes()

	// Create and register Dashboard API handler
	dashboardHandler := handlers.NewDashboardHandler(
		cfg,
		services.grpcClient,
		services.workerPool,
		services.logManager,
		services.statsManager,
		services.connectionMgr,
	)
	router.RegisterHandler(api.DashboardPrefix, dashboardHandler)

	// Create and register Webhook API handler
	webhookHandler := handlers.NewWebhookHandler(
		cfg,
		services.grpcClient,
		services.workerPool,
		services.statsManager,
	)
	router.RegisterHandler(api.APIv1Prefix, webhookHandler)
	router.RegisterHandler(api.APILatestPrefix, webhookHandler)

	// TODO: Setup modern dashboard web interface
	// dashboardWeb := modern.NewDashboard(cfg)
	// dashboardWeb.SetupRoutes(router.GetEngine())

	log.Printf("✅ All routes configured successfully")
	log.Printf("📋 Available endpoints:")
	log.Printf("   • Dashboard: /")
	log.Printf("   • API Info: /api")
	log.Printf("   • Dashboard API: %s", api.DashboardPrefix)
	log.Printf("   • Webhook API v1: %s", api.APIv1Prefix)
	log.Printf("   • Webhook API Latest: %s", api.APILatestPrefix)
	log.Printf("   • Health Check: /health")
	log.Printf("   • Ping: /ping")
}

// Example of how to add custom middleware or routes
func setupCustomRoutes(router *api.Router) {
	engine := router.GetEngine()

	// Add custom routes if needed
	engine.GET("/version", func(c *gin.Context) {
		api.Success(c, gin.H{
			"version":    "2.0.0-unified",
			"build_time": time.Now().Format(time.RFC3339),
			"go_version": "1.21+",
		}, "Version information")
	})

	// Add metrics endpoint
	engine.GET("/metrics", func(c *gin.Context) {
		// TODO: Implement Prometheus metrics
		api.Success(c, gin.H{
			"message": "Metrics endpoint",
			"status":  "not_implemented",
		}, "Metrics not yet implemented")
	})
}

// Health check with detailed status
func detailedHealthCheck(services *Services) gin.HandlerFunc {
	return func(c *gin.Context) {
		status := gin.H{
			"status":    "healthy",
			"timestamp": time.Now().Format(time.RFC3339),
			"version":   "2.0.0-unified",
			"services": gin.H{
				"grpc": gin.H{
					"connected": services.grpcClient != nil && services.grpcClient.IsConnected(),
					"status":    "ok",
				},
				"workers": gin.H{
					"count":  services.workerPool.GetWorkerCount(),
					"active": services.workerPool.GetActiveWorkerCount(),
					"status": "ok",
				},
				"logs": gin.H{
					"status": "ok",
				},
				"stats": gin.H{
					"status": "ok",
				},
			},
		}

		api.Success(c, status, "Detailed health check")
	}
}
