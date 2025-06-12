# Webhook Bridge - Unified Architecture

## Overview

Webhook Bridge 2.0 introduces a unified architecture that combines multiple executables into a single, powerful binary. This design eliminates the complexity of managing multiple processes while maintaining the high performance of Go and the flexibility of Python plugins.

## Architecture Diagram

```
┌─────────────────────────────────────────────────────────┐
│                webhook-bridge.exe                       │
│                 (Unified Binary)                        │
│                                                         │
│  ┌─────────────────┐    Internal    ┌─────────────────┐ │
│  │   Go HTTP       │◄──────────────►│ Python Executor │ │
│  │   Server        │   Management   │ Manager         │ │
│  │                 │                │                 │ │
│  │ • Gin Router    │                │ • Auto-start    │ │
│  │ • Request       │                │ • Process Mgmt  │ │
│  │   Validation    │                │ • gRPC Client   │ │
│  │ • Load Balancing│                │ • Health Check  │ │
│  │ • Monitoring    │                │                 │ │
│  └─────────────────┘                └─────────────────┘ │
│           │                                  │          │
│           │                                  ▼          │
│           │                    ┌─────────────────┐      │
│           │                    │ Python Executor │      │
│           │                    │ Service         │      │
│           │                    │                 │      │
│           │                    │ • Plugin Loader │      │
│           │                    │ • gRPC Server   │      │
│           │                    │ • Compatibility │      │
│           │                    │   Layer         │      │
│           │                    └─────────────────┘      │
└───────────┼─────────────────────────────────────────────┘
            │                            │
            ▼                            ▼
┌─────────────────┐             ┌─────────────────┐
│   HTTP Clients  │             │ Python Plugins  │
│                 │             │                 │
│ • Webhooks      │             │ • Existing Code │
│ • API Calls     │             │ • New Plugins   │
│ • Browsers      │             │ • Libraries     │
└─────────────────┘             └─────────────────┘
```

## Components

### 1. Unified CLI (`cmd/webhook-bridge/`)
- **Framework**: Cobra CLI with subcommands
- **Responsibilities**:
  - Command routing and argument parsing
  - Service orchestration and management
  - Configuration loading and validation
  - Development and production workflows

### 2. Go HTTP Server (Integrated)
- **Framework**: Gin (high-performance HTTP router)
- **Responsibilities**:
  - HTTP request handling and routing
  - Request validation and preprocessing
  - Response formatting and error handling
  - Load balancing and concurrency management
  - Health checks and monitoring

### 3. Python Executor Manager (Integrated)
- **Framework**: Process management with gRPC client
- **Responsibilities**:
  - Automatic Python executor startup
  - Process lifecycle management
  - gRPC connection management
  - Health monitoring and recovery

### 4. Python Plugin Executor (`python_executor/`)
- **Framework**: gRPC server with existing plugin system
- **Responsibilities**:
  - Plugin discovery and loading
  - Plugin execution with existing logic
  - Backward compatibility with current plugins
  - Error handling and logging

### 5. Python Interpreter Management (`internal/python/`)
- **Strategies**:
  1. **UV Virtual Environment** (preferred)
  2. **Custom Path** (configured)
  3. **PATH Discovery** (fallback)
- **Features**:
  - Automatic environment setup
  - Dependency isolation
  - Version compatibility checks

## Configuration

### config.yaml
```yaml
server:
  host: "0.0.0.0"
  port: 8000
  mode: "debug"

python:
  strategy: "auto"  # auto, uv, path, custom
  uv:
    enabled: true
    project_path: "."
    venv_name: "webhook-bridge"
  plugin_dirs:
    - "./plugins"
    - "./example_plugins"

executor:
  host: "localhost"
  port: 50051
  timeout: 30

logging:
  level: "info"
  format: "text"
```

## Python Interpreter Discovery

The system uses a priority-based approach to find the Python interpreter:

1. **Custom Path** - If explicitly configured
2. **UV Virtual Environment** - If UV is available and enabled
3. **PATH Discovery** - System PATH lookup
4. **Error** - If none found

### UV Integration
```bash
# Create virtual environment
uv venv webhook-bridge

# Install dependencies
uv sync

# Run with UV
uv run python python_executor/main.py
```

## Deployment Options

### 1. Unified Service (Recommended)
```bash
# Build unified binary
go build -o webhook-bridge.exe ./cmd/webhook-bridge

# Run unified service (Python executor + Go server)
./webhook-bridge unified --port 8080
```

### 2. Individual Components
```bash
# Run Go server only
./webhook-bridge serve --port 8080

# Run backend server with gRPC client
./webhook-bridge server --port 8080

# Manage Python environment
./webhook-bridge python info
```

### 3. Docker Container
```dockerfile
FROM golang:1.21-alpine AS go-builder
# ... Go build steps

FROM python:3.11-slim
# ... Python setup
COPY --from=go-builder /app/webhook-bridge /usr/local/bin/
CMD ["webhook-bridge", "unified"]
```

## Performance Benefits

### Go HTTP Server
- **Concurrency**: Native goroutines for handling thousands of concurrent requests
- **Memory**: Lower memory footprint compared to Python
- **Startup**: Faster startup times
- **Throughput**: Higher requests per second

### Python Plugin Compatibility
- **Existing Code**: No changes required for current plugins
- **Libraries**: Full access to Python ecosystem
- **Development**: Rapid prototyping and development

## Migration Path

### Phase 1: Hybrid Setup ✅
- Go HTTP server with gRPC communication
- Python executor service
- Full backward compatibility

### Phase 2: Performance Optimization
- Identify high-traffic plugins
- Optional Go plugin implementations
- A/B testing between implementations

### Phase 3: Advanced Features
- Plugin hot-reloading
- Multi-language plugin support
- Advanced monitoring and metrics

## Development Workflow

### Setup
```bash
# Install dependencies
make install-deps

# Generate gRPC code
make proto

# Setup development environment
make dev-setup
```

### Running
```bash
# Run both services
make run-all

# Or run separately
make run-python  # Terminal 1
make run         # Terminal 2
```

### Testing
```bash
# Run all tests
make test

# Run specific tests
make test-go
make test-python
```

## API Compatibility

The new architecture maintains 100% API compatibility with the existing webhook bridge:

- Same HTTP endpoints (`/api/v1/webhook/:plugin`)
- Same request/response formats
- Same plugin interface
- Same configuration options

## Monitoring and Observability

### Health Checks
- Go server: `GET /health`
- Python executor: gRPC health check
- End-to-end: Plugin execution test

### Metrics
- Request latency and throughput
- Plugin execution times
- Error rates and types
- Resource utilization

### Logging
- Structured logging (JSON/text)
- Correlation IDs across services
- Plugin execution traces

## Security Considerations

### Network Security
- gRPC communication over localhost by default
- TLS support for production deployments
- Network isolation options

### Plugin Security
- Sandboxed Python execution
- Resource limits and timeouts
- Input validation and sanitization

## Troubleshooting

### Common Issues

1. **Python Interpreter Not Found**
   ```bash
   # Check configuration
   cat config.yaml
   
   # Test Python discovery
   ./bin/webhook-bridge-server --test-python
   ```

2. **gRPC Connection Failed**
   ```bash
   # Check if Python executor is running
   ps aux | grep python_executor
   
   # Test gRPC connection
   grpcurl -plaintext localhost:50051 list
   ```

3. **Plugin Execution Errors**
   ```bash
   # Check plugin logs
   tail -f logs/python_executor.log
   
   # Test plugin directly
   python -c "from example_plugins.test_plugin import *"
   ```

## Future Enhancements

- **Multi-language Support**: Add support for Node.js, Rust plugins
- **Plugin Marketplace**: Plugin discovery and installation
- **Visual Plugin Builder**: GUI for creating simple plugins
- **Advanced Routing**: Content-based routing and load balancing
- **Caching Layer**: Redis integration for plugin results
- **Metrics Dashboard**: Real-time monitoring interface
