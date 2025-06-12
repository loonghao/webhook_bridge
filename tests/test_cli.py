"""CLI tests."""
# Import built-in modules
from pathlib import Path
import subprocess
import sys


try:
    # Import local modules
    from webhook_bridge import cli
except ImportError:
    # Handle case where package is not installed
    cli = None


def test_cli_import():
    """Test that CLI module can be imported."""
    if cli is None:
        # Check that the file exists at least
        project_root = Path(__file__).parent.parent
        cli_file = project_root / "webhook_bridge" / "cli.py"
        assert cli_file.exists(), "CLI file not found"
    else:
        assert cli


def test_cli_has_main():
    """Test that CLI module has main function."""
    if cli is not None:
        assert hasattr(cli, 'main')
        assert callable(cli.main)


def test_go_binary_builds():
    """Test that Go binary can be built."""
    project_root = Path(__file__).parent.parent

    # Try to build the binary
    result = subprocess.run(
        ["go", "build", "-o", "test-webhook-bridge", "./cmd/webhook-bridge"],
        cwd=project_root,
        capture_output=True,
        text=True, check=False,
    )

    # Clean up
    test_binary = project_root / "test-webhook-bridge"
    if test_binary.exists():
        test_binary.unlink()

    assert result.returncode == 0, f"Go build failed: {result.stderr}"


def test_python_executor_structure():
    """Test Python executor structure."""
    project_root = Path(__file__).parent.parent
    python_executor_dir = project_root / "python_executor"

    assert python_executor_dir.exists(), "python_executor directory not found"
    assert (python_executor_dir / "__init__.py").exists(), "python_executor/__init__.py not found"
    assert (python_executor_dir / "main.py").exists(), "python_executor/main.py not found"


