# Webhook Bridge 现代化插件系统 PRD

## 📋 产品概述

### 产品名称
Webhook Bridge 现代化插件系统 v3.0

### 产品愿景
构建业界领先的低代码 webhook 桥接平台，支持多语言插件生态，提供从零代码配置到高性能原生插件的完整解决方案。

### 目标用户
- **运维工程师**: 需要快速配置 webhook 集成，无需编程
- **后端开发者**: 需要高性能插件处理复杂业务逻辑
- **前端开发者**: 熟悉 JavaScript，希望参与 webhook 处理
- **DevOps 团队**: 需要可视化管理和监控 webhook 流程

## 🎯 核心目标

### 主要目标
1. **降低使用门槛**: 提供零代码 YAML 配置插件
2. **提升性能**: 支持 Go 原生高性能插件
3. **扩展生态**: 支持 Python、Go、JavaScript 多语言插件
4. **可视化管理**: Dashboard 拖拽式插件构建器
5. **企业级特性**: 插件市场、模板库、版本管理

### 成功指标
- **开发效率**: 新插件创建时间从 2 小时降低到 10 分钟
- **性能提升**: 高频 webhook 处理性能提升 5-10 倍
- **用户采用**: 3 个月内 80% 用户使用新插件系统
- **生态丰富**: 6 个月内插件模板库达到 50+ 个

## 🏗️ 技术架构

### 整体架构
```
┌─────────────────────────────────────────────────────────────┐
│                    Webhook Bridge v3.0                     │
├─────────────────────────────────────────────────────────────┤
│  Go HTTP Server (Port 8000)                                │
│  ├─ Smart Router & Event Classifier                        │
│  ├─ Webhook Signature Validation                           │
│  └─ Plugin Type Detection & Routing                        │
├─────────────────────────────────────────────────────────────┤
│  Multi-Language Plugin Execution Layer                     │
│  ├─ Python Executor (gRPC Port 50051)                      │
│  ├─ Go Native Plugin Loader                                │
│  ├─ JavaScript Runtime (Node.js/V8)                        │
│  └─ YAML Config Engine                                     │
├─────────────────────────────────────────────────────────────┤
│  Integration & Output Layer                                │
│  ├─ Database Connectors                                    │
│  ├─ Message Queue Publishers                               │
│  ├─ External API Clients                                   │
│  └─ Notification Services                                  │
└─────────────────────────────────────────────────────────────┘
```

### 插件类型支持
1. **YAML 配置插件**: 零代码声明式配置
2. **Python 插件**: 现有生态兼容 + 增强
3. **Go 原生插件**: 高性能关键路径处理
4. **JavaScript 插件**: 前端开发者友好

## 📋 功能需求

### 1. YAML 配置插件系统

#### 1.1 配置结构
```yaml
name: "github_webhook"
version: "1.0.0"
description: "GitHub webhook processor"

source:
  type: "github"
  events: ["push", "pull_request"]
  signature_validation:
    enabled: true
    secret_env: "GITHUB_WEBHOOK_SECRET"

pipeline:
  - name: "extract_data"
    type: "jq"
    expression: |
      {
        event: .action // "push",
        repository: .repository.name,
        author: .sender.login
      }
  
  - name: "filter_conditions"
    type: "conditional"
    conditions:
      - if: '.event == "push" and (.branch | contains("main"))'
        then: "continue"
      - else: "skip"

outputs:
  - name: "slack_notification"
    type: "webhook"
    url_env: "SLACK_WEBHOOK_URL"
    method: "POST"
```

#### 1.2 支持的处理器类型
- **jq**: JSON 数据提取和转换
- **conditional**: 条件分支处理
- **template**: Go template 模板渲染
- **script**: 内联 JavaScript 脚本
- **http**: HTTP 请求调用
- **database**: 数据库操作

#### 1.3 输出目标类型
- **webhook**: HTTP webhook 调用
- **database**: 数据库写入
- **message_queue**: 消息队列发布
- **file**: 文件写入
- **email**: 邮件发送
- **slack/teams**: 即时通讯通知

### 2. Go 原生插件系统

#### 2.1 插件接口定义
```go
type GoPlugin interface {
    Handle(ctx context.Context, data map[string]interface{}) (*Result, error)
    Metadata() *PluginMetadata
    Validate(config map[string]interface{}) error
}

type PluginMetadata struct {
    Name        string
    Version     string
    Description string
    Author      string
    SupportedEvents []string
}
```

#### 2.2 插件加载机制
- **编译时加载**: 插件编译为 .so 文件
- **运行时发现**: 自动扫描插件目录
- **热重载**: 支持插件热更新
- **版本管理**: 插件版本控制和回滚

#### 2.3 性能优化
- **并发处理**: 支持并发输出处理
- **连接池**: 数据库和 HTTP 连接复用
- **内存管理**: 智能内存回收
- **缓存机制**: 插件结果缓存

### 3. JavaScript 插件系统

#### 3.1 运行时环境
- **Node.js 集成**: 支持 npm 包生态
- **V8 引擎**: 高性能 JavaScript 执行
- **ES6+ 支持**: 现代 JavaScript 特性
- **TypeScript**: 可选类型支持

#### 3.2 插件 API
```javascript
class WebhookPlugin {
    async handle(data, context) {
        // 插件处理逻辑
        return {
            status: 'success',
            data: processedData
        };
    }
    
    async validate(config) {
        // 配置验证
        return { valid: true };
    }
}
```

#### 3.3 内置模块
- **fetch**: HTTP 请求客户端
- **crypto**: 加密和签名验证
- **lodash**: 数据处理工具
- **moment**: 时间处理
- **joi**: 数据验证

### 4. 可视化配置界面

#### 4.1 插件构建器
- **拖拽式界面**: 可视化流程构建
- **组件库**: 预定义处理器组件
- **实时预览**: 配置实时效果预览
- **代码生成**: 自动生成 YAML 配置

#### 4.2 插件管理
- **插件列表**: 所有插件的统一管理
- **状态监控**: 插件运行状态和性能指标
- **版本控制**: 插件版本管理和回滚
- **测试工具**: 插件功能测试和调试

#### 4.3 模板市场
- **模板库**: 常用 webhook 处理模板
- **分类浏览**: 按服务类型分类
- **一键部署**: 模板一键安装和配置
- **社区贡献**: 用户贡献模板机制

### 5. 企业级特性

#### 5.1 安全性
- **签名验证**: Webhook 签名自动验证
- **访问控制**: 基于角色的权限管理
- **审计日志**: 完整的操作审计记录
- **加密存储**: 敏感配置加密存储

#### 5.2 可观测性
- **指标监控**: Prometheus 指标导出
- **链路追踪**: 分布式链路追踪
- **日志聚合**: 结构化日志收集
- **告警机制**: 异常情况自动告警

#### 5.3 高可用性
- **负载均衡**: 多实例负载均衡
- **故障转移**: 自动故障检测和转移
- **数据备份**: 配置和数据自动备份
- **灾难恢复**: 快速恢复机制

## 🚀 实施计划

### Phase 1: 基础架构扩展 (4 周)
**目标**: 扩展现有架构支持多插件类型

**任务**:
- [ ] 扩展 gRPC 协议支持多插件类型
- [ ] 实现插件类型检测和路由
- [ ] 添加插件元数据管理
- [ ] 更新 Dashboard 基础框架

**交付物**:
- 新的 gRPC 协议定义
- 插件路由器实现
- 基础 Dashboard 更新

### Phase 2: YAML 配置插件引擎 (6 周)
**目标**: 实现零代码 YAML 配置插件

**任务**:
- [ ] YAML 配置解析器
- [ ] 数据处理管道引擎
- [ ] 条件分支和模板系统
- [ ] 输出目标连接器
- [ ] 配置验证和错误处理

**交付物**:
- YAML 插件引擎
- 处理器组件库
- 配置验证系统

### Phase 3: Go 原生插件支持 (5 周)
**目标**: 高性能 Go 插件系统

**任务**:
- [ ] Go 插件接口定义
- [ ] 插件编译和加载系统
- [ ] 热重载机制
- [ ] 性能优化和连接池
- [ ] 插件版本管理

**交付物**:
- Go 插件 SDK
- 插件加载器
- 性能优化框架

### Phase 4: JavaScript 运行时集成 (4 周)
**目标**: JavaScript 插件支持

**任务**:
- [ ] Node.js/V8 运行时集成
- [ ] JavaScript 插件 API
- [ ] npm 包管理集成
- [ ] TypeScript 支持
- [ ] 安全沙箱环境

**交付物**:
- JavaScript 运行时
- 插件 API 库
- 安全沙箱

### Phase 5: 可视化配置界面 (6 周)
**目标**: Dashboard 插件构建器

**任务**:
- [ ] 拖拽式流程构建器
- [ ] 组件库和模板系统
- [ ] 实时预览和测试
- [ ] 插件管理界面
- [ ] 用户体验优化

**交付物**:
- 可视化插件构建器
- 插件管理界面
- 用户指南和文档

### Phase 6: 企业级特性和插件市场 (4 周)
**目标**: 企业级功能和生态建设

**任务**:
- [ ] 插件模板市场
- [ ] 高级安全特性
- [ ] 监控和告警系统
- [ ] 高可用性部署
- [ ] 社区贡献机制

**交付物**:
- 插件市场平台
- 企业级安全框架
- 监控告警系统

## 📊 风险评估

### 技术风险
- **多语言集成复杂性**: 中等风险
  - 缓解措施: 分阶段实施，充分测试
- **性能影响**: 低风险
  - 缓解措施: 性能基准测试，优化关键路径
- **向后兼容性**: 低风险
  - 缓解措施: 保持现有 API 兼容

### 业务风险
- **用户学习成本**: 中等风险
  - 缓解措施: 详细文档，渐进式功能发布
- **生态建设时间**: 中等风险
  - 缓解措施: 提供丰富的初始模板库

### 运营风险
- **维护复杂性增加**: 中等风险
  - 缓解措施: 自动化测试，完善监控

## 📈 成功指标

### 技术指标
- **性能**: 高频 webhook 处理延迟 < 10ms
- **可用性**: 系统可用性 > 99.9%
- **扩展性**: 支持 10,000+ 并发 webhook

### 业务指标
- **用户采用率**: 3 个月内 80% 用户使用新功能
- **开发效率**: 插件开发时间减少 90%
- **生态丰富度**: 6 个月内模板库 50+ 个

### 用户体验指标
- **易用性**: 新用户 10 分钟内完成首个插件配置
- **满意度**: 用户满意度评分 > 4.5/5
- **文档完整性**: 文档覆盖率 > 95%

## 📚 参考资料

### 竞品分析
- **Zapier**: 低代码集成平台
- **Microsoft Power Automate**: 企业级自动化
- **GitHub Actions**: 基于 YAML 的工作流
- **GitLab CI/CD**: 声明式配置管道

### 技术参考
- **Go Plugin 系统**: Go 1.8+ plugin 包
- **gRPC 多语言支持**: Protocol Buffers
- **JavaScript V8 集成**: Node.js C++ Addons
- **YAML 处理**: gopkg.in/yaml.v3

## 💡 具体实现示例

### GitHub Webhook 低代码配置示例

#### YAML 配置文件
```yaml
# plugins/github_ci_cd.yaml
name: "github_ci_cd"
version: "1.0.0"
description: "GitHub CI/CD webhook processor"

source:
  type: "github"
  events: ["push", "pull_request"]
  signature_validation:
    enabled: true
    secret_env: "GITHUB_WEBHOOK_SECRET"
  filters:
    branches: ["main", "develop"]
    actions: ["opened", "synchronize", "closed"]

pipeline:
  - name: "extract_metadata"
    type: "jq"
    expression: |
      {
        event_type: .action // "push",
        repository: .repository.name,
        branch: .ref // .pull_request.head.ref,
        author: .sender.login,
        commit_sha: .after // .pull_request.head.sha,
        commit_message: .head_commit.message // .pull_request.title,
        pr_number: .pull_request.number // null
      }

  - name: "determine_environment"
    type: "conditional"
    conditions:
      - if: '.branch == "main" and .event_type == "push"'
        then:
          set: { environment: "production", deploy: true }
      - if: '.branch == "develop" and .event_type == "push"'
        then:
          set: { environment: "staging", deploy: true }
      - if: '.event_type == "pull_request" and .action == "opened"'
        then:
          set: { environment: "preview", deploy: true }
      - else:
        set: { deploy: false }

  - name: "skip_if_no_deploy"
    type: "conditional"
    conditions:
      - if: '.deploy == false'
        then: "skip"

  - name: "format_notification"
    type: "template"
    template: |
      {
        "text": "🚀 Deployment triggered for {{ .repository }}",
        "blocks": [
          {
            "type": "section",
            "text": {
              "type": "mrkdwn",
              "text": "*Repository:* {{ .repository }}\n*Branch:* {{ .branch }}\n*Environment:* {{ .environment }}\n*Author:* {{ .author }}"
            }
          },
          {
            "type": "section",
            "text": {
              "type": "mrkdwn",
              "text": "*Commit:* `{{ .commit_sha | slice 0 8 }}`\n*Message:* {{ .commit_message }}"
            }
          }
        ]
      }

outputs:
  - name: "trigger_deployment"
    type: "webhook"
    url_env: "DEPLOYMENT_WEBHOOK_URL"
    method: "POST"
    headers:
      Authorization: "Bearer ${DEPLOY_TOKEN}"
      Content-Type: "application/json"
    body_template: |
      {
        "repository": "{{ .repository }}",
        "branch": "{{ .branch }}",
        "commit_sha": "{{ .commit_sha }}",
        "environment": "{{ .environment }}"
      }

  - name: "slack_notification"
    type: "webhook"
    url_env: "SLACK_WEBHOOK_URL"
    method: "POST"
    headers:
      Content-Type: "application/json"
    condition: '.environment != "preview"'

  - name: "database_log"
    type: "database"
    connection_env: "DATABASE_URL"
    table: "deployment_logs"
    data_template: |
      {
        "repository": "{{ .repository }}",
        "branch": "{{ .branch }}",
        "commit_sha": "{{ .commit_sha }}",
        "environment": "{{ .environment }}",
        "author": "{{ .author }}",
        "triggered_at": "{{ now | date "2006-01-02T15:04:05Z07:00" }}"
      }
```

#### Dashboard 可视化配置界面
```typescript
// 插件构建器组件示例
interface PluginBuilderState {
  source: WebhookSource;
  pipeline: PipelineStep[];
  outputs: OutputDestination[];
}

const GitHubPluginBuilder: React.FC = () => {
  const [config, setConfig] = useState<PluginBuilderState>({
    source: {
      type: 'github',
      events: ['push', 'pull_request'],
      signatureValidation: true
    },
    pipeline: [],
    outputs: []
  });

  return (
    <div className="plugin-builder">
      <SourceConfigPanel
        value={config.source}
        onChange={(source) => setConfig({...config, source})}
      />

      <PipelineBuilder
        steps={config.pipeline}
        onStepsChange={(pipeline) => setConfig({...config, pipeline})}
      />

      <OutputBuilder
        outputs={config.outputs}
        onChange={(outputs) => setConfig({...config, outputs})}
      />

      <PreviewPanel config={config} />
    </div>
  );
};
```

### Go 原生插件示例

```go
// plugins/go/high_performance_processor.go
package main

import (
    "context"
    "encoding/json"
    "fmt"
    "sync"
    "time"

    "github.com/loonghao/webhook_bridge/pkg/plugin"
)

type HighPerformanceProcessor struct {
    plugin.BaseGoPlugin

    // 连接池
    httpClient *http.Client
    dbPool     *sql.DB

    // 缓存
    cache sync.Map

    // 指标
    processedCount int64
    errorCount     int64
}

func (p *HighPerformanceProcessor) Handle(ctx context.Context, data map[string]interface{}) (*plugin.Result, error) {
    start := time.Now()
    defer func() {
        atomic.AddInt64(&p.processedCount, 1)
        // 记录处理时间指标
        plugin.RecordProcessingTime(time.Since(start))
    }()

    // 并发处理多个输出
    var wg sync.WaitGroup
    results := make(chan error, 3)

    // 异步发送通知
    wg.Add(1)
    go func() {
        defer wg.Done()
        results <- p.sendNotification(ctx, data)
    }()

    // 异步写入数据库
    wg.Add(1)
    go func() {
        defer wg.Done()
        results <- p.saveToDatabase(ctx, data)
    }()

    // 异步发布到消息队列
    wg.Add(1)
    go func() {
        defer wg.Done()
        results <- p.publishToQueue(ctx, data)
    }()

    // 等待所有操作完成
    wg.Wait()
    close(results)

    // 检查错误
    var errors []string
    for err := range results {
        if err != nil {
            errors = append(errors, err.Error())
            atomic.AddInt64(&p.errorCount, 1)
        }
    }

    if len(errors) > 0 {
        return &plugin.Result{
            Status: "partial_success",
            Data: map[string]interface{}{
                "errors": errors,
                "processed_outputs": 3 - len(errors),
            },
        }, nil
    }

    return &plugin.Result{
        Status: "success",
        Data: map[string]interface{}{
            "processed_at": time.Now().Unix(),
            "processing_time_ms": time.Since(start).Milliseconds(),
        },
    }, nil
}

func (p *HighPerformanceProcessor) Metadata() *plugin.PluginMetadata {
    return &plugin.PluginMetadata{
        Name:        "high_performance_processor",
        Version:     "1.0.0",
        Description: "High-performance webhook processor with concurrent outputs",
        Author:      "Webhook Bridge Team",
        SupportedEvents: []string{"github.push", "github.pull_request"},
        Capabilities: []string{"concurrent", "metrics", "caching"},
    }
}

// 插件入口点
func NewPlugin() plugin.GoPlugin {
    return &HighPerformanceProcessor{
        httpClient: &http.Client{Timeout: 10 * time.Second},
        // 初始化数据库连接池等
    }
}
```

### JavaScript 插件示例

```javascript
// plugins/js/advanced_github_processor.js
const { BaseJSPlugin } = require('@webhook-bridge/plugin-js');
const { Octokit } = require('@octokit/rest');
const axios = require('axios');

class AdvancedGitHubProcessor extends BaseJSPlugin {
    constructor() {
        super();
        this.octokit = new Octokit({
            auth: process.env.GITHUB_TOKEN
        });
    }

    async handle(data) {
        const { action, repository, pull_request, sender } = data;

        try {
            // 根据事件类型处理
            switch (action) {
                case 'opened':
                    return await this.handlePROpened(pull_request, repository);
                case 'closed':
                    return await this.handlePRClosed(pull_request, repository);
                case 'synchronize':
                    return await this.handlePRUpdated(pull_request, repository);
                default:
                    return { status: 'skipped', reason: 'unsupported action' };
            }
        } catch (error) {
            this.logger.error('Plugin execution failed:', error);
            return {
                status: 'error',
                error: error.message
            };
        }
    }

    async handlePROpened(pr, repo) {
        // 自动添加标签
        await this.addLabels(repo, pr.number);

        // 触发 CI 检查
        await this.triggerCIChecks(repo, pr);

        // 发送团队通知
        await this.notifyTeam(repo, pr, 'opened');

        return {
            status: 'success',
            data: {
                action: 'pr_opened_processed',
                pr_number: pr.number,
                labels_added: true,
                ci_triggered: true,
                team_notified: true
            }
        };
    }

    async addLabels(repo, prNumber) {
        const labels = ['needs-review', 'auto-processed'];

        await this.octokit.issues.addLabels({
            owner: repo.owner.login,
            repo: repo.name,
            issue_number: prNumber,
            labels: labels
        });
    }

    async triggerCIChecks(repo, pr) {
        // 调用外部 CI 系统
        const response = await axios.post(process.env.CI_WEBHOOK_URL, {
            repository: repo.full_name,
            branch: pr.head.ref,
            commit_sha: pr.head.sha,
            pr_number: pr.number
        }, {
            headers: {
                'Authorization': `Bearer ${process.env.CI_TOKEN}`,
                'Content-Type': 'application/json'
            }
        });

        return response.status === 200;
    }

    async notifyTeam(repo, pr, action) {
        const message = {
            text: `🔔 PR ${action}: ${pr.title}`,
            attachments: [{
                color: 'good',
                fields: [
                    { title: 'Repository', value: repo.full_name, short: true },
                    { title: 'Author', value: pr.user.login, short: true },
                    { title: 'Branch', value: pr.head.ref, short: true },
                    { title: 'PR Number', value: `#${pr.number}`, short: true }
                ],
                actions: [{
                    type: 'button',
                    text: 'View PR',
                    url: pr.html_url
                }]
            }]
        };

        await axios.post(process.env.SLACK_WEBHOOK_URL, message);
    }

    // 插件配置验证
    async validate(config) {
        const required = ['GITHUB_TOKEN', 'CI_WEBHOOK_URL', 'SLACK_WEBHOOK_URL'];
        const missing = required.filter(key => !process.env[key]);

        if (missing.length > 0) {
            return {
                valid: false,
                errors: [`Missing required environment variables: ${missing.join(', ')}`]
            };
        }

        return { valid: true };
    }
}

module.exports = AdvancedGitHubProcessor;
```

## 🔧 技术实现细节

### gRPC 协议扩展

```protobuf
// api/proto/webhook_v3.proto
syntax = "proto3";

package webhook.v3;

service WebhookExecutorV3 {
    // 执行插件
    rpc ExecutePlugin(ExecutePluginRequest) returns (ExecutePluginResponse);

    // 验证插件配置
    rpc ValidatePlugin(ValidatePluginRequest) returns (ValidatePluginResponse);

    // 获取插件元数据
    rpc GetPluginMetadata(GetPluginMetadataRequest) returns (GetPluginMetadataResponse);

    // 插件生命周期管理
    rpc LoadPlugin(LoadPluginRequest) returns (LoadPluginResponse);
    rpc UnloadPlugin(UnloadPluginRequest) returns (UnloadPluginResponse);
    rpc ReloadPlugin(ReloadPluginRequest) returns (ReloadPluginResponse);

    // 插件性能监控
    rpc GetPluginMetrics(GetPluginMetricsRequest) returns (GetPluginMetricsResponse);
}

message ExecutePluginRequest {
    string plugin_name = 1;
    PluginType plugin_type = 2;
    string http_method = 3;
    map<string, string> data = 4;
    map<string, string> headers = 5;
    WebhookContext context = 6;
}

message WebhookContext {
    string source_type = 1;      // github, gitlab, slack
    string event_type = 2;       // push, pull_request, issue
    string signature = 3;        // webhook signature
    int64 timestamp = 4;         // event timestamp
    string request_id = 5;       // unique request ID
    map<string, string> metadata = 6;  // additional context
}

enum PluginType {
    PYTHON = 0;
    YAML_CONFIG = 1;
    GO_NATIVE = 2;
    JAVASCRIPT = 3;
}

message PluginMetadata {
    string name = 1;
    string version = 2;
    string description = 3;
    string author = 4;
    repeated string supported_events = 5;
    repeated string capabilities = 6;
    map<string, string> config_schema = 7;
}
```

### 插件市场数据结构

```yaml
# plugin_market.yaml
plugins:
  - id: "github-ci-cd"
    name: "GitHub CI/CD Integration"
    description: "Complete CI/CD pipeline integration for GitHub webhooks"
    version: "1.2.0"
    author: "Webhook Bridge Team"
    category: "ci-cd"
    tags: ["github", "deployment", "automation"]
    type: "yaml_config"
    downloads: 1250
    rating: 4.8

    # 插件配置模板
    template: |
      name: "github_ci_cd"
      source:
        type: "github"
        events: ["push", "pull_request"]
      pipeline:
        - name: "extract_data"
          type: "jq"
          expression: "{ repo: .repository.name, branch: .ref }"
      outputs:
        - name: "deploy"
          type: "webhook"
          url_env: "DEPLOY_WEBHOOK_URL"

    # 必需的环境变量
    required_env:
      - name: "GITHUB_WEBHOOK_SECRET"
        description: "GitHub webhook secret for signature validation"
      - name: "DEPLOY_WEBHOOK_URL"
        description: "Deployment service webhook URL"

    # 使用示例
    examples:
      - name: "Basic Setup"
        description: "Simple push-to-deploy configuration"
        config: |
          # Basic configuration example
          source:
            events: ["push"]
            filters:
              branches: ["main"]

      - name: "Advanced Setup"
        description: "Multi-environment deployment"
        config: |
          # Advanced configuration with multiple environments
          pipeline:
            - name: "determine_env"
              type: "conditional"
              conditions:
                - if: '.branch == "main"'
                  then: { environment: "production" }

  - id: "slack-notifications"
    name: "Slack Notifications"
    description: "Rich Slack notifications for webhook events"
    version: "1.0.5"
    author: "Community"
    category: "notifications"
    tags: ["slack", "notifications", "messaging"]
    type: "javascript"
    downloads: 890
    rating: 4.6

    # JavaScript 插件包
    package: "@webhook-bridge/slack-notifications"
    npm_version: "1.0.5"

    # 插件特性
    features:
      - "Rich message formatting"
      - "Custom emoji and colors"
      - "Thread replies support"
      - "User mentions"
      - "Interactive buttons"

    # 配置示例
    config_example: |
      {
        "webhook_url": "${SLACK_WEBHOOK_URL}",
        "channel": "#deployments",
        "username": "Webhook Bridge",
        "icon_emoji": ":rocket:",
        "template": "custom"
      }
```

---

**文档版本**: v1.0
**创建日期**: 2024-01-01
**最后更新**: 2024-01-01
**负责人**: Webhook Bridge 团队
