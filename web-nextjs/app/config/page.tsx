'use client'

import { Settings, Save, RotateCcw } from 'lucide-react'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Layout } from '@/components/Layout'

export default function Configuration() {
  return (
    <Layout>
      <div className="space-y-6">
        <div className="flex items-center justify-between">
          <div>
            <h1 className="text-3xl font-bold tracking-tight">Configuration</h1>
            <p className="text-muted-foreground">
              Manage system settings and configuration
            </p>
          </div>
          <div className="flex items-center space-x-2">
            <Button variant="outline">
              <RotateCcw className="h-4 w-4 mr-2" />
              Reset to Defaults
            </Button>
            <Button>
              <Save className="h-4 w-4 mr-2" />
              Save Changes
            </Button>
          </div>
        </div>

        <Card>
          <CardHeader>
            <CardTitle className="flex items-center space-x-2">
              <Settings className="h-5 w-5" />
              <span>System Configuration</span>
            </CardTitle>
            <CardDescription>
              Configure webhook bridge system settings
            </CardDescription>
          </CardHeader>
          <CardContent>
            <div className="text-center py-12 text-muted-foreground">
              <Settings className="h-12 w-12 mx-auto mb-4 opacity-50" />
              <p className="text-lg font-medium">Configuration Panel</p>
              <p>This page will contain system configuration options.</p>
            </div>
          </CardContent>
        </Card>
      </div>
    </Layout>
  )
}
