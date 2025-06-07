'use client'

import { useState, useEffect } from 'react'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import { Layout } from '@/components/Layout'
import { apiClient } from '@/services/api'
import { monitorWebSocket } from '@/services/websocket'
import { checkBridgeHealth, getBridgeHealthScore, getBridgeConnectionSummary, BridgeStatus } from '@/lib/bridgeStatus'
import { CheckCircle, XCircle, AlertTriangle, RefreshCw, Wifi } from 'lucide-react'

interface TestResult {
  name: string
  status: 'success' | 'error' | 'pending'
  message: string
  data?: any
}

export default function DebugPage() {
  const [tests, setTests] = useState<TestResult[]>([])
  const [isRunning, setIsRunning] = useState(false)
  const [wsStatus, setWsStatus] = useState<string>('disconnected')
  const [bridgeStatus, setBridgeStatus] = useState<BridgeStatus | null>(null)
  const [healthScore, setHealthScore] = useState<number>(0)

  const testEndpoints = [
    { name: 'Health Check', endpoint: '/health', method: 'GET' },
    { name: 'Dashboard Status', endpoint: '/api/dashboard/status', method: 'GET' },
    { name: 'Dashboard Stats', endpoint: '/api/dashboard/stats', method: 'GET' },
    { name: 'Plugins List', endpoint: '/api/dashboard/plugins', method: 'GET' },
    { name: 'Workers List', endpoint: '/api/dashboard/workers', method: 'GET' },
    { name: 'Logs', endpoint: '/api/dashboard/logs', method: 'GET' },
  ]

  const runTests = async () => {
    setIsRunning(true)
    const results: TestResult[] = []

    for (const test of testEndpoints) {
      try {
        const response = await fetch(test.endpoint)
        const data = await response.json()
        
        results.push({
          name: test.name,
          status: response.ok ? 'success' : 'error',
          message: response.ok ? `${response.status} OK` : `${response.status} ${response.statusText}`,
          data: data
        })
      } catch (error) {
        results.push({
          name: test.name,
          status: 'error',
          message: error instanceof Error ? error.message : 'Unknown error',
        })
      }
    }

    setTests(results)
    setIsRunning(false)
  }

  const testApiClient = async () => {
    setIsRunning(true)
    const results: TestResult[] = []

    try {
      const stats = await apiClient.getStats()
      results.push({
        name: 'API Client - Stats',
        status: 'success',
        message: 'Successfully fetched stats',
        data: stats
      })
    } catch (error) {
      results.push({
        name: 'API Client - Stats',
        status: 'error',
        message: error instanceof Error ? error.message : 'Unknown error'
      })
    }

    try {
      const status = await apiClient.getStatus()
      results.push({
        name: 'API Client - Status',
        status: 'success',
        message: 'Successfully fetched status',
        data: status
      })
    } catch (error) {
      results.push({
        name: 'API Client - Status',
        status: 'error',
        message: error instanceof Error ? error.message : 'Unknown error'
      })
    }

    try {
      const plugins = await apiClient.getPlugins()
      results.push({
        name: 'API Client - Plugins',
        status: 'success',
        message: `Successfully fetched ${plugins.length} plugins`,
        data: plugins
      })
    } catch (error) {
      results.push({
        name: 'API Client - Plugins',
        status: 'error',
        message: error instanceof Error ? error.message : 'Unknown error'
      })
    }

    setTests(results)
    setIsRunning(false)
  }

  const testWebSocket = async () => {
    try {
      await monitorWebSocket.connect()
      setWsStatus('connected')
    } catch (error) {
      setWsStatus('error')
    }
  }

  const checkBridgeStatus = async () => {
    setIsRunning(true)
    try {
      const status = await checkBridgeHealth()
      setBridgeStatus(status)
      setHealthScore(getBridgeHealthScore())
    } catch (error) {
      console.error('Bridge status check failed:', error)
    } finally {
      setIsRunning(false)
    }
  }

  useEffect(() => {
    // Monitor WebSocket status
    const checkWsStatus = () => {
      setWsStatus(monitorWebSocket.connectionState)
    }

    const interval = setInterval(checkWsStatus, 1000)
    return () => clearInterval(interval)
  }, [])

  const getStatusIcon = (status: string) => {
    switch (status) {
      case 'success':
        return <CheckCircle className="h-4 w-4 text-green-600" />
      case 'error':
        return <XCircle className="h-4 w-4 text-red-600" />
      case 'pending':
        return <AlertTriangle className="h-4 w-4 text-yellow-600" />
      default:
        return <AlertTriangle className="h-4 w-4 text-gray-600" />
    }
  }

  const getStatusColor = (status: string) => {
    switch (status) {
      case 'success': return 'default'
      case 'error': return 'destructive'
      case 'pending': return 'secondary'
      default: return 'outline'
    }
  }

  return (
    <Layout>
      <div className="space-y-6">
        <div>
          <h1 className="text-3xl font-bold tracking-tight">Debug & Integration Test</h1>
          <p className="text-muted-foreground">
            Test API connectivity and data transformation
          </p>
        </div>

        {/* Environment Info */}
        <Card>
          <CardHeader>
            <CardTitle>Environment Configuration</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="grid gap-2 text-sm">
              <div className="flex justify-between">
                <span>API Base URL:</span>
                <code>{process.env.NEXT_PUBLIC_API_BASE_URL || 'relative'}</code>
              </div>
              <div className="flex justify-between">
                <span>WebSocket URL:</span>
                <code>{process.env.NEXT_PUBLIC_WS_BASE_URL || 'relative'}</code>
              </div>
              <div className="flex justify-between">
                <span>Development Mode:</span>
                <Badge variant={process.env.NEXT_PUBLIC_DEV_MODE === 'true' ? 'default' : 'outline'}>
                  {process.env.NEXT_PUBLIC_DEV_MODE || 'false'}
                </Badge>
              </div>
              <div className="flex justify-between">
                <span>WebSocket Status:</span>
                <Badge variant={wsStatus === 'connected' ? 'default' : 'destructive'}>
                  {wsStatus}
                </Badge>
              </div>
            </div>
          </CardContent>
        </Card>

        {/* Bridge Status */}
        {bridgeStatus && (
          <Card>
            <CardHeader>
              <CardTitle className="flex items-center justify-between">
                <span>Bridge Status</span>
                <div className="flex items-center space-x-2">
                  <Badge variant={healthScore > 80 ? 'default' : healthScore > 50 ? 'secondary' : 'destructive'}>
                    {healthScore}% Health
                  </Badge>
                  <Wifi className={`h-4 w-4 ${healthScore > 80 ? 'text-green-600' : 'text-red-600'}`} />
                </div>
              </CardTitle>
              <CardDescription>
                {getBridgeConnectionSummary()}
              </CardDescription>
            </CardHeader>
            <CardContent>
              <div className="grid gap-4 md:grid-cols-3">
                <div className="space-y-2">
                  <h4 className="font-medium">API Connection</h4>
                  <div className="flex items-center space-x-2">
                    {bridgeStatus.api.connected ?
                      <CheckCircle className="h-4 w-4 text-green-600" /> :
                      <XCircle className="h-4 w-4 text-red-600" />
                    }
                    <span className="text-sm">{bridgeStatus.api.connected ? 'Connected' : 'Disconnected'}</span>
                  </div>
                  <p className="text-xs text-muted-foreground">{bridgeStatus.api.baseUrl}</p>
                  {bridgeStatus.api.error && (
                    <p className="text-xs text-red-600">{bridgeStatus.api.error}</p>
                  )}
                </div>

                <div className="space-y-2">
                  <h4 className="font-medium">Go Server</h4>
                  <div className="flex items-center space-x-2">
                    {bridgeStatus.backend.goServer ?
                      <CheckCircle className="h-4 w-4 text-green-600" /> :
                      <XCircle className="h-4 w-4 text-red-600" />
                    }
                    <span className="text-sm">{bridgeStatus.backend.goServer ? 'Running' : 'Not Running'}</span>
                  </div>
                </div>

                <div className="space-y-2">
                  <h4 className="font-medium">Python Executor</h4>
                  <div className="flex items-center space-x-2">
                    {bridgeStatus.backend.pythonExecutor ?
                      <CheckCircle className="h-4 w-4 text-green-600" /> :
                      <XCircle className="h-4 w-4 text-red-600" />
                    }
                    <span className="text-sm">{bridgeStatus.backend.pythonExecutor ? 'Connected' : 'Disconnected'}</span>
                  </div>
                </div>
              </div>
            </CardContent>
          </Card>
        )}

        {/* Test Controls */}
        <div className="flex space-x-4">
          <Button onClick={checkBridgeStatus} disabled={isRunning}>
            {isRunning && <RefreshCw className="h-4 w-4 mr-2 animate-spin" />}
            Check Bridge Status
          </Button>
          <Button onClick={runTests} disabled={isRunning} variant="outline">
            {isRunning && <RefreshCw className="h-4 w-4 mr-2 animate-spin" />}
            Test Raw Endpoints
          </Button>
          <Button onClick={testApiClient} disabled={isRunning} variant="outline">
            {isRunning && <RefreshCw className="h-4 w-4 mr-2 animate-spin" />}
            Test API Client
          </Button>
          <Button onClick={testWebSocket} variant="outline">
            Test WebSocket
          </Button>
        </div>

        {/* Test Results */}
        {tests.length > 0 && (
          <Card>
            <CardHeader>
              <CardTitle>Test Results</CardTitle>
              <CardDescription>
                API connectivity and data transformation tests
              </CardDescription>
            </CardHeader>
            <CardContent>
              <div className="space-y-4">
                {tests.map((test, index) => (
                  <div key={index} className="flex items-start space-x-3 p-3 border rounded">
                    {getStatusIcon(test.status)}
                    <div className="flex-1">
                      <div className="flex items-center justify-between">
                        <h4 className="font-medium">{test.name}</h4>
                        <Badge variant={getStatusColor(test.status)}>
                          {test.status}
                        </Badge>
                      </div>
                      <p className="text-sm text-muted-foreground mt-1">
                        {test.message}
                      </p>
                      {test.data && (
                        <details className="mt-2">
                          <summary className="text-xs text-muted-foreground cursor-pointer">
                            View Response Data
                          </summary>
                          <pre className="text-xs bg-muted p-2 rounded mt-1 overflow-auto max-h-40">
                            {JSON.stringify(test.data, null, 2)}
                          </pre>
                        </details>
                      )}
                    </div>
                  </div>
                ))}
              </div>
            </CardContent>
          </Card>
        )}
      </div>
    </Layout>
  )
}
