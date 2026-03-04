// Package websocket 提供 WebSocket 实时推送功能
// Package websocket provides WebSocket real-time push capabilities
package websocket

import (
	"encoding/json"
	"log"
	"sync"
	"time"
)

// Hub WebSocket 连接管理中心
// Hub manages WebSocket connections
type Hub struct {
	clients    map[*Client]bool
	broadcast  chan *Message
	register   chan *Client
	unregister chan *Client
	mu         sync.RWMutex
	running    bool
	done       chan struct{}
}

// Message WebSocket 消息
// Message represents a WebSocket message
type Message struct {
	Type      string      `json:"type"`       // traffic, vuln, intercept, etc.
	Action    string      `json:"action"`     // subscribe, unsubscribe, data, etc.
	Channel   string      `json:"channel"`    // traffic, vulns, intercept
	Data      interface{} `json:"data"`       // 消息数据
	Timestamp time.Time   `json:"timestamp"`  // 时间戳
	ClientID  string      `json:"client_id"`  // 发送客户端 ID (可选)
}

// NewHub 创建 WebSocket Hub
func NewHub() *Hub {
	return &Hub{
		clients:    make(map[*Client]bool),
		broadcast:  make(chan *Message, 256),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		done:       make(chan struct{}),
	}
}

// Run 运行 Hub
func (h *Hub) Run() {
	h.mu.Lock()
	h.running = true
	h.mu.Unlock()

	for {
		select {
		case <-h.done:
			return
		case client := <-h.register:
			h.mu.Lock()
			h.clients[client] = true
			h.mu.Unlock()
			log.Printf("[WebSocket] Client connected: %s, total: %d", client.ID, len(h.clients))

		case client := <-h.unregister:
			h.mu.Lock()
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.send)
			}
			h.mu.Unlock()
			log.Printf("[WebSocket] Client disconnected: %s, total: %d", client.ID, len(h.clients))

		case message := <-h.broadcast:
			h.mu.RLock()
			data, err := json.Marshal(message)
			if err != nil {
				log.Printf("[WebSocket] Error marshaling message: %v", err)
				h.mu.RUnlock()
				continue
			}

			for client := range h.clients {
				// 只发送给订阅了相应 channel 的客户端
				if client.IsSubscribed(message.Channel) {
					select {
					case client.send <- data:
					default:
						// 客户端缓冲区满，关闭连接
						close(client.send)
						delete(h.clients, client)
					}
				}
			}
			h.mu.RUnlock()
		}
	}
}

// Stop 停止 Hub
func (h *Hub) Stop() {
	h.mu.Lock()
	defer h.mu.Unlock()

	if !h.running {
		return
	}

	close(h.done)
	h.running = false

	// 关闭所有客户端连接
	for client := range h.clients {
		close(client.send)
		delete(h.clients, client)
	}
}

// Register 注册客户端
func (h *Hub) Register(client *Client) {
	h.register <- client
}

// Unregister 注销客户端
func (h *Hub) Unregister(client *Client) {
	h.unregister <- client
}

// Broadcast 广播消息
func (h *Hub) Broadcast(msg *Message) {
	msg.Timestamp = time.Now()
	h.broadcast <- msg
}

// BroadcastToChannel 向特定频道广播
func (h *Hub) BroadcastToChannel(channel string, data interface{}) {
	h.Broadcast(&Message{
		Type:    "data",
		Action:  "update",
		Channel: channel,
		Data:    data,
	})
}

// ClientCount 获取客户端数量
func (h *Hub) ClientCount() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.clients)
}

// IsRunning 检查是否在运行
func (h *Hub) IsRunning() bool {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.running
}
