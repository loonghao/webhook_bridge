"""NOX actions for web-related tasks.

This module provides NOX sessions for running and testing the web server.
"""
# Import future modules
from __future__ import annotations

# Import built-in modules
from pathlib import Path
import webbrowser

# Import third-party modules
import nox


@nox.session
def start_server(session: nox.Session) -> None:
    """Start the webhook bridge server for development."""
    # Install dependencies
    session.install("uvicorn")
    session.install("-e", ".")

    # Create plugin directory if it doesn't exist
    plugin_dir = Path("example_plugins")
    plugin_dir.mkdir(parents=True, exist_ok=True)

    # Create a test plugin
    test_plugin = plugin_dir / "test_plugin.py"
    test_plugin.write_text('''
from webhook_bridge.plugin import BasePlugin

class Plugin(BasePlugin):
    def run(self) -> dict:
        return {"status": "success", "message": "Test plugin executed"}
    ''')

    # Open API documentation in browser
    host = "127.0.0.1"
    port = "54012"
    webbrowser.open_new_tab(f"http://{host}:{port}")

    # Start server
    session.run(
        "python",
        "-m",
        "webhook_bridge.cli",
        "--host",
        host,
        "--port",
        port,
        "--plugin-dir",
        str(plugin_dir),
        "--log-level",
        "DEBUG",
    )
