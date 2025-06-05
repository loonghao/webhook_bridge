import { useState, useEffect } from 'react'
import {
  Play,
  RefreshCw,
  AlertCircle,
  CheckCircle,
  Clock,
  Activity,

  Settings,
  BarChart3,
  FileText,
  AlertTriangle,
  TrendingUp,
  Zap,
  Timer,
  Target,
  Wifi,
  WifiOff
} from 'lucide-react'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import { Textarea } from '@/components/ui/textarea'
import { Label } from '@/components/ui/label'
import { Alert, AlertDescription } from '@/components/ui/alert'
import { SystemHealthBanner } from '@/components/SystemHealthBanner'
import { apiClient } from '@/services/api'
import { useRealTimeMonitoring, useSystemMetrics, usePluginUpdates } from '@/hooks/useRealTimeMonitoring'
import type { PluginInfo, PluginExecutionRequest, PluginExecutionResult, LogEntry } from '@/types/api'



interface PluginStatistics {
  plugin: string
  totalExecutions: number
  successfulExecutions: number
  failedExecutions: number
  avgExecutionTime: number
  lastExecution: string
  errorRate: number
}

export function PluginManager() {
  const [plugins, setPlugins] = useState<PluginInfo[]>([])
  const [selectedPlugin, setSelectedPlugin] = useState<PluginInfo | null>(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [lastUpdated, setLastUpdated] = useState<Date | null>(null)

  // Plugin execution state
  const [executionMethod, setExecutionMethod] = useState<'GET' | 'POST' | 'PUT' | 'DELETE'>('POST')
  const [executionData, setExecutionData] = useState('{\n  "message": "Hello from Plugin Manager!",\n  "timestamp": "' + new Date().toISOString() + '"\n}')
  const [executionResult, setExecutionResult] = useState<PluginExecutionResult | null>(null)
  const [executing, setExecuting] = useState(false)

  // Plugin metrics and statistics
  const [pluginStatistics, setPluginStatistics] = useState<PluginStatistics[]>([])
  const [pluginLogs, setPluginLogs] = useState<LogEntry[]>([])
  const [selectedLogLevel, setSelectedLogLevel] = useState<string>('all')

  // Real-time monitoring hooks
  const { metrics: realtimeMetrics, isConnected: metricsConnected, error: metricsError } = useSystemMetrics()
  const { updates: pluginUpdates, clearUpdates } = usePluginUpdates()
  const [, monitoringActions] = useRealTimeMonitoring()

  const fetchPlugins = async () => {
    try {
      setLoading(true)
      setError(null)
      const response = await apiClient.getPlugins()

      if (response.success) {
        const data = response.data || []
        setPlugins(data)
        setLastUpdated(new Date())

        // If no plugin is selected and we have plugins, select the first one
        if (!selectedPlugin && data.length > 0) {
          setSelectedPlugin(data[0])
        }
      } else {
        setError(response.error?.message || 'Failed to fetch plugins')
      }
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to fetch plugins')
    } finally {
      setLoading(false)
    }
  }



  const fetchPluginStatistics = async () => {
    try {
      const stats = await apiClient.getPluginStatistics()
      setPluginStatistics(stats)
    } catch (err) {
      console.error('Failed to fetch plugin statistics:', err)
    }
  }

  const fetchPluginLogs = async (pluginName?: string) => {
    try {
      const logs = await apiClient.getAllPluginLogs({ 
        limit: 100, 
        plugin: pluginName,
        level: selectedLogLevel === 'all' ? undefined : selectedLogLevel
      })
      setPluginLogs(logs)
    } catch (err) {
      console.error('Failed to fetch plugin logs:', err)
    }
  }

  const executePlugin = async () => {
    if (!selectedPlugin) return

    try {
      setExecuting(true)
      setExecutionResult(null)

      let data: Record<string, any> = {}
      if (executionData.trim()) {
        try {
          data = JSON.parse(executionData)
        } catch (err) {
          throw new Error('Invalid JSON in execution data: ' + (err instanceof Error ? err.message : 'Unknown JSON error'))
        }
      }

      const request: PluginExecutionRequest = {
        plugin: selectedPlugin.name,
        method: executionMethod,
        data
      }

      const result = await apiClient.executePlugin(request)
      setExecutionResult(result)

      // Refresh data after execution (real-time metrics will update automatically)
      await Promise.all([
        fetchPlugins(),
        fetchPluginStatistics(),
        fetchPluginLogs(selectedPlugin.name)
      ])
    } catch (err) {
      let errorMessage = 'Execution failed'
      let errorDetails = ''

      if (err instanceof Error) {
        errorMessage = err.message

        // Check for specific error types
        if (err.message.includes('503') || err.message.includes('Service Unavailable')) {
          errorMessage = 'Python executor is not available'
          errorDetails = 'The Python executor service is not running or not connected. Please check the system status.'
        } else if (err.message.includes('404')) {
          errorMessage = 'Plugin not found'
          errorDetails = `The plugin "${selectedPlugin.name}" was not found on the server.`
        } else if (err.message.includes('timeout')) {
          errorMessage = 'Execution timeout'
          errorDetails = 'The plugin execution took too long and was cancelled.'
        } else if (err.message.includes('JSON')) {
          errorDetails = 'Please check your JSON syntax and try again.'
        }
      }

      setExecutionResult({
        success: false,
        error: errorMessage,
        details: errorDetails,
        timestamp: new Date().toISOString()
      } as any)
    } finally {
      setExecuting(false)
    }
  }

  const testPluginConnection = async (pluginName: string) => {
    try {
      const result = await apiClient.testPluginConnection(pluginName)
      return result
    } catch (err) {
      console.error('Failed to test plugin connection:', err)
      return { success: false, error: 'Connection test failed' }
    }
  }

  useEffect(() => {
    const loadData = async () => {
      await Promise.all([
        fetchPlugins(),
        fetchPluginStatistics()
      ])
    }
    loadData()
  }, [])

  useEffect(() => {
    if (selectedPlugin) {
      fetchPluginLogs(selectedPlugin.name)
    } else {
      fetchPluginLogs()
    }
  }, [selectedPlugin, selectedLogLevel])

  const formatTime = (date: Date | null) => {
    if (!date) return 'Never'
    return date.toLocaleTimeString()
  }

  const getStatusIcon = (status: string) => {
    switch (status) {
      case 'active':
        return <CheckCircle className="h-4 w-4 text-green-500" />
      case 'inactive':
        return <Clock className="h-4 w-4 text-yellow-500" />
      case 'error':
        return <AlertCircle className="h-4 w-4 text-red-500" />
      default:
        return <AlertCircle className="h-4 w-4 text-gray-500" />
    }
  }

  const getStatusBadge = (status: string) => {
    const variants = {
      active: 'default',
      inactive: 'secondary',
      error: 'destructive'
    } as const
    
    return (
      <Badge variant={variants[status as keyof typeof variants] || 'secondary'}>
        {status}
      </Badge>
    )
  }

  if (loading) {
    return (
      <div className="flex items-center justify-center h-64">
        <RefreshCw className="h-8 w-8 animate-spin" />
        <span className="ml-2">Loading plugin manager...</span>
      </div>
    )
  }

  if (error) {
    return (
      <div className="flex items-center justify-center h-64">
        <AlertCircle className="h-8 w-8 text-red-500" />
        <span className="ml-2 text-red-500">{error}</span>
        <Button onClick={fetchPlugins} variant="outline" className="ml-4">
          <RefreshCw className="h-4 w-4 mr-2" />
          Retry
        </Button>
      </div>
    )
  }

  return (
    <div className="space-y-6">
      {/* System Health Banner */}
      <SystemHealthBanner />

      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-3xl font-bold">Plugin Manager</h1>
          <p className="text-muted-foreground">
            Advanced plugin management, monitoring and testing
            {lastUpdated && (
              <span className="ml-2">• Last updated: {formatTime(lastUpdated)}</span>
            )}
          </p>
        </div>
        <div className="flex space-x-2">
          <Button onClick={() => {
            fetchPlugins()
            fetchPluginStatistics()
            // Real-time metrics refresh automatically
          }} variant="outline">
            <RefreshCw className="h-4 w-4 mr-2" />
            Refresh Data
          </Button>
        </div>
      </div>

      {/* Real-time Connection Status */}
      {!metricsConnected && (
        <Alert variant="destructive">
          <WifiOff className="h-4 w-4" />
          <AlertDescription className="flex items-center justify-between">
            <span>Real-time monitoring disconnected. {metricsError}</span>
            <Button
              variant="outline"
              size="sm"
              onClick={monitoringActions.retry}
              className="ml-2"
            >
              Retry
            </Button>
          </AlertDescription>
        </Alert>
      )}

      {/* Plugin Metrics Overview */}
      {realtimeMetrics && (
        <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-6">
          <Card className={metricsConnected ? 'border-green-200' : 'border-gray-200'}>
            <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
              <CardTitle className="text-sm font-medium">Total Executions</CardTitle>
              <div className="flex items-center space-x-1">
                <Activity className="h-4 w-4 text-muted-foreground" />
                {metricsConnected && <Wifi className="h-3 w-3 text-green-500" />}
              </div>
            </CardHeader>
            <CardContent>
              <div className="text-2xl font-bold">{realtimeMetrics.total_executions}</div>
              <p className="text-xs text-muted-foreground">
                +{realtimeMetrics.last_hour_executions} in last hour
              </p>
            </CardContent>
          </Card>

          <Card>
            <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
              <CardTitle className="text-sm font-medium">Success Rate</CardTitle>
              <TrendingUp className="h-4 w-4 text-muted-foreground" />
            </CardHeader>
            <CardContent>
              <div className="text-2xl font-bold">{realtimeMetrics.success_rate.toFixed(1)}%</div>
              <p className="text-xs text-muted-foreground">
                {realtimeMetrics.success_rate >= 95 ? 'Excellent' : realtimeMetrics.success_rate >= 90 ? 'Good' : 'Needs attention'}
              </p>
            </CardContent>
          </Card>

          <Card>
            <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
              <CardTitle className="text-sm font-medium">Avg Response Time</CardTitle>
              <Timer className="h-4 w-4 text-muted-foreground" />
            </CardHeader>
            <CardContent>
              <div className="text-2xl font-bold">{realtimeMetrics.avg_execution_time.toFixed(0)}ms</div>
              <p className="text-xs text-muted-foreground">
                {realtimeMetrics.avg_execution_time < 100 ? 'Fast' : realtimeMetrics.avg_execution_time < 500 ? 'Normal' : 'Slow'}
              </p>
            </CardContent>
          </Card>

          <Card>
            <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
              <CardTitle className="text-sm font-medium">Active Plugins</CardTitle>
              <Zap className="h-4 w-4 text-muted-foreground" />
            </CardHeader>
            <CardContent>
              <div className="text-2xl font-bold">{realtimeMetrics.active_plugins}</div>
              <p className="text-xs text-muted-foreground">
                of {plugins.length} total
              </p>
            </CardContent>
          </Card>

          <Card>
            <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
              <CardTitle className="text-sm font-medium">Error Rate</CardTitle>
              <AlertTriangle className="h-4 w-4 text-muted-foreground" />
            </CardHeader>
            <CardContent>
              <div className="text-2xl font-bold">{realtimeMetrics.error_rate.toFixed(1)}%</div>
              <p className="text-xs text-muted-foreground">
                {realtimeMetrics.error_rate < 1 ? 'Excellent' : realtimeMetrics.error_rate < 5 ? 'Good' : 'High'}
              </p>
            </CardContent>
          </Card>

          <Card>
            <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
              <CardTitle className="text-sm font-medium">System Health</CardTitle>
              <Target className="h-4 w-4 text-muted-foreground" />
            </CardHeader>
            <CardContent>
              <div className="text-2xl font-bold">
                {realtimeMetrics.success_rate >= 95 && realtimeMetrics.error_rate < 1 ? '🟢' :
                 realtimeMetrics.success_rate >= 90 && realtimeMetrics.error_rate < 5 ? '🟡' : '🔴'}
              </div>
              <p className="text-xs text-muted-foreground">
                {realtimeMetrics.success_rate >= 95 && realtimeMetrics.error_rate < 1 ? 'Healthy' :
                 realtimeMetrics.success_rate >= 90 && realtimeMetrics.error_rate < 5 ? 'Warning' : 'Critical'}
              </p>
            </CardContent>
          </Card>
        </div>
      )}

      {/* Main Content */}
      <div className="grid gap-6 lg:grid-cols-3">
        {/* Plugin List */}
        <Card className="lg:col-span-1">
          <CardHeader>
            <CardTitle>Plugin List</CardTitle>
            <CardDescription>
              Select a plugin to manage and monitor
            </CardDescription>
          </CardHeader>
          <CardContent className="space-y-2">
            {plugins.map((plugin) => {
              const pluginStats = pluginStatistics.find(s => s.plugin === plugin.name)
              return (
                <div
                  key={plugin.name}
                  className={`p-3 rounded-lg border cursor-pointer transition-colors ${
                    selectedPlugin?.name === plugin.name
                      ? 'bg-accent border-accent-foreground'
                      : 'hover:bg-accent/50'
                  }`}
                  onClick={() => setSelectedPlugin(plugin)}
                >
                  <div className="flex items-center justify-between">
                    <div className="flex items-center space-x-2">
                      {getStatusIcon(plugin.status)}
                      <span className="font-medium">{plugin.name}</span>
                    </div>
                    {getStatusBadge(plugin.status)}
                  </div>
                  <p className="text-sm text-muted-foreground mt-1">
                    {plugin.description}
                  </p>
                  <div className="flex items-center justify-between text-xs text-muted-foreground mt-2">
                    <span>v{plugin.version}</span>
                    <span>{pluginStats?.totalExecutions || 0} executions</span>
                  </div>
                  {pluginStats && (
                    <div className="flex items-center justify-between text-xs text-muted-foreground mt-1">
                      <span>Success: {pluginStats.successfulExecutions}</span>
                      <span>Errors: {pluginStats.failedExecutions}</span>
                    </div>
                  )}
                </div>
              )
            })}
          </CardContent>
        </Card>

        {/* Plugin Management Interface */}
        <div className="lg:col-span-2">
          {selectedPlugin ? (
            <Tabs defaultValue="overview" className="space-y-4">
              <TabsList>
                <TabsTrigger value="overview">Overview</TabsTrigger>
                <TabsTrigger value="test">Test & Execute</TabsTrigger>
                <TabsTrigger value="statistics">Statistics</TabsTrigger>
                <TabsTrigger value="logs">Logs</TabsTrigger>
                <TabsTrigger value="monitoring">Monitoring</TabsTrigger>
              </TabsList>

              <TabsContent value="overview">
                <Card>
                  <CardHeader>
                    <CardTitle className="flex items-center space-x-2">
                      <span>{selectedPlugin.name}</span>
                      {getStatusBadge(selectedPlugin.status)}
                    </CardTitle>
                    <CardDescription>{selectedPlugin.description}</CardDescription>
                  </CardHeader>
                  <CardContent className="space-y-4">
                    <div className="grid gap-4 md:grid-cols-2">
                      <div>
                        <Label className="text-sm font-medium">Version</Label>
                        <p className="text-sm text-muted-foreground">{selectedPlugin.version}</p>
                      </div>
                      <div>
                        <Label className="text-sm font-medium">Status</Label>
                        <p className="text-sm text-muted-foreground">{selectedPlugin.status}</p>
                      </div>
                      <div>
                        <Label className="text-sm font-medium">Execution Count</Label>
                        <p className="text-sm text-muted-foreground">{selectedPlugin.executionCount || 0}</p>
                      </div>
                      <div>
                        <Label className="text-sm font-medium">Error Count</Label>
                        <p className="text-sm text-muted-foreground">{selectedPlugin.errorCount || 0}</p>
                      </div>
                      {selectedPlugin.avgExecutionTime && (
                        <div>
                          <Label className="text-sm font-medium">Avg Execution Time</Label>
                          <p className="text-sm text-muted-foreground">{selectedPlugin.avgExecutionTime}</p>
                        </div>
                      )}
                      {selectedPlugin.lastExecuted && (
                        <div>
                          <Label className="text-sm font-medium">Last Executed</Label>
                          <p className="text-sm text-muted-foreground">
                            {new Date(selectedPlugin.lastExecuted).toLocaleString()}
                          </p>
                        </div>
                      )}
                    </div>

                    {selectedPlugin.supportedMethods && (
                      <div>
                        <Label className="text-sm font-medium">Supported Methods</Label>
                        <div className="flex space-x-2 mt-1">
                          {selectedPlugin.supportedMethods.map((method) => (
                            <Badge key={method} variant="outline">
                              {method}
                            </Badge>
                          ))}
                        </div>
                      </div>
                    )}

                    {selectedPlugin.path && (
                      <div>
                        <Label className="text-sm font-medium">Path</Label>
                        <p className="text-sm text-muted-foreground font-mono">{selectedPlugin.path}</p>
                      </div>
                    )}

                    <div className="flex space-x-2 pt-4">
                      <Button
                        onClick={() => testPluginConnection(selectedPlugin.name)}
                        variant="outline"
                        size="sm"
                      >
                        <Zap className="h-4 w-4 mr-2" />
                        Test Connection
                      </Button>
                      <Button
                        onClick={() => fetchPluginLogs(selectedPlugin.name)}
                        variant="outline"
                        size="sm"
                      >
                        <RefreshCw className="h-4 w-4 mr-2" />
                        Refresh Data
                      </Button>
                    </div>
                  </CardContent>
                </Card>
              </TabsContent>

              <TabsContent value="test">
                <Card>
                  <CardHeader>
                    <CardTitle className="flex items-center space-x-2">
                      <Play className="h-5 w-5" />
                      <span>Test & Execute Plugin</span>
                    </CardTitle>
                    <CardDescription>
                      Execute the plugin with custom data to test its functionality
                    </CardDescription>
                  </CardHeader>
                  <CardContent className="space-y-4">
                    <div className="grid gap-4 md:grid-cols-2">
                      <div>
                        <Label htmlFor="method">HTTP Method</Label>
                        <Select value={executionMethod} onValueChange={(value: any) => setExecutionMethod(value)}>
                          <SelectTrigger>
                            <SelectValue />
                          </SelectTrigger>
                          <SelectContent>
                            <SelectItem value="GET">GET</SelectItem>
                            <SelectItem value="POST">POST</SelectItem>
                            <SelectItem value="PUT">PUT</SelectItem>
                            <SelectItem value="DELETE">DELETE</SelectItem>
                          </SelectContent>
                        </Select>
                      </div>
                    </div>

                    <div>
                      <Label htmlFor="data">Test Data (JSON)</Label>
                      <Textarea
                        id="data"
                        value={executionData}
                        onChange={(e) => setExecutionData(e.target.value)}
                        placeholder="Enter JSON data for the plugin..."
                        className="min-h-[120px] font-mono"
                      />
                    </div>

                    <Button
                      onClick={executePlugin}
                      disabled={executing}
                      className="w-full"
                    >
                      {executing ? (
                        <>
                          <RefreshCw className="h-4 w-4 mr-2 animate-spin" />
                          Executing...
                        </>
                      ) : (
                        <>
                          <Play className="h-4 w-4 mr-2" />
                          Execute Plugin
                        </>
                      )}
                    </Button>

                    {executionResult && (
                      <div className="mt-4">
                        <Label>Execution Result</Label>
                        <div className={`p-4 rounded-lg border mt-2 ${
                          executionResult.success ? 'bg-green-50 border-green-200' : 'bg-red-50 border-red-200'
                        }`}>
                          <div className="flex items-center space-x-2 mb-2">
                            {executionResult.success ? (
                              <CheckCircle className="h-4 w-4 text-green-600" />
                            ) : (
                              <AlertCircle className="h-4 w-4 text-red-600" />
                            )}
                            <span className={`font-medium ${
                              executionResult.success ? 'text-green-800' : 'text-red-800'
                            }`}>
                              {executionResult.success ? 'Success' : 'Error'}
                            </span>
                            <span className="text-sm text-muted-foreground">
                              {new Date(executionResult.timestamp).toLocaleString()}
                            </span>
                          </div>

                          {executionResult.error && (
                            <div className="mb-2">
                              <p className="text-red-700 font-medium">{executionResult.error}</p>
                              {(executionResult as any).details && (
                                <p className="text-red-600 text-sm mt-1">{(executionResult as any).details}</p>
                              )}
                            </div>
                          )}

                          {!executionResult.success && !executionResult.data && (
                            <Alert variant="destructive" className="mt-2">
                              <AlertTriangle className="h-4 w-4" />
                              <AlertDescription>
                                <div className="space-y-2">
                                  <p>Plugin execution failed. Common causes:</p>
                                  <ul className="list-disc list-inside text-sm space-y-1">
                                    <li>Python executor service is not running</li>
                                    <li>Plugin has syntax errors or missing dependencies</li>
                                    <li>Invalid input data format</li>
                                    <li>Network connectivity issues</li>
                                  </ul>
                                  <p className="text-sm">Check the system status and plugin logs for more details.</p>
                                </div>
                              </AlertDescription>
                            </Alert>
                          )}

                          {executionResult.data && (
                            <pre className="text-sm bg-white p-2 rounded border overflow-auto max-h-64">
                              {JSON.stringify(executionResult.data, null, 2)}
                            </pre>
                          )}

                          {executionResult.executionTime && (
                            <p className="text-sm text-muted-foreground mt-2">
                              Execution time: {executionResult.executionTime}ms
                            </p>
                          )}
                        </div>
                      </div>
                    )}
                  </CardContent>
                </Card>
              </TabsContent>

              <TabsContent value="statistics">
                <Card>
                  <CardHeader>
                    <CardTitle className="flex items-center space-x-2">
                      <BarChart3 className="h-5 w-5" />
                      <span>Plugin Statistics</span>
                    </CardTitle>
                    <CardDescription>
                      Detailed performance metrics and execution statistics
                    </CardDescription>
                  </CardHeader>
                  <CardContent>
                    <div className="space-y-4">
                      {pluginStatistics
                        .filter(stat => stat.plugin === selectedPlugin.name)
                        .map((stat, index) => (
                          <div key={index} className="p-4 border rounded-lg">
                            <div className="flex items-center justify-between mb-4">
                              <h4 className="font-medium">{stat.plugin} Performance</h4>
                              <Badge variant="outline">
                                {stat.errorRate.toFixed(1)}% error rate
                              </Badge>
                            </div>
                            <div className="grid gap-4 md:grid-cols-3">
                              <div>
                                <Label className="text-sm font-medium">Total Executions</Label>
                                <p className="text-2xl font-bold">{stat.totalExecutions}</p>
                              </div>
                              <div>
                                <Label className="text-sm font-medium">Successful</Label>
                                <p className="text-2xl font-bold text-green-600">{stat.successfulExecutions}</p>
                              </div>
                              <div>
                                <Label className="text-sm font-medium">Failed</Label>
                                <p className="text-2xl font-bold text-red-600">{stat.failedExecutions}</p>
                              </div>
                            </div>
                            <div className="grid gap-4 md:grid-cols-2 mt-4">
                              <div>
                                <Label className="text-sm font-medium">Avg Execution Time</Label>
                                <p className="text-lg font-semibold">{stat.avgExecutionTime}ms</p>
                              </div>
                              <div>
                                <Label className="text-sm font-medium">Last Execution</Label>
                                <p className="text-lg font-semibold">
                                  {new Date(stat.lastExecution).toLocaleString()}
                                </p>
                              </div>
                            </div>
                          </div>
                        ))}

                      {pluginStatistics.filter(stat => stat.plugin === selectedPlugin.name).length === 0 && (
                        <div className="text-center py-8">
                          <BarChart3 className="h-12 w-12 text-muted-foreground mx-auto mb-4" />
                          <h3 className="text-lg font-medium mb-2">No Statistics Available</h3>
                          <p className="text-muted-foreground">
                            Execute the plugin to generate performance statistics
                          </p>
                        </div>
                      )}
                    </div>
                  </CardContent>
                </Card>
              </TabsContent>

              <TabsContent value="logs">
                <Card>
                  <CardHeader>
                    <CardTitle className="flex items-center space-x-2">
                      <FileText className="h-5 w-5" />
                      <span>Plugin Logs</span>
                    </CardTitle>
                    <CardDescription>
                      View execution logs and debug information
                    </CardDescription>
                  </CardHeader>
                  <CardContent>
                    <div className="space-y-4">
                      <div className="flex items-center space-x-2">
                        <Label htmlFor="logLevel">Filter by Level</Label>
                        <Select value={selectedLogLevel} onValueChange={setSelectedLogLevel}>
                          <SelectTrigger className="w-32">
                            <SelectValue />
                          </SelectTrigger>
                          <SelectContent>
                            <SelectItem value="all">All</SelectItem>
                            <SelectItem value="info">Info</SelectItem>
                            <SelectItem value="warn">Warning</SelectItem>
                            <SelectItem value="error">Error</SelectItem>
                            <SelectItem value="debug">Debug</SelectItem>
                          </SelectContent>
                        </Select>
                        <Button
                          onClick={() => fetchPluginLogs(selectedPlugin.name)}
                          variant="outline"
                          size="sm"
                        >
                          <RefreshCw className="h-4 w-4 mr-2" />
                          Refresh
                        </Button>
                      </div>

                      <div className="space-y-2 max-h-96 overflow-auto">
                        {pluginLogs.map((log, index) => (
                          <div key={index} className="p-3 border rounded-lg">
                            <div className="flex items-center justify-between mb-1">
                              <Badge variant={
                                log.level === 'error' ? 'destructive' :
                                log.level === 'warn' ? 'secondary' : 'default'
                              }>
                                {log.level.toUpperCase()}
                              </Badge>
                              <span className="text-sm text-muted-foreground">
                                {new Date(log.timestamp).toLocaleString()}
                              </span>
                            </div>
                            <p className="text-sm">{log.message}</p>
                            {log.source && (
                              <p className="text-xs text-muted-foreground mt-1">
                                Source: {log.source}
                              </p>
                            )}
                          </div>
                        ))}

                        {pluginLogs.length === 0 && (
                          <div className="text-center py-8">
                            <FileText className="h-12 w-12 text-muted-foreground mx-auto mb-4" />
                            <h3 className="text-lg font-medium mb-2">No Logs Available</h3>
                            <p className="text-muted-foreground">
                              No logs found for this plugin with the selected filter
                            </p>
                          </div>
                        )}
                      </div>
                    </div>
                  </CardContent>
                </Card>
              </TabsContent>

              <TabsContent value="monitoring">
                <Card>
                  <CardHeader>
                    <CardTitle className="flex items-center space-x-2">
                      <Activity className="h-5 w-5" />
                      <span>Real-time Monitoring</span>
                    </CardTitle>
                    <CardDescription>
                      Monitor plugin performance and health in real-time
                    </CardDescription>
                  </CardHeader>
                  <CardContent>
                    <div className="space-y-6">
                      {/* Plugin Health Status */}
                      <div className="grid gap-4 md:grid-cols-2">
                        <div className="p-4 border rounded-lg">
                          <div className="flex items-center space-x-2 mb-2">
                            <div className={`w-3 h-3 rounded-full ${
                              selectedPlugin.status === 'active' ? 'bg-green-500' :
                              selectedPlugin.status === 'inactive' ? 'bg-yellow-500' : 'bg-red-500'
                            }`} />
                            <Label className="font-medium">Plugin Status</Label>
                          </div>
                          <p className="text-2xl font-bold capitalize">{selectedPlugin.status}</p>
                        </div>

                        <div className="p-4 border rounded-lg">
                          <Label className="font-medium">Last Activity</Label>
                          <p className="text-2xl font-bold">
                            {selectedPlugin.lastExecuted
                              ? new Date(selectedPlugin.lastExecuted).toLocaleTimeString()
                              : 'Never'
                            }
                          </p>
                        </div>
                      </div>

                      {/* Quick Actions */}
                      <div>
                        <Label className="font-medium mb-2 block">Quick Actions</Label>
                        <div className="flex space-x-2">
                          <Button
                            onClick={() => testPluginConnection(selectedPlugin.name)}
                            variant="outline"
                            size="sm"
                          >
                            <Zap className="h-4 w-4 mr-2" />
                            Test Connection
                          </Button>
                          <Button
                            onClick={() => fetchPluginLogs(selectedPlugin.name)}
                            variant="outline"
                            size="sm"
                          >
                            <RefreshCw className="h-4 w-4 mr-2" />
                            Refresh Logs
                          </Button>
                          <Button
                            onClick={() => fetchPluginStatistics()}
                            variant="outline"
                            size="sm"
                          >
                            <BarChart3 className="h-4 w-4 mr-2" />
                            Update Stats
                          </Button>
                        </div>
                      </div>

                      {/* Performance Indicators */}
                      <div>
                        <Label className="font-medium mb-2 block">Performance Indicators</Label>
                        <div className="grid gap-4 md:grid-cols-3">
                          <div className="p-3 border rounded">
                            <div className="text-sm text-muted-foreground">Execution Count</div>
                            <div className="text-xl font-bold">{selectedPlugin.executionCount || 0}</div>
                          </div>
                          <div className="p-3 border rounded">
                            <div className="text-sm text-muted-foreground">Error Count</div>
                            <div className="text-xl font-bold text-red-600">{selectedPlugin.errorCount || 0}</div>
                          </div>
                          <div className="p-3 border rounded">
                            <div className="text-sm text-muted-foreground">Avg Response Time</div>
                            <div className="text-xl font-bold">{selectedPlugin.avgExecutionTime || 'N/A'}</div>
                          </div>
                        </div>
                      </div>

                      {/* Real-time Plugin Updates */}
                      <div>
                        <div className="flex items-center justify-between mb-2">
                          <Label className="font-medium">Recent Activity</Label>
                          <div className="flex items-center space-x-2">
                            <Badge variant={metricsConnected ? 'default' : 'secondary'}>
                              {metricsConnected ? 'Live' : 'Offline'}
                            </Badge>
                            <Button
                              onClick={clearUpdates}
                              variant="outline"
                              size="sm"
                            >
                              Clear
                            </Button>
                          </div>
                        </div>
                        <div className="space-y-2 max-h-64 overflow-auto">
                          {pluginUpdates
                            .filter(update => update.plugin_name === selectedPlugin.name)
                            .slice(0, 10)
                            .map((update, index) => (
                              <div key={index} className="p-3 border rounded-lg">
                                <div className="flex items-center justify-between mb-1">
                                  <div className="flex items-center space-x-2">
                                    {update.success ? (
                                      <CheckCircle className="h-4 w-4 text-green-500" />
                                    ) : (
                                      <AlertCircle className="h-4 w-4 text-red-500" />
                                    )}
                                    <Badge variant={update.success ? 'default' : 'destructive'}>
                                      {update.status}
                                    </Badge>
                                  </div>
                                  <span className="text-sm text-muted-foreground">
                                    {update.last_executed && new Date(update.last_executed).toLocaleTimeString()}
                                  </span>
                                </div>
                                <div className="flex items-center justify-between text-sm">
                                  <span>Execution time: {update.execution_time}ms</span>
                                  {update.error && (
                                    <span className="text-red-600 text-xs truncate max-w-48">
                                      {update.error}
                                    </span>
                                  )}
                                </div>
                              </div>
                            ))}

                          {pluginUpdates.filter(update => update.plugin_name === selectedPlugin.name).length === 0 && (
                            <div className="text-center py-8">
                              <Activity className="h-12 w-12 text-muted-foreground mx-auto mb-4" />
                              <h3 className="text-lg font-medium mb-2">No Recent Activity</h3>
                              <p className="text-muted-foreground">
                                Execute the plugin to see real-time updates here
                              </p>
                            </div>
                          )}
                        </div>
                      </div>
                    </div>
                  </CardContent>
                </Card>
              </TabsContent>
            </Tabs>
          ) : (
            <Card>
              <CardContent className="flex items-center justify-center h-64">
                <div className="text-center">
                  <Settings className="h-12 w-12 text-muted-foreground mx-auto mb-4" />
                  <h3 className="text-lg font-medium mb-2">No Plugin Selected</h3>
                  <p className="text-muted-foreground">
                    Select a plugin from the list to access advanced management features
                  </p>
                </div>
              </CardContent>
            </Card>
          )}
        </div>
      </div>
    </div>
  )
}
