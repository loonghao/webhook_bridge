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

    # Common response definitions for all HTTP methods
    plugin_responses = {
        200: {
            "description": "Plugin executed successfully",
            "content": {
                "application/json": {
                    "example": {
                        "status_code": 200,
                        "message": "success",
                        "data": {
                            "plugin": "example",
                            "plugin_data": {
                                "input_data": {"key": "value"},
                                "additional_info": "This is some additional information.",
                                "http_method": "POST",
                                "result": {
                                    "status": "success",
                                    "data": {"key": "value"},
                                },
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
    }

    # Helper function to execute a plugin
    async def _execute_plugin(
        request: Request,
        plugin_name: str,
        data: dict[str, Any],
        http_method: str,
    ) -> WebhookResponse:
        """Execute a specific webhook plugin.

        Args:
            request: FastAPI request object
            plugin_name: Name of the plugin to execute
            data: Data to pass to the plugin
            http_method: HTTP method used to call the plugin

        Returns:
            WebhookResponse: Response containing the plugin execution results
        """
        logger = logging.getLogger(__name__)
        logger.info("Executing plugin %r with method %r and data: %s", plugin_name, http_method, data)

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
            plugin_instance = plugin_class(data, http_method=http_method)
            result = plugin_instance.execute()

            logger.info("Successfully executed plugin %r with method %r", plugin_name, http_method)
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
                "Failed to execute plugin %r with method %r: %s\n%s",
                plugin_name,
                http_method,
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

    # POST endpoint
    @router.post(
        "/plugin/{plugin_name}",
        response_model=WebhookResponseData,
        summary="Execute a plugin with POST",
        description="Execute a specific webhook plugin with the provided data using POST method",
        status_code=status.HTTP_200_OK,
        responses=plugin_responses,
    )
    @version(1)
    async def execute_plugin_post(
        request: Request,
        plugin_name: str,
        data: dict[str, Any],
    ) -> WebhookResponse:
        """Execute a specific webhook plugin using POST method.

        Args:
            request: FastAPI request object
            plugin_name: Name of the plugin to execute
            data: Data to pass to the plugin

        Returns:
            WebhookResponse: Response containing the plugin execution results
        """
        # For backward compatibility, log a message if the plugin uses the run method
        logger = logging.getLogger(__name__)
        logger.info(f"Executing plugin {plugin_name} with POST method")
        return await _execute_plugin(request, plugin_name, data, "POST")

    # GET endpoint
    @router.get(
        "/plugin/{plugin_name}",
        response_model=WebhookResponseData,
        summary="Execute a plugin with GET",
        description="Execute a specific webhook plugin using GET method",
        status_code=status.HTTP_200_OK,
        responses=plugin_responses,
    )
    @version(1)
    async def execute_plugin_get(
        request: Request,
        plugin_name: str,
    ) -> WebhookResponse:
        """Execute a specific webhook plugin using GET method.

        Args:
            request: FastAPI request object
            plugin_name: Name of the plugin to execute

        Returns:
            WebhookResponse: Response containing the plugin execution results
        """
        # For GET requests, we extract query parameters as data
        data = dict(request.query_params)
        return await _execute_plugin(request, plugin_name, data, "GET")

    # PUT endpoint
    @router.put(
        "/plugin/{plugin_name}",
        response_model=WebhookResponseData,
        summary="Execute a plugin with PUT",
        description="Execute a specific webhook plugin with the provided data using PUT method",
        status_code=status.HTTP_200_OK,
        responses=plugin_responses,
    )
    @version(1)
    async def execute_plugin_put(
        request: Request,
        plugin_name: str,
        data: dict[str, Any],
    ) -> WebhookResponse:
        """Execute a specific webhook plugin using PUT method.

        Args:
            request: FastAPI request object
            plugin_name: Name of the plugin to execute
            data: Data to pass to the plugin

        Returns:
            WebhookResponse: Response containing the plugin execution results
        """
        return await _execute_plugin(request, plugin_name, data, "PUT")

    # DELETE endpoint
    @router.delete(
        "/plugin/{plugin_name}",
        response_model=WebhookResponseData,
        summary="Execute a plugin with DELETE",
        description="Execute a specific webhook plugin using DELETE method",
        status_code=status.HTTP_200_OK,
        responses=plugin_responses,
    )
    @version(1)
    async def execute_plugin_delete(
        request: Request,
        plugin_name: str,
    ) -> WebhookResponse:
        """Execute a specific webhook plugin using DELETE method.

        Args:
            request: FastAPI request object
            plugin_name: Name of the plugin to execute

        Returns:
            WebhookResponse: Response containing the plugin execution results
        """
        # For DELETE requests, we extract query parameters as data
        data = dict(request.query_params)
        return await _execute_plugin(request, plugin_name, data, "DELETE")

    app.include_router(router)
