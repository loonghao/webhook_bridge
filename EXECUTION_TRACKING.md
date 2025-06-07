# 🔍 Execution Tracking System

Webhook Bridge现在包含了一个强大的执行追踪系统，可以记录、分析和监控所有插件的执行历史。

## ✨ 功能特性

### 📊 **执行历史记录**
- 记录每次插件执行的完整生命周期
- 包含输入数据、输出结果、错误信息
- 支持执行时间、重试次数、状态追踪
- 自动分类错误类型（超时、连接错误、权限等）

### 🗄️ **持久化存储**
- **SQLite** (默认): 嵌入式数据库，零配置
- **MySQL** (计划中): 高性能关系型数据库
- **PostgreSQL** (计划中): 企业级数据库支持

### 📈 **实时监控**
- 插件执行统计和性能指标
- 成功率、平均执行时间分析
- 错误趋势和模式识别
- 实时内存指标缓存

### 🧹 **自动维护**
- 可配置的数据保留策略
- 自动清理过期记录
- 后台清理任务
- 数据库健康检查

## 🚀 快速开始

### 1. 配置启用

在 `config.yaml` 中启用执行追踪：

```yaml
# 存储配置
storage:
  type: "sqlite"
  sqlite:
    database_path: "data/executions.db"
    max_connections: 10
    retention_days: 30
    enable_wal: true
    enable_foreign_keys: true

# 执行追踪配置
execution_tracking:
  enabled: true
  track_input: true
  track_output: true
  track_errors: true
  max_input_size: 1048576   # 1MB
  max_output_size: 1048576  # 1MB
  cleanup_interval: "24h"
  metrics_aggregation_interval: "1h"
```

### 2. 启动服务

```bash
./webhook-bridge.exe
```

启动日志会显示：
```
✅ Execution storage initialized successfully
✅ Execution cleanup worker started
```

### 3. API端点

执行追踪提供了以下REST API端点：

#### 📋 获取执行历史
```http
GET /api/v1/executions?limit=50&offset=0&plugin=my_plugin&status=completed
```

#### 📊 获取执行统计
```http
GET /api/v1/executions/stats?days=7&plugin=my_plugin
```

#### 🔍 获取特定执行详情
```http
GET /api/v1/executions/{execution_id}
```

#### 🗄️ 获取存储信息
```http
GET /api/v1/executions/storage/info
```

#### 🧹 清理旧记录
```http
DELETE /api/v1/executions/cleanup
```

## 📊 数据结构

### ExecutionRecord
```json
{
  "id": "uuid",
  "plugin_name": "my_plugin",
  "http_method": "POST",
  "start_time": "2025-06-06T14:33:48Z",
  "end_time": "2025-06-06T14:33:49Z",
  "status": "completed",
  "input": "{\"key\": \"value\"}",
  "output": "{\"result\": \"success\"}",
  "error": "",
  "error_type": "",
  "duration": 1000000000,
  "attempts": 1,
  "retry_count": 0,
  "trace_id": "trace-uuid",
  "user_agent": "Mozilla/5.0...",
  "remote_ip": "192.168.1.100",
  "tags": "{\"job_id\": \"job-123\"}",
  "metadata": "{\"priority\": 1}",
  "created_at": "2025-06-06T14:33:48Z",
  "updated_at": "2025-06-06T14:33:49Z"
}
```

### ExecutionStats
```json
{
  "total_executions": 1000,
  "successful_executions": 950,
  "failed_executions": 50,
  "timeout_executions": 5,
  "success_rate": 95.0,
  "avg_duration": 500000000,
  "min_duration": 100000000,
  "max_duration": 2000000000,
  "unique_plugins": 10,
  "daily_stats": [...]
}
```

## 🔧 高级配置

### 存储优化
```yaml
storage:
  sqlite:
    enable_wal: true          # 启用WAL模式提升并发性能
    enable_foreign_keys: true # 启用外键约束
    max_connections: 20       # 增加连接池大小
```

### 数据保留策略
```yaml
execution_tracking:
  cleanup_interval: "12h"    # 更频繁的清理
  retention_days: 90         # 保留90天数据
```

### 性能调优
```yaml
execution_tracking:
  max_input_size: 2097152    # 2MB输入数据限制
  max_output_size: 2097152   # 2MB输出数据限制
  track_input: false         # 禁用输入追踪以节省空间
```

## 🛠️ 故障排除

### 常见问题

1. **数据库连接失败**
   ```
   ⚠️ Failed to initialize storage: failed to open database
   ```
   - 检查数据目录权限
   - 确保磁盘空间充足

2. **执行追踪被禁用**
   ```
   🔄 Execution tracking will be disabled
   ```
   - 检查配置文件中的 `execution_tracking.enabled`
   - 验证存储配置是否正确

3. **性能问题**
   - 调整 `max_input_size` 和 `max_output_size`
   - 减少 `retention_days`
   - 增加 `cleanup_interval` 频率

### 数据库维护

查看数据库状态：
```bash
curl http://localhost:8001/api/v1/executions/storage/info
```

手动清理：
```bash
curl -X DELETE http://localhost:8001/api/v1/executions/cleanup
```

## 🔮 未来计划

- [ ] **MySQL支持**: 企业级数据库支持
- [ ] **PostgreSQL支持**: 高级查询和分析功能
- [ ] **数据导出**: CSV/JSON格式导出
- [ ] **可视化Dashboard**: 图表和趋势分析
- [ ] **告警系统**: 基于阈值的自动告警
- [ ] **性能分析**: 插件性能瓶颈识别

## 📝 示例用法

查看完整的测试示例，请参考 `test_execution_tracking.py` 文件。

---

**注意**: 执行追踪系统设计为高性能、低侵入性。即使在追踪失败的情况下，也不会影响插件的正常执行。
