# Stratavore Makefile

.PHONY: all build install clean test lint migration-up migration-down docker-setup

BINARY_NAME=stratavore
DAEMON_NAME=stratavored
AGENT_NAME=stratavore-agent
VERSION?=0.1.0
BUILD_TIME=$(shell date -u '+%Y-%m-%d_%H:%M:%S')
COMMIT=$(shell git rev-parse --short HEAD 2>/dev/null || echo "dev")
LDFLAGS=-ldflags "-X main.Version=${VERSION} -X main.BuildTime=${BUILD_TIME} -X main.Commit=${COMMIT}"

all: build

deps:
	go mod download
	go mod tidy

build:
	@echo "Building Stratavore components..."
	go build ${LDFLAGS} -o bin/${BINARY_NAME} ./cmd/stratavore
	go build ${LDFLAGS} -o bin/${DAEMON_NAME} ./cmd/stratavored
	go build ${LDFLAGS} -o bin/${AGENT_NAME} ./cmd/stratavore-agent
	@echo "Build complete: bin/"

install: build
	@echo "Installing Stratavore to /usr/local/bin..."
	sudo cp bin/${BINARY_NAME} /usr/local/bin/
	sudo cp bin/${DAEMON_NAME} /usr/local/bin/
	sudo cp bin/${AGENT_NAME} /usr/local/bin/
	sudo chmod +x /usr/local/bin/${BINARY_NAME}
	sudo chmod +x /usr/local/bin/${DAEMON_NAME}
	sudo chmod +x /usr/local/bin/${AGENT_NAME}
	@echo "Creating config directory..."
	mkdir -p ~/.config/stratavore
	@echo "Installation complete!"

clean:
	rm -rf bin/
	rm -f stratavore.db

test:
	go test -v -race -coverprofile=coverage.out ./...

test-integration:
	go test -v -race -tags=integration ./test/integration/...

lint:
	go vet ./...
	staticcheck ./...
	golangci-lint run

migration-up:
	@echo "Running database migrations (up)..."
	./scripts/migrate.sh up

migration-down:
	@echo "Rolling back database migrations..."
	./scripts/migrate.sh down

docker-setup:
	@echo "Setting up Docker infrastructure integration..."
	./scripts/setup-docker-integration.sh

systemd-install:
	@echo "Installing systemd service..."
	sudo cp deployments/systemd/stratavored.service /etc/systemd/system/
	sudo systemctl daemon-reload
	sudo systemctl enable stratavored
	@echo "Service installed. Start with: sudo systemctl start stratavored"

format:
	gofmt -w -s .
	goimports -w .

run-daemon:
	go run ./cmd/stratavored

run-cli:
	go run ./cmd/stratavore

proto:
	protoc --go_out=. --go-grpc_out=. pkg/api/stratavore.proto

proto:
	protoc --go_out=. --go-grpc_out=. pkg/api/stratavore.proto

.PHONY: help
help:
	@echo "Stratavore Build System"
	@echo ""
	@echo "Targets:"
	@echo "  all              - Build all components (default)"
	@echo "  deps             - Download dependencies"
	@echo "  build            - Build CLI, daemon, and agent"
	@echo "  install          - Install binaries to /usr/local/bin"
	@echo "  clean            - Remove build artifacts"
	@echo "  test             - Run unit tests"
	@echo "  test-integration - Run integration tests"
	@echo "  lint             - Run linters"
	@echo "  proto            - Generate protobuf Go code"
	@echo "  migration-up     - Apply database migrations"
	@echo "  migration-down   - Rollback database migrations"
	@echo "  docker-setup     - Configure Docker integration"
	@echo "  systemd-install  - Install systemd service"
	@echo "  format           - Format Go code"
	@echo "  run-daemon       - Run daemon in dev mode"
	@echo "  run-cli          - Run CLI in dev mode"
