'use client'

import { TestTube, Send, Copy } from 'lucide-react'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Layout } from '@/components/Layout'

export default function ApiTest() {
  return (
    <Layout>
      <div className="space-y-6">
        <div className="flex items-center justify-between">
          <div>
            <h1 className="text-3xl font-bold tracking-tight">API Testing</h1>
            <p className="text-muted-foreground">
              Test webhook endpoints and API functionality
            </p>
          </div>
          <div className="flex items-center space-x-2">
            <Button variant="outline">
              <Copy className="h-4 w-4 mr-2" />
              Copy cURL
            </Button>
            <Button>
              <Send className="h-4 w-4 mr-2" />
              Send Request
            </Button>
          </div>
        </div>

        <Card>
          <CardHeader>
            <CardTitle className="flex items-center space-x-2">
              <TestTube className="h-5 w-5" />
              <span>API Testing Tool</span>
            </CardTitle>
            <CardDescription>
              Test webhook endpoints and debug API responses
            </CardDescription>
          </CardHeader>
          <CardContent>
            <div className="text-center py-12 text-muted-foreground">
              <TestTube className="h-12 w-12 mx-auto mb-4 opacity-50" />
              <p className="text-lg font-medium">API Testing</p>
              <p>This page will contain API testing tools and utilities.</p>
            </div>
          </CardContent>
        </Card>
      </div>
    </Layout>
  )
}
