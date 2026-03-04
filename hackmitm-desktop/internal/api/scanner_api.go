package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"

	"hackmitm-desktop/internal/scanner"
)

// ScannerAPI handles scanner-related API calls
type ScannerAPI struct {
	apiEndpoint string
	client      *http.Client
	ctx         context.Context

	// Embedded scanner for direct access
	embeddedScanner *scanner.Scanner
	findings        []*scanner.Vulnerability
	findingsMu      sync.RWMutex
}

// NewScannerAPI creates a new ScannerAPI instance
func NewScannerAPI() *ScannerAPI {
	s := &ScannerAPI{
		client:   &http.Client{},
		findings: make([]*scanner.Vulnerability, 0),
	}

	// Initialize embedded scanner
	s.embeddedScanner = scanner.NewScanner()
	s.embeddedScanner.OnVulnerability = func(v *scanner.Vulnerability) {
		s.findingsMu.Lock()
		s.findings = append(s.findings, v)
		s.findingsMu.Unlock()
	}

	return s
}

// SetContext sets the application context
func (a *ScannerAPI) SetContext(ctx context.Context) {
	a.ctx = ctx
}

// SetAPIEndpoint sets the API endpoint
func (a *ScannerAPI) SetAPIEndpoint(endpoint string) {
	a.apiEndpoint = endpoint
}

// Rule represents a scanner rule
type Rule struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Severity    string   `json:"severity"`
	Enabled     bool     `json:"enabled"`
	Priority    int      `json:"priority"`
	Tags        []string `json:"tags"`
	Remediation string   `json:"remediation,omitempty"`
}

// GetRules fetches all scanner rules
func (a *ScannerAPI) GetRules() ([]Rule, error) {
	resp, err := a.client.Get(fmt.Sprintf("%s/api/scanner/rules", a.apiEndpoint))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var result struct {
		Data  []Rule `json:"data"`
		Total int    `json:"total"`
	}

	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}

	return result.Data, nil
}

// GetRule fetches a single rule by ID
func (a *ScannerAPI) GetRule(id string) (*Rule, error) {
	resp, err := a.client.Get(fmt.Sprintf("%s/api/scanner/rules/%s", a.apiEndpoint, id))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("rule not found")
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var rule Rule
	if err := json.Unmarshal(body, &rule); err != nil {
		return nil, err
	}

	return &rule, nil
}

// EnableRule enables a rule
func (a *ScannerAPI) EnableRule(id string) error {
	return a.updateRuleStatus(id, true)
}

// DisableRule disables a rule
func (a *ScannerAPI) DisableRule(id string) error {
	return a.updateRuleStatus(id, false)
}

func (a *ScannerAPI) updateRuleStatus(id string, enabled bool) error {
	data := map[string]interface{}{"enabled": enabled}
	jsonData, err := json.Marshal(data)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("PATCH",
		fmt.Sprintf("%s/api/scanner/rules/%s", a.apiEndpoint, id),
		bytes.NewReader(jsonData))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := a.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to update rule status")
	}

	return nil
}

// CreateRule creates a new custom rule
func (a *ScannerAPI) CreateRule(rule Rule) error {
	jsonData, err := json.Marshal(rule)
	if err != nil {
		return err
	}

	resp, err := a.client.Post(
		fmt.Sprintf("%s/api/scanner/rules", a.apiEndpoint),
		"application/json",
		bytes.NewReader(jsonData))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to create rule: %s", string(body))
	}

	return nil
}

// ReloadRules reloads all rules from disk
func (a *ScannerAPI) ReloadRules() (int, error) {
	resp, err := a.client.Post(
		fmt.Sprintf("%s/api/scanner/reload", a.apiEndpoint),
		"application/json",
		nil)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, err
	}

	var result struct {
		Count int `json:"count"`
	}

	if err := json.Unmarshal(body, &result); err != nil {
		return 0, err
	}

	return result.Count, nil
}

// ============ Embedded Scanner Methods (Direct Access) ============

// ScanRequest scans an HTTP request using embedded scanner
func (a *ScannerAPI) ScanRequest(msg *scanner.HTTPMessage) []*scanner.Vulnerability {
	if a.embeddedScanner == nil {
		return nil
	}
	return a.embeddedScanner.ScanRequest(msg)
}

// ScanResponse scans an HTTP response using embedded scanner
func (a *ScannerAPI) ScanResponse(msg *scanner.HTTPMessage) []*scanner.Vulnerability {
	if a.embeddedScanner == nil {
		return nil
	}
	return a.embeddedScanner.ScanResponse(msg)
}

// GetEmbeddedFindings returns all findings from embedded scanner
func (a *ScannerAPI) GetEmbeddedFindings() []*scanner.Vulnerability {
	a.findingsMu.RLock()
	defer a.findingsMu.RUnlock()
	return a.findings
}

// ClearEmbeddedFindings clears all findings from embedded scanner
func (a *ScannerAPI) ClearEmbeddedFindings() {
	a.findingsMu.Lock()
	defer a.findingsMu.Unlock()
	a.findings = make([]*scanner.Vulnerability, 0)
	if a.embeddedScanner != nil {
		a.embeddedScanner.ClearFindings()
	}
}

// GetEmbeddedRules returns all rules from embedded scanner
func (a *ScannerAPI) GetEmbeddedRules() []*scanner.ScanRule {
	if a.embeddedScanner == nil {
		return nil
	}
	return a.embeddedScanner.GetRules()
}

// EnableEmbeddedRule enables a rule in embedded scanner
func (a *ScannerAPI) EnableEmbeddedRule(ruleID string) {
	if a.embeddedScanner != nil {
		a.embeddedScanner.EnableRule(ruleID)
	}
}

// DisableEmbeddedRule disables a rule in embedded scanner
func (a *ScannerAPI) DisableEmbeddedRule(ruleID string) {
	if a.embeddedScanner != nil {
		a.embeddedScanner.DisableRule(ruleID)
	}
}

// SetScannerEnabled enables or disables the embedded scanner
func (a *ScannerAPI) SetScannerEnabled(enabled bool) {
	if a.embeddedScanner != nil {
		a.embeddedScanner.SetEnabled(enabled)
	}
}

// IsScannerEnabled returns whether the embedded scanner is enabled
func (a *ScannerAPI) IsScannerEnabled() bool {
	if a.embeddedScanner == nil {
		return false
	}
	return a.embeddedScanner.IsEnabled()
}

// ScanTrafficMessage is a convenience method to scan both request and response
func (a *ScannerAPI) ScanTrafficMessage(url, method string, headers map[string]string, reqBody, respBody string, statusCode int) []*scanner.Vulnerability {
	var allFindings []*scanner.Vulnerability

	// Scan request
	reqMsg := &scanner.HTTPMessage{
		URL:       url,
		Method:    method,
		Headers:   headers,
		Body:      reqBody,
		IsRequest: true,
	}
	allFindings = append(allFindings, a.ScanRequest(reqMsg)...)

	// Scan response
	respMsg := &scanner.HTTPMessage{
		URL:        url,
		Method:     method,
		Headers:    headers,
		Body:       respBody,
		StatusCode: statusCode,
		IsRequest:  false,
	}
	allFindings = append(allFindings, a.ScanResponse(respMsg)...)

	return allFindings
}
