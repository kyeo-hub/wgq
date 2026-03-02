package server

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/wgq-bot/wgq/internal/crypto"
	"github.com/wgq-bot/wgq/internal/wechat"
)

// CallbackServer 企业微信回调服务器
type CallbackServer struct {
	crypto   *crypto.WXCrypto
	handler  MessageHandler
	addr     string
}

// MessageHandler 消息处理接口
type MessageHandler interface {
	HandleMessage(msg *wechat.Message) (*wechat.ReplyMessage, error)
}

// NewCallbackServer 创建回调服务器
func NewCallbackServer(aesKey, token string, handler MessageHandler, addr string) (*CallbackServer, error) {
	wxCrypto, err := crypto.NewWXCrypto(aesKey, token)
	if err != nil {
		return nil, fmt.Errorf("init crypto failed: %w", err)
	}
	return &CallbackServer{
		crypto:  wxCrypto,
		handler: handler,
		addr:    addr,
	}, nil
}

// Start 启动服务器
func (s *CallbackServer) Start() error {
	http.HandleFunc("/wechat/callback", s.handleCallback)
	http.HandleFunc("/health", s.handleHealth)
	
	log.Printf("Starting callback server on %s", s.addr)
	return http.ListenAndServe(s.addr, nil)
}

// handleHealth 健康检查
func (s *CallbackServer) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

// handleCallback 处理回调请求
func (s *CallbackServer) handleCallback(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// 获取 URL 参数
	query := r.URL.Query()
	timestamp := query.Get("timestamp")
	nonce := query.Get("nonce")
	echoStr := query.Get("echostr")
	signature := query.Get("msg_signature")

	// 验证签名
	if !s.crypto.VerifySignature(timestamp, nonce, echoStr, signature) {
		http.Error(w, "Invalid signature", http.StatusForbidden)
		return
	}

	// 读取请求体
	body, err := io.ReadAll(r.Body)
	if err != nil {
		log.Printf("Read body failed: %v", err)
		http.Error(w, "Internal error", http.StatusInternalServerError)
		return
	}
	defer r.Body.Close()

	// 解密消息
	plaintext, err := s.crypto.Decrypt(string(body))
	if err != nil {
		log.Printf("Decrypt failed: %v", err)
		http.Error(w, "Decrypt failed", http.StatusInternalServerError)
		return
	}

	// 解析消息
	var msg wechat.Message
	if err := json.Unmarshal(plaintext, &msg); err != nil {
		log.Printf("Unmarshal message failed: %v", err)
		http.Error(w, "Parse message failed", http.StatusInternalServerError)
		return
	}

	log.Printf("Received message: msgid=%s, from=%s, type=%s", msg.MsgID, msg.From.UserID, msg.MsgType)

	// 处理消息
	reply, err := s.handler.HandleMessage(&msg)
	if err != nil {
		log.Printf("Handle message failed: %v", err)
		http.Error(w, "Handle message failed", http.StatusInternalServerError)
		return
	}

	// 回复消息
	if reply != nil {
		if err := s.sendReply(msg.ResponseURL, reply); err != nil {
			log.Printf("Send reply failed: %v", err)
		}
	}

	// 返回成功响应
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("success"))
}

// sendReply 发送回复消息
func (s *CallbackServer) sendReply(responseURL string, reply *wechat.ReplyMessage) error {
	data, err := json.Marshal(reply)
	if err != nil {
		return fmt.Errorf("marshal reply failed: %w", err)
	}

	// 加密回复
	encrypted, err := s.crypto.Encrypt(data)
	if err != nil {
		return fmt.Errorf("encrypt reply failed: %w", err)
	}

	// 生成签名
	timestamp := strconv.FormatInt(time.Now().Unix(), 10)
	nonce := generateNonce()
	signature := s.crypto.GetSignature(timestamp, nonce, encrypted)

	// 发送请求
	respURL := fmt.Sprintf("%s&timestamp=%s&nonce=%s&msg_signature=%s", 
		responseURL, timestamp, nonce, signature)
	
	resp, err := http.Post(respURL, "application/json", 
		strings.NewReader(encrypted))
	if err != nil {
		return fmt.Errorf("post reply failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("reply response status: %d", resp.StatusCode)
	}

	log.Printf("Reply sent successfully")
	return nil
}

// generateNonce 生成随机 nonce
func generateNonce() string {
	return fmt.Sprintf("%d", time.Now().UnixNano())
}
