# Import built-in modules
import importlib.machinery
import importlib.util
import logging
import os
from abc import ABC
from abc import abstractmethod

# Import third-party modules
from addict import Dict


class BasePlugin(ABC):

    def __init__(self, data, logger=None):
        self.data = Dict(data)
        self.logger = logger or logging.getLogger(__name__)

    @abstractmethod
    def run(self):
        pass


def load_plugin(pyfile):
    """Load plugin by give python file.

    Args:
        pyfile (str): Absolute path of the python file.

    Returns:
        webhook_bridge

    """
    name = os.path.basename(pyfile).split(".")[0]
    loader = importlib.machinery.SourceFileLoader(name, pyfile)
    spec = importlib.util.spec_from_loader(loader.name, loader)
    mod = importlib.util.module_from_spec(spec)
    loader.exec_module(mod)
    return mod.Plugin
