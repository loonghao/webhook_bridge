PROJECT_NAME := webhook-bridge
BIN := webhook-bridge

.PHONY: help build release-build run admin worker test fmt lint dashboard-check clean

help:
	@echo "$(PROJECT_NAME) 4.0"
	@echo "  build           Build debug CLI"
	@echo "  release-build   Build release CLI"
	@echo "  run             Run API + configured workers"
	@echo "  admin           Show admin summary"
	@echo "  worker          Start one standalone worker"
	@echo "  test            Run Rust and dashboard checks"
	@echo "  fmt             Format Rust"
	@echo "  lint            Check Rust formatting"
	@echo "  dashboard-check Type-check and lint dashboard"
	@echo "  clean           Remove Rust build output"

build:
	cargo build -p webhook-bridge-server --bin $(BIN)

release-build:
	cargo build --release -p webhook-bridge-server --bin $(BIN)

run: build
	./target/debug/$(BIN) run --config config.4.0.yaml

admin: build
	./target/debug/$(BIN) admin --config config.4.0.yaml

worker: build
	./target/debug/$(BIN) worker start --config config.4.0.yaml --index 0

test:
	cargo test
	cd web-nextjs && npm run type-check && npm run lint

fmt:
	cargo fmt

lint:
	cargo fmt --check

dashboard-check:
	cd web-nextjs && npm run type-check && npm run lint

clean:
	cargo clean
