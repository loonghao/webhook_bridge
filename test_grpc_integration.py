#!/usr/bin/env python3
"""Integration test for gRPC communication between Go and Python services."""

# Import built-in modules
import asyncio
from pathlib import Path
import sys


# Add the project root to Python path
project_root = Path(__file__).parent
sys.path.insert(0, str(project_root))

# Import third-party modules
from api.proto import webhook_pb2
from api.proto import webhook_pb2_grpc
import grpc


async def test_python_executor():
    """Test the Python executor gRPC service."""
    print("🔍 Testing Python executor gRPC service...")

    # Connect to the Python executor
    channel = grpc.aio.insecure_channel('localhost:50051')
    stub = webhook_pb2_grpc.WebhookExecutorStub(channel)

    try:
        # Test health check
        print("  ✓ Testing health check...")
        health_request = webhook_pb2.HealthCheckRequest(service="test")
        health_response = await stub.HealthCheck(health_request)
        print(f"    Health status: {health_response.status}")
        print(f"    Message: {health_response.message}")

        # Test list plugins
        print("  ✓ Testing list plugins...")
        list_request = webhook_pb2.ListPluginsRequest()
        list_response = await stub.ListPlugins(list_request)
        print(f"    Found {list_response.total_count} plugins")
        for plugin in list_response.plugins:
            print(f"    - {plugin.name}: {plugin.description}")

        # Test plugin execution (if plugins are available)
        if list_response.total_count > 0:
            plugin_name = list_response.plugins[0].name
            print(f"  ✓ Testing plugin execution: {plugin_name}")

            exec_request = webhook_pb2.ExecutePluginRequest(
                plugin_name=plugin_name,
                http_method="GET",
                data={"test": "data"},
            )
            exec_response = await stub.ExecutePlugin(exec_request)
            print(f"    Status: {exec_response.status_code}")
            print(f"    Message: {exec_response.message}")
            print(f"    Execution time: {exec_response.execution_time:.3f}s")

        print("✅ Python executor tests passed!")
        return True

    except Exception as e:
        print(f"❌ Python executor test failed: {e}")
        return False
    finally:
        await channel.close()


def test_go_server():
    """Test the Go HTTP server."""
    print("🔍 Testing Go HTTP server...")

    try:
        # Import third-party modules
        import requests

        # Test health endpoint
        print("  ✓ Testing health endpoint...")
        response = requests.get("http://localhost:8000/health", timeout=5)
        if response.status_code == 200:
            print(f"    Health check: {response.json()}")
        else:
            print(f"    Health check failed: {response.status_code}")
            return False

        # Test list plugins endpoint
        print("  ✓ Testing list plugins endpoint...")
        response = requests.get("http://localhost:8000/api/v1/plugins", timeout=5)
        if response.status_code == 200:
            data = response.json()
            print(f"    Found {data.get('total_count', 0)} plugins")
        else:
            print(f"    List plugins failed: {response.status_code}")
            return False

        print("✅ Go server tests passed!")
        return True

    except ImportError:
        print("⚠️  requests library not available, skipping HTTP tests")
        return True
    except Exception as e:
        print(f"❌ Go server test failed: {e}")
        return False


async def main():
    """Main test function."""
    print("🚀 Starting webhook bridge integration tests...")
    print()

    # Check if services are running
    print("📋 Checking service availability...")

    # Test Python executor
    python_ok = await test_python_executor()
    print()

    # Test Go server
    go_ok = test_go_server()
    print()

    # Summary
    print("📊 Test Summary:")
    print(f"  Python Executor: {'✅ PASS' if python_ok else '❌ FAIL'}")
    print(f"  Go HTTP Server:  {'✅ PASS' if go_ok else '❌ FAIL'}")
    print()

    if python_ok and go_ok:
        print("🎉 All integration tests passed!")
        return 0
    else:
        print("💥 Some tests failed!")
        return 1


if __name__ == "__main__":
    try:
        exit_code = asyncio.run(main())
        sys.exit(exit_code)
    except KeyboardInterrupt:
        print("\n⏹️  Tests interrupted by user")
        sys.exit(1)
    except Exception as e:
        print(f"\n💥 Test runner error: {e}")
        sys.exit(1)
