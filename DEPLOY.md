# wgq-bot 部署说明

## 快速部署

### 1. 上传文件到服务器

将 `wgq-linux-amd64.tar.gz` 上传到服务器：

```bash
# 使用 scp
scp wgq-linux-amd64.tar.gz user@your-server:/opt/wgq-bot/

# 或使用 rz 命令
rz
```

### 2. 解压并配置

```bash
cd /opt/wgq-bot
tar -xzf wgq-linux-amd64.tar.gz

# 复制配置文件
cp config.example.json config.json

# 编辑配置（填入你的企业微信信息）
vim config.json
```

### 3. 安装 qwen

```bash
# 安装 Node.js (如果未安装)
curl -fsSL https://deb.nodesource.com/setup_18.x | bash -
apt-get install -y nodejs

# 安装 qwen
npm install -g @qwen-code/qwen-code@latest

# 验证安装
qwen --version
```

### 4. 启动服务

```bash
# 直接运行
./wgq -config config.json

# 或使用 systemd 后台运行（见下方）
```

### 5. 配置 systemd 服务（推荐）

创建服务文件：

```bash
cat > /etc/systemd/system/wgq-bot.service << 'EOF'
[Unit]
Description=wgq-bot 企业微信智能机器人
After=network.target

[Service]
Type=simple
User=root
WorkingDirectory=/opt/wgq-bot
ExecStart=/opt/wgq-bot/wgq -config /opt/wgq-bot/config.json
Restart=always
RestartSec=10

[Install]
WantedBy=multi-user.target
EOF

# 启动服务
systemctl daemon-reload
systemctl start wgq-bot
systemctl enable wgq-bot

# 查看状态
systemctl status wgq-bot

# 查看日志
journalctl -u wgq-bot -f
```

### 6. 配置防火墙

```bash
# 开放 8080 端口
ufw allow 8080/tcp

# 或使用 firewalld
firewall-cmd --permanent --add-port=8080/tcp
firewall-cmd --reload
```

### 7. 配置企业微信

在企业微信管理后台配置：
- **URL**: `http://你的服务器IP:8080/wechat/callback`
- **Token**: (config.json 中的 token)
- **EncodingAESKey**: (config.json 中的 encoding_aes_key)

---

## 使用 Nginx 反向代理（可选，推荐 HTTPS）

### 安装 Nginx

```bash
apt-get install -y nginx
```

### 配置 Nginx

```bash
cat > /etc/nginx/sites-available/wgq-bot << 'EOF'
server {
    listen 80;
    server_name your-domain.com;

    location /wechat/callback {
        proxy_pass http://127.0.0.1:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
}
EOF

# 启用配置
ln -s /etc/nginx/sites-available/wgq-bot /etc/nginx/sites-enabled/
nginx -t
systemctl restart nginx
```

### 配置 HTTPS（使用 Let's Encrypt）

```bash
apt-get install -y certbot python3-certbot-nginx
certbot --nginx -d your-domain.com
```

---

## 故障排查

### 查看日志

```bash
# systemd 日志
journalctl -u wgq-bot -f

# 或直接运行查看输出
./wgq -config config.json
```

### 检查端口

```bash
netstat -tlnp | grep 8080
```

### 测试回调

```bash
# 本地测试
curl http://localhost:8080/health

# 应该返回：OK
```

### 常见问题

**1. 权限问题**
```bash
chmod +x wgq
```

**2. 端口被占用**
```bash
# 修改 config.json 中的 server.addr
# 例如：":8081"
```

**3. qwen 未找到**
```bash
# 确认安装
which qwen

# 如果未找到，检查 npm 全局路径
npm config get prefix
```
