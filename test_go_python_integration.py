#!/usr/bin/env python3
"""Integration test for Go-Python gRPC communication."""

import asyncio
import json
import socket
import subprocess
import sys
import time
import requests
from pathlib import Path

# Add project root to path
project_root = Path(__file__).parent
sys.path.insert(0, str(project_root))

def get_free_port():
    """Find and return a free port."""
    with socket.socket(socket.AF_INET, socket.SOCK_STREAM) as s:
        s.bind(('localhost', 0))
        s.listen(1)
        port = s.getsockname()[1]
    return port


class GoServerTester:
    """Test the Go HTTP server with Python executor backend."""
    
    def __init__(self, go_port: int = None, python_port: int = None):
        # Auto-assign ports if not specified
        self.python_port = python_port or get_free_port()
        self.go_port = go_port or get_free_port()
        self.go_base_url = f"http://localhost:{self.go_port}"
        self.python_process = None
        self.go_process = None
        
    def start_python_executor(self):
        """Start the Python executor service."""
        print(f"🐍 Starting Python executor on port {self.python_port}...")
        
        python_path = ".venv/Scripts/python.exe"
        self.python_process = subprocess.Popen([
            python_path, "python_executor/main.py",
            "--host", "localhost",
            "--port", str(self.python_port),
            "--plugin-dirs", "example_plugins",
            "--log-level", "INFO"
        ], stdout=subprocess.PIPE, stderr=subprocess.PIPE, text=True)
        
        # Wait for Python service to start
        time.sleep(3)
        
        if self.python_process.poll() is not None:
            stdout, stderr = self.python_process.communicate()
            raise RuntimeError(f"Python executor failed to start:\nSTDOUT: {stdout}\nSTDERR: {stderr}")
        
        print("✅ Python executor started successfully")
    
    def start_go_server(self):
        """Start the Go HTTP server."""
        print(f"🚀 Starting Go server on port {self.go_port}...")
        
        # Build the Go server first
        go_exe = r"C:\Program Files\Go\bin\go.exe"
        build_result = subprocess.run([go_exe, "build", "-o", "bin/webhook-bridge-server", "./cmd/server"],
                                    capture_output=True, text=True)
        if build_result.returncode != 0:
            raise RuntimeError(f"Failed to build Go server: {build_result.stderr}")
        
        # Start the Go server with environment variables for port configuration
        import os
        env = os.environ.copy()  # Copy current environment
        env.update({
            "WEBHOOK_BRIDGE_PORT": str(self.go_port),
            "WEBHOOK_BRIDGE_EXECUTOR_PORT": str(self.python_port)
        })
        self.go_process = subprocess.Popen([
            "./bin/webhook-bridge-server"
        ], stdout=subprocess.PIPE, stderr=subprocess.PIPE, text=True, env=env)
        
        # Wait for Go service to start
        time.sleep(3)
        
        if self.go_process.poll() is not None:
            stdout, stderr = self.go_process.communicate()
            print(f"❌ Go server failed to start:")
            print(f"STDOUT: {stdout}")
            print(f"STDERR: {stderr}")
            raise RuntimeError(f"Go server failed to start")

        print("✅ Go server started successfully")
    
    def stop_services(self):
        """Stop both services."""
        print("\n🛑 Stopping services...")
        
        if self.go_process:
            self.go_process.terminate()
            try:
                self.go_process.wait(timeout=5)
            except subprocess.TimeoutExpired:
                self.go_process.kill()
                self.go_process.wait()
            print("✅ Go server stopped")
        
        if self.python_process:
            self.python_process.terminate()
            try:
                self.python_process.wait(timeout=5)
            except subprocess.TimeoutExpired:
                self.python_process.kill()
                self.python_process.wait()
            print("✅ Python executor stopped")
    
    def test_health_check(self) -> bool:
        """Test health check endpoint."""
        print("🔍 Testing health check...")
        
        try:
            response = requests.get(f"{self.go_base_url}/health", timeout=10)
            
            if response.status_code == 200:
                data = response.json()
                print(f"  Status: {data.get('status')}")
                print(f"  Service: {data.get('service')}")
                print(f"  Version: {data.get('version')}")
                
                grpc_check = data.get('checks', {}).get('grpc', {})
                print(f"  gRPC Status: {grpc_check.get('status')}")
                print(f"  gRPC Message: {grpc_check.get('message')}")
                
                if data.get('status') in ['healthy', 'degraded']:
                    print("  ✅ Health check passed")
                    return True
                else:
                    print("  ❌ Health check failed")
                    return False
            else:
                print(f"  ❌ Health check failed with status {response.status_code}")
                return False
                
        except Exception as e:
            print(f"  ❌ Health check error: {e}")
            return False
    
    def test_list_plugins(self) -> bool:
        """Test plugin listing endpoint."""
        print("🔍 Testing plugin listing...")
        
        try:
            response = requests.get(f"{self.go_base_url}/api/v1/plugins", timeout=10)
            
            if response.status_code == 200:
                data = response.json()
                plugins = data.get('plugins', [])
                total_count = data.get('total_count', 0)
                
                print(f"  Found {total_count} plugins:")
                for plugin in plugins:
                    status = "✅" if plugin.get('is_available') else "❌"
                    print(f"    {status} {plugin.get('name')}: {plugin.get('description')}")
                    print(f"      Methods: {', '.join(plugin.get('supported_methods', []))}")
                
                if total_count > 0:
                    print("  ✅ Plugin listing passed")
                    return True
                else:
                    print("  ⚠️  No plugins found")
                    return True  # Not necessarily a failure
            else:
                print(f"  ❌ Plugin listing failed with status {response.status_code}")
                return False
                
        except Exception as e:
            print(f"  ❌ Plugin listing error: {e}")
            return False
    
    def test_plugin_execution(self) -> bool:
        """Test plugin execution endpoint."""
        print("🔍 Testing plugin execution...")
        
        try:
            # First get available plugins
            response = requests.get(f"{self.go_base_url}/api/v1/plugins", timeout=10)
            if response.status_code != 200:
                print("  ❌ Failed to get plugin list")
                return False
            
            plugins = response.json().get('plugins', [])
            if not plugins:
                print("  ⚠️  No plugins available for testing")
                return True
            
            # Test with the first available plugin
            plugin = plugins[0]
            plugin_name = plugin['name']
            
            print(f"  Testing plugin: {plugin_name}")
            
            # Test different HTTP methods
            test_data = {"test_key": "test_value", "timestamp": str(time.time())}
            methods_to_test = ["GET", "POST", "PUT", "DELETE"]
            
            success_count = 0
            for method in methods_to_test:
                if method in plugin.get('supported_methods', []):
                    try:
                        if method == "GET":
                            resp = requests.get(
                                f"{self.go_base_url}/api/v1/webhook/{plugin_name}",
                                params=test_data,
                                timeout=30
                            )
                        elif method == "POST":
                            resp = requests.post(
                                f"{self.go_base_url}/api/v1/webhook/{plugin_name}",
                                json=test_data,
                                timeout=30
                            )
                        elif method == "PUT":
                            resp = requests.put(
                                f"{self.go_base_url}/api/v1/webhook/{plugin_name}",
                                json=test_data,
                                timeout=30
                            )
                        elif method == "DELETE":
                            resp = requests.delete(
                                f"{self.go_base_url}/api/v1/webhook/{plugin_name}",
                                params=test_data,
                                timeout=30
                            )
                        
                        if resp.status_code == 200:
                            data = resp.json()
                            exec_time = data.get('execution_time', 'N/A')
                            print(f"    ✅ {method}: {data.get('message')} ({exec_time})")
                            success_count += 1
                        else:
                            print(f"    ❌ {method}: Status {resp.status_code}")
                            
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
    
    def test_plugin_info(self) -> bool:
        """Test plugin info endpoint."""
        print("🔍 Testing plugin info...")
        
        try:
            # First get available plugins
            response = requests.get(f"{self.go_base_url}/api/v1/plugins", timeout=10)
            if response.status_code != 200:
                print("  ❌ Failed to get plugin list")
                return False
            
            plugins = response.json().get('plugins', [])
            if not plugins:
                print("  ⚠️  No plugins available for testing")
                return True
            
            # Test with the first available plugin
            plugin_name = plugins[0]['name']
            
            response = requests.get(f"{self.go_base_url}/api/v1/plugins/{plugin_name}", timeout=10)
            
            if response.status_code == 200:
                data = response.json()
                print(f"  Plugin: {data.get('name')}")
                print(f"  Description: {data.get('description')}")
                print(f"  Available: {data.get('is_available')}")
                print(f"  Methods: {', '.join(data.get('supported_methods', []))}")
                print("  ✅ Plugin info passed")
                return True
            else:
                print(f"  ❌ Plugin info failed with status {response.status_code}")
                return False
                
        except Exception as e:
            print(f"  ❌ Plugin info error: {e}")
            return False


def main():
    """Main test function."""
    print("🧪 Go-Python gRPC Integration Test Suite")
    print("=" * 60)

    tester = GoServerTester()

    print(f"📋 Test Configuration:")
    print(f"  Python Executor Port: {tester.python_port}")
    print(f"  Go Server Port: {tester.go_port}")
    print(f"  Go Server URL: {tester.go_base_url}")
    print()

    try:
        # Start services
        tester.start_python_executor()
        tester.start_go_server()

        # Wait a bit for services to fully initialize
        time.sleep(2)
        
        # Run tests
        tests = [
            ("Health Check", tester.test_health_check),
            ("List Plugins", tester.test_list_plugins),
            ("Plugin Execution", tester.test_plugin_execution),
            ("Plugin Info", tester.test_plugin_info),
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
        
        print(f"\n{'='*60}")
        print(f"📊 Test Results: {passed}/{total} tests passed")
        
        if passed == total:
            print("🎉 All tests passed!")
            return 0
        else:
            print("💥 Some tests failed!")
            return 1
    
    finally:
        tester.stop_services()


if __name__ == "__main__":
    try:
        exit_code = main()
        sys.exit(exit_code)
    except KeyboardInterrupt:
        print("\n⏹️  Tests interrupted by user")
        sys.exit(1)
    except Exception as e:
        print(f"\n💥 Test runner error: {e}")
        sys.exit(1)
