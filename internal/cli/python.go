package cli

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/loonghao/webhook_bridge/internal/config"
	"github.com/loonghao/webhook_bridge/internal/python"
)

// NewPythonCommand creates the python command (integrated from python-manager.exe)
func NewPythonCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "python",
		Short: "Python environment management (integrated from python-manager.exe)",
		Long: `Manage Python interpreter and environment for webhook bridge.
This command integrates the functionality previously provided by python-manager.exe.

Features:
- Python interpreter discovery and validation
- Virtual environment management
- Package installation and dependency management
- Environment information and diagnostics`,
		RunE: runPython,
	}

	// Subcommands
	cmd.AddCommand(
		newPythonInfoCommand(),
		newPythonValidateCommand(),
		newPythonInstallCommand(),
		newPythonExecutorCommand(),
	)

	// Global flags for python command
	cmd.PersistentFlags().String("strategy", "", "Python discovery strategy (auto, uv, path, custom)")
	cmd.PersistentFlags().String("config", "", "Configuration file path")

	return cmd
}

func runPython(cmd *cobra.Command, args []string) error {
	verbose, _ := cmd.Flags().GetBool("verbose")
	strategy, _ := cmd.Flags().GetString("strategy")
	configPath, _ := cmd.Flags().GetString("config")

	if verbose {
		fmt.Printf("🐍 Python Manager Tool (Integrated)\n")
		fmt.Printf("===================================\n")
	}

	// Load configuration
	var cfg *config.Config
	var err error

	if configPath != "" {
		cfg, err = config.LoadFromFile(configPath)
	} else {
		cfg, err = config.Load()
	}
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	// Override interpreter if provided
	if strategy != "" {
		cfg.Python.Interpreter = strategy
	}

	// Create Python manager
	manager := python.NewManager(&cfg.Python)

	if verbose {
		fmt.Printf("Using Python configuration:\n")
		fmt.Printf("  Interpreter: %s\n", cfg.Python.Interpreter)
		fmt.Printf("  UV Enabled: %v\n", cfg.Python.UV.Enabled)
		fmt.Printf("  UV Project Path: %s\n", cfg.Python.UV.ProjectPath)
		fmt.Printf("  UV Venv Name: %s\n", cfg.Python.UV.VenvName)
		fmt.Printf("  Plugin Dirs: %v\n", cfg.Python.PluginDirs)
		fmt.Printf("\n")
	}

	// If no subcommand specified, show basic info
	fmt.Println("🐍 Python Manager Tool")
	fmt.Println("======================")

	// Discover interpreter
	interpreterPath, err := manager.DiscoverInterpreter()
	if err != nil {
		return fmt.Errorf("failed to discover Python interpreter: %w", err)
	}

	fmt.Printf("✅ Python interpreter found: %s\n", interpreterPath)
	fmt.Printf("📋 Interpreter used: %s\n", cfg.Python.Interpreter)

	// Get basic info
	if interpreterInfo, err := manager.GetInterpreterInfo(); err == nil {
		fmt.Printf("🔢 Python version: %s\n", interpreterInfo.Version)
		fmt.Printf("🏠 Virtual environment: %v\n", interpreterInfo.IsVirtual)
		if interpreterInfo.VenvPath != "" {
			fmt.Printf("📁 Venv path: %s\n", interpreterInfo.VenvPath)
		}

		// Show key capabilities
		fmt.Println("🔧 Key capabilities:")
		keyCapabilities := []string{"grpc", "grpcio", "fastapi", "requests", "async_support"}
		for _, cap := range keyCapabilities {
			status := "❌"
			if interpreterInfo.Capabilities[cap] {
				status = "✅"
			}
			fmt.Printf("  %s %s\n", status, cap)
		}
	}

	fmt.Printf("\n💡 Use 'webhook-bridge python --help' to see available subcommands\n")
	return nil
}

func newPythonInfoCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "info",
		Short: "Show detailed Python interpreter information",
		RunE:  runPythonInfo,
	}
	return cmd
}

func runPythonInfo(cmd *cobra.Command, args []string) error {
	verbose, _ := cmd.Flags().GetBool("verbose")
	strategy, _ := cmd.Flags().GetString("strategy")
	configPath, _ := cmd.Flags().GetString("config")

	// Load configuration
	var cfg *config.Config
	var err error

	if configPath != "" {
		cfg, err = config.LoadFromFile(configPath)
	} else {
		cfg, err = config.Load()
	}
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	// Override interpreter if provided
	if strategy != "" {
		cfg.Python.Interpreter = strategy
	}

	// Create Python manager
	manager := python.NewManager(&cfg.Python)

	if verbose {
		fmt.Println("🔍 Discovering Python interpreter...")
	}

	interpreterInfo, err := manager.GetInterpreterInfo()
	if err != nil {
		return fmt.Errorf("failed to get interpreter info: %w", err)
	}

	// Pretty print JSON
	jsonData, err := json.MarshalIndent(interpreterInfo, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal interpreter info: %w", err)
	}

	fmt.Println("📋 Python Interpreter Information:")
	fmt.Println(string(jsonData))
	return nil
}

func newPythonValidateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "validate",
		Short: "Validate Python environment",
		RunE:  runPythonValidate,
	}
	return cmd
}

func runPythonValidate(cmd *cobra.Command, args []string) error {
	strategy, _ := cmd.Flags().GetString("strategy")
	configPath, _ := cmd.Flags().GetString("config")

	// Load configuration
	var cfg *config.Config
	var err error

	if configPath != "" {
		cfg, err = config.LoadFromFile(configPath)
	} else {
		cfg, err = config.Load()
	}
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	// Override interpreter if provided
	if strategy != "" {
		cfg.Python.Interpreter = strategy
	}

	// Create Python manager
	manager := python.NewManager(&cfg.Python)

	fmt.Println("✅ Validating Python environment...")
	if err := manager.ValidateEnvironment(); err != nil {
		return fmt.Errorf("environment validation failed: %w", err)
	}
	fmt.Println("✅ Environment validation passed!")
	return nil
}

func newPythonInstallCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "install [packages...]",
		Short: "Install Python packages",
		Long:  "Install Python packages using the configured package manager",
		RunE:  runPythonInstall,
	}
	cmd.Flags().String("packages", "", "Comma-separated list of packages to install")
	return cmd
}

func runPythonInstall(cmd *cobra.Command, args []string) error {
	strategy, _ := cmd.Flags().GetString("strategy")
	configPath, _ := cmd.Flags().GetString("config")
	packagesFlag, _ := cmd.Flags().GetString("packages")

	// Determine packages to install
	var packages []string
	if packagesFlag != "" {
		packages = parsePackageList(packagesFlag)
	}
	if len(args) > 0 {
		packages = append(packages, args...)
	}

	if len(packages) == 0 {
		return fmt.Errorf("no packages specified")
	}

	// Load configuration
	var cfg *config.Config
	var err error

	if configPath != "" {
		cfg, err = config.LoadFromFile(configPath)
	} else {
		cfg, err = config.Load()
	}
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	// Override interpreter if provided
	if strategy != "" {
		cfg.Python.Interpreter = strategy
	}

	// Create Python manager
	manager := python.NewManager(&cfg.Python)

	fmt.Printf("📦 Installing packages: %v\n", packages)
	if err := manager.InstallDependencies(packages); err != nil {
		return fmt.Errorf("failed to install dependencies: %w", err)
	}
	fmt.Println("✅ Packages installed successfully!")
	return nil
}

func newPythonExecutorCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "executor",
		Short: "Start Python executor service",
		Long:  "Start the Python executor gRPC service for plugin execution",
		RunE:  runPythonExecutor,
	}
	cmd.Flags().String("host", "127.0.0.1", "Host to bind the executor to")
	cmd.Flags().Int("port", 50051, "Port to bind the executor to")
	cmd.Flags().StringSlice("plugin-dirs", nil, "Additional plugin directories")
	return cmd
}

func runPythonExecutor(cmd *cobra.Command, args []string) error {
	host, _ := cmd.Flags().GetString("host")
	port, _ := cmd.Flags().GetInt("port")
	pluginDirs, _ := cmd.Flags().GetStringSlice("plugin-dirs")
	verbose, _ := cmd.Flags().GetBool("verbose")

	if verbose {
		fmt.Printf("🐍 Starting Python executor service...\n")
		fmt.Printf("🌐 Host: %s\n", host)
		fmt.Printf("🌐 Port: %d\n", port)
		fmt.Printf("📁 Plugin directories: %v\n", pluginDirs)
	}

	// This would start the Python executor service
	// For now, we'll show instructions on how to start it manually
	fmt.Printf("🐍 Python Executor Service\n")
	fmt.Printf("==========================\n")
	fmt.Printf("To start the Python executor service, run:\n")
	fmt.Printf("  python python_executor/main.py --host %s --port %d\n", host, port)
	
	if len(pluginDirs) > 0 {
		fmt.Printf("  --plugin-dirs %s\n", strings.Join(pluginDirs, ","))
	}
	
	fmt.Printf("\nAlternatively, use the 'start' command for full service management:\n")
	fmt.Printf("  webhook-bridge start\n")

	return nil
}

// parsePackageList parses a comma-separated list of packages
func parsePackageList(packages string) []string {
	if packages == "" {
		return nil
	}
	
	parts := strings.Split(packages, ",")
	result := make([]string, 0, len(parts))
	
	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}
	
	return result
}
