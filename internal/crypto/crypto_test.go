package crypto

import (
	"encoding/base64"
	"encoding/json"
	"testing"
)

func TestWXCrypto(t *testing.T) {
	// 生成一个有效的 32 字节 AESKey (base64 编码)
	aesKeyBytes := []byte("abcdefghijklmnopqrstuvwxyz012345") // 32 字节
	aesKey := base64.StdEncoding.EncodeToString(aesKeyBytes)
	token := "testtoken"

	crypto, err := NewWXCrypto(aesKey, token)
	if err != nil {
		t.Fatalf("NewWXCrypto failed: %v", err)
	}

	// 测试加密和解密
	original := []byte(`{"msgid":"123","msgtype":"text","text":{"content":"hello"}}`)

	encrypted, err := crypto.Encrypt(original)
	if err != nil {
		t.Fatalf("Encrypt failed: %v", err)
	}
	t.Logf("Encrypted: %s", encrypted)

	decrypted, err := crypto.Decrypt(encrypted)
	if err != nil {
		t.Fatalf("Decrypt failed: %v", err)
	}

	if string(decrypted) != string(original) {
		t.Errorf("Decrypted != Original\nGot: %s\nWant: %s", string(decrypted), string(original))
	}

	// 验证解密后的 JSON 可以正确解析
	var msg map[string]interface{}
	if err := json.Unmarshal(decrypted, &msg); err != nil {
		t.Errorf("Unmarshal decrypted data failed: %v", err)
	}
}

func TestVerifySignature(t *testing.T) {
	aesKeyBytes := []byte("abcdefghijklmnopqrstuvwxyz012345")
	aesKey := base64.StdEncoding.EncodeToString(aesKeyBytes)
	token := "testtoken"

	crypto, err := NewWXCrypto(aesKey, token)
	if err != nil {
		t.Fatalf("NewWXCrypto failed: %v", err)
	}

	timestamp := "1234567890"
	nonce := "abcdef"
	echoStr := "test"

	signature := crypto.GetSignature(timestamp, nonce, echoStr)

	if !crypto.VerifySignature(timestamp, nonce, echoStr, signature) {
		t.Error("Signature verification failed")
	}

	// 测试错误签名
	if crypto.VerifySignature(timestamp, nonce, echoStr, "wrong_signature") {
		t.Error("Should reject wrong signature")
	}
}
