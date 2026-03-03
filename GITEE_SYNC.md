# Gitee 同步与发布指南

本文档说明如何将 wgq-bot 同步到 Gitee 并在中国大陆地区使用一键安装脚本。

## 仓库地址

| 平台 | 地址 |
|------|------|
| **GitHub (主仓库)** | https://github.com/kyeo-hub/wgq |
| **Gitee (镜像)** | https://gitee.com/kyeo/wgq |

## 自动同步流程

### 1. GitHub Actions 工作流

项目使用 GitHub Actions 自动同步代码和 Releases 到 Gitee：

```yaml
# .github/workflows/release.yml
```

**触发条件**：
- 推送 tag（如 `v1.0.0`）
- 手动触发 workflow_dispatch

**执行步骤**：
1. **构建**：编译 Linux amd64 二进制文件
2. **GitHub Release**：创建 GitHub Release 并上传 artifacts
3. **Gitee Release**：同步创建 Gitee Release
4. **代码镜像**：同步代码到 Gitee

### 2. 配置 Secrets

在 GitHub 仓库设置中配置以下 Secrets：

1. 进入 `Settings` → `Secrets and variables` → `Actions`
2. 添加以下 repository secrets：

| Secret Name | 说明 | 获取方式 |
|-------------|------|----------|
| `GITEE_PRIVATE_KEY` | Gitee SSH 私钥 | 生成 SSH key 后添加到 Gitee |
| `GITEE_TOKEN` | Gitee API Token | Gitee → 设置 → 个人令牌（勾选 **projects** 权限） |

### 3. 生成 Gitee SSH Key

```bash
# 生成 SSH key
ssh-keygen -t rsa -b 4096 -C "your_email@example.com" -f gitee_key

# 查看公钥
cat gitee_key.pub

# 将公钥添加到 Gitee:
# Gitee → 设置 → SSH 公钥 → 添加公钥
```

### 4. 获取 Gitee Token

1. 登录 Gitee
2. 进入 **设置** → **个人令牌**
3. 点击 **生成新令牌**
4. 勾选权限：**`projects`** （此权限已包含 Release 读写）
5. 复制令牌并保存到 GitHub Secrets

## 发布新版本

### 方法一：推送 Tag（推荐）

```bash
# 本地打 tag
git tag v1.0.0
git push origin v1.0.0

# GitHub Actions 会自动：
# 1. 构建二进制文件
# 2. 创建 GitHub Release
# 3. 创建 Gitee Release
# 4. 同步代码到 Gitee
```

### 方法二：手动触发 Workflow

1. 进入 GitHub Actions → Release workflow
2. 点击 **Run workflow**
3. 选择分支（通常是 main）
4. 点击 **Run workflow**

## 一键安装脚本

### 中国大陆用户（推荐）

```bash
bash <(curl -fsSL https://gitee.com/kyeo/wgq/raw/main/install-remote.sh) \
  <YOUR_TOKEN> \
  <YOUR_AES_KEY> \
  8888
```

### 海外用户

```bash
bash <(curl -fsSL https://raw.githubusercontent.com/kyeo-hub/wgq/main/install-remote.sh) \
  <YOUR_TOKEN> \
  <YOUR_AES_KEY> \
  8888
```

## 下载源说明

安装脚本会从以下源按优先级下载：

1. **Gitee Releases** (中国大陆最快)
   - `https://gitee.com/kyeo/wgq/releases/latest/download/wgq-linux-amd64.tar.gz`

2. **Gitee Raw** (备用)
   - `https://gitee.com/kyeo/wgq/raw/main/wgq-linux-amd64.tar.gz`

3. **ghproxy.com** (GitHub 加速)
   - `https://ghproxy.com/https://github.com/kyeo-hub/wgq/releases/...`

4. **GitHub Releases** (原始地址)
   - `https://github.com/kyeo-hub/wgq/releases/...`

## 验证同步

### 检查 Gitee Release

访问：https://gitee.com/kyeo/wgq/releases

应该看到与 GitHub 相同的 Release 和 `wgq-linux-amd64.tar.gz` 文件。

### 测试安装脚本

```bash
# 在干净的 Linux 环境中测试
docker run --rm -it alpine:latest /bin/sh

# 在容器中安装 curl
apk add --no-cache curl bash

# 运行安装脚本（测试模式）
bash <(curl -fsSL https://gitee.com/kyeo/wgq/raw/main/install-remote.sh) --help
```

## 故障排查

### 1. Gitee Release 未创建

**原因**：Gitee Token 权限不足或过期

**解决**：
1. 重新生成 Gitee Token
2. 确保勾选 `projects` 权限（已包含 Release 读写）
3. 更新 GitHub Secrets

### 2. 代码同步失败

**原因**：SSH Key 配置问题

**解决**：
1. 检查 SSH Key 是否正确添加到 Gitee
2. 验证 GitHub Secret `GITEE_PRIVATE_KEY` 格式正确
3. 手动触发 workflow 查看日志

### 3. 下载速度慢

**解决**：
- 中国大陆用户优先使用 Gitee 源
- 检查网络连接
- 尝试备用下载源

## 手动同步（备用方案）

如果自动同步失败，可以手动操作：

### 同步代码

```bash
# 克隆 Gitee 仓库
git clone git@gitee.com:kyeo/wgq.git
cd wgq

# 添加 GitHub 远程
git remote add github https://github.com/kyeo-hub/wgq.git

# 从 GitHub 拉取
git pull github main

# 推送到 Gitee
git push origin main
```

### 上传 Release

1. 访问 https://gitee.com/kyeo/wgq/releases
2. 点击 **创建发布**
3. 填写版本号和说明
4. 上传 `wgq-linux-amd64.tar.gz`

## 相关文件

| 文件 | 说明 |
|------|------|
| `.github/workflows/release.yml` | GitHub Actions 工作流 |
| `install-remote.sh` | 一键安装脚本 |
| `install.sh` | 本地安装脚本 |
| `gitee_key` | Gitee SSH 私钥（本地使用） |
| `gitee_key.pub` | Gitee SSH 公钥 |
| `MIRROR.md` | 镜像仓库说明 |

## 联系支持

如有问题，请提交 Issue：
- GitHub: https://github.com/kyeo-hub/wgq/issues
- Gitee: https://gitee.com/kyeo/wgq/issues
