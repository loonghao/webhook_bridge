"""Test cases for the API endpoints."""
# Import built-in modules
from pathlib import Path

# Import third-party modules
from fastapi import FastAPI
from fastapi.testclient import TestClient
import pytest
from starlette import status

# Import local modules
from webhook_bridge.api.call_plugin import api as call_plugin_api
from webhook_bridge.api.list_plugins import api as list_plugins_api


@pytest.fixture
def app() -> FastAPI:
    """Create a FastAPI test application.

    Returns:
        FastAPI: The test application instance
    """
    app = FastAPI()
    app.state.plugin_dir = str(Path(__file__).parent / "test_plugins")
    call_plugin_api(app)
    list_plugins_api(app)
    return app


@pytest.fixture
def client(app: FastAPI) -> TestClient:
    """Create a test client.

    Args:
        app: The FastAPI test application

    Returns:
        TestClient: The test client instance
    """
    return TestClient(app)


@pytest.fixture
def test_plugin_dir(tmp_path: Path) -> str:
    """Create a temporary plugin directory for testing.

    Args:
        tmp_path: Pytest temporary path fixture

    Returns:
        str: Path to the temporary plugin directory
    """
    plugin_dir = tmp_path / "plugins"
    plugin_dir.mkdir()

    # Create a test plugin
    test_plugin = plugin_dir / "test_plugin.py"
    test_plugin.write_text('''
from webhook_bridge.plugin import BasePlugin

class Plugin(BasePlugin):
    def run(self):
        return {"input_data": self.data}
    ''')

    return str(plugin_dir)


def test_list_plugins_empty(client: TestClient, app: FastAPI, tmp_path: Path) -> None:
    """Test listing plugins when no plugins are available.

    Args:
        client: The test client
        app: The FastAPI application
        tmp_path: Pytest temporary path fixture
    """
    empty_dir = tmp_path / "empty"
    empty_dir.mkdir()
    app.state.plugin_dir = str(empty_dir)

    response = client.get("/plugins")
    assert response.status_code == status.HTTP_200_OK

    data = response.json()
    assert data["status_code"] == status.HTTP_200_OK
    assert data["message"] == "success"
    assert data["data"]["plugins"] == []


def test_list_plugins(client: TestClient, app: FastAPI, test_plugin_dir: str) -> None:
    """Test listing available plugins.

    Args:
        client: The test client
        app: The FastAPI application
        test_plugin_dir: Path to test plugin directory
    """
    app.state.plugin_dir = test_plugin_dir

    response = client.get("/plugins")
    assert response.status_code == status.HTTP_200_OK

    data = response.json()
    assert data["status_code"] == status.HTTP_200_OK
    assert data["message"] == "success"
    assert "test_plugin" in data["data"]["plugins"]


def test_execute_plugin_not_found(client: TestClient) -> None:
    """Test executing a non-existent plugin.

    Args:
        client: The test client
    """
    response = client.post(
        "/plugin/non_existent",
        json={"test": "data"},
    )
    assert response.status_code == status.HTTP_404_NOT_FOUND

    data = response.json()
    assert data["status_code"] == status.HTTP_404_NOT_FOUND
    assert "not found" in data["message"].lower()


def test_execute_plugin_success(
    client: TestClient,
    app: FastAPI,
    test_plugin_dir: str,
) -> None:
    """Test successful plugin execution.

    Args:
        client: The test client
        app: The FastAPI application
        test_plugin_dir: Path to test plugin directory
    """
    app.state.plugin_dir = test_plugin_dir
    test_data = {"test": "data"}

    response = client.post(
        "/plugin/test_plugin",
        json=test_data,
    )
    assert response.status_code == status.HTTP_200_OK

    data = response.json()
    assert data["status_code"] == status.HTTP_200_OK
    assert data["message"] == "success"
    assert data["data"]["plugin"] == "test_plugin"
    assert data["data"]["plugin_data"]["input_data"] == test_data
