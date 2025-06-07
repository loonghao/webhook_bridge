'use client'

import { Wifi, RefreshCw, AlertTriangle } from 'lucide-react'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Layout } from '@/components/Layout'

export default function ConnectionStatus() {
  return (
    <Layout>
      <div className="space-y-6">
        <div className="flex items-center justify-between">
          <div>
            <h1 className="text-3xl font-bold tracking-tight">Connection Status</h1>
            <p className="text-muted-foreground">
              Monitor external connections and service health
            </p>
          </div>
          <Button>
            <RefreshCw className="h-4 w-4 mr-2" />
            Refresh All
          </Button>
        </div>

        <Card>
          <CardHeader>
            <CardTitle className="flex items-center space-x-2">
              <Wifi className="h-5 w-5" />
              <span>Connection Monitoring</span>
            </CardTitle>
            <CardDescription>
              Monitor connections to external services and APIs
            </CardDescription>
          </CardHeader>
          <CardContent>
            <div className="text-center py-12 text-muted-foreground">
              <Wifi className="h-12 w-12 mx-auto mb-4 opacity-50" />
              <p className="text-lg font-medium">Connection Status</p>
              <p>This page will show external connection monitoring.</p>
            </div>
          </CardContent>
        </Card>
      </div>
    </Layout>
  )
}
