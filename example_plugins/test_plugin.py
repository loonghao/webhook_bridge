
# Import local modules
from webhook_bridge.plugin import BasePlugin


class Plugin(BasePlugin):
    def run(self) -> dict:
        return {"status": "success", "message": "Test plugin executed"}
