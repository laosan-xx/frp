#!/bin/bash

# FRP GoReleaser Docker 构建脚本
set -e

echo "🚀 开始构建 FRP GoReleaser Docker 镜像..."

# 检查 Docker 是否运行
if ! docker info > /dev/null 2>&1; then
    echo "❌ Docker 未运行，请先启动 Docker"
    exit 1
fi

# 构建镜像
echo "📦 构建 Docker 镜像..."
docker build -f Dockerfile.goreleaser -t frp-goreleaser:latest .

echo "✅ Docker 镜像构建完成！"

# 显示镜像信息
echo "📋 镜像信息："
docker images frp-goreleaser:latest

echo ""
echo "🎯 使用方法："
echo "1. 直接运行容器："
echo "   docker run -it --rm -v \$(pwd):/workspace frp-goreleaser:latest"
echo ""
echo "2. 使用 docker-compose："
echo "   docker-compose -f docker-compose.goreleaser.yml up -d"
echo "   docker-compose -f docker-compose.goreleaser.yml exec frp-builder bash"
echo ""
echo "3. 在容器中执行构建："
echo "   ./package.sh"
echo ""
echo "4. 运行 GoReleaser："
echo "   goreleaser release --clean --release-notes=./Release.md"
