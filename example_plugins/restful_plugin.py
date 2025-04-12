"""Example RESTful plugin demonstrating support for different HTTP methods."""
# Import local modules
from webhook_bridge.plugin import BasePlugin


class Plugin(BasePlugin):
    """Example RESTful plugin demonstrating support for different HTTP methods.

    This plugin implements handlers for GET, POST, PUT, and DELETE methods.
    """

    def handle(self) -> dict:
        """Generic handler for the plugin.

        This method is called when no specific method handler is available
        or as a fallback.

        Returns:
            dict: Dictionary containing the plugin's results
        """
        return {
            "status": "success",
            "message": f"Generic handler called with method {self.http_method}",
            "data": self.data,
        }

    def get(self) -> dict:
        """Handle GET requests.

        Returns:
            dict: Dictionary containing the plugin's results
        """
        return {
            "status": "success",
            "message": "GET request processed",
            "data": self.data,
            "resource_id": self.data.get("id", "all"),
        }

    def post(self) -> dict:
        """Handle POST requests.

        Returns:
            dict: Dictionary containing the plugin's results
        """
        return {
            "status": "success",
            "message": "Resource created",
            "data": self.data,
            "resource_id": "new_id_123",
        }

    def put(self) -> dict:
        """Handle PUT requests.

        Returns:
            dict: Dictionary containing the plugin's results
        """
        return {
            "status": "success",
            "message": "Resource updated",
            "data": self.data,
            "resource_id": self.data.get("id", "unknown"),
        }

    def delete(self) -> dict:
        """Handle DELETE requests.

        Returns:
            dict: Dictionary containing the plugin's results
        """
        return {
            "status": "success",
            "message": "Resource deleted",
            "data": self.data,
            "resource_id": self.data.get("id", "unknown"),
        }
