// Package proxy 提供代理服务器功能
// Package proxy provides proxy server functionality
package proxy

import (
	"bufio"
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"sync"
	"time"

	"hackmitm/pkg/cert"
	"hackmitm/pkg/config"
	"hackmitm/pkg/logger"
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
	// httpServer HTTP服务器
	httpServer *http.Server
	// client HTTP客户端
	client *http.Client
	// connPool 连接池
	connPool sync.Pool
	// activeConns 活跃连接数
	activeConns int64
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

	// 创建HTTP客户端
	client := &http.Client{
		Timeout: cfg.GetProxy().UpstreamTimeout,
		Transport: &http.Transport{
			MaxIdleConns:        cfg.GetProxy().MaxIdleConns,
			MaxIdleConnsPerHost: 20,
			IdleConnTimeout:     90 * time.Second,
			TLSHandshakeTimeout: 10 * time.Second,
			DisableCompression:  !cfg.GetProxy().EnableCompression,
		},
	}

	server := &Server{
		config:      cfg,
		certManager: certMgr,
		processor:   processor,
		client:      client,
		ctx:         ctx,
		cancel:      cancel,
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

	go func() {
		if err := s.httpServer.Serve(listener); err != nil && err != http.ErrServerClosed {
			logger.Errorf("HTTP服务器运行失败: %v", err)
		}
	}()

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

	logger.Info("代理服务器已停止")
	return nil
}

// ServeHTTP 实现http.Handler接口
// ServeHTTP implements http.Handler interface
func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// 增加活跃连接数
	s.mutex.Lock()
	s.activeConns++
	s.mutex.Unlock()

	defer func() {
		s.mutex.Lock()
		s.activeConns--
		s.mutex.Unlock()
	}()

	// 处理CONNECT方法（HTTPS代理）
	if r.Method == http.MethodConnect {
		s.handleConnect(w, r)
		return
	}

	// 处理HTTP代理
	s.handleHTTP(w, r)
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

	for {
		// 读取客户端请求
		reader := bufio.NewReader(clientConn)
		req, err := http.ReadRequest(reader)
		if err != nil {
			if err != io.EOF {
				logger.Errorf("读取HTTPS请求失败: %v", err)
			}
			return
		}

		// 设置请求URL
		req.URL.Scheme = "https"
		req.URL.Host = targetHost

		// 处理请求
		if err := s.processor.ProcessRequest(req); err != nil {
			logger.Errorf("处理HTTPS请求失败: %v", err)
			return
		}

		// 转发请求到目标服务器
		resp, err := s.client.Do(req)
		if err != nil {
			logger.Errorf("转发HTTPS请求失败: %v", err)
			return
		}

		// 处理响应
		if err := s.processor.ProcessResponse(resp, req); err != nil {
			logger.Errorf("处理HTTPS响应失败: %v", err)
			resp.Body.Close()
			return
		}

		// 发送响应到客户端
		if err := resp.Write(clientConn); err != nil {
			logger.Errorf("发送HTTPS响应失败: %v", err)
			resp.Body.Close()
			return
		}

		resp.Body.Close()

		// 检查是否需要保持连接
		if req.Header.Get("Connection") == "close" ||
			resp.Header.Get("Connection") == "close" {
			return
		}
	}
}

// handleHTTP 处理HTTP请求
// handleHTTP handles HTTP requests
func (s *Server) handleHTTP(w http.ResponseWriter, r *http.Request) {
	logger.Debugf("处理HTTP请求: %s %s", r.Method, r.URL.String())

	// 构建完整URL
	if r.URL.Scheme == "" {
		r.URL.Scheme = "http"
	}
	if r.URL.Host == "" {
		r.URL.Host = r.Host
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

	// 转发请求到目标服务器
	resp, err := s.client.Do(newReq)
	if err != nil {
		logger.Errorf("转发HTTP请求失败: %v", err)
		http.Error(w, "转发请求失败", http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

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

// GetStats 获取服务器统计信息
// GetStats returns server statistics
func (s *Server) GetStats() map[string]interface{} {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	return map[string]interface{}{
		"active_connections": s.activeConns,
		"cert_cache_stats":   s.certManager.GetCacheStats(),
	}
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
