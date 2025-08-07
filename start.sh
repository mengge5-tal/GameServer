#!/bin/bash

# GameServer v2.0 启动脚本
echo "🚀 Starting GameServer v2.0 with Clean Architecture..."

# 检查.env文件是否存在
if [ ! -f .env ]; then
    echo "❌ .env file not found. Please copy .env.example to .env and configure it."
    exit 1
fi

# 导出环境变量
echo "📝 Loading environment variables..."
export $(grep -v '^#' .env | grep -v '^$' | xargs)

# 检查必要的环境变量
if [ -z "$DB_HOST" ] || [ -z "$DB_NAME" ] || [ -z "$DB_USER" ] || [ -z "$DB_PASSWORD" ]; then
    echo "❌ Missing required database environment variables in .env file"
    echo "Required: DB_HOST, DB_NAME, DB_USER, DB_PASSWORD"
    exit 1
fi

echo "✅ Environment variables loaded"
echo "🗄️  Database: $DB_HOST:$DB_PORT/$DB_NAME"
echo "🌐 Server will start on: ${SERVER_HOST:-101.201.51.135}:${SERVER_PORT:-8080}"

# 编译和启动
echo "🔨 Building server..."
go build -o gameserver ./cmd/server

if [ $? -eq 0 ]; then
    echo "✅ Build successful"
    echo "🎮 Starting GameServer..."
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    ./gameserver
else
    echo "❌ Build failed"
    exit 1
fi