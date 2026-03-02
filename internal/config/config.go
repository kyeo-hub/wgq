package config

import (
	"encoding/json"
	"fmt"
	"os"
)

// Config 应用配置
type Config struct {
	// 企业微信配置
	WeChat WeChatConfig `json:"wechat"`

	// 服务器配置
	Server ServerConfig `json:"server"`

	// qwen-code 配置
	Qwen QwenConfig `json:"qwen"`

	// 用户白名单
	AllowedUsers []string `json:"allowed_users"`
}

// WeChatConfig 企业微信配置
type WeChatConfig struct {
	// 企业微信后台配置的 EncodingAESKey (base64 编码)
	EncodingAESKey string `json:"encoding_aes_key"`

	// 企业微信后台配置的 Token
	Token string `json:"token"`

	// 机器人 ID (可选)
	BotID string `json:"bot_id"`
}

// ServerConfig 服务器配置
type ServerConfig struct {
	// 监听地址
	Addr string `json:"addr"`

	// 回调路径
	CallbackPath string `json:"callback_path"`
}

// QwenConfig qwen-code 配置
type QwenConfig struct {
	// 工作目录
	WorkDir string `json:"work_dir"`

	// 命令超时时间（秒）
	TimeoutSeconds int `json:"timeout_seconds"`

	// 最大输出行数
	MaxOutputLines int `json:"max_output_lines"`
}

// DefaultConfig 默认配置
func DefaultConfig() *Config {
	return &Config{
		WeChat: WeChatConfig{
			EncodingAESKey: "",
			Token:          "",
			BotID:          "",
		},
		Server: ServerConfig{
			Addr:         ":8080",
			CallbackPath: "/wechat/callback",
		},
		Qwen: QwenConfig{
			WorkDir:        "/tmp/qwen-workspace",
			TimeoutSeconds: 300,
			MaxOutputLines: 500,
		},
		AllowedUsers: []string{},
	}
}

// LoadConfig 从文件加载配置
func LoadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read config file failed: %w", err)
	}

	config := DefaultConfig()
	if err := json.Unmarshal(data, config); err != nil {
		return nil, fmt.Errorf("parse config file failed: %w", err)
	}

	return config, nil
}

// SaveConfig 保存配置到文件
func SaveConfig(path string, config *Config) error {
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal config failed: %w", err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("write config file failed: %w", err)
	}

	return nil
}

// Validate 验证配置
func (c *Config) Validate() error {
	if c.WeChat.EncodingAESKey == "" {
		return fmt.Errorf("wechat encoding_aes_key is required")
	}

	if c.WeChat.Token == "" {
		return fmt.Errorf("wechat token is required")
	}

	if c.Server.Addr == "" {
		return fmt.Errorf("server addr is required")
	}

	if c.Qwen.WorkDir == "" {
		return fmt.Errorf("qwen work_dir is required")
	}

	if c.Qwen.TimeoutSeconds <= 0 {
		return fmt.Errorf("qwen timeout_seconds must be positive")
	}

	return nil
}
