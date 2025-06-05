"""
Webhook Bridge Manager - Manages the Go binary lifecycle.

This module handles downloading, installing, updating, and running the
Go-based webhook bridge server binary.
"""

# Import built-in modules
import json
import os
from pathlib import Path
import platform
import shutil
import subprocess
import tarfile
import tempfile
from typing import Optional
from urllib.request import urlopen
from urllib.request import urlretrieve
import zipfile


class WebhookBridgeManager:
    """Manages the webhook bridge Go binary."""

    GITHUB_REPO = "loonghao/webhook_bridge"
    BINARY_NAME = "webhook-bridge"

    def __init__(self, verbose: bool = False, config_path: Optional[Path] = None):
        """Initialize the manager."""
        self.verbose = verbose
        self.config_path = config_path
        self.install_dir = self._get_install_dir()
        self.binary_path = self._find_binary_path()

    def _get_install_dir(self) -> Path:
        """Get the installation directory."""
        if os.name == "nt":  # Windows
            base_dir = Path.home() / "AppData" / "Local" / "webhook-bridge"
        else:  # Unix-like
            base_dir = Path.home() / ".local" / "bin" / "webhook-bridge"

        base_dir.mkdir(parents=True, exist_ok=True)
        return base_dir

    def _get_binary_name(self) -> str:
        """Get the binary name for the current platform."""
        if os.name == "nt":
            return f"{self.BINARY_NAME}.exe"
        return self.BINARY_NAME

    def _find_binary_path(self) -> Path:
        """Find the binary path, checking multiple locations."""
        binary_name = self._get_binary_name()

        # 1. Check current working directory (for local development)
        local_binary = Path.cwd() / binary_name
        if local_binary.exists():
            self._log(f"Found local binary: {local_binary}")
            return local_binary

        # 2. Check if binary is in the same directory as this Python package
        package_dir = Path(__file__).parent.parent
        package_binary = package_dir / binary_name
        if package_binary.exists():
            self._log(f"Found package binary: {package_binary}")
            return package_binary

        # 3. Check system PATH
        system_binary = shutil.which(binary_name)
        if system_binary:
            self._log(f"Found system binary: {system_binary}")
            return Path(system_binary)

        # 4. Fall back to installation directory
        install_binary = self.install_dir / binary_name
        self._log(f"Using install directory: {install_binary}")
        return install_binary

    def _get_platform_info(self) -> tuple[str, str]:
        """Get platform and architecture information."""
        system = platform.system().lower()
        machine = platform.machine().lower()

        # Normalize system name
        if system == "windows":
            system = "windows"
        elif system == "darwin":
            system = "darwin"
        else:
            system = "linux"

        # Normalize architecture
        if machine in ("x86_64", "amd64"):
            arch = "amd64"
        elif machine in ("aarch64", "arm64"):
            arch = "arm64"
        else:
            arch = "amd64"  # Default fallback

        return system, arch

    def _log(self, message: str) -> None:
        """Log a message if verbose mode is enabled."""
        if self.verbose:
            print(f"🔧 {message}")

    def _get_latest_release(self) -> dict:
        """Get information about the latest release."""
        url = f"https://api.github.com/repos/{self.GITHUB_REPO}/releases/latest"
        self._log(f"Fetching latest release info from {url}")

        with urlopen(url) as response:
            return json.loads(response.read().decode())

    def _download_binary(self, version: Optional[str] = None) -> Path:
        """Download the binary for the current platform."""
        if version:
            release_url = f"https://api.github.com/repos/{self.GITHUB_REPO}/releases/tags/{version}"
            with urlopen(release_url) as response:
                release_info = json.loads(response.read().decode())
        else:
            release_info = self._get_latest_release()
            version = release_info["tag_name"]

        system, arch = self._get_platform_info()

        # Find the appropriate asset
        asset_name = f"{self.BINARY_NAME}-{system}-{arch}"
        if system == "windows":
            asset_name += ".zip"
        else:
            asset_name += ".tar.gz"

        download_url = None
        for asset in release_info["assets"]:
            if asset["name"] == asset_name:
                download_url = asset["browser_download_url"]
                break

        if not download_url:
            raise RuntimeError(f"No binary found for {system}/{arch}")

        self._log(f"Downloading {asset_name} from {download_url}")

        # Download to temporary file
        with tempfile.NamedTemporaryFile(delete=False, suffix=Path(asset_name).suffix) as tmp_file:
            urlretrieve(download_url, tmp_file.name)
            return Path(tmp_file.name)

    def _extract_binary(self, archive_path: Path) -> Path:
        """Extract the binary from the downloaded archive."""
        extract_dir = Path(tempfile.mkdtemp())

        if archive_path.suffix == ".zip":
            with zipfile.ZipFile(archive_path, 'r') as zip_file:
                zip_file.extractall(extract_dir)
        else:  # .tar.gz
            with tarfile.open(archive_path, 'r:gz') as tar_file:
                tar_file.extractall(extract_dir)

        # Find the binary in the extracted files
        binary_name = self._get_binary_name()
        for file_path in extract_dir.rglob("*"):
            if file_path.name == binary_name or file_path.name.startswith(self.BINARY_NAME):
                return file_path

        raise RuntimeError(f"Binary {binary_name} not found in archive")

    @staticmethod
    def get_version() -> str:
        """Get the version of this Python package."""
        from . import __version__
        return __version__

    def install(self, force: bool = False, version: Optional[str] = None) -> int:
        """Install the webhook bridge binary."""
        if self.binary_path.exists() and not force:
            print(f"✅ Webhook bridge is already installed at {self.binary_path}")
            print("   Use --force to reinstall")
            return 0

        try:
            print("📦 Installing webhook bridge...")

            # Download the binary
            archive_path = self._download_binary(version)

            # Extract the binary
            binary_path = self._extract_binary(archive_path)

            # Move to installation directory
            if self.binary_path.exists():
                self.binary_path.unlink()

            shutil.move(str(binary_path), str(self.binary_path))
            self.binary_path.chmod(0o755)  # Make executable

            # Cleanup
            archive_path.unlink()
            shutil.rmtree(binary_path.parent)

            print(f"✅ Webhook bridge installed successfully to {self.binary_path}")
            print("   Run 'webhook-bridge run' to start the server")

            return 0

        except Exception as e:
            print(f"❌ Installation failed: {e}")
            return 1

    def run(self, port: int = 8000, host: str = "0.0.0.0", daemon: bool = False) -> int:
        """Run the webhook bridge server."""
        if not self.binary_path.exists():
            print(f"❌ Webhook bridge binary not found at: {self.binary_path}")
            print("   Available options:")
            print("   1. Run 'webhook-bridge install' to download the binary")
            print("   2. Build locally with 'uvx nox -s build-local'")
            print("   3. Ensure binary is in PATH or current directory")
            return 1

        cmd = [str(self.binary_path)]

        # Add configuration if provided
        if self.config_path:
            cmd.extend(["--config", str(self.config_path)])

        # Add host and port arguments
        cmd.extend(["--host", host, "--port", str(port)])

        # Set environment variables as backup
        env = os.environ.copy()
        env["WEBHOOK_BRIDGE_HOST"] = host
        env["WEBHOOK_BRIDGE_PORT"] = str(port)

        try:
            if daemon:
                print(f"🚀 Starting webhook bridge server in daemon mode on {host}:{port}")
                subprocess.Popen(cmd, env=env)
                print("✅ Server started in background")
                return 0
            else:
                print(f"🚀 Starting webhook bridge server on {host}:{port}")
                print(f"   Binary: {self.binary_path}")
                print("   Press Ctrl+C to stop")
                return subprocess.run(cmd, env=env, check=False).returncode

        except KeyboardInterrupt:
            print("\n⚠️  Server stopped by user")
            return 0
        except Exception as e:
            print(f"❌ Failed to start server: {e}")
            return 1

    def status(self) -> int:
        """Check the status of the webhook bridge server."""
        if not self.binary_path.exists():
            print("❌ Webhook bridge is not installed")
            return 1

        # Try to get version info
        try:
            result = subprocess.run(
                [str(self.binary_path), "--version"],
                capture_output=True,
                text=True,
                timeout=5, check=False,
            )
            if result.returncode == 0:
                print(f"✅ Webhook bridge is installed: {self.binary_path}")
                print(f"   Version: {result.stdout.strip()}")
            else:
                print(f"⚠️  Binary exists but may be corrupted: {self.binary_path}")
            return result.returncode
        except Exception as e:
            print(f"❌ Failed to check status: {e}")
            return 1

    def stop(self) -> int:
        """Stop the running webhook bridge server."""
        # This is a simple implementation - in a real scenario,
        # you might want to use PID files or process management
        try:
            if os.name == "nt":
                subprocess.run(["taskkill", "/f", "/im", self._get_binary_name()], check=False)
            else:
                subprocess.run(["pkill", "-f", self.BINARY_NAME], check=False)
            print("✅ Server stop signal sent")
            return 0
        except Exception as e:
            print(f"❌ Failed to stop server: {e}")
            return 1

    def update(self, check_only: bool = False) -> int:
        """Update to the latest version."""
        try:
            release_info = self._get_latest_release()
            latest_version = release_info["tag_name"]

            if check_only:
                print(f"📋 Latest version: {latest_version}")
                return 0

            print(f"🔄 Updating to {latest_version}...")
            return self.install(force=True)

        except Exception as e:
            print(f"❌ Update failed: {e}")
            return 1

    def config(self, action: str = "show") -> int:
        """Handle configuration management."""
        if action == "show":
            if self.config_path and self.config_path.exists():
                print(f"📋 Configuration file: {self.config_path}")
                with open(self.config_path) as f:
                    print(f.read())
            else:
                print("📋 No configuration file specified or found")
                print("   Use --config to specify a configuration file")
            return 0
        elif action == "init":
            config_path = Path("config.yaml")
            if config_path.exists():
                print(f"⚠️  Configuration file already exists: {config_path}")
                return 1

            # Create a basic configuration file
            config_content = """# Webhook Bridge Configuration
server:
  host: "0.0.0.0"
  port: 8000
  mode: "debug"

python:
  interpreter: "python"
  venv_path: ".venv"

logging:
  level: "info"
  format: "text"

directories:
  working_dir: "."
  log_dir: "logs"
  plugin_dir: "plugins"
"""
            with open(config_path, 'w') as f:
                f.write(config_content)
            print(f"✅ Configuration file created: {config_path}")
            return 0
        elif action == "validate":
            if not self.config_path or not self.config_path.exists():
                print("❌ No configuration file to validate")
                return 1

            # Basic validation - in a real implementation, you'd use a schema
            try:
                # Import third-party modules
                import yaml
                with open(self.config_path) as f:
                    yaml.safe_load(f)
                print("✅ Configuration file is valid")
                return 0
            except Exception as e:
                print(f"❌ Configuration file is invalid: {e}")
                return 1
        else:
            print(f"❌ Unknown config action: {action}")
            return 1
