"""FastAPI server configuration and startup."""
# Import future modules
from __future__ import annotations

# Import built-in modules
from pathlib import Path
from typing import Any

# Import third-party modules
from fastapi import FastAPI
from fastapi import Request
from fastapi.middleware.cors import CORSMiddleware
from fastapi.templating import Jinja2Templates
import httpx
import markdown2

# Import local modules
from webhook_bridge.__version__ import __version__
from webhook_bridge.api import setup_api


async def get_github_readme() -> str | None:
    """Fetch README content from GitHub.

    Returns:
        Optional[str]: The README content if successfully fetched, None otherwise.

    """
    url = "https://raw.githubusercontent.com/loonghao/webhook_bridge/main/README.md"
    async with httpx.AsyncClient() as client:
        try:
            response = await client.get(url)
            if response.status_code == 200:
                content: str = response.text.replace(
                    "](docs/",
                    "](https://raw.githubusercontent.com/loonghao/webhook_bridge/main/docs/",
                )
                return content
        except httpx.HTTPError:
            pass
    return None


# Configure templates
templates = Jinja2Templates(directory=str(Path(__file__).parent / "templates"))


def create_app(
    *,
    title: str = "Webhook Bridge API",
    description: str = "A flexible webhook integration platform",
    version: str = __version__,
    plugin_dir: str | None = None,
    **kwargs: Any,
) -> FastAPI:
    """Create and configure the FastAPI application."""
    # Create main app
    main_app = FastAPI(
        title=title,
        description=description,
        version=version,
        **kwargs,
    )

    # Setup plugin directory
    if plugin_dir is not None:
        main_app.state.plugin_dir = plugin_dir


    # Add CORS middleware
    main_app.add_middleware(
        CORSMiddleware,
        allow_origins=["*"],
        allow_credentials=True,
        allow_methods=["*"],
        allow_headers=["*"],
    )

    versioned_app = setup_api(main_app)

    @versioned_app.get("/", include_in_schema=False)
    async def read_root(request: Request) -> dict[str, Any]:
        """Render the README as the home page."""
        readme_content = await get_github_readme()

        if readme_content is None:
            return {"error": "Failed to load README from GitHub"}

        html_content = markdown2.markdown(
            readme_content,
            extras=[
                "fenced-code-blocks",
                "tables",
                "toc",
            ],
        )

        return {
            "request": request,
            "content": html_content,
            "version": version,
        }

    # Setup API routes
    return versioned_app
