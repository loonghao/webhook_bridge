"""
CLI module for webhook-bridge Python components.

Note: The main CLI is implemented by the Rust webhook-bridge binary.
This module provides Python-specific utilities and development helpers.
"""

# Import built-in modules
import sys
from typing import Optional


def main(args: Optional[list] = None) -> int:
    """
    Main entry point for Python CLI utilities.

    This is a minimal implementation that redirects users to the Rust CLI.
    """
    if args is None:
        args = sys.argv[1:]
    
    print("webhook-bridge Python hook utilities")
    print("Note: the main CLI is implemented by the Rust webhook-bridge binary.")
    print("Use: webhook-bridge --help")
    print("For Python development utilities, use: uvx nox -l")
    
    return 0


if __name__ == "__main__":
    sys.exit(main())
