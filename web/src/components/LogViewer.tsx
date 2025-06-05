import { useState, useEffect, useRef } from 'react'
import { useSearchParams } from 'react-router-dom'
import { Search, Pause, Play, Trash2, Download, Filter, X } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import { LogEntry } from '@/types/api'
import { apiClient } from '@/services/api'

interface LogViewerProps {
  className?: string
}

export function LogViewer({ className }: LogViewerProps) {
  const [searchParams, setSearchParams] = useSearchParams()
  const [logs, setLogs] = useState<LogEntry[]>([])
  const [filteredLogs, setFilteredLogs] = useState<LogEntry[]>([])
  const [ws, setWs] = useState<WebSocket | null>(null)
  const [isConnected, setIsConnected] = useState(false)
  const [isPaused, setIsPaused] = useState(false)
  const [searchTerm, setSearchTerm] = useState(searchParams.get('search') || '')
  const [levelFilter, setLevelFilter] = useState<string>(searchParams.get('level') || 'all')
  const [pluginFilter, setPluginFilter] = useState<string>(searchParams.get('plugin') || 'all')
  const [availablePlugins, setAvailablePlugins] = useState<string[]>([])
  const [autoScroll, setAutoScroll] = useState(true)
  const [showFilters, setShowFilters] = useState(
    searchParams.get('search') !== null ||
    searchParams.get('level') !== null ||
    searchParams.get('plugin') !== null
  )
  const logsEndRef = useRef<HTMLDivElement>(null)
  const logsContainerRef = useRef<HTMLDivElement>(null)

  // Fetch available plugins for filtering
  useEffect(() => {
    const fetchPlugins = async () => {
      try {
        const response = await apiClient.getPlugins()
        if (response.success && response.data) {
          const pluginNames = response.data.map((plugin: any) => plugin.name)
          setAvailablePlugins(pluginNames)
        }
      } catch (error) {
        console.error('Failed to fetch plugins:', error)
      }
    }

    fetchPlugins()
  }, [])

  // Extract unique plugins from logs
  useEffect(() => {
    const pluginsFromLogs = [...new Set(logs.filter(log => log.plugin).map(log => log.plugin!))]
    const allPlugins = [...new Set([...availablePlugins, ...pluginsFromLogs])]
    setAvailablePlugins(allPlugins)
  }, [logs])

  // Sync filters with URL parameters
  useEffect(() => {
    const params = new URLSearchParams()

    if (searchTerm) params.set('search', searchTerm)
    if (levelFilter !== 'all') params.set('level', levelFilter)
    if (pluginFilter !== 'all') params.set('plugin', pluginFilter)

    setSearchParams(params, { replace: true })
  }, [searchTerm, levelFilter, pluginFilter, setSearchParams])

  // WebSocket connection
  useEffect(() => {
    const connectWebSocket = () => {
      const websocket = new WebSocket('ws://localhost:8000/api/dashboard/logs/stream')
      
      websocket.onopen = () => {
        console.log('WebSocket connected')
        setIsConnected(true)
      }
      
      websocket.onmessage = (event) => {
        if (isPaused) return
        
        try {
          const logEntry: LogEntry = JSON.parse(event.data)
          setLogs(prev => [...prev, logEntry])
        } catch (error) {
          console.error('Failed to parse log entry:', error)
        }
      }
      
      websocket.onclose = () => {
        console.log('WebSocket disconnected')
        setIsConnected(false)
        // Attempt to reconnect after 3 seconds
        setTimeout(connectWebSocket, 3000)
      }
      
      websocket.onerror = (error) => {
        console.error('WebSocket error:', error)
        setIsConnected(false)
      }
      
      setWs(websocket)
    }

    connectWebSocket()

    return () => {
      if (ws) {
        ws.close()
      }
    }
  }, [isPaused])

  // Filter logs based on search term, level, and plugin
  useEffect(() => {
    let filtered = logs

    if (levelFilter !== 'all') {
      filtered = filtered.filter(log => log.level === levelFilter)
    }

    if (pluginFilter !== 'all') {
      filtered = filtered.filter(log => {
        if (pluginFilter === 'no-plugin') {
          return !log.plugin
        }
        return log.plugin === pluginFilter
      })
    }

    if (searchTerm) {
      filtered = filtered.filter(log =>
        log.message.toLowerCase().includes(searchTerm.toLowerCase()) ||
        (log.source && log.source.toLowerCase().includes(searchTerm.toLowerCase())) ||
        (log.plugin && log.plugin.toLowerCase().includes(searchTerm.toLowerCase()))
      )
    }

    setFilteredLogs(filtered)
  }, [logs, searchTerm, levelFilter, pluginFilter])

  // Auto scroll to bottom
  useEffect(() => {
    if (autoScroll && logsEndRef.current) {
      logsEndRef.current.scrollIntoView({ behavior: 'smooth' })
    }
  }, [filteredLogs, autoScroll])

  const togglePause = () => {
    setIsPaused(!isPaused)
  }

  const clearLogs = () => {
    setLogs([])
    setFilteredLogs([])
  }

  const clearFilters = () => {
    setSearchTerm('')
    setLevelFilter('all')
    setPluginFilter('all')
  }

  const hasActiveFilters = () => {
    return searchTerm !== '' || levelFilter !== 'all' || pluginFilter !== 'all'
  }

  const exportLogs = () => {
    const logText = filteredLogs.map(log => 
      `[${log.timestamp}] [${log.level.toUpperCase()}] ${log.source ? `[${log.source}] ` : ''}${log.message}`
    ).join('\n')
    
    const blob = new Blob([logText], { type: 'text/plain' })
    const url = URL.createObjectURL(blob)
    const a = document.createElement('a')
    a.href = url
    a.download = `webhook-bridge-logs-${new Date().toISOString().slice(0, 19)}.txt`
    document.body.appendChild(a)
    a.click()
    document.body.removeChild(a)
    URL.revokeObjectURL(url)
  }

  const getLevelColor = (level: string) => {
    switch (level) {
      case 'error':
        return 'text-red-500 bg-red-50 border-red-200'
      case 'warn':
        return 'text-yellow-600 bg-yellow-50 border-yellow-200'
      case 'info':
        return 'text-blue-500 bg-blue-50 border-blue-200'
      case 'debug':
        return 'text-gray-500 bg-gray-50 border-gray-200'
      default:
        return 'text-gray-600 bg-gray-50 border-gray-200'
    }
  }

  const formatTimestamp = (timestamp: string) => {
    return new Date(timestamp).toLocaleTimeString()
  }

  return (
    <Card className={className}>
      <CardHeader>
        <div className="flex items-center justify-between">
          <CardTitle className="flex items-center gap-2">
            Real-time Logs
            <div className={`w-2 h-2 rounded-full ${isConnected ? 'bg-green-500' : 'bg-red-500'}`} />
          </CardTitle>
          
          <div className="flex items-center gap-2">
            <Button
              variant="outline"
              size="sm"
              onClick={() => setShowFilters(!showFilters)}
              className="flex items-center gap-1"
            >
              <Filter className="h-4 w-4" />
              Filters
              {hasActiveFilters() && (
                <Badge variant="secondary" className="ml-1 h-4 w-4 p-0 text-xs">
                  !
                </Badge>
              )}
            </Button>

            <Button
              variant="outline"
              size="sm"
              onClick={togglePause}
              className="flex items-center gap-1"
            >
              {isPaused ? <Play className="h-4 w-4" /> : <Pause className="h-4 w-4" />}
              {isPaused ? 'Resume' : 'Pause'}
            </Button>

            <Button
              variant="outline"
              size="sm"
              onClick={clearLogs}
              className="flex items-center gap-1"
            >
              <Trash2 className="h-4 w-4" />
              Clear
            </Button>

            <Button
              variant="outline"
              size="sm"
              onClick={exportLogs}
              className="flex items-center gap-1"
            >
              <Download className="h-4 w-4" />
              Export
            </Button>
          </div>
        </div>
        
        {/* Quick Search */}
        <div className="flex items-center gap-4">
          <div className="relative flex-1">
            <Search className="absolute left-3 top-1/2 transform -translate-y-1/2 h-4 w-4 text-muted-foreground" />
            <input
              type="text"
              placeholder="Search logs..."
              value={searchTerm}
              onChange={(e) => setSearchTerm(e.target.value)}
              className="w-full pl-10 pr-4 py-2 border border-input rounded-md bg-background text-sm"
            />
          </div>

          <label className="flex items-center gap-2 text-sm">
            <input
              type="checkbox"
              checked={autoScroll}
              onChange={(e) => setAutoScroll(e.target.checked)}
              className="rounded"
            />
            Auto-scroll
          </label>
        </div>

        {/* Advanced Filters */}
        {showFilters && (
          <div className="space-y-4 p-4 bg-muted/30 rounded-md">
            <div className="flex items-center justify-between">
              <h4 className="text-sm font-medium">Advanced Filters</h4>
              {hasActiveFilters() && (
                <Button
                  variant="ghost"
                  size="sm"
                  onClick={clearFilters}
                  className="flex items-center gap-1 text-xs"
                >
                  <X className="h-3 w-3" />
                  Clear All
                </Button>
              )}
            </div>

            <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
              <div>
                <label className="text-xs font-medium text-muted-foreground mb-1 block">
                  Log Level
                </label>
                <Select value={levelFilter} onValueChange={setLevelFilter}>
                  <SelectTrigger className="h-8">
                    <SelectValue />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem value="all">All Levels</SelectItem>
                    <SelectItem value="error">Error</SelectItem>
                    <SelectItem value="warn">Warning</SelectItem>
                    <SelectItem value="info">Info</SelectItem>
                    <SelectItem value="debug">Debug</SelectItem>
                  </SelectContent>
                </Select>
              </div>

              <div>
                <label className="text-xs font-medium text-muted-foreground mb-1 block">
                  Plugin
                </label>
                <Select value={pluginFilter} onValueChange={setPluginFilter}>
                  <SelectTrigger className="h-8">
                    <SelectValue />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem value="all">All Plugins</SelectItem>
                    <SelectItem value="no-plugin">System (No Plugin)</SelectItem>
                    {availablePlugins.map((plugin) => (
                      <SelectItem key={plugin} value={plugin}>
                        {plugin}
                      </SelectItem>
                    ))}
                  </SelectContent>
                </Select>
              </div>

              <div className="flex items-end">
                <div className="text-xs text-muted-foreground">
                  <div>Total: {logs.length} logs</div>
                  <div>Filtered: {filteredLogs.length} logs</div>
                </div>
              </div>
            </div>

            {/* Active Filters Display */}
            {hasActiveFilters() && (
              <div className="flex items-center gap-2 flex-wrap">
                <span className="text-xs font-medium text-muted-foreground">Active filters:</span>
                {searchTerm && (
                  <Badge variant="secondary" className="text-xs">
                    Search: "{searchTerm}"
                    <button
                      onClick={() => setSearchTerm('')}
                      className="ml-1 hover:bg-muted-foreground/20 rounded-full p-0.5"
                    >
                      <X className="h-2 w-2" />
                    </button>
                  </Badge>
                )}
                {levelFilter !== 'all' && (
                  <Badge variant="secondary" className="text-xs">
                    Level: {levelFilter}
                    <button
                      onClick={() => setLevelFilter('all')}
                      className="ml-1 hover:bg-muted-foreground/20 rounded-full p-0.5"
                    >
                      <X className="h-2 w-2" />
                    </button>
                  </Badge>
                )}
                {pluginFilter !== 'all' && (
                  <Badge variant="secondary" className="text-xs">
                    Plugin: {pluginFilter === 'no-plugin' ? 'System' : pluginFilter}
                    <button
                      onClick={() => setPluginFilter('all')}
                      className="ml-1 hover:bg-muted-foreground/20 rounded-full p-0.5"
                    >
                      <X className="h-2 w-2" />
                    </button>
                  </Badge>
                )}
              </div>
            )}
          </div>
        )}
      </CardHeader>
      
      <CardContent>
        <div 
          ref={logsContainerRef}
          className="h-96 overflow-y-auto bg-muted/30 rounded-md p-4 font-mono text-sm"
        >
          {filteredLogs.length === 0 ? (
            <div className="text-center text-muted-foreground py-8">
              {logs.length === 0 ? 'No logs yet...' : 'No logs match your filters'}
            </div>
          ) : (
            <div className="space-y-1">
              {filteredLogs.map((log, index) => (
                <div
                  key={index}
                  className={`p-2 rounded border-l-4 ${getLevelColor(log.level)}`}
                >
                  <div className="flex items-start gap-2">
                    <span className="text-xs text-muted-foreground whitespace-nowrap">
                      {formatTimestamp(log.timestamp)}
                    </span>
                    <span className={`text-xs font-medium px-1 rounded ${
                      log.level === 'error' ? 'bg-red-100 text-red-700' :
                      log.level === 'warn' ? 'bg-yellow-100 text-yellow-700' :
                      log.level === 'info' ? 'bg-blue-100 text-blue-700' :
                      'bg-gray-100 text-gray-700'
                    }`}>
                      {log.level.toUpperCase()}
                    </span>
                    {log.source && (
                      <span className="text-xs text-muted-foreground">
                        [{log.source}]
                      </span>
                    )}
                    {log.plugin && (
                      <Badge variant="outline" className="text-xs px-1 py-0">
                        {log.plugin}
                      </Badge>
                    )}
                    <span className="flex-1 break-words">
                      {log.message}
                    </span>
                  </div>
                  {log.details && (
                    <div className="mt-1 ml-16 text-xs text-muted-foreground">
                      <pre className="whitespace-pre-wrap">
                        {typeof log.details === 'string' ? log.details : JSON.stringify(log.details, null, 2)}
                      </pre>
                    </div>
                  )}
                </div>
              ))}
              <div ref={logsEndRef} />
            </div>
          )}
        </div>
        
        <div className="mt-4 flex items-center justify-between text-xs text-muted-foreground">
          <div className="flex items-center gap-4">
            <span>
              Showing {filteredLogs.length} of {logs.length} logs
            </span>
            {hasActiveFilters() && (
              <span className="text-orange-600">
                Filters active
              </span>
            )}
            {availablePlugins.length > 0 && (
              <span>
                {availablePlugins.length} plugins available
              </span>
            )}
          </div>
          <span>
            Status: {isConnected ? 'Connected' : 'Disconnected'}
            {isPaused && ' (Paused)'}
          </span>
        </div>
      </CardContent>
    </Card>
  )
}
