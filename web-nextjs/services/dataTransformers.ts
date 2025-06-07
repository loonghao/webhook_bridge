import { DashboardStats, SystemStatus, PluginInfo, WorkerInfo, LogEntry } from '@/types/api'

/**
 * Transform backend stats response to frontend format
 */
export function transformDashboardStats(backendStats: any): DashboardStats {
  // Handle both snake_case and camelCase formats from backend
  const totalRequests = backendStats.totalRequests || backendStats.total_requests || 0
  const activePlugins = backendStats.activePlugins || backendStats.plugin_count || 0
  const workers = backendStats.workers || backendStats.active_connections || 0
  const uptime = backendStats.uptime || 'Unknown'

  return {
    // Backend fields (snake_case for API compatibility)
    total_requests: totalRequests,
    successful_requests: backendStats.successful_requests || 0,
    failed_requests: backendStats.failed_requests || 0,
    average_response_time: backendStats.average_response_time || 0,
    active_connections: workers,
    plugin_count: activePlugins,
    error_rate: backendStats.error_rate || 0,
    uptime: uptime,
    plugin_stats: backendStats.plugin_stats || {},

    // Computed fields for UI compatibility (camelCase)
    totalRequests: totalRequests,
    activePlugins: activePlugins,
    workers: workers,
    requestsGrowth: backendStats.requestsGrowth || calculateGrowth(totalRequests, backendStats.previous_requests),
    pluginsGrowth: backendStats.pluginsGrowth || `${activePlugins} plugins`,
    workersStatus: backendStats.workersStatus || (workers > 0 ? 'Active' : 'Idle'),
    uptimePercentage: backendStats.uptimePercentage || calculateUptimePercentage(backendStats.error_rate || 0),
  }
}

/**
 * Transform backend system status to frontend format
 */
export function transformSystemStatus(backendStatus: any): SystemStatus {
  // Handle the actual backend response format
  const status = backendStatus.status || 'unknown'
  const isHealthy = status === 'healthy'
  const checks = backendStatus.checks || {}

  return {
    // Backend fields (adapt to actual response format)
    server_status: status,
    grpc_connected: checks.grpc?.status || false,
    worker_count: backendStatus.worker_count || 0,
    active_workers: backendStatus.active_workers || 0,
    total_jobs: backendStatus.total_jobs || 0,
    completed_jobs: backendStatus.completed_jobs || 0,
    failed_jobs: backendStatus.failed_jobs || 0,
    uptime: backendStatus.uptime || 'Unknown',

    // Computed fields for UI compatibility
    service: backendStatus.service || 'Webhook Bridge',
    status: status,
    version: backendStatus.version || '2.0.0-hybrid',
    goVersion: 'Go 1.21+',
    pythonVersion: checks.grpc?.status ? 'Python 3.8+' : undefined,
  }
}

/**
 * Transform backend plugin info to frontend format
 */
export function transformPluginInfo(backendPlugin: any): PluginInfo {
  // Handle the actual backend response format
  const status = backendPlugin.status || (backendPlugin.is_available ? 'active' : 'inactive')

  return {
    // Backend fields (adapt to actual response)
    name: backendPlugin.name || 'Unknown',
    path: backendPlugin.path || '',
    description: backendPlugin.description || 'No description available',
    supported_methods: backendPlugin.supported_methods || [],
    is_available: status === 'active',
    last_modified: backendPlugin.last_modified || backendPlugin.lastExecuted || '',

    // Computed fields for UI compatibility
    id: backendPlugin.name || generatePluginId(backendPlugin.path || ''),
    version: backendPlugin.version || '1.0.0',
    status: status as 'active' | 'inactive' | 'error' | 'loading',
    enabled: status === 'active',
    type: detectPluginType(backendPlugin.path || ''),
    executionCount: backendPlugin.executionCount || 0,
    successRate: 100, // Default success rate
    avgExecutionTime: backendPlugin.avgExecutionTime || '0ms',
    lastExecuted: backendPlugin.lastExecuted,
    errorCount: backendPlugin.errorCount || 0,
  }
}

/**
 * Transform backend worker info to frontend format
 */
export function transformWorkerInfo(backendWorker: any): WorkerInfo {
  return {
    id: backendWorker.id || generateWorkerId(),
    status: mapWorkerStatus(backendWorker.status),
    currentJob: backendWorker.currentJob || backendWorker.current_job,
    completedJobs: backendWorker.completedJobs || backendWorker.completed_jobs || 0,
    totalJobs: backendWorker.totalJobs || backendWorker.total_jobs || 0,
    failedJobs: backendWorker.failedJobs || backendWorker.failed_jobs || 0,
    uptime: backendWorker.uptime || 'Unknown',
    lastActivity: backendWorker.lastActivity || backendWorker.last_activity,
    performance: backendWorker.performance ? {
      avgExecutionTime: backendWorker.performance.avg_execution_time || 0,
      successRate: backendWorker.performance.success_rate || 100,
      errorCount: backendWorker.performance.error_count || 0,
    } : undefined,
  }
}

/**
 * Transform backend log entry to frontend format
 */
export function transformLogEntry(backendLog: any): LogEntry {
  return {
    id: backendLog.id || generateLogId(),
    timestamp: backendLog.timestamp || new Date().toISOString(),
    level: mapLogLevel(backendLog.level),
    message: backendLog.message || '',
    source: backendLog.source || backendLog.component,
    component: backendLog.component,
    plugin: backendLog.plugin || backendLog.plugin_name,
    worker: backendLog.worker,
    metadata: backendLog.metadata || backendLog.details,
    stackTrace: backendLog.stack_trace,
  }
}

// Helper functions

function calculateGrowth(current: number, previous?: number): string {
  if (!previous || previous === 0) return '+0%'
  const growth = ((current - previous) / previous) * 100
  return `${growth > 0 ? '+' : ''}${growth.toFixed(1)}%`
}

function calculateUptimePercentage(errorRate: number): string {
  const uptime = Math.max(0, 100 - (errorRate * 100))
  return `${uptime.toFixed(1)}%`
}

function generatePluginId(path: string): string {
  return path.split('/').pop()?.replace(/\.[^/.]+$/, '') || 'unknown'
}

function detectPluginType(path: string): 'python' | 'go' | 'javascript' | 'yaml' {
  const ext = path.split('.').pop()?.toLowerCase()
  switch (ext) {
    case 'py': return 'python'
    case 'go': return 'go'
    case 'js': case 'ts': return 'javascript'
    case 'yaml': case 'yml': return 'yaml'
    default: return 'python' // Default to python
  }
}

function mapWorkerStatus(status: string): 'idle' | 'busy' | 'error' | 'stopped' {
  switch (status?.toLowerCase()) {
    case 'active': case 'busy': return 'busy'
    case 'idle': return 'idle'
    case 'error': case 'failed': return 'error'
    case 'stopped': case 'inactive': return 'stopped'
    default: return 'idle'
  }
}

function mapLogLevel(level: string): 'debug' | 'info' | 'warn' | 'error' | 'fatal' {
  switch (level?.toLowerCase()) {
    case 'debug': return 'debug'
    case 'info': case 'information': return 'info'
    case 'warn': case 'warning': return 'warn'
    case 'error': return 'error'
    case 'fatal': case 'critical': return 'fatal'
    default: return 'info'
  }
}

function generateWorkerId(): string {
  return `worker-${Math.random().toString(36).substr(2, 9)}`
}

function generateLogId(): string {
  return `log-${Date.now()}-${Math.random().toString(36).substr(2, 5)}`
}

/**
 * Transform array of backend items using the appropriate transformer
 */
export function transformArray<T, U>(
  items: T[], 
  transformer: (item: T) => U
): U[] {
  return items.map(transformer)
}

/**
 * Create mock activity entries from various data sources
 */
export function createActivityEntries(
  plugins: PluginInfo[],
  logs: LogEntry[],
  workers: WorkerInfo[]
): any[] {
  const activities: any[] = []
  
  // Add recent plugin activities
  plugins.forEach(plugin => {
    if (plugin.lastExecuted) {
      activities.push({
        id: `activity-plugin-${plugin.id}`,
        timestamp: plugin.lastExecuted,
        type: 'plugin_execution',
        title: `Plugin ${plugin.name} executed`,
        status: plugin.status === 'active' ? 'success' : 'error',
        plugin: plugin.name,
      })
    }
  })
  
  // Add recent error logs as activities
  logs.filter(log => log.level === 'error').slice(0, 5).forEach(log => {
    activities.push({
      id: `activity-error-${log.id}`,
      timestamp: log.timestamp,
      type: 'system_event',
      title: 'System Error',
      description: log.message,
      status: 'error',
      component: log.source,
    })
  })
  
  // Sort by timestamp (newest first)
  return activities.sort((a, b) => 
    new Date(b.timestamp).getTime() - new Date(a.timestamp).getTime()
  )
}
