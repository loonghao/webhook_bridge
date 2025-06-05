import { useState, useEffect } from 'react'
import { AlertTriangle, CheckCircle, XCircle, RefreshCw, Wifi, WifiOff } from 'lucide-react'
import { Alert, AlertDescription } from '@/components/ui/alert'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import { apiClient } from '@/services/api'

interface SystemHealth {
  goServer: {
    status: 'healthy' | 'unhealthy' | 'unknown'
    message?: string
  }
  pythonExecutor: {
    status: 'connected' | 'disconnected' | 'unknown'
    message?: string
    lastConnected?: string
  }
  plugins: {
    total: number
    active: number
    errors: number
  }
}

interface SystemHealthBannerProps {
  className?: string
  showDetails?: boolean
}

export function SystemHealthBanner({ className = '', showDetails = false }: SystemHealthBannerProps) {
  const [health, setHealth] = useState<SystemHealth | null>(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [lastCheck, setLastCheck] = useState<Date | null>(null)

  const checkSystemHealth = async () => {
    try {
      setLoading(true)
      setError(null)

      // Check system status and connection
      const [statusResponse, connectionResponse, pluginsResponse] = await Promise.allSettled([
        apiClient.getStatus(),
        apiClient.get('/api/dashboard/connection'),
        apiClient.getPlugins()
      ])

      const systemHealth: SystemHealth = {
        goServer: {
          status: 'unknown',
          message: 'Unable to determine server status'
        },
        pythonExecutor: {
          status: 'unknown',
          message: 'Unable to determine executor status'
        },
        plugins: {
          total: 0,
          active: 0,
          errors: 0
        }
      }

      // Process status response
      if (statusResponse.status === 'fulfilled') {
        const status = statusResponse.value
        systemHealth.goServer = {
          status: status.status === 'healthy' ? 'healthy' : 'unhealthy',
          message: (status as any).message || 'Server is running'
        }
      } else {
        systemHealth.goServer = {
          status: 'unhealthy',
          message: 'Failed to connect to server'
        }
      }

      // Process connection response
      if (connectionResponse.status === 'fulfilled') {
        const connection = connectionResponse.value.data
        systemHealth.pythonExecutor = {
          status: connection.status === 'connected' ? 'connected' : 'disconnected',
          message: connection.last_error || (connection.status === 'connected' ? 'Executor is running' : 'Executor is not connected'),
          lastConnected: connection.last_connected
        }
      } else {
        systemHealth.pythonExecutor = {
          status: 'disconnected',
          message: 'Failed to check executor status'
        }
      }

      // Process plugins response
      if (pluginsResponse.status === 'fulfilled') {
        const pluginsApiResponse = pluginsResponse.value
        const plugins = pluginsApiResponse.success ? (pluginsApiResponse.data || []) : []
        systemHealth.plugins = {
          total: plugins.length,
          active: plugins.filter((p: any) => p.status === 'active').length,
          errors: plugins.filter((p: any) => p.status === 'error').length
        }
      }

      setHealth(systemHealth)
      setLastCheck(new Date())
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to check system health')
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => {
    checkSystemHealth()
    
    // Auto-refresh every 30 seconds
    const interval = setInterval(checkSystemHealth, 30000)
    return () => clearInterval(interval)
  }, [])

  const getOverallStatus = (): 'healthy' | 'warning' | 'critical' => {
    if (!health) return 'critical'
    
    if (health.goServer.status === 'unhealthy') return 'critical'
    if (health.pythonExecutor.status === 'disconnected') return 'warning'
    if (health.plugins.errors > 0) return 'warning'
    
    return 'healthy'
  }

  const getStatusIcon = (status: string) => {
    switch (status) {
      case 'healthy':
      case 'connected':
        return <CheckCircle className="h-4 w-4 text-green-500" />
      case 'warning':
        return <AlertTriangle className="h-4 w-4 text-yellow-500" />
      case 'critical':
      case 'unhealthy':
      case 'disconnected':
        return <XCircle className="h-4 w-4 text-red-500" />
      default:
        return <RefreshCw className="h-4 w-4 text-gray-500" />
    }
  }

  const getStatusBadge = (status: string) => {
    const variants = {
      healthy: 'default',
      connected: 'default',
      warning: 'secondary',
      critical: 'destructive',
      unhealthy: 'destructive',
      disconnected: 'destructive',
      unknown: 'secondary'
    } as const
    
    return (
      <Badge variant={variants[status as keyof typeof variants] || 'secondary'}>
        {status}
      </Badge>
    )
  }

  if (loading && !health) {
    return (
      <Alert className={className}>
        <RefreshCw className="h-4 w-4 animate-spin" />
        <AlertDescription>
          Checking system health...
        </AlertDescription>
      </Alert>
    )
  }

  if (error && !health) {
    return (
      <Alert variant="destructive" className={className}>
        <XCircle className="h-4 w-4" />
        <AlertDescription className="flex items-center justify-between">
          <span>System health check failed: {error}</span>
          <Button onClick={checkSystemHealth} variant="outline" size="sm">
            <RefreshCw className="h-4 w-4 mr-2" />
            Retry
          </Button>
        </AlertDescription>
      </Alert>
    )
  }

  if (!health) return null

  const overallStatus = getOverallStatus()

  return (
    <div className={`space-y-2 ${className}`}>
      {/* Main status banner */}
      <Alert variant={overallStatus === 'critical' ? 'destructive' : overallStatus === 'warning' ? 'default' : 'default'}>
        {getStatusIcon(overallStatus)}
        <AlertDescription className="flex items-center justify-between">
          <div className="flex items-center space-x-4">
            <span>
              System Status: {overallStatus === 'healthy' ? 'All systems operational' : 
                            overallStatus === 'warning' ? 'Some issues detected' : 
                            'Critical issues detected'}
            </span>
            {lastCheck && (
              <span className="text-xs text-muted-foreground">
                Last checked: {lastCheck.toLocaleTimeString()}
              </span>
            )}
          </div>
          <Button onClick={checkSystemHealth} variant="outline" size="sm" disabled={loading}>
            <RefreshCw className={`h-4 w-4 mr-2 ${loading ? 'animate-spin' : ''}`} />
            Refresh
          </Button>
        </AlertDescription>
      </Alert>

      {/* Detailed status (if requested or there are issues) */}
      {(showDetails || overallStatus !== 'healthy') && (
        <div className="grid gap-2 md:grid-cols-3">
          {/* Go Server Status */}
          <div className="flex items-center justify-between p-3 border rounded-lg">
            <div className="flex items-center space-x-2">
              {getStatusIcon(health.goServer.status)}
              <span className="text-sm font-medium">Go Server</span>
            </div>
            {getStatusBadge(health.goServer.status)}
          </div>

          {/* Python Executor Status */}
          <div className="flex items-center justify-between p-3 border rounded-lg">
            <div className="flex items-center space-x-2">
              {health.pythonExecutor.status === 'connected' ? 
                <Wifi className="h-4 w-4 text-green-500" /> : 
                <WifiOff className="h-4 w-4 text-red-500" />
              }
              <span className="text-sm font-medium">Python Executor</span>
            </div>
            {getStatusBadge(health.pythonExecutor.status)}
          </div>

          {/* Plugins Status */}
          <div className="flex items-center justify-between p-3 border rounded-lg">
            <div className="flex items-center space-x-2">
              {health.plugins.errors > 0 ? 
                <AlertTriangle className="h-4 w-4 text-yellow-500" /> : 
                <CheckCircle className="h-4 w-4 text-green-500" />
              }
              <span className="text-sm font-medium">
                Plugins ({health.plugins.active}/{health.plugins.total})
              </span>
            </div>
            {health.plugins.errors > 0 ? 
              <Badge variant="destructive">{health.plugins.errors} errors</Badge> :
              <Badge variant="default">OK</Badge>
            }
          </div>
        </div>
      )}

      {/* Error details */}
      {overallStatus !== 'healthy' && (
        <div className="text-sm text-muted-foreground space-y-1">
          {health.goServer.status === 'unhealthy' && (
            <div>• Go Server: {health.goServer.message}</div>
          )}
          {health.pythonExecutor.status === 'disconnected' && (
            <div>• Python Executor: {health.pythonExecutor.message}</div>
          )}
          {health.plugins.errors > 0 && (
            <div>• {health.plugins.errors} plugin(s) have errors</div>
          )}
        </div>
      )}
    </div>
  )
}
