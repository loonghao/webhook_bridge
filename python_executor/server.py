"""gRPC Server Implementation for Python Webhook Executor.

This module implements the gRPC service interface for executing Python webhook plugins.
It maintains full compatibility with the existing webhook_bridge plugin system.
"""
# Import future modules
from __future__ import annotations

# Import built-in modules
import logging
import os
import time
import traceback
from typing import Any
from typing import Dict
from typing import Optional

# Import third-party modules
from api.proto import webhook_pb2
from api.proto import webhook_pb2_grpc
import grpc

# Import local modules
from webhook_bridge.filesystem import get_plugins
from webhook_bridge.plugin import BasePlugin
from webhook_bridge.plugin import load_plugin


class WebhookExecutorServicer(webhook_pb2_grpc.WebhookExecutorServicer):
    """gRPC service implementation for webhook plugin execution."""

    def __init__(self, plugin_dirs: Optional[list[str]] = None):
        """Initialize the webhook executor service.

        Args:
            plugin_dirs: Optional list of plugin directories to search
        """
        self.logger = logging.getLogger(__name__)
        self.plugin_dirs = plugin_dirs or []
        self.plugin_cache = {}  # Cache for loaded plugin classes
        self.execution_stats = {
            "total_executions": 0,
            "successful_executions": 0,
            "failed_executions": 0,
            "total_execution_time": 0.0,
        }

        # Setup logging format
        logging.basicConfig(
            level=logging.INFO,
            format='%(asctime)s - %(name)s - %(levelname)s - %(message)s',
        )

        self.logger.info("WebhookExecutorServicer initialized")
        self.logger.info(f"Plugin directories: {self.plugin_dirs}")

        # Validate environment on startup
        self._validate_environment()

    def _validate_environment(self) -> None:
        """Validate the Python environment for plugin execution."""
        try:
            # Check if we can import required modules
            # Import local modules
            from webhook_bridge.filesystem import get_plugins
            from webhook_bridge.plugin import BasePlugin
            from webhook_bridge.plugin import load_plugin

            # Test plugin discovery
            plugins = get_plugins()
            self.logger.info(f"Environment validation passed. Found {len(plugins)} plugins.")

        except ImportError as e:
            self.logger.error(f"Environment validation failed: {e}")
            raise
        except Exception as e:
            self.logger.warning(f"Environment validation warning: {e}")

    def _get_plugins_with_extra_dirs(self) -> Dict[str, str]:
        """Get plugins including extra directories."""
        # Start with default plugins
        plugins = get_plugins()

        # Add plugins from extra directories
        for plugin_dir in self.plugin_dirs:
            try:
                # Use absolute path
                # Import built-in modules
                import os
                abs_plugin_dir = os.path.abspath(plugin_dir)
                self.logger.debug(f"Searching for plugins in: {abs_plugin_dir}")

                # Manually discover plugins in the directory
                extra_plugins = self._discover_plugins_in_dir(abs_plugin_dir)
                plugins.update(extra_plugins)
                self.logger.info(f"Added {len(extra_plugins)} plugins from {plugin_dir}")
            except Exception as e:
                self.logger.warning(f"Failed to load plugins from {plugin_dir}: {e}")

        return plugins

    def _discover_plugins_in_dir(self, plugin_dir: str) -> Dict[str, str]:
        """Manually discover plugins in a directory."""
        plugins = {}

        if not os.path.exists(plugin_dir):
            self.logger.warning(f"Plugin directory does not exist: {plugin_dir}")
            return plugins

        if not os.path.isdir(plugin_dir):
            self.logger.warning(f"Plugin path is not a directory: {plugin_dir}")
            return plugins

        # Look for Python files
        for filename in os.listdir(plugin_dir):
            if filename.endswith('.py') and not filename.startswith('__'):
                plugin_path = os.path.join(plugin_dir, filename)
                plugin_name = filename[:-3]  # Remove .py extension
                plugins[plugin_name] = plugin_path
                self.logger.debug(f"Found plugin: {plugin_name} at {plugin_path}")

        return plugins

    def _load_plugin_with_cache(self, plugin_path: str) -> type[BasePlugin]:
        """Load plugin with caching to improve performance."""
        # Use file modification time as cache key
        try:
            mtime = os.path.getmtime(plugin_path)
            cache_key = f"{plugin_path}:{mtime}"

            if cache_key in self.plugin_cache:
                self.logger.debug(f"Using cached plugin: {plugin_path}")
                return self.plugin_cache[cache_key]

            # Load plugin
            plugin_class = load_plugin(plugin_path)

            # Cache the result
            self.plugin_cache[cache_key] = plugin_class
            self.logger.debug(f"Cached plugin: {plugin_path}")

            return plugin_class

        except Exception as e:
            self.logger.error(f"Failed to load plugin {plugin_path}: {e}")
            raise

    def _execute_plugin_safely(self, plugin_class: type[BasePlugin], data: Dict[str, str], http_method: str) -> Dict[str, Any]:
        """Execute plugin with proper error handling and method routing."""
        try:
            # Create plugin instance
            plugin_instance = plugin_class(data, http_method=http_method)

            # Use the new execute method which calls run() and formats results
            result = plugin_instance.execute()

            self.logger.debug(f"Plugin execution result: {type(result)}")
            return result

        except Exception as e:
            self.logger.error(f"Plugin execution failed: {e}")
            self.logger.debug(f"Plugin execution traceback: {traceback.format_exc()}")
            raise

    def _format_plugin_result(self, result: Any) -> Dict[str, str]:
        """Format plugin result for gRPC response."""
        result_data = {}

        if isinstance(result, dict):
            for key, value in result.items():
                if isinstance(value, (dict, list)):
                    # Convert complex types to JSON strings
                    # Import built-in modules
                    import json
                    result_data[key] = json.dumps(value)
                else:
                    result_data[key] = str(value)
        else:
            result_data["result"] = str(result)

        return result_data

    def _update_execution_stats(self, success: bool, execution_time: float) -> None:
        """Update execution statistics."""
        self.execution_stats["total_executions"] += 1
        self.execution_stats["total_execution_time"] += execution_time

        if success:
            self.execution_stats["successful_executions"] += 1
        else:
            self.execution_stats["failed_executions"] += 1

    def ExecutePlugin(
        self,
        request: webhook_pb2.ExecutePluginRequest,
        context: grpc.ServicerContext,
    ) -> webhook_pb2.ExecutePluginResponse:
        """Execute a webhook plugin.

        Args:
            request: The plugin execution request
            context: gRPC service context

        Returns:
            ExecutePluginResponse: The plugin execution result
        """
        start_time = time.time()
        plugin_name = request.plugin_name
        http_method = request.http_method
        success = False

        self.logger.info(f"Executing plugin: {plugin_name} with method: {http_method}")
        self.logger.debug(f"Request data: {dict(request.data)}")

        try:
            # Get available plugins (including extra directories)
            plugins = self._get_plugins_with_extra_dirs()

            if plugin_name not in plugins:
                error_msg = f"Plugin '{plugin_name}' not found. Available plugins: {list(plugins.keys())}"
                self.logger.error(error_msg)
                execution_time = time.time() - start_time
                self._update_execution_stats(False, execution_time)

                return webhook_pb2.ExecutePluginResponse(
                    status_code=404,
                    message="Plugin not found",
                    error=error_msg,
                    execution_time=execution_time,
                )

            # Convert gRPC request data to Python dict
            data = dict(request.data)

            # Load plugin with caching
            plugin_src_file = plugins[plugin_name]
            plugin_class = self._load_plugin_with_cache(plugin_src_file)

            # Execute plugin safely
            result = self._execute_plugin_safely(plugin_class, data, http_method)

            # Format result for gRPC
            result_data = self._format_plugin_result(result)

            execution_time = time.time() - start_time
            success = True
            self._update_execution_stats(True, execution_time)

            self.logger.info(f"Successfully executed plugin {plugin_name} in {execution_time:.3f}s")

            return webhook_pb2.ExecutePluginResponse(
                status_code=200,
                message="success",
                data=result_data,
                execution_time=execution_time,
            )

        except Exception as e:
            execution_time = time.time() - start_time
            error_msg = str(e)
            error_trace = traceback.format_exc()

            self._update_execution_stats(False, execution_time)

            self.logger.error(f"Error executing plugin {plugin_name}: {error_msg}")
            self.logger.debug(f"Error trace: {error_trace}")

            # Determine appropriate status code based on error type
            status_code = 500
            if "not found" in error_msg.lower():
                status_code = 404
            elif "permission" in error_msg.lower() or "access" in error_msg.lower():
                status_code = 403
            elif "timeout" in error_msg.lower():
                status_code = 408

            return webhook_pb2.ExecutePluginResponse(
                status_code=status_code,
                message="Plugin execution failed",
                error=error_msg,
                execution_time=execution_time,
            )

    def ListPlugins(
        self,
        request: webhook_pb2.ListPluginsRequest,
        context: grpc.ServicerContext,
    ) -> webhook_pb2.ListPluginsResponse:
        """List available webhook plugins.

        Args:
            request: The list plugins request
            context: gRPC service context

        Returns:
            ListPluginsResponse: List of available plugins
        """
        self.logger.info("Listing available plugins")
        self.logger.debug(f"Filter: {request.filter}")

        try:
            # Get plugins from all directories
            plugins = self._get_plugins_with_extra_dirs()
            plugin_infos = []

            for name, path in plugins.items():
                # Apply filter if provided
                if request.filter and request.filter.lower() not in name.lower():
                    continue

                try:
                    # Try to load plugin to get more info
                    plugin_class = self._load_plugin_with_cache(path)

                    # Get supported methods by checking which methods are implemented
                    supported_methods = self._get_supported_methods(plugin_class)

                    # Get description from docstring if available
                    description = self._get_plugin_description(plugin_class)

                    # Get file modification time
                    try:
                        # Import built-in modules
                        import datetime
                        mtime = os.path.getmtime(path)
                        last_modified = datetime.datetime.fromtimestamp(mtime).isoformat()
                    except:
                        last_modified = ""

                    plugin_info = webhook_pb2.PluginInfo(
                        name=name,
                        path=path,
                        description=description,
                        supported_methods=supported_methods,
                        is_available=True,
                        last_modified=last_modified,
                    )

                except Exception as e:
                    self.logger.warning(f"Failed to load plugin {name}: {e}")
                    plugin_info = webhook_pb2.PluginInfo(
                        name=name,
                        path=path,
                        description=f"Failed to load: {e}",
                        supported_methods=[],
                        is_available=False,
                        last_modified="",
                    )

                plugin_infos.append(plugin_info)

            # Sort plugins by name for consistent ordering
            plugin_infos.sort(key=lambda p: p.name)

            self.logger.info(f"Found {len(plugin_infos)} plugins (filtered from {len(plugins)} total)")

            return webhook_pb2.ListPluginsResponse(
                plugins=plugin_infos,
                total_count=len(plugin_infos),
            )

        except Exception as e:
            self.logger.error(f"Error listing plugins: {e}")
            context.set_code(grpc.StatusCode.INTERNAL)
            context.set_details(f"Failed to list plugins: {e}")
            return webhook_pb2.ListPluginsResponse()

    def _get_supported_methods(self, plugin_class: type[BasePlugin]) -> list[str]:
        """Get supported HTTP methods for a plugin."""
        supported_methods = []

        # Check if plugin has method-specific handlers
        for method in ["GET", "POST", "PUT", "DELETE"]:
            method_name = method.lower()
            if hasattr(plugin_class, method_name):
                # Check if the method is overridden (not just inherited from BasePlugin)
                plugin_method = getattr(plugin_class, method_name)
                base_method = getattr(BasePlugin, method_name)
                if plugin_method != base_method:
                    supported_methods.append(method)

        # If no specific methods are implemented, assume all are supported
        if not supported_methods:
            supported_methods = ["GET", "POST", "PUT", "DELETE"]

        return supported_methods

    def _get_plugin_description(self, plugin_class: type[BasePlugin]) -> str:
        """Get plugin description from docstring."""
        description = ""

        if plugin_class.__doc__:
            # Get first line of docstring
            lines = plugin_class.__doc__.strip().split('\n')
            description = lines[0].strip()

            # Remove common prefixes
            if description.startswith('"""'):
                description = description[3:].strip()
            if description.endswith('"""'):
                description = description[:-3].strip()

        # Fallback to class name if no description
        if not description:
            description = f"Plugin class: {plugin_class.__name__}"

        return description

    def GetPluginInfo(
        self,
        request: webhook_pb2.GetPluginInfoRequest,
        context: grpc.ServicerContext,
    ) -> webhook_pb2.GetPluginInfoResponse:
        """Get information about a specific plugin.
        
        Args:
            request: The get plugin info request
            context: gRPC service context
            
        Returns:
            GetPluginInfoResponse: Plugin information
        """
        plugin_name = request.plugin_name
        self.logger.info(f"Getting info for plugin: {plugin_name}")

        try:
            plugins = get_plugins()

            if plugin_name not in plugins:
                return webhook_pb2.GetPluginInfoResponse(found=False)

            path = plugins[plugin_name]

            try:
                plugin_class = load_plugin(path)

                description = ""
                if plugin_class.__doc__:
                    description = plugin_class.__doc__.strip()

                plugin_info = webhook_pb2.PluginInfo(
                    name=plugin_name,
                    path=path,
                    description=description,
                    supported_methods=["GET", "POST", "PUT", "DELETE"],
                    is_available=True,
                )

                return webhook_pb2.GetPluginInfoResponse(
                    plugin=plugin_info,
                    found=True,
                )

            except Exception as e:
                self.logger.warning(f"Failed to load plugin {plugin_name}: {e}")
                plugin_info = webhook_pb2.PluginInfo(
                    name=plugin_name,
                    path=path,
                    description=f"Failed to load: {e}",
                    supported_methods=[],
                    is_available=False,
                )

                return webhook_pb2.GetPluginInfoResponse(
                    plugin=plugin_info,
                    found=True,
                )

        except Exception as e:
            self.logger.error(f"Error getting plugin info: {e}")
            context.set_code(grpc.StatusCode.INTERNAL)
            context.set_details(f"Failed to get plugin info: {e}")
            return webhook_pb2.GetPluginInfoResponse(found=False)

    def HealthCheck(
        self,
        request: webhook_pb2.HealthCheckRequest,
        context: grpc.ServicerContext,
    ) -> webhook_pb2.HealthCheckResponse:
        """Perform health check.

        Args:
            request: The health check request
            context: gRPC service context

        Returns:
            HealthCheckResponse: Health status
        """
        self.logger.debug("Health check requested")

        try:
            # Test plugin system
            plugins = self._get_plugins_with_extra_dirs()
            plugin_count = len(plugins)

            # Get execution statistics
            stats = self.execution_stats
            success_rate = 0.0
            if stats["total_executions"] > 0:
                success_rate = (stats["successful_executions"] / stats["total_executions"]) * 100

            avg_execution_time = 0.0
            if stats["total_executions"] > 0:
                avg_execution_time = stats["total_execution_time"] / stats["total_executions"]

            # Check if we can load at least one plugin
            plugin_test_status = "unknown"
            if plugins:
                try:
                    # Try to load the first plugin as a test
                    first_plugin = next(iter(plugins.values()))
                    self._load_plugin_with_cache(first_plugin)
                    plugin_test_status = "ok"
                except Exception as e:
                    plugin_test_status = f"failed: {e}"

            details = {
                "plugin_count": str(plugin_count),
                "service": "python-executor",
                "total_executions": str(stats["total_executions"]),
                "successful_executions": str(stats["successful_executions"]),
                "failed_executions": str(stats["failed_executions"]),
                "success_rate": f"{success_rate:.2f}%",
                "avg_execution_time": f"{avg_execution_time:.3f}s",
                "plugin_test": plugin_test_status,
                "cache_size": str(len(self.plugin_cache)),
            }

            # Determine overall health status
            status = "healthy"
            message = f"Python executor is healthy. {plugin_count} plugins available."

            if plugin_count == 0:
                status = "degraded"
                message = "No plugins found"
            elif plugin_test_status.startswith("failed"):
                status = "degraded"
                message = f"Plugin loading issues detected: {plugin_test_status}"
            elif stats["total_executions"] > 10 and success_rate < 80:
                status = "degraded"
                message = f"Low success rate: {success_rate:.1f}%"

            return webhook_pb2.HealthCheckResponse(
                status=status,
                message=message,
                details=details,
            )

        except Exception as e:
            self.logger.error(f"Health check failed: {e}")
            return webhook_pb2.HealthCheckResponse(
                status="unhealthy",
                message=f"Health check failed: {e}",
                details={"error": str(e)},
            )
