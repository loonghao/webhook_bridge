'use client'

import Link from 'next/link'
import { usePathname } from 'next/navigation'
import {
  Home,
  BarChart3,
  Puzzle,
  Cog,
  Code,
  Wifi,
  ScrollText,
  Settings,
  TestTube,
  Zap,
  ChevronRight
} from 'lucide-react'
import { cn } from '@/lib/utils'

const navigationGroups = [
  {
    title: 'Overview',
    items: [
      { name: 'Dashboard', href: '/', icon: Home },
      { name: 'System Status', href: '/status', icon: BarChart3 },
    ]
  },
  {
    title: 'Plugin Management',
    items: [
      { name: 'Plugins', href: '/plugins', icon: Puzzle },
      { name: 'Plugin Manager', href: '/plugin-manager', icon: Cog },
    ]
  },
  {
    title: 'System',
    items: [
      { name: 'Python Interpreters', href: '/interpreters', icon: Code },
      { name: 'Connection Status', href: '/connection', icon: Wifi },
      { name: 'Logs', href: '/logs', icon: ScrollText },
    ]
  },
  {
    title: 'Tools',
    items: [
      { name: 'Configuration', href: '/config', icon: Settings },
      { name: 'API Test', href: '/api-test', icon: TestTube },
      { name: 'Debug', href: '/debug', icon: Code },
    ]
  }
]

export function Sidebar() {
  const pathname = usePathname()

  return (
    <div className="w-64 h-full flex flex-col modern-sidebar">
      {/* Logo Section */}
      <div className="p-6 modern-sidebar-header">
        <div className="flex items-center space-x-3">
          <div className="flex h-12 w-12 items-center justify-center rounded-xl bg-white bg-opacity-20">
            <Zap className="h-6 w-6 text-white" />
          </div>
          <div>
            <h2 className="text-xl font-bold text-white">Webhook Bridge</h2>
            <p className="text-sm text-white text-opacity-70">Admin Panel</p>
          </div>
        </div>
      </div>

      {/* Navigation */}
      <nav className="flex-1 px-4 py-6 overflow-y-auto">
        <div className="space-y-8">
          {navigationGroups.map((group) => (
            <div key={group.title}>
              <h3 className="mb-3 px-3 text-xs font-semibold text-slate-400 uppercase tracking-wider">
                {group.title}
              </h3>
              <ul className="space-y-1">
                {group.items.map((item) => {
                  const isActive = pathname === item.href ||
                    (item.href !== '/' && pathname.startsWith(item.href))

                  return (
                    <li key={item.name}>
                      <Link
                        href={item.href}
                        className={cn(
                          'modern-sidebar-nav-item',
                          isActive && 'active'
                        )}
                      >
                        <item.icon className="h-5 w-5 transition-colors" />
                        <span className="font-medium">{item.name}</span>
                        <ChevronRight className={cn(
                          "h-4 w-4 ml-auto transition-all duration-200",
                          isActive ? "rotate-90 text-blue-400" : "text-slate-500 group-hover:text-slate-300"
                        )} />
                      </Link>
                    </li>
                  )
                })}
              </ul>
            </div>
          ))}
        </div>
      </nav>

      {/* Footer */}
      <div className="p-4 border-t border-slate-700">
        <div className="modern-glass rounded-lg p-3">
          <div className="flex items-center space-x-3">
            <div className="h-3 w-3 rounded-full bg-green-400 animate-pulse"></div>
            <span className="text-sm text-slate-300 font-medium">System Online</span>
          </div>
          <div className="mt-2 text-xs text-slate-400">
            All services operational
          </div>
        </div>
      </div>
    </div>
  )
}
