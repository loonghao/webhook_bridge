import { useState, useEffect } from 'react'
import { Save, RefreshCw } from 'lucide-react'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { apiClient } from '@/services/api'

export function Configuration() {
  const [config, setConfig] = useState<any>(null)
  const [loading, setLoading] = useState(true)
  const [saving, setSaving] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const [success, setSuccess] = useState<string | null>(null)

  const loadConfig = async () => {
    try {
      setLoading(true)
      setError(null)
      const data = await apiClient.getConfig()
      setConfig(data)
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to load configuration')
    } finally {
      setLoading(false)
    }
  }

  const saveConfig = async () => {
    if (!config) return
    
    try {
      setSaving(true)
      setError(null)
      setSuccess(null)
      
      await apiClient.saveConfig(config)
      setSuccess('Configuration saved successfully')
      
      // Clear success message after 3 seconds
      setTimeout(() => setSuccess(null), 3000)
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to save configuration')
    } finally {
      setSaving(false)
    }
  }

  const handleInputChange = (section: string, key: string, value: any) => {
    setConfig((prev: any) => ({
      ...prev,
      [section]: {
        ...prev[section],
        [key]: value
      }
    }))
  }

  useEffect(() => {
    loadConfig()
  }, [])

  if (loading) {
    return (
      <div className="space-y-6">
        <div>
          <h1 className="text-3xl font-bold tracking-tight">Configuration</h1>
          <p className="text-muted-foreground">Loading configuration...</p>
        </div>
      </div>
    )
  }

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-3xl font-bold tracking-tight">Configuration</h1>
          <p className="text-muted-foreground">
            Manage system configuration and settings
          </p>
        </div>
        <div className="flex items-center space-x-2">
          <Button variant="outline" onClick={loadConfig} disabled={loading}>
            <RefreshCw className={`h-4 w-4 mr-2 ${loading ? 'animate-spin' : ''}`} />
            Refresh
          </Button>
          <Button onClick={saveConfig} disabled={saving || !config}>
            <Save className="h-4 w-4 mr-2" />
            {saving ? 'Saving...' : 'Save Changes'}
          </Button>
        </div>
      </div>

      {error && (
        <Card className="border-destructive">
          <CardContent className="pt-6">
            <div className="flex items-center space-x-2 text-destructive">
              <span className="text-sm">{error}</span>
            </div>
          </CardContent>
        </Card>
      )}

      {success && (
        <Card className="border-green-500">
          <CardContent className="pt-6">
            <div className="flex items-center space-x-2 text-green-600">
              <span className="text-sm">{success}</span>
            </div>
          </CardContent>
        </Card>
      )}

      {config && (
        <div className="grid gap-6">
          {/* Server Configuration */}
          <Card>
            <CardHeader>
              <CardTitle>Server Configuration</CardTitle>
              <CardDescription>
                Configure server host, port, and mode settings
              </CardDescription>
            </CardHeader>
            <CardContent className="space-y-4">
              <div className="grid grid-cols-2 gap-4">
                <div>
                  <label className="text-sm font-medium">Host</label>
                  <input
                    type="text"
                    value={config.server?.host || ''}
                    onChange={(e) => handleInputChange('server', 'host', e.target.value)}
                    className="w-full mt-1 px-3 py-2 border border-input rounded-md bg-background"
                    placeholder="0.0.0.0"
                  />
                </div>
                <div>
                  <label className="text-sm font-medium">Port</label>
                  <input
                    type="number"
                    value={config.server?.port || ''}
                    onChange={(e) => handleInputChange('server', 'port', parseInt(e.target.value))}
                    className="w-full mt-1 px-3 py-2 border border-input rounded-md bg-background"
                    placeholder="8000"
                  />
                </div>
              </div>
              <div>
                <label className="text-sm font-medium">Mode</label>
                <select
                  value={config.server?.mode || 'debug'}
                  onChange={(e) => handleInputChange('server', 'mode', e.target.value)}
                  className="w-full mt-1 px-3 py-2 border border-input rounded-md bg-background"
                >
                  <option value="debug">Debug</option>
                  <option value="release">Release</option>
                </select>
              </div>
            </CardContent>
          </Card>

          {/* Python Configuration */}
          <Card>
            <CardHeader>
              <CardTitle>Python Configuration</CardTitle>
              <CardDescription>
                Configure Python interpreter and virtual environment settings
              </CardDescription>
            </CardHeader>
            <CardContent className="space-y-4">
              <div>
                <label className="text-sm font-medium">Python Interpreter</label>
                <input
                  type="text"
                  value={config.python?.interpreter || ''}
                  onChange={(e) => handleInputChange('python', 'interpreter', e.target.value)}
                  className="w-full mt-1 px-3 py-2 border border-input rounded-md bg-background"
                  placeholder="python"
                />
              </div>
              <div>
                <label className="text-sm font-medium">Virtual Environment Path</label>
                <input
                  type="text"
                  value={config.python?.venv_path || ''}
                  onChange={(e) => handleInputChange('python', 'venv_path', e.target.value)}
                  className="w-full mt-1 px-3 py-2 border border-input rounded-md bg-background"
                  placeholder=".venv"
                />
              </div>
              <div className="flex items-center space-x-2">
                <input
                  type="checkbox"
                  id="auto_download_uv"
                  checked={config.python?.auto_download_uv || false}
                  onChange={(e) => handleInputChange('python', 'auto_download_uv', e.target.checked)}
                  className="rounded border-input"
                />
                <label htmlFor="auto_download_uv" className="text-sm font-medium">
                  Auto Download UV
                </label>
              </div>
            </CardContent>
          </Card>

          {/* Logging Configuration */}
          <Card>
            <CardHeader>
              <CardTitle>Logging Configuration</CardTitle>
              <CardDescription>
                Configure logging level, format, and output settings
              </CardDescription>
            </CardHeader>
            <CardContent className="space-y-4">
              <div className="grid grid-cols-2 gap-4">
                <div>
                  <label className="text-sm font-medium">Log Level</label>
                  <select
                    value={config.logging?.level || 'info'}
                    onChange={(e) => handleInputChange('logging', 'level', e.target.value)}
                    className="w-full mt-1 px-3 py-2 border border-input rounded-md bg-background"
                  >
                    <option value="debug">Debug</option>
                    <option value="info">Info</option>
                    <option value="warn">Warning</option>
                    <option value="error">Error</option>
                  </select>
                </div>
                <div>
                  <label className="text-sm font-medium">Log Format</label>
                  <select
                    value={config.logging?.format || 'text'}
                    onChange={(e) => handleInputChange('logging', 'format', e.target.value)}
                    className="w-full mt-1 px-3 py-2 border border-input rounded-md bg-background"
                  >
                    <option value="text">Text</option>
                    <option value="json">JSON</option>
                  </select>
                </div>
              </div>
              <div>
                <label className="text-sm font-medium">Log File Path</label>
                <input
                  type="text"
                  value={config.logging?.file || ''}
                  onChange={(e) => handleInputChange('logging', 'file', e.target.value)}
                  className="w-full mt-1 px-3 py-2 border border-input rounded-md bg-background"
                  placeholder="logs/webhook-bridge.log"
                />
              </div>
            </CardContent>
          </Card>

          {/* Directory Configuration */}
          <Card>
            <CardHeader>
              <CardTitle>Directory Configuration</CardTitle>
              <CardDescription>
                Configure working directories and data paths
              </CardDescription>
            </CardHeader>
            <CardContent className="space-y-4">
              <div className="grid grid-cols-2 gap-4">
                <div>
                  <label className="text-sm font-medium">Working Directory</label>
                  <input
                    type="text"
                    value={config.directories?.working_dir || ''}
                    onChange={(e) => handleInputChange('directories', 'working_dir', e.target.value)}
                    className="w-full mt-1 px-3 py-2 border border-input rounded-md bg-background"
                    placeholder="."
                  />
                </div>
                <div>
                  <label className="text-sm font-medium">Log Directory</label>
                  <input
                    type="text"
                    value={config.directories?.log_dir || ''}
                    onChange={(e) => handleInputChange('directories', 'log_dir', e.target.value)}
                    className="w-full mt-1 px-3 py-2 border border-input rounded-md bg-background"
                    placeholder="logs"
                  />
                </div>
              </div>
              <div className="grid grid-cols-2 gap-4">
                <div>
                  <label className="text-sm font-medium">Plugin Directory</label>
                  <input
                    type="text"
                    value={config.directories?.plugin_dir || ''}
                    onChange={(e) => handleInputChange('directories', 'plugin_dir', e.target.value)}
                    className="w-full mt-1 px-3 py-2 border border-input rounded-md bg-background"
                    placeholder="plugins"
                  />
                </div>
                <div>
                  <label className="text-sm font-medium">Data Directory</label>
                  <input
                    type="text"
                    value={config.directories?.data_dir || ''}
                    onChange={(e) => handleInputChange('directories', 'data_dir', e.target.value)}
                    className="w-full mt-1 px-3 py-2 border border-input rounded-md bg-background"
                    placeholder="data"
                  />
                </div>
              </div>
            </CardContent>
          </Card>
        </div>
      )}
    </div>
  )
}
