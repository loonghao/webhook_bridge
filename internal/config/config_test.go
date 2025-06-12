package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoad(t *testing.T) {
	tests := []struct {
		name    string
		setup   func() (cleanup func())
		wantErr bool
	}{
		{
			name: "load default config",
			setup: func() func() {
				return func() {}
			},
			wantErr: false,
		},
		{
			name: "load from environment variables",
			setup: func() func() {
				os.Setenv("WEBHOOK_BRIDGE_PORT", "9090")
				os.Setenv("WEBHOOK_BRIDGE_EXECUTOR_PORT", "50052")
				return func() {
					os.Unsetenv("WEBHOOK_BRIDGE_PORT")
					os.Unsetenv("WEBHOOK_BRIDGE_EXECUTOR_PORT")
				}
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cleanup := tt.setup()
			defer cleanup()

			cfg, err := Load()
			if (err != nil) != tt.wantErr {
				t.Errorf("Load() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if cfg == nil {
					t.Error("Load() returned nil config")
				}
			}
		})
	}
}

func TestConfig_AssignPorts(t *testing.T) {
	cfg := &Config{}
	cfg.setDefaults()

	err := cfg.AssignPorts()
	if err != nil {
		t.Errorf("AssignPorts() error = %v", err)
	}

	if cfg.Server.Port == 0 {
		t.Error("Server port was not assigned")
	}

	if cfg.Executor.Port == 0 {
		t.Error("Executor port was not assigned")
	}
}

func TestConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		config  *Config
		wantErr bool
	}{
		{
			name: "valid config",
			config: &Config{
				Server: ServerConfig{
					Port: 8080,
				},
				Executor: ExecutorConfig{
					Port:    50051,
					Timeout: 30,
				},
			},
			wantErr: false,
		},
		{
			name: "invalid server port",
			config: &Config{
				Server: ServerConfig{
					Port: -1,
				},
				Executor: ExecutorConfig{
					Port:    50051,
					Timeout: 30,
				},
			},
			wantErr: true,
		},
		{
			name: "invalid executor timeout",
			config: &Config{
				Server: ServerConfig{
					Port: 8080,
				},
				Executor: ExecutorConfig{
					Port:    50051,
					Timeout: -1,
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestConfig_GetServerAddress(t *testing.T) {
	cfg := &Config{
		Server: ServerConfig{
			Host: "localhost",
			Port: 8080,
		},
	}

	expected := "localhost:8080"
	actual := cfg.GetServerAddress()

	if actual != expected {
		t.Errorf("GetServerAddress() = %v, want %v", actual, expected)
	}
}

func TestConfig_GetExecutorAddress(t *testing.T) {
	cfg := &Config{
		Executor: ExecutorConfig{
			Host: "localhost",
			Port: 50051,
		},
	}

	expected := "localhost:50051"
	actual := cfg.GetExecutorAddress()

	if actual != expected {
		t.Errorf("GetExecutorAddress() = %v, want %v", actual, expected)
	}
}

func TestLoadFromFile(t *testing.T) {
	// Create a temporary config file
	tempDir := t.TempDir()
	configFile := filepath.Join(tempDir, "test_config.yaml")

	configContent := `
server:
  host: "0.0.0.0"
  port: 9090
  mode: "release"

executor:
  host: "localhost"
  port: 50052
  timeout: 60

logging:
  level: "debug"
  format: "json"
`

	err := os.WriteFile(configFile, []byte(configContent), 0600)
	if err != nil {
		t.Fatalf("Failed to create test config file: %v", err)
	}

	cfg, err := LoadFromFile(configFile)
	if err != nil {
		t.Fatalf("LoadFromFile() error = %v", err)
	}

	if cfg.Server.Host != "0.0.0.0" {
		t.Errorf("Server.Host = %v, want %v", cfg.Server.Host, "0.0.0.0")
	}

	if cfg.Server.Port != 9090 {
		t.Errorf("Server.Port = %v, want %v", cfg.Server.Port, 9090)
	}

	if cfg.Server.Mode != "release" {
		t.Errorf("Server.Mode = %v, want %v", cfg.Server.Mode, "release")
	}

	if cfg.Executor.Port != 50052 {
		t.Errorf("Executor.Port = %v, want %v", cfg.Executor.Port, 50052)
	}

	if cfg.Executor.Timeout != 60 {
		t.Errorf("Executor.Timeout = %v, want %v", cfg.Executor.Timeout, 60)
	}

	if cfg.Logging.Level != "debug" {
		t.Errorf("Logging.Level = %v, want %v", cfg.Logging.Level, "debug")
	}
}

func TestConfig_setDefaults(t *testing.T) {
	cfg := &Config{}
	cfg.setDefaults()

	// Test server defaults
	if cfg.Server.Host != "0.0.0.0" {
		t.Errorf("Server.Host = %v, want %v", cfg.Server.Host, "0.0.0.0")
	}

	if cfg.Server.Port != 0 {
		t.Errorf("Server.Port = %v, want %v", cfg.Server.Port, 0)
	}

	if cfg.Server.Mode != "debug" {
		t.Errorf("Server.Mode = %v, want %v", cfg.Server.Mode, "debug")
	}

	// Test executor defaults
	if cfg.Executor.Host != "localhost" {
		t.Errorf("Executor.Host = %v, want %v", cfg.Executor.Host, "localhost")
	}

	if cfg.Executor.Port != 0 {
		t.Errorf("Executor.Port = %v, want %v", cfg.Executor.Port, 0)
	}

	if cfg.Executor.Timeout != 30 {
		t.Errorf("Executor.Timeout = %v, want %v", cfg.Executor.Timeout, 30)
	}

	// Test logging defaults
	if cfg.Logging.Level != "info" {
		t.Errorf("Logging.Level = %v, want %v", cfg.Logging.Level, "info")
	}

	if cfg.Logging.Format != "text" {
		t.Errorf("Logging.Format = %v, want %v", cfg.Logging.Format, "text")
	}
}
