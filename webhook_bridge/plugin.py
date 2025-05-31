"""Plugin system for webhook_bridge.

This module provides the base plugin class and utilities for loading
and executing webhook plugins.
"""

import importlib.util
import sys
import traceback
from abc import ABC, abstractmethod
from typing import Any, Dict, Optional


class BasePlugin(ABC):
    """Base class for all webhook plugins."""
    
    def __init__(self, data: Dict[str, Any], http_method: str = "POST"):
        """
        Initialize the plugin.
        
        Args:
            data: The webhook data received.
            http_method: The HTTP method used for the request.
        """
        self.data = data
        self.http_method = http_method.upper()
        self.result = {}
        
    @abstractmethod
    def run(self) -> Dict[str, Any]:
        """
        Execute the plugin logic.
        
        This method must be implemented by all plugins.
        
        Returns:
            Dictionary containing the plugin execution result.
        """
        pass
    
    def execute(self) -> Dict[str, Any]:
        """
        Execute the plugin and format the result.
        
        Returns:
            Formatted execution result.
        """
        try:
            # Call the plugin's run method
            result = self.run()
            
            # Ensure result is a dictionary
            if not isinstance(result, dict):
                result = {"result": result}
            
            # Add metadata
            result.update({
                "plugin_executed": True,
                "http_method": self.http_method,
                "status": "success"
            })
            
            return result
            
        except Exception as e:
            return {
                "plugin_executed": False,
                "error": str(e),
                "traceback": traceback.format_exc(),
                "http_method": self.http_method,
                "status": "error"
            }
    
    def get_data(self, key: str, default: Any = None) -> Any:
        """
        Get data from the webhook payload.
        
        Args:
            key: The key to retrieve.
            default: Default value if key is not found.
            
        Returns:
            The value for the key, or default if not found.
        """
        return self.data.get(key, default)
    
    def set_result(self, key: str, value: Any) -> None:
        """
        Set a result value.
        
        Args:
            key: The result key.
            value: The result value.
        """
        self.result[key] = value
    
    def get_result(self) -> Dict[str, Any]:
        """
        Get the current result dictionary.
        
        Returns:
            The result dictionary.
        """
        return self.result.copy()


def load_plugin(plugin_path: str) -> type:
    """
    Load a plugin from a file path.
    
    Args:
        plugin_path: Path to the plugin file.
        
    Returns:
        The plugin class.
        
    Raises:
        ImportError: If the plugin cannot be loaded.
        AttributeError: If the plugin doesn't have the required class.
    """
    try:
        # Create module spec
        spec = importlib.util.spec_from_file_location("plugin_module", plugin_path)
        if spec is None or spec.loader is None:
            raise ImportError(f"Cannot create module spec for {plugin_path}")
        
        # Load the module
        module = importlib.util.module_from_spec(spec)
        spec.loader.exec_module(module)
        
        # Find the plugin class
        plugin_class = None
        for attr_name in dir(module):
            attr = getattr(module, attr_name)
            if (isinstance(attr, type) and 
                issubclass(attr, BasePlugin) and 
                attr is not BasePlugin):
                plugin_class = attr
                break
        
        if plugin_class is None:
            raise AttributeError(f"No plugin class found in {plugin_path}")
        
        return plugin_class
        
    except Exception as e:
        raise ImportError(f"Failed to load plugin from {plugin_path}: {e}")


def validate_plugin_class(plugin_class: type) -> bool:
    """
    Validate that a class is a proper plugin.
    
    Args:
        plugin_class: The class to validate.
        
    Returns:
        True if the class is a valid plugin, False otherwise.
    """
    try:
        return (
            isinstance(plugin_class, type) and
            issubclass(plugin_class, BasePlugin) and
            plugin_class is not BasePlugin and
            hasattr(plugin_class, 'run') and
            callable(getattr(plugin_class, 'run'))
        )
    except Exception:
        return False


class PluginManager:
    """Manager for loading and executing plugins."""
    
    def __init__(self):
        self._plugin_cache = {}
    
    def load_plugin(self, plugin_path: str, use_cache: bool = True) -> type:
        """
        Load a plugin with optional caching.
        
        Args:
            plugin_path: Path to the plugin file.
            use_cache: Whether to use cached plugins.
            
        Returns:
            The plugin class.
        """
        if use_cache and plugin_path in self._plugin_cache:
            return self._plugin_cache[plugin_path]
        
        plugin_class = load_plugin(plugin_path)
        
        if use_cache:
            self._plugin_cache[plugin_path] = plugin_class
        
        return plugin_class
    
    def clear_cache(self) -> None:
        """Clear the plugin cache."""
        self._plugin_cache.clear()
    
    def execute_plugin(self, plugin_path: str, data: Dict[str, Any], 
                      http_method: str = "POST") -> Dict[str, Any]:
        """
        Execute a plugin.
        
        Args:
            plugin_path: Path to the plugin file.
            data: The webhook data.
            http_method: The HTTP method.
            
        Returns:
            The plugin execution result.
        """
        plugin_class = self.load_plugin(plugin_path)
        plugin_instance = plugin_class(data, http_method)
        return plugin_instance.execute()
