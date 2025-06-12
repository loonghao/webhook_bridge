# Webhook Bridge Next.js Dashboard

A modern, responsive dashboard built with Next.js, TypeScript, and shadcn/ui for the Webhook Bridge service.

## 🚀 Features

- **Modern UI**: Built with shadcn/ui components and Tailwind CSS
- **Real-time Data**: Connects to Go backend APIs for live dashboard updates
- **Responsive Design**: Works seamlessly on desktop and mobile devices
- **Dark Theme**: Professional dark mode interface
- **Type Safety**: Full TypeScript support for better development experience
- **Auto Refresh**: Automatic data refresh with manual refresh option
- **Static Export**: Optimized for static deployment

## 🛠️ Tech Stack

- **Next.js 15** - React framework with App Router
- **TypeScript** - Type-safe development
- **shadcn/ui** - High-quality, accessible UI components
- **Tailwind CSS** - Utility-first CSS framework
- **Radix UI** - Unstyled, accessible UI primitives
- **Lucide React** - Beautiful, customizable icons

## 📦 Installation

```bash
# Install dependencies
npm install

# Start development server
npm run dev

# Build for production
npm run build

# Start production server
npm run start

# Type checking
npm run type-check
```

## 🔧 Development

### Project Structure

```
web-nextjs/
├── app/                          # Next.js App Router pages
│   ├── layout.tsx               # Root layout
│   ├── page.tsx                 # Dashboard page
│   ├── status/page.tsx          # System status
│   ├── plugins/page.tsx         # Plugin management
│   ├── logs/page.tsx            # Log viewer
│   ├── config/page.tsx          # Configuration
│   ├── interpreters/page.tsx    # Python interpreters
│   ├── connection/page.tsx      # Connection status
│   ├── api-test/page.tsx        # API testing
│   ├── plugin-manager/page.tsx  # Plugin manager
│   └── not-found.tsx           # 404 page
├── components/                   # Reusable components
│   ├── ui/                      # shadcn/ui components
│   ├── Layout.tsx               # Main layout
│   ├── Sidebar.tsx              # Navigation sidebar
│   ├── Header.tsx               # Top header
│   ├── Breadcrumb.tsx           # Breadcrumb navigation
│   ├── SystemHealthBanner.tsx   # Health status banner
│   └── LogViewer.tsx            # Log viewing component
├── hooks/                        # Custom React hooks
│   ├── useDashboard.ts          # Dashboard data management
│   └── useRealTimeMonitoring.ts # Real-time monitoring
├── services/                     # API and external services
│   ├── api.ts                   # HTTP API client
│   └── websocket.ts             # WebSocket service
├── types/                        # TypeScript type definitions
│   └── api.ts                   # API response types
└── lib/                         # Utility functions
    └── utils.ts                 # Common utilities
```

### API Integration

The dashboard connects to the Go backend through the `/api/dashboard` endpoints:

- `GET /api/dashboard/status` - System status
- `GET /api/dashboard/stats` - Dashboard statistics
- `GET /api/dashboard/plugins` - Plugin information
- `GET /api/dashboard/workers` - Worker pool status
- `GET /api/dashboard/logs` - System logs
- `GET /api/dashboard/activity` - Recent activity

### Adding New Components

To add new shadcn/ui components:

```bash
# Example: Add a new component
npx shadcn-ui@latest add dialog
```

### Customizing Theme

Edit `app/globals.css` to customize the color scheme and design tokens.

## 🌐 Deployment

### Static Export

The application is configured for static export:

```bash
npm run build
```

This generates a static site in the `dist/` directory that can be deployed to any static hosting service.

### Environment Variables

Create a `.env.local` file for local development:

```env
# API Base URL (optional, defaults to relative URLs)
NEXT_PUBLIC_API_BASE_URL=http://localhost:8000
```

## 🔄 Real-time Features

- **WebSocket Connection**: Real-time system monitoring
- **Auto-refresh**: Automatic data updates every 30 seconds
- **Live Logs**: Real-time log streaming with filters
- **System Health**: Live system status updates

## 📱 Pages

### Dashboard (`/`)
- System overview with key metrics
- Real-time activity feed
- Quick action buttons
- System health banner

### System Status (`/status`)
- Detailed system health monitoring
- Resource usage (CPU, memory, disk)
- Worker pool status
- Plugin health overview

### Plugins (`/plugins`)
- Plugin management interface
- Enable/disable plugins
- Execute plugins manually
- View plugin statistics and logs

### Logs (`/logs`)
- Real-time log viewer
- Advanced filtering and search
- Log level filtering
- Export functionality

### Configuration (`/config`)
- System configuration management
- Settings editor
- Configuration validation

### Python Interpreters (`/interpreters`)
- Python environment management
- Interpreter testing
- Package management

### Connection Status (`/connection`)
- External service monitoring
- Connection health checks
- Latency monitoring

### API Test (`/api-test`)
- Webhook endpoint testing
- Request/response debugging
- cURL generation

### Plugin Manager (`/plugin-manager`)
- Plugin installation
- Plugin updates
- Plugin marketplace

## 🎨 Design System

The dashboard follows the shadcn/ui design system with:

- **Consistent Typography**: Standardized text styles
- **Color Palette**: Professional dark theme
- **Spacing System**: Consistent spacing throughout
- **Component Library**: Reusable, accessible components
- **Responsive Grid**: Mobile-first responsive design

## 🔒 Error Handling

- **Graceful Degradation**: Shows fallback content when APIs fail
- **Retry Logic**: Automatic retry for failed requests
- **User Feedback**: Clear error messages and loading states
- **Error Boundaries**: Prevents crashes from component errors

## 🚧 Development Status

### ✅ Completed
- Core infrastructure and routing
- All page layouts and navigation
- API client and WebSocket services
- Real-time monitoring hooks
- System health monitoring
- Plugin management interface
- Log viewer with advanced features
- Responsive design and dark theme

### 🔄 In Progress
- Advanced plugin configuration
- Metrics visualization
- User authentication

### 📋 Planned
- Interactive API documentation
- Advanced webhook testing
- Export and reporting features
- Notification system

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests if applicable
5. Submit a pull request

## 📄 License

This project is part of the Webhook Bridge service.
