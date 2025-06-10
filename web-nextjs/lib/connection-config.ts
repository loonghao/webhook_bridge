/**
 * Connection Configuration
 * Centralized configuration for all frontend-backend connections
 */

export interface ConnectionConfig {
  api: {
    baseUrl: string
    timeout: number
    retries: {
      maxRetries: number
      baseDelay: number
      maxDelay: number
    }
    healthCheck: {
      enabled: boolean
      interval: number
      endpoint: string
    }
  }
  websocket: {
    baseUrl: string
    reconnect: {
      enabled: boolean
      maxAttempts: number
      baseDelay: number
      maxDelay: number
    }
    heartbeat: {
      enabled: boolean
      interval: number
    }
  }
  monitoring: {
    enabled: boolean
    stagewise: boolean
    performanceThresholds: {
      slowRequest: number
      errorRate: number
    }
  }
}

// Default configuration
const defaultConfig: ConnectionConfig = {
  api: {
    baseUrl: getApiBaseUrl(),
    timeout: 30000, // 30 seconds
    retries: {
      maxRetries: 3,
      baseDelay: 1000,
      maxDelay: 5000
    },
    healthCheck: {
      enabled: true,
      interval: 30000, // 30 seconds
      endpoint: '/status'
    }
  },
  websocket: {
    baseUrl: getWebSocketBaseUrl(),
    reconnect: {
      enabled: true,
      maxAttempts: 5,
      baseDelay: 1000,
      maxDelay: 30000
    },
    heartbeat: {
      enabled: true,
      interval: 30000 // 30 seconds
    }
  },
  monitoring: {
    enabled: process.env.NODE_ENV === 'development',
    stagewise: process.env.NEXT_PUBLIC_ENABLE_STAGEWISE === 'true',
    performanceThresholds: {
      slowRequest: 2000, // 2 seconds
      errorRate: 0.1 // 10%
    }
  }
}

function getApiBaseUrl(): string {
  if (typeof window !== 'undefined') {
    // Client-side
    if (process.env.NEXT_PUBLIC_API_BASE_URL) {
      return `${process.env.NEXT_PUBLIC_API_BASE_URL}/api/dashboard`
    }
    
    // Default to backend server (port 8080) when running in development
    const protocol = window.location.protocol
    const hostname = window.location.hostname
    const backendPort = '8080'
    return `${protocol}//${hostname}:${backendPort}/api/dashboard`
  }
  
  // Server-side
  return '/api/dashboard'
}

function getWebSocketBaseUrl(): string {
  if (typeof window !== 'undefined') {
    if (process.env.NEXT_PUBLIC_WS_BASE_URL) {
      return process.env.NEXT_PUBLIC_WS_BASE_URL
    }
    
    if (process.env.NEXT_PUBLIC_API_BASE_URL) {
      return process.env.NEXT_PUBLIC_API_BASE_URL.replace('http', 'ws')
    }
    
    // Default to backend server
    const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:'
    const hostname = window.location.hostname
    const backendPort = '8080'
    return `${protocol}//${hostname}:${backendPort}`
  }
  
  return 'ws://localhost:8080'
}

// Configuration manager
class ConnectionConfigManager {
  private config: ConnectionConfig = { ...defaultConfig }
  private listeners: Set<(config: ConnectionConfig) => void> = new Set()

  getConfig(): ConnectionConfig {
    return { ...this.config }
  }

  updateConfig(updates: Partial<ConnectionConfig>): void {
    this.config = this.mergeConfig(this.config, updates)
    this.notifyListeners()
  }

  updateApiConfig(updates: Partial<ConnectionConfig['api']>): void {
    this.config.api = { ...this.config.api, ...updates }
    this.notifyListeners()
  }

  updateWebSocketConfig(updates: Partial<ConnectionConfig['websocket']>): void {
    this.config.websocket = { ...this.config.websocket, ...updates }
    this.notifyListeners()
  }

  updateMonitoringConfig(updates: Partial<ConnectionConfig['monitoring']>): void {
    this.config.monitoring = { ...this.config.monitoring, ...updates }
    this.notifyListeners()
  }

  // Auto-optimize based on network conditions
  optimizeForNetwork(networkType: 'fast' | 'slow' | 'unreliable'): void {
    switch (networkType) {
      case 'fast':
        this.updateConfig({
          api: {
            ...this.config.api,
            timeout: 10000,
            retries: { maxRetries: 2, baseDelay: 500, maxDelay: 2000 }
          },
          websocket: {
            ...this.config.websocket,
            reconnect: { enabled: true, maxAttempts: 3, baseDelay: 500, maxDelay: 5000 }
          }
        })
        break
        
      case 'slow':
        this.updateConfig({
          api: {
            ...this.config.api,
            timeout: 60000,
            retries: { maxRetries: 5, baseDelay: 2000, maxDelay: 10000 }
          },
          websocket: {
            ...this.config.websocket,
            reconnect: { enabled: true, maxAttempts: 10, baseDelay: 2000, maxDelay: 60000 }
          }
        })
        break
        
      case 'unreliable':
        this.updateConfig({
          api: {
            ...this.config.api,
            timeout: 45000,
            retries: { maxRetries: 8, baseDelay: 1500, maxDelay: 15000 }
          },
          websocket: {
            ...this.config.websocket,
            reconnect: { enabled: true, maxAttempts: 15, baseDelay: 1000, maxDelay: 30000 }
          }
        })
        break
    }
  }

  // Auto-optimize based on error patterns
  optimizeForErrors(errorPatterns: { timeouts: number, networkErrors: number, serverErrors: number }): void {
    const totalErrors = errorPatterns.timeouts + errorPatterns.networkErrors + errorPatterns.serverErrors
    
    if (totalErrors === 0) return

    // High timeout rate - increase timeout and retries
    if (errorPatterns.timeouts / totalErrors > 0.3) {
      this.updateApiConfig({
        timeout: Math.min(this.config.api.timeout * 1.5, 120000),
        retries: {
          ...this.config.api.retries,
          maxRetries: Math.min(this.config.api.retries.maxRetries + 2, 10)
        }
      })
    }

    // High network error rate - increase retry delays
    if (errorPatterns.networkErrors / totalErrors > 0.4) {
      this.updateApiConfig({
        retries: {
          ...this.config.api.retries,
          baseDelay: Math.min(this.config.api.retries.baseDelay * 1.5, 5000),
          maxDelay: Math.min(this.config.api.retries.maxDelay * 1.5, 30000)
        }
      })
      
      this.updateWebSocketConfig({
        reconnect: {
          ...this.config.websocket.reconnect,
          baseDelay: Math.min(this.config.websocket.reconnect.baseDelay * 1.5, 5000),
          maxDelay: Math.min(this.config.websocket.reconnect.maxDelay * 1.5, 60000)
        }
      })
    }
  }

  subscribe(listener: (config: ConnectionConfig) => void): () => void {
    this.listeners.add(listener)
    return () => this.listeners.delete(listener)
  }

  private mergeConfig(base: ConnectionConfig, updates: Partial<ConnectionConfig>): ConnectionConfig {
    return {
      api: { ...base.api, ...updates.api },
      websocket: { ...base.websocket, ...updates.websocket },
      monitoring: { ...base.monitoring, ...updates.monitoring }
    }
  }

  private notifyListeners(): void {
    this.listeners.forEach(listener => {
      try {
        listener(this.config)
      } catch (error) {
        console.error('Error in config listener:', error)
      }
    })
  }

  // Load configuration from environment or storage
  loadFromEnvironment(): void {
    const envConfig: Partial<ConnectionConfig> = {}

    // API configuration
    if (process.env.NEXT_PUBLIC_API_TIMEOUT) {
      envConfig.api = {
        ...this.config.api,
        timeout: parseInt(process.env.NEXT_PUBLIC_API_TIMEOUT)
      }
    }

    // WebSocket configuration
    if (process.env.NEXT_PUBLIC_WS_RECONNECT_ATTEMPTS) {
      envConfig.websocket = {
        ...this.config.websocket,
        reconnect: {
          ...this.config.websocket.reconnect,
          maxAttempts: parseInt(process.env.NEXT_PUBLIC_WS_RECONNECT_ATTEMPTS)
        }
      }
    }

    // Monitoring configuration
    if (process.env.NEXT_PUBLIC_MONITORING_ENABLED) {
      envConfig.monitoring = {
        ...this.config.monitoring,
        enabled: process.env.NEXT_PUBLIC_MONITORING_ENABLED === 'true'
      }
    }

    if (Object.keys(envConfig).length > 0) {
      this.updateConfig(envConfig)
    }
  }

  // Save configuration to local storage
  saveToStorage(): void {
    if (typeof window !== 'undefined') {
      try {
        localStorage.setItem('webhook_bridge_connection_config', JSON.stringify(this.config))
      } catch (error) {
        console.warn('Failed to save connection config to storage:', error)
      }
    }
  }

  // Load configuration from local storage
  loadFromStorage(): void {
    if (typeof window !== 'undefined') {
      try {
        const stored = localStorage.getItem('webhook_bridge_connection_config')
        if (stored) {
          const storedConfig = JSON.parse(stored)
          this.updateConfig(storedConfig)
        }
      } catch (error) {
        console.warn('Failed to load connection config from storage:', error)
      }
    }
  }
}

// Global instance
export const connectionConfig = new ConnectionConfigManager()

// Initialize configuration
if (typeof window !== 'undefined') {
  connectionConfig.loadFromEnvironment()
  connectionConfig.loadFromStorage()
}

// React hook for using connection config
export function useConnectionConfig() {
  const [config, setConfig] = React.useState<ConnectionConfig>(connectionConfig.getConfig())

  React.useEffect(() => {
    return connectionConfig.subscribe(setConfig)
  }, [])

  return {
    config,
    updateConfig: connectionConfig.updateConfig.bind(connectionConfig),
    updateApiConfig: connectionConfig.updateApiConfig.bind(connectionConfig),
    updateWebSocketConfig: connectionConfig.updateWebSocketConfig.bind(connectionConfig),
    updateMonitoringConfig: connectionConfig.updateMonitoringConfig.bind(connectionConfig),
    optimizeForNetwork: connectionConfig.optimizeForNetwork.bind(connectionConfig),
    optimizeForErrors: connectionConfig.optimizeForErrors.bind(connectionConfig),
    saveToStorage: connectionConfig.saveToStorage.bind(connectionConfig)
  }
}

// Add React import
import React from 'react'
