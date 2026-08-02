#!/usr/bin/env bash
set -euo pipefail

PROJECT_ROOT="$(cd "$(dirname "$0")" && pwd)"
SERVER_DIST="$PROJECT_ROOT/server/dist"
DIST_DIR="$PROJECT_ROOT/dist"
UPX="${UPX:-$(command -v upx || echo "$HOME/bin/upx")}"

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
echo "  前端构建完成 -> server/cmd/server/static/（每次构建自动清空，防止旧产物累积进二进制）"

echo ""
echo "[2/3] 后端检查与编译（-s -w 去除调试信息与符号表，-trimpath 可复现构建）..."
cd "$PROJECT_ROOT/server"
go vet ./...
go test ./...
go build -trimpath -ldflags "-s -w" -o dist/probig-server ./cmd/server/
echo "  Go 编译完成 -> server/dist/probig-server"

echo ""
echo "[3/3] 构建生产交付文件（UPX 压缩）..."
mkdir -p "$DIST_DIR"
cp "$SERVER_DIST/probig-server" "$DIST_DIR/probig-server"
if [ -x "$UPX" ]; then
  "$UPX" --best --lzma "$DIST_DIR/probig-server"
  echo "  UPX 压缩完成"
else
  echo "  警告: 未找到 upx（可设置 UPX 环境变量指定路径），跳过压缩"
fi

echo ""
SIZE=$(du -h "$DIST_DIR/probig-server" | cut -f1)
echo "=== 构建完成（dist/probig-server 体积: $SIZE）==="
echo ""
echo "开发测试:  cd server && ./dist/probig-server"
echo "生产部署:  cd dist && ./probig-server"
echo " (首次启动会自动生成 config.yaml 并初始化数据库)"
