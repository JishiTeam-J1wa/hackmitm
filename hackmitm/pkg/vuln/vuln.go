// Package vuln 提供漏洞管理功能
// Package vuln provides vulnerability management capabilities
package vuln

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"sync"
	"time"
)

// Vulnerability 漏洞数据模型
// Vulnerability represents a security vulnerability
type Vulnerability struct {
	ID           string                 `json:"id"`
	SessionID    string                 `json:"session_id"`
	TrafficID    string                 `json:"traffic_id"`
	RuleID       string                 `json:"rule_id"`
	Hash         string                 `json:"hash"` // 用于去重的唯一哈希
	Name         string                 `json:"name"`
	Description  string                 `json:"description"`
	Severity     Severity               `json:"severity"`
	Confidence   float64                `json:"confidence"`
	URL          string                 `json:"url"`
	Parameter    string                 `json:"parameter"`
	Evidence     string                 `json:"evidence"`
	Remediation  string                 `json:"remediation"`
	Status       Status                 `json:"status"`
	Occurrences  int                    `json:"occurrences"` // 出现次数
	FirstSeen    time.Time              `json:"first_seen"`
	LastSeen     time.Time              `json:"last_seen"`
	Timestamp    time.Time              `json:"timestamp"`
	Metadata     map[string]interface{} `json:"metadata"`
	Request      *HTTPRequestSnapshot  `json:"request"`
	Response     *HTTPResponseSnapshot `json:"response"`
	Tags         []string               `json:"tags"`
	Notes        string                 `json:"notes"`
}

// HTTPRequestSnapshot HTTP 请求快照
type HTTPRequestSnapshot struct {
	Method  string            `json:"method"`
	URL     string            `json:"url"`
	Headers map[string]string `json:"headers"`
	Body    string            `json:"body"`
}

// HTTPResponseSnapshot HTTP 响应快照
type HTTPResponseSnapshot struct {
	StatusCode int               `json:"status_code"`
	Headers    map[string]string `json:"headers"`
	Body       string            `json:"body"`
	Size       int64             `json:"size"`
}

// Severity 漏洞严重程度
type Severity string

const (
	SeverityCritical Severity = "critical"
	SeverityHigh     Severity = "high"
	SeverityMedium   Severity = "medium"
	SeverityLow      Severity = "low"
	SeverityInfo     Severity = "info"
)

// Status 漏洞状态
type Status string

const (
	StatusOpen          Status = "open"
	StatusConfirmed     Status = "confirmed"
	StatusFalsePositive Status = "false_positive"
	StatusFixed         Status = "fixed"
)

// CalculateHash 计算漏洞唯一哈希用于去重
// CalculateHash generates a unique hash for deduplication
func (v *Vulnerability) CalculateHash() string {
	data := fmt.Sprintf("%s|%s|%s", v.URL, v.Parameter, v.RuleID)
	hash := sha256.Sum256([]byte(data))
	return hex.EncodeToString(hash[:16])
}

// GenerateID 生成漏洞 ID
func (v *Vulnerability) GenerateID() string {
	if v.Hash == "" {
		v.Hash = v.CalculateHash()
	}
	timestamp := v.Timestamp.UnixNano()
	data := fmt.Sprintf("%s-%d", v.Hash, timestamp)
	hash := sha256.Sum256([]byte(data))
	return fmt.Sprintf("vuln_%s", hex.EncodeToString(hash[:12]))
}

// ToJSON 转换为 JSON
func (v *Vulnerability) ToJSON() ([]byte, error) {
	return json.Marshal(v)
}

// FromJSON 从 JSON 解析
func FromJSON(data []byte) (*Vulnerability, error) {
	var v Vulnerability
	err := json.Unmarshal(data, &v)
	if err != nil {
		return nil, err
	}
	return &v, nil
}

// Filter 漏洞过滤器
type Filter struct {
	SessionID  string   `json:"session_id"`
	Severity   []Severity `json:"severity"`
	Status     []Status  `json:"status"`
	RuleID     string    `json:"rule_id"`
	URLPattern string    `json:"url_pattern"`
	Search     string    `json:"search"`
	FromDate   time.Time `json:"from_date"`
	ToDate     time.Time `json:"to_date"`
	Tags       []string  `json:"tags"`
}

// Matches 检查漏洞是否匹配过滤器
func (f *Filter) Matches(v *Vulnerability) bool {
	if f.SessionID != "" && v.SessionID != f.SessionID {
		return false
	}
	if len(f.Severity) > 0 {
		found := false
		for _, s := range f.Severity {
			if v.Severity == s {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}
	if len(f.Status) > 0 {
		found := false
		for _, s := range f.Status {
			if v.Status == s {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}
	if f.RuleID != "" && v.RuleID != f.RuleID {
		return false
	}
	if !f.FromDate.IsZero() && v.Timestamp.Before(f.FromDate) {
		return false
	}
	if !f.ToDate.IsZero() && v.Timestamp.After(f.ToDate) {
		return false
	}
	return true
}

// Stats 漏洞统计
type Stats struct {
	Total        int            `json:"total"`
	BySeverity   map[Severity]int `json:"by_severity"`
	ByStatus     map[Status]int   `json:"by_status"`
	ByRule       map[string]int   `json:"by_rule"`
	LastUpdated  time.Time       `json:"last_updated"`
}

// NewStats 创建统计对象
func NewStats() *Stats {
	return &Stats{
		BySeverity: make(map[Severity]int),
		ByStatus:   make(map[Status]int),
		ByRule:     make(map[string]int),
	}
}

// Update 更新统计
func (s *Stats) Update(vulns []*Vulnerability) {
	s.Total = len(vulns)
	s.BySeverity = make(map[Severity]int)
	s.ByStatus = make(map[Status]int)
	s.ByRule = make(map[string]int)

	for _, v := range vulns {
		s.BySeverity[v.Severity]++
		s.ByStatus[v.Status]++
		s.ByRule[v.RuleID]++
	}
	s.LastUpdated = time.Now()
}

// Manager 漏洞管理器接口
type Manager interface {
	// Create 创建漏洞
	Create(vuln *Vulnerability) error
	// Get 获取漏洞
	Get(id string) (*Vulnerability, error)
	// Update 更新漏洞
	Update(vuln *Vulnerability) error
	// Delete 删除漏洞
	Delete(id string) error
	// List 列出漏洞
	List(filter *Filter, offset, limit int) ([]*Vulnerability, int, error)
	// UpdateStatus 更新漏洞状态
	UpdateStatus(id string, status Status, notes string) error
	// GetStats 获取统计信息
	GetStats(sessionID string) (*Stats, error)
	// FindByHash 根据哈希查找漏洞
	FindByHash(hash string) (*Vulnerability, error)
	// IncrementOccurrences 增加出现次数
	IncrementOccurrences(id string) error
}

// MemoryStore 内存存储实现
type MemoryStore struct {
	vulns  map[string]*Vulnerability
	hashes map[string]string // hash -> id
	mu     sync.RWMutex
}

// NewMemoryStore 创建内存存储
func NewMemoryStore() *MemoryStore {
	return &MemoryStore{
		vulns:  make(map[string]*Vulnerability),
		hashes: make(map[string]string),
	}
}

func (s *MemoryStore) Create(vuln *Vulnerability) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if vuln.ID == "" {
		vuln.Hash = vuln.CalculateHash()
		vuln.ID = vuln.GenerateID()
	}

	// 检查重复
	if existingID, exists := s.hashes[vuln.Hash]; exists {
		// 增加现有漏洞的出现次数
		if existing, ok := s.vulns[existingID]; ok {
			existing.Occurrences++
			existing.LastSeen = time.Now()
		}
		return nil
	}

	vuln.FirstSeen = time.Now()
	vuln.LastSeen = time.Now()
	vuln.Occurrences = 1
	if vuln.Timestamp.IsZero() {
		vuln.Timestamp = time.Now()
	}

	s.vulns[vuln.ID] = vuln
	s.hashes[vuln.Hash] = vuln.ID
	return nil
}

func (s *MemoryStore) Get(id string) (*Vulnerability, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	vuln, ok := s.vulns[id]
	if !ok {
		return nil, fmt.Errorf("vulnerability not found: %s", id)
	}
	return vuln, nil
}

func (s *MemoryStore) Update(vuln *Vulnerability) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.vulns[vuln.ID]; !ok {
		return fmt.Errorf("vulnerability not found: %s", vuln.ID)
	}
	s.vulns[vuln.ID] = vuln
	return nil
}

func (s *MemoryStore) Delete(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	vuln, ok := s.vulns[id]
	if !ok {
		return fmt.Errorf("vulnerability not found: %s", id)
	}
	delete(s.hashes, vuln.Hash)
	delete(s.vulns, id)
	return nil
}

func (s *MemoryStore) List(filter *Filter, offset, limit int) ([]*Vulnerability, int, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var result []*Vulnerability
	for _, v := range s.vulns {
		if filter == nil || filter.Matches(v) {
			result = append(result, v)
		}
	}

	total := len(result)
	if offset >= total {
		return []*Vulnerability{}, total, nil
	}

	end := offset + limit
	if end > total || limit <= 0 {
		end = total
	}

	return result[offset:end], total, nil
}

func (s *MemoryStore) UpdateStatus(id string, status Status, notes string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	vuln, ok := s.vulns[id]
	if !ok {
		return fmt.Errorf("vulnerability not found: %s", id)
	}
	vuln.Status = status
	vuln.Notes = notes
	return nil
}

func (s *MemoryStore) GetStats(sessionID string) (*Stats, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	stats := NewStats()
	for _, v := range s.vulns {
		if sessionID == "" || v.SessionID == sessionID {
			stats.Total++
			stats.BySeverity[v.Severity]++
			stats.ByStatus[v.Status]++
			stats.ByRule[v.RuleID]++
		}
	}
	stats.LastUpdated = time.Now()
	return stats, nil
}

func (s *MemoryStore) FindByHash(hash string) (*Vulnerability, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	id, ok := s.hashes[hash]
	if !ok {
		return nil, fmt.Errorf("vulnerability not found for hash: %s", hash)
	}
	return s.vulns[id], nil
}

func (s *MemoryStore) IncrementOccurrences(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	vuln, ok := s.vulns[id]
	if !ok {
		return fmt.Errorf("vulnerability not found: %s", id)
	}
	vuln.Occurrences++
	vuln.LastSeen = time.Now()
	return nil
}
