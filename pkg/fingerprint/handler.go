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
		fh.logger.Info("┌─────────────────────────────────────────────────────────────────────")
		fh.logger.Infof("│ 🎯 指纹识别成功 - %d 个匹配项", len(result.Fingerprint))
		fh.logger.Info("├─────────────────────────────────────────────────────────────────────")
		fh.logger.Infof("│ 🌐 URL: %s", result.URL)
		fh.logger.Infof("│ 📊 状态码: %d", httpResp.StatusCode)
		fh.logger.Infof("│ 📄 标题: %s", httpResp.Title)
		fh.logger.Infof("│ ⚡ 处理时间: %v", result.ProcessTime)
		fh.logger.Infof("│ 🎲 置信度: %.2f%%", result.Confidence*100)
		fh.logger.Info("├─────────────────────────────────────────────────────────────────────")
		fh.logger.Info("│ 🔍 识别到的指纹:")

		// 详细输出每个识别到的指纹
		for i, fingerprint := range result.Fingerprint {
			if i == len(result.Fingerprint)-1 {
				fh.logger.Infof("│   └─ [%d] %s", i+1, fingerprint)
			} else {
				fh.logger.Infof("│   ├─ [%d] %s", i+1, fingerprint)
			}
		}
		fh.logger.Info("└─────────────────────────────────────────────────────────────────────")
	} else {
		// 未识别的情况
		fh.logger.Info("┌─────────────────────────────────────────────────────────────────────")
		fh.logger.Info("│ ❓ 指纹识别 - 暂未识别")
		fh.logger.Info("├─────────────────────────────────────────────────────────────────────")
		fh.logger.Infof("│ 🌐 URL: %s", result.URL)
		fh.logger.Infof("│ 📊 状态码: %d", httpResp.StatusCode)
		fh.logger.Infof("│ 📄 标题: %s", httpResp.Title)
		fh.logger.Infof("│ ⚡ 处理时间: %v", result.ProcessTime)
		fh.logger.Info("│ 💡 提示: 该网站可能使用了未知的技术栈或自定义系统")
		fh.logger.Info("└─────────────────────────────────────────────────────────────────────")
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
