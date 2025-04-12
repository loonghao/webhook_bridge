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
    returns a result. Additionally, plugins can implement method-specific handlers
    for different HTTP methods (get, post, put, delete).

    Attributes:
        data: Dictionary containing the plugin's input data
        logger: Logger instance for the plugin
        http_method: The HTTP method used to call the plugin (GET, POST, PUT, DELETE)
    """

    def __init__(
        self,
        data: dict[str, Any],
        logger: logging.Logger | None = None,
        http_method: str = "POST",
    ) -> None:
        """Initialize the plugin with data and logger.

        Args:
            data: Dictionary containing the plugin's input data
            logger: Optional logger instance. If None, a new logger will be created
            http_method: The HTTP method used to call the plugin (GET, POST, PUT, DELETE)
        """
        self.data = data
        if logger is None:
            logger = logging.getLogger(self.__class__.__name__)
        self.logger = logger
        self.http_method = http_method.upper()

    def run(self) -> dict[str, Any]:
        """Execute the plugin's main functionality.

        This method is deprecated and will be removed in a future version.
        Please implement the `handle` method instead, or method-specific handlers
        (`get`, `post`, `put`, `delete`) for more control.

        Returns:
            dict: Dictionary containing the plugin's results
        """
        # Log a deprecation warning
        self.logger.warning(
            "The 'run' method is deprecated and will be removed in a future version. "
            "Please implement the 'handle' method instead, or method-specific handlers "
            "('get', 'post', 'put', 'delete') for more control."
        )

        # If the plugin has implemented the run method, use it
        if hasattr(self.__class__, "run") and self.__class__.run is not BasePlugin.run:
            return self._legacy_run()

        # Otherwise, dispatch to the appropriate method handler
        if self.http_method == "GET" and hasattr(self, "get"):
            return self.get()
        elif self.http_method == "POST" and hasattr(self, "post"):
            return self.post()
        elif self.http_method == "PUT" and hasattr(self, "put"):
            return self.put()
        elif self.http_method == "DELETE" and hasattr(self, "delete"):
            return self.delete()
        else:
            # Fall back to the generic handler
            return self.handle()

    def _legacy_run(self) -> dict[str, Any]:
        """Legacy implementation for plugins that override the run method.

        This is a compatibility method for plugins that still override the run method.
        It will be removed in a future version.

        Returns:
            dict: Dictionary containing the plugin's results
        """
        # Call the subclass implementation of run
        # We need to use super().__class__ to get the parent class method
        # and avoid infinite recursion
        method = super().__getattribute__("run")
        return method()

    @abstractmethod
    def handle(self) -> dict[str, Any]:
        """Generic handler for the plugin's functionality.

        This method must be implemented by all plugin classes. It should process
        the plugin's input data (stored in self.data) and return a dictionary
        containing the results.

        Returns:
            dict: Dictionary containing the plugin's results
        """
        raise NotImplementedError("Plugin must implement handle() method")

    def get(self) -> dict[str, Any]:
        """Handle GET requests.

        By default, this method calls the generic handler. Override this method
        to provide GET-specific functionality.

        Returns:
            dict: Dictionary containing the plugin's results
        """
        return self.handle()

    def post(self) -> dict[str, Any]:
        """Handle POST requests.

        By default, this method calls the generic handler. Override this method
        to provide POST-specific functionality.

        Returns:
            dict: Dictionary containing the plugin's results
        """
        return self.handle()

    def put(self) -> dict[str, Any]:
        """Handle PUT requests.

        By default, this method calls the generic handler. Override this method
        to provide PUT-specific functionality.

        Returns:
            dict: Dictionary containing the plugin's results
        """
        return self.handle()

    def delete(self) -> dict[str, Any]:
        """Handle DELETE requests.

        By default, this method calls the generic handler. Override this method
        to provide DELETE-specific functionality.

        Returns:
            dict: Dictionary containing the plugin's results
        """
        return self.handle()

    def execute(self) -> dict[str, Any]:
        """Execute the plugin and format the results.

        This method calls the run method and formats the results in a standard format.

        Returns:
            dict: Dictionary containing the formatted plugin results
        """
        data = self.run() or {}
        result = {
            "input_data": self.data,
            "additional_info": "This is some additional information.",
            "http_method": self.http_method,
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
