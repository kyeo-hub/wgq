package main

import (
	"bytes"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/wgq-bot/wgq/internal/crypto"
	"github.com/wgq-bot/wgq/internal/wechat"
)

// 测试工具：模拟企业微信发送消息到回调服务器

func main() {
	url := flag.String("url", "http://localhost:8080/wechat/callback", "回调服务器 URL")
	aesKey := flag.String("aeskey", "", "EncodingAESKey (base64)")
	token := flag.String("token", "", "Token")
	userID := flag.String("user", "testuser", "发送者用户 ID")
	message := flag.String("msg", "帮我写一个 Hello World 程序", "要发送的消息内容")
	flag.Parse()

	if *aesKey == "" || *token == "" {
		fmt.Println("使用方法:")
		fmt.Println("  go run ./cmd/testclient/main.go -aeskey <YOUR_AES_KEY> -token <YOUR_TOKEN>")
		fmt.Println()
		fmt.Println("可选参数:")
		fmt.Println("  -url     回调服务器地址 (默认：http://localhost:8080/wechat/callback)")
		fmt.Println("  -user    发送者用户 ID (默认：testuser)")
		fmt.Println("  -msg     消息内容 (默认：帮我写一个 Hello World 程序)")
		return
	}

	// 创建加解密工具
	wxCrypto, err := crypto.NewWXCrypto(*aesKey, *token)
	if err != nil {
		fmt.Printf("创建加解密工具失败：%v\n", err)
		return
	}

	// 构建测试消息
	msg := wechat.Message{
		MsgID:       generateMsgID(),
		AIBotID:     "testbot",
		ChatType:    "single",
		From:        wechat.UserInfo{UserID: *userID},
		ResponseURL: "http://localhost:8080/wechat/reply",
		MsgType:     "text",
		Text: &wechat.TextContent{
			Content: *message,
		},
	}

	// 序列化消息
	msgData, err := json.Marshal(msg)
	if err != nil {
		fmt.Printf("序列化消息失败：%v\n", err)
		return
	}

	// 加密消息
	encrypted, err := wxCrypto.Encrypt(msgData)
	if err != nil {
		fmt.Printf("加密消息失败：%v\n", err)
		return
	}

	// 生成签名
	timestamp := strconv.FormatInt(time.Now().Unix(), 10)
	nonce := generateNonce()
	signature := wxCrypto.GetSignature(timestamp, nonce, encrypted)

	// 构建请求 URL
	reqURL := fmt.Sprintf("%s?timestamp=%s&nonce=%s&msg_signature=%s",
		*url, timestamp, nonce, signature)

	// 发送请求
	resp, err := http.Post(reqURL, "application/json", bytes.NewReader([]byte(encrypted)))
	if err != nil {
		fmt.Printf("发送请求失败：%v\n", err)
		return
	}
	defer resp.Body.Close()

	// 读取响应
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("读取响应失败：%v\n", err)
		return
	}

	fmt.Printf("发送成功!\n")
	fmt.Printf("状态码：%d\n", resp.StatusCode)
	fmt.Printf("响应内容：%s\n", string(body))

	if resp.StatusCode == 200 {
		fmt.Println("\n✅ 消息已成功发送到回调服务器")
		fmt.Println("请检查服务器日志查看处理结果")
	} else {
		fmt.Println("\n❌ 发送失败，请检查配置")
	}
}

func generateMsgID() string {
	return fmt.Sprintf("test_%d", time.Now().UnixNano())
}

func generateNonce() string {
	b := make([]byte, 16)
	rand.Read(b)
	return base64.RawURLEncoding.EncodeToString(b)
}
