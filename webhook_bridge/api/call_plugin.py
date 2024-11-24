"""API endpoint for executing webhook plugins."""
# Import future modules
from __future__ import annotations

# Import built-in modules
import logging
import traceback
from typing import Any

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
from webhook_bridge.plugin import BasePlugin
from webhook_bridge.plugin import load_plugin


def api(app: FastAPI) -> None:
    """Register the plugin execution endpoint with the FastAPI application.

    Args:
        app: The FastAPI application instance to register the endpoint with
    """
    router = APIRouter()

    @router.post(
        "/plugin/{plugin_name}",
        response_model=WebhookResponseData,
        summary="Execute a plugin",
        description="Execute a specific webhook plugin with the provided data",
        status_code=status.HTTP_200_OK,
        responses={
            200: {
                "description": "Plugin executed successfully",
                "content": {
                    "application/json": {
                        "example": {
                            "status_code": 200,
                            "message": "success",
                            "data": {
                                "plugin": "example",
                                "src_data": {"key": "value"},
                                "result": {
                                    "status": "success",
                                    "data": {"key": "value"},
                                },
                            },
                        },
                    },
                },
            },
            404: {
                "description": "Plugin not found",
                "content": {
                    "application/json": {
                        "example": {
                            "status_code": 404,
                            "message": "Plugin not found",
                            "data": {
                                "error": "Plugin not found",
                            },
                        },
                    },
                },
            },
            500: {
                "description": "Plugin execution failed",
                "content": {
                    "application/json": {
                        "example": {
                            "status_code": 500,
                            "message": "Plugin execution failed",
                            "data": {
                                "details": "Error message",
                            },
                        },
                    },
                },
            },
        },
    )
    @version(1)
    async def execute_plugin(
        request: Request,
        plugin_name: str,
        data: dict[str, Any],
    ) -> WebhookResponse:
        """Execute a specific webhook plugin.

        Args:
            request: FastAPI request object
            plugin_name: Name of the plugin to execute
            data: Data to pass to the plugin

        Returns:
            WebhookResponse: Response containing the plugin execution results
        """
        logger = logging.getLogger(__name__)
        logger.info("Executing plugin %r with data: %s", plugin_name, data)

        # Get plugin directory from app state
        plugin_dir = app.state.plugin_dir
        if plugin_dir is None:
            return WebhookResponse(
                status_code=status.HTTP_500_INTERNAL_SERVER_ERROR,
                message="Plugin execution failed",
                data={
                    "error": "Plugin directory not configured",
                    "details": "Plugin directory not set in application state",
                },
            )

        # Get available plugins
        plugins = get_plugins(plugin_dir)
        plugin_src_file = plugins.get(plugin_name)

        if not plugin_src_file:
            logger.error("Plugin %r not found in directory: %s", plugin_name, plugin_dir)
            return WebhookResponse(
                status_code=status.HTTP_404_NOT_FOUND,
                message="Plugin not found",
                data={
                    "error": "Plugin not found",
                },
            )

        try:
            # Load and execute plugin
            plugin_class: type[BasePlugin] = load_plugin(plugin_src_file)
            plugin_instance = plugin_class(data)
            result = plugin_instance.execute()

            logger.info("Successfully executed plugin %r", plugin_name)
            return WebhookResponse(
                status_code=status.HTTP_200_OK,
                message="success",
                data={
                    "plugin": plugin_name,
                    "plugin_data": result,
                },
            )
        except Exception as e:
            logger.error(f"Error executing plugin: {e}")
            logger.debug(traceback.format_exc())
            error_msg = str(e)
            logger.error(
                "Failed to execute plugin %r: %s\n%s",
                plugin_name,
                error_msg,
                traceback.format_exc(),
            )
            return WebhookResponse(
                status_code=status.HTTP_500_INTERNAL_SERVER_ERROR,
                message="Plugin execution failed",
                data={
                    "details": error_msg,
                },
            )

    app.include_router(router)
