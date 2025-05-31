package cli

import (
	"fmt"
	"runtime"

	"github.com/spf13/cobra"

	"github.com/loonghao/webhook_bridge/internal/service"
)

// NewServiceCommand creates the service management command
func NewServiceCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "service",
		Short: "Manage system service",
		Long:  "Install, uninstall, start, stop, and manage webhook bridge as a system service",
	}

	// Add subcommands
	cmd.AddCommand(
		newServiceInstallCommand(),
		newServiceUninstallCommand(),
		newServiceStartCommand(),
		newServiceStopCommand(),
		newServiceStatusCommand(),
		newServiceRunCommand(),
	)

	return cmd
}

func newServiceInstallCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "install",
		Short: "Install webhook bridge as system service",
		RunE:  runServiceInstall,
	}

	cmd.Flags().String("name", "webhook-bridge", "Service name")
	cmd.Flags().String("display-name", "Webhook Bridge", "Service display name")
	cmd.Flags().String("description", "Webhook Bridge Service - High-performance webhook processing with Go and Python", "Service description")
	cmd.Flags().String("config", "", "Configuration file path")
	cmd.Flags().Int("workers", 0, "Number of worker processes (default: CPU count)")
	cmd.Flags().String("log-path", "", "Log file path")

	return cmd
}

func newServiceUninstallCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "uninstall",
		Short: "Uninstall webhook bridge system service",
		RunE:  runServiceUninstall,
	}

	cmd.Flags().String("name", "webhook-bridge", "Service name")
	cmd.Flags().String("display-name", "Webhook Bridge", "Service display name")
	cmd.Flags().String("description", "Webhook Bridge Service", "Service description")

	return cmd
}

func newServiceStartCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "start",
		Short: "Start webhook bridge system service",
		RunE:  runServiceStart,
	}

	cmd.Flags().String("name", "webhook-bridge", "Service name")
	cmd.Flags().String("display-name", "Webhook Bridge", "Service display name")
	cmd.Flags().String("description", "Webhook Bridge Service", "Service description")

	return cmd
}

func newServiceStopCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "stop",
		Short: "Stop webhook bridge system service",
		RunE:  runServiceStop,
	}

	cmd.Flags().String("name", "webhook-bridge", "Service name")
	cmd.Flags().String("display-name", "Webhook Bridge", "Service display name")
	cmd.Flags().String("description", "Webhook Bridge Service", "Service description")

	return cmd
}

func newServiceStatusCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "status",
		Short: "Check webhook bridge system service status",
		RunE:  runServiceStatus,
	}

	cmd.Flags().String("name", "webhook-bridge", "Service name")
	cmd.Flags().String("display-name", "Webhook Bridge", "Service display name")
	cmd.Flags().String("description", "Webhook Bridge Service", "Service description")

	return cmd
}

func newServiceRunCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "run",
		Short: "Run webhook bridge service (called by service manager)",
		RunE:  runServiceRun,
	}

	cmd.Flags().String("name", "webhook-bridge", "Service name")
	cmd.Flags().String("display-name", "Webhook Bridge", "Service display name")
	cmd.Flags().String("description", "Webhook Bridge Service", "Service description")
	cmd.Flags().String("config", "", "Configuration file path")
	cmd.Flags().Int("workers", 0, "Number of worker processes (default: CPU count)")
	cmd.Flags().String("log-path", "", "Log file path")

	return cmd
}

func runServiceInstall(cmd *cobra.Command, args []string) error {
	verbose, _ := cmd.Flags().GetBool("verbose")
	name, _ := cmd.Flags().GetString("name")
	displayName, _ := cmd.Flags().GetString("display-name")
	description, _ := cmd.Flags().GetString("description")
	configPath, _ := cmd.Flags().GetString("config")
	workers, _ := cmd.Flags().GetInt("workers")
	logPath, _ := cmd.Flags().GetString("log-path")

	if verbose {
		fmt.Printf("🔧 Installing webhook bridge as system service...\n")
		fmt.Printf("📋 Service name: %s\n", name)
		fmt.Printf("📋 Display name: %s\n", displayName)
		fmt.Printf("📋 Description: %s\n", description)
		fmt.Printf("🖥️  Platform: %s\n", runtime.GOOS)
	}

	cfg := &service.ServiceConfig{
		Name:        name,
		DisplayName: displayName,
		Description: description,
		ConfigPath:  configPath,
		WorkerCount: workers,
		LogPath:     logPath,
	}

	if err := service.InstallService(cfg); err != nil {
		return fmt.Errorf("failed to install service: %w", err)
	}

	fmt.Printf("✅ Service installed successfully!\n")
	fmt.Printf("🚀 Start the service with: webhook-bridge service start\n")
	fmt.Printf("📊 Check status with: webhook-bridge service status\n")

	return nil
}

func runServiceUninstall(cmd *cobra.Command, args []string) error {
	verbose, _ := cmd.Flags().GetBool("verbose")
	name, _ := cmd.Flags().GetString("name")
	displayName, _ := cmd.Flags().GetString("display-name")
	description, _ := cmd.Flags().GetString("description")

	if verbose {
		fmt.Printf("🗑️  Uninstalling webhook bridge system service...\n")
		fmt.Printf("📋 Service name: %s\n", name)
	}

	cfg := &service.ServiceConfig{
		Name:        name,
		DisplayName: displayName,
		Description: description,
	}

	if err := service.UninstallService(cfg); err != nil {
		return fmt.Errorf("failed to uninstall service: %w", err)
	}

	fmt.Printf("✅ Service uninstalled successfully!\n")
	return nil
}

func runServiceStart(cmd *cobra.Command, args []string) error {
	verbose, _ := cmd.Flags().GetBool("verbose")
	name, _ := cmd.Flags().GetString("name")
	displayName, _ := cmd.Flags().GetString("display-name")
	description, _ := cmd.Flags().GetString("description")

	if verbose {
		fmt.Printf("🚀 Starting webhook bridge system service...\n")
		fmt.Printf("📋 Service name: %s\n", name)
	}

	cfg := &service.ServiceConfig{
		Name:        name,
		DisplayName: displayName,
		Description: description,
	}

	if err := service.StartService(cfg); err != nil {
		return fmt.Errorf("failed to start service: %w", err)
	}

	fmt.Printf("✅ Service started successfully!\n")
	fmt.Printf("📊 Check status with: webhook-bridge service status\n")
	return nil
}

func runServiceStop(cmd *cobra.Command, args []string) error {
	verbose, _ := cmd.Flags().GetBool("verbose")
	name, _ := cmd.Flags().GetString("name")
	displayName, _ := cmd.Flags().GetString("display-name")
	description, _ := cmd.Flags().GetString("description")

	if verbose {
		fmt.Printf("🛑 Stopping webhook bridge system service...\n")
		fmt.Printf("📋 Service name: %s\n", name)
	}

	cfg := &service.ServiceConfig{
		Name:        name,
		DisplayName: displayName,
		Description: description,
	}

	if err := service.StopService(cfg); err != nil {
		return fmt.Errorf("failed to stop service: %w", err)
	}

	fmt.Printf("✅ Service stopped successfully!\n")
	return nil
}

func runServiceStatus(cmd *cobra.Command, args []string) error {
	verbose, _ := cmd.Flags().GetBool("verbose")
	name, _ := cmd.Flags().GetString("name")
	displayName, _ := cmd.Flags().GetString("display-name")
	description, _ := cmd.Flags().GetString("description")

	cfg := &service.ServiceConfig{
		Name:        name,
		DisplayName: displayName,
		Description: description,
	}

	status, err := service.GetServiceStatus(cfg)
	if err != nil {
		return fmt.Errorf("failed to get service status: %w", err)
	}

	fmt.Printf("📊 Service Status: %s\n", status)

	if verbose {
		fmt.Printf("📋 Service name: %s\n", name)
		fmt.Printf("📋 Display name: %s\n", displayName)
		fmt.Printf("🖥️  Platform: %s\n", runtime.GOOS)
	}

	switch status {
	case "running":
		fmt.Printf("✅ Service is running\n")
	case "stopped":
		fmt.Printf("⏹️  Service is stopped\n")
		fmt.Printf("🚀 Start with: webhook-bridge service start\n")
	default:
		fmt.Printf("❓ Service status is unknown\n")
	}

	return nil
}

func runServiceRun(cmd *cobra.Command, args []string) error {
	name, _ := cmd.Flags().GetString("name")
	displayName, _ := cmd.Flags().GetString("display-name")
	description, _ := cmd.Flags().GetString("description")
	configPath, _ := cmd.Flags().GetString("config")
	workers, _ := cmd.Flags().GetInt("workers")
	logPath, _ := cmd.Flags().GetString("log-path")

	cfg := &service.ServiceConfig{
		Name:        name,
		DisplayName: displayName,
		Description: description,
		ConfigPath:  configPath,
		WorkerCount: workers,
		LogPath:     logPath,
	}

	// This is called by the service manager
	return service.RunService(cfg)
}
