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

// API base URL configuration
const getApiBase = () => {
  if (typeof window !== 'undefined') {
    // Client-side: use environment variable or default to backend URL
    if (process.env.NEXT_PUBLIC_API_BASE_URL) {
      return `${process.env.NEXT_PUBLIC_API_BASE_URL}/api/dashboard`
    }

    // Default to backend server (port 8080) when running in development
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
  private async request<T>(endpoint: string, options?: RequestInit): Promise<T> {
    const url = `${API_BASE}${endpoint}`

    try {
      const response = await fetch(url, {
        headers: {
          'Content-Type': 'application/json',
          ...options?.headers,
        },
        ...options,
      })

      if (!response.ok) {
        throw new Error(`HTTP error! status: ${response.status}`)
      }

      const data = await response.json()

      // Handle backend API response format
      if (data && typeof data === 'object' && 'success' in data) {
        if (!data.success) {
          throw new Error(data.message || data.error || 'API request failed')
        }
        // Return the data field if it exists, otherwise return the whole response
        return data.data !== undefined ? data.data : data
      }

      // Return raw data if not in standard API format
      return data
    } catch (error) {
      console.error(`API request failed: ${url}`, error)
      throw error
    }
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
