# Import third-party modules
from fastapi import FastAPI
from fastapi_versioning import VersionedFastAPI
from webhook_bridge.api import call_plugin
from webhook_bridge.api import list_plugins


def setup_api(app: FastAPI) -> VersionedFastAPI:
    # All public APIs.
    list_plugins.api(app)
    call_plugin.api(app)

    return VersionedFastAPI(app, version_format="{major}", prefix_format="/api/v{major}")
