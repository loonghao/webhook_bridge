"""CLI tests."""
# Import future modules
from __future__ import annotations

# Import built-in modules
from pathlib import Path
from unittest.mock import MagicMock
from unittest.mock import patch

# Import third-party modules
from click.testing import CliRunner

# Import local modules
from webhook_bridge.cli import ServerConfig
from webhook_bridge.cli import main
from webhook_bridge.cli import run_server


def test_cli_default_values() -> None:
    """Test CLI with default values."""
    runner = CliRunner()
    with patch("webhook_bridge.cli.run_server") as mock_run_server:
        result = runner.invoke(main, [])
        assert result.exit_code == 0
        mock_run_server.assert_called_once()

        # Check the config passed to run_server
        config = mock_run_server.call_args[0][0]
        assert config.host == "0.0.0.0"
        assert config.port == 8000
        assert config.log_level == "INFO"
        assert config.title == "Webhook Bridge API"
        assert config.description == "A flexible webhook integration platform"


def test_cli_custom_values() -> None:
    """Test CLI with custom values."""
    runner = CliRunner()
    with runner.isolated_filesystem():
        # Create a temporary plugin directory
        plugin_dir = Path("test_plugins")
        plugin_dir.mkdir()

        with patch("webhook_bridge.cli.run_server") as mock_run_server:
            result = runner.invoke(main, [
                "--host", "localhost",
                "--port", "9000",
                "--plugin-dir", str(plugin_dir),
                "--log-level", "DEBUG",
                "--title", "Custom API",
                "--description", "Custom Description",
                "--workers", "2",
                "--worker-class", "uvicorn.workers.UvicornH11Worker",
                "--reload",
                "--no-access-log",
                "--no-use-colors",
                "--timeout-keep-alive", "10",
            ])
            assert result.exit_code == 0
            mock_run_server.assert_called_once()

            # Check the config passed to run_server
            config = mock_run_server.call_args[0][0]
            assert config.host == "localhost"
            assert config.port == 9000
            assert config.log_level == "DEBUG"
            assert config.title == "Custom API"
            assert config.description == "Custom Description"
            assert config.workers == 2
            assert config.worker_class == "uvicorn.workers.UvicornH11Worker"
            assert config.reload is True
            assert config.access_log is False
            assert config.use_colors is False
            assert config.timeout_keep_alive == 10


@patch("webhook_bridge.cli.uvicorn.run")
def test_run_server_with_api_config(mock_run: MagicMock) -> None:
    """Test server running with API configuration."""
    plugin_dir = Path("/plugins").absolute()
    config = ServerConfig(
        host="localhost",
        port=9000,
        plugin_dir=plugin_dir,
        log_level="DEBUG",
        title="Custom API",
        description="Custom Description",
    )
    run_server(config)

    mock_run.assert_called_once()


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
        use_colors=False,
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


def test_server_config_validation() -> None:
    """Test ServerConfig Pydantic validation."""
    # Test valid config
    config = ServerConfig(
        host="localhost",
        port=8080,
        plugin_dir=Path.cwd() / "test_plugins",
        log_level="DEBUG",
    )
    assert config.host == "localhost"
    assert config.port == 8080
    assert config.log_level == "DEBUG"

    # Test default values
    config_defaults = ServerConfig()
    assert config_defaults.host == "0.0.0.0"
    assert config_defaults.port == 8000
    assert config_defaults.log_level == "INFO"
    assert config_defaults.workers == 1
    assert config_defaults.access_log is True
    assert config_defaults.use_colors is True
