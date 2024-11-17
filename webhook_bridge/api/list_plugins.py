# Import built-in modules
import logging

# Import third-party modules
from fastapi import FastAPI
from fastapi import status
from fastapi_versioning import version
from webhook_bridge.filesystem import get_plugins
from webhook_bridge.models import PluginResponse


def api(app: FastAPI) -> None:
    @app.get("/plugins", response_model=PluginResponse)  # type: ignore
    @version(1)  # type: ignore
    async def list_plugins() -> PluginResponse:
        plugins_data = get_plugins()
        plugins = list(plugins_data.keys())
        logger = logging.getLogger(__name__)
        logger.info("Get plugins list: {}".format(plugins))
        response = PluginResponse(status_code=status.HTTP_200_OK, message="successes", data={"plugins": plugins})
        return response
