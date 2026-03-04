// Package traffic 流量模式识别
package traffic

import (
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"hackmitm/pkg/logger"
)

// PatternType 流量模式类型
type PatternType string

const (
	PatternTypeAPI       PatternType = "api"       // API调用
	PatternTypeWebPage   PatternType = "webpage"   // 网页访问
	PatternTypeDownload  PatternType = "download"  // 文件下载
	PatternTypeUpload    PatternType = "upload"    // 文件上传
	PatternTypeWebSocket PatternType = "websocket" // WebSocket连接
	PatternTypeAjax      PatternType = "ajax"      // AJAX请求
	PatternTypeBot       PatternType = "bot"       // 爬虫/机器人
	PatternTypeAttack    PatternType = "attack"    // 攻击行为
	PatternTypeStatic    PatternType = "static"    // 静态资源
	PatternTypeAuth      PatternType = "auth"      // 认证相关
	PatternTypeAdmin     PatternType = "admin"     // 管理后台
	PatternTypeSearch    PatternType = "search"    // 搜索请求
	PatternTypeForm      PatternType = "form"      // 表单提交
	PatternTypeRedirect  PatternType = "redirect"  // 重定向
	PatternTypeError     PatternType = "error"     // 错误页面
	PatternTypeUnknown   PatternType = "unknown"   // 未知模式
)

// TrafficPattern 流量模式定义
type TrafficPattern struct {
	Type        PatternType `json:"type"`
	Name        string      `json:"name"`
	Description string      `json:"description"`
	Rules       []Rule      `json:"rules"`
	Priority    int         `json:"priority"`   // 优先级，数字越小优先级越高
	Confidence  float64     `json:"confidence"` // 置信度 0-1
	Enabled     bool        `json:"enabled"`
}

// Rule 模式识别规则
type Rule struct {
	Type     RuleType       `json:"type"`
	Field    string         `json:"field"`    // 检查的字段
	Operator Operator       `json:"operator"` // 操作符
	Value    string         `json:"value"`    // 匹配值
	Regex    *regexp.Regexp `json:"-"`        // 编译后的正则表达式
	Weight   float64        `json:"weight"`   // 权重
	Required bool           `json:"required"` // 是否必需匹配
}

// RuleType 规则类型
type RuleType string

const (
	RuleTypeMethod      RuleType = "method"       // HTTP方法
	RuleTypePath        RuleType = "path"         // URL路径
	RuleTypeQuery       RuleType = "query"        // 查询参数
	RuleTypeHeader      RuleType = "header"       // 请求头
	RuleTypeUserAgent   RuleType = "user_agent"   // User-Agent
	RuleTypeContentType RuleType = "content_type" // Content-Type
	RuleTypeReferer     RuleType = "referer"      // Referer
	RuleTypeHost        RuleType = "host"         // Host
	RuleTypeSize        RuleType = "size"         // 内容大小
	RuleTypeStatusCode  RuleType = "status_code"  // 状态码
	RuleTypeBody        RuleType = "body"         // 请求/响应体
)

// Operator 操作符
type Operator string

const (
	OperatorEquals     Operator = "equals"      // 等于
	OperatorContains   Operator = "contains"    // 包含
	OperatorStartsWith Operator = "starts_with" // 开始于
	OperatorEndsWith   Operator = "ends_with"   // 结束于
	OperatorRegex      Operator = "regex"       // 正则匹配
	OperatorGreater    Operator = "greater"     // 大于
	OperatorLess       Operator = "less"        // 小于
	OperatorIn         Operator = "in"          // 在列表中
	OperatorNotIn      Operator = "not_in"      // 不在列表中
)

// PatternResult 模式识别结果
type PatternResult struct {
	Pattern      *TrafficPattern `json:"pattern"`
	Matched      bool            `json:"matched"`
	Confidence   float64         `json:"confidence"`
	MatchedRules []string        `json:"matched_rules"`
	Timestamp    time.Time       `json:"timestamp"`
	RequestID    string          `json:"request_id"`
}

// TrafficInfo 流量信息
type TrafficInfo struct {
	Request      *http.Request     `json:"-"`
	Response     *http.Response    `json:"-"`
	Method       string            `json:"method"`
	URL          string            `json:"url"`
	Path         string            `json:"path"`
	Query        string            `json:"query"`
	Headers      map[string]string `json:"headers"`
	UserAgent    string            `json:"user_agent"`
	ContentType  string            `json:"content_type"`
	Referer      string            `json:"referer"`
	Host         string            `json:"host"`
	RequestSize  int64             `json:"request_size"`
	ResponseSize int64             `json:"response_size"`
	StatusCode   int               `json:"status_code"`
	Duration     time.Duration     `json:"duration"`
	ClientIP     string            `json:"client_ip"`
	Timestamp    time.Time         `json:"timestamp"`
}

// PatternRecognizer 流量模式识别器
type PatternRecognizer struct {
	patterns   []*TrafficPattern
	stats      *RecognizerStats
	mutex      sync.RWMutex
	enabled    bool
	confidence float64 // 最小置信度阈值
}

// RecognizerStats 识别器统计信息
type RecognizerStats struct {
	TotalRequests   int64                 `json:"total_requests"`
	RecognizedCount int64                 `json:"recognized_count"`
	PatternStats    map[PatternType]int64 `json:"pattern_stats"`
	LastUpdate      time.Time             `json:"last_update"`
	ProcessingTime  time.Duration         `json:"processing_time"`
}

// NewPatternRecognizer 创建流量模式识别器
func NewPatternRecognizer() *PatternRecognizer {
	pr := &PatternRecognizer{
		patterns:   make([]*TrafficPattern, 0),
		enabled:    true,
		confidence: 0.6, // 默认最小置信度60%
		stats: &RecognizerStats{
			PatternStats: make(map[PatternType]int64),
			LastUpdate:   time.Now(),
		},
	}

	// 加载默认模式
	pr.loadDefaultPatterns()

	return pr
}

// loadDefaultPatterns 加载默认流量模式
func (pr *PatternRecognizer) loadDefaultPatterns() {
	defaultPatterns := []*TrafficPattern{
		// API调用模式
		{
			Type:        PatternTypeAPI,
			Name:        "REST API",
			Description: "RESTful API调用",
			Priority:    10,
			Confidence:  0.8,
			Enabled:     true,
			Rules: []Rule{
				{Type: RuleTypePath, Operator: OperatorRegex, Value: `^/api/`, Weight: 0.4, Required: false},
				{Type: RuleTypeContentType, Operator: OperatorContains, Value: "application/json", Weight: 0.3},
				{Type: RuleTypeMethod, Operator: OperatorIn, Value: "GET,POST,PUT,DELETE,PATCH", Weight: 0.3},
			},
		},
		// 网页访问模式
		{
			Type:        PatternTypeWebPage,
			Name:        "Web Page",
			Description: "网页访问",
			Priority:    20,
			Confidence:  0.7,
			Enabled:     true,
			Rules: []Rule{
				{Type: RuleTypeMethod, Operator: OperatorEquals, Value: "GET", Weight: 0.2},
				{Type: RuleTypeHeader, Operator: OperatorContains, Value: "text/html", Weight: 0.4},
				{Type: RuleTypeUserAgent, Operator: OperatorRegex, Value: `Mozilla|Chrome|Safari|Firefox`, Weight: 0.4},
			},
		},
		// 文件下载模式
		{
			Type:        PatternTypeDownload,
			Name:        "File Download",
			Description: "文件下载",
			Priority:    15,
			Confidence:  0.8,
			Enabled:     true,
			Rules: []Rule{
				{Type: RuleTypeMethod, Operator: OperatorEquals, Value: "GET", Weight: 0.2},
				{Type: RuleTypePath, Operator: OperatorRegex, Value: `\.(zip|rar|exe|dmg|pkg|deb|rpm|tar|gz|pdf|doc|xls|ppt)$`, Weight: 0.5},
				{Type: RuleTypeHeader, Operator: OperatorContains, Value: "application/octet-stream", Weight: 0.3},
			},
		},
		// 爬虫/机器人模式
		{
			Type:        PatternTypeBot,
			Name:        "Bot/Crawler",
			Description: "爬虫或机器人访问",
			Priority:    5,
			Confidence:  0.9,
			Enabled:     true,
			Rules: []Rule{
				{Type: RuleTypeUserAgent, Operator: OperatorRegex, Value: `(?i)(bot|crawler|spider|scraper|curl|wget|python|java)`, Weight: 0.6},
				{Type: RuleTypeHeader, Operator: OperatorContains, Value: "bot", Weight: 0.4},
			},
		},
		// 攻击行为模式
		{
			Type:        PatternTypeAttack,
			Name:        "Attack Pattern",
			Description: "可疑攻击行为",
			Priority:    1,
			Confidence:  0.8,
			Enabled:     true,
			Rules: []Rule{
				{Type: RuleTypePath, Operator: OperatorRegex, Value: `(?i)(union|select|insert|update|delete|drop|exec|script|alert|onload|onerror)`, Weight: 0.4},
				{Type: RuleTypeQuery, Operator: OperatorRegex, Value: `(?i)(\.\./|\.\.\\|/etc/passwd|cmd\.exe|powershell)`, Weight: 0.4},
				{Type: RuleTypeUserAgent, Operator: OperatorRegex, Value: `(?i)(sqlmap|nmap|nikto|burp|zap)`, Weight: 0.2},
			},
		},
		// 静态资源模式
		{
			Type:        PatternTypeStatic,
			Name:        "Static Resource",
			Description: "静态资源访问",
			Priority:    30,
			Confidence:  0.9,
			Enabled:     true,
			Rules: []Rule{
				{Type: RuleTypeMethod, Operator: OperatorEquals, Value: "GET", Weight: 0.2},
				{Type: RuleTypePath, Operator: OperatorRegex, Value: `\.(css|js|png|jpg|jpeg|gif|svg|ico|woff|woff2|ttf|eot)$`, Weight: 0.6},
				{Type: RuleTypeContentType, Operator: OperatorRegex, Value: `^(text/css|application/javascript|image/|font/)`, Weight: 0.2},
			},
		},
		// 认证相关模式
		{
			Type:        PatternTypeAuth,
			Name:        "Authentication",
			Description: "认证相关请求",
			Priority:    8,
			Confidence:  0.8,
			Enabled:     true,
			Rules: []Rule{
				{Type: RuleTypePath, Operator: OperatorRegex, Value: `(?i)/(login|logout|auth|signin|signup|register|oauth)`, Weight: 0.5},
				{Type: RuleTypeMethod, Operator: OperatorEquals, Value: "POST", Weight: 0.3},
				{Type: RuleTypeContentType, Operator: OperatorContains, Value: "application/x-www-form-urlencoded", Weight: 0.2},
			},
		},
	}

	// 编译正则表达式
	for _, pattern := range defaultPatterns {
		for i := range pattern.Rules {
			if pattern.Rules[i].Operator == OperatorRegex {
				if regex, err := regexp.Compile(pattern.Rules[i].Value); err == nil {
					pattern.Rules[i].Regex = regex
				} else {
					logger.Errorf("编译正则表达式失败: %s, 错误: %v", pattern.Rules[i].Value, err)
				}
			}
		}
		pr.patterns = append(pr.patterns, pattern)
	}

	logger.Infof("加载了 %d 个默认流量模式", len(defaultPatterns))
}

// RecognizePattern 识别流量模式
func (pr *PatternRecognizer) RecognizePattern(info *TrafficInfo) *PatternResult {
	if !pr.enabled {
		return &PatternResult{
			Pattern:    nil,
			Matched:    false,
			Confidence: 0,
			Timestamp:  time.Now(),
		}
	}

	startTime := time.Now()
	atomic.AddInt64(&pr.stats.TotalRequests, 1)

	pr.mutex.RLock()
	patterns := make([]*TrafficPattern, len(pr.patterns))
	copy(patterns, pr.patterns)
	pr.mutex.RUnlock()

	var bestMatch *PatternResult
	bestConfidence := 0.0

	// 按优先级排序检查模式
	for _, pattern := range patterns {
		if !pattern.Enabled {
			continue
		}

		result := pr.matchPattern(pattern, info)
		if result.Matched && result.Confidence > bestConfidence && result.Confidence >= pr.confidence {
			bestConfidence = result.Confidence
			bestMatch = result
		}
	}

	// 更新统计信息
	processingTime := time.Since(startTime)
	pr.stats.ProcessingTime = processingTime
	pr.stats.LastUpdate = time.Now()

	if bestMatch != nil {
		atomic.AddInt64(&pr.stats.RecognizedCount, 1)
		// 修复：使用临时变量来避免地址操作错误
		if _, exists := pr.stats.PatternStats[bestMatch.Pattern.Type]; !exists {
			pr.stats.PatternStats[bestMatch.Pattern.Type] = 0
		}
		pr.stats.PatternStats[bestMatch.Pattern.Type]++

		logger.Debugf("识别到流量模式: %s, 置信度: %.2f, 处理时间: %v",
			bestMatch.Pattern.Name, bestMatch.Confidence, processingTime)
		return bestMatch
	}

	// 未识别的模式
	return &PatternResult{
		Pattern: &TrafficPattern{
			Type: PatternTypeUnknown,
			Name: "Unknown Pattern",
		},
		Matched:    false,
		Confidence: 0,
		Timestamp:  time.Now(),
	}
}

// matchPattern 匹配单个模式
func (pr *PatternRecognizer) matchPattern(pattern *TrafficPattern, info *TrafficInfo) *PatternResult {
	result := &PatternResult{
		Pattern:      pattern,
		Matched:      false,
		Confidence:   0,
		MatchedRules: make([]string, 0),
		Timestamp:    time.Now(),
	}

	totalWeight := 0.0
	matchedWeight := 0.0
	requiredMatched := true

	for _, rule := range pattern.Rules {
		matched := pr.matchRule(&rule, info)
		totalWeight += rule.Weight

		if matched {
			matchedWeight += rule.Weight
			result.MatchedRules = append(result.MatchedRules, fmt.Sprintf("%s:%s", rule.Type, rule.Value))
		} else if rule.Required {
			requiredMatched = false
			break
		}
	}

	// 计算置信度
	if requiredMatched && totalWeight > 0 {
		confidence := (matchedWeight / totalWeight) * pattern.Confidence
		if confidence >= pr.confidence {
			result.Matched = true
			result.Confidence = confidence
		}
	}

	return result
}

// matchRule 匹配单个规则
func (pr *PatternRecognizer) matchRule(rule *Rule, info *TrafficInfo) bool {
	var fieldValue string

	// 获取字段值
	switch rule.Type {
	case RuleTypeMethod:
		fieldValue = info.Method
	case RuleTypePath:
		fieldValue = info.Path
	case RuleTypeQuery:
		fieldValue = info.Query
	case RuleTypeUserAgent:
		fieldValue = info.UserAgent
	case RuleTypeContentType:
		fieldValue = info.ContentType
	case RuleTypeReferer:
		fieldValue = info.Referer
	case RuleTypeHost:
		fieldValue = info.Host
	case RuleTypeStatusCode:
		fieldValue = fmt.Sprintf("%d", info.StatusCode)
	case RuleTypeHeader:
		if info.Headers != nil {
			for key, value := range info.Headers {
				if strings.Contains(strings.ToLower(key), strings.ToLower(rule.Field)) ||
					strings.Contains(strings.ToLower(value), strings.ToLower(rule.Value)) {
					fieldValue = value
					break
				}
			}
		}
	case RuleTypeSize:
		if strings.Contains(rule.Field, "request") {
			fieldValue = fmt.Sprintf("%d", info.RequestSize)
		} else {
			fieldValue = fmt.Sprintf("%d", info.ResponseSize)
		}
	default:
		return false
	}

	// 执行匹配
	return pr.executeMatch(rule.Operator, fieldValue, rule.Value, rule.Regex)
}

// executeMatch 执行匹配操作
func (pr *PatternRecognizer) executeMatch(operator Operator, fieldValue, ruleValue string, regex *regexp.Regexp) bool {
	switch operator {
	case OperatorEquals:
		return strings.EqualFold(fieldValue, ruleValue)
	case OperatorContains:
		return strings.Contains(strings.ToLower(fieldValue), strings.ToLower(ruleValue))
	case OperatorStartsWith:
		return strings.HasPrefix(strings.ToLower(fieldValue), strings.ToLower(ruleValue))
	case OperatorEndsWith:
		return strings.HasSuffix(strings.ToLower(fieldValue), strings.ToLower(ruleValue))
	case OperatorRegex:
		if regex != nil {
			return regex.MatchString(fieldValue)
		}
		return false
	case OperatorGreater:
		if fieldVal, err := parseNumber(fieldValue); err == nil {
			if ruleVal, err := parseNumber(ruleValue); err == nil {
				return fieldVal > ruleVal
			}
		}
		return false
	case OperatorLess:
		if fieldVal, err := parseNumber(fieldValue); err == nil {
			if ruleVal, err := parseNumber(ruleValue); err == nil {
				return fieldVal < ruleVal
			}
		}
		return false
	case OperatorIn:
		values := strings.Split(ruleValue, ",")
		for _, v := range values {
			if strings.EqualFold(strings.TrimSpace(v), fieldValue) {
				return true
			}
		}
		return false
	case OperatorNotIn:
		values := strings.Split(ruleValue, ",")
		for _, v := range values {
			if strings.EqualFold(strings.TrimSpace(v), fieldValue) {
				return false
			}
		}
		return true
	default:
		return false
	}
}

// parseNumber 解析数字
func parseNumber(s string) (float64, error) {
	var f float64
	if n, err := fmt.Sscanf(s, "%f", &f); err == nil && n == 1 {
		return f, nil
	}
	return 0, fmt.Errorf("无法解析数字: %s", s)
}

// ExtractTrafficInfo 从HTTP请求/响应中提取流量信息
func (pr *PatternRecognizer) ExtractTrafficInfo(req *http.Request, resp *http.Response, clientIP string, duration time.Duration) *TrafficInfo {
	info := &TrafficInfo{
		Request:     req,
		Response:    resp,
		Method:      req.Method,
		URL:         req.URL.String(),
		Path:        req.URL.Path,
		Query:       req.URL.RawQuery,
		Headers:     make(map[string]string),
		UserAgent:   req.UserAgent(),
		ContentType: req.Header.Get("Content-Type"),
		Referer:     req.Header.Get("Referer"),
		Host:        req.Host,
		RequestSize: req.ContentLength,
		Duration:    duration,
		ClientIP:    clientIP,
		Timestamp:   time.Now(),
	}

	// 提取请求头
	for name, values := range req.Header {
		if len(values) > 0 {
			info.Headers[name] = values[0]
		}
	}

	// 提取响应信息
	if resp != nil {
		info.StatusCode = resp.StatusCode
		info.ResponseSize = resp.ContentLength
	}

	return info
}

// AddPattern 添加自定义模式
func (pr *PatternRecognizer) AddPattern(pattern *TrafficPattern) error {
	if pattern == nil {
		return fmt.Errorf("模式不能为空")
	}

	// 编译正则表达式
	for i := range pattern.Rules {
		if pattern.Rules[i].Operator == OperatorRegex {
			regex, err := regexp.Compile(pattern.Rules[i].Value)
			if err != nil {
				return fmt.Errorf("编译正则表达式失败: %w", err)
			}
			pattern.Rules[i].Regex = regex
		}
	}

	pr.mutex.Lock()
	defer pr.mutex.Unlock()

	pr.patterns = append(pr.patterns, pattern)
	logger.Infof("添加自定义流量模式: %s", pattern.Name)
	return nil
}

// RemovePattern 移除模式
func (pr *PatternRecognizer) RemovePattern(patternType PatternType) error {
	pr.mutex.Lock()
	defer pr.mutex.Unlock()

	for i, pattern := range pr.patterns {
		if pattern.Type == patternType {
			pr.patterns = append(pr.patterns[:i], pr.patterns[i+1:]...)
			logger.Infof("移除流量模式: %s", pattern.Name)
			return nil
		}
	}

	return fmt.Errorf("未找到模式: %s", patternType)
}

// GetPatterns 获取所有模式
func (pr *PatternRecognizer) GetPatterns() []*TrafficPattern {
	pr.mutex.RLock()
	defer pr.mutex.RUnlock()

	patterns := make([]*TrafficPattern, len(pr.patterns))
	copy(patterns, pr.patterns)
	return patterns
}

// GetStats 获取统计信息
func (pr *PatternRecognizer) GetStats() *RecognizerStats {
	return &RecognizerStats{
		TotalRequests:   atomic.LoadInt64(&pr.stats.TotalRequests),
		RecognizedCount: atomic.LoadInt64(&pr.stats.RecognizedCount),
		PatternStats:    pr.copyPatternStats(),
		LastUpdate:      pr.stats.LastUpdate,
		ProcessingTime:  pr.stats.ProcessingTime,
	}
}

// copyPatternStats 复制模式统计信息
func (pr *PatternRecognizer) copyPatternStats() map[PatternType]int64 {
	result := make(map[PatternType]int64)
	for k, v := range pr.stats.PatternStats {
		result[k] = v
	}
	return result
}

// SetEnabled 设置启用状态
func (pr *PatternRecognizer) SetEnabled(enabled bool) {
	pr.enabled = enabled
	logger.Infof("流量模式识别器状态: %v", enabled)
}

// SetConfidenceThreshold 设置置信度阈值
func (pr *PatternRecognizer) SetConfidenceThreshold(threshold float64) {
	if threshold >= 0 && threshold <= 1 {
		pr.confidence = threshold
		logger.Infof("设置置信度阈值: %.2f", threshold)
	}
}

// IsEnabled 检查是否启用
func (pr *PatternRecognizer) IsEnabled() bool {
	return pr.enabled
}

// LoadPatternsFromJSON 从JSON加载模式
func (pr *PatternRecognizer) LoadPatternsFromJSON(jsonData []byte) error {
	var patterns []*TrafficPattern
	if err := json.Unmarshal(jsonData, &patterns); err != nil {
		return fmt.Errorf("解析JSON失败: %w", err)
	}

	pr.mutex.Lock()
	defer pr.mutex.Unlock()

	// 编译正则表达式
	for _, pattern := range patterns {
		for i := range pattern.Rules {
			if pattern.Rules[i].Operator == OperatorRegex {
				regex, err := regexp.Compile(pattern.Rules[i].Value)
				if err != nil {
					logger.Errorf("编译正则表达式失败: %s, 错误: %v", pattern.Rules[i].Value, err)
					continue
				}
				pattern.Rules[i].Regex = regex
			}
		}
	}

	pr.patterns = patterns
	logger.Infof("从JSON加载了 %d 个流量模式", len(patterns))
	return nil
}

// ExportPatternsToJSON 导出模式到JSON
func (pr *PatternRecognizer) ExportPatternsToJSON() ([]byte, error) {
	pr.mutex.RLock()
	defer pr.mutex.RUnlock()

	return json.MarshalIndent(pr.patterns, "", "  ")
}
