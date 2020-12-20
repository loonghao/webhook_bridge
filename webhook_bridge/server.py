# Import built-in modules
from typing import Dict

# Import third-party modules
import uvicorn
from fastapi import FastAPI
from fastapi.responses import JSONResponse

# Import local modules
from webhook_bridge import paths
from webhook_bridge.plugin import load_plugin

APP = FastAPI(debug=True,
              title="Webhook bridge server")


@APP.router.post("/api/plugin/{plugin_name}")
async def plugin_integrated(plugin_name: str,
                            data: Dict):
    plugins = paths.get_plugins()
    plugin_src_file = plugins.get(plugin_name)
    if not plugin_src_file:
        mgs = (f"Plugin {plugin_name} not found. "
               f"The currently available plugins are {list(plugins.keys())}")
        return JSONResponse(status_code=404, content={"message": mgs})
    else:
        plugin = load_plugin(plugin_src_file)
        plugin_instance = plugin(data)
        plugin_instance.run()
        content = f"{plugin_name}: executed successfully."
        return JSONResponse(status_code=200, content={"message": content})


def start_server():
    port = 5001
    uvicorn.run("webhook_bridge.server:APP",
                host="localhost",
                port=port,
                reload=True,
                log_level="info")


if __name__ == "__main__":
    start_server()
