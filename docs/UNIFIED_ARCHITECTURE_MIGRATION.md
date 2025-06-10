# Webhook Bridge 统一架构迁移指南

## 概述

Webhook Bridge v2.0+ 已从多个可执行文件的架构迁移到统一的单一可执行文件架构。本文档详细说明了这一重大变化以及如何从旧架构迁移到新架构。

## 架构对比

### 旧架构 (v1.x - 多可执行文件)

```
❌ webhook-bridge-server.exe    (Go HTTP服务器)
❌ python-manager.exe           (Python环境管理器)
❌ python python_executor/main.py (Python执行器服务)

问题：
- 需要管理多个进程
- 进程间协调复杂
- 部署和维护困难
- 容易出现进程同步问题
```

### 新架构 (v2.0+ - 统一可执行文件)

```
✅ webhook-bridge.exe           (统一CLI + 所有功能)
   ├── unified                  (统一服务 - 推荐!)
   ├── serve                    (独立Go服务器)
   ├── server                   (后端服务器 + gRPC客户端)
   ├── python                   (Python环境管理)
   └── 其他管理命令...

优势：
- 单一可执行文件
- 统一进程管理
- 自动服务协调
- 简化部署和维护
```

## 迁移步骤

### 1. 停止旧服务

```bash
# 停止所有旧版本服务
taskkill /f /im webhook-bridge-server.exe
taskkill /f /im python-manager.exe
taskkill /f /im python.exe  # 如果运行了Python执行器
```

### 2. 备份配置和数据

```bash
# 备份重要文件
cp config.yaml config.yaml.backup
cp -r plugins plugins.backup
cp -r logs logs.backup
cp -r data data.backup
```

### 3. 下载新版本

```bash
# 下载最新的统一版本
wget https://github.com/loonghao/webhook_bridge/releases/latest/download/webhook_bridge_Windows_x86_64.zip
unzip webhook_bridge_Windows_x86_64.zip

# 或者从源码构建
go build -o webhook-bridge.exe ./cmd/webhook-bridge
```

### 4. 更新配置文件

新版本的配置文件增加了统一服务相关配置：

```yaml
# 新增：统一服务配置
unified:
  enabled: true
  shutdown_timeout: 10
  startup:
    python_executor_delay: 3
    health_check_interval: 5

# 更新：执行器配置
executor:
  host: "127.0.0.1"
  port: 50051
  timeout: 30
  auto_start: true  # 新增：自动启动
```

### 5. 测试新架构

```bash
# 测试统一服务
webhook-bridge.exe unified --port 8080 --verbose

# 测试Python环境
webhook-bridge.exe python info

# 测试独立服务器
webhook-bridge.exe serve --port 8081
```

## 命令迁移对照表

| 旧命令 | 新命令 | 说明 |
|--------|--------|------|
| `webhook-bridge-server.exe` | `webhook-bridge.exe unified` | 推荐使用统一服务 |
| `webhook-bridge-server.exe --port 8080` | `webhook-bridge.exe unified --port 8080` | 统一服务指定端口 |
| `python-manager.exe --info` | `webhook-bridge.exe python info` | Python环境信息 |
| `python-manager.exe --validate` | `webhook-bridge.exe python validate` | Python环境验证 |
| 手动启动Python执行器 | 自动管理 | 统一服务自动启动Python执行器 |

## 启动方式对比

### 旧方式 (多步骤)

```bash
# 步骤1: 启动Python执行器
python python_executor/main.py &

# 步骤2: 启动Go服务器
webhook-bridge-server.exe --port 8080

# 问题：需要手动管理两个进程
```

### 新方式 (一步完成)

```bash
# 方式1: 统一服务 (推荐)
webhook-bridge.exe unified --port 8080

# 方式2: 独立服务器 (仅Go)
webhook-bridge.exe serve --port 8080

# 方式3: 后端服务器 (Go + gRPC客户端)
webhook-bridge.exe server --port 8080
```

## 功能对比

| 功能 | 旧架构 | 新架构 |
|------|--------|--------|
| HTTP服务器 | ✅ webhook-bridge-server.exe | ✅ 所有模式 |
| Python插件执行 | ✅ 手动启动 | ✅ 自动管理 |
| Python环境管理 | ✅ python-manager.exe | ✅ `python` 子命令 |
| 进程管理 | ❌ 手动 | ✅ 自动 |
| 服务协调 | ❌ 复杂 | ✅ 简单 |
| 部署复杂度 | ❌ 高 | ✅ 低 |
| 错误恢复 | ❌ 手动 | ✅ 自动 |

## 配置文件变化

### 新增配置项

```yaml
# 统一服务配置
unified:
  enabled: true
  shutdown_timeout: 10
  startup:
    python_executor_delay: 3
    health_check_interval: 5

# 执行器自动启动
executor:
  auto_start: true
```

### 保持兼容的配置项

```yaml
# 这些配置项保持不变
server:
  host: "0.0.0.0"
  port: 8080
  mode: "debug"

python:
  strategy: "auto"
  plugin_dirs:
    - "./plugins"
    - "./example_plugins"

logging:
  level: "info"
  format: "text"
```

## 最佳实践

### 1. 生产环境推荐

```bash
# 使用统一服务，生产模式
webhook-bridge.exe unified --mode release --port 8080
```

### 2. 开发环境推荐

```bash
# 使用统一服务，详细输出
webhook-bridge.exe unified --verbose
```

### 3. 轻量级部署

```bash
# 仅Go服务器，无Python功能
webhook-bridge.exe serve --port 8080
```

### 4. API测试

```bash
# 健康检查
curl http://localhost:8080/health

# 插件列表
curl http://localhost:8080/api/v1/plugins

# 插件执行
curl "http://localhost:8080/api/v1/webhook/test_plugin?test=data"
```

## 故障排除

### 1. 迁移后服务无法启动

```bash
# 检查配置文件
webhook-bridge.exe config --show

# 检查Python环境
webhook-bridge.exe python info

# 重新构建
webhook-bridge.exe build --force
```

### 2. Python插件不工作

```bash
# 验证Python环境
webhook-bridge.exe python validate

# 检查插件目录
webhook-bridge.exe python info

# 使用统一服务而不是serve
webhook-bridge.exe unified --verbose
```

### 3. 端口冲突

```bash
# 检查端口占用
netstat -ano | findstr :8080

# 使用不同端口
webhook-bridge.exe unified --port 8081
```

## 回滚方案

如果需要回滚到旧架构：

1. 停止新服务
2. 恢复备份的配置文件
3. 使用旧版本的可执行文件
4. 手动启动各个服务

```bash
# 停止新服务
taskkill /f /im webhook-bridge.exe

# 恢复配置
cp config.yaml.backup config.yaml

# 启动旧服务 (如果还有旧文件)
python python_executor/main.py &
webhook-bridge-server.exe
```

## 总结

统一架构带来的主要改进：

1. **简化部署**: 从多个exe文件到单个exe文件
2. **自动管理**: Python执行器自动启动和管理
3. **统一接口**: 所有功能通过一个CLI访问
4. **更好的错误处理**: 统一的错误处理和恢复机制
5. **简化运维**: 单一进程监控和管理

推荐所有用户迁移到新的统一架构，享受更简单、更可靠的webhook bridge体验。
