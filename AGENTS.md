# AGENTS.md — webhook_bridge

> Rust webhook gateway with a single `webhook-bridge` executable (CLI, HTTP
> server, admin commands, Python worker management), a Python hook runtime, a
> gRPC Python executor and a Next.js dashboard. Currently `4.0.0-alpha.x`.
> Navigation map for AI agents, not a reference manual. Follow the links; do
> not read everything up front.

## Build & test

No justfile. The `Makefile` at the root is the canonical entry point — run
`make help` for the live list.

```bash
make build                    # cargo build -p webhook-bridge-server --bin webhook-bridge
make release-build            # release-optimised binary
make test                     # cargo test + dashboard type-check and lint
make fmt                      # cargo fmt
make lint                     # cargo fmt --check
make dashboard-check          # cd web-nextjs && npm run type-check && npm run lint
make run                      # build, then run --config config.4.0.yaml
make admin / make worker      # admin summary / start one standalone worker
make clean                    # cargo clean
```

Python hook runtime tests and lint run through nox:

```bash
nox -s pytest                 # pytest for the webhook_bridge Python package
nox -s lint                   # autoflake, ruff, black, mypy
nox -s lint-fix               # autofix variants
nox -s dev / start-server / build-local / test-local / run-local   # local stack
nox -s clean-local / clean-all
```

Raw commands, when you only need one side:

```bash
cargo test                    # Rust workspace tests (crates/bridge-core, crates/bridge-server)
cargo run -p webhook-bridge-server --bin webhook-bridge -- run --config config.4.0.yaml
cd web-nextjs && npm run type-check && npm run lint
```

`tasks.yaml` (Taskfile) mirrors a subset of the above as `task build`,
`task test`, `task lint`, `task dev`, `task release`, `task clean`.

Local endpoints: dashboard <http://127.0.0.1:3002>, API health
<http://127.0.0.1:8080/health>, unified gateway <http://127.0.0.1:8080/gateway>.

## Repo layout

| Path | Role |
|---|---|
| `crates/bridge-core/` | Domain library — `config.rs`, `executor.rs`, `runtime.rs`, `proto.rs`, `storage.rs` |
| `crates/bridge-server/` | Binary crate producing `webhook-bridge` (`src/main.rs`, `src/storage.rs`) |
| `webhook_bridge/` | Python hook runtime package: `plugin.py`, `filesystem.py`, `cli.py`, `__version__.py` |
| `python_executor/` | Standalone gRPC Python executor (`server.py`, `main.py`, `utils.py`) |
| `api/proto/` | Protobuf contracts shared by the Rust server and Python executor |
| `web-nextjs/` | Next.js dashboard (`app/`, `components/`, `hooks/`, `lib/`, `services/`, `types/`) |
| `example_plugins/` | Sample Python hooks: `github.py`, `sentry.py`, `restful_plugin.py`, `test_plugin.py` |
| `tests/` | Pytest suite for the Python package (`test_api.py`, `test_cli.py`, `test_paths.py`, `test_data/`) |
| `docs/` | `WEBHOOK_BRIDGE_4_ARCHITECTURE.md`, `DASHBOARD_GUIDE.md` |
| `tools/` | MCP debug helpers for the dashboard |
| `config.4.0.yaml`, `config.example.yaml` | Runtime configuration; `check-config` validates a file |
| `mkdocs.yml`, `noxfile.py`, `nox_actions/` | Docs site and nox session implementations |
| `Dockerfile`, `docker-compose.yml` | Container build and local stack |

## Release

- release-please drives versioning from Conventional Commits on `main`
  (`release-please-config.json`, `release-type: simple`).
  `.release-please-manifest.json` is the single source of version truth.
- `feat:` → minor, `fix:` → patch, `chore:`/`docs:`/`ci:` → **no release**.
- Use `chore:`/`docs:` for config and doc work so release-please does not cut a
  valueless version.
- release-please also rewrites the version in **four** files via `extra-files`:
  `crates/bridge-core/Cargo.toml`, `crates/bridge-server/Cargo.toml`,
  `pyproject.toml` and `web-nextjs/package.json`. Never hand-edit those
  versions; `[tool.commitizen]` in `pyproject.toml` holds the same file list
  for manual bumps.
- After a release is created, `.github/workflows/release.yml` builds and
  uploads `linux-x86_64`, `windows-x86_64` and `darwin-arm64` binaries.

## Do / Don't

- **Do** put new server domain logic in `crates/bridge-core`, not in the
  binary crate.
- **Do** keep `api/proto/` authoritative when changing the Rust ↔ Python
  executor contract — both sides are generated from it.
- **Do** run `make dashboard-check` whenever you touch `web-nextjs/`.
- **Don't** assume the Python package is the server. `webhook_bridge/` is the
  hook/plugin runtime; the HTTP server is Rust.
- **Don't** edit the version in the four release-please `extra-files` by hand.
- **Don't** add `CLAUDE.md` / `GEMINI.md` / `CURSOR.md` / `ANTHROPIC.md` /
  `OPENAI.md` / `COPILOT.md` / `CODEBUDDY.md` / `.cursorrules` / `.clinerules` /
  `.windsurfrules` at the root. This file is the only agent contract file.
- **Don't** hardcode an exact version in tests (`assert __version__ == "X.Y.Z"`)
  — release-please bumps will break it. Use `>=` or read package metadata.
- **Don't** commit build artifacts to the repo root (`target/`, `dist/`,
  `*.log`, `*.pid`, `coverage.xml`, `node_modules/`).
