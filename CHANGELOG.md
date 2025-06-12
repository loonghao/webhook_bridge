## v3.0.0 (2025-06-12)

### BREAKING CHANGE

- Removed separate server and python-manager binaries in favor of unified webhook-bridge CLI
- Docker containers now use unified CLI and port 8080 instead of 8000

### Refactor

- clean up codebase and remove unnecessary test files (#85)

## v2.2.0 (2025-06-01)

### Feat

- optimize CI test flow and timeout settings
- enhance container startup detection and error diagnostics
- resolve Go version mismatch issues and optimize CI/CD pipeline
- update architecture flowchart to reflect hybrid Go/Python design
- enhance GoReleaser and Docker deployment

### Fix

- synchronize Python executor port configuration
- improve Docker entrypoint configuration path logic
- add docker-entrypoint.sh to GoReleaser extra_files
- standardize Go version configuration across all CI workflows
- resolve Go version mismatch in CI environments
- resolve CI docker-compose and protobuf generation issues
- improve Docker Python environment setup for reliable builds
- resolve Docker build Go version mismatch
- resolve documentation build issues and Go module dependencies
- add protoc installation to Docker test workflows

## v2.1.0 (2025-06-01)

### Feat

- add comprehensive code quality check scripts
- add webhook-bridge main CLI to build pipeline
- configure commitizen to auto-update Go version files

### Fix

- complete code quality check system and dashboard build issues
- add missing build:prod script for dashboard production builds
- resolve Go lint issues in development tools
- remove problematic PR preview job from docs workflow
- resolve CI build issues for dashboard and documentation
- resolve Python lint issues and code style violations
- unify version to 2.0.0 across Go and Python packages
- resolve build issues and version consistency
- resolve dashboard, logging, and Python executor issues

## v2.0.0 (2025-05-31)

### BREAKING CHANGE

- Complete rewrite with Go HTTP server and Python CLI tool

### Feat

- optimize CI for Go-first architecture and clean up project
- release v1.0.0 with hybrid Go/Python architecture

### Fix

- resolve golangci-lint errcheck issues
- add path validation to prevent directory traversal attacks (G304)
- resolve security vulnerabilities identified by gosec
- resolve port range conflicts in CI environment
- resolve CI PowerShell compatibility issues
- replace deprecated gosec GitHub Action with direct installation
- optimize Python lint configuration for CI compatibility
- modernize golangci-lint configuration for CI compatibility

## v0.6.0 (2025-05-30)

### Feat

- add enhanced uvicorn options support

### Fix

- use Python 3.11 for mypy session to resolve Pydantic compatibility
- add click and pydantic to mypy dependencies
- remove unsupported poetry --no-update option

### Refactor

- modernize CLI with Click and Pydantic
- fix code complexity and argument count issues

## v0.5.0 (2025-04-12)

### Feat

- add support for RESTful methods (GET, POST, PUT, DELETE)
- add support for RESTful methods (GET, POST, PUT, DELETE)

## v0.4.0 (2025-02-07)

### Feat

- disable docs when deploying to internet-accessible machines according to company security policy

## v0.3.0 (2024-11-25)

### Feat

- add CLI support and improve code style

## v0.2.0 (2024-11-17)

### Feat

- **api**: add FastAPI versioning and plugin system

### Fix

- Other computers cannot post
