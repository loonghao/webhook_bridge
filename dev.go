//go:build ignore

// Development tool entry point
// Usage: go run dev.go <command>
package main

import (
	"os"
	"os/exec"
)

func main() {
	args := append([]string{"run", "tools/dev/main.go"}, os.Args[1:]...)
	cmd := exec.Command("go", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	if err := cmd.Run(); err != nil {
		os.Exit(1)
	}
}
