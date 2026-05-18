#!/bin/bash
# Clean Development Environment Script
# This script cleans up all development artifacts and temporary files

set -e

echo "🧹 Cleaning webhook-bridge development environment..."

# Function to safely remove files/directories
remove_safely() {
    local path="$1"
    local description="$2"
    
    if [ -e "$path" ]; then
        if rm -rf "$path" 2>/dev/null; then
            echo "✅ Removed $description"
        else
            echo "⚠️  Warning: Could not remove $description"
        fi
    fi
}

# Function to remove files by pattern
remove_by_pattern() {
    local pattern="$1"
    local description="$2"
    
    find . -name "$pattern" -type f 2>/dev/null | while read -r file; do
        remove_safely "$file" "$description ($file)"
    done
}

echo "📦 Cleaning build artifacts..."

# Remove native build artifacts
remove_by_pattern "*.exe" "Native executable"
remove_by_pattern "*.dll" "Dynamic library"
remove_by_pattern "*.so" "Shared object"
remove_by_pattern "*.dylib" "Dynamic library (macOS)"
remove_by_pattern "*.test" "Native test binary"

# Remove build directories
remove_safely "build" "Build directory"

# Remove dist but preserve python-deps
if [ -d "dist/python-deps" ]; then
    # Move python-deps temporarily
    mv dist/python-deps /tmp/webhook-bridge-python-deps-backup 2>/dev/null || true
    remove_safely "dist" "Distribution directory"
    mkdir -p dist
    mv /tmp/webhook-bridge-python-deps-backup dist/python-deps 2>/dev/null || true
else
    remove_safely "dist" "Distribution directory"
    mkdir -p dist
fi

echo "🐍 Cleaning Python artifacts..."

# Remove Python cache
find . -name "__pycache__" -type d 2>/dev/null | while read -r dir; do
    remove_safely "$dir" "Python cache directory"
done

remove_by_pattern "*.pyc" "Python compiled file"
remove_by_pattern "*.pyo" "Python optimized file"
remove_safely ".pytest_cache" "Pytest cache"
remove_safely ".ruff_cache" "Ruff cache"
remove_safely ".nox" "Nox cache"

echo "🌐 Cleaning frontend artifacts..."

# Remove Node.js artifacts
remove_safely "web-nextjs/dist" "Frontend build output"
remove_safely "web-nextjs/.next" "Next.js build cache"
remove_safely "static" "Static files directory"
remove_safely "package-lock.json" "Package lock file (root)"

echo "📝 Cleaning logs and data..."

# Remove runtime directories
remove_safely "logs" "Logs directory"
remove_safely "data" "Data directory"

# Remove log files
remove_by_pattern "*.log" "Log file"

echo "⚙️  Cleaning configuration files..."

# Remove test configuration files
remove_safely "config.test.yaml" "Test configuration"
remove_safely "config.quick.yaml" "Quick configuration"
remove_safely "config.dev.yaml" "Development configuration"
remove_safely "config.local.yaml" "Local configuration"

echo "🔍 Cleaning coverage and analysis files..."

# Remove coverage files
remove_by_pattern "coverage.*" "Coverage file"
remove_by_pattern "*.sarif" "Security analysis file"

echo "🗑️  Cleaning temporary files..."

# Remove temporary files
remove_by_pattern "*.tmp" "Temporary file"
remove_by_pattern "*.temp" "Temporary file"
remove_by_pattern "*.pid" "Process ID file"
remove_by_pattern "*.bak" "Backup file"
remove_by_pattern "*.backup" "Backup file"
remove_by_pattern "*.orig" "Original file"

# Remove OS-specific files
remove_by_pattern ".DS_Store" "macOS metadata file"
remove_by_pattern "Thumbs.db" "Windows thumbnail cache"
remove_by_pattern "desktop.ini" "Windows desktop configuration"

echo "🎉 Development environment cleaned successfully!"
echo ""
echo "📋 Summary:"
echo "   - Build artifacts removed"
echo "   - Python cache cleared"
echo "   - Frontend build outputs removed"
echo "   - Logs and data directories cleaned"
echo "   - Temporary files removed"
echo "   - Native build artifacts cleaned"
echo ""
echo "💡 To start fresh development:"
echo "   uvx nox -s quick    # Quick start"
echo "   uvx nox -s dev      # Full development"
