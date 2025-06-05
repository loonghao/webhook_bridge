import { useState, useEffect } from 'react'
import {
  Play,
  RefreshCw,
  AlertCircle,
  CheckCircle,
  Clock,
  Activity,
  Code,
  Settings,
  BarChart3,
  FileText,
  AlertTriangle
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
import type { PluginInfo, PluginExecutionRequest, PluginExecutionResult, PluginStats, LogEntry } from '@/types/api'

export function Plugins() {
  const [plugins, setPlugins] = useState<PluginInfo[]>([])
  const [selectedPlugin, setSelectedPlugin] = useState<PluginInfo | null>(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [lastUpdated, setLastUpdated] = useState<Date | null>(null)
  
  // Plugin execution state
  const [executionMethod, setExecutionMethod] = useState<'GET' | 'POST' | 'PUT' | 'DELETE'>('POST')
  const [executionData, setExecutionData] = useState('{\n  "message": "Hello from dashboard!",\n  "timestamp": "' + new Date().toISOString() + '"\n}')
  const [executionResult, setExecutionResult] = useState<PluginExecutionResult | null>(null)
  const [executing, setExecuting] = useState(false)
  
  // Plugin stats and logs
  const [pluginStats, setPluginStats] = useState<PluginStats[]>([])
  const [pluginLogs, setPluginLogs] = useState<LogEntry[]>([])

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

  const fetchPluginStats = async () => {
    try {
      const stats = await apiClient.getPluginStats()
      setPluginStats(stats)
    } catch (err) {
      console.error('Failed to fetch plugin stats:', err)
    }
  }

  const fetchPluginLogs = async (pluginName: string) => {
    try {
      const logs = await apiClient.getPluginLogs(pluginName, { limit: 50 })
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

      // Refresh plugin data after execution
      await fetchPlugins()
      await fetchPluginStats()
      if (selectedPlugin) {
        await fetchPluginLogs(selectedPlugin.name)
      }
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

  useEffect(() => {
    fetchPlugins()
    fetchPluginStats()
  }, [])

  useEffect(() => {
    if (selectedPlugin) {
      fetchPluginLogs(selectedPlugin.name)
    }
  }, [selectedPlugin])

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
        <span className="ml-2">Loading plugins...</span>
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
          <h1 className="text-3xl font-bold">Plugins</h1>
          <p className="text-muted-foreground">
            Manage and test webhook plugins
            {lastUpdated && (
              <span className="ml-2">• Last updated: {formatTime(lastUpdated)}</span>
            )}
          </p>
        </div>
        <Button onClick={fetchPlugins} variant="outline">
          <RefreshCw className="h-4 w-4 mr-2" />
          Refresh
        </Button>
      </div>

      {/* Plugin Overview Cards */}
      <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-4">
        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">Total Plugins</CardTitle>
            <Code className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">{plugins.length}</div>
          </CardContent>
        </Card>
        
        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">Active Plugins</CardTitle>
            <CheckCircle className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">
              {plugins.filter(p => p.status === 'active').length}
            </div>
          </CardContent>
        </Card>
        
        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">Total Executions</CardTitle>
            <Activity className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">
              {plugins.reduce((sum, p) => sum + (p.executionCount || 0), 0)}
            </div>
          </CardContent>
        </Card>
        
        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">Error Rate</CardTitle>
            <AlertCircle className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">
              {plugins.length > 0 
                ? `${Math.round((plugins.reduce((sum, p) => sum + (p.errorCount || 0), 0) / plugins.reduce((sum, p) => sum + (p.executionCount || 0), 1)) * 100)}%`
                : '0%'
              }
            </div>
          </CardContent>
        </Card>
      </div>

      {/* Main Content */}
      <div className="grid gap-6 lg:grid-cols-3">
        {/* Plugin List */}
        <Card className="lg:col-span-1">
          <CardHeader>
            <CardTitle>Plugin List</CardTitle>
            <CardDescription>
              Select a plugin to view details and test
            </CardDescription>
          </CardHeader>
          <CardContent className="space-y-2">
            {plugins.map((plugin) => (
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
                  <span>{plugin.executionCount || 0} executions</span>
                </div>
              </div>
            ))}
          </CardContent>
        </Card>

        {/* Plugin Details and Testing */}
        <div className="lg:col-span-2">
          {selectedPlugin ? (
            <Tabs defaultValue="details" className="space-y-4">
              <TabsList>
                <TabsTrigger value="details">Details</TabsTrigger>
                <TabsTrigger value="test">Test</TabsTrigger>
                <TabsTrigger value="stats">Statistics</TabsTrigger>
                <TabsTrigger value="logs">Logs</TabsTrigger>
              </TabsList>

              <TabsContent value="details">
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
                  </CardContent>
                </Card>
              </TabsContent>

              <TabsContent value="test">
                <Card>
                  <CardHeader>
                    <CardTitle className="flex items-center space-x-2">
                      <Play className="h-5 w-5" />
                      <span>Test Plugin</span>
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

              <TabsContent value="stats">
                <Card>
                  <CardHeader>
                    <CardTitle className="flex items-center space-x-2">
                      <BarChart3 className="h-5 w-5" />
                      <span>Plugin Statistics</span>
                    </CardTitle>
                  </CardHeader>
                  <CardContent>
                    <div className="space-y-4">
                      {pluginStats
                        .filter(stat => stat.plugin === selectedPlugin.name)
                        .map((stat, index) => (
                          <div key={index} className="p-4 border rounded-lg">
                            <div className="flex items-center justify-between mb-2">
                              <span className="font-medium">{stat.method} Method</span>
                              <Badge variant="outline">{stat.count} executions</Badge>
                            </div>
                            <div className="grid gap-2 md:grid-cols-3 text-sm">
                              <div>
                                <span className="text-muted-foreground">Errors: </span>
                                <span className={stat.errors > 0 ? 'text-red-600' : 'text-green-600'}>
                                  {stat.errors}
                                </span>
                              </div>
                              <div>
                                <span className="text-muted-foreground">Avg Time: </span>
                                <span>{stat.avgTime}</span>
                              </div>
                              <div>
                                <span className="text-muted-foreground">Last Execution: </span>
                                <span>{new Date(stat.lastExecution).toLocaleString()}</span>
                              </div>
                            </div>
                          </div>
                        ))}
                      
                      {pluginStats.filter(stat => stat.plugin === selectedPlugin.name).length === 0 && (
                        <p className="text-muted-foreground text-center py-8">
                          No statistics available for this plugin
                        </p>
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
                  </CardHeader>
                  <CardContent>
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
                        <p className="text-muted-foreground text-center py-8">
                          No logs available for this plugin
                        </p>
                      )}
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
                    Select a plugin from the list to view details and test functionality
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
