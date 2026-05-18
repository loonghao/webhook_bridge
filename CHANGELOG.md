# Changelog

## [5.0.0-alpha.0](https://github.com/loonghao/webhook_bridge/compare/webhook-bridge-v4.0.0-alpha.0...webhook-bridge-v5.0.0-alpha.0) (2026-05-18)


### ⚠ BREAKING CHANGES

* Removed separate server and python-manager binaries in favor of unified webhook-bridge CLI
* Complete rewrite with Go HTTP server and Python CLI tool

### Features

* add CLI support and improve code style ([285a907](https://github.com/loonghao/webhook_bridge/commit/285a907c031b8cf8c88db7b8b9e75f36bb27be8a))
* add comprehensive code quality check scripts ([a37994e](https://github.com/loonghao/webhook_bridge/commit/a37994e55252f837065e7e45a4d4bc1566dffa13))
* add enhanced uvicorn options support ([ecfab71](https://github.com/loonghao/webhook_bridge/commit/ecfab71a167cb16be784b8d327b914836dbff7d9))
* add support for RESTful methods (GET, POST, PUT, DELETE) ([1ae99c1](https://github.com/loonghao/webhook_bridge/commit/1ae99c1d507822ea9f6ee07d64ec6d113c63712c))
* add support for RESTful methods (GET, POST, PUT, DELETE) ([e26d193](https://github.com/loonghao/webhook_bridge/commit/e26d193fac9bfcd365d1989876e46744d6a86d12))
* add webhook-bridge main CLI to build pipeline ([59122a0](https://github.com/loonghao/webhook_bridge/commit/59122a03807f0431c6dd2eeb32961237a09f9183))
* **api:** add FastAPI versioning and plugin system ([2bad5df](https://github.com/loonghao/webhook_bridge/commit/2bad5dfdce7f9b73cce82a7cbbc37ce687a2deea))
* configure commitizen to auto-update Go version files ([dceb0d2](https://github.com/loonghao/webhook_bridge/commit/dceb0d27d0385006065c036982fc147fb955286c))
* disable docs when deploying to internet-accessible machines according to company security policy ([fcb44f3](https://github.com/loonghao/webhook_bridge/commit/fcb44f3be331911617032e012714b81cde77ce83))
* enhance container startup detection and error diagnostics ([4fb5ce0](https://github.com/loonghao/webhook_bridge/commit/4fb5ce0cc1976e37af299f415eef5bd1f0ba2387))
* enhance GoReleaser and Docker deployment ([f787858](https://github.com/loonghao/webhook_bridge/commit/f78785876294775224be7dc0c6c078895065e38a))
* make coverage tests informational only, not blocking ([8aacedf](https://github.com/loonghao/webhook_bridge/commit/8aacedffb5f30204afd8acf4e67bafcebd457d01))
* optimize CI for Go-first architecture and clean up project ([bb181f1](https://github.com/loonghao/webhook_bridge/commit/bb181f1ec3ad391eb46dc94b29e708c656250323))
* optimize CI test flow and timeout settings ([a1b2f18](https://github.com/loonghao/webhook_bridge/commit/a1b2f1830e98c2570ee44ec168dc7b526bc96c1f))
* rebuild webhook bridge 4.0 runtime ([e64bff6](https://github.com/loonghao/webhook_bridge/commit/e64bff69b2fba6577a3af3a1112ca0e032bc0c37))
* release v1.0.0 with hybrid Go/Python architecture ([3f0097d](https://github.com/loonghao/webhook_bridge/commit/3f0097dca81376e7814e6ea4ce4b2a7a80fa8f40))
* resolve Go version mismatch issues and optimize CI/CD pipeline ([c0906a6](https://github.com/loonghao/webhook_bridge/commit/c0906a6384fdccca28b05623980669db225badf1))
* update architecture flowchart to reflect hybrid Go/Python design ([bd1d7b4](https://github.com/loonghao/webhook_bridge/commit/bd1d7b4488762e82c8fac7204ab3261afbed7249))


### Bug Fixes

* add click and pydantic to mypy dependencies ([f2b36e3](https://github.com/loonghao/webhook_bridge/commit/f2b36e33dd7a9038501830e95e4e685896f9e234))
* add docker-entrypoint.sh to GoReleaser extra_files ([07ac1b5](https://github.com/loonghao/webhook_bridge/commit/07ac1b510a0b41f7093b86bd4df99aa71c948259))
* add missing build:prod script for dashboard production builds ([a7a238e](https://github.com/loonghao/webhook_bridge/commit/a7a238ee79b7907c55badd3b68bb6a76a82b5486))
* add path validation to prevent directory traversal attacks (G304) ([55cd7f7](https://github.com/loonghao/webhook_bridge/commit/55cd7f75fcdb021590545020bdbea717d4375e2e))
* add protoc installation to Docker test workflows ([5acd97f](https://github.com/loonghao/webhook_bridge/commit/5acd97fe7a545b749c109232de7baeb3246d1ca1))
* complete code quality check system and dashboard build issues ([58935b6](https://github.com/loonghao/webhook_bridge/commit/58935b64bbb2d2b26842286aa39e2463b61e5970))
* **deps:** update dependency @modelcontextprotocol/sdk to v1 [security] ([#103](https://github.com/loonghao/webhook_bridge/issues/103)) ([01d26fa](https://github.com/loonghao/webhook_bridge/commit/01d26fa985d82ae2f3493ab4d33ef5429d2ca3ca))
* **deps:** update dependency lucide-react to ^0.516.0 ([35060da](https://github.com/loonghao/webhook_bridge/commit/35060da4a935804abff533e47b51a27f3fbc6187))
* improve Docker entrypoint configuration path logic ([f5cca58](https://github.com/loonghao/webhook_bridge/commit/f5cca586fc0efbf16c56b99a4b135b72191a0196))
* improve Docker Python environment setup for reliable builds ([b9b9d8f](https://github.com/loonghao/webhook_bridge/commit/b9b9d8ff08e51152dff1c0961b5972a93aff6110))
* modernize golangci-lint configuration for CI compatibility ([ba5e161](https://github.com/loonghao/webhook_bridge/commit/ba5e16141f54ce3d68221240ebcc5f76782af009))
* optimize Python lint configuration for CI compatibility ([f1c85a0](https://github.com/loonghao/webhook_bridge/commit/f1c85a00835940844f8b0ef17dcf2b843d121740))
* Other computers cannot post ([f7fcbc7](https://github.com/loonghao/webhook_bridge/commit/f7fcbc73f0da794d34ca44397b75ff688ff8e649))
* remove problematic PR preview job from docs workflow ([a3d03f0](https://github.com/loonghao/webhook_bridge/commit/a3d03f09dba3a34f4a955f26ef22c96afa26260d))
* remove unsupported poetry --no-update option ([df13acb](https://github.com/loonghao/webhook_bridge/commit/df13acb48be83a46f510c540f883534251055c03))
* replace deprecated gosec GitHub Action with direct installation ([90fae7c](https://github.com/loonghao/webhook_bridge/commit/90fae7cabc53b05b0030a4e6c9b09a09830f0017))
* resolve build issues and version consistency ([71cdf3d](https://github.com/loonghao/webhook_bridge/commit/71cdf3d6258ef29fc88dc88418a980278e485b24))
* resolve CI build issues for dashboard and documentation ([8f035ca](https://github.com/loonghao/webhook_bridge/commit/8f035cace2907ce81f56d3d03b3725cbd92b4694))
* resolve CI docker-compose and protobuf generation issues ([f593847](https://github.com/loonghao/webhook_bridge/commit/f5938474591cc86a713dd700e7025ab81d4d9aad))
* resolve CI PowerShell compatibility issues ([da6a4ae](https://github.com/loonghao/webhook_bridge/commit/da6a4aef825c2e06167a8b6c20ced4d0c2275b72))
* resolve dashboard, logging, and Python executor issues ([13c3456](https://github.com/loonghao/webhook_bridge/commit/13c3456a8732d5d61e25f25095d9732a8bfeb392))
* resolve Docker build Go version mismatch ([f9d38c2](https://github.com/loonghao/webhook_bridge/commit/f9d38c2ce5f142b8869b9bd82a858515a50a20f3))
* resolve documentation build issues and Go module dependencies ([8eb8742](https://github.com/loonghao/webhook_bridge/commit/8eb8742f18ea4605c8b34310f9c7a9c0b944b23a))
* resolve Go lint issues in development tools ([c9a1dbc](https://github.com/loonghao/webhook_bridge/commit/c9a1dbc2580f4193e7b5830db913c23988a716f0))
* resolve Go version mismatch in CI environments ([53bc873](https://github.com/loonghao/webhook_bridge/commit/53bc873a74ac8e3e24e27cc767a3b900eda38706))
* resolve golangci-lint errcheck issues ([0a1df23](https://github.com/loonghao/webhook_bridge/commit/0a1df23f85def1b70425e674a074fee1e2a4eb51))
* resolve GoReleaser 404 errors and update dependencies ([cf695dc](https://github.com/loonghao/webhook_bridge/commit/cf695dc76eab7063c83e38bccd79a515f62bde99))
* resolve macOS file system race condition in plugin stats storage ([bef8efe](https://github.com/loonghao/webhook_bridge/commit/bef8efec383853ed8c8fa0a7215c99a3e69dbe25))
* resolve poetry lock and PyPI publish issues ([452da5f](https://github.com/loonghao/webhook_bridge/commit/452da5f79e8ba7ab8dbf5219680999487a895ab0))
* resolve port range conflicts in CI environment ([34170a5](https://github.com/loonghao/webhook_bridge/commit/34170a5ecc0dc6189fefd5fb23905269ce96b6cb))
* resolve Python lint issues and code style violations ([b12b386](https://github.com/loonghao/webhook_bridge/commit/b12b386b5071e1a87e851216d790ff60b6d2f1a9))
* resolve security vulnerabilities identified by gosec ([dd08a7b](https://github.com/loonghao/webhook_bridge/commit/dd08a7bd3920e774b165612e5f47e81e4ea5c49a))
* standardize Go version configuration across all CI workflows ([4cd4c53](https://github.com/loonghao/webhook_bridge/commit/4cd4c5322c5257918f8f4c13486d0bb1f993574e))
* synchronize Python executor port configuration ([9acefed](https://github.com/loonghao/webhook_bridge/commit/9acefed2f5e5a8671713c958e51cb5f366d9e8c3))
* unify version to 2.0.0 across Go and Python packages ([403672b](https://github.com/loonghao/webhook_bridge/commit/403672bc18d8746321d3049c40d2e4a559432409))
* use Python 3.11 for mypy session to resolve Pydantic compatibility ([c360756](https://github.com/loonghao/webhook_bridge/commit/c360756b0ec540c705313a97c2bdaff266463188))


### Code Refactoring

* clean up codebase and remove unnecessary test files ([#85](https://github.com/loonghao/webhook_bridge/issues/85)) ([cab1fd0](https://github.com/loonghao/webhook_bridge/commit/cab1fd04c5dda6a7a8fa01a3afe3f16a7cf894e8))

## 4.0.0-alpha.0

- Rebuilt Webhook Bridge as a Rust control plane with one `webhook-bridge`
  executable for CLI, server, worker, and admin workflows.
- Added a unified `/gateway` webhook ingress that detects providers such as
  GitHub, GitLab, and Sentry.
- Added Python hook execution through Rust-managed workers with optional `uv`
  environment management.
- Added local script routes and parallel script groups for PowerShell, Python,
  forwarding, and other command-driven integrations.
- Added SQLite-backed execution records and runtime logs.
- Added a Next.js dashboard for routes, workers, logs, and runtime status.
- Added release-please based release automation for platform executables.

## v4.0.0 (2026-05-18)

### Feat

- rebuild webhook bridge 4.0 runtime

### Fix

- **deps**: update dependency @modelcontextprotocol/sdk to v1 [security] (#103)

## v3.1.1 (2025-06-17)

### Fix

- **deps**: update dependency lucide-react to ^0.516.0

## v3.1.0 (2025-06-13)

### Feat

- make coverage tests informational only, not blocking

### Fix

- resolve macOS file system race condition in plugin stats storage
- resolve GoReleaser 404 errors and update dependencies
- resolve poetry lock and PyPI publish issues

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
