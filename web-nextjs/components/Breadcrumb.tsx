'use client'

import Link from 'next/link'
import { usePathname } from 'next/navigation'
import { ChevronRight, Home, LucideIcon } from 'lucide-react'
import { cn } from '@/lib/utils'

interface BreadcrumbItem {
  label: string
  href?: string
  icon?: LucideIcon
}

const routeMap: Record<string, BreadcrumbItem[]> = {
  '/': [
    { label: 'Dashboard', icon: Home }
  ],
  '/dashboard': [
    { label: 'Dashboard', icon: Home }
  ],
  '/status': [
    { label: 'Dashboard', href: '/', icon: Home },
    { label: 'System Status' }
  ],
  '/plugins': [
    { label: 'Dashboard', href: '/', icon: Home },
    { label: 'Plugin Management', href: '/plugins' },
    { label: 'Plugins' }
  ],
  '/plugin-manager': [
    { label: 'Dashboard', href: '/', icon: Home },
    { label: 'Plugin Management', href: '/plugins' },
    { label: 'Plugin Manager' }
  ],
  '/interpreters': [
    { label: 'Dashboard', href: '/', icon: Home },
    { label: 'System', href: '/status' },
    { label: 'Python Interpreters' }
  ],
  '/connection': [
    { label: 'Dashboard', href: '/', icon: Home },
    { label: 'System', href: '/status' },
    { label: 'Connection Status' }
  ],
  '/logs': [
    { label: 'Dashboard', href: '/', icon: Home },
    { label: 'System', href: '/status' },
    { label: 'Logs' }
  ],
  '/config': [
    { label: 'Dashboard', href: '/', icon: Home },
    { label: 'Tools' },
    { label: 'Configuration' }
  ],
  '/api-test': [
    { label: 'Dashboard', href: '/', icon: Home },
    { label: 'Tools' },
    { label: 'API Test' }
  ],
  '/debug': [
    { label: 'Dashboard', href: '/', icon: Home },
    { label: 'Tools' },
    { label: 'Debug' }
  ]
}

export function Breadcrumb() {
  const pathname = usePathname()
  const breadcrumbs = routeMap[pathname] || [
    { label: 'Dashboard', href: '/', icon: Home },
    { label: 'Unknown Page' }
  ]

  if (breadcrumbs.length <= 1) {
    return null
  }

  return (
    <nav className="flex items-center space-x-1 text-sm text-muted-foreground mb-6">
      {breadcrumbs.map((item, index) => {
        const isLast = index === breadcrumbs.length - 1
        const Icon = item.icon

        return (
          <div key={index} className="flex items-center">
            {index > 0 && (
              <ChevronRight className="h-4 w-4 mx-2 text-muted-foreground/50" />
            )}
            
            {item.href && !isLast ? (
              <Link
                href={item.href}
                className={cn(
                  'flex items-center space-x-1 hover:text-foreground transition-colors',
                  Icon && 'space-x-1'
                )}
              >
                {Icon && <Icon className="h-4 w-4" />}
                <span>{item.label}</span>
              </Link>
            ) : (
              <div className={cn(
                'flex items-center space-x-1',
                isLast ? 'text-foreground font-medium' : 'text-muted-foreground',
                Icon && 'space-x-1'
              )}>
                {Icon && <Icon className="h-4 w-4" />}
                <span>{item.label}</span>
              </div>
            )}
          </div>
        )
      })}
    </nav>
  )
}
