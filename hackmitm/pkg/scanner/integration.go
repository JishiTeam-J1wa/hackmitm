package scanner

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"sync"
	"time"

	"hackmitm/pkg/logger"
)

// ScannerHandler 扫描器处理器，集成到流量处理链
// ScannerHandler integrates the scanner into traffic processing chain
type ScannerHandler struct {
	scanner   *PipelineScanner
	vulnStore VulnStore
	enabled   bool
	mu        sync.RWMutex
}

// VulnStore 漏洞存储接口
type VulnStore interface {
	SaveVulnerability(vuln interface{}) error
}

// NewScannerHandler 创建扫描器处理器
func NewScannerHandler(scanner *PipelineScanner) *ScannerHandler {
	return &ScannerHandler{
		scanner: scanner,
		enabled: true,
	}
}

// SetVulnStore 设置漏洞存储
func (h *ScannerHandler) SetVulnStore(store VulnStore) {
	h.vulnStore = store
}

// Enable 启用扫描
func (h *ScannerHandler) Enable() {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.enabled = true
}

// Disable 禁用扫描
func (h *ScannerHandler) Disable() {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.enabled = false
}

// IsEnabled 检查是否启用
func (h *ScannerHandler) IsEnabled() bool {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.enabled
}

// HandleResponse 处理响应，触发被动扫描
// 实现 traffic.ResponseHandler 接口
func (h *ScannerHandler) HandleResponse(resp *http.Response, req *http.Request) error {
	if !h.IsEnabled() {
		return nil
	}

	// 转换为扫描流量格式
	traffic := h.convertToTraffic(req, resp)

	// 异步提交扫描任务
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		_, err := h.scanner.Scan(ctx, traffic)
		if err != nil {
			logger.Debugf("Scan task error: %v", err)
		}
	}()

	return nil
}

// HandleRequest 处理请求（仅扫描请求）
func (h *ScannerHandler) HandleRequest(req *http.Request) error {
	if !h.IsEnabled() {
		return nil
	}

	// 转换为扫描流量格式（仅请求）
	traffic := h.convertRequestToTraffic(req)

	// 异步提交扫描任务
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		_, err := h.scanner.Scan(ctx, traffic)
		if err != nil {
			logger.Debugf("Scan task error: %v", err)
		}
	}()

	return nil
}

// convertToTraffic 将 HTTP 请求/响应转换为扫描流量
func (h *ScannerHandler) convertToTraffic(req *http.Request, resp *http.Response) *HTTPTraffic {
	traffic := &HTTPTraffic{
		ID:          generateTrafficID(),
		URL:         req.URL.String(),
		Method:      req.Method,
		Headers:     headersToMap(req.Header),
		ContentType: req.Header.Get("Content-Type"),
		UserAgent:   req.UserAgent(),
		Timestamp:   time.Now(),
	}

	// 读取请求体
	if req.Body != nil {
		body, _ := io.ReadAll(req.Body)
		req.Body = io.NopCloser(bytes.NewBuffer(body))
		traffic.Body = body
	}

	// 转换响应
	if resp != nil {
		traffic.Response = &HTTPResponse{
			StatusCode: resp.StatusCode,
			Headers:    headersToMap(resp.Header),
		}

		// 读取响应体
		if resp.Body != nil {
			body, _ := io.ReadAll(resp.Body)
			resp.Body = io.NopCloser(bytes.NewBuffer(body))
			traffic.Response.Body = body
			traffic.Response.Size = int64(len(body))
		}
	}

	return traffic
}

// convertRequestToTraffic 将 HTTP 请求转换为扫描流量
func (h *ScannerHandler) convertRequestToTraffic(req *http.Request) *HTTPTraffic {
	traffic := &HTTPTraffic{
		ID:          generateTrafficID(),
		URL:         req.URL.String(),
		Method:      req.Method,
		Headers:     headersToMap(req.Header),
		ContentType: req.Header.Get("Content-Type"),
		UserAgent:   req.UserAgent(),
		Timestamp:   time.Now(),
	}

	// 读取请求体
	if req.Body != nil {
		body, _ := io.ReadAll(req.Body)
		req.Body = io.NopCloser(bytes.NewBuffer(body))
		traffic.Body = body
	}

	return traffic
}

// headersToMap 将 http.Header 转换为 map
func headersToMap(h http.Header) map[string]string {
	result := make(map[string]string)
	for k, v := range h {
		if len(v) > 0 {
			result[k] = v[0]
		}
	}
	return result
}

// generateTrafficID 生成流量 ID
func generateTrafficID() string {
	return time.Now().Format("20060102150405.999999999")
}

// ScannerMiddleware 扫描器中间件，用于代理服务器
// ScannerMiddleware is a middleware for proxy server
type ScannerMiddleware struct {
	handler *ScannerHandler
}

// NewScannerMiddleware 创建扫描器中间件
func NewScannerMiddleware(handler *ScannerHandler) *ScannerMiddleware {
	return &ScannerMiddleware{handler: handler}
}

// ProcessRequest 处理请求
func (m *ScannerMiddleware) ProcessRequest(req *http.Request) error {
	return m.handler.HandleRequest(req)
}

// ProcessResponse 处理响应
func (m *ScannerMiddleware) ProcessResponse(resp *http.Response, req *http.Request) error {
	return m.handler.HandleResponse(resp, req)
}

// StartVulnCollector 启动漏洞收集器
// 从扫描器的漏洞通道读取漏洞并存储
func (h *ScannerHandler) StartVulnCollector(ctx context.Context) {
	vulnChan := h.scanner.GetVulnChannel()
	if vulnChan == nil {
		return
	}

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case vuln, ok := <-vulnChan:
				if !ok {
					return
				}
				if h.vulnStore != nil {
					if err := h.vulnStore.SaveVulnerability(vuln); err != nil {
						logger.Errorf("Failed to save vulnerability: %v", err)
					}
				}
				logger.Infof("[Vuln] Found: %s - %s (%s)", vuln.Name, vuln.URL, vuln.Severity)
			}
		}
	}()
}

// GetStats 获取扫描统计
func (h *ScannerHandler) GetStats() *ScannerStats {
	return h.scanner.Stats()
}
