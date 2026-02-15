#!/usr/bin/env bash
# stack-down.sh — stop the full Stratavore Docker stack
#
# Usage: ./scripts/stack-down.sh [--volumes]
#   --volumes   also remove persistent volumes (wipes database state)

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_DIR="$(dirname "$SCRIPT_DIR")"

VOLUMES=false
for arg in "$@"; do
  case "$arg" in
    --volumes) VOLUMES=true ;;
    *) echo "Unknown argument: $arg"; exit 1 ;;
  esac
done

cd "$PROJECT_DIR"

if [ "$VOLUMES" = true ]; then
  echo "==> Stopping stack and removing volumes (database state will be lost)..."
  sudo docker-compose down -v
else
  echo "==> Stopping stack (volumes preserved)..."
  sudo docker-compose down
fi

echo "==> Stack stopped."
