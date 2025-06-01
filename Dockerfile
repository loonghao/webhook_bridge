# Multi-stage Dockerfile for webhook bridge hybrid architecture

# Stage 1: Build Go application
FROM golang:1.21-alpine AS go-builder

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
FROM python:3.13-slim AS python-base

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
FROM python:3.13-slim

WORKDIR /app

# Install runtime dependencies
RUN apt-get update && apt-get install -y \
    ca-certificates \
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

# Copy configuration
COPY config.yaml ./

# Create non-root user
RUN useradd -r -s /bin/false webhook && \
    chown -R webhook:webhook /app

USER webhook

# Expose ports
EXPOSE 8000 50051

# Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
    CMD curl -f http://localhost:8000/health || exit 1

# Default command (can be overridden)
CMD ["sh", "-c", "python python_executor/main.py & webhook-bridge-server"]
