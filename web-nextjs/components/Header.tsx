'use client'

import { Bell, Search, Settings, User, Moon, Sun } from 'lucide-react'
import { Button } from './ui/button'
import { Badge } from './ui/badge'

export function Header() {
  return (
    <header className="modern-header px-6 py-4 sticky top-0 z-50">
      <div className="flex items-center justify-between">
        <div className="flex items-center space-x-4">
          <h1 className="text-2xl font-bold text-blue-600">Dashboard</h1>
          <Badge variant="secondary" className="bg-blue-100 text-blue-700 dark:bg-blue-900 dark:bg-opacity-30 dark:text-blue-300">
            v2.0.0-hybrid
          </Badge>
        </div>

        <div className="flex items-center space-x-4">
          <div className="relative">
            <Search className="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-slate-400" />
            <input
              type="search"
              placeholder="Search logs, plugins..."
              className="h-10 w-80 rounded-xl border-0 bg-slate-100 dark:bg-slate-800 pl-10 pr-4 text-sm placeholder:text-slate-400 focus:outline-none focus:ring-2 focus:ring-blue-500 transition-all duration-200"
            />
          </div>

          <div className="flex items-center space-x-2">
            <Button variant="ghost" size="icon" className="h-10 w-10 rounded-xl hover:bg-slate-100 dark:hover:bg-slate-800">
              <Bell className="h-5 w-5" />
            </Button>

            <Button variant="ghost" size="icon" className="h-10 w-10 rounded-xl hover:bg-slate-100 dark:hover:bg-slate-800">
              <Settings className="h-5 w-5" />
            </Button>

            <Button variant="ghost" size="icon" className="h-10 w-10 rounded-xl hover:bg-slate-100 dark:hover:bg-slate-800">
              <Sun className="h-4 w-4 rotate-0 scale-100 transition-all dark:-rotate-90 dark:scale-0" />
              <Moon className="absolute h-4 w-4 rotate-90 scale-0 transition-all dark:rotate-0 dark:scale-100" />
            </Button>

            <div className="h-8 w-8 rounded-full modern-button flex items-center justify-center">
              <User className="h-4 w-4 text-white" />
            </div>
          </div>
        </div>
      </div>
    </header>
  )
}
