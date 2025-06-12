# Webhook Bridge CLI 使用指南

Webhook Bridge 提供了一个统一的命令行工具，将原本的多个可执行文件整合为单一二进制文件，支持开发、测试、部署和运维的完整生命周期。

## 快速开始

### 安装和基本使用

```bash
# 下载并解压 webhook-bridge 发布包
# 或者从源码构建
go build -o webhook-bridge cmd/webhook-bridge/main.go

# 查看帮助
./webhook-bridge --help

# 快速启动（推荐新用户）
./webhook-bridge serve

# 完整开发模式启动
./webhook-bridge start
```

## 核心命令详解

### 1. `unified` - 统一服务模式 ⭐ (推荐)

**最完整的启动方式，自动管理Python执行器和Go服务器**

```bash
# 基本启动（推荐）
webhook-bridge unified

# 指定端口
webhook-bridge unified --port 8080

# 生产环境模式
webhook-bridge unified --mode release --port 8080

# 指定配置文件
webhook-bridge unified --config /path/to/config.yaml

# 详细输出
webhook-bridge unified --verbose

# API模式（不启动Python执行器）
webhook-bridge unified --no-python
```

**特点：**
- ✅ 单一命令启动所有服务
- ✅ 自动管理Python执行器
- ✅ 完整的插件功能支持
- ✅ 统一进程管理
- ✅ 优雅的服务关闭

### 2. `serve` - 独立服务器模式

**仅启动Go HTTP服务器，不包含Python执行器**

```bash
# 基本启动
webhook-bridge serve

# 指定端口
webhook-bridge serve --port 9000

# 生产环境模式
webhook-bridge serve --env prod --port 8080
```

**特点：**
- ✅ 轻量级，快速启动
- ✅ 无需Python环境
- ⚠️ Python插件功能不可用

### 3. `server` - 后端服务器模式

**启动后端服务器，包含gRPC客户端功能**

```bash
# 基本启动
webhook-bridge server

# 指定端口
webhook-bridge server --port 8080

# 详细输出
webhook-bridge server --verbose
```

**特点：**
- ✅ 包含gRPC客户端
- ✅ 支持连接外部Python执行器
- 🔧 需要单独启动Python执行器

### 4. `python` - Python环境管理

**管理Python环境和依赖**

```bash
# 显示Python环境信息
webhook-bridge python info

# 验证Python环境
webhook-bridge python validate

# 安装Python包
webhook-bridge python install grpcio requests

# 启动Python执行器服务
webhook-bridge python executor
```

**特点：**
- ✅ 统一的Python环境管理
- ✅ 自动环境检测
- ✅ 依赖安装和验证

### 5. `start` - 完整开发模式

**传统的完整功能模式，包含Go服务器和Python执行器**

```bash
# 开发模式启动
webhook-bridge start

# 生产模式启动
webhook-bridge start --env prod

# 强制重新构建
webhook-bridge start --force-build

# 后台运行
webhook-bridge start --daemon
```

**特点：**
- ✅ 完整的Python插件支持
- ✅ 智能构建检测
- ✅ 自动Python环境检测
- 🔧 传统多进程管理方式

### 6. `dashboard` - Web管理界面

**启动服务并打开Web管理界面**

```bash
# 启动并打开浏览器
webhook-bridge dashboard

# 不自动打开浏览器
webhook-bridge dashboard --no-browser

# 指定端口
webhook-bridge dashboard --port 9000

# 生产模式
webhook-bridge dashboard --env prod
```

**访问地址：**
- 🌐 Dashboard界面: `http://localhost:8080/dashboard`
- 🔍 API文档: `http://localhost:8080/api`
- ❤️ 健康检查: `http://localhost:8080/health`

## 开发和构建命令

### 4. `build` - 构建项目

```bash
# 构建所有组件
webhook-bridge build

# 只构建Go组件
webhook-bridge build --go-only

# 只构建Python环境
webhook-bridge build --python-only

# 强制重新构建
webhook-bridge build --force

# 详细输出
webhook-bridge build --verbose
```

### 5. `test` - 运行测试

```bash
# 运行所有测试
webhook-bridge test

# 只运行Go测试
webhook-bridge test --go --no-python

# 只运行Python测试
webhook-bridge test --python --no-go

# 运行集成测试
webhook-bridge test --integration

# 生成覆盖率报告
webhook-bridge test --coverage
```

### 6. `clean` - 清理构建产物

```bash
# 清理所有构建产物
webhook-bridge clean

# 详细输出
webhook-bridge clean --verbose
```

## 运维和管理命令

### 7. `status` - 查看服务状态

```bash
# 查看服务状态
webhook-bridge status

# 详细状态信息
webhook-bridge status --verbose
```

**输出示例：**
```
📊 Webhook Bridge Service Status
================================
🚀 Go Server: ✅ Running (PID: 12345)
🐍 Python Executor: ✅ Running (PID: 12346)

🔨 Build Status:
  🚀 Go Server: ✅ Built
  🔧 Python Manager: ✅ Built
  🐍 Python Environment: ✅ Ready

📝 Configuration:
  📝 config.yaml: ✅ Present
  📝 config.dev.yaml: ✅ Present
  📝 config.prod.yaml: ✅ Present
```

### 8. `stop` - 停止服务

```bash
# 停止所有服务
webhook-bridge stop

# 详细输出
webhook-bridge stop --verbose
```

### 9. `config` - 配置管理

```bash
# 显示当前配置
webhook-bridge config --show

# 设置开发环境
webhook-bridge config --env dev

# 设置生产环境
webhook-bridge config --env prod
```

## 部署命令

### 10. `deploy` - 部署应用

```bash
# 标准部署
webhook-bridge deploy

# 生产环境部署
webhook-bridge deploy --env prod

# 跳过测试
webhook-bridge deploy --skip-tests

# Docker部署
webhook-bridge deploy --docker
```

## 环境配置

### 开发环境设置

```bash
# 1. 初始构建
webhook-bridge build

# 2. 启动开发服务
webhook-bridge start --env dev

# 3. 打开管理界面
webhook-bridge dashboard
```

### 生产环境设置

```bash
# 1. 构建生产版本
webhook-bridge build

# 2. 部署
webhook-bridge deploy --env prod

# 3. 启动生产服务
webhook-bridge serve --env prod --port 8080
```

### 测试环境设置

```bash
# 1. 运行所有测试
webhook-bridge test --coverage

# 2. 启动测试服务
webhook-bridge serve --env dev --port 9000

# 3. 运行集成测试
webhook-bridge test --integration
```

## 调试指南

### 1. 详细日志输出

```bash
# 所有命令都支持 --verbose 参数
webhook-bridge start --verbose
webhook-bridge serve --verbose
webhook-bridge build --verbose
```

### 2. 检查服务状态

```bash
# 查看详细状态
webhook-bridge status --verbose

# 检查配置
webhook-bridge config --show
```

### 3. 重新构建解决问题

```bash
# 清理并重新构建
webhook-bridge clean
webhook-bridge build --force --verbose
```

### 4. 测试连接

```bash
# 检查服务健康状态
curl http://localhost:8000/health

# 检查API端点
curl http://localhost:8000/api/v1/plugins

# 查看指标
curl http://localhost:8000/metrics
```

### 5. 日志文件位置

- **服务日志**: `logs/webhook-bridge.log`
- **Python执行器日志**: `logs/python-executor.log`
- **构建日志**: 控制台输出（使用 `--verbose`）

## 常见使用场景

### 场景1：快速演示（推荐）
```bash
webhook-bridge unified --verbose
# 访问 http://localhost:8080/dashboard
# 完整功能，包含Python插件支持
```

### 场景2：轻量级API服务
```bash
webhook-bridge serve --verbose
# 仅Go服务器，无Python插件功能
```

### 场景3：开发调试
```bash
webhook-bridge unified --verbose
# 或者使用传统方式
webhook-bridge start --env dev --verbose
```

### 场景4：生产部署
```bash
webhook-bridge deploy --env prod
webhook-bridge unified --mode release --port 8080
```

### 场景5：Python环境管理
```bash
webhook-bridge python info
webhook-bridge python validate
webhook-bridge python install grpcio
```

### 场景6：CI/CD集成
```bash
webhook-bridge clean
webhook-bridge build
webhook-bridge test --coverage
webhook-bridge deploy --skip-tests
```

### 场景7：问题排查
```bash
webhook-bridge status --verbose
webhook-bridge config --show
webhook-bridge python info
webhook-bridge clean && webhook-bridge build --force --verbose
```

## 全局参数

所有命令都支持以下全局参数：

- `--verbose, -v`: 启用详细输出
- `--config, -c`: 指定配置文件路径
- `--help, -h`: 显示帮助信息

## 配置文件优先级

1. 命令行参数（最高优先级）
2. `--config` 指定的配置文件
3. `config.yaml`（当前目录）
4. `config.dev.yaml` 或 `config.prod.yaml`
5. `config.example.yaml`
6. 默认配置（最低优先级）

## 端口配置

- **默认服务器端口**: 8000
- **默认执行器端口**: 50051
- **自动端口检测**: 如果端口被占用，会自动选择可用端口
- **端口覆盖**: 使用 `--port` 或 `--server-port` 参数

## 故障排除

### 常见问题和解决方案

#### 1. 端口被占用
```bash
# 问题：Error: listen tcp :8000: bind: address already in use
# 解决：使用不同端口或停止占用进程
webhook-bridge serve --port 9000

# 或者找到并停止占用进程（Windows）
netstat -ano | findstr :8000
taskkill /PID <PID> /F

# Linux/macOS
lsof -ti:8000 | xargs kill -9
```

#### 2. Python环境问题
```bash
# 问题：Python executor failed to start
# 解决：重新构建Python环境
webhook-bridge clean
webhook-bridge build --python-only --verbose

# 检查Python版本
python --version
# 确保Python 3.8+
```

#### 3. 构建失败
```bash
# 问题：Build failed
# 解决：清理并强制重新构建
webhook-bridge clean
webhook-bridge build --force --verbose

# 检查Go版本
go version
# 确保Go 1.19+
```

#### 4. 配置文件问题
```bash
# 问题：Configuration validation failed
# 解决：检查配置文件语法
webhook-bridge config --show

# 重置为默认配置
cp config.example.yaml config.yaml
```

#### 5. 权限问题
```bash
# 问题：Permission denied
# 解决：检查文件权限（Linux/macOS）
chmod +x webhook-bridge
chmod +x build/webhook-bridge-server

# Windows：以管理员身份运行
```

### 性能调优

#### 1. 生产环境优化
```bash
# 使用生产模式
webhook-bridge serve --env prod

# 配置文件优化（config.yaml）
server:
  mode: "release"
  workers: 8  # CPU核心数的2倍

logging:
  level: "info"  # 减少日志输出

executor:
  pool_size: 10  # 根据负载调整
```

#### 2. 内存优化
```bash
# 限制日志文件大小
logging:
  max_size: 100  # MB
  max_age: 7     # 天
  compress: true
```

#### 3. 并发优化
```bash
# 调整工作池大小
server:
  read_timeout: 30s
  write_timeout: 30s
  idle_timeout: 120s
```

### 监控和日志

#### 1. 实时监控
```bash
# 查看实时日志
tail -f logs/webhook-bridge.log

# Windows
Get-Content logs/webhook-bridge.log -Wait
```

#### 2. 健康检查
```bash
# 基本健康检查
curl http://localhost:8000/health

# 详细指标
curl http://localhost:8000/metrics

# 工作池状态
curl http://localhost:8000/workers
```

#### 3. 性能指标
```bash
# API响应时间测试
curl -w "@curl-format.txt" -o /dev/null -s http://localhost:8000/api/v1/plugins

# curl-format.txt 内容：
#     time_namelookup:  %{time_namelookup}\n
#        time_connect:  %{time_connect}\n
#     time_appconnect:  %{time_appconnect}\n
#    time_pretransfer:  %{time_pretransfer}\n
#       time_redirect:  %{time_redirect}\n
#  time_starttransfer:  %{time_starttransfer}\n
#                     ----------\n
#          time_total:  %{time_total}\n
```

## 高级用法

### 1. 自定义插件开发

```bash
# 创建新插件
mkdir -p plugins/my_plugin
cat > plugins/my_plugin/plugin.py << 'EOF'
from webhook_bridge.plugin import BasePlugin

class MyPlugin(BasePlugin):
    def handle(self):
        return {
            "message": "Hello from my plugin!",
            "data": self.data
        }
EOF

# 测试插件
webhook-bridge start --verbose
curl -X POST http://localhost:8000/api/v1/webhook/my_plugin \
  -H "Content-Type: application/json" \
  -d '{"test": "data"}'
```

### 2. 环境变量配置

```bash
# 设置环境变量
export WEBHOOK_BRIDGE_PORT=9000
export WEBHOOK_BRIDGE_MODE=release
export WEBHOOK_BRIDGE_LOG_LEVEL=info

# 启动服务（会使用环境变量）
webhook-bridge serve
```

### 3. Docker部署

```bash
# 构建Docker镜像
webhook-bridge deploy --docker

# 运行容器
docker run -d \
  --name webhook-bridge \
  -p 8000:8000 \
  -v $(pwd)/plugins:/app/plugins \
  -v $(pwd)/config.yaml:/app/config.yaml \
  webhook-bridge:latest
```

### 4. 系统服务集成

```bash
# 创建systemd服务（Linux）
sudo tee /etc/systemd/system/webhook-bridge.service << 'EOF'
[Unit]
Description=Webhook Bridge Service
After=network.target

[Service]
Type=simple
User=webhook
WorkingDirectory=/opt/webhook-bridge
ExecStart=/opt/webhook-bridge/webhook-bridge serve --env prod
Restart=always
RestartSec=5

[Install]
WantedBy=multi-user.target
EOF

# 启用并启动服务
sudo systemctl enable webhook-bridge
sudo systemctl start webhook-bridge
sudo systemctl status webhook-bridge
```

### 5. 负载均衡配置

```bash
# Nginx配置示例
upstream webhook_bridge {
    server 127.0.0.1:8000;
    server 127.0.0.1:8001;
    server 127.0.0.1:8002;
}

server {
    listen 80;
    server_name webhook.example.com;

    location / {
        proxy_pass http://webhook_bridge;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
    }
}
```

## 开发工作流

### 1. 日常开发流程

```bash
# 1. 拉取最新代码
git pull origin main

# 2. 清理并构建
webhook-bridge clean
webhook-bridge build --verbose

# 3. 运行测试
webhook-bridge test --coverage

# 4. 启动开发服务
webhook-bridge start --env dev --verbose

# 5. 开发完成后测试
webhook-bridge test --integration

# 6. 提交代码
git add .
git commit -m "feat: add new feature"
git push origin feature-branch
```

### 2. 发布流程

```bash
# 1. 版本测试
webhook-bridge clean
webhook-bridge build
webhook-bridge test --coverage --integration

# 2. 构建发布版本
webhook-bridge deploy --env prod --skip-tests

# 3. 创建发布包
tar -czf webhook-bridge-v1.0.0.tar.gz \
  webhook-bridge \
  build/ \
  config.example.yaml \
  README.md \
  LICENSE

# 4. 部署到生产环境
scp webhook-bridge-v1.0.0.tar.gz user@server:/opt/
ssh user@server "cd /opt && tar -xzf webhook-bridge-v1.0.0.tar.gz"
ssh user@server "cd /opt/webhook-bridge && ./webhook-bridge serve --env prod"
```

### 3. 持续集成示例

```yaml
# .github/workflows/ci.yml
name: CI
on: [push, pull_request]

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
    - uses: actions/checkout@v3
    - uses: actions/setup-go@v3
      with:
        go-version: '1.19'

    - name: Build
      run: go run cmd/webhook-bridge/main.go build

    - name: Test
      run: go run cmd/webhook-bridge/main.go test --coverage

    - name: Deploy
      if: github.ref == 'refs/heads/main'
      run: go run cmd/webhook-bridge/main.go deploy --env prod
```

## 最佳实践

### 1. 安全配置

```yaml
# config.yaml - 生产环境安全配置
server:
  mode: "release"
  host: "127.0.0.1"  # 只监听本地，通过反向代理暴露

logging:
  level: "info"  # 不记录敏感的debug信息

security:
  rate_limit: 100  # 每分钟请求限制
  timeout: 30s     # 请求超时
```

### 2. 性能配置

```yaml
# 高性能配置
server:
  workers: 16        # 根据CPU核心数调整
  max_connections: 1000

executor:
  pool_size: 20      # Python执行器池大小
  timeout: 60s       # 插件执行超时

logging:
  async: true        # 异步日志
  buffer_size: 1000  # 日志缓冲区
```

### 3. 监控配置

```yaml
# 监控和指标配置
monitoring:
  enabled: true
  metrics_path: "/metrics"
  health_path: "/health"

logging:
  structured: true   # 结构化日志，便于分析
  format: "json"     # JSON格式，便于日志聚合
```

这个CLI使用指南涵盖了从基础使用到高级配置的所有方面，帮助用户在不同场景下有效使用webhook-bridge。
