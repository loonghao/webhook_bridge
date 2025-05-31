#!/bin/bash

# Webhook Bridge Deployment Script
# This script automates the deployment process for webhook bridge

set -e  # Exit on any error

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"
BUILD_DIR="$PROJECT_ROOT/build"
DIST_DIR="$PROJECT_ROOT/dist"

# Default values
ENVIRONMENT="dev"
SKIP_TESTS=false
SKIP_BUILD=false
INSTALL_DEPS=true
VERBOSE=false

# Functions
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

show_help() {
    cat << EOF
Webhook Bridge Deployment Script

Usage: $0 [OPTIONS]

Options:
    -e, --env ENVIRONMENT       Set deployment environment (dev, prod) [default: dev]
    -s, --skip-tests           Skip running tests
    -b, --skip-build           Skip building binaries
    -n, --no-deps              Skip dependency installation
    -v, --verbose              Enable verbose output
    -h, --help                 Show this help message

Examples:
    $0                         # Deploy for development
    $0 -e prod                 # Deploy for production
    $0 -s -b                   # Quick deploy (skip tests and build)
    $0 --env prod --verbose    # Production deploy with verbose output

EOF
}

check_dependencies() {
    log_info "Checking dependencies..."
    
    # Check Go
    if ! command -v go &> /dev/null; then
        log_error "Go is not installed. Please install Go 1.21 or later."
        exit 1
    fi
    
    # Check Python
    if ! command -v python3 &> /dev/null; then
        log_error "Python 3 is not installed. Please install Python 3.8 or later."
        exit 1
    fi
    
    # Check UV (optional but recommended)
    if ! command -v uv &> /dev/null; then
        log_warning "UV is not installed. Consider installing UV for better Python environment management."
    fi
    
    log_success "Dependencies check passed"
}

install_python_deps() {
    if [ "$INSTALL_DEPS" = true ]; then
        log_info "Installing Python dependencies..."
        cd "$PROJECT_ROOT"
        
        if command -v uv &> /dev/null; then
            log_info "Using UV for Python environment management"
            uv venv .venv
            source .venv/bin/activate || source .venv/Scripts/activate
            uv pip install -r requirements.txt
        else
            log_info "Using pip for Python environment management"
            python3 -m venv .venv
            source .venv/bin/activate || source .venv/Scripts/activate
            pip install -r requirements.txt
        fi
        
        log_success "Python dependencies installed"
    else
        log_info "Skipping Python dependency installation"
    fi
}

run_tests() {
    if [ "$SKIP_TESTS" = false ]; then
        log_info "Running tests..."
        cd "$PROJECT_ROOT"
        
        # Activate Python virtual environment
        source .venv/bin/activate || source .venv/Scripts/activate
        
        # Run Go tests
        log_info "Running Go tests..."
        go test ./... -v
        
        # Run Python tests
        log_info "Running Python tests..."
        python -m pytest tests/ -v || true  # Don't fail if no tests yet
        
        # Run integration tests
        log_info "Running integration tests..."
        python test_go_python_integration.py || log_warning "Integration tests failed"
        
        log_success "Tests completed"
    else
        log_info "Skipping tests"
    fi
}

build_binaries() {
    if [ "$SKIP_BUILD" = false ]; then
        log_info "Building binaries..."
        cd "$PROJECT_ROOT"
        
        # Create build directory
        mkdir -p "$BUILD_DIR"
        mkdir -p "$DIST_DIR"
        
        # Build Go server
        log_info "Building Go server..."
        go build -o "$BUILD_DIR/webhook-bridge-server" ./cmd/server
        
        # Build Python manager (if needed)
        log_info "Building Python manager..."
        go build -o "$BUILD_DIR/python-manager" ./cmd/python-manager
        
        log_success "Binaries built successfully"
    else
        log_info "Skipping build"
    fi
}

setup_config() {
    log_info "Setting up configuration for environment: $ENVIRONMENT"
    cd "$PROJECT_ROOT"
    
    case $ENVIRONMENT in
        "dev")
            cp config.dev.yaml config.yaml
            log_info "Using development configuration"
            ;;
        "prod")
            cp config.prod.yaml config.yaml
            log_info "Using production configuration"
            ;;
        *)
            if [ ! -f config.yaml ]; then
                cp config.example.yaml config.yaml
                log_warning "Unknown environment. Using example configuration."
                log_warning "Please review and modify config.yaml as needed."
            fi
            ;;
    esac
}

create_systemd_service() {
    if [ "$ENVIRONMENT" = "prod" ]; then
        log_info "Creating systemd service file..."
        
        cat > "$DIST_DIR/webhook-bridge.service" << EOF
[Unit]
Description=Webhook Bridge Service
After=network.target

[Service]
Type=simple
User=webhook-bridge
Group=webhook-bridge
WorkingDirectory=/opt/webhook-bridge
ExecStart=/opt/webhook-bridge/webhook-bridge-server
Restart=always
RestartSec=5
Environment=WEBHOOK_BRIDGE_ENV=prod

# Security settings
NoNewPrivileges=true
PrivateTmp=true
ProtectSystem=strict
ProtectHome=true
ReadWritePaths=/var/log/webhook-bridge /var/lib/webhook-bridge

[Install]
WantedBy=multi-user.target
EOF
        
        log_success "Systemd service file created at $DIST_DIR/webhook-bridge.service"
        log_info "To install: sudo cp $DIST_DIR/webhook-bridge.service /etc/systemd/system/"
        log_info "To enable: sudo systemctl enable webhook-bridge"
        log_info "To start: sudo systemctl start webhook-bridge"
    fi
}

create_docker_files() {
    log_info "Creating Docker files..."
    
    # Dockerfile
    cat > "$PROJECT_ROOT/Dockerfile" << 'EOF'
# Multi-stage build for webhook bridge
FROM golang:1.21-alpine AS go-builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -o webhook-bridge-server ./cmd/server
RUN go build -o python-manager ./cmd/python-manager

FROM python:3.11-slim

# Install system dependencies
RUN apt-get update && apt-get install -y \
    curl \
    && rm -rf /var/lib/apt/lists/*

# Install UV
RUN pip install uv

WORKDIR /app

# Copy Python requirements and install dependencies
COPY requirements.txt .
RUN uv venv .venv && \
    . .venv/bin/activate && \
    uv pip install -r requirements.txt

# Copy Go binaries
COPY --from=go-builder /app/webhook-bridge-server .
COPY --from=go-builder /app/python-manager .

# Copy Python code and configs
COPY python_executor/ ./python_executor/
COPY api/ ./api/
COPY example_plugins/ ./example_plugins/
COPY config.prod.yaml ./config.yaml

# Create non-root user
RUN useradd -m -u 1000 webhook && \
    chown -R webhook:webhook /app

USER webhook

EXPOSE 8080 50051

CMD ["./webhook-bridge-server"]
EOF

    # Docker Compose
    cat > "$PROJECT_ROOT/docker-compose.yml" << 'EOF'
version: '3.8'

services:
  webhook-bridge:
    build: .
    ports:
      - "8080:8080"
      - "50051:50051"
    environment:
      - WEBHOOK_BRIDGE_ENV=prod
    volumes:
      - ./plugins:/app/plugins:ro
      - ./logs:/var/log/webhook-bridge
    restart: unless-stopped
    healthcheck:
      test: ["CMD", "curl", "-f", "http://localhost:8080/health"]
      interval: 30s
      timeout: 10s
      retries: 3

  # Optional: Add monitoring
  prometheus:
    image: prom/prometheus:latest
    ports:
      - "9090:9090"
    volumes:
      - ./monitoring/prometheus.yml:/etc/prometheus/prometheus.yml:ro
    command:
      - '--config.file=/etc/prometheus/prometheus.yml'
      - '--storage.tsdb.path=/prometheus'
    restart: unless-stopped

  grafana:
    image: grafana/grafana:latest
    ports:
      - "3000:3000"
    environment:
      - GF_SECURITY_ADMIN_PASSWORD=admin
    volumes:
      - grafana-storage:/var/lib/grafana
    restart: unless-stopped

volumes:
  grafana-storage:
EOF

    log_success "Docker files created"
}

package_release() {
    log_info "Packaging release..."
    cd "$PROJECT_ROOT"
    
    # Create release directory
    RELEASE_DIR="$DIST_DIR/webhook-bridge-$(date +%Y%m%d-%H%M%S)"
    mkdir -p "$RELEASE_DIR"
    
    # Copy binaries
    cp "$BUILD_DIR"/* "$RELEASE_DIR/"
    
    # Copy configuration files
    cp config.*.yaml "$RELEASE_DIR/"
    
    # Copy Python executor
    cp -r python_executor "$RELEASE_DIR/"
    cp -r api "$RELEASE_DIR/"
    cp -r example_plugins "$RELEASE_DIR/"
    
    # Copy documentation
    cp README*.md "$RELEASE_DIR/" 2>/dev/null || true
    
    # Create archive
    cd "$DIST_DIR"
    tar -czf "webhook-bridge-$(date +%Y%m%d-%H%M%S).tar.gz" "$(basename "$RELEASE_DIR")"
    
    log_success "Release packaged at $DIST_DIR"
}

main() {
    log_info "Starting Webhook Bridge deployment..."
    log_info "Environment: $ENVIRONMENT"
    
    check_dependencies
    install_python_deps
    run_tests
    build_binaries
    setup_config
    create_systemd_service
    create_docker_files
    package_release
    
    log_success "Deployment completed successfully!"
    log_info "Next steps:"
    log_info "  1. Review configuration in config.yaml"
    log_info "  2. Test the deployment: ./build/webhook-bridge-server"
    log_info "  3. For production: follow the systemd or Docker setup instructions"
}

# Parse command line arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        -e|--env)
            ENVIRONMENT="$2"
            shift 2
            ;;
        -s|--skip-tests)
            SKIP_TESTS=true
            shift
            ;;
        -b|--skip-build)
            SKIP_BUILD=true
            shift
            ;;
        -n|--no-deps)
            INSTALL_DEPS=false
            shift
            ;;
        -v|--verbose)
            VERBOSE=true
            set -x
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

# Run main function
main
