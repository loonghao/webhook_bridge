# Fix GoReleaser Build Issues Script (PowerShell)

Write-Host "🔧 Fixing GoReleaser build issues..." -ForegroundColor Green

# 1. Clean up previous builds
Write-Host "🧹 Cleaning up previous builds..." -ForegroundColor Yellow
Remove-Item -Path "dist" -Recurse -Force -ErrorAction SilentlyContinue
go clean -cache

# 2. Ensure all required files exist
Write-Host "📁 Checking required files..." -ForegroundColor Yellow

# Check if docker-entrypoint.sh exists
if (-not (Test-Path "docker-entrypoint.sh")) {
    Write-Host "📝 Creating docker-entrypoint.sh..." -ForegroundColor Cyan
    @'
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
'@ | Out-File -FilePath "docker-entrypoint.sh" -Encoding UTF8
}

# Check if requirements.txt exists
if (-not (Test-Path "requirements.txt")) {
    Write-Host "📝 Creating requirements.txt..." -ForegroundColor Cyan
    @'
# Core dependencies for webhook-bridge Python executor
grpcio>=1.50.0
grpcio-tools>=1.50.0
pyyaml>=6.0
fastapi>=0.100.0
uvicorn>=0.20.0
httpx>=0.24.0
click>=8.0.0
'@ | Out-File -FilePath "requirements.txt" -Encoding UTF8
}

# 3. Update go.mod and download dependencies
Write-Host "📦 Updating Go dependencies..." -ForegroundColor Yellow
go mod tidy
go mod download

# 4. Generate protobuf files if needed
Write-Host "🔧 Generating protobuf files..." -ForegroundColor Yellow
if (-not (Test-Path "api\proto\webhook.pb.go")) {
    try {
        go run dev.go proto
    } catch {
        Write-Host "⚠️ Protobuf generation failed, continuing..." -ForegroundColor Yellow
    }
}

# 5. Build dashboard
Write-Host "🏗️ Building dashboard..." -ForegroundColor Yellow
try {
    go run dev.go dashboard build --production
} catch {
    Write-Host "⚠️ Dashboard build failed, creating minimal structure..." -ForegroundColor Yellow
    New-Item -ItemType Directory -Path "web-nextjs\dist" -Force | Out-Null
    '<!DOCTYPE html><html><head><title>Webhook Bridge</title></head><body><h1>Dashboard</h1></body></html>' | Out-File -FilePath "web-nextjs\dist\index.html" -Encoding UTF8
}

# 6. Test build locally first
Write-Host "🧪 Testing local build..." -ForegroundColor Yellow
go build -o test-webhook-bridge.exe .\cmd\webhook-bridge
if (Test-Path "test-webhook-bridge.exe") {
    Write-Host "✅ Local build successful" -ForegroundColor Green
    Remove-Item "test-webhook-bridge.exe"
} else {
    Write-Host "❌ Local build failed" -ForegroundColor Red
    exit 1
}

# 7. Check Git status
Write-Host "📋 Checking Git status..." -ForegroundColor Yellow
$gitStatus = git status --porcelain
if ($gitStatus) {
    Write-Host "⚠️ Working directory has uncommitted changes:" -ForegroundColor Yellow
    git status --porcelain
    Write-Host "💡 Consider committing changes before release" -ForegroundColor Cyan
} else {
    Write-Host "✅ Working directory is clean" -ForegroundColor Green
}

# 8. Check if we're on a tag
Write-Host "🏷️ Checking Git tags..." -ForegroundColor Yellow
try {
    $tag = git describe --exact-match --tags HEAD 2>$null
    Write-Host "✅ On tag: $tag" -ForegroundColor Green
} catch {
    Write-Host "⚠️ Not on a tag. For release, create a tag first:" -ForegroundColor Yellow
    Write-Host "   git tag v1.0.0" -ForegroundColor White
    Write-Host "   git push origin v1.0.0" -ForegroundColor White
}

Write-Host "✅ GoReleaser fix script completed!" -ForegroundColor Green
Write-Host "📋 Next steps:" -ForegroundColor Cyan
Write-Host "  1. For snapshot build: go run dev.go release-snapshot" -ForegroundColor White
Write-Host "  2. For dry run: goreleaser release --skip=publish --clean" -ForegroundColor White
Write-Host "  3. For full release: goreleaser release --clean" -ForegroundColor White
