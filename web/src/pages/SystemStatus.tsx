import { Activity, Server, Users, Database, HardDrive, Cpu } from 'lucide-react'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { useDashboard } from '@/hooks/useDashboard'

export function SystemStatus() {
  const { status, workers, plugins, loading, error, lastUpdated } = useDashboard()

  const formatTime = (date: Date | null) => {
    if (!date) return 'Never'
    return date.toLocaleTimeString()
  }

  return (
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
            <CardTitle className="text-sm font-medium">gRPC Connection</CardTitle>
            <Activity className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">
              {loading ? '...' : status?.checks?.grpc?.status ? 'Connected' : 'Disconnected'}
            </div>
            <p className={`text-xs ${
              status?.checks?.grpc?.status ? 'text-green-600' : 'text-red-600'
            }`}>
              {status?.checks?.grpc?.message || 'Unknown'}
            </p>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">Database</CardTitle>
            <Database className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">
              {loading ? '...' : status?.checks?.database?.status ? 'Online' : 'Offline'}
            </div>
            <p className={`text-xs ${
              status?.checks?.database?.status ? 'text-green-600' : 'text-red-600'
            }`}>
              {status?.checks?.database?.message || 'Unknown'}
            </p>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">Storage</CardTitle>
            <HardDrive className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">
              {loading ? '...' : status?.checks?.storage?.status ? 'Available' : 'Unavailable'}
            </div>
            <p className={`text-xs ${
              status?.checks?.storage?.status ? 'text-green-600' : 'text-red-600'
            }`}>
              {status?.checks?.storage?.message || 'Unknown'}
            </p>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">Worker Pool</CardTitle>
            <Users className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">
              {loading ? '...' : workers.length}
            </div>
            <p className="text-xs text-muted-foreground">
              {workers.filter(w => w.status === 'idle').length} idle, {workers.filter(w => w.status === 'busy').length} busy
            </p>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">Active Plugins</CardTitle>
            <Cpu className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">
              {loading ? '...' : plugins.filter(p => p.status === 'active').length}
            </div>
            <p className="text-xs text-muted-foreground">
              {plugins.length} total plugins
            </p>
          </CardContent>
        </Card>
      </div>

      {/* Detailed Status */}
      <div className="grid gap-4 md:grid-cols-2">
        <Card>
          <CardHeader>
            <CardTitle>System Information</CardTitle>
            <CardDescription>
              Server details and runtime information
            </CardDescription>
          </CardHeader>
          <CardContent>
            <div className="space-y-4">
              <div className="flex justify-between">
                <span className="text-sm font-medium">Version</span>
                <span className="text-sm text-muted-foreground">{status?.version || 'Unknown'}</span>
              </div>
              <div className="flex justify-between">
                <span className="text-sm font-medium">Build</span>
                <span className="text-sm text-muted-foreground">{status?.build || 'Unknown'}</span>
              </div>
              <div className="flex justify-between">
                <span className="text-sm font-medium">Uptime</span>
                <span className="text-sm text-muted-foreground">{status?.uptime || 'Unknown'}</span>
              </div>
              <div className="flex justify-between">
                <span className="text-sm font-medium">Last Updated</span>
                <span className="text-sm text-muted-foreground">{formatTime(lastUpdated)}</span>
              </div>
            </div>
          </CardContent>
        </Card>

        <Card>
          <CardHeader>
            <CardTitle>Worker Details</CardTitle>
            <CardDescription>
              Individual worker status and performance
            </CardDescription>
          </CardHeader>
          <CardContent>
            <div className="space-y-4">
              {loading ? (
                <div className="text-center py-4 text-muted-foreground">Loading workers...</div>
              ) : workers.length > 0 ? (
                workers.map((worker) => (
                  <div key={worker.id} className="flex justify-between items-center">
                    <div>
                      <span className="text-sm font-medium">{worker.id}</span>
                      <p className="text-xs text-muted-foreground">
                        {worker.completedJobs}/{worker.totalJobs} jobs completed
                      </p>
                    </div>
                    <span className={`text-xs px-2 py-1 rounded-full ${
                      worker.status === 'idle' ? 'bg-green-100 text-green-800' :
                      worker.status === 'busy' ? 'bg-yellow-100 text-yellow-800' :
                      'bg-gray-100 text-gray-800'
                    }`}>
                      {worker.status}
                    </span>
                  </div>
                ))
              ) : (
                <div className="text-center py-4 text-muted-foreground">No workers found</div>
              )}
            </div>
          </CardContent>
        </Card>
      </div>
    </div>
  )
}
