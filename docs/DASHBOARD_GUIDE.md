# Webhook Bridge Dashboard 使用指南

Webhook Bridge 提供了现代化的Web管理界面，让您可以通过浏览器轻松管理和监控webhook服务。

## 🌐 访问Dashboard

### 启动Dashboard

```bash
# 方式1：直接启动并打开浏览器
webhook-bridge dashboard

# 方式2：启动服务后手动访问
webhook-bridge serve
# 然后访问 http://localhost:8000/dashboard

# 方式3：完整开发模式
webhook-bridge start
# 然后访问 http://localhost:8000/dashboard
```

### 访问地址

| 界面 | URL | 说明 |
|------|-----|------|
| 🎛️ **Dashboard** | `http://localhost:8000/dashboard` | 主要管理界面，基于Tailwind CSS |
| 📖 **API文档** | `http://localhost:8000/api` | 交互式API文档 |
| ❤️ **健康检查** | `http://localhost:8000/health` | 服务状态检查 |
| 📈 **指标监控** | `http://localhost:8000/metrics` | 性能指标 |

## 🎛️ Dashboard功能

### 1. 概览页面 (Overview)

**主要功能：**
- 📊 服务状态实时监控
- 🔌 插件数量统计
- 📈 请求处理统计
- ⚡ 系统性能指标

**显示信息：**
```
🚀 服务状态: ✅ 运行中
🐍 Python执行器: ✅ 连接正常
🔌 已加载插件: 5个
📊 今日请求: 1,234次
⚡ 平均响应时间: 45ms
💾 内存使用: 128MB
```

### 2. 插件管理 (Plugins)

**功能列表：**
- 📋 查看所有已加载的插件
- 🔍 插件详细信息查看
- 🧪 插件在线测试
- 📝 插件文档查看
- 🔄 插件重新加载

**插件信息显示：**
- 插件名称和描述
- 支持的HTTP方法 (GET/POST/PUT/DELETE)
- 最后修改时间
- 执行统计信息

**在线测试功能：**
```json
{
  "method": "POST",
  "url": "/api/v1/webhook/example",
  "headers": {
    "Content-Type": "application/json"
  },
  "body": {
    "test": "data"
  }
}
```

### 3. 实时日志 (Logs)

**功能特性：**
- 📜 实时日志流显示
- 🔍 日志搜索和过滤
- 📊 日志级别筛选 (DEBUG/INFO/WARN/ERROR)
- 💾 日志下载功能
- 🎨 语法高亮显示

**日志过滤选项：**
- 时间范围筛选
- 日志级别筛选
- 关键词搜索
- 来源模块筛选

### 4. 配置管理 (Configuration)

**配置功能：**
- ⚙️ 当前配置查看
- 📝 配置在线编辑
- 🔄 配置重新加载
- 💾 配置备份和恢复
- ✅ 配置验证

**支持的配置项：**
- 服务器设置 (端口、主机、模式)
- 日志配置 (级别、文件路径)
- Python执行器设置
- 安全配置
- 性能调优参数

### 5. 监控指标 (Metrics)

**性能指标：**
- 📈 请求处理速度
- 💾 内存使用情况
- 🔌 插件执行统计
- 🌐 网络连接状态
- ⏱️ 响应时间分布

**图表类型：**
- 实时折线图
- 饼状图
- 柱状图
- 仪表盘

### 6. 系统信息 (System)

**系统状态：**
- 🖥️ 服务器信息 (OS、CPU、内存)
- 🐍 Python环境信息
- 📦 依赖包版本
- 🔧 构建信息
- 📊 运行时统计



## 🧪 插件测试功能

### 在线测试工具

**测试步骤：**
1. 选择要测试的插件
2. 选择HTTP方法 (GET/POST/PUT/DELETE)
3. 输入测试数据
4. 点击"发送请求"
5. 查看响应结果

**测试示例：**
```bash
# 插件: example
# 方法: POST
# 数据:
{
  "message": "Hello World",
  "timestamp": "2024-01-01T00:00:00Z"
}

# 响应:
{
  "status": "success",
  "data": {
    "processed_message": "Hello World",
    "plugin": "example",
    "execution_time": "0.045s"
  }
}
```

### 批量测试

**功能：**
- 📋 创建测试套件
- 🔄 批量执行测试
- 📊 测试结果统计
- 📝 测试报告生成

## 🔧 配置管理

### 在线配置编辑

**支持格式：**
- YAML配置文件
- JSON格式
- 环境变量

**配置验证：**
- ✅ 语法检查
- ✅ 类型验证
- ✅ 必填项检查
- ⚠️ 警告提示

**配置示例：**
```yaml
server:
  host: "0.0.0.0"
  port: 8000
  mode: "debug"

executor:
  host: "localhost"
  port: 50051
  timeout: 30

logging:
  level: "info"
  file: "logs/webhook-bridge.log"
  max_size: 100
  max_age: 7

plugins:
  directories:
    - "example_plugins"
    - "custom_plugins"
  auto_reload: true
```

### 配置备份

**备份功能：**
- 📥 自动备份
- 💾 手动备份
- 🔄 配置恢复
- 📋 备份历史

## 📈 监控和告警

### 实时监控

**监控项目：**
- 🚀 服务运行状态
- 📊 请求处理量
- ⚡ 响应时间
- 💾 资源使用率
- 🔌 插件执行状态

### 告警设置

**告警条件：**
- 响应时间超过阈值
- 错误率超过限制
- 内存使用过高
- 插件执行失败

**告警方式：**
- 🔔 浏览器通知
- 📧 邮件告警
- 📱 Webhook通知

## 🔒 安全功能

### 访问控制

**安全特性：**
- 🔐 基本认证支持
- 🛡️ CSRF保护
- 🔒 HTTPS支持
- 📝 访问日志记录

### 权限管理

**权限级别：**
- 👀 **只读** - 查看状态和日志
- 🔧 **操作** - 测试插件和重载配置
- 👑 **管理** - 完整管理权限

## 📱 移动端支持

### 响应式设计

**特点：**
- 📱 移动设备优化
- 💻 平板电脑支持
- 🖥️ 桌面端完整功能
- 🎨 自适应布局

### 移动端功能

**主要功能：**
- 📊 状态监控
- 🔍 日志查看
- 🧪 简单测试
- ⚙️ 基本配置

## 🚀 使用技巧

### 快捷键

| 快捷键 | 功能 |
|--------|------|
| `Ctrl + R` | 刷新页面 |
| `Ctrl + F` | 搜索日志 |
| `Ctrl + S` | 保存配置 |
| `F5` | 重新加载 |

### 浏览器兼容性

**支持的浏览器：**
- ✅ Chrome 90+
- ✅ Firefox 88+
- ✅ Safari 14+
- ✅ Edge 90+

### 性能优化

**优化建议：**
- 🔄 启用浏览器缓存
- 📊 限制日志显示数量
- ⚡ 使用WebSocket实时更新
- 💾 定期清理历史数据

## 🆘 故障排除

### 常见问题

**1. Dashboard无法访问**
```bash
# 检查服务状态
webhook-bridge status

# 检查端口占用
netstat -ano | findstr :8000

# 重启服务
webhook-bridge stop
webhook-bridge serve
```

**2. 插件不显示**
```bash
# 检查插件目录
webhook-bridge config --show

# 重新加载插件
curl -X POST http://localhost:8000/api/v1/reload
```

**3. 实时日志不更新**
- 检查WebSocket连接
- 刷新浏览器页面
- 检查防火墙设置

### 调试模式

**启用调试：**
```bash
# 启动调试模式
webhook-bridge serve --verbose --env dev

# 查看详细日志
tail -f logs/webhook-bridge.log
```

这个Dashboard使用指南涵盖了Web界面的所有功能，帮助用户充分利用可视化管理工具。
