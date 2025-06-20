// Package traffic 提供流量处理功能
// Package traffic provides traffic processing functionality
package traffic

import (
	"bufio"
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"

	"hackmitm/pkg/logger"

	"github.com/sirupsen/logrus"
)

// Processor 流量处理器
// Processor traffic processor
type Processor struct {
	// requestHandlers HTTP请求处理器链
	requestHandlers []RequestHandler
	// responseHandlers HTTP响应处理器链
	responseHandlers []ResponseHandler
	// compressionEnabled 启用压缩
	compressionEnabled bool
	// maxBodySize 最大请求/响应体大小
	maxBodySize int64
	// mutex 保护处理器链的互斥锁
	mutex sync.RWMutex
}

// RequestHandler HTTP请求处理器接口
// RequestHandler interface for HTTP request processing
type RequestHandler interface {
	HandleRequest(req *http.Request) error
}

// ResponseHandler HTTP响应处理器接口
// ResponseHandler interface for HTTP response processing
type ResponseHandler interface {
	HandleResponse(resp *http.Response, req *http.Request) error
}

// ProcessorOptions 流量处理器选项
// ProcessorOptions traffic processor options
type ProcessorOptions struct {
	// CompressionEnabled 启用压缩
	CompressionEnabled bool
	// MaxBodySize 最大请求/响应体大小
	MaxBodySize int64
}

// NewProcessor 创建新的流量处理器
// NewProcessor creates a new traffic processor
func NewProcessor(opts ProcessorOptions) *Processor {
	if opts.MaxBodySize == 0 {
		opts.MaxBodySize = 10 * 1024 * 1024 // 默认10MB
	}

	return &Processor{
		requestHandlers:    make([]RequestHandler, 0),
		responseHandlers:   make([]ResponseHandler, 0),
		compressionEnabled: opts.CompressionEnabled,
		maxBodySize:        opts.MaxBodySize,
	}
}

// AddRequestHandler 添加请求处理器
// AddRequestHandler adds a request handler
func (p *Processor) AddRequestHandler(handler RequestHandler) {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	p.requestHandlers = append(p.requestHandlers, handler)
}

// AddResponseHandler 添加响应处理器
// AddResponseHandler adds a response handler
func (p *Processor) AddResponseHandler(handler ResponseHandler) {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	p.responseHandlers = append(p.responseHandlers, handler)
}

// ProcessRequest 处理HTTP请求
// ProcessRequest processes HTTP request
func (p *Processor) ProcessRequest(req *http.Request) error {
	// 限制请求体大小
	if req.ContentLength > p.maxBodySize {
		return fmt.Errorf("请求体过大: %d bytes", req.ContentLength)
	}

	// 执行请求处理器链
	p.mutex.RLock()
	handlers := make([]RequestHandler, len(p.requestHandlers))
	copy(handlers, p.requestHandlers)
	p.mutex.RUnlock()

	for _, handler := range handlers {
		if err := handler.HandleRequest(req); err != nil {
			logger.Errorf("请求处理器执行失败: %v", err)
			return err
		}
	}

	logger.Debugf("请求处理完成: %s %s", req.Method, req.URL.String())
	return nil
}

// ProcessResponse 处理HTTP响应
// ProcessResponse processes HTTP response
func (p *Processor) ProcessResponse(resp *http.Response, req *http.Request) error {
	// 限制响应体大小
	if resp.ContentLength > p.maxBodySize {
		return fmt.Errorf("响应体过大: %d bytes", resp.ContentLength)
	}

	// 执行响应处理器链
	p.mutex.RLock()
	handlers := make([]ResponseHandler, len(p.responseHandlers))
	copy(handlers, p.responseHandlers)
	p.mutex.RUnlock()

	for _, handler := range handlers {
		if err := handler.HandleResponse(resp, req); err != nil {
			logger.Errorf("响应处理器执行失败: %v", err)
			return err
		}
	}

	logger.Debugf("响应处理完成: %s %d", req.URL.String(), resp.StatusCode)
	return nil
}

// LoggingHandler 日志记录处理器
// LoggingHandler logging handler
type LoggingHandler struct{}

// HandleRequest 处理请求日志
// HandleRequest handles request logging
func (h *LoggingHandler) HandleRequest(req *http.Request) error {
	logger.Infof("[REQUEST] %s %s %s", req.Method, req.URL.String(), req.Proto)

	// 记录请求头
	if logger.DefaultLogger.Logger.Level <= logrus.InfoLevel {
		for name, values := range req.Header {
			for _, value := range values {
				logger.Debugf("[REQUEST HEADER] %s: %s", name, value)
			}
		}
	}

	return nil
}

// HandleResponse 处理响应日志
// HandleResponse handles response logging
func (h *LoggingHandler) HandleResponse(resp *http.Response, req *http.Request) error {
	logger.Infof("[RESPONSE] %s %d %s", req.URL.String(), resp.StatusCode, resp.Status)

	// 记录响应头
	if logger.DefaultLogger.Logger.Level <= logrus.InfoLevel {
		for name, values := range resp.Header {
			for _, value := range values {
				logger.Debugf("[RESPONSE HEADER] %s: %s", name, value)
			}
		}
	}

	return nil
}

// HeaderModifierHandler 请求头修改处理器
// HeaderModifierHandler header modification handler
type HeaderModifierHandler struct {
	// AddHeaders 要添加的请求头
	AddHeaders map[string]string
	// RemoveHeaders 要移除的请求头
	RemoveHeaders []string
}

// HandleRequest 处理请求头修改
// HandleRequest handles request header modification
func (h *HeaderModifierHandler) HandleRequest(req *http.Request) error {
	// 添加请求头
	for name, value := range h.AddHeaders {
		req.Header.Set(name, value)
		logger.Debugf("添加请求头: %s: %s", name, value)
	}

	// 移除请求头
	for _, name := range h.RemoveHeaders {
		if req.Header.Get(name) != "" {
			req.Header.Del(name)
			logger.Debugf("移除请求头: %s", name)
		}
	}

	return nil
}

// CompressionHandler 压缩处理器
// CompressionHandler compression handler
type CompressionHandler struct {
	// enabled 是否启用压缩
	enabled bool
}

// NewCompressionHandler 创建压缩处理器
// NewCompressionHandler creates a compression handler
func NewCompressionHandler(enabled bool) *CompressionHandler {
	return &CompressionHandler{enabled: enabled}
}

// HandleResponse 处理响应压缩
// HandleResponse handles response compression
func (h *CompressionHandler) HandleResponse(resp *http.Response, req *http.Request) error {
	if !h.enabled {
		return nil
	}

	// 检查是否需要压缩
	contentType := resp.Header.Get("Content-Type")
	if !h.shouldCompress(contentType) {
		return nil
	}

	// 检查是否已经压缩
	contentEncoding := resp.Header.Get("Content-Encoding")
	if contentEncoding != "" {
		logger.Debugf("响应已压缩，跳过: %s", contentEncoding)
		return nil
	}

	// 读取响应体
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("读取响应体失败: %w", err)
	}
	resp.Body.Close()

	// 压缩响应体
	var compressed bytes.Buffer
	gzWriter := gzip.NewWriter(&compressed)
	if _, err := gzWriter.Write(body); err != nil {
		return fmt.Errorf("压缩响应体失败: %w", err)
	}
	gzWriter.Close()

	// 更新响应
	resp.Body = io.NopCloser(&compressed)
	resp.Header.Set("Content-Encoding", "gzip")
	resp.Header.Set("Content-Length", strconv.Itoa(compressed.Len()))
	resp.ContentLength = int64(compressed.Len())

	logger.Debugf("响应压缩完成: %d -> %d bytes", len(body), compressed.Len())
	return nil
}

// shouldCompress 判断是否应该压缩
// shouldCompress determines if content should be compressed
func (h *CompressionHandler) shouldCompress(contentType string) bool {
	compressibleTypes := []string{
		"text/",
		"application/json",
		"application/javascript",
		"application/xml",
		"application/xhtml+xml",
	}

	contentType = strings.ToLower(contentType)
	for _, t := range compressibleTypes {
		if strings.HasPrefix(contentType, t) {
			return true
		}
	}

	return false
}

// TLSInfoHandler TLS信息处理器
// TLSInfoHandler TLS information handler
type TLSInfoHandler struct{}

// HandleRequest 处理TLS信息记录
// HandleRequest handles TLS information logging
func (h *TLSInfoHandler) HandleRequest(req *http.Request) error {
	if req.TLS != nil {
		logger.Debugf("[TLS] Version: %d, Cipher: %d, ServerName: %s",
			req.TLS.Version, req.TLS.CipherSuite, req.TLS.ServerName)

		if len(req.TLS.PeerCertificates) > 0 {
			cert := req.TLS.PeerCertificates[0]
			logger.Debugf("[TLS CERT] Subject: %s, Issuer: %s",
				cert.Subject.String(), cert.Issuer.String())
		}
	}

	return nil
}

// URLRewriteHandler URL重写处理器
// URLRewriteHandler URL rewrite handler
type URLRewriteHandler struct {
	// rules URL重写规则
	rules []URLRewriteRule
}

// URLRewriteRule URL重写规则
// URLRewriteRule URL rewrite rule
type URLRewriteRule struct {
	// Pattern 匹配模式
	Pattern string
	// Replacement 替换目标
	Replacement string
}

// NewURLRewriteHandler 创建URL重写处理器
// NewURLRewriteHandler creates a URL rewrite handler
func NewURLRewriteHandler(rules []URLRewriteRule) *URLRewriteHandler {
	return &URLRewriteHandler{rules: rules}
}

// HandleRequest 处理URL重写
// HandleRequest handles URL rewriting
func (h *URLRewriteHandler) HandleRequest(req *http.Request) error {
	originalURL := req.URL.String()

	for _, rule := range h.rules {
		if strings.Contains(req.URL.String(), rule.Pattern) {
			newURL := strings.Replace(req.URL.String(), rule.Pattern, rule.Replacement, -1)

			parsedURL, err := url.Parse(newURL)
			if err != nil {
				logger.Errorf("解析重写URL失败: %v", err)
				continue
			}

			req.URL = parsedURL
			logger.Debugf("URL重写: %s -> %s", originalURL, newURL)
			break
		}
	}

	return nil
}

// ParseHTTPRequest 解析HTTP请求
// ParseHTTPRequest parses HTTP request from raw data
func ParseHTTPRequest(data []byte) (*http.Request, error) {
	reader := bufio.NewReader(bytes.NewReader(data))
	req, err := http.ReadRequest(reader)
	if err != nil {
		return nil, fmt.Errorf("解析HTTP请求失败: %w", err)
	}

	return req, nil
}

// ParseHTTPResponse 解析HTTP响应
// ParseHTTPResponse parses HTTP response from raw data
func ParseHTTPResponse(data []byte, req *http.Request) (*http.Response, error) {
	reader := bufio.NewReader(bytes.NewReader(data))
	resp, err := http.ReadResponse(reader, req)
	if err != nil {
		return nil, fmt.Errorf("解析HTTP响应失败: %w", err)
	}

	return resp, nil
}

// SerializeRequest 序列化HTTP请求
// SerializeRequest serializes HTTP request to raw data
func SerializeRequest(req *http.Request) ([]byte, error) {
	var buf bytes.Buffer

	if err := req.Write(&buf); err != nil {
		return nil, fmt.Errorf("序列化HTTP请求失败: %w", err)
	}

	return buf.Bytes(), nil
}

// SerializeResponse 序列化HTTP响应
// SerializeResponse serializes HTTP response to raw data
func SerializeResponse(resp *http.Response) ([]byte, error) {
	var buf bytes.Buffer

	if err := resp.Write(&buf); err != nil {
		return nil, fmt.Errorf("序列化HTTP响应失败: %w", err)
	}

	return buf.Bytes(), nil
}
