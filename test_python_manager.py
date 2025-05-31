#!/usr/bin/env python3
"""Test script for Python Manager functionality."""

# Import built-in modules
import json
import os
from pathlib import Path
import subprocess
import sys


def run_command(cmd, capture_output=True, check=True):
    """Run a command and return the result."""
    print(f"Running: {' '.join(cmd)}")
    result = subprocess.run(
        cmd,
        capture_output=capture_output,
        text=True,
        check=check,
    )
    if capture_output:
        print(f"Output: {result.stdout}")
        if result.stderr:
            print(f"Error: {result.stderr}")
    return result


def test_python_manager_binary():
    """Test the Python manager binary."""
    print("🔍 Testing Python Manager Binary...")

    # Build the binary first
    print("Building python-manager binary...")
    try:
        run_command(["go", "build", "-o", "bin/python-manager", "./cmd/python-manager"])
        print("✅ Binary built successfully")
    except subprocess.CalledProcessError as e:
        print(f"❌ Failed to build binary: {e}")
        return False

    # Test basic functionality
    print("\n📋 Testing basic info...")
    try:
        result = run_command(["./bin/python-manager"])
        print("✅ Basic info test passed")
    except subprocess.CalledProcessError as e:
        print(f"❌ Basic info test failed: {e}")
        return False

    # Test detailed info
    print("\n📊 Testing detailed info...")
    try:
        result = run_command(["./bin/python-manager", "--info"])
        # Try to parse JSON output
        lines = result.stdout.split('\n')
        json_started = False
        json_lines = []

        for line in lines:
            if line.strip().startswith('{'):
                json_started = True
            if json_started:
                json_lines.append(line)
                if line.strip().endswith('}') and line.count('}') >= line.count('{'):
                    break

        if json_lines:
            json_str = '\n'.join(json_lines)
            interpreter_info = json.loads(json_str)
            print("✅ Detailed info test passed")
            print(f"   Python path: {interpreter_info.get('path', 'N/A')}")
            print(f"   Python version: {interpreter_info.get('version', 'N/A')}")
            print(f"   Strategy: {interpreter_info.get('strategy', 'N/A')}")
            print(f"   Virtual env: {interpreter_info.get('is_virtual', 'N/A')}")
        else:
            print("⚠️  Could not parse JSON output")

    except subprocess.CalledProcessError as e:
        print(f"❌ Detailed info test failed: {e}")
        return False
    except json.JSONDecodeError as e:
        print(f"⚠️  JSON parsing failed: {e}")

    # Test validation
    print("\n✅ Testing environment validation...")
    try:
        result = run_command(["./bin/python-manager", "--validate"])
        print("✅ Validation test passed")
    except subprocess.CalledProcessError as e:
        print(f"❌ Validation test failed: {e}")
        return False

    # Test different strategies
    strategies = ["auto", "path"]
    for strategy in strategies:
        print(f"\n🔧 Testing strategy: {strategy}")
        try:
            result = run_command([
                "./bin/python-manager",
                "--strategy", strategy,
                "--info",
            ])
            print(f"✅ Strategy {strategy} test passed")
        except subprocess.CalledProcessError as e:
            print(f"⚠️  Strategy {strategy} test failed: {e}")

    return True


def test_uv_integration():
    """Test UV integration if available."""
    print("\n🔧 Testing UV Integration...")

    # Check if UV is available
    try:
        result = run_command(["uv", "--version"])
        print(f"✅ UV is available: {result.stdout.strip()}")
    except (subprocess.CalledProcessError, FileNotFoundError):
        print("⚠️  UV not available, skipping UV tests")
        return True

    # Test UV strategy
    print("Testing UV strategy...")
    try:
        result = run_command([
            "./bin/python-manager",
            "--strategy", "uv",
            "--info",
        ])
        print("✅ UV strategy test passed")
    except subprocess.CalledProcessError as e:
        print(f"⚠️  UV strategy test failed: {e}")

    return True


def test_virtual_environment_detection():
    """Test virtual environment detection."""
    print("\n🏠 Testing Virtual Environment Detection...")

    # Check current environment
    venv_path = os.environ.get('VIRTUAL_ENV')
    if venv_path:
        print(f"✅ Currently in virtual environment: {venv_path}")
    else:
        print("ℹ️  Not currently in a virtual environment")

    # Test with current environment
    try:
        result = run_command(["./bin/python-manager", "--info", "--verbose"])
        print("✅ Virtual environment detection test completed")
    except subprocess.CalledProcessError as e:
        print(f"❌ Virtual environment detection test failed: {e}")
        return False

    return True


def test_capability_detection():
    """Test Python capability detection."""
    print("\n🔧 Testing Capability Detection...")

    # Get detailed info to check capabilities
    try:
        result = run_command(["./bin/python-manager", "--info"])

        # Look for capability information in output
        if "capabilities" in result.stdout.lower():
            print("✅ Capability detection working")
        else:
            print("⚠️  Capability information not found in output")

    except subprocess.CalledProcessError as e:
        print(f"❌ Capability detection test failed: {e}")
        return False

    return True


def test_error_handling():
    """Test error handling scenarios."""
    print("\n❌ Testing Error Handling...")

    # Test with invalid strategy
    print("Testing invalid strategy...")
    try:
        result = run_command([
            "./bin/python-manager",
            "--strategy", "invalid_strategy",
        ], check=False)

        if result.returncode != 0:
            print("✅ Invalid strategy properly rejected")
        else:
            print("⚠️  Invalid strategy was accepted")

    except Exception as e:
        print(f"⚠️  Error handling test exception: {e}")

    # Test with invalid config
    print("Testing with non-existent config...")
    try:
        result = run_command([
            "./bin/python-manager",
            "--config", "/nonexistent/config.yaml",
        ], check=False)

        if result.returncode != 0:
            print("✅ Non-existent config properly handled")
        else:
            print("⚠️  Non-existent config was accepted")

    except Exception as e:
        print(f"⚠️  Config error handling test exception: {e}")

    return True


def main():
    """Main test function."""
    print("🚀 Python Manager Test Suite")
    print("=" * 50)

    # Change to project root
    project_root = Path(__file__).parent
    os.chdir(project_root)

    tests = [
        ("Python Manager Binary", test_python_manager_binary),
        ("UV Integration", test_uv_integration),
        ("Virtual Environment Detection", test_virtual_environment_detection),
        ("Capability Detection", test_capability_detection),
        ("Error Handling", test_error_handling),
    ]

    passed = 0
    total = len(tests)

    for test_name, test_func in tests:
        print(f"\n{'='*20} {test_name} {'='*20}")
        try:
            if test_func():
                print(f"✅ {test_name} PASSED")
                passed += 1
            else:
                print(f"❌ {test_name} FAILED")
        except Exception as e:
            print(f"💥 {test_name} ERROR: {e}")

    print(f"\n{'='*50}")
    print(f"📊 Test Results: {passed}/{total} tests passed")

    if passed == total:
        print("🎉 All tests passed!")
        return 0
    else:
        print("💥 Some tests failed!")
        return 1


if __name__ == "__main__":
    sys.exit(main())
