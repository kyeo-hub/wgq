#!/bin/bash

# wgq-bot 快速安装脚本
# 用法：bash <(curl -fsSL https://raw.githubusercontent.com/kyeo-hub/wgq/main/install-remote.sh) <TOKEN> <AES_KEY> [PORT]

set -e

echo "========================================"
echo "  wgq-bot 一键安装脚本"
echo "========================================"
echo ""

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

log_info() { echo -e "${GREEN}[INFO]${NC} $1"; }
log_warn() { echo -e "${YELLOW}[WARN]${NC} $1"; }
log_error() { echo -e "${RED}[ERROR]${NC} $1"; }

# 检查是否 root
if [ "$EUID" -ne 0 ]; then 
    log_error "请使用 root 用户运行 (sudo -i 或 su -)"
    exit 1
fi

# 检查参数
if [ $# -lt 2 ]; then
    echo "用法：bash <(curl -fsSL https://raw.githubusercontent.com/kyeo-hub/wgq/main/install-remote.sh) <TOKEN> <AES_KEY> [PORT]"
    echo ""
    echo "参数:"
    echo "  TOKEN     - 企业微信 Token"
    echo "  AES_KEY   - 企业微信 EncodingAESKey"  
    echo "  PORT      - 服务端口 (可选，默认 8080)"
    echo ""
    echo "示例:"
    echo "  bash <(curl -fsSL ...) 2dpvh5TUIFM8l5Kaq60GtclcRflLAy DIE4GdISYuzuC3yYIYWVD9u3gSHoja4fCYWdKL6iz4X 8080"
    exit 1
fi

TOKEN="$1"
AES_KEY="$2"
PORT="${3:-8080}"

INSTALL_DIR="/opt/wgq-bot"

log_info "开始安装 wgq-bot..."
log_info "安装目录：$INSTALL_DIR"
log_info "服务端口：$PORT"
echo ""

# 创建目录
log_info "创建安装目录..."
mkdir -p "$INSTALL_DIR"
cd "$INSTALL_DIR"

# 下载二进制
log_info "下载 wgq-bot..."

# 尝试多个下载源
DOWNLOAD_URLS=(
    "https://github.com/kyeo-hub/wgq/releases/latest/download/wgq-linux-amd64.tar.gz"
    "https://raw.githubusercontent.com/kyeo-hub/wgq/main/wgq-linux-amd64.tar.gz"
)

DOWNLOADED=false
for url in "${DOWNLOAD_URLS[@]}"; do
    if curl -fsSL --max-time 60 "$url" -o wgq.tar.gz 2>/dev/null; then
        log_info "从 $url 下载成功"
        DOWNLOADED=true
        break
    fi
done

if [ "$DOWNLOADED" = true ]; then
    tar -xzf wgq.tar.gz
    rm wgq.tar.gz
    chmod +x wgq
else
    log_error "所有下载源失败，请手动上传 wgq-linux-amd64.tar.gz 或使用 GitHub Releases"
    log_info "下载地址：https://github.com/kyeo-hub/wgq/releases"
    exit 1
fi

# 生成配置
log_info "生成配置文件..."
cat > config.json << EOF
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

# 安装 Node.js
if ! command -v node &> /dev/null; then
    log_info "安装 Node.js..."
    if [ -f /etc/debian_version ]; then
        # Debian/Ubuntu - 安装 Node.js 20 LTS
        curl -fsSL https://deb.nodesource.com/setup_20.x | bash -
        apt-get install -y nodejs
    elif [ -f /etc/redhat-release ]; then
        # CentOS/RHEL - 安装 Node.js 20 LTS
        curl -fsSL https://rpm.nodesource.com/setup_20.x | bash -
        yum install -y nodejs
    else
        log_error "不支持的系统，请手动安装 Node.js (建议 v20+)"
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

# 安装 qwen
log_info "安装 qwen..."
npm install -g @qwen-code/qwen-code@latest

# 配置 systemd
log_info "配置 systemd 服务..."
cat > /etc/systemd/system/wgq-bot.service << EOF
[Unit]
Description=wgq-bot 企业微信智能机器人
After=network.target

[Service]
Type=simple
User=root
WorkingDirectory=$INSTALL_DIR
ExecStart=$INSTALL_DIR/wgq -config $INSTALL_DIR/config.json
Restart=always
RestartSec=10

[Install]
WantedBy=multi-user.target
EOF

systemctl daemon-reload
systemctl enable wgq-bot

# 防火墙
if command -v ufw &> /dev/null && ufw status | grep -q "active"; then
    ufw allow "$PORT"/tcp 2>/dev/null || true
fi

# 启动
log_info "启动服务..."
systemctl start wgq-bot
sleep 2

# 验证
if systemctl is-active --quiet wgq-bot; then
    echo ""
    echo "========================================"
    log_info "✅ 安装成功!"
    echo "========================================"
    echo ""
    echo "回调 URL: http://<服务器 IP>:$PORT/wechat/callback"
    echo ""
    echo "管理命令:"
    echo "  systemctl status wgq-bot   # 查看状态"
    echo "  journalctl -u wgq-bot -f   # 查看日志"
    echo ""
else
    log_error "服务启动失败，查看日志：journalctl -u wgq-bot -f"
    exit 1
fi
