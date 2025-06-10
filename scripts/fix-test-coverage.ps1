# Fix Test Coverage Issues Script (PowerShell)
# This script addresses the Python and Dashboard build failures

Write-Host "🔧 Fixing webhook-bridge test coverage issues..." -ForegroundColor Green

# 1. Clean up previous builds
Write-Host "🧹 Cleaning up previous builds..." -ForegroundColor Yellow
Remove-Item -Path "dist", "build", "*.egg-info" -Recurse -Force -ErrorAction SilentlyContinue
Remove-Item -Path "web-nextjs\node_modules\.cache" -Recurse -Force -ErrorAction SilentlyContinue

# 2. Fix Python package structure
Write-Host "📦 Setting up Python package structure..." -ForegroundColor Yellow
New-Item -ItemType Directory -Path "webhook_bridge" -Force | Out-Null
if (-not (Test-Path "webhook_bridge\__init__.py")) {
    Write-Host "Creating webhook_bridge\__init__.py..." -ForegroundColor Cyan
    @'
"""Webhook Bridge Python Components"""
__version__ = "2.2.0"
'@ | Out-File -FilePath "webhook_bridge\__init__.py" -Encoding UTF8
}

# 3. Install Python dependencies
Write-Host "🐍 Installing Python dependencies..." -ForegroundColor Yellow
if (Get-Command uv -ErrorAction SilentlyContinue) {
    Write-Host "Using uv for Python package management..." -ForegroundColor Cyan
    uv pip install --upgrade pip
    uv pip install -e .
} else {
    Write-Host "Using pip for Python package management..." -ForegroundColor Cyan
    python -m pip install --upgrade pip
    pip install -e .
}

# 4. Fix Node.js dependencies
Write-Host "📱 Fixing Node.js dependencies..." -ForegroundColor Yellow
Set-Location web-nextjs
if (Test-Path "package-lock.json") {
    Write-Host "Cleaning npm cache..." -ForegroundColor Cyan
    npm cache clean --force
    Write-Host "Installing dependencies..." -ForegroundColor Cyan
    npm ci
} else {
    Write-Host "No package-lock.json found, running npm install..." -ForegroundColor Cyan
    npm install
}
Set-Location ..

# 5. Run tests
Write-Host "🧪 Running tests..." -ForegroundColor Yellow
Write-Host "Testing Python components..." -ForegroundColor Cyan
try {
    python -m pytest tests\ -v
} catch {
    Write-Host "Python tests completed with issues" -ForegroundColor Yellow
}

Write-Host "Testing Go components..." -ForegroundColor Cyan
try {
    go test .\... -v
} catch {
    Write-Host "Go tests completed with issues" -ForegroundColor Yellow
}

Write-Host "Testing Dashboard..." -ForegroundColor Cyan
Set-Location web-nextjs
try {
    npm run type-check
} catch {
    Write-Host "TypeScript check completed with issues" -ForegroundColor Yellow
}
try {
    npm run lint
} catch {
    Write-Host "Linting completed with issues" -ForegroundColor Yellow
}
Set-Location ..

Write-Host "✅ Test coverage fix script completed!" -ForegroundColor Green
Write-Host "📋 Next steps:" -ForegroundColor Cyan
Write-Host "  1. Review any remaining test failures" -ForegroundColor White
Write-Host "  2. Run: go run dev.go test-coverage" -ForegroundColor White
Write-Host "  3. Check CI pipeline status" -ForegroundColor White
