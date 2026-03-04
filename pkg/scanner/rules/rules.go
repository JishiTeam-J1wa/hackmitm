// Package rules 提供扫描规则定义和管理
// Package rules provides scanning rule definitions and management
package rules

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"

	"gopkg.in/yaml.v3"
)

// RuleConfig 规则配置
type RuleConfig struct {
	ID          string         `yaml:"id"`
	Name        string         `yaml:"name"`
	Description string         `yaml:"description"`
	Severity    string         `yaml:"severity"`
	Enabled     bool           `yaml:"enabled"`
	Priority    int            `yaml:"priority"`
	Tags        []string       `yaml:"tags"`
	Matchers    []Matcher      `yaml:"matchers"`
	Remediation string         `yaml:"remediation"`
	References  []string       `yaml:"references"`
}

// Matcher 匹配器配置
type Matcher struct {
	Part      string   `yaml:"part"`       // url, headers, body, params, response
	Type      string   `yaml:"type"`       // regex, contains, equals, starts_with, ends_with
	Pattern   string   `yaml:"pattern"`    // 匹配模式
	Patterns  []string `yaml:"patterns"`   // 多个模式 (OR 关系)
	Condition string   `yaml:"condition"`  // and, or (多个 matcher 之间的关系)
	Negative  bool     `yaml:"negative"`   // 取反
	Group     int      `yaml:"group"`      // 分组编号，同组的用 AND 连接
}

// BaseRule 基础规则实现
type BaseRule struct {
	config      *RuleConfig
	compiled    []*regexp.Regexp
	enabled     bool
	mu          sync.RWMutex
}

// NewBaseRule 创建基础规则
func NewBaseRule(config *RuleConfig) (*BaseRule, error) {
	rule := &BaseRule{
		config:  config,
		enabled: config.Enabled,
	}

	// 预编译正则表达式
	for _, m := range config.Matchers {
		if m.Type == "regex" {
			if m.Pattern != "" {
				re, err := regexp.Compile(m.Pattern)
				if err != nil {
					return nil, fmt.Errorf("invalid regex pattern: %s: %w", m.Pattern, err)
				}
				rule.compiled = append(rule.compiled, re)
			}
			for _, p := range m.Patterns {
				re, err := regexp.Compile(p)
				if err != nil {
					return nil, fmt.Errorf("invalid regex pattern: %s: %w", p, err)
				}
				rule.compiled = append(rule.compiled, re)
			}
		}
	}

	return rule, nil
}

func (r *BaseRule) ID() string {
	return r.config.ID
}

func (r *BaseRule) Name() string {
	return r.config.Name
}

func (r *BaseRule) Description() string {
	return r.config.Description
}

func (r *BaseRule) Severity() string {
	return r.config.Severity
}

func (r *BaseRule) Enabled() bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.enabled
}

func (r *BaseRule) SetEnabled(enabled bool) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.enabled = enabled
}

func (r *BaseRule) Priority() int {
	return r.config.Priority
}

func (r *BaseRule) Remediation() string {
	return r.config.Remediation
}

func (r *BaseRule) Tags() []string {
	return r.config.Tags
}

// MatchParts 匹配多个部分
func (r *BaseRule) MatchParts(parts map[string]string) (bool, string) {
	groupMatches := make(map[int][]bool)

	for i, m := range r.config.Matchers {
		var content string
		switch m.Part {
		case "url":
			content = parts["url"]
		case "headers":
			content = parts["headers"]
		case "body":
			content = parts["body"]
		case "params":
			content = parts["params"]
		case "response":
			content = parts["response"]
		case "all":
			content = strings.Join([]string{
				parts["url"],
				parts["headers"],
				parts["body"],
				parts["params"],
				parts["response"],
			}, "\n")
		default:
			content = parts[m.Part]
		}

		matched, evidence := r.matchSingle(m, content, i)
		group := m.Group
		if group == 0 {
			group = i + 1
		}
		groupMatches[group] = append(groupMatches[group], matched)

		if matched && evidence != "" {
			return true, evidence
		}
	}

	// 检查分组匹配结果
	for _, matches := range groupMatches {
		allMatched := true
		for _, m := range matches {
			if !m {
				allMatched = false
				break
			}
		}
		if allMatched && len(matches) > 0 {
			return true, ""
		}
	}

	return false, ""
}

// matchSingle 单个匹配器匹配
func (r *BaseRule) matchSingle(m Matcher, content string, idx int) (bool, string) {
	if content == "" {
		return false, ""
	}

	var matched bool
	var evidence string

	patterns := m.Patterns
	if m.Pattern != "" {
		patterns = append(patterns, m.Pattern)
	}

	for _, pattern := range patterns {
		switch m.Type {
		case "regex":
			if idx < len(r.compiled) && r.compiled[idx] != nil {
				loc := r.compiled[idx].FindStringIndex(content)
				if loc != nil {
					matched = true
					start := loc[0]
					end := loc[1]
					// 扩展证据范围
					if start > 50 {
						start -= 50
					}
					if end+50 < len(content) {
						end += 50
					}
					evidence = content[start:end]
					break
				}
			}
		case "contains":
			if strings.Contains(content, pattern) {
				matched = true
				evidence = pattern
			}
		case "equals":
			if content == pattern {
				matched = true
				evidence = content
			}
		case "starts_with":
			if strings.HasPrefix(content, pattern) {
				matched = true
				evidence = content[:min(100, len(content))]
			}
		case "ends_with":
			if strings.HasSuffix(content, pattern) {
				matched = true
				evidence = content[max(0, len(content)-100):]
			}
		}

		if matched {
			break
		}
	}

	// 取反
	if m.Negative {
		matched = !matched
	}

	return matched, evidence
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// RuleLoader 规则加载器
type RuleLoader struct {
	ruleDir string
	rules   map[string]*BaseRule
	mu      sync.RWMutex
}

// NewRuleLoader 创建规则加载器
func NewRuleLoader(ruleDir string) *RuleLoader {
	return &RuleLoader{
		ruleDir: ruleDir,
		rules:   make(map[string]*BaseRule),
	}
}

// Load 加载规则
func (l *RuleLoader) Load() error {
	l.mu.Lock()
	defer l.mu.Unlock()

	// 遍历规则目录
	files, err := filepath.Glob(filepath.Join(l.ruleDir, "*.yaml"))
	if err != nil {
		return err
	}

	files2, _ := filepath.Glob(filepath.Join(l.ruleDir, "*.yml"))
	files = append(files, files2...)

	for _, file := range files {
		data, err := os.ReadFile(file)
		if err != nil {
			continue
		}

		var config RuleConfig
		if err := yaml.Unmarshal(data, &config); err != nil {
			continue
		}

		rule, err := NewBaseRule(&config)
		if err != nil {
			continue
		}

		l.rules[config.ID] = rule
	}

	return nil
}

// LoadFromBytes 从字节加载规则
func (l *RuleLoader) LoadFromBytes(data []byte) error {
	var config RuleConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return err
	}

	rule, err := NewBaseRule(&config)
	if err != nil {
		return err
	}

	l.mu.Lock()
	l.rules[config.ID] = rule
	l.mu.Unlock()

	return nil
}

// Get 获取规则
func (l *RuleLoader) Get(id string) *BaseRule {
	l.mu.RLock()
	defer l.mu.RUnlock()
	return l.rules[id]
}

// GetAll 获取所有规则
func (l *RuleLoader) GetAll() []*BaseRule {
	l.mu.RLock()
	defer l.mu.RUnlock()
	rules := make([]*BaseRule, 0, len(l.rules))
	for _, r := range l.rules {
		rules = append(rules, r)
	}
	return rules
}

// Enable 启用规则
func (l *RuleLoader) Enable(id string) error {
	l.mu.Lock()
	defer l.mu.Unlock()
	if rule, ok := l.rules[id]; ok {
		rule.SetEnabled(true)
		return nil
	}
	return fmt.Errorf("rule not found: %s", id)
}

// Disable 禁用规则
func (l *RuleLoader) Disable(id string) error {
	l.mu.Lock()
	defer l.mu.Unlock()
	if rule, ok := l.rules[id]; ok {
		rule.SetEnabled(false)
		return nil
	}
	return fmt.Errorf("rule not found: %s", id)
}

// Count 获取规则数量
func (l *RuleLoader) Count() int {
	l.mu.RLock()
	defer l.mu.RUnlock()
	return len(l.rules)
}

// RuleManager 规则管理器
type RuleManager struct {
	loader    *RuleLoader
	lastLoad  time.Time
	autoReload bool
}

// NewRuleManager 创建规则管理器
func NewRuleManager(ruleDir string) *RuleManager {
	return &RuleManager{
		loader: NewRuleLoader(ruleDir),
	}
}

// Load 加载规则
func (m *RuleManager) Load() error {
	if err := m.loader.Load(); err != nil {
		return err
	}
	m.lastLoad = time.Now()
	return nil
}

// Reload 重载规则
func (m *RuleManager) Reload() error {
	return m.Load()
}

// GetLoader 获取加载器
func (m *RuleManager) GetLoader() *RuleLoader {
	return m.loader
}
