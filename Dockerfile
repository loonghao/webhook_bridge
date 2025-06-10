# Multi-stage Dockerfile for webhook bridge unified architecture

# Stage 1: Build Go application
FROM golang:1.23-alpine AS go-builder

WORKDIR /app

# Install build dependencies
RUN apk add --no-cache git

# Copy Go modules files
COPY go.mod go.sum ./

# Copy local packages first (needed for replace directives)
COPY web-nextjs/ web-nextjs/
COPY pkg/ pkg/

RUN go mod download

# Copy source code
COPY cmd/ cmd/
COPY internal/ internal/
COPY api/ api/

# Build the unified application
ARG VERSION=2.0.0-unified
ARG GIT_COMMIT=unknown
ARG BUILD_DATE=unknown

RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags "-X main.version=${VERSION} \
              -X main.buildTime=${BUILD_DATE} \
              -X main.goVersion=$(go version | cut -d' ' -f3)" \
    -o webhook-bridge ./cmd/webhook-bridge

# Stage 2: Python environment
FROM python:3.11-slim AS python-base

WORKDIR /app

# Install system dependencies for building Python packages
RUN apt-get update && apt-get install -y \
    gcc \
    g++ \
    make \
    pkg-config \
    libffi-dev \
    libssl-dev \
    python3-dev \
    && rm -rf /var/lib/apt/lists/*

# Install UV
RUN pip install --no-cache-dir uv

# Copy Python requirements
COPY requirements.txt ./

# Create virtual environment and install dependencies using uv
RUN uv venv /opt/venv
ENV VIRTUAL_ENV=/opt/venv
ENV PATH="/opt/venv/bin:$PATH"

# Install Python dependencies using uv with verbose output
RUN echo "Installing Python dependencies..." && \
    cat requirements.txt && \
    uv pip install --verbose -r requirements.txt

# Stage 3: Final runtime image
FROM python:3.11-slim

WORKDIR /app

# Install runtime dependencies
RUN apt-get update && apt-get install -y \
    ca-certificates \
    wget \
    netcat-openbsd \
    && rm -rf /var/lib/apt/lists/*

# Copy Python virtual environment
COPY --from=python-base /opt/venv /opt/venv
ENV PATH="/opt/venv/bin:$PATH"

# Copy unified Go binary
COPY --from=go-builder /app/webhook-bridge /usr/local/bin/

# Copy Python source code
COPY python_executor/ python_executor/
COPY api/ api/
COPY example_plugins/ example_plugins/

# Copy dashboard files
COPY web-nextjs/dist/ web-nextjs/dist/

# Copy configuration and entrypoint script
COPY config.yaml ./
COPY docker-entrypoint.sh /usr/local/bin/

# Create directories for configuration and plugins
RUN mkdir -p /app/config /app/plugins /app/logs /app/data

# Make entrypoint script executable
RUN chmod +x /usr/local/bin/docker-entrypoint.sh

# Environment variables for configuration
ENV WEBHOOK_BRIDGE_CONFIG_PATH="/app/config"
ENV WEBHOOK_BRIDGE_PLUGINS_PATH="/app/plugins:/app/example_plugins"
ENV WEBHOOK_BRIDGE_LOG_PATH="/app/logs"
ENV WEBHOOK_BRIDGE_DATA_PATH="/app/data"
ENV WEBHOOK_BRIDGE_WEB_PATH="/app/web-nextjs/dist"
ENV WEBHOOK_BRIDGE_PYTHON_PATH="/app/python_executor"

# Create non-root user
RUN useradd -r -s /bin/false webhook && \
    chown -R webhook:webhook /app

USER webhook

# Expose ports (unified service uses 8080 by default)
EXPOSE 8080 50051

# Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1

# Volume mounts for external configuration
VOLUME ["/app/config", "/app/plugins", "/app/logs", "/app/data"]

# Default command (can be overridden)
# Use start command to automatically manage Python executor and Go server
ENTRYPOINT ["docker-entrypoint.sh"]
CMD ["webhook-bridge", "start"]
