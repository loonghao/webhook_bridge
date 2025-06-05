import { useState, useEffect, useCallback, useRef } from 'react'
import { 
  monitorWebSocket, 
  connectToMonitor, 
  MonitorMessage, 
  PluginStatusUpdate, 
  SystemMetricsUpdate 
} from '@/services/websocket'

export interface RealTimeMonitoringState {
  isConnected: boolean
  systemMetrics: SystemMetricsUpdate | null
  pluginUpdates: PluginStatusUpdate[]
  connectionError: string | null
  lastUpdate: Date | null
}

export interface RealTimeMonitoringActions {
  connect: () => Promise<void>
  disconnect: () => void
  clearPluginUpdates: () => void
  retry: () => Promise<void>
}

export function useRealTimeMonitoring(): [RealTimeMonitoringState, RealTimeMonitoringActions] {
  const [state, setState] = useState<RealTimeMonitoringState>({
    isConnected: false,
    systemMetrics: null,
    pluginUpdates: [],
    connectionError: null,
    lastUpdate: null
  })

  const retryTimeoutRef = useRef<NodeJS.Timeout>()
  const maxPluginUpdates = 50 // Keep only the last 50 plugin updates

  const handleSystemMetrics = useCallback((message: MonitorMessage) => {
    if (message.type === 'system_metrics') {
      setState(prev => ({
        ...prev,
        systemMetrics: message.data as SystemMetricsUpdate,
        lastUpdate: new Date()
      }))
    }
  }, [])

  const handlePluginStatus = useCallback((message: MonitorMessage) => {
    if (message.type === 'plugin_status') {
      const update = message.data as PluginStatusUpdate
      setState(prev => ({
        ...prev,
        pluginUpdates: [
          update,
          ...prev.pluginUpdates.slice(0, maxPluginUpdates - 1)
        ],
        lastUpdate: new Date()
      }))
    }
  }, [])

  const handleConnection = useCallback((message: MonitorMessage) => {
    if (message.type === 'connected') {
      setState(prev => ({
        ...prev,
        isConnected: true,
        connectionError: null
      }))
    } else if (message.type === 'disconnected') {
      setState(prev => ({
        ...prev,
        isConnected: false,
        connectionError: message.data.reason || 'Connection lost'
      }))
    } else if (message.type === 'error') {
      setState(prev => ({
        ...prev,
        isConnected: false,
        connectionError: 'Connection error'
      }))
    }
  }, [])

  const connect = useCallback(async () => {
    try {
      setState(prev => ({ ...prev, connectionError: null }))
      await connectToMonitor()
    } catch (error) {
      setState(prev => ({
        ...prev,
        connectionError: error instanceof Error ? error.message : 'Failed to connect'
      }))
      throw error
    }
  }, [])

  const disconnect = useCallback(() => {
    monitorWebSocket.disconnect()
    setState(prev => ({
      ...prev,
      isConnected: false,
      connectionError: null
    }))
  }, [])

  const clearPluginUpdates = useCallback(() => {
    setState(prev => ({
      ...prev,
      pluginUpdates: []
    }))
  }, [])

  const retry = useCallback(async () => {
    if (retryTimeoutRef.current) {
      clearTimeout(retryTimeoutRef.current)
    }
    
    try {
      await connect()
    } catch (error) {
      // Schedule retry after 5 seconds
      retryTimeoutRef.current = setTimeout(() => {
        retry()
      }, 5000)
    }
  }, [connect])

  useEffect(() => {
    // Set up event handlers
    monitorWebSocket.on('system_metrics', handleSystemMetrics)
    monitorWebSocket.on('plugin_status', handlePluginStatus)
    monitorWebSocket.on('connected', handleConnection)
    monitorWebSocket.on('disconnected', handleConnection)
    monitorWebSocket.on('error', handleConnection)

    // Auto-connect on mount
    connect().catch(console.error)

    return () => {
      // Clean up event handlers
      monitorWebSocket.off('system_metrics', handleSystemMetrics)
      monitorWebSocket.off('plugin_status', handlePluginStatus)
      monitorWebSocket.off('connected', handleConnection)
      monitorWebSocket.off('disconnected', handleConnection)
      monitorWebSocket.off('error', handleConnection)
      
      // Clear retry timeout
      if (retryTimeoutRef.current) {
        clearTimeout(retryTimeoutRef.current)
      }
    }
  }, [connect, handleSystemMetrics, handlePluginStatus, handleConnection])

  return [
    state,
    {
      connect,
      disconnect,
      clearPluginUpdates,
      retry
    }
  ]
}

// Hook for getting real-time system metrics only
export function useSystemMetrics() {
  const [monitoring] = useRealTimeMonitoring()
  
  return {
    metrics: monitoring.systemMetrics,
    isConnected: monitoring.isConnected,
    lastUpdate: monitoring.lastUpdate,
    error: monitoring.connectionError
  }
}

// Hook for getting real-time plugin updates only
export function usePluginUpdates() {
  const [monitoring, actions] = useRealTimeMonitoring()
  
  return {
    updates: monitoring.pluginUpdates,
    isConnected: monitoring.isConnected,
    lastUpdate: monitoring.lastUpdate,
    error: monitoring.connectionError,
    clearUpdates: actions.clearPluginUpdates
  }
}

// Hook for connection management
export function useMonitoringConnection() {
  const [monitoring, actions] = useRealTimeMonitoring()
  
  return {
    isConnected: monitoring.isConnected,
    error: monitoring.connectionError,
    connect: actions.connect,
    disconnect: actions.disconnect,
    retry: actions.retry
  }
}
