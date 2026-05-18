# Changelog

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
