'use client'

import { useState, useEffect } from 'react'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
import { ScrollArea } from '@/components/ui/scroll-area'
import { useStagewise } from '@/hooks/useStagewise'
import { 
  Play, 
  Square, 
  Download, 
  Upload, 
  Trash2, 
  Clock, 
  CheckCircle, 
  XCircle, 
  AlertTriangle,
  Loader2,
  Network,
  Terminal,
  Activity,
  FileText
} from 'lucide-react'
import type { Stage, StageStep, NetworkRequest, ConsoleEntry } from '@/types/stagewise'

interface StagewiseDebuggerProps {
  className?: string
}

export function StagewiseDebugger({ className }: StagewiseDebuggerProps) {
  const stagewise = useStagewise({
    captureConsole: true,
    captureNetwork: true,
    enablePerformanceMetrics: true,
    maxLogEntries: 500
  })

  const [sessionName, setSessionName] = useState('')
  const [isExporting, setIsExporting] = useState(false)

  const handleStartSession = () => {
    const name = sessionName || `Debug Session ${new Date().toLocaleTimeString()}`
    stagewise.startSession(name, 'AI-assisted debugging session')
    setSessionName('')
  }

  const handleExportSession = async () => {
    if (!stagewise.session) return
    
    setIsExporting(true)
    try {
      const data = stagewise.exportSession()
      const blob = new Blob([data], { type: 'application/json' })
      const url = URL.createObjectURL(blob)
      const a = document.createElement('a')
      a.href = url
      a.download = `stagewise-${stagewise.session.name}-${Date.now()}.json`
      document.body.appendChild(a)
      a.click()
      document.body.removeChild(a)
      URL.revokeObjectURL(url)
    } catch (error) {
      console.error('Export failed:', error)
    } finally {
      setIsExporting(false)
    }
  }

  const handleImportSession = (event: React.ChangeEvent<HTMLInputElement>) => {
    const file = event.target.files?.[0]
    if (!file) return

    const reader = new FileReader()
    reader.onload = (e) => {
      try {
        const data = e.target?.result as string
        stagewise.importSession(data)
      } catch (error) {
        console.error('Import failed:', error)
      }
    }
    reader.readAsText(file)
  }

  const getStatusIcon = (status: string) => {
    switch (status) {
      case 'success':
        return <CheckCircle className="h-4 w-4 text-green-600" />
      case 'error':
        return <XCircle className="h-4 w-4 text-red-600" />
      case 'running':
        return <Loader2 className="h-4 w-4 text-blue-600 animate-spin" />
      case 'pending':
        return <Clock className="h-4 w-4 text-gray-600" />
      case 'skipped':
        return <AlertTriangle className="h-4 w-4 text-yellow-600" />
      default:
        return <Clock className="h-4 w-4 text-gray-600" />
    }
  }

  const getStatusColor = (status: string) => {
    switch (status) {
      case 'success': return 'default'
      case 'error': return 'destructive'
      case 'running': return 'default'
      case 'pending': return 'secondary'
      case 'skipped': return 'outline'
      default: return 'outline'
    }
  }

  const formatDuration = (duration?: number) => {
    if (!duration) return 'N/A'
    if (duration < 1000) return `${duration}ms`
    return `${(duration / 1000).toFixed(2)}s`
  }

  const StageComponent = ({ stage }: { stage: Stage }) => (
    <Card className="mb-4">
      <CardHeader className="pb-3">
        <div className="flex items-center justify-between">
          <div className="flex items-center space-x-2">
            {getStatusIcon(stage.status)}
            <CardTitle className="text-lg">{stage.name}</CardTitle>
            <Badge variant={getStatusColor(stage.status)}>{stage.status}</Badge>
          </div>
          <div className="text-sm text-muted-foreground">
            {formatDuration(stage.duration)}
          </div>
        </div>
        {stage.description && (
          <CardDescription>{stage.description}</CardDescription>
        )}
      </CardHeader>
      <CardContent>
        <div className="space-y-2">
          {stage.steps.map((step) => (
            <StepComponent key={step.id} step={step} />
          ))}
        </div>
      </CardContent>
    </Card>
  )

  const StepComponent = ({ step }: { step: StageStep }) => (
    <div className="flex items-start space-x-3 p-3 border rounded-lg">
      {getStatusIcon(step.status)}
      <div className="flex-1 min-w-0">
        <div className="flex items-center justify-between">
          <h4 className="font-medium truncate">{step.name}</h4>
          <div className="flex items-center space-x-2">
            <Badge variant={getStatusColor(step.status)} className="text-xs">
              {step.status}
            </Badge>
            <span className="text-xs text-muted-foreground">
              {formatDuration(step.duration)}
            </span>
          </div>
        </div>
        {step.description && (
          <p className="text-sm text-muted-foreground mt-1">{step.description}</p>
        )}
        {step.error && (
          <p className="text-sm text-red-600 mt-1">{step.error}</p>
        )}
        {step.data && (
          <details className="mt-2">
            <summary className="text-xs text-muted-foreground cursor-pointer">
              View Data
            </summary>
            <pre className="text-xs bg-muted p-2 rounded mt-1 overflow-auto max-h-32">
              {JSON.stringify(step.data, null, 2)}
            </pre>
          </details>
        )}
        {step.logs && step.logs.length > 0 && (
          <details className="mt-2">
            <summary className="text-xs text-muted-foreground cursor-pointer">
              View Logs ({step.logs.length})
            </summary>
            <div className="text-xs bg-muted p-2 rounded mt-1 max-h-32 overflow-auto">
              {step.logs.map((log, index) => (
                <div key={index} className="font-mono">{log}</div>
              ))}
            </div>
          </details>
        )}
      </div>
    </div>
  )

  const NetworkRequestComponent = ({ request }: { request: NetworkRequest }) => (
    <div className="flex items-start space-x-3 p-3 border rounded-lg">
      <Network className="h-4 w-4 mt-0.5 text-blue-600" />
      <div className="flex-1 min-w-0">
        <div className="flex items-center justify-between">
          <div className="flex items-center space-x-2">
            <Badge variant="outline" className="text-xs">{request.method}</Badge>
            <span className="font-medium truncate">{request.url}</span>
          </div>
          <div className="flex items-center space-x-2">
            {request.status && (
              <Badge variant={request.status >= 400 ? 'destructive' : 'default'} className="text-xs">
                {request.status}
              </Badge>
            )}
            <span className="text-xs text-muted-foreground">
              {formatDuration(request.duration)}
            </span>
          </div>
        </div>
        {request.error && (
          <p className="text-sm text-red-600 mt-1">{request.error}</p>
        )}
        <details className="mt-2">
          <summary className="text-xs text-muted-foreground cursor-pointer">
            View Details
          </summary>
          <div className="text-xs bg-muted p-2 rounded mt-1 max-h-32 overflow-auto">
            <div><strong>URL:</strong> {request.url}</div>
            <div><strong>Method:</strong> {request.method}</div>
            <div><strong>Status:</strong> {request.status} {request.statusText}</div>
            <div><strong>Duration:</strong> {formatDuration(request.duration)}</div>
            {request.requestHeaders && (
              <div><strong>Request Headers:</strong> {JSON.stringify(request.requestHeaders, null, 2)}</div>
            )}
          </div>
        </details>
      </div>
    </div>
  )

  const ConsoleEntryComponent = ({ entry }: { entry: ConsoleEntry }) => (
    <div className="flex items-start space-x-3 p-2 border-b">
      <Terminal className={`h-4 w-4 mt-0.5 ${
        entry.level === 'error' ? 'text-red-600' :
        entry.level === 'warn' ? 'text-yellow-600' :
        entry.level === 'info' ? 'text-blue-600' :
        'text-gray-600'
      }`} />
      <div className="flex-1 min-w-0">
        <div className="flex items-center space-x-2">
          <Badge variant="outline" className="text-xs">{entry.level}</Badge>
          <span className="text-xs text-muted-foreground">
            {entry.timestamp.toLocaleTimeString()}
          </span>
        </div>
        <pre className="text-sm font-mono mt-1 whitespace-pre-wrap">{entry.message}</pre>
        {entry.stack && (
          <details className="mt-1">
            <summary className="text-xs text-muted-foreground cursor-pointer">
              Stack Trace
            </summary>
            <pre className="text-xs bg-muted p-2 rounded mt-1 overflow-auto max-h-32">
              {entry.stack}
            </pre>
          </details>
        )}
      </div>
    </div>
  )

  return (
    <div className={className}>
      <Card>
        <CardHeader>
          <div className="flex items-center justify-between">
            <div>
              <CardTitle className="flex items-center space-x-2">
                <Activity className="h-5 w-5" />
                <span>Stagewise Debugger</span>
              </CardTitle>
              <CardDescription>
                AI-assisted debugging with stage-wise execution tracking
              </CardDescription>
            </div>
            <div className="flex items-center space-x-2">
              {stagewise.session ? (
                <>
                  <Button
                    onClick={stagewise.endSession}
                    variant="outline"
                    size="sm"
                  >
                    <Square className="h-4 w-4 mr-2" />
                    End Session
                  </Button>
                  <Button
                    onClick={handleExportSession}
                    variant="outline"
                    size="sm"
                    disabled={isExporting}
                  >
                    {isExporting ? (
                      <Loader2 className="h-4 w-4 mr-2 animate-spin" />
                    ) : (
                      <Download className="h-4 w-4 mr-2" />
                    )}
                    Export
                  </Button>
                </>
              ) : (
                <>
                  <input
                    type="text"
                    placeholder="Session name (optional)"
                    value={sessionName}
                    onChange={(e) => setSessionName(e.target.value)}
                    className="px-3 py-1 text-sm border rounded"
                  />
                  <Button onClick={handleStartSession} size="sm">
                    <Play className="h-4 w-4 mr-2" />
                    Start Session
                  </Button>
                </>
              )}
              <input
                type="file"
                accept=".json"
                onChange={handleImportSession}
                className="hidden"
                id="import-session"
              />
              <Button
                onClick={() => document.getElementById('import-session')?.click()}
                variant="outline"
                size="sm"
              >
                <Upload className="h-4 w-4 mr-2" />
                Import
              </Button>
              <Button
                onClick={stagewise.clearHistory}
                variant="outline"
                size="sm"
              >
                <Trash2 className="h-4 w-4 mr-2" />
                Clear
              </Button>
            </div>
          </div>
        </CardHeader>
        <CardContent>
          {stagewise.session ? (
            <div className="space-y-4">
              <div className="flex items-center justify-between p-4 bg-muted rounded-lg">
                <div>
                  <h3 className="font-medium">{stagewise.session.name}</h3>
                  <p className="text-sm text-muted-foreground">
                    {stagewise.session.description}
                  </p>
                </div>
                <div className="text-right">
                  <Badge variant={getStatusColor(stagewise.session.status)}>
                    {stagewise.session.status}
                  </Badge>
                  <p className="text-sm text-muted-foreground mt-1">
                    {formatDuration(stagewise.session.duration)}
                  </p>
                </div>
              </div>

              <Tabs defaultValue="stages" className="w-full">
                <TabsList className="grid w-full grid-cols-4">
                  <TabsTrigger value="stages">
                    Stages ({stagewise.session.stages.length})
                  </TabsTrigger>
                  <TabsTrigger value="network">
                    Network ({stagewise.networkRequests.length})
                  </TabsTrigger>
                  <TabsTrigger value="console">
                    Console ({stagewise.consoleEntries.length})
                  </TabsTrigger>
                  <TabsTrigger value="metrics">
                    Metrics ({stagewise.performanceMetrics.length})
                  </TabsTrigger>
                </TabsList>

                <TabsContent value="stages" className="space-y-4">
                  <ScrollArea className="h-96">
                    {stagewise.session.stages.length > 0 ? (
                      stagewise.session.stages.map((stage) => (
                        <StageComponent key={stage.id} stage={stage} />
                      ))
                    ) : (
                      <div className="text-center text-muted-foreground py-8">
                        No stages yet. Start debugging to see stages here.
                      </div>
                    )}
                  </ScrollArea>
                </TabsContent>

                <TabsContent value="network" className="space-y-4">
                  <ScrollArea className="h-96">
                    {stagewise.networkRequests.length > 0 ? (
                      <div className="space-y-2">
                        {stagewise.networkRequests.map((request) => (
                          <NetworkRequestComponent key={request.id} request={request} />
                        ))}
                      </div>
                    ) : (
                      <div className="text-center text-muted-foreground py-8">
                        No network requests captured yet.
                      </div>
                    )}
                  </ScrollArea>
                </TabsContent>

                <TabsContent value="console" className="space-y-4">
                  <ScrollArea className="h-96">
                    {stagewise.consoleEntries.length > 0 ? (
                      <div className="space-y-1">
                        {stagewise.consoleEntries.map((entry) => (
                          <ConsoleEntryComponent key={entry.id} entry={entry} />
                        ))}
                      </div>
                    ) : (
                      <div className="text-center text-muted-foreground py-8">
                        No console entries captured yet.
                      </div>
                    )}
                  </ScrollArea>
                </TabsContent>

                <TabsContent value="metrics" className="space-y-4">
                  <ScrollArea className="h-96">
                    {stagewise.performanceMetrics.length > 0 ? (
                      <div className="space-y-2">
                        {stagewise.performanceMetrics.map((metric) => (
                          <div key={metric.id} className="flex items-center justify-between p-2 border rounded">
                            <span className="font-medium">{metric.name}</span>
                            <div className="text-right">
                              <span className="font-mono">{metric.value} {metric.unit}</span>
                              <p className="text-xs text-muted-foreground">
                                {metric.timestamp.toLocaleTimeString()}
                              </p>
                            </div>
                          </div>
                        ))}
                      </div>
                    ) : (
                      <div className="text-center text-muted-foreground py-8">
                        No performance metrics captured yet.
                      </div>
                    )}
                  </ScrollArea>
                </TabsContent>
              </Tabs>
            </div>
          ) : (
            <div className="text-center text-muted-foreground py-8">
              <FileText className="h-12 w-12 mx-auto mb-4 opacity-50" />
              <h3 className="font-medium mb-2">No Active Session</h3>
              <p className="text-sm">
                Start a debugging session to begin tracking stages, network requests, and console output.
              </p>
            </div>
          )}
        </CardContent>
      </Card>
    </div>
  )
}
