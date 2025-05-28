#!/bin/bash

# 开发模式启动脚本

echo "🛠️  启动开发模式..."

# 切换到项目根目录
cd "$(dirname "$0")/.."

# 检查是否有 Go 环境
if ! command -v go &> /dev/null; then
    echo "❌ Go 未安装，请先安装 Go 1.19+"
    exit 1
fi

# 检查是否有 Node.js 环境
if ! command -v node &> /dev/null; then
    echo "❌ Node.js 未安装，请先安装 Node.js 16+"
    exit 1
fi

echo "🗄️  启动 MySQL 数据库..."
docker-compose up -d mysql

echo "⏳ 等待 MySQL 启动..."
sleep 15

echo "📦 安装后端依赖..."
cd backend
go mod tidy

echo "🔧 启动后端服务 (端口 8080)..."
go run main.go &
BACKEND_PID=$!

cd ../frontend

echo "📦 安装前端依赖..."
npm install

echo "🎨 启动前端服务 (端口 3000)..."
npm run dev &
FRONTEND_PID=$!

echo ""
echo "✅ 开发环境启动完成！"
echo ""
echo "📱 前端地址: http://localhost:3000"
echo "🔧 后端API: http://localhost:8080/api"
echo "🗄️  数据库: localhost:3306"
echo ""
echo "⏹️  按 Ctrl+C 停止所有服务"

# 设置信号处理器，在脚本退出时清理后台进程
trap "echo '🛑 停止服务...'; kill $BACKEND_PID $FRONTEND_PID; docker-compose stop mysql; exit" INT TERM

# 等待进程结束
wait 