#!/usr/bin/env bash
# stack-up.sh — bring up the full Stratavore Docker stack
#
# Usage: ./scripts/stack-up.sh [--build]
#   --build   rebuild the daemon image before starting

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

echo "==> Bringing up Stratavore stack..."

if [ "$BUILD" = true ]; then
  echo "==> Rebuilding daemon image..."
  sudo docker-compose build stratavored
fi

sudo docker-compose up -d

echo ""
echo "==> Stack is up. Services:"
sudo docker-compose ps --format "table {{.Name}}\t{{.Status}}\t{{.Ports}}" 2>/dev/null \
  || sudo docker-compose ps
echo ""
echo "==> Daemon API: http://localhost:8080"
echo "==> RabbitMQ UI: http://localhost:15672  (guest/guest)"
echo "==> Grafana:     http://localhost:3000"
echo ""
echo "==> Use 'stratavore status' to verify daemon health."
