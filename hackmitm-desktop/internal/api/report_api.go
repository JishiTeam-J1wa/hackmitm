package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// ReportAPI handles report-related API calls
type ReportAPI struct {
	apiEndpoint string
	client      *http.Client
}

// NewReportAPI creates a new ReportAPI instance
func NewReportAPI() *ReportAPI {
	return &ReportAPI{
		client: &http.Client{},
	}
}

// SetAPIEndpoint sets the API endpoint
func (a *ReportAPI) SetAPIEndpoint(endpoint string) {
	a.apiEndpoint = endpoint
}

// ReportOptions contains options for report generation
type ReportOptions struct {
	SessionID string   `json:"session_id"`
	Title     string   `json:"title"`
	Format    string   `json:"format"` // json, markdown, html
	Severity  []string `json:"severity"`
	Status    []string `json:"status"`
}

// ReportData represents the generated report
type ReportData struct {
	Title       string                 `json:"title"`
	GeneratedAt string                 `json:"generated_at"`
	SessionID   string                 `json:"session_id"`
	Summary     ReportSummary          `json:"summary"`
	Vulns       []VulnerabilityReport  `json:"vulns"`
	Metadata    map[string]interface{} `json:"metadata"`
}

// ReportSummary contains report summary statistics
type ReportSummary struct {
	Total        int            `json:"total"`
	BySeverity   map[string]int `json:"by_severity"`
	ByStatus     map[string]int `json:"by_status"`
	TopVulnTypes []VulnTypeCount `json:"top_vuln_types"`
	TopHosts     []HostCount    `json:"top_hosts"`
}

// VulnerabilityReport represents a vulnerability in a report
type VulnerabilityReport struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Severity    string `json:"severity"`
	Confidence  float64 `json:"confidence"`
	URL         string `json:"url"`
	Parameter   string `json:"parameter"`
	Status      string `json:"status"`
	Evidence    string `json:"evidence"`
	Remediation string `json:"remediation"`
	FirstSeen   string `json:"first_seen"`
	LastSeen    string `json:"last_seen"`
	Occurrences int    `json:"occurrences"`
}

// VulnTypeCount represents a vulnerability type count
type VulnTypeCount struct {
	Type  string `json:"type"`
	Count int    `json:"count"`
}

// HostCount represents a host count
type HostCount struct {
	Host  string `json:"host"`
	Count int    `json:"count"`
}

// GenerateReport generates a report with the given options
func (a *ReportAPI) GenerateReport(options ReportOptions) ([]byte, string, error) {
	jsonData, err := json.Marshal(options)
	if err != nil {
		return nil, "", err
	}

	resp, err := a.client.Post(
		fmt.Sprintf("%s/api/reports/generate", a.apiEndpoint),
		"application/json",
		bytes.NewReader(jsonData))
	if err != nil {
		return nil, "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, "", fmt.Errorf("failed to generate report: %s", string(body))
	}

	// Get content type
	contentType := resp.Header.Get("Content-Type")

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, "", err
	}

	return body, contentType, nil
}

// GenerateJSONReport generates a JSON report
func (a *ReportAPI) GenerateJSONReport(sessionID, title string, severity, status []string) (*ReportData, error) {
	options := ReportOptions{
		SessionID: sessionID,
		Title:     title,
		Format:    "json",
		Severity:  severity,
		Status:    status,
	}

	data, _, err := a.GenerateReport(options)
	if err != nil {
		return nil, err
	}

	var report ReportData
	if err := json.Unmarshal(data, &report); err != nil {
		return nil, err
	}

	return &report, nil
}

// GenerateHTMLReport generates an HTML report
func (a *ReportAPI) GenerateHTMLReport(sessionID, title string, severity, status []string) ([]byte, error) {
	options := ReportOptions{
		SessionID: sessionID,
		Title:     title,
		Format:    "html",
		Severity:  severity,
		Status:    status,
	}

	data, _, err := a.GenerateReport(options)
	return data, err
}

// GenerateMarkdownReport generates a Markdown report
func (a *ReportAPI) GenerateMarkdownReport(sessionID, title string, severity, status []string) ([]byte, error) {
	options := ReportOptions{
		SessionID: sessionID,
		Title:     title,
		Format:    "markdown",
		Severity:  severity,
		Status:    status,
	}

	data, _, err := a.GenerateReport(options)
	return data, err
}

// ListReports lists available reports
func (a *ReportAPI) ListReports() ([]map[string]interface{}, error) {
	resp, err := a.client.Get(fmt.Sprintf("%s/api/reports", a.apiEndpoint))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var result struct {
		Data []map[string]interface{} `json:"data"`
	}

	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}

	return result.Data, nil
}
