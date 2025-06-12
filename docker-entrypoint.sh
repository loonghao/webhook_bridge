#!/bin/bash
set -e

# Docker entrypoint script for webhook-bridge unified architecture
# This script can start the unified service or individual components

# Default configuration
CONFIG_FILE="/app/config.yaml"
PYTHON_EXECUTOR_HOST="0.0.0.0"
PYTHON_EXECUTOR_PORT="50051"
LOG_LEVEL="${LOG_LEVEL:-INFO}"

# Function to log messages
log() {
    echo "[$(date '+%Y-%m-%d %H:%M:%S')] $1"
}

# Function to check if a port is available with enhanced diagnostics
wait_for_port() {
    local host=$1
    local port=$2
    local timeout=${3:-30}
    local count=0
    local check_interval=1
    local last_progress_report=0
    local progress_interval=5

    log "🔍 Waiting for $host:$port to be available (timeout: ${timeout}s)..."

    while ! nc -z "$host" "$port" 2>/dev/null; do
        if [ $count -ge $timeout ]; then
            log "❌ TIMEOUT: Port $host:$port not available after ${timeout} seconds"
            log "🔧 Troubleshooting information:"
            log "   - Check if the service is running"
            log "   - Verify port configuration matches expected value"
            log "   - Check for port conflicts or firewall issues"
            if command -v netstat >/dev/null 2>&1; then
                log "   - Active connections on port $port:"
                netstat -an | grep ":$port " || log "     No connections found on port $port"
            fi
            return 1
        fi

        # Progress reporting every 5 seconds
        if [ $((count - last_progress_report)) -ge $progress_interval ]; then
            log "⏳ Still waiting for $host:$port... (${count}/${timeout}s elapsed)"
            last_progress_report=$count
        fi

        sleep $check_interval
        count=$((count + check_interval))
    done

    log "✅ Port $host:$port is now available (took ${count}s)"
}

# Function to validate configuration
validate_config() {
    log "Validating configuration..."

    # Check if config file exists
    if [ ! -f "$CONFIG_FILE" ]; then
        log "ERROR: Config file not found at $CONFIG_FILE"
        return 1
    fi

    # Check if config file is readable
    if [ ! -r "$CONFIG_FILE" ]; then
        log "ERROR: Config file $CONFIG_FILE is not readable"
        return 1
    fi

    log "Configuration file validated: $CONFIG_FILE"
    return 0
}

# Function to check if a port is in use
check_port_in_use() {
    local port=$1
    if nc -z localhost "$port" 2>/dev/null; then
        return 0  # Port is in use
    else
        return 1  # Port is free
    fi
}

# Function to perform comprehensive service health check
check_service_health() {
    local service_name=$1
    local pid=$2
    local host=$3
    local port=$4

    log "🏥 Performing health check for $service_name..."

    # Check if process is running
    if [ -n "$pid" ]; then
        if kill -0 "$pid" 2>/dev/null; then
            log "✅ Process $service_name (PID: $pid) is running"
        else
            log "❌ Process $service_name (PID: $pid) is not running"
            return 1
        fi
    fi

    # Check port accessibility
    if [ -n "$port" ]; then
        if nc -z "$host" "$port" 2>/dev/null; then
            log "✅ Port $host:$port is accessible"
        else
            log "❌ Port $host:$port is not accessible"
            return 1
        fi
    fi

    # Additional system checks
    log "📊 System resource status:"
    if command -v free >/dev/null 2>&1; then
        local mem_info=$(free -h | grep "Mem:" | awk '{print "Used: " $3 "/" $2 " (" $3/$2*100 "%)"}')
        log "   Memory: $mem_info"
    fi

    if command -v df >/dev/null 2>&1; then
        local disk_info=$(df -h / | tail -1 | awk '{print "Used: " $3 "/" $2 " (" $5 ")"}')
        log "   Disk: $disk_info"
    fi

    log "✅ Health check for $service_name completed successfully"
    return 0
}

# Function to diagnose startup failures
diagnose_startup_failure() {
    local service_name=$1
    local pid=$2
    local expected_port=$3

    log "🔍 Diagnosing startup failure for $service_name..."

    # Check if process exists
    if [ -n "$pid" ]; then
        if kill -0 "$pid" 2>/dev/null; then
            log "ℹ️  Process is still running (PID: $pid)"

            # Check what ports the process is using
            if command -v lsof >/dev/null 2>&1; then
                log "📡 Ports used by process $pid:"
                lsof -p "$pid" -i 2>/dev/null || log "   No network connections found"
            elif command -v netstat >/dev/null 2>&1; then
                log "📡 Network connections:"
                netstat -tulpn 2>/dev/null | grep "$pid" || log "   No connections found for PID $pid"
            fi
        else
            log "💀 Process has terminated (PID: $pid)"
        fi
    fi

    # Check port status
    if [ -n "$expected_port" ]; then
        if check_port_in_use "$expected_port"; then
            log "⚠️  Port $expected_port is in use by another process"
            if command -v lsof >/dev/null 2>&1; then
                log "🔍 Process using port $expected_port:"
                lsof -i ":$expected_port" 2>/dev/null || log "   Unable to determine process"
            fi
        else
            log "ℹ️  Port $expected_port is available"
        fi
    fi

    # Check system resources
    log "🖥️  System status:"
    if command -v uptime >/dev/null 2>&1; then
        log "   Load: $(uptime | awk -F'load average:' '{print $2}')"
    fi

    log "🔍 Diagnosis completed for $service_name"
}

# Function to start Python executor
start_python_executor() {
    log "Starting Python executor..."

    # Check if config file exists and is readable
    if [ ! -f "$CONFIG_FILE" ]; then
        log "WARNING: Config file not found at $CONFIG_FILE"
        # Try alternative locations
        if [ -f "/app/config/config.yaml" ]; then
            CONFIG_FILE="/app/config/config.yaml"
            log "Using config file at $CONFIG_FILE"
        elif [ -f "/app/config.yaml" ]; then
            CONFIG_FILE="/app/config.yaml"
            log "Using config file at $CONFIG_FILE"
        else
            log "ERROR: No config file found in expected locations"
            log "Searched: /app/config.yaml, /app/config/config.yaml"
            exit 1
        fi
    else
        log "Using config file: $CONFIG_FILE"
    fi

    # Check if the Python executor port is available
    if check_port_in_use "$PYTHON_EXECUTOR_PORT"; then
        log "WARNING: Port $PYTHON_EXECUTOR_PORT is already in use"
        log "This may cause Python executor startup to fail"
    else
        log "Port $PYTHON_EXECUTOR_PORT is available for Python executor"
    fi
    
    # Verify Python environment before starting
    log "🔍 Verifying Python environment..."
    if ! python --version >/dev/null 2>&1; then
        log "❌ ERROR: Python is not available"
        exit 1
    fi
    log "✅ Python version: $(python --version 2>&1)"

    # Check if Python executor script exists
    if [ ! -f "/app/python_executor/main.py" ]; then
        log "❌ ERROR: Python executor script not found at /app/python_executor/main.py"
        exit 1
    fi
    log "✅ Python executor script found"

    # Test Python import capabilities
    log "🔍 Testing Python dependencies..."
    if ! python -c "import grpc, yaml, asyncio; print('Dependencies OK')" 2>/dev/null; then
        log "❌ ERROR: Required Python dependencies not available"
        log "🔧 Attempting to list installed packages..."
        python -m pip list 2>/dev/null || log "Unable to list packages"
        exit 1
    fi
    log "✅ Python dependencies verified"

    # Start Python executor in background with explicit port configuration
    log "Starting Python executor with explicit port: $PYTHON_EXECUTOR_PORT"
    log "Command: python /app/python_executor/main.py --host $PYTHON_EXECUTOR_HOST --port $PYTHON_EXECUTOR_PORT --log-level $LOG_LEVEL --config $CONFIG_FILE"

    python /app/python_executor/main.py \
        --host "$PYTHON_EXECUTOR_HOST" \
        --port "$PYTHON_EXECUTOR_PORT" \
        --log-level "$LOG_LEVEL" \
        --config "$CONFIG_FILE" &

    PYTHON_PID=$!
    log "Python executor started with PID: $PYTHON_PID"

    # Give Python executor a moment to initialize
    sleep 2

    # Initial process health check
    log "🔍 Performing initial health check..."
    if ! kill -0 $PYTHON_PID 2>/dev/null; then
        log "❌ ERROR: Python executor process died immediately after startup"
        log "🔧 This usually indicates a configuration or dependency issue"
        diagnose_startup_failure "Python executor" "$PYTHON_PID" "$PYTHON_EXECUTOR_PORT"
        exit 1
    fi
    log "✅ Python executor process is running (PID: $PYTHON_PID)"

    # Wait for Python executor to be ready with enhanced monitoring
    log "⏳ Waiting for Python executor to bind to port $PYTHON_EXECUTOR_PORT..."
    if ! wait_for_port "$PYTHON_EXECUTOR_HOST" "$PYTHON_EXECUTOR_PORT" 45; then
        log "❌ ERROR: Python executor failed to start or bind to port"
        log "🔍 Performing detailed diagnosis..."
        diagnose_startup_failure "Python executor" "$PYTHON_PID" "$PYTHON_EXECUTOR_PORT"

        # Attempt graceful shutdown before exit
        log "🛑 Attempting graceful shutdown of Python executor..."
        kill $PYTHON_PID 2>/dev/null || true
        sleep 2
        kill -9 $PYTHON_PID 2>/dev/null || true
        exit 1
    fi

    # Final health verification
    log "🏥 Performing final health verification..."
    if check_service_health "Python executor" "$PYTHON_PID" "$PYTHON_EXECUTOR_HOST" "$PYTHON_EXECUTOR_PORT"; then
        log "🎉 Python executor is ready and healthy on $PYTHON_EXECUTOR_HOST:$PYTHON_EXECUTOR_PORT"
    else
        log "⚠️  Python executor started but health check failed"
        diagnose_startup_failure "Python executor" "$PYTHON_PID" "$PYTHON_EXECUTOR_PORT"
        exit 1
    fi
}

# Function to verify Go server configuration
verify_go_server_config() {
    log "Verifying Go server configuration..."

    # Check if config file contains executor port configuration
    if command -v grep >/dev/null 2>&1; then
        if grep -q "port: $PYTHON_EXECUTOR_PORT" "$CONFIG_FILE"; then
            log "✓ Config file contains matching executor port: $PYTHON_EXECUTOR_PORT"
        else
            log "⚠ Config file may not contain matching executor port"
            log "Expected: port: $PYTHON_EXECUTOR_PORT"
        fi
    else
        log "grep not available, skipping config verification"
    fi
}

# Function to start Go server
start_go_server() {
    log "Starting Go server..."

    # Verify config file is still accessible (should be set by Python executor function)
    if [ ! -f "$CONFIG_FILE" ]; then
        log "ERROR: Config file $CONFIG_FILE is not accessible"
        log "This should not happen if Python executor started successfully"
        exit 1
    fi

    log "Using config file: $CONFIG_FILE"

    # Verify configuration consistency
    verify_go_server_config

    # Comprehensive pre-startup verification
    log "🔍 Performing pre-startup verification..."

    # Verify Python executor is still running and accessible
    if ! check_port_in_use "$PYTHON_EXECUTOR_PORT"; then
        log "❌ ERROR: Python executor is no longer accessible on port $PYTHON_EXECUTOR_PORT"
        log "🔧 Go server will not be able to connect to Python executor"
        if [ -n "$PYTHON_PID" ]; then
            diagnose_startup_failure "Python executor" "$PYTHON_PID" "$PYTHON_EXECUTOR_PORT"
        fi
        exit 1
    fi

    # Final health check of Python executor before starting Go server
    if [ -n "$PYTHON_PID" ]; then
        if ! check_service_health "Python executor" "$PYTHON_PID" "$PYTHON_EXECUTOR_HOST" "$PYTHON_EXECUTOR_PORT"; then
            log "❌ ERROR: Python executor health check failed before starting Go server"
            exit 1
        fi
    fi

    log "✅ All pre-startup checks passed, starting Go server..."
    log "🚀 Executing: webhook-bridge serve --config $CONFIG_FILE"

    # Start Go server (this will block)
    exec webhook-bridge serve --config "$CONFIG_FILE"
}

# Function to handle shutdown
shutdown() {
    log "Received shutdown signal, stopping services..."
    
    # Kill Python executor if it's running
    if [ ! -z "$PYTHON_PID" ]; then
        log "Stopping Python executor (PID: $PYTHON_PID)"
        kill $PYTHON_PID 2>/dev/null || true
        wait $PYTHON_PID 2>/dev/null || true
    fi
    
    log "Services stopped"
    exit 0
}

# Set up signal handlers
trap shutdown SIGTERM SIGINT

# Function to display startup banner and system info
display_startup_info() {
    log "🚀 =============================================="
    log "🚀 Starting webhook-bridge unified architecture"
    log "🚀 =============================================="
    log "📋 Configuration:"
    log "   Config file: $CONFIG_FILE"
    log "   Python executor: $PYTHON_EXECUTOR_HOST:$PYTHON_EXECUTOR_PORT"
    log "   Log level: $LOG_LEVEL"
    log "   Command: $*"
    log ""
    log "🖥️  System Information:"
    if command -v uname >/dev/null 2>&1; then
        log "   OS: $(uname -s) $(uname -r)"
    fi
    if command -v python3 >/dev/null 2>&1; then
        log "   Python: $(python3 --version 2>&1)"
    elif command -v python >/dev/null 2>&1; then
        log "   Python: $(python --version 2>&1)"
    fi
    log "   Container: Docker"
    log ""
    log "🔧 Available tools:"
    for tool in nc netstat lsof grep free df uptime; do
        if command -v "$tool" >/dev/null 2>&1; then
            log "   ✅ $tool"
        else
            log "   ❌ $tool (not available)"
        fi
    done
    log "🚀 =============================================="
}

# Main execution
display_startup_info "$@"

# Check if we should use start command or legacy mode
if [ "$1" = "webhook-bridge" ] && [ "$2" = "start" ]; then
    log "🚀 Starting unified service mode..."

    # Validate initial configuration
    if ! validate_config; then
        log "ERROR: Configuration validation failed, attempting to find config file..."
        # Try to find config file in common locations
        if [ -f "/app/config.yaml" ]; then
            CONFIG_FILE="/app/config.yaml"
            log "Found config file at $CONFIG_FILE"
        elif [ -f "/app/config/config.yaml" ]; then
            CONFIG_FILE="/app/config/config.yaml"
            log "Found config file at $CONFIG_FILE"
        else
            log "FATAL: No valid config file found"
            exit 1
        fi
    fi

    # Start unified service (this will block and manage everything)
    log "🚀 Executing unified service: webhook-bridge start --config $CONFIG_FILE"
    exec webhook-bridge start --config "$CONFIG_FILE"

elif [ "$1" = "webhook-bridge" ]; then
    log "🚀 Starting webhook-bridge with custom command: $*"

    # Validate configuration
    if ! validate_config; then
        log "WARNING: Configuration validation failed, using default config"
    fi

    # Execute the webhook-bridge command directly
    exec "$@"

else
    log "🚀 Starting legacy mode (separate Python executor and Go server)..."

    # Validate initial configuration
    if ! validate_config; then
        log "ERROR: Configuration validation failed, attempting to find config file..."
        # Try to find config file in common locations
        if [ -f "/app/config.yaml" ]; then
            CONFIG_FILE="/app/config.yaml"
            log "Found config file at $CONFIG_FILE"
        elif [ -f "/app/config/config.yaml" ]; then
            CONFIG_FILE="/app/config/config.yaml"
            log "Found config file at $CONFIG_FILE"
        else
            log "FATAL: No valid config file found"
            exit 1
        fi
    fi

    # Start Python executor first
    log "📍 Phase 1: Starting Python executor..."
    start_python_executor

    log "📍 Phase 2: Starting Go server..."
    # Start Go server (this will block)
    start_go_server
fi
