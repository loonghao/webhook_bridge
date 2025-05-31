#!/usr/bin/env python3
"""HTTP Server Performance and Feature Test Script."""

import asyncio
import json
import time
import sys
from typing import Dict, Any

try:
    import aiohttp
    import requests
except ImportError:
    print("⚠️  Installing required packages...")
    import subprocess
    subprocess.check_call([sys.executable, "-m", "pip", "install", "aiohttp", "requests"])
    import aiohttp
    import requests


class HTTPServerTester:
    """Test the Go HTTP server functionality and performance."""
    
    def __init__(self, base_url: str = "http://localhost:8000"):
        self.base_url = base_url
        self.session = None
        
    async def __aenter__(self):
        self.session = aiohttp.ClientSession()
        return self
        
    async def __aexit__(self, exc_type, exc_val, exc_tb):
        if self.session:
            await self.session.close()
    
    def test_sync_endpoints(self) -> Dict[str, Any]:
        """Test endpoints synchronously."""
        results = {}
        
        endpoints = [
            ("GET", "/health", "Health Check"),
            ("GET", "/metrics", "Metrics"),
            ("GET", "/", "Root Documentation"),
            ("GET", "/api/v1/plugins", "List Plugins"),
            ("GET", "/api/latest/plugins", "List Plugins (Latest)"),
            ("GET", "/nonexistent", "404 Handler"),
        ]
        
        for method, path, description in endpoints:
            url = f"{self.base_url}{path}"
            start_time = time.time()
            
            try:
                response = requests.request(method, url, timeout=10)
                duration = (time.time() - start_time) * 1000
                
                results[path] = {
                    "description": description,
                    "status_code": response.status_code,
                    "duration_ms": round(duration, 2),
                    "headers": dict(response.headers),
                    "success": 200 <= response.status_code < 500,
                }
                
                # Try to parse JSON response
                try:
                    results[path]["response"] = response.json()
                except:
                    results[path]["response"] = response.text[:200]
                    
            except Exception as e:
                results[path] = {
                    "description": description,
                    "error": str(e),
                    "success": False,
                }
        
        return results
    
    async def test_async_endpoints(self) -> Dict[str, Any]:
        """Test endpoints asynchronously for performance."""
        results = {}
        
        endpoints = [
            ("GET", "/health"),
            ("GET", "/metrics"),
            ("GET", "/api/v1/plugins"),
        ]
        
        # Test concurrent requests
        tasks = []
        for method, path in endpoints:
            for i in range(5):  # 5 concurrent requests per endpoint
                task = self.make_request(method, path, f"{path}_{i}")
                tasks.append(task)
        
        start_time = time.time()
        responses = await asyncio.gather(*tasks, return_exceptions=True)
        total_duration = (time.time() - start_time) * 1000
        
        # Process results
        successful = sum(1 for r in responses if isinstance(r, dict) and r.get("success", False))
        failed = len(responses) - successful
        
        results["concurrent_test"] = {
            "total_requests": len(responses),
            "successful": successful,
            "failed": failed,
            "total_duration_ms": round(total_duration, 2),
            "avg_duration_ms": round(total_duration / len(responses), 2),
            "requests_per_second": round(len(responses) / (total_duration / 1000), 2),
        }
        
        return results
    
    async def make_request(self, method: str, path: str, identifier: str) -> Dict[str, Any]:
        """Make an async HTTP request."""
        url = f"{self.base_url}{path}"
        start_time = time.time()
        
        try:
            async with self.session.request(method, url) as response:
                duration = (time.time() - start_time) * 1000
                
                return {
                    "identifier": identifier,
                    "status_code": response.status,
                    "duration_ms": round(duration, 2),
                    "success": 200 <= response.status < 500,
                }
        except Exception as e:
            return {
                "identifier": identifier,
                "error": str(e),
                "success": False,
            }
    
    def test_middleware_features(self) -> Dict[str, Any]:
        """Test middleware functionality."""
        results = {}
        
        # Test CORS
        response = requests.options(f"{self.base_url}/api/v1/plugins")
        results["cors"] = {
            "status_code": response.status_code,
            "access_control_allow_origin": response.headers.get("Access-Control-Allow-Origin"),
            "access_control_allow_methods": response.headers.get("Access-Control-Allow-Methods"),
        }
        
        # Test Request ID
        response = requests.get(f"{self.base_url}/health")
        results["request_id"] = {
            "header_present": "X-Request-ID" in response.headers,
            "request_id": response.headers.get("X-Request-ID"),
        }
        
        # Test Execution Time
        results["execution_time"] = {
            "header_present": "X-Execution-Time" in response.headers,
            "execution_time": response.headers.get("X-Execution-Time"),
        }
        
        # Test Security Headers
        security_headers = [
            "X-Content-Type-Options",
            "X-Frame-Options", 
            "X-XSS-Protection",
            "Referrer-Policy",
        ]
        
        results["security_headers"] = {}
        for header in security_headers:
            results["security_headers"][header] = response.headers.get(header)
        
        return results


async def main():
    """Main test function."""
    print("🚀 Starting HTTP Server Tests...")
    print("=" * 60)
    
    async with HTTPServerTester() as tester:
        # Test basic endpoints
        print("📋 Testing Basic Endpoints...")
        sync_results = tester.test_sync_endpoints()
        
        for path, result in sync_results.items():
            status = "✅" if result.get("success", False) else "❌"
            duration = result.get("duration_ms", "N/A")
            status_code = result.get("status_code", "N/A")
            print(f"  {status} {result['description']}: {status_code} ({duration}ms)")
        
        print()
        
        # Test middleware features
        print("🔧 Testing Middleware Features...")
        middleware_results = tester.test_middleware_features()
        
        print(f"  ✅ CORS: {middleware_results['cors']['access_control_allow_origin']}")
        print(f"  ✅ Request ID: {middleware_results['request_id']['header_present']}")
        print(f"  ✅ Execution Time: {middleware_results['execution_time']['header_present']}")
        print(f"  ✅ Security Headers: {len([h for h in middleware_results['security_headers'].values() if h])}/4")
        
        print()
        
        # Test performance
        print("⚡ Testing Performance (Concurrent Requests)...")
        async_results = await tester.test_async_endpoints()
        
        perf = async_results["concurrent_test"]
        print(f"  📊 Total Requests: {perf['total_requests']}")
        print(f"  ✅ Successful: {perf['successful']}")
        print(f"  ❌ Failed: {perf['failed']}")
        print(f"  ⏱️  Total Duration: {perf['total_duration_ms']}ms")
        print(f"  📈 Requests/Second: {perf['requests_per_second']}")
        print(f"  📊 Avg Response Time: {perf['avg_duration_ms']}ms")
        
        print()
        
        # Summary
        print("📊 Test Summary:")
        total_endpoints = len(sync_results)
        successful_endpoints = sum(1 for r in sync_results.values() if r.get("success", False))
        
        print(f"  Endpoint Tests: {successful_endpoints}/{total_endpoints} passed")
        print(f"  Performance Test: {perf['successful']}/{perf['total_requests']} requests successful")
        print(f"  Average Response Time: {perf['avg_duration_ms']}ms")
        
        if successful_endpoints == total_endpoints and perf['failed'] == 0:
            print("  🎉 All tests passed!")
            return 0
        else:
            print("  💥 Some tests failed!")
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
