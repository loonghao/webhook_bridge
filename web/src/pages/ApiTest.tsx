import { useState } from 'react'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import { Alert, AlertDescription } from '@/components/ui/alert'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
import { CheckCircle, AlertCircle, RefreshCw, Activity, Server, Database } from 'lucide-react'
import { apiClient } from '@/services/api'

export function ApiTest() {
  const [loading, setLoading] = useState(false)
  const [results, setResults] = useState<Record<string, any>>({})
  const [errors, setErrors] = useState<Record<string, string>>({})

  const testEndpoint = async (name: string, testFn: () => Promise<any>) => {
    try {
      setLoading(true)
      setErrors(prev => ({ ...prev, [name]: '' }))
      
      const result = await testFn()
      setResults(prev => ({ ...prev, [name]: result }))
    } catch (err) {
      const errorMessage = err instanceof Error ? err.message : 'Unknown error'
      setErrors(prev => ({ ...prev, [name]: errorMessage }))
    } finally {
      setLoading(false)
    }
  }

  const testAllEndpoints = async () => {
    const tests = [
      { name: 'health', fn: () => fetch('http://localhost:8001/health').then(r => r.json()) },
      { name: 'api-info', fn: () => fetch('http://localhost:8001/api').then(r => r.json()) },
      { name: 'system-status', fn: () => apiClient.getSystemStatus() },
      { name: 'system-info', fn: () => apiClient.getSystemInfo() },
      { name: 'connection-status', fn: () => apiClient.getConnectionStatus() },
      { name: 'config', fn: () => apiClient.getConfig() },
      { name: 'workers', fn: () => apiClient.getWorkers() },
      { name: 'stats', fn: () => apiClient.getStats() },
      { name: 'metrics', fn: () => apiClient.getMetrics() },
      { name: 'plugins', fn: () => apiClient.getPlugins() },
    ]

    for (const test of tests) {
      await testEndpoint(test.name, test.fn)
      // Small delay between requests
      await new Promise(resolve => setTimeout(resolve, 100))
    }
  }

  const getStatusIcon = (name: string) => {
    if (errors[name]) {
      return <AlertCircle className="h-4 w-4 text-red-500" />
    }
    if (results[name]) {
      return <CheckCircle className="h-4 w-4 text-green-500" />
    }
    return <RefreshCw className="h-4 w-4 text-gray-400" />
  }

  const getStatusBadge = (name: string) => {
    if (errors[name]) {
      return <Badge variant="destructive">Error</Badge>
    }
    if (results[name]) {
      return <Badge variant="default">Success</Badge>
    }
    return <Badge variant="secondary">Not tested</Badge>
  }

  const formatResponse = (data: any) => {
    if (!data) return 'No data'
    return JSON.stringify(data, null, 2)
  }

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-3xl font-bold">API Test Dashboard</h1>
          <p className="text-muted-foreground">
            Test unified API endpoints and verify responses
          </p>
        </div>
        <Button onClick={testAllEndpoints} disabled={loading}>
          {loading ? (
            <>
              <RefreshCw className="h-4 w-4 mr-2 animate-spin" />
              Testing...
            </>
          ) : (
            <>
              <Activity className="h-4 w-4 mr-2" />
              Test All Endpoints
            </>
          )}
        </Button>
      </div>

      <Tabs defaultValue="overview" className="space-y-4">
        <TabsList>
          <TabsTrigger value="overview">Overview</TabsTrigger>
          <TabsTrigger value="dashboard">Dashboard API</TabsTrigger>
          <TabsTrigger value="webhook">Webhook API</TabsTrigger>
          <TabsTrigger value="system">System API</TabsTrigger>
        </TabsList>

        <TabsContent value="overview">
          <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
            {[
              { name: 'health', title: 'Health Check', icon: <Server className="h-4 w-4" /> },
              { name: 'api-info', title: 'API Info', icon: <Database className="h-4 w-4" /> },
              { name: 'system-status', title: 'System Status', icon: <Activity className="h-4 w-4" /> },
              { name: 'system-info', title: 'System Info', icon: <Server className="h-4 w-4" /> },
              { name: 'connection-status', title: 'Connection Status', icon: <Database className="h-4 w-4" /> },
              { name: 'config', title: 'Configuration', icon: <Server className="h-4 w-4" /> },
              { name: 'workers', title: 'Workers', icon: <Activity className="h-4 w-4" /> },
              { name: 'stats', title: 'Statistics', icon: <Database className="h-4 w-4" /> },
              { name: 'metrics', title: 'Metrics', icon: <Activity className="h-4 w-4" /> },
              { name: 'plugins', title: 'Plugins', icon: <Server className="h-4 w-4" /> },
            ].map((endpoint) => (
              <Card key={endpoint.name}>
                <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                  <CardTitle className="text-sm font-medium flex items-center space-x-2">
                    {endpoint.icon}
                    <span>{endpoint.title}</span>
                  </CardTitle>
                  {getStatusIcon(endpoint.name)}
                </CardHeader>
                <CardContent>
                  <div className="flex items-center justify-between">
                    {getStatusBadge(endpoint.name)}
                    <Button
                      size="sm"
                      variant="outline"
                      onClick={() => {
                        const testMap: Record<string, () => Promise<any>> = {
                          'health': () => fetch('http://localhost:8001/health').then(r => r.json()),
                          'api-info': () => fetch('http://localhost:8001/api').then(r => r.json()),
                          'system-status': () => apiClient.getSystemStatus(),
                          'system-info': () => apiClient.getSystemInfo(),
                          'connection-status': () => apiClient.getConnectionStatus(),
                          'config': () => apiClient.getConfig(),
                          'workers': () => apiClient.getWorkers(),
                          'stats': () => apiClient.getStats(),
                          'metrics': () => apiClient.getMetrics(),
                          'plugins': () => apiClient.getPlugins(),
                        }
                        testEndpoint(endpoint.name, testMap[endpoint.name])
                      }}
                      disabled={loading}
                    >
                      Test
                    </Button>
                  </div>
                  {errors[endpoint.name] && (
                    <Alert variant="destructive" className="mt-2">
                      <AlertCircle className="h-4 w-4" />
                      <AlertDescription className="text-xs">
                        {errors[endpoint.name]}
                      </AlertDescription>
                    </Alert>
                  )}
                </CardContent>
              </Card>
            ))}
          </div>
        </TabsContent>

        <TabsContent value="dashboard">
          <div className="space-y-4">
            {['system-status', 'system-info', 'connection-status', 'config', 'workers', 'stats', 'metrics'].map((name) => (
              <Card key={name}>
                <CardHeader>
                  <CardTitle className="flex items-center justify-between">
                    <span>{name.replace('-', ' ').toUpperCase()}</span>
                    {getStatusBadge(name)}
                  </CardTitle>
                </CardHeader>
                <CardContent>
                  {errors[name] ? (
                    <Alert variant="destructive">
                      <AlertCircle className="h-4 w-4" />
                      <AlertDescription>{errors[name]}</AlertDescription>
                    </Alert>
                  ) : results[name] ? (
                    <pre className="text-xs bg-muted p-4 rounded-lg overflow-auto max-h-64">
                      {formatResponse(results[name])}
                    </pre>
                  ) : (
                    <p className="text-muted-foreground">Not tested yet</p>
                  )}
                </CardContent>
              </Card>
            ))}
          </div>
        </TabsContent>

        <TabsContent value="webhook">
          <Card>
            <CardHeader>
              <CardTitle>Webhook API Tests</CardTitle>
              <CardDescription>
                Test webhook-related endpoints
              </CardDescription>
            </CardHeader>
            <CardContent>
              <div className="space-y-4">
                <Card>
                  <CardHeader>
                    <CardTitle className="flex items-center justify-between">
                      <span>PLUGINS</span>
                      {getStatusBadge('plugins')}
                    </CardTitle>
                  </CardHeader>
                  <CardContent>
                    {errors['plugins'] ? (
                      <Alert variant="destructive">
                        <AlertCircle className="h-4 w-4" />
                        <AlertDescription>{errors['plugins']}</AlertDescription>
                      </Alert>
                    ) : results['plugins'] ? (
                      <pre className="text-xs bg-muted p-4 rounded-lg overflow-auto max-h-64">
                        {formatResponse(results['plugins'])}
                      </pre>
                    ) : (
                      <p className="text-muted-foreground">Not tested yet</p>
                    )}
                  </CardContent>
                </Card>
              </div>
            </CardContent>
          </Card>
        </TabsContent>

        <TabsContent value="system">
          <Card>
            <CardHeader>
              <CardTitle>System API Tests</CardTitle>
              <CardDescription>
                Test system-level endpoints
              </CardDescription>
            </CardHeader>
            <CardContent>
              <div className="space-y-4">
                {['health', 'api-info'].map((name) => (
                  <Card key={name}>
                    <CardHeader>
                      <CardTitle className="flex items-center justify-between">
                        <span>{name.replace('-', ' ').toUpperCase()}</span>
                        {getStatusBadge(name)}
                      </CardTitle>
                    </CardHeader>
                    <CardContent>
                      {errors[name] ? (
                        <Alert variant="destructive">
                          <AlertCircle className="h-4 w-4" />
                          <AlertDescription>{errors[name]}</AlertDescription>
                        </Alert>
                      ) : results[name] ? (
                        <pre className="text-xs bg-muted p-4 rounded-lg overflow-auto max-h-64">
                          {formatResponse(results[name])}
                        </pre>
                      ) : (
                        <p className="text-muted-foreground">Not tested yet</p>
                      )}
                    </CardContent>
                  </Card>
                ))}
              </div>
            </CardContent>
          </Card>
        </TabsContent>
      </Tabs>
    </div>
  )
}
