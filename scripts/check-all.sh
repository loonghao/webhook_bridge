#!/bin/bash

# Comprehensive code quality check script for webhook-bridge
# This script checks Go, Python, and Node.js/TypeScript code quality

set -e

echo "🔍 Starting comprehensive code quality checks..."
echo "================================================"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Function to print colored output
print_status() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Check if we're in the right directory
if [ ! -f "go.mod" ] || [ ! -f "pyproject.toml" ] || [ ! -d "web" ]; then
    print_error "Please run this script from the webhook-bridge root directory"
    exit 1
fi

# Track overall status
OVERALL_STATUS=0

echo ""
echo "🐹 Go Code Quality Checks"
echo "========================="

print_status "Running Go formatting check..."
if ! gofmt -l . | grep -q .; then
    print_success "Go code is properly formatted"
else
    print_warning "Go code needs formatting. Running gofmt..."
    gofmt -w .
    print_success "Go code formatted"
fi

print_status "Running Go linting..."
if go run dev.go lint; then
    print_success "Go linting passed"
else
    print_error "Go linting failed"
    OVERALL_STATUS=1
fi

print_status "Running Go tests..."
if go run dev.go test; then
    print_success "Go tests passed"
else
    print_error "Go tests failed"
    OVERALL_STATUS=1
fi

print_status "Running Go build test..."
if go run dev.go build; then
    print_success "Go build successful"
else
    print_error "Go build failed"
    OVERALL_STATUS=1
fi

echo ""
echo "🐍 Python Code Quality Checks"
echo "============================="

print_status "Checking if uv is available..."
if ! command -v uvx &> /dev/null; then
    print_error "uvx is not available. Please install uv first."
    OVERALL_STATUS=1
else
    print_status "Running Python linting..."
    if uvx nox -s lint; then
        print_success "Python linting passed"
    else
        print_error "Python linting failed"
        OVERALL_STATUS=1
    fi

    print_status "Running Python tests..."
    if uvx nox -s pytest; then
        print_success "Python tests passed"
    else
        print_error "Python tests failed"
        OVERALL_STATUS=1
    fi
fi

echo ""
echo "🌐 Node.js/TypeScript Code Quality Checks"
echo "========================================="

if [ -d "web" ]; then
    cd web
    
    print_status "Checking if npm is available..."
    if ! command -v npm &> /dev/null; then
        print_error "npm is not available. Please install Node.js first."
        OVERALL_STATUS=1
    else
        print_status "Installing dependencies..."
        if npm install; then
            print_success "Dependencies installed"
        else
            print_error "Failed to install dependencies"
            OVERALL_STATUS=1
        fi

        print_status "Running TypeScript linting..."
        if npm run lint; then
            print_success "TypeScript linting passed"
        else
            print_error "TypeScript linting failed"
            OVERALL_STATUS=1
        fi

        print_status "Running TypeScript type checking..."
        if npm run type-check; then
            print_success "TypeScript type checking passed"
        else
            print_error "TypeScript type checking failed"
            OVERALL_STATUS=1
        fi

        print_status "Running TypeScript build..."
        if npm run build; then
            print_success "TypeScript build successful"
        else
            print_error "TypeScript build failed"
            OVERALL_STATUS=1
        fi
    fi
    
    cd ..
else
    print_warning "Web directory not found, skipping Node.js checks"
fi

echo ""
echo "📋 Summary"
echo "=========="

if [ $OVERALL_STATUS -eq 0 ]; then
    print_success "All code quality checks passed! ✨"
    echo ""
    echo "Your code is ready for:"
    echo "  ✅ Commit and push"
    echo "  ✅ Pull request creation"
    echo "  ✅ Production deployment"
else
    print_error "Some checks failed. Please fix the issues above."
    echo ""
    echo "Common fixes:"
    echo "  🔧 Run 'gofmt -w .' for Go formatting"
    echo "  🔧 Run 'uvx nox -s lint' for Python issues"
    echo "  🔧 Run 'npm run lint:fix' in web/ for TypeScript issues"
fi

exit $OVERALL_STATUS
