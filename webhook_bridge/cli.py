"""Command-line interface for the webhook bridge."""
# Import future modules
from __future__ import annotations

# Import built-in modules
import argparse
import logging
import os
from pathlib import Path
import sys
from typing import Any
from typing import Sequence

# Import third-party modules
import uvicorn

# Import local modules
from webhook_bridge.server import create_app


def validate_url_path(value: str) -> str:
    """Validate URL path.

    Args:
        value: URL path to validate

    Returns:
        str: Validated URL path

    Raises:
        argparse.ArgumentTypeError: If URL path is invalid
    """
    if not value:
        raise argparse.ArgumentTypeError("URL path cannot be empty")
    if not value.startswith("/"):
        raise argparse.ArgumentTypeError("URL path must start with /")
    return value


def create_parser() -> argparse.ArgumentParser:
    """Create the argument parser.

    Returns:
        argparse.ArgumentParser: The configured argument parser
    """
    parser = argparse.ArgumentParser(
        description="Start the webhook bridge server.",
    )

    # Server configuration
    parser.add_argument(
        "--host",
        default=os.environ.get("WEBHOOK_BRIDGE_HOST", "0.0.0.0"),
        help="Host to bind the server to (default: 0.0.0.0)",
    )
    parser.add_argument(
        "--port",
        type=int,
        default=int(os.environ.get("WEBHOOK_BRIDGE_PORT", "8000")),
        help="Port to bind the server to (default: 8000)",
    )
    parser.add_argument(
        "--log-level",
        choices=["DEBUG", "INFO", "WARNING", "ERROR", "CRITICAL"],
        default=os.environ.get("WEBHOOK_BRIDGE_LOG_LEVEL", "INFO"),
        help="Logging level (default: INFO)",
    )

    # Worker configuration
    parser.add_argument(
        "--workers",
        type=int,
        default=int(os.environ.get("WEBHOOK_BRIDGE_WORKERS", "1")),
        help="Number of worker processes (default: 1)",
    )
    parser.add_argument(
        "--worker-class",
        default=os.environ.get("WEBHOOK_BRIDGE_WORKER_CLASS", "uvicorn.workers.UvicornWorker"),
        help="Worker class to use (default: uvicorn.workers.UvicornWorker)",
    )

    # Development and debugging
    parser.add_argument(
        "--reload",
        action="store_true",
        default=os.environ.get("WEBHOOK_BRIDGE_RELOAD", "").lower() in ("true", "1", "yes"),
        help="Enable auto-reload for development",
    )
    parser.add_argument(
        "--reload-dirs",
        nargs="*",
        default=None,
        help="Directories to watch for reload (space-separated)",
    )

    # Logging configuration
    parser.add_argument(
        "--access-log",
        action="store_true",
        default=True,
        help="Enable access log (default: enabled)",
    )
    parser.add_argument(
        "--no-access-log",
        action="store_true",
        help="Disable access log",
    )
    parser.add_argument(
        "--use-colors",
        action="store_true",
        default=True,
        help="Use colors in log output (default: enabled)",
    )
    parser.add_argument(
        "--no-use-colors",
        action="store_true",
        help="Disable colors in log output",
    )

    # SSL/TLS configuration
    parser.add_argument(
        "--ssl-keyfile",
        type=Path,
        default=None,
        help="SSL key file path",
    )
    parser.add_argument(
        "--ssl-certfile",
        type=Path,
        default=None,
        help="SSL certificate file path",
    )
    parser.add_argument(
        "--ssl-ca-certs",
        type=Path,
        default=None,
        help="SSL CA certificates file path",
    )

    # Performance configuration
    parser.add_argument(
        "--limit-concurrency",
        type=int,
        default=None,
        help="Maximum number of concurrent connections",
    )
    parser.add_argument(
        "--limit-max-requests",
        type=int,
        default=None,
        help="Maximum number of requests before restarting worker",
    )
    parser.add_argument(
        "--timeout-keep-alive",
        type=int,
        default=5,
        help="Keep-alive timeout in seconds (default: 5)",
    )

    # API configuration
    parser.add_argument(
        "--title",
        default="Webhook Bridge API",
        help="API title (default: Webhook Bridge API)",
    )
    parser.add_argument(
        "--description",
        default="A flexible webhook integration platform",
        help="API description (default: A flexible webhook integration platform)",
    )
    parser.add_argument(
        "--disable-docs",
        action="store_true",
        help="Disable the API documentation endpoints (/docs and /redoc)",
    )

    # Plugin configuration
    parser.add_argument(
        "--plugin-dir",
        type=Path,
        default=Path.cwd() / "plugins",
        help="Directory containing webhook plugins (default: ./plugins)",
    )

    return parser


def run_server(
    host: str,
    port: int,
    plugin_dir: Path,
    log_level: str,
    *,
    disable_docs: bool = False,
    workers: int = 1,
    worker_class: str = "uvicorn.workers.UvicornWorker",
    reload: bool = False,
    reload_dirs: list[str] | None = None,
    access_log: bool = True,
    no_access_log: bool = False,
    use_colors: bool = True,
    no_use_colors: bool = False,
    ssl_keyfile: Path | None = None,
    ssl_certfile: Path | None = None,
    ssl_ca_certs: Path | None = None,
    limit_concurrency: int | None = None,
    limit_max_requests: int | None = None,
    timeout_keep_alive: int = 5,
    **kwargs: Any,
) -> None:
    """Run the webhook bridge server.

    Args:
        host: Host to bind the server to
        port: Port to bind the server to
        plugin_dir: Directory containing webhook plugins
        log_level: Logging level
        disable_docs: Whether to disable API documentation
        workers: Number of worker processes
        worker_class: Worker class to use
        reload: Enable auto-reload for development
        reload_dirs: Directories to watch for reload
        access_log: Enable access log
        no_access_log: Disable access log (overrides access_log)
        use_colors: Use colors in log output
        no_use_colors: Disable colors in log output (overrides use_colors)
        ssl_keyfile: SSL key file path
        ssl_certfile: SSL certificate file path
        ssl_ca_certs: SSL CA certificates file path
        limit_concurrency: Maximum number of concurrent connections
        limit_max_requests: Maximum number of requests before restarting worker
        timeout_keep_alive: Keep-alive timeout in seconds
        **kwargs: Additional arguments to pass to FastAPI
    """
    # Configure logging
    logging.basicConfig(
        level=getattr(logging, log_level.upper()),
        format="%(asctime)s - %(name)s - %(levelname)s - %(message)s",
    )

    kwargs["enable_docs"] = not disable_docs

    # Create FastAPI app
    app = create_app(plugin_dir=str(plugin_dir), **kwargs)

    # Prepare uvicorn configuration
    uvicorn_config = {
        "app": app,
        "host": host,
        "port": port,
        "log_level": log_level.lower(),
        "workers": workers,
        "reload": reload,
        "access_log": not no_access_log if no_access_log else access_log,
        "use_colors": not no_use_colors if no_use_colors else use_colors,
        "timeout_keep_alive": timeout_keep_alive,
    }

    # Add optional configurations
    if reload_dirs:
        uvicorn_config["reload_dirs"] = reload_dirs

    if ssl_keyfile:
        uvicorn_config["ssl_keyfile"] = str(ssl_keyfile)

    if ssl_certfile:
        uvicorn_config["ssl_certfile"] = str(ssl_certfile)

    if ssl_ca_certs:
        uvicorn_config["ssl_ca_certs"] = str(ssl_ca_certs)

    if limit_concurrency is not None:
        uvicorn_config["limit_concurrency"] = limit_concurrency

    if limit_max_requests is not None:
        uvicorn_config["limit_max_requests"] = limit_max_requests

    # For multiple workers, we need to use a different approach
    if workers > 1:
        # When using multiple workers, we can't pass the app instance directly
        # Set environment variables for the app factory
        os.environ["WEBHOOK_BRIDGE_PLUGIN_DIR"] = str(plugin_dir)
        if "title" in kwargs:
            os.environ["WEBHOOK_BRIDGE_TITLE"] = kwargs["title"]
        if "description" in kwargs:
            os.environ["WEBHOOK_BRIDGE_DESCRIPTION"] = kwargs["description"]
        if "enable_docs" in kwargs:
            os.environ["WEBHOOK_BRIDGE_ENABLE_DOCS"] = str(kwargs["enable_docs"])

        uvicorn_config["app"] = "webhook_bridge.cli:get_app"

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


def main(argv: Sequence[str] | None = None) -> None:
    """Main entry point.

    Args:
        argv: Command-line arguments
    """
    parser = create_parser()
    args = parser.parse_args(argv)

    try:
        run_server(
            host=args.host,
            port=args.port,
            plugin_dir=args.plugin_dir,
            log_level=args.log_level,
            title=args.title,
            description=args.description,
            disable_docs=args.disable_docs,
            workers=args.workers,
            worker_class=args.worker_class,
            reload=args.reload,
            reload_dirs=args.reload_dirs,
            access_log=args.access_log,
            no_access_log=args.no_access_log,
            use_colors=args.use_colors,
            no_use_colors=args.no_use_colors,
            ssl_keyfile=args.ssl_keyfile,
            ssl_certfile=args.ssl_certfile,
            ssl_ca_certs=args.ssl_ca_certs,
            limit_concurrency=args.limit_concurrency,
            limit_max_requests=args.limit_max_requests,
            timeout_keep_alive=args.timeout_keep_alive,
        )
        sys.exit(0)
    except Exception as e:
        logging.error("Error running server: %s", e)
        sys.exit(1)


if __name__ == "__main__":
    main()
