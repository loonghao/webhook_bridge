"""Test plugin for webhook bridge."""
# Import local modules
from webhook_bridge.plugin import BasePlugin


class Plugin(BasePlugin):
    def handle(self) -> dict:
        """Generic handler for the plugin.

        Returns:
            dict: Dictionary containing the plugin's results
        """
        return {"status": "success", "message": f"Test plugin executed with {self.http_method} method"}
