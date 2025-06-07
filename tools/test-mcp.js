#!/usr/bin/env node

/**
 * 测试 React Debug MCP 服务器
 */

import { spawn } from 'child_process';
import { fileURLToPath } from 'url';
import { dirname, join } from 'path';

const __filename = fileURLToPath(import.meta.url);
const __dirname = dirname(__filename);

console.log('🧪 测试 React Debug MCP 服务器...\n');

// 启动 MCP Inspector
const inspectorProcess = spawn('npx', [
  '@modelcontextprotocol/inspector',
  'node',
  join(__dirname, 'react-debug-mcp.js')
], {
  stdio: 'inherit',
  shell: true
});

console.log('🔍 MCP Inspector 已启动');
console.log('📋 可用工具：');
console.log('  - react_navigate: 导航到 React 应用');
console.log('  - react_component_inspect: 检查组件状态');
console.log('  - react_state_debug: 调试组件状态');
console.log('  - nextjs_route_test: 测试 Next.js 路由');
console.log('  - react_performance_check: 性能检查');
console.log('  - react_screenshot: 截图功能');
console.log('\n💡 在浏览器中测试这些工具！');

inspectorProcess.on('close', (code) => {
  console.log(`\n✅ MCP Inspector 已关闭 (代码: ${code})`);
});

process.on('SIGINT', () => {
  console.log('\n⚠️ 正在关闭...');
  inspectorProcess.kill();
  process.exit(0);
});
