package fingerprint

import (
	"crypto/md5"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
)

// FingerprintRule 指纹规则结构
type FingerprintRule struct {
	CMS      string   `json:"cms"`
	Method   string   `json:"method"`   // keyword, faviconhash, regula
	Location string   `json:"location"` // body, header, title
	Keyword  []string `json:"keyword"`
	// 预编译的正则表达式
	CompiledRegexes []*regexp.Regexp `json:"-"`
	// 预编译的favicon URL查找正则表达式
	CompiledFaviconRegexes []*regexp.Regexp `json:"-"`
}

// FingerprintDB 指纹数据库结构
type FingerprintDB struct {
	Fingerprint []FingerprintRule `json:"fingerprint"`
}

// FingerprintEngine 指纹识别引擎
type FingerprintEngine struct {
	rules         []FingerprintRule
	lruCache      *LRUCache
	layeredIndex  *LayeredIndex
	logger        *logrus.Logger
	cacheHits     int64
	cacheMisses   int64
	cleanupTicker *time.Ticker
	stopCleanup   chan bool

	// 性能配置
	maxMatches int  // 最大匹配数量，用于早期退出
	useLayered bool // 是否使用分层索引

	// 预编译的favicon查找正则表达式
	faviconRegexes []*regexp.Regexp
}

// FingerprintResult 指纹识别结果
type FingerprintResult struct {
	URL         string            `json:"url"`
	Fingerprint []string          `json:"fingerprint"`
	Confidence  float64           `json:"confidence"`
	ProcessTime time.Duration     `json:"process_time"`
	Headers     map[string]string `json:"headers,omitempty"`
}

// HTTPResponse HTTP响应结构
type HTTPResponse struct {
	URL        string
	StatusCode int
	Headers    map[string]string
	Body       string
	Title      string
}

// NewFingerprintEngine 创建指纹识别引擎
func NewFingerprintEngine(logger *logrus.Logger) *FingerprintEngine {
	fe := &FingerprintEngine{
		rules:        make([]FingerprintRule, 0),
		lruCache:     NewLRUCache(1000, 30*time.Minute), // 1000个条目，30分钟TTL
		layeredIndex: NewLayeredIndex(),
		logger:       logger,
		stopCleanup:  make(chan bool),
		maxMatches:   10,   // 默认最大匹配10个指纹
		useLayered:   true, // 默认启用分层索引
	}

	// 预编译favicon查找正则表达式
	faviconPatterns := []string{
		`<link[^>]*rel=["\']shortcut icon["\'][^>]*href=["\']([^"\']+)["\']`,
		`<link[^>]*rel=["\']icon["\'][^>]*href=["\']([^"\']+)["\']`,
		`<link[^>]*href=["\']([^"\']+)["\'][^>]*rel=["\']shortcut icon["\']`,
		`<link[^>]*href=["\']([^"\']+)["\'][^>]*rel=["\']icon["\']`,
	}

	fe.faviconRegexes = make([]*regexp.Regexp, 0, len(faviconPatterns))
	for _, pattern := range faviconPatterns {
		if compiled, err := regexp.Compile(pattern); err == nil {
			fe.faviconRegexes = append(fe.faviconRegexes, compiled)
		} else {
			logger.Warnf("Failed to compile favicon regex: %s, error: %v", pattern, err)
		}
	}

	// 启动缓存清理定时器
	fe.cleanupTicker = time.NewTicker(10 * time.Minute) // 每10分钟清理一次过期缓存
	go fe.runCacheCleanup()

	return fe
}

// runCacheCleanup 运行缓存清理任务
func (fe *FingerprintEngine) runCacheCleanup() {
	for {
		select {
		case <-fe.cleanupTicker.C:
			if cleaned := fe.lruCache.CleanExpired(); cleaned > 0 {
				fe.logger.Debugf("Cleaned %d expired cache entries", cleaned)
			}
		case <-fe.stopCleanup:
			fe.cleanupTicker.Stop()
			return
		}
	}
}

// LoadFingerprints 加载指纹库
func (fe *FingerprintEngine) LoadFingerprints(filepath string) error {
	fe.logger.Info("")
	fe.logger.Info("╭─────────────────────────────────────────────────────────────╮")
	fe.logger.Info("│                🔍 指纹识别系统初始化                        │")
	fe.logger.Info("╰─────────────────────────────────────────────────────────────╯")
	fe.logger.Infof("📂 指纹库路径: %s", filepath)

	data, err := os.ReadFile(filepath)
	if err != nil {
		fe.logger.Errorf("❌ 指纹库加载失败: %v", err)
		return fmt.Errorf("failed to read fingerprint file: %v", err)
	}

	var db FingerprintDB
	if err := json.Unmarshal(data, &db); err != nil {
		fe.logger.Errorf("❌ 指纹库解析失败: %v", err)
		return fmt.Errorf("failed to parse fingerprint file: %v", err)
	}

	// 预编译正则表达式
	for i := range db.Fingerprint {
		rule := &db.Fingerprint[i]

		// 预编译关键字中的正则表达式（如果method为regex或regula）
		if rule.Method == "regex" || rule.Method == "regula" {
			rule.CompiledRegexes = make([]*regexp.Regexp, 0, len(rule.Keyword))
			for _, pattern := range rule.Keyword {
				if compiled, err := regexp.Compile(pattern); err == nil {
					rule.CompiledRegexes = append(rule.CompiledRegexes, compiled)
				} else {
					fe.logger.Warnf("⚠️ 正则表达式编译失败: %s", pattern)
				}
			}
		}
	}

	fe.rules = db.Fingerprint

	// 构建分层索引
	if fe.useLayered {
		fe.layeredIndex.BuildIndex(fe.rules)
		fe.logger.Info("🔧 分层索引构建完成")
	}

	// 统计不同类型的指纹规则
	methodStats := make(map[string]int)
	for _, rule := range fe.rules {
		methodStats[rule.Method]++
	}

	fe.logger.Infof("📊 加载指纹规则: %d 条", len(fe.rules))
	for method, count := range methodStats {
		switch method {
		case "keyword":
			fe.logger.Infof("   🔤 关键字匹配: %d 条", count)
		case "faviconhash":
			fe.logger.Infof("   🎨 Favicon Hash: %d 条", count)
		case "regula", "regex":
			fe.logger.Infof("   🔧 正则表达式: %d 条", count)
		default:
			fe.logger.Infof("   ❓ %s: %d 条", method, count)
		}
	}
	fe.logger.Info("✅ 指纹库加载完成")
	fe.logger.Info("")
	return nil
}

// IdentifyFingerprint 识别指纹
func (fe *FingerprintEngine) IdentifyFingerprint(response *HTTPResponse) *FingerprintResult {
	startTime := time.Now()

	// 检查LRU缓存
	cacheKey := fe.generateCacheKey(response)
	if cached, exists := fe.lruCache.Get(cacheKey); exists {
		fe.cacheHits++
		return &FingerprintResult{
			URL:         response.URL,
			Fingerprint: cached,
			Confidence:  1.0,
			ProcessTime: time.Since(startTime),
		}
	}
	fe.cacheMisses++

	fingerprints := make([]string, 0)
	matchedRules := make(map[string]bool)

	// 使用分层索引进行快速匹配
	if fe.useLayered {
		layeredResult := fe.layeredIndex.SearchLayered(response, fe.maxMatches)

		// 收集所有层的匹配结果
		allMatches := make([]*FingerprintRule, 0)
		allMatches = append(allMatches, layeredResult.Layer1Matches...)
		allMatches = append(allMatches, layeredResult.Layer2Matches...)
		allMatches = append(allMatches, layeredResult.Layer3Matches...)

		// 去重并收集CMS名称
		for _, rule := range allMatches {
			if !matchedRules[rule.CMS] {
				fingerprints = append(fingerprints, rule.CMS)
				matchedRules[rule.CMS] = true
			}
		}
	} else {
		// 回退到传统的遍历方式
		for _, rule := range fe.rules {
			if fe.matchRule(response, rule) {
				if !matchedRules[rule.CMS] {
					fingerprints = append(fingerprints, rule.CMS)
					matchedRules[rule.CMS] = true
				}
			}

			// 早期退出优化
			if len(matchedRules) >= fe.maxMatches {
				break
			}
		}
	}

	// 缓存结果到LRU缓存
	fe.lruCache.Put(cacheKey, fingerprints)

	confidence := fe.calculateConfidence(len(fingerprints))

	return &FingerprintResult{
		URL:         response.URL,
		Fingerprint: fingerprints,
		Confidence:  confidence,
		ProcessTime: time.Since(startTime),
		Headers:     response.Headers,
	}
}

// matchRule 匹配单个规则
func (fe *FingerprintEngine) matchRule(response *HTTPResponse, rule FingerprintRule) bool {
	var content string

	// 根据location选择匹配内容
	switch rule.Location {
	case "body":
		content = response.Body
	case "header":
		content = fe.headersToString(response.Headers)
	case "title":
		content = response.Title
	default:
		content = response.Body
	}

	// 根据method选择匹配方式
	switch rule.Method {
	case "keyword":
		return fe.matchKeyword(content, rule.Keyword)
	case "faviconhash":
		return fe.matchFaviconHash(response, rule.Keyword)
	case "regula", "regex":
		return fe.matchRegex(content, rule)
	default:
		return fe.matchKeyword(content, rule.Keyword)
	}
}

// matchKeyword 关键字匹配
func (fe *FingerprintEngine) matchKeyword(content string, keywords []string) bool {
	if len(keywords) == 0 {
		return false
	}

	content = strings.ToLower(content)

	// 所有关键字都必须匹配
	for _, keyword := range keywords {
		if !strings.Contains(content, strings.ToLower(keyword)) {
			return false
		}
	}

	return true
}

// matchFaviconHash favicon hash匹配
func (fe *FingerprintEngine) matchFaviconHash(response *HTTPResponse, hashes []string) bool {
	if len(hashes) == 0 {
		return false
	}

	// 从响应中提取favicon
	faviconHash := fe.extractFaviconHash(response)
	if faviconHash == "" {
		return false
	}

	// 检查是否匹配任何一个hash
	for _, hash := range hashes {
		if faviconHash == hash {
			return true
		}
	}

	return false
}

// matchRegex 正则表达式匹配（使用预编译的正则表达式）
func (fe *FingerprintEngine) matchRegex(content string, rule FingerprintRule) bool {
	if len(rule.CompiledRegexes) == 0 {
		return false
	}

	// 所有正则表达式都必须匹配
	for _, compiledRegex := range rule.CompiledRegexes {
		if !compiledRegex.MatchString(content) {
			return false
		}
	}

	return true
}

// extractFaviconHash 提取favicon hash
func (fe *FingerprintEngine) extractFaviconHash(response *HTTPResponse) string {
	// 从HTML中查找favicon链接
	faviconURL := fe.findFaviconURL(response.Body, response.URL)
	if faviconURL == "" {
		return ""
	}

	// 下载favicon并计算hash
	return fe.calculateFaviconHash(faviconURL)
}

// findFaviconURL 查找favicon URL（使用预编译的正则表达式）
func (fe *FingerprintEngine) findFaviconURL(body, baseURL string) string {
	// 使用预编译的正则表达式查找link标签中的favicon
	for _, regex := range fe.faviconRegexes {
		matches := regex.FindStringSubmatch(body)
		if len(matches) > 1 {
			return fe.resolveURL(baseURL, matches[1])
		}
	}

	// 默认favicon路径
	return fe.resolveURL(baseURL, "/favicon.ico")
}

// calculateFaviconHash 计算favicon hash
func (fe *FingerprintEngine) calculateFaviconHash(url string) string {
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	resp, err := client.Get(url)
	if err != nil {
		return ""
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return ""
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return ""
	}

	// 计算MD5 hash
	hash := md5.Sum(data)
	encoded := base64.StdEncoding.EncodeToString(hash[:])

	// 转换为数字格式（类似EHole的格式）
	return fe.hashToNumber(encoded)
}

// hashToNumber 将hash转换为数字格式
func (fe *FingerprintEngine) hashToNumber(hash string) string {
	// 简化实现，实际应该使用mmh3算法
	sum := int32(0)
	for i, char := range hash {
		sum += int32(char) * int32(i+1)
	}
	return fmt.Sprintf("%d", sum)
}

// resolveURL 解析相对URL
func (fe *FingerprintEngine) resolveURL(baseURL, relativeURL string) string {
	if strings.HasPrefix(relativeURL, "http") {
		return relativeURL
	}

	if strings.HasPrefix(relativeURL, "//") {
		return "http:" + relativeURL
	}

	if strings.HasPrefix(relativeURL, "/") {
		// 提取base URL的协议和域名
		parts := strings.SplitN(baseURL, "://", 2)
		if len(parts) != 2 {
			return baseURL + relativeURL
		}

		protocol := parts[0]
		domainPart := strings.Split(parts[1], "/")[0]
		return protocol + "://" + domainPart + relativeURL
	}

	return baseURL + "/" + relativeURL
}

// headersToString 将headers转换为字符串
func (fe *FingerprintEngine) headersToString(headers map[string]string) string {
	var parts []string
	for key, value := range headers {
		parts = append(parts, fmt.Sprintf("%s: %s", key, value))
	}
	return strings.Join(parts, "\n")
}

// calculateConfidence 计算置信度
func (fe *FingerprintEngine) calculateConfidence(matchCount int) float64 {
	if matchCount == 0 {
		return 0.0
	}

	// 简单的置信度计算
	confidence := 0.5 + float64(matchCount)*0.1
	if confidence > 1.0 {
		confidence = 1.0
	}

	return confidence
}

// generateCacheKey 生成缓存键
func (fe *FingerprintEngine) generateCacheKey(response *HTTPResponse) string {
	content := response.Body + response.Title + fe.headersToString(response.Headers)
	hash := md5.Sum([]byte(content))
	return fmt.Sprintf("%x", hash)
}

// GetStats 获取统计信息
func (fe *FingerprintEngine) GetStats() map[string]interface{} {
	stats := map[string]interface{}{
		"total_rules":    len(fe.rules),
		"cache_size":     fe.lruCache.Size(),
		"cache_hits":     fe.cacheHits,
		"cache_misses":   fe.cacheMisses,
		"rule_breakdown": fe.getRuleBreakdown(),
	}

	// 添加LRU缓存统计信息
	cacheStats := fe.lruCache.GetStats()
	for key, value := range cacheStats {
		stats["cache_"+key] = value
	}

	// 计算缓存命中率
	totalRequests := fe.cacheHits + fe.cacheMisses
	if totalRequests > 0 {
		stats["cache_hit_rate"] = float64(fe.cacheHits) / float64(totalRequests) * 100
	} else {
		stats["cache_hit_rate"] = 0.0
	}

	// 添加分层索引统计信息
	if fe.useLayered && fe.layeredIndex != nil {
		indexStats := fe.layeredIndex.GetStats()
		stats["layered_index"] = map[string]interface{}{
			"enabled":             fe.useLayered,
			"total_rules":         indexStats.TotalRules,
			"indexed_rules":       indexStats.IndexedRules,
			"header_index_size":   indexStats.HeaderIndexSize,
			"keyword_index_size":  indexStats.KeywordIndexSize,
			"title_index_size":    indexStats.TitleIndexSize,
			"regex_rules_count":   indexStats.RegexRulesCount,
			"favicon_rules_count": indexStats.FaviconRulesCount,
		}
	} else {
		stats["layered_index"] = map[string]interface{}{
			"enabled": false,
		}
	}

	// 添加性能配置
	stats["performance_config"] = map[string]interface{}{
		"max_matches": fe.maxMatches,
		"use_layered": fe.useLayered,
	}

	return stats
}

// getRuleBreakdown 获取规则分类统计
func (fe *FingerprintEngine) getRuleBreakdown() map[string]int {
	breakdown := make(map[string]int)

	for _, rule := range fe.rules {
		breakdown[rule.Method]++
	}

	return breakdown
}

// ClearCache 清理缓存
func (fe *FingerprintEngine) ClearCache() {
	fe.lruCache.Clear()
	fe.cacheHits = 0
	fe.cacheMisses = 0
}

// Stop 停止指纹识别引擎
func (fe *FingerprintEngine) Stop() {
	if fe.stopCleanup != nil {
		close(fe.stopCleanup)
	}
	if fe.cleanupTicker != nil {
		fe.cleanupTicker.Stop()
	}
}

// SetMaxMatches 设置最大匹配数量
func (fe *FingerprintEngine) SetMaxMatches(maxMatches int) {
	if maxMatches > 0 {
		fe.maxMatches = maxMatches
	}
}

// SetLayeredEnabled 设置是否启用分层索引
func (fe *FingerprintEngine) SetLayeredEnabled(enabled bool) {
	fe.useLayered = enabled
	if enabled && fe.layeredIndex != nil && len(fe.rules) > 0 {
		// 重建索引
		fe.layeredIndex.BuildIndex(fe.rules)
		fe.logger.Info("分层索引已启用并重建")
	} else if !enabled {
		fe.logger.Info("分层索引已禁用")
	}
}

// GetLayeredEnabled 获取分层索引启用状态
func (fe *FingerprintEngine) GetLayeredEnabled() bool {
	return fe.useLayered
}

// RebuildIndex 重建分层索引
func (fe *FingerprintEngine) RebuildIndex() {
	if fe.useLayered && fe.layeredIndex != nil && len(fe.rules) > 0 {
		fe.layeredIndex.RebuildIndex(fe.rules)
		fe.logger.Info("分层索引已重建")
	}
}
