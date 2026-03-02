package wechat

// Message 企业微信消息结构
type Message struct {
	MsgID      string   `json:"msgid"`       // 消息唯一标识（用于排重）
	AIBotID    string   `json:"aibotid"`     // 智能机器人 ID
	ChatID     string   `json:"chatid"`      // 会话 ID（群聊时返回）
	ChatType   string   `json:"chattype"`    // single/group
	From       UserInfo `json:"from"`        // 发送者信息
	ResponseURL string  `json:"response_url"`// 主动回复临时 URL
	MsgType    string   `json:"msgtype"`     // 消息类型

	// 文本消息
	Text *TextContent `json:"text,omitempty"`

	// 图片消息
	Image *ImageContent `json:"image,omitempty"`

	// 图文混排消息
	Mixed *MixedContent `json:"mixed,omitempty"`

	// 语音消息
	Voice *VoiceContent `json:"voice,omitempty"`

	// 文件消息
	File *FileContent `json:"file,omitempty"`

	// 流式刷新消息
	Stream *StreamContent `json:"stream,omitempty"`

	// 引用消息
	Quote *QuoteContent `json:"quote,omitempty"`
}

// UserInfo 用户信息
type UserInfo struct {
	UserID string `json:"userid"`
}

// TextContent 文本消息内容
type TextContent struct {
	Content string `json:"content"`
}

// ImageContent 图片消息内容
type ImageContent struct {
	URL string `json:"url"`
}

// MixedContent 图文混排消息内容
type MixedContent struct {
	MsgItem []MixedItem `json:"msg_item"`
}

// MixedItem 图文混排消息项
type MixedItem struct {
	Type    string `json:"type"`    // text/image
	Content string `json:"content"` // 文本内容或图片 URL
}

// VoiceContent 语音消息内容
type VoiceContent struct {
	Content string `json:"content"` // 语音文件 URL
}

// FileContent 文件消息内容
type FileContent struct {
	URL string `json:"url"` // 文件 URL (≤100M)
}

// StreamContent 流式刷新消息内容
type StreamContent struct {
	ID string `json:"id"` // 流式消息 ID
}

// QuoteContent 引用消息内容
type QuoteContent struct {
	Content string `json:"content"` // 引用内容
}

// ReplyMessage 回复消息结构
type ReplyMessage struct {
	MsgType string `json:"msgtype"`

	// 文本消息
	Text *TextContent `json:"text,omitempty"`

	// 模板卡片消息
	TemplateCard *TemplateCard `json:"template_card,omitempty"`

	// 流式消息
	Stream *StreamContent `json:"stream,omitempty"`
}

// TemplateCard 模板卡片消息
type TemplateCard struct {
	CardType string        `json:"card_type"`
	Source   CardSource    `json:"source"`
	MainTitle CardTitle    `json:"main_title"`
	SubTitle string        `json:"sub_title"`
	EmphasisContent CardEmphasis `json:"emphasis_content,omitempty"`
	QuoteArea CardQuoteArea `json:"quote_area,omitempty"`
	Actions   []CardAction  `json:"action_menu,omitempty"`
}

// CardSource 卡片来源
type CardSource struct {
	IconURL   string `json:"icon_url"`
	Desc      string `json:"desc"`
	DescColor int    `json:"desc_color"`
}

// CardTitle 卡片标题
type CardTitle struct {
	Title string `json:"title"`
	Desc  string `json:"desc"`
}

// CardEmphasis 强调内容
type CardEmphasis struct {
	Title string `json:"title"`
	Desc  string `json:"desc"`
}

// CardQuoteArea 引用区域
type CardQuoteArea struct {
	Type  string `json:"type"` // text/image
	Content string `json:"content"`
}

// CardAction 卡片动作
type CardAction struct {
	Text string `json:"text"`
	Key  string `json:"key"`
}
