"""Test cases for basic functionality."""
import sys
from pathlib import Path

# Add project root to Python path
project_root = Path(__file__).parent.parent
sys.path.insert(0, str(project_root))


def test_webhook_bridge_import():
    """Test that webhook_bridge package can be imported."""
    import webhook_bridge

    assert webhook_bridge
    assert hasattr(webhook_bridge, "__version__")


def test_python_executor_exists():
    """Test that python_executor directory exists."""
    python_executor_dir = project_root / "python_executor"
    assert python_executor_dir.exists()
    assert (python_executor_dir / "__init__.py").exists()
    assert (python_executor_dir / "main.py").exists()


def test_project_structure():
    """Test basic project structure."""
    assert (project_root / "Cargo.toml").exists()
    assert (project_root / "crates" / "bridge-core").exists()
    assert (project_root / "crates" / "bridge-server").exists()
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
