#!/bin/bash

# Quick Start Script for Webhook Bridge
# This script provides a simple way to start the webhook bridge service

set -e

# Colors
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
RED='\033[0;31m'
NC='\033[0m'

# Configuration
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"
CONFIG_FILE="$PROJECT_ROOT/config.yaml"
PID_FILE="$PROJECT_ROOT/webhook-bridge.pid"

# Default values
ENVIRONMENT="dev"
DAEMON=false
STOP=false
STATUS=false
RESTART=false

show_help() {
    cat << EOF
Webhook Bridge Quick Start Script

Usage: $0 [OPTIONS]

Options:
    -e, --env ENVIRONMENT    Set environment (dev, prod) [default: dev]
    -d, --daemon            Run as daemon (background)
    -s, --stop              Stop running service
    -r, --restart           Restart service
    --status                Show service status
    -h, --help              Show this help message

Examples:
    $0                      # Start in development mode
    $0 -e prod -d           # Start in production mode as daemon
    $0 --stop               # Stop the service
    $0 --status             # Check service status

EOF
}

log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

check_prerequisites() {
    log_info "Checking prerequisites..."
    
    # Check if binaries exist
    if [ ! -f "$PROJECT_ROOT/build/webhook-bridge-server" ]; then
        log_error "webhook-bridge-server not found. Please run deployment script first."
        log_info "Run: ./scripts/deploy.sh"
        exit 1
    fi
    
    # Check if Python virtual environment exists
    if [ ! -d "$PROJECT_ROOT/.venv" ]; then
        log_error "Python virtual environment not found. Please run deployment script first."
        log_info "Run: ./scripts/deploy.sh"
        exit 1
    fi
    
    # Check configuration
    if [ ! -f "$CONFIG_FILE" ]; then
        log_warning "Configuration file not found. Creating default configuration..."
        case $ENVIRONMENT in
            "prod")
                cp "$PROJECT_ROOT/config.prod.yaml" "$CONFIG_FILE"
                ;;
            *)
                cp "$PROJECT_ROOT/config.dev.yaml" "$CONFIG_FILE"
                ;;
        esac
    fi
    
    log_success "Prerequisites check passed"
}

start_python_executor() {
    log_info "Starting Python executor..."
    cd "$PROJECT_ROOT"
    
    # Activate Python virtual environment
    source .venv/bin/activate || source .venv/Scripts/activate
    
    # Start Python executor in background
    if [ "$DAEMON" = true ]; then
        nohup python python_executor/main.py --config config.yaml > python_executor.log 2>&1 &
        echo $! > python_executor.pid
    else
        python python_executor/main.py --config config.yaml &
        echo $! > python_executor.pid
    fi
    
    # Wait a moment for Python executor to start
    sleep 2
    
    log_success "Python executor started (PID: $(cat python_executor.pid))"
}

start_go_server() {
    log_info "Starting Go server..."
    cd "$PROJECT_ROOT"
    
    # Start Go server
    if [ "$DAEMON" = true ]; then
        nohup ./build/webhook-bridge-server > webhook-bridge.log 2>&1 &
        echo $! > "$PID_FILE"
        log_success "Webhook bridge started as daemon (PID: $(cat "$PID_FILE"))"
        log_info "Logs: tail -f webhook-bridge.log"
    else
        ./build/webhook-bridge-server &
        echo $! > "$PID_FILE"
        log_success "Webhook bridge started (PID: $(cat "$PID_FILE"))"
        
        # Wait for the process
        wait
    fi
}

stop_service() {
    log_info "Stopping webhook bridge service..."
    
    # Stop Go server
    if [ -f "$PID_FILE" ]; then
        PID=$(cat "$PID_FILE")
        if kill -0 "$PID" 2>/dev/null; then
            kill "$PID"
            rm -f "$PID_FILE"
            log_success "Go server stopped (PID: $PID)"
        else
            log_warning "Go server process not found"
            rm -f "$PID_FILE"
        fi
    else
        log_warning "PID file not found for Go server"
    fi
    
    # Stop Python executor
    if [ -f "$PROJECT_ROOT/python_executor.pid" ]; then
        PID=$(cat "$PROJECT_ROOT/python_executor.pid")
        if kill -0 "$PID" 2>/dev/null; then
            kill "$PID"
            rm -f "$PROJECT_ROOT/python_executor.pid"
            log_success "Python executor stopped (PID: $PID)"
        else
            log_warning "Python executor process not found"
            rm -f "$PROJECT_ROOT/python_executor.pid"
        fi
    else
        log_warning "PID file not found for Python executor"
    fi
}

show_status() {
    log_info "Checking service status..."
    
    # Check Go server
    if [ -f "$PID_FILE" ]; then
        PID=$(cat "$PID_FILE")
        if kill -0 "$PID" 2>/dev/null; then
            log_success "Go server is running (PID: $PID)"
        else
            log_error "Go server is not running (stale PID file)"
        fi
    else
        log_error "Go server is not running"
    fi
    
    # Check Python executor
    if [ -f "$PROJECT_ROOT/python_executor.pid" ]; then
        PID=$(cat "$PROJECT_ROOT/python_executor.pid")
        if kill -0 "$PID" 2>/dev/null; then
            log_success "Python executor is running (PID: $PID)"
        else
            log_error "Python executor is not running (stale PID file)"
        fi
    else
        log_error "Python executor is not running"
    fi
    
    # Check if services are responding
    if command -v curl &> /dev/null; then
        log_info "Testing service endpoints..."
        
        # Try to get port from config or use default
        PORT=$(grep -E "^\s*port:" "$CONFIG_FILE" | head -1 | awk '{print $2}' || echo "8080")
        
        if curl -s "http://localhost:$PORT/health" > /dev/null; then
            log_success "HTTP server is responding on port $PORT"
        else
            log_warning "HTTP server is not responding on port $PORT"
        fi
    fi
}

restart_service() {
    log_info "Restarting webhook bridge service..."
    stop_service
    sleep 2
    start_service
}

start_service() {
    check_prerequisites
    start_python_executor
    start_go_server
}

# Parse command line arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        -e|--env)
            ENVIRONMENT="$2"
            shift 2
            ;;
        -d|--daemon)
            DAEMON=true
            shift
            ;;
        -s|--stop)
            STOP=true
            shift
            ;;
        -r|--restart)
            RESTART=true
            shift
            ;;
        --status)
            STATUS=true
            shift
            ;;
        -h|--help)
            show_help
            exit 0
            ;;
        *)
            log_error "Unknown option: $1"
            show_help
            exit 1
            ;;
    esac
done

# Execute requested action
if [ "$STOP" = true ]; then
    stop_service
elif [ "$STATUS" = true ]; then
    show_status
elif [ "$RESTART" = true ]; then
    restart_service
else
    # Default action: start
    log_info "Starting Webhook Bridge in $ENVIRONMENT mode..."
    if [ "$DAEMON" = true ]; then
        log_info "Running as daemon..."
    fi
    
    # Setup signal handlers for graceful shutdown
    trap 'log_info "Received interrupt signal, stopping..."; stop_service; exit 0' INT TERM
    
    start_service
fi
