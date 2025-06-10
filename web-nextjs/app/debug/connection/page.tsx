'use client'

import React from 'react'
import { ConnectionHealthMonitor } from '@/components/ConnectionHealthMonitor'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Alert, AlertDescription } from '@/components/ui/alert'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
import { 
  Activity, 
  AlertTriangle, 
  CheckCircle, 
  Settings, 
  Wifi,
  Server,
  Database,
  Globe
} from 'lucide-react'
import { useConnectionManager } from '@/lib/connection-manager'
import { useStagewise } from '@/hooks/useStagewise'

export default function ConnectionDebugPage() {
  const { state, isFullyConnected, healthScore, performFullCheck, runDiagnostics } = useConnectionManager()
  const stagewise = useStagewise()
  const [diagnostics, setDiagnostics] = React.useState<any>(null)
  const [isRunningDiagnostics, setIsRunningDiagnostics] = React.useState(false)

  const handleRunDiagnostics = async () => {
    setIsRunningDiagnostics(true)
    try {
      const result = await runDiagnostics()
      setDiagnostics(result)
    } catch (error) {
      console.error('Diagnostics failed:', error)
    } finally {
      setIsRunningDiagnostics(false)
    }
  }

  const getStatusColor = (status: string) => {
    switch (status) {
      case 'connected': return 'text-green-500'
      case 'connecting': return 'text-yellow-500'
      case 'disconnected': return 'text-gray-500'
      case 'error': return 'text-red-500'
      default: return 'text-gray-500'
    }
  }

  const getStatusIcon = (status: string) => {
    switch (status) {
      case 'connected': return <CheckCircle className="h-4 w-4 text-green-500" />
      case 'connecting': return <Activity className="h-4 w-4 text-yellow-500 animate-spin" />
      case 'disconnected': return <AlertTriangle className="h-4 w-4 text-gray-500" />
      case 'error': return <AlertTriangle className="h-4 w-4 text-red-500" />
      default: return <AlertTriangle className="h-4 w-4 text-gray-500" />
    }
  }

  React.useEffect(() => {
    // Start stagewise session for connection debugging
    stagewise.startSession('Connection Debug Session')
    const stageId = stagewise.startStage('connection-analysis', 'Analyzing connection health')

    return () => {
      if (stageId) {
        stagewise.endStage(stageId)
      }
      stagewise.endSession()
    }
  }, [stagewise])

  return (
    <div className="container mx-auto p-6 space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-3xl font-bold">Connection Debug</h1>
          <p className="text-muted-foreground">
            Monitor and optimize frontend-backend connections
          </p>
        </div>
        <div className="flex items-center space-x-2">
          <Badge variant={isFullyConnected ? 'default' : 'destructive'}>
            Health Score: {healthScore}%
          </Badge>
          <Button onClick={performFullCheck} variant="outline">
            <Activity className="h-4 w-4 mr-2" />
            Refresh
          </Button>
        </div>
      </div>

      <Tabs defaultValue="overview" className="w-full">
        <TabsList className="grid w-full grid-cols-4">
          <TabsTrigger value="overview">Overview</TabsTrigger>
          <TabsTrigger value="diagnostics">Diagnostics</TabsTrigger>
          <TabsTrigger value="optimization">Optimization</TabsTrigger>
          <TabsTrigger value="stagewise">Stagewise Data</TabsTrigger>
        </TabsList>

        <TabsContent value="overview" className="space-y-6">
          {/* Connection Status Cards */}
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
            {/* API Status */}
            <Card>
              <CardHeader className="pb-2">
                <div className="flex items-center justify-between">
                  <CardTitle className="text-sm font-medium">API Connection</CardTitle>
                  <Globe className="h-4 w-4 text-muted-foreground" />
                </div>
              </CardHeader>
              <CardContent>
                <div className="flex items-center space-x-2">
                  {getStatusIcon(state.api.status)}
                  <span className={`font-medium ${getStatusColor(state.api.status)}`}>
                    {state.api.status.toUpperCase()}
                  </span>
                </div>
                {state.api.latency && (
                  <p className="text-xs text-muted-foreground mt-1">
                    Latency: {state.api.latency}ms
                  </p>
                )}
                {state.api.error && (
                  <p className="text-xs text-red-500 mt-1">{state.api.error}</p>
                )}
              </CardContent>
            </Card>

            {/* Go Server Status */}
            <Card>
              <CardHeader className="pb-2">
                <div className="flex items-center justify-between">
                  <CardTitle className="text-sm font-medium">Go Server</CardTitle>
                  <Server className="h-4 w-4 text-muted-foreground" />
                </div>
              </CardHeader>
              <CardContent>
                <div className="flex items-center space-x-2">
                  {state.backend.goServer ? (
                    <CheckCircle className="h-4 w-4 text-green-500" />
                  ) : (
                    <AlertTriangle className="h-4 w-4 text-red-500" />
                  )}
                  <span className={`font-medium ${state.backend.goServer ? 'text-green-500' : 'text-red-500'}`}>
                    {state.backend.goServer ? 'RUNNING' : 'STOPPED'}
                  </span>
                </div>
              </CardContent>
            </Card>

            {/* Python Executor Status */}
            <Card>
              <CardHeader className="pb-2">
                <div className="flex items-center justify-between">
                  <CardTitle className="text-sm font-medium">Python Executor</CardTitle>
                  <Database className="h-4 w-4 text-muted-foreground" />
                </div>
              </CardHeader>
              <CardContent>
                <div className="flex items-center space-x-2">
                  {state.backend.pythonExecutor ? (
                    <CheckCircle className="h-4 w-4 text-green-500" />
                  ) : (
                    <AlertTriangle className="h-4 w-4 text-red-500" />
                  )}
                  <span className={`font-medium ${state.backend.pythonExecutor ? 'text-green-500' : 'text-red-500'}`}>
                    {state.backend.pythonExecutor ? 'CONNECTED' : 'DISCONNECTED'}
                  </span>
                </div>
              </CardContent>
            </Card>

            {/* WebSocket Status */}
            <Card>
              <CardHeader className="pb-2">
                <div className="flex items-center justify-between">
                  <CardTitle className="text-sm font-medium">WebSocket</CardTitle>
                  <Wifi className="h-4 w-4 text-muted-foreground" />
                </div>
              </CardHeader>
              <CardContent>
                <div className="space-y-1">
                  <div className="flex items-center space-x-2">
                    {getStatusIcon(state.websocket.monitor.status)}
                    <span className="text-xs">Monitor: {state.websocket.monitor.status}</span>
                  </div>
                  <div className="flex items-center space-x-2">
                    {getStatusIcon(state.websocket.logs.status)}
                    <span className="text-xs">Logs: {state.websocket.logs.status}</span>
                  </div>
                </div>
              </CardContent>
            </Card>
          </div>

          {/* Overall Health */}
          <Card>
            <CardHeader>
              <CardTitle>Overall Health</CardTitle>
            </CardHeader>
            <CardContent>
              <div className="space-y-4">
                <div className="flex items-center justify-between">
                  <span>Health Score</span>
                  <div className="flex items-center space-x-2">
                    <div className="w-32 bg-gray-200 rounded-full h-2">
                      <div 
                        className={`h-2 rounded-full ${healthScore >= 80 ? 'bg-green-500' : healthScore >= 60 ? 'bg-yellow-500' : 'bg-red-500'}`}
                        style={{ width: `${healthScore}%` }}
                      />
                    </div>
                    <span className="font-bold">{healthScore}%</span>
                  </div>
                </div>
                
                {!isFullyConnected && (
                  <Alert>
                    <AlertTriangle className="h-4 w-4" />
                    <AlertDescription>
                      Some connections are not fully established. This may affect functionality.
                    </AlertDescription>
                  </Alert>
                )}
              </div>
            </CardContent>
          </Card>
        </TabsContent>

        <TabsContent value="diagnostics" className="space-y-6">
          <Card>
            <CardHeader>
              <div className="flex items-center justify-between">
                <CardTitle>Connection Diagnostics</CardTitle>
                <Button 
                  onClick={handleRunDiagnostics}
                  disabled={isRunningDiagnostics}
                >
                  <Settings className="h-4 w-4 mr-2" />
                  {isRunningDiagnostics ? 'Running...' : 'Run Diagnostics'}
                </Button>
              </div>
            </CardHeader>
            <CardContent>
              {diagnostics ? (
                <div className="space-y-4">
                  <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
                    <div>
                      <h4 className="font-medium mb-2">API</h4>
                      <p className={`text-sm ${diagnostics.api.reachable ? 'text-green-600' : 'text-red-600'}`}>
                        {diagnostics.api.reachable ? '✓ Reachable' : '✗ Not reachable'}
                      </p>
                      {diagnostics.api.latency && (
                        <p className="text-sm text-muted-foreground">
                          Latency: {diagnostics.api.latency}ms
                        </p>
                      )}
                      {diagnostics.api.error && (
                        <p className="text-sm text-red-600">{diagnostics.api.error}</p>
                      )}
                    </div>
                    
                    <div>
                      <h4 className="font-medium mb-2">WebSocket</h4>
                      <p className={`text-sm ${diagnostics.websocket.monitor ? 'text-green-600' : 'text-red-600'}`}>
                        Monitor: {diagnostics.websocket.monitor ? '✓ Connected' : '✗ Disconnected'}
                      </p>
                      <p className={`text-sm ${diagnostics.websocket.logs ? 'text-green-600' : 'text-red-600'}`}>
                        Logs: {diagnostics.websocket.logs ? '✓ Connected' : '✗ Disconnected'}
                      </p>
                    </div>
                    
                    <div>
                      <h4 className="font-medium mb-2">Backend</h4>
                      <p className={`text-sm ${diagnostics.backend.go ? 'text-green-600' : 'text-red-600'}`}>
                        Go: {diagnostics.backend.go ? '✓ Running' : '✗ Stopped'}
                      </p>
                      <p className={`text-sm ${diagnostics.backend.python ? 'text-green-600' : 'text-red-600'}`}>
                        Python: {diagnostics.backend.python ? '✓ Connected' : '✗ Disconnected'}
                      </p>
                    </div>
                  </div>
                  
                  {diagnostics.recommendations.length > 0 && (
                    <div>
                      <h4 className="font-medium mb-2">Recommendations</h4>
                      <ul className="space-y-1">
                        {diagnostics.recommendations.map((rec: string, index: number) => (
                          <li key={index} className="text-sm text-yellow-600">
                            • {rec}
                          </li>
                        ))}
                      </ul>
                    </div>
                  )}
                </div>
              ) : (
                <p className="text-muted-foreground">Click &quot;Run Diagnostics&quot; to analyze connections</p>
              )}
            </CardContent>
          </Card>
        </TabsContent>

        <TabsContent value="optimization">
          <ConnectionHealthMonitor />
        </TabsContent>

        <TabsContent value="stagewise" className="space-y-6">
          <Card>
            <CardHeader>
              <CardTitle>Stagewise Network Data</CardTitle>
            </CardHeader>
            <CardContent>
              <div className="space-y-4">
                <div>
                  <h4 className="font-medium mb-2">Network Requests ({stagewise.networkRequests.length})</h4>
                  <div className="max-h-96 overflow-y-auto">
                    {stagewise.networkRequests.length > 0 ? (
                      <div className="space-y-2">
                        {stagewise.networkRequests.slice(-10).map((request, index) => (
                          <div key={index} className="border rounded p-2 text-sm">
                            <div className="flex items-center justify-between">
                              <span className="font-mono">{request.method} {request.url}</span>
                              <span className={`px-2 py-1 rounded text-xs ${
                                request.status && request.status >= 200 && request.status < 300 
                                  ? 'bg-green-100 text-green-800' 
                                  : 'bg-red-100 text-red-800'
                              }`}>
                                {request.status || 'Failed'}
                              </span>
                            </div>
                            {request.duration && (
                              <p className="text-muted-foreground">Duration: {request.duration}ms</p>
                            )}
                            {request.error && (
                              <p className="text-red-600">Error: {request.error}</p>
                            )}
                          </div>
                        ))}
                      </div>
                    ) : (
                      <p className="text-muted-foreground">No network requests captured yet</p>
                    )}
                  </div>
                </div>
              </div>
            </CardContent>
          </Card>
        </TabsContent>
      </Tabs>
    </div>
  )
}
