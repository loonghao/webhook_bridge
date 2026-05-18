'use client'

import { Activity, Server, Users, Zap, RefreshCw, AlertCircle, ScrollText, Route, Settings, TestTube, Cpu, FileCode2 } from 'lucide-react'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import { SystemHealthBanner } from '@/components/SystemHealthBanner'
import { Layout } from '@/components/Layout'
import { useDashboard } from '@/hooks/useDashboard'

export default function Dashboard() {
  const { stats, status, plugins, workers, activity, loading, error, lastUpdated, refresh } = useDashboard()

  const formatTime = (date: Date | null) => {
    if (!date) return 'Never'
    return date.toLocaleTimeString()
  }

  return (
    <Layout>
      <div className="space-y-6">
        {/* System Health Banner */}
        <SystemHealthBanner showDetails={true} />

        <div className="flex items-center justify-between">
          <div>
            <div className="flex flex-wrap items-center gap-2 mb-3">
              <Badge variant="outline" className="border-slate-300 text-slate-700 dark:border-slate-700 dark:text-slate-300">
                <Cpu className="h-3.5 w-3.5 mr-1.5" />
                Rust 4.0 control plane
              </Badge>
              <Badge variant="outline" className="border-emerald-300 text-emerald-700 dark:border-emerald-800 dark:text-emerald-300">
                <FileCode2 className="h-3.5 w-3.5 mr-1.5" />
                Python hooks executor
              </Badge>
              <Badge variant="outline" className="border-cyan-300 text-cyan-700 dark:border-cyan-800 dark:text-cyan-300">
                <Route className="h-3.5 w-3.5 mr-1.5" />
                Request forwarding first
              </Badge>
            </div>
            <h1 className="text-4xl font-bold text-slate-950 dark:text-white mb-2">Webhook Bridge Dashboard</h1>
            <p className="text-lg text-slate-600 dark:text-slate-400">
              Operate webhook forwarding, Python hook transforms, and worker health from one compact console
            </p>
          </div>
          <div className="flex items-center space-x-4">
            <div className="text-right">
              <p className="text-sm text-slate-500 dark:text-slate-400">Last updated</p>
              <p className="text-sm font-medium text-slate-700 dark:text-slate-300">
                {formatTime(lastUpdated)}
              </p>
            </div>
            <Button
              className="modern-button"
              size="sm"
              onClick={refresh}
              disabled={loading}
            >
              <RefreshCw className={`h-4 w-4 mr-2 ${loading ? 'animate-spin' : ''}`} />
              Refresh
            </Button>
          </div>
        </div>

        {error && (
          <Card className="border-destructive">
            <CardContent className="pt-6">
              <div className="flex items-center space-x-2 text-destructive">
                <AlertCircle className="h-4 w-4" />
                <span className="text-sm">{error}</span>
              </div>
            </CardContent>
          </Card>
        )}

        {/* Stats Cards */}
        <div className="grid gap-6 md:grid-cols-2 lg:grid-cols-4">
          <div className="modern-card modern-stat-card blue p-6 hover:shadow-lg transition-shadow duration-200">
            <div className="flex items-center justify-between mb-4">
              <div className="p-3 rounded-xl bg-blue-100 dark:bg-blue-900/30">
                <Zap className="h-6 w-6 text-blue-600 dark:text-blue-400" />
              </div>
              <div className="text-right">
                <p className="text-sm font-medium text-slate-600 dark:text-slate-400">Total Requests</p>
                <p className="text-3xl font-bold text-slate-900 dark:text-white">
                  {loading ? '...' : stats?.totalRequests?.toLocaleString() || '0'}
                </p>
              </div>
            </div>
            <div className="flex items-center text-sm">
              <span className="text-green-600 dark:text-green-400 font-medium">
                {stats?.requestsGrowth || '+0%'}
              </span>
              <span className="text-slate-500 dark:text-slate-400 ml-1">from last month</span>
            </div>
          </div>

          <div className="modern-card modern-stat-card green p-6 hover:shadow-lg transition-shadow duration-200">
            <div className="flex items-center justify-between mb-4">
              <div className="p-3 rounded-xl bg-green-100 dark:bg-green-900/30">
                <Server className="h-6 w-6 text-green-600 dark:text-green-400" />
              </div>
              <div className="text-right">
                <p className="text-sm font-medium text-slate-600 dark:text-slate-400">Active Routes</p>
                <p className="text-3xl font-bold text-slate-900 dark:text-white">
                  {loading ? '...' : plugins.filter(p => p.status === 'active').length}
                </p>
              </div>
            </div>
            <div className="flex items-center text-sm">
              <span className="text-green-600 dark:text-green-400 font-medium">
                {stats?.pluginsGrowth || '+0'}
              </span>
              <span className="text-slate-500 dark:text-slate-400 ml-1">new this week</span>
            </div>
          </div>

          <div className="modern-card modern-stat-card purple p-6 hover:shadow-lg transition-shadow duration-200">
            <div className="flex items-center justify-between mb-4">
              <div className="p-3 rounded-xl bg-purple-100 dark:bg-purple-900/30">
                <Users className="h-6 w-6 text-purple-600 dark:text-purple-400" />
              </div>
              <div className="text-right">
                <p className="text-sm font-medium text-slate-600 dark:text-slate-400">Active Workers</p>
                <p className="text-3xl font-bold text-slate-900 dark:text-white">
                  {loading ? '...' : workers.length}
                </p>
              </div>
            </div>
            <div className="flex items-center text-sm">
              <span className="text-green-600 dark:text-green-400 font-medium">
                {stats?.workersStatus || 'All operational'}
              </span>
            </div>
          </div>

          <div className="modern-card modern-stat-card orange p-6 hover:shadow-lg transition-shadow duration-200">
            <div className="flex items-center justify-between mb-4">
              <div className="p-3 rounded-xl bg-orange-100 dark:bg-orange-900/30">
                <Activity className="h-6 w-6 text-orange-600 dark:text-orange-400" />
              </div>
              <div className="text-right">
                <p className="text-sm font-medium text-slate-600 dark:text-slate-400">System Uptime</p>
                <p className="text-3xl font-bold text-slate-900 dark:text-white">
                  {loading ? '...' : stats?.uptime || 'Unknown'}
                </p>
              </div>
            </div>
            <div className="flex items-center text-sm">
              <span className="text-green-600 dark:text-green-400 font-medium">
                {stats?.uptimePercentage || '99.9%'}
              </span>
              <span className="text-slate-500 dark:text-slate-400 ml-1">availability</span>
            </div>
          </div>
        </div>

        {/* Activity and Quick Actions */}
        <div className="grid gap-6 md:grid-cols-2">
          <div className="modern-card p-6">
            <div className="flex items-center justify-between mb-6">
              <div>
                <h3 className="text-xl font-semibold text-slate-900 dark:text-white">Recent Activity</h3>
                <p className="text-sm text-slate-600 dark:text-slate-400">Latest webhook processing events</p>
              </div>
              <Activity className="h-5 w-5 text-blue-500" />
            </div>
            <div className="space-y-3">
              {loading ? (
                <div className="text-center py-8 text-slate-500">
                  <div className="animate-spin h-6 w-6 border-2 border-blue-500 border-t-transparent rounded-full mx-auto mb-2"></div>
                  Loading activity...
                </div>
              ) : activity.length > 0 ? (
                activity.slice(0, 5).map((item) => (
                  <div key={item.id} className="flex items-center justify-between p-3 bg-slate-50 dark:bg-slate-800/50 rounded-lg">
                    <span className="text-sm font-medium text-slate-700 dark:text-slate-300">{item.title}</span>
                    <Badge variant={
                      item.status === 'success' ? 'default' :
                      item.status === 'error' ? 'destructive' :
                      item.status === 'warning' ? 'secondary' : 'outline'
                    } className="text-xs">
                      {item.status}
                    </Badge>
                  </div>
                ))
              ) : (
                <div className="text-center py-8 text-slate-500">
                  <Activity className="h-12 w-12 mx-auto mb-3 text-slate-300" />
                  No recent activity
                </div>
              )}
            </div>
          </div>

          <div className="modern-card p-6">
            <div className="flex items-center justify-between mb-6">
              <div>
                <h3 className="text-xl font-semibold text-slate-900 dark:text-white">Quick Actions</h3>
                <p className="text-sm text-slate-600 dark:text-slate-400">Common management tasks</p>
              </div>
              <Zap className="h-5 w-5 text-purple-500" />
            </div>
            <div className="grid gap-3">
              <Button className="w-full justify-start h-12 bg-blue-50 dark:bg-blue-900 dark:bg-opacity-20 border border-blue-200 dark:border-blue-800 text-blue-700 dark:text-blue-300 hover:bg-blue-100 dark:hover:bg-blue-900 dark:hover:bg-opacity-30" variant="outline">
                <ScrollText className="h-4 w-4 mr-3" />
                View Logs
              </Button>
              <Button className="w-full justify-start h-12 bg-green-50 dark:bg-green-900 dark:bg-opacity-20 border border-green-200 dark:border-green-800 text-green-700 dark:text-green-300 hover:bg-green-100 dark:hover:bg-green-900 dark:hover:bg-opacity-30" variant="outline">
                <Route className="h-4 w-4 mr-3" />
                Manage Routes
              </Button>
              <Button className="w-full justify-start h-12 bg-purple-50 dark:bg-purple-900 dark:bg-opacity-20 border border-purple-200 dark:border-purple-800 text-purple-700 dark:text-purple-300 hover:bg-purple-100 dark:hover:bg-purple-900 dark:hover:bg-opacity-30" variant="outline">
                <Settings className="h-4 w-4 mr-3" />
                System Configuration
              </Button>
              <Button className="w-full justify-start h-12 bg-orange-50 dark:bg-orange-900 dark:bg-opacity-20 border border-orange-200 dark:border-orange-800 text-orange-700 dark:text-orange-300 hover:bg-orange-100 dark:hover:bg-orange-900 dark:hover:bg-opacity-30" variant="outline">
                <TestTube className="h-4 w-4 mr-3" />
                API Testing
              </Button>
            </div>
          </div>
        </div>
      </div>
    </Layout>
  )
}
