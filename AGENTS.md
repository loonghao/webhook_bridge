# AGENTS.md — Webhook Bridge

> Navigation map, not a reference manual. Follow the links; don't read
> everything upfront.

Webhook Bridge is a hybrid Rust/Python webhook gateway. A single
`webhook-bridge` executable provides the CLI, HTTP server, admin commands,
and Python worker management; the Rust core identifies the provider,
persists the event, and triggers Python hooks, script groups, or
forwarding routes.

---

## Repository Contract

**The `Makefile` is the task entrypoint; use it rather than raw cargo commands.**

| Task | Command |
|------|---------|
| Build (debug) | `make build` |
| Build (release) | `make release-build` |
| Test | `make test` |
| Format | `make fmt` |
| Lint | `make lint` |
| Admin dashboard check | `make dashboard-check` |
| Clean | `make clean` |
| List all targets | `make help` |

**Repository layout**

| Path | Role |
|------|------|
| `crates/` | Rust workspace members (gateway, server, providers) |
| `webhook_bridge/` | Python package — hooks and plugin surface |
| `python_executor/` | Python worker execution layer |
| `web-nextjs/` | Admin dashboard (Next.js/TypeScript) |
| `api/`, `docs/` | API definitions and documentation (mkdocs) |
| `example_plugins/` | Reference Python hook plugins |
| `tests/` | Test suite |
| `noxfile.py`, `Makefile` | Task entrypoints |

**Release flow** — `release-please` on `main` drives `CHANGELOG.md` and the version in
`pyproject.toml` from Conventional Commit subjects. Tagging and
publishing run in CI. Never edit `CHANGELOG.md` or a version string by hand.

**Prohibitions**

- Do not bypass the Makefile for routine build/test/lint work.
- Do not edit `CHANGELOG.md` or version strings manually.
- Do not add a second agent contract file at the repository root; `AGENTS.md` is the single source.
- Do not commit configuration secrets — `config.example.yaml` is the template; real configs stay local.

---

## Agent Contract Files

`AGENTS.md` is the **only** agent contract file at the repository root. It is the
native instruction file for Codex, OpenCode, Cursor, GitHub Copilot, Windsurf,
Cline, Roo Code, Kiro, Trae, and Augment, and Claude Code falls back to it when
no `CLAUDE.md` exists — so do not add `CLAUDE.md`, `GEMINI.md`, `CURSOR.md`, or
any other vendor-specific variant.

**Gemini CLI exception:** Gemini CLI defaults its context file to `GEMINI.md`. To
make it read `AGENTS.md`, set `context.fileName` once in `~/.gemini/settings.json`:

```json
{
  "context": {
    "fileName": ["AGENTS.md", "GEMINI.md"]
  }
}
```
