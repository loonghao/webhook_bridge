# Import third-party modules
import nox
from nox_actions.utils import PACKAGE_NAME

@nox.session
def mypy(session):
    session.install("mypy", "types-Markdown", "no_implicit_optional")
    session.run("mypy", PACKAGE_NAME)
