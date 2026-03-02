package main

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/wgq-bot/wgq/internal/config"
	"github.com/wgq-bot/wgq/internal/crypto"
	"github.com/wgq-bot/wgq/internal/qwen"
)

// 配置验证工具：检查配置是否正确

func main() {
	fmt.Println("========================================")
	fmt.Println("  wgq-bot 配置验证工具")
	fmt.Println("========================================")
	fmt.Println()

	// 加载配置
	cfgPath := "config.json"
	if len(os.Args) > 1 {
		cfgPath = os.Args[1]
	}

	fmt.Printf("📄 加载配置文件：%s\n", cfgPath)
	cfg, err := config.LoadConfig(cfgPath)
	if err != nil {
		fmt.Printf("❌ 加载配置失败：%v\n", err)
		return
	}
	fmt.Println("✅ 配置文件加载成功")
	fmt.Println()

	// 验证配置
	fmt.Println("🔍 验证配置项...")
	if err := cfg.Validate(); err != nil {
		fmt.Printf("❌ 配置验证失败：%v\n", err)
		return
	}
	fmt.Println("✅ 配置项验证通过")
	fmt.Println()

	// 测试加解密
	fmt.Println("🔐 测试加解密模块...")
	wxCrypto, err := crypto.NewWXCrypto(cfg.WeChat.EncodingAESKey, cfg.WeChat.Token)
	if err != nil {
		fmt.Printf("❌ 加解密模块初始化失败：%v\n", err)
		fmt.Println()
		fmt.Println("可能的原因:")
		fmt.Println("  1. EncodingAESKey 格式不正确（应该是 base64 编码的 32 字节字符串）")
		fmt.Println("  2. Token 为空")
		return
	}
	fmt.Println("✅ 加解密模块初始化成功")

	// 测试加密解密
	testData := []byte(`{"test":"hello"}`)
	encrypted, err := wxCrypto.Encrypt(testData)
	if err != nil {
		fmt.Printf("❌ 加密测试失败：%v\n", err)
		return
	}
	fmt.Printf("   加密测试：成功 (密文长度：%d)\n", len(encrypted))

	decrypted, err := wxCrypto.Decrypt(encrypted)
	if err != nil {
		fmt.Printf("❌ 解密测试失败：%v\n", err)
		return
	}
	if string(decrypted) != string(testData) {
		fmt.Printf("❌ 解密结果不匹配\n")
		return
	}
	fmt.Println("   解密测试：成功")
	fmt.Println("✅ 加解密测试通过")
	fmt.Println()

	// 检查 qwen
	fmt.Println("🤖 检查 qwen 安装状态...")
	if qwen.CheckInstalled() {
		fmt.Println("✅ qwen 已安装")
		// 尝试获取版本
		cmd := exec.Command("qwen", "--version")
		if output, err := cmd.Output(); err == nil {
			fmt.Printf("   版本：%s\n", string(output))
		}
	} else {
		fmt.Println("⚠️  qwen 未安装或不在 PATH 中")
		fmt.Println("   请运行：npm install -g @qwen-code/qwen-code@latest")
	}
	fmt.Println()

	// 显示配置摘要
	fmt.Println("📋 配置摘要:")
	fmt.Printf("   服务器地址：%s\n", cfg.Server.Addr)
	fmt.Printf("   工作目录：%s\n", cfg.Qwen.WorkDir)
	fmt.Printf("   超时时间：%d 秒\n", cfg.Qwen.TimeoutSeconds)
	fmt.Printf("   最大输出：%d 行\n", cfg.Qwen.MaxOutputLines)
	if len(cfg.AllowedUsers) > 0 {
		fmt.Printf("   白名单用户：%d 人\n", len(cfg.AllowedUsers))
	} else {
		fmt.Println("   白名单用户：无限制")
	}
	fmt.Println()

	fmt.Println("========================================")
	fmt.Println("  ✅ 配置验证完成！")
	fmt.Println("========================================")
	fmt.Println()
	fmt.Println("下一步:")
	fmt.Println("  1. 启动服务器：go run ./cmd/main.go -config config.json")
	fmt.Println("  2. 运行测试客户端：go run ./cmd/testclient/main.go -aeskey <AES_KEY> -token <TOKEN>")
	fmt.Println("  3. 配置企业微信回调 URL: http://你的服务器 IP:8080/wechat/callback")
}
