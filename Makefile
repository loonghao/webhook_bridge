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
	@echo "$(GREEN)Building:$(NC)"
	@echo "  build        - Build Go binaries"
	@echo "  build-race   - Build with race detection"
	@echo "  build-all    - Build for all platforms"
	@echo "  build-linux  - Build for Linux"
	@echo "  build-windows - Build for Windows"
	@echo "  build-darwin - Build for macOS"
	@echo ""
	@echo "$(GREEN)Testing:$(NC)"
	@echo "  test         - Run all tests"
	@echo "  test-go      - Run Go tests"
	@echo "  test-go-race - Run Go tests with race detection"
	@echo "  test-python  - Run Python tests"
	@echo "  test-integration - Run integration tests"
	@echo "  test-coverage - Run tests with coverage"
	@echo "  verify       - Run all verification checks"
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
	@echo "  release-snapshot - Create snapshot release"
	@echo "  release-goreleaser - Release with GoReleaser"
	@echo ""
	@echo "$(GREEN)Utilities:$(NC)"
	@echo "  clean        - Clean build artifacts"
	@echo "  version      - Show version information"

# Aliases for convenience
deps: install-deps

# Install dependencies
install-deps: install-go-deps install-python-deps install-tools

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

install-tools:
	@echo "Installing development tools..."
	@echo "Installing protobuf tools..."
	go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
	go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
	@echo "Installing linting tools..."
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	@echo "Installing release tools..."
	go install github.com/goreleaser/goreleaser@latest
	@echo "Installing task runner..."
	go install github.com/go-task/task/v3/cmd/task@latest

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
build: proto
	@echo "$(BLUE)Building Go binaries...$(NC)"
	mkdir -p $(BUILD_DIR)
	go build $(LDFLAGS) -o $(BUILD_DIR)/webhook-bridge-server ./cmd/server
	go build $(LDFLAGS) -o $(BUILD_DIR)/python-manager ./cmd/python-manager
	@echo "$(GREEN)Build completed$(NC)"

# Build with race detection (for development)
build-race: proto
	@echo "$(BLUE)Building with race detection...$(NC)"
	mkdir -p $(BUILD_DIR)
	go build -race $(LDFLAGS) -o $(BUILD_DIR)/webhook-bridge-server-race ./cmd/server
	go build -race $(LDFLAGS) -o $(BUILD_DIR)/python-manager-race ./cmd/python-manager
	@echo "$(GREEN)Race detection build completed$(NC)"

# Cross-platform builds (Kubernetes style)
build-linux: proto
	@echo "$(BLUE)Building for Linux...$(NC)"
	mkdir -p $(BUILD_DIR)
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o $(BUILD_DIR)/webhook-bridge-server-linux-amd64 ./cmd/server
	CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build $(LDFLAGS) -o $(BUILD_DIR)/webhook-bridge-server-linux-arm64 ./cmd/server

build-windows: proto
	@echo "$(BLUE)Building for Windows...$(NC)"
	mkdir -p $(BUILD_DIR)
	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build $(LDFLAGS) -o $(BUILD_DIR)/webhook-bridge-server-windows-amd64.exe ./cmd/server

build-darwin: proto
	@echo "$(BLUE)Building for macOS...$(NC)"
	mkdir -p $(BUILD_DIR)
	CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build $(LDFLAGS) -o $(BUILD_DIR)/webhook-bridge-server-darwin-amd64 ./cmd/server
	CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 go build $(LDFLAGS) -o $(BUILD_DIR)/webhook-bridge-server-darwin-arm64 ./cmd/server

build-all: build-linux build-windows build-darwin
	@echo "$(GREEN)All platform builds completed$(NC)"

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

test-go-race:
	@echo "Running Go tests with race detection..."
	go test -race -v ./...

test-go-short:
	@echo "Running Go short tests..."
	go test -short -v ./...

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

test-coverage-ci:
	@echo "$(BLUE)Running tests with coverage for CI...$(NC)"
	go test -race -coverprofile=coverage.out -covermode=atomic ./...

# Verification targets (Kubernetes style)
verify: verify-gofmt verify-govet verify-golint verify-deps

verify-gofmt:
	@echo "Verifying gofmt..."
	@if [ -n "$$(gofmt -l . | grep -v vendor)" ]; then \
		echo "Files not formatted with gofmt:"; \
		gofmt -l . | grep -v vendor; \
		exit 1; \
	fi

verify-govet:
	@echo "Verifying go vet..."
	go vet ./...

verify-golint:
	@echo "Verifying golangci-lint..."
	golangci-lint run

verify-deps:
	@echo "Verifying dependencies..."
	go mod verify
	go mod tidy
	@if [ -n "$$(git status --porcelain go.mod go.sum)" ]; then \
		echo "go.mod or go.sum is not up to date"; \
		git diff go.mod go.sum; \
		exit 1; \
	fi

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

release: clean verify test
	@echo "$(BLUE)Creating release package...$(NC)"
	@chmod +x $(SCRIPTS_DIR)/deploy.sh
	$(SCRIPTS_DIR)/deploy.sh -e prod -s
	@echo "$(GREEN)Release package created$(NC)"

# GoReleaser targets (Lazydocker style)
release-snapshot:
	@echo "$(BLUE)Creating snapshot release...$(NC)"
	goreleaser release --snapshot --clean

release-dry-run:
	@echo "$(BLUE)Dry run release...$(NC)"
	goreleaser release --skip=publish --clean

release-goreleaser:
	@echo "$(BLUE)Creating release with GoReleaser...$(NC)"
	goreleaser release --clean

# Check if we can release
check-release:
	@echo "Checking release readiness..."
	@if [ -z "$$(git tag --points-at HEAD)" ]; then \
		echo "❌ No tag found at HEAD. Please create a tag first."; \
		exit 1; \
	fi
	@echo "✅ Release check passed"

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
