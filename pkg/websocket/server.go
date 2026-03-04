package websocket

import (
	"log"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

// Server WebSocket 服务器
type Server struct {
	hub           *Hub
	upgrader      websocket.Upgrader
	onMessage     func(*Client, *Message)
	onConnect     func(*Client)
	onDisconnect  func(*Client)
}

// ServerConfig 服务器配置
type ServerConfig struct {
	ReadBufferSize  int
	WriteBufferSize int
	CheckOrigin     func(r *http.Request) bool
}

// DefaultServerConfig 默认配置
func DefaultServerConfig() *ServerConfig {
	return &ServerConfig{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin: func(r *http.Request) bool {
			return true // 允许所有来源
		},
	}
}

// NewServer 创建 WebSocket 服务器
func NewServer(config *ServerConfig) *Server {
	if config == nil {
		config = DefaultServerConfig()
	}

	return &Server{
		hub: NewHub(),
		upgrader: websocket.Upgrader{
			ReadBufferSize:  config.ReadBufferSize,
			WriteBufferSize: config.WriteBufferSize,
			CheckOrigin:     config.CheckOrigin,
		},
	}
}

// OnMessage 设置消息处理回调
func (s *Server) OnMessage(fn func(*Client, *Message)) {
	s.onMessage = fn
}

// OnConnect 设置连接回调
func (s *Server) OnConnect(fn func(*Client)) {
	s.onConnect = fn
}

// OnDisconnect 设置断开回调
func (s *Server) OnDisconnect(fn func(*Client)) {
	s.onDisconnect = fn
}

// Start 启动服务器
func (s *Server) Start() {
	go s.hub.Run()
	log.Printf("[WebSocket] Server started")
}

// Stop 停止服务器
func (s *Server) Stop() {
	s.hub.Stop()
	log.Printf("[WebSocket] Server stopped")
}

// HandleWebSocket 处理 WebSocket 连接请求
// 用作 http.HandlerFunc
func (s *Server) HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := s.upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("[WebSocket] Upgrade error: %v", err)
		return
	}

	clientID := generateClientID()
	client := NewClient(clientID, s.hub, conn)

	// 注册客户端
	s.hub.Register(client)

	// 调用连接回调
	if s.onConnect != nil {
		s.onConnect(client)
	}

	// 启动读写循环
	go client.WritePump()
	go func() {
		defer func() {
			if s.onDisconnect != nil {
				s.onDisconnect(client)
			}
		}()
		client.ReadPump(s.onMessage)
	}()
}

// Broadcast 广播消息
func (s *Server) Broadcast(msg *Message) {
	s.hub.Broadcast(msg)
}

// BroadcastToChannel 向频道广播
func (s *Server) BroadcastToChannel(channel string, data interface{}) {
	s.hub.BroadcastToChannel(channel, data)
}

// BroadcastTraffic 广播流量更新
func (s *Server) BroadcastTraffic(traffic interface{}) {
	s.BroadcastToChannel("traffic", traffic)
}

// BroadcastVulnerability 广播漏洞更新
func (s *Server) BroadcastVulnerability(vuln interface{}) {
	s.BroadcastToChannel("vulns", vuln)
}

// ClientCount 获取客户端数量
func (s *Server) ClientCount() int {
	return s.hub.ClientCount()
}

// GetHub 获取 Hub
func (s *Server) GetHub() *Hub {
	return s.hub
}

// generateClientID 生成客户端 ID
func generateClientID() string {
	return uuid.New().String()[:8]
}

// ============ 事件适配器 ============

// TrafficEvent 流量事件
type TrafficEvent struct {
	Type      string      `json:"type"` // new, update, delete
	ID        string      `json:"id"`
	Data      interface{} `json:"data"`
	Timestamp time.Time   `json:"timestamp"`
}

// VulnEvent 漏洞事件
type VulnEvent struct {
	Type      string      `json:"type"` // new, update, delete
	ID        string      `json:"id"`
	Data      interface{} `json:"data"`
	Timestamp time.Time   `json:"timestamp"`
}

// NewTrafficEvent 创建流量事件
func NewTrafficEvent(eventType, id string, data interface{}) *TrafficEvent {
	return &TrafficEvent{
		Type:      eventType,
		ID:        id,
		Data:      data,
		Timestamp: time.Now(),
	}
}

// NewVulnEvent 创建漏洞事件
func NewVulnEvent(eventType, id string, data interface{}) *VulnEvent {
	return &VulnEvent{
		Type:      eventType,
		ID:        id,
		Data:      data,
		Timestamp: time.Now(),
	}
}
