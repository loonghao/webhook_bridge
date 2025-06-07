// Unified API Response Types
export interface ApiResponse<T = any> {
  success: boolean
  data?: T
  error?: {
    code: string
    message: string
    details?: string
  }
  message?: string
  timestamp: string
  request_id?: string
}

export interface PaginatedResponse<T = any> extends ApiResponse<T> {
  pagination?: {
    page: number
    page_size: number
    total: number
    total_pages: number
  }
}

// Legacy API Response (for backward compatibility)
export interface LegacyApiResponse<T = any> {
  success: boolean
  data?: T
  error?: string
  message?: string
}

// Dashboard Stats (matches backend getStats response)
export interface DashboardStats {
  total_requests: number
  successful_requests: number
  failed_requests: number
  average_response_time: number
  active_connections: number
  plugin_count: number
  error_rate: number
  uptime?: string
  plugin_stats?: Record<string, any>

  // Computed fields for UI
  totalRequests?: number
  activePlugins?: number
  workers?: number
  requestsGrowth?: string
  pluginsGrowth?: string
  workersStatus?: string
  uptimePercentage?: string
}

// System Status (matches backend getStatus response)
export interface SystemStatus {
  server_status: string
  grpc_connected: boolean
  worker_count: number
  active_workers: number
  total_jobs: number
  completed_jobs: number
  failed_jobs: number
  uptime: string

  // Computed fields for UI compatibility
  service?: string
  status?: 'healthy' | 'unhealthy' | 'unknown'
  version?: string
  goVersion?: string
  pythonVersion?: string
  memory?: {
    used: number
    total: number
    percentage: number
  }
  cpu?: {
    usage: number
    cores: number
  }
  disk?: {
    used: number
    total: number
    percentage: number
  }
}

// Plugin Information (matches backend/gRPC response)
export interface PluginInfo {
  name: string
  path: string
  description: string
  supported_methods: string[]
  is_available: boolean
  last_modified?: string

  // Computed fields for UI compatibility
  id?: string
  version?: string
  status?: 'active' | 'inactive' | 'error' | 'loading'
  author?: string
  lastExecuted?: string
  executionCount?: number
  successRate?: number
  avgExecutionTime?: number | string
  errorCount?: number
  config?: Record<string, any>
  dependencies?: string[]
  error?: string
  enabled?: boolean
  type?: 'python' | 'go' | 'javascript' | 'yaml'
  size?: number
}

// Worker Information
export interface WorkerInfo {
  id: string
  status: 'idle' | 'busy' | 'error' | 'stopped'
  currentJob?: string
  completedJobs: number
  totalJobs: number
  failedJobs?: number
  uptime: string
  lastActivity?: string
  performance?: {
    avgExecutionTime: number
    successRate: number
    errorCount: number
  }
}

// Log Entry
export interface LogEntry {
  id: string
  timestamp: string
  level: 'debug' | 'info' | 'warn' | 'error' | 'fatal'
  message: string
  source?: string
  component?: string
  plugin?: string
  worker?: string
  metadata?: Record<string, any>
  stackTrace?: string
}

// Activity Entry
export interface ActivityEntry {
  id: string
  timestamp: string
  type: 'plugin_execution' | 'webhook_received' | 'system_event' | 'user_action'
  title: string
  description?: string
  status: 'success' | 'error' | 'warning' | 'info'
  duration?: number
  plugin?: string
  worker?: string
  metadata?: Record<string, any>
}

// Configuration Types
export interface ConfigSection {
  name: string
  description?: string
  fields: ConfigField[]
}

export interface ConfigField {
  key: string
  label: string
  type: 'string' | 'number' | 'boolean' | 'select' | 'textarea' | 'password'
  value: any
  defaultValue?: any
  description?: string
  required?: boolean
  options?: { label: string; value: any }[]
  validation?: {
    min?: number
    max?: number
    pattern?: string
    message?: string
  }
}

// Python Environment
export interface PythonEnvironment {
  id: string
  name: string
  path: string
  version: string
  status: 'available' | 'unavailable' | 'error'
  isDefault: boolean
  packages?: PythonPackage[]
  lastChecked?: string
  error?: string
}

export interface PythonPackage {
  name: string
  version: string
  description?: string
  required: boolean
  status: 'installed' | 'missing' | 'outdated'
}

// Connection Status
export interface ConnectionInfo {
  name: string
  type: 'database' | 'api' | 'service' | 'websocket'
  status: 'connected' | 'disconnected' | 'error' | 'connecting'
  url?: string
  lastConnected?: string
  latency?: number
  error?: string
  metadata?: Record<string, any>
}

// API Test
export interface ApiTestRequest {
  method: 'GET' | 'POST' | 'PUT' | 'DELETE' | 'PATCH'
  url: string
  headers?: Record<string, string>
  body?: string
  timeout?: number
}

export interface ApiTestResponse {
  status: number
  statusText: string
  headers: Record<string, string>
  body: string
  duration: number
  size: number
  timestamp: string
}

// WebSocket Types
export interface MonitorMessage {
  type: string
  timestamp: string
  data: any
}

export interface PluginStatusUpdate {
  plugin_name: string
  status: string
  last_executed?: string
  execution_time?: number
  success: boolean
  error?: string
}

export interface SystemMetricsUpdate {
  total_executions: number
  success_rate: number
  avg_execution_time: number
  active_plugins: number
  error_rate: number
  last_hour_executions: number
}

// Chart Data Types
export interface ChartDataPoint {
  timestamp: string
  value: number
  label?: string
}

export interface MetricsData {
  requests: ChartDataPoint[]
  errors: ChartDataPoint[]
  latency: ChartDataPoint[]
  plugins: ChartDataPoint[]
}
