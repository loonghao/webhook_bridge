"""NOX actions for web-related tasks.

This module provides NOX sessions for running and testing the web server.
"""
# Import future modules
from __future__ import annotations

# Import built-in modules
import os
from pathlib import Path
import subprocess
import sys
import time
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


@nox.session
def build_local(session: nox.Session) -> None:
    """Build the project locally for testing."""
    session.log("🔧 Building webhook-bridge locally...")

    # Build frontend first
    session.log("📦 Building frontend...")
    session.run("go", "run", "dev.go", "dashboard", "install", external=True)
    session.run("go", "run", "dev.go", "dashboard", "build", external=True)

    # Build Go binaries
    session.log("🔨 Building Go binaries...")
    session.run("go", "build", "-o", "webhook-bridge.exe", "./cmd/webhook-bridge", external=True)
    session.run("go", "build", "-o", "webhook-bridge-server.exe", "./cmd/server", external=True)
    session.run("go", "build", "-o", "python-manager.exe", "./cmd/python-manager", external=True)

    session.log("✅ Local build completed!")
    session.log("📁 Binaries created:")
    session.log("   - webhook-bridge.exe")
    session.log("   - webhook-bridge-server.exe")
    session.log("   - python-manager.exe")


@nox.session
def test_local(session: nox.Session) -> None:
    """Test the locally built webhook-bridge."""
    session.log("🧪 Testing locally built webhook-bridge...")

    # Ensure binaries exist
    binaries = ["webhook-bridge.exe", "webhook-bridge-server.exe", "python-manager.exe"]
    for binary in binaries:
        if not Path(binary).exists():
            session.error(f"❌ Binary {binary} not found. Run 'uvx nox -s build-local' first.")
            return

    # Test webhook-bridge CLI
    session.log("🔍 Testing webhook-bridge CLI...")
    session.run("./webhook-bridge.exe", "--version", external=True)

    # Test server binary
    session.log("🔍 Testing webhook-bridge-server...")
    session.run("./webhook-bridge-server.exe", "--version", external=True)

    # Test python-manager
    session.log("🔍 Testing python-manager...")
    session.run("./python-manager.exe", "--version", external=True)

    session.log("✅ All binaries tested successfully!")


@nox.session
def run_local(session: nox.Session) -> None:
    """Run the locally built webhook-bridge server for manual testing."""
    session.log("🚀 Starting locally built webhook-bridge server...")

    # Ensure binary exists
    if not Path("webhook-bridge.exe").exists():
        session.error("❌ webhook-bridge.exe not found. Run 'uvx nox -s build-local' first.")
        return

    # Create test configuration
    config_path = Path("config.test.yaml")
    if not config_path.exists():
        config_content = """
# Test configuration for local webhook-bridge
server:
  host: "127.0.0.1"
  port: 8000
  dashboard_port: 8001

logging:
  level: "debug"
  file: "logs/webhook-bridge.log"

plugins:
  directory: "example_plugins"

python:
  executor_port: 50051
  auto_install: false

dashboard:
  enabled: true
  auto_open: true
"""
        config_path.write_text(config_content.strip())
        session.log(f"📝 Created test configuration: {config_path}")

    # Open dashboard in browser
    dashboard_url = "http://127.0.0.1:8001"
    session.log(f"🌐 Opening dashboard: {dashboard_url}")
    webbrowser.open_new_tab(dashboard_url)

    # Start the server
    session.log("🎯 Starting webhook-bridge server...")
    session.log("   Server: http://127.0.0.1:8000")
    session.log("   Dashboard: http://127.0.0.1:8001")
    session.log("   Press Ctrl+C to stop")

    try:
        session.run("webhook-bridge.exe", "start", "--config", str(config_path), "--verbose", external=True)
    except KeyboardInterrupt:
        session.log("\n⚠️  Server stopped by user")


@nox.session
def dev(session: nox.Session) -> None:
    """Build and run the webhook-bridge server for development (one command)."""
    session.log("🚀 Building and starting webhook-bridge development server...")

    # Build first
    session.log("📦 Building project...")
    session.run("go", "run", "dev.go", "dashboard", "install", external=True)
    session.run("go", "run", "dev.go", "dashboard", "build", external=True)
    session.run("go", "build", "-o", "webhook-bridge.exe", "./cmd/webhook-bridge", external=True)
    session.run("go", "build", "-o", "webhook-bridge-server.exe", "./cmd/server", external=True)
    session.run("go", "build", "-o", "python-manager.exe", "./cmd/python-manager", external=True)

    session.log("✅ Build completed!")

    # Create test configuration
    config_path = Path("config.test.yaml")
    if not config_path.exists():
        config_content = """# Test configuration for local webhook-bridge
server:
  host: "127.0.0.1"
  port: 8000
  dashboard_port: 8001

logging:
  level: "debug"
  file: "logs/webhook-bridge.log"

plugins:
  directory: "example_plugins"

python:
  executor_port: 50051
  auto_install: false

dashboard:
  enabled: true
  auto_open: true
"""
        config_path.write_text(config_content)
        session.log(f"📝 Created test configuration: {config_path}")

    # Create necessary directories
    logs_dir = Path("logs")
    logs_dir.mkdir(exist_ok=True)

    plugins_dir = Path("example_plugins")
    plugins_dir.mkdir(exist_ok=True)

    # Open dashboard in browser
    dashboard_url = "http://127.0.0.1:8001"
    session.log(f"🌐 Opening dashboard: {dashboard_url}")
    webbrowser.open_new_tab(dashboard_url)

    # Start the server
    session.log("🎯 Starting webhook-bridge server...")
    session.log("   API: http://127.0.0.1:8000")
    session.log("   Dashboard: http://127.0.0.1:8001")
    session.log("   Press Ctrl+C to stop")

    try:
        session.run("webhook-bridge.exe", "start", "--config", str(config_path), "--verbose", external=True)
    except KeyboardInterrupt:
        session.log("\n⚠️  Server stopped by user")


@nox.session
def quick(session: nox.Session) -> None:
    """Quick start: build and run server with minimal output."""
    session.log("⚡ Quick start webhook-bridge...")

    # Build binaries only (skip frontend for speed)
    session.run("go", "build", "-o", "webhook-bridge.exe", "./cmd/webhook-bridge", external=True)

    # Create minimal config
    config_path = Path("config.quick.yaml")
    config_content = """server:
  host: "127.0.0.1"
  port: 8000
  dashboard_port: 8001

logging:
  level: "info"
  file: "logs/webhook-bridge.log"

plugins:
  directory: "example_plugins"

python:
  executor_port: 50051
  auto_install: false

dashboard:
  enabled: true
"""
    config_path.write_text(config_content)
    session.log(f"📝 Created configuration: {config_path}")

    # Create logs directory
    logs_dir = Path("logs")
    logs_dir.mkdir(exist_ok=True)

    # Create plugins directory
    plugins_dir = Path("example_plugins")
    plugins_dir.mkdir(exist_ok=True)

    session.log("🚀 Starting server...")
    session.log("   API: http://127.0.0.1:8000")
    session.log("   Dashboard: http://127.0.0.1:8001")
    session.log("   Press Ctrl+C to stop")

    try:
        session.run("webhook-bridge.exe", "start", "--config", str(config_path), "--verbose", external=True)
    except KeyboardInterrupt:
        session.log("\n⚠️  Server stopped")


@nox.session
def clean_local(session: nox.Session) -> None:
    """Clean up locally built binaries and test files."""
    session.log("🧹 Cleaning up local build artifacts...")

    # Remove binaries
    binaries = ["webhook-bridge.exe", "webhook-bridge-server.exe", "python-manager.exe", "unified-server.exe"]
    for binary in binaries:
        binary_path = Path(binary)
        if binary_path.exists():
            binary_path.unlink()
            session.log(f"🗑️  Removed {binary}")

    # Remove test configs
    test_configs = ["config.test.yaml", "config.quick.yaml", "config.dev.yaml", "config.local.yaml"]
    for config in test_configs:
        config_path = Path(config)
        if config_path.exists():
            config_path.unlink()
            session.log(f"🗑️  Removed {config}")

    # Clean Go build cache
    session.run("go", "clean", "-cache", external=True)
    session.log("🗑️  Cleaned Go build cache")

    session.log("✅ Cleanup completed!")


@nox.session
def clean_all(session: nox.Session) -> None:
    """Deep clean: remove all development artifacts and caches."""
    session.log("🧹 Deep cleaning development environment...")

    # Use the platform-specific clean script
    if os.name == "nt":  # Windows
        script_path = Path("scripts/clean-dev.ps1")
        if script_path.exists():
            session.run("powershell", "-ExecutionPolicy", "Bypass", "-File", str(script_path), external=True)
        else:
            session.log("⚠️  Clean script not found, running basic cleanup...")
            # Fallback to basic cleanup
            clean_local(session)
    else:  # Unix-like
        script_path = Path("scripts/clean-dev.sh")
        if script_path.exists():
            session.run("bash", str(script_path), external=True)
        else:
            session.log("⚠️  Clean script not found, running basic cleanup...")
            # Fallback to basic cleanup
            clean_local(session)

    session.log("✅ Deep clean completed!")
