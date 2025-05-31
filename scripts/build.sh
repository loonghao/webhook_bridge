#!/bin/bash
# Build script for webhook bridge

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Build information
VERSION=${VERSION:-"2.0.0-hybrid"}
GIT_COMMIT=$(git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_DATE=$(date -u +"%Y-%m-%dT%H:%M:%SZ")
GO_VERSION=$(go version | awk '{print $3}')

echo -e "${GREEN}Building webhook bridge...${NC}"
echo "Version: $VERSION"
echo "Git Commit: $GIT_COMMIT"
echo "Build Date: $BUILD_DATE"
echo "Go Version: $GO_VERSION"

# Create bin directory
mkdir -p bin

# Build flags
LDFLAGS="-X github.com/loonghao/webhook_bridge/pkg/version.Version=$VERSION"
LDFLAGS="$LDFLAGS -X github.com/loonghao/webhook_bridge/pkg/version.GitCommit=$GIT_COMMIT"
LDFLAGS="$LDFLAGS -X github.com/loonghao/webhook_bridge/pkg/version.BuildDate=$BUILD_DATE"

# Build Go server
echo -e "${YELLOW}Building Go server...${NC}"
go build -ldflags "$LDFLAGS" -o bin/webhook-bridge-server ./cmd/server

# Make executable
chmod +x bin/webhook-bridge-server

echo -e "${GREEN}Build completed successfully!${NC}"
echo "Binary: bin/webhook-bridge-server"

# Show binary info
echo -e "${YELLOW}Binary information:${NC}"
ls -lh bin/webhook-bridge-server
file bin/webhook-bridge-server
