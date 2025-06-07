# React/Next.js Debug MCP 工具集

专门为 React 和 Next.js 开发调试设计的 MCP 服务器集合。

## 🚀 推荐的 MCP 工具

### 1. 自定义 React Debug MCP (最推荐)
专门为你的项目定制的 React/Next.js 调试工具。

**功能特性：**
- ✅ React 组件状态检查
- ✅ Props 和 State 调试
- ✅ Next.js 路由测试
- ✅ 性能分析
- ✅ 组件截图
- ✅ React DevTools 集成

### 2. Puppeteer MCP
强大的浏览器自动化工具。

### 3. Browser MCP
基础的浏览器操作工具。

## 📦 安装步骤

### 1. 安装依赖
```bash
cd tools
npm install
```

### 2. 配置 Claude Desktop
将 `claude-desktop-config.json` 的内容复制到：
```
%APPDATA%\Claude\claude_desktop_config.json
```

### 3. 重启 Claude Desktop
完全关闭并重新启动 Claude Desktop。

## 🎯 使用方法

### React 组件调试
```
请帮我调试 React 组件的状态
- 导航到 http://localhost:3000
- 检查 UserProfile 组件的 props
- 截图保存当前状态
```

### Next.js 路由测试
```
测试 Next.js API 路由
- 测试 /api/users GET 请求
- 检查返回的数据格式
- 验证状态码
```

### 性能分析
```
分析 React 应用性能
- 检查加载时间
- 分析内存使用
- 查看 JS/CSS 覆盖率
```

## 🔧 可用工具

| 工具名称 | 功能描述 | 使用场景 |
|---------|---------|---------|
| `react_navigate` | 导航到 React 应用 | 打开开发服务器 |
| `react_component_inspect` | 检查组件信息 | 调试组件 props/state |
| `react_state_debug` | 调试组件状态 | 深度状态分析 |
| `nextjs_route_test` | 测试 API 路由 | API 接口验证 |
| `react_performance_check` | 性能检查 | 性能优化 |
| `react_screenshot` | 截图功能 | 视觉验证 |

## 🌟 优势对比

| 特性 | Browser MCP | Puppeteer MCP | React Debug MCP |
|-----|------------|---------------|-----------------|
| React 专用 | ❌ | ⚠️ | ✅ |
| 组件调试 | ❌ | ⚠️ | ✅ |
| Next.js 支持 | ❌ | ⚠️ | ✅ |
| 性能分析 | ❌ | ✅ | ✅ |
| 自定义功能 | ❌ | ⚠️ | ✅ |
| 开发友好 | ⚠️ | ✅ | ✅ |

## 🚀 快速开始

1. **启动你的 React/Next.js 应用**
   ```bash
   npm run dev  # 通常在 http://localhost:3000
   ```

2. **在 Claude 中使用**
   ```
   请帮我调试 React 应用：
   1. 导航到 localhost:3000
   2. 检查 Header 组件的状态
   3. 测试 /api/auth 路由
   ```

3. **查看结果**
   - 组件信息会以 JSON 格式显示
   - 截图会保存到项目目录
   - 性能报告包含详细指标

## 🔍 调试技巧

### 1. 组件选择器
```javascript
// CSS 选择器
"[data-testid='user-profile']"
".header-component"
"#main-content"

// React 组件名
"UserProfile"
"Header"
"MainContent"
```

### 2. 性能优化
- 使用性能检查工具定期监控
- 关注内存使用和加载时间
- 分析 JS/CSS 覆盖率

### 3. 路由测试
- 测试所有 API 端点
- 验证错误处理
- 检查响应格式

## 🛠️ 故障排除

### 常见问题

1. **MCP 服务器未连接**
   - 检查 Claude Desktop 配置
   - 确保路径正确
   - 重启 Claude Desktop

2. **React 组件未检测到**
   - 确保应用正在运行
   - 检查 React DevTools 是否可用
   - 验证组件选择器

3. **性能检查失败**
   - 确保网络连接正常
   - 检查应用是否响应
   - 验证 URL 正确性

### 日志调试
```bash
# 查看 MCP 服务器日志
node react-debug-mcp.js 2>&1 | tee debug.log
```

## 📚 扩展功能

你可以根据需要扩展更多功能：
- Redux/Zustand 状态调试
- React Query 缓存检查
- 组件渲染性能分析
- 自动化测试集成

## 🤝 贡献

欢迎提交 Issue 和 Pull Request 来改进这个工具！
