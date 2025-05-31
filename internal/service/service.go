package service

import (
	"context"
	"fmt"
	"log"
	"os"
	"runtime"

	"github.com/gin-gonic/gin"
	"github.com/kardianos/service"
	"github.com/loonghao/webhook_bridge/internal/config"
	"github.com/loonghao/webhook_bridge/internal/server"
	"github.com/loonghao/webhook_bridge/internal/worker"
)

// WebhookBridgeService represents the system service
type WebhookBridgeService struct {
	config     *config.Config
	server     *server.Server
	workerPool *worker.Pool
	logger     service.Logger
	ctx        context.Context
	cancel     context.CancelFunc
}

// ServiceConfig represents service configuration
type ServiceConfig struct {
	Name        string
	DisplayName string
	Description string
	ConfigPath  string
	WorkerCount int
	LogPath     string
}

// NewService creates a new webhook bridge service
func NewService(cfg *ServiceConfig) (*WebhookBridgeService, error) {
	// Load application configuration
	var appConfig *config.Config
	var err error

	if cfg.ConfigPath != "" {
		appConfig, err = config.LoadFromFile(cfg.ConfigPath)
	} else {
		appConfig, err = config.Load()
	}

	if err != nil {
		appConfig = config.Default()
	}

	// Assign ports
	if err := appConfig.AssignPorts(); err != nil {
		return nil, fmt.Errorf("failed to assign ports: %w", err)
	}

	// Create context
	ctx, cancel := context.WithCancel(context.Background())

	return &WebhookBridgeService{
		config:     appConfig,
		ctx:        ctx,
		cancel:     cancel,
		workerPool: worker.NewPool(cfg.WorkerCount),
	}, nil
}

// Start implements service.Interface
func (s *WebhookBridgeService) Start(svc service.Service) error {
	if s.logger != nil {
		s.logger.Info("Starting webhook bridge service...")
	}

	// Start in a goroutine to avoid blocking
	go s.run()
	return nil
}

// Stop implements service.Interface
func (s *WebhookBridgeService) Stop(svc service.Service) error {
	if s.logger != nil {
		s.logger.Info("Stopping webhook bridge service...")
	}

	// Cancel context to stop all components
	s.cancel()

	// Stop worker pool
	if s.workerPool != nil {
		s.workerPool.Stop()
	}

	// Stop server
	if s.server != nil {
		s.server.Stop()
	}

	if s.logger != nil {
		s.logger.Info("Webhook bridge service stopped")
	}

	return nil
}

// run is the main service loop
func (s *WebhookBridgeService) run() {
	// Create and start server
	s.server = server.New(s.config)

	// Start gRPC connection
	if err := s.server.Start(); err != nil {
		if s.logger != nil {
			s.logger.Warningf("Failed to connect to Python executor: %v", err)
			s.logger.Info("Server will start in API-only mode")
		}
	}

	// Start worker pool
	s.workerPool.Start(s.ctx)

	// Setup HTTP server
	router := gin.New()
	s.server.SetupRoutes(router)

	// Start HTTP server
	httpServer := &server.HTTPServer{
		Config: s.config,
		Router: router,
	}

	if err := httpServer.Start(s.ctx); err != nil {
		if s.logger != nil {
			s.logger.Errorf("Failed to start HTTP server: %v", err)
		}
		return
	}

	if s.logger != nil {
		s.logger.Infof("Webhook bridge service running on %s", s.config.GetServerAddress())
		s.logger.Infof("Worker pool started with %d workers", s.workerPool.Size())
	}

	// Wait for context cancellation
	<-s.ctx.Done()
}

// InstallService installs the service to the system
func InstallService(cfg *ServiceConfig) error {
	svcConfig := &service.Config{
		Name:        cfg.Name,
		DisplayName: cfg.DisplayName,
		Description: cfg.Description,
		Arguments:   []string{"service", "run"},
	}

	// Set service executable path
	execPath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("failed to get executable path: %w", err)
	}
	svcConfig.Executable = execPath

	// Add configuration arguments
	if cfg.ConfigPath != "" {
		svcConfig.Arguments = append(svcConfig.Arguments, "--config", cfg.ConfigPath)
	}
	if cfg.WorkerCount > 0 {
		svcConfig.Arguments = append(svcConfig.Arguments, "--workers", fmt.Sprintf("%d", cfg.WorkerCount))
	}

	// Platform-specific configuration
	switch runtime.GOOS {
	case "windows":
		svcConfig.Option = service.KeyValue{
			"StartType":   "automatic",
			"Description": cfg.Description,
		}
	case "linux":
		svcConfig.Dependencies = []string{
			"Requires=network.target",
			"After=network-online.target syslog.target",
		}
		svcConfig.Option = service.KeyValue{
			"LimitNOFILE": 1048576,
		}
	case "darwin":
		svcConfig.Option = service.KeyValue{
			"KeepAlive": true,
			"RunAtLoad": true,
		}
	}

	// Create service
	webhookService, err := NewService(cfg)
	if err != nil {
		return fmt.Errorf("failed to create service: %w", err)
	}

	svc, err := service.New(webhookService, svcConfig)
	if err != nil {
		return fmt.Errorf("failed to create system service: %w", err)
	}

	// Install service
	if err := svc.Install(); err != nil {
		return fmt.Errorf("failed to install service: %w", err)
	}

	fmt.Printf("✅ Service '%s' installed successfully\n", cfg.Name)
	return nil
}

// UninstallService removes the service from the system
func UninstallService(cfg *ServiceConfig) error {
	svcConfig := &service.Config{
		Name:        cfg.Name,
		DisplayName: cfg.DisplayName,
		Description: cfg.Description,
	}

	webhookService, err := NewService(cfg)
	if err != nil {
		return fmt.Errorf("failed to create service: %w", err)
	}

	svc, err := service.New(webhookService, svcConfig)
	if err != nil {
		return fmt.Errorf("failed to create system service: %w", err)
	}

	// Stop service if running
	if err := svc.Stop(); err != nil {
		// Ignore error if service is not running
	}

	// Uninstall service
	if err := svc.Uninstall(); err != nil {
		return fmt.Errorf("failed to uninstall service: %w", err)
	}

	fmt.Printf("✅ Service '%s' uninstalled successfully\n", cfg.Name)
	return nil
}

// StartService starts the installed service
func StartService(cfg *ServiceConfig) error {
	svcConfig := &service.Config{
		Name:        cfg.Name,
		DisplayName: cfg.DisplayName,
		Description: cfg.Description,
	}

	webhookService, err := NewService(cfg)
	if err != nil {
		return fmt.Errorf("failed to create service: %w", err)
	}

	svc, err := service.New(webhookService, svcConfig)
	if err != nil {
		return fmt.Errorf("failed to create system service: %w", err)
	}

	if err := svc.Start(); err != nil {
		return fmt.Errorf("failed to start service: %w", err)
	}

	fmt.Printf("✅ Service '%s' started successfully\n", cfg.Name)
	return nil
}

// StopService stops the running service
func StopService(cfg *ServiceConfig) error {
	svcConfig := &service.Config{
		Name:        cfg.Name,
		DisplayName: cfg.DisplayName,
		Description: cfg.Description,
	}

	webhookService, err := NewService(cfg)
	if err != nil {
		return fmt.Errorf("failed to create service: %w", err)
	}

	svc, err := service.New(webhookService, svcConfig)
	if err != nil {
		return fmt.Errorf("failed to create system service: %w", err)
	}

	if err := svc.Stop(); err != nil {
		return fmt.Errorf("failed to stop service: %w", err)
	}

	fmt.Printf("✅ Service '%s' stopped successfully\n", cfg.Name)
	return nil
}

// GetServiceStatus returns the service status
func GetServiceStatus(cfg *ServiceConfig) (string, error) {
	svcConfig := &service.Config{
		Name:        cfg.Name,
		DisplayName: cfg.DisplayName,
		Description: cfg.Description,
	}

	webhookService, err := NewService(cfg)
	if err != nil {
		return "", fmt.Errorf("failed to create service: %w", err)
	}

	svc, err := service.New(webhookService, svcConfig)
	if err != nil {
		return "", fmt.Errorf("failed to create system service: %w", err)
	}

	status, err := svc.Status()
	if err != nil {
		return "", fmt.Errorf("failed to get service status: %w", err)
	}

	switch status {
	case service.StatusRunning:
		return "running", nil
	case service.StatusStopped:
		return "stopped", nil
	case service.StatusUnknown:
		return "unknown", nil
	default:
		return "unknown", nil
	}
}

// RunService runs the service (called by service manager)
func RunService(cfg *ServiceConfig) error {
	webhookService, err := NewService(cfg)
	if err != nil {
		return fmt.Errorf("failed to create service: %w", err)
	}

	svcConfig := &service.Config{
		Name:        cfg.Name,
		DisplayName: cfg.DisplayName,
		Description: cfg.Description,
	}

	svc, err := service.New(webhookService, svcConfig)
	if err != nil {
		return fmt.Errorf("failed to create system service: %w", err)
	}

	// Setup logger
	logger, err := svc.Logger(nil)
	if err != nil {
		log.Printf("Failed to create service logger: %v", err)
	} else {
		webhookService.logger = logger
	}

	// Run service
	if err := svc.Run(); err != nil {
		if logger != nil {
			logger.Errorf("Service run error: %v", err)
		}
		return err
	}

	return nil
}
