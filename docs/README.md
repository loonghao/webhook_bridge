# Webhook Bridge 文档中心

欢迎来到 Webhook Bridge 的文档中心！这里包含了使用 Webhook Bridge 所需的所有文档和指南。

## 📚 文档导航

### 🚀 快速开始

- **[主要README](../README.md)** - 项目概述、安装和基本使用
- **[CLI快速参考](CLI_QUICK_REFERENCE.md)** - 常用命令速查表，适合快速查阅

### 📖 详细指南

- **[CLI使用指南](CLI_USAGE.md)** - 完整的命令行工具使用文档
  - 所有命令详解
  - 开发、测试、生产环境配置
  - 故障排除和最佳实践
  - 高级用法和集成示例

- **[Dashboard使用指南](DASHBOARD_GUIDE.md)** - Web管理界面使用文档
  - 现代化Dashboard功能
  - 插件管理和测试
  - 实时监控和日志
  - 配置管理

## 🎯 按使用场景查找

### 新用户入门
1. 阅读 [主要README](../README.md) 了解项目概述
2. 查看 [CLI快速参考](CLI_QUICK_REFERENCE.md) 学习基本命令
3. 使用 `webhook-bridge serve` 快速启动
4. 访问 `http://localhost:8000/dashboard` 体验Web界面

### 开发者
1. 阅读 [CLI使用指南](CLI_USAGE.md) 的开发环境部分
2. 使用 `webhook-bridge start --env dev --verbose` 启动开发环境
3. 查看 [Dashboard使用指南](DASHBOARD_GUIDE.md) 学习插件测试
4. 参考插件开发示例

### 运维人员
1. 查看 [CLI使用指南](CLI_USAGE.md) 的生产环境部分
2. 学习 [Dashboard使用指南](DASHBOARD_GUIDE.md) 的监控功能
3. 配置系统服务和负载均衡
4. 设置监控告警

### 故障排除
1. 查看 [CLI使用指南](CLI_USAGE.md) 的故障排除章节
2. 使用 `webhook-bridge status --verbose` 检查状态
3. 查看 [Dashboard使用指南](DASHBOARD_GUIDE.md) 的调试功能
4. 检查日志文件

## 🔧 命令速查

### 最常用命令
```bash
# 快速启动
webhook-bridge serve

# 开发模式
webhook-bridge start --env dev --verbose

# 查看状态
webhook-bridge status

# 获取帮助
webhook-bridge --help
```

### 重要URL
- **Dashboard**: http://localhost:8000/dashboard
- **API文档**: http://localhost:8000/api
- **健康检查**: http://localhost:8000/health

## 📋 文档更新

### 版本信息
- **当前版本**: v1.0.0
- **最后更新**: 2024年1月
- **兼容性**: 支持 v0.6.0 插件系统

### 贡献文档
如果您发现文档有误或需要改进，请：
1. 在 [GitHub Issues](https://github.com/loonghao/webhook_bridge/issues) 报告问题
2. 提交 Pull Request 改进文档
3. 在 [Discussions](https://github.com/loonghao/webhook_bridge/discussions) 提出建议

## 🆘 获取帮助

### 在线帮助
- **GitHub Issues**: 报告bug和功能请求
- **GitHub Discussions**: 社区讨论和问答
- **CLI帮助**: `webhook-bridge --help`

### 社区资源
- **示例插件**: `example_plugins/` 目录
- **配置示例**: `config.example.yaml`
- **测试用例**: `tests/` 目录

## 📊 文档统计

| 文档 | 内容 | 适用对象 |
|------|------|----------|
| [CLI快速参考](CLI_QUICK_REFERENCE.md) | 命令速查表 | 所有用户 |
| [CLI使用指南](CLI_USAGE.md) | 完整CLI文档 | 开发者、运维 |
| [Dashboard指南](DASHBOARD_GUIDE.md) | Web界面使用 | 所有用户 |

## 🔄 版本历史

### v1.0.0 (当前版本)
- ✅ 全新Go/Python混合架构
- ✅ 现代化Web Dashboard
- ✅ 统一CLI工具
- ✅ 向后兼容v0.6.0插件系统
- ✅ 完整文档体系

### v0.6.0 (历史版本)
- Python单体架构
- 基础Web界面
- 简单CLI工具

---

💡 **提示**: 建议将此文档页面加入书签，方便快速访问各种指南。

🔗 **相关链接**:
- [项目主页](https://github.com/loonghao/webhook_bridge)
- [发布页面](https://github.com/loonghao/webhook_bridge/releases)
- [PyPI包](https://pypi.org/project/webhook-bridge/)
