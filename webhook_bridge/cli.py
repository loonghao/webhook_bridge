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
    **kwargs: Any,
) -> None:
    """Run the webhook bridge server.

    Args:
        host: Host to bind the server to
        port: Port to bind the server to
        plugin_dir: Directory containing webhook plugins
        log_level: Logging level
        disable_docs: Whether to disable API documentation
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

    # Run server
    uvicorn.run(
        app=app,
        host=host,
        port=port,
        log_level=log_level.lower(),
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
        )
        sys.exit(0)
    except Exception as e:
        logging.error("Error running server: %s", e)
        sys.exit(1)


if __name__ == "__main__":
    main()
