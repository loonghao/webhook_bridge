#!/bin/bash
# Development environment setup script

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${GREEN}Setting up webhook bridge development environment...${NC}"

# Check prerequisites
echo -e "${YELLOW}Checking prerequisites...${NC}"

# Check Go
if ! command -v go &> /dev/null; then
    echo -e "${RED}Go is not installed. Please install Go 1.21 or later.${NC}"
    exit 1
fi

# Check Python
if ! command -v python &> /dev/null && ! command -v python3 &> /dev/null; then
    echo -e "${RED}Python is not installed. Please install Python 3.8 or later.${NC}"
    exit 1
fi

# Check UV
if ! command -v uv &> /dev/null; then
    echo -e "${YELLOW}UV is not installed. Installing UV...${NC}"
    curl -LsSf https://astral.sh/uv/install.sh | sh
    source $HOME/.cargo/env
fi

# Check protoc
if ! command -v protoc &> /dev/null; then
    echo -e "${RED}protoc is not installed. Please install Protocol Buffers compiler.${NC}"
    echo "On macOS: brew install protobuf"
    echo "On Ubuntu: apt-get install protobuf-compiler"
    echo "On Windows: choco install protoc"
    exit 1
fi

echo -e "${GREEN}Prerequisites check passed!${NC}"

# Setup Go dependencies
echo -e "${YELLOW}Installing Go dependencies...${NC}"
go mod download
go mod tidy

# Install Go tools
echo -e "${YELLOW}Installing Go tools...${NC}"
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest

# Setup Python environment
echo -e "${YELLOW}Setting up Python environment...${NC}"
if [ ! -d "webhook-bridge" ]; then
    uv venv webhook-bridge
fi

# Activate virtual environment and install dependencies
source webhook-bridge/bin/activate || source webhook-bridge/Scripts/activate
uv pip install grpcio grpcio-tools

# Generate protobuf code
echo -e "${YELLOW}Generating protobuf code...${NC}"
make proto

# Build project
echo -e "${YELLOW}Building project...${NC}"
make build

echo -e "${GREEN}Development environment setup completed!${NC}"
echo ""
echo "To activate the Python virtual environment:"
echo "  source webhook-bridge/bin/activate    # Linux/macOS"
echo "  webhook-bridge\\Scripts\\activate      # Windows"
echo ""
echo "To run the services:"
echo "  make run-python    # Start Python executor"
echo "  make run           # Start Go server"
echo ""
echo "Or run both:"
echo "  make run-all"
