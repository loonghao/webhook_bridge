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

// Dashboard Stats
export interface DashboardStats {
  totalRequests: number
  activePlugins: number
  workers: number
  uptime: string
  requestsGrowth: string
  pluginsGrowth: string
  workersStatus: string
  uptimePercentage: string
}

// System Status
export interface SystemStatus {
  service: string
  status: string
  version: string
  timestamp: string
  uptime: string
  build: string
  checks: {
    grpc: {
      status: boolean
      message: string
    }
    database: {
      status: boolean
      message: string
    }
    storage: {
      status: boolean
      message: string
    }
  }
}

// Plugin Info
export interface PluginInfo {
  name: string
  version: string
  status: 'active' | 'inactive' | 'error'
  description: string
  lastExecuted?: string
  executionCount: number
  errorCount?: number
  avgExecutionTime?: string
  path?: string
  supportedMethods?: string[]
  isAvailable?: boolean
  lastModified?: string
}

// Plugin Execution Request
export interface PluginExecutionRequest {
  plugin: string
  method: 'GET' | 'POST' | 'PUT' | 'DELETE'
  data: Record<string, any>
}

// Plugin Execution Result
export interface PluginExecutionResult {
  success: boolean
  data?: any
  error?: string
  executionTime?: number
  timestamp: string
}

// Plugin Stats
export interface PluginStats {
  plugin: string
  method: string
  count: number
  errors: number
  lastExecution: string
  avgTime: string
}

// Worker Info
export interface WorkerInfo {
  id: string
  status: 'active' | 'idle' | 'busy'
  currentJob?: string
  totalJobs: number
  completedJobs: number
  failedJobs: number
}

// Log Entry
export interface LogEntry {
  timestamp: string
  level: 'info' | 'warn' | 'error' | 'debug'
  message: string
  source?: string
  plugin?: string
  details?: any
}

// Activity Entry
export interface ActivityEntry {
  id: string
  type: 'webhook' | 'plugin' | 'system'
  message: string
  timestamp: string
  status: 'success' | 'error' | 'warning'
  source?: string
}
