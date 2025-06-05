import { LogViewer } from '@/components/LogViewer'

export function Logs() {
  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-3xl font-bold tracking-tight">Real-time Logs</h1>
        <p className="text-muted-foreground">
          Monitor system logs in real-time with filtering and search capabilities.
        </p>
      </div>
      
      <LogViewer />
    </div>
  )
}
