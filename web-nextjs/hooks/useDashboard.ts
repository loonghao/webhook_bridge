'use client'

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

  const fetchData = useCallback(async () => {
    try {
      setState(prev => ({ ...prev, loading: true, error: null }))

      // Fetch all data in parallel with retry logic
      const [stats, status, plugins, workers, logs, activity] = await Promise.allSettled([
        withRetry(() => apiClient.getStats()),
        withRetry(() => apiClient.getStatus()),
        withRetry(() => apiClient.getPlugins()),
        withRetry(() => apiClient.getWorkers()),
        withRetry(() => apiClient.getLogs({ limit: 50 })),
        withRetry(() => apiClient.getActivity(20)),
      ])

      setState(prev => ({
        ...prev,
        stats: stats.status === 'fulfilled' ? stats.value : prev.stats,
        status: status.status === 'fulfilled' ? status.value : prev.status,
        plugins: plugins.status === 'fulfilled' ? plugins.value : prev.plugins,
        workers: workers.status === 'fulfilled' ? workers.value : prev.workers,
        logs: logs.status === 'fulfilled' ? logs.value : prev.logs,
        activity: activity.status === 'fulfilled' ? activity.value : prev.activity,
        loading: false,
        lastUpdated: new Date(),
        error: null,
      }))

    } catch (error) {
      console.error('Dashboard data fetch failed:', error)
      setState(prev => ({
        ...prev,
        loading: false,
        error: error instanceof Error ? error.message : 'Failed to fetch dashboard data',
      }))
    }
  }, [])

  const refresh = useCallback(() => {
    fetchData()
  }, [fetchData])

  // Initial data fetch
  useEffect(() => {
    fetchData()
  }, [fetchData])

  // Auto-refresh setup
  useEffect(() => {
    if (!autoRefresh) return

    const interval = setInterval(() => {
      fetchData()
    }, refreshInterval)

    return () => clearInterval(interval)
  }, [autoRefresh, refreshInterval, fetchData])

  // Plugin management functions
  const enablePlugin = useCallback(async (pluginId: string) => {
    try {
      await apiClient.enablePlugin(pluginId)
      // Update local state optimistically
      setState(prev => ({
        ...prev,
        plugins: prev.plugins.map(plugin =>
          plugin.id === pluginId ? { ...plugin, status: 'active', enabled: true } : plugin
        ),
      }))
      // Refresh to get actual state
      setTimeout(fetchData, 1000)
    } catch (error) {
      console.error('Failed to enable plugin:', error)
      throw error
    }
  }, [fetchData])

  const disablePlugin = useCallback(async (pluginId: string) => {
    try {
      await apiClient.disablePlugin(pluginId)
      // Update local state optimistically
      setState(prev => ({
        ...prev,
        plugins: prev.plugins.map(plugin =>
          plugin.id === pluginId ? { ...plugin, status: 'inactive', enabled: false } : plugin
        ),
      }))
      // Refresh to get actual state
      setTimeout(fetchData, 1000)
    } catch (error) {
      console.error('Failed to disable plugin:', error)
      throw error
    }
  }, [fetchData])

  const executePlugin = useCallback(async (pluginId: string, payload?: any) => {
    try {
      const result = await apiClient.executePlugin(pluginId, payload)
      // Refresh data to show updated execution stats
      setTimeout(fetchData, 1000)
      return result
    } catch (error) {
      console.error('Failed to execute plugin:', error)
      throw error
    }
  }, [fetchData])

  // Worker management functions
  const submitJob = useCallback(async (jobData: any) => {
    try {
      const result = await apiClient.submitJob(jobData)
      // Refresh workers to show new job
      setTimeout(fetchData, 1000)
      return result
    } catch (error) {
      console.error('Failed to submit job:', error)
      throw error
    }
  }, [fetchData])

  return {
    // Data
    ...state,
    
    // Actions
    refresh,
    enablePlugin,
    disablePlugin,
    executePlugin,
    submitJob,
    
    // Computed values
    isHealthy: state.status?.status === 'healthy',
    activePluginsCount: state.plugins.filter(p => p.status === 'active').length,
    totalPluginsCount: state.plugins.length,
    busyWorkersCount: state.workers.filter(w => w.status === 'busy').length,
    totalWorkersCount: state.workers.length,
    recentErrors: state.logs.filter(log => 
      log.level === 'error' && 
      new Date(log.timestamp) > new Date(Date.now() - 24 * 60 * 60 * 1000)
    ).length,
  }
}

export default useDashboard
