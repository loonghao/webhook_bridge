import { useState, useEffect, useCallback } from 'react'
import { apiClient, withRetry } from '@/services/api'
import { 
  DashboardStats, 
  SystemStatus, 
  PluginInfo, 
  WorkerInfo, 
  LogEntry,
  ActivityEntry 
} from '@/types/api'

interface DashboardData {
  stats: DashboardStats | null
  status: SystemStatus | null
  plugins: PluginInfo[]
  workers: WorkerInfo[]
  logs: LogEntry[]
  activity: ActivityEntry[]
}

interface DashboardState extends DashboardData {
  loading: boolean
  error: string | null
  lastUpdated: Date | null
}

export function useDashboard(autoRefresh = true, refreshInterval = 30000) {
  const [state, setState] = useState<DashboardState>({
    stats: null,
    status: null,
    plugins: [],
    workers: [],
    logs: [],
    activity: [],
    loading: true,
    error: null,
    lastUpdated: null,
  })

  const loadData = useCallback(async () => {
    try {
      setState(prev => ({ ...prev, loading: true, error: null }))

      // Load all data in parallel
      const [stats, status, plugins, workers, logs, activity] = await Promise.allSettled([
        withRetry(() => apiClient.getStats()),
        withRetry(() => apiClient.getStatus()),
        withRetry(() => apiClient.getPlugins()),
        withRetry(() => apiClient.getWorkers()),
        withRetry(() => apiClient.getLogs({ limit: 50 })),
        withRetry(() => apiClient.getRecentActivity()),
      ])

      setState(prev => ({
        ...prev,
        stats: stats.status === 'fulfilled' ? stats.value : null,
        status: status.status === 'fulfilled' ? status.value : null,
        plugins: plugins.status === 'fulfilled' ? plugins.value : [],
        workers: workers.status === 'fulfilled' ? workers.value : [],
        logs: logs.status === 'fulfilled' ? logs.value : [],
        activity: activity.status === 'fulfilled' ? activity.value : [],
        loading: false,
        lastUpdated: new Date(),
      }))

      // Check for any errors
      const errors = [stats, status, plugins, workers, logs, activity]
        .filter(result => result.status === 'rejected')
        .map(result => (result as PromiseRejectedResult).reason.message)

      if (errors.length > 0) {
        setState(prev => ({ 
          ...prev, 
          error: `Some data failed to load: ${errors.join(', ')}` 
        }))
      }

    } catch (error) {
      setState(prev => ({
        ...prev,
        loading: false,
        error: error instanceof Error ? error.message : 'Failed to load dashboard data',
      }))
    }
  }, [])

  const refresh = useCallback(() => {
    loadData()
  }, [loadData])

  // Initial load
  useEffect(() => {
    loadData()
  }, [loadData])

  // Auto refresh
  useEffect(() => {
    if (!autoRefresh) return

    const interval = setInterval(() => {
      loadData()
    }, refreshInterval)

    return () => clearInterval(interval)
  }, [autoRefresh, refreshInterval, loadData])

  return {
    ...state,
    refresh,
  }
}
