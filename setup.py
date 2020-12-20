"""Describe the distribution to distutils."""
import re

from setuptools import find_packages, setup

with open("README.md", "r") as fh:
    long_description = fh.read()

setup(
    name="webhook_bridge",
    author="Long Hao",
    package_dir={"": "."},
    url="https://github.com/loonghao/webhook_bridge",
    packages=find_packages("."),
    use_scm_version=True,
    setup_requires=["setuptools_scm"],
    author_email="hal.long@outlook.com",
    description="Bridge Webhook into your tool or internal integration.",
    long_description=long_description,
    entry_points={
        "console_scripts": ["webhook-bridge = webhook_bridge.server:start_server"]
    },
    long_description_content_type="text/markdown",
    classifiers=[
        "Development Status :: 4 - Beta",
        "License :: OSI Approved :: MIT License",
        "Programming Language :: Python :: 3 :: Only",
        "Programming Language :: Python :: 3.6",
        "Programming Language :: Python :: 3.7",
        "Programming Language :: Python :: 3.8",
        "Topic :: Software Development :: Libraries :: Python Modules",
    ],
)
