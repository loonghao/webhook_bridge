# Quick Start Script for Webhook Bridge (Windows)
# This script provides a simple way to start the webhook bridge service on Windows

param(
    [string]$Environment = "dev",
    [switch]$Daemon,
    [switch]$Stop,
    [switch]$Status,
    [switch]$Restart,
    [switch]$Help
)

# Colors
$Green = "Green"
$Yellow = "Yellow"
$Blue = "Blue"
$Red = "Red"

# Configuration
$ScriptDir = Split-Path -Parent $MyInvocation.MyCommand.Path
$ProjectRoot = Split-Path -Parent $ScriptDir
$ConfigFile = Join-Path $ProjectRoot "config.yaml"
$PidFile = Join-Path $ProjectRoot "webhook-bridge.pid"
$PythonPidFile = Join-Path $ProjectRoot "python_executor.pid"

# Global variables for process tracking
$GoServerProcess = $null
$PythonExecutorProcess = $null

function Show-Help {
    @"
Webhook Bridge Quick Start Script for Windows

Usage: .\start.ps1 [OPTIONS]

Options:
    -Environment ENVIRONMENT    Set environment (dev, prod) [default: dev]
    -Daemon                    Run as background service
    -Stop                      Stop running service
    -Restart                   Restart service
    -Status                    Show service status
    -Help                      Show this help message

Examples:
    .\start.ps1                      # Start in development mode
    .\start.ps1 -Environment prod -Daemon  # Start in production mode as service
    .\start.ps1 -Stop                # Stop the service
    .\start.ps1 -Status              # Check service status

"@
}

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

function Test-Prerequisites {
    Write-Info "Checking prerequisites..."
    
    # Check if binaries exist
    $serverPath = Join-Path $ProjectRoot "build\webhook-bridge-server.exe"
    if (-not (Test-Path $serverPath)) {
        Write-Error "webhook-bridge-server.exe not found. Please run deployment script first."
        Write-Info "Run: .\scripts\deploy.ps1"
        exit 1
    }
    
    # Check if Python virtual environment exists
    $venvPath = Join-Path $ProjectRoot ".venv"
    if (-not (Test-Path $venvPath)) {
        Write-Error "Python virtual environment not found. Please run deployment script first."
        Write-Info "Run: .\scripts\deploy.ps1"
        exit 1
    }
    
    # Check configuration
    if (-not (Test-Path $ConfigFile)) {
        Write-Warning "Configuration file not found. Creating default configuration..."
        switch ($Environment) {
            "prod" {
                Copy-Item (Join-Path $ProjectRoot "config.prod.yaml") $ConfigFile
            }
            default {
                Copy-Item (Join-Path $ProjectRoot "config.dev.yaml") $ConfigFile
            }
        }
    }
    
    Write-Success "Prerequisites check passed"
}

function Start-PythonExecutor {
    Write-Info "Starting Python executor..."
    Set-Location $ProjectRoot
    
    # Activate Python virtual environment and start executor
    $pythonScript = Join-Path $ProjectRoot "python_executor\main.py"
    $activateScript = Join-Path $ProjectRoot ".venv\Scripts\Activate.ps1"
    
    if ($Daemon) {
        # Start as background job
        $job = Start-Job -ScriptBlock {
            param($ProjectRoot, $ActivateScript, $PythonScript, $ConfigFile)
            Set-Location $ProjectRoot
            & $ActivateScript
            & python $PythonScript --config $ConfigFile
        } -ArgumentList $ProjectRoot, $activateScript, $pythonScript, $ConfigFile
        
        $job.Id | Out-File $PythonPidFile
        Write-Success "Python executor started as background job (Job ID: $($job.Id))"
    }
    else {
        # Start in current session
        & $activateScript
        $global:PythonExecutorProcess = Start-Process -FilePath "python" -ArgumentList $pythonScript, "--config", $ConfigFile -PassThru -NoNewWindow
        $global:PythonExecutorProcess.Id | Out-File $PythonPidFile
        Write-Success "Python executor started (PID: $($global:PythonExecutorProcess.Id))"
    }
    
    # Wait a moment for Python executor to start
    Start-Sleep -Seconds 2
}

function Start-GoServer {
    Write-Info "Starting Go server..."
    Set-Location $ProjectRoot
    
    $serverPath = Join-Path $ProjectRoot "build\webhook-bridge-server.exe"
    
    if ($Daemon) {
        # Start as background process
        $global:GoServerProcess = Start-Process -FilePath $serverPath -PassThru -WindowStyle Hidden
        $global:GoServerProcess.Id | Out-File $PidFile
        Write-Success "Webhook bridge started as background service (PID: $($global:GoServerProcess.Id))"
        Write-Info "Logs: Check webhook-bridge.log"
    }
    else {
        # Start in current session
        $global:GoServerProcess = Start-Process -FilePath $serverPath -PassThru -NoNewWindow
        $global:GoServerProcess.Id | Out-File $PidFile
        Write-Success "Webhook bridge started (PID: $($global:GoServerProcess.Id))"
        
        # Wait for the process if not daemon
        $global:GoServerProcess.WaitForExit()
    }
}

function Stop-Service {
    Write-Info "Stopping webhook bridge service..."
    
    # Stop Go server
    if (Test-Path $PidFile) {
        $pid = Get-Content $PidFile
        try {
            $process = Get-Process -Id $pid -ErrorAction Stop
            $process.Kill()
            Remove-Item $PidFile
            Write-Success "Go server stopped (PID: $pid)"
        }
        catch {
            Write-Warning "Go server process not found or already stopped"
            Remove-Item $PidFile -ErrorAction SilentlyContinue
        }
    }
    else {
        Write-Warning "PID file not found for Go server"
    }
    
    # Stop Python executor
    if (Test-Path $PythonPidFile) {
        $jobId = Get-Content $PythonPidFile
        try {
            # Try to stop as job first
            $job = Get-Job -Id $jobId -ErrorAction SilentlyContinue
            if ($job) {
                Stop-Job -Id $jobId
                Remove-Job -Id $jobId
                Write-Success "Python executor job stopped (Job ID: $jobId)"
            }
            else {
                # Try to stop as process
                $process = Get-Process -Id $jobId -ErrorAction Stop
                $process.Kill()
                Write-Success "Python executor stopped (PID: $jobId)"
            }
            Remove-Item $PythonPidFile
        }
        catch {
            Write-Warning "Python executor process not found or already stopped"
            Remove-Item $PythonPidFile -ErrorAction SilentlyContinue
        }
    }
    else {
        Write-Warning "PID file not found for Python executor"
    }
}

function Show-Status {
    Write-Info "Checking service status..."
    
    # Check Go server
    if (Test-Path $PidFile) {
        $pid = Get-Content $PidFile
        try {
            $process = Get-Process -Id $pid -ErrorAction Stop
            Write-Success "Go server is running (PID: $pid)"
        }
        catch {
            Write-Error "Go server is not running (stale PID file)"
        }
    }
    else {
        Write-Error "Go server is not running"
    }
    
    # Check Python executor
    if (Test-Path $PythonPidFile) {
        $jobId = Get-Content $PythonPidFile
        try {
            # Check as job first
            $job = Get-Job -Id $jobId -ErrorAction SilentlyContinue
            if ($job) {
                Write-Success "Python executor is running (Job ID: $jobId, State: $($job.State))"
            }
            else {
                # Check as process
                $process = Get-Process -Id $jobId -ErrorAction Stop
                Write-Success "Python executor is running (PID: $jobId)"
            }
        }
        catch {
            Write-Error "Python executor is not running (stale PID file)"
        }
    }
    else {
        Write-Error "Python executor is not running"
    }
    
    # Check if services are responding
    Write-Info "Testing service endpoints..."
    
    # Try to get port from config or use default
    try {
        $configContent = Get-Content $ConfigFile | Out-String
        $port = 8080  # default
        if ($configContent -match "port:\s*(\d+)") {
            $port = $matches[1]
        }
        
        $response = Invoke-WebRequest -Uri "http://localhost:$port/health" -TimeoutSec 5 -ErrorAction Stop
        Write-Success "HTTP server is responding on port $port"
    }
    catch {
        Write-Warning "HTTP server is not responding on port $port"
    }
}

function Restart-Service {
    Write-Info "Restarting webhook bridge service..."
    Stop-Service
    Start-Sleep -Seconds 2
    Start-Service
}

function Start-Service {
    Test-Prerequisites
    Start-PythonExecutor
    Start-GoServer
}

# Show help if requested
if ($Help) {
    Show-Help
    exit 0
}

# Setup cleanup on exit
Register-EngineEvent -SourceIdentifier PowerShell.Exiting -Action {
    if ($global:GoServerProcess -and -not $global:GoServerProcess.HasExited) {
        $global:GoServerProcess.Kill()
    }
    if ($global:PythonExecutorProcess -and -not $global:PythonExecutorProcess.HasExited) {
        $global:PythonExecutorProcess.Kill()
    }
}

# Execute requested action
try {
    if ($Stop) {
        Stop-Service
    }
    elseif ($Status) {
        Show-Status
    }
    elseif ($Restart) {
        Restart-Service
    }
    else {
        # Default action: start
        Write-Info "Starting Webhook Bridge in $Environment mode..."
        if ($Daemon) {
            Write-Info "Running as background service..."
        }
        
        Start-Service
        
        if (-not $Daemon) {
            Write-Info "Press Ctrl+C to stop the service"
            # Keep the script running
            try {
                while ($true) {
                    Start-Sleep -Seconds 1
                }
            }
            catch {
                Write-Info "Stopping services..."
                Stop-Service
            }
        }
    }
}
catch {
    Write-Error "Operation failed: $_"
    exit 1
}
