import os

from webhook_bridge import paths


def test_get_default_plugin_path():
    assert paths.get_default_plugin_path()


def test_get_plugins(monkeypatch, test_data_root):
    monkeypatch.setenv("WEBHOOK_BRIDGE_SERVER_PLUGINS",
                       os.path.join(test_data_root, "custom_plugins"))
    print(paths.get_plugins().keys())
    assert len(paths.get_plugins().keys()) == 2
