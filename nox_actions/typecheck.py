# Import third-party modules
import nox


@nox.session
def mypy(session):
    session.install("mypy", "types-Markdown", "no_implicit_optional")
    session.run("mypy")
