# Import built-in modules

# Import third-party modules
from fastapi import FastAPI
from fastapi.responses import HTMLResponse
from markdown import markdown  # You may need to install this package
from webhook_bridge.api import setup_api


APP = FastAPI(debug=True, title="Webhook Bridge Server")

# Markdown content for the homepage
markdown_content = """
# Welcome to the Webhook Bridge Server

This server allows you to integrate various plugins.

You can find the source code on [GitHub](https://github.com/loonghao/webhook_bridge).

## Available Plugins
- Plugin1
- Plugin2
- Plugin3
"""

@APP.get("/", response_class=HTMLResponse) # type: ignore
async def read_root() -> HTMLResponse:
    html_content = markdown(markdown_content)
    return HTMLResponse(content=html_content)

# Setup API.
APP = setup_api(APP)
