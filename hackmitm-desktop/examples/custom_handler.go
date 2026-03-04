// Package main 演示如何创建和使用自定义处理器
// Package main demonstrates how to create and use custom handlers
package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"hackmitm/pkg/cert"
	"hackmitm/pkg/config"
	"hackmitm/pkg/logger"
	"hackmitm/pkg/proxy"
	"hackmitm/pkg/traffic"
)

// AuthenticationHandler 认证处理器示例
// AuthenticationHandler example authentication handler
type AuthenticationHandler struct {
	apiKey string
	users  map[string]string
}

// HandleRequest 处理请求认证
// HandleRequest handles request authentication
func (h *AuthenticationHandler) HandleRequest(req *http.Request) error {
	// 添加 API Key
	if h.apiKey != "" {
		req.Header.Set("X-API-Key", h.apiKey)
	}

	// 基本认证示例
	if username, password, ok := req.BasicAuth(); ok {
		if storedPassword, exists := h.users[username]; exists && storedPassword == password {
			req.Header.Set("X-Authenticated-User", username)
			logger.Infof("用户认证成功: %s", username)
		} else {
			logger.Warnf("用户认证失败: %s", username)
		}
	}

	return nil
}

// RequestLoggingHandler 请求日志处理器
// RequestLoggingHandler request logging handler
type RequestLoggingHandler struct {
	logFile string
}

// HandleRequest 记录详细的请求信息
// HandleRequest logs detailed request information
func (h *RequestLoggingHandler) HandleRequest(req *http.Request) error {
	logData := map[string]interface{}{
		"timestamp":   time.Now().Format(time.RFC3339),
		"method":      req.Method,
		"url":         req.URL.String(),
		"host":        req.Host,
		"remote_addr": req.RemoteAddr,
		"user_agent":  req.Header.Get("User-Agent"),
		"referer":     req.Header.Get("Referer"),
		"headers":     req.Header,
	}

	// 记录请求体（仅用于演示，生产环境需要谨慎处理）
	if req.ContentLength > 0 && req.ContentLength < 1024 { // 限制大小
		body, err := io.ReadAll(req.Body)
		if err == nil {
			req.Body = io.NopCloser(strings.NewReader(string(body)))
			logData["body"] = string(body)
		}
	}

	logJSON, _ := json.MarshalIndent(logData, "", "  ")
	logger.Infof("请求详情: %s", string(logJSON))

	return nil
}

// ContentFilterHandler 内容过滤处理器
// ContentFilterHandler content filtering handler
type ContentFilterHandler struct {
	blockedDomains  []string
	blockedKeywords []string
}

// HandleRequest 过滤请求
// HandleRequest filters requests
func (h *ContentFilterHandler) HandleRequest(req *http.Request) error {
	// 检查被阻止的域名
	for _, domain := range h.blockedDomains {
		if strings.Contains(req.Host, domain) {
			return fmt.Errorf("域名被阻止: %s", domain)
		}
	}

	// 检查 URL 中的关键词
	for _, keyword := range h.blockedKeywords {
		if strings.Contains(req.URL.String(), keyword) {
			logger.Warnf("检测到被阻止的关键词: %s in %s", keyword, req.URL.String())
			return fmt.Errorf("URL 包含被阻止的关键词: %s", keyword)
		}
	}

	return nil
}

// HandleResponse 过滤响应内容
// HandleResponse filters response content
func (h *ContentFilterHandler) HandleResponse(resp *http.Response, req *http.Request) error {
	// 仅处理 HTML 内容
	contentType := resp.Header.Get("Content-Type")
	if !strings.Contains(contentType, "text/html") {
		return nil
	}

	// 读取响应体
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	resp.Body.Close()

	// 过滤关键词
	content := string(body)
	for _, keyword := range h.blockedKeywords {
		content = strings.ReplaceAll(content, keyword, "***")
	}

	// 更新响应体
	resp.Body = io.NopCloser(strings.NewReader(content))
	resp.ContentLength = int64(len(content))

	return nil
}

// PerformanceMonitorHandler 性能监控处理器
// PerformanceMonitorHandler performance monitoring handler
type PerformanceMonitorHandler struct {
	requestTimes map[string]time.Time
}

// HandleRequest 记录请求开始时间
// HandleRequest records request start time
func (h *PerformanceMonitorHandler) HandleRequest(req *http.Request) error {
	requestID := fmt.Sprintf("%s-%d", req.RemoteAddr, time.Now().UnixNano())
	req.Header.Set("X-Request-ID", requestID)

	if h.requestTimes == nil {
		h.requestTimes = make(map[string]time.Time)
	}
	h.requestTimes[requestID] = time.Now()

	return nil
}

// HandleResponse 计算响应时间
// HandleResponse calculates response time
func (h *PerformanceMonitorHandler) HandleResponse(resp *http.Response, req *http.Request) error {
	requestID := req.Header.Get("X-Request-ID")
	if startTime, exists := h.requestTimes[requestID]; exists {
		duration := time.Since(startTime)
		logger.Infof("请求 %s 处理时间: %v", req.URL.String(), duration)

		// 添加响应时间头
		resp.Header.Set("X-Response-Time", duration.String())

		// 清理记录
		delete(h.requestTimes, requestID)
	}

	return nil
}

// CacheHandler 简单缓存处理器
// CacheHandler simple cache handler
type CacheHandler struct {
	cache map[string]CacheEntry
}

// CacheEntry 缓存条目
// CacheEntry cache entry
type CacheEntry struct {
	Data      []byte
	Headers   http.Header
	Status    int
	Timestamp time.Time
	TTL       time.Duration
}

// HandleRequest 检查缓存
// HandleRequest checks cache
func (h *CacheHandler) HandleRequest(req *http.Request) error {
	// 仅对 GET 请求启用缓存
	if req.Method != http.MethodGet {
		return nil
	}

	if h.cache == nil {
		h.cache = make(map[string]CacheEntry)
	}

	cacheKey := req.URL.String()
	if entry, exists := h.cache[cacheKey]; exists {
		if time.Since(entry.Timestamp) < entry.TTL {
			logger.Debugf("缓存命中: %s", cacheKey)
			// 这里应该直接返回缓存的响应，但需要更复杂的实现
		} else {
			// 缓存过期，删除条目
			delete(h.cache, cacheKey)
		}
	}

	return nil
}

func main() {
	logger.Info("启动 HackMITM 自定义处理器示例")

	// 加载配置
	cfg, err := config.LoadConfig("../configs/config.json")
	if err != nil {
		logger.Fatalf("加载配置失败: %v", err)
	}

	// 初始化证书管理器
	certMgr, err := cert.NewCertManager(cert.CertOptions{
		CertDir:     cfg.GetTLS().CertDir,
		EnableCache: cfg.GetTLS().EnableCertCache,
		CacheTTL:    cfg.GetTLS().CertCacheTTL,
	})
	if err != nil {
		logger.Fatalf("初始化证书管理器失败: %v", err)
	}

	// 创建代理服务器
	server, err := proxy.NewServer(cfg, certMgr)
	if err != nil {
		logger.Fatalf("创建代理服务器失败: %v", err)
	}

	// 添加自定义处理器
	setupCustomHandlers(server)

	// 启动服务器
	if err := server.Start(); err != nil {
		logger.Fatalf("启动服务器失败: %v", err)
	}

	logger.Info("服务器已启动，按 Ctrl+C 退出")

	// 等待中断信号
	select {}
}

// setupCustomHandlers 设置自定义处理器
// setupCustomHandlers sets up custom handlers
func setupCustomHandlers(server *proxy.Server) {
	// 1. 认证处理器
	authHandler := &AuthenticationHandler{
		apiKey: "your-api-key-here",
		users: map[string]string{
			"admin": "password123",
			"user":  "userpass",
		},
	}
	server.AddRequestHandler(authHandler)

	// 2. 请求日志处理器
	requestLogger := &RequestLoggingHandler{
		logFile: "requests.log",
	}
	server.AddRequestHandler(requestLogger)

	// 3. 内容过滤处理器
	contentFilter := &ContentFilterHandler{
		blockedDomains: []string{
			"malicious.com",
			"blocked-site.net",
		},
		blockedKeywords: []string{
			"malware",
			"virus",
			"spam",
		},
	}
	server.AddRequestHandler(contentFilter)
	server.AddResponseHandler(contentFilter)

	// 4. 性能监控处理器
	perfMonitor := &PerformanceMonitorHandler{}
	server.AddRequestHandler(perfMonitor)
	server.AddResponseHandler(perfMonitor)

	// 5. 缓存处理器
	cacheHandler := &CacheHandler{}
	server.AddRequestHandler(cacheHandler)

	// 6. 添加自定义请求头
	headerModifier := &traffic.HeaderModifierHandler{
		AddHeaders: map[string]string{
			"X-Proxy-Agent":   "HackMITM-Custom/1.0",
			"X-Custom-Header": "Custom-Value",
			"X-Request-Time":  time.Now().Format(time.RFC3339),
		},
		RemoveHeaders: []string{
			"X-Forwarded-For",
			"X-Real-IP",
		},
	}
	server.AddRequestHandler(headerModifier)

	logger.Info("自定义处理器设置完成")
	logger.Info("包含以下处理器:")
	logger.Info("  - 认证处理器")
	logger.Info("  - 请求日志处理器")
	logger.Info("  - 内容过滤处理器")
	logger.Info("  - 性能监控处理器")
	logger.Info("  - 缓存处理器")
	logger.Info("  - 请求头修改处理器")
}
