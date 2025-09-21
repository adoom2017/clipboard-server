#!/bin/bash

# Docker 容器启动脚本
set -e

echo "Starting Clipboard Sync Server..."
echo "Working directory: $(pwd)"
echo "User: $(whoami)"
echo "Groups: $(groups)"

# 创建必要的目录
echo "Creating necessary directories..."
mkdir -p data logs uploads

# 检查目录权限
echo "Checking directory permissions..."
ls -la data logs uploads || echo "Some directories may not exist yet"

# 设置正确的权限
echo "Setting directory permissions..."
chmod -R 755 data logs uploads || echo "Permission setting completed"

# 检查环境变量
echo "Environment variables:"
echo "  GO_ENV: ${GO_ENV:-development}"
echo "  DB_PATH: ${DB_PATH:-data/clipboard.db}"
echo "  SERVER_HOST: ${SERVER_HOST:-localhost}"
echo "  SERVER_PORT: ${SERVER_PORT:-8080}"

# 启动应用
echo "Starting application..."
exec "$@"