# Import third-party modules
import nox
from nox_actions.utils import PACKAGE_NAME


@nox.session(python="3.11")
def mypy(session):
    session.install("mypy", "types-Markdown", "no_implicit_optional", "click", "pydantic")
    session.run("mypy", PACKAGE_NAME)
