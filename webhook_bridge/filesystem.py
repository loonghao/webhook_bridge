# Import built-in modules
from functools import lru_cache
import glob
import os
from typing import List


def get_default_plugin_path() -> str:
    return os.path.join(os.path.dirname(__file__), "plugins")


def get_plugin_paths() -> List[str]:
    try:
        paths = os.getenv("WEBHOOK_BRIDGE_SERVER_PLUGINS", "").split(os.pathsep)
    except AttributeError:
        paths = []
    paths.insert(0, get_default_plugin_path())
    return paths


@lru_cache
def get_plugins() -> dict[str, str]:
    plugins = {}

    def _get_plugin_name(file_name: str) -> str:
        return os.path.basename(file_name).split(".py")[0]

    for plugin in get_plugin_paths():
        data = {
            _get_plugin_name(p): p for p in
            glob.glob(os.path.join(plugin, "*.py"))
            if "__init__" not in p
        }
        plugins.update(data)

    return plugins
