# Webhook Bridge Makefile
# Enhanced build system for hybrid Go/Python architecture

# Variables
PROJECT_NAME := webhook-bridge
VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
BUILD_TIME := $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
GO_VERSION := $(shell go version | awk '{print $$3}')

# Directories
BUILD_DIR := build
DIST_DIR := dist
SCRIPTS_DIR := scripts

# Go build flags
LDFLAGS := -ldflags "-X main.version=$(VERSION) -X main.buildTime=$(BUILD_TIME) -X main.goVersion=$(GO_VERSION)"

# Colors for output
BLUE := \033[34m
GREEN := \033[32m
YELLOW := \033[33m
RED := \033[31m
NC := \033[0m

.PHONY: help build run test clean proto install-deps dev-setup deploy start stop status

# Default target
help:
	@echo "$(BLUE)Webhook Bridge - Enhanced Build System$(NC)"
	@echo ""
	@echo "$(GREEN)Development:$(NC)"
	@echo "  dev-setup    - Setup development environment"
	@echo "  deps         - Install all dependencies"
	@echo "  build        - Build Go binaries"
	@echo "  start        - Start services in development mode"
	@echo "  stop         - Stop running services"
	@echo "  restart      - Restart services"
	@echo "  status       - Show service status"
	@echo ""
	@echo "$(GREEN)Testing:$(NC)"
	@echo "  test         - Run all tests"
	@echo "  test-go      - Run Go tests"
	@echo "  test-python  - Run Python tests"
	@echo "  test-integration - Run integration tests"
	@echo "  test-coverage - Run tests with coverage"
	@echo ""
	@echo "$(GREEN)Code Quality:$(NC)"
	@echo "  lint         - Run linters"
	@echo "  format       - Format code"
	@echo "  proto        - Generate gRPC code"
	@echo ""
	@echo "$(GREEN)Deployment:$(NC)"
	@echo "  deploy-dev   - Deploy for development"
	@echo "  deploy-prod  - Deploy for production"
	@echo "  docker-build - Build Docker image"
	@echo "  docker-run   - Run Docker container"
	@echo "  release      - Create release package"
	@echo ""
	@echo "$(GREEN)Utilities:$(NC)"
	@echo "  clean        - Clean build artifacts"
	@echo "  version      - Show version information"

# Aliases for convenience
deps: install-deps

# Install dependencies
install-deps: install-go-deps install-python-deps

install-go-deps:
	@echo "Installing Go dependencies..."
	go mod download
	go mod tidy

install-python-deps:
	@echo "Installing Python dependencies..."
	@if command -v uv >/dev/null 2>&1; then \
		echo "Using uv for Python dependencies..."; \
		uv sync; \
	else \
		echo "Using pip for Python dependencies..."; \
		pip install -r requirements.txt; \
		pip install -r requirements-dev.txt; \
	fi

# Development setup
dev-setup: install-deps proto
	@echo "Development environment setup complete!"

# Generate gRPC code
proto:
	@echo "Generating gRPC code..."
	@mkdir -p api/proto
	protoc --go_out=. --go_opt=paths=source_relative \
		--go-grpc_out=. --go-grpc_opt=paths=source_relative \
		api/proto/webhook.proto
	@echo "Generating Python gRPC code..."
	.venv/Scripts/python.exe -m grpc_tools.protoc -I. --python_out=. --grpc_python_out=. api/proto/webhook.proto

# Build Go binaries
build:
	@echo "$(BLUE)Building Go binaries...$(NC)"
	mkdir -p $(BUILD_DIR)
	go build $(LDFLAGS) -o $(BUILD_DIR)/webhook-bridge-server ./cmd/server
	go build $(LDFLAGS) -o $(BUILD_DIR)/python-manager ./cmd/python-manager
	@echo "$(GREEN)Build completed$(NC)"

# Run Go server
run: build
	@echo "Starting webhook bridge server..."
	./bin/webhook-bridge-server

# Run Python executor service
run-python:
	@echo "Starting Python executor service..."
	.venv/Scripts/python.exe python_executor/main.py

# Run both services
run-all:
	@echo "Starting both services..."
	@make run-python &
	@sleep 2
	@make run

# Service management using scripts
start: build
	@echo "$(BLUE)Starting webhook bridge in development mode...$(NC)"
	@chmod +x $(SCRIPTS_DIR)/start.sh
	$(SCRIPTS_DIR)/start.sh -e dev

start-daemon: build
	@echo "$(BLUE)Starting webhook bridge as daemon...$(NC)"
	@chmod +x $(SCRIPTS_DIR)/start.sh
	$(SCRIPTS_DIR)/start.sh -e dev -d

stop:
	@echo "$(BLUE)Stopping webhook bridge...$(NC)"
	@chmod +x $(SCRIPTS_DIR)/start.sh
	$(SCRIPTS_DIR)/start.sh --stop

restart:
	@echo "$(BLUE)Restarting webhook bridge...$(NC)"
	@chmod +x $(SCRIPTS_DIR)/start.sh
	$(SCRIPTS_DIR)/start.sh --restart

status:
	@chmod +x $(SCRIPTS_DIR)/start.sh
	$(SCRIPTS_DIR)/start.sh --status

# Tests
test: test-go test-python

test-go:
	@echo "Running Go tests..."
	go test -v ./...

test-python:
	@echo "Running Python tests..."
	.venv/Scripts/python.exe -m pytest tests/

test-integration:
	@echo "$(BLUE)Running integration tests...$(NC)"
	.venv/Scripts/python.exe test_go_python_integration.py

test-coverage:
	@echo "$(BLUE)Running tests with coverage...$(NC)"
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "$(GREEN)Coverage report generated: coverage.html$(NC)"

# Clean
clean:
	@echo "$(BLUE)Cleaning build artifacts...$(NC)"
	rm -rf bin/
	rm -rf $(DIST_DIR)/
	rm -rf $(BUILD_DIR)/
	rm -rf *.egg-info/
	rm -f *.log *.pid coverage.out coverage.html
	go clean
	@echo "$(GREEN)Clean completed$(NC)"

# Deployment
deploy-dev:
	@echo "$(BLUE)Deploying for development...$(NC)"
	@chmod +x $(SCRIPTS_DIR)/deploy.sh
	$(SCRIPTS_DIR)/deploy.sh -e dev

deploy-prod:
	@echo "$(BLUE)Deploying for production...$(NC)"
	@chmod +x $(SCRIPTS_DIR)/deploy.sh
	$(SCRIPTS_DIR)/deploy.sh -e prod

release: clean build test
	@echo "$(BLUE)Creating release package...$(NC)"
	@chmod +x $(SCRIPTS_DIR)/deploy.sh
	$(SCRIPTS_DIR)/deploy.sh -e prod -s
	@echo "$(GREEN)Release package created$(NC)"

# Docker
docker-build:
	@echo "Building Docker image..."
	docker build -t webhook-bridge:latest .

docker-run:
	@echo "Running Docker container..."
	docker run -p 8000:8000 -p 50051:50051 webhook-bridge:latest

# Development helpers
fmt:
	@echo "Formatting Go code..."
	go fmt ./...
	@echo "Formatting Python code..."
	@if command -v uv >/dev/null 2>&1; then \
		uv run ruff format .; \
	else \
		ruff format .; \
	fi

lint:
	@echo "Linting Go code..."
	golangci-lint run
	@echo "Linting Python code..."
	@if command -v uv >/dev/null 2>&1; then \
		uv run ruff check .; \
	else \
		ruff check .; \
	fi

# Version information
version:
	@echo "Project: $(PROJECT_NAME)"
	@echo "Version: $(VERSION)"
	@echo "Build Time: $(BUILD_TIME)"
	@echo "Go Version: $(GO_VERSION)"

# Quick development workflow
quick: clean deps build test
	@echo "$(GREEN)Quick development workflow completed$(NC)"
