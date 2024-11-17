# Import built-in modules
import os
import webbrowser

# Import third-party modules
import nox
from nox_actions.utils import THIS_ROOT


def local_test(session: nox.Session) -> None:
    print( os.path.join(THIS_ROOT, "example_plugins"))
    os.environ["WEBHOOK_BRIDGE_SERVER_PLUGINS"] = os.path.join(THIS_ROOT, "example_plugins")
    port = "54002"
    webbrowser.open_new_tab(f"http://127.0.0.1:{port}/api/v1/docs")
    session.run("uvicorn", "webhook_bridge.server:APP", "--reload", "--host", "127.0.0.1", "--port", port)
