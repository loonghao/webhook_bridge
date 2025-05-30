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
from webhook_bridge.cli import ServerConfig
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
    plugin_dir = Path("/plugins").absolute()
    config = ServerConfig(
        host="localhost",
        port=9000,
        plugin_dir=plugin_dir,
        log_level="DEBUG",
        kwargs={
            "title": "Custom API",
            "description": "Custom Description",
        },
    )
    run_server(config)

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
        "--workers", "2",
        "--worker-class", "uvicorn.workers.UvicornH11Worker",
        "--reload",
        "--no-access-log",
        "--no-use-colors",
        "--timeout-keep-alive", "10",
    ]

    with patch("sys.argv", ["webhook-bridge"] + test_args):
        with pytest.raises(SystemExit) as exc_info:
            main()
        assert exc_info.value.code == 0


@patch("webhook_bridge.cli.uvicorn.run")
def test_run_server_with_workers(mock_run: MagicMock) -> None:
    """Test server running with multiple workers."""
    plugin_dir = Path("/plugins").absolute()
    config = ServerConfig(
        host="localhost",
        port=9000,
        plugin_dir=plugin_dir,
        log_level="DEBUG",
        workers=2,
        worker_class="uvicorn.workers.UvicornH11Worker",
        reload=False,
        access_log=False,
        no_access_log=True,
        use_colors=False,
        no_use_colors=True,
        timeout_keep_alive=10,
    )
    run_server(config)

    mock_run.assert_called_once()
    call_args = mock_run.call_args[1]
    assert call_args["workers"] == 2
    assert call_args["access_log"] is False
    assert call_args["use_colors"] is False
    assert call_args["timeout_keep_alive"] == 10


@patch("webhook_bridge.cli.uvicorn.run")
def test_run_server_with_ssl(mock_run: MagicMock) -> None:
    """Test server running with SSL configuration."""
    plugin_dir = Path("/plugins").absolute()
    ssl_keyfile = Path("/path/to/key.pem")
    ssl_certfile = Path("/path/to/cert.pem")

    config = ServerConfig(
        host="localhost",
        port=9000,
        plugin_dir=plugin_dir,
        log_level="INFO",
        ssl_keyfile=ssl_keyfile,
        ssl_certfile=ssl_certfile,
    )
    run_server(config)

    mock_run.assert_called_once()
    call_args = mock_run.call_args[1]
    assert call_args["ssl_keyfile"] == str(ssl_keyfile)
    assert call_args["ssl_certfile"] == str(ssl_certfile)


@patch("webhook_bridge.cli.uvicorn.run")
def test_run_server_with_performance_limits(mock_run: MagicMock) -> None:
    """Test server running with performance limits."""
    plugin_dir = Path("/plugins").absolute()
    config = ServerConfig(
        host="localhost",
        port=9000,
        plugin_dir=plugin_dir,
        log_level="INFO",
        limit_concurrency=100,
        limit_max_requests=1000,
    )
    run_server(config)

    mock_run.assert_called_once()
    call_args = mock_run.call_args[1]
    assert call_args["limit_concurrency"] == 100
    assert call_args["limit_max_requests"] == 1000


@patch("webhook_bridge.cli.uvicorn.run")
def test_main_default_values(mock_run: MagicMock) -> None:
    """Test main function with default values."""
    with patch("sys.argv", ["webhook-bridge"]):
        with pytest.raises(SystemExit) as exc_info:
            main()
        assert exc_info.value.code == 0
