package handler

import (
	"context"
	"fmt"
	"log"
	"regexp"
	"strings"
	"time"

	"github.com/wgq-bot/wgq/internal/qwen"
	"github.com/wgq-bot/wgq/internal/wechat"
)

// MessageHandler 消息处理器
type MessageHandler struct {
	executor      *qwen.Executor
	allowedUsers  map[string]bool // 允许的用户白名单
	sessionCache  map[string]*Session // 会话缓存
}

// Session 会话状态
type Session struct {
	UserID    string
	StartTime time.Time
	LastTime  time.Time
	TaskID    string
	Status    string
}

// NewMessageHandler 创建消息处理器
func NewMessageHandler(executor *qwen.Executor, allowedUsers []string) *MessageHandler {
	userMap := make(map[string]bool)
	for _, user := range allowedUsers {
		userMap[user] = true
	}

	return &MessageHandler{
		executor:     executor,
		allowedUsers: userMap,
		sessionCache: make(map[string]*Session),
	}
}

// HandleMessage 处理企业微信消息
func (h *MessageHandler) HandleMessage(msg *wechat.Message) (*wechat.ReplyMessage, error) {
	log.Printf("Processing message from user: %s", msg.From.UserID)

	// 检查用户白名单
	if !h.isUserAllowed(msg.From.UserID) {
		return h.createTextReply("❌ 您没有权限使用此机器人，请联系管理员。"), nil
	}

	// 根据消息类型处理
	switch msg.MsgType {
	case "text":
		return h.handleTextMessage(msg)
	case "mixed":
		return h.handleMixedMessage(msg)
	default:
		return h.createTextReply("⚠️ 暂不支持此消息类型，请发送文本消息。"), nil
	}
}

// handleTextMessage 处理文本消息
func (h *MessageHandler) handleTextMessage(msg *wechat.Message) (*wechat.ReplyMessage, error) {
	if msg.Text == nil {
		return h.createTextReply("⚠️ 无效的消息格式。"), nil
	}

	content := strings.TrimSpace(msg.Text.Content)
	if content == "" {
		return h.createTextReply("⚠️ 消息内容不能为空。"), nil
	}

	// 解析命令
	cmd := h.parseCommand(content)

	// 执行命令
	ctx := context.Background()
	result, err := h.executor.Execute(ctx, cmd.Prompt)
	if err != nil {
		log.Printf("Execute failed: %v", err)
		return h.createTextReply(fmt.Sprintf("❌ 执行失败：%v", err)), nil
	}

	// 构建回复
	reply := h.buildReply(result, cmd)
	return reply, nil
}

// handleMixedMessage 处理图文混排消息
func (h *MessageHandler) handleMixedMessage(msg *wechat.Message) (*wechat.ReplyMessage, error) {
	if msg.Mixed == nil || len(msg.Mixed.MsgItem) == 0 {
		return h.createTextReply("⚠️ 无效的消息格式。"), nil
	}

	// 提取文本内容
	var textParts []string
	for _, item := range msg.Mixed.MsgItem {
		if item.Type == "text" {
			textParts = append(textParts, item.Content)
		}
	}

	if len(textParts) == 0 {
		return h.createTextReply("⚠️ 请提供文本内容。"), nil
	}

	prompt := strings.Join(textParts, " ")
	ctx := context.Background()
	result, err := h.executor.Execute(ctx, prompt)
	if err != nil {
		return h.createTextReply(fmt.Sprintf("❌ 执行失败：%v", err)), nil
	}

	reply := h.buildReply(result, Command{Type: CmdNormal})
	return reply, nil
}

// Command 解析后的命令
type Command struct {
	Type    CommandType
	Prompt  string
	Args    []string
}

// CommandType 命令类型
type CommandType int

const (
	CmdNormal CommandType = iota // 普通执行
	CmdHelp                      // 帮助命令
	CmdStatus                    // 状态查询
	CmdCancel                    // 取消任务
	CmdVersion                   // 版本查询
)

// parseCommand 解析命令
func (h *MessageHandler) parseCommand(content string) Command {
	content = strings.TrimSpace(content)

	// 帮助命令
	if content == "/help" || content == "帮助" || content == "help" {
		return Command{
			Type:   CmdHelp,
			Prompt: "",
		}
	}

	// 状态命令
	if content == "/status" || content == "状态" {
		return Command{
			Type:   CmdStatus,
			Prompt: "",
		}
	}

	// 版本命令
	if content == "/version" || content == "版本" {
		return Command{
			Type:   CmdVersion,
			Prompt: "",
		}
	}

	// 取消命令
	if content == "/cancel" || content == "取消" {
		return Command{
			Type:   CmdCancel,
			Prompt: "",
		}
	}

	// 默认普通命令，直接作为 prompt 传递给 qwen-code
	return Command{
		Type:   CmdNormal,
		Prompt: content,
	}
}

// buildReply 构建回复消息
func (h *MessageHandler) buildReply(result *qwen.ExecutionResult, cmd Command) *wechat.ReplyMessage {
	switch cmd.Type {
	case CmdHelp:
		return h.createTextReply(h.getHelpText())
	case CmdStatus:
		return h.createTextReply(h.getStatusText())
	case CmdVersion:
		version, _ := h.executor.GetVersion()
		if version == "" {
			version = "未知"
		}
		return h.createTextReply(fmt.Sprintf("📦 qwen-code 版本：%s", version))
	case CmdCancel:
		return h.createTextReply("✅ 任务已取消（如果有正在执行的任务）。")
	default:
		// 普通执行结果
		return h.createExecutionResultReply(result)
	}
}

// createExecutionResultReply 创建执行结果回复
func (h *MessageHandler) createExecutionResultReply(result *qwen.ExecutionResult) *wechat.ReplyMessage {
	var builder strings.Builder

	// 状态图标
	statusIcon := "✅"
	if !result.Success {
		statusIcon = "❌"
	}

	// 执行时间
	durationStr := fmt.Sprintf("%.2f 秒", result.Duration.Seconds())

	// 构建消息
	builder.WriteString(fmt.Sprintf("%s 执行完成\n", statusIcon))
	builder.WriteString(fmt.Sprintf("⏱️ 耗时：%s\n\n", durationStr))

	if result.Error != "" {
		builder.WriteString("❌ 错误信息:\n")
		builder.WriteString(h.truncateForReply(result.Error))
	} else {
		builder.WriteString("📋 输出:\n")
		builder.WriteString(h.truncateForReply(result.Output))
	}

	return h.createTextReply(builder.String())
}

// truncateForReply 截断消息以适应企业微信限制
func (h *MessageHandler) truncateForReply(text string) string {
	const maxLen = 1500 // 企业微信文本消息长度限制
	if len(text) <= maxLen {
		return text
	}
	return text[:maxLen-50] + "\n\n... [内容过长，已截断]"
}

// createTextReply 创建文本回复
func (h *MessageHandler) createTextReply(content string) *wechat.ReplyMessage {
	return &wechat.ReplyMessage{
		MsgType: "text",
		Text: &wechat.TextContent{
			Content: content,
		},
	}
}

// isUserAllowed 检查用户是否在白名单中
func (h *MessageHandler) isUserAllowed(userID string) bool {
	if len(h.allowedUsers) == 0 {
		return true // 空名单表示允许所有用户
	}
	return h.allowedUsers[userID]
}

// getHelpText 获取帮助文本
func (h *MessageHandler) getHelpText() string {
	return `🤖 企业微信智能机器人 - 使用帮助

📌 支持的命令:
  /help     - 显示此帮助信息
  /status   - 查看系统状态
  /version  - 查看 qwen-code 版本
  /cancel   - 取消当前任务

💡 使用方式:
直接发送您的编程问题或任务描述，机器人将调用 qwen-code 为您处理。

示例:
  "帮我写一个快速排序函数"
  "解释这段代码的作用"
  "帮我修复这个 bug"

⚠️ 注意事项:
  - 每个命令最多等待 5 分钟
  - 输出内容过长会被截断
  - 请勿发送敏感信息`
}

// getStatusText 获取状态文本
func (h *MessageHandler) getStatusText() string {
	var builder strings.Builder

	builder.WriteString("📊 系统状态\n\n")
	
	// 检查 qwen-code 是否安装
	if qwen.CheckInstalled() {
		builder.WriteString("✅ qwen-code: 已安装\n")
	} else {
		builder.WriteString("❌ qwen-code: 未安装\n")
	}

	// 会话数量
	builder.WriteString(fmt.Sprintf("👥 活跃会话：%d\n", len(h.sessionCache)))

	// 清理过期会话
	h.cleanupSessions()

	return builder.String()
}

// cleanupSessions 清理过期会话
func (h *MessageHandler) cleanupSessions() {
	now := time.Now()
	for userID, session := range h.sessionCache {
		if now.Sub(session.LastTime) > 30*time.Minute {
			delete(h.sessionCache, userID)
		}
	}
}

// AddUserToSession 更新用户会话
func (h *MessageHandler) AddUserToSession(userID string) {
	session := &Session{
		UserID:    userID,
		StartTime: time.Now(),
		LastTime:  time.Now(),
		Status:    "active",
	}
	h.sessionCache[userID] = session
}

// regex patterns for command parsing
var (
	cmdPattern = regexp.MustCompile(`^/(\w+)\s*(.*)`)
)
