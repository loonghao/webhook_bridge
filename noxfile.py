# Import built-in modules
import os
import sys

# Import third-party modules
import nox


ROOT = os.path.dirname(__file__)

# Ensure maya_umbrella is importable.
if ROOT not in sys.path:
    sys.path.append(ROOT)

# Import third-party modules
from nox_actions import codetest
from nox_actions import lint
from nox_actions import web


nox.session(lint.lint, name="lint")
nox.session(lint.lint_fix, name="lint-fix")
nox.session(codetest.pytest, name="pytest")
nox.session(web.start_server, name="start-server")
nox.session(web.dev, name="dev")
nox.session(web.quick, name="quick")
nox.session(web.build_local, name="build-local")
nox.session(web.test_local, name="test-local")
nox.session(web.run_local, name="run-local")
nox.session(web.clean_local, name="clean-local")
