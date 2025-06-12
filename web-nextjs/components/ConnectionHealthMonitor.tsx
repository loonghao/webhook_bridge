'use client'

import React, { useState, useEffect } from 'react'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Alert, AlertDescription } from '@/components/ui/alert'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
import { 
  Activity, 
  AlertTriangle, 
  CheckCircle, 
  Clock, 
  RefreshCw, 
  TrendingUp,
  Wifi,
  WifiOff
} from 'lucide-react'
import { useConnectionOptimizer, ConnectionMetrics, OptimizationSuggestion } from '@/lib/connection-optimizer'
import { useStagewise } from '@/hooks/useStagewise'

export function ConnectionHealthMonitor() {
  const { analyzeConnections, logOptimizationReport } = useConnectionOptimizer()
  const stagewise = useStagewise()
  const [metrics, setMetrics] = useState<ConnectionMetrics | null>(null)
  const [suggestions, setSuggestions] = useState<OptimizationSuggestion[]>([])
  const [isAnalyzing, setIsAnalyzing] = useState(false)
  const [autoRefresh, setAutoRefresh] = useState(false)

  const performAnalysis = async () => {
    setIsAnalyzing(true)
    try {
      const result = analyzeConnections()
      setMetrics(result.metrics)
      setSuggestions(result.suggestions)
    } catch (error) {
      console.error('Analysis failed:', error)
    } finally {
      setIsAnalyzing(false)
    }
  }

  useEffect(() => {
    performAnalysis()
  }, []) // eslint-disable-line react-hooks/exhaustive-deps

  useEffect(() => {
    if (autoRefresh) {
      const interval = setInterval(performAnalysis, 10000) // 10 seconds
      return () => clearInterval(interval)
    }
  }, [autoRefresh]) // eslint-disable-line react-hooks/exhaustive-deps

  const getHealthStatus = () => {
    if (!metrics) return { status: 'unknown', color: 'gray' }
    
    if (metrics.errorRate > 0.1) return { status: 'critical', color: 'red' }
    if (metrics.errorRate > 0.05 || metrics.slowRequests > metrics.totalRequests * 0.2) {
      return { status: 'warning', color: 'yellow' }
    }
    return { status: 'healthy', color: 'green' }
  }

  const health = getHealthStatus()

  const formatLatency = (ms: number) => {
    if (ms < 1000) return `${ms.toFixed(0)}ms`
    return `${(ms / 1000).toFixed(1)}s`
  }

  const getPriorityIcon = (priority: string) => {
    switch (priority) {
      case 'high': return <AlertTriangle className="h-4 w-4 text-red-500" />
      case 'medium': return <Clock className="h-4 w-4 text-yellow-500" />
      case 'low': return <TrendingUp className="h-4 w-4 text-blue-500" />
      default: return null
    }
  }

  const getPriorityColor = (priority: string) => {
    switch (priority) {
      case 'high': return 'destructive'
      case 'medium': return 'default'
      case 'low': return 'secondary'
      default: return 'outline'
    }
  }

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div className="flex items-center space-x-2">
          {health.status === 'healthy' ? (
            <Wifi className="h-5 w-5 text-green-500" />
          ) : (
            <WifiOff className="h-5 w-5 text-red-500" />
          )}
          <h2 className="text-2xl font-bold">Connection Health</h2>
          <Badge variant={health.color === 'green' ? 'default' : 'destructive'}>
            {health.status.toUpperCase()}
          </Badge>
        </div>
        
        <div className="flex items-center space-x-2">
          <Button
            variant="outline"
            size="sm"
            onClick={() => setAutoRefresh(!autoRefresh)}
          >
            <RefreshCw className={`h-4 w-4 mr-2 ${autoRefresh ? 'animate-spin' : ''}`} />
            Auto Refresh
          </Button>
          <Button
            variant="outline"
            size="sm"
            onClick={performAnalysis}
            disabled={isAnalyzing}
          >
            <Activity className="h-4 w-4 mr-2" />
            Analyze Now
          </Button>
          <Button
            variant="outline"
            size="sm"
            onClick={logOptimizationReport}
          >
            Export Report
          </Button>
        </div>
      </div>

      {/* Metrics Overview */}
      {metrics && (
        <div className="grid grid-cols-1 md:grid-cols-4 gap-4">
          <Card>
            <CardHeader className="pb-2">
              <CardTitle className="text-sm font-medium">Total Requests</CardTitle>
            </CardHeader>
            <CardContent>
              <div className="text-2xl font-bold">{metrics.totalRequests}</div>
            </CardContent>
          </Card>
          
          <Card>
            <CardHeader className="pb-2">
              <CardTitle className="text-sm font-medium">Error Rate</CardTitle>
            </CardHeader>
            <CardContent>
              <div className="text-2xl font-bold text-red-500">
                {(metrics.errorRate * 100).toFixed(1)}%
              </div>
              <p className="text-xs text-muted-foreground">
                {metrics.failedRequests} failed
              </p>
            </CardContent>
          </Card>
          
          <Card>
            <CardHeader className="pb-2">
              <CardTitle className="text-sm font-medium">Avg Latency</CardTitle>
            </CardHeader>
            <CardContent>
              <div className="text-2xl font-bold">
                {formatLatency(metrics.averageLatency)}
              </div>
            </CardContent>
          </Card>
          
          <Card>
            <CardHeader className="pb-2">
              <CardTitle className="text-sm font-medium">Slow Requests</CardTitle>
            </CardHeader>
            <CardContent>
              <div className="text-2xl font-bold text-yellow-500">
                {metrics.slowRequests}
              </div>
              <p className="text-xs text-muted-foreground">
                &gt; 2s response time
              </p>
            </CardContent>
          </Card>
        </div>
      )}

      {/* Detailed Analysis */}
      <Tabs defaultValue="suggestions" className="w-full">
        <TabsList>
          <TabsTrigger value="suggestions">
            Optimization Suggestions ({suggestions.length})
          </TabsTrigger>
          <TabsTrigger value="endpoints">
            Endpoint Performance
          </TabsTrigger>
          <TabsTrigger value="errors">
            Common Errors
          </TabsTrigger>
        </TabsList>

        <TabsContent value="suggestions" className="space-y-4">
          {suggestions.length === 0 ? (
            <Alert>
              <CheckCircle className="h-4 w-4" />
              <AlertDescription>
                No optimization suggestions at this time. Your connections are performing well!
              </AlertDescription>
            </Alert>
          ) : (
            suggestions.map((suggestion, index) => (
              <Card key={index}>
                <CardHeader>
                  <div className="flex items-center justify-between">
                    <div className="flex items-center space-x-2">
                      {getPriorityIcon(suggestion.priority)}
                      <CardTitle className="text-lg">{suggestion.title}</CardTitle>
                    </div>
                    <Badge variant={getPriorityColor(suggestion.priority) as any}>
                      {suggestion.priority.toUpperCase()}
                    </Badge>
                  </div>
                </CardHeader>
                <CardContent className="space-y-3">
                  <p className="text-muted-foreground">{suggestion.description}</p>
                  <p className="font-medium">{suggestion.action}</p>
                  {suggestion.code && (
                    <pre className="bg-muted p-3 rounded-md text-sm overflow-x-auto">
                      <code>{suggestion.code.trim()}</code>
                    </pre>
                  )}
                </CardContent>
              </Card>
            ))
          )}
        </TabsContent>

        <TabsContent value="endpoints" className="space-y-4">
          {metrics && Object.keys(metrics.endpointPerformance).length > 0 ? (
            Object.entries(metrics.endpointPerformance).map(([endpoint, perf]) => (
              <Card key={endpoint}>
                <CardHeader>
                  <CardTitle className="text-lg font-mono">{endpoint}</CardTitle>
                </CardHeader>
                <CardContent>
                  <div className="grid grid-cols-3 gap-4">
                    <div>
                      <p className="text-sm text-muted-foreground">Requests</p>
                      <p className="text-lg font-bold">{perf.count}</p>
                    </div>
                    <div>
                      <p className="text-sm text-muted-foreground">Avg Latency</p>
                      <p className="text-lg font-bold">{formatLatency(perf.averageLatency)}</p>
                    </div>
                    <div>
                      <p className="text-sm text-muted-foreground">Error Rate</p>
                      <p className={`text-lg font-bold ${perf.errorRate > 0.1 ? 'text-red-500' : 'text-green-500'}`}>
                        {(perf.errorRate * 100).toFixed(1)}%
                      </p>
                    </div>
                  </div>
                  {perf.lastError && (
                    <Alert className="mt-3">
                      <AlertTriangle className="h-4 w-4" />
                      <AlertDescription>
                        Last error: {perf.lastError}
                      </AlertDescription>
                    </Alert>
                  )}
                </CardContent>
              </Card>
            ))
          ) : (
            <Alert>
              <AlertDescription>
                No endpoint performance data available. Make some API requests to see analysis.
              </AlertDescription>
            </Alert>
          )}
        </TabsContent>

        <TabsContent value="errors" className="space-y-4">
          {metrics && Object.keys(metrics.commonErrors).length > 0 ? (
            Object.entries(metrics.commonErrors)
              .sort(([,a], [,b]) => b - a)
              .map(([error, count]) => (
                <Card key={error}>
                  <CardContent className="pt-6">
                    <div className="flex items-center justify-between">
                      <p className="font-mono text-sm">{error}</p>
                      <Badge variant="destructive">{count} occurrences</Badge>
                    </div>
                  </CardContent>
                </Card>
              ))
          ) : (
            <Alert>
              <CheckCircle className="h-4 w-4" />
              <AlertDescription>
                No errors detected. Great job!
              </AlertDescription>
            </Alert>
          )}
        </TabsContent>
      </Tabs>
    </div>
  )
}
