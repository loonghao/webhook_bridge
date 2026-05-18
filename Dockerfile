FROM rust:1-bookworm AS builder

WORKDIR /app
COPY Cargo.toml Cargo.lock ./
COPY crates ./crates
COPY api ./api
COPY python_executor ./python_executor
COPY webhook_bridge ./webhook_bridge

RUN cargo build --release -p webhook-bridge-server --bin webhook-bridge

FROM python:3.12-slim

WORKDIR /app
COPY python_executor/requirements.txt /tmp/requirements.txt
RUN pip install --no-cache-dir uv -r /tmp/requirements.txt
COPY --from=builder /app/target/release/webhook-bridge /usr/local/bin/webhook-bridge
COPY config.4.0.yaml ./config.4.0.yaml
COPY example_plugins ./example_plugins

EXPOSE 8080 50051 50052
CMD ["webhook-bridge", "run", "--config", "config.4.0.yaml"]
