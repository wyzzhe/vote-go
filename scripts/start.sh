#!/bin/bash

# 启动脚本 - 实时投票系统

echo "🚀 启动实时投票系统..."

# 检查 Docker 是否安装
if ! command -v docker &> /dev/null; then
    echo "❌ Docker 未安装，请先安装 Docker"
    exit 1
fi

if ! command -v docker-compose &> /dev/null; then
    echo "❌ Docker Compose 未安装，请先安装 Docker Compose"
    exit 1
fi

# 切换到项目根目录
cd "$(dirname "$0")/.."

echo "📦 启动 Docker 容器..."
docker-compose up -d

echo "⏳ 等待服务启动..."
sleep 10

# 检查服务状态
echo "🔍 检查服务状态..."
docker-compose ps

echo ""
echo "✅ 服务启动完成！"
echo ""
echo "📱 前端地址: http://localhost:3000"
echo "🔧 后端API: http://localhost:8080/api"
echo "🗄️  数据库: localhost:3306 (用户名: root, 密码: password)"
echo ""
echo "📝 查看日志:"
echo "  前端: docker-compose logs -f frontend"
echo "  后端: docker-compose logs -f backend"
echo "  数据库: docker-compose logs -f mysql"
echo ""
echo "⏹️  停止服务: docker-compose down" 