package fingerprint

import (
	"strings"
	"sync"
)

// LayeredIndex 分层过滤索引
type LayeredIndex struct {
	// 第一层：快速过滤索引
	headerIndex map[string][]*FingerprintRule // HTTP头特征索引
	statusIndex map[int][]*FingerprintRule    // 状态码索引
	pathIndex   map[string][]*FingerprintRule // URL路径特征索引

	// 第二层：内容特征索引
	titleIndex   map[string][]*FingerprintRule // Title关键字索引
	keywordIndex map[string][]*FingerprintRule // 关键字索引

	// 第三层：深度匹配索引
	regexRules   []*FingerprintRule // 正则表达式规则
	faviconRules []*FingerprintRule // Favicon规则

	// 索引元数据
	totalRules   int
	indexedRules int
	mutex        sync.RWMutex
}

// IndexStats 索引统计信息
type IndexStats struct {
	TotalRules        int `json:"total_rules"`
	IndexedRules      int `json:"indexed_rules"`
	HeaderIndexSize   int `json:"header_index_size"`
	KeywordIndexSize  int `json:"keyword_index_size"`
	TitleIndexSize    int `json:"title_index_size"`
	RegexRulesCount   int `json:"regex_rules_count"`
	FaviconRulesCount int `json:"favicon_rules_count"`
}

// LayeredResult 分层识别结果
type LayeredResult struct {
	Layer1Matches  []*FingerprintRule `json:"layer1_matches"`
	Layer2Matches  []*FingerprintRule `json:"layer2_matches"`
	Layer3Matches  []*FingerprintRule `json:"layer3_matches"`
	TotalMatches   int                `json:"total_matches"`
	StoppedAtLayer int                `json:"stopped_at_layer"`
}

// NewLayeredIndex 创建分层索引
func NewLayeredIndex() *LayeredIndex {
	return &LayeredIndex{
		headerIndex:  make(map[string][]*FingerprintRule),
		statusIndex:  make(map[int][]*FingerprintRule),
		pathIndex:    make(map[string][]*FingerprintRule),
		titleIndex:   make(map[string][]*FingerprintRule),
		keywordIndex: make(map[string][]*FingerprintRule),
		regexRules:   make([]*FingerprintRule, 0),
		faviconRules: make([]*FingerprintRule, 0),
	}
}

// BuildIndex 构建分层索引
func (li *LayeredIndex) BuildIndex(rules []FingerprintRule) {
	li.mutex.Lock()
	defer li.mutex.Unlock()

	li.totalRules = len(rules)
	li.indexedRules = 0

	// 清空现有索引
	li.clearIndexes()

	for i := range rules {
		rule := &rules[i]
		li.indexRule(rule)
		li.indexedRules++
	}
}

// clearIndexes 清空所有索引
func (li *LayeredIndex) clearIndexes() {
	li.headerIndex = make(map[string][]*FingerprintRule)
	li.statusIndex = make(map[int][]*FingerprintRule)
	li.pathIndex = make(map[string][]*FingerprintRule)
	li.titleIndex = make(map[string][]*FingerprintRule)
	li.keywordIndex = make(map[string][]*FingerprintRule)
	li.regexRules = make([]*FingerprintRule, 0)
	li.faviconRules = make([]*FingerprintRule, 0)
}

// indexRule 为单个规则建立索引
func (li *LayeredIndex) indexRule(rule *FingerprintRule) {
	switch rule.Method {
	case "keyword":
		li.indexKeywordRule(rule)
	case "faviconhash":
		li.faviconRules = append(li.faviconRules, rule)
	case "regex", "regula":
		li.regexRules = append(li.regexRules, rule)
	default:
		li.indexKeywordRule(rule) // 默认按关键字处理
	}
}

// indexKeywordRule 为关键字规则建立索引
func (li *LayeredIndex) indexKeywordRule(rule *FingerprintRule) {
	for _, keyword := range rule.Keyword {
		normalizedKeyword := strings.ToLower(strings.TrimSpace(keyword))
		if normalizedKeyword == "" {
			continue
		}

		// 根据location和keyword内容分类索引
		switch rule.Location {
		case "header":
			li.addToHeaderIndex(normalizedKeyword, rule)
		case "title":
			li.addToTitleIndex(normalizedKeyword, rule)
		case "body":
			li.addToKeywordIndex(normalizedKeyword, rule)
		default:
			li.addToKeywordIndex(normalizedKeyword, rule)
		}

		// 特殊路径特征索引
		if strings.Contains(normalizedKeyword, "/") || strings.Contains(normalizedKeyword, "?") {
			li.addToPathIndex(normalizedKeyword, rule)
		}
	}
}

// addToHeaderIndex 添加到HTTP头索引
func (li *LayeredIndex) addToHeaderIndex(keyword string, rule *FingerprintRule) {
	if li.headerIndex[keyword] == nil {
		li.headerIndex[keyword] = make([]*FingerprintRule, 0)
	}
	li.headerIndex[keyword] = append(li.headerIndex[keyword], rule)
}

// addToTitleIndex 添加到标题索引
func (li *LayeredIndex) addToTitleIndex(keyword string, rule *FingerprintRule) {
	if li.titleIndex[keyword] == nil {
		li.titleIndex[keyword] = make([]*FingerprintRule, 0)
	}
	li.titleIndex[keyword] = append(li.titleIndex[keyword], rule)
}

// addToKeywordIndex 添加到关键字索引
func (li *LayeredIndex) addToKeywordIndex(keyword string, rule *FingerprintRule) {
	if li.keywordIndex[keyword] == nil {
		li.keywordIndex[keyword] = make([]*FingerprintRule, 0)
	}
	li.keywordIndex[keyword] = append(li.keywordIndex[keyword], rule)
}

// addToPathIndex 添加到路径索引
func (li *LayeredIndex) addToPathIndex(keyword string, rule *FingerprintRule) {
	if li.pathIndex[keyword] == nil {
		li.pathIndex[keyword] = make([]*FingerprintRule, 0)
	}
	li.pathIndex[keyword] = append(li.pathIndex[keyword], rule)
}

// SearchLayered 分层搜索指纹
func (li *LayeredIndex) SearchLayered(response *HTTPResponse, maxMatches int) *LayeredResult {
	li.mutex.RLock()
	defer li.mutex.RUnlock()

	result := &LayeredResult{
		Layer1Matches: make([]*FingerprintRule, 0),
		Layer2Matches: make([]*FingerprintRule, 0),
		Layer3Matches: make([]*FingerprintRule, 0),
	}

	matchedRules := make(map[*FingerprintRule]bool)

	// 第一层：快速过滤
	li.searchLayer1(response, result, matchedRules)
	if len(matchedRules) >= maxMatches {
		result.StoppedAtLayer = 1
		result.TotalMatches = len(matchedRules)
		return result
	}

	// 第二层：内容特征
	li.searchLayer2(response, result, matchedRules)
	if len(matchedRules) >= maxMatches {
		result.StoppedAtLayer = 2
		result.TotalMatches = len(matchedRules)
		return result
	}

	// 第三层：深度匹配
	li.searchLayer3(response, result, matchedRules)
	result.StoppedAtLayer = 3
	result.TotalMatches = len(matchedRules)

	return result
}

// searchLayer1 第一层：快速过滤
func (li *LayeredIndex) searchLayer1(response *HTTPResponse, result *LayeredResult, matched map[*FingerprintRule]bool) {
	// 搜索HTTP头特征
	for headerName, headerValue := range response.Headers {
		normalizedHeader := strings.ToLower(headerName + ":" + headerValue)
		for keyword, rules := range li.headerIndex {
			if strings.Contains(normalizedHeader, keyword) {
				for _, rule := range rules {
					if !matched[rule] && li.quickMatchRule(response, rule) {
						matched[rule] = true
						result.Layer1Matches = append(result.Layer1Matches, rule)
					}
				}
			}
		}
	}

	// 搜索状态码特征
	if rules, exists := li.statusIndex[response.StatusCode]; exists {
		for _, rule := range rules {
			if !matched[rule] && li.quickMatchRule(response, rule) {
				matched[rule] = true
				result.Layer1Matches = append(result.Layer1Matches, rule)
			}
		}
	}

	// 搜索路径特征
	normalizedURL := strings.ToLower(response.URL)
	for pathKeyword, rules := range li.pathIndex {
		if strings.Contains(normalizedURL, pathKeyword) {
			for _, rule := range rules {
				if !matched[rule] && li.quickMatchRule(response, rule) {
					matched[rule] = true
					result.Layer1Matches = append(result.Layer1Matches, rule)
				}
			}
		}
	}
}

// searchLayer2 第二层：内容特征
func (li *LayeredIndex) searchLayer2(response *HTTPResponse, result *LayeredResult, matched map[*FingerprintRule]bool) {
	// 搜索标题特征
	normalizedTitle := strings.ToLower(response.Title)
	for titleKeyword, rules := range li.titleIndex {
		if strings.Contains(normalizedTitle, titleKeyword) {
			for _, rule := range rules {
				if !matched[rule] && li.contentMatchRule(response, rule) {
					matched[rule] = true
					result.Layer2Matches = append(result.Layer2Matches, rule)
				}
			}
		}
	}

	// 搜索内容关键字
	normalizedBody := strings.ToLower(response.Body)
	for keyword, rules := range li.keywordIndex {
		if strings.Contains(normalizedBody, keyword) {
			for _, rule := range rules {
				if !matched[rule] && li.contentMatchRule(response, rule) {
					matched[rule] = true
					result.Layer2Matches = append(result.Layer2Matches, rule)
				}
			}
		}
	}
}

// searchLayer3 第三层：深度匹配
func (li *LayeredIndex) searchLayer3(response *HTTPResponse, result *LayeredResult, matched map[*FingerprintRule]bool) {
	// 正则表达式匹配
	for _, rule := range li.regexRules {
		if !matched[rule] && li.regexMatchRule(response, rule) {
			matched[rule] = true
			result.Layer3Matches = append(result.Layer3Matches, rule)
		}
	}

	// Favicon匹配
	for _, rule := range li.faviconRules {
		if !matched[rule] && li.faviconMatchRule(response, rule) {
			matched[rule] = true
			result.Layer3Matches = append(result.Layer3Matches, rule)
		}
	}
}

// quickMatchRule 快速匹配规则（第一层）
func (li *LayeredIndex) quickMatchRule(response *HTTPResponse, rule *FingerprintRule) bool {
	if rule.Method != "keyword" {
		return false
	}

	// 简单的关键字匹配，不做复杂验证
	switch rule.Location {
	case "header":
		headerStr := li.headersToString(response.Headers)
		return li.containsAllKeywords(strings.ToLower(headerStr), rule.Keyword)
	case "title":
		return li.containsAllKeywords(strings.ToLower(response.Title), rule.Keyword)
	default:
		return li.containsAllKeywords(strings.ToLower(response.Body), rule.Keyword)
	}
}

// contentMatchRule 内容匹配规则（第二层）
func (li *LayeredIndex) contentMatchRule(response *HTTPResponse, rule *FingerprintRule) bool {
	if rule.Method != "keyword" {
		return false
	}

	// 更严格的匹配验证
	var content string
	switch rule.Location {
	case "header":
		content = li.headersToString(response.Headers)
	case "title":
		content = response.Title
	case "body":
		content = response.Body
	default:
		content = response.Body
	}

	return li.containsAllKeywords(strings.ToLower(content), rule.Keyword)
}

// regexMatchRule 正则匹配规则（第三层）
func (li *LayeredIndex) regexMatchRule(response *HTTPResponse, rule *FingerprintRule) bool {
	if len(rule.CompiledRegexes) == 0 {
		return false
	}

	var content string
	switch rule.Location {
	case "header":
		content = li.headersToString(response.Headers)
	case "title":
		content = response.Title
	case "body":
		content = response.Body
	default:
		content = response.Body
	}

	// 所有正则表达式都必须匹配
	for _, compiledRegex := range rule.CompiledRegexes {
		if !compiledRegex.MatchString(content) {
			return false
		}
	}

	return true
}

// faviconMatchRule Favicon匹配规则（第三层）
func (li *LayeredIndex) faviconMatchRule(_ *HTTPResponse, _ *FingerprintRule) bool {
	// 这里可以实现favicon匹配逻辑
	// 为了保持性能，可以简化或异步处理
	// TODO: 实现favicon匹配逻辑
	// 参数response和rule将在未来的实现中使用
	return false // 暂时返回false，避免性能影响
}

// containsAllKeywords 检查是否包含所有关键字
func (li *LayeredIndex) containsAllKeywords(content string, keywords []string) bool {
	for _, keyword := range keywords {
		if !strings.Contains(content, strings.ToLower(keyword)) {
			return false
		}
	}
	return true
}

// headersToString 将headers转换为字符串
func (li *LayeredIndex) headersToString(headers map[string]string) string {
	var parts []string
	for key, value := range headers {
		parts = append(parts, key+":"+value)
	}
	return strings.Join(parts, " ")
}

// GetStats 获取索引统计信息
func (li *LayeredIndex) GetStats() *IndexStats {
	li.mutex.RLock()
	defer li.mutex.RUnlock()

	return &IndexStats{
		TotalRules:        li.totalRules,
		IndexedRules:      li.indexedRules,
		HeaderIndexSize:   len(li.headerIndex),
		KeywordIndexSize:  len(li.keywordIndex),
		TitleIndexSize:    len(li.titleIndex),
		RegexRulesCount:   len(li.regexRules),
		FaviconRulesCount: len(li.faviconRules),
	}
}

// RebuildIndex 重建索引
func (li *LayeredIndex) RebuildIndex(rules []FingerprintRule) {
	li.BuildIndex(rules)
}
