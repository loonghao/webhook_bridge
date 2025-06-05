import { 
  ApiResponse, 
  DashboardStats, 
  SystemStatus, 
  PluginInfo, 
  WorkerInfo, 
  LogEntry,
  ActivityEntry 
} from '@/types/api'

const API_BASE = '/api/dashboard'

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
      return data
    } catch (error) {
      console.error(`API request failed: ${url}`, error)
      throw error
    }
  }

  // Dashboard endpoints
  async getStats(): Promise<DashboardStats> {
    return this.request<DashboardStats>('/stats')
  }

  async getStatus(): Promise<SystemStatus> {
    return this.request<SystemStatus>('/status')
  }

  async getPlugins(): Promise<ApiResponse<PluginInfo[]>> {
    return this.request<ApiResponse<PluginInfo[]>>('/plugins')
  }

  async getWorkers(): Promise<WorkerInfo[]> {
    const response = await this.request<ApiResponse<WorkerInfo[]>>('/workers')
    return response.data || []
  }

  async getLogs(params?: { limit?: number; level?: string }): Promise<LogEntry[]> {
    const searchParams = new URLSearchParams()
    if (params?.limit) searchParams.append('limit', params.limit.toString())
    if (params?.level) searchParams.append('level', params.level)
    
    const endpoint = `/logs${searchParams.toString() ? `?${searchParams}` : ''}`
    const response = await this.request<ApiResponse<LogEntry[]>>(endpoint)
    return response.data || []
  }

  async getRecentActivity(): Promise<ActivityEntry[]> {
    // This would be a custom endpoint for recent activity
    // For now, we'll use logs as activity
    const logs = await this.getLogs({ limit: 10 })
    return logs.map((log, index) => ({
      id: `activity-${index}`,
      type: 'system' as const,
      message: log.message,
      timestamp: log.timestamp,
      status: log.level === 'error' ? 'error' : log.level === 'warn' ? 'warning' : 'success',
      source: log.source
    }))
  }

  // Configuration endpoints
  async getConfig(): Promise<any> {
    return this.request('/config')
  }

  async saveConfig(config: any): Promise<ApiResponse> {
    return this.request('/config', {
      method: 'POST',
      body: JSON.stringify(config),
    })
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
  async getInterpreters(): Promise<any> {
    return this.request('/interpreters')
  }

  async addInterpreter(interpreterData: any): Promise<ApiResponse> {
    return this.request('/interpreters', {
      method: 'POST',
      body: JSON.stringify(interpreterData),
    })
  }

  async removeInterpreter(name: string): Promise<ApiResponse> {
    return this.request(`/interpreters/${name}`, {
      method: 'DELETE',
    })
  }

  async validateInterpreter(name: string): Promise<ApiResponse> {
    return this.request(`/interpreters/${name}/validate`, {
      method: 'POST',
    })
  }

  async activateInterpreter(name: string): Promise<ApiResponse> {
    return this.request(`/interpreters/${name}/activate`, {
      method: 'POST',
    })
  }

  async discoverInterpreters(): Promise<any> {
    return this.request('/interpreters/discover')
  }

  // Connection Management
  async getConnectionStatus(): Promise<any> {
    return this.request('/connection')
  }

  async reconnectService(interpreterName?: string): Promise<ApiResponse> {
    const data = interpreterName ? { interpreter_name: interpreterName } : {}
    return this.request('/connection/reconnect', {
      method: 'POST',
      body: JSON.stringify(data),
    })
  }

  async testConnection(): Promise<ApiResponse> {
    return this.request('/connection/test', {
      method: 'POST',
    })
  }

  // New unified API methods
  async getSystemStatus(): Promise<ApiResponse<SystemStatus>> {
    return this.request('/status')
  }

  async getSystemInfo(): Promise<ApiResponse<any>> {
    return this.request('/system')
  }

  async getMetrics(): Promise<ApiResponse<any>> {
    return this.request('/metrics')
  }

  // Plugin-specific methods
  async getPluginStats(): Promise<any> {
    return this.request('/plugins/stats')
  }

  async getPluginLogs(pluginName: string, options?: { limit?: number }): Promise<any> {
    const params = options ? `?limit=${options.limit || 50}` : ''
    return this.request(`/plugins/${pluginName}/logs${params}`)
  }

  async executePlugin(request: any): Promise<any> {
    return this.request(`/plugins/${request.plugin}/execute`, {
      method: 'POST',
      body: JSON.stringify({
        method: request.method,
        data: request.data
      })
    })
  }

  // Enhanced Plugin Management methods
  async getPluginDetails(pluginName: string): Promise<any> {
    return this.request(`/plugins/${pluginName}`)
  }

  async getPluginStatistics(pluginName?: string): Promise<any> {
    const endpoint = pluginName ? `/plugins/${pluginName}/statistics` : '/plugins/statistics'
    return this.request(endpoint)
  }

  async getPluginExecutionHistory(pluginName: string, options?: { limit?: number; offset?: number }): Promise<any> {
    const params = new URLSearchParams()
    if (options?.limit) params.append('limit', options.limit.toString())
    if (options?.offset) params.append('offset', options.offset.toString())

    const endpoint = `/plugins/${pluginName}/history${params.toString() ? `?${params}` : ''}`
    return this.request(endpoint)
  }

  async getAllPluginLogs(options?: { limit?: number; plugin?: string; level?: string }): Promise<any> {
    const params = new URLSearchParams()
    if (options?.limit) params.append('limit', options.limit.toString())
    if (options?.plugin) params.append('plugin', options.plugin)
    if (options?.level) params.append('level', options.level)

    const endpoint = `/plugins/logs${params.toString() ? `?${params}` : ''}`
    return this.request(endpoint)
  }

  async testPluginConnection(pluginName: string): Promise<any> {
    return this.request(`/plugins/${pluginName}/test`, {
      method: 'POST'
    })
  }

  async getPluginMetrics(): Promise<any> {
    return this.request('/plugins/metrics')
  }
}

export const apiClient = new ApiClient()

// Utility functions for error handling and retries
export async function withRetry<T>(
  fn: () => Promise<T>,
  retries: number = 3,
  delay: number = 1000
): Promise<T> {
  try {
    return await fn()
  } catch (error) {
    if (retries > 0) {
      await new Promise(resolve => setTimeout(resolve, delay))
      return withRetry(fn, retries - 1, delay * 2)
    }
    throw error
  }
}

export function isApiError(error: any): error is Error {
  return error instanceof Error
}
