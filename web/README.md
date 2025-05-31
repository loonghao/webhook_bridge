# Webhook Bridge Dashboard

A modern, responsive dashboard built with React, TypeScript, and shadcn/ui for the Webhook Bridge service.

## 🚀 Features

- **Modern UI**: Built with shadcn/ui components and Tailwind CSS
- **Real-time Data**: Connects to Go backend APIs for live dashboard updates
- **Responsive Design**: Works seamlessly on desktop and mobile devices
- **Dark Theme**: Professional dark mode interface
- **Type Safety**: Full TypeScript support for better development experience
- **Auto Refresh**: Automatic data refresh with manual refresh option

## 🛠️ Tech Stack

- **React 18** - Modern React with hooks
- **TypeScript** - Type-safe development
- **Vite** - Fast build tool and dev server
- **shadcn/ui** - High-quality, accessible UI components
- **Tailwind CSS** - Utility-first CSS framework
- **Radix UI** - Unstyled, accessible UI primitives
- **Lucide React** - Beautiful, customizable icons
- **React Router** - Client-side routing

## 📦 Installation

```bash
# Install dependencies
npm install

# Start development server
npm run dev

# Build for production
npm run build

# Preview production build
npm run preview
```

## 🔧 Development

### Project Structure

```
src/
├── components/          # Reusable UI components
│   ├── ui/             # shadcn/ui components
│   ├── Header.tsx      # Top navigation bar
│   ├── Sidebar.tsx     # Side navigation
│   └── Layout.tsx      # Main layout wrapper
├── pages/              # Page components
│   └── Dashboard.tsx   # Main dashboard page
├── hooks/              # Custom React hooks
│   └── useDashboard.ts # Dashboard data management
├── services/           # API services
│   └── api.ts          # API client and utilities
├── types/              # TypeScript type definitions
│   └── api.ts          # API response types
├── lib/                # Utility functions
│   └── utils.ts        # Common utilities
├── App.tsx             # Main app component
├── main.tsx            # App entry point
└── index.css           # Global styles and CSS variables
```

### API Integration

The dashboard connects to the Go backend through the `/api/dashboard` endpoints:

- `GET /api/dashboard/status` - System status
- `GET /api/dashboard/stats` - Dashboard statistics
- `GET /api/dashboard/plugins` - Plugin information
- `GET /api/dashboard/workers` - Worker pool status
- `GET /api/dashboard/logs` - System logs

### Adding New Components

To add new shadcn/ui components:

```bash
# Example: Add a new component (if using shadcn CLI)
npx shadcn-ui@latest add dialog
```

### Customizing Theme

Edit `src/index.css` to customize the color scheme and design tokens.

## 🌐 Deployment

The dashboard builds to static files in the `dist/` directory and can be served by the Go backend or any static file server.

### Integration with Go Backend

The Go backend should serve the built files from the `dist/` directory and proxy API requests to the appropriate handlers.

## 📱 Features Overview

### Dashboard Page
- **System Statistics**: Real-time metrics and KPIs
- **Recent Activity**: Latest webhook events and system activities
- **System Status**: Health checks for all services
- **Auto Refresh**: Configurable automatic data updates

### Responsive Design
- **Mobile-first**: Optimized for mobile devices
- **Adaptive Layout**: Adjusts to different screen sizes
- **Touch-friendly**: Large touch targets for mobile interaction

### Error Handling
- **Graceful Degradation**: Shows fallback content when APIs fail
- **Retry Logic**: Automatic retry for failed requests
- **User Feedback**: Clear error messages and loading states

## 🔄 Data Flow

1. **useDashboard Hook**: Manages all dashboard data and state
2. **API Client**: Handles HTTP requests with retry logic
3. **Type Safety**: TypeScript ensures data consistency
4. **Real-time Updates**: Automatic refresh every 30 seconds
5. **Error Recovery**: Graceful handling of network issues

## 🎨 Design System

The dashboard follows the shadcn/ui design system with:

- **Consistent Spacing**: Using Tailwind's spacing scale
- **Color Palette**: Professional dark theme with accent colors
- **Typography**: Clear hierarchy with appropriate font weights
- **Interactive States**: Hover, focus, and active states for all interactive elements

This modern dashboard provides a professional, user-friendly interface for managing and monitoring the Webhook Bridge service.
