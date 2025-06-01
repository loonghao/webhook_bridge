# Python Plugin Development Guide

本指南详细介绍如何为 Webhook Bridge 开发 Python 插件。

## 🚀 快速开始

### 1. 安装 Python API 包

首先安装 `webhook-bridge` Python 包以获取插件开发 API：

```bash
# 使用 pip 安装
pip install webhook-bridge

# 或使用 uv (推荐)
uv pip install webhook-bridge

# 验证安装
python -c "from webhook_bridge.plugin import BasePlugin; print('安装成功!')"
```

### 2. 创建第一个插件

在插件目录中创建 Python 文件：

```python
# hello_plugin.py
from typing import Dict, Any
from webhook_bridge.plugin import BasePlugin


class Plugin(BasePlugin):
    """Hello World 插件示例
    
    注意：类名必须是 'Plugin' 才能被自动发现
    """

    def handle(self) -> Dict[str, Any]:
        """通用处理器，处理所有 HTTP 方法"""
        name = self.data.get("name", "World")
        
        self.logger.info(f"Hello plugin 处理 {self.http_method} 请求")
        
        return {
            "status": "success",
            "data": {
                "message": f"Hello, {name}!",
                "method": self.http_method,
                "plugin": "hello_plugin"
            }
        }
```

### 3. 测试插件

```bash
# 启动服务器
./webhook-bridge-server

# 测试插件
curl -X POST "http://localhost:8000/api/v1/webhook/hello_plugin" \
     -H "Content-Type: application/json" \
     -d '{"name": "Developer"}'
```

## 📚 BasePlugin API 详解

### 可用属性

```python
class Plugin(BasePlugin):
    def handle(self):
        # 访问 webhook 数据
        data = self.data  # Dict[str, Any]
        
        # 获取 HTTP 方法
        method = self.http_method  # str: GET/POST/PUT/DELETE
        
        # 使用日志记录
        self.logger.info("插件执行中...")
        self.logger.error("发生错误")
        
        # 访问执行结果 (可选)
        result = self.result  # Dict[str, Any]
```

### 方法重写

```python
class Plugin(BasePlugin):
    def handle(self) -> Dict[str, Any]:
        """通用处理器 - 必须实现"""
        pass
    
    def get(self) -> Dict[str, Any]:
        """处理 GET 请求"""
        pass
    
    def post(self) -> Dict[str, Any]:
        """处理 POST 请求"""
        pass
    
    def put(self) -> Dict[str, Any]:
        """处理 PUT 请求"""
        pass
    
    def delete(self) -> Dict[str, Any]:
        """处理 DELETE 请求"""
        pass
    
    def run(self) -> Dict[str, Any]:
        """向后兼容方法 (v0.6.0)"""
        pass
```

## 🔄 插件执行流程

### 混合架构执行流程

```
1. HTTP 请求 → Go HTTP 服务器 (端口 8000)
   ├─ 请求验证和路由
   └─ 提取插件名称和数据

2. gRPC 调用 → Python 执行器 (端口 50051)
   ├─ 加载插件类
   ├─ 创建插件实例
   └─ 方法路由

3. 插件执行
   ├─ 根据 HTTP 方法调用相应处理器
   ├─ 访问 self.data, self.http_method, self.logger
   └─ 返回结果字典

4. 响应处理
   ├─ gRPC 响应 → Go 服务器
   ├─ 格式化 HTTP 响应
   └─ 返回给客户端
```

### 方法路由逻辑

```python
# 插件方法调用优先级
if hasattr(plugin, method.lower()):  # get/post/put/delete
    result = getattr(plugin, method.lower())()
else:
    result = plugin.handle()  # 回退到通用处理器
```

## 🧪 Dashboard 可视化测试

### 访问测试界面

1. 启动服务器：`./webhook-bridge-server`
2. 打开浏览器：`http://localhost:8000/`
3. 导航到 **Plugins** 标签页
4. 选择要测试的插件

### 测试界面功能

- **🎯 插件选择器**: 下拉菜单选择插件
- **🔧 HTTP 方法**: GET/POST/PUT/DELETE 切换
- **📝 数据编辑器**: JSON 格式测试数据
- **⚡ 执行按钮**: 一键执行插件
- **📊 结果显示**: 实时结果和性能指标
- **🐛 错误调试**: 详细错误信息

### 测试示例

**输入数据：**
```json
{
  "message": "Hello from Dashboard!",
  "user_id": 12345,
  "timestamp": "2024-01-01T00:00:00Z"
}
```

**执行结果：**
```json
{
  "status_code": 200,
  "message": "success", 
  "execution_time": "0.045s",
  "data": {
    "status": "success",
    "data": {
      "processed_message": "Processed: Hello from Dashboard!",
      "method": "POST"
    }
  }
}
```

## 📝 插件开发最佳实践

### 1. 错误处理

```python
class Plugin(BasePlugin):
    def handle(self) -> Dict[str, Any]:
        try:
            # 插件逻辑
            result = self.process_data()
            return {
                "status": "success",
                "data": result
            }
        except ValueError as e:
            self.logger.error(f"数据验证错误: {e}")
            return {
                "status": "error",
                "error": f"Invalid data: {e}"
            }
        except Exception as e:
            self.logger.error(f"插件执行失败: {e}")
            return {
                "status": "error", 
                "error": "Internal plugin error"
            }
```

### 2. 数据验证

```python
class Plugin(BasePlugin):
    def handle(self) -> Dict[str, Any]:
        # 验证必需字段
        required_fields = ["user_id", "action"]
        for field in required_fields:
            if field not in self.data:
                return {
                    "status": "error",
                    "error": f"Missing required field: {field}"
                }
        
        # 验证数据类型
        if not isinstance(self.data.get("user_id"), int):
            return {
                "status": "error",
                "error": "user_id must be an integer"
            }
        
        # 处理数据
        return self.process_validated_data()
```

### 3. 日志记录

```python
class Plugin(BasePlugin):
    def handle(self) -> Dict[str, Any]:
        self.logger.info(f"开始处理 {self.http_method} 请求")
        self.logger.debug(f"接收数据: {self.data}")
        
        try:
            result = self.process_data()
            self.logger.info("插件执行成功")
            return result
        except Exception as e:
            self.logger.error(f"插件执行失败: {e}", exc_info=True)
            raise
```

## 🔧 高级功能

### 1. 配置管理

```python
import os
from webhook_bridge.plugin import BasePlugin

class Plugin(BasePlugin):
    def __init__(self, *args, **kwargs):
        super().__init__(*args, **kwargs)
        # 从环境变量读取配置
        self.api_key = os.getenv("MY_PLUGIN_API_KEY")
        self.timeout = int(os.getenv("MY_PLUGIN_TIMEOUT", "30"))
```

### 2. 外部 API 调用

```python
import requests
from webhook_bridge.plugin import BasePlugin

class Plugin(BasePlugin):
    def handle(self) -> Dict[str, Any]:
        try:
            response = requests.post(
                "https://api.example.com/webhook",
                json=self.data,
                timeout=30
            )
            response.raise_for_status()
            
            return {
                "status": "success",
                "data": {
                    "external_response": response.json(),
                    "status_code": response.status_code
                }
            }
        except requests.RequestException as e:
            self.logger.error(f"外部 API 调用失败: {e}")
            return {
                "status": "error",
                "error": f"External API error: {e}"
            }
```

## 📁 插件目录结构

```
plugins/
├── hello_plugin.py          # 简单插件
├── notification/             # 复杂插件包
│   ├── __init__.py
│   ├── plugin.py            # 主插件类
│   ├── config.py            # 配置管理
│   └── utils.py             # 工具函数
└── requirements.txt         # 插件依赖
```

## 🚀 部署和分发

### 1. 插件依赖管理

```bash
# 在插件目录创建 requirements.txt
echo "requests>=2.25.0" > plugins/requirements.txt
echo "pydantic>=1.8.0" >> plugins/requirements.txt

# 安装插件依赖
pip install -r plugins/requirements.txt
```

### 2. Docker 部署

```dockerfile
# 在 Dockerfile 中安装插件依赖
COPY plugins/requirements.txt /app/plugins/
RUN pip install -r /app/plugins/requirements.txt

# 复制插件文件
COPY plugins/ /app/plugins/
```

## 🔍 调试和故障排除

### 1. 启用调试日志

```bash
# 启动服务器时启用调试模式
./webhook-bridge-server --log-level debug

# 或设置环境变量
export WEBHOOK_BRIDGE_LOG_LEVEL=debug
./webhook-bridge-server
```

### 2. 插件测试脚本

```python
# test_plugin.py
from webhook_bridge.plugin import BasePlugin
import sys
import os

# 添加插件路径
sys.path.insert(0, os.path.dirname(__file__))

# 导入插件
from my_plugin import Plugin

# 测试插件
test_data = {"message": "test"}
plugin = Plugin(test_data, http_method="POST")
result = plugin.handle()
print(f"测试结果: {result}")
```

## 📚 更多资源

- [API 文档](API.md) - 完整的 API 参考
- [配置指南](CONFIGURATION.md) - 服务器配置选项
- [Docker 指南](DOCKER_GUIDE.md) - 容器化部署
- [示例插件](../example_plugins/) - 更多插件示例
