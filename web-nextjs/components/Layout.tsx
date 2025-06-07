'use client'

import { ReactNode } from 'react'
import { Sidebar } from './Sidebar'
import { Header } from './Header'
import { Breadcrumb } from './Breadcrumb'

interface LayoutProps {
  children: ReactNode
}

export function Layout({ children }: LayoutProps) {
  return (
    <div className="min-h-screen bg-gradient-to-br from-slate-50 via-blue-50 to-indigo-100 dark:from-slate-900 dark:via-slate-900 dark:to-slate-800">
      {/* Background Pattern */}
      <div className="absolute inset-0 bg-grid-slate-100 dark:bg-grid-slate-700/25 bg-[size:20px_20px] opacity-60"></div>
      <div className="absolute inset-0 bg-gradient-to-r from-blue-500/5 to-purple-500/5"></div>

      <div className="relative flex h-screen">
        <Sidebar />
        <div className="flex-1 flex flex-col overflow-hidden">
          <Header />
          <main className="flex-1 overflow-y-auto">
            <div className="p-8 space-y-8">
              <Breadcrumb />
              {children}
            </div>
          </main>
        </div>
      </div>
    </div>
  )
}
