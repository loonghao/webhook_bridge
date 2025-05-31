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
