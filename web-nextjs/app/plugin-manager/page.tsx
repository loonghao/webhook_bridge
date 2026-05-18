'use client'

import { useEffect, useMemo, useState } from 'react'
import {
  Activity,
  AlertCircle,
  CheckCircle2,
  ChevronRight,
  Copy,
  FileCode2,
  Play,
  RefreshCw,
  Route,
  Server,
  ShieldCheck,
  TerminalSquare,
} from 'lucide-react'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Card, CardContent } from '@/components/ui/card'
import { Textarea } from '@/components/ui/textarea'
import { Layout } from '@/components/Layout'
import { apiClient } from '@/services/api'
import { PluginInfo } from '@/types/api'
import { cn } from '@/lib/utils'

type ExecutionState = {
  plugin: string
  ok: boolean
  message: string
  data: unknown
  duration?: number
} | null

const samplePayload = JSON.stringify(
  {
    event: 'deployment.finished',
    project: 'webhook-bridge',
    actor: 'dashboard',
  },
  null,
  2
)

export default function PluginManager() {
  const [plugins, setPlugins] = useState<PluginInfo[]>([])
  const [selectedPlugin, setSelectedPlugin] = useState<string>('')
  const [payload, setPayload] = useState(samplePayload)
  const [loading, setLoading] = useState(true)
  const [executing, setExecuting] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const [result, setResult] = useState<ExecutionState>(null)

  const selected = useMemo(
    () => plugins.find(plugin => plugin.name === selectedPlugin) || plugins[0],
    [plugins, selectedPlugin]
  )

  const activeCount = plugins.filter(plugin => plugin.status === 'active').length
  const methodCount = new Set(plugins.flatMap(plugin => plugin.supported_methods || [])).size

  const refresh = async () => {
    try {
      setLoading(true)
      setError(null)
      const nextPlugins = await apiClient.getPlugins()
      setPlugins(nextPlugins)
      setSelectedPlugin(current =>
        current && nextPlugins.some(plugin => plugin.name === current)
          ? current
          : nextPlugins[0]?.name || ''
      )
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to load routes')
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => {
    refresh()
  }, [])

  const executeSelected = async () => {
    if (!selected) return

    try {
      setExecuting(true)
      setError(null)
      const parsedPayload = payload.trim() ? JSON.parse(payload) : {}
      const started = performance.now()
      const response = await apiClient.executePlugin(selected.name, parsedPayload)

      setResult({
        plugin: selected.name,
        ok: response.success !== false,
        message: response.message || 'Webhook processed',
        data: response.data || response,
        duration: Math.round(performance.now() - started),
      })
      await refresh()
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Webhook execution failed')
    } finally {
      setExecuting(false)
    }
  }

  const copyCurl = async () => {
    if (!selected) return
    const curl = `curl -X POST http://127.0.0.1:8080/api/webhook/${selected.name} -H "Content-Type: application/json" -d '${payload.replace(/\s+/g, ' ')}'`
    await navigator.clipboard.writeText(curl)
  }

  return (
    <Layout>
      <div className="space-y-6">
        <section className="overflow-hidden rounded-lg border border-slate-200 bg-white shadow-sm dark:border-slate-800 dark:bg-slate-950">
          <div className="grid gap-0 lg:grid-cols-[1.1fr_0.9fr]">
            <div className="p-6 md:p-8">
              <div className="flex flex-wrap items-center gap-2">
                <Badge variant="outline" className="border-cyan-300 text-cyan-700 dark:border-cyan-900 dark:text-cyan-300">
                  <Server className="mr-1.5 h-3.5 w-3.5" />
                  Rust API online
                </Badge>
                <Badge variant="outline" className="border-slate-300 text-slate-700 dark:border-slate-700 dark:text-slate-300">
                  <Route className="mr-1.5 h-3.5 w-3.5" />
                  Request forwarding
                </Badge>
                <Badge variant="outline" className="border-emerald-300 text-emerald-700 dark:border-emerald-900 dark:text-emerald-300">
                  <FileCode2 className="mr-1.5 h-3.5 w-3.5" />
                  Python hooks
                </Badge>
              </div>
              <h1 className="mt-5 text-3xl font-semibold tracking-tight text-slate-950 dark:text-white">
                Webhook Routes
              </h1>
              <p className="mt-2 max-w-2xl text-sm leading-6 text-slate-600 dark:text-slate-400">
                Route incoming webhooks to downstream services first, then use Python hooks when a payload needs local logic or transformation.
              </p>
            </div>
            <div className="grid grid-cols-3 border-t border-slate-200 bg-slate-950 text-white lg:border-l lg:border-t-0 dark:border-slate-800">
              <Metric label="Routes" value={plugins.length} />
              <Metric label="Active" value={activeCount} />
              <Metric label="Methods" value={methodCount} />
            </div>
          </div>
        </section>

        {error && (
          <div className="flex items-center gap-2 rounded-lg border border-red-200 bg-red-50 px-4 py-3 text-sm text-red-700 dark:border-red-900 dark:bg-red-950/40 dark:text-red-300">
            <AlertCircle className="h-4 w-4" />
            <span>{error}</span>
          </div>
        )}

        <div className="grid gap-6 xl:grid-cols-[minmax(280px,380px)_1fr]">
          <section className="rounded-lg border border-slate-200 bg-white shadow-sm dark:border-slate-800 dark:bg-slate-950">
            <div className="flex items-center justify-between border-b border-slate-200 px-4 py-3 dark:border-slate-800">
              <div>
                <h2 className="text-sm font-semibold text-slate-950 dark:text-white">Ingress Routes</h2>
                <p className="text-xs text-slate-500">Forwarding routes and Python hooks</p>
              </div>
              <Button variant="outline" size="sm" onClick={refresh} disabled={loading}>
                <RefreshCw className={cn('h-4 w-4', loading && 'animate-spin')} />
              </Button>
            </div>
            <div className="divide-y divide-slate-100 dark:divide-slate-900">
              {plugins.map(plugin => (
                <button
                  key={plugin.name}
                  type="button"
                  onClick={() => setSelectedPlugin(plugin.name)}
                  className={cn(
                    'flex w-full items-center gap-3 px-4 py-4 text-left transition-colors hover:bg-slate-50 dark:hover:bg-slate-900/70',
                    selected?.name === plugin.name && 'bg-cyan-50/70 dark:bg-cyan-950/20'
                  )}
                >
                  <div className="flex h-10 w-10 shrink-0 items-center justify-center rounded-md border border-slate-200 bg-slate-50 dark:border-slate-800 dark:bg-slate-900">
                    {plugin.type === 'forward' ? (
                      <Route className="h-5 w-5 text-cyan-600 dark:text-cyan-300" />
                    ) : (
                      <FileCode2 className="h-5 w-5 text-slate-700 dark:text-slate-300" />
                    )}
                  </div>
                  <div className="min-w-0 flex-1">
                    <div className="flex items-center gap-2">
                      <p className="truncate text-sm font-semibold text-slate-950 dark:text-white">{plugin.name}</p>
                      <StatusBadge status={plugin.status || 'inactive'} />
                    </div>
                    <p className="mt-1 truncate text-xs text-slate-500">{plugin.path}</p>
                  </div>
                  <ChevronRight className="h-4 w-4 text-slate-400" />
                </button>
              ))}
              {!loading && plugins.length === 0 && (
                <div className="px-4 py-10 text-center text-sm text-slate-500">
                  No routes or hook files were discovered.
                </div>
              )}
            </div>
          </section>

          <section className="grid gap-6 lg:grid-cols-[1fr_360px]">
            <Card className="overflow-hidden rounded-lg border-slate-200 shadow-sm dark:border-slate-800">
              <CardContent className="p-0">
                <div className="border-b border-slate-200 p-5 dark:border-slate-800">
                  <div className="flex flex-wrap items-start justify-between gap-3">
                    <div>
                      <h2 className="text-xl font-semibold text-slate-950 dark:text-white">
                        {selected?.name || 'Select a route'}
                      </h2>
                      <p className="mt-1 text-sm text-slate-500">
                        {selected?.description || 'Choose a forwarding route or Python hook to inspect and execute.'}
                      </p>
                    </div>
                    {selected && <StatusBadge status={selected.status || 'inactive'} />}
                  </div>
                </div>

                <div className="grid gap-0 md:grid-cols-3">
                  <InfoCell label="Mode" value={selected?.type === 'forward' ? 'forward' : selected?.type === 'script-group' ? 'script fanout' : selected?.type === 'powershell' || selected?.type === 'pwsh' ? 'PowerShell script' : 'python hook'} />
                  <InfoCell label="Version" value={selected?.version || '1.0.0'} />
                  <InfoCell label="Success Rate" value={`${selected?.successRate || 100}%`} />
                </div>

                <div className="space-y-4 border-t border-slate-200 p-5 dark:border-slate-800">
                  <div>
                    <div className="mb-2 flex items-center justify-between">
                      <label className="text-sm font-medium text-slate-800 dark:text-slate-200">Payload</label>
                      <Button variant="outline" size="sm" onClick={copyCurl} disabled={!selected}>
                        <Copy className="mr-2 h-4 w-4" />
                        Copy cURL
                      </Button>
                    </div>
                    <Textarea
                      value={payload}
                      onChange={event => setPayload(event.target.value)}
                      className="min-h-[220px] resize-y rounded-md border-slate-300 bg-slate-950 font-mono text-xs leading-5 text-slate-100 shadow-inner dark:border-slate-800"
                      spellCheck={false}
                    />
                  </div>
                  <Button
                    className="h-11 w-full bg-slate-950 text-white hover:bg-slate-800 dark:bg-cyan-500 dark:text-slate-950 dark:hover:bg-cyan-400"
                    onClick={executeSelected}
                    disabled={!selected || executing}
                  >
                    {executing ? <RefreshCw className="mr-2 h-4 w-4 animate-spin" /> : <Play className="mr-2 h-4 w-4" />}
                    {selected?.type === 'forward' ? 'Forward webhook' : selected?.type === 'script-group' ? 'Run script fanout' : selected?.type === 'powershell' || selected?.type === 'pwsh' ? 'Run PowerShell script' : 'Execute Python hook'}
                  </Button>
                </div>
              </CardContent>
            </Card>

            <div className="space-y-6">
              <Card className="rounded-lg border-slate-200 shadow-sm dark:border-slate-800">
                <CardContent className="p-5">
                  <h3 className="flex items-center gap-2 text-sm font-semibold text-slate-950 dark:text-white">
                    <ShieldCheck className="h-4 w-4 text-emerald-500" />
                    Supported Methods
                  </h3>
                  <div className="mt-4 flex flex-wrap gap-2">
                    {(selected?.supported_methods?.length ? selected.supported_methods : ['POST']).map(method => (
                      <Badge key={method} variant="outline" className="rounded-md border-slate-300 px-2.5 py-1 font-mono dark:border-slate-700">
                        {method}
                      </Badge>
                    ))}
                  </div>
                </CardContent>
              </Card>

              <Card className="rounded-lg border-slate-200 shadow-sm dark:border-slate-800">
                <CardContent className="p-5">
                  <h3 className="flex items-center gap-2 text-sm font-semibold text-slate-950 dark:text-white">
                    <TerminalSquare className="h-4 w-4 text-cyan-500" />
                    Execution Result
                  </h3>
                  {result ? (
                    <div className="mt-4 space-y-3">
                      <div className="flex items-center justify-between rounded-md bg-slate-100 px-3 py-2 text-sm dark:bg-slate-900">
                        <span className="flex items-center gap-2">
                          {result.ok ? <CheckCircle2 className="h-4 w-4 text-emerald-500" /> : <AlertCircle className="h-4 w-4 text-red-500" />}
                          {result.plugin}
                        </span>
                        <span className="text-xs text-slate-500">{result.duration}ms</span>
                      </div>
                      <pre className="max-h-[320px] overflow-auto rounded-md bg-slate-950 p-3 text-xs leading-5 text-slate-100">
                        {JSON.stringify(result.data, null, 2)}
                      </pre>
                    </div>
                  ) : (
                    <div className="mt-4 rounded-md border border-dashed border-slate-300 p-6 text-center text-sm text-slate-500 dark:border-slate-700">
                      <Activity className="mx-auto mb-2 h-6 w-6" />
                      Send a webhook to see the response.
                    </div>
                  )}
                </CardContent>
              </Card>
            </div>
          </section>
        </div>
      </div>
    </Layout>
  )
}

function Metric({ label, value }: { label: string; value: number }) {
  return (
    <div className="border-r border-white/10 p-5 last:border-r-0">
      <div className="text-3xl font-semibold tabular-nums">{value}</div>
      <div className="mt-1 text-xs uppercase tracking-wide text-slate-400">{label}</div>
    </div>
  )
}

function InfoCell({ label, value }: { label: string; value: string }) {
  return (
    <div className="border-r border-slate-200 p-5 last:border-r-0 dark:border-slate-800">
      <div className="text-xs uppercase tracking-wide text-slate-500">{label}</div>
      <div className="mt-1 truncate text-sm font-semibold text-slate-950 dark:text-white">{value}</div>
    </div>
  )
}

function StatusBadge({ status }: { status: string }) {
  const active = status === 'active'
  const error = status === 'error'

  return (
    <Badge
      variant="outline"
      className={cn(
        'rounded-md px-2 py-0.5 text-[11px]',
        active && 'border-emerald-300 bg-emerald-50 text-emerald-700 dark:border-emerald-900 dark:bg-emerald-950/40 dark:text-emerald-300',
        error && 'border-red-300 bg-red-50 text-red-700 dark:border-red-900 dark:bg-red-950/40 dark:text-red-300',
        !active && !error && 'border-slate-300 bg-slate-50 text-slate-600 dark:border-slate-700 dark:bg-slate-900 dark:text-slate-300'
      )}
    >
      {status}
    </Badge>
  )
}
