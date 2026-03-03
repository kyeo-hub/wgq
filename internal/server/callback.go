package server

import (
	"crypto/sha1"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/wgq-bot/wgq/internal/crypto"
	"github.com/wgq-bot/wgq/internal/wechat"
)

// CallbackServer 企业微信回调服务器
type CallbackServer struct {
	crypto       *crypto.WXCrypto
	handler      MessageHandler
	addr         string
	callbackPath string
	token        string
}

// MessageHandler 消息处理接口
type MessageHandler interface {
	HandleMessage(msg *wechat.Message) (*wechat.ReplyMessage, error)
}

// NewCallbackServer 创建回调服务器
func NewCallbackServer(aesKey, token, callbackPath string, handler MessageHandler, addr string, corpID string) (*CallbackServer, error) {
	wxCrypto, err := crypto.NewWXCrypto(aesKey, token)
	if err != nil {
		return nil, fmt.Errorf("init crypto failed: %w", err)
	}
	return &CallbackServer{
		crypto:       wxCrypto,
		handler:      handler,
		addr:         addr,
		callbackPath: callbackPath,
		token:        token,
	}, nil
}

// Start 启动服务器
func (s *CallbackServer) Start() error {
	http.HandleFunc(s.callbackPath, s.handleCallback)
	http.HandleFunc("/health", s.handleHealth)

	log.Printf("Starting callback server on %s", s.addr)
	log.Printf("Callback path: %s", s.callbackPath)
	return http.ListenAndServe(s.addr, nil)
}

// handleHealth 健康检查
func (s *CallbackServer) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

// handleCallback 处理回调请求（支持 GET 验证和 POST 消息）
func (s *CallbackServer) handleCallback(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	timestamp := query.Get("timestamp")
	nonce := query.Get("nonce")
	echoStr := query.Get("echostr")
	signature := query.Get("msg_signature")

	log.Printf("Callback received: method=%s, ts=%s, nonce=%s, sig=%s, echoStr_len=%d",
		r.Method, timestamp, nonce, signature, len(echoStr))

	// 验证签名
	if !s.verifySignature(timestamp, nonce, echoStr, signature) {
		log.Printf("Invalid signature: expected=%s", signature)
		http.Error(w, "Invalid signature", http.StatusForbidden)
		return
	}

	switch r.Method {
	case http.MethodGet:
		// URL 验证 - 解密并返回 echostr
		s.handleGetCallback(w, timestamp, nonce, echoStr)

	case http.MethodPost:
		// 消息处理
		s.handlePostCallback(w, r, timestamp, nonce)

	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// handleGetCallback 处理 GET 回调（URL 验证）
func (s *CallbackServer) handleGetCallback(w http.ResponseWriter, timestamp, nonce, echoStr string) {
	plaintext, err := s.crypto.Decrypt(echoStr)
	if err != nil {
		log.Printf("Decrypt echostr failed: %v", err)
		http.Error(w, fmt.Sprintf("Decrypt failed: %v", err), http.StatusInternalServerError)
		return
	}

	// 返回解密后的 echostr
	w.Write(plaintext)
	log.Printf("URL verified successfully, echostr: %s", string(plaintext))
}

// handlePostCallback 处理 POST 回调（消息接收）
func (s *CallbackServer) handlePostCallback(w http.ResponseWriter, r *http.Request, timestamp, nonce string) {
	// 读取请求体
	body, err := io.ReadAll(r.Body)
	if err != nil {
		log.Printf("Read body failed: %v", err)
		http.Error(w, "Internal error", http.StatusInternalServerError)
		return
	}
	defer r.Body.Close()

	if len(body) == 0 {
		log.Printf("Empty request body")
		http.Error(w, "Empty body", http.StatusBadRequest)
		return
	}

	log.Printf("Received POST body length: %d", len(body))

	// 解密消息
	plaintext, err := s.crypto.Decrypt(string(body))
	if err != nil {
		log.Printf("Decrypt message failed: %v", err)
		http.Error(w, fmt.Sprintf("Decrypt failed: %v", err), http.StatusInternalServerError)
		return
	}

	log.Printf("Decrypted message: %s", string(plaintext))

	// 解析消息
	var msg wechat.Message
	if err := json.Unmarshal(plaintext, &msg); err != nil {
		log.Printf("Unmarshal message failed: %v", err)
		http.Error(w, fmt.Sprintf("Parse message failed: %v", err), http.StatusInternalServerError)
		return
	}

	log.Printf("Received message: msgid=%s, from=%s, type=%s, chatid=%s",
		msg.MsgID, msg.From.UserID, msg.MsgType, msg.ChatID)

	// 处理消息
	reply, err := s.handler.HandleMessage(&msg)
	if err != nil {
		log.Printf("Handle message failed: %v", err)
		http.Error(w, fmt.Sprintf("Handle message failed: %v", err), http.StatusInternalServerError)
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

// verifySignature 验证签名
func (s *CallbackServer) verifySignature(timestamp, nonce, echoStr, signature string) bool {
	tmp := []string{s.token, timestamp, nonce}
	sort.Strings(tmp)
	h := sha1.New()
	h.Write([]byte(strings.Join(tmp, "")))
	hash := fmt.Sprintf("%x", h.Sum(nil))
	log.Printf("Verify signature: token=%s, ts=%s, nonce=%s, expected=%s, got=%s",
		s.token, timestamp, nonce, hash, signature)
	return hash == signature
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
	signature := s.getSignature(timestamp, nonce, encrypted)

	// 发送请求
	respURL := fmt.Sprintf("%s&timestamp=%s&nonce=%s&msg_signature=%s",
		responseURL, timestamp, nonce, signature)

	log.Printf("Sending reply to: %s", respURL)

	resp, err := http.Post(respURL, "application/json",
		strings.NewReader(encrypted))
	if err != nil {
		return fmt.Errorf("post reply failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	log.Printf("Reply response: status=%d, body=%s", resp.StatusCode, string(respBody))

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("reply response status: %d", resp.StatusCode)
	}

	log.Printf("Reply sent successfully")
	return nil
}

// getSignature 获取签名
func (s *CallbackServer) getSignature(timestamp, nonce, encrypted string) string {
	tmp := []string{s.token, timestamp, nonce}
	sort.Strings(tmp)
	h := sha1.New()
	h.Write([]byte(strings.Join(tmp, "")))
	return fmt.Sprintf("%x", h.Sum(nil))
}

// generateNonce 生成随机 nonce
func generateNonce() string {
	return fmt.Sprintf("%d", time.Now().UnixNano())
}
