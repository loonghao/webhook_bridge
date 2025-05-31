package utils

import (
	"fmt"
	"net"
	"testing"
)

func TestGetFreePort(t *testing.T) {
	port, err := GetFreePort()
	if err != nil {
		t.Fatalf("GetFreePort() error = %v", err)
	}

	if port <= 0 || port > 65535 {
		t.Errorf("GetFreePort() = %v, want port in range 1-65535", port)
	}

	// Verify the port is actually free
	if !IsPortFree(port) {
		t.Errorf("GetFreePort() returned port %v which is not free", port)
	}
}

func TestIsPortFree(t *testing.T) {
	// Test with a port that is in use
	listener, err := net.Listen("tcp", "localhost:0")
	if err != nil {
		t.Fatalf("Failed to create listener for testing: %v", err)
	}
	defer listener.Close()

	usedPort := listener.Addr().(*net.TCPAddr).Port

	// The port should not be free since we have a listener on it
	if IsPortFree(usedPort) {
		t.Errorf("IsPortFree(%v) = true, want false (port should be in use)", usedPort)
	}

	// Test with a port that should be free
	// Use a high port number that's unlikely to be in use
	testPort := 60000
	for i := 0; i < 10; i++ { // Try a few ports to find a free one
		if IsPortFree(testPort + i) {
			// Found a free port, test passed
			return
		}
	}
	t.Error("Could not find any free port in range 60000-60009 for testing")
}

func TestParsePort(t *testing.T) {
	tests := []struct {
		name    string
		portStr string
		want    int
		wantErr bool
	}{
		{
			name:    "valid port",
			portStr: "8080",
			want:    8080,
			wantErr: false,
		},
		{
			name:    "valid port at lower bound",
			portStr: "1",
			want:    1,
			wantErr: false,
		},
		{
			name:    "valid port at upper bound",
			portStr: "65535",
			want:    65535,
			wantErr: false,
		},
		{
			name:    "empty string",
			portStr: "",
			want:    0,
			wantErr: true,
		},
		{
			name:    "invalid number",
			portStr: "abc",
			want:    0,
			wantErr: true,
		},
		{
			name:    "port too low",
			portStr: "0",
			want:    0,
			wantErr: true,
		},
		{
			name:    "port too high",
			portStr: "65536",
			want:    0,
			wantErr: true,
		},
		{
			name:    "negative port",
			portStr: "-1",
			want:    0,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParsePort(tt.portStr)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParsePort() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("ParsePort() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetPortWithFallback(t *testing.T) {
	// Test with a port that is in use - should return a different port
	listener, err := net.Listen("tcp", "localhost:0")
	if err != nil {
		t.Fatalf("Failed to create listener for testing: %v", err)
	}
	defer listener.Close()

	usedPort := listener.Addr().(*net.TCPAddr).Port
	result, err := GetPortWithFallback(usedPort)
	if err != nil {
		t.Fatalf("GetPortWithFallback() error = %v", err)
	}

	if result == usedPort {
		t.Errorf("GetPortWithFallback(%v) = %v, should have returned different port", usedPort, result)
	}

	if result <= 0 || result > 65535 {
		t.Errorf("GetPortWithFallback() = %v, want port in range 1-65535", result)
	}

	// Verify the returned port is actually free
	if !IsPortFree(result) {
		t.Errorf("GetPortWithFallback() returned port %v which is not free", result)
	}

	// Test with a high port number that should be free
	testPort := 60010
	for i := 0; i < 10; i++ {
		candidatePort := testPort + i
		if IsPortFree(candidatePort) {
			result, err := GetPortWithFallback(candidatePort)
			if err != nil {
				t.Fatalf("GetPortWithFallback() error = %v", err)
			}
			if result != candidatePort {
				t.Errorf("GetPortWithFallback(%v) = %v, want %v (port should be free)", candidatePort, result, candidatePort)
			}
			return
		}
	}
	t.Error("Could not find any free port in range 60010-60019 for testing")
}

func TestGetFreePortInRange(t *testing.T) {
	// Test with a reasonable range
	start := 60000
	end := 60010

	port, err := GetFreePortInRange(start, end)
	if err != nil {
		t.Fatalf("GetFreePortInRange() error = %v", err)
	}

	if port < start || port > end {
		t.Errorf("GetFreePortInRange(%v, %v) = %v, want port in range", start, end, port)
	}

	if !IsPortFree(port) {
		t.Errorf("GetFreePortInRange() returned port %v which is not free", port)
	}
}

func TestGetFreePortInRange_NoFreePort(t *testing.T) {
	// Create a very small range and occupy all ports
	start := 60020
	end := 60022

	var listeners []net.Listener
	defer func() {
		for _, l := range listeners {
			l.Close()
		}
	}()

	// Occupy all ports in the range
	for port := start; port <= end; port++ {
		addr := fmt.Sprintf("localhost:%d", port)
		listener, err := net.Listen("tcp", addr)
		if err == nil {
			listeners = append(listeners, listener)
		} else {
			t.Logf("Failed to bind to port %d: %v", port, err)
		}
	}

	// Verify we actually bound to some ports
	if len(listeners) == 0 {
		t.Skip("Could not bind to any ports in the test range, skipping test")
	}

	// Now try to get a free port in the range
	_, err := GetFreePortInRange(start, end)
	if err == nil {
		t.Error("GetFreePortInRange() should have returned error when no ports are free")
	}
}

func TestGetPortsWithFallback(t *testing.T) {
	preferredPorts := []int{60030, 60031, 60032}

	ports, err := GetPortsWithFallback(preferredPorts)
	if err != nil {
		t.Fatalf("GetPortsWithFallback() error = %v", err)
	}

	if len(ports) != len(preferredPorts) {
		t.Errorf("GetPortsWithFallback() returned %v ports, want %v", len(ports), len(preferredPorts))
	}

	// Check that all returned ports are valid and free
	for i, port := range ports {
		if port <= 0 || port > 65535 {
			t.Errorf("GetPortsWithFallback() returned invalid port %v at index %v", port, i)
		}

		if !IsPortFree(port) {
			t.Errorf("GetPortsWithFallback() returned port %v which is not free", port)
		}
	}

	// Check that all ports are unique
	portMap := make(map[int]bool)
	for _, port := range ports {
		if portMap[port] {
			t.Errorf("GetPortsWithFallback() returned duplicate port %v", port)
		}
		portMap[port] = true
	}
}
