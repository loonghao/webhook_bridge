# Webhook Bridge Deployment Script for Windows
# This script automates the deployment process for webhook bridge on Windows

param(
    [string]$Environment = "dev",
    [switch]$SkipTests,
    [switch]$SkipBuild,
    [switch]$NoDeps,
    [switch]$Verbose,
    [switch]$Help
)

# Colors for output
$Red = "Red"
$Green = "Green"
$Yellow = "Yellow"
$Blue = "Blue"

# Configuration
$ScriptDir = Split-Path -Parent $MyInvocation.MyCommand.Path
$ProjectRoot = Split-Path -Parent $ScriptDir
$BuildDir = Join-Path $ProjectRoot "build"
$DistDir = Join-Path $ProjectRoot "dist"

# Functions
function Write-Info {
    param([string]$Message)
    Write-Host "[INFO] $Message" -ForegroundColor $Blue
}

function Write-Success {
    param([string]$Message)
    Write-Host "[SUCCESS] $Message" -ForegroundColor $Green
}

function Write-Warning {
    param([string]$Message)
    Write-Host "[WARNING] $Message" -ForegroundColor $Yellow
}

function Write-Error {
    param([string]$Message)
    Write-Host "[ERROR] $Message" -ForegroundColor $Red
}

function Show-Help {
    @"
Webhook Bridge Deployment Script for Windows

Usage: .\deploy.ps1 [OPTIONS]

Options:
    -Environment ENVIRONMENT    Set deployment environment (dev, prod) [default: dev]
    -SkipTests                 Skip running tests
    -SkipBuild                 Skip building binaries
    -NoDeps                    Skip dependency installation
    -Verbose                   Enable verbose output
    -Help                      Show this help message

Examples:
    .\deploy.ps1                         # Deploy for development
    .\deploy.ps1 -Environment prod       # Deploy for production
    .\deploy.ps1 -SkipTests -SkipBuild   # Quick deploy (skip tests and build)
    .\deploy.ps1 -Environment prod -Verbose  # Production deploy with verbose output

"@
}

function Test-Dependencies {
    Write-Info "Checking dependencies..."
    
    # Check Go
    try {
        $goVersion = & go version 2>$null
        Write-Info "Found Go: $goVersion"
    }
    catch {
        Write-Error "Go is not installed. Please install Go 1.21 or later."
        exit 1
    }
    
    # Check Python
    try {
        $pythonVersion = & python --version 2>$null
        Write-Info "Found Python: $pythonVersion"
    }
    catch {
        Write-Error "Python is not installed. Please install Python 3.8 or later."
        exit 1
    }
    
    # Check UV (optional but recommended)
    try {
        $uvVersion = & uv --version 2>$null
        Write-Info "Found UV: $uvVersion"
    }
    catch {
        Write-Warning "UV is not installed. Consider installing UV for better Python environment management."
    }
    
    Write-Success "Dependencies check passed"
}

function Install-PythonDeps {
    if (-not $NoDeps) {
        Write-Info "Installing Python dependencies..."
        Set-Location $ProjectRoot
        
        try {
            $uvVersion = & uv --version 2>$null
            Write-Info "Using UV for Python environment management"
            & uv venv .venv
            & .venv\Scripts\Activate.ps1
            & uv pip install -r requirements.txt
        }
        catch {
            Write-Info "Using pip for Python environment management"
            & python -m venv .venv
            & .venv\Scripts\Activate.ps1
            & pip install -r requirements.txt
        }
        
        Write-Success "Python dependencies installed"
    }
    else {
        Write-Info "Skipping Python dependency installation"
    }
}

function Invoke-Tests {
    if (-not $SkipTests) {
        Write-Info "Running tests..."
        Set-Location $ProjectRoot
        
        # Activate Python virtual environment
        & .venv\Scripts\Activate.ps1
        
        # Run Go tests
        Write-Info "Running Go tests..."
        & go test ./... -v
        
        # Run Python tests
        Write-Info "Running Python tests..."
        try {
            & python -m pytest tests/ -v
        }
        catch {
            Write-Warning "Python tests not found or failed"
        }
        
        # Run integration tests
        Write-Info "Running integration tests..."
        try {
            & python test_go_python_integration.py
        }
        catch {
            Write-Warning "Integration tests failed"
        }
        
        Write-Success "Tests completed"
    }
    else {
        Write-Info "Skipping tests"
    }
}

function Build-Binaries {
    if (-not $SkipBuild) {
        Write-Info "Building binaries..."
        Set-Location $ProjectRoot
        
        # Create build directory
        if (-not (Test-Path $BuildDir)) {
            New-Item -ItemType Directory -Path $BuildDir | Out-Null
        }
        if (-not (Test-Path $DistDir)) {
            New-Item -ItemType Directory -Path $DistDir | Out-Null
        }
        
        # Build Go server
        Write-Info "Building Go server..."
        & go build -o "$BuildDir\webhook-bridge-server.exe" .\cmd\server
        
        # Build Python manager
        Write-Info "Building Python manager..."
        & go build -o "$BuildDir\python-manager.exe" .\cmd\python-manager
        
        Write-Success "Binaries built successfully"
    }
    else {
        Write-Info "Skipping build"
    }
}

function Set-Configuration {
    Write-Info "Setting up configuration for environment: $Environment"
    Set-Location $ProjectRoot
    
    switch ($Environment) {
        "dev" {
            Copy-Item "config.dev.yaml" "config.yaml" -Force
            Write-Info "Using development configuration"
        }
        "prod" {
            Copy-Item "config.prod.yaml" "config.yaml" -Force
            Write-Info "Using production configuration"
        }
        default {
            if (-not (Test-Path "config.yaml")) {
                Copy-Item "config.example.yaml" "config.yaml" -Force
                Write-Warning "Unknown environment. Using example configuration."
                Write-Warning "Please review and modify config.yaml as needed."
            }
        }
    }
}

function New-WindowsService {
    if ($Environment -eq "prod") {
        Write-Info "Creating Windows service installer..."
        
        $serviceScript = @"
# Windows Service Installation Script
# Run as Administrator

`$serviceName = "WebhookBridge"
`$serviceDisplayName = "Webhook Bridge Service"
`$serviceDescription = "High-performance webhook processing service"
`$servicePath = "C:\Program Files\WebhookBridge\webhook-bridge-server.exe"

# Stop and remove existing service if it exists
if (Get-Service `$serviceName -ErrorAction SilentlyContinue) {
    Write-Host "Stopping existing service..."
    Stop-Service `$serviceName -Force
    Write-Host "Removing existing service..."
    sc.exe delete `$serviceName
}

# Create new service
Write-Host "Creating Windows service..."
New-Service -Name `$serviceName -BinaryPathName `$servicePath -DisplayName `$serviceDisplayName -Description `$serviceDescription -StartupType Automatic

# Start service
Write-Host "Starting service..."
Start-Service `$serviceName

Write-Host "Service installed and started successfully!"
Write-Host "Service Name: `$serviceName"
Write-Host "Display Name: `$serviceDisplayName"
"@
        
        $serviceScript | Out-File -FilePath "$DistDir\install-service.ps1" -Encoding UTF8
        
        Write-Success "Windows service installer created at $DistDir\install-service.ps1"
        Write-Info "To install: Run PowerShell as Administrator and execute the script"
    }
}

function New-DockerFiles {
    Write-Info "Creating Docker files..."
    
    # Dockerfile (same as Linux version)
    $dockerfile = @'
# Multi-stage build for webhook bridge
FROM golang:1.21-alpine AS go-builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -o webhook-bridge-server ./cmd/server
RUN go build -o python-manager ./cmd/python-manager

FROM python:3.11-slim

# Install system dependencies
RUN apt-get update && apt-get install -y \
    curl \
    && rm -rf /var/lib/apt/lists/*

# Install UV
RUN pip install uv

WORKDIR /app

# Copy Python requirements and install dependencies
COPY requirements.txt .
RUN uv venv .venv && \
    . .venv/bin/activate && \
    uv pip install -r requirements.txt

# Copy Go binaries
COPY --from=go-builder /app/webhook-bridge-server .
COPY --from=go-builder /app/python-manager .

# Copy Python code and configs
COPY python_executor/ ./python_executor/
COPY api/ ./api/
COPY example_plugins/ ./example_plugins/
COPY config.prod.yaml ./config.yaml

# Create non-root user
RUN useradd -m -u 1000 webhook && \
    chown -R webhook:webhook /app

USER webhook

EXPOSE 8080 50051

CMD ["./webhook-bridge-server"]
'@
    
    $dockerfile | Out-File -FilePath "$ProjectRoot\Dockerfile" -Encoding UTF8
    
    Write-Success "Docker files created"
}

function New-Release {
    Write-Info "Packaging release..."
    Set-Location $ProjectRoot
    
    # Create release directory
    $timestamp = Get-Date -Format "yyyyMMdd-HHmmss"
    $releaseDir = Join-Path $DistDir "webhook-bridge-$timestamp"
    New-Item -ItemType Directory -Path $releaseDir | Out-Null
    
    # Copy binaries
    Copy-Item "$BuildDir\*" $releaseDir -Recurse
    
    # Copy configuration files
    Copy-Item "config.*.yaml" $releaseDir
    
    # Copy Python executor
    Copy-Item "python_executor" $releaseDir -Recurse
    Copy-Item "api" $releaseDir -Recurse
    Copy-Item "example_plugins" $releaseDir -Recurse
    
    # Copy documentation
    Get-ChildItem "README*.md" | Copy-Item -Destination $releaseDir -ErrorAction SilentlyContinue
    
    # Create ZIP archive
    $archivePath = Join-Path $DistDir "webhook-bridge-$timestamp.zip"
    Compress-Archive -Path $releaseDir -DestinationPath $archivePath
    
    Write-Success "Release packaged at $DistDir"
}

function Main {
    Write-Info "Starting Webhook Bridge deployment..."
    Write-Info "Environment: $Environment"
    
    Test-Dependencies
    Install-PythonDeps
    Invoke-Tests
    Build-Binaries
    Set-Configuration
    New-WindowsService
    New-DockerFiles
    New-Release
    
    Write-Success "Deployment completed successfully!"
    Write-Info "Next steps:"
    Write-Info "  1. Review configuration in config.yaml"
    Write-Info "  2. Test the deployment: .\build\webhook-bridge-server.exe"
    Write-Info "  3. For production: follow the Windows service or Docker setup instructions"
}

# Show help if requested
if ($Help) {
    Show-Help
    exit 0
}

# Enable verbose output if requested
if ($Verbose) {
    $VerbosePreference = "Continue"
}

# Run main function
try {
    Main
}
catch {
    Write-Error "Deployment failed: $_"
    exit 1
}
