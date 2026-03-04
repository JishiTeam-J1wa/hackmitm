package websocket

import (
	"encoding/json"
	"log"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

const (
	// writeWait 写超时时间
	writeWait = 10 * time.Second
	// pongWait Pong 消息等待时间
	pongWait = 60 * time.Second
	// pingPeriod Ping 消息发送周期
	pingPeriod = (pongWait * 9) / 10
	// maxMessageSize 最大消息大小
	maxMessageSize = 512 * 1024 // 512KB
)

// Client WebSocket 客户端
// Client represents a WebSocket client connection
type Client struct {
	ID         string
	hub        *Hub
	conn       *websocket.Conn
	send       chan []byte
	subscriptions map[string]bool // 订阅的频道
	mu         sync.RWMutex
	lastPing   time.Time
}

// NewClient 创建 WebSocket 客户端
func NewClient(id string, hub *Hub, conn *websocket.Conn) *Client {
	return &Client{
		ID:            id,
		hub:           hub,
		conn:          conn,
		send:          make(chan []byte, 256),
		subscriptions: make(map[string]bool),
		lastPing:      time.Now(),
	}
}

// ReadPump 读取消息循环
func (c *Client) ReadPump(onMessage func(*Client, *Message)) {
	defer func() {
		c.hub.Unregister(c)
		c.conn.Close()
	}()

	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(pongWait))
		c.mu.Lock()
		c.lastPing = time.Now()
		c.mu.Unlock()
		return nil
	})

	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("[WebSocket] Read error: %v", err)
			}
			break
		}

		var msg Message
		if err := json.Unmarshal(message, &msg); err != nil {
			log.Printf("[WebSocket] Error unmarshaling message: %v", err)
			continue
		}

		// 处理订阅/取消订阅
		c.handleMessage(&msg, onMessage)
	}
}

// WritePump 写入消息循环
func (c *Client) WritePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			// 批量发送队列中的消息
			n := len(c.send)
			for i := 0; i < n; i++ {
				w.Write([]byte{'\n'})
				w.Write(<-c.send)
			}

			if err := w.Close(); err != nil {
				return
			}

		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// handleMessage 处理客户端消息
func (c *Client) handleMessage(msg *Message, onMessage func(*Client, *Message)) {
	switch msg.Action {
	case "subscribe":
		if msg.Channel != "" {
			c.Subscribe(msg.Channel)
			c.sendAck("subscribed", msg.Channel)
		}
	case "unsubscribe":
		if msg.Channel != "" {
			c.Unsubscribe(msg.Channel)
			c.sendAck("unsubscribed", msg.Channel)
		}
	default:
		// 自定义消息处理
		if onMessage != nil {
			onMessage(c, msg)
		}
	}
}

// sendAck 发送确认消息
func (c *Client) sendAck(action, channel string) {
	ack := &Message{
		Type:    "ack",
		Action:  action,
		Channel: channel,
	}
	data, _ := json.Marshal(ack)
	select {
	case c.send <- data:
	default:
	}
}

// Subscribe 订阅频道
func (c *Client) Subscribe(channel string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.subscriptions[channel] = true
}

// Unsubscribe 取消订阅频道
func (c *Client) Unsubscribe(channel string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.subscriptions, channel)
}

// IsSubscribed 检查是否订阅了频道
func (c *Client) IsSubscribed(channel string) bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.subscriptions[channel] || c.subscriptions["*"] // * 表示订阅所有
}

// GetSubscriptions 获取订阅列表
func (c *Client) GetSubscriptions() []string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	channels := make([]string, 0, len(c.subscriptions))
	for ch := range c.subscriptions {
		channels = append(channels, ch)
	}
	return channels
}

// Send 发送消息
func (c *Client) Send(data []byte) error {
	select {
	case c.send <- data:
		return nil
	default:
		return ErrBufferFull
	}
}

// SendJSON 发送 JSON 消息
func (c *Client) SendJSON(v interface{}) error {
	data, err := json.Marshal(v)
	if err != nil {
		return err
	}
	return c.Send(data)
}

// LastPing 获取最后 ping 时间
func (c *Client) LastPing() time.Time {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.lastPing
}

// 错误定义
var ErrBufferFull = &ClientError{Msg: "client buffer full"}

// ClientError 客户端错误
type ClientError struct {
	Msg string
}

func (e *ClientError) Error() string {
	return e.Msg
}
