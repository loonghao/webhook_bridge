import { useState, useEffect } from 'react'
import { Plus, Play, Trash2, CheckCircle, XCircle, RefreshCw, Search, Settings } from 'lucide-react'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import { apiClient } from '@/services/api'

interface InterpreterInfo {
  name: string
  path: string
  status: 'ready' | 'validating' | 'error' | 'unavailable'
  validated: boolean
  last_validated?: string
  version?: string
  use_uv: boolean
  venv_path?: string
  required_packages: string[]
  validation_error?: string
}

interface InterpretersData {
  active: string
  interpreters: Record<string, InterpreterInfo>
}

export function PythonInterpreters() {
  const [interpreters, setInterpreters] = useState<InterpretersData | null>(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [showAddForm, setShowAddForm] = useState(false)
  const [discovering, setDiscovering] = useState(false)

  const loadInterpreters = async () => {
    try {
      setLoading(true)
      setError(null)
      const data = await apiClient.get('/api/dashboard/interpreters')
      setInterpreters(data.data)
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to load interpreters')
    } finally {
      setLoading(false)
    }
  }

  const activateInterpreter = async (name: string) => {
    try {
      await apiClient.post(`/api/dashboard/interpreters/${name}/activate`)
      await loadInterpreters() // Reload data
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to activate interpreter')
    }
  }

  const validateInterpreter = async (name: string) => {
    try {
      await apiClient.post(`/api/dashboard/interpreters/${name}/validate`)
      await loadInterpreters() // Reload data
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to validate interpreter')
    }
  }

  const removeInterpreter = async (name: string) => {
    if (!confirm(`Are you sure you want to remove interpreter "${name}"?`)) {
      return
    }

    try {
      await apiClient.delete(`/api/dashboard/interpreters/${name}`)
      await loadInterpreters() // Reload data
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to remove interpreter')
    }
  }

  const discoverInterpreters = async () => {
    try {
      setDiscovering(true)
      const data = await apiClient.get('/api/dashboard/interpreters/discover')
      console.log('Discovered interpreters:', data.data)
      // TODO: Show discovered interpreters in a modal or add them automatically
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to discover interpreters')
    } finally {
      setDiscovering(false)
    }
  }

  useEffect(() => {
    loadInterpreters()
  }, [])

  const getStatusBadge = (status: string, validated: boolean) => {
    if (status === 'ready' && validated) {
      return <Badge className="bg-green-100 text-green-800"><CheckCircle className="w-3 h-3 mr-1" />Ready</Badge>
    } else if (status === 'validating') {
      return <Badge className="bg-yellow-100 text-yellow-800"><RefreshCw className="w-3 h-3 mr-1 animate-spin" />Validating</Badge>
    } else if (status === 'error') {
      return <Badge className="bg-red-100 text-red-800"><XCircle className="w-3 h-3 mr-1" />Error</Badge>
    } else {
      return <Badge className="bg-gray-100 text-gray-800">Unknown</Badge>
    }
  }

  if (loading) {
    return (
      <div className="space-y-6">
        <div>
          <h1 className="text-3xl font-bold tracking-tight">Python Interpreters</h1>
          <p className="text-muted-foreground">Loading interpreters...</p>
        </div>
      </div>
    )
  }

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-3xl font-bold tracking-tight">Python Interpreters</h1>
          <p className="text-muted-foreground">
            Manage Python interpreters and their configurations
          </p>
        </div>
        <div className="flex items-center space-x-2">
          <Button 
            variant="outline" 
            onClick={discoverInterpreters} 
            disabled={discovering}
          >
            <Search className={`h-4 w-4 mr-2 ${discovering ? 'animate-spin' : ''}`} />
            {discovering ? 'Discovering...' : 'Discover'}
          </Button>
          <Button variant="outline" onClick={loadInterpreters} disabled={loading}>
            <RefreshCw className={`h-4 w-4 mr-2 ${loading ? 'animate-spin' : ''}`} />
            Refresh
          </Button>
          <Button onClick={() => setShowAddForm(true)}>
            <Plus className="h-4 w-4 mr-2" />
            Add Interpreter
          </Button>
        </div>
      </div>

      {error && (
        <Card className="border-destructive">
          <CardContent className="pt-6">
            <div className="flex items-center space-x-2 text-destructive">
              <XCircle className="h-4 w-4" />
              <span className="text-sm">{error}</span>
            </div>
          </CardContent>
        </Card>
      )}

      {interpreters && (
        <div className="space-y-4">
          {/* Active Interpreter Info */}
          {interpreters.active && (
            <Card className="border-green-200 bg-green-50">
              <CardHeader>
                <CardTitle className="text-green-800">Active Interpreter</CardTitle>
                <CardDescription className="text-green-600">
                  Currently active Python interpreter: <strong>{interpreters.active}</strong>
                </CardDescription>
              </CardHeader>
            </Card>
          )}

          {/* Interpreters List */}
          <div className="grid gap-4">
            {Object.entries(interpreters.interpreters).map(([key, interpreter]) => (
              <Card key={key} className={interpreter.status === 'ready' && interpreter.validated ? 'border-green-200' : ''}>
                <CardHeader>
                  <div className="flex items-center justify-between">
                    <div className="flex items-center space-x-3">
                      <CardTitle className="text-lg">{interpreter.name}</CardTitle>
                      {getStatusBadge(interpreter.status, interpreter.validated)}
                      {interpreters.active === key && (
                        <Badge className="bg-blue-100 text-blue-800">Active</Badge>
                      )}
                    </div>
                    <div className="flex items-center space-x-2">
                      <Button
                        variant="outline"
                        size="sm"
                        onClick={() => validateInterpreter(key)}
                        disabled={interpreter.status === 'validating'}
                      >
                        <RefreshCw className={`h-3 w-3 mr-1 ${interpreter.status === 'validating' ? 'animate-spin' : ''}`} />
                        Validate
                      </Button>
                      {interpreters.active !== key && (
                        <Button
                          variant="outline"
                          size="sm"
                          onClick={() => activateInterpreter(key)}
                        >
                          <Play className="h-3 w-3 mr-1" />
                          Activate
                        </Button>
                      )}
                      <Button
                        variant="outline"
                        size="sm"
                        onClick={() => removeInterpreter(key)}
                        disabled={interpreters.active === key}
                      >
                        <Trash2 className="h-3 w-3 mr-1" />
                        Remove
                      </Button>
                    </div>
                  </div>
                  <CardDescription>
                    <div className="space-y-1">
                      <div><strong>Path:</strong> {interpreter.path}</div>
                      {interpreter.version && <div><strong>Version:</strong> {interpreter.version}</div>}
                      {interpreter.venv_path && <div><strong>Virtual Environment:</strong> {interpreter.venv_path}</div>}
                      {interpreter.use_uv && <div><strong>Uses UV:</strong> Yes</div>}
                    </div>
                  </CardDescription>
                </CardHeader>
                <CardContent>
                  <div className="space-y-3">
                    {interpreter.validation_error && (
                      <div className="p-3 bg-red-50 border border-red-200 rounded-md">
                        <div className="text-sm text-red-800">
                          <strong>Validation Error:</strong> {interpreter.validation_error}
                        </div>
                      </div>
                    )}
                    
                    {interpreter.required_packages && interpreter.required_packages.length > 0 && (
                      <div>
                        <div className="text-sm font-medium mb-2">Required Packages:</div>
                        <div className="flex flex-wrap gap-1">
                          {interpreter.required_packages.map((pkg, index) => (
                            <Badge key={index} variant="secondary" className="text-xs">
                              {pkg}
                            </Badge>
                          ))}
                        </div>
                      </div>
                    )}

                    {interpreter.last_validated && (
                      <div className="text-xs text-muted-foreground">
                        Last validated: {new Date(interpreter.last_validated).toLocaleString()}
                      </div>
                    )}
                  </div>
                </CardContent>
              </Card>
            ))}
          </div>

          {Object.keys(interpreters.interpreters).length === 0 && (
            <Card>
              <CardContent className="pt-6">
                <div className="text-center py-8">
                  <Settings className="h-12 w-12 mx-auto text-muted-foreground mb-4" />
                  <h3 className="text-lg font-medium mb-2">No Interpreters Configured</h3>
                  <p className="text-muted-foreground mb-4">
                    Add a Python interpreter to get started with webhook execution.
                  </p>
                  <Button onClick={() => setShowAddForm(true)}>
                    <Plus className="h-4 w-4 mr-2" />
                    Add Your First Interpreter
                  </Button>
                </div>
              </CardContent>
            </Card>
          )}
        </div>
      )}

      {/* TODO: Add Interpreter Form Modal */}
      {showAddForm && (
        <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center p-4 z-50">
          <Card className="w-full max-w-md">
            <CardHeader>
              <CardTitle>Add Python Interpreter</CardTitle>
              <CardDescription>
                Configure a new Python interpreter for webhook execution
              </CardDescription>
            </CardHeader>
            <CardContent>
              <div className="space-y-4">
                <div>
                  <label className="text-sm font-medium">Name</label>
                  <input
                    type="text"
                    className="w-full mt-1 px-3 py-2 border border-input rounded-md bg-background"
                    placeholder="e.g., Python 3.11"
                  />
                </div>
                <div>
                  <label className="text-sm font-medium">Path</label>
                  <input
                    type="text"
                    className="w-full mt-1 px-3 py-2 border border-input rounded-md bg-background"
                    placeholder="e.g., /usr/bin/python3"
                  />
                </div>
                <div className="flex items-center space-x-2">
                  <input
                    type="checkbox"
                    id="use_uv"
                    className="rounded border-input"
                  />
                  <label htmlFor="use_uv" className="text-sm font-medium">
                    Use UV virtual environment
                  </label>
                </div>
                <div className="flex justify-end space-x-2">
                  <Button variant="outline" onClick={() => setShowAddForm(false)}>
                    Cancel
                  </Button>
                  <Button onClick={() => setShowAddForm(false)}>
                    Add Interpreter
                  </Button>
                </div>
              </div>
            </CardContent>
          </Card>
        </div>
      )}
    </div>
  )
}
