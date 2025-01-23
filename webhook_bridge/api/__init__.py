"""API initialization and configuration module.

This module sets up the FastAPI application with versioned endpoints and
configures all available API routes. It provides a clean interface for
registering API endpoints and managing API versions.

Example:
    >>> from fastapi import FastAPI
    >>> app = FastAPI()
    >>> versioned_app = setup_api(app)
"""
# Import future modules
from __future__ import annotations

# Import third-party modules
from fastapi import FastAPI
from fastapi_versioning import VersionedFastAPI

# Import local modules
from webhook_bridge.api import call_plugin
from webhook_bridge.api import list_plugins


def setup_api(app: FastAPI) -> FastAPI:
    """Set up API routes and versioning for the application.

    This function registers all available API endpoints and wraps the application
    with versioning support. The API will be accessible under /api/v{major}
    where {major} is the major version number.

    Args:
        app: The FastAPI application instance to configure

    Returns:
        FastAPI: The configured application with versioning support

    Example:
        >>> app = FastAPI()
        >>> configured_app = setup_api(app)
        >>> # API now available at /api/v1/...
    """
    # Register versioned APIs directly on the main app
    list_plugins.api(app)
    call_plugin.api(app)

    # Configure versioned API
    versioned_app = VersionedFastAPI(
        app,
        version_format="{major}",
        prefix_format="/api/v{major}",
        enable_latest=True,
        docs_url=app.docs_url,
        redoc_url=app.redoc_url,
        openapi_url=app.openapi_url,
    )

    return versioned_app
