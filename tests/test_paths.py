# Import local modules
from webhook_bridge import __version__


def test_version_exists():
    """Test that version is defined."""
    assert __version__
    assert isinstance(__version__, str)


def test_package_import():
    """Test that the package can be imported."""
    # Import local modules
    import webhook_bridge
    assert webhook_bridge


def test_cli_import():
    """Test that CLI module can be imported."""
    # Import local modules
    from webhook_bridge import cli
    assert cli


def test_project_structure():
    """Test that project structure is correct."""
    import os
    from pathlib import Path

    project_root = Path(__file__).parent.parent

    # Check that essential directories exist
    assert (project_root / "cmd").exists()
    assert (project_root / "internal").exists()
    assert (project_root / "python_executor").exists()
    assert (project_root / "webhook_bridge").exists()

    # Check that essential files exist
    assert (project_root / "go.mod").exists()
    assert (project_root / "pyproject.toml").exists()
    assert (project_root / "webhook_bridge" / "__init__.py").exists()
    assert (project_root / "webhook_bridge" / "cli.py").exists()
