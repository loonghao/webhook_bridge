/**
 * Connection Optimizer based on Stagewise Debug Data
 * Analyzes network requests and optimizes frontend-backend connections
 */

import { useStagewise } from '@/hooks/useStagewise'
import { NetworkRequest } from '@/types/stagewise'

export interface ConnectionMetrics {
  totalRequests: number
  failedRequests: number
  averageLatency: number
  slowRequests: number
  errorRate: number
  commonErrors: Record<string, number>
  endpointPerformance: Record<string, {
    count: number
    averageLatency: number
    errorRate: number
    lastError?: string
  }>
}

export interface OptimizationSuggestion {
  type: 'error' | 'performance' | 'reliability' | 'configuration'
  priority: 'high' | 'medium' | 'low'
  title: string
  description: string
  action: string
  code?: string
}

export class ConnectionOptimizer {
  private stagewise: any
  private slowRequestThreshold = 2000 // 2 seconds
  private highErrorRateThreshold = 0.1 // 10%

  constructor(stagewise: any) {
    this.stagewise = stagewise
  }

  analyzeNetworkRequests(): ConnectionMetrics {
    const requests = this.stagewise.networkRequests as NetworkRequest[]
    
    if (!requests || requests.length === 0) {
      return this.getEmptyMetrics()
    }

    const totalRequests = requests.length
    const failedRequests = requests.filter(r => r.error || (r.status && r.status >= 400)).length
    const completedRequests = requests.filter(r => r.duration !== undefined)
    
    const averageLatency = completedRequests.length > 0
      ? completedRequests.reduce((sum, r) => sum + (r.duration || 0), 0) / completedRequests.length
      : 0

    const slowRequests = completedRequests.filter(r => (r.duration || 0) > this.slowRequestThreshold).length
    const errorRate = totalRequests > 0 ? failedRequests / totalRequests : 0

    // Analyze common errors
    const commonErrors: Record<string, number> = {}
    requests.forEach(r => {
      if (r.error) {
        commonErrors[r.error] = (commonErrors[r.error] || 0) + 1
      } else if (r.status && r.status >= 400) {
        const errorKey = `HTTP ${r.status}: ${r.statusText || 'Unknown'}`
        commonErrors[errorKey] = (commonErrors[errorKey] || 0) + 1
      }
    })

    // Analyze endpoint performance
    const endpointPerformance: Record<string, any> = {}
    requests.forEach(r => {
      const endpoint = this.extractEndpoint(r.url)
      if (!endpointPerformance[endpoint]) {
        endpointPerformance[endpoint] = {
          count: 0,
          totalLatency: 0,
          errors: 0,
          lastError: undefined
        }
      }

      const ep = endpointPerformance[endpoint]
      ep.count++
      
      if (r.duration) {
        ep.totalLatency += r.duration
      }
      
      if (r.error || (r.status && r.status >= 400)) {
        ep.errors++
        ep.lastError = r.error || `HTTP ${r.status}`
      }
    })

    // Calculate averages and error rates for endpoints
    Object.keys(endpointPerformance).forEach(endpoint => {
      const ep = endpointPerformance[endpoint]
      ep.averageLatency = ep.count > 0 ? ep.totalLatency / ep.count : 0
      ep.errorRate = ep.count > 0 ? ep.errors / ep.count : 0
      delete ep.totalLatency
      delete ep.errors
    })

    return {
      totalRequests,
      failedRequests,
      averageLatency,
      slowRequests,
      errorRate,
      commonErrors,
      endpointPerformance
    }
  }

  generateOptimizationSuggestions(metrics: ConnectionMetrics): OptimizationSuggestion[] {
    const suggestions: OptimizationSuggestion[] = []

    // High error rate
    if (metrics.errorRate > this.highErrorRateThreshold) {
      suggestions.push({
        type: 'reliability',
        priority: 'high',
        title: 'High Error Rate Detected',
        description: `${(metrics.errorRate * 100).toFixed(1)}% of requests are failing`,
        action: 'Implement retry logic and improve error handling',
        code: `
// Add retry logic to API client
import { withRetry } from '@/services/api'

const result = await withRetry(
  () => apiClient.get('/endpoint'),
  3, // max retries
  1000 // delay
)`
      })
    }

    // Slow requests
    if (metrics.slowRequests > 0) {
      const slowPercentage = (metrics.slowRequests / metrics.totalRequests * 100).toFixed(1)
      suggestions.push({
        type: 'performance',
        priority: metrics.slowRequests > metrics.totalRequests * 0.2 ? 'high' : 'medium',
        title: 'Slow Requests Detected',
        description: `${slowPercentage}% of requests are taking longer than ${this.slowRequestThreshold}ms`,
        action: 'Implement request caching and optimize backend endpoints',
        code: `
// Add request caching
const cache = new Map()
const getCachedData = async (key: string, fetcher: () => Promise<any>) => {
  if (cache.has(key)) return cache.get(key)
  const data = await fetcher()
  cache.set(key, data)
  return data
}`
      })
    }

    // Connection errors
    const connectionErrors = Object.keys(metrics.commonErrors).filter(error => 
      error.includes('fetch') || error.includes('network') || error.includes('ECONNREFUSED')
    )
    
    if (connectionErrors.length > 0) {
      suggestions.push({
        type: 'configuration',
        priority: 'high',
        title: 'Connection Issues Detected',
        description: 'Network connectivity problems between frontend and backend',
        action: 'Check backend server status and API base URL configuration',
        code: `
// Check current API configuration
console.log('API Base URL:', process.env.NEXT_PUBLIC_API_BASE_URL)
console.log('Backend Status:', await apiClient.getStatus())

// Update .env.local if needed:
// NEXT_PUBLIC_API_BASE_URL=http://localhost:8080`
      })
    }

    // Specific endpoint issues
    Object.entries(metrics.endpointPerformance).forEach(([endpoint, perf]) => {
      if (perf.errorRate > 0.2) { // 20% error rate for specific endpoint
        suggestions.push({
          type: 'error',
          priority: 'medium',
          title: `Endpoint Issues: ${endpoint}`,
          description: `${(perf.errorRate * 100).toFixed(1)}% error rate for ${endpoint}`,
          action: `Review ${endpoint} implementation and add specific error handling`,
          code: `
// Add specific error handling for ${endpoint}
try {
  const result = await apiClient.get('${endpoint}')
  return result
} catch (error) {
  console.error('${endpoint} failed:', error)
  // Implement fallback or user notification
  throw new Error('Service temporarily unavailable')
}`
        })
      }
    })

    // WebSocket connection issues
    const wsErrors = Object.keys(metrics.commonErrors).filter(error => 
      error.includes('WebSocket') || error.includes('ws://')
    )
    
    if (wsErrors.length > 0) {
      suggestions.push({
        type: 'configuration',
        priority: 'medium',
        title: 'WebSocket Connection Issues',
        description: 'Real-time features may not work properly',
        action: 'Configure WebSocket URL and implement reconnection logic',
        code: `
// Update WebSocket configuration
// In .env.local:
// NEXT_PUBLIC_WS_BASE_URL=ws://localhost:8080

// Implement auto-reconnection
const connectWithRetry = async (maxAttempts = 5) => {
  for (let i = 0; i < maxAttempts; i++) {
    try {
      await monitorWebSocket.connect()
      break
    } catch (error) {
      if (i === maxAttempts - 1) throw error
      await new Promise(resolve => setTimeout(resolve, 1000 * Math.pow(2, i)))
    }
  }
}`
      })
    }

    return suggestions.sort((a, b) => {
      const priorityOrder = { high: 3, medium: 2, low: 1 }
      return priorityOrder[b.priority] - priorityOrder[a.priority]
    })
  }

  private extractEndpoint(url: string): string {
    try {
      const urlObj = new URL(url)
      return urlObj.pathname.replace('/api/dashboard', '')
    } catch {
      return url
    }
  }

  private getEmptyMetrics(): ConnectionMetrics {
    return {
      totalRequests: 0,
      failedRequests: 0,
      averageLatency: 0,
      slowRequests: 0,
      errorRate: 0,
      commonErrors: {},
      endpointPerformance: {}
    }
  }
}

// Hook for using connection optimizer
export function useConnectionOptimizer() {
  const stagewise = useStagewise()
  
  const optimizer = new ConnectionOptimizer(stagewise)
  
  const analyzeConnections = () => {
    const metrics = optimizer.analyzeNetworkRequests()
    const suggestions = optimizer.generateOptimizationSuggestions(metrics)
    
    return { metrics, suggestions }
  }
  
  const logOptimizationReport = () => {
    const { metrics, suggestions } = analyzeConnections()
    
    console.group('🔍 Connection Optimization Report')
    console.log('📊 Metrics:', metrics)
    console.log('💡 Suggestions:', suggestions)
    console.groupEnd()
    
    return { metrics, suggestions }
  }
  
  return {
    analyzeConnections,
    logOptimizationReport,
    optimizer
  }
}
