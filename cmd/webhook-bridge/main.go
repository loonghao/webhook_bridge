package main

import (
	"log"

	"github.com/spf13/cobra"

	"github.com/loonghao/webhook_bridge/internal/cli"
	_ "github.com/loonghao/webhook_bridge/web-nextjs" // Import to ensure embed directives are executed
)

var (
	version   = "dev"
	buildTime = "unknown"
	goVersion = "unknown"
)

func main() {
	// Set version information
	cli.SetVersionInfo(version, buildTime, goVersion)

	// Create root command
	rootCmd := &cobra.Command{
		Use:   "webhook-bridge",
		Short: "Webhook Bridge - High-performance webhook processing service",
		Long: `Webhook Bridge is a high-performance webhook processing service
that combines Go's performance with Python's flexibility.

This unified CLI tool provides everything you need to build, deploy,
and manage your webhook bridge service.`,
		Version: version,
	}

	// Add subcommands
	rootCmd.AddCommand(
		// Core commands
		cli.NewStartCommand(),   // Start all services (frontend, backend, Python executor)
		cli.NewServeCommand(),   // Standalone server (no Python dependencies)
		cli.NewServiceCommand(), // System service management

		// Management commands
		cli.NewStopCommand(),   // Stop services
		cli.NewStatusCommand(), // Show service status
		cli.NewConfigCommand(), // Configuration management

		// Development tools
		cli.NewPythonCommand(),    // Python environment management
		cli.NewBuildCommand(),     // Build command
		cli.NewTestCommand(),      // Test command
		cli.NewCleanCommand(),     // Clean command
		cli.NewDashboardCommand(), // Dashboard management

		// Other
		cli.NewVersionCommand(), // Version information
	)

	// Global flags
	rootCmd.PersistentFlags().BoolP("verbose", "v", false, "Enable verbose output")
	rootCmd.PersistentFlags().StringP("config", "c", "", "Configuration file path")

	// Execute command
	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}
