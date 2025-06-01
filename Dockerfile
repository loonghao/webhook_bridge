# Multi-stage Dockerfile for webhook bridge hybrid architecture

# Stage 1: Build Go application
FROM golang:1.23-alpine AS go-builder

WORKDIR /app

# Install build dependencies
RUN apk add --no-cache git

# Copy Go modules files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY cmd/ cmd/
COPY internal/ internal/
COPY api/ api/
COPY pkg/ pkg/

# Build the application
ARG VERSION=2.0.0-hybrid
ARG GIT_COMMIT=unknown
ARG BUILD_DATE=unknown

RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags "-X github.com/loonghao/webhook_bridge/pkg/version.Version=${VERSION} \
              -X github.com/loonghao/webhook_bridge/pkg/version.GitCommit=${GIT_COMMIT} \
              -X github.com/loonghao/webhook_bridge/pkg/version.BuildDate=${BUILD_DATE}" \
    -o webhook-bridge-server ./cmd/server

# Stage 2: Python environment
FROM python:3.11-slim AS python-base

WORKDIR /app

# Install system dependencies
RUN apt-get update && apt-get install -y \
    gcc \
    && rm -rf /var/lib/apt/lists/*

# Install UV
RUN pip install uv

# Copy Python requirements
COPY pyproject.toml ./
COPY requirements*.txt ./

# Install Python dependencies
RUN uv venv /opt/venv && \
    /opt/venv/bin/pip install grpcio grpcio-tools && \
    /opt/venv/bin/pip install -r requirements.txt

# Stage 3: Final runtime image
FROM python:3.11-slim

WORKDIR /app

# Install runtime dependencies
RUN apt-get update && apt-get install -y \
    ca-certificates \
    wget \
    && rm -rf /var/lib/apt/lists/*

# Copy Python virtual environment
COPY --from=python-base /opt/venv /opt/venv
ENV PATH="/opt/venv/bin:$PATH"

# Copy Go binary
COPY --from=go-builder /app/webhook-bridge-server /usr/local/bin/

# Copy Python source code
COPY webhook_bridge/ webhook_bridge/
COPY python_executor/ python_executor/
COPY api/ api/
COPY example_plugins/ example_plugins/

# Copy dashboard files
COPY web/static/js/dist/ web/static/js/dist/

# Copy configuration
COPY config.yaml ./

# Create directories for configuration and plugins
RUN mkdir -p /app/config /app/plugins /app/logs /app/data

# Environment variables for configuration
ENV WEBHOOK_BRIDGE_CONFIG_PATH="/app/config"
ENV WEBHOOK_BRIDGE_PLUGINS_PATH="/app/plugins:/app/example_plugins"
ENV WEBHOOK_BRIDGE_LOG_PATH="/app/logs"
ENV WEBHOOK_BRIDGE_DATA_PATH="/app/data"
ENV WEBHOOK_BRIDGE_WEB_PATH="/app/web/static/js/dist"
ENV WEBHOOK_BRIDGE_PYTHON_PATH="/app/python_executor"

# Create non-root user
RUN useradd -r -s /bin/false webhook && \
    chown -R webhook:webhook /app

USER webhook

# Expose ports
EXPOSE 8000 50051

# Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8000/health || exit 1

# Volume mounts for external configuration
VOLUME ["/app/config", "/app/plugins", "/app/logs", "/app/data"]

# Default command (can be overridden)
CMD ["webhook-bridge-server", "--config", "/app/config.yaml"]
