# Webhook Bridge - 统一CLI工具

## 概述

`webhook-bridge.exe` 是一个统一的命令行工具，提供了构建、部署和管理webhook bridge服务的所有功能。它自动处理Go编译、Python环境设置和服务管理。

## 特性

✅ **自动环境配置** - 自动检测和配置Go、Python环境
✅ **一键构建** - 构建Go二进制文件和设置Python虚拟环境
✅ **服务管理** - 启动、停止、重启和状态检查
✅ **跨平台支持** - Windows、Linux、macOS
✅ **智能依赖管理** - 优先使用UV，自动回退到pip
✅ **配置管理** - 多环境配置支持

## 快速开始

### 1. 构建项目
```bash
# 构建所有组件（Go二进制 + Python环境）
.\webhook-bridge.exe build --verbose

# 仅构建Go二进制，跳过Python环境
.\webhook-bridge.exe build --skip-python

# 跨平台构建
.\webhook-bridge.exe build --cross-platform
```

### 2. 启动服务
```bash
# 开发模式启动
.\webhook-bridge.exe start --env dev --verbose

# 生产模式启动
.\webhook-bridge.exe start --env prod

# 后台运行
.\webhook-bridge.exe start --daemon

# 自定义端口
.\webhook-bridge.exe start --server-port 8080 --executor-port 50051
```

### 3. 管理服务
```bash
# 检查服务状态
.\webhook-bridge.exe status

# 停止服务
.\webhook-bridge.exe stop

# 重启服务（等同于stop + start）
.\webhook-bridge.exe start --env dev
```

### 4. 测试
```bash
# 运行所有测试
.\webhook-bridge.exe test --verbose

# 仅运行Go测试
.\webhook-bridge.exe test --python=false

# 运行集成测试
.\webhook-bridge.exe test --integration

# 生成覆盖率报告
.\webhook-bridge.exe test --coverage
```

### 5. 配置管理
```bash
# 设置开发环境配置
.\webhook-bridge.exe config --env dev

# 设置生产环境配置
.\webhook-bridge.exe config --env prod

# 查看当前配置
.\webhook-bridge.exe config --show
```

### 6. 部署
```bash
# 开发环境部署
.\webhook-bridge.exe deploy --env dev

# 生产环境部署
.\webhook-bridge.exe deploy --env prod --skip-tests

# Docker部署
.\webhook-bridge.exe deploy --docker
```

### 7. 清理
```bash
# 清理构建产物
.\webhook-bridge.exe clean
```

## 命令详解

### build - 构建命令
自动检测Go和Python环境，构建所有必要的组件：

- 检测Go编译器（支持标准安装路径）
- 构建Go二进制文件（server + python-manager）
- 创建Python虚拟环境（优先使用UV）
- 安装Python依赖

### start - 启动命令
智能启动服务，包含以下功能：

- 自动构建（可选）
- 配置文件管理
- 端口自动分配
- 进程管理
- 优雅关闭

### status - 状态检查
显示详细的服务状态：

- Go服务器状态和PID
- Python执行器状态和PID
- 构建状态检查
- 配置文件状态

### test - 测试命令
运行各种测试：

- Go单元测试
- Python测试
- 集成测试
- 覆盖率报告

## 环境要求

### 必需
- **Go 1.21+** - 用于编译Go组件
- **Python 3.8+** - 用于Python执行器

### 推荐
- **UV** - 更快的Python包管理器
- **Git** - 版本信息获取

## 配置文件

工具支持多种配置文件：

- `config.dev.yaml` - 开发环境配置
- `config.prod.yaml` - 生产环境配置
- `config.example.yaml` - 示例配置
- `config.yaml` - 当前使用的配置

## 自动化特性

### 智能路径检测
- 自动检测Go安装路径（Windows: `C:\Program Files\Go\bin\go.exe`）
- 自动检测Python命令（python3, python）
- 自动检测UV工具

### 依赖管理
- 优先使用UV进行Python包管理
- 自动回退到标准pip
- 虚拟环境自动创建和管理

### 端口管理
- 自动端口分配避免冲突
- 支持环境变量覆盖
- 配置文件端口设置

## 故障排除

### Go编译器未找到
```bash
# 检查Go安装
go version

# 或者设置完整路径
export PATH=$PATH:/usr/local/go/bin  # Linux/macOS
# 或在Windows中添加到系统PATH
```

### Python环境问题
```bash
# 检查Python安装
python3 --version

# 手动创建虚拟环境
python3 -m venv .venv

# 激活虚拟环境并安装依赖
source .venv/bin/activate  # Linux/macOS
.venv\Scripts\activate     # Windows
pip install -r requirements.txt
```

### 端口冲突
```bash
# 使用自定义端口
.\webhook-bridge.exe start --server-port 8081 --executor-port 50052

# 或设置环境变量
set WEBHOOK_BRIDGE_PORT=8081
set WEBHOOK_BRIDGE_EXECUTOR_PORT=50052
```

## 开发工作流

### 典型开发流程
```bash
# 1. 初始设置
.\webhook-bridge.exe build --verbose

# 2. 开发模式启动
.\webhook-bridge.exe start --env dev --verbose

# 3. 运行测试
.\webhook-bridge.exe test

# 4. 清理和重新构建
.\webhook-bridge.exe clean
.\webhook-bridge.exe build
```

### 生产部署流程
```bash
# 1. 完整部署
.\webhook-bridge.exe deploy --env prod

# 2. 或分步执行
.\webhook-bridge.exe clean
.\webhook-bridge.exe build --cross-platform
.\webhook-bridge.exe test
.\webhook-bridge.exe config --env prod
.\webhook-bridge.exe start --env prod --daemon
```

## 版本信息

```bash
# 查看版本信息
.\webhook-bridge.exe version

# 输出示例：
# Webhook Bridge dev
# Build Time: unknown
# Go Version: unknown
# Runtime: windows/amd64
```

这个统一的CLI工具大大简化了webhook bridge的开发和部署流程，提供了生产级的自动化管理能力。
