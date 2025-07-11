// Package proxy 提供代理服务器功能
// Package proxy provides proxy server functionality
package proxy

import (
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"hackmitm/pkg/cert"
	"hackmitm/pkg/config"
	"hackmitm/pkg/logger"
	"hackmitm/pkg/plugin"
	"hackmitm/pkg/security"
	"hackmitm/pkg/traffic"
)

// Server 代理服务器
// Server proxy server
type Server struct {
	// config 配置
	config *config.Config
	// certManager 证书管理器
	certManager *cert.CertManager
	// processor 流量处理器
	processor *traffic.Processor
	// accessController 访问控制器
	accessController *security.AccessController
	// pluginManager 插件管理器
	pluginManager *plugin.Manager
	// httpServer HTTP服务器
	httpServer *http.Server
	// client HTTP客户端
	client *http.Client
	// connPool 连接池
	connPool sync.Pool
	// activeConns 活跃连接数
	activeConns int64
	// totalRequests 总请求数
	totalRequests int64
	// startTime 启动时间
	startTime time.Time
	// mutex 互斥锁
	mutex sync.RWMutex
	// ctx 上下文
	ctx context.Context
	// cancel 取消函数
	cancel context.CancelFunc
}

// NewServer 创建新的代理服务器
// NewServer creates a new proxy server
func NewServer(cfg *config.Config, certMgr *cert.CertManager) (*Server, error) {
	ctx, cancel := context.WithCancel(context.Background())

	// 创建流量处理器
	processor := traffic.NewProcessor(traffic.ProcessorOptions{
		CompressionEnabled: cfg.GetProxy().EnableCompression,
		MaxBodySize:        10 * 1024 * 1024, // 10MB
	})

	// 添加默认处理器
	processor.AddRequestHandler(&traffic.LoggingHandler{})
	processor.AddRequestHandler(&traffic.TLSInfoHandler{})
	processor.AddResponseHandler(&traffic.LoggingHandler{})

	if cfg.GetProxy().EnableCompression {
		processor.AddResponseHandler(traffic.NewCompressionHandler(true))
	}

	// 创建访问控制器
	accessController := security.NewAccessController()

	// 创建插件管理器
	pluginManager := plugin.NewManager("./plugins")

	// 创建HTTP客户端
	transport := &http.Transport{
		MaxIdleConns:        cfg.GetProxy().MaxIdleConns,
		MaxIdleConnsPerHost: 20,
		IdleConnTimeout:     90 * time.Second,
		TLSHandshakeTimeout: 10 * time.Second,
		DisableCompression:  !cfg.GetProxy().EnableCompression,
		// 增加更多性能优化
		WriteBufferSize:   4096,
		ReadBufferSize:    4096,
		MaxConnsPerHost:   0, // 无限制
		DisableKeepAlives: false,
	}

	client := &http.Client{
		Timeout:   cfg.GetProxy().UpstreamTimeout,
		Transport: transport,
		// 禁用自动重定向以便代理处理
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	server := &Server{
		config:           cfg,
		certManager:      certMgr,
		processor:        processor,
		accessController: accessController,
		pluginManager:    pluginManager,
		client:           client,
		startTime:        time.Now(),
		ctx:              ctx,
		cancel:           cancel,
	}

	// 初始化连接池
	server.connPool.New = func() interface{} {
		return make([]byte, 32*1024) // 32KB缓冲区
	}

	return server, nil
}

// Start 启动代理服务器
// Start starts the proxy server
func (s *Server) Start() error {
	serverConfig := s.config.GetServer()

	// 创建HTTP服务器
	s.httpServer = &http.Server{
		Addr:         fmt.Sprintf("%s:%d", serverConfig.ListenAddr, serverConfig.ListenPort),
		Handler:      s,
		ReadTimeout:  serverConfig.ReadTimeout,
		WriteTimeout: serverConfig.WriteTimeout,
		TLSConfig: &tls.Config{
			GetCertificate: s.getCertificate,
			NextProtos:     []string{"h2", "http/1.1"},
		},
	}

	logger.Infof("代理服务器启动，监听地址: %s", s.httpServer.Addr)

	// 启动HTTP服务器
	listener, err := net.Listen("tcp", s.httpServer.Addr)
	if err != nil {
		return fmt.Errorf("创建监听器失败: %w", err)
	}

	// 创建启动完成通道
	started := make(chan error, 1)

	go func() {
		if err := s.httpServer.Serve(listener); err != nil && err != http.ErrServerClosed {
			logger.Errorf("HTTP服务器运行失败: %v", err)
			started <- err
		}
		close(started)
	}()

	// 等待服务器启动
	time.Sleep(100 * time.Millisecond)
	// 检查端口是否真正在监听
	for i := 0; i < 3; i++ { // 重试3次
		conn, err := net.DialTimeout("tcp", s.httpServer.Addr, time.Second)
		if err == nil {
			conn.Close()
			logger.Info("代理服务器启动完成")
			return nil
		}
		time.Sleep(100 * time.Millisecond)
	}

	// 检查是否有错误发生
	select {
	case err := <-started:
		if err != nil {
			return fmt.Errorf("启动服务器失败: %w", err)
		}
	default:
		return fmt.Errorf("服务器端口未就绪")
	}

	return nil
}

// Stop 停止代理服务器
// Stop stops the proxy server
func (s *Server) Stop() error {
	logger.Info("正在停止代理服务器...")

	s.cancel()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if s.httpServer != nil {
		if err := s.httpServer.Shutdown(ctx); err != nil {
			logger.Errorf("停止HTTP服务器失败: %v", err)
			return err
		}
	}

	// 关闭插件管理器
	if s.pluginManager != nil {
		if err := s.pluginManager.Shutdown(); err != nil {
			logger.Errorf("关闭插件管理器失败: %v", err)
		}
	}

	logger.Info("代理服务器已停止")
	return nil
}

// ServeHTTP 实现http.Handler接口
// ServeHTTP implements http.Handler interface
func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// 增加活跃连接数和总请求数
	atomic.AddInt64(&s.activeConns, 1)
	atomic.AddInt64(&s.totalRequests, 1)

	defer func() {
		atomic.AddInt64(&s.activeConns, -1)
	}()

	// 访问控制检查
	if err := s.accessController.IsAllowed(r); err != nil {
		logger.Warnf("访问被拒绝: %v", err)
		http.Error(w, "访问被拒绝", http.StatusForbidden)
		return
	}

	// 插件过滤检查
	if allowed, err := s.checkPluginFilters(r); err != nil {
		logger.Errorf("插件过滤检查失败: %v", err)
		http.Error(w, "内部错误", http.StatusInternalServerError)
		return
	} else if !allowed {
		logger.Warnf("请求被插件过滤器阻止: %s %s", r.Method, r.URL.String())
		http.Error(w, "请求被阻止", http.StatusForbidden)
		return
	}

	// 检查WebSocket升级
	if s.isWebSocketUpgrade(r) {
		s.handleWebSocket(w, r)
		return
	}

	// 处理CONNECT方法（HTTPS代理）
	if r.Method == http.MethodConnect {
		s.handleConnect(w, r)
		return
	}

	// 处理HTTP代理
	s.handleHTTP(w, r)
}

// isWebSocketUpgrade 检查是否为WebSocket升级请求
func (s *Server) isWebSocketUpgrade(r *http.Request) bool {
	return strings.ToLower(r.Header.Get("Connection")) == "upgrade" &&
		strings.ToLower(r.Header.Get("Upgrade")) == "websocket"
}

// handleWebSocket 处理WebSocket连接
func (s *Server) handleWebSocket(w http.ResponseWriter, r *http.Request) {
	logger.Debugf("处理WebSocket请求: %s", r.URL.String())

	// 劫持连接
	hijacker, ok := w.(http.Hijacker)
	if !ok {
		http.Error(w, "不支持WebSocket", http.StatusInternalServerError)
		return
	}

	clientConn, _, err := hijacker.Hijack()
	if err != nil {
		logger.Errorf("劫持WebSocket连接失败: %v", err)
		return
	}
	defer clientConn.Close()

	// 连接到目标服务器
	targetURL := *r.URL
	if targetURL.Scheme == "" {
		targetURL.Scheme = "ws"
		if r.TLS != nil {
			targetURL.Scheme = "wss"
		}
	}
	if targetURL.Host == "" {
		targetURL.Host = r.Host
	}

	// 创建到目标的连接
	serverConn, err := net.Dial("tcp", targetURL.Host)
	if err != nil {
		logger.Errorf("连接WebSocket目标失败: %v", err)
		return
	}
	defer serverConn.Close()

	// 转发初始HTTP请求
	if err := r.Write(serverConn); err != nil {
		logger.Errorf("转发WebSocket握手失败: %v", err)
		return
	}

	// 双向代理WebSocket流量
	go s.proxyWebSocketData(clientConn, serverConn, "client->server")
	s.proxyWebSocketData(serverConn, clientConn, "server->client")
}

// proxyWebSocketData 代理WebSocket数据
func (s *Server) proxyWebSocketData(src, dst net.Conn, direction string) {
	buffer := s.connPool.Get().([]byte)
	defer s.connPool.Put(buffer)

	_, err := io.CopyBuffer(dst, src, buffer)
	if err != nil && !strings.Contains(err.Error(), "use of closed network connection") {
		logger.Debugf("WebSocket数据传输结束 (%s): %v", direction, err)
	}
}

// handleConnect 处理CONNECT请求（HTTPS代理）
// handleConnect handles CONNECT requests (HTTPS proxy)
func (s *Server) handleConnect(w http.ResponseWriter, r *http.Request) {
	logger.Debugf("处理CONNECT请求: %s", r.Host)

	// 劫持连接
	hijacker, ok := w.(http.Hijacker)
	if !ok {
		http.Error(w, "不支持连接劫持", http.StatusInternalServerError)
		return
	}

	clientConn, _, err := hijacker.Hijack()
	if err != nil {
		logger.Errorf("劫持连接失败: %v", err)
		return
	}
	defer clientConn.Close()

	// 发送连接建立响应
	_, err = clientConn.Write([]byte("HTTP/1.1 200 Connection Established\r\n\r\n"))
	if err != nil {
		logger.Errorf("发送CONNECT响应失败: %v", err)
		return
	}

	// 获取目标主机名
	host, _, err := net.SplitHostPort(r.Host)
	if err != nil {
		host = r.Host
	}

	// 获取证书
	certificate, err := s.certManager.GetCertificate(host)
	if err != nil {
		logger.Errorf("获取证书失败: %v", err)
		return
	}

	// 创建TLS配置
	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{*certificate},
		ServerName:   host,
	}

	// 升级到TLS连接
	tlsConn := tls.Server(clientConn, tlsConfig)
	if err := tlsConn.Handshake(); err != nil {
		logger.Errorf("TLS握手失败: %v", err)
		return
	}

	// 处理HTTPS流量
	s.handleHTTPS(tlsConn, r.Host)
}

// handleHTTPS 处理HTTPS流量
// handleHTTPS handles HTTPS traffic
func (s *Server) handleHTTPS(clientConn *tls.Conn, targetHost string) {
	defer clientConn.Close()

	logger.Debugf("开始处理HTTPS连接: %s", targetHost)

	// 创建到目标服务器的连接
	logger.Debugf("正在连接目标服务器: %s", targetHost)
	targetConn, err := tls.Dial("tcp", targetHost, &tls.Config{
		InsecureSkipVerify: true,                              // 忽略目标服务器证书验证
		ServerName:         strings.Split(targetHost, ":")[0], // 设置SNI
	})
	if err != nil {
		logger.Errorf("连接目标服务器失败: %v", err)
		return
	}
	defer targetConn.Close()
	logger.Debugf("成功连接到目标服务器: %s", targetHost)

	// 创建双向数据拷贝的通道
	errChan := make(chan error, 2)

	// 客户端到服务器
	go func() {
		buffer := s.connPool.Get().([]byte)
		defer s.connPool.Put(buffer)
		logger.Debugf("开始转发客户端到服务器的数据: %s", targetHost)
		n, err := io.CopyBuffer(targetConn, clientConn, buffer)
		logger.Debugf("客户端到服务器数据传输完成: %s, 传输字节: %d", targetHost, n)
		errChan <- err
	}()

	// 服务器到客户端
	go func() {
		buffer := s.connPool.Get().([]byte)
		defer s.connPool.Put(buffer)
		logger.Debugf("开始转发服务器到客户端的数据: %s", targetHost)
		n, err := io.CopyBuffer(clientConn, targetConn, buffer)
		logger.Debugf("服务器到客户端数据传输完成: %s, 传输字节: %d", targetHost, n)
		errChan <- err
	}()

	// 等待任一方向出错或连接关闭
	for i := 0; i < 2; i++ {
		if err := <-errChan; err != nil && err != io.EOF {
			logger.Debugf("HTTPS数据传输结束: %v", err)
		}
	}

	logger.Debugf("HTTPS连接处理完成: %s", targetHost)
}

// handleHTTP 处理HTTP请求（增强版，集成插件）
func (s *Server) handleHTTP(w http.ResponseWriter, r *http.Request) {
	startTime := time.Now()
	logger.Debugf("处理HTTP请求: %s %s", r.Method, r.URL.String())

	// 构建完整URL
	if r.URL.Scheme == "" {
		r.URL.Scheme = "http"
	}
	if r.URL.Host == "" {
		r.URL.Host = r.Host
	}

	// 创建请求上下文
	requestCtx := s.buildRequestContext(r, startTime)

	// 处理请求插件链
	if err := s.processRequestPlugins(r, requestCtx); err != nil {
		logger.Errorf("请求插件处理失败: %v", err)
		http.Error(w, "请求处理失败", http.StatusInternalServerError)
		return
	}

	// 处理请求
	if err := s.processor.ProcessRequest(r); err != nil {
		logger.Errorf("处理HTTP请求失败: %v", err)
		http.Error(w, "请求处理失败", http.StatusInternalServerError)
		return
	}

	// 创建新的请求（避免修改原始请求）
	newReq := r.Clone(r.Context())
	newReq.RequestURI = ""

	// 设置超时上下文
	ctx, cancel := context.WithTimeout(r.Context(), s.config.GetProxy().UpstreamTimeout)
	defer cancel()
	newReq = newReq.WithContext(ctx)

	// 转发请求到目标服务器
	resp, err := s.client.Do(newReq)
	if err != nil {
		logger.Errorf("转发HTTP请求失败: %v", err)
		if strings.Contains(err.Error(), "timeout") {
			http.Error(w, "请求超时", http.StatusGatewayTimeout)
		} else {
			http.Error(w, "转发请求失败", http.StatusBadGateway)
		}
		return
	}
	defer resp.Body.Close()

	// 创建响应上下文
	responseCtx := s.buildResponseContext(resp, time.Since(startTime))

	// 处理响应插件链
	if err := s.processResponsePlugins(resp, r, responseCtx); err != nil {
		logger.Errorf("响应插件处理失败: %v", err)
		http.Error(w, "响应处理失败", http.StatusInternalServerError)
		return
	}

	// 处理响应
	if err := s.processor.ProcessResponse(resp, r); err != nil {
		logger.Errorf("处理HTTP响应失败: %v", err)
		http.Error(w, "响应处理失败", http.StatusInternalServerError)
		return
	}

	// 复制响应头
	for name, values := range resp.Header {
		for _, value := range values {
			w.Header().Add(name, value)
		}
	}

	// 设置状态码
	w.WriteHeader(resp.StatusCode)

	// 复制响应体
	buffer := s.connPool.Get().([]byte)
	defer s.connPool.Put(buffer)

	if _, err := io.CopyBuffer(w, resp.Body, buffer); err != nil {
		logger.Errorf("复制HTTP响应体失败: %v", err)
	}
}

// buildRequestContext 构建请求上下文
func (s *Server) buildRequestContext(r *http.Request, startTime time.Time) *plugin.RequestContext {
	// 读取请求体（如果有）
	var body []byte
	if r.Body != nil {
		if data, err := io.ReadAll(r.Body); err == nil {
			body = data
			// 重新设置请求体
			r.Body = io.NopCloser(strings.NewReader(string(data)))
		}
	}

	// 构建请求头映射
	headers := make(map[string]string)
	for name, values := range r.Header {
		if len(values) > 0 {
			headers[name] = values[0] // 取第一个值
		}
	}

	return &plugin.RequestContext{
		StartTime: startTime,
		ClientIP:  s.getClientIP(r),
		UserAgent: r.UserAgent(),
		Method:    r.Method,
		URL:       r.URL.String(),
		Headers:   headers,
		Body:      body,
		Metadata:  make(map[string]interface{}),
	}
}

// buildResponseContext 构建响应上下文
func (s *Server) buildResponseContext(resp *http.Response, duration time.Duration) *plugin.ResponseContext {
	// 构建响应头映射
	headers := make(map[string]string)
	for name, values := range resp.Header {
		if len(values) > 0 {
			headers[name] = values[0] // 取第一个值
		}
	}

	return &plugin.ResponseContext{
		StatusCode: resp.StatusCode,
		Headers:    headers,
		Body:       nil, // 响应体在流式处理中，这里为空
		Size:       resp.ContentLength,
		Duration:   duration,
		Metadata:   make(map[string]interface{}),
	}
}

// processRequestPlugins 处理请求插件链
func (s *Server) processRequestPlugins(r *http.Request, ctx *plugin.RequestContext) error {
	if s.pluginManager == nil {
		return nil
	}

	return s.pluginManager.ProcessRequest(r, ctx)
}

// processResponsePlugins 处理响应插件链
func (s *Server) processResponsePlugins(resp *http.Response, req *http.Request, ctx *plugin.ResponseContext) error {
	if s.pluginManager == nil {
		return nil
	}

	return s.pluginManager.ProcessResponse(resp, req, ctx)
}

// getCertificate 获取TLS证书
// getCertificate gets TLS certificate
func (s *Server) getCertificate(hello *tls.ClientHelloInfo) (*tls.Certificate, error) {
	return s.certManager.GetCertificate(hello.ServerName)
}

// AddRequestHandler 添加请求处理器
// AddRequestHandler adds a request handler
func (s *Server) AddRequestHandler(handler traffic.RequestHandler) {
	s.processor.AddRequestHandler(handler)
}

// AddResponseHandler 添加响应处理器
// AddResponseHandler adds a response handler
func (s *Server) AddResponseHandler(handler traffic.ResponseHandler) {
	s.processor.AddResponseHandler(handler)
}

// LoadPlugins 加载插件
func (s *Server) LoadPlugins(configs []*plugin.PluginConfig) error {
	if s.pluginManager == nil {
		return fmt.Errorf("插件管理器未初始化")
	}

	var errors []string
	for _, config := range configs {
		if err := s.pluginManager.LoadPlugin(config); err != nil {
			errors = append(errors, fmt.Sprintf("%s: %v", config.Name, err))
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("加载插件失败: %s", strings.Join(errors, "; "))
	}

	return nil
}

// StartPlugins 启动所有插件
func (s *Server) StartPlugins() error {
	if s.pluginManager == nil {
		return nil
	}

	return s.pluginManager.StartAll()
}

// GetPluginManager 获取插件管理器
func (s *Server) GetPluginManager() *plugin.Manager {
	return s.pluginManager
}

// GetStats 获取服务器统计信息（增强版）
func (s *Server) GetStats() map[string]interface{} {
	uptime := time.Since(s.startTime)

	stats := map[string]interface{}{
		"uptime":              uptime.String(),
		"active_connections":  atomic.LoadInt64(&s.activeConns),
		"total_requests":      atomic.LoadInt64(&s.totalRequests),
		"requests_per_second": float64(atomic.LoadInt64(&s.totalRequests)) / uptime.Seconds(),
		"cert_cache_stats":    s.certManager.GetCacheStats(),
		"access_control":      s.accessController.GetStats(),
		"start_time":          s.startTime.Format(time.RFC3339),
	}

	// 添加插件统计信息
	if s.pluginManager != nil {
		stats["plugins"] = s.pluginManager.GetStats()
	}

	return stats
}

// SetUpstreamProxy 设置上游代理
// SetUpstreamProxy sets upstream proxy
func (s *Server) SetUpstreamProxy(proxyURL string) error {
	if proxyURL == "" {
		return nil
	}

	parsedURL, err := url.Parse(proxyURL)
	if err != nil {
		return fmt.Errorf("解析上游代理URL失败: %w", err)
	}

	transport := s.client.Transport.(*http.Transport)
	transport.Proxy = http.ProxyURL(parsedURL)

	logger.Infof("设置上游代理: %s", proxyURL)
	return nil
}

// SetAccessControl 设置访问控制
func (s *Server) SetAccessControl(username, password string) {
	s.accessController.SetAuth(username, password)
}

// AddToWhitelist 添加IP到白名单
func (s *Server) AddToWhitelist(ip string) {
	s.accessController.AddToWhitelist(ip)
}

// AddToBlacklist 添加IP到黑名单
func (s *Server) AddToBlacklist(ip string) {
	s.accessController.AddToBlacklist(ip)
}

// checkPluginFilters 检查插件过滤器
func (s *Server) checkPluginFilters(r *http.Request) (bool, error) {
	if s.pluginManager == nil {
		return true, nil
	}

	clientIP := s.getClientIP(r)
	filterCtx := &plugin.FilterContext{
		ClientIP:     clientIP,
		UserAgent:    r.UserAgent(),
		RequestCount: atomic.LoadInt64(&s.totalRequests),
		LastRequest:  time.Now(),
		Metadata:     make(map[string]interface{}),
	}

	return s.pluginManager.ShouldAllow(r, filterCtx)
}

// getClientIP 获取客户端IP
func (s *Server) getClientIP(r *http.Request) string {
	// 检查X-Forwarded-For头
	xff := r.Header.Get("X-Forwarded-For")
	if xff != "" {
		ips := strings.Split(xff, ",")
		return strings.TrimSpace(ips[0])
	}

	// 检查X-Real-IP头
	xri := r.Header.Get("X-Real-IP")
	if xri != "" {
		return strings.TrimSpace(xri)
	}

	// 使用RemoteAddr
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}
