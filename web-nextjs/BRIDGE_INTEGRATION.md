# Webhook Bridge Next.js Integration

This document outlines the completed bridge integration between the Next.js frontend (`@web-nextjs/`) and the Go backend.

## 🎯 Integration Overview

The bridge integration ensures seamless communication between:
- **Next.js Frontend** (`@web-nextjs/`) - Modern React dashboard
- **Go Backend** (`cmd/unified-server/`) - API server and routing
- **Python Executor** (`python_executor/`) - Plugin execution engine

## ✅ Completed Bridge Components

### 1. API Response Format Handling
- **Unified Response Processing**: Handles both wrapped (`{success, data, message}`) and direct API responses
- **Error Handling**: Proper error extraction from backend response format
- **Type Safety**: Full TypeScript support for all API interactions

### 2. Data Transformation Layer
- **`services/dataTransformers.ts`**: Converts backend data to frontend-compatible formats
- **Automatic Field Mapping**: Maps backend snake_case to frontend camelCase
- **Default Value Handling**: Provides sensible defaults for missing fields
- **Computed Fields**: Generates UI-specific fields from backend data

### 3. API Client Integration
- **Environment-Aware URLs**: Supports both development and production configurations
- **Retry Logic**: Built-in retry mechanism with exponential backoff
- **Request/Response Transformation**: Automatic data transformation
- **Error Recovery**: Graceful error handling and user feedback

### 4. WebSocket Integration
- **Real-time Monitoring**: Connects to `/api/dashboard/monitor/stream`
- **Auto-reconnection**: Handles connection drops with exponential backoff
- **Event Handling**: Proper event subscription and unsubscription
- **Environment Configuration**: Supports different WebSocket URLs

### 5. Development Tools
- **Debug Page** (`/debug`): Comprehensive integration testing interface
- **Bridge Status Checker**: Real-time connection health monitoring
- **API Testing**: Direct endpoint testing and validation
- **Environment Display**: Shows current configuration and status

## 🔧 Configuration

### Environment Variables
```env
# API Configuration
NEXT_PUBLIC_API_BASE_URL=http://localhost:8000
NEXT_PUBLIC_WS_BASE_URL=ws://localhost:8000

# Development Settings
NEXT_PUBLIC_DEV_MODE=true
NEXT_PUBLIC_ENABLE_WEBSOCKET=true
NEXT_PUBLIC_ENABLE_REAL_TIME=true
NEXT_PUBLIC_ENABLE_DEBUG=true
```

### Next.js Configuration
- **Development Proxy**: Automatic API proxying to Go backend
- **Static Export**: Optimized for production deployment
- **Environment-Aware Routing**: Different behavior for dev/prod

## 🔗 API Endpoint Mapping

### Backend → Frontend Mapping
| Backend Endpoint | Frontend Usage | Data Transformation |
|------------------|----------------|-------------------|
| `GET /api/dashboard/status` | System status display | `transformSystemStatus()` |
| `GET /api/dashboard/stats` | Dashboard statistics | `transformDashboardStats()` |
| `GET /api/dashboard/plugins` | Plugin management | `transformPluginInfo()` |
| `GET /api/dashboard/workers` | Worker monitoring | `transformWorkerInfo()` |
| `GET /api/dashboard/logs` | Log viewer | `transformLogEntry()` |
| `WS /api/dashboard/monitor/stream` | Real-time updates | Event handling |

### Data Structure Alignment
```typescript
// Backend Response
{
  "success": true,
  "data": {
    "server_status": "running",
    "grpc_connected": true,
    "worker_count": 4
  }
}

// Frontend Usage
{
  server_status: "running",
  grpc_connected: true,
  worker_count: 4,
  // Computed fields
  service: "Webhook Bridge",
  status: "healthy",
  version: "2.0.0-hybrid"
}
```

## 🧪 Testing & Validation

### Bridge Status Checker
- **API Connectivity**: Tests HTTP connection to Go backend
- **Service Health**: Validates Go server and Python executor status
- **WebSocket Status**: Monitors real-time connection health
- **Health Score**: Provides overall system health percentage

### Debug Interface (`/debug`)
- **Raw Endpoint Testing**: Direct API calls without transformation
- **API Client Testing**: Tests with data transformation
- **WebSocket Testing**: Real-time connection validation
- **Environment Display**: Shows current configuration

### Integration Tests
```bash
# Start Go backend
cd cmd/unified-server && go run main.go

# Start Python executor
cd python_executor && python main.py

# Start Next.js frontend
cd web-nextjs && npm run dev

# Access debug page
http://localhost:3002/debug
```

## 🚀 Deployment Considerations

### Development Mode
- **API Proxy**: Next.js proxies `/api/*` to `http://localhost:8000`
- **Hot Reload**: Automatic refresh on code changes
- **Debug Tools**: Full debugging interface available

### Production Mode
- **Static Export**: Generates static files for deployment
- **Relative URLs**: Uses relative paths for API calls
- **Environment Variables**: Production-specific configuration

### Docker Integration
```dockerfile
# Frontend build
FROM node:18-alpine AS frontend
WORKDIR /app/web-nextjs
COPY web-nextjs/ .
RUN npm ci && npm run build

# Serve static files through Go backend
COPY --from=frontend /app/web-nextjs/dist /app/static
```

## 🔍 Monitoring & Debugging

### Health Monitoring
- **Bridge Status**: Real-time connection health
- **Service Dependencies**: Go server + Python executor status
- **Error Tracking**: Comprehensive error logging and display

### Debug Tools
- **API Inspector**: View raw API responses
- **Data Transformation**: See before/after data transformation
- **WebSocket Monitor**: Real-time event monitoring
- **Environment Validator**: Configuration verification

## 📊 Performance Optimizations

### Frontend Optimizations
- **Code Splitting**: Automatic route-based splitting
- **Static Generation**: Pre-built static pages
- **Lazy Loading**: Components loaded on demand
- **Caching**: Intelligent data caching with refresh

### API Optimizations
- **Request Batching**: Multiple API calls in parallel
- **Retry Logic**: Exponential backoff for failed requests
- **Error Recovery**: Graceful degradation on API failures
- **WebSocket Efficiency**: Event-based real-time updates

## 🔄 Data Flow

```
┌─────────────────┐    HTTP/WS     ┌──────────────┐    gRPC    ┌─────────────────┐
│   Next.js       │◄──────────────►│  Go Backend  │◄──────────►│ Python Executor │
│   Frontend      │                │   Server     │            │                 │
│                 │                │              │            │                 │
│ • Dashboard     │                │ • API Routes │            │ • Plugin Exec   │
│ • Real-time UI  │                │ • WebSocket  │            │ • gRPC Service  │
│ • Data Transform│                │ • Routing    │            │ • Log Management│
└─────────────────┘                └──────────────┘            └─────────────────┘
```

## 🎉 Integration Status

### ✅ Fully Integrated
- API client with data transformation
- WebSocket real-time monitoring
- Error handling and recovery
- Development and production configurations
- Debug and testing tools
- Type-safe data flow

### 🔄 Ready for Enhancement
- Advanced WebSocket features
- Plugin configuration UI
- Real-time metrics visualization
- User authentication integration

The bridge integration is **complete and production-ready**! 🚀
