// Package traffic 流量模式识别处理器
package traffic

import (
	"fmt"
	"net/http"
	"sync"
	"time"

	"hackmitm/pkg/logger"
)

// PatternHandler 流量模式识别处理器
type PatternHandler struct {
	recognizer  *PatternRecognizer
	enabled     bool
	mutex       sync.RWMutex
	resultCache map[string]*PatternResult
	cacheSize   int
	cacheTTL    time.Duration
}

// NewPatternHandler 创建流量模式识别处理器
func NewPatternHandler() *PatternHandler {
	return &PatternHandler{
		recognizer:  NewPatternRecognizer(),
		enabled:     true,
		resultCache: make(map[string]*PatternResult),
		cacheSize:   1000, // 缓存最近1000个结果
		cacheTTL:    5 * time.Minute,
	}
}

// HandleRequest 处理请求（实现RequestHandler接口）
func (ph *PatternHandler) HandleRequest(req *http.Request) error {
	if !ph.enabled {
		return nil
	}

	startTime := time.Now()

	// 提取流量信息（请求阶段）
	info := ph.recognizer.ExtractTrafficInfo(req, nil, ph.getClientIP(req), 0)

	// 识别流量模式
	result := ph.recognizer.RecognizePattern(info)

	// 将结果存储在请求上下文中
	if result.Matched {
		// 在请求头中添加识别结果（用于后续处理）
		req.Header.Set("X-Traffic-Pattern", string(result.Pattern.Type))
		req.Header.Set("X-Pattern-Confidence", formatFloat(result.Confidence))

		logger.Debugf("请求流量模式识别: %s (置信度: %.2f) - %s %s",
			result.Pattern.Name, result.Confidence, req.Method, req.URL.Path)
	}

	processingTime := time.Since(startTime)
	logger.Debugf("流量模式识别处理时间: %v", processingTime)

	return nil
}

// HandleResponse 处理响应（实现ResponseHandler接口）
func (ph *PatternHandler) HandleResponse(resp *http.Response, req *http.Request) error {
	if !ph.enabled {
		return nil
	}

	startTime := time.Now()

	// 计算请求处理时间
	var duration time.Duration
	if reqStartTime := req.Header.Get("X-Request-Start-Time"); reqStartTime != "" {
		if t, err := time.Parse(time.RFC3339Nano, reqStartTime); err == nil {
			duration = time.Since(t)
		}
	}

	// 提取完整的流量信息（包含响应）
	info := ph.recognizer.ExtractTrafficInfo(req, resp, ph.getClientIP(req), duration)

	// 重新识别流量模式（包含响应信息）
	result := ph.recognizer.RecognizePattern(info)

	// 缓存结果
	ph.cacheResult(req.URL.String(), result)

	// 在响应头中添加识别结果
	if result.Matched {
		resp.Header.Set("X-Traffic-Pattern", string(result.Pattern.Type))
		resp.Header.Set("X-Pattern-Confidence", formatFloat(result.Confidence))

		logger.Infof("流量模式识别: %s (置信度: %.2f) - %s %s [%d]",
			result.Pattern.Name, result.Confidence, req.Method, req.URL.Path, resp.StatusCode)
	}

	processingTime := time.Since(startTime)
	logger.Debugf("响应流量模式识别处理时间: %v", processingTime)

	return nil
}

// getClientIP 获取客户端IP
func (ph *PatternHandler) getClientIP(req *http.Request) string {
	// 检查X-Forwarded-For头
	if xff := req.Header.Get("X-Forwarded-For"); xff != "" {
		return xff
	}

	// 检查X-Real-IP头
	if xri := req.Header.Get("X-Real-IP"); xri != "" {
		return xri
	}

	// 使用RemoteAddr
	return req.RemoteAddr
}

// cacheResult 缓存识别结果
func (ph *PatternHandler) cacheResult(key string, result *PatternResult) {
	ph.mutex.Lock()
	defer ph.mutex.Unlock()

	// 检查缓存大小，清理过期项
	if len(ph.resultCache) >= ph.cacheSize {
		ph.cleanExpiredCache()
	}

	// 如果仍然超过限制，移除最旧的项
	if len(ph.resultCache) >= ph.cacheSize {
		for k := range ph.resultCache {
			delete(ph.resultCache, k)
			break
		}
	}

	ph.resultCache[key] = result
}

// cleanExpiredCache 清理过期缓存
func (ph *PatternHandler) cleanExpiredCache() {
	now := time.Now()
	for key, result := range ph.resultCache {
		if now.Sub(result.Timestamp) > ph.cacheTTL {
			delete(ph.resultCache, key)
		}
	}
}

// GetCachedResult 获取缓存的识别结果
func (ph *PatternHandler) GetCachedResult(url string) (*PatternResult, bool) {
	ph.mutex.RLock()
	defer ph.mutex.RUnlock()

	result, exists := ph.resultCache[url]
	if !exists {
		return nil, false
	}

	// 检查是否过期
	if time.Since(result.Timestamp) > ph.cacheTTL {
		return nil, false
	}

	return result, true
}

// GetRecognizer 获取模式识别器
func (ph *PatternHandler) GetRecognizer() *PatternRecognizer {
	return ph.recognizer
}

// SetEnabled 设置启用状态
func (ph *PatternHandler) SetEnabled(enabled bool) {
	ph.mutex.Lock()
	defer ph.mutex.Unlock()

	ph.enabled = enabled
	ph.recognizer.SetEnabled(enabled)
	logger.Infof("流量模式识别处理器状态: %v", enabled)
}

// IsEnabled 检查是否启用
func (ph *PatternHandler) IsEnabled() bool {
	ph.mutex.RLock()
	defer ph.mutex.RUnlock()

	return ph.enabled
}

// GetStats 获取统计信息
func (ph *PatternHandler) GetStats() map[string]interface{} {
	ph.mutex.RLock()
	defer ph.mutex.RUnlock()

	stats := ph.recognizer.GetStats()

	return map[string]interface{}{
		"enabled":          ph.enabled,
		"total_requests":   stats.TotalRequests,
		"recognized_count": stats.RecognizedCount,
		"pattern_stats":    stats.PatternStats,
		"cache_size":       len(ph.resultCache),
		"cache_limit":      ph.cacheSize,
		"cache_ttl":        ph.cacheTTL.String(),
		"last_update":      stats.LastUpdate.Format(time.RFC3339),
		"processing_time":  stats.ProcessingTime.String(),
		"recognition_rate": ph.calculateRecognitionRate(stats),
	}
}

// calculateRecognitionRate 计算识别率
func (ph *PatternHandler) calculateRecognitionRate(stats *RecognizerStats) float64 {
	if stats.TotalRequests == 0 {
		return 0.0
	}
	return float64(stats.RecognizedCount) / float64(stats.TotalRequests) * 100.0
}

// ClearCache 清空缓存
func (ph *PatternHandler) ClearCache() {
	ph.mutex.Lock()
	defer ph.mutex.Unlock()

	ph.resultCache = make(map[string]*PatternResult)
	logger.Info("流量模式识别缓存已清空")
}

// SetCacheConfig 设置缓存配置
func (ph *PatternHandler) SetCacheConfig(size int, ttl time.Duration) {
	ph.mutex.Lock()
	defer ph.mutex.Unlock()

	ph.cacheSize = size
	ph.cacheTTL = ttl

	// 如果新大小小于当前缓存，清理多余项
	if len(ph.resultCache) > size {
		count := 0
		for key := range ph.resultCache {
			if count >= size {
				delete(ph.resultCache, key)
			}
			count++
		}
	}

	logger.Infof("更新缓存配置: 大小=%d, TTL=%v", size, ttl)
}

// GetPatternsByType 按类型获取模式
func (ph *PatternHandler) GetPatternsByType(patternType PatternType) []*TrafficPattern {
	patterns := ph.recognizer.GetPatterns()
	var result []*TrafficPattern

	for _, pattern := range patterns {
		if pattern.Type == patternType {
			result = append(result, pattern)
		}
	}

	return result
}

// GetTopPatterns 获取最常识别的模式
func (ph *PatternHandler) GetTopPatterns(limit int) []map[string]interface{} {
	stats := ph.recognizer.GetStats()

	// 转换为可排序的切片
	type patternStat struct {
		Type  PatternType
		Count int64
	}

	var patternStats []patternStat
	for pType, count := range stats.PatternStats {
		patternStats = append(patternStats, patternStat{
			Type:  pType,
			Count: count,
		})
	}

	// 简单的冒泡排序（按计数降序）
	for i := 0; i < len(patternStats)-1; i++ {
		for j := 0; j < len(patternStats)-i-1; j++ {
			if patternStats[j].Count < patternStats[j+1].Count {
				patternStats[j], patternStats[j+1] = patternStats[j+1], patternStats[j]
			}
		}
	}

	// 限制结果数量
	if limit > 0 && limit < len(patternStats) {
		patternStats = patternStats[:limit]
	}

	// 转换为返回格式
	var result []map[string]interface{}
	for _, stat := range patternStats {
		result = append(result, map[string]interface{}{
			"type":  string(stat.Type),
			"count": stat.Count,
		})
	}

	return result
}

// formatFloat 格式化浮点数
func formatFloat(f float64) string {
	return fmt.Sprintf("%.2f", f)
}
