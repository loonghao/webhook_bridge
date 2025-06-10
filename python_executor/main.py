#!/usr/bin/env python3
"""Python gRPC Executor Service for Webhook Bridge.

This service provides a gRPC interface for executing Python webhook plugins,
maintaining full compatibility with the existing plugin system.
"""
# Import future modules
from __future__ import annotations

# Import built-in modules
import argparse
import asyncio
from concurrent import futures
import logging
import os
from pathlib import Path
import signal
import sys

# Import third-party modules
import grpc


# Add project root to Python path
project_root = Path(__file__).parent.parent
sys.path.insert(0, str(project_root))

# Import third-party modules
from api.proto import webhook_pb2_grpc
from python_executor.server import WebhookExecutorServicer
from python_executor.utils import get_port_with_fallback
from python_executor.utils import is_port_free


def setup_logging(level: str = "INFO") -> None:
    """Setup logging configuration."""
    logging.basicConfig(
        level=getattr(logging, level.upper()),
        format="%(asctime)s - %(name)s - %(levelname)s - %(message)s",
        handlers=[
            logging.StreamHandler(sys.stdout),
            logging.FileHandler("python_executor.log"),
        ],
    )


def create_server(host: str = "localhost", port: int = 50051, plugin_dirs: list[str] = None) -> grpc.aio.Server:
    """Create and configure the gRPC server."""
    server = grpc.aio.server(futures.ThreadPoolExecutor(max_workers=10))

    # Add the webhook executor service with plugin directories
    servicer = WebhookExecutorServicer(plugin_dirs=plugin_dirs)
    webhook_pb2_grpc.add_WebhookExecutorServicer_to_server(servicer, server)

    # Add insecure port
    listen_addr = f"{host}:{port}"
    server.add_insecure_port(listen_addr)

    return server


async def serve(host: str = "0.0.0.0", port: int = 50051, plugin_dirs: list[str] = None) -> None:
    """Start the gRPC server."""
    logger = logging.getLogger(__name__)

    server = create_server(host, port, plugin_dirs)

    # Start server
    await server.start()
    logger.info(f"Python Executor gRPC server started on {host}:{port}")
    if plugin_dirs:
        logger.info(f"Additional plugin directories: {plugin_dirs}")

    # Setup graceful shutdown with proper async handling
    shutdown_event = asyncio.Event()

    def signal_handler(_signum, _frame):
        logger.info("Received shutdown signal, initiating graceful shutdown...")
        shutdown_event.set()

    signal.signal(signal.SIGINT, signal_handler)
    signal.signal(signal.SIGTERM, signal_handler)

    try:
        # Wait for either termination or shutdown signal
        shutdown_task = asyncio.create_task(shutdown_event.wait())
        termination_task = asyncio.create_task(server.wait_for_termination())

        done, pending = await asyncio.wait(
            [shutdown_task, termination_task],
            return_when=asyncio.FIRST_COMPLETED
        )

        # Cancel pending tasks
        for task in pending:
            task.cancel()
            try:
                await task
            except asyncio.CancelledError:
                pass

        # If shutdown was triggered by signal, stop the server
        if shutdown_task in done:
            logger.info("Stopping server gracefully...")
            await server.stop(grace=5)
            logger.info("Server stopped successfully")

    except KeyboardInterrupt:
        logger.info("Keyboard interrupt received, stopping server...")
        await server.stop(grace=5)
        logger.info("Server stopped successfully")


def main() -> None:
    """Main entry point."""
    parser = argparse.ArgumentParser(description="Python Executor gRPC Service")
    parser.add_argument(
        "--host",
        default="0.0.0.0",
        help="Host to bind the server to (default: 0.0.0.0)",
    )
    parser.add_argument(
        "--port",
        type=int,
        default=0,
        help="Port to bind the server to (default: 0 for auto-assign)",
    )
    parser.add_argument(
        "--log-level",
        default="INFO",
        choices=["DEBUG", "INFO", "WARNING", "ERROR"],
        help="Log level (default: INFO)",
    )
    parser.add_argument(
        "--plugin-dirs",
        nargs="*",
        help="Additional plugin directories to search",
    )
    parser.add_argument(
        "--config",
        help="Configuration file path (YAML format)",
    )

    args = parser.parse_args()

    # Setup logging
    setup_logging(args.log_level)

    # Load configuration if provided
    plugin_dirs = args.plugin_dirs or []
    config_port = None
    if args.config:
        try:
            # Import third-party modules
            try:
                # Import third-party modules
                import yaml
            except ImportError:
                logging.error("PyYAML is not installed. Please install it with: pip install PyYAML")
                sys.exit(1)

            if not os.path.exists(args.config):
                logging.error(f"Configuration file not found: {args.config}")
                sys.exit(1)

            with open(args.config) as f:
                config_data = yaml.safe_load(f)

                # Load plugin directories
                if 'python' in config_data and 'plugin_dirs' in config_data['python']:
                    plugin_dirs.extend(config_data['python']['plugin_dirs'])
                    logging.info(f"Loaded plugin directories from config: {config_data['python']['plugin_dirs']}")

                # Load executor port configuration
                if 'executor' in config_data and 'port' in config_data['executor']:
                    config_port = config_data['executor']['port']
                    logging.info(f"Loaded executor port from config: {config_port}")

        except Exception as e:
            logging.warning(f"Failed to load configuration from {args.config}: {e}")

    # Assign port with priority: command line > config file > auto-assign
    port = args.port
    if port == 0 and config_port is not None:
        port = config_port
        logging.info(f"Using port from config file: {port}")
    if port == 0:
        port = get_port_with_fallback(50051, args.host)  # Prefer 50051, fallback to any free port
        logging.info(f"Auto-assigned port: {port}")
    elif not is_port_free(port, args.host):
        logging.warning(f"Port {port} is not available on {args.host}, finding alternative...")
        port = get_port_with_fallback(port, args.host)
        logging.info(f"Using alternative port: {port}")

    # Start server
    try:
        logging.info(f"Starting Python executor on {args.host}:{port}")
        logging.info(f"Plugin directories: {plugin_dirs}")
        asyncio.run(serve(args.host, port, plugin_dirs))
    except KeyboardInterrupt:
        logging.info("Server stopped by user")
    except Exception as e:
        logging.error(f"Server error: {e}")
        logging.error(f"Failed to start server on {args.host}:{port}")
        # Import built-in modules
        import traceback
        logging.error(f"Traceback: {traceback.format_exc()}")
        sys.exit(1)


if __name__ == "__main__":
    main()
