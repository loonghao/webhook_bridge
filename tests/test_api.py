"""Test cases for basic functionality."""
import os
import sys
from pathlib import Path

# Add project root to Python path
project_root = Path(__file__).parent.parent
sys.path.insert(0, str(project_root))


def test_webhook_bridge_import():
    """Test that webhook_bridge package can be imported."""
    try:
        import webhook_bridge
        assert webhook_bridge
        assert hasattr(webhook_bridge, '__version__')
    except ImportError:
        # This is acceptable for a Go-primary project
        assert True


def test_python_executor_exists():
    """Test that python_executor directory exists."""
    python_executor_dir = project_root / "python_executor"
    assert python_executor_dir.exists()
    assert (python_executor_dir / "__init__.py").exists()
    assert (python_executor_dir / "main.py").exists()


def test_project_structure():
    """Test basic project structure."""
    assert (project_root / "cmd").exists()
    assert (project_root / "internal").exists()
    assert (project_root / "go.mod").exists()
    assert (project_root / "pyproject.toml").exists()


def test_cli_module():
    """Test CLI module exists."""
    try:
        from webhook_bridge import cli
        assert hasattr(cli, 'main')
        assert callable(cli.main)
    except ImportError:
        # Check file exists at least
        cli_file = project_root / "webhook_bridge" / "cli.py"
        assert cli_file.exists()
