"""Nox sessions for the Rust 4.0 bridge and Next.js dashboard."""

from __future__ import annotations

from pathlib import Path

import nox


ROOT = Path(__file__).resolve().parent.parent


def _run_bridge(session: nox.Session, *args: str) -> None:
    with session.chdir(str(ROOT)):
        session.run(
            "cargo",
            "run",
            "-p",
            "webhook-bridge-server",
            "--bin",
            "webhook-bridge",
            "--",
            *args,
            external=True,
        )


def start_server(session: nox.Session) -> None:
    """Start the Rust API with configured Python workers."""
    _run_bridge(session, "run", "--config", "config.4.0.yaml")


def dev(session: nox.Session) -> None:
    """Run the local development server."""
    start_server(session)


def quick(session: nox.Session) -> None:
    """Run the core fast checks."""
    with session.chdir(str(ROOT)):
        session.run("cargo", "fmt", "--check", external=True)
        session.run("cargo", "test", external=True)


def build_local(session: nox.Session) -> None:
    """Build the release binary."""
    with session.chdir(str(ROOT)):
        session.run(
            "cargo",
            "build",
            "--release",
            "-p",
            "webhook-bridge-server",
            "--bin",
            "webhook-bridge",
            external=True,
        )


def test_local(session: nox.Session) -> None:
    """Run Rust tests."""
    with session.chdir(str(ROOT)):
        session.run("cargo", "test", external=True)


def run_local(session: nox.Session) -> None:
    """Run the built bridge from Cargo."""
    _run_bridge(session, "run", "--config", "config.4.0.yaml")


def clean_local(session: nox.Session) -> None:
    """Clean Rust build artifacts."""
    with session.chdir(str(ROOT)):
        session.run("cargo", "clean", external=True)


def clean_all(session: nox.Session) -> None:
    """Clean local build artifacts."""
    clean_local(session)
