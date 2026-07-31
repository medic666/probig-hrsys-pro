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
bunx vite build
echo "  前端构建完成 -> server/cmd/server/static/"

echo ""
echo "[2/3] 编译 Go 二进制..."
cd "$PROJECT_ROOT/server"
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