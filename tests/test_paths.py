# Import local modules
from webhook_bridge import __version__


def test_version_exists():
    """Test that version is defined."""
    assert __version__
    assert isinstance(__version__, str)


def test_package_import():
    """Test that the package can be imported."""
    import webhook_bridge
    assert webhook_bridge


def test_cli_import():
    """Test that CLI module can be imported."""
    from webhook_bridge import cli
    assert cli


def test_manager_import():
    """Test that manager module can be imported."""
    from webhook_bridge import manager
    assert manager
