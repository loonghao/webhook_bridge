#!/bin/bash
set -e

# Docker entrypoint script for webhook-bridge hybrid architecture
# This script starts both the Python executor and Go server

# Default configuration
CONFIG_FILE="/app/config.yaml"
PYTHON_EXECUTOR_HOST="0.0.0.0"
PYTHON_EXECUTOR_PORT="50051"
LOG_LEVEL="${LOG_LEVEL:-INFO}"

# Function to log messages
log() {
    echo "[$(date '+%Y-%m-%d %H:%M:%S')] $1"
}

# Function to check if a port is available
wait_for_port() {
    local host=$1
    local port=$2
    local timeout=${3:-30}
    local count=0

    log "Waiting for $host:$port to be available..."
    while ! nc -z "$host" "$port" 2>/dev/null; do
        if [ $count -ge $timeout ]; then
            log "ERROR: Timeout waiting for $host:$port"
            return 1
        fi
        sleep 1
        count=$((count + 1))
    done
    log "$host:$port is available"
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

    # Check if the process is still running
    if ! kill -0 $PYTHON_PID 2>/dev/null; then
        log "ERROR: Python executor process died immediately after startup"
        log "This usually indicates a configuration or dependency issue"
        exit 1
    fi

    # Wait for Python executor to be ready
    log "Waiting for Python executor to bind to port $PYTHON_EXECUTOR_PORT..."
    if ! wait_for_port "$PYTHON_EXECUTOR_HOST" "$PYTHON_EXECUTOR_PORT" 30; then
        log "ERROR: Python executor failed to start or bind to port"
        log "Checking if process is still running..."
        if kill -0 $PYTHON_PID 2>/dev/null; then
            log "Process is running but not responding on expected port"
        else
            log "Process has terminated"
        fi
        kill $PYTHON_PID 2>/dev/null || true
        exit 1
    fi

    log "Python executor is ready and listening on $PYTHON_EXECUTOR_HOST:$PYTHON_EXECUTOR_PORT"
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

    # Verify Python executor is still running and accessible
    if ! check_port_in_use "$PYTHON_EXECUTOR_PORT"; then
        log "ERROR: Python executor is no longer accessible on port $PYTHON_EXECUTOR_PORT"
        log "Go server will not be able to connect to Python executor"
        exit 1
    fi

    log "✓ Python executor is accessible, starting Go server..."

    # Start Go server
    exec webhook-bridge-server --config "$CONFIG_FILE"
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

# Main execution
log "Starting webhook-bridge hybrid architecture..."
log "Config file: $CONFIG_FILE"
log "Python executor: $PYTHON_EXECUTOR_HOST:$PYTHON_EXECUTOR_PORT"
log "Log level: $LOG_LEVEL"

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
start_python_executor

# Start Go server (this will block)
start_go_server
