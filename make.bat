@echo off
REM Webhook Bridge Build System for Windows
REM Provides make-like functionality on Windows

setlocal enabledelayedexpansion

REM Colors (limited support in Windows)
set "BLUE=[34m"
set "GREEN=[32m"
set "YELLOW=[33m"
set "RED=[31m"
set "NC=[0m"

REM Configuration
set "PROJECT_NAME=webhook-bridge"
set "BUILD_DIR=build"
set "DIST_DIR=dist"
set "SCRIPTS_DIR=scripts"
set "GO_EXE=C:\Program Files\Go\bin\go.exe"

REM Parse command line arguments
if "%1"=="" goto help
if "%1"=="help" goto help
if "%1"=="deps" goto deps
if "%1"=="build" goto build
if "%1"=="start" goto start
if "%1"=="stop" goto stop
if "%1"=="restart" goto restart
if "%1"=="status" goto status
if "%1"=="test" goto test
if "%1"=="test-integration" goto test-integration
if "%1"=="clean" goto clean
if "%1"=="deploy-dev" goto deploy-dev
if "%1"=="deploy-prod" goto deploy-prod
if "%1"=="version" goto version
if "%1"=="quick" goto quick

echo Unknown target: %1
goto help

:help
echo %BLUE%Webhook Bridge - Enhanced Build System%NC%
echo.
echo %GREEN%Development:%NC%
echo   deps         - Install all dependencies
echo   build        - Build Go binaries
echo   start        - Start services in development mode
echo   stop         - Stop running services
echo   restart      - Restart services
echo   status       - Show service status
echo.
echo %GREEN%Testing:%NC%
echo   test         - Run all tests
echo   test-integration - Run integration tests
echo.
echo %GREEN%Deployment:%NC%
echo   deploy-dev   - Deploy for development
echo   deploy-prod  - Deploy for production
echo   clean        - Clean build artifacts
echo.
echo %GREEN%Utilities:%NC%
echo   version      - Show version information
echo   quick        - Quick development workflow
goto end

:deps
echo %BLUE%Installing dependencies...%NC%
echo Installing Go dependencies...
"%GO_EXE%" mod download
"%GO_EXE%" mod tidy
echo Installing Python dependencies...
if exist .venv (
    .venv\Scripts\python.exe -m pip install -r requirements.txt
) else (
    python -m venv .venv
    .venv\Scripts\python.exe -m pip install -r requirements.txt
)
echo %GREEN%Dependencies installed%NC%
goto end

:build
echo %BLUE%Building Go binaries...%NC%
if not exist %BUILD_DIR% mkdir %BUILD_DIR%
"%GO_EXE%" build -o %BUILD_DIR%\webhook-bridge-server.exe .\cmd\server
"%GO_EXE%" build -o %BUILD_DIR%\python-manager.exe .\cmd\python-manager
echo %GREEN%Build completed%NC%
goto end

:start
echo %BLUE%Starting webhook bridge in development mode...%NC%
powershell -ExecutionPolicy Bypass -File %SCRIPTS_DIR%\start.ps1 -Environment dev
goto end

:stop
echo %BLUE%Stopping webhook bridge...%NC%
powershell -ExecutionPolicy Bypass -File %SCRIPTS_DIR%\start.ps1 -Stop
goto end

:restart
echo %BLUE%Restarting webhook bridge...%NC%
powershell -ExecutionPolicy Bypass -File %SCRIPTS_DIR%\start.ps1 -Restart
goto end

:status
powershell -ExecutionPolicy Bypass -File %SCRIPTS_DIR%\start.ps1 -Status
goto end

:test
echo %BLUE%Running tests...%NC%
echo Running Go tests...
"%GO_EXE%" test -v .\...
echo Running Python tests...
if exist .venv (
    .venv\Scripts\python.exe -m pytest tests\ -v
) else (
    echo Python virtual environment not found. Run 'make deps' first.
)
goto end

:test-integration
echo %BLUE%Running integration tests...%NC%
if exist .venv (
    .venv\Scripts\python.exe test_go_python_integration.py
) else (
    echo Python virtual environment not found. Run 'make deps' first.
)
goto end

:clean
echo %BLUE%Cleaning build artifacts...%NC%
if exist bin rmdir /s /q bin
if exist %DIST_DIR% rmdir /s /q %DIST_DIR%
if exist %BUILD_DIR% rmdir /s /q %BUILD_DIR%
del /q *.log *.pid 2>nul
"%GO_EXE%" clean
echo %GREEN%Clean completed%NC%
goto end

:deploy-dev
echo %BLUE%Deploying for development...%NC%
powershell -ExecutionPolicy Bypass -File %SCRIPTS_DIR%\deploy.ps1 -Environment dev
goto end

:deploy-prod
echo %BLUE%Deploying for production...%NC%
powershell -ExecutionPolicy Bypass -File %SCRIPTS_DIR%\deploy.ps1 -Environment prod
goto end

:version
echo Project: %PROJECT_NAME%
for /f "tokens=*" %%i in ('git describe --tags --always --dirty 2^>nul') do set VERSION=%%i
if "%VERSION%"=="" set VERSION=dev
echo Version: %VERSION%
echo Build Time: %date% %time%
for /f "tokens=3" %%i in ('"%GO_EXE%" version') do set GO_VERSION=%%i
echo Go Version: %GO_VERSION%
goto end

:quick
echo %BLUE%Running quick development workflow...%NC%
call :clean
call :deps
call :build
call :test
echo %GREEN%Quick development workflow completed%NC%
goto end

:end
endlocal
