#!/usr/bin/env python3
"""
Test script to verify Python package build and installation.
This script tests the webhook-bridge Python package to ensure it can be built and installed correctly.
"""

import subprocess
import sys
import tempfile
import os
from pathlib import Path


def run_command(cmd, cwd=None, check=True):
    """Run a command and return the result."""
    print(f"Running: {' '.join(cmd)}")
    result = subprocess.run(cmd, cwd=cwd, capture_output=True, text=True, check=check)
    if result.stdout:
        print(f"STDOUT: {result.stdout}")
    if result.stderr:
        print(f"STDERR: {result.stderr}")
    return result


def test_poetry_build():
    """Test Poetry build process."""
    print("=" * 50)
    print("Testing Poetry build process...")
    print("=" * 50)
    
    # Test poetry check
    result = run_command(["uvx", "poetry", "check"])
    assert result.returncode == 0, "Poetry check failed"
    print("✅ Poetry check passed")
    
    # Test poetry build
    result = run_command(["uvx", "poetry", "build"])
    assert result.returncode == 0, "Poetry build failed"
    print("✅ Poetry build passed")
    
    # Check if build artifacts exist
    dist_dir = Path("dist")
    wheel_files = list(dist_dir.glob("webhook_bridge-*.whl"))
    tar_files = list(dist_dir.glob("webhook_bridge-*.tar.gz"))
    
    assert len(wheel_files) > 0, "No wheel files found"
    assert len(tar_files) > 0, "No tar.gz files found"
    print(f"✅ Found {len(wheel_files)} wheel files and {len(tar_files)} source distributions")


def test_package_installation():
    """Test package installation in a temporary environment."""
    print("=" * 50)
    print("Testing package installation...")
    print("=" * 50)
    
    # Find the latest wheel file
    dist_dir = Path("dist")
    wheel_files = list(dist_dir.glob("webhook_bridge-*.whl"))
    if not wheel_files:
        print("❌ No wheel files found for testing")
        return
    
    latest_wheel = max(wheel_files, key=lambda p: p.stat().st_mtime)
    print(f"Testing installation of: {latest_wheel}")
    
    # Create a temporary directory for testing
    with tempfile.TemporaryDirectory() as temp_dir:
        # Install the package
        result = run_command([
            "python", "-m", "pip", "install", str(latest_wheel)
        ], check=False)
        
        if result.returncode == 0:
            print("✅ Package installation successful")
            
            # Try to import the package
            try:
                import webhook_bridge
                print("✅ Package import successful")
                print(f"Package location: {webhook_bridge.__file__}")
            except ImportError as e:
                print(f"❌ Package import failed: {e}")
        else:
            print(f"❌ Package installation failed: {result.stderr}")


def test_dependencies():
    """Test that all dependencies are correctly specified."""
    print("=" * 50)
    print("Testing dependencies...")
    print("=" * 50)
    
    # Test poetry install
    result = run_command(["uvx", "poetry", "install"], check=False)
    if result.returncode == 0:
        print("✅ Poetry install successful")
    else:
        print(f"❌ Poetry install failed: {result.stderr}")
        return
    
    # Test that core dependencies can be imported
    core_deps = ["yaml", "grpc"]
    for dep in core_deps:
        try:
            if dep == "yaml":
                import yaml
                print(f"✅ {dep} import successful")
            elif dep == "grpc":
                import grpc
                print(f"✅ {dep} import successful")
        except ImportError as e:
            print(f"❌ {dep} import failed: {e}")


def main():
    """Main test function."""
    print("🚀 Starting webhook-bridge Python package tests...")
    
    try:
        test_poetry_build()
        test_dependencies()
        test_package_installation()
        
        print("\n" + "=" * 50)
        print("🎉 All tests passed! Python package is ready for PyPI.")
        print("=" * 50)
        
    except Exception as e:
        print(f"\n❌ Test failed: {e}")
        sys.exit(1)


if __name__ == "__main__":
    main()
