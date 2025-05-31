package utils

import (
	"fmt"
	"net"
	"strconv"
)

// GetFreePort finds and returns a free port on the local machine
func GetFreePort() (int, error) {
	addr, err := net.ResolveTCPAddr("tcp", "localhost:0")
	if err != nil {
		return 0, err
	}

	l, err := net.ListenTCP("tcp", addr)
	if err != nil {
		return 0, err
	}
	defer l.Close()

	return l.Addr().(*net.TCPAddr).Port, nil
}

// GetFreePortInRange finds a free port within the specified range
func GetFreePortInRange(start, end int) (int, error) {
	for port := start; port <= end; port++ {
		if IsPortFree(port) {
			return port, nil
		}
	}
	return 0, fmt.Errorf("no free port found in range %d-%d", start, end)
}

// IsPortFree checks if a port is available for use
func IsPortFree(port int) bool {
	addr := fmt.Sprintf("localhost:%d", port)
	conn, err := net.Listen("tcp", addr)
	if err != nil {
		return false
	}
	conn.Close()
	return true
}

// ParsePort parses a port string and validates it
func ParsePort(portStr string) (int, error) {
	if portStr == "" {
		return 0, fmt.Errorf("port cannot be empty")
	}
	
	port, err := strconv.Atoi(portStr)
	if err != nil {
		return 0, fmt.Errorf("invalid port number: %s", portStr)
	}
	
	if port < 1 || port > 65535 {
		return 0, fmt.Errorf("port must be between 1 and 65535, got %d", port)
	}
	
	return port, nil
}

// GetPortWithFallback tries to use the specified port, falls back to a free port if occupied
func GetPortWithFallback(preferredPort int) (int, error) {
	if preferredPort > 0 && IsPortFree(preferredPort) {
		return preferredPort, nil
	}
	
	// If preferred port is not available, find a free one
	return GetFreePort()
}

// GetPortsWithFallback gets multiple ports with fallback logic
func GetPortsWithFallback(preferredPorts []int) ([]int, error) {
	var ports []int
	
	for _, preferred := range preferredPorts {
		port, err := GetPortWithFallback(preferred)
		if err != nil {
			return nil, fmt.Errorf("failed to get port (preferred: %d): %w", preferred, err)
		}
		ports = append(ports, port)
	}
	
	return ports, nil
}
