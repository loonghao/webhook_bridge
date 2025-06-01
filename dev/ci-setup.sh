#!/bin/bash

# CI Setup Script for webhook-bridge
# This script ensures consistent Go environment setup across all CI jobs
# and resolves common Go version mismatch issues

set -euo pipefail

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Logging functions
log_info() {
    echo -e "${BLUE}ℹ️  $1${NC}"
}

log_success() {
    echo -e "${GREEN}✅ $1${NC}"
}

log_warning() {
    echo -e "${YELLOW}⚠️  $1${NC}"
}

log_error() {
    echo -e "${RED}❌ $1${NC}"
}

# Check if we're in CI environment
is_ci() {
    [[ "${CI:-false}" == "true" ]] || [[ "${GITHUB_ACTIONS:-false}" == "true" ]]
}

# Load environment variables from .github/env
load_env_vars() {
    if [[ -f ".github/env" ]]; then
        log_info "Loading environment variables from .github/env"
        # Export variables from .github/env (skip comments and empty lines)
        set -a
        while IFS= read -r line; do
            # Skip empty lines and comments
            if [[ -n "$line" && ! "$line" =~ ^[[:space:]]*# ]]; then
                export "$line"
            fi
        done < .github/env
        set +a
        log_success "Environment variables loaded"
    else
        log_warning ".github/env file not found, using defaults"
        export GO_VERSION=${GO_VERSION:-"1.23"}
        export GOLANGCI_LINT_VERSION=${GOLANGCI_LINT_VERSION:-"v1.64.6"}
    fi
}

# Verify Go installation and version
verify_go() {
    log_info "Verifying Go installation..."
    
    if ! command -v go &> /dev/null; then
        log_error "Go is not installed or not in PATH"
        exit 1
    fi
    
    local go_version=$(go version | awk '{print $3}' | sed 's/go//')
    log_info "Detected Go version: $go_version"
    log_info "Expected Go version: $GO_VERSION"
    
    # Check if versions match (allowing for patch version differences)
    local major_minor=$(echo "$go_version" | cut -d. -f1,2)
    local expected_major_minor=$(echo "$GO_VERSION" | cut -d. -f1,2)
    
    if [[ "$major_minor" != "$expected_major_minor" ]]; then
        log_warning "Go version mismatch detected"
        log_warning "This may cause 'compile version does not match go tool version' errors"
        return 1
    fi
    
    log_success "Go version verification passed"
    return 0
}

# Clean Go caches to resolve version mismatch issues
clean_go_caches() {
    log_info "Cleaning Go caches to resolve potential version conflicts..."
    
    # Clean build cache
    if go clean -cache; then
        log_success "Build cache cleaned"
    else
        log_warning "Failed to clean build cache"
    fi
    
    # Clean module cache (only in CI to avoid affecting local development)
    if is_ci; then
        if go clean -modcache; then
            log_success "Module cache cleaned"
        else
            log_warning "Failed to clean module cache"
        fi
    fi
    
    # Clean test cache
    if go clean -testcache; then
        log_success "Test cache cleaned"
    else
        log_warning "Failed to clean test cache"
    fi
}

# Install Go tools with version consistency
install_go_tools() {
    log_info "Installing Go protobuf tools..."
    
    local tools=(
        "google.golang.org/protobuf/cmd/protoc-gen-go@latest"
        "google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest"
    )
    
    for tool in "${tools[@]}"; do
        log_info "Installing $tool"
        if go install "$tool"; then
            log_success "Installed $tool"
        else
            log_error "Failed to install $tool"
            exit 1
        fi
    done
}

# Verify protobuf compiler installation
verify_protoc() {
    log_info "Verifying protobuf compiler..."
    
    if ! command -v protoc &> /dev/null; then
        log_error "protoc is not installed or not in PATH"
        log_error "Please ensure protobuf compiler is installed"
        exit 1
    fi
    
    local protoc_version=$(protoc --version)
    log_success "protoc found: $protoc_version"
}

# Download and verify Go modules
setup_go_modules() {
    log_info "Setting up Go modules..."
    
    # Download modules
    if go mod download; then
        log_success "Go modules downloaded"
    else
        log_error "Failed to download Go modules"
        exit 1
    fi
    
    # Verify modules
    if go mod verify; then
        log_success "Go modules verified"
    else
        log_error "Go module verification failed"
        exit 1
    fi
}

# Main setup function
main() {
    log_info "Starting CI setup for webhook-bridge..."
    
    # Load environment variables
    load_env_vars
    
    # Verify Go installation
    if ! verify_go; then
        log_warning "Go version mismatch detected, cleaning caches..."
        clean_go_caches
    fi
    
    # Clean caches if in CI environment
    if is_ci; then
        clean_go_caches
    fi
    
    # Verify protobuf compiler
    verify_protoc
    
    # Install Go tools
    install_go_tools
    
    # Setup Go modules
    setup_go_modules
    
    log_success "CI setup completed successfully!"
    
    # Display environment info
    log_info "Environment Information:"
    echo "  Go version: $(go version)"
    echo "  GOPATH: $(go env GOPATH)"
    echo "  GOCACHE: $(go env GOCACHE)"
    echo "  GOMODCACHE: $(go env GOMODCACHE)"
}

# Run main function
main "$@"
