# Build Dashboard TypeScript
# PowerShell script to build the TypeScript dashboard

param(
    [switch]$Watch,
    [switch]$Production,
    [switch]$Clean
)

$ErrorActionPreference = "Stop"

# Colors for output
$Green = "`e[32m"
$Yellow = "`e[33m"
$Red = "`e[31m"
$Reset = "`e[0m"

function Write-ColorOutput {
    param($Color, $Message)
    Write-Host "${Color}${Message}${Reset}"
}

function Test-Command {
    param($Command)
    try {
        Get-Command $Command -ErrorAction Stop | Out-Null
        return $true
    } catch {
        return $false
    }
}

# Check if we're in the right directory
if (-not (Test-Path "web/tsconfig.json")) {
    Write-ColorOutput $Red "❌ Error: Must run from project root directory"
    exit 1
}

# Check for Node.js
if (-not (Test-Command "node")) {
    Write-ColorOutput $Red "❌ Error: Node.js is not installed or not in PATH"
    Write-ColorOutput $Yellow "Please install Node.js from https://nodejs.org/"
    exit 1
}

# Check for npm
if (-not (Test-Command "npm")) {
    Write-ColorOutput $Red "❌ Error: npm is not installed or not in PATH"
    exit 1
}

# Change to web directory
Push-Location "web"

try {
    # Clean if requested
    if ($Clean) {
        Write-ColorOutput $Yellow "🧹 Cleaning build directory..."
        if (Test-Path "static/js/dist") {
            Remove-Item -Recurse -Force "static/js/dist"
        }
        Write-ColorOutput $Green "✅ Clean completed"
    }

    # Check if node_modules exists
    if (-not (Test-Path "node_modules")) {
        Write-ColorOutput $Yellow "📦 Installing dependencies..."
        npm install
        if ($LASTEXITCODE -ne 0) {
            Write-ColorOutput $Red "❌ Failed to install dependencies"
            exit 1
        }
        Write-ColorOutput $Green "✅ Dependencies installed"
    }

    # Create dist directory if it doesn't exist
    if (-not (Test-Path "static/js/dist")) {
        New-Item -ItemType Directory -Path "static/js/dist" -Force | Out-Null
    }

    if ($Watch) {
        Write-ColorOutput $Yellow "👀 Starting TypeScript watch mode..."
        Write-ColorOutput $Yellow "Press Ctrl+C to stop"
        npm run build:watch
    } elseif ($Production) {
        Write-ColorOutput $Yellow "🏗️ Building for production..."
        npm run build:prod
        if ($LASTEXITCODE -eq 0) {
            Write-ColorOutput $Green "✅ Production build completed"
            Write-ColorOutput $Green "📁 Output: web/static/js/dist/"
        } else {
            Write-ColorOutput $Red "❌ Production build failed"
            exit 1
        }
    } else {
        Write-ColorOutput $Yellow "🏗️ Building TypeScript dashboard..."
        npm run build
        if ($LASTEXITCODE -eq 0) {
            Write-ColorOutput $Green "✅ Build completed successfully"
            Write-ColorOutput $Green "📁 Output: web/static/js/dist/"
        } else {
            Write-ColorOutput $Red "❌ Build failed"
            exit 1
        }
    }

} finally {
    Pop-Location
}

Write-ColorOutput $Green "🎉 Dashboard build process completed!"
