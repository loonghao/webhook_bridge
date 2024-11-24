# Import built-in modules
import json
import os
from tempfile import mkdtemp

# Import local modules
from webhook_bridge.plugin import BasePlugin


class Plugin(BasePlugin):

    def run(self):
        root = mkdtemp("webhook-bridge")
        with open(os.path.join(root, "info.json"), "w") as f:
            json.dump(self.data, f, indent=2)
        os.startfile(root)
        # Perform some operations with self.data
        result = {
            "message": "Plugin executed successfully",
            "input_data": self.data,
            "additional_info": "This is some additional information.",
        }
        return result
