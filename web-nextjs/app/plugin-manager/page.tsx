'use client'

import { Cog, Upload, Download, Trash2 } from 'lucide-react'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Layout } from '@/components/Layout'

export default function PluginManager() {
  return (
    <Layout>
      <div className="space-y-6">
        <div className="flex items-center justify-between">
          <div>
            <h1 className="text-3xl font-bold tracking-tight">Plugin Manager</h1>
            <p className="text-muted-foreground">
              Install, update, and manage webhook plugins
            </p>
          </div>
          <div className="flex items-center space-x-2">
            <Button variant="outline">
              <Download className="h-4 w-4 mr-2" />
              Install from URL
            </Button>
            <Button>
              <Upload className="h-4 w-4 mr-2" />
              Upload Plugin
            </Button>
          </div>
        </div>

        <Card>
          <CardHeader>
            <CardTitle className="flex items-center space-x-2">
              <Cog className="h-5 w-5" />
              <span>Plugin Management</span>
            </CardTitle>
            <CardDescription>
              Install, configure, and manage webhook processing plugins
            </CardDescription>
          </CardHeader>
          <CardContent>
            <div className="text-center py-12 text-muted-foreground">
              <Cog className="h-12 w-12 mx-auto mb-4 opacity-50" />
              <p className="text-lg font-medium">Plugin Manager</p>
              <p>This page will contain plugin installation and management tools.</p>
            </div>
          </CardContent>
        </Card>
      </div>
    </Layout>
  )
}
