"""Command-line interface for the webhook bridge."""
# Import future modules
from __future__ import annotations

# Import built-in modules
import logging
import os
from pathlib import Path
import sys
from typing import Any

# Import third-party modules
import click
from pydantic import BaseModel
from pydantic import Field
import uvicorn

# Import local modules
from webhook_bridge.server import create_app


class ServerConfig(BaseModel):
    """Server configuration using Pydantic."""

    host: str = Field(default="0.0.0.0", description="Host to bind the server to")
    port: int = Field(default=8000, description="Port to bind the server to")
    plugin_dir: Path = Field(default_factory=lambda: Path.cwd() / "plugins", description="Plugin directory")
    log_level: str = Field(default="INFO", description="Logging level")

    # Worker configuration
    workers: int = Field(default=1, description="Number of worker processes")
    worker_class: str = Field(default="uvicorn.workers.UvicornWorker", description="Worker class")

    # Development options
    reload: bool = Field(default=False, description="Enable auto-reload")
    reload_dirs: list[str] | None = Field(default=None, description="Directories to watch for reload")

    # Logging options
    access_log: bool = Field(default=True, description="Enable access log")
    use_colors: bool = Field(default=True, description="Use colors in log output")

    # SSL/TLS options
    ssl_keyfile: Path | None = Field(default=None, description="SSL key file path")
    ssl_certfile: Path | None = Field(default=None, description="SSL certificate file path")
    ssl_ca_certs: Path | None = Field(default=None, description="SSL CA certificates file path")

    # Performance options
    limit_concurrency: int | None = Field(default=None, description="Maximum concurrent connections")
    limit_max_requests: int | None = Field(default=None, description="Maximum requests before restart")
    timeout_keep_alive: int = Field(default=5, description="Keep-alive timeout in seconds")

    # API options
    title: str = Field(default="Webhook Bridge API", description="API title")
    description: str = Field(default="A flexible webhook integration platform", description="API description")
    disable_docs: bool = Field(default=False, description="Disable API documentation")


def _configure_logging(log_level: str) -> None:
    """Configure logging for the server."""
    logging.basicConfig(
        level=getattr(logging, log_level.upper()),
        format="%(asctime)s - %(name)s - %(levelname)s - %(message)s",
    )


def _build_uvicorn_config(config: ServerConfig) -> dict[str, Any]:
    """Build uvicorn configuration from server config."""
    uvicorn_config = {
        "host": config.host,
        "port": config.port,
        "log_level": config.log_level.lower(),
        "workers": config.workers,
        "reload": config.reload,
        "access_log": config.access_log,
        "use_colors": config.use_colors,
        "timeout_keep_alive": config.timeout_keep_alive,
    }

    # Add optional configurations
    if config.reload_dirs:
        uvicorn_config["reload_dirs"] = config.reload_dirs

    if config.ssl_keyfile:
        uvicorn_config["ssl_keyfile"] = str(config.ssl_keyfile)

    if config.ssl_certfile:
        uvicorn_config["ssl_certfile"] = str(config.ssl_certfile)

    if config.ssl_ca_certs:
        uvicorn_config["ssl_ca_certs"] = str(config.ssl_ca_certs)

    if config.limit_concurrency is not None:
        uvicorn_config["limit_concurrency"] = config.limit_concurrency

    if config.limit_max_requests is not None:
        uvicorn_config["limit_max_requests"] = config.limit_max_requests

    return uvicorn_config


def _setup_multi_worker_env(config: ServerConfig) -> None:
    """Set up environment variables for multi-worker mode."""
    os.environ["WEBHOOK_BRIDGE_PLUGIN_DIR"] = str(config.plugin_dir)
    os.environ["WEBHOOK_BRIDGE_TITLE"] = config.title
    os.environ["WEBHOOK_BRIDGE_DESCRIPTION"] = config.description
    os.environ["WEBHOOK_BRIDGE_ENABLE_DOCS"] = str(not config.disable_docs)


def run_server(config: ServerConfig) -> None:
    """Run the webhook bridge server.

    Args:
        config: Server configuration object
    """
    # Configure logging
    _configure_logging(config.log_level)

    # Build uvicorn configuration
    uvicorn_config = _build_uvicorn_config(config)

    # For multiple workers, we need to use a different approach
    if config.workers > 1:
        # Set up environment variables for the app factory
        _setup_multi_worker_env(config)
        uvicorn_config["app"] = "webhook_bridge.cli:get_app"
    else:
        # Create FastAPI app for single worker
        app = create_app(
            plugin_dir=str(config.plugin_dir),
            title=config.title,
            description=config.description,
            enable_docs=not config.disable_docs,
        )
        uvicorn_config["app"] = app

    # Run server
    uvicorn.run(**uvicorn_config)


def get_app() -> Any:
    """Get the FastAPI app instance for multi-worker mode."""
    plugin_dir = os.environ.get("WEBHOOK_BRIDGE_PLUGIN_DIR", str(Path.cwd() / "plugins"))
    title = os.environ.get("WEBHOOK_BRIDGE_TITLE", "Webhook Bridge API")
    description = os.environ.get(
        "WEBHOOK_BRIDGE_DESCRIPTION",
        "A flexible webhook integration platform",
    )
    enable_docs = os.environ.get("WEBHOOK_BRIDGE_ENABLE_DOCS", "True").lower() in (
        "true",
        "1",
        "yes",
    )

    return create_app(
        plugin_dir=plugin_dir,
        title=title,
        description=description,
        enable_docs=enable_docs,
    )


# Click CLI implementation
@click.command()
@click.option("--host", default="0.0.0.0", envvar="WEBHOOK_BRIDGE_HOST", help="Host to bind the server to")
@click.option("--port", default=8000, envvar="WEBHOOK_BRIDGE_PORT", help="Port to bind the server to")
@click.option("--plugin-dir", type=click.Path(file_okay=False, dir_okay=True, path_type=Path),
              default=Path.cwd() / "plugins", help="Directory containing webhook plugins")
@click.option("--log-level", type=click.Choice(["DEBUG", "INFO", "WARNING", "ERROR", "CRITICAL"]),
              default="INFO", envvar="WEBHOOK_BRIDGE_LOG_LEVEL", help="Logging level")
@click.option("--workers", default=1, envvar="WEBHOOK_BRIDGE_WORKERS", help="Number of worker processes")
@click.option("--worker-class", default="uvicorn.workers.UvicornWorker", envvar="WEBHOOK_BRIDGE_WORKER_CLASS",
              help="Worker class to use")
@click.option("--reload", is_flag=True, envvar="WEBHOOK_BRIDGE_RELOAD", help="Enable auto-reload for development")
@click.option("--reload-dirs", multiple=True, help="Directories to watch for reload")
@click.option("--access-log/--no-access-log", default=True, help="Enable/disable access log")
@click.option("--use-colors/--no-use-colors", default=True, help="Enable/disable colors in log output")
@click.option("--ssl-keyfile", type=click.Path(path_type=Path), help="SSL key file path")
@click.option("--ssl-certfile", type=click.Path(path_type=Path), help="SSL certificate file path")
@click.option("--ssl-ca-certs", type=click.Path(path_type=Path), help="SSL CA certificates file path")
@click.option("--limit-concurrency", type=int, help="Maximum number of concurrent connections")
@click.option("--limit-max-requests", type=int, help="Maximum number of requests before restarting worker")
@click.option("--timeout-keep-alive", default=5, help="Keep-alive timeout in seconds")
@click.option("--title", default="Webhook Bridge API", help="API title")
@click.option("--description", default="A flexible webhook integration platform", help="API description")
@click.option("--disable-docs", is_flag=True, help="Disable the API documentation endpoints")
def main(
    host: str,
    port: int,
    plugin_dir: Path,
    log_level: str,
    workers: int,
    worker_class: str,
    reload: bool,
    reload_dirs: tuple[str, ...],
    access_log: bool,
    use_colors: bool,
    ssl_keyfile: Path | None,
    ssl_certfile: Path | None,
    ssl_ca_certs: Path | None,
    limit_concurrency: int | None,
    limit_max_requests: int | None,
    timeout_keep_alive: int,
    title: str,
    description: str,
    disable_docs: bool,
) -> None:
    """Start the webhook bridge server."""
    try:
        config = ServerConfig(
            host=host,
            port=port,
            plugin_dir=plugin_dir,
            log_level=log_level,
            workers=workers,
            worker_class=worker_class,
            reload=reload,
            reload_dirs=list(reload_dirs) if reload_dirs else None,
            access_log=access_log,
            use_colors=use_colors,
            ssl_keyfile=ssl_keyfile,
            ssl_certfile=ssl_certfile,
            ssl_ca_certs=ssl_ca_certs,
            limit_concurrency=limit_concurrency,
            limit_max_requests=limit_max_requests,
            timeout_keep_alive=timeout_keep_alive,
            title=title,
            description=description,
            disable_docs=disable_docs,
        )
        run_server(config)
    except Exception as e:
        click.echo(f"Error running server: {e}", err=True)
        sys.exit(1)


if __name__ == "__main__":
    main()
