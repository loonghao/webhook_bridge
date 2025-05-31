#!/usr/bin/env python3
"""Test script for Python Executor gRPC Service."""

import asyncio
import json
import subprocess
import sys
import time
from pathlib import Path

# Add project root to path
project_root = Path(__file__).parent
sys.path.insert(0, str(project_root))

import grpc
from api.proto import webhook_pb2
from api.proto import webhook_pb2_grpc


class PythonExecutorTester:
    """Test the Python executor gRPC service."""
    
    def __init__(self, host: str = "localhost", port: int = 50051):
        self.host = host
        self.port = port
        self.channel = None
        self.stub = None
        
    async def __aenter__(self):
        self.channel = grpc.aio.insecure_channel(f'{self.host}:{self.port}')
        self.stub = webhook_pb2_grpc.WebhookExecutorStub(self.channel)
        return self
        
    async def __aexit__(self, exc_type, exc_val, exc_tb):
        if self.channel:
            await self.channel.close()
    
    async def test_health_check(self) -> bool:
        """Test health check functionality."""
        print("🔍 Testing health check...")
        
        try:
            request = webhook_pb2.HealthCheckRequest(service="test")
            response = await self.stub.HealthCheck(request)
            
            print(f"  Status: {response.status}")
            print(f"  Message: {response.message}")
            print(f"  Details: {dict(response.details)}")
            
            if response.status in ["healthy", "degraded"]:
                print("  ✅ Health check passed")
                return True
            else:
                print("  ❌ Health check failed")
                return False
                
        except Exception as e:
            print(f"  ❌ Health check error: {e}")
            return False
    
    async def test_list_plugins(self) -> bool:
        """Test plugin listing functionality."""
        print("🔍 Testing plugin listing...")
        
        try:
            request = webhook_pb2.ListPluginsRequest()
            response = await self.stub.ListPlugins(request)
            
            print(f"  Found {response.total_count} plugins:")
            for plugin in response.plugins:
                status = "✅" if plugin.is_available else "❌"
                print(f"    {status} {plugin.name}: {plugin.description}")
                print(f"      Path: {plugin.path}")
                print(f"      Methods: {', '.join(plugin.supported_methods)}")
                if plugin.last_modified:
                    print(f"      Modified: {plugin.last_modified}")
            
            if response.total_count > 0:
                print("  ✅ Plugin listing passed")
                return True
            else:
                print("  ⚠️  No plugins found")
                return True  # Not necessarily a failure
                
        except Exception as e:
            print(f"  ❌ Plugin listing error: {e}")
            return False
    
    async def test_plugin_execution(self) -> bool:
        """Test plugin execution functionality."""
        print("🔍 Testing plugin execution...")
        
        try:
            # First get available plugins
            list_request = webhook_pb2.ListPluginsRequest()
            list_response = await self.stub.ListPlugins(list_request)
            
            if list_response.total_count == 0:
                print("  ⚠️  No plugins available for testing")
                return True
            
            # Test with the first available plugin
            plugin = list_response.plugins[0]
            plugin_name = plugin.name
            
            print(f"  Testing plugin: {plugin_name}")
            
            # Test different HTTP methods
            test_data = {"test_key": "test_value", "timestamp": str(time.time())}
            methods_to_test = ["GET", "POST", "PUT", "DELETE"]
            
            success_count = 0
            for method in methods_to_test:
                if method in plugin.supported_methods:
                    try:
                        request = webhook_pb2.ExecutePluginRequest(
                            plugin_name=plugin_name,
                            http_method=method,
                            data=test_data
                        )
                        
                        response = await self.stub.ExecutePlugin(request)
                        
                        if response.status_code == 200:
                            print(f"    ✅ {method}: {response.message} ({response.execution_time:.3f}s)")
                            success_count += 1
                        else:
                            print(f"    ❌ {method}: {response.message} (status: {response.status_code})")
                            if response.error:
                                print(f"       Error: {response.error}")
                    except Exception as e:
                        print(f"    ❌ {method}: Exception - {e}")
                else:
                    print(f"    ⏭️  {method}: Not supported")
            
            if success_count > 0:
                print(f"  ✅ Plugin execution passed ({success_count} methods successful)")
                return True
            else:
                print("  ❌ Plugin execution failed (no methods successful)")
                return False
                
        except Exception as e:
            print(f"  ❌ Plugin execution error: {e}")
            return False
    
    async def test_plugin_info(self) -> bool:
        """Test plugin info functionality."""
        print("🔍 Testing plugin info...")
        
        try:
            # First get available plugins
            list_request = webhook_pb2.ListPluginsRequest()
            list_response = await self.stub.ListPlugins(list_request)
            
            if list_response.total_count == 0:
                print("  ⚠️  No plugins available for testing")
                return True
            
            # Test with the first available plugin
            plugin_name = list_response.plugins[0].name
            
            request = webhook_pb2.GetPluginInfoRequest(plugin_name=plugin_name)
            response = await self.stub.GetPluginInfo(request)
            
            if response.found:
                plugin = response.plugin
                print(f"  Plugin: {plugin.name}")
                print(f"  Description: {plugin.description}")
                print(f"  Path: {plugin.path}")
                print(f"  Available: {plugin.is_available}")
                print(f"  Methods: {', '.join(plugin.supported_methods)}")
                print("  ✅ Plugin info passed")
                return True
            else:
                print(f"  ❌ Plugin {plugin_name} not found")
                return False
                
        except Exception as e:
            print(f"  ❌ Plugin info error: {e}")
            return False
    
    async def test_error_handling(self) -> bool:
        """Test error handling scenarios."""
        print("🔍 Testing error handling...")
        
        try:
            # Test with non-existent plugin
            request = webhook_pb2.ExecutePluginRequest(
                plugin_name="non_existent_plugin",
                http_method="GET",
                data={"test": "data"}
            )
            
            response = await self.stub.ExecutePlugin(request)
            
            if response.status_code == 404:
                print("  ✅ Non-existent plugin properly handled (404)")
                return True
            else:
                print(f"  ❌ Unexpected status code for non-existent plugin: {response.status_code}")
                return False
                
        except Exception as e:
            print(f"  ❌ Error handling test error: {e}")
            return False


def start_python_executor(port: int = 50051) -> subprocess.Popen:
    """Start the Python executor service."""
    print(f"🚀 Starting Python executor on port {port}...")
    
    # Use the virtual environment Python
    python_path = ".venv/Scripts/python.exe"
    
    process = subprocess.Popen([
        python_path, "python_executor/main.py",
        "--host", "localhost",
        "--port", str(port),
        "--log-level", "INFO",
        "--plugin-dirs", "example_plugins"
    ], stdout=subprocess.PIPE, stderr=subprocess.PIPE, text=True)
    
    # Wait a moment for the server to start
    time.sleep(3)
    
    return process


async def main():
    """Main test function."""
    print("🧪 Python Executor gRPC Service Test Suite")
    print("=" * 60)
    
    port = 50055
    process = None
    
    try:
        # Start the Python executor service
        process = start_python_executor(port)
        
        # Check if process started successfully
        if process.poll() is not None:
            stdout, stderr = process.communicate()
            print(f"❌ Failed to start Python executor:")
            print(f"STDOUT: {stdout}")
            print(f"STDERR: {stderr}")
            return 1
        
        print("✅ Python executor started successfully")
        
        # Run tests
        async with PythonExecutorTester("localhost", port) as tester:
            tests = [
                ("Health Check", tester.test_health_check),
                ("List Plugins", tester.test_list_plugins),
                ("Plugin Execution", tester.test_plugin_execution),
                ("Plugin Info", tester.test_plugin_info),
                ("Error Handling", tester.test_error_handling),
            ]
            
            passed = 0
            total = len(tests)
            
            for test_name, test_func in tests:
                print(f"\n{'='*20} {test_name} {'='*20}")
                try:
                    if await test_func():
                        print(f"✅ {test_name} PASSED")
                        passed += 1
                    else:
                        print(f"❌ {test_name} FAILED")
                except Exception as e:
                    print(f"💥 {test_name} ERROR: {e}")
            
            print(f"\n{'='*60}")
            print(f"📊 Test Results: {passed}/{total} tests passed")
            
            if passed == total:
                print("🎉 All tests passed!")
                return 0
            else:
                print("💥 Some tests failed!")
                return 1
    
    finally:
        # Stop the Python executor service
        if process:
            print("\n🛑 Stopping Python executor...")
            process.terminate()
            try:
                process.wait(timeout=5)
            except subprocess.TimeoutExpired:
                process.kill()
                process.wait()
            print("✅ Python executor stopped")


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
