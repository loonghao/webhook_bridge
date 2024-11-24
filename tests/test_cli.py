"""CLI tests."""
# Import future modules
from __future__ import annotations

# Import built-in modules
from pathlib import Path
from unittest.mock import MagicMock
from unittest.mock import patch

# Import third-party modules
import pytest

# Import local modules
from webhook_bridge.cli import create_parser
from webhook_bridge.cli import main
from webhook_bridge.cli import run_server


def test_create_parser() -> None:
    """Test argument parser creation."""
    parser = create_parser()
    args = parser.parse_args([])

    assert args.host == "0.0.0.0"
    assert args.port == 8000
    assert str(args.plugin_dir) == str(Path("plugins").absolute())
    assert args.log_level == "INFO"
    assert args.title == "Webhook Bridge API"
    assert args.description == "A flexible webhook integration platform"


def test_create_parser_custom_values() -> None:
    """Test argument parser with custom values."""
    parser = create_parser()
    plugin_dir = str(Path("/plugins").absolute())
    args = parser.parse_args([
        "--host", "localhost",
        "--port", "9000",
        "--plugin-dir", plugin_dir,
        "--log-level", "DEBUG",
        "--title", "Custom API",
        "--description", "Custom Description",
    ])

    assert args.host == "localhost"
    assert args.port == 9000
    assert str(args.plugin_dir) == plugin_dir
    assert args.log_level == "DEBUG"
    assert args.title == "Custom API"
    assert args.description == "Custom Description"


@patch("webhook_bridge.cli.uvicorn.run")
def test_run_server_with_api_config(mock_run: MagicMock) -> None:
    """Test server running with API configuration."""
    plugin_dir = str(Path("/plugins").absolute())
    run_server(
        host="localhost",
        port=9000,
        plugin_dir=plugin_dir,
        log_level="DEBUG",
        title="Custom API",
        description="Custom Description",
    )

    mock_run.assert_called_once()


@patch("webhook_bridge.cli.uvicorn.run")
def test_main_with_all_options(mock_run: MagicMock) -> None:
    """Test main function with all options."""
    plugin_dir = str(Path("/plugins").absolute())
    test_args = [
        "--host", "localhost",
        "--port", "9000",
        "--plugin-dir", plugin_dir,
        "--log-level", "DEBUG",
        "--title", "Custom API",
        "--description", "Custom Description",
    ]

    with patch("sys.argv", ["webhook-bridge"] + test_args):
        with pytest.raises(SystemExit) as exc_info:
            main()
        assert exc_info.value.code == 0


@patch("webhook_bridge.cli.uvicorn.run")
def test_main_default_values(mock_run: MagicMock) -> None:
    """Test main function with default values."""
    with patch("sys.argv", ["webhook-bridge"]):
        with pytest.raises(SystemExit) as exc_info:
            main()
        assert exc_info.value.code == 0
