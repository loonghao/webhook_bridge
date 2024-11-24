"""Data models for the webhook bridge.

This module defines the Pydantic models used for data validation and serialization
in the webhook bridge application.

Example:
    >>> from webhook_bridge.models import WebhookRequest
    >>> request = WebhookRequest(plugin="test.py", data={"key": "value"})
    >>> request.model_dump()
    {'plugin': 'test.py', 'data': {'key': 'value'}}
"""
# Import future modules
from __future__ import annotations

# Import built-in modules
from typing import Any

# Import third-party modules
from fastapi import status
from fastapi.responses import JSONResponse
from pydantic import BaseModel
from pydantic import ConfigDict
from pydantic import Field


class WebhookRequest(BaseModel):
    """Model representing a webhook request.

    This model validates incoming webhook requests, ensuring they contain
    the required plugin name and data fields.

    Attributes:
        plugin: Name of the plugin to execute
        data: Dictionary containing the data to pass to the plugin
    """

    model_config = ConfigDict(
        json_schema_extra={
            "examples": [
                {
                    "plugin": "example.py",
                    "data": {"message": "Hello, World!"},
                },
            ],
        },
    )

    plugin: str = Field(
        ...,
        description="Name of the plugin to execute",
        examples=["example.py"],
    )
    data: dict[str, Any] = Field(
        ...,
        description="Data to pass to the plugin",
        examples=[{"message": "Hello, World!"}],
    )


class WebhookResponseData(BaseModel):
    """Model representing a webhook response data.

    This model validates and serializes the response from webhook plugins,
    ensuring they contain the required status code, message and data fields.

    Attributes:
        status_code: HTTP status code
        message: Response message
        data: Dictionary containing the response data
    """

    model_config = ConfigDict(
        json_schema_extra={
            "examples": [
                {
                    "status_code": status.HTTP_200_OK,
                    "message": "success",
                    "data": {"message": "Operation completed"},
                },
                {
                    "status_code": status.HTTP_500_INTERNAL_SERVER_ERROR,
                    "message": "Plugin execution failed",
                    "data": {"error": "Plugin not found"},
                },
            ],
        },
    )

    status_code: int = Field(
        ...,
        description="HTTP status code",
        examples=[status.HTTP_200_OK, status.HTTP_500_INTERNAL_SERVER_ERROR],
    )
    message: str = Field(
        ...,
        description="Response message",
        examples=["success", "error"],
    )
    data: dict[str, Any] = Field(
        default_factory=dict,
        description="Response data",
    )


class WebhookResponse(JSONResponse):
    """FastAPI response model for webhook responses.

    This class extends JSONResponse to provide a consistent response format
    for webhook API endpoints.
    """

    def __init__(
        self,
        status_code: int,
        message: str,
        data: dict[str, Any] | None = None,
        **kwargs: Any,
    ) -> None:
        """Initialize a new WebhookResponse.

        Args:
            status_code: HTTP status code
            message: Response message
            data: Response data
            **kwargs: Additional arguments to pass to JSONResponse
        """
        response_data = WebhookResponseData(
            status_code=status_code,
            message=message,
            data=data or {},
        )

        # If data contains HTML content, return HTML response
        if data and "readme" in data:
            super().__init__(
                content=data["readme"],
                status_code=status_code,
                media_type="text/html",
            )
        else:
            # Otherwise return JSON response
            super().__init__(
                status_code=status_code,
                content=response_data.model_dump(),
                **kwargs,
            )
