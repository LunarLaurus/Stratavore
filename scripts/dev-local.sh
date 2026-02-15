#!/usr/bin/env bash
# dev-local.sh — run the daemon from source for local development
#
# Starts infrastructure containers (postgres, rabbitmq, redis, qdrant),
# stops the Docker daemon container, builds from source, and runs the
# daemon binary directly so code changes can be tested without rebuilding
# a Docker image.
#
# Usage: ./scripts/dev-local.sh [--no-build] [--go-path /path/to/go/bin]
#   --no-build        skip make build (use existing bin/stratavored)
#   --go-path PATH    path to go binary directory (default: /home/meridian/go/bin)

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_DIR="$(dirname "$SCRIPT_DIR")"

BUILD=true
GO_BIN_PATH="/home/meridian/go/bin"

while [[ $# -gt 0 ]]; do
  case "$1" in
    --no-build)    BUILD=false; shift ;;
    --go-path)     GO_BIN_PATH="$2"; shift 2 ;;
    *) echo "Unknown argument: $1"; exit 1 ;;
  esac
done

export PATH="$GO_BIN_PATH:$PATH"
cd "$PROJECT_DIR"

# Start infrastructure only (exclude the daemon container)
echo "==> Starting infrastructure containers..."
sudo docker-compose up -d postgres rabbitmq redis qdrant

# Stop the Docker daemon container so it doesn't compete on port 8080
echo "==> Stopping Docker daemon container (if running)..."
sudo docker-compose stop stratavored 2>/dev/null || true

# Wait for postgres to be ready
echo "==> Waiting for PostgreSQL..."
for i in $(seq 1 15); do
  if sudo docker-compose exec -T postgres pg_isready -U postgres -q 2>/dev/null; then
    echo "    PostgreSQL ready."
    break
  fi
  if [ "$i" -eq 15 ]; then
    echo "    ERROR: PostgreSQL did not become ready in time."
    exit 1
  fi
  sleep 2
done

# Build from source
if [ "$BUILD" = true ]; then
  echo "==> Building..."
  if ! command -v go &>/dev/null; then
    echo "    ERROR: 'go' not found. Add $GO_BIN_PATH to PATH or pass --go-path."
    exit 1
  fi
  make build
  echo "    Build complete."
fi

echo ""
echo "==> Starting local daemon (Ctrl+C to stop)..."
echo "==> API:  http://localhost:8080"
echo "==> Use 'stratavore status' in another terminal to verify."
echo ""

exec ./bin/stratavored
