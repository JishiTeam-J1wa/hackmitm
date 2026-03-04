// Package main 安全检查插件
package main

import (
	"context"
	"fmt"
	"net/http"
	"regexp"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"hackmitm/pkg/logger"
	"hackmitm/pkg/plugin"
)

// SecurityPlugin 安全检查插件
type SecurityPlugin struct {
	*plugin.BasePlugin
	// 配置项
	config struct {
		EnableDebug bool
		// SQL注入检测
		SQLInjectionCheck bool
		// XSS检测
		XSSCheck bool
		// 路径遍历检测
		PathTraversalCheck bool
		// 命令注入检测
		CommandInjectionCheck bool
		// 敏感文件检测
		SensitiveFileCheck bool
		// 黑名单路径
		BlacklistPaths []string
		// 黑名单IP
		BlacklistIPs []string
		// 请求频率限制
		RateLimit struct {
			Enabled     bool
			MaxRequests int
			TimeWindow  int // 秒
		}
	}
	// 状态数据
	stats struct {
		sync.RWMutex
		totalChecks     int64
		blockedRequests int64
		threats         map[string]int64 // 威胁类型统计
		lastUpdate      time.Time
		// 频率限制
		requestCounts map[string]*requestCounter
	}
}

// requestCounter 请求计数器
type requestCounter struct {
	count     int
	startTime time.Time
}

// NewPlugin 创建插件实例
func NewPlugin(config map[string]interface{}) (plugin.Plugin, error) {
	p := &SecurityPlugin{
		BasePlugin: plugin.NewBasePlugin("security", "1.0.0", "安全控制插件"),
	}
	return p, nil
}

// Initialize 初始化插件
func (p *SecurityPlugin) Initialize(config map[string]interface{}) error {
	if err := p.BasePlugin.Initialize(config); err != nil {
		return fmt.Errorf("初始化基础插件失败: %w", err)
	}

	// 加载配置
	p.config.EnableDebug = p.GetConfigBool("enable_debug", false)
	p.config.SQLInjectionCheck = p.GetConfigBool("sql_injection_check", true)
	p.config.XSSCheck = p.GetConfigBool("xss_check", true)
	p.config.PathTraversalCheck = p.GetConfigBool("path_traversal_check", true)
	p.config.CommandInjectionCheck = p.GetConfigBool("command_injection_check", true)
	p.config.SensitiveFileCheck = p.GetConfigBool("sensitive_file_check", true)

	// 加载黑名单
	if blacklistPaths, ok := config["blacklist_paths"].([]interface{}); ok {
		for _, path := range blacklistPaths {
			if strPath, ok := path.(string); ok {
				p.config.BlacklistPaths = append(p.config.BlacklistPaths, strPath)
			}
		}
	}
	if blacklistIPs, ok := config["blacklist_ips"].([]interface{}); ok {
		for _, ip := range blacklistIPs {
			if strIP, ok := ip.(string); ok {
				p.config.BlacklistIPs = append(p.config.BlacklistIPs, strIP)
			}
		}
	}

	// 加载频率限制配置
	p.config.RateLimit.Enabled = p.GetConfigBool("rate_limit.enabled", true)
	p.config.RateLimit.MaxRequests = p.GetConfigInt("rate_limit.max_requests", 100)
	p.config.RateLimit.TimeWindow = p.GetConfigInt("rate_limit.time_window", 60)

	// 初始化状态
	p.stats.threats = make(map[string]int64)
	p.stats.requestCounts = make(map[string]*requestCounter)

	logger.Infof("安全插件初始化完成，SQL注入检测: %v, XSS检测: %v, 路径遍历检测: %v",
		p.config.SQLInjectionCheck, p.config.XSSCheck, p.config.PathTraversalCheck)
	return nil
}

// Start 启动插件
func (p *SecurityPlugin) Start(ctx context.Context) error {
	logger.Info("安全插件启动")
	// 启动清理任务
	go p.cleanupTask(ctx)
	return nil
}

// Stop 停止插件
func (p *SecurityPlugin) Stop(ctx context.Context) error {
	logger.Info("安全插件停止")
	return nil
}

// Priority 返回插件优先级
func (p *SecurityPlugin) Priority() int {
	return p.GetConfigInt("priority", 100) // 安全检查优先级较高
}

// ShouldAllow 判断请求是否允许通过
func (p *SecurityPlugin) ShouldAllow(req *http.Request, ctx *plugin.FilterContext) (bool, error) {
	atomic.AddInt64(&p.stats.totalChecks, 1)

	// 检查IP黑名单
	if p.isIPBlocked(ctx.ClientIP) {
		p.recordThreat("ip_blacklist", ctx.ClientIP)
		return false, nil
	}

	// 检查路径黑名单
	if p.isPathBlocked(req.URL.Path) {
		p.recordThreat("path_blacklist", req.URL.Path)
		return false, nil
	}

	// 频率限制检查
	if p.config.RateLimit.Enabled {
		if !p.checkRateLimit(ctx.ClientIP) {
			p.recordThreat("rate_limit", ctx.ClientIP)
			return false, nil
		}
	}

	// SQL注入检查
	if p.config.SQLInjectionCheck && p.detectSQLInjection(req) {
		p.recordThreat("sql_injection", req.URL.String())
		return false, nil
	}

	// XSS检查
	if p.config.XSSCheck && p.detectXSS(req) {
		p.recordThreat("xss", req.URL.String())
		return false, nil
	}

	// 路径遍历检查
	if p.config.PathTraversalCheck && p.detectPathTraversal(req) {
		p.recordThreat("path_traversal", req.URL.Path)
		return false, nil
	}

	// 命令注入检查
	if p.config.CommandInjectionCheck && p.detectCommandInjection(req) {
		p.recordThreat("command_injection", req.URL.String())
		return false, nil
	}

	// 敏感文件检查
	if p.config.SensitiveFileCheck && p.detectSensitiveFile(req) {
		p.recordThreat("sensitive_file", req.URL.Path)
		return false, nil
	}

	return true, nil
}

// ProcessRequest 处理HTTP请求
func (p *SecurityPlugin) ProcessRequest(req *http.Request, ctx *plugin.RequestContext) error {
	if !p.IsStarted() {
		return nil
	}

	// 检查请求是否应该被允许
	allowed, err := p.ShouldAllow(req, &plugin.FilterContext{
		ClientIP:  ctx.ClientIP,
		UserAgent: ctx.UserAgent,
	})

	if err != nil {
		return fmt.Errorf("安全检查失败: %w", err)
	}

	if !allowed {
		return fmt.Errorf("请求被安全策略拒绝")
	}

	return nil
}

// 记录威胁
func (p *SecurityPlugin) recordThreat(threatType, details string) {
	p.stats.Lock()
	defer p.stats.Unlock()

	atomic.AddInt64(&p.stats.blockedRequests, 1)
	p.stats.threats[threatType]++
	p.stats.lastUpdate = time.Now()

	logger.Warnf("[安全插件] 检测到威胁: %s, 详情: %s", threatType, details)
}

// 检查IP是否在黑名单中
func (p *SecurityPlugin) isIPBlocked(ip string) bool {
	for _, blockedIP := range p.config.BlacklistIPs {
		if ip == blockedIP {
			return true
		}
	}
	return false
}

// 检查路径是否在黑名单中
func (p *SecurityPlugin) isPathBlocked(path string) bool {
	for _, blockedPath := range p.config.BlacklistPaths {
		if strings.HasPrefix(path, blockedPath) {
			return true
		}
	}
	return false
}

// 检查请求频率
func (p *SecurityPlugin) checkRateLimit(ip string) bool {
	p.stats.Lock()
	defer p.stats.Unlock()

	now := time.Now()
	counter, exists := p.stats.requestCounts[ip]
	if !exists || now.Sub(counter.startTime).Seconds() > float64(p.config.RateLimit.TimeWindow) {
		p.stats.requestCounts[ip] = &requestCounter{
			count:     1,
			startTime: now,
		}
		return true
	}

	counter.count++
	if counter.count > p.config.RateLimit.MaxRequests {
		return false
	}
	return true
}

// SQL注入检测
func (p *SecurityPlugin) detectSQLInjection(req *http.Request) bool {
	patterns := []string{
		"(?i)(\\b(SELECT|INSERT|UPDATE|DELETE|DROP|UNION|ALTER)\\b.*\\b(FROM|INTO|WHERE|TABLE)\\b)",
		"(?i)'.*--",
		"(?i)\\b(OR|AND)\\b.*=.*",
	}
	return p.checkPatterns(req, patterns)
}

// XSS检测
func (p *SecurityPlugin) detectXSS(req *http.Request) bool {
	patterns := []string{
		"(?i)<script.*?>",
		"(?i)javascript:",
		"(?i)on(load|click|mouseover|submit|focus)=",
	}
	return p.checkPatterns(req, patterns)
}

// 路径遍历检测
func (p *SecurityPlugin) detectPathTraversal(req *http.Request) bool {
	patterns := []string{
		"\\.\\./",
		"\\.\\.\\\\",
		"/etc/passwd",
		"c:\\\\windows",
	}
	return p.checkPatterns(req, patterns)
}

// 命令注入检测
func (p *SecurityPlugin) detectCommandInjection(req *http.Request) bool {
	patterns := []string{
		"(?i)[;&|`].*\\$\\(",
		"(?i)\\b(cat|echo|rm|mv|cp|wget|curl)\\b",
	}
	return p.checkPatterns(req, patterns)
}

// 敏感文件检测
func (p *SecurityPlugin) detectSensitiveFile(req *http.Request) bool {
	patterns := []string{
		"(?i)\\.(git|svn|env|config|ini|log|bak|swp)$",
		"(?i)(wp-config\\.php|config\\.php|\\.htaccess|web\\.config)",
	}
	return p.checkPatterns(req, patterns)
}

// 检查请求中是否包含指定模式
func (p *SecurityPlugin) checkPatterns(req *http.Request, patterns []string) bool {
	// 检查URL参数
	query := req.URL.Query()
	for _, values := range query {
		for _, value := range values {
			if p.matchPatterns(value, patterns) {
				return true
			}
		}
	}

	// 检查POST表单
	if err := req.ParseForm(); err == nil {
		for _, values := range req.Form {
			for _, value := range values {
				if p.matchPatterns(value, patterns) {
					return true
				}
			}
		}
	}

	return false
}

// 匹配正则表达式模式
func (p *SecurityPlugin) matchPatterns(value string, patterns []string) bool {
	for _, pattern := range patterns {
		if matched, err := regexp.MatchString(pattern, value); err == nil && matched {
			return true
		}
	}
	return false
}

// cleanupTask 清理过期的请求计数器
func (p *SecurityPlugin) cleanupTask(ctx context.Context) {
	ticker := time.NewTicker(time.Duration(p.config.RateLimit.TimeWindow) * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			p.stats.Lock()
			now := time.Now()
			for ip, counter := range p.stats.requestCounts {
				if now.Sub(counter.startTime).Seconds() > float64(p.config.RateLimit.TimeWindow) {
					delete(p.stats.requestCounts, ip)
				}
			}
			p.stats.Unlock()
		}
	}
}

// GetStatistics 获取统计信息
func (p *SecurityPlugin) GetStatistics() map[string]interface{} {
	p.stats.RLock()
	defer p.stats.RUnlock()

	// 复制威胁统计数据
	threats := make(map[string]int64)
	for k, v := range p.stats.threats {
		threats[k] = v
	}

	return map[string]interface{}{
		"total_checks":            p.stats.totalChecks,
		"blocked_requests":        p.stats.blockedRequests,
		"threats":                 threats,
		"last_update":             p.stats.lastUpdate,
		"active_ips":              len(p.stats.requestCounts),
		"plugin_version":          p.Version(),
		"plugin_enabled":          p.IsStarted(),
		"sql_injection_check":     p.config.SQLInjectionCheck,
		"xss_check":               p.config.XSSCheck,
		"path_traversal_check":    p.config.PathTraversalCheck,
		"command_injection_check": p.config.CommandInjectionCheck,
		"sensitive_file_check":    p.config.SensitiveFileCheck,
		"rate_limit": map[string]interface{}{
			"enabled":      p.config.RateLimit.Enabled,
			"max_requests": p.config.RateLimit.MaxRequests,
			"time_window":  p.config.RateLimit.TimeWindow,
		},
	}
}

// 确保实现了所有必要的接口
var (
	_ plugin.Plugin       = (*SecurityPlugin)(nil)
	_ plugin.FilterPlugin = (*SecurityPlugin)(nil)
)
