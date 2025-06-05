# Python Interpreter Management

This document describes the new Python interpreter management features in webhook-bridge, which allow you to configure and manage multiple Python interpreters dynamically through the dashboard.

## Overview

The Python interpreter management system provides:

- **Multiple Interpreter Support**: Configure and manage multiple Python interpreters
- **Dynamic Switching**: Switch between interpreters without restarting the service
- **Validation System**: Validate interpreters and their dependencies
- **Dashboard Integration**: Manage interpreters through the web dashboard
- **Connection Management**: Monitor and control the connection to Python executor service

## Features

### 1. Multiple Python Interpreters

You can now configure multiple Python interpreters with different settings:

```yaml
python:
  active_interpreter: "system-python"
  interpreters:
    system-python:
      name: "System Python 3.9"
      path: "/usr/bin/python3"
      venv_path: ".venv"
      use_uv: false
      required_packages:
        - "grpcio"
        - "grpcio-tools"
      validated: true
    
    uv-python:
      name: "UV Python Environment"
      path: ".venv/bin/python"
      use_uv: true
      required_packages:
        - "grpcio"
        - "grpcio-tools"
        - "fastapi"
      validated: true
```

### 2. Dashboard Management

Access the Python interpreter management through the dashboard:

- **Python Interpreters Page** (`/interpreters`): Manage all configured interpreters
- **Connection Status Page** (`/connection`): Monitor service connection status
- **Configuration Page** (`/config`): Basic Python configuration

### 3. Dynamic Connection Management

The connection manager handles:

- Automatic reconnection on interpreter changes
- Connection status monitoring
- Process management
- Error handling and recovery

## Usage

### Adding a New Interpreter

1. Navigate to the **Python Interpreters** page in the dashboard
2. Click **Add Interpreter**
3. Fill in the interpreter details:
   - **Name**: Display name for the interpreter
   - **Path**: Path to the Python executable
   - **Virtual Environment**: Optional venv path
   - **Use UV**: Whether to use UV for package management
   - **Required Packages**: List of required packages
4. Click **Add Interpreter**

### Activating an Interpreter

1. Go to the **Python Interpreters** page
2. Find the interpreter you want to activate
3. Click **Activate** next to the interpreter
4. The service will automatically reconnect using the new interpreter

### Validating Interpreters

1. On the **Python Interpreters** page
2. Click **Validate** next to any interpreter
3. The system will check:
   - Python executable exists and is accessible
   - Required packages are installed
   - Virtual environment is properly configured

### Monitoring Connection Status

1. Visit the **Connection Status** page
2. View real-time connection information:
   - Connection status (connected/disconnected/reconnecting)
   - Active interpreter
   - Process information
   - Uptime and statistics

### Testing Connection

1. On the **Connection Status** page
2. Click **Test Connection** to verify the connection works
3. View test results and response times

## API Endpoints

The system provides REST API endpoints for programmatic access:

### Interpreter Management
- `GET /api/dashboard/interpreters` - List all interpreters
- `POST /api/dashboard/interpreters` - Add new interpreter
- `DELETE /api/dashboard/interpreters/{name}` - Remove interpreter
- `POST /api/dashboard/interpreters/{name}/validate` - Validate interpreter
- `POST /api/dashboard/interpreters/{name}/activate` - Activate interpreter
- `GET /api/dashboard/interpreters/discover` - Auto-discover interpreters

### Connection Management
- `GET /api/dashboard/connection` - Get connection status
- `POST /api/dashboard/connection/reconnect` - Force reconnection
- `POST /api/dashboard/connection/test` - Test connection

## Configuration

### Basic Configuration

```yaml
python:
  # Legacy single interpreter configuration
  interpreter: "auto"  # or path to Python executable
  active_interpreter: "system-python"  # Name of active interpreter
  auto_download_uv: true
  venv_path: ".venv"
```

### Advanced Configuration

```yaml
python:
  interpreters:
    my-interpreter:
      name: "My Custom Python"
      path: "/path/to/python"
      venv_path: "/path/to/venv"
      use_uv: true
      required_packages:
        - "grpcio"
        - "grpcio-tools"
        - "custom-package"
      environment:
        CUSTOM_VAR: "value"
        PYTHONPATH: "/custom/path"
      validated: false
```

## Environment Variables

You can override configuration using environment variables:

- `WEBHOOK_BRIDGE_PYTHON_PATH` - Override Python interpreter path
- `WEBHOOK_BRIDGE_ACTIVE_INTERPRETER` - Set active interpreter name
- `WEBHOOK_BRIDGE_EXECUTOR_HOST` - Executor service host
- `WEBHOOK_BRIDGE_EXECUTOR_PORT` - Executor service port

## Troubleshooting

### Common Issues

1. **Interpreter Not Found**
   - Verify the Python executable path is correct
   - Check file permissions
   - Ensure Python is in the system PATH

2. **Package Installation Errors**
   - Verify required packages are installed
   - Check virtual environment activation
   - Review package compatibility

3. **Connection Failures**
   - Check if the Python executor process is running
   - Verify network connectivity
   - Review executor service logs

### Debugging

1. Enable debug logging in configuration:
   ```yaml
   logging:
     level: "debug"
   ```

2. Check the **Connection Status** page for detailed error information

3. Review logs in the **Logs** page or log files

## Migration from Legacy Configuration

Existing configurations will continue to work. To migrate to the new system:

1. Keep your existing `python.interpreter` setting
2. Add new interpreters to the `python.interpreters` section
3. Set `python.active_interpreter` to specify which interpreter to use
4. Use the dashboard to manage interpreters going forward

## Best Practices

1. **Validation**: Always validate interpreters after adding them
2. **Dependencies**: Keep required packages lists up to date
3. **Monitoring**: Regularly check connection status
4. **Backup**: Keep configuration backups before making changes
5. **Testing**: Test connections after making changes

## Security Considerations

1. **File Permissions**: Ensure Python executables have appropriate permissions
2. **Path Validation**: Only use trusted Python interpreter paths
3. **Environment Variables**: Be careful with custom environment variables
4. **Network Security**: Secure the executor service network connection
