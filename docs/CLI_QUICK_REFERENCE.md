# Webhook Bridge CLI 快速参考

## 🚀 快速开始

```bash
# 最简单的启动方式
webhook-bridge serve

# 完整功能启动
webhook-bridge start

# 打开Web管理界面
webhook-bridge dashboard
```

## 📋 核心命令

| 命令 | 用途 | 示例 |
|------|------|------|
| `serve` | 独立服务器（推荐） | `webhook-bridge serve --port 8080` |
| `start` | 完整开发模式 | `webhook-bridge start --env dev` |
| `dashboard` | Web管理界面 | `webhook-bridge dashboard --no-browser` |
| `build` | 构建项目 | `webhook-bridge build --verbose` |
| `test` | 运行测试 | `webhook-bridge test --coverage` |
| `status` | 查看状态 | `webhook-bridge status --verbose` |
| `stop` | 停止服务 | `webhook-bridge stop` |
| `clean` | 清理构建 | `webhook-bridge clean` |

## 🌐 重要URL

- **Dashboard**: http://localhost:8000/dashboard
- **API文档**: http://localhost:8000/api
- **健康检查**: http://localhost:8000/health
- **指标监控**: http://localhost:8000/metrics

## ⚙️ 常用参数

| 参数 | 说明 | 示例 |
|------|------|------|
| `--verbose, -v` | 详细输出 | `webhook-bridge serve -v` |
| `--env, -e` | 环境模式 | `webhook-bridge start -e prod` |
| `--port` | 服务器端口 | `webhook-bridge serve --port 9000` |
| `--config, -c` | 配置文件 | `webhook-bridge serve -c config.prod.yaml` |
| `--help, -h` | 帮助信息 | `webhook-bridge serve -h` |

## 🔧 开发环境

```bash
# 1. 初始化
webhook-bridge build

# 2. 开发启动
webhook-bridge start --env dev --verbose

# 3. 运行测试
webhook-bridge test --coverage

# 4. 清理重建
webhook-bridge clean && webhook-bridge build --force
```

## 🚀 生产环境

```bash
# 1. 构建生产版本
webhook-bridge build

# 2. 运行测试
webhook-bridge test --coverage --integration

# 3. 启动生产服务
webhook-bridge serve --env prod --port 8080

# 4. 检查状态
webhook-bridge status
```

## 🐛 故障排除

```bash
# 端口被占用
webhook-bridge serve --port 9000

# Python环境问题
webhook-bridge clean
webhook-bridge build --python-only --verbose

# 构建失败
webhook-bridge clean
webhook-bridge build --force --verbose

# 查看详细状态
webhook-bridge status --verbose

# 检查配置
webhook-bridge config --show
```

## 📊 监控命令

```bash
# 健康检查
curl http://localhost:8000/health

# 性能指标
curl http://localhost:8000/metrics

# 工作池状态
curl http://localhost:8000/workers

# 实时日志
tail -f logs/webhook-bridge.log
```

## 🔌 插件测试

```bash
# 列出所有插件
curl http://localhost:8000/api/v1/plugins

# 测试插件
curl -X POST http://localhost:8000/api/v1/webhook/example \
  -H "Content-Type: application/json" \
  -d '{"test": "data"}'

# 查看插件信息
curl http://localhost:8000/api/v1/plugins/example
```

## 📁 目录结构

```
webhook-bridge/
├── webhook-bridge          # 主执行文件
├── config.yaml            # 主配置文件
├── logs/                  # 日志目录
├── plugins/               # 插件目录
├── build/                 # 构建产物
└── .venv/                 # Python虚拟环境
```

## 🔑 环境变量

```bash
export WEBHOOK_BRIDGE_PORT=8080
export WEBHOOK_BRIDGE_MODE=release
export WEBHOOK_BRIDGE_LOG_LEVEL=info
export WEBHOOK_BRIDGE_CONFIG_PATH=/path/to/config.yaml
```

## 📝 配置优先级

1. 命令行参数（最高）
2. 环境变量
3. 指定的配置文件
4. config.yaml
5. config.{env}.yaml
6. config.example.yaml
7. 默认配置（最低）

## 🆘 获取帮助

```bash
# 主帮助
webhook-bridge --help

# 命令帮助
webhook-bridge serve --help
webhook-bridge start --help
webhook-bridge test --help

# 版本信息
webhook-bridge version
```

## 🎯 常见场景

### 快速演示
```bash
webhook-bridge serve --verbose
# 访问 http://localhost:8000/dashboard
```

### 开发调试
```bash
webhook-bridge start --env dev --verbose
# 完整功能，包含Python插件
```

### 生产部署
```bash
webhook-bridge serve --env prod --port 8080
# 单进程，高性能
```

### CI/CD集成
```bash
webhook-bridge clean
webhook-bridge build
webhook-bridge test --coverage
```

---

💡 **提示**: 所有命令都支持 `--verbose` 参数来获取详细输出，有助于调试问题。

📖 **详细文档**: 查看 `docs/CLI_USAGE.md` 获取完整使用指南。
