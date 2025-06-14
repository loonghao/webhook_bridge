import type { Metadata } from 'next'
import { Inter } from 'next/font/google'
import './globals.css'
import { StagewiseProvider } from '@/components/StagewiseProvider'

const inter = Inter({ subsets: ['latin'] })

export const metadata: Metadata = {
  title: 'Webhook Bridge - Dashboard',
  description: 'Modern webhook processing dashboard built with Next.js',
}

export default function RootLayout({
  children,
}: {
  children: React.ReactNode
}) {
  return (
    <html lang="en" className="dark">
      <body className={`${inter.className} bg-background text-foreground antialiased`}>
        <StagewiseProvider>
          <div id="root">
            {children}
          </div>
        </StagewiseProvider>
      </body>
    </html>
  )
}
