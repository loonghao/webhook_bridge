package cli

import (
	"fmt"
	"runtime"

	"github.com/spf13/cobra"
)

var (
	version   = "dev"
	buildTime = "unknown"
	goVersion = "unknown"
)

// SetVersionInfo sets the version information
func SetVersionInfo(v, bt, gv string) {
	version = v
	buildTime = bt
	goVersion = gv
}

// NewVersionCommand creates the version command
func NewVersionCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Show version information",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("Webhook Bridge %s\n", version)
			fmt.Printf("Build Time: %s\n", buildTime)
			fmt.Printf("Go Version: %s\n", goVersion)
			fmt.Printf("Runtime: %s/%s\n", runtime.GOOS, runtime.GOARCH)
		},
	}
}
