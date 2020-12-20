import json

from webhook_bridge.plugin import BasePlugin


class Plugin(BasePlugin):

    def run(self):
        with open("d:/test.json", "w") as f:
            json.dump(self.data.event.tags, f, indent=2)
