"""Filesystem utilities for plugin management.

This module provides utilities for managing plugin paths and discovering plugin
files in the webhook bridge system. It handles both default and custom plugin
locations specified through environment variables.

Example:
    >>> from webhook_bridge.filesystem import get_plugins
    >>> plugins = get_plugins()
    >>> print(plugins)
    {'example_plugin': '/path/to/example_plugin.py', ...}
"""
# Import future modules
from __future__ import annotations

# Import built-in modules
from functools import lru_cache
import glob
import logging
import os
from pathlib import Path
from typing import Any
from typing import Dict


def get_default_plugin_path() -> Path:
    """Get the default path where plugins are stored.

    Returns:
        Path: Absolute path to the default plugins directory.
    """
    return Path(os.path.dirname(__file__)) / "plugins"


def get_plugin_paths(extra_path: str | Path | None = None) -> list[str]:
    """Get all plugin paths including custom paths from environment variable.

    The function reads the WEBHOOK_BRIDGE_SERVER_PLUGINS environment variable
    which can contain multiple paths separated by the system path separator.
    The default plugin path is always included as the first path.

    Args:
        extra_path: Additional plugin path to include (e.g., from app.state.plugin_dir)

    Returns:
        List[str]: List of paths where plugins can be found.
    """
    logger = logging.getLogger(__name__)
    paths = []

    # Add environment variable paths
    try:
        env_paths = os.getenv("WEBHOOK_BRIDGE_SERVER_PLUGINS", "")
        if env_paths:
            paths.extend(env_paths.split(os.pathsep))
            logger.info("Added plugin paths from environment: %s", paths)
    except AttributeError:
        logger.warning("Failed to get plugin paths from environment")

    # Add default path
    default_path = str(get_default_plugin_path())
    paths.insert(0, default_path)
    logger.info("Added default plugin path: %s", default_path)

    # Add extra path if provided
    if extra_path:
        extra_path_str = str(extra_path)
        if extra_path_str not in paths:
            paths.append(extra_path_str)
            logger.info("Added extra plugin path: %s", extra_path_str)

    return paths


@lru_cache
def get_plugins(extra_path: str | Path | None = None) -> Dict[str, Any]:
    """Get a dictionary of available plugins and their paths.

    This function searches for Python files in all configured plugin directories
    and returns a mapping of plugin names to their absolute paths.

    Args:
        extra_path: Additional plugin path to include (e.g., from app.state.plugin_dir)

    Returns:
        Dict[str, str]: Dictionary mapping plugin names to their absolute paths.
    """
    logger = logging.getLogger(__name__)
    plugins = {}
    logger.info("Extra path: %s", extra_path)

    # Get all plugin paths
    paths = get_plugin_paths(extra_path)
    logger.info("Searching for plugins in paths: %s", paths)

    # Search for plugins in each path
    for path in paths:
        if not os.path.isdir(path):
            logger.warning("Plugin path not found or not a directory: %s", path)
            continue

        # Find all .py files in the directory
        pattern = os.path.join(path, "*.py")
        for plugin_path in glob.glob(pattern):
            plugin_name = os.path.basename(plugin_path).split(".py")[0]
            if plugin_name != "__init__":
                if plugin_name not in plugins:
                    plugins[plugin_name] = plugin_path
                    logger.debug("Found plugin: %s at %s", plugin_name, plugin_path)

    logger.info("Found %d plugins: %s", len(plugins), list(plugins.keys()))
    return plugins
