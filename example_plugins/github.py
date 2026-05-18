"""GitHub webhook hook example for Webhook Bridge 4.0."""
from __future__ import annotations

from webhook_bridge.plugin import BasePlugin


class Plugin(BasePlugin):
    """Handle common GitHub webhook payloads."""

    def post(self) -> dict:
        repository = self.data.get("repository.full_name") or self.data.get("repository.name") or "unknown"
        action = self.data.get("action")
        ref = self.data.get("ref", "")

        if ref.startswith("refs/heads/"):
            action = action or "push"
            branch = ref.removeprefix("refs/heads/")
        else:
            branch = self.data.get("branch", "")

        return {
            "status": "success",
            "provider": "github",
            "repository": repository,
            "action": action or "webhook",
            "branch": branch,
            "sender": self.data.get("sender.login", "unknown"),
        }
