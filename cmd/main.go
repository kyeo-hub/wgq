package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/wgq-bot/wgq/internal/config"
	"github.com/wgq-bot/wgq/internal/handler"
	"github.com/wgq-bot/wgq/internal/qwen"
	"github.com/wgq-bot/wgq/internal/server"
)

var (
	version = "dev"
	buildTime = "unknown"
)

func main() {
	// 解析命令行参数
	configPath := flag.String("config", "config.json", "配置文件路径")
	showVersion := flag.Bool("version", false, "显示版本信息")
	flag.Parse()

	if *showVersion {
		fmt.Printf("wgq-bot version %s (built at %s)\n", version, buildTime)
		os.Exit(0)
	}

	// 加载配置
	cfg, err := config.LoadConfig(*configPath)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// 验证配置
	if err := cfg.Validate(); err != nil {
		log.Fatalf("Invalid config: %v", err)
	}

	log.Println("Configuration loaded successfully")

	// 检查 qwen 是否安装
	if !qwen.CheckInstalled() {
		log.Println("Warning: qwen is not installed or not in PATH")
		log.Println("Please install qwen: npm install -g @qwen-code/qwen-code@latest")
	}

	// 创建 qwen 执行器
	executor := qwen.NewExecutor(qwen.ExecutorConfig{
		WorkDir:   cfg.Qwen.WorkDir,
		Timeout:   time.Duration(cfg.Qwen.TimeoutSeconds) * time.Second,
		MaxOutput: cfg.Qwen.MaxOutputLines,
	})

	// 创建消息处理器
	msgHandler := handler.NewMessageHandler(executor, cfg.AllowedUsers)

	// 创建回调服务器
	callbackServer, err := server.NewCallbackServer(
		cfg.WeChat.EncodingAESKey,
		cfg.WeChat.Token,
		msgHandler,
		cfg.Server.Addr,
	)
	if err != nil {
		log.Fatalf("Failed to create callback server: %v", err)
	}

	log.Printf("Starting wgq-bot server...")
	log.Printf("Listening on %s", cfg.Server.Addr)
	log.Printf("Callback path: %s", cfg.Server.CallbackPath)

	// 启动服务器
	if err := callbackServer.Start(); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
