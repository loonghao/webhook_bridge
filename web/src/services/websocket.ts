export interface MonitorMessage {
  type: string
  timestamp: string
  data: any
}

export interface PluginStatusUpdate {
  plugin_name: string
  status: string
  last_executed?: string
  execution_time?: number
  success: boolean
  error?: string
}

export interface SystemMetricsUpdate {
  total_executions: number
  success_rate: number
  avg_execution_time: number
  active_plugins: number
  error_rate: number
  last_hour_executions: number
}

export type MonitorEventHandler = (message: MonitorMessage) => void

export class WebSocketService {
  private ws: WebSocket | null = null
  private reconnectAttempts = 0
  private maxReconnectAttempts = 5
  private reconnectDelay = 1000
  private handlers: Map<string, MonitorEventHandler[]> = new Map()
  private isConnecting = false
  private shouldReconnect = true

  constructor(private baseUrl: string = '') {
    if (!this.baseUrl) {
      this.baseUrl = `${window.location.protocol === 'https:' ? 'wss:' : 'ws:'}//${window.location.host}`
    }
  }

  connect(endpoint: string): Promise<void> {
    return new Promise((resolve, reject) => {
      if (this.isConnecting) {
        reject(new Error('Already connecting'))
        return
      }

      if (this.ws && this.ws.readyState === WebSocket.OPEN) {
        resolve()
        return
      }

      this.isConnecting = true
      const wsUrl = `${this.baseUrl}/api${endpoint}`
      
      try {
        this.ws = new WebSocket(wsUrl)
        
        this.ws.onopen = () => {
          console.log(`WebSocket connected to ${endpoint}`)
          this.isConnecting = false
          this.reconnectAttempts = 0
          this.emit('connected', { endpoint })
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
          console.log(`WebSocket disconnected from ${endpoint}:`, event.code, event.reason)
          this.isConnecting = false
          this.ws = null
          this.emit('disconnected', { endpoint, code: event.code, reason: event.reason })
          
          if (this.shouldReconnect && this.reconnectAttempts < this.maxReconnectAttempts) {
            this.scheduleReconnect(endpoint)
          }
        }

        this.ws.onerror = (error) => {
          console.error(`WebSocket error on ${endpoint}:`, error)
          this.isConnecting = false
          this.emit('error', { endpoint, error })
          reject(error)
        }
      } catch (error) {
        this.isConnecting = false
        reject(error)
      }
    })
  }

  private scheduleReconnect(endpoint: string) {
    this.reconnectAttempts++
    const delay = this.reconnectDelay * Math.pow(2, this.reconnectAttempts - 1)
    
    console.log(`Scheduling reconnect attempt ${this.reconnectAttempts} in ${delay}ms`)
    
    setTimeout(() => {
      if (this.shouldReconnect) {
        this.connect(endpoint).catch(error => {
          console.error('Reconnect failed:', error)
        })
      }
    }, delay)
  }

  private handleMessage(message: MonitorMessage) {
    const handlers = this.handlers.get(message.type) || []
    handlers.forEach(handler => {
      try {
        handler(message)
      } catch (error) {
        console.error('Error in message handler:', error)
      }
    })

    // Also emit to 'message' handlers
    this.emit('message', message)
  }

  on(event: string, handler: MonitorEventHandler) {
    if (!this.handlers.has(event)) {
      this.handlers.set(event, [])
    }
    this.handlers.get(event)!.push(handler)
  }

  off(event: string, handler: MonitorEventHandler) {
    const handlers = this.handlers.get(event)
    if (handlers) {
      const index = handlers.indexOf(handler)
      if (index > -1) {
        handlers.splice(index, 1)
      }
    }
  }

  private emit(event: string, data: any) {
    const handlers = this.handlers.get(event) || []
    handlers.forEach(handler => {
      try {
        handler({ type: event, timestamp: new Date().toISOString(), data })
      } catch (error) {
        console.error('Error in event handler:', error)
      }
    })
  }

  send(data: any) {
    if (this.ws && this.ws.readyState === WebSocket.OPEN) {
      this.ws.send(JSON.stringify(data))
    } else {
      console.warn('WebSocket is not connected')
    }
  }

  disconnect() {
    this.shouldReconnect = false
    if (this.ws) {
      this.ws.close()
      this.ws = null
    }
  }

  isConnected(): boolean {
    return this.ws !== null && this.ws.readyState === WebSocket.OPEN
  }

  getReadyState(): number | null {
    return this.ws ? this.ws.readyState : null
  }
}

// Singleton instances for different endpoints
export const logWebSocket = new WebSocketService()
export const monitorWebSocket = new WebSocketService()

// Helper functions for common operations
export const connectToLogs = () => logWebSocket.connect('/logs/stream')
export const connectToMonitor = () => monitorWebSocket.connect('/monitor/stream')

export const disconnectAll = () => {
  logWebSocket.disconnect()
  monitorWebSocket.disconnect()
}

// Auto-reconnect on page visibility change
document.addEventListener('visibilitychange', () => {
  if (!document.hidden) {
    // Page became visible, try to reconnect if needed
    if (!logWebSocket.isConnected()) {
      connectToLogs().catch(console.error)
    }
    if (!monitorWebSocket.isConnected()) {
      connectToMonitor().catch(console.error)
    }
  }
})

// Cleanup on page unload
window.addEventListener('beforeunload', () => {
  disconnectAll()
})
