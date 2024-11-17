# Import built-in modules
from abc import ABC
from abc import abstractmethod
import importlib.machinery
import importlib.util
import logging
import os
from typing import Any
from typing import Dict
from typing import Type

# Import third-party modules
from addict import Addict


class BasePlugin(ABC):

    def __init__(self, data: Dict[str, Any], logger: logging.Logger = None):
        self.data = Addict(data)
        self.logger = logger or logging.getLogger(__name__)

    @abstractmethod
    def run(self) -> Dict[str, Any]:
        """Subclasses should implement this method to return a dictionary."""
        raise NotImplementedError("Subclasses should implement this method")


def load_plugin(pyfile: str) -> Type[BasePlugin]:
    """Load a plugin from a Python file."""
    name = os.path.basename(pyfile).split(".")[0]
    loader = importlib.machinery.SourceFileLoader(name, pyfile)
    spec = importlib.util.spec_from_loader(loader.name, loader)

    if spec is None:
        raise ImportError(f"Could not create a module spec for {pyfile}")

    mod = importlib.util.module_from_spec(spec)
    loader.exec_module(mod)
    if not issubclass(mod.Plugin, BasePlugin):
        raise TypeError(f"{mod.Plugin} is not a subclass of BasePlugin")
    return mod.Plugin  # type: ignore
