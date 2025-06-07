# Clean Development Environment Script
# This script cleans up all development artifacts and temporary files

Write-Host "🧹 Cleaning webhook-bridge development environment..." -ForegroundColor Cyan

# Function to safely remove files/directories
function Remove-SafelyIfExists {
    param([string]$Path, [string]$Description)
    
    if (Test-Path $Path) {
        try {
            Remove-Item -Recurse -Force $Path
            Write-Host "✅ Removed $Description" -ForegroundColor Green
        } catch {
            Write-Host "⚠️  Warning: Could not remove $Description - $($_.Exception.Message)" -ForegroundColor Yellow
        }
    }
}

# Function to remove files by pattern
function Remove-FilesByPattern {
    param([string]$Pattern, [string]$Description)
    
    $files = Get-ChildItem -Path . -Name $Pattern -ErrorAction SilentlyContinue
    if ($files) {
        foreach ($file in $files) {
            Remove-SafelyIfExists $file "$Description ($file)"
        }
    }
}

Write-Host "📦 Cleaning build artifacts..." -ForegroundColor Yellow

# Remove Go build artifacts
Remove-FilesByPattern "*.exe" "Go executable"
Remove-FilesByPattern "*.dll" "Dynamic library"
Remove-FilesByPattern "*.so" "Shared object"
Remove-FilesByPattern "*.dylib" "Dynamic library (macOS)"
Remove-FilesByPattern "*.test" "Go test binary"

# Remove build directories
Remove-SafelyIfExists "build" "Build directory"
Remove-SafelyIfExists "dist" "Distribution directory (keeping python-deps)"

# Recreate dist with python-deps if it existed
if (Test-Path "dist/python-deps") {
    # python-deps will be preserved by the above removal
} else {
    New-Item -ItemType Directory -Path "dist" -Force | Out-Null
}

Write-Host "🐍 Cleaning Python artifacts..." -ForegroundColor Yellow

# Remove Python cache
Get-ChildItem -Path . -Recurse -Name "__pycache__" -ErrorAction SilentlyContinue | ForEach-Object {
    Remove-SafelyIfExists $_ "Python cache directory"
}

Remove-FilesByPattern "*.pyc" "Python compiled file"
Remove-FilesByPattern "*.pyo" "Python optimized file"
Remove-SafelyIfExists ".pytest_cache" "Pytest cache"
Remove-SafelyIfExists ".ruff_cache" "Ruff cache"
Remove-SafelyIfExists ".nox" "Nox cache"

Write-Host "🌐 Cleaning frontend artifacts..." -ForegroundColor Yellow

# Remove Node.js artifacts
Remove-SafelyIfExists "web/dist" "Frontend build output"
Remove-SafelyIfExists "web/static" "Frontend static files"
Remove-SafelyIfExists "static" "Static files directory"
Remove-SafelyIfExists "package-lock.json" "Package lock file (root)"

Write-Host "📝 Cleaning logs and data..." -ForegroundColor Yellow

# Remove runtime directories
Remove-SafelyIfExists "logs" "Logs directory"
Remove-SafelyIfExists "data" "Data directory"

# Remove log files
Remove-FilesByPattern "*.log" "Log file"

Write-Host "⚙️  Cleaning configuration files..." -ForegroundColor Yellow

# Remove test configuration files
Remove-SafelyIfExists "config.test.yaml" "Test configuration"
Remove-SafelyIfExists "config.quick.yaml" "Quick configuration"
Remove-SafelyIfExists "config.dev.yaml" "Development configuration"
Remove-SafelyIfExists "config.local.yaml" "Local configuration"

Write-Host "🔍 Cleaning coverage and analysis files..." -ForegroundColor Yellow

# Remove coverage files
Remove-FilesByPattern "coverage.*" "Coverage file"
Remove-FilesByPattern "*.sarif" "Security analysis file"

Write-Host "🗑️  Cleaning temporary files..." -ForegroundColor Yellow

# Remove temporary files
Remove-FilesByPattern "*.tmp" "Temporary file"
Remove-FilesByPattern "*.temp" "Temporary file"
Remove-FilesByPattern "*.pid" "Process ID file"
Remove-FilesByPattern "*.bak" "Backup file"
Remove-FilesByPattern "*.backup" "Backup file"
Remove-FilesByPattern "*.orig" "Original file"

# Remove OS-specific files
Remove-FilesByPattern ".DS_Store" "macOS metadata file"
Remove-FilesByPattern "Thumbs.db" "Windows thumbnail cache"
Remove-FilesByPattern "desktop.ini" "Windows desktop configuration"

Write-Host "🧽 Running Go clean..." -ForegroundColor Yellow

# Clean Go cache
try {
    go clean -cache -testcache -modcache 2>$null
    Write-Host "✅ Go cache cleaned" -ForegroundColor Green
} catch {
    Write-Host "⚠️  Warning: Could not clean Go cache - $($_.Exception.Message)" -ForegroundColor Yellow
}

Write-Host "🎉 Development environment cleaned successfully!" -ForegroundColor Green
Write-Host ""
Write-Host "📋 Summary:" -ForegroundColor Cyan
Write-Host "   - Build artifacts removed" -ForegroundColor White
Write-Host "   - Python cache cleared" -ForegroundColor White
Write-Host "   - Frontend build outputs removed" -ForegroundColor White
Write-Host "   - Logs and data directories cleaned" -ForegroundColor White
Write-Host "   - Temporary files removed" -ForegroundColor White
Write-Host "   - Go cache cleaned" -ForegroundColor White
Write-Host ""
Write-Host "💡 To start fresh development:" -ForegroundColor Cyan
Write-Host "   uvx nox -s quick    # Quick start" -ForegroundColor White
Write-Host "   uvx nox -s dev      # Full development" -ForegroundColor White
