# Import built-in modules
import glob
import os
from functools import lru_cache


def get_default_plugin_path():
    return os.path.join(os.path.dirname(__file__), "plugins")


def get_plugin_paths():
    try:
        paths = os.getenv("WEBHOOK_BRIDGE_SERVER_PLUGINS").split(os.pathsep)
    except AttributeError:
        paths = []
    paths.insert(0, get_default_plugin_path())
    return paths


@lru_cache()
def get_plugins():
    plugins = {}

    def _get_plugin_name(file_name):
        return os.path.basename(file_name).split(".py")[0]

    for plugin in get_plugin_paths():
        data = {
            _get_plugin_name(p): p for p in
            glob.glob(os.path.join(plugin, "*.py"))
            if "__init__" not in p
        }
        plugins.update(data)

    return plugins
