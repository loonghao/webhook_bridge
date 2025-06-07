# 🚀 Local Development Guide

This guide covers local development workflows for webhook-bridge using the new nox-based development commands.

## Quick Start

### ⚡ Super Quick Start
```bash
# Build and run server immediately (fastest)
uvx nox -s quick
```
This command:
- Builds only the Go binary (skips frontend for speed)
- Creates minimal configuration
- Starts server on http://127.0.0.1:8001

### 🔧 Full Development Start
```bash
# Build everything and run with full features
uvx nox -s dev
```
This command:
- Builds frontend dashboard
- Builds all Go binaries
- Creates comprehensive test configuration
- Opens dashboard in browser
- Starts server with debug logging

## Available Commands

### Development Commands

| Command | Description | Use Case |
|---------|-------------|----------|
| `uvx nox -s quick` | Super fast start (Go only) | Quick testing, API development |
| `uvx nox -s dev` | Full development environment | Frontend development, full testing |
| `uvx nox -s build-local` | Build all components | CI/CD, manual builds |
| `uvx nox -s test-local` | Test built binaries | Verification, debugging |
| `uvx nox -s run-local` | Run pre-built server | Testing existing builds |
| `uvx nox -s clean-local` | Clean build artifacts | Reset environment |

### Workflow Examples

#### Frontend Development
```bash
# Full build with dashboard
uvx nox -s dev
# Dashboard available at http://127.0.0.1:8001
```

#### API Development
```bash
# Quick start without frontend build
uvx nox -s quick
# API available at http://127.0.0.1:8000
```

#### Testing Builds
```bash
# Build everything
uvx nox -s build-local

# Test the builds
uvx nox -s test-local

# Run for manual testing
uvx nox -s run-local
```

#### Clean Development
```bash
# Clean up everything
uvx nox -s clean-local

# Fresh start
uvx nox -s dev
```

## Configuration Files

The development commands create test configuration files:

- `config.test.yaml` - Full development configuration (created by `dev`)
- `config.quick.yaml` - Minimal configuration (created by `quick`)

### Test Configuration Features
- Debug logging enabled
- Local host binding (127.0.0.1)
- Standard ports (8000 for API, 8001 for dashboard)
- Example plugins directory
- Auto-open dashboard in browser

## Server Endpoints

When running locally, you can access:

- **API Server**: http://127.0.0.1:8000
  - Health check: http://127.0.0.1:8000/health
  - API docs: http://127.0.0.1:8000/api
  - Plugin API: http://127.0.0.1:8000/api/v1/plugins

- **Dashboard**: http://127.0.0.1:8001
  - Main dashboard: http://127.0.0.1:8001/dashboard
  - Plugin management: http://127.0.0.1:8001/plugins
  - System logs: http://127.0.0.1:8001/logs
  - Configuration: http://127.0.0.1:8001/config

## Troubleshooting

### Binary Not Found
```bash
# Error: webhook-bridge.exe not found
# Solution: Build first
uvx nox -s build-local
```

### Port Already in Use
```bash
# Error: Port 8000/8001 already in use
# Solution: Stop existing processes or modify config
```

### Frontend Build Issues
```bash
# Error: Dashboard build fails
# Solution: Use quick start to skip frontend
uvx nox -s quick
```

### Python Executor Issues
```bash
# Warning: Python executor not available
# This is normal for local development
# Server runs in API-only mode
```

## Development Tips

1. **Use `quick` for rapid iteration** when working on Go code
2. **Use `dev` for full-stack development** when working on dashboard
3. **Use `clean-local`** if you encounter build issues
4. **Check logs** in `logs/webhook-bridge.log` for debugging
5. **Dashboard auto-opens** in your default browser with `dev` command

## Integration with Python CLI

The webhook_bridge Python CLI can also start the Go server:

```bash
# Using Python CLI (requires binary in PATH or current directory)
python -m webhook_bridge.cli run --port 8000

# Check if binary is detected
python -m webhook_bridge.cli status
```

## Next Steps

- See [PLUGIN_DEVELOPMENT.md](PLUGIN_DEVELOPMENT.md) for plugin development
- See [DASHBOARD_GUIDE.md](DASHBOARD_GUIDE.md) for dashboard features
- See [CLI_USAGE.md](CLI_USAGE.md) for production deployment
