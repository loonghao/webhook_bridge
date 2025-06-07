# Docker Deployment Guide

This guide explains how to deploy webhook-bridge using Docker and Docker Compose.

## 🐳 Quick Start

### Using Docker Compose (Recommended)

1. **Clone the repository**:
   ```bash
   git clone https://github.com/loonghao/webhook_bridge.git
   cd webhook_bridge
   ```

2. **Create required directories**:
   ```bash
   mkdir -p config plugins logs data
   ```

3. **Copy configuration**:
   ```bash
   cp config.yaml config/webhook-bridge.yaml
   ```

4. **Start the service**:
   ```bash
   docker-compose up -d
   ```

5. **Access the dashboard**:
   Open http://localhost:8000 in your browser

### Using Docker directly

```bash
# Pull the latest image
docker pull ghcr.io/loonghao/webhook-bridge:latest

# Create directories
mkdir -p config plugins logs data

# Run the container
docker run -d \
  --name webhook-bridge \
  -p 8000:8000 \
  -p 50051:50051 \
  -v $(pwd)/config:/app/config \
  -v $(pwd)/plugins:/app/plugins \
  -v $(pwd)/logs:/app/logs \
  -v $(pwd)/data:/app/data \
  -v $(pwd)/config.yaml:/app/config.yaml:ro \
  -e WEBHOOK_BRIDGE_CONFIG_PATH=/app/config \
  -e WEBHOOK_BRIDGE_PLUGINS_PATH=/app/plugins:/app/example_plugins \
  ghcr.io/loonghao/webhook-bridge:latest
```

## 📁 Directory Structure

The Docker container expects the following directory structure:

```
/app/
├── config/          # Configuration files
├── plugins/         # Custom plugins
├── logs/           # Log files
├── data/           # Persistent data
├── example_plugins/ # Built-in example plugins
└── config.yaml     # Main configuration file
```

## 🔧 Environment Variables

### Core Configuration

| Variable | Description | Default |
|----------|-------------|---------|
| `WEBHOOK_BRIDGE_CONFIG_PATH` | Configuration directory | `/app/config` |
| `WEBHOOK_BRIDGE_PLUGINS_PATH` | Plugin search paths (colon-separated) | `/app/plugins:/app/example_plugins` |
| `WEBHOOK_BRIDGE_LOG_PATH` | Log directory | `/app/logs` |
| `WEBHOOK_BRIDGE_DATA_PATH` | Data directory | `/app/data` |
| `WEBHOOK_BRIDGE_WEB_PATH` | Web dashboard path | `/app/web-nextjs/dist` |
| `WEBHOOK_BRIDGE_PYTHON_PATH` | Python executor path | `/app/python_executor` |

### Server Configuration

| Variable | Description | Default |
|----------|-------------|---------|
| `WEBHOOK_BRIDGE_HOST` | Server bind address | `0.0.0.0` |
| `WEBHOOK_BRIDGE_PORT` | HTTP server port | `8000` |
| `WEBHOOK_BRIDGE_GRPC_PORT` | gRPC server port | `50051` |
| `WEBHOOK_BRIDGE_MODE` | Server mode (debug/release) | `release` |

### Python Configuration

| Variable | Description | Default |
|----------|-------------|---------|
| `PYTHONPATH` | Python module search path | `/app` |

## 📋 Docker Compose Profiles

The `docker-compose.yml` includes several profiles for different use cases:

### Production (default)
```bash
docker-compose up -d
```

### Development
```bash
docker-compose --profile dev up -d
```

### Separate Services (for debugging)
```bash
docker-compose --profile dev-separate up -d
```

### With Optional Services
```bash
docker-compose --profile optional up -d
```

## 🔍 Health Checks

The container includes built-in health checks:

```bash
# Check container health
docker ps

# Manual health check
docker exec webhook-bridge wget --no-verbose --tries=1 --spider http://localhost:8000/health
```

## 📊 Monitoring and Logs

### View logs
```bash
# All logs
docker-compose logs -f

# Specific service
docker-compose logs -f webhook-bridge

# Follow logs
docker logs -f webhook-bridge
```

### Access metrics
- Health endpoint: http://localhost:8000/health
- Dashboard: http://localhost:8000/
- API endpoints: http://localhost:8000/api/

## 🔧 Configuration Examples

### Basic Configuration

Create `config/webhook-bridge.yaml`:

```yaml
server:
  host: "0.0.0.0"
  port: 8000
  mode: "release"

logging:
  level: "info"
  file: "/app/logs/webhook-bridge.log"

plugins:
  directories:
    - "/app/plugins"
    - "/app/example_plugins"

python:
  executable: "python"
  grpc_port: 50051
```

### Advanced Configuration with External Services

```yaml
server:
  host: "0.0.0.0"
  port: 8000
  mode: "release"

database:
  type: "postgres"
  host: "postgres"
  port: 5432
  name: "webhook_bridge"
  user: "webhook"
  password: "webhook_password"

cache:
  type: "redis"
  host: "redis"
  port: 6379

logging:
  level: "info"
  file: "/app/logs/webhook-bridge.log"
```

## 🚀 Production Deployment

### Using Docker Swarm

```yaml
version: '3.8'

services:
  webhook-bridge:
    image: ghcr.io/loonghao/webhook-bridge:latest
    ports:
      - "8000:8000"
      - "50051:50051"
    volumes:
      - webhook_config:/app/config
      - webhook_plugins:/app/plugins
      - webhook_logs:/app/logs
      - webhook_data:/app/data
    environment:
      - WEBHOOK_BRIDGE_MODE=release
    deploy:
      replicas: 2
      restart_policy:
        condition: on-failure
      resources:
        limits:
          memory: 512M
        reservations:
          memory: 256M

volumes:
  webhook_config:
  webhook_plugins:
  webhook_logs:
  webhook_data:
```

### Using Kubernetes

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: webhook-bridge
spec:
  replicas: 2
  selector:
    matchLabels:
      app: webhook-bridge
  template:
    metadata:
      labels:
        app: webhook-bridge
    spec:
      containers:
      - name: webhook-bridge
        image: ghcr.io/loonghao/webhook-bridge:latest
        ports:
        - containerPort: 8000
        - containerPort: 50051
        env:
        - name: WEBHOOK_BRIDGE_MODE
          value: "release"
        volumeMounts:
        - name: config
          mountPath: /app/config
        - name: plugins
          mountPath: /app/plugins
        - name: logs
          mountPath: /app/logs
        - name: data
          mountPath: /app/data
      volumes:
      - name: config
        configMap:
          name: webhook-bridge-config
      - name: plugins
        persistentVolumeClaim:
          claimName: webhook-bridge-plugins
      - name: logs
        persistentVolumeClaim:
          claimName: webhook-bridge-logs
      - name: data
        persistentVolumeClaim:
          claimName: webhook-bridge-data
```

## 🛠 Troubleshooting

### Common Issues

1. **Container fails to start**:
   ```bash
   # Check logs
   docker logs webhook-bridge
   
   # Check configuration
   docker exec webhook-bridge cat /app/config.yaml
   ```

2. **Permission issues**:
   ```bash
   # Fix ownership
   sudo chown -R 1000:1000 config plugins logs data
   ```

3. **Port conflicts**:
   ```bash
   # Use different ports
   docker run -p 8080:8000 -p 50052:50051 ...
   ```

4. **Plugin not found**:
   ```bash
   # Check plugin directory
   docker exec webhook-bridge ls -la /app/plugins
   
   # Check environment variables
   docker exec webhook-bridge env | grep WEBHOOK_BRIDGE
   ```

### Debug Mode

Run container in debug mode:

```bash
docker run -it --rm \
  -p 8000:8000 \
  -p 50051:50051 \
  -v $(pwd)/config:/app/config \
  -e WEBHOOK_BRIDGE_MODE=debug \
  ghcr.io/loonghao/webhook-bridge:latest
```

## 🔐 Docker Registry Authentication

### GitHub Container Registry (GHCR)

The project is configured to automatically publish Docker images to GitHub Container Registry (ghcr.io). **No additional token configuration is required** for the repository owner.

#### Automatic Configuration
- ✅ **GITHUB_TOKEN**: Automatically provided by GitHub Actions
- ✅ **Permissions**: Already configured in `.github/workflows/release.yml`
  ```yaml
  permissions:
    contents: write
    packages: write
    id-token: write
  ```

#### Manual Docker Push (for maintainers)
If you need to manually push images:

```bash
# Login to GHCR
echo $GITHUB_TOKEN | docker login ghcr.io -u USERNAME --password-stdin

# Build and push
docker build -t ghcr.io/loonghao/webhook-bridge:latest .
docker push ghcr.io/loonghao/webhook-bridge:latest
```

#### For Contributors
Contributors don't need any special configuration. Docker images are automatically built and published when:
1. A new tag is pushed (e.g., `v1.0.0`)
2. The release workflow runs successfully
3. GoReleaser handles the Docker build and push automatically

## 📚 Additional Resources

- [Plugin Development](./PLUGIN_DEVELOPMENT.md)
- [Dashboard Guide](./DASHBOARD_GUIDE.md)
- [GitHub Repository](https://github.com/loonghao/webhook_bridge)
- [Docker Hub](https://hub.docker.com/r/loonghao/webhook-bridge)
