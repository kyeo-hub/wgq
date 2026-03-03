#!/bin/bash

# 手动发布到 Gitee Release
# 使用方法：./scripts/release-to-gitee.sh v1.0.0

set -e

VERSION="${1:-v1.0.0}"
GITEE_OWNER="kyeo"
GITEE_REPO="wgq"
ARTIFACT="wgq-linux-amd64.tar.gz"

echo "========================================"
echo "  发布到 Gitee Release"
echo "========================================"
echo ""
echo "版本：$VERSION"
echo "仓库：$GITEE_OWNER/$GITEE_REPO"
echo "文件：$ARTIFACT"
echo ""

# 检查 GITEE_TOKEN
if [ -z "$GITEE_TOKEN" ]; then
    echo "❌ 错误：GITEE_TOKEN 环境变量未设置"
    echo ""
    echo "请在 Gitee 生成 API Token:"
    echo "1. 访问：https://gitee.com/profile/personal_access_tokens"
    echo "2. 生成新令牌（只需勾选 projects 权限）"
    echo "3. 设置环境变量：export GITEE_TOKEN=your_token"
    echo ""
    echo "注意：projects 权限已包含 Release 的读写权限"
    exit 1
fi

# 检查构建文件
if [ ! -f "$ARTIFACT" ]; then
    echo "❌ 构建文件不存在：$ARTIFACT"
    echo ""
    echo "正在构建..."
    CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o wgq ./cmd/main.go
    tar -czvf "$ARTIFACT" wgq config.example.json README.md
fi

# 创建 Release
echo "📦 创建/更新 Gitee Release..."

# 检查 release 是否已存在
EXISTING_RELEASE=$(curl -s -H "Authorization: Bearer $GITEE_TOKEN" \
    "https://gitee.com/api/v5/repos/$GITEE_OWNER/$GITEE_REPO/releases/tags/$VERSION" \
    2>/dev/null | jq -r '.tag_name' 2>/dev/null || echo "")

if [ "$EXISTING_RELEASE" = "$VERSION" ]; then
    echo "⚠️  Release $VERSION 已存在，正在更新..."
    
    # 获取 release ID
    RELEASE_ID=$(curl -s -H "Authorization: Bearer $GITEE_TOKEN" \
        "https://gitee.com/api/v5/repos/$GITEE_OWNER/$GITEE_REPO/releases/tags/$VERSION" \
        | jq -r '.id')
    
    # 删除旧的 release asset
    echo "🗑️  删除旧的 assets..."
    curl -s -X DELETE -H "Authorization: Bearer $GITEE_TOKEN" \
        "https://gitee.com/api/v5/repos/$GITEE_OWNER/$GITEE_REPO/releases/$RELEASE_ID/attach_files" \
        2>/dev/null || true
    
    # 更新 release
    RESPONSE=$(curl -s -X PUT -H "Authorization: Bearer $GITEE_TOKEN" \
        -H "Content-Type: application/json" \
        "https://gitee.com/api/v5/repos/$GITEE_OWNER/$GITEE_REPO/releases/$RELEASE_ID" \
        -d "{
            \"access_token\": \"$GITEE_TOKEN\",
            \"tag_name\": \"$VERSION\",
            \"name\": \"$VERSION\",
            \"body\": \"wgq-bot v$VERSION - 企业微信智能机器人\n\n## 安装说明\n\n### 中国大陆（推荐）\n\`\`\`bash\nbash <(curl -fsSL https://gitee.com/kyeo/wgq/raw/main/install-remote.sh) <TOKEN> <AES_KEY> 8888\n\`\`\`\n\n### 全球\n\`\`\`bash\nbash <(curl -fsSL https://raw.githubusercontent.com/kyeo-hub/wgq/main/install-remote.sh) <TOKEN> <AES_KEY> 8888\n\`\`\`\"\n        }")
else
    echo "✨ 创建新的 Release..."
    
    # 创建新 release
    RESPONSE=$(curl -s -X POST -H "Authorization: Bearer $GITEE_TOKEN" \
        -H "Content-Type: application/json" \
        "https://gitee.com/api/v5/repos/$GITEE_OWNER/$GITEE_REPO/releases" \
        -d "{
            \"access_token\": \"$GITEE_TOKEN\",
            \"tag_name\": \"$VERSION\",
            \"name\": \"$VERSION\",
            \"body\": \"wgq-bot v$VERSION - 企业微信智能机器人\n\n## 安装说明\n\n### 中国大陆（推荐）\n\`\`\`bash\nbash <(curl -fsSL https://gitee.com/kyeo/wgq/raw/main/install-remote.sh) <TOKEN> <AES_KEY> 8888\n\`\`\`\n\n### 全球\n\`\`\`bash\nbash <(curl -fsSL https://raw.githubusercontent.com/kyeo-hub/wgq/main/install-remote.sh) <TOKEN> <AES_KEY> 8888\n\`\`\`\",
            \"prerelease\": false,
            \"target_commitish\": \"main\"
        }")
fi

RELEASE_ID=$(echo "$RESPONSE" | jq -r '.id')

if [ -z "$RELEASE_ID" ] || [ "$RELEASE_ID" = "null" ]; then
    echo "❌ 创建/更新 Release 失败"
    echo "响应：$RESPONSE"
    exit 1
fi

echo "✅ Release 创建/更新成功 (ID: $RELEASE_ID)"

# 上传文件
echo "📤 上传构建文件..."
UPLOAD_RESPONSE=$(curl -s -X POST -H "Authorization: Bearer $GITEE_TOKEN" \
    -F "file=@$ARTIFACT" \
    "https://gitee.com/api/v5/repos/$GITEE_OWNER/$GITEE_REPO/releases/$RELEASE_ID/attach_files")

ATTACH_ID=$(echo "$UPLOAD_RESPONSE" | jq -r '.id')

if [ -z "$ATTACH_ID" ] || [ "$ATTACH_ID" = "null" ]; then
    echo "⚠️  文件上传可能失败，请检查"
    echo "响应：$UPLOAD_RESPONSE"
else
    echo "✅ 文件上传成功 (Attach ID: $ATTACH_ID)"
fi

echo ""
echo "========================================"
echo "  ✅ 发布完成!"
echo "========================================"
echo ""
echo "Gitee Release: https://gitee.com/$GITEE_OWNER/$GITEE_REPO/releases/$VERSION"
echo ""
echo "测试下载:"
echo "  curl -LO https://gitee.com/$GITEE_OWNER/$GITEE_REPO/releases/download/$VERSION/$ARTIFACT"
echo ""
