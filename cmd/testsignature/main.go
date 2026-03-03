package main

import (
	"crypto/sha1"
	"fmt"
	"sort"
	"strings"
)

// 测试签名验证
func main() {
	// 企业微信测试参数
	token := "YOUR_TOKEN_HERE" // 替换为你的 token
	timestamp := "1772504484"
	nonce := "6zwnkbkpgai"
	signature := "f939272e1f375ba90ff8e8e218720bc2224057e0"

	// 计算签名
	tmp := []string{token, timestamp, nonce}
	sort.Strings(tmp)
	h := sha1.New()
	h.Write([]byte(strings.Join(tmp, "")))
	hash := fmt.Sprintf("%x", h.Sum(nil))

	fmt.Printf("Token: %s\n", token)
	fmt.Printf("Timestamp: %s\n", timestamp)
	fmt.Printf("Nonce: %s\n", nonce)
	fmt.Printf("Expected signature: %s\n", signature)
	fmt.Printf("Calculated signature: %s\n", hash)
	fmt.Printf("Match: %v\n", hash == signature)
}
