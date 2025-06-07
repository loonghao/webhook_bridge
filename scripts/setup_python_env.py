#!/usr/bin/env python3
"""
Setup Python environment for webhook_bridge.

This script ensures that the Python environment is properly configured
for running the Python executor.
"""

# Import built-in modules
import logging
import os
from pathlib import Path
import subprocess
import sys


# Setup logging
logging.basicConfig(level=logging.INFO, format='%(asctime)s - %(levelname)s - %(message)s')
logger = logging.getLogger(__name__)

def check_python_version():
    """Check if Python version is compatible."""
    version = sys.version_info
    if version.major < 3 or (version.major == 3 and version.minor < 8):
        logger.error(f"Python 3.8+ required, found {version.major}.{version.minor}")
        return False
    logger.info(f"Python version: {version.major}.{version.minor}.{version.micro}")
    return True

def check_virtual_env():
    """Check if we're in a virtual environment."""
    in_venv = hasattr(sys, 'real_prefix') or (hasattr(sys, 'base_prefix') and sys.base_prefix != sys.prefix)
    if in_venv:
        logger.info(f"Virtual environment detected: {sys.prefix}")
    else:
        logger.warning("Not in a virtual environment")
    return in_venv

def install_requirements():
    """Install Python requirements."""
    requirements_file = Path("python_executor/requirements.txt")
    if not requirements_file.exists():
        logger.error(f"Requirements file not found: {requirements_file}")
        return False
    
    try:
        logger.info("Installing Python requirements...")
        subprocess.run([sys.executable, "-m", "pip", "install", "-r", str(requirements_file)],
                      check=True, capture_output=True, text=True)
        logger.info("Requirements installed successfully")
        return True
    except subprocess.CalledProcessError as e:
        logger.error(f"Failed to install requirements: {e}")
        logger.error(f"STDOUT: {e.stdout}")
        logger.error(f"STDERR: {e.stderr}")
        return False

def check_grpc_installation():
    """Check if gRPC is properly installed."""
    try:
        # Import third-party modules
        import grpc
        logger.info(f"gRPC version: {grpc.__version__}")
        return True
    except ImportError:
        logger.error("gRPC not installed")
        return False

def check_webhook_bridge_module():
    """Check if webhook_bridge module can be imported."""
    try:
        # Add current directory to Python path
        current_dir = Path.cwd()
        if str(current_dir) not in sys.path:
            sys.path.insert(0, str(current_dir))
        
        # Import local modules
        from webhook_bridge.filesystem import get_plugins
        from webhook_bridge.plugin import BasePlugin
        logger.info("webhook_bridge module imported successfully")
        
        # Test plugin discovery
        plugins = get_plugins()
        logger.info(f"Found {len(plugins)} plugins")
        return True
    except ImportError as e:
        logger.error(f"Failed to import webhook_bridge module: {e}")
        return False

def create_example_config():
    """Create example configuration if it doesn't exist."""
    config_file = Path("config.yaml")
    if config_file.exists():
        logger.info("Configuration file already exists")
        return True
    
    config_content = """# Webhook Bridge Configuration
server:
  host: "0.0.0.0"
  port: 8000
  mode: "debug"

executor:
  host: "localhost"
  port: 50051

logging:
  level: "info"
  file: "logs/webhook_bridge.log"
  max_size: 100
  max_backups: 3
  max_age: 30

plugins:
  directories:
    - "example_plugins"
"""
    
    try:
        config_file.write_text(config_content)
        logger.info(f"Created example configuration: {config_file}")
        return True
    except Exception as e:
        logger.error(f"Failed to create configuration: {e}")
        return False

def main():
    """Main setup function."""
    logger.info("🔧 Setting up Python environment for webhook_bridge")
    
    success = True
    
    # Check Python version
    if not check_python_version():
        success = False
    
    # Check virtual environment
    check_virtual_env()
    
    # Install requirements
    if not install_requirements():
        success = False
    
    # Check gRPC installation
    if not check_grpc_installation():
        success = False
    
    # Check webhook_bridge module
    if not check_webhook_bridge_module():
        success = False
    
    # Create example config
    if not create_example_config():
        success = False
    
    if success:
        logger.info("✅ Python environment setup completed successfully")
        return 0
    else:
        logger.error("❌ Python environment setup failed")
        return 1

if __name__ == "__main__":
    sys.exit(main())
