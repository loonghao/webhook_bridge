"""Compatibility plugin API for Python webhook hooks.

Webhook Bridge 4.0 keeps the v0.6.x mental model: a hook is a Python file with
a ``Plugin`` class that inherits ``BasePlugin`` and implements ``handle`` or an
HTTP method-specific handler such as ``post``.
"""
from __future__ import annotations

import importlib.util
from pathlib import Path
from types import ModuleType
from typing import Any, Type


class BasePlugin:
    """Base class for Python hook plugins."""

    def __init__(self, data: dict[str, Any] | None = None, http_method: str = "POST") -> None:
        self.data = data or {}
        self.http_method = http_method.upper()

    def execute(self) -> Any:
        """Route execution to a method-specific handler, then ``handle``."""
        method_handler = getattr(self, self.http_method.lower(), None)
        if callable(method_handler):
            return method_handler()
        return self.handle()

    def handle(self) -> Any:
        """Default hook body. Plugins usually override this."""
        raise NotImplementedError("Plugin must implement handle() or an HTTP method handler")

    def get(self) -> Any:
        return self.handle()

    def post(self) -> Any:
        return self.handle()

    def put(self) -> Any:
        return self.handle()

    def delete(self) -> Any:
        return self.handle()

    def patch(self) -> Any:
        return self.handle()


def load_plugin(plugin_path: str | Path) -> Type[BasePlugin]:
    """Load and return the ``Plugin`` class from a Python file."""
    path = Path(plugin_path).resolve()
    if not path.exists():
        raise FileNotFoundError(f"Plugin file not found: {path}")

    module = _load_module(path)
    plugin_class = getattr(module, "Plugin", None)
    if plugin_class is None:
        raise AttributeError(f"{path} does not define a Plugin class")
    if not issubclass(plugin_class, BasePlugin):
        raise TypeError(f"{path} Plugin must inherit from webhook_bridge.plugin.BasePlugin")
    return plugin_class


def _load_module(path: Path) -> ModuleType:
    module_name = f"webhook_bridge_hook_{path.stem}_{abs(hash(path))}"
    spec = importlib.util.spec_from_file_location(module_name, path)
    if spec is None or spec.loader is None:
        raise ImportError(f"Cannot load plugin module from {path}")

    module = importlib.util.module_from_spec(spec)
    spec.loader.exec_module(module)
    return module
