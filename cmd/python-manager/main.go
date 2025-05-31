package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"

	"github.com/loonghao/webhook_bridge/internal/config"
	"github.com/loonghao/webhook_bridge/internal/python"
)

func main() {
	var (
		strategy = flag.String("strategy", "", "Python discovery strategy (auto, uv, path, custom)")
		validate = flag.Bool("validate", false, "Validate Python environment")
		install  = flag.String("install", "", "Comma-separated list of packages to install")
		info     = flag.Bool("info", false, "Show detailed interpreter information")
		verbose  = flag.Bool("verbose", false, "Verbose output")
	)
	flag.Parse()

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Override interpreter if provided
	if *strategy != "" {
		cfg.Python.Interpreter = *strategy
	}

	// Create Python manager
	manager := python.NewManager(&cfg.Python)

	if *verbose {
		fmt.Printf("Using Python configuration:\n")
		fmt.Printf("  Interpreter: %s\n", cfg.Python.Interpreter)
		fmt.Printf("  UV Enabled: %v\n", cfg.Python.UV.Enabled)
		fmt.Printf("  UV Project Path: %s\n", cfg.Python.UV.ProjectPath)
		fmt.Printf("  UV Venv Name: %s\n", cfg.Python.UV.VenvName)
		fmt.Printf("  Venv Path: %s\n", cfg.Python.VenvPath)
		fmt.Printf("  Plugin Dirs: %v\n", cfg.Python.PluginDirs)
		fmt.Println()
	}

	// Show interpreter information
	if *info {
		fmt.Println("🔍 Discovering Python interpreter...")
		interpreterInfo, err := manager.GetInterpreterInfo()
		if err != nil {
			log.Fatalf("Failed to get interpreter info: %v", err)
		}

		// Pretty print JSON
		jsonData, err := json.MarshalIndent(interpreterInfo, "", "  ")
		if err != nil {
			log.Fatalf("Failed to marshal interpreter info: %v", err)
		}

		fmt.Println("📋 Python Interpreter Information:")
		fmt.Println(string(jsonData))
		fmt.Println()
	}

	// Validate environment
	if *validate {
		fmt.Println("✅ Validating Python environment...")
		if err := manager.ValidateEnvironment(); err != nil {
			log.Fatalf("Environment validation failed: %v", err)
		}
		fmt.Println("✅ Environment validation passed!")
		fmt.Println()
	}

	// Install packages
	if *install != "" {
		packages := parsePackageList(*install)
		fmt.Printf("📦 Installing packages: %v\n", packages)
		if err := manager.InstallDependencies(packages); err != nil {
			log.Fatalf("Failed to install dependencies: %v", err)
		}
		fmt.Println("✅ Packages installed successfully!")
		fmt.Println()
	}

	// If no specific action requested, show basic info
	if !*info && !*validate && *install == "" {
		fmt.Println("🐍 Python Manager Tool")
		fmt.Println("======================")

		// Discover interpreter
		interpreterPath, err := manager.DiscoverInterpreter()
		if err != nil {
			log.Fatalf("Failed to discover Python interpreter: %v", err)
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

		fmt.Println()
		fmt.Println("Use --help for more options")
	}
}

// parsePackageList parses a comma-separated list of packages
func parsePackageList(packageStr string) []string {
	if packageStr == "" {
		return nil
	}

	packages := []string{}
	for _, pkg := range splitAndTrim(packageStr, ",") {
		if pkg != "" {
			packages = append(packages, pkg)
		}
	}
	return packages
}

// splitAndTrim splits a string and trims whitespace from each part
func splitAndTrim(s, sep string) []string {
	parts := []string{}
	for _, part := range splitString(s, sep) {
		trimmed := trimSpace(part)
		if trimmed != "" {
			parts = append(parts, trimmed)
		}
	}
	return parts
}

// splitString splits a string by separator
func splitString(s, sep string) []string {
	if s == "" {
		return []string{}
	}

	parts := []string{}
	start := 0
	for i := 0; i < len(s); i++ {
		if i+len(sep) <= len(s) && s[i:i+len(sep)] == sep {
			parts = append(parts, s[start:i])
			start = i + len(sep)
			i += len(sep) - 1
		}
	}
	parts = append(parts, s[start:])
	return parts
}

// trimSpace trims whitespace from a string
func trimSpace(s string) string {
	start := 0
	end := len(s)

	// Trim leading whitespace
	for start < end && isSpace(s[start]) {
		start++
	}

	// Trim trailing whitespace
	for end > start && isSpace(s[end-1]) {
		end--
	}

	return s[start:end]
}

// isSpace checks if a character is whitespace
func isSpace(c byte) bool {
	return c == ' ' || c == '\t' || c == '\n' || c == '\r'
}
