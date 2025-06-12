'use client'

import { useState, useEffect, useRef, useCallback, Suspense } from 'react'
import { useSearchParams } from 'next/navigation'
import { Search, Pause, Play, Trash2, Download, Filter, X, Wifi, WifiOff } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import { LogEntry } from '@/types/api'
import { apiClient } from '@/services/api'
import { subscribeToLogs, unsubscribeFromLogs, connectToLogs, disconnectFromLogs, logsWebSocket } from '@/services/websocket'

interface LogViewerProps {
  className?: string
}

function LogViewerContent({ className }: LogViewerProps) {
  const [logs, setLogs] = useState<LogEntry[]>([])
  const [filteredLogs, setFilteredLogs] = useState<LogEntry[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [isPaused, setIsPaused] = useState(false)
  const [searchTerm, setSearchTerm] = useState('')
  const [levelFilter, setLevelFilter] = useState<string>('all')
  const [sourceFilter, setSourceFilter] = useState<string>('all')
  const [autoScroll, setAutoScroll] = useState(true)
  const [isWebSocketConnected, setIsWebSocketConnected] = useState(false)
  const [useRealTime, setUseRealTime] = useState(true)

  const logContainerRef = useRef<HTMLDivElement>(null)
  const searchParams = useSearchParams()

  // Get initial filters from URL params
  useEffect(() => {
    const level = searchParams.get('level')
    const source = searchParams.get('source')
    const search = searchParams.get('search')

    if (level) setLevelFilter(level)
    if (source) setSourceFilter(source)
    if (search) setSearchTerm(search)
  }, [searchParams])

  // Fetch logs
  const fetchLogs = useCallback(async () => {
    try {
      setError(null)
      const params: any = { limit: 500 }
      if (levelFilter !== 'all') params.level = levelFilter
      if (sourceFilter !== 'all') params.source = sourceFilter

      const data = await apiClient.getLogs(params)
      setLogs(data)
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to fetch logs')
    } finally {
      setLoading(false)
    }
  }, [levelFilter, sourceFilter])

  // Initial fetch
  useEffect(() => {
    fetchLogs()
  }, [levelFilter, sourceFilter]) // eslint-disable-line react-hooks/exhaustive-deps

  // WebSocket connection management
  useEffect(() => {
    if (!useRealTime || isPaused) return

    const connectWebSocket = async () => {
      try {
        await connectToLogs()
        setIsWebSocketConnected(true)
        setError(null)
      } catch (err) {
        console.error('Failed to connect to WebSocket:', err)
        setIsWebSocketConnected(false)
        setError('Failed to connect to real-time log streaming')
      }
    }

    connectWebSocket()

    // Monitor connection status
    const checkConnection = () => {
      setIsWebSocketConnected(logsWebSocket.isConnected)
    }

    const connectionInterval = setInterval(checkConnection, 5000)

    return () => {
      clearInterval(connectionInterval)
      disconnectFromLogs()
      setIsWebSocketConnected(false)
    }
  }, [useRealTime, isPaused])

  // Handle real-time log updates
  useEffect(() => {
    if (!useRealTime || !isWebSocketConnected) return

    const handleLogUpdate = (logEntry: any) => {
      // Convert WebSocket log entry to our LogEntry format
      const newLog: LogEntry = {
        id: logEntry.id || Date.now().toString(),
        timestamp: logEntry.timestamp || new Date().toISOString(),
        level: logEntry.level || 'info',
        message: logEntry.message || '',
        source: logEntry.source || 'system',
        component: logEntry.component || logEntry.plugin_name || undefined,
      }

      setLogs(prevLogs => {
        const updatedLogs = [newLog, ...prevLogs]
        // Keep only the latest 500 logs to prevent memory issues
        return updatedLogs.slice(0, 500)
      })
    }

    subscribeToLogs(handleLogUpdate)

    return () => {
      unsubscribeFromLogs(handleLogUpdate)
    }
  }, [useRealTime, isWebSocketConnected])

  // Auto-refresh logs (fallback when WebSocket is not available)
  useEffect(() => {
    if (isPaused || (useRealTime && isWebSocketConnected)) return

    const interval = setInterval(fetchLogs, 2000)
    return () => clearInterval(interval)
  }, [isPaused, levelFilter, sourceFilter, useRealTime, isWebSocketConnected, fetchLogs])

  // Filter logs based on search term
  useEffect(() => {
    let filtered = logs
    
    if (searchTerm) {
      const term = searchTerm.toLowerCase()
      filtered = logs.filter(log => 
        log.message.toLowerCase().includes(term) ||
        log.source?.toLowerCase().includes(term) ||
        log.component?.toLowerCase().includes(term)
      )
    }
    
    setFilteredLogs(filtered)
  }, [logs, searchTerm])

  // Auto-scroll to bottom
  useEffect(() => {
    if (autoScroll && logContainerRef.current) {
      logContainerRef.current.scrollTop = logContainerRef.current.scrollHeight
    }
  }, [filteredLogs, autoScroll])

  const clearLogs = () => {
    setLogs([])
    setFilteredLogs([])
  }

  const downloadLogs = () => {
    const logText = filteredLogs.map(log => 
      `[${log.timestamp}] ${log.level.toUpperCase()} ${log.source || 'system'}: ${log.message}`
    ).join('\n')
    
    const blob = new Blob([logText], { type: 'text/plain' })
    const url = URL.createObjectURL(blob)
    const a = document.createElement('a')
    a.href = url
    a.download = `webhook-bridge-logs-${new Date().toISOString().split('T')[0]}.txt`
    document.body.appendChild(a)
    a.click()
    document.body.removeChild(a)
    URL.revokeObjectURL(url)
  }

  const getLevelColor = (level: string) => {
    switch (level.toLowerCase()) {
      case 'error':
      case 'fatal':
        return 'destructive'
      case 'warn':
        return 'secondary'
      case 'info':
        return 'default'
      case 'debug':
        return 'outline'
      default:
        return 'outline'
    }
  }

  const formatTimestamp = (timestamp: string) => {
    return new Date(timestamp).toLocaleTimeString()
  }

  // Get unique sources for filter
  const uniqueSources = Array.from(new Set(logs.map(log => log.source).filter(Boolean)))

  return (
    <Card className={className}>
      <CardHeader>
        <div className="flex items-center justify-between">
          <CardTitle className="flex items-center space-x-2">
            <span>System Logs</span>
            {loading && <div className="animate-spin rounded-full h-4 w-4 border-b-2 border-primary" />}
          </CardTitle>
          
          <div className="flex items-center space-x-2">
            <Button
              variant="outline"
              size="sm"
              onClick={() => setUseRealTime(!useRealTime)}
              className={isWebSocketConnected ? 'text-green-600' : 'text-gray-500'}
            >
              {isWebSocketConnected ? <Wifi className="h-4 w-4" /> : <WifiOff className="h-4 w-4" />}
              {useRealTime ? 'Real-time' : 'Polling'}
            </Button>

            <Button
              variant="outline"
              size="sm"
              onClick={() => setIsPaused(!isPaused)}
            >
              {isPaused ? <Play className="h-4 w-4" /> : <Pause className="h-4 w-4" />}
              {isPaused ? 'Resume' : 'Pause'}
            </Button>

            <Button
              variant="outline"
              size="sm"
              onClick={clearLogs}
            >
              <Trash2 className="h-4 w-4" />
              Clear
            </Button>

            <Button
              variant="outline"
              size="sm"
              onClick={downloadLogs}
              disabled={filteredLogs.length === 0}
            >
              <Download className="h-4 w-4" />
              Download
            </Button>
          </div>
        </div>
        
        {/* Filters */}
        <div className="flex items-center space-x-4">
          <div className="relative flex-1">
            <Search className="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground" />
            <input
              type="text"
              placeholder="Search logs..."
              value={searchTerm}
              onChange={(e) => setSearchTerm(e.target.value)}
              className="h-9 w-full rounded-md border border-input bg-background pl-10 pr-3 text-sm ring-offset-background placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2"
            />
            {searchTerm && (
              <Button
                variant="ghost"
                size="sm"
                className="absolute right-1 top-1/2 h-7 w-7 -translate-y-1/2 p-0"
                onClick={() => setSearchTerm('')}
              >
                <X className="h-3 w-3" />
              </Button>
            )}
          </div>
          
          <Select value={levelFilter} onValueChange={setLevelFilter}>
            <SelectTrigger className="w-32">
              <SelectValue placeholder="Level" />
            </SelectTrigger>
            <SelectContent>
              <SelectItem value="all">All Levels</SelectItem>
              <SelectItem value="debug">Debug</SelectItem>
              <SelectItem value="info">Info</SelectItem>
              <SelectItem value="warn">Warning</SelectItem>
              <SelectItem value="error">Error</SelectItem>
              <SelectItem value="fatal">Fatal</SelectItem>
            </SelectContent>
          </Select>
          
          <Select value={sourceFilter} onValueChange={setSourceFilter}>
            <SelectTrigger className="w-32">
              <SelectValue placeholder="Source" />
            </SelectTrigger>
            <SelectContent>
              <SelectItem value="all">All Sources</SelectItem>
              {uniqueSources.map(source => (
                <SelectItem key={source} value={source!}>
                  {source}
                </SelectItem>
              ))}
            </SelectContent>
          </Select>
        </div>
        
        <div className="flex items-center justify-between text-sm text-muted-foreground">
          <div className="flex items-center space-x-4">
            <span>
              Showing {filteredLogs.length} of {logs.length} logs
            </span>
            {useRealTime && (
              <span className={`flex items-center space-x-1 ${isWebSocketConnected ? 'text-green-600' : 'text-red-500'}`}>
                {isWebSocketConnected ? <Wifi className="h-3 w-3" /> : <WifiOff className="h-3 w-3" />}
                <span>{isWebSocketConnected ? 'Connected' : 'Disconnected'}</span>
              </span>
            )}
          </div>
          <label className="flex items-center space-x-2">
            <input
              type="checkbox"
              checked={autoScroll}
              onChange={(e) => setAutoScroll(e.target.checked)}
              className="rounded border-input"
            />
            <span>Auto-scroll</span>
          </label>
        </div>
      </CardHeader>
      
      <CardContent>
        {error && (
          <div className="mb-4 p-3 bg-destructive/10 border border-destructive/20 rounded-md">
            <p className="text-sm text-destructive">{error}</p>
          </div>
        )}
        
        <div
          ref={logContainerRef}
          className="h-96 overflow-auto space-y-1 font-mono text-sm"
        >
          {filteredLogs.length === 0 ? (
            <div className="flex items-center justify-center h-full text-muted-foreground">
              {loading ? 'Loading logs...' : 'No logs found'}
            </div>
          ) : (
            filteredLogs.map((log) => (
              <div
                key={log.id}
                className="flex items-start space-x-3 p-2 rounded hover:bg-muted/50"
              >
                <span className="text-xs text-muted-foreground whitespace-nowrap">
                  {formatTimestamp(log.timestamp)}
                </span>
                
                <Badge variant={getLevelColor(log.level)} className="text-xs">
                  {log.level.toUpperCase()}
                </Badge>
                
                {log.source && (
                  <span className="text-xs text-muted-foreground">
                    [{log.source}]
                  </span>
                )}
                
                <span className="flex-1 break-words">
                  {log.message}
                </span>
              </div>
            ))
          )}
        </div>
      </CardContent>
    </Card>
  )
}

export function LogViewer({ className }: LogViewerProps) {
  return (
    <Suspense fallback={
      <Card className={className}>
        <CardContent className="flex items-center justify-center h-96">
          <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-primary"></div>
        </CardContent>
      </Card>
    }>
      <LogViewerContent className={className} />
    </Suspense>
  )
}
