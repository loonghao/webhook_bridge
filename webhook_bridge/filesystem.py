"""Plugin discovery helpers for Python webhook hooks."""
from __future__ import annotations

import os
from pathlib import Path


DEFAULT_PLUGIN_DIRS = ("plugins", "example_plugins")


def get_plugins(plugin_dirs: list[str] | tuple[str, ...] | None = None) -> dict[str, str]:
    """Discover hook plugins in the configured directories.

    A plugin is any ``*.py`` file that does not start with ``_``. The plugin name
    is the file stem, matching the historical ``/webhook/<plugin>`` route.
    """
    roots = plugin_dirs or DEFAULT_PLUGIN_DIRS
    plugins: dict[str, str] = {}

    for root in roots:
        path = Path(os.path.expandvars(root)).expanduser()
        if not path.exists() or not path.is_dir():
            continue

        for candidate in path.glob("*.py"):
            if candidate.name.startswith("_"):
                continue
            plugins[candidate.stem] = str(candidate.resolve())

    return plugins
