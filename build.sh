#!/usr/bin/env bash
set -euo pipefail

PROJECT_ROOT="$(cd "$(dirname "$0")" && pwd)"
SERVER_DIST="$PROJECT_ROOT/server/dist"
DIST_DIR="$PROJECT_ROOT/dist"

echo "=== Probig 单二进制构建 ==="

echo ""
echo "[1/3] 构建前端..."
cd "$PROJECT_ROOT/web"
if [ ! -d "node_modules" ]; then
  echo "  安装前端依赖..."
  bun install
fi
bunx vue-tsc --noEmit
bun run test
bunx eslint . --ext .vue,.ts,.tsx
bunx vite build
echo "  前端构建完成 -> server/cmd/server/static/"

echo ""
echo "[2/3] 后端检查与编译..."
cd "$PROJECT_ROOT/server"
go vet ./...
go test ./...
go build -o dist/probig-server ./cmd/server/
echo "  Go 编译完成 -> server/dist/probig-server"

echo ""
echo "[3/3] 构建生产交付文件..."
mkdir -p "$DIST_DIR"
cp "$SERVER_DIST/probig-server" "$DIST_DIR/probig-server"

echo ""
echo "=== 构建完成 ==="
echo ""
echo "开发测试:  cd server && ./dist/probig-server"
echo "生产部署:  cd dist && ./probig-server"
echo " (首次启动会自动生成 config.yaml 并初始化数据库)"