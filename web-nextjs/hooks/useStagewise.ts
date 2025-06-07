'use client'

import { useState, useCallback, useRef, useEffect } from 'react'
import { v4 as uuidv4 } from 'uuid'
import type {
  StagewiseSession,
  Stage,
  StageStep,
  StagewiseConfig,
  NetworkRequest,
  ConsoleEntry,
  PerformanceMetric,
  StagewiseContext,
  StagewiseExport
} from '@/types/stagewise'

const defaultConfig: StagewiseConfig = {
  autoStart: false,
  captureConsole: true,
  captureNetwork: true,
  captureErrors: true,
  maxLogEntries: 1000,
  enableScreenshots: false,
  enablePerformanceMetrics: true
}

export function useStagewise(initialConfig?: Partial<StagewiseConfig>): StagewiseContext {
  const [session, setSession] = useState<StagewiseSession | null>(null)
  const [config] = useState<StagewiseConfig>({ ...defaultConfig, ...initialConfig })
  const [networkRequests, setNetworkRequests] = useState<NetworkRequest[]>([])
  const [consoleEntries, setConsoleEntries] = useState<ConsoleEntry[]>([])
  const [performanceMetrics, setPerformanceMetrics] = useState<PerformanceMetric[]>([])
  
  const originalConsole = useRef<Console>()
  const originalFetch = useRef<typeof fetch>()

  // Console capture
  useEffect(() => {
    if (!config.captureConsole) return

    originalConsole.current = window.console
    
    const captureConsoleMethod = (level: ConsoleEntry['level']) => {
      const original = originalConsole.current![level]
      return (...args: any[]) => {
        // Call original method
        original.apply(originalConsole.current, args)
        
        // Capture for stagewise
        const entry: ConsoleEntry = {
          id: uuidv4(),
          level,
          message: args.map(arg => 
            typeof arg === 'object' ? JSON.stringify(arg, null, 2) : String(arg)
          ).join(' '),
          args,
          timestamp: new Date(),
          stack: level === 'error' ? new Error().stack : undefined
        }
        
        setConsoleEntries(prev => {
          const newEntries = [...prev, entry]
          return newEntries.slice(-config.maxLogEntries!)
        })
      }
    }

    window.console.log = captureConsoleMethod('log')
    window.console.info = captureConsoleMethod('info')
    window.console.warn = captureConsoleMethod('warn')
    window.console.error = captureConsoleMethod('error')
    window.console.debug = captureConsoleMethod('debug')

    return () => {
      if (originalConsole.current) {
        window.console = originalConsole.current
      }
    }
  }, [config.captureConsole, config.maxLogEntries])

  // Network capture
  useEffect(() => {
    if (!config.captureNetwork) return

    originalFetch.current = window.fetch
    
    window.fetch = async (...args) => {
      const requestId = uuidv4()
      const startTime = new Date()
      const url = typeof args[0] === 'string' ? args[0] : (args[0] as Request).url
      const method = args[1]?.method || 'GET'
      
      const request: NetworkRequest = {
        id: requestId,
        url,
        method,
        startTime,
        requestHeaders: args[1]?.headers as Record<string, string>,
        requestBody: args[1]?.body
      }
      
      try {
        const response = await originalFetch.current!(...args)
        const endTime = new Date()
        
        const updatedRequest: NetworkRequest = {
          ...request,
          status: response.status,
          statusText: response.statusText,
          responseHeaders: Object.fromEntries(response.headers.entries()),
          endTime,
          duration: endTime.getTime() - startTime.getTime()
        }
        
        setNetworkRequests(prev => [...prev, updatedRequest])
        return response
      } catch (error) {
        const endTime = new Date()
        const updatedRequest: NetworkRequest = {
          ...request,
          endTime,
          duration: endTime.getTime() - startTime.getTime(),
          error: error instanceof Error ? error.message : String(error)
        }
        
        setNetworkRequests(prev => [...prev, updatedRequest])
        throw error
      }
    }

    return () => {
      if (originalFetch.current) {
        window.fetch = originalFetch.current
      }
    }
  }, [config.captureNetwork])

  const startSession = useCallback((name: string, description?: string) => {
    const newSession: StagewiseSession = {
      id: uuidv4(),
      name,
      description: description || '',
      status: 'running',
      stages: [],
      startTime: new Date(),
      metadata: {},
      tags: []
    }
    
    setSession(newSession)
    setNetworkRequests([])
    setConsoleEntries([])
    setPerformanceMetrics([])
    
    console.log(`🎬 Stagewise session started: ${name}`)
  }, [])

  const endSession = useCallback(() => {
    if (!session) return
    
    const endTime = new Date()
    const updatedSession: StagewiseSession = {
      ...session,
      status: 'success',
      endTime,
      duration: endTime.getTime() - session.startTime.getTime()
    }
    
    setSession(updatedSession)
    console.log(`🎬 Stagewise session ended: ${session.name}`)
  }, [session])

  const cancelSession = useCallback(() => {
    if (!session) return
    
    const endTime = new Date()
    const updatedSession: StagewiseSession = {
      ...session,
      status: 'cancelled',
      endTime,
      duration: endTime.getTime() - session.startTime.getTime()
    }
    
    setSession(updatedSession)
    console.log(`🎬 Stagewise session cancelled: ${session.name}`)
  }, [session])

  const startStage = useCallback((name: string, description?: string): string => {
    if (!session) throw new Error('No active session')
    
    const stageId = uuidv4()
    const newStage: Stage = {
      id: stageId,
      name,
      description: description || '',
      status: 'running',
      steps: [],
      startTime: new Date(),
      metadata: {}
    }
    
    setSession(prev => prev ? {
      ...prev,
      stages: [...prev.stages, newStage]
    } : null)
    
    console.log(`🎭 Stage started: ${name}`)
    return stageId
  }, [session])

  const endStage = useCallback((stageId: string, error?: string) => {
    if (!session) return
    
    const endTime = new Date()
    setSession(prev => {
      if (!prev) return null
      
      return {
        ...prev,
        stages: prev.stages.map(stage => 
          stage.id === stageId ? {
            ...stage,
            status: error ? 'error' : 'success',
            endTime,
            duration: endTime.getTime() - (stage.startTime?.getTime() || 0),
            error
          } : stage
        )
      }
    })
    
    console.log(`🎭 Stage ended: ${stageId}${error ? ` (error: ${error})` : ''}`)
  }, [session])

  const skipStage = useCallback((stageId: string, reason?: string) => {
    if (!session) return
    
    setSession(prev => {
      if (!prev) return null
      
      return {
        ...prev,
        stages: prev.stages.map(stage => 
          stage.id === stageId ? {
            ...stage,
            status: 'skipped',
            endTime: new Date(),
            error: reason
          } : stage
        )
      }
    })
    
    console.log(`🎭 Stage skipped: ${stageId}${reason ? ` (${reason})` : ''}`)
  }, [session])

  const startStep = useCallback((stageId: string, name: string, description?: string): string => {
    if (!session) throw new Error('No active session')
    
    const stepId = uuidv4()
    const newStep: StageStep = {
      id: stepId,
      name,
      description: description || '',
      status: 'running',
      startTime: new Date(),
      logs: [],
      metadata: {}
    }
    
    setSession(prev => {
      if (!prev) return null
      
      return {
        ...prev,
        stages: prev.stages.map(stage => 
          stage.id === stageId ? {
            ...stage,
            steps: [...stage.steps, newStep]
          } : stage
        )
      }
    })
    
    console.log(`🎯 Step started: ${name}`)
    return stepId
  }, [session])

  const endStep = useCallback((stageId: string, stepId: string, error?: string) => {
    if (!session) return
    
    const endTime = new Date()
    setSession(prev => {
      if (!prev) return null
      
      return {
        ...prev,
        stages: prev.stages.map(stage => 
          stage.id === stageId ? {
            ...stage,
            steps: stage.steps.map(step => 
              step.id === stepId ? {
                ...step,
                status: error ? 'error' : 'success',
                endTime,
                duration: endTime.getTime() - (step.startTime?.getTime() || 0),
                error
              } : step
            )
          } : stage
        )
      }
    })
    
    console.log(`🎯 Step ended: ${stepId}${error ? ` (error: ${error})` : ''}`)
  }, [session])

  const skipStep = useCallback((stageId: string, stepId: string, reason?: string) => {
    if (!session) return
    
    setSession(prev => {
      if (!prev) return null
      
      return {
        ...prev,
        stages: prev.stages.map(stage => 
          stage.id === stageId ? {
            ...stage,
            steps: stage.steps.map(step => 
              step.id === stepId ? {
                ...step,
                status: 'skipped',
                endTime: new Date(),
                error: reason
              } : step
            )
          } : stage
        )
      }
    })
    
    console.log(`🎯 Step skipped: ${stepId}${reason ? ` (${reason})` : ''}`)
  }, [session])

  const logData = useCallback((stageId: string, stepId: string, data: any) => {
    if (!session) return
    
    setSession(prev => {
      if (!prev) return null
      
      return {
        ...prev,
        stages: prev.stages.map(stage => 
          stage.id === stageId ? {
            ...stage,
            steps: stage.steps.map(step => 
              step.id === stepId ? {
                ...step,
                data: { ...step.data, ...data }
              } : step
            )
          } : stage
        )
      }
    })
  }, [session])

  const logMessage = useCallback((stageId: string, stepId: string, message: string) => {
    if (!session) return
    
    setSession(prev => {
      if (!prev) return null
      
      return {
        ...prev,
        stages: prev.stages.map(stage => 
          stage.id === stageId ? {
            ...stage,
            steps: stage.steps.map(step => 
              step.id === stepId ? {
                ...step,
                logs: [...(step.logs || []), `${new Date().toISOString()}: ${message}`]
              } : step
            )
          } : stage
        )
      }
    })
  }, [session])

  const captureScreenshot = useCallback(async (stageId: string, stepId: string, name?: string) => {
    if (!config.enableScreenshots) return
    
    // Note: This would require additional setup for actual screenshot capture
    // For now, we'll just log the intent
    console.log(`📸 Screenshot captured: ${name || 'unnamed'} for step ${stepId}`)
  }, [config.enableScreenshots])

  const exportSession = useCallback((): string => {
    if (!session) throw new Error('No active session')
    
    const exportData: StagewiseExport = {
      version: '1.0.0',
      exportTime: new Date(),
      session,
      networkRequests,
      consoleEntries,
      performanceMetrics,
      metadata: {
        userAgent: navigator.userAgent,
        url: window.location.href,
        timestamp: new Date().toISOString()
      }
    }
    
    return JSON.stringify(exportData, null, 2)
  }, [session, networkRequests, consoleEntries, performanceMetrics])

  const importSession = useCallback((data: string) => {
    try {
      const importData: StagewiseExport = JSON.parse(data)
      setSession(importData.session)
      setNetworkRequests(importData.networkRequests)
      setConsoleEntries(importData.consoleEntries)
      setPerformanceMetrics(importData.performanceMetrics)
      console.log('📥 Session imported successfully')
    } catch (error) {
      console.error('Failed to import session:', error)
      throw new Error('Invalid session data')
    }
  }, [])

  const clearHistory = useCallback(() => {
    setNetworkRequests([])
    setConsoleEntries([])
    setPerformanceMetrics([])
    console.log('🧹 History cleared')
  }, [])

  return {
    session,
    config,
    networkRequests,
    consoleEntries,
    performanceMetrics,
    startSession,
    endSession,
    cancelSession,
    startStage,
    endStage,
    skipStage,
    startStep,
    endStep,
    skipStep,
    logData,
    logMessage,
    captureScreenshot,
    exportSession,
    importSession,
    clearHistory
  }
}
