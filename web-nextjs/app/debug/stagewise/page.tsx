'use client'

import { useState, useEffect } from 'react'
import { Layout } from '@/components/Layout'
import { StagewiseDebugger } from '@/components/StagewiseDebugger'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import { useStagewise } from '@/hooks/useStagewise'
import { apiClient } from '@/services/api'
import { 
  Play, 
  CheckCircle, 
  XCircle, 
  Loader2,
  Brain,
  Code,
  Network,
  Database
} from 'lucide-react'

export default function StagewiseDebugPage() {
  const stagewise = useStagewise()
  const [isRunningDemo, setIsRunningDemo] = useState(false)

  const runDemoWorkflow = async () => {
    setIsRunningDemo(true)
    
    try {
      // Start a demo session
      stagewise.startSession('API Integration Demo', 'Demonstrating stagewise debugging with API calls')
      
      // Stage 1: Environment Check
      const stage1 = stagewise.startStage('Environment Check', 'Verify environment and configuration')
      
      const step1_1 = stagewise.startStep(stage1, 'Check API Base URL', 'Verify API endpoint configuration')
      stagewise.logData(stage1, step1_1, { 
        apiBaseUrl: process.env.NEXT_PUBLIC_API_BASE_URL || 'relative',
        timestamp: new Date().toISOString()
      })
      stagewise.logMessage(stage1, step1_1, 'API base URL configured')
      stagewise.endStep(stage1, step1_1)
      
      const step1_2 = stagewise.startStep(stage1, 'Check Browser Environment', 'Verify browser capabilities')
      stagewise.logData(stage1, step1_2, {
        userAgent: navigator.userAgent,
        language: navigator.language,
        cookieEnabled: navigator.cookieEnabled,
        onLine: navigator.onLine
      })
      stagewise.endStep(stage1, step1_2)
      
      stagewise.endStage(stage1)
      
      // Stage 2: API Health Check
      const stage2 = stagewise.startStage('API Health Check', 'Test basic API connectivity')
      
      const step2_1 = stagewise.startStep(stage2, 'Health Endpoint', 'Test /health endpoint')
      try {
        const response = await fetch('/health')
        const data = await response.json()
        stagewise.logData(stage2, step2_1, { response: data, status: response.status })
        stagewise.logMessage(stage2, step2_1, `Health check ${response.ok ? 'passed' : 'failed'}`)
        stagewise.endStep(stage2, step2_1, response.ok ? undefined : 'Health check failed')
      } catch (error) {
        stagewise.endStep(stage2, step2_1, error instanceof Error ? error.message : 'Unknown error')
      }
      
      const step2_2 = stagewise.startStep(stage2, 'Dashboard Status', 'Test dashboard status endpoint')
      try {
        const status = await apiClient.getStatus()
        stagewise.logData(stage2, step2_2, status)
        stagewise.logMessage(stage2, step2_2, 'Dashboard status retrieved successfully')
        stagewise.endStep(stage2, step2_2)
      } catch (error) {
        stagewise.endStep(stage2, step2_2, error instanceof Error ? error.message : 'Unknown error')
      }
      
      stagewise.endStage(stage2)
      
      // Stage 3: Data Retrieval
      const stage3 = stagewise.startStage('Data Retrieval', 'Fetch application data')
      
      const step3_1 = stagewise.startStep(stage3, 'Get Statistics', 'Retrieve dashboard statistics')
      try {
        const stats = await apiClient.getStats()
        stagewise.logData(stage3, step3_1, stats)
        stagewise.logMessage(stage3, step3_1, 'Statistics retrieved successfully')
        stagewise.endStep(stage3, step3_1)
      } catch (error) {
        stagewise.endStep(stage3, step3_1, error instanceof Error ? error.message : 'Unknown error')
      }
      
      const step3_2 = stagewise.startStep(stage3, 'Get Plugins', 'Retrieve plugin information')
      try {
        const plugins = await apiClient.getPlugins()
        stagewise.logData(stage3, step3_2, { pluginCount: plugins.length, plugins })
        stagewise.logMessage(stage3, step3_2, `Retrieved ${plugins.length} plugins`)
        stagewise.endStep(stage3, step3_2)
      } catch (error) {
        stagewise.endStep(stage3, step3_2, error instanceof Error ? error.message : 'Unknown error')
      }
      
      const step3_3 = stagewise.startStep(stage3, 'Get Workers', 'Retrieve worker information')
      try {
        const workers = await apiClient.getWorkers()
        stagewise.logData(stage3, step3_3, workers)
        stagewise.logMessage(stage3, step3_3, 'Worker information retrieved')
        stagewise.endStep(stage3, step3_3)
      } catch (error) {
        stagewise.endStep(stage3, step3_3, error instanceof Error ? error.message : 'Unknown error')
      }
      
      stagewise.endStage(stage3)
      
      // Stage 4: Performance Test
      const stage4 = stagewise.startStage('Performance Test', 'Test multiple concurrent requests')
      
      const step4_1 = stagewise.startStep(stage4, 'Concurrent Requests', 'Make multiple API calls simultaneously')
      try {
        const startTime = performance.now()
        const promises = [
          apiClient.getStatus(),
          apiClient.getStats(),
          apiClient.getPlugins(),
          apiClient.getWorkers()
        ]
        
        const results = await Promise.allSettled(promises)
        const endTime = performance.now()
        
        stagewise.logData(stage4, step4_1, {
          totalTime: endTime - startTime,
          results: results.map(r => r.status),
          successCount: results.filter(r => r.status === 'fulfilled').length,
          errorCount: results.filter(r => r.status === 'rejected').length
        })
        
        stagewise.logMessage(stage4, step4_1, `Completed ${results.length} concurrent requests in ${(endTime - startTime).toFixed(2)}ms`)
        stagewise.endStep(stage4, step4_1)
      } catch (error) {
        stagewise.endStep(stage4, step4_1, error instanceof Error ? error.message : 'Unknown error')
      }
      
      stagewise.endStage(stage4)
      
      // End the session
      stagewise.endSession()
      
    } catch (error) {
      console.error('Demo workflow failed:', error)
      if (stagewise.session) {
        stagewise.cancelSession()
      }
    } finally {
      setIsRunningDemo(false)
    }
  }

  const runErrorDemo = async () => {
    setIsRunningDemo(true)
    
    try {
      stagewise.startSession('Error Handling Demo', 'Demonstrating error capture and debugging')
      
      const stage1 = stagewise.startStage('Error Scenarios', 'Test various error conditions')
      
      // Intentional error
      const step1_1 = stagewise.startStep(stage1, 'Network Error', 'Test network failure handling')
      try {
        await fetch('/nonexistent-endpoint')
      } catch (error) {
        stagewise.endStep(stage1, step1_1, 'Expected network error for demo')
      }
      
      // Console error
      const step1_2 = stagewise.startStep(stage1, 'Console Error', 'Generate console error')
      console.error('This is a demo error for stagewise debugging')
      console.warn('This is a demo warning')
      console.log('This is a demo log message')
      stagewise.endStep(stage1, step1_2)
      
      // JavaScript error
      const step1_3 = stagewise.startStep(stage1, 'JavaScript Error', 'Test JavaScript error handling')
      try {
        // @ts-ignore - intentional error for demo
        const result = undefined.someProperty.anotherProperty
      } catch (error) {
        stagewise.endStep(stage1, step1_3, error instanceof Error ? error.message : 'Unknown error')
      }
      
      stagewise.endStage(stage1)
      stagewise.endSession()
      
    } catch (error) {
      console.error('Error demo failed:', error)
      if (stagewise.session) {
        stagewise.cancelSession()
      }
    } finally {
      setIsRunningDemo(false)
    }
  }

  return (
    <Layout>
      <div className="space-y-6">
        <div>
          <h1 className="text-3xl font-bold tracking-tight">Stagewise Debugging</h1>
          <p className="text-muted-foreground">
            AI-assisted debugging with stage-wise execution tracking and comprehensive logging
          </p>
        </div>

        {/* Demo Controls */}
        <Card>
          <CardHeader>
            <CardTitle className="flex items-center space-x-2">
              <Brain className="h-5 w-5" />
              <span>AI Debugging Demos</span>
            </CardTitle>
            <CardDescription>
              Run demo workflows to see stagewise debugging in action
            </CardDescription>
          </CardHeader>
          <CardContent>
            <div className="flex space-x-4">
              <Button 
                onClick={runDemoWorkflow} 
                disabled={isRunningDemo}
                className="flex items-center space-x-2"
              >
                {isRunningDemo ? (
                  <Loader2 className="h-4 w-4 animate-spin" />
                ) : (
                  <Play className="h-4 w-4" />
                )}
                <span>Run API Demo</span>
              </Button>
              
              <Button 
                onClick={runErrorDemo} 
                disabled={isRunningDemo}
                variant="outline"
                className="flex items-center space-x-2"
              >
                {isRunningDemo ? (
                  <Loader2 className="h-4 w-4 animate-spin" />
                ) : (
                  <XCircle className="h-4 w-4" />
                )}
                <span>Run Error Demo</span>
              </Button>
            </div>
          </CardContent>
        </Card>

        {/* Features Overview */}
        <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-4">
          <Card>
            <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
              <CardTitle className="text-sm font-medium">Stage Tracking</CardTitle>
              <Code className="h-4 w-4 text-muted-foreground" />
            </CardHeader>
            <CardContent>
              <div className="text-2xl font-bold">{stagewise.session?.stages.length || 0}</div>
              <p className="text-xs text-muted-foreground">
                Execution stages tracked
              </p>
            </CardContent>
          </Card>
          
          <Card>
            <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
              <CardTitle className="text-sm font-medium">Network Requests</CardTitle>
              <Network className="h-4 w-4 text-muted-foreground" />
            </CardHeader>
            <CardContent>
              <div className="text-2xl font-bold">{stagewise.networkRequests.length}</div>
              <p className="text-xs text-muted-foreground">
                HTTP requests captured
              </p>
            </CardContent>
          </Card>
          
          <Card>
            <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
              <CardTitle className="text-sm font-medium">Console Entries</CardTitle>
              <Database className="h-4 w-4 text-muted-foreground" />
            </CardHeader>
            <CardContent>
              <div className="text-2xl font-bold">{stagewise.consoleEntries.length}</div>
              <p className="text-xs text-muted-foreground">
                Console messages logged
              </p>
            </CardContent>
          </Card>
          
          <Card>
            <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
              <CardTitle className="text-sm font-medium">Session Status</CardTitle>
              {stagewise.session ? (
                stagewise.session.status === 'running' ? (
                  <Loader2 className="h-4 w-4 text-blue-600 animate-spin" />
                ) : stagewise.session.status === 'success' ? (
                  <CheckCircle className="h-4 w-4 text-green-600" />
                ) : (
                  <XCircle className="h-4 w-4 text-red-600" />
                )
              ) : (
                <div className="h-4 w-4 rounded-full bg-gray-300" />
              )}
            </CardHeader>
            <CardContent>
              <div className="text-2xl font-bold">
                {stagewise.session ? (
                  <Badge variant={
                    stagewise.session.status === 'running' ? 'default' :
                    stagewise.session.status === 'success' ? 'default' :
                    'destructive'
                  }>
                    {stagewise.session.status}
                  </Badge>
                ) : (
                  <Badge variant="outline">inactive</Badge>
                )}
              </div>
              <p className="text-xs text-muted-foreground">
                Current session state
              </p>
            </CardContent>
          </Card>
        </div>

        {/* Main Debugger Component */}
        <StagewiseDebugger />
      </div>
    </Layout>
  )
}
