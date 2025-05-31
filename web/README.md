# Modern Dashboard - Simplified Version

This directory contains the simplified modern dashboard implementation for the Webhook Bridge project.

## Structure

```
web/
├── templates/
│   ├── dashboard.html          # Original Bootstrap-based template
│   └── modern-dashboard.html   # New simplified modern template
├── static/
│   ├── css/
│   │   └── modern-dashboard.css # Simplified shadcn/ui inspired styles
│   ├── js/
│   │   ├── dashboard.js        # Original JavaScript
│   │   └── modern-dashboard.js # New simplified JavaScript
│   └── favicon.ico             # Favicon placeholder
└── README.md                   # This file
```

## Key Improvements

### 1. **Separation of Concerns**
- **HTML**: Clean template files without inline styles/scripts
- **CSS**: Dedicated stylesheet with modern design system
- **JavaScript**: Modular, class-based approach
- **Go**: Simplified backend with template rendering

### 2. **Modern Design System**
- **Tailwind CSS**: Utility-first CSS framework
- **shadcn/ui inspired**: Modern color scheme and components
- **Lucide Icons**: Clean, consistent iconography
- **Dark Theme**: Professional dark mode interface

### 3. **Simplified Architecture**
- **Template-based**: Uses Go's html/template package
- **Fallback Support**: Embedded template for reliability
- **Clean API**: RESTful endpoints for data
- **Modular JavaScript**: Object-oriented dashboard class

### 4. **Reduced Complexity**
- **From 888 lines to 327 lines** in dashboard.go (63% reduction)
- **Removed inline HTML/CSS/JS** from Go code
- **Cleaner file organization**
- **Better maintainability**

## Usage

The modern dashboard is served through the `ModernDashboardHandler` which:

1. **Loads templates** from `web/templates/modern-dashboard.html`
2. **Falls back** to embedded template if file not found
3. **Serves static assets** from `web/static/`
4. **Provides API endpoints** for dynamic data

## API Endpoints

- `GET /` - Main dashboard page
- `GET /api/v1/status` - System status
- `GET /api/v1/metrics` - Performance metrics
- `GET /api/v1/plugins` - Plugin information
- `GET /api/v1/logs` - Recent logs
- `GET /api/v1/config` - Configuration data
- `GET /api/v1/workers` - Worker pool status
- `POST /api/v1/workers/jobs` - Submit new job

## Features

- **Real-time updates** every 30 seconds
- **Responsive design** for mobile and desktop
- **Interactive navigation** between sections
- **Error handling** with graceful fallbacks
- **Loading states** for better UX
- **Accessibility** with proper focus management

## Development

To extend the dashboard:

1. **Add new sections** in the HTML template
2. **Update JavaScript** to handle new data loading
3. **Add API endpoints** in the Go handler
4. **Style with CSS** using the design system

The modular architecture makes it easy to add new features without affecting existing functionality.
