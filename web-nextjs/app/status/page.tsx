'use client'

import { Activity, Server, Users, Database, HardDrive, Cpu } from 'lucide-react'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Layout } from '@/components/Layout'
import { useDashboard } from '@/hooks/useDashboard'

export default function SystemStatus() {
  const { status, workers, plugins, loading, error, lastUpdated } = useDashboard()

  const formatTime = (date: Date | null) => {
    if (!date) return 'Never'
    return date.toLocaleTimeString()
  }

  return (
    <Layout>
      <div className="space-y-6">
        <div>
          <h1 className="text-3xl font-bold tracking-tight">System Status</h1>
          <p className="text-muted-foreground">
            Monitor system health and component status
          </p>
        </div>

        {error && (
          <Card className="border-destructive">
            <CardContent className="pt-6">
              <div className="flex items-center space-x-2 text-destructive">
                <Activity className="h-4 w-4" />
                <span className="text-sm">{error}</span>
              </div>
            </CardContent>
          </Card>
        )}

        {/* System Health Overview */}
        <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
          <Card>
            <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
              <CardTitle className="text-sm font-medium">Go Server</CardTitle>
              <Server className="h-4 w-4 text-muted-foreground" />
            </CardHeader>
            <CardContent>
              <div className="text-2xl font-bold">
                {loading ? '...' : status?.service || 'Unknown'}
              </div>
              <p className={`text-xs ${
                status?.status === 'healthy' ? 'text-green-600' : 'text-red-600'
              }`}>
                {status?.status || 'Unknown'}
              </p>
            </CardContent>
          </Card>

          <Card>
            <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
              <CardTitle className="text-sm font-medium">System Uptime</CardTitle>
              <Activity className="h-4 w-4 text-muted-foreground" />
            </CardHeader>
            <CardContent>
              <div className="text-2xl font-bold">
                {loading ? '...' : status?.uptime || 'Unknown'}
              </div>
              <p className="text-xs text-muted-foreground">
                Since last restart
              </p>
            </CardContent>
          </Card>

          <Card>
            <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
              <CardTitle className="text-sm font-medium">Version</CardTitle>
              <Database className="h-4 w-4 text-muted-foreground" />
            </CardHeader>
            <CardContent>
              <div className="text-2xl font-bold">
                {loading ? '...' : status?.version || 'Unknown'}
              </div>
              <p className="text-xs text-muted-foreground">
                Go {status?.goVersion || 'Unknown'}
              </p>
            </CardContent>
          </Card>
        </div>

        {/* Resource Usage */}
        {status?.memory && (
          <div className="grid gap-4 md:grid-cols-3">
            <Card>
              <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                <CardTitle className="text-sm font-medium">Memory Usage</CardTitle>
                <HardDrive className="h-4 w-4 text-muted-foreground" />
              </CardHeader>
              <CardContent>
                <div className="text-2xl font-bold">
                  {status.memory.percentage.toFixed(1)}%
                </div>
                <p className="text-xs text-muted-foreground">
                  {(status.memory.used / 1024 / 1024).toFixed(0)}MB / {(status.memory.total / 1024 / 1024).toFixed(0)}MB
                </p>
              </CardContent>
            </Card>

            <Card>
              <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                <CardTitle className="text-sm font-medium">CPU Usage</CardTitle>
                <Cpu className="h-4 w-4 text-muted-foreground" />
              </CardHeader>
              <CardContent>
                <div className="text-2xl font-bold">
                  {status.cpu?.usage.toFixed(1) || '0'}%
                </div>
                <p className="text-xs text-muted-foreground">
                  {status.cpu?.cores || 'Unknown'} cores
                </p>
              </CardContent>
            </Card>

            <Card>
              <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                <CardTitle className="text-sm font-medium">Disk Usage</CardTitle>
                <Database className="h-4 w-4 text-muted-foreground" />
              </CardHeader>
              <CardContent>
                <div className="text-2xl font-bold">
                  {status.disk?.percentage.toFixed(1) || '0'}%
                </div>
                <p className="text-xs text-muted-foreground">
                  {status.disk ? `${(status.disk.used / 1024 / 1024 / 1024).toFixed(1)}GB / ${(status.disk.total / 1024 / 1024 / 1024).toFixed(1)}GB` : 'Unknown'}
                </p>
              </CardContent>
            </Card>
          </div>
        )}

        {/* Workers Status */}
        <div className="grid gap-4 md:grid-cols-2">
          <Card>
            <CardHeader>
              <CardTitle>Worker Pool Status</CardTitle>
              <CardDescription>
                Current worker pool performance and status
              </CardDescription>
            </CardHeader>
            <CardContent>
              <div className="space-y-4">
                <div className="flex justify-between items-center">
                  <span className="text-sm font-medium">Total Workers</span>
                  <span className="text-2xl font-bold">{workers.length}</span>
                </div>
                <div className="flex justify-between items-center">
                  <span className="text-sm font-medium">Active Workers</span>
                  <span className="text-lg font-semibold text-green-600">
                    {workers.filter(w => w.status === 'busy').length}
                  </span>
                </div>
                <div className="flex justify-between items-center">
                  <span className="text-sm font-medium">Idle Workers</span>
                  <span className="text-lg font-semibold text-blue-600">
                    {workers.filter(w => w.status === 'idle').length}
                  </span>
                </div>
                <div className="flex justify-between items-center">
                  <span className="text-sm font-medium">Error Workers</span>
                  <span className="text-lg font-semibold text-red-600">
                    {workers.filter(w => w.status === 'error').length}
                  </span>
                </div>
              </div>
            </CardContent>
          </Card>

          <Card>
            <CardHeader>
              <CardTitle>Plugin Status</CardTitle>
              <CardDescription>
                Overview of plugin health and performance
              </CardDescription>
            </CardHeader>
            <CardContent>
              <div className="space-y-4">
                <div className="flex justify-between items-center">
                  <span className="text-sm font-medium">Total Plugins</span>
                  <span className="text-2xl font-bold">{plugins.length}</span>
                </div>
                <div className="flex justify-between items-center">
                  <span className="text-sm font-medium">Active Plugins</span>
                  <span className="text-lg font-semibold text-green-600">
                    {plugins.filter(p => p.status === 'active').length}
                  </span>
                </div>
                <div className="flex justify-between items-center">
                  <span className="text-sm font-medium">Inactive Plugins</span>
                  <span className="text-lg font-semibold text-gray-600">
                    {plugins.filter(p => p.status === 'inactive').length}
                  </span>
                </div>
                <div className="flex justify-between items-center">
                  <span className="text-sm font-medium">Error Plugins</span>
                  <span className="text-lg font-semibold text-red-600">
                    {plugins.filter(p => p.status === 'error').length}
                  </span>
                </div>
              </div>
            </CardContent>
          </Card>
        </div>

        {/* Last Updated */}
        <div className="text-center text-sm text-muted-foreground">
          Last updated: {formatTime(lastUpdated)}
        </div>
      </div>
    </Layout>
  )
}
