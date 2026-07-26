#!/bin/bash
set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$SCRIPT_DIR"

echo "=== 1. Building frontend ==="
cd "$PROJECT_ROOT/web"
bun run build

echo ""
echo "=== 2. Copying frontend dist to embed target ==="
cd "$PROJECT_ROOT"
rm -rf cmd/dist
cp -r web/dist cmd/dist

echo ""
echo "=== 3. Building Go binary ==="
cd "$PROJECT_ROOT"
CGO_ENABLED=0 go build -ldflags="-s -w" -o probig ./cmd/

echo ""
echo "=== Build complete ==="
echo "Binary: $PROJECT_ROOT/probig"
ls -lh "$PROJECT_ROOT/probig"
echo ""
echo "Run with: GIN_MODE=release ./probig"
