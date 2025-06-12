package cli

import (
	"fmt"
	"strconv"
)

// parsePort parses a port string to integer
func parsePort(portStr string) (int, error) {
	if portStr == "" {
		return 0, fmt.Errorf("empty port string")
	}

	// Handle special cases
	switch portStr {
	case "auto", "0":
		return 0, nil
	}

	// Parse as integer
	port, err := strconv.Atoi(portStr)
	if err != nil {
		return 0, fmt.Errorf("invalid port format: %s", portStr)
	}

	// Validate port range
	if port < 1 || port > 65535 {
		return 0, fmt.Errorf("port out of range: %d (must be 1-65535)", port)
	}

	return port, nil
}
