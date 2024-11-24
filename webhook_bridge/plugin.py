"""API endpoint for executing webhook plugins.

This module provides the API endpoint for executing specific webhook plugins
with provided data. It handles plugin loading, execution, and error handling.

Example:
    >>> # Import plugin class
    >>> from webhook_bridge.plugin import load_plugin
    >>> # Load and execute plugin
    >>> plugin_class = load_plugin('example_plugin.py')
    >>> plugin = plugin_class({"key": "value"})
    >>> result = plugin.run()
"""
# Import future modules
from __future__ import annotations

# Import built-in modules
from abc import ABC
from abc import abstractmethod
import importlib.machinery
import importlib.util
import logging
from pathlib import Path
from typing import Any
from typing import TypeVar


T = TypeVar("T", bound="BasePlugin")


class BasePlugin(ABC):
    """Abstract base class for all webhook bridge plugins.

    This class defines the interface that all webhook bridge plugins must implement.
    Each plugin must provide a `run` method that processes the input data and
    returns a result.

    Attributes:
        data: Dictionary containing the plugin's input data
        logger: Logger instance for the plugin
    """

    def __init__(
        self,
        data: dict[str, Any],
        logger: logging.Logger | None = None,
    ) -> None:
        """Initialize the plugin with data and logger.

        Args:
            data: Dictionary containing the plugin's input data
            logger: Optional logger instance. If None, a new logger will be created
        """
        self.data = data
        if logger is None:
            logger = logging.getLogger(self.__class__.__name__)
        self.logger = logger

    @abstractmethod
    def run(self) -> dict[str, Any]:
        """Execute the plugin's main functionality.

        This method must be implemented by all plugin classes. It should process
        the plugin's input data (stored in self.data) and return a dictionary
        containing the results.

        Returns:
            dict: Dictionary containing the plugin's results
        """
        raise NotImplementedError("Plugin must implement run() method")

    def execute(self) -> dict[str, Any]:
        data = self.run() or {}
        result = {
            "input_data": self.data,
            "additional_info": "This is some additional information.",
            "result": data,
        }
        return result

def load_plugin(pyfile: str) -> type[BasePlugin]:
    """Load a plugin class from a Python file.

    This function dynamically loads a Python file and returns the plugin class
    defined within it. The plugin class must inherit from BasePlugin.

    Args:
        pyfile: Path to the Python file containing the plugin class

    Returns:
        type[BasePlugin]: The plugin class type

    Raises:
        AttributeError: If no plugin class is found in the file
        ImportError: If there is an error loading the plugin file

    Example:
        >>> plugin_class = load_plugin('example_plugin.py')
        >>> plugin = plugin_class({"key": "value"})
        >>> result = plugin.run()
    """
    name = Path(pyfile).stem
    loader = importlib.machinery.SourceFileLoader(name, pyfile)
    spec = importlib.util.spec_from_loader(loader.name, loader)
    if spec is None:
        raise ImportError(f"Failed to create module spec for {pyfile}")

    module = importlib.util.module_from_spec(spec)
    loader.exec_module(module)

    # Find plugin class
    plugin_class = None
    for _, obj in module.__dict__.items():
        if (
            isinstance(obj, type)
            and issubclass(obj, BasePlugin)
            and obj is not BasePlugin
        ):
            plugin_class = obj
            break

    if plugin_class is None:
        raise AttributeError(
            f"No Plugin class found in {pyfile}. "
            "Make sure the module defines a class that inherits from BasePlugin.",
        )

    return plugin_class
