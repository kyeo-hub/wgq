package main

import (
	"encoding/base64"
	"fmt"
	"strings"

	"github.com/wgq-bot/wgq/internal/config"
)

func main() {
	fmt.Println("========================================")
	fmt.Println("  EncodingAESKey 诊断工具")
	fmt.Println("========================================")
	fmt.Println()

	cfg, err := config.LoadConfig("config.json")
	if err != nil {
		fmt.Printf("加载配置失败：%v\n", err)
		return
	}

	aesKey := cfg.WeChat.EncodingAESKey

	fmt.Printf("当前 EncodingAESKey:\n")
	fmt.Printf("  原始值：%q\n", aesKey)
	fmt.Printf("  长度：%d 字符\n", len(aesKey))
	fmt.Println()

	// 检查常见问题
	fmt.Println("🔍 检查问题:")

	// 1. 检查是否有空格
	if strings.Contains(aesKey, " ") {
		fmt.Println("  ❌ 包含空格")
	} else {
		fmt.Println("  ✅ 不包含空格")
	}

	// 2. 检查是否有换行
	if strings.Contains(aesKey, "\n") || strings.Contains(aesKey, "\r") {
		fmt.Println("  ❌ 包含换行符")
	} else {
		fmt.Println("  ✅ 不包含换行符")
	}

	// 3. 检查长度
	if len(aesKey) != 43 {
		fmt.Printf("  ❌ 长度应为 43，实际为 %d\n", len(aesKey))
	} else {
		fmt.Println("  ✅ 长度正确 (43)")
	}

	// 4. 尝试 base64 解码
	decoded, err := base64.StdEncoding.DecodeString(aesKey)
	if err != nil {
		fmt.Printf("  ❌ Base64 解码失败：%v\n", err)
	} else {
		fmt.Println("  ✅ Base64 解码成功")
		fmt.Printf("     解码后长度：%d 字节 (应为 32)\n", len(decoded))
		if len(decoded) != 32 {
			fmt.Printf("  ❌ 解码后应为 32 字节，实际为 %d\n", len(decoded))
		}
	}
	fmt.Println()

	// 清理建议
	fmt.Println("📋 修复建议:")
	fmt.Println("  1. 从企业微信后台重新复制 EncodingAESKey")
	fmt.Println("  2. 确保只复制 43 个字符，不要包含空格或换行")
	fmt.Println("  3. 在 config.json 中粘贴时确保格式正确:")
	fmt.Println()
	fmt.Println("     \"wechat\": {")
	fmt.Println("       \"encoding_aes_key\": \"这里粘贴 43 字符的密钥\",")
	fmt.Println("       \"token\": \"你的 token\"")
	fmt.Println("     }")
	fmt.Println()

	// 显示当前配置的值（脱敏）
	if len(aesKey) > 10 {
		fmt.Printf("当前配置（前 10 字符）: %q...\n", aesKey[:10])
	}
}
