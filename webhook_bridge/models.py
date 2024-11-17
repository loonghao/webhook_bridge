# Import built-in modules
from typing import Any
from typing import Dict

# Import third-party modules
from fastapi import status
from pydantic import BaseModel


class PluginNotFoundResponse(BaseModel):  # type: ignore
    status_code: int = status.HTTP_404_NOT_FOUND
    message: str


class PluginResponse(PluginNotFoundResponse):
    status_code: int = status.HTTP_200_OK
    data: Dict[str, Any]
