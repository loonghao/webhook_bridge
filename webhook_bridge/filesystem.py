"""Filesystem utilities for webhook_bridge.

This module provides utilities for discovering and managing webhook plugins
in the filesystem.
"""

import os
import glob
from pathlib import Path
from typing import Dict, List


def get_plugins(plugin_dirs: List[str] = None) -> Dict[str, str]:
    """
    Discover webhook plugins in the filesystem.
    
    Args:
        plugin_dirs: List of directories to search for plugins.
                    If None, searches in default locations.
    
    Returns:
        Dictionary mapping plugin names to their file paths.
    """
    plugins = {}
    
    # Default plugin directories
    if plugin_dirs is None:
        plugin_dirs = [
            "plugins",
            "example_plugins",
            "webhook_plugins",
        ]
    
    for plugin_dir in plugin_dirs:
        if not os.path.exists(plugin_dir):
            continue
            
        # Search for Python files in the plugin directory
        pattern = os.path.join(plugin_dir, "*.py")
        for plugin_file in glob.glob(pattern):
            # Skip __init__.py files
            if os.path.basename(plugin_file) == "__init__.py":
                continue
                
            # Extract plugin name from filename
            plugin_name = os.path.splitext(os.path.basename(plugin_file))[0]
            plugins[plugin_name] = plugin_file
    
    return plugins


def get_plugin_directories() -> List[str]:
    """
    Get list of available plugin directories.
    
    Returns:
        List of directory paths that contain plugins.
    """
    directories = []
    
    default_dirs = [
        "plugins",
        "example_plugins", 
        "webhook_plugins",
    ]
    
    for directory in default_dirs:
        if os.path.exists(directory) and os.path.isdir(directory):
            directories.append(directory)
    
    return directories


def validate_plugin_file(plugin_path: str) -> bool:
    """
    Validate that a plugin file exists and is readable.
    
    Args:
        plugin_path: Path to the plugin file.
        
    Returns:
        True if the plugin file is valid, False otherwise.
    """
    try:
        return (
            os.path.exists(plugin_path) and
            os.path.isfile(plugin_path) and
            os.access(plugin_path, os.R_OK) and
            plugin_path.endswith('.py')
        )
    except Exception:
        return False


def create_plugin_directory(directory: str) -> bool:
    """
    Create a plugin directory if it doesn't exist.
    
    Args:
        directory: Directory path to create.
        
    Returns:
        True if directory was created or already exists, False on error.
    """
    try:
        Path(directory).mkdir(parents=True, exist_ok=True)
        return True
    except Exception:
        return False
