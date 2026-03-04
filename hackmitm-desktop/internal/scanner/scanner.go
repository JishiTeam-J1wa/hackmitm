// Package scanner implements passive vulnerability scanning for HTTP traffic
package scanner

import (
	"regexp"
	"strings"
	"sync"
	"time"
)

// Severity levels for vulnerabilities
type Severity string

const (
	SeverityHigh   Severity = "high"
	SeverityMedium Severity = "medium"
	SeverityLow    Severity = "low"
	SeverityInfo   Severity = "info"
)

// ScanRule defines a vulnerability detection rule
type ScanRule struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Severity    Severity `json:"severity"`
	Pattern     string   `json:"pattern"`
	Location    string   `json:"location"` // "request", "response", "both", "headers"
	Description string   `json:"description"`
	Remediation string   `json:"remediation"`
	Category    string   `json:"category"`
	Enabled     bool     `json:"enabled"`

	// Compiled regex (internal)
	compiledPattern *regexp.Regexp `json:"-"`
}

// Vulnerability represents a detected vulnerability
type Vulnerability struct {
	ID           string            `json:"id"`
	RuleID       string            `json:"ruleId"`
	Name         string            `json:"name"`
	Severity     Severity          `json:"severity"`
	Description  string            `json:"description"`
	Remediation  string            `json:"remediation"`
	URL          string            `json:"url"`
	Method       string            `json:"method"`
	Evidence     string            `json:"evidence"`
	Location     string            `json:"location"`
	Request      string            `json:"request"`
	Response     string            `json:"response"`
	Headers      map[string]string `json:"headers"`
	Timestamp    time.Time         `json:"timestamp"`
	FalsePositive bool             `json:"falsePositive"`
}

// HTTPMessage represents an HTTP request or response for scanning
type HTTPMessage struct {
	URL       string            `json:"url"`
	Method    string            `json:"method"`
	Headers   map[string]string `json:"headers"`
	Body      string            `json:"body"`
	StatusCode int              `json:"statusCode"`
	IsRequest  bool             `json:"isRequest"`
}

// Scanner implements passive vulnerability scanning
type Scanner struct {
	rules    []*ScanRule
	findings []*Vulnerability
	mu       sync.RWMutex
	enabled  bool

	// Callback when vulnerability is found
	OnVulnerability func(*Vulnerability)
}

// NewScanner creates a new scanner with default rules
func NewScanner() *Scanner {
	s := &Scanner{
		rules:    getDefaultRules(),
		findings: make([]*Vulnerability, 0),
		enabled:  true,
	}
	// Compile all regex patterns
	for _, rule := range s.rules {
		if rule.Pattern != "" {
			rule.compiledPattern = regexp.MustCompile(rule.Pattern)
		}
	}
	return s
}

// getDefaultRules returns built-in vulnerability detection rules
func getDefaultRules() []*ScanRule {
	return []*ScanRule{
		// SQL Injection patterns
		{
			ID:          "SQL-INJECTION-001",
			Name:        "Potential SQL Injection",
			Severity:    SeverityHigh,
			Pattern:     `(?i)(\b(SELECT|INSERT|UPDATE|DELETE|DROP|UNION|ALTER|CREATE|TRUNCATE)\b.*\b(FROM|INTO|TABLE|DATABASE)\b|'.*(OR|AND)\s*'.*=)`,
			Location:    "request",
			Description: "Request parameter may be vulnerable to SQL injection",
			Remediation: "Use parameterized queries or prepared statements",
			Category:    "injection",
			Enabled:     true,
		},
		// XSS patterns
		{
			ID:          "XSS-001",
			Name:        "Potential Cross-Site Scripting (XSS)",
			Severity:    SeverityHigh,
			Pattern:     `(?i)<script[^>]*>.*?</script>|javascript:|on\w+\s*=|<img[^>]+onerror`,
			Location:    "request",
			Description: "Request contains potential XSS payload",
			Remediation: "Implement proper output encoding and Content-Security-Policy",
			Category:    "xss",
			Enabled:     true,
		},
		// Credit card detection
		{
			ID:          "SENSITIVE-001",
			Name:        "Credit Card Number Exposure",
			Severity:    SeverityHigh,
			Pattern:     `\b(?:\d{4}[-\s]?){3}\d{4}\b`,
			Location:    "response",
			Description: "Response may contain credit card numbers",
			Remediation: "Mask or tokenize sensitive payment data",
			Category:    "sensitive-data",
			Enabled:     true,
		},
		// API key patterns
		{
			ID:          "SENSITIVE-002",
			Name:        "API Key/Secret Exposure",
			Severity:    SeverityHigh,
			Pattern:     `(?i)(api[_-]?key|secret[_-]?key|access[_-]?token|bearer)\s*[=:]\s*['"]?[a-zA-Z0-9_\-]{20,}['"]?`,
			Location:    "response",
			Description: "Response may contain API keys or secrets",
			Remediation: "Remove sensitive credentials from responses",
			Category:    "sensitive-data",
			Enabled:     true,
		},
		// Password in response
		{
			ID:          "SENSITIVE-003",
			Name:        "Password in Response",
			Severity:    SeverityHigh,
			Pattern:     `(?i)["']password["']\s*:\s*["'][^"']+["']`,
			Location:    "response",
			Description: "Response contains password field with value",
			Remediation: "Never return passwords in API responses",
			Category:    "sensitive-data",
			Enabled:     true,
		},
		// Missing security headers
		{
			ID:          "HEADERS-001",
			Name:        "Missing X-Frame-Options Header",
			Severity:    SeverityMedium,
			Pattern:     ``,
			Location:    "headers",
			Description: "Response does not include X-Frame-Options header",
			Remediation: "Add X-Frame-Options: DENY or SAMEORIGIN header",
			Category:    "security-headers",
			Enabled:     true,
		},
		{
			ID:          "HEADERS-002",
			Name:        "Missing Content-Security-Policy Header",
			Severity:    SeverityMedium,
			Pattern:     ``,
			Location:    "headers",
			Description: "Response does not include Content-Security-Policy header",
			Remediation: "Add Content-Security-Policy header with appropriate directives",
			Category:    "security-headers",
			Enabled:     true,
		},
		{
			ID:          "HEADERS-003",
			Name:        "Missing X-Content-Type-Options Header",
			Severity:    SeverityLow,
			Pattern:     ``,
			Location:    "headers",
			Description: "Response does not include X-Content-Type-Options header",
			Remediation: "Add X-Content-Type-Options: nosniff header",
			Category:    "security-headers",
			Enabled:     true,
		},
		// Server version disclosure
		{
			ID:          "INFO-001",
			Name:        "Server Version Disclosure",
			Severity:    SeverityLow,
			Pattern:     `(?i)(Server|X-Powered-By|X-AspNet-Version):\s*[a-zA-Z0-9_\-\./]+`,
			Location:    "headers",
			Description: "Server header reveals technology version",
			Remediation: "Remove or obfuscate server version headers",
			Category:    "information-disclosure",
			Enabled:     true,
		},
		// Debug information
		{
			ID:          "INFO-002",
			Name:        "Debug Information Exposure",
			Severity:    SeverityMedium,
			Pattern:     `(?i)(stack\s*trace|debug\s*mode|error\s*reporting|phpinfo)`,
			Location:    "response",
			Description: "Response may contain debug information",
			Remediation: "Disable debug mode in production environments",
			Category:    "information-disclosure",
			Enabled:     true,
		},
		// Path traversal
		{
			ID:          "TRAVERSAL-001",
			Name:        "Path Traversal Attempt",
			Severity:    SeverityHigh,
			Pattern:     `(\.\.\/|\.\.\\|%2e%2e%2f|%2e%2e\/)`,
			Location:    "request",
			Description: "Request contains path traversal patterns",
			Remediation: "Validate and sanitize file path inputs",
			Category:    "injection",
			Enabled:     true,
		},
		// Open redirect
		{
			ID:          "REDIRECT-001",
			Name:        "Open Redirect Potential",
			Severity:    SeverityMedium,
			Pattern:     `(?i)(redirect|next|url|return|goto|dest|target)\s*[=:]\s*['"]?https?://`,
			Location:    "request",
			Description: "Request parameter may enable open redirect",
			Remediation: "Validate redirect URLs against whitelist",
			Category:    "redirect",
			Enabled:     true,
		},
	}
}

// ScanRequest scans an HTTP request for vulnerabilities
func (s *Scanner) ScanRequest(msg *HTTPMessage) []*Vulnerability {
	return s.Scan(msg, true)
}

// ScanResponse scans an HTTP response for vulnerabilities
func (s *Scanner) ScanResponse(msg *HTTPMessage) []*Vulnerability {
	return s.Scan(msg, false)
}

// Scan performs vulnerability scanning on an HTTP message
func (s *Scanner) Scan(msg *HTTPMessage, isRequest bool) []*Vulnerability {
	if !s.enabled {
		return nil
	}

	var findings []*Vulnerability

	for _, rule := range s.rules {
		if !rule.Enabled {
			continue
		}

		// Check if rule applies to this message type
		if !s.ruleAppliesTo(rule, isRequest) {
			continue
		}

		var content string
		switch rule.Location {
		case "headers":
			findings = append(findings, s.scanHeaders(msg, rule)...)
			continue
		case "request", "response", "both":
			content = s.getContentForLocation(msg, rule.Location, isRequest)
		default:
			content = msg.Body
		}

		if rule.compiledPattern == nil {
			continue
		}

		matches := rule.compiledPattern.FindAllString(content, -1)
		for _, match := range matches {
			vuln := &Vulnerability{
				ID:          generateVulnID(rule.ID, msg.URL),
				RuleID:      rule.ID,
				Name:        rule.Name,
				Severity:    rule.Severity,
				Description: rule.Description,
				Remediation: rule.Remediation,
				URL:         msg.URL,
				Method:      msg.Method,
				Evidence:    truncateEvidence(match, 500),
				Location:    rule.Location,
				Timestamp:   time.Now(),
			}

			findings = append(findings, vuln)
			s.addFinding(vuln)
		}
	}

	return findings
}

// scanHeaders checks for missing security headers
func (s *Scanner) scanHeaders(msg *HTTPMessage, rule *ScanRule) []*Vulnerability {
	var findings []*Vulnerability

	headerName := ""
	switch rule.ID {
	case "HEADERS-001":
		headerName = "X-Frame-Options"
	case "HEADERS-002":
		headerName = "Content-Security-Policy"
	case "HEADERS-003":
		headerName = "X-Content-Type-Options"
	default:
		// For version disclosure, check if header exists and matches pattern
		if rule.compiledPattern != nil {
			for name, value := range msg.Headers {
				headerLine := name + ": " + value
				if rule.compiledPattern.MatchString(headerLine) {
					vuln := &Vulnerability{
						ID:          generateVulnID(rule.ID, msg.URL),
						RuleID:      rule.ID,
						Name:        rule.Name,
						Severity:    rule.Severity,
						Description: rule.Description,
						Remediation: rule.Remediation,
						URL:         msg.URL,
						Method:      msg.Method,
						Evidence:    headerLine,
						Location:    "headers",
						Timestamp:   time.Now(),
					}
					findings = append(findings, vuln)
					s.addFinding(vuln)
				}
			}
		}
		return findings
	}

	// Check for missing header
	if headerName != "" {
		found := false
		for name := range msg.Headers {
			if strings.EqualFold(name, headerName) {
				found = true
				break
			}
		}
		if !found {
			vuln := &Vulnerability{
				ID:          generateVulnID(rule.ID, msg.URL),
				RuleID:      rule.ID,
				Name:        rule.Name,
				Severity:    rule.Severity,
				Description: rule.Description,
				Remediation: rule.Remediation,
				URL:         msg.URL,
				Method:      msg.Method,
				Evidence:    "Missing: " + headerName,
				Location:    "headers",
				Timestamp:   time.Now(),
			}
			findings = append(findings, vuln)
			s.addFinding(vuln)
		}
	}

	return findings
}

// ruleAppliesTo checks if a rule applies to the message type
func (s *Scanner) ruleAppliesTo(rule *ScanRule, isRequest bool) bool {
	switch rule.Location {
	case "request":
		return isRequest
	case "response":
		return !isRequest
	case "both", "headers":
		return !isRequest // headers check is for response
	default:
		return true
	}
}

// getContentForLocation gets the content to scan based on location
func (s *Scanner) getContentForLocation(msg *HTTPMessage, location string, isRequest bool) string {
	switch location {
	case "request":
		if isRequest {
			return msg.Body
		}
	case "response":
		if !isRequest {
			return msg.Body
		}
	case "both":
		return msg.Body
	}
	return ""
}

// addFinding adds a vulnerability finding to the scanner
func (s *Scanner) addFinding(vuln *Vulnerability) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.findings = append(s.findings, vuln)

	// Call callback if set
	if s.OnVulnerability != nil {
		go s.OnVulnerability(vuln)
	}
}

// GetFindings returns all vulnerability findings
func (s *Scanner) GetFindings() []*Vulnerability {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.findings
}

// ClearFindings clears all stored findings
func (s *Scanner) ClearFindings() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.findings = make([]*Vulnerability, 0)
}

// GetRules returns all scan rules
func (s *Scanner) GetRules() []*ScanRule {
	return s.rules
}

// EnableRule enables a specific rule
func (s *Scanner) EnableRule(ruleID string) {
	for _, rule := range s.rules {
		if rule.ID == ruleID {
			rule.Enabled = true
			break
		}
	}
}

// DisableRule disables a specific rule
func (s *Scanner) DisableRule(ruleID string) {
	for _, rule := range s.rules {
		if rule.ID == ruleID {
			rule.Enabled = false
			break
		}
	}
}

// SetEnabled enables or disables the scanner
func (s *Scanner) SetEnabled(enabled bool) {
	s.enabled = enabled
}

// IsEnabled returns whether the scanner is enabled
func (s *Scanner) IsEnabled() bool {
	return s.enabled
}

// Helper functions

func generateVulnID(ruleID, url string) string {
	return ruleID + "-" + url + "-" + time.Now().Format("20060102150405")
}

func truncateEvidence(evidence string, maxLen int) string {
	if len(evidence) <= maxLen {
		return evidence
	}
	return evidence[:maxLen] + "..."
}
