'use client'

import { useState, useEffect } from 'react'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import { connectToLogs, disconnectFromLogs, subscribeToLogs, unsubscribeFromLogs, logsWebSocket } from '@/services/websocket'

export default function WebSocketDebugPage() {
  const [connectionState, setConnectionState] = useState<string>('disconnected')
  const [messages, setMessages] = useState<any[]>([])
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    // Monitor connection state changes
    const checkConnection = () => {
      setConnectionState(logsWebSocket.connectionState)
    }

    const interval = setInterval(checkConnection, 1000)
    return () => clearInterval(interval)
  }, [])

  const handleConnect = async () => {
    try {
      setError(null)
      await connectToLogs()
      
      // Subscribe to log messages
      const logHandler = (logData: any) => {
        setMessages(prev => [{
          type: 'log',
          timestamp: new Date().toISOString(),
          data: logData
        }, ...prev.slice(0, 49)]) // Keep only last 50 messages
      }
      
      subscribeToLogs(logHandler)
      
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Connection failed')
    }
  }

  const handleDisconnect = () => {
    disconnectFromLogs()
    setMessages([])
  }

  const handleClearMessages = () => {
    setMessages([])
  }

  const getConnectionBadgeColor = () => {
    switch (connectionState) {
      case 'connected': return 'bg-green-500'
      case 'connecting': return 'bg-yellow-500'
      case 'disconnected': return 'bg-red-500'
      default: return 'bg-gray-500'
    }
  }

  return (
    <div className="container mx-auto p-6 space-y-6">
      <div className="flex items-center justify-between">
        <h1 className="text-3xl font-bold">WebSocket Debug</h1>
        <Badge className={getConnectionBadgeColor()}>
          {connectionState}
        </Badge>
      </div>

      {error && (
        <Card className="border-red-200 bg-red-50">
          <CardContent className="pt-6">
            <p className="text-red-600">Error: {error}</p>
          </CardContent>
        </Card>
      )}

      <Card>
        <CardHeader>
          <CardTitle>Connection Control</CardTitle>
        </CardHeader>
        <CardContent className="space-y-4">
          <div className="flex space-x-2">
            <Button 
              onClick={handleConnect} 
              disabled={connectionState === 'connected' || connectionState === 'connecting'}
            >
              Connect
            </Button>
            <Button 
              onClick={handleDisconnect} 
              variant="outline"
              disabled={connectionState === 'disconnected'}
            >
              Disconnect
            </Button>
            <Button 
              onClick={handleClearMessages} 
              variant="outline"
            >
              Clear Messages
            </Button>
          </div>
          
          <div className="text-sm text-gray-600">
            <p>WebSocket URL: ws://localhost:8080/api/dashboard/logs/stream</p>
            <p>Connection State: {connectionState}</p>
            <p>Messages Received: {messages.length}</p>
          </div>
        </CardContent>
      </Card>

      <Card>
        <CardHeader>
          <CardTitle>Real-time Messages</CardTitle>
        </CardHeader>
        <CardContent>
          <div className="space-y-2 max-h-96 overflow-y-auto">
            {messages.length === 0 ? (
              <p className="text-gray-500 text-center py-4">No messages received yet</p>
            ) : (
              messages.map((message, index) => (
                <div key={index} className="p-3 bg-gray-50 rounded border text-sm">
                  <div className="flex justify-between items-start mb-2">
                    <Badge variant="outline">{message.type}</Badge>
                    <span className="text-xs text-gray-500">{message.timestamp}</span>
                  </div>
                  <pre className="text-xs overflow-x-auto">
                    {JSON.stringify(message.data, null, 2)}
                  </pre>
                </div>
              ))
            )}
          </div>
        </CardContent>
      </Card>
    </div>
  )
}
