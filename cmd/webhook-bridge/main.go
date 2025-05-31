package main

import (
	"log"

	"github.com/spf13/cobra"
	"github.com/loonghao/webhook_bridge/internal/cli"
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
		cli.NewServeCommand(),    // Standalone server (no dependencies)
		cli.NewServiceCommand(),  // System service management
		cli.NewBuildCommand(),
		cli.NewStartCommand(),    // Full development mode
		cli.NewStopCommand(),
		cli.NewStatusCommand(),
		cli.NewDashboardCommand(),
		cli.NewDeployCommand(),
		cli.NewTestCommand(),
		cli.NewCleanCommand(),
		cli.NewConfigCommand(),
		cli.NewVersionCommand(),
	)

	// Global flags
	rootCmd.PersistentFlags().BoolP("verbose", "v", false, "Enable verbose output")
	rootCmd.PersistentFlags().StringP("config", "c", "", "Configuration file path")

	// Execute command
	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}
