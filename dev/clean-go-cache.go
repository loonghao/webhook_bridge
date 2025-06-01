package main

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
)

func main() {
	fmt.Println("🧹 Cleaning Go build cache and module cache...")

	// Clean build cache
	if err := runCommand("go", "clean", "-cache"); err != nil {
		fmt.Printf("❌ Failed to clean build cache: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("✅ Build cache cleaned")

	// Clean module cache
	if err := runCommand("go", "clean", "-modcache"); err != nil {
		fmt.Printf("❌ Failed to clean module cache: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("✅ Module cache cleaned")

	// Clean test cache
	if err := runCommand("go", "clean", "-testcache"); err != nil {
		fmt.Printf("❌ Failed to clean test cache: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("✅ Test cache cleaned")

	// Remove any potential Go tool binaries that might be cached
	if err := runCommand("go", "clean", "-i", "all"); err != nil {
		fmt.Printf("⚠️  Warning: Failed to clean installed packages: %v\n", err)
	} else {
		fmt.Println("✅ Installed packages cleaned")
	}

	fmt.Println("🎉 Go cache cleanup completed successfully!")
	fmt.Println("💡 You may need to run 'go mod download' to re-download dependencies")
}

func runCommand(name string, args ...string) error {
	cmd := exec.Command(name, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	// Set environment for cross-platform compatibility
	if runtime.GOOS == "windows" {
		cmd.Env = append(os.Environ(), "GOOS=windows")
	}

	return cmd.Run()
}
