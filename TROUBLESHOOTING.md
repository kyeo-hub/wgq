# 企业微信回调 "Invalid signature" 问题诊断

## 问题现象

```
请求：http://120.48.19.58:8080/wechat/callback?msg_signature=f939272e1f375ba90ff8e8e218720bc2224057e0&timestamp=1772504484&nonce=6zwnkbkpgai&echostr=...
返回：Invalid signature
```

## 原因分析

**签名验证失败** = 服务器配置的 Token 与企业微信后台不一致

## 验证步骤

### 1. 检查企业微信后台 Token

登录 [企业微信管理后台](https://work.weixin.qq.com/) → 应用管理 → 自建应用 → 接收消息设置

记录 Token 值，例如：`2dpvh5TUIFM8l5Kaq60GtclcRflLAy`

### 2. 检查服务器 config.json

登录服务器查看配置：

```bash
# SSH 登录服务器
ssh root@120.48.19.58

# 查看当前 token 配置
cat /opt/wgq-bot/config.json | jq -r '.wechat.token'
```

### 3. 对比两个 Token 是否一致

- 企业微信后台：`________________`
- 服务器配置：`________________`

**如果不一致**，需要更新服务器配置。

## 解决方案

### 方案 A：更新服务器配置

```bash
# 1. SSH 登录服务器
ssh root@120.48.19.58

# 2. 停止服务
systemctl stop wgq-bot

# 3. 编辑配置文件
vim /opt/wgq-bot/config.json

# 修改 wechat.token 为企业微信后台的 Token
# 保存退出

# 4. 重启服务
systemctl start wgq-bot

# 5. 验证
systemctl status wgq-bot
journalctl -u wgq-bot -f
```

### 方案 B：重新运行一键安装脚本

```bash
# 在服务器上运行（使用正确的 Token）
bash <(curl -fsSL https://gitee.com/kyeo/wgq/raw/main/install-remote.sh) \
  <企业微信后台的_TOKEN> \
  <企业微信后台的_AES_KEY> \
  8080
```

## 测试验证

更新配置后，在企业微信后台点击「保存」触发回调测试，应该显示：

```
✅ 验证成功
```

## 常见问题

### Q: Token 复制时多了空格？
A: 确保 Token 前后没有空格或换行

### Q: 修改配置后还是失败？
A: 确认服务已重启，配置已生效

### Q: 如何查看服务器日志？
A: `journalctl -u wgq-bot -f --no-pager`
