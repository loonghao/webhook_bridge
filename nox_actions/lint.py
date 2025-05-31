# Import third-party modules
import nox
from nox_actions.utils import PACKAGE_NAME


def lint(session: nox.Session) -> None:
    session.install("isort", "ruff")
    session.run("isort", "--check-only", PACKAGE_NAME)
    session.run("ruff", "check")


def lint_fix(session: nox.Session) -> None:
    session.install("isort", "ruff", "autoflake")
    session.run("ruff", "check", "--fix")
    session.run("isort", ".")
    # Skip pre-commit for now due to Windows compatibility issues
    # session.run("pre-commit", "run", "--all-files")
    session.run("autoflake", "--in-place", "--remove-all-unused-imports", "--remove-unused-variables", ".")
