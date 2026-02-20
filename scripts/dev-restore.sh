#!/usr/bin/env bash
# dev-restore.sh — switch back from local dev mode to Docker daemon
#
# Kills any local stratavored process and restarts the Docker daemon
# container. Use this after dev-local.sh when you're done testing.
#
# Usage: ./scripts/dev-restore.sh [--build]
#   --build   rebuild the Docker image before starting

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_DIR="$(dirname "$SCRIPT_DIR")"

BUILD=false
for arg in "$@"; do
  case "$arg" in
    --build) BUILD=true ;;
    *) echo "Unknown argument: $arg"; exit 1 ;;
  esac
done

cd "$PROJECT_DIR"

echo "==> Stopping local daemon (if running)..."
pkill -f "bin/stratavored" 2>/dev/null && echo "    Stopped." || echo "    (not running)"
sleep 1

if [ "$BUILD" = true ]; then
  echo "==> Rebuilding Docker daemon image..."
  sudo docker-compose build stratavored
fi

echo "==> Starting Docker daemon container..."
sudo docker-compose up -d stratavored

echo ""
echo "==> Restored to Docker mode. Daemon: http://localhost:8080"
