'use client'

import { useState, useEffect, useCallback, useRef } from 'react'
import {
  monitorWebSocket,
  connectToMonitor
} from '@/services/websocket'
import {
  MonitorMessage,
  PluginStatusUpdate,
  SystemMetricsUpdate
} from '@/types/api'

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

export function useRealTimeMonitoring(autoConnect = true): RealTimeMonitoringState & RealTimeMonitoringActions {
  const [state, setState] = useState<RealTimeMonitoringState>({
    isConnected: false,
    systemMetrics: null,
    pluginUpdates: [],
    connectionError: null,
    lastUpdate: null,
  })

  const retryTimeoutRef = useRef<NodeJS.Timeout>()
  const maxPluginUpdates = 100 // Keep only last 100 updates

  // Handle system metrics updates
  const handleSystemMetrics = useCallback((metrics: SystemMetricsUpdate) => {
    setState(prev => ({
      ...prev,
      systemMetrics: metrics,
      lastUpdate: new Date(),
    }))
  }, [])

  // Handle plugin status updates
  const handlePluginUpdate = useCallback((update: PluginStatusUpdate) => {
    setState(prev => ({
      ...prev,
      pluginUpdates: [
        update,
        ...prev.pluginUpdates.slice(0, maxPluginUpdates - 1)
      ],
      lastUpdate: new Date(),
    }))
  }, [])

  // Handle connection status changes
  const handleConnectionChange = useCallback(() => {
    const isConnected = monitorWebSocket.isConnected
    setState(prev => ({
      ...prev,
      isConnected,
      connectionError: isConnected ? null : prev.connectionError,
    }))
  }, [])

  // Connect to WebSocket
  const connect = useCallback(async () => {
    try {
      setState(prev => ({ ...prev, connectionError: null }))
      await connectToMonitor()
      handleConnectionChange()
    } catch (error) {
      const errorMessage = error instanceof Error ? error.message : 'Connection failed'
      setState(prev => ({
        ...prev,
        connectionError: errorMessage,
        isConnected: false,
      }))
      throw error
    }
  }, [handleConnectionChange])

  // Disconnect from WebSocket
  const disconnect = useCallback(() => {
    if (retryTimeoutRef.current) {
      clearTimeout(retryTimeoutRef.current)
    }
    monitorWebSocket.disconnect()
    setState(prev => ({
      ...prev,
      isConnected: false,
      connectionError: null,
    }))
  }, [])

  // Retry connection with exponential backoff
  const retry = useCallback(async () => {
    if (retryTimeoutRef.current) {
      clearTimeout(retryTimeoutRef.current)
    }

    try {
      await connect()
    } catch (error) {
      // Schedule retry with exponential backoff
      const delay = Math.min(1000 * Math.pow(2, Math.random() * 3), 30000)
      retryTimeoutRef.current = setTimeout(retry, delay)
    }
  }, [connect])

  // Clear plugin updates
  const clearPluginUpdates = useCallback(() => {
    setState(prev => ({
      ...prev,
      pluginUpdates: [],
    }))
  }, [])

  // Setup WebSocket event listeners
  useEffect(() => {
    // Subscribe to system metrics
    monitorWebSocket.on('system_metrics', (message: MonitorMessage) => {
      handleSystemMetrics(message.data as SystemMetricsUpdate)
    })

    // Subscribe to plugin updates
    monitorWebSocket.on('plugin_status', (message: MonitorMessage) => {
      handlePluginUpdate(message.data as PluginStatusUpdate)
    })

    // Monitor connection state changes
    const checkConnection = () => {
      handleConnectionChange()
    }

    // Check connection state periodically
    const connectionCheckInterval = setInterval(checkConnection, 5000)

    return () => {
      clearInterval(connectionCheckInterval)
      if (retryTimeoutRef.current) {
        clearTimeout(retryTimeoutRef.current)
      }
    }
  }, [handleSystemMetrics, handlePluginUpdate, handleConnectionChange])

  // Auto-connect on mount
  useEffect(() => {
    if (autoConnect && !state.isConnected) {
      connect().catch(error => {
        console.error('Auto-connect failed:', error)
        // Start retry cycle
        retry()
      })
    }

    return () => {
      if (!autoConnect) {
        disconnect()
      }
    }
  }, [autoConnect, connect, disconnect, retry, state.isConnected])

  // Cleanup on unmount
  useEffect(() => {
    return () => {
      if (retryTimeoutRef.current) {
        clearTimeout(retryTimeoutRef.current)
      }
    }
  }, [])

  return {
    ...state,
    connect,
    disconnect,
    clearPluginUpdates,
    retry,
  }
}

export default useRealTimeMonitoring
