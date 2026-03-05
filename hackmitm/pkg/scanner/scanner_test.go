package scanner

import (
	"testing"
	"time"
)

// MockRule is a mock implementation of Rule for testing
type MockRule struct {
	id          string
	name        string
	description string
	severity    Severity
	enabled     bool
	priority    int
	matchFunc   func(traffic *HTTPTraffic) (bool, *MatchResult)
}

func (r *MockRule) ID() string          { return r.id }
func (r *MockRule) Name() string        { return r.name }
func (r *MockRule) Description() string { return r.description }
func (r *MockRule) Severity() Severity  { return r.severity }
func (r *MockRule) Enabled() bool       { return r.enabled }
func (r *MockRule) SetEnabled(e bool)   { r.enabled = e }
func (r *MockRule) Priority() int       { return r.priority }
func (r *MockRule) Match(traffic *HTTPTraffic) (bool, *MatchResult) {
	if r.matchFunc != nil {
		return r.matchFunc(traffic)
	}
	return false, &MatchResult{Matched: false}
}

// TestBaseScanner_AddRule tests adding rules to the scanner
func TestBaseScanner_AddRule(t *testing.T) {
	scanner := NewBaseScanner("test-scanner", nil)

	rule := &MockRule{
		id:          "test-rule-1",
		name:        "Test Rule",
		description: "A test rule",
		severity:    SeverityHigh,
		enabled:     true,
		priority:    1,
	}

	err := scanner.AddRule(rule)
	if err != nil {
		t.Fatalf("AddRule failed: %v", err)
	}

	rules := scanner.Rules()
	if len(rules) != 1 {
		t.Errorf("Expected 1 rule, got %d", len(rules))
	}

	if rules[0].ID() != "test-rule-1" {
		t.Errorf("Expected rule ID 'test-rule-1', got '%s'", rules[0].ID())
	}
}

// TestBaseScanner_RemoveRule tests removing rules from the scanner
func TestBaseScanner_RemoveRule(t *testing.T) {
	scanner := NewBaseScanner("test-scanner", nil)

	rule := &MockRule{
		id:      "test-rule-1",
		name:    "Test Rule",
		enabled: true,
	}

	_ = scanner.AddRule(rule)

	err := scanner.RemoveRule("test-rule-1")
	if err != nil {
		t.Fatalf("RemoveRule failed: %v", err)
	}

	rules := scanner.Rules()
	if len(rules) != 0 {
		t.Errorf("Expected 0 rules after removal, got %d", len(rules))
	}
}

// TestBaseScanner_Name tests the scanner name
func TestBaseScanner_Name(t *testing.T) {
	scanner := NewBaseScanner("my-scanner", nil)
	if scanner.Name() != "my-scanner" {
		t.Errorf("Expected name 'my-scanner', got '%s'", scanner.Name())
	}
}

// TestDefaultScannerConfig tests the default configuration
func TestDefaultScannerConfig(t *testing.T) {
	config := DefaultScannerConfig()

	if config.MaxConcurrent != 10 {
		t.Errorf("Expected MaxConcurrent 10, got %d", config.MaxConcurrent)
	}

	if config.QueueSize != 1000 {
		t.Errorf("Expected QueueSize 1000, got %d", config.QueueSize)
	}

	if config.Timeout != 30*time.Second {
		t.Errorf("Expected Timeout 30s, got %v", config.Timeout)
	}

	if !config.Enabled {
		t.Error("Expected Enabled to be true")
	}
}

// TestSeverity tests severity constants
func TestSeverity(t *testing.T) {
	tests := []struct {
		severity Severity
		expected string
	}{
		{SeverityCritical, "critical"},
		{SeverityHigh, "high"},
		{SeverityMedium, "medium"},
		{SeverityLow, "low"},
		{SeverityInfo, "info"},
	}

	for _, tt := range tests {
		if string(tt.severity) != tt.expected {
			t.Errorf("Expected severity '%s', got '%s'", tt.expected, tt.severity)
		}
	}
}

// TestVulnStatus tests vulnerability status constants
func TestVulnStatus(t *testing.T) {
	tests := []struct {
		status   VulnStatus
		expected string
	}{
		{StatusOpen, "open"},
		{StatusConfirmed, "confirmed"},
		{StatusFalsePositive, "false_positive"},
		{StatusFixed, "fixed"},
	}

	for _, tt := range tests {
		if string(tt.status) != tt.expected {
			t.Errorf("Expected status '%s', got '%s'", tt.expected, tt.status)
		}
	}
}

// TestHTTPTraffic tests HTTP traffic structure
func TestHTTPTraffic(t *testing.T) {
	traffic := &HTTPTraffic{
		ID:         "req-123",
		SessionID:  "sess-456",
		URL:        "https://example.com/api/test?id=1",
		Method:     "GET",
		Headers:    map[string]string{"Content-Type": "application/json"},
		Body:       []byte(`{"test": "data"}`),
		Timestamp:  time.Now(),
		SourceIP:   "192.168.1.1",
		UserAgent:  "Mozilla/5.0",
	}

	if traffic.ID != "req-123" {
		t.Errorf("Expected ID 'req-123', got '%s'", traffic.ID)
	}

	if traffic.Method != "GET" {
		t.Errorf("Expected Method 'GET', got '%s'", traffic.Method)
	}
}

// TestHTTPResponse tests HTTP response structure
func TestHTTPResponse(t *testing.T) {
	response := &HTTPResponse{
		StatusCode: 200,
		Headers:    map[string]string{"Content-Type": "application/json"},
		Body:       []byte(`{"status": "ok"}`),
		Size:       19,
		Duration:   100 * time.Millisecond,
	}

	if response.StatusCode != 200 {
		t.Errorf("Expected StatusCode 200, got %d", response.StatusCode)
	}

	if response.Size != 19 {
		t.Errorf("Expected Size 19, got %d", response.Size)
	}
}

// TestMatchResult tests match result structure
func TestMatchResult(t *testing.T) {
	result := &MatchResult{
		Matched:    true,
		Evidence:   "SQL error message detected",
		Confidence: 0.85,
		Metadata: map[string]interface{}{
			"pattern": "UNION SELECT",
		},
	}

	if !result.Matched {
		t.Error("Expected Matched to be true")
	}

	if result.Confidence != 0.85 {
		t.Errorf("Expected Confidence 0.85, got %f", result.Confidence)
	}
}

// TestVulnerability tests vulnerability structure
func TestVulnerability(t *testing.T) {
	vuln := &Vulnerability{
		ID:          "vuln-123",
		SessionID:   "sess-456",
		TrafficID:   "req-789",
		RuleID:      "rule-sql-001",
		Name:        "SQL Injection",
		Description: "Potential SQL injection vulnerability",
		Severity:    SeverityHigh,
		Confidence:  0.9,
		URL:         "https://example.com/api/users",
		Parameter:   "id",
		Evidence:    "UNION SELECT detected",
		Status:      StatusOpen,
		Timestamp:   time.Now(),
	}

	if vuln.ID != "vuln-123" {
		t.Errorf("Expected ID 'vuln-123', got '%s'", vuln.ID)
	}

	if vuln.Severity != SeverityHigh {
		t.Errorf("Expected Severity '%s', got '%s'", SeverityHigh, vuln.Severity)
	}
}

// TestScannerStats tests scanner statistics
func TestScannerStats(t *testing.T) {
	stats := &ScannerStats{
		TotalScanned:    100,
		TotalVulnsFound: 15,
		BySeverity: map[Severity]int64{
			SeverityCritical: 2,
			SeverityHigh:     5,
			SeverityMedium:   5,
			SeverityLow:      3,
		},
		AverageScanTime: 50 * time.Millisecond,
		QueueSize:       100,
		ActiveScans:     5,
	}

	if stats.TotalScanned != 100 {
		t.Errorf("Expected TotalScanned 100, got %d", stats.TotalScanned)
	}

	if stats.BySeverity[SeverityCritical] != 2 {
		t.Errorf("Expected 2 critical vulns, got %d", stats.BySeverity[SeverityCritical])
	}
}

// TestMockRule_Match tests the mock rule matching
func TestMockRule_Match(t *testing.T) {
	rule := &MockRule{
		id:          "test-sqli",
		name:        "SQL Injection Detector",
		description: "Detects SQL injection patterns",
		severity:    SeverityHigh,
		enabled:     true,
		priority:    1,
		matchFunc: func(traffic *HTTPTraffic) (bool, *MatchResult) {
			// Simple SQL injection pattern detection
			body := string(traffic.Body)
			if len(body) > 0 && (contains(body, "' OR '1'='1") || contains(body, "UNION SELECT")) {
				return true, &MatchResult{
					Matched:    true,
					Evidence:   "SQL injection pattern found in request body",
					Confidence: 0.9,
				}
			}
			return false, &MatchResult{Matched: false}
		},
	}

	// Test matching traffic
	matchingTraffic := &HTTPTraffic{
		ID:      "req-1",
		Body:    []byte(`{"query": "SELECT * FROM users WHERE id='1' OR '1'='1'"}`),
		Method:  "POST",
		URL:     "https://example.com/api/search",
	}

	matched, result := rule.Match(matchingTraffic)
	if !matched {
		t.Error("Expected rule to match SQL injection traffic")
	}
	if result.Confidence != 0.9 {
		t.Errorf("Expected confidence 0.9, got %f", result.Confidence)
	}

	// Test non-matching traffic
	nonMatchingTraffic := &HTTPTraffic{
		ID:      "req-2",
		Body:    []byte(`{"query": "SELECT * FROM users WHERE id=1"}`),
		Method:  "POST",
		URL:     "https://example.com/api/search",
	}

	matched, _ = rule.Match(nonMatchingTraffic)
	if matched {
		t.Error("Expected rule NOT to match safe traffic")
	}
}

// TestMockRule_EnableDisable tests enabling and disabling rules
func TestMockRule_EnableDisable(t *testing.T) {
	rule := &MockRule{
		id:      "test-rule",
		enabled: true,
	}

	if !rule.Enabled() {
		t.Error("Expected rule to be enabled initially")
	}

	rule.SetEnabled(false)
	if rule.Enabled() {
		t.Error("Expected rule to be disabled after SetEnabled(false)")
	}

	rule.SetEnabled(true)
	if !rule.Enabled() {
		t.Error("Expected rule to be enabled after SetEnabled(true)")
	}
}

// Helper function
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
