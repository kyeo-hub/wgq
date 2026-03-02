package crypto

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha1"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"sort"
	"strings"
)

var (
	ErrInvalidPadding = errors.New("invalid padding size")
	ErrInvalidAESKey  = errors.New("invalid AES key length")
)

// WXCrypto 企业微信加解密工具
type WXCrypto struct {
	aesKey []byte
	token  string
}

// NewWXCrypto 创建加解密工具
// aesKey: 企业微信后台配置的 EncodingAESKey (base64 编码，43 字符)
// token: 企业微信后台配置的 Token
func NewWXCrypto(aesKeyBase64, token string) (*WXCrypto, error) {
	// 企业微信的 EncodingAESKey 是 43 字符，需要补 padding 到 44 字符
	// 43 % 4 = 3，所以需要补 1 个 '='
	aesKeyBase64 = strings.TrimRight(aesKeyBase64, "=") // 先移除可能已有的 padding
	for len(aesKeyBase64)%4 != 0 {
		aesKeyBase64 += "="
	}

	aesKey, err := base64.StdEncoding.DecodeString(aesKeyBase64)
	if err != nil {
		return nil, err
	}
	if len(aesKey) != 32 {
		return nil, ErrInvalidAESKey
	}
	return &WXCrypto{
		aesKey: aesKey,
		token:  token,
	}, nil
}

// PKCS7Pad PKCS#7 填充
func PKCS7Pad(data []byte, blockSize int) []byte {
	padding := blockSize - len(data)%blockSize
	padText := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(data, padText...)
}

// PKCS7Unpad PKCS#7 去填充
func PKCS7Unpad(data []byte, blockSize int) ([]byte, error) {
	if len(data)%blockSize != 0 {
		return nil, ErrInvalidPadding
	}
	padding := int(data[len(data)-1])
	if padding > blockSize || padding > len(data) {
		return nil, ErrInvalidPadding
	}
	for i := 0; i < padding; i++ {
		if data[len(data)-1-i] != byte(padding) {
			return nil, ErrInvalidPadding
		}
	}
	return data[:len(data)-padding], nil
}

// Encrypt 加密消息
// 返回 base64 编码的密文
func (c *WXCrypto) Encrypt(plaintext []byte) (string, error) {
	block, err := aes.NewCipher(c.aesKey)
	if err != nil {
		return "", err
	}

	// PKCS#7 填充
	plaintext = PKCS7Pad(plaintext, aes.BlockSize)

	// 生成随机 IV (AESKey 前 16 字节作为 IV 的初始向量，但实际使用随机 IV)
	iv := make([]byte, aes.BlockSize)
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return "", err
	}

	// CBC 加密
	mode := cipher.NewCBCEncrypter(block, iv)
	ciphertext := make([]byte, len(plaintext))
	mode.CryptBlocks(ciphertext, plaintext)

	// 拼接 IV + 密文，然后 base64 编码
	result := append(iv, ciphertext...)
	return base64.StdEncoding.EncodeToString(result), nil
}

// Decrypt 解密消息
// ciphertext: base64 编码的密文
func (c *WXCrypto) Decrypt(ciphertext string) ([]byte, error) {
	data, err := base64.StdEncoding.DecodeString(ciphertext)
	if err != nil {
		return nil, err
	}

	block, err := aes.NewCipher(c.aesKey)
	if err != nil {
		return nil, err
	}

	if len(data) < aes.BlockSize {
		return nil, ErrInvalidPadding
	}

	// 提取 IV
	iv := data[:aes.BlockSize]
	data = data[aes.BlockSize:]

	// CBC 解密
	mode := cipher.NewCBCDecrypter(block, iv)
	mode.CryptBlocks(data, data)

	// 去填充
	plaintext, err := PKCS7Unpad(data, aes.BlockSize)
	if err != nil {
		return nil, err
	}

	return plaintext, nil
}

// VerifySignature 验证签名
// timestamp: 时间戳
// nonce: 随机数
// echoStr: 回显字符串（验证模式）或密文（普通模式）
// signature: 签名
func (c *WXCrypto) VerifySignature(timestamp, nonce, echoStr, signature string) bool {
	// 将 token, timestamp, nonce 排序后 sha1
	tmp := []string{c.token, timestamp, nonce}
	sort.Strings(tmp)
	h := sha1.New()
	h.Write([]byte(strings.Join(tmp, "")))
	hash := fmt.Sprintf("%x", h.Sum(nil))
	return hash == signature
}

// GetSignature 获取签名
func (c *WXCrypto) GetSignature(timestamp, nonce, echoStr string) string {
	tmp := []string{c.token, timestamp, nonce}
	sort.Strings(tmp)
	h := sha1.New()
	h.Write([]byte(strings.Join(tmp, "")))
	return fmt.Sprintf("%x", h.Sum(nil))
}
