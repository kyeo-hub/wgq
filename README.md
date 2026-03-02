# wgq-bot - 企业微信智能机器人

基于 Go 语言实现的企业微信智能机器人，用于控制服务器上的 qwen CLI，实现类似 OpenCLI 的通讯工具控制服务器体验。

## 功能特性

- ✅ 企业微信消息加解密（AES-256-CBC）
- ✅ 支持文本消息和图文混排消息
- ✅ qwen-code CLI 自动化执行
- ✅ 用户白名单权限控制
- ✅ 命令超时和输出限制
- ✅ Docker 容器化部署支持

## 架构图

```
┌─────────────┐     ┌──────────────────┐     ┌─────────────────┐
│ 企业微信用户 │ ──→ │ 企业微信服务器   │ ──→ │ wgq-bot 服务器   │
│ (发送命令)   │     │ (回调推送)       │     │ (Go 服务)        │
└─────────────┘     └──────────────────┘     └────────┬────────┘
                                                      │
                                                      ↓
                                             ┌─────────────────┐
                                             │ qwen CLI        │
                                             │ (AI 编码助手)    │
                                             └─────────────────┘
```

## 快速开始

### 1. 企业微信配置

1. 登录 [企业微信管理后台](https://work.weixin.qq.com/)
2. 进入「应用管理」→「自建应用」
3. 创建「智能机器人」应用
4. 获取以下配置信息：
   - **Token**: 在回调配置页面获取
   - **EncodingAESKey**: 在回调配置页面获取
5. 配置回调 URL：`http://你的服务器 IP:8080/wechat/callback`

### 2. 本地运行

#### 前置要求

- Go 1.21+
- Node.js 18+
- qwen CLI

#### 安装 qwen

```bash
npm install -g @qwen-code/qwen-code@latest
```

#### 编译运行

```bash
# 下载依赖
go mod tidy

# 复制配置文件
cp config.example.json config.json

# 编辑配置文件，填入企业微信的 Token 和 EncodingAESKey
vim config.json

# 运行
go run ./cmd/main.go -config config.json
```

### 3. Docker 部署

```bash
# 复制配置文件
cp config.example.json config.json

# 编辑配置
vim config.json

# 启动容器
docker-compose up -d

# 查看日志
docker-compose logs -f
```

### 4. 中国大陆一键安装（推荐）

```bash
# 使用 Gitee 镜像源（速度更快）
bash <(curl -fsSL https://gitee.com/kyeo/wgq/raw/main/install-remote.sh) \
  <YOUR_TOKEN> \
  <YOUR_AES_KEY> \
  8888

# 或使用 GitHub + ghproxy 加速
bash <(curl -fsSL https://raw.githubusercontent.com/kyeo-hub/wgq/main/install-remote.sh) \
  <YOUR_TOKEN> \
  <YOUR_AES_KEY> \
  8888
```

## 配置说明

### config.json

```json
{
  "wechat": {
    "encoding_aes_key": "企业微信后台的 EncodingAESKey",
    "token": "企业微信后台的 Token",
    "bot_id": ""
  },
  "server": {
    "addr": ":8080",
    "callback_path": "/wechat/callback"
  },
  "qwen": {
    "work_dir": "/tmp/qwen-workspace",
    "timeout_seconds": 300,
    "max_output_lines": 500
  },
  "allowed_users": ["user1", "user2"]
}
```

### 配置项说明

| 字段 | 说明 | 必填 |
|------|------|------|
| `wechat.encoding_aes_key` | 企业微信后台的 EncodingAESKey | 是 |
| `wechat.token` | 企业微信后台的 Token | 是 |
| `server.addr` | 服务监听地址 | 是 |
| `qwen.work_dir` | qwen 工作目录 | 是 |
| `qwen.timeout_seconds` | 命令执行超时时间（秒） | 否 |
| `allowed_users` | 允许使用的用户 ID 列表（空表示允许所有） | 否 |

## 使用方式

### 发送消息

在企业微信中向机器人发送消息即可：

```
帮我写一个快速排序函数
```

### 内置命令

| 命令 | 说明 |
|------|------|
| `/help` | 显示帮助信息 |
| `/status` | 查看系统状态 |
| `/version` | 查看 qwen-code 版本 |
| `/cancel` | 取消当前任务 |

### 示例对话

**用户**: 帮我写一个 Python 的快速排序函数

**机器人**: 
```
✅ 执行完成
⏱️ 耗时：3.45 秒

📋 输出:
def quick_sort(arr):
    if len(arr) <= 1:
        return arr
    pivot = arr[len(arr) // 2]
    left = [x for x in arr if x < pivot]
    middle = [x for x in arr if x == pivot]
    right = [x for x in arr if x > pivot]
    return quick_sort(left) + middle + quick_sort(right)
```

## 项目结构

```
wgq/
├── cmd/
│   └── main.go              # 程序入口
├── internal/
│   ├── config/
│   │   └── config.go        # 配置管理
│   ├── crypto/
│   │   └── crypto.go        # 加解密模块
│   ├── handler/
│   │   └── message.go       # 消息处理
│   ├── qwen/
│   │   └── executor.go      # qwen 执行器
│   ├── server/
│   │   └── callback.go      # 回调服务器
│   └── wechat/
│       └── message.go       # 消息数据结构
├── config.example.json      # 配置示例
├── Dockerfile               # Docker 构建文件
├── docker-compose.yml       # Docker Compose 配置
└── README.md                # 本文档
```

## 安全注意事项

1. **不要将 config.json 提交到代码仓库**（已加入 .gitignore）
2. **使用用户白名单**限制可访问的用户
3. **配置防火墙**仅允许企业微信服务器访问回调端口
4. **使用 HTTPS**（生产环境建议配合 Nginx 反向代理）

## 故障排查

### 日志查看

```bash
# 本地运行
go run ./cmd/main.go -config config.json

# Docker
docker-compose logs -f
```

### 常见问题

1. **签名验证失败**
   - 检查 Token 是否配置正确
   - 检查服务器时间是否同步

2. **解密失败**
   - 检查 EncodingAESKey 是否正确
   - 确认是完整的 base64 字符串

3. **qwen 未找到**
   - 确认已安装：`npm install -g @qwen-code/qwen-code@latest`
   - 检查 PATH 环境变量

## 开发计划

- [ ] 支持流式消息输出
- [ ] 支持图片和文件消息
- [ ] 会话上下文支持
- [ ] 模板卡片消息回复
- [ ] 多租户支持

## License

MIT License
