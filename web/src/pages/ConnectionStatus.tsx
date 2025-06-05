import { useState, useEffect } from 'react'
import { RefreshCw, Wifi, WifiOff, Play, TestTube, AlertCircle, CheckCircle } from 'lucide-react'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import { apiClient } from '@/services/api'

interface ConnectionInfo {
  status: 'connected' | 'connecting' | 'disconnected' | 'reconnecting' | 'failed'
  reconnect_attempts: number
  max_reconnects: number
  executor_host: string
  executor_port: number
  active_interpreter: string
  last_connected?: string
  uptime?: string
  process_pid?: number
  last_error?: string
}

export function ConnectionStatus() {
  const [connection, setConnection] = useState<ConnectionInfo | null>(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [reconnecting, setReconnecting] = useState(false)
  const [testing, setTesting] = useState(false)

  const loadConnectionStatus = async () => {
    try {
      setLoading(true)
      setError(null)
      const response = await apiClient.getConnectionStatus()
      if (response.success) {
        setConnection(response.data)
      } else {
        setError(response.error?.message || 'Failed to load connection status')
      }
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to load connection status')
    } finally {
      setLoading(false)
    }
  }

  const reconnectService = async (interpreterName?: string) => {
    try {
      setReconnecting(true)
      const response = await apiClient.reconnectService(interpreterName)

      if (!response.success) {
        setError(response.error?.message || 'Failed to reconnect service')
        setReconnecting(false)
        return
      }

      // Wait a moment then reload status
      setTimeout(() => {
        loadConnectionStatus()
        setReconnecting(false)
      }, 2000)
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to reconnect service')
      setReconnecting(false)
    }
  }

  const testConnection = async () => {
    try {
      setTesting(true)
      setError(null)
      const response = await apiClient.testConnection()

      if (!response.success) {
        setError(response.error?.message || 'Connection test failed')
      } else {
        console.log('Connection test result:', response.data)
        // Reload status after successful test
        await loadConnectionStatus()
      }
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Connection test failed')
    } finally {
      setTesting(false)
    }
  }

  useEffect(() => {
    loadConnectionStatus()
    
    // Auto-refresh every 30 seconds
    const interval = setInterval(loadConnectionStatus, 30000)
    return () => clearInterval(interval)
  }, [])

  const getStatusBadge = (status: string) => {
    switch (status) {
      case 'connected':
        return <Badge className="bg-green-100 text-green-800"><CheckCircle className="w-3 h-3 mr-1" />Connected</Badge>
      case 'connecting':
        return <Badge className="bg-blue-100 text-blue-800"><RefreshCw className="w-3 h-3 mr-1 animate-spin" />Connecting</Badge>
      case 'reconnecting':
        return <Badge className="bg-yellow-100 text-yellow-800"><RefreshCw className="w-3 h-3 mr-1 animate-spin" />Reconnecting</Badge>
      case 'failed':
        return <Badge className="bg-red-100 text-red-800"><AlertCircle className="w-3 h-3 mr-1" />Failed</Badge>
      case 'disconnected':
        return <Badge className="bg-gray-100 text-gray-800"><WifiOff className="w-3 h-3 mr-1" />Disconnected</Badge>
      default:
        return <Badge className="bg-gray-100 text-gray-800">Unknown</Badge>
    }
  }

  const getStatusIcon = (status: string) => {
    switch (status) {
      case 'connected':
        return <Wifi className="h-8 w-8 text-green-600" />
      case 'connecting':
      case 'reconnecting':
        return <RefreshCw className="h-8 w-8 text-blue-600 animate-spin" />
      case 'failed':
        return <AlertCircle className="h-8 w-8 text-red-600" />
      case 'disconnected':
        return <WifiOff className="h-8 w-8 text-gray-600" />
      default:
        return <WifiOff className="h-8 w-8 text-gray-600" />
    }
  }

  if (loading) {
    return (
      <div className="space-y-6">
        <div>
          <h1 className="text-3xl font-bold tracking-tight">Connection Status</h1>
          <p className="text-muted-foreground">Loading connection status...</p>
        </div>
      </div>
    )
  }

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-3xl font-bold tracking-tight">Connection Status</h1>
          <p className="text-muted-foreground">
            Monitor and manage the connection to Python executor service
          </p>
        </div>
        <div className="flex items-center space-x-2">
          <Button variant="outline" onClick={loadConnectionStatus} disabled={loading}>
            <RefreshCw className={`h-4 w-4 mr-2 ${loading ? 'animate-spin' : ''}`} />
            Refresh
          </Button>
          <Button 
            variant="outline" 
            onClick={testConnection} 
            disabled={testing || !connection || connection.status !== 'connected'}
          >
            <TestTube className={`h-4 w-4 mr-2 ${testing ? 'animate-spin' : ''}`} />
            {testing ? 'Testing...' : 'Test Connection'}
          </Button>
          <Button 
            onClick={() => reconnectService()} 
            disabled={reconnecting}
          >
            <Play className={`h-4 w-4 mr-2 ${reconnecting ? 'animate-spin' : ''}`} />
            {reconnecting ? 'Reconnecting...' : 'Reconnect'}
          </Button>
        </div>
      </div>

      {error && (
        <Card className="border-destructive">
          <CardContent className="pt-6">
            <div className="flex items-center space-x-2 text-destructive">
              <AlertCircle className="h-4 w-4" />
              <span className="text-sm">{error}</span>
            </div>
          </CardContent>
        </Card>
      )}

      {connection && (
        <div className="grid gap-6 md:grid-cols-2">
          {/* Connection Overview */}
          <Card className={connection.status === 'connected' ? 'border-green-200 bg-green-50' : 
                          connection.status === 'failed' ? 'border-red-200 bg-red-50' : ''}>
            <CardHeader>
              <div className="flex items-center space-x-3">
                {getStatusIcon(connection.status)}
                <div>
                  <CardTitle>Service Connection</CardTitle>
                  <CardDescription>Python executor service status</CardDescription>
                </div>
              </div>
            </CardHeader>
            <CardContent>
              <div className="space-y-3">
                <div className="flex items-center justify-between">
                  <span className="text-sm font-medium">Status:</span>
                  {getStatusBadge(connection.status)}
                </div>
                
                <div className="flex items-center justify-between">
                  <span className="text-sm font-medium">Executor Address:</span>
                  <span className="text-sm">{connection.executor_host}:{connection.executor_port}</span>
                </div>

                {connection.active_interpreter && (
                  <div className="flex items-center justify-between">
                    <span className="text-sm font-medium">Active Interpreter:</span>
                    <span className="text-sm">{connection.active_interpreter}</span>
                  </div>
                )}

                {connection.process_pid && (
                  <div className="flex items-center justify-between">
                    <span className="text-sm font-medium">Process PID:</span>
                    <span className="text-sm">{connection.process_pid}</span>
                  </div>
                )}

                {connection.uptime && (
                  <div className="flex items-center justify-between">
                    <span className="text-sm font-medium">Uptime:</span>
                    <span className="text-sm">{connection.uptime}</span>
                  </div>
                )}

                {connection.last_connected && (
                  <div className="flex items-center justify-between">
                    <span className="text-sm font-medium">Last Connected:</span>
                    <span className="text-sm">{new Date(connection.last_connected).toLocaleString()}</span>
                  </div>
                )}
              </div>
            </CardContent>
          </Card>

          {/* Connection Details */}
          <Card>
            <CardHeader>
              <CardTitle>Connection Details</CardTitle>
              <CardDescription>Detailed connection information and statistics</CardDescription>
            </CardHeader>
            <CardContent>
              <div className="space-y-3">
                <div className="flex items-center justify-between">
                  <span className="text-sm font-medium">Reconnect Attempts:</span>
                  <span className="text-sm">{connection.reconnect_attempts} / {connection.max_reconnects}</span>
                </div>

                {connection.reconnect_attempts > 0 && (
                  <div className="w-full bg-gray-200 rounded-full h-2">
                    <div 
                      className="bg-blue-600 h-2 rounded-full" 
                      style={{ width: `${(connection.reconnect_attempts / connection.max_reconnects) * 100}%` }}
                    ></div>
                  </div>
                )}

                {connection.last_error && (
                  <div className="p-3 bg-red-50 border border-red-200 rounded-md">
                    <div className="text-sm text-red-800">
                      <strong>Last Error:</strong> {connection.last_error}
                    </div>
                  </div>
                )}

                <div className="pt-4 space-y-2">
                  <Button 
                    variant="outline" 
                    className="w-full" 
                    onClick={() => reconnectService(connection.active_interpreter)}
                    disabled={reconnecting}
                  >
                    <RefreshCw className={`h-4 w-4 mr-2 ${reconnecting ? 'animate-spin' : ''}`} />
                    Reconnect with Current Interpreter
                  </Button>
                  
                  <Button 
                    variant="outline" 
                    className="w-full" 
                    onClick={() => reconnectService()}
                    disabled={reconnecting}
                  >
                    <RefreshCw className={`h-4 w-4 mr-2 ${reconnecting ? 'animate-spin' : ''}`} />
                    Force Reconnect
                  </Button>
                </div>
              </div>
            </CardContent>
          </Card>
        </div>
      )}

      {/* Connection Actions */}
      <Card>
        <CardHeader>
          <CardTitle>Connection Actions</CardTitle>
          <CardDescription>
            Manage the connection to the Python executor service
          </CardDescription>
        </CardHeader>
        <CardContent>
          <div className="grid gap-4 md:grid-cols-3">
            <Button 
              variant="outline" 
              onClick={testConnection}
              disabled={testing || !connection || connection.status !== 'connected'}
            >
              <TestTube className={`h-4 w-4 mr-2 ${testing ? 'animate-spin' : ''}`} />
              Test Connection
            </Button>
            
            <Button 
              variant="outline" 
              onClick={() => reconnectService()}
              disabled={reconnecting}
            >
              <RefreshCw className={`h-4 w-4 mr-2 ${reconnecting ? 'animate-spin' : ''}`} />
              Reconnect Service
            </Button>
            
            <Button 
              variant="outline" 
              onClick={loadConnectionStatus}
              disabled={loading}
            >
              <RefreshCw className={`h-4 w-4 mr-2 ${loading ? 'animate-spin' : ''}`} />
              Refresh Status
            </Button>
          </div>
        </CardContent>
      </Card>
    </div>
  )
}
