package websocket

import (
	"encoding/json"
	"log"
	"sync"
	"time"
)

// Interceptor 拦截管理器
// 用于实时拦截和修改请求
type Interceptor struct {
	hub           *Hub
	enabled       bool
	interceptMode InterceptMode
	pending       map[string]*InterceptedRequest // request_id -> request
	mu            sync.RWMutex
	onForward     func(requestID string, modified *ModifiedRequest)
	onDrop        func(requestID string)
}

// InterceptMode 拦截模式
type InterceptMode string

const (
	InterceptModeOff    InterceptMode = "off"
	InterceptModeAll    InterceptMode = "all"
	InterceptModeFilter InterceptMode = "filter"
)

// InterceptedRequest 被拦截的请求
type InterceptedRequest struct {
	ID          string                 `json:"id"`
	Method      string                 `json:"method"`
	URL         string                 `json:"url"`
	Headers     map[string]string      `json:"headers"`
	Body        string                 `json:"body"`
	Timestamp   time.Time              `json:"timestamp"`
	Waiting     bool                   `json:"waiting"`
	Metadata    map[string]interface{} `json:"metadata"`
}

// ModifiedRequest 修改后的请求
type ModifiedRequest struct {
	ID      string            `json:"id"`
	Method  string            `json:"method"`
	URL     string            `json:"url"`
	Headers map[string]string `json:"headers"`
	Body    string            `json:"body"`
}

// InterceptFilter 拦截过滤器
type InterceptFilter struct {
	Methods    []string `json:"methods"`
	Hosts      []string `json:"hosts"`
	PathPrefix []string `json:"path_prefix"`
	Extensions []string `json:"extensions"`
}

// NewInterceptor 创建拦截器
func NewInterceptor(hub *Hub) *Interceptor {
	return &Interceptor{
		hub:           hub,
		enabled:       false,
		interceptMode: InterceptModeOff,
		pending:       make(map[string]*InterceptedRequest),
	}
}

// SetMode 设置拦截模式
func (i *Interceptor) SetMode(mode InterceptMode) {
	i.mu.Lock()
	defer i.mu.Unlock()
	i.interceptMode = mode
	i.enabled = mode != InterceptModeOff
	log.Printf("[Interceptor] Mode set to: %s", mode)
}

// GetMode 获取拦截模式
func (i *Interceptor) GetMode() InterceptMode {
	i.mu.RLock()
	defer i.mu.RUnlock()
	return i.interceptMode
}

// IsEnabled 是否启用
func (i *Interceptor) IsEnabled() bool {
	i.mu.RLock()
	defer i.mu.RUnlock()
	return i.enabled
}

// ShouldIntercept 判断是否应该拦截
func (i *Interceptor) ShouldIntercept(method, host, path string) bool {
	i.mu.RLock()
	defer i.mu.RUnlock()

	if !i.enabled {
		return false
	}

	if i.interceptMode == InterceptModeAll {
		return true
	}

	// TODO: 实现过滤模式
	return false
}

// Intercept 拦截请求
func (i *Interceptor) Intercept(req *InterceptedRequest) {
	i.mu.Lock()
	req.Waiting = true
	i.pending[req.ID] = req
	i.mu.Unlock()

	// 广播拦截事件
	i.hub.BroadcastToChannel("intercept", map[string]interface{}{
		"type": "intercepted",
		"data": req,
	})

	log.Printf("[Interceptor] Request intercepted: %s %s", req.Method, req.URL)
}

// Forward 放行请求
func (i *Interceptor) Forward(requestID string, modified *ModifiedRequest) {
	i.mu.Lock()
	delete(i.pending, requestID)
	i.mu.Unlock()

	// 广播放行事件
	i.hub.BroadcastToChannel("intercept", map[string]interface{}{
		"type":      "forwarded",
		"request_id": requestID,
	})

	// 调用回调
	if i.onForward != nil {
		i.onForward(requestID, modified)
	}

	log.Printf("[Interceptor] Request forwarded: %s", requestID)
}

// Drop 丢弃请求
func (i *Interceptor) Drop(requestID string) {
	i.mu.Lock()
	delete(i.pending, requestID)
	i.mu.Unlock()

	// 广播丢弃事件
	i.hub.BroadcastToChannel("intercept", map[string]interface{}{
		"type":       "dropped",
		"request_id": requestID,
	})

	// 调用回调
	if i.onDrop != nil {
		i.onDrop(requestID)
	}

	log.Printf("[Interceptor] Request dropped: %s", requestID)
}

// GetPending 获取待处理请求列表
func (i *Interceptor) GetPending() []*InterceptedRequest {
	i.mu.RLock()
	defer i.mu.RUnlock()

	result := make([]*InterceptedRequest, 0, len(i.pending))
	for _, req := range i.pending {
		result = append(result, req)
	}
	return result
}

// OnForward 设置放行回调
func (i *Interceptor) OnForward(fn func(string, *ModifiedRequest)) {
	i.onForward = fn
}

// OnDrop 设置丢弃回调
func (i *Interceptor) OnDrop(fn func(string)) {
	i.onDrop = fn
}

// HandleClientMessage 处理客户端消息
func (i *Interceptor) HandleClientMessage(client *Client, msg *Message) {
	switch msg.Action {
	case "intercept_enable":
		var mode InterceptMode = InterceptModeAll
		if data, ok := msg.Data.(map[string]interface{}); ok {
			if m, ok := data["mode"].(string); ok {
				mode = InterceptMode(m)
			}
		}
		i.SetMode(mode)
		client.SendJSON(map[string]interface{}{
			"type": "ack",
			"action": "intercept_enabled",
			"mode": mode,
		})

	case "intercept_disable":
		i.SetMode(InterceptModeOff)
		client.SendJSON(map[string]interface{}{
			"type": "ack",
			"action": "intercept_disabled",
		})

	case "forward":
		if data, ok := msg.Data.(map[string]interface{}); ok {
			requestID, _ := data["request_id"].(string)
			modified := &ModifiedRequest{
				ID: requestID,
			}
			if m, ok := data["method"].(string); ok {
				modified.Method = m
			}
			if u, ok := data["url"].(string); ok {
				modified.URL = u
			}
			if h, ok := data["headers"].(map[string]interface{}); ok {
				modified.Headers = make(map[string]string)
				for k, v := range h {
					if vs, ok := v.(string); ok {
						modified.Headers[k] = vs
					}
				}
			}
			if b, ok := data["body"].(string); ok {
				modified.Body = b
			}
			i.Forward(requestID, modified)
		}

	case "drop":
		if data, ok := msg.Data.(map[string]interface{}); ok {
			if requestID, ok := data["request_id"].(string); ok {
				i.Drop(requestID)
			}
		}

	case "get_pending":
		pending := i.GetPending()
		data, _ := json.Marshal(map[string]interface{}{
			"type": "pending_list",
			"data": pending,
		})
		client.Send(data)
	}
}

// ClearPending 清空待处理请求
func (i *Interceptor) ClearPending() {
	i.mu.Lock()
	defer i.mu.Unlock()

	for id := range i.pending {
		i.Drop(id)
	}
}
