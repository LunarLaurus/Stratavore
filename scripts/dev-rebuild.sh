#!/usr/bin/env bash
# dev-rebuild.sh — rebuild binaries and hot-swap the local daemon
#
# Kills any running local stratavored process, rebuilds from source,
# and starts the new binary. Infrastructure containers are left untouched.
#
# Usage: ./scripts/dev-rebuild.sh [--go-path /path/to/go/bin]
#   --go-path PATH    path to go binary directory (default: /home/meridian/go/bin)
#
# Intended to be run in a shell where dev-local.sh was previously used.

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_DIR="$(dirname "$SCRIPT_DIR")"

GO_BIN_PATH="/home/meridian/go/bin"
while [[ $# -gt 0 ]]; do
  case "$1" in
    --go-path) GO_BIN_PATH="$2"; shift 2 ;;
    *) echo "Unknown argument: $1"; exit 1 ;;
  esac
done

export PATH="$GO_BIN_PATH:$PATH"
cd "$PROJECT_DIR"

echo "==> Stopping local daemon..."
pkill -f "bin/stratavored" 2>/dev/null && echo "    Stopped." || echo "    (not running)"
sleep 1

echo "==> Rebuilding..."
make build
echo "    Build complete."

echo ""
echo "==> Starting new daemon (Ctrl+C to stop)..."
echo "==> API:  http://localhost:8080"
echo ""

exec ./bin/stratavored
