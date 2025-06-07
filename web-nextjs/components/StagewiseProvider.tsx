'use client'

import React, { createContext, useContext, ReactNode } from 'react'
import { useStagewise } from '@/hooks/useStagewise'
import type { StagewiseContext } from '@/types/stagewise'

const StagewiseContext = createContext<StagewiseContext | null>(null)

interface StagewiseProviderProps {
  children: ReactNode
  config?: {
    autoStart?: boolean
    captureConsole?: boolean
    captureNetwork?: boolean
    captureErrors?: boolean
    maxLogEntries?: number
    enableScreenshots?: boolean
    enablePerformanceMetrics?: boolean
  }
}

export function StagewiseProvider({ children, config }: StagewiseProviderProps) {
  const stagewise = useStagewise(config)

  return (
    <StagewiseContext.Provider value={stagewise}>
      {children}
    </StagewiseContext.Provider>
  )
}

export function useStagewiseContext(): StagewiseContext {
  const context = useContext(StagewiseContext)
  if (!context) {
    throw new Error('useStagewiseContext must be used within a StagewiseProvider')
  }
  return context
}

// Optional: Global stagewise instance for use outside of React components
let globalStagewise: StagewiseContext | null = null

export function setGlobalStagewise(stagewise: StagewiseContext) {
  globalStagewise = stagewise
}

export function getGlobalStagewise(): StagewiseContext | null {
  return globalStagewise
}

// Utility functions for common debugging patterns
export const stagewise = {
  // Quick session management
  quickStart: (name?: string) => {
    const sw = getGlobalStagewise()
    if (sw) {
      sw.startSession(name || `Quick Debug ${new Date().toLocaleTimeString()}`)
      return sw.session?.id
    }
    return null
  },

  quickEnd: () => {
    const sw = getGlobalStagewise()
    if (sw) {
      sw.endSession()
    }
  },

  // Quick stage management
  stage: (name: string, description?: string) => {
    const sw = getGlobalStagewise()
    if (sw && sw.session) {
      return sw.startStage(name, description)
    }
    return null
  },

  endStage: (stageId: string, error?: string) => {
    const sw = getGlobalStagewise()
    if (sw) {
      sw.endStage(stageId, error)
    }
  },

  // Quick step management
  step: (stageId: string, name: string, description?: string) => {
    const sw = getGlobalStagewise()
    if (sw && sw.session) {
      return sw.startStep(stageId, name, description)
    }
    return null
  },

  endStep: (stageId: string, stepId: string, error?: string) => {
    const sw = getGlobalStagewise()
    if (sw) {
      sw.endStep(stageId, stepId, error)
    }
  },

  // Quick logging
  log: (stageId: string, stepId: string, message: string) => {
    const sw = getGlobalStagewise()
    if (sw) {
      sw.logMessage(stageId, stepId, message)
    }
  },

  data: (stageId: string, stepId: string, data: any) => {
    const sw = getGlobalStagewise()
    if (sw) {
      sw.logData(stageId, stepId, data)
    }
  },

  // Utility for wrapping async functions with automatic stage/step tracking
  wrap: <T extends (...args: any[]) => Promise<any>>(
    fn: T,
    stageName: string,
    stepName?: string
  ): T => {
    return (async (...args: any[]) => {
      const sw = getGlobalStagewise()
      if (!sw || !sw.session) {
        return fn(...args)
      }

      const stageId = sw.startStage(stageName, `Auto-wrapped: ${fn.name || 'anonymous'}`)
      const stepId = sw.startStep(stageId, stepName || fn.name || 'execute', 'Auto-wrapped function execution')

      try {
        sw.logData(stageId, stepId, { args, functionName: fn.name })
        const result = await fn(...args)
        sw.logData(stageId, stepId, { result })
        sw.endStep(stageId, stepId)
        sw.endStage(stageId)
        return result
      } catch (error) {
        const errorMessage = error instanceof Error ? error.message : String(error)
        sw.endStep(stageId, stepId, errorMessage)
        sw.endStage(stageId, errorMessage)
        throw error
      }
    }) as T
  },

  // Utility for timing operations
  time: async function <T>(
    operation: () => Promise<T>,
    stageId: string,
    stepName: string
  ): Promise<T> {
    const sw = getGlobalStagewise()
    if (!sw || !sw.session) {
      return operation()
    }

    const stepId = sw.startStep(stageId, stepName, 'Timed operation')
    const startTime = performance.now()

    try {
      const result = await operation()
      const endTime = performance.now()
      sw.logData(stageId, stepId, {
        duration: endTime - startTime,
        startTime,
        endTime
      })
      sw.endStep(stageId, stepId)
      return result
    } catch (error) {
      const endTime = performance.now()
      sw.logData(stageId, stepId, {
        duration: endTime - startTime,
        startTime,
        endTime,
        error: error instanceof Error ? error.message : String(error)
      })
      sw.endStep(stageId, stepId, error instanceof Error ? error.message : String(error))
      throw error
    }
  }
}
