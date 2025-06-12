/**
 * Bridge Status Checker
 * Utilities to check the connection status between frontend and backend
 */

export interface BridgeStatus {
  api: {
    connected: boolean
    baseUrl: string
    lastCheck: Date | null
    error?: string
  }
  websocket: {
    connected: boolean
    url: string
    lastCheck: Date | null
    error?: string
  }
  backend: {
    goServer: boolean
    pythonExecutor: boolean
    lastCheck: Date | null
  }
}

class BridgeStatusChecker {
  private status: BridgeStatus = {
    api: {
      connected: false,
      baseUrl: '',
      lastCheck: null,
    },
    websocket: {
      connected: false,
      url: '',
      lastCheck: null,
    },
    backend: {
      goServer: false,
      pythonExecutor: false,
      lastCheck: null,
    }
  }

  async checkApiConnection(): Promise<boolean> {
    try {
      const baseUrl = this.getApiBaseUrl()
      this.status.api.baseUrl = baseUrl
      
      const response = await fetch(`${baseUrl}/status`, {
        method: 'GET',
        headers: { 'Content-Type': 'application/json' }
      })
      
      const connected = response.ok
      this.status.api.connected = connected
      this.status.api.lastCheck = new Date()
      
      if (!connected) {
        this.status.api.error = `HTTP ${response.status}: ${response.statusText}`
      } else {
        delete this.status.api.error
      }
      
      return connected
    } catch (error) {
      this.status.api.connected = false
      this.status.api.lastCheck = new Date()
      this.status.api.error = error instanceof Error ? error.message : 'Unknown error'
      return false
    }
  }

  async checkBackendServices(): Promise<void> {
    try {
      const baseUrl = this.getApiBaseUrl()
      const response = await fetch(`${baseUrl}/status`)
      
      if (response.ok) {
        const data = await response.json()
        
        // Handle both direct response and wrapped response
        const statusData = data.data || data
        
        this.status.backend.goServer = statusData.server_status === 'running'
        this.status.backend.pythonExecutor = statusData.grpc_connected === true
        this.status.backend.lastCheck = new Date()
      }
    } catch (error) {
      this.status.backend.goServer = false
      this.status.backend.pythonExecutor = false
      this.status.backend.lastCheck = new Date()
    }
  }

  checkWebSocketConnection(): boolean {
    // This would need to be implemented with actual WebSocket instance
    // For now, return a basic check
    this.status.websocket.url = this.getWebSocketUrl()
    this.status.websocket.lastCheck = new Date()
    
    // TODO: Implement actual WebSocket connection check
    this.status.websocket.connected = false
    
    return this.status.websocket.connected
  }

  async performFullCheck(): Promise<BridgeStatus> {
    await Promise.all([
      this.checkApiConnection(),
      this.checkBackendServices(),
    ])
    
    this.checkWebSocketConnection()
    
    return this.getStatus()
  }

  getStatus(): BridgeStatus {
    return { ...this.status }
  }

  private getApiBaseUrl(): string {
    if (typeof window !== 'undefined') {
      return process.env.NEXT_PUBLIC_API_BASE_URL 
        ? `${process.env.NEXT_PUBLIC_API_BASE_URL}/api/dashboard`
        : '/api/dashboard'
    }
    return '/api/dashboard'
  }

  private getWebSocketUrl(): string {
    if (typeof window !== 'undefined') {
      const wsBase = process.env.NEXT_PUBLIC_WS_BASE_URL || 
                    (process.env.NEXT_PUBLIC_API_BASE_URL?.replace('http', 'ws')) ||
                    `ws://${window.location.host}`
      return `${wsBase}/api/dashboard/monitor/stream`
    }
    return '/api/dashboard/monitor/stream'
  }

  isFullyConnected(): boolean {
    return this.status.api.connected && 
           this.status.backend.goServer && 
           this.status.backend.pythonExecutor
  }

  getConnectionSummary(): string {
    const { api, backend } = this.status
    
    if (!api.connected) {
      return 'API connection failed'
    }
    
    if (!backend.goServer) {
      return 'Go server not running'
    }
    
    if (!backend.pythonExecutor) {
      return 'Python executor not connected'
    }
    
    return 'All services connected'
  }

  getHealthScore(): number {
    let score = 0
    let total = 0
    
    // API connection (40% weight)
    total += 40
    if (this.status.api.connected) score += 40
    
    // Go server (30% weight)
    total += 30
    if (this.status.backend.goServer) score += 30
    
    // Python executor (30% weight)
    total += 30
    if (this.status.backend.pythonExecutor) score += 30
    
    return Math.round((score / total) * 100)
  }
}

// Export singleton instance
export const bridgeStatusChecker = new BridgeStatusChecker()

// Utility functions
export async function checkBridgeHealth(): Promise<BridgeStatus> {
  return bridgeStatusChecker.performFullCheck()
}

export function getBridgeStatus(): BridgeStatus {
  return bridgeStatusChecker.getStatus()
}

export function isBridgeHealthy(): boolean {
  return bridgeStatusChecker.isFullyConnected()
}

export function getBridgeHealthScore(): number {
  return bridgeStatusChecker.getHealthScore()
}

export function getBridgeConnectionSummary(): string {
  return bridgeStatusChecker.getConnectionSummary()
}
