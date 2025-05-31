# Webhook Bridge Python CLI

This Python package provides a command-line interface for managing the Webhook Bridge Go binary. It automatically downloads, installs, and manages the appropriate binary for your platform.

## Installation

### Using uvx (Recommended)
```bash
# Run directly without installation
uvx webhook-bridge --help

# Install and run
uvx webhook-bridge install
uvx webhook-bridge run
```

### Using pip
```bash
pip install webhook-bridge
webhook-bridge --help
```

## Quick Start

1. **Install the binary**:
   ```bash
   webhook-bridge install
   ```

2. **Start the server**:
   ```bash
   webhook-bridge run
   ```

3. **Check status**:
   ```bash
   webhook-bridge status
   ```

## Commands

- `install` - Download and install the webhook bridge server binary
- `run` - Start the webhook bridge server
- `status` - Check the status of the installation
- `stop` - Stop the running server
- `update` - Update to the latest version
- `config` - Configuration management

## Examples

```bash
# Install specific version
webhook-bridge install --version v1.0.0

# Run on custom port
webhook-bridge run --port 9000

# Run in daemon mode
webhook-bridge run --daemon

# Check for updates
webhook-bridge update --check-only

# Initialize configuration
webhook-bridge config init
```

## Architecture

This Python CLI tool is a lightweight wrapper that:

1. **Downloads** the appropriate Go binary for your platform
2. **Manages** the binary lifecycle (install, update, remove)
3. **Provides** a consistent interface across platforms
4. **Handles** configuration and process management

The actual webhook bridge server is implemented in Go for high performance, while this Python tool provides easy installation and management.

## Platform Support

- **Linux**: AMD64, ARM64
- **Windows**: AMD64
- **macOS**: AMD64, ARM64 (Apple Silicon)

## Configuration

The CLI tool supports YAML configuration files:

```yaml
server:
  host: "0.0.0.0"
  port: 8000
  mode: "debug"

python:
  interpreter: "python"
  venv_path: ".venv"

logging:
  level: "info"
  format: "text"

directories:
  working_dir: "."
  log_dir: "logs"
  plugin_dir: "plugins"
```

## Development

This package is part of the Webhook Bridge project. For the full source code and documentation, visit:

- **Repository**: https://github.com/loonghao/webhook_bridge
- **Issues**: https://github.com/loonghao/webhook_bridge/issues
- **Documentation**: https://github.com/loonghao/webhook_bridge#readme
