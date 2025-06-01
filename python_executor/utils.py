"""Utility functions for Python executor."""

# Import built-in modules
import socket
from typing import Optional


def get_free_port() -> int:
    """Find and return a free port on the local machine."""
    with socket.socket(socket.AF_INET, socket.SOCK_STREAM) as s:
        s.bind(('localhost', 0))
        s.listen(1)
        port = s.getsockname()[1]
    return port


def get_free_port_in_range(start: int, end: int) -> Optional[int]:
    """Find a free port within the specified range."""
    for port in range(start, end + 1):
        if is_port_free(port):
            return port
    return None


def is_port_free(port: int, host: str = 'localhost') -> bool:
    """Check if a port is available for use."""
    try:
        with socket.socket(socket.AF_INET, socket.SOCK_STREAM) as s:
            s.bind((host, port))
            return True
    except OSError:
        return False


def get_port_with_fallback(preferred_port: int, host: str = 'localhost') -> int:
    """Try to use the specified port, fall back to a free port if occupied."""
    if preferred_port > 0 and is_port_free(preferred_port, host):
        return preferred_port

    # If preferred port is not available, find a free one
    return get_free_port()


def parse_port(port_str: str) -> int:
    """Parse a port string and validate it."""
    if not port_str:
        raise ValueError("Port cannot be empty")

    try:
        port = int(port_str)
    except ValueError:
        raise ValueError(f"Invalid port number: {port_str}")

    if port < 1 or port > 65535:
        raise ValueError(f"Port must be between 1 and 65535, got {port}")

    return port
