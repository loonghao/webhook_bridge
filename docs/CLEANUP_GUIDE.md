# 🧹 Cleanup Guide

This guide covers the comprehensive cleanup system for webhook-bridge development environment.

## Quick Reference

| Command | Description | Use Case |
|---------|-------------|----------|
| `uvx nox -s clean-local` | Basic cleanup | Remove binaries and test configs |
| `uvx nox -s clean-all` | Deep cleanup | Remove all development artifacts |
| `scripts/clean-dev.ps1` | Windows script | Direct PowerShell cleanup |
| `scripts/clean-dev.sh` | Unix script | Direct Bash cleanup |

## What Gets Cleaned

### 🔨 Build Artifacts
- **Go binaries**: `*.exe`, `*.dll`, `*.so`, `*.dylib`, `*.test`
- **Build directories**: `build/`, `dist/` (preserves `dist/python-deps/`)
- **Coverage files**: `coverage.*`, `*.sarif`

### 🐍 Python Artifacts
- **Cache directories**: `__pycache__/`, `.pytest_cache/`, `.ruff_cache/`, `.nox/`
- **Compiled files**: `*.pyc`, `*.pyo`
- **Package files**: `poetry.lock` (when appropriate)

### 🌐 Frontend Artifacts
- **Build outputs**: `web-nextjs/dist/`, `web-nextjs/.next/`, `static/`
- **Dependencies**: `package-lock.json` (root level)
- **Node modules**: Preserved in `web-nextjs/node_modules/`

### 📝 Runtime Data
- **Logs**: `logs/`, `*.log`
- **Data**: `data/`
- **Process files**: `*.pid`

### ⚙️ Configuration Files
- **Test configs**: `config.test.yaml`, `config.quick.yaml`, `config.dev.yaml`
- **Local configs**: `config.local.yaml`

### 🗑️ Temporary Files
- **Temp files**: `*.tmp`, `*.temp`
- **Backup files**: `*.bak`, `*.backup`, `*.orig`
- **OS files**: `.DS_Store`, `Thumbs.db`, `desktop.ini`

## Cleanup Commands

### Basic Cleanup (`clean-local`)
```bash
uvx nox -s clean-local
```
**Removes**:
- Go binaries (*.exe)
- Test configuration files
- Go build cache

**Use when**: Quick cleanup between builds

### Deep Cleanup (`clean-all`)
```bash
uvx nox -s clean-all
```
**Removes**:
- All build artifacts
- All cache directories
- All temporary files
- All runtime data
- All test configurations

**Use when**: 
- Starting fresh development
- Preparing for release
- Troubleshooting build issues
- Cleaning up before committing

### Direct Script Usage

#### Windows
```powershell
.\scripts\clean-dev.ps1
```

#### Unix/Linux/macOS
```bash
./scripts/clean-dev.sh
```

## .gitignore Coverage

The updated `.gitignore` ensures these file types are never committed:

### 📦 Build and Distribution
```gitignore
# Go build artifacts
*.exe
*.dll
*.so
*.dylib
build/
dist/
!dist/python-deps/

# Frontend builds
web-nextjs/dist/
web-nextjs/.next/
static/
package-lock.json
```

### 🐍 Python Development
```gitignore
# Python cache and testing
__pycache__/
*.pyc
.pytest_cache/
.ruff_cache/
.nox/

# Virtual environments
.venv/
venv/
```

### 📝 Development Files
```gitignore
# Test configurations
config.test.yaml
config.quick.yaml
config.dev.yaml
config.local.yaml

# Runtime data
logs/
data/
*.log
*.pid
```

## Best Practices

### 🔄 Regular Cleanup
```bash
# Before starting work
uvx nox -s clean-local

# Weekly deep clean
uvx nox -s clean-all

# Before committing
git status  # Check for untracked files
```

### 🚀 Development Workflow
```bash
# 1. Clean environment
uvx nox -s clean-all

# 2. Start fresh development
uvx nox -s quick

# 3. Work on features...

# 4. Clean before commit
uvx nox -s clean-local
git add .
git commit -m "feat: your changes"
```

### 🔍 Troubleshooting
```bash
# Build issues? Deep clean first
uvx nox -s clean-all
uvx nox -s dev

# Cache corruption? Clean Go cache
go clean -cache -testcache -modcache

# Frontend issues? Clean Node modules
cd web && rm -rf node_modules && npm install
```

## Automation

### Pre-commit Hook
Add to `.git/hooks/pre-commit`:
```bash
#!/bin/bash
# Clean up before commit
uvx nox -s clean-local
```

### CI/CD Integration
```yaml
# In GitHub Actions
- name: Clean environment
  run: uvx nox -s clean-all
```

## Safety Features

### 🛡️ Protected Files
- **Source code**: Never removed
- **Configuration templates**: `config.example.yaml` preserved
- **Documentation**: All `docs/` content preserved
- **Python dependencies**: `dist/python-deps/` preserved
- **Web dependencies**: `web-nextjs/node_modules/` preserved

### ⚠️ Warnings
- Scripts show warnings for files that couldn't be removed
- Go cache cleaning may fail on Windows (file locks)
- Some operations require elevated permissions

## Recovery

If you accidentally remove important files:

### 🔄 Restore from Git
```bash
# Restore specific file
git checkout HEAD -- filename

# Restore all tracked files
git checkout HEAD -- .
```

### 🔨 Rebuild Environment
```bash
# Rebuild everything
uvx nox -s clean-all
uvx nox -s dev
```

### 📦 Reinstall Dependencies
```bash
# Go dependencies
go mod download

# Frontend dependencies
cd web && npm install

# Python dependencies
uv sync
```
