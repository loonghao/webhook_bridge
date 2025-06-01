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
    
    # Start Python executor in background
    python /app/python_executor/main.py \
        --host "$PYTHON_EXECUTOR_HOST" \
        --port "$PYTHON_EXECUTOR_PORT" \
        --log-level "$LOG_LEVEL" \
        --config "$CONFIG_FILE" &
    
    PYTHON_PID=$!
    log "Python executor started with PID: $PYTHON_PID"
    
    # Wait for Python executor to be ready
    if ! wait_for_port "$PYTHON_EXECUTOR_HOST" "$PYTHON_EXECUTOR_PORT" 30; then
        log "ERROR: Python executor failed to start"
        kill $PYTHON_PID 2>/dev/null || true
        exit 1
    fi
    
    log "Python executor is ready"
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
