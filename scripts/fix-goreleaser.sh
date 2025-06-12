#!/bin/bash

# Fix GoReleaser Build Issues Script

set -e

echo "🔧 Fixing GoReleaser build issues..."

# 1. Clean up previous builds
echo "🧹 Cleaning up previous builds..."
rm -rf dist/ || true
go clean -cache || true

# 2. Ensure all required files exist
echo "📁 Checking required files..."

# Check if docker-entrypoint.sh exists
if [ ! -f "docker-entrypoint.sh" ]; then
    echo "📝 Creating docker-entrypoint.sh..."
    cat > docker-entrypoint.sh << 'EOF'
#!/bin/bash
set -e

# Docker entrypoint script for webhook-bridge

echo "🚀 Starting webhook-bridge container..."

# Set default values if not provided
export WEBHOOK_BRIDGE_CONFIG_PATH="${WEBHOOK_BRIDGE_CONFIG_PATH:-/app/config}"
export WEBHOOK_BRIDGE_PLUGINS_PATH="${WEBHOOK_BRIDGE_PLUGINS_PATH:-/app/plugins:/app/example_plugins}"
export WEBHOOK_BRIDGE_LOG_PATH="${WEBHOOK_BRIDGE_LOG_PATH:-/app/logs}"
export WEBHOOK_BRIDGE_DATA_PATH="${WEBHOOK_BRIDGE_DATA_PATH:-/app/data}"
export WEBHOOK_BRIDGE_WEB_PATH="${WEBHOOK_BRIDGE_WEB_PATH:-/app/web-nextjs/dist}"
export WEBHOOK_BRIDGE_PYTHON_PATH="${WEBHOOK_BRIDGE_PYTHON_PATH:-/app/python_executor}"

# Create directories if they don't exist
mkdir -p "$WEBHOOK_BRIDGE_CONFIG_PATH"
mkdir -p "$WEBHOOK_BRIDGE_LOG_PATH"
mkdir -p "$WEBHOOK_BRIDGE_DATA_PATH"

# Copy default config if none exists
if [ ! -f "$WEBHOOK_BRIDGE_CONFIG_PATH/config.yaml" ] && [ -f "/app/config.yaml" ]; then
    echo "📋 Copying default configuration..."
    cp /app/config.yaml "$WEBHOOK_BRIDGE_CONFIG_PATH/"
fi

# Execute the command
echo "🎯 Executing: $@"
exec "$@"
EOF
    chmod +x docker-entrypoint.sh
fi

# Check if requirements.txt exists
if [ ! -f "requirements.txt" ]; then
    echo "📝 Creating requirements.txt..."
    cat > requirements.txt << 'EOF'
# Core dependencies for webhook-bridge Python executor
grpcio>=1.50.0
grpcio-tools>=1.50.0
pyyaml>=6.0
fastapi>=0.100.0
uvicorn>=0.20.0
httpx>=0.24.0
click>=8.0.0
EOF
fi

# 3. Update go.mod and download dependencies
echo "📦 Updating Go dependencies..."
go mod tidy
go mod download

# 4. Generate protobuf files if needed
echo "🔧 Generating protobuf files..."
if [ ! -f "api/proto/webhook.pb.go" ]; then
    go run dev.go proto || echo "⚠️ Protobuf generation failed, continuing..."
fi

# 5. Build dashboard
echo "🏗️ Building dashboard..."
go run dev.go dashboard build --production || {
    echo "⚠️ Dashboard build failed, creating minimal structure..."
    mkdir -p web-nextjs/dist
    echo '<!DOCTYPE html><html><head><title>Webhook Bridge</title></head><body><h1>Dashboard</h1></body></html>' > web-nextjs/dist/index.html
}

# 6. Test build locally first
echo "🧪 Testing local build..."
go build -o test-webhook-bridge ./cmd/webhook-bridge
if [ -f "test-webhook-bridge" ]; then
    echo "✅ Local build successful"
    rm test-webhook-bridge
else
    echo "❌ Local build failed"
    exit 1
fi

# 7. Check Git status
echo "📋 Checking Git status..."
if ! git status --porcelain | grep -q .; then
    echo "✅ Working directory is clean"
else
    echo "⚠️ Working directory has uncommitted changes:"
    git status --porcelain
    echo "💡 Consider committing changes before release"
fi

# 8. Check if we're on a tag
echo "🏷️ Checking Git tags..."
if git describe --exact-match --tags HEAD >/dev/null 2>&1; then
    TAG=$(git describe --exact-match --tags HEAD)
    echo "✅ On tag: $TAG"
else
    echo "⚠️ Not on a tag. For release, create a tag first:"
    echo "   git tag v1.0.0"
    echo "   git push origin v1.0.0"
fi

echo "✅ GoReleaser fix script completed!"
echo "📋 Next steps:"
echo "  1. For snapshot build: go run dev.go release-snapshot"
echo "  2. For dry run: goreleaser release --skip=publish --clean"
echo "  3. For full release: goreleaser release --clean"
