#!/usr/bin/env node

/**
 * React/Next.js Debug MCP Server
 * 专门为 React 和 Next.js 开发调试设计的 MCP 服务器
 */

import { Server } from '@modelcontextprotocol/sdk/server/index.js';
import { StdioServerTransport } from '@modelcontextprotocol/sdk/server/stdio.js';
import {
  CallToolRequestSchema,
  ListToolsRequestSchema,
} from '@modelcontextprotocol/sdk/types.js';
import puppeteer from 'puppeteer';

class ReactDebugMCP {
  constructor() {
    this.server = new Server(
      {
        name: 'react-debug-mcp',
        version: '1.0.0',
      },
      {
        capabilities: {
          tools: {},
        },
      }
    );

    this.browser = null;
    this.page = null;
    this.setupToolHandlers();
  }

  setupToolHandlers() {
    // 列出可用工具
    this.server.setRequestHandler(ListToolsRequestSchema, async () => {
      return {
        tools: [
          {
            name: 'react_navigate',
            description: '导航到 React/Next.js 应用页面',
            inputSchema: {
              type: 'object',
              properties: {
                url: {
                  type: 'string',
                  description: 'URL 地址 (默认: http://localhost:3000)',
                },
              },
            },
          },
          {
            name: 'react_component_inspect',
            description: '检查 React 组件状态和 props',
            inputSchema: {
              type: 'object',
              properties: {
                selector: {
                  type: 'string',
                  description: 'CSS 选择器或组件名称',
                },
              },
              required: ['selector'],
            },
          },
          {
            name: 'react_state_debug',
            description: '调试 React 组件状态',
            inputSchema: {
              type: 'object',
              properties: {
                component: {
                  type: 'string',
                  description: '组件名称或选择器',
                },
              },
              required: ['component'],
            },
          },
          {
            name: 'nextjs_route_test',
            description: '测试 Next.js 路由',
            inputSchema: {
              type: 'object',
              properties: {
                route: {
                  type: 'string',
                  description: '路由路径 (如: /api/users, /dashboard)',
                },
                method: {
                  type: 'string',
                  description: 'HTTP 方法',
                  enum: ['GET', 'POST', 'PUT', 'DELETE'],
                  default: 'GET',
                },
              },
              required: ['route'],
            },
          },
          {
            name: 'react_performance_check',
            description: '检查 React 应用性能',
            inputSchema: {
              type: 'object',
              properties: {
                url: {
                  type: 'string',
                  description: 'URL 地址',
                  default: 'http://localhost:3000',
                },
              },
            },
          },
          {
            name: 'react_screenshot',
            description: '截取 React 应用截图',
            inputSchema: {
              type: 'object',
              properties: {
                selector: {
                  type: 'string',
                  description: '要截图的元素选择器 (可选)',
                },
                filename: {
                  type: 'string',
                  description: '保存文件名',
                  default: 'react-debug-screenshot.png',
                },
              },
            },
          },
        ],
      };
    });

    // 处理工具调用
    this.server.setRequestHandler(CallToolRequestSchema, async (request) => {
      const { name, arguments: args } = request.params;

      try {
        switch (name) {
          case 'react_navigate':
            return await this.navigateToReactApp(args.url || 'http://localhost:3000');

          case 'react_component_inspect':
            return await this.inspectReactComponent(args.selector);

          case 'react_state_debug':
            return await this.debugReactState(args.component);

          case 'nextjs_route_test':
            return await this.testNextjsRoute(args.route, args.method || 'GET');

          case 'react_performance_check':
            return await this.checkReactPerformance(args.url || 'http://localhost:3000');

          case 'react_screenshot':
            return await this.takeReactScreenshot(args.selector, args.filename);

          default:
            throw new Error(`Unknown tool: ${name}`);
        }
      } catch (error) {
        return {
          content: [
            {
              type: 'text',
              text: `Error: ${error.message}`,
            },
          ],
        };
      }
    });
  }

  async ensureBrowser() {
    if (!this.browser) {
      this.browser = await puppeteer.launch({
        headless: false,
        devtools: true,
        args: ['--no-sandbox', '--disable-setuid-sandbox'],
      });
    }
    if (!this.page) {
      this.page = await this.browser.newPage();
      // 启用 React DevTools
      await this.page.evaluateOnNewDocument(() => {
        window.__REACT_DEVTOOLS_GLOBAL_HOOK__ = window.__REACT_DEVTOOLS_GLOBAL_HOOK__ || {};
      });
    }
    return this.page;
  }

  async navigateToReactApp(url) {
    const page = await this.ensureBrowser();
    await page.goto(url, { waitUntil: 'networkidle0' });
    
    // 检查是否是 React 应用
    const isReact = await page.evaluate(() => {
      return !!(window.React || window.__REACT_DEVTOOLS_GLOBAL_HOOK__);
    });

    return {
      content: [
        {
          type: 'text',
          text: `✅ 已导航到: ${url}\n${isReact ? '🎯 检测到 React 应用' : '⚠️ 未检测到 React'}`,
        },
      ],
    };
  }

  async inspectReactComponent(selector) {
    const page = await this.ensureBrowser();
    
    const componentInfo = await page.evaluate((sel) => {
      const element = document.querySelector(sel);
      if (!element) return null;

      // 尝试获取 React 组件信息
      const reactKey = Object.keys(element).find(key => 
        key.startsWith('__reactInternalInstance') || key.startsWith('__reactFiber')
      );

      if (reactKey) {
        const fiber = element[reactKey];
        return {
          componentName: fiber.type?.name || fiber.elementType?.name || 'Unknown',
          props: fiber.memoizedProps || {},
          state: fiber.memoizedState || {},
          key: fiber.key,
        };
      }

      return {
        tagName: element.tagName,
        className: element.className,
        id: element.id,
        textContent: element.textContent?.substring(0, 100),
      };
    }, selector);

    return {
      content: [
        {
          type: 'text',
          text: componentInfo 
            ? `🔍 组件信息:\n${JSON.stringify(componentInfo, null, 2)}`
            : `❌ 未找到选择器: ${selector}`,
        },
      ],
    };
  }

  async debugReactState(component) {
    const page = await this.ensureBrowser();
    
    const stateInfo = await page.evaluate((comp) => {
      // 查找所有 React 组件
      const allElements = document.querySelectorAll('*');
      const components = [];

      for (const element of allElements) {
        const reactKey = Object.keys(element).find(key => 
          key.startsWith('__reactInternalInstance') || key.startsWith('__reactFiber')
        );

        if (reactKey) {
          const fiber = element[reactKey];
          const componentName = fiber.type?.name || fiber.elementType?.name;
          
          if (componentName && componentName.toLowerCase().includes(comp.toLowerCase())) {
            components.push({
              name: componentName,
              props: fiber.memoizedProps || {},
              state: fiber.memoizedState || {},
              hooks: fiber.hooks || [],
            });
          }
        }
      }

      return components;
    }, component);

    return {
      content: [
        {
          type: 'text',
          text: stateInfo.length > 0
            ? `🎯 找到 ${stateInfo.length} 个匹配组件:\n${JSON.stringify(stateInfo, null, 2)}`
            : `❌ 未找到组件: ${component}`,
        },
      ],
    };
  }

  async testNextjsRoute(route, method) {
    const page = await this.ensureBrowser();
    
    try {
      const response = await page.evaluate(async (route, method) => {
        const res = await fetch(route, { method });
        return {
          status: res.status,
          statusText: res.statusText,
          headers: Object.fromEntries(res.headers.entries()),
          body: await res.text(),
        };
      }, route, method);

      return {
        content: [
          {
            type: 'text',
            text: `🌐 Next.js 路由测试结果:\n${JSON.stringify(response, null, 2)}`,
          },
        ],
      };
    } catch (error) {
      return {
        content: [
          {
            type: 'text',
            text: `❌ 路由测试失败: ${error.message}`,
          },
        ],
      };
    }
  }

  async checkReactPerformance(url) {
    const page = await this.ensureBrowser();
    
    // 启用性能监控
    await page.coverage.startJSCoverage();
    await page.coverage.startCSSCoverage();
    
    const startTime = Date.now();
    await page.goto(url, { waitUntil: 'networkidle0' });
    const loadTime = Date.now() - startTime;

    const metrics = await page.metrics();
    const jsCoverage = await page.coverage.stopJSCoverage();
    const cssCoverage = await page.coverage.stopCSSCoverage();

    return {
      content: [
        {
          type: 'text',
          text: `📊 React 应用性能报告:
⏱️ 加载时间: ${loadTime}ms
🧠 内存使用: ${Math.round(metrics.JSHeapUsedSize / 1024 / 1024)}MB
📦 JS 覆盖率: ${jsCoverage.length} 个文件
🎨 CSS 覆盖率: ${cssCoverage.length} 个文件
📈 详细指标: ${JSON.stringify(metrics, null, 2)}`,
        },
      ],
    };
  }

  async takeReactScreenshot(selector, filename) {
    const page = await this.ensureBrowser();
    
    const screenshotOptions = {
      path: filename || 'react-debug-screenshot.png',
      fullPage: !selector,
    };

    if (selector) {
      const element = await page.$(selector);
      if (element) {
        await element.screenshot(screenshotOptions);
      } else {
        throw new Error(`Element not found: ${selector}`);
      }
    } else {
      await page.screenshot(screenshotOptions);
    }

    return {
      content: [
        {
          type: 'text',
          text: `📸 截图已保存: ${screenshotOptions.path}`,
        },
      ],
    };
  }

  async run() {
    const transport = new StdioServerTransport();
    await this.server.connect(transport);
    console.error('React Debug MCP Server running...');
  }
}

// 启动服务器
const server = new ReactDebugMCP();
server.run().catch(console.error);
