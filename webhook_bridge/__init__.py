"""
Webhook Bridge - A flexible webhook integration platform with hybrid Go/Python architecture.

This Python package provides a CLI tool to download, manage, and run the Go-based
webhook bridge server.
"""

__version__ = "1.0.0"
__author__ = "hal.long <hal.long@outlook.com>"
__description__ = "A flexible webhook integration platform with hybrid Go/Python architecture"

# Re-export main components
from .cli import main
from .manager import WebhookBridgeManager

__all__ = ["main", "WebhookBridgeManager"]
