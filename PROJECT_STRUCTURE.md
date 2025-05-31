# Project Structure

This document describes the project structure and organization of the webhook bridge hybrid architecture.

## Directory Layout

```
webhook_bridge/
├── api/                    # API definitions and generated code
│   └── proto/             # Protocol Buffer definitions
│       ├── webhook.proto  # gRPC service definition
│       ├── webhook.pb.go  # Generated Go code
│       ├── webhook_grpc.pb.go
│       ├── webhook_pb2.py # Generated Python code
│       └── webhook_pb2_grpc.py
├── cmd/                   # Application entry points
│   └── server/           # Go HTTP server main
│       └── main.go
├── internal/             # Private application code
│   ├── config/          # Configuration management
│   ├── grpc/            # gRPC client implementation
│   ├── python/          # Python interpreter management
│   └── server/          # HTTP server implementation
├── pkg/                  # Public library code
│   └── version/         # Version information
├── python_executor/      # Python gRPC service
│   ├── __init__.py
│   ├── main.py          # Service entry point
│   └── server.py        # gRPC service implementation
├── webhook_bridge/       # Original Python package (preserved)
│   ├── api/
│   ├── cli.py
│   ├── filesystem.py
│   ├── models.py
│   ├── plugin.py
│   └── server.py
├── example_plugins/      # Example webhook plugins
├── tests/               # Test files
├── scripts/             # Build and deployment scripts
│   ├── build.sh
│   └── setup-dev.sh
├── config.yaml          # Configuration file
├── Dockerfile           # Container definition
├── docker-compose.yml   # Container orchestration
├── Makefile            # Build automation
├── go.mod              # Go module definition
├── pyproject.toml      # Python project definition
└── README.md           # Project documentation
```

## Architecture Components

### Go Components

#### `cmd/server/`
- **Purpose**: Main entry point for the Go HTTP server
- **Responsibilities**: 
  - Application initialization
  - Configuration loading
  - Server startup and shutdown

#### `internal/config/`
- **Purpose**: Configuration management
- **Features**:
  - YAML configuration loading
  - Environment variable overrides
  - Multiple Python interpreter strategies

#### `internal/server/`
- **Purpose**: HTTP server implementation using Gin
- **Features**:
  - RESTful API endpoints
  - Request routing and validation
  - gRPC client integration

#### `internal/grpc/`
- **Purpose**: gRPC client for communicating with Python executor
- **Features**:
  - Connection management
  - Request/response handling
  - Error handling and timeouts

#### `internal/python/`
- **Purpose**: Python interpreter discovery and management
- **Features**:
  - UV virtual environment support
  - Custom path configuration
  - PATH-based discovery

#### `pkg/version/`
- **Purpose**: Version information and build metadata
- **Features**:
  - Version strings
  - Build information
  - Runtime details

### Python Components

#### `python_executor/`
- **Purpose**: gRPC service for executing Python plugins
- **Features**:
  - Plugin discovery and loading
  - Backward compatibility with existing plugins
  - Health checks and monitoring

#### `webhook_bridge/` (Preserved)
- **Purpose**: Original Python package
- **Status**: Maintained for compatibility
- **Usage**: Used by Python executor for plugin execution

### API Definition

#### `api/proto/`
- **Purpose**: gRPC service definitions
- **Files**:
  - `webhook.proto`: Service interface definition
  - Generated Go and Python code

### Configuration

#### `config.yaml`
- **Purpose**: Main configuration file
- **Sections**:
  - Server settings
  - Python interpreter configuration
  - Executor service settings
  - Logging configuration

### Build and Deployment

#### `scripts/`
- **Purpose**: Build and deployment automation
- **Files**:
  - `build.sh`: Go application build script
  - `setup-dev.sh`: Development environment setup

#### `Dockerfile`
- **Purpose**: Container image definition
- **Features**:
  - Multi-stage build
  - Go and Python runtime
  - Security best practices

#### `docker-compose.yml`
- **Purpose**: Container orchestration
- **Profiles**:
  - Production deployment
  - Development environment
  - Separate services for debugging

## Best Practices Implemented

### Go Project Structure
- ✅ Standard Go project layout
- ✅ Clear separation of concerns
- ✅ Internal packages for private code
- ✅ Public packages for reusable code

### Python Project Structure
- ✅ Preserved existing structure
- ✅ Clear module organization
- ✅ Proper package initialization

### Configuration Management
- ✅ Centralized configuration
- ✅ Environment variable support
- ✅ Validation and defaults

### Build and Deployment
- ✅ Automated build scripts
- ✅ Version management
- ✅ Container support
- ✅ Development environment setup

### Code Quality
- ✅ Consistent naming conventions
- ✅ Proper error handling
- ✅ Logging and monitoring
- ✅ Documentation

## Development Workflow

1. **Setup**: Run `make dev-setup` or `scripts/setup-dev.sh`
2. **Build**: Run `make build`
3. **Test**: Run `make test`
4. **Run**: Run `make run-all` or separate services
5. **Deploy**: Use Docker or build binaries

## Deployment Options

### 1. Single Container
```bash
docker-compose up webhook-bridge
```

### 2. Development Mode
```bash
docker-compose --profile dev up
```

### 3. Separate Services
```bash
docker-compose --profile dev-separate up
```

### 4. Binary Deployment
```bash
make build
./bin/webhook-bridge-server &
python python_executor/main.py
```

## Migration Path

This structure supports gradual migration:

1. **Phase 1**: Hybrid architecture (current)
2. **Phase 2**: Performance optimization
3. **Phase 3**: Advanced features

The structure is designed to be flexible and maintainable while preserving backward compatibility.
