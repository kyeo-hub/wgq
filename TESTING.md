# wgq-bot 测试指南

## 测试流程

### 步骤 1: 验证配置

```bash
go run ./cmd/checkconfig/main.go
```

如果看到 `✅ 配置验证完成`，说明配置正确。

**常见错误：**

```
❌ 加解密模块初始化失败：illegal base64 data at input byte 40
```

这表示 EncodingAESKey 格式不正确。企业微信的 EncodingAESKey 应该是：
- 43 个字符的 base64 编码字符串
- 解码后为 32 字节

**解决方法：**
1. 登录企业微信管理后台
2. 进入应用 → 智能机器人 → 回调配置
3. 重新复制 EncodingAESKey（注意不要复制多余空格）

---

### 步骤 2: 启动服务器

```bash
# 方式 1: 直接运行
go run ./cmd/main.go -config config.json

# 方式 2: 后台运行
nohup go run ./cmd/main.go -config config.json > wgq-bot.log 2>&1 &
```

启动后应该看到：
```
Configuration loaded successfully
Starting wgq-bot server...
Listening on :8080
Callback path: /wechat/callback
```

---

### 步骤 3: 本地测试（不通过企业微信）

```bash
go run ./cmd/testclient/main.go \
  -aeskey "YOUR_AES_KEY" \
  -token "YOUR_TOKEN" \
  -url "http://localhost:8080/wechat/callback" \
  -user "testuser" \
  -msg "帮我写一个 Hello World"
```

如果看到 `✅ 消息已成功发送到回调服务器`，说明本地通信正常。

---

### 步骤 4: 企业微信真实测试

1. **配置回调 URL**
   - 登录企业微信管理后台
   - 进入应用 → 智能机器人 → 回调配置
   - 填写 URL: `http://你的服务器IP:8080/wechat/callback`
   - 填写 Token（与 config.json 一致）
   - 填写 EncodingAESKey（与 config.json 一致）
   - 点击「保存」

2. **验证回调**
   - 保存后企业微信会自动发送验证请求
   - 如果验证失败，检查服务器日志

3. **发送消息测试**
   - 在企业微信中找到机器人
   - 发送消息：`/help`
   - 应该收到帮助信息回复

---

## 使用内网穿透（本地开发）

如果你的服务器不在公网，可以使用内网穿透工具：

### ngrok

```bash
# 安装 ngrok
npm install -g ngrok

# 启动穿透
ngrok http 8080

# 复制生成的公网地址（如：https://abc123.ngrok.io）
# 在企业微信后台配置回调 URL:
# https://abc123.ngrok.io/wechat/callback
```

### frp

```ini
# frpc.ini
[common]
server_addr = your-frp-server.com
server_port = 7000

[wechat]
type = http
local_port = 8080
custom_domains = your-domain.com
```

---

## 故障排查

### 1. 签名验证失败

```
HTTP 403: Invalid signature
```

**原因：** Token 不一致

**解决：**
- 检查 config.json 中的 token 是否与企业微信后台一致
- 注意大小写，不要有多余空格

### 2. 解密失败

```
HTTP 500: Decrypt failed
```

**原因：** EncodingAESKey 不正确

**解决：**
- 重新从企业微信后台复制 EncodingAESKey
- 确认是完整的 43 字符字符串

### 3. qwen 未找到

```
Warning: qwen is not installed or not in PATH
```

**解决：**
```bash
npm install -g @qwen-code/qwen-code@latest

# 验证安装
qwen --version
```

### 4. 回调 URL 无法访问

**检查清单：**
- [ ] 服务器是否启动（`ps aux | grep wgq`）
- [ ] 端口是否监听（`netstat -tlnp | grep 8080`）
- [ ] 防火墙是否开放 8080 端口
- [ ] 回调 URL 是否配置正确（包含 `/wechat/callback`）

---

## 日志查看

```bash
# 实时查看日志
tail -f wgq-bot.log

# 查看最近的错误
grep ERROR wgq-bot.log
```

---

## 完整测试示例

```bash
# 1. 验证配置
go run ./cmd/checkconfig/main.go

# 2. 启动服务器（终端 1）
go run ./cmd/main.go -config config.json

# 3. 发送测试消息（终端 2）
go run ./cmd/testclient/main.go \
  -aeskey "$(jq -r '.wechat.encoding_aes_key' config.json)" \
  -token "$(jq -r '.wechat.token' config.json)" \
  -msg "/help"

# 4. 查看服务器日志，应该看到：
# Received message: msgid=test_xxx, from=testuser, type=text
# Processing message from user: testuser
# Reply sent successfully
```

---

## 性能测试

```bash
# 发送多条消息测试并发
for i in {1..10}; do
  go run ./cmd/testclient/main.go \
    -aeskey "YOUR_AES_KEY" \
    -token "YOUR_TOKEN" \
    -msg "测试消息 $i" &
done
wait
```
