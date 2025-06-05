import {
  BarChart3,
  Home,
  ScrollText,
  Settings,
  Zap,
  Code,
  Wifi,
  Puzzle,
  TestTube,
  Cog
} from 'lucide-react'
import { NavLink } from 'react-router-dom'
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
    ]
  }
]

export function Sidebar() {
  return (
    <div className="w-64 bg-card border-r">
      <div className="p-6">
        <div className="flex items-center space-x-2">
          <div className="flex h-8 w-8 items-center justify-center rounded-lg bg-primary text-primary-foreground">
            <Zap className="h-4 w-4" />
          </div>
          <div>
            <h2 className="text-lg font-semibold">Webhook Bridge</h2>
            <p className="text-xs text-muted-foreground">Admin Panel</p>
          </div>
        </div>
      </div>
      
      <nav className="px-4 pb-4">
        <div className="space-y-6">
          {navigationGroups.map((group) => (
            <div key={group.title}>
              <h3 className="mb-2 px-3 text-xs font-semibold text-muted-foreground uppercase tracking-wider">
                {group.title}
              </h3>
              <ul className="space-y-1">
                {group.items.map((item) => (
                  <li key={item.name}>
                    <NavLink
                      to={item.href}
                      className={({ isActive }) =>
                        cn(
                          'flex items-center space-x-3 rounded-lg px-3 py-2 text-sm font-medium transition-colors',
                          isActive
                            ? 'bg-accent text-accent-foreground'
                            : 'text-muted-foreground hover:bg-accent hover:text-accent-foreground'
                        )
                      }
                    >
                      <item.icon className="h-4 w-4" />
                      <span>{item.name}</span>
                    </NavLink>
                  </li>
                ))}
              </ul>
            </div>
          ))}
        </div>
      </nav>
    </div>
  )
}
