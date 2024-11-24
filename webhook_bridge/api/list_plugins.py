"""API endpoint for listing available plugins.

This module provides the API endpoint for retrieving a list of all available
plugins in the webhook bridge system. It exposes a GET endpoint that returns
the names of all registered plugins.
"""
# Import future modules
from __future__ import annotations

# Import built-in modules
import logging
import traceback
from typing import Any
from typing import Dict

# Import third-party modules
from fastapi import FastAPI
from fastapi import Request
from fastapi import status
from fastapi.routing import APIRouter
from fastapi_versioning import version

# Import local modules
from webhook_bridge.filesystem import get_plugins
from webhook_bridge.models import WebhookResponse
from webhook_bridge.models import WebhookResponseData


def api(app: FastAPI) -> None:
    """Register the plugin listing endpoint with the FastAPI application.

    Args:
        app: The FastAPI application instance to register the endpoint with
    """
    router = APIRouter()

    @router.get(
        "/plugins",
        response_model=WebhookResponseData,
        summary="List available plugins",
        description="Retrieve a list of all available webhook plugins",
        status_code=status.HTTP_200_OK,
        responses={
            200: {
                "description": "List of available plugins",
                "content": {
                    "application/json": {
                        "example": {
                            "status_code": 200,
                            "message": "success",
                            "data": {
                                "plugins": ["plugin1", "plugin2"],
                            },
                        },
                    },
                },
            },
            500: {
                "description": "Internal server error",
                "content": {
                    "application/json": {
                        "example": {
                            "status_code": 500,
                            "message": "Failed to retrieve plugins",
                            "data": {
                                "error": "Internal server error",
                                "details": "Error message",
                            },
                        },
                    },
                },
            },
        },
    )
    @version(1)
    async def list_plugins(request: Request) -> WebhookResponse:
        """List all available webhook plugins.

        Returns:
            WebhookResponse: Response containing the list of available plugins.

        Raises:
            HTTPException: If there's an error retrieving plugins.
        """
        logger = logging.getLogger(__name__)

        try:
            # Get plugin directory from app state
            plugin_dir = app.state.plugin_dir
            # Get list of available plugins
            plugins_data: Dict[str, Any] = get_plugins(plugin_dir)
            if not plugins_data:
                logger.warning("No plugins found in directory: %s", plugin_dir)
                return WebhookResponse(
                    status_code=status.HTTP_200_OK,
                    message="success",
                    data={
                        "plugins": [],
                    },
                )

            logger.info("Found %d plugins in directory: %s", len(plugins_data), plugin_dir)
            return WebhookResponse(
                status_code=status.HTTP_200_OK,
                message="success",
                data={
                    "plugins": list(plugins_data.keys()),
                },
            )

        except Exception as e:
            logger.error(f"Error listing plugins: {e}")
            logger.debug(traceback.format_exc())
            logger.error("Failed to list plugins: %s", e)
            return WebhookResponse(
                status_code=status.HTTP_500_INTERNAL_SERVER_ERROR,
                message="Plugin listing failed",
                data={
                    "error": "Internal server error",
                    "details": str(e),
                },
            )

    app.include_router(router)
