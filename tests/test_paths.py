# Import built-in modules
import os

# Import third-party modules
from webhook_bridge.filesystem import get_default_plugin_path
from webhook_bridge.filesystem import get_plugins


def test_get_default_plugin_path():
    assert get_default_plugin_path()


def test_get_plugins(monkeypatch, test_data_root):
    monkeypatch.setenv("WEBHOOK_BRIDGE_SERVER_PLUGINS",
                       os.path.join(test_data_root, "custom_plugins"))
    assert len(get_plugins().keys()) == 2
