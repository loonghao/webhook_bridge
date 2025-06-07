/**
 * Stagewise debugging types for AI-assisted web debugging
 */

export interface StageStep {
  id: string
  name: string
  description: string
  status: 'pending' | 'running' | 'success' | 'error' | 'skipped'
  startTime?: Date
  endTime?: Date
  duration?: number
  error?: string
  data?: any
  logs?: string[]
  metadata?: Record<string, any>
}

export interface Stage {
  id: string
  name: string
  description: string
  status: 'pending' | 'running' | 'success' | 'error' | 'skipped'
  steps: StageStep[]
  startTime?: Date
  endTime?: Date
  duration?: number
  error?: string
  metadata?: Record<string, any>
}

export interface StagewiseSession {
  id: string
  name: string
  description: string
  status: 'pending' | 'running' | 'success' | 'error' | 'cancelled'
  stages: Stage[]
  startTime: Date
  endTime?: Date
  duration?: number
  metadata?: Record<string, any>
  tags?: string[]
}

export interface StagewiseConfig {
  autoStart?: boolean
  captureConsole?: boolean
  captureNetwork?: boolean
  captureErrors?: boolean
  maxLogEntries?: number
  enableScreenshots?: boolean
  enablePerformanceMetrics?: boolean
}

export interface NetworkRequest {
  id: string
  url: string
  method: string
  status?: number
  statusText?: string
  requestHeaders?: Record<string, string>
  responseHeaders?: Record<string, string>
  requestBody?: any
  responseBody?: any
  startTime: Date
  endTime?: Date
  duration?: number
  error?: string
}

export interface ConsoleEntry {
  id: string
  level: 'log' | 'info' | 'warn' | 'error' | 'debug'
  message: string
  args?: any[]
  timestamp: Date
  stack?: string
}

export interface PerformanceMetric {
  id: string
  name: string
  value: number
  unit: string
  timestamp: Date
  metadata?: Record<string, any>
}

export interface StagewiseContext {
  session: StagewiseSession | null
  config: StagewiseConfig
  networkRequests: NetworkRequest[]
  consoleEntries: ConsoleEntry[]
  performanceMetrics: PerformanceMetric[]
  
  // Session management
  startSession: (name: string, description?: string) => void
  endSession: () => void
  cancelSession: () => void
  
  // Stage management
  startStage: (name: string, description?: string) => string
  endStage: (stageId: string, error?: string) => void
  skipStage: (stageId: string, reason?: string) => void
  
  // Step management
  startStep: (stageId: string, name: string, description?: string) => string
  endStep: (stageId: string, stepId: string, error?: string) => void
  skipStep: (stageId: string, stepId: string, reason?: string) => void
  
  // Data capture
  logData: (stageId: string, stepId: string, data: any) => void
  logMessage: (stageId: string, stepId: string, message: string) => void
  captureScreenshot: (stageId: string, stepId: string, name?: string) => Promise<void>
  
  // Utilities
  exportSession: () => string
  importSession: (data: string) => void
  clearHistory: () => void
}

export interface StagewiseExport {
  version: string
  exportTime: Date
  session: StagewiseSession
  networkRequests: NetworkRequest[]
  consoleEntries: ConsoleEntry[]
  performanceMetrics: PerformanceMetric[]
  metadata?: Record<string, any>
}
