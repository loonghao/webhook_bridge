<p align="center">
<img src="https://i.imgur.com/d9UWkck.png" alt="logo"></a>
</p>

# Webhook Bridge

A high-performance webhook integration platform with **hybrid Go/Python architecture**. Features a blazing-fast Go HTTP server with flexible Python plugin execution environment.

[![Go Version](https://img.shields.io/github/go-mod/go-version/loonghao/webhook_bridge)](https://golang.org/)
[![Python Version](https://img.shields.io/pypi/pyversions/webhook-bridge)](https://pypi.org/project/webhook-bridge/)
[![PyPI Version](https://img.shields.io/pypi/v/webhook-bridge?color=green)](https://pypi.org/project/webhook-bridge/)
[![Go CI](https://github.com/loonghao/webhook_bridge/workflows/Go%20CI/CD/badge.svg)](https://github.com/loonghao/webhook_bridge/actions)
[![Python Tests](https://github.com/loonghao/webhook_bridge/workflows/Tests/badge.svg)](https://github.com/loonghao/webhook_bridge/actions)
[![License](https://img.shields.io/github/license/loonghao/webhook_bridge)](https://github.com/loonghao/webhook_bridge/blob/main/LICENSE)
[![Release](https://img.shields.io/github/v/release/loonghao/webhook_bridge)](https://github.com/loonghao/webhook_bridge/releases)
[![Downloads](https://static.pepy.tech/badge/webhook-bridge)](https://pepy.tech/project/webhook-bridge)

<p align="center">
<img src="https://i.imgur.com/31RO4xN.png" alt="logo"></a>
</p>

## 🚀 **v1.0.0 - Major Architecture Upgrade**

**Breaking Change**: Webhook Bridge has been completely rewritten with a hybrid Go/Python architecture for maximum performance and flexibility.

### **🏗️ New Architecture**
- **⚡ Go HTTP Server**: High-performance server built with Gin framework
- **🐍 Python Executor**: Flexible plugin execution environment via gRPC
- **🎨 Modern Dashboard**: Beautiful web interface with Tailwind CSS
- **📦 Easy Installation**: Simple CLI tool for binary management

## Features

- 🚀 **API Versioning**: Support for versioned API endpoints (`/v1`, `/latest`)
- 🔌 **Plugin System**: Dynamic plugin loading and execution
- 🌐 **RESTful API**: Support for GET, POST, PUT, DELETE HTTP methods
- 🛠️ **Flexible Configuration**: Extensive CLI and programmatic configuration options
- 📝 **Rich Documentation**: Interactive API documentation with Swagger UI and ReDoc
- 🔒 **Secure**: Built-in security features and error handling
- 📊 **Logging**: Comprehensive logging and error tracking

## 📦 Installation

### **🎯 Quick Start (Recommended)**

Using `uvx` (no installation required):
```bash
# Install and run in one command
uvx webhook-bridge install
uvx webhook-bridge run

# Or run directly
uvx webhook-bridge --help
```

### **🐍 Python Package Installation**

```bash
# Install via pip
pip install webhook-bridge

# Install and run the server
webhook-bridge install
webhook-bridge run
```

### **⚡ Direct Binary Download**

Download pre-built binaries from [GitHub Releases](https://github.com/loonghao/webhook_bridge/releases):

```bash
# Linux AMD64
wget https://github.com/loonghao/webhook_bridge/releases/latest/download/webhook-bridge-linux-amd64.tar.gz
tar -xzf webhook-bridge-linux-amd64.tar.gz
./webhook-bridge-linux-amd64

# Windows AMD64
# Download webhook-bridge-windows-amd64.zip and extract

# macOS (Intel)
wget https://github.com/loonghao/webhook_bridge/releases/latest/download/webhook-bridge-darwin-amd64.tar.gz
tar -xzf webhook-bridge-darwin-amd64.tar.gz
./webhook-bridge-darwin-amd64

# macOS (Apple Silicon)
wget https://github.com/loonghao/webhook_bridge/releases/latest/download/webhook-bridge-darwin-arm64.tar.gz
tar -xzf webhook-bridge-darwin-arm64.tar.gz
./webhook-bridge-darwin-arm64
```

### **🐳 Docker**

```bash
# Run with Docker
docker run -p 8000:8000 ghcr.io/loonghao/webhook-bridge:latest

# Or with docker-compose
docker-compose up
```

## 🚀 Quick Start

### **1. Install and Start Server**

```bash
# Using uvx (recommended)
uvx webhook-bridge install
uvx webhook-bridge run

# Or using pip
pip install webhook-bridge
webhook-bridge install
webhook-bridge run --port 8000
```

### **2. Access the Modern Dashboard**

Open your browser and navigate to:
- **Dashboard**: `http://localhost:8000/` - Modern web interface
- **API Documentation**: `http://localhost:8000/api` - API reference

### **3. Test with Sample Request**

```bash
# Test webhook endpoint
curl -X POST "http://localhost:8000/v1/webhook/example" \
     -H "Content-Type: application/json" \
     -d '{"message": "Hello, World!"}'

# Check server status
curl "http://localhost:8000/health"
```

### **4. CLI Commands**

```bash
# Check installation status
webhook-bridge status

# Update to latest version
webhook-bridge update

# Stop the server
webhook-bridge stop

# View configuration
webhook-bridge config show

# Initialize configuration file
webhook-bridge config init
```

## Configuration

### Command Line Options

```bash
webhook-bridge --help
```

#### Server Configuration
- `--host`: Host to bind the server to (default: "0.0.0.0")
- `--port`: Port to bind the server to (default: 8000)
- `--log-level`: Logging level (DEBUG/INFO/WARNING/ERROR/CRITICAL)

#### Worker Configuration
- `--workers`: Number of worker processes (default: 1)
- `--worker-class`: Worker class to use (default: uvicorn.workers.UvicornWorker)

#### Development and Debugging
- `--reload`: Enable auto-reload for development
- `--reload-dirs`: Directories to watch for reload (space-separated)

#### Logging Configuration
- `--access-log`: Enable access log (default: enabled)
- `--no-access-log`: Disable access log
- `--use-colors`: Use colors in log output (default: enabled)
- `--no-use-colors`: Disable colors in log output

#### SSL/TLS Configuration
- `--ssl-keyfile`: SSL key file path
- `--ssl-certfile`: SSL certificate file path
- `--ssl-ca-certs`: SSL CA certificates file path

#### Performance Configuration
- `--limit-concurrency`: Maximum number of concurrent connections
- `--limit-max-requests`: Maximum number of requests before restarting worker
- `--timeout-keep-alive`: Keep-alive timeout in seconds (default: 5)

#### API Configuration
- `--title`: API title
- `--description`: API description
- `--disable-docs`: Disable the API documentation endpoints (/docs and /redoc)

#### Plugin Configuration
- `--plugin-dir`: Directory containing webhook plugins

### Environment Variables

All command-line options can also be set via environment variables:

- `WEBHOOK_BRIDGE_HOST`: Host to bind the server to
- `WEBHOOK_BRIDGE_PORT`: Port to bind the server to
- `WEBHOOK_BRIDGE_LOG_LEVEL`: Logging level
- `WEBHOOK_BRIDGE_WORKERS`: Number of worker processes
- `WEBHOOK_BRIDGE_WORKER_CLASS`: Worker class to use
- `WEBHOOK_BRIDGE_RELOAD`: Enable auto-reload (true/false)
- `WEBHOOK_BRIDGE_SERVER_PLUGINS`: Additional plugin directories (separated by system path separator)

### Usage Examples

#### Basic Usage
```bash
# Start server with default settings
webhook-bridge

# Start server on specific host and port
webhook-bridge --host 127.0.0.1 --port 9000
```

#### Production Deployment
```bash
# Start with multiple workers for production
webhook-bridge --workers 4 --host 0.0.0.0 --port 8000

# Start with SSL/TLS support
webhook-bridge --ssl-keyfile /path/to/key.pem --ssl-certfile /path/to/cert.pem

# Start with performance limits
webhook-bridge --limit-concurrency 1000 --limit-max-requests 10000
```

#### Development Mode
```bash
# Start with auto-reload for development
webhook-bridge --reload --reload-dirs webhook_bridge --log-level DEBUG

# Start without access logs and colors for cleaner output
webhook-bridge --no-access-log --no-use-colors
```

### Modern CLI Features

The webhook bridge now uses **Click** for a modern CLI experience with:

- **Rich help system**: `webhook-bridge --help`
- **Environment variable support**: All options can be set via `WEBHOOK_BRIDGE_*` environment variables
- **Type validation**: Automatic validation of paths, integers, and choices
- **Boolean flags**: Use `--flag/--no-flag` syntax for boolean options
- **Multiple values**: Use `--reload-dirs dir1 --reload-dirs dir2` for multiple directories

### Configuration with Pydantic

Server configuration is now managed with **Pydantic** for:

- **Type safety**: Automatic type validation and conversion
- **Default values**: Sensible defaults for all configuration options
- **Documentation**: Built-in field descriptions and validation
- **Serialization**: Easy conversion to/from JSON and other formats

### CI/CD Improvements

The project now features an upgraded CI/CD pipeline with:

- **macOS-14 runners**: Upgraded from macOS-12 for better resource availability
- **Apple Silicon support**: Native ARM64 testing on macOS
- **Multi-Python testing**: Python 3.10, 3.11, and 3.12 support
- **Dependency caching**: Faster builds with Poetry and pip caching
- **Optimized matrix**: Reduced resource usage with strategic test combinations

## Plugin Development

Create a Python file in your plugin directory:

```python
from typing import Dict, Any
from webhook_bridge.plugin import BasePlugin

class MyPlugin(BasePlugin):
    """Custom webhook plugin."""

    def handle(self) -> Dict[str, Any]:
        """Generic handler for all HTTP methods.

        Returns:
            Dict[str, Any]: Processed result
        """
        # Process your webhook data here
        result = {
            "status": "success",
            "data": {"message": f"Processed: {self.data}"}
        }
        return result

    def get(self) -> Dict[str, Any]:
        """Handle GET requests.

        Returns:
            Dict[str, Any]: Processed result
        """
        # Process GET request
        return {
            "status": "success",
            "data": {"message": "GET request processed"}
        }

    def post(self) -> Dict[str, Any]:
        """Handle POST requests.

        Returns:
            Dict[str, Any]: Processed result
        """
        # Process POST request
        return {
            "status": "success",
            "data": {"message": "POST request processed"}
        }
```

The plugin must:
1. Inherit from `BasePlugin`
2. Implement at least the `handle` method (generic handler)
3. Optionally implement method-specific handlers: `get`, `post`, `put`, `delete`
4. Return a dictionary containing at least:
   - `status`: String indicating success or failure
   - `data`: Dictionary containing the processed result

## Development

### Prerequisites

- Python 3.8 or higher
- nox for development environment management

### Setup Development Environment

1. Clone the repository:
```bash
git clone https://github.com/loonghao/webhook_bridge.git
cd webhook_bridge
```

2. Install nox:
```bash
pip install -r requirements-dev.txt
```

3. Run tests:
```bash
nox -s pytest
```

4. Run linting:
```bash
nox -s lint-fix
```

### Project Structure

```
webhook_bridge/
├── webhook_bridge/      # Main package directory
│   ├── api/            # API endpoints
│   ├── models/         # Data models
│   └── templates/      # HTML templates
├── tests/              # Test files
├── pyproject.toml      # Project metadata and dependencies
└── README.md          # This file
```

## API Endpoints

### Version 1 (`api/v1`)

#### List Plugins
- `GET api/v1/plugins`: List all available webhook plugins
  - Response 200:
    ```json
    {
        "status_code": 200,
        "message": "success",
        "data": {
            "plugins": ["plugin1", "plugin2"]
        }
    }
    ```

#### Execute Plugin
- `GET api/v1/plugin/{plugin_name}`: Execute a specific webhook plugin with GET method
  - Parameters:
    - `plugin_name`: Name of the plugin to execute
    - Query parameters: Data to be processed by the plugin
  - Response 200: Standard response format

- `POST api/v1/plugin/{plugin_name}`: Execute a specific webhook plugin with POST method
  - Parameters:
    - `plugin_name`: Name of the plugin to execute
    - Request body: JSON data to be processed by the plugin
  - Response 200: Standard response format

- `PUT api/v1/plugin/{plugin_name}`: Execute a specific webhook plugin with PUT method
  - Parameters:
    - `plugin_name`: Name of the plugin to execute
    - Request body: JSON data to be processed by the plugin
  - Response 200: Standard response format

- `DELETE api/v1/plugin/{plugin_name}`: Execute a specific webhook plugin with DELETE method
  - Parameters:
    - `plugin_name`: Name of the plugin to execute
    - Query parameters: Data to be processed by the plugin
  - Response 200:
    ```json
    {
        "status_code": 200,
        "message": "success",
        "data": {
            "plugin": "example",
            "src_data": {"key": "value"},
            "result": {
                "status": "success",
                "data": {"key": "value"}
            }
        }
    }
    ```
  - Error Responses:
    - 404: Plugin not found
    - 500: Plugin execution failed

## Error Handling

The API uses standard HTTP status codes and returns detailed error messages:

```json
{
    "status_code": 404,
    "message": "Plugin not found",
    "data": {
        "error": "Plugin not found"
    }
}
```

## Contributing

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request
