@echo off
REM Comprehensive code quality check script for webhook-bridge (Windows)
REM This script checks Go, Python, and Node.js/TypeScript code quality

setlocal enabledelayedexpansion

echo Starting comprehensive code quality checks...
echo ================================================

REM Check if we're in the right directory
if not exist "go.mod" (
    echo [ERROR] Please run this script from the webhook-bridge root directory
    exit /b 1
)
if not exist "pyproject.toml" (
    echo [ERROR] Please run this script from the webhook-bridge root directory
    exit /b 1
)
if not exist "web" (
    echo [ERROR] Please run this script from the webhook-bridge root directory
    exit /b 1
)

REM Track overall status
set OVERALL_STATUS=0

echo.
echo Go Code Quality Checks
echo =======================

echo [INFO] Running Go formatting check...
gofmt -l . > temp_gofmt_output.txt
for /f %%i in (temp_gofmt_output.txt) do set GOFMT_NEEDED=1
del temp_gofmt_output.txt

if defined GOFMT_NEEDED (
    echo [WARNING] Go code needs formatting. Running gofmt...
    gofmt -w .
    echo [SUCCESS] Go code formatted
) else (
    echo [SUCCESS] Go code is properly formatted
)

echo [INFO] Running Go linting...
go run dev.go lint
if !errorlevel! neq 0 (
    echo [ERROR] Go linting failed
    set OVERALL_STATUS=1
) else (
    echo [SUCCESS] Go linting passed
)

echo [INFO] Running Go tests...
go run dev.go test
if !errorlevel! neq 0 (
    echo [ERROR] Go tests failed
    set OVERALL_STATUS=1
) else (
    echo [SUCCESS] Go tests passed
)

echo [INFO] Running Go build test...
go run dev.go build
if !errorlevel! neq 0 (
    echo [ERROR] Go build failed
    set OVERALL_STATUS=1
) else (
    echo [SUCCESS] Go build successful
)

echo.
echo Python Code Quality Checks
echo ===========================

echo [INFO] Checking if uv is available...
where uvx >nul 2>&1
if !errorlevel! neq 0 (
    echo [ERROR] uvx is not available. Please install uv first.
    set OVERALL_STATUS=1
) else (
    echo [INFO] Running Python linting...
    uvx nox -s lint
    if !errorlevel! neq 0 (
        echo [ERROR] Python linting failed
        set OVERALL_STATUS=1
    ) else (
        echo [SUCCESS] Python linting passed
    )

    echo [INFO] Running Python tests...
    uvx nox -s pytest
    if !errorlevel! neq 0 (
        echo [ERROR] Python tests failed
        set OVERALL_STATUS=1
    ) else (
        echo [SUCCESS] Python tests passed
    )
)

echo.
echo Node.js/TypeScript Code Quality Checks
echo =======================================

if exist "web" (
    echo [INFO] Checking if npm is available...
    where npm >nul 2>&1
    if !errorlevel! neq 0 (
        echo [ERROR] npm is not available. Please install Node.js first.
        set OVERALL_STATUS=1
    ) else (
        echo [INFO] Installing dependencies...
        pushd web
        npm install --no-audit >nul 2>&1
        popd
        echo [SUCCESS] Dependencies installed

        echo [INFO] Running TypeScript linting...
        pushd web
        npm run lint
        set LINT_RESULT=!errorlevel!
        popd
        if !LINT_RESULT! neq 0 (
            echo [ERROR] TypeScript linting failed
            set OVERALL_STATUS=1
        ) else (
            echo [SUCCESS] TypeScript linting passed
        )

        echo [INFO] Running TypeScript type checking...
        pushd web
        npm run type-check
        set TYPECHECK_RESULT=!errorlevel!
        popd
        if !TYPECHECK_RESULT! neq 0 (
            echo [ERROR] TypeScript type checking failed
            set OVERALL_STATUS=1
        ) else (
            echo [SUCCESS] TypeScript type checking passed
        )

        echo [INFO] Running TypeScript build...
        pushd web
        npm run build
        set BUILD_RESULT=!errorlevel!
        popd
        if !BUILD_RESULT! neq 0 (
            echo [ERROR] TypeScript build failed
            set OVERALL_STATUS=1
        ) else (
            echo [SUCCESS] TypeScript build successful
        )
    )
) else (
    echo [WARNING] Web directory not found, skipping Node.js checks
)

echo.
echo Summary
echo =======

if !OVERALL_STATUS! equ 0 (
    echo [SUCCESS] All code quality checks passed!
    echo.
    echo Your code is ready for:
    echo   - Commit and push
    echo   - Pull request creation
    echo   - Production deployment
) else (
    echo [ERROR] Some checks failed. Please fix the issues above.
    echo.
    echo Common fixes:
    echo   - Run 'gofmt -w .' for Go formatting
    echo   - Run 'uvx nox -s lint' for Python issues
    echo   - Run 'npm run lint:fix' in web/ for TypeScript issues
)

exit /b !OVERALL_STATUS!
