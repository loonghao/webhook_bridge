<p align="center">
<img src="https://i.imgur.com/d9UWkck.png" alt="logo"></a>
</p>

# Webhook Bridge

A flexible and powerful webhook integration platform that allows you to bridge webhooks into your tools or internal integrations.

[![Python Version](https://img.shields.io/pypi/pyversions/webhook-bridge)](https://img.shields.io/pypi/pyversions/webhook-bridge)
[![Nox](https://img.shields.io/badge/%F0%9F%A6%8A-Nox-D85E00.svg)](https://github.com/wntrblm/nox)
[![PyPI Version](https://img.shields.io/pypi/v/webhook-bridge?color=green)](https://pypi.org/project/webhook-bridge/)
[![Downloads](https://static.pepy.tech/badge/webhook-bridge)](https://pepy.tech/project/webhook-bridge)
[![Downloads](https://static.pepy.tech/badge/webhook-bridge/month)](https://pepy.tech/project/webhook-bridge)
[![Downloads](https://static.pepy.tech/badge/webhook-bridge/week)](https://pepy.tech/project/webhook-bridge)
[![License](https://img.shields.io/pypi/l/webhook-bridge)](https://pypi.org/project/webhook-bridge/)
[![PyPI Format](https://img.shields.io/pypi/format/webhook-bridge)](https://pypi.org/project/webhook-bridge/)
[![Maintenance](https://img.shields.io/badge/Maintained%3F-yes-green.svg)](https://github.com/loonghao/webhook-bridge/graphs/commit-activity)

<p align="center">
<img src="https://i.imgur.com/31RO4xN.png" alt="logo"></a>
</p>

## Features

- 🚀 **API Versioning**: Support for versioned API endpoints (`/v1`, `/latest`)
- 🔌 **Plugin System**: Dynamic plugin loading and execution
- 🛠️ **Flexible Configuration**: Extensive CLI and programmatic configuration options
- 📝 **Rich Documentation**: Interactive API documentation with Swagger UI and ReDoc
- 🔒 **Secure**: Built-in security features and error handling
- 📊 **Logging**: Comprehensive logging and error tracking

## Installation

You can install via pip:

```bash
pip install webhook_bridge
```

Or install from source:

```bash
git clone https://github.com/loonghao/webhook_bridge.git
cd webhook_bridge
pip install -e .
```

## Quick Start

1. Launch the server:

```bash
webhook-bridge --host localhost --port 8000
```

2. Test with a sample request:

```bash
curl -X POST "http://localhost:8000/v1/plugin/example" \
     -H "Content-Type: application/json" \
     -d '{"message": "Hello, World!"}'
```

3. Access the API documentation:
   - Swagger UI: `http://localhost:8000/docs`
   - ReDoc: `http://localhost:8000/redoc`

## Configuration

### Command Line Options

```bash
webhook-bridge --help
```

#### Server Configuration
- `--host`: Host to bind the server to (default: "0.0.0.0")
- `--port`: Port to bind the server to (default: 8000)
- `--log-level`: Logging level (DEBUG/INFO/WARNING/ERROR/CRITICAL)

#### API Configuration
- `--title`: API title
- `--description`: API description
- `--docs-url`: URL for API documentation
- `--redoc-url`: URL for ReDoc documentation
- `--openapi-url`: URL for OpenAPI schema

#### Plugin Configuration
- `--plugin-dir`: Directory containing webhook plugins

### Environment Variables
- `WEBHOOK_BRIDGE_SERVER_PLUGINS`: Additional plugin directories (separated by system path separator)

## Plugin Development

Create a Python file in your plugin directory:

```python
from typing import Dict, Any
from webhook_bridge.plugin import BasePlugin

class MyPlugin(BasePlugin):
    """Custom webhook plugin."""

    def run(self, data: Dict[str, Any]) -> Dict[str, Any]:
        """Process webhook data.

        Args:
            data: Input data from webhook

        Returns:
            Dict[str, Any]: Processed result

            Example:
            {
                "status": "success",
                "data": {"key": "processed_value"}
            }
        """
        # Process your webhook data here
        result = {
            "status": "success",
            "data": {"message": f"Processed: {data}"}
        }
        return result
```

The plugin must:
1. Inherit from `BasePlugin`
2. Implement the `run` method
3. Return a dictionary containing at least:
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
- `POST api/v1/plugin/{plugin_name}`: Execute a specific webhook plugin
  - Parameters:
    - `plugin_name`: Name of the plugin to execute
    - Request body: JSON data to be processed by the plugin
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
