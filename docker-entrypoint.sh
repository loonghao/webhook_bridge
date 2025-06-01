#!/bin/bash
set -e

# Docker entrypoint script for webhook-bridge hybrid architecture
# This script starts both the Python executor and Go server

# Default configuration
CONFIG_FILE="/app/config/config.yaml"
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

# Function to start Python executor
start_python_executor() {
    log "Starting Python executor..."
    
    # Check if config file exists, use default if not
    if [ ! -f "$CONFIG_FILE" ]; then
        log "Config file not found at $CONFIG_FILE, using default config"
        CONFIG_FILE="/app/config.yaml"
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
    
    # Check if config file exists, use default if not
    if [ ! -f "$CONFIG_FILE" ]; then
        log "Config file not found at $CONFIG_FILE, using default config"
        CONFIG_FILE="/app/config.yaml"
    fi
    
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

# Start Python executor first
start_python_executor

# Start Go server (this will block)
start_go_server
