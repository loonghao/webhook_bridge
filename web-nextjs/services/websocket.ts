import { MonitorMessage, PluginStatusUpdate, SystemMetricsUpdate } from '@/types/api'

export type MonitorEventHandler = (message: MonitorMessage) => void

class WebSocketManager {
  private ws: WebSocket | null = null
  private reconnectAttempts = 0
  private maxReconnectAttempts = 5
  private reconnectDelay = 1000
  private handlers: Map<string, MonitorEventHandler[]> = new Map()
  private isConnecting = false

  constructor(private url: string = '/api/dashboard/logs/stream') {
    // Update URL based on environment configuration
    if (typeof window !== 'undefined' && process.env.NEXT_PUBLIC_WS_BASE_URL) {
      this.url = `${process.env.NEXT_PUBLIC_WS_BASE_URL}/api/dashboard/logs/stream`
    }
  }

  connect(): Promise<void> {
    return new Promise((resolve, reject) => {
      if (this.ws?.readyState === WebSocket.OPEN) {
        resolve()
        return
      }

      if (this.isConnecting) {
        reject(new Error('Connection already in progress'))
        return
      }

      this.isConnecting = true

      try {
        // Convert relative URL to WebSocket URL
        let wsUrl: string
        if (this.url.startsWith('ws')) {
          wsUrl = this.url
        } else {
          // Use backend port (8080) instead of frontend port
          const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:'
          const hostname = window.location.hostname
          const backendPort = '8080' // Backend server port
          wsUrl = `${protocol}//${hostname}:${backendPort}${this.url}`
        }

        console.log('Connecting to WebSocket:', wsUrl)
        this.ws = new WebSocket(wsUrl)

        this.ws.onopen = () => {
          console.log('WebSocket connected')
          this.isConnecting = false
          this.reconnectAttempts = 0
          resolve()
        }

        this.ws.onmessage = (event) => {
          try {
            const message: MonitorMessage = JSON.parse(event.data)
            this.handleMessage(message)
          } catch (error) {
            console.error('Failed to parse WebSocket message:', error)
          }
        }

        this.ws.onclose = (event) => {
          console.log('WebSocket disconnected:', event.code, event.reason)
          this.isConnecting = false
          this.ws = null
          
          if (!event.wasClean && this.reconnectAttempts < this.maxReconnectAttempts) {
            this.scheduleReconnect()
          }
        }

        this.ws.onerror = (error) => {
          console.error('WebSocket error:', error)
          this.isConnecting = false
          reject(error)
        }

      } catch (error) {
        this.isConnecting = false
        reject(error)
      }
    })
  }

  disconnect(): void {
    if (this.ws) {
      this.ws.close(1000, 'Client disconnect')
      this.ws = null
    }
    this.reconnectAttempts = this.maxReconnectAttempts // Prevent reconnection
  }

  private scheduleReconnect(): void {
    this.reconnectAttempts++
    const delay = this.reconnectDelay * Math.pow(2, this.reconnectAttempts - 1)

    console.log(`Scheduling reconnect attempt ${this.reconnectAttempts} in ${delay}ms`)

    // Log reconnection attempt for stagewise
    if (typeof window !== 'undefined' && (window as any).__stagewise_log) {
      (window as any).__stagewise_log('websocket_reconnect_attempt', {
        url: this.url,
        attempt: this.reconnectAttempts,
        delay,
        timestamp: new Date().toISOString()
      })
    }

    setTimeout(() => {
      if (this.reconnectAttempts <= this.maxReconnectAttempts) {
        this.connect().catch(error => {
          console.error('Reconnection failed:', error)

          // Log reconnection failure for stagewise
          if (typeof window !== 'undefined' && (window as any).__stagewise_log) {
            (window as any).__stagewise_log('websocket_reconnect_failed', {
              url: this.url,
              attempt: this.reconnectAttempts,
              error: error instanceof Error ? error.message : String(error),
              timestamp: new Date().toISOString()
            })
          }
        })
      }
    }, delay)
  }

  private handleMessage(message: MonitorMessage): void {
    const handlers = this.handlers.get(message.type) || []
    handlers.forEach(handler => {
      try {
        handler(message)
      } catch (error) {
        console.error('Error in message handler:', error)
      }
    })

    // Also call handlers for 'all' type
    const allHandlers = this.handlers.get('all') || []
    allHandlers.forEach(handler => {
      try {
        handler(message)
      } catch (error) {
        console.error('Error in all-type handler:', error)
      }
    })
  }

  on(type: string, handler: MonitorEventHandler): void {
    if (!this.handlers.has(type)) {
      this.handlers.set(type, [])
    }
    this.handlers.get(type)!.push(handler)
  }

  off(type: string, handler: MonitorEventHandler): void {
    const handlers = this.handlers.get(type)
    if (handlers) {
      const index = handlers.indexOf(handler)
      if (index > -1) {
        handlers.splice(index, 1)
      }
    }
  }

  send(message: any): void {
    if (this.ws?.readyState === WebSocket.OPEN) {
      this.ws.send(JSON.stringify(message))
    } else {
      console.warn('WebSocket not connected, cannot send message')
    }
  }

  get isConnected(): boolean {
    return this.ws?.readyState === WebSocket.OPEN
  }

  get connectionState(): string {
    if (!this.ws) return 'disconnected'
    
    switch (this.ws.readyState) {
      case WebSocket.CONNECTING: return 'connecting'
      case WebSocket.OPEN: return 'connected'
      case WebSocket.CLOSING: return 'closing'
      case WebSocket.CLOSED: return 'disconnected'
      default: return 'unknown'
    }
  }
}

// Create singleton instances
export const monitorWebSocket = new WebSocketManager('/api/dashboard/monitor/stream')
export const logsWebSocket = new WebSocketManager('/api/dashboard/logs/stream')

// Store handler mappings for proper unsubscription
const handlerMappings = new WeakMap<Function, MonitorEventHandler>()

// Convenience functions
export function connectToMonitor(): Promise<void> {
  return monitorWebSocket.connect()
}

export function disconnectFromMonitor(): void {
  monitorWebSocket.disconnect()
}

export function connectToLogs(): Promise<void> {
  return logsWebSocket.connect()
}

export function disconnectFromLogs(): void {
  logsWebSocket.disconnect()
}

export function subscribeToPluginUpdates(handler: (update: PluginStatusUpdate) => void): void {
  const wrappedHandler: MonitorEventHandler = (message) => {
    handler(message.data as PluginStatusUpdate)
  }
  handlerMappings.set(handler, wrappedHandler)
  monitorWebSocket.on('plugin_status', wrappedHandler)
}

export function subscribeToSystemMetrics(handler: (metrics: SystemMetricsUpdate) => void): void {
  const wrappedHandler: MonitorEventHandler = (message) => {
    handler(message.data as SystemMetricsUpdate)
  }
  handlerMappings.set(handler, wrappedHandler)
  monitorWebSocket.on('system_metrics', wrappedHandler)
}

export function subscribeToLogs(handler: (log: any) => void): void {
  const wrappedHandler: MonitorEventHandler = (message) => {
    handler(message.data)
  }
  handlerMappings.set(handler, wrappedHandler)
  logsWebSocket.on('log_entry', wrappedHandler)
}

export function unsubscribeFromPluginUpdates(handler: (update: PluginStatusUpdate) => void): void {
  const wrappedHandler = handlerMappings.get(handler)
  if (wrappedHandler) {
    monitorWebSocket.off('plugin_status', wrappedHandler)
    handlerMappings.delete(handler)
  }
}

export function unsubscribeFromSystemMetrics(handler: (metrics: SystemMetricsUpdate) => void): void {
  const wrappedHandler = handlerMappings.get(handler)
  if (wrappedHandler) {
    monitorWebSocket.off('system_metrics', wrappedHandler)
    handlerMappings.delete(handler)
  }
}

export function unsubscribeFromLogs(handler: (log: any) => void): void {
  const wrappedHandler = handlerMappings.get(handler)
  if (wrappedHandler) {
    logsWebSocket.off('log_entry', wrappedHandler)
    handlerMappings.delete(handler)
  }
}

// Export the WebSocket manager for advanced usage
export default monitorWebSocket
