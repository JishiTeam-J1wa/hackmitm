package fingerprint

import (
	"fmt"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
)

// FingerprintHandler 指纹识别处理器
type FingerprintHandler struct {
	engine *FingerprintEngine
	logger *logrus.Logger
}

// NewFingerprintHandler 创建指纹识别处理器
func NewFingerprintHandler(logger *logrus.Logger) *FingerprintHandler {
	return &FingerprintHandler{
		engine: NewFingerprintEngine(logger),
		logger: logger,
	}
}

// Initialize 初始化处理器
func (fh *FingerprintHandler) Initialize(fingerprintPath string) error {
	if err := fh.engine.LoadFingerprints(fingerprintPath); err != nil {
		return fmt.Errorf("failed to initialize fingerprint handler: %v", err)
	}

	fh.logger.Info("🎯 指纹识别处理器已就绪，开始监控网络流量...")
	return nil
}

// InitializeWithConfig 使用配置初始化处理器
func (fh *FingerprintHandler) InitializeWithConfig(fingerprintPath string, cacheSize int, cacheTTL int) error {
	if err := fh.engine.LoadFingerprints(fingerprintPath); err != nil {
		return fmt.Errorf("failed to initialize fingerprint handler: %v", err)
	}

	// 配置LRU缓存参数
	if cacheSize > 0 {
		fh.engine.lruCache.SetCapacity(cacheSize)
	}
	if cacheTTL > 0 {
		fh.engine.lruCache.SetTTL(time.Duration(cacheTTL) * time.Second)
	}

	fh.logger.Infof("🎯 指纹识别处理器已就绪，缓存配置: 容量=%d, TTL=%ds",
		fh.engine.lruCache.capacity, cacheTTL)
	return nil
}

// InitializeWithAdvancedConfig 使用高级配置初始化处理器
func (fh *FingerprintHandler) InitializeWithAdvancedConfig(fingerprintPath string, cacheSize int, cacheTTL int, useLayeredIndex bool, maxMatches int) error {
	if err := fh.engine.LoadFingerprints(fingerprintPath); err != nil {
		return fmt.Errorf("failed to initialize fingerprint handler: %v", err)
	}

	// 配置LRU缓存参数
	if cacheSize > 0 {
		fh.engine.lruCache.SetCapacity(cacheSize)
	}
	if cacheTTL > 0 {
		fh.engine.lruCache.SetTTL(time.Duration(cacheTTL) * time.Second)
	}

	// 配置分层索引
	fh.engine.SetLayeredEnabled(useLayeredIndex)

	// 配置最大匹配数量
	if maxMatches > 0 {
		fh.engine.SetMaxMatches(maxMatches)
	}

	fh.logger.Infof("🎯 指纹识别处理器已就绪，配置: 缓存=%d/%ds, 分层索引=%v, 最大匹配=%d",
		cacheSize, cacheTTL, useLayeredIndex, maxMatches)
	return nil
}

// HandleRequest 处理HTTP请求
func (fh *FingerprintHandler) HandleRequest(req *http.Request, resp *http.Response, body []byte) {
	if fh.engine == nil {
		fh.logger.Warn("⚠️ 指纹识别引擎未初始化")
		return
	}

	fh.logger.Debugf("🔄 开始指纹识别: %s", req.URL.String())

	// 构建HTTP响应结构
	httpResp := &HTTPResponse{
		URL:        req.URL.String(),
		StatusCode: resp.StatusCode,
		Headers:    fh.extractHeaders(resp.Header),
		Body:       string(body),
		Title:      fh.extractTitle(string(body)),
	}

	fh.logger.Debugf("📝 响应内容长度: %d 字节", len(body))
	fh.logger.Debugf("📄 页面标题: %s", httpResp.Title)

	// 执行指纹识别
	result := fh.engine.IdentifyFingerprint(httpResp)

	// 美化的控制台输出
	if len(result.Fingerprint) > 0 {
		// 成功识别的情况
		fh.logger.Info("")
		fh.logger.Info("╭─────────────────────────────────────────────────────────────╮")
		fh.logger.Infof("│          🎯 指纹识别成功 - 发现 %d 个技术栈                  │", len(result.Fingerprint))
		fh.logger.Info("╰─────────────────────────────────────────────────────────────╯")
		
		// 基本信息
		fh.logger.Infof("🌐 目标URL: %s", fh.truncateString(result.URL, 80))
		fh.logger.Infof("📊 HTTP状态: %d %s", httpResp.StatusCode, fh.getStatusDescription(httpResp.StatusCode))
		
		title := httpResp.Title
		if title == "" {
			title = "(无标题)"
		}
		fh.logger.Infof("📄 页面标题: %s", fh.truncateString(title, 80))
		fh.logger.Infof("⚡ 识别耗时: %s", result.ProcessTime.String())
		fh.logger.Infof("🎲 置信度: %.1f%% %s", result.Confidence*100, fh.getConfidenceLevel(result.Confidence))
		
		fh.logger.Info("")
		fh.logger.Info("🔍 技术栈识别结果:")
		
		// 指纹列表
		for i, fingerprint := range result.Fingerprint {
			icon := fh.getTechIcon(fingerprint)
			if i == len(result.Fingerprint)-1 {
				fh.logger.Infof("   └─ %s [%d] %s", icon, i+1, fingerprint)
			} else {
				fh.logger.Infof("   ├─ %s [%d] %s", icon, i+1, fingerprint)
			}
		}
		
		fh.logger.Info("")
	} else {
		// 未识别的情况
		fh.logger.Info("")
		fh.logger.Info("╭─────────────────────────────────────────────────────────────╮")
		fh.logger.Info("│              ❓ 未发现已知技术栈                            │")
		fh.logger.Info("╰─────────────────────────────────────────────────────────────╯")
		
		// 基本信息
		fh.logger.Infof("🌐 目标URL: %s", fh.truncateString(result.URL, 80))
		fh.logger.Infof("📊 HTTP状态: %d %s", httpResp.StatusCode, fh.getStatusDescription(httpResp.StatusCode))
		
		title := httpResp.Title
		if title == "" {
			title = "(无标题)"
		}
		fh.logger.Infof("📄 页面标题: %s", fh.truncateString(title, 80))
		fh.logger.Infof("⚡ 识别耗时: %s", result.ProcessTime.String())
		
		fh.logger.Info("")
		fh.logger.Info("💡 可能的原因:")
		fh.logger.Info("   • 使用了自定义或较新的技术栈")
		fh.logger.Info("   • 网站进行了指纹混淆处理")
		fh.logger.Info("   • 指纹库需要更新以支持该技术")
		fh.logger.Info("   • 静态页面或CDN缓存页面")
		fh.logger.Info("")
	}
}

// truncateString 截断字符串到指定长度
func (fh *FingerprintHandler) truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	if maxLen <= 3 {
		return s[:maxLen]
	}
	return s[:maxLen-3] + "..."
}

// getStatusDescription 获取HTTP状态码描述
func (fh *FingerprintHandler) getStatusDescription(code int) string {
	switch {
	case code >= 200 && code < 300:
		return "✅ 成功"
	case code >= 300 && code < 400:
		return "🔄 重定向"
	case code >= 400 && code < 500:
		return "❌ 客户端错误"
	case code >= 500:
		return "💥 服务器错误"
	default:
		return "❓ 未知状态"
	}
}

// getConfidenceLevel 获取置信度级别描述
func (fh *FingerprintHandler) getConfidenceLevel(confidence float64) string {
	switch {
	case confidence >= 0.9:
		return "(极高)"
	case confidence >= 0.7:
		return "(高)"
	case confidence >= 0.5:
		return "(中等)"
	case confidence >= 0.3:
		return "(较低)"
	default:
		return "(低)"
	}
}

// getTechIcon 根据技术名称获取对应图标
func (fh *FingerprintHandler) getTechIcon(tech string) string {
	techLower := strings.ToLower(tech)
	switch {
	case strings.Contains(techLower, "nginx"):
		return "🌐"
	case strings.Contains(techLower, "apache"):
		return "🪶"
	case strings.Contains(techLower, "php"):
		return "🐘"
	case strings.Contains(techLower, "mysql"):
		return "🐬"
	case strings.Contains(techLower, "redis"):
		return "🔴"
	case strings.Contains(techLower, "node") || strings.Contains(techLower, "express"):
		return "💚"
	case strings.Contains(techLower, "python") || strings.Contains(techLower, "django") || strings.Contains(techLower, "flask") || strings.Contains(techLower, "gunicorn"):
		return "🐍"
	case strings.Contains(techLower, "java") || strings.Contains(techLower, "tomcat") || strings.Contains(techLower, "spring"):
		return "☕"
	case strings.Contains(techLower, "wordpress"):
		return "📝"
	case strings.Contains(techLower, "react"):
		return "⚛️"
	case strings.Contains(techLower, "vue"):
		return "💚"
	case strings.Contains(techLower, "angular"):
		return "🅰️"
	case strings.Contains(techLower, "docker"):
		return "🐳"
	case strings.Contains(techLower, "kubernetes"):
		return "☸️"
	case strings.Contains(techLower, "cloudflare"):
		return "☁️"
	case strings.Contains(techLower, "aws"):
		return "🟠"
	case strings.Contains(techLower, "cdn"):
		return "🚀"
	default:
		return "🔧"
	}
}

// extractHeaders 提取响应头
func (fh *FingerprintHandler) extractHeaders(headers http.Header) map[string]string {
	result := make(map[string]string)

	for key, values := range headers {
		if len(values) > 0 {
			result[strings.ToLower(key)] = values[0]
		}
	}

	return result
}

// extractTitle 提取页面标题
func (fh *FingerprintHandler) extractTitle(body string) string {
	// 使用正则表达式提取title标签内容
	titleRegex := regexp.MustCompile(`(?i)<title[^>]*>([^<]*)</title>`)
	matches := titleRegex.FindStringSubmatch(body)

	if len(matches) > 1 {
		return strings.TrimSpace(matches[1])
	}

	return ""
}

// GetStats 获取统计信息
func (fh *FingerprintHandler) GetStats() map[string]interface{} {
	if fh.engine == nil {
		return map[string]interface{}{
			"status": "not_initialized",
		}
	}

	stats := fh.engine.GetStats()
	stats["status"] = "active"
	return stats
}

// GetFingerprintResults 获取指纹识别结果
func (fh *FingerprintHandler) GetFingerprintResults(url string) *FingerprintResult {
	if fh.engine == nil {
		return nil
	}

	// 这里应该从缓存或存储中获取结果
	// 简化实现，返回空结果
	return &FingerprintResult{
		URL:         url,
		Fingerprint: []string{},
		Confidence:  0.0,
		ProcessTime: 0,
	}
}

// IdentifyURL 对指定URL进行指纹识别
func (fh *FingerprintHandler) IdentifyURL(url string) (*FingerprintResult, error) {
	if fh.engine == nil {
		return nil, fmt.Errorf("fingerprint engine not initialized")
	}

	// 创建HTTP客户端
	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	// 发送请求
	resp, err := client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch URL: %v", err)
	}
	defer resp.Body.Close()

	// 读取响应内容
	body := make([]byte, 1024*1024) // 1MB limit
	n, _ := resp.Body.Read(body)
	body = body[:n]

	// 构建HTTP响应结构
	httpResp := &HTTPResponse{
		URL:        url,
		StatusCode: resp.StatusCode,
		Headers:    fh.extractHeaders(resp.Header),
		Body:       string(body),
		Title:      fh.extractTitle(string(body)),
	}

	// 执行指纹识别
	result := fh.engine.IdentifyFingerprint(httpResp)

	return result, nil
}

// Stop 停止指纹识别处理器
func (fh *FingerprintHandler) Stop() {
	if fh.engine != nil {
		fh.engine.Stop()
	}
}

// ClearCache 清理缓存
func (fh *FingerprintHandler) ClearCache() {
	if fh.engine != nil {
		fh.engine.ClearCache()
	}
}
