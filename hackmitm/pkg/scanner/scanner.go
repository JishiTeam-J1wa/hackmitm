// Package scanner 提供被动扫描引擎功能
// Package scanner provides passive vulnerability scanning capabilities
package scanner

import (
	"context"
	"sync"
	"time"
)

// Scanner 被动扫描引擎接口
// Scanner interface for passive vulnerability scanning
type Scanner interface {
	// Name 返回扫描器名称
	Name() string
	// Scan 扫描 HTTP 流量
	Scan(ctx context.Context, traffic *HTTPTraffic) ([]*Vulnerability, error)
	// Rules 返回扫描器使用的规则列表
	Rules() []Rule
	// AddRule 添加扫描规则
	AddRule(rule Rule) error
	// RemoveRule 移除扫描规则
	RemoveRule(ruleID string) error
	// Start 启动扫描器
	Start(ctx context.Context) error
	// Stop 停止扫描器
	Stop(ctx context.Context) error
}

// HTTPTraffic HTTP 流量数据
// HTTPTraffic represents HTTP request and response data
type HTTPTraffic struct {
	ID         string            `json:"id"`
	SessionID  string            `json:"session_id"`
	URL        string            `json:"url"`
	Method     string            `json:"method"`
	Headers    map[string]string `json:"headers"`
	Body       []byte            `json:"body"`
	Response   *HTTPResponse     `json:"response"`
	Timestamp  time.Time         `json:"timestamp"`
	SourceIP   string            `json:"source_ip"`
	UserAgent  string            `json:"user_agent"`
	ContentType string           `json:"content_type"`
}

// HTTPResponse HTTP 响应数据
// HTTPResponse represents HTTP response data
type HTTPResponse struct {
	StatusCode int               `json:"status_code"`
	Headers    map[string]string `json:"headers"`
	Body       []byte            `json:"body"`
	Size       int64             `json:"size"`
	Duration   time.Duration     `json:"duration"`
}

// Rule 扫描规则接口
// Rule interface for scanning rules
type Rule interface {
	// ID 返回规则唯一标识
	ID() string
	// Name 返回规则名称
	Name() string
	// Description 返回规则描述
	Description() string
	// Severity 返回漏洞严重程度
	Severity() Severity
	// Enabled 返回规则是否启用
	Enabled() bool
	// SetEnabled 设置规则启用状态
	SetEnabled(enabled bool)
	// Match 检查流量是否匹配规则
	Match(traffic *HTTPTraffic) (bool, *MatchResult)
	// Priority 返回规则优先级
	Priority() int
}

// MatchResult 匹配结果
// MatchResult represents the result of a rule match
type MatchResult struct {
	Matched    bool                   `json:"matched"`
	Evidence   string                 `json:"evidence"`
	Confidence float64                `json:"confidence"` // 0.0 - 1.0
	Metadata   map[string]interface{} `json:"metadata"`
}

// Vulnerability 漏洞数据结构
// Vulnerability represents a detected vulnerability
type Vulnerability struct {
	ID           string            `json:"id"`
	SessionID    string            `json:"session_id"`
	TrafficID    string            `json:"traffic_id"`
	RuleID       string            `json:"rule_id"`
	Name         string            `json:"name"`
	Description  string            `json:"description"`
	Severity     Severity          `json:"severity"`
	Confidence   float64           `json:"confidence"` // 0.0 - 1.0
	URL          string            `json:"url"`
	Parameter    string            `json:"parameter"`
	Evidence     string            `json:"evidence"`
	Remediation  string            `json:"remediation"`
	Status       VulnStatus        `json:"status"`
	Timestamp    time.Time         `json:"timestamp"`
	Metadata     map[string]interface{} `json:"metadata"`
	Request      *HTTPRequestData  `json:"request"`
	Response     *HTTPResponseData `json:"response"`
}

// HTTPRequestData HTTP 请求数据（用于漏洞记录）
type HTTPRequestData struct {
	Method  string            `json:"method"`
	URL     string            `json:"url"`
	Headers map[string]string `json:"headers"`
	Body    string            `json:"body"`
}

// HTTPResponseData HTTP 响应数据（用于漏洞记录）
type HTTPResponseData struct {
	StatusCode int               `json:"status_code"`
	Headers    map[string]string `json:"headers"`
	Body       string            `json:"body"`
}

// Severity 漏洞严重程度
// Severity represents vulnerability severity level
type Severity string

const (
	SeverityCritical Severity = "critical"
	SeverityHigh     Severity = "high"
	SeverityMedium   Severity = "medium"
	SeverityLow      Severity = "low"
	SeverityInfo     Severity = "info"
)

// VulnStatus 漏洞状态
// VulnStatus represents vulnerability status
type VulnStatus string

const (
	StatusOpen          VulnStatus = "open"
	StatusConfirmed     VulnStatus = "confirmed"
	StatusFalsePositive VulnStatus = "false_positive"
	StatusFixed         VulnStatus = "fixed"
)

// ScannerConfig 扫描器配置
// ScannerConfig represents scanner configuration
type ScannerConfig struct {
	MaxConcurrent int           `json:"max_concurrent"` // 最大并发扫描数
	QueueSize     int           `json:"queue_size"`     // 任务队列大小
	Timeout       time.Duration `json:"timeout"`        // 扫描超时时间
	Enabled       bool          `json:"enabled"`        // 是否启用扫描
}

// DefaultScannerConfig 返回默认扫描器配置
func DefaultScannerConfig() *ScannerConfig {
	return &ScannerConfig{
		MaxConcurrent: 10,
		QueueSize:     1000,
		Timeout:       30 * time.Second,
		Enabled:       true,
	}
}

// ScanResult 扫描结果
// ScanResult represents the result of a scan operation
type ScanResult struct {
	TrafficID     string          `json:"traffic_id"`
	Vulnerabilities []*Vulnerability `json:"vulnerabilities"`
	Duration      time.Duration   `json:"duration"`
	Error         error           `json:"error"`
	Timestamp     time.Time       `json:"timestamp"`
}

// ScannerStats 扫描器统计信息
// ScannerStats represents scanner statistics
type ScannerStats struct {
	TotalScanned      int64         `json:"total_scanned"`
	TotalVulnsFound   int64         `json:"total_vulns_found"`
	BySeverity        map[Severity]int64 `json:"by_severity"`
	AverageScanTime   time.Duration `json:"average_scan_time"`
	QueueSize         int           `json:"queue_size"`
	ActiveScans       int           `json:"active_scans"`
}

// BaseScanner 基础扫描器实现
// BaseScanner provides base implementation for scanners
type BaseScanner struct {
	name     string
	rules    []Rule
	ruleMap  map[string]Rule
	config   *ScannerConfig
	stats    *ScannerStats
	mu       sync.RWMutex
	ctx      context.Context
	cancel   context.CancelFunc
	running  bool
}

// NewBaseScanner 创建基础扫描器
func NewBaseScanner(name string, config *ScannerConfig) *BaseScanner {
	if config == nil {
		config = DefaultScannerConfig()
	}
	return &BaseScanner{
		name:    name,
		rules:   make([]Rule, 0),
		ruleMap: make(map[string]Rule),
		config:  config,
		stats: &ScannerStats{
			BySeverity: make(map[Severity]int64),
		},
	}
}

func (s *BaseScanner) Name() string {
	return s.name
}

func (s *BaseScanner) Rules() []Rule {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.rules
}

func (s *BaseScanner) AddRule(rule Rule) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.rules = append(s.rules, rule)
	s.ruleMap[rule.ID()] = rule
	return nil
}

func (s *BaseScanner) RemoveRule(ruleID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.ruleMap, ruleID)
	for i, r := range s.rules {
		if r.ID() == ruleID {
			s.rules = append(s.rules[:i], s.rules[i+1:]...)
			break
		}
	}
	return nil
}

func (s *BaseScanner) Stats() *ScannerStats {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.stats
}
