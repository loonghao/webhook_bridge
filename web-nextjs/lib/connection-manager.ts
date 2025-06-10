/**
 * Connection Manager
 * Manages and monitors all frontend-backend connections
 */

import { apiClient } from '@/services/api'
import { monitorWebSocket, logsWebSocket } from '@/services/websocket'

export interface ConnectionState {
  api: {
    status: 'connected' | 'connecting' | 'disconnected' | 'error'
    lastCheck: Date | null
    latency: number | null
    error?: string
  }
  websocket: {
    monitor: {
      status: 'connected' | 'connecting' | 'disconnected' | 'error'
      lastCheck: Date | null
      error?: string
    }
    logs: {
      status: 'connected' | 'connecting' | 'disconnected' | 'error'
      lastCheck: Date | null
      error?: string
    }
  }
  backend: {
    goServer: boolean
    pythonExecutor: boolean
    lastCheck: Date | null
  }
}

export type ConnectionEventType = 'api_status_change' | 'websocket_status_change' | 'backend_status_change'

export interface ConnectionEvent {
  type: ConnectionEventType
  data: any
  timestamp: Date
}

class ConnectionManager {
  private state: ConnectionState = {
    api: {
      status: 'disconnected',
      lastCheck: null,
      latency: null
    },
    websocket: {
      monitor: {
        status: 'disconnected',
        lastCheck: null
      },
      logs: {
        status: 'disconnected',
        lastCheck: null
      }
    },
    backend: {
      goServer: false,
      pythonExecutor: false,
      lastCheck: null
    }
  }

  private listeners: Map<ConnectionEventType, Function[]> = new Map()
  private checkInterval: NodeJS.Timeout | null = null
  private isChecking = false

  constructor() {
    this.initializeEventListeners()
  }

  private initializeEventListeners() {
    // Listen to WebSocket events
    monitorWebSocket.on('open', () => {
      this.updateWebSocketStatus('monitor', 'connected')
    })

    monitorWebSocket.on('close', () => {
      this.updateWebSocketStatus('monitor', 'disconnected')
    })

    monitorWebSocket.on('error', (error: any) => {
      this.updateWebSocketStatus('monitor', 'error', error.message)
    })

    logsWebSocket.on('open', () => {
      this.updateWebSocketStatus('logs', 'connected')
    })

    logsWebSocket.on('close', () => {
      this.updateWebSocketStatus('logs', 'disconnected')
    })

    logsWebSocket.on('error', (error: any) => {
      this.updateWebSocketStatus('logs', 'error', error.message)
    })
  }

  async checkApiConnection(): Promise<boolean> {
    const startTime = Date.now()
    
    try {
      this.state.api.status = 'connecting'
      this.emit('api_status_change', this.state.api)

      const response = await apiClient.getStatus()
      const latency = Date.now() - startTime

      this.state.api = {
        status: 'connected',
        lastCheck: new Date(),
        latency,
        error: undefined
      }

      // Update backend status from API response
      this.state.backend = {
        goServer: response.server_status === 'running',
        pythonExecutor: response.grpc_connected === true,
        lastCheck: new Date()
      }

      this.emit('api_status_change', this.state.api)
      this.emit('backend_status_change', this.state.backend)

      return true
    } catch (error) {
      const latency = Date.now() - startTime

      this.state.api = {
        status: 'error',
        lastCheck: new Date(),
        latency,
        error: error instanceof Error ? error.message : String(error)
      }

      this.emit('api_status_change', this.state.api)
      return false
    }
  }

  async checkWebSocketConnections(): Promise<void> {
    // Check monitor WebSocket
    if (!monitorWebSocket.isConnected) {
      try {
        this.updateWebSocketStatus('monitor', 'connecting')
        await monitorWebSocket.connect()
      } catch (error) {
        this.updateWebSocketStatus('monitor', 'error', error instanceof Error ? error.message : String(error))
      }
    }

    // Check logs WebSocket
    if (!logsWebSocket.isConnected) {
      try {
        this.updateWebSocketStatus('logs', 'connecting')
        await logsWebSocket.connect()
      } catch (error) {
        this.updateWebSocketStatus('logs', 'error', error instanceof Error ? error.message : String(error))
      }
    }
  }

  private updateWebSocketStatus(type: 'monitor' | 'logs', status: ConnectionState['websocket']['monitor']['status'], error?: string) {
    this.state.websocket[type] = {
      status,
      lastCheck: new Date(),
      error
    }
    this.emit('websocket_status_change', this.state.websocket)
  }

  async performFullCheck(): Promise<ConnectionState> {
    if (this.isChecking) {
      return this.state
    }

    this.isChecking = true

    try {
      await Promise.all([
        this.checkApiConnection(),
        this.checkWebSocketConnections()
      ])
    } finally {
      this.isChecking = false
    }

    return this.state
  }

  startPeriodicChecks(interval: number = 30000): void {
    if (this.checkInterval) {
      clearInterval(this.checkInterval)
    }

    this.checkInterval = setInterval(() => {
      this.performFullCheck()
    }, interval)

    // Perform initial check
    this.performFullCheck()
  }

  stopPeriodicChecks(): void {
    if (this.checkInterval) {
      clearInterval(this.checkInterval)
      this.checkInterval = null
    }
  }

  getState(): ConnectionState {
    return { ...this.state }
  }

  isFullyConnected(): boolean {
    return this.state.api.status === 'connected' &&
           this.state.backend.goServer &&
           this.state.backend.pythonExecutor
  }

  getHealthScore(): number {
    let score = 0
    let total = 0

    // API connection (40% weight)
    total += 40
    if (this.state.api.status === 'connected') score += 40
    else if (this.state.api.status === 'connecting') score += 20

    // Backend services (40% weight)
    total += 40
    if (this.state.backend.goServer) score += 20
    if (this.state.backend.pythonExecutor) score += 20

    // WebSocket connections (20% weight)
    total += 20
    if (this.state.websocket.monitor.status === 'connected') score += 10
    if (this.state.websocket.logs.status === 'connected') score += 10

    return total > 0 ? Math.round((score / total) * 100) : 0
  }

  on(event: ConnectionEventType, listener: Function): void {
    if (!this.listeners.has(event)) {
      this.listeners.set(event, [])
    }
    this.listeners.get(event)!.push(listener)
  }

  off(event: ConnectionEventType, listener: Function): void {
    const listeners = this.listeners.get(event)
    if (listeners) {
      const index = listeners.indexOf(listener)
      if (index > -1) {
        listeners.splice(index, 1)
      }
    }
  }

  private emit(event: ConnectionEventType, data: any): void {
    const listeners = this.listeners.get(event)
    if (listeners) {
      const eventData: ConnectionEvent = {
        type: event,
        data,
        timestamp: new Date()
      }
      listeners.forEach(listener => {
        try {
          listener(eventData)
        } catch (error) {
          console.error('Connection event listener error:', error)
        }
      })
    }
  }

  // Diagnostic methods
  async runDiagnostics(): Promise<{
    api: { reachable: boolean, latency: number | null, error?: string }
    websocket: { monitor: boolean, logs: boolean }
    backend: { go: boolean, python: boolean }
    recommendations: string[]
  }> {
    const diagnostics = {
      api: { reachable: false, latency: null as number | null, error: undefined as string | undefined },
      websocket: { monitor: false, logs: false },
      backend: { go: false, python: false },
      recommendations: [] as string[]
    }

    // Test API connection
    try {
      const startTime = Date.now()
      await apiClient.getStatus()
      diagnostics.api.reachable = true
      diagnostics.api.latency = Date.now() - startTime
    } catch (error) {
      diagnostics.api.error = error instanceof Error ? error.message : String(error)
      diagnostics.recommendations.push('Check if backend server is running on the correct port')
    }

    // Check WebSocket connections
    diagnostics.websocket.monitor = monitorWebSocket.isConnected
    diagnostics.websocket.logs = logsWebSocket.isConnected

    if (!diagnostics.websocket.monitor || !diagnostics.websocket.logs) {
      diagnostics.recommendations.push('WebSocket connections are not established - real-time features may not work')
    }

    // Check backend services
    diagnostics.backend.go = this.state.backend.goServer
    diagnostics.backend.python = this.state.backend.pythonExecutor

    if (!diagnostics.backend.python) {
      diagnostics.recommendations.push('Python executor is not connected - plugin execution may fail')
    }

    return diagnostics
  }
}

// Global instance
export const connectionManager = new ConnectionManager()

// React hook for using connection manager
export function useConnectionManager() {
  const [state, setState] = React.useState<ConnectionState>(connectionManager.getState())

  React.useEffect(() => {
    const handleStateChange = () => {
      setState(connectionManager.getState())
    }

    connectionManager.on('api_status_change', handleStateChange)
    connectionManager.on('websocket_status_change', handleStateChange)
    connectionManager.on('backend_status_change', handleStateChange)

    // Start periodic checks
    connectionManager.startPeriodicChecks()

    return () => {
      connectionManager.off('api_status_change', handleStateChange)
      connectionManager.off('websocket_status_change', handleStateChange)
      connectionManager.off('backend_status_change', handleStateChange)
      connectionManager.stopPeriodicChecks()
    }
  }, [])

  return {
    state,
    isFullyConnected: connectionManager.isFullyConnected(),
    healthScore: connectionManager.getHealthScore(),
    performFullCheck: () => connectionManager.performFullCheck(),
    runDiagnostics: () => connectionManager.runDiagnostics()
  }
}

// Add React import for the hook
import React from 'react'
