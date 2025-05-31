"""
Command-line interface for webhook-bridge.

This CLI tool manages the Go-based webhook bridge server, including downloading,
installing, and running the appropriate binary for the current platform.
"""

import argparse
import sys
from pathlib import Path

from .manager import WebhookBridgeManager


def create_parser() -> argparse.ArgumentParser:
    """Create the argument parser for the CLI."""
    parser = argparse.ArgumentParser(
        prog="webhook-bridge",
        description="Webhook Bridge - A flexible webhook integration platform",
        formatter_class=argparse.RawDescriptionHelpFormatter,
        epilog="""
Examples:
  webhook-bridge --version              Show version information
  webhook-bridge install               Download and install the server binary
  webhook-bridge run                   Start the webhook bridge server
  webhook-bridge run --port 9000       Start server on custom port
  webhook-bridge status                Check server status
  webhook-bridge stop                  Stop the running server
  webhook-bridge update                Update to the latest version
  webhook-bridge config                Show configuration options

For more information, visit: https://github.com/loonghao/webhook_bridge
        """,
    )

    parser.add_argument(
        "--version",
        action="version",
        version=f"webhook-bridge {WebhookBridgeManager.get_version()}",
    )

    parser.add_argument(
        "--verbose", "-v",
        action="store_true",
        help="Enable verbose output",
    )

    parser.add_argument(
        "--config", "-c",
        type=Path,
        help="Path to configuration file",
    )

    subparsers = parser.add_subparsers(dest="command", help="Available commands")

    # Install command
    install_parser = subparsers.add_parser(
        "install",
        help="Download and install the webhook bridge server binary",
    )
    install_parser.add_argument(
        "--force", "-f",
        action="store_true",
        help="Force reinstallation even if already installed",
    )
    install_parser.add_argument(
        "--version",
        help="Specific version to install (default: latest)",
    )

    # Run command
    run_parser = subparsers.add_parser(
        "run",
        help="Start the webhook bridge server",
    )
    run_parser.add_argument(
        "--port", "-p",
        type=int,
        default=8000,
        help="Port to run the server on (default: 8000)",
    )
    run_parser.add_argument(
        "--host",
        default="0.0.0.0",
        help="Host to bind the server to (default: 0.0.0.0)",
    )
    run_parser.add_argument(
        "--daemon", "-d",
        action="store_true",
        help="Run server in daemon mode",
    )

    # Status command
    subparsers.add_parser(
        "status",
        help="Check the status of the webhook bridge server",
    )

    # Stop command
    subparsers.add_parser(
        "stop",
        help="Stop the running webhook bridge server",
    )

    # Update command
    update_parser = subparsers.add_parser(
        "update",
        help="Update to the latest version",
    )
    update_parser.add_argument(
        "--check-only",
        action="store_true",
        help="Only check for updates, don't install",
    )

    # Config command
    config_parser = subparsers.add_parser(
        "config",
        help="Configuration management",
    )
    config_subparsers = config_parser.add_subparsers(dest="config_action")
    config_subparsers.add_parser("show", help="Show current configuration")
    config_subparsers.add_parser("init", help="Initialize default configuration")
    config_subparsers.add_parser("validate", help="Validate configuration file")

    return parser


def main() -> int:
    """Main entry point for the CLI."""
    parser = create_parser()
    args = parser.parse_args()

    # If no command is provided, show help
    if not args.command:
        parser.print_help()
        return 0

    try:
        manager = WebhookBridgeManager(verbose=args.verbose, config_path=args.config)

        if args.command == "install":
            return manager.install(force=args.force, version=getattr(args, 'version', None))
        elif args.command == "run":
            return manager.run(
                port=args.port,
                host=args.host,
                daemon=args.daemon,
            )
        elif args.command == "status":
            return manager.status()
        elif args.command == "stop":
            return manager.stop()
        elif args.command == "update":
            return manager.update(check_only=args.check_only)
        elif args.command == "config":
            return manager.config(args.config_action or "show")
        else:
            parser.print_help()
            return 1

    except KeyboardInterrupt:
        print("\n⚠️  Operation cancelled by user")
        return 130
    except Exception as e:
        print(f"❌ Error: {e}")
        if args.verbose:
            import traceback
            traceback.print_exc()
        return 1


if __name__ == "__main__":
    sys.exit(main())
