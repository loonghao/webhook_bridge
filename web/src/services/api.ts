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

  async getPlugins(): Promise<PluginInfo[]> {
    const response = await this.request<ApiResponse<PluginInfo[]>>('/plugins')
    return response.data || []
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
