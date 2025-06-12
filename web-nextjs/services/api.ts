import {
  ApiResponse,
  DashboardStats,
  SystemStatus,
  PluginInfo,
  WorkerInfo,
  LogEntry,
  ActivityEntry,
  ConfigSection,
  PythonEnvironment,
  ConnectionInfo,
  ApiTestRequest,
  ApiTestResponse
} from '@/types/api'
import {
  transformDashboardStats,
  transformSystemStatus,
  transformPluginInfo,
  transformWorkerInfo,
  transformLogEntry,
  transformArray,
  createActivityEntries
} from './dataTransformers'

// API base URL configuration with dynamic port detection
const getApiBase = () => {
  if (typeof window !== 'undefined') {
    // Client-side: check for runtime configuration first
    const runtimeConfig = (window as any).__WEBHOOK_BRIDGE_CONFIG__
    if (runtimeConfig?.apiBaseUrl) {
      return `${runtimeConfig.apiBaseUrl}/api/dashboard`
    }

    // Use environment variable if available
    if (process.env.NEXT_PUBLIC_API_BASE_URL) {
      return `${process.env.NEXT_PUBLIC_API_BASE_URL}/api/dashboard`
    }

    // For embedded deployment: use same origin (same port as frontend)
    if (window.location.pathname.startsWith('/')) {
      return '/api/dashboard'
    }

    // Fallback to default backend port for development
    const protocol = window.location.protocol
    const hostname = window.location.hostname
    const backendPort = '8080' // Backend server port
    return `${protocol}//${hostname}:${backendPort}/api/dashboard`
  }
  // Server-side: always use relative URL
  return '/api/dashboard'
}

const API_BASE = getApiBase()

class ApiClient {
  private retryConfig = {
    maxRetries: 3,
    baseDelay: 1000,
    maxDelay: 5000
  }

  private async request<T>(endpoint: string, options?: RequestInit): Promise<T> {
    const url = `${API_BASE}${endpoint}`
    const startTime = Date.now()

    // Log request start for stagewise
    if (typeof window !== 'undefined' && (window as any).__stagewise_log) {
      (window as any).__stagewise_log('api_request_start', {
        url,
        method: options?.method || 'GET',
        timestamp: new Date().toISOString()
      })
    }

    try {
      const response = await this.requestWithRetry(url, {
        headers: {
          'Content-Type': 'application/json',
          'X-Request-ID': this.generateRequestId(),
          ...options?.headers,
        },
        ...options,
      })

      if (!response.ok) {
        const errorText = await response.text()
        let errorData
        try {
          errorData = JSON.parse(errorText)
        } catch {
          errorData = { message: errorText }
        }

        const error = new Error(`HTTP ${response.status}: ${errorData.message || response.statusText}`)
        ;(error as any).status = response.status
        ;(error as any).data = errorData
        throw error
      }

      const data = await response.json()
      const duration = Date.now() - startTime

      // Log successful request for stagewise
      if (typeof window !== 'undefined' && (window as any).__stagewise_log) {
        (window as any).__stagewise_log('api_request_success', {
          url,
          method: options?.method || 'GET',
          status: response.status,
          duration,
          timestamp: new Date().toISOString()
        })
      }

      // Handle backend API response format
      if (data && typeof data === 'object' && 'success' in data) {
        if (!data.success) {
          throw new Error(data.message || data.error || 'API request failed')
        }
        return data.data !== undefined ? data.data : data
      }

      return data
    } catch (error) {
      const duration = Date.now() - startTime

      // Log failed request for stagewise
      if (typeof window !== 'undefined' && (window as any).__stagewise_log) {
        (window as any).__stagewise_log('api_request_error', {
          url,
          method: options?.method || 'GET',
          error: error instanceof Error ? error.message : String(error),
          duration,
          timestamp: new Date().toISOString()
        })
      }

      console.error(`API request failed: ${url}`, error)
      throw error
    }
  }

  private async requestWithRetry(url: string, options: RequestInit): Promise<Response> {
    let lastError: Error

    for (let attempt = 0; attempt <= this.retryConfig.maxRetries; attempt++) {
      try {
        return await fetch(url, options)
      } catch (error) {
        lastError = error as Error

        if (attempt === this.retryConfig.maxRetries) {
          break
        }

        // Only retry on network errors, not HTTP errors
        if (this.isRetryableError(error)) {
          const delay = Math.min(
            this.retryConfig.baseDelay * Math.pow(2, attempt),
            this.retryConfig.maxDelay
          )

          console.warn(`Request failed, retrying in ${delay}ms (attempt ${attempt + 1}/${this.retryConfig.maxRetries})`)
          await new Promise(resolve => setTimeout(resolve, delay))
        } else {
          break
        }
      }
    }

    throw lastError!
  }

  private isRetryableError(error: any): boolean {
    // Retry on network errors, timeouts, and 5xx server errors
    return error.name === 'TypeError' || // Network error
           error.code === 'ECONNREFUSED' ||
           error.code === 'ENOTFOUND' ||
           error.code === 'TIMEOUT'
  }

  private generateRequestId(): string {
    return `req_${Date.now()}_${Math.random().toString(36).substr(2, 9)}`
  }

  // Dashboard endpoints
  async getStats(): Promise<DashboardStats> {
    const rawStats = await this.request<any>('/stats')
    return transformDashboardStats(rawStats)
  }

  async getStatus(): Promise<SystemStatus> {
    const rawStatus = await this.request<any>('/status')
    return transformSystemStatus(rawStatus)
  }

  async getPlugins(): Promise<PluginInfo[]> {
    const rawPlugins = await this.request<any[]>('/plugins')
    return transformArray(rawPlugins, transformPluginInfo)
  }

  async getWorkers(): Promise<WorkerInfo[]> {
    const rawWorkers = await this.request<any[]>('/workers')
    return transformArray(rawWorkers, transformWorkerInfo)
  }

  async getLogs(params?: {
    level?: string
    source?: string
    limit?: number
    offset?: number
    since?: string
  }): Promise<LogEntry[]> {
    const searchParams = new URLSearchParams()
    if (params) {
      Object.entries(params).forEach(([key, value]) => {
        if (value !== undefined) {
          searchParams.append(key, value.toString())
        }
      })
    }
    const query = searchParams.toString()
    const rawLogs = await this.request<any[]>(`/logs${query ? `?${query}` : ''}`)
    return transformArray(rawLogs, transformLogEntry)
  }

  async getActivity(limit = 50): Promise<ActivityEntry[]> {
    // Since backend doesn't have activity endpoint, create from other data
    const [plugins, logs, workers] = await Promise.all([
      this.getPlugins(),
      this.getLogs({ limit: 20 }),
      this.getWorkers()
    ])
    return createActivityEntries(plugins, logs, workers).slice(0, limit)
  }

  // Plugin management
  async enablePlugin(pluginId: string): Promise<ApiResponse> {
    return this.request<ApiResponse>(`/plugins/${pluginId}/enable`, { method: 'POST' })
  }

  async disablePlugin(pluginId: string): Promise<ApiResponse> {
    return this.request<ApiResponse>(`/plugins/${pluginId}/disable`, { method: 'POST' })
  }

  async executePlugin(pluginId: string, payload?: any): Promise<ApiResponse> {
    return this.request<ApiResponse>(`/plugins/${pluginId}/execute`, {
      method: 'POST',
      body: payload ? JSON.stringify(payload) : undefined,
    })
  }

  async getPluginLogs(pluginId: string, limit = 100): Promise<LogEntry[]> {
    return this.request<LogEntry[]>(`/plugins/${pluginId}/logs?limit=${limit}`)
  }

  async updatePluginConfig(pluginId: string, config: Record<string, any>): Promise<ApiResponse> {
    return this.request<ApiResponse>(`/plugins/${pluginId}/config`, {
      method: 'PUT',
      body: JSON.stringify(config),
    })
  }

  // Configuration management
  async getConfig(): Promise<ConfigSection[]> {
    return this.request<ConfigSection[]>('/config')
  }

  async updateConfig(config: Record<string, any>): Promise<ApiResponse> {
    return this.request<ApiResponse>('/config', {
      method: 'PUT',
      body: JSON.stringify(config),
    })
  }

  async resetConfig(): Promise<ApiResponse> {
    return this.request<ApiResponse>('/config/reset', { method: 'POST' })
  }

  // Python environment endpoints
  async getPythonEnvStatus(): Promise<any> {
    return this.request('/python-env')
  }

  async downloadUV(): Promise<ApiResponse> {
    return this.request('/download-uv', { method: 'POST' })
  }

  async downloadPython(): Promise<ApiResponse> {
    return this.request('/download-python', { method: 'POST' })
  }

  async setupVirtualEnv(): Promise<ApiResponse> {
    return this.request('/setup-venv', { method: 'POST' })
  }

  async testPythonEnv(): Promise<ApiResponse> {
    return this.request('/test-python', { method: 'POST' })
  }

  // Worker management
  async submitJob(jobData: any): Promise<ApiResponse> {
    return this.request('/workers/jobs', {
      method: 'POST',
      body: JSON.stringify(jobData),
    })
  }

  // Generic HTTP methods for new endpoints
  async get<T = any>(endpoint: string): Promise<T> {
    return this.request<T>(endpoint)
  }

  async post<T = any>(endpoint: string, data?: any): Promise<T> {
    return this.request<T>(endpoint, {
      method: 'POST',
      body: data ? JSON.stringify(data) : undefined,
    })
  }

  async put<T = any>(endpoint: string, data?: any): Promise<T> {
    return this.request<T>(endpoint, {
      method: 'PUT',
      body: data ? JSON.stringify(data) : undefined,
    })
  }

  async delete<T = any>(endpoint: string): Promise<T> {
    return this.request<T>(endpoint, {
      method: 'DELETE',
    })
  }

  // Python Interpreter Management
  async getInterpreters(): Promise<PythonEnvironment[]> {
    return this.request('/interpreters')
  }

  async setDefaultInterpreter(interpreterId: string): Promise<ApiResponse> {
    return this.request(`/interpreters/${interpreterId}/default`, { method: 'POST' })
  }

  async testInterpreter(interpreterId: string): Promise<ApiResponse> {
    return this.request(`/interpreters/${interpreterId}/test`, { method: 'POST' })
  }

  // Connection Status
  async getConnections(): Promise<ConnectionInfo[]> {
    return this.request('/connections')
  }

  async testConnection(connectionId: string): Promise<ApiResponse> {
    return this.request(`/connections/${connectionId}/test`, { method: 'POST' })
  }

  // API Testing
  async testApiEndpoint(request: ApiTestRequest): Promise<ApiTestResponse> {
    return this.request('/test-api', {
      method: 'POST',
      body: JSON.stringify(request),
    })
  }
}

// Retry utility function
export async function withRetry<T>(
  fn: () => Promise<T>,
  maxRetries = 3,
  delay = 1000
): Promise<T> {
  let lastError: Error

  for (let i = 0; i <= maxRetries; i++) {
    try {
      return await fn()
    } catch (error) {
      lastError = error as Error
      if (i === maxRetries) break
      
      // Exponential backoff
      await new Promise(resolve => setTimeout(resolve, delay * Math.pow(2, i)))
    }
  }

  throw lastError!
}

// Create and export the API client instance
export const apiClient = new ApiClient()
export default apiClient
