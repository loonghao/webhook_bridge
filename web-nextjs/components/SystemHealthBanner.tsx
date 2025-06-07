'use client'

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

export function SystemHealthBanner({ className, showDetails = false }: SystemHealthBannerProps) {
  const [health, setHealth] = useState<SystemHealth>({
    goServer: { status: 'unknown' },
    pythonExecutor: { status: 'unknown' },
    plugins: { total: 0, active: 0, errors: 0 }
  })
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  const fetchHealth = async () => {
    try {
      setLoading(true)
      setError(null)

      // Fetch system status and plugins in parallel
      const [status, plugins] = await Promise.allSettled([
        apiClient.getStatus(),
        apiClient.getPlugins()
      ])

      const systemStatus = status.status === 'fulfilled' ? status.value : null
      const pluginList = plugins.status === 'fulfilled' ? plugins.value : []

      setHealth({
        goServer: {
          status: systemStatus?.status || 'unknown',
          message: systemStatus?.service || 'Go server status unknown'
        },
        pythonExecutor: {
          status: systemStatus?.pythonVersion ? 'connected' : 'disconnected',
          message: systemStatus?.pythonVersion 
            ? `Python ${systemStatus.pythonVersion}` 
            : 'Python executor not available'
        },
        plugins: {
          total: pluginList.length,
          active: pluginList.filter(p => p.status === 'active').length,
          errors: pluginList.filter(p => p.status === 'error').length
        }
      })
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to fetch system health')
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => {
    fetchHealth()
    
    // Refresh health status every 30 seconds
    const interval = setInterval(fetchHealth, 30000)
    return () => clearInterval(interval)
  }, [])

  const getOverallStatus = (): 'healthy' | 'warning' | 'error' => {
    if (health.goServer.status === 'unhealthy' || health.pythonExecutor.status === 'disconnected') {
      return 'error'
    }
    if (health.plugins.errors > 0 || health.goServer.status === 'unknown') {
      return 'warning'
    }
    return 'healthy'
  }

  const overallStatus = getOverallStatus()

  const getStatusIcon = () => {
    switch (overallStatus) {
      case 'healthy':
        return <CheckCircle className="h-4 w-4 text-green-600" />
      case 'warning':
        return <AlertTriangle className="h-4 w-4 text-yellow-600" />
      case 'error':
        return <XCircle className="h-4 w-4 text-red-600" />
    }
  }

  const getStatusMessage = () => {
    if (error) return `System health check failed: ${error}`
    
    switch (overallStatus) {
      case 'healthy':
        return 'All systems operational'
      case 'warning':
        return `System operational with ${health.plugins.errors} plugin error(s)`
      case 'error':
        return 'System issues detected'
    }
  }

  const getAlertVariant = () => {
    switch (overallStatus) {
      case 'error':
        return 'destructive'
      default:
        return 'default'
    }
  }

  if (loading && !health.goServer.status) {
    return (
      <Alert className={className}>
        <RefreshCw className="h-4 w-4 animate-spin" />
        <AlertDescription>Checking system health...</AlertDescription>
      </Alert>
    )
  }

  return (
    <Alert variant={getAlertVariant()} className={className}>
      <div className="flex items-center justify-between">
        <div className="flex items-center space-x-2">
          {getStatusIcon()}
          <AlertDescription className="font-medium">
            {getStatusMessage()}
          </AlertDescription>
        </div>
        
        <div className="flex items-center space-x-2">
          {showDetails && (
            <div className="flex items-center space-x-4 text-sm">
              <div className="flex items-center space-x-1">
                <Badge variant={health.goServer.status === 'healthy' ? 'default' : 'destructive'}>
                  Go Server
                </Badge>
              </div>
              
              <div className="flex items-center space-x-1">
                {health.pythonExecutor.status === 'connected' ? (
                  <Wifi className="h-3 w-3 text-green-600" />
                ) : (
                  <WifiOff className="h-3 w-3 text-red-600" />
                )}
                <span className="text-muted-foreground">Python</span>
              </div>
              
              <div className="flex items-center space-x-1">
                <span className="text-muted-foreground">
                  {health.plugins.active}/{health.plugins.total} plugins
                </span>
              </div>
            </div>
          )}
          
          <Button
            variant="ghost"
            size="sm"
            onClick={fetchHealth}
            disabled={loading}
          >
            <RefreshCw className={`h-3 w-3 ${loading ? 'animate-spin' : ''}`} />
          </Button>
        </div>
      </div>
    </Alert>
  )
}
