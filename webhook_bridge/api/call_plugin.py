# Import built-in modules

# Import built-in modules
import traceback
from typing import Any
from typing import Dict
from typing import Type

# Import third-party modules
from fastapi import FastAPI
from fastapi_versioning import version
from webhook_bridge.filesystem import get_plugins
from webhook_bridge.models import PluginResponse
from webhook_bridge.plugin import BasePlugin
from webhook_bridge.plugin import load_plugin


def api(app: FastAPI) -> None:
    @app.post("/plugin/{plugin_name}", response_model=PluginResponse)  # type: ignore
    @version(1)  # type: ignore
    async def plugin_integrated(plugin_name: str, data: Dict[str, Any]) -> PluginResponse:
        plugins = get_plugins()
        plugin_src_file = plugins.get(plugin_name)

        if not plugin_src_file:
            msg = (f"Plugin {plugin_name} not found. "
                   f"The currently available plugins are {list(plugins.keys())}")
            return PluginResponse(status_code=404, message=msg, data={
                "plugin_name": plugin_name, "src_data": data
            })
        else:
            try:
                plugin_class: Type[BasePlugin] = load_plugin(plugin_src_file)
                plugin_instance = plugin_class(data)
                result = plugin_instance.run()
            except Exception as err:
                result = {
                    "error": str(err), "traceback": traceback.format_exc()
                }
                return PluginResponse(status_code=500, message="Plugin did not return a dictionary.",
                                      data={
                                          "plugin_name": plugin_name, "src_data": data,
                                          "result": result
                                      }
                                      )
            # Ensure the result is a dictionary
            if not isinstance(result, dict):
                return PluginResponse(status_code=500, message="Plugin did not return a dictionary.",
                                      data={
                                          "plugin_name": plugin_name, "src_data": data,
                                          "result": result
                                      }
                                      )

            content: Dict[str, Any] = {
                "plugin_name": plugin_name,
                "src_data": data,
                "result": result
            }
            return PluginResponse(data=content, message="Plugin executed successfully")
