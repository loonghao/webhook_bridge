"""CLI tests."""
# Import local modules
from webhook_bridge import cli


def test_cli_import():
    """Test that CLI module can be imported."""
    assert cli


def test_cli_has_main():
    """Test that CLI module has main function."""
    assert hasattr(cli, 'main')


