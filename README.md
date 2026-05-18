# Webhook Bridge

Webhook Bridge 4.0 is a Rust-based webhook gateway with a single
`webhook-bridge` executable for the CLI, HTTP server, admin commands, and
Python worker management.

Expose one public URL, configure it in GitHub/GitLab/Sentry, and let the Rust
gateway identify the provider, persist the event, and trigger local Python
hooks, PowerShell/Python scripts, parallel script groups, or forwarding routes.

## Quick Start

```bash
cargo run -p webhook-bridge-server --bin webhook-bridge -- run --config config.4.0.yaml
```

Open:

- Dashboard: http://127.0.0.1:3002
- API health: http://127.0.0.1:8080/health
- Unified webhook gateway: http://127.0.0.1:8080/gateway

## CLI

```bash
webhook-bridge run --config config.4.0.yaml
webhook-bridge admin --config config.4.0.yaml
webhook-bridge worker start --config config.4.0.yaml --index 0
webhook-bridge check-config --config config.4.0.yaml
```

## Unified Gateway

Configure external providers to send webhooks to a single URL:

```text
https://your-host.example.com/gateway
```

The gateway detects common provider headers such as `X-GitHub-Event`,
`X-Gitlab-Event`, and Sentry hook headers, then maps them through
`gateway.provider_routes`:

```yaml
gateway:
  enabled: true
  public_path: "/gateway"
  provider_routes:
    github: "github-fanout"
    gitlab: "gitlab"
    sentry: "sentry"
```

You can still force a route for debugging or custom providers:

```bash
curl -X POST "http://127.0.0.1:8080/gateway?route=github" \
  -H "Content-Type: application/json" \
  -d '{"repository":{"full_name":"loonghao/webhook_bridge"}}'
```

## Script Fanout

Routes can run local scripts directly. This lets one GitHub delivery fan out to
multiple local actions, for example writing the raw webhook JSON with
PowerShell and sending a short notification with Python:

```yaml
gateway:
  provider_routes:
    github: "github-fanout"

scripts:
  groups:
    - name: "github-fanout"
      mode: "parallel"
      routes:
        - "github-save-json"
        - "github-wechat"
  routes:
    - name: "github-save-json"
      shell: "powershell"
      script_path: "scripts/save-github-hook.ps1"
      env:
        WEBHOOK_BRIDGE_OUTPUT_DIR: "data/github-hooks"
    - name: "github-wechat"
      shell: "python"
      script_path: "scripts/notify-wechat.py"
      env:
        WEBHOOK_BRIDGE_WECHAT_WEBHOOK_URL: "https://qyapi.weixin.qq.com/cgi-bin/webhook/send?key=00000000-0000-0000-0000-000000000000"
```

Use a real webhook URL only in local config or environment variables. Do not
commit production webhook keys.

## Python Hook Example

```python
from webhook_bridge.plugin import BasePlugin


class Plugin(BasePlugin):
    def post(self):
        return {
            "status": "success",
            "payload": self.data,
        }
```

Send a webhook:

```bash
curl -X POST http://127.0.0.1:8080/gateway?route=test_plugin \
  -H "Content-Type: application/json" \
  -d '{"event":"hello"}'
```

The executor can run under `uv` by default:

```yaml
python:
  environment_manager: "uv"
  uv_project_dir: "."
```

If `uv` is not installed, local development falls back to the configured Python
interpreter when allowed.

## Development

```bash
cargo test
cd web-nextjs
npm install
npm run dev
```

The release build emits one executable named `webhook-bridge`. It embeds the
Python executor runtime and materializes it locally when workers start.

## Release

Releases are managed by `googleapis/release-please-action`. When the
release-please PR is merged to `main`, CI builds and uploads platform-specific
single executable assets to the GitHub Release.

See [Webhook Bridge 4.0 Architecture](docs/WEBHOOK_BRIDGE_4_ARCHITECTURE.md).
