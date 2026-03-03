#!/bin/bash

# wgq-bot 一键安装脚本
# 用法：curl -fsSL https://your-server/wgq-linux-amd64.tar.gz | bash -s -- <TOKEN> <AES_KEY>

set -e

echo "========================================"
echo "  wgq-bot 一键安装脚本"
echo "========================================"
echo ""

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# 日志函数
log_info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

log_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# 检查是否 root
if [ "$EUID" -ne 0 ]; then
    log_error "请使用 root 用户运行此脚本 (sudo -i)"
    exit 1
fi

# 安装目录
INSTALL_DIR="/opt/wgq-bot"
CONFIG_FILE="$INSTALL_DIR/config.json"
SERVICE_FILE="/etc/systemd/system/wgq-bot.service"

# 检查参数
if [ $# -lt 2 ]; then
    echo "用法：bash install.sh <TOKEN> <AES_KEY> [PORT]"
    echo ""
    echo "参数说明:"
    echo "  TOKEN     - 企业微信 Token"
    echo "  AES_KEY   - 企业微信 EncodingAESKey"
    echo "  PORT      - 服务端口 (可选，默认 8080)"
    echo ""
    echo "示例:"
    echo "  bash install.sh 2dpvh5TUIFM8l5Kaq60GtclcRflLAy DIE4GdISYuzuC3yYIYWVD9u3gSHoja4fCYWdKL6iz4X 8080"
    exit 1
fi

TOKEN="$1"
AES_KEY="$2"
PORT="${3:-8888}"

log_info "配置信息:"
echo "  安装目录：$INSTALL_DIR"
echo "  服务端口：$PORT"
echo ""

# 检查端口是否被占用
check_port() {
    if command -v ss &> /dev/null; then
        ss -tlnp | grep -q ":$PORT "
    elif command -v netstat &> /dev/null; then
        netstat -tlnp | grep -q ":$PORT "
    else
        return 1
    fi
}

if check_port; then
    log_error "端口 $PORT 已被占用!"
    log_info "查看占用进程：lsof -i :$PORT"
    log_info "解决方案:"
    echo "  1. 停止占用端口的进程"
    echo "  2. 或使用其他端口运行本脚本:"
    echo "     ./install.sh $TOKEN $AES_KEY 9000"
    exit 1
fi
log_info "端口 $PORT 可用"

# 1. 创建安装目录
log_info "创建安装目录..."
mkdir -p "$INSTALL_DIR"
cd "$INSTALL_DIR"

# 2. 下载 wgq-bot
log_info "下载 wgq-bot..."
DOWNLOAD_URLS=(
    # Gitee 镜像（中国大陆推荐）
    "https://gitee.com/kyeo/wgq/releases/latest/download/wgq-linux-amd64.tar.gz"
    # Gitee raw 直链（备用）
    "https://gitee.com/kyeo/wgq/raw/main/wgq-linux-amd64.tar.gz"
    # GitHub CDN
    "https://ghproxy.com/https://github.com/kyeo-hub/wgq/releases/latest/download/wgq-linux-amd64.tar.gz"
    # GitHub 原始地址
    "https://github.com/kyeo-hub/wgq/releases/latest/download/wgq-linux-amd64.tar.gz"
    # raw.githubusercontent.com
    "https://raw.githubusercontent.com/kyeo-hub/wgq/main/wgq-linux-amd64.tar.gz"
)

DOWNLOADED=false
for url in "${DOWNLOAD_URLS[@]}"; do
    log_info "尝试从 $url 下载..."
    if curl -fsSL --max-time 60 "$url" -o wgq.tar.gz 2>/dev/null; then
        if tar -tzf wgq.tar.gz &>/dev/null; then
            log_info "从 $url 下载成功"
            DOWNLOADED=true
            break
        else
            rm -f wgq.tar.gz
        fi
    fi
done

if [ "$DOWNLOADED" = true ]; then
    tar -xzf wgq.tar.gz
    rm -f wgq.tar.gz
    chmod +x wgq
else
    log_error "下载失败！"
    log_info "请手动下载并上传 wgq-linux-amd64.tar.gz 到 $INSTALL_DIR"
    log_info "Gitee: https://gitee.com/kyeo/wgq/releases"
    log_info "GitHub: https://github.com/kyeo-hub/wgq/releases"
    exit 1
fi

# 3. 解压
log_info "解压文件..."
tar -xzf wgq.tar.gz
rm -f wgq.tar.gz
chmod +x wgq

# 4. 生成配置文件
log_info "生成配置文件..."
cat > "$CONFIG_FILE" << EOF
{
  "wechat": {
    "encoding_aes_key": "$AES_KEY",
    "token": "$TOKEN",
    "bot_id": ""
  },
  "server": {
    "addr": ":$PORT",
    "callback_path": "/wechat/callback"
  },
  "qwen": {
    "work_dir": "/tmp/qwen-workspace",
    "timeout_seconds": 300,
    "max_output_lines": 500
  },
  "allowed_users": []
}
EOF

log_info "配置文件已生成：$CONFIG_FILE"

# 5. 检查并安装 Node.js
log_info "检查 Node.js..."
if ! command -v node &> /dev/null; then
    log_info "Node.js 未安装，开始安装..."
    
    # 检测系统类型
    if [ -f /etc/debian_version ]; then
        # Debian/Ubuntu - 安装 Node.js 20 LTS
        curl -fsSL https://deb.nodesource.com/setup_20.x | bash -
        apt-get install -y nodejs
    elif [ -f /etc/redhat-release ]; then
        # CentOS/RHEL - 安装 Node.js 20 LTS
        curl -fsSL https://rpm.nodesource.com/setup_20.x | bash -
        yum install -y nodejs
    else
        log_error "不支持的系统类型，请手动安装 Node.js (建议 v20+)"
        exit 1
    fi
else
    NODE_VERSION=$(node --version)
    log_info "Node.js 已安装：$NODE_VERSION"
    # 检查版本是否 >= 18
    if [ "$(node --version | cut -d'.' -f1 | tr -d 'v')" -lt 18 ]; then
        log_warn "Node.js 版本过低，建议升级到 v20 LTS"
    fi
fi

# 6. 安装 qwen
log_info "检查 qwen..."
if ! command -v qwen &> /dev/null; then
    log_info "安装 qwen..."
    npm install -g @qwen-code/qwen-code@latest
else
    log_info "qwen 已安装：$(qwen --version 2>/dev/null || echo '未知版本')"
fi

# 7. 配置 systemd 服务
log_info "配置 systemd 服务..."
cat > "$SERVICE_FILE" << EOF
[Unit]
Description=wgq-bot 企业微信智能机器人
After=network.target

[Service]
Type=simple
User=root
WorkingDirectory=$INSTALL_DIR
ExecStart=$INSTALL_DIR/wgq -config $CONFIG_FILE
Restart=always
RestartSec=10
LimitNOFILE=65535

[Install]
WantedBy=multi-user.target
EOF

# 重载 systemd
systemctl daemon-reload
systemctl enable wgq-bot

# 8. 配置防火墙
log_info "配置防火墙..."
if command -v ufw &> /dev/null && ufw status | grep -q "Status: active"; then
    ufw allow "$PORT"/tcp
    log_info "已开放端口：$PORT (ufw)"
elif command -v firewall-cmd &> /dev/null && systemctl is-active --quiet firewalld; then
    firewall-cmd --permanent --add-port="$PORT"/tcp
    firewall-cmd --reload
    log_info "已开放端口：$PORT (firewalld)"
else
    log_warn "未检测到活动的防火墙，请手动配置"
fi

# 9. 启动服务
log_info "启动服务..."
systemctl start wgq-bot

# 等待服务启动
sleep 3

# 10. 检查服务状态
log_info "检查服务状态..."
if systemctl is-active --quiet wgq-bot; then
    log_info "✅ wgq-bot 服务已启动"
else
    log_error "❌ wgq-bot 服务启动失败"
    log_info "查看日志：journalctl -u wgq-bot -f"
    exit 1
fi

# 11. 测试健康检查
log_info "测试健康检查..."
if curl -s "http://localhost:$PORT/health" | grep -q "OK"; then
    log_info "✅ 健康检查通过"
else
    log_warn "⚠️  健康检查失败，请查看日志"
fi

# 完成
echo ""
echo "========================================"
echo -e "  ${GREEN}✅ 安装完成！${NC}"
echo "========================================"
echo ""
echo "📋 服务信息:"
echo "  服务状态：active"
echo "  监听端口：$PORT"
echo "  回调 URL: http://<你的服务器 IP>:$PORT/wechat/callback"
echo ""
echo "📋 管理命令:"
echo "  查看状态：systemctl status wgq-bot"
echo "  查看日志：journalctl -u wgq-bot -f"
echo "  重启服务：systemctl restart wgq-bot"
echo "  停止服务：systemctl stop wgq-bot"
echo ""
echo "📋 下一步:"
echo "  1. 在企业微信管理后台配置回调 URL:"
echo "     http://<你的服务器 IP>:$PORT/wechat/callback"
echo ""
echo "  2. 配置 Token: $TOKEN"
echo "     EncodingAESKey: $AES_KEY"
echo ""
echo "  3. 在企业微信中向机器人发送 /help 测试"
echo ""
echo "========================================"
