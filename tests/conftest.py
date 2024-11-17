# Import built-in modules
import os

# Import third-party modules
import pytest


@pytest.fixture()
def test_data_root():
    return os.path.join(os.path.dirname(__file__), "test_data")
