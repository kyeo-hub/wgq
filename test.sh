#!/bin/bash

# 测试脚本：启动服务器并发送测试消息

set -e

echo "========================================"
echo "  wgq-bot 测试脚本"
echo "========================================"
echo ""

# 检查配置文件
if [ ! -f "config.json" ]; then
    echo "❌ 配置文件 config.json 不存在"
    echo "请先复制并编辑配置文件："
    echo "  cp config.example.json config.json"
    exit 1
fi

# 检查 qwen 是否安装
if ! command -v qwen &> /dev/null; then
    echo "⚠️  警告：qwen 未安装，请运行：npm install -g @qwen-code/qwen-code@latest"
fi

# 启动服务器（后台运行）
echo "🚀 启动服务器..."
go run ./cmd/main.go -config config.json &
SERVER_PID=$!
echo "服务器进程 ID: $SERVER_PID"

# 等待服务器启动
echo "等待服务器启动..."
sleep 2

# 检查服务器是否启动
if ! kill -0 $SERVER_PID 2>/dev/null; then
    echo "❌ 服务器启动失败"
    exit 1
fi

echo "✅ 服务器已启动"
echo ""

# 读取配置
AES_KEY=$(jq -r '.wechat.encoding_aes_key' config.json)
TOKEN=$(jq -r '.wechat.token' config.json)

# 发送测试消息
echo "📨 发送测试消息..."
go run ./cmd/testclient/main.go \
    -aeskey "$AES_KEY" \
    -token "$TOKEN" \
    -url "http://localhost:8080/wechat/callback" \
    -user "testuser" \
    -msg "帮我写一个快速排序函数"

echo ""
echo "========================================"
echo "测试完成"
echo "========================================"
echo ""
echo "停止服务器..."
kill $SERVER_PID 2>/dev/null || true

echo "✅ 测试结束"
