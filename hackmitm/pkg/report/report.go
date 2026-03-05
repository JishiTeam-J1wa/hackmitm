// Package report 提供报告生成功能
// Package report provides report generation functionality
package report

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"text/template"
	"time"

	"hackmitm/pkg/storage"
)

// Generator 报告生成器
type Generator struct {
	storage *storage.SQLiteStorage
}

// NewGenerator 创建报告生成器
func NewGenerator(storage *storage.SQLiteStorage) *Generator {
	return &Generator{storage: storage}
}

// ReportData 报告数据
type ReportData struct {
	Title       string                 `json:"title"`
	GeneratedAt time.Time              `json:"generated_at"`
	SessionID   string                 `json:"session_id"`
	Summary     *ReportSummary         `json:"summary"`
	Vulns       []VulnerabilityReport  `json:"vulns"`
	Metadata    map[string]interface{} `json:"metadata"`
}

// ReportSummary 报告摘要
type ReportSummary struct {
	Total          int            `json:"total"`
	BySeverity     map[string]int `json:"by_severity"`
	ByStatus       map[string]int `json:"by_status"`
	TopVulnTypes   []VulnTypeCount `json:"top_vuln_types"`
	TopHosts       []HostCount    `json:"top_hosts"`
}

// VulnerabilityReport 漏洞报告
type VulnerabilityReport struct {
	ID           string    `json:"id"`
	Name         string    `json:"name"`
	Severity     string    `json:"severity"`
	Confidence   float64   `json:"confidence"`
	URL          string    `json:"url"`
	Parameter    string    `json:"parameter"`
	Status       string    `json:"status"`
	Evidence     string    `json:"evidence"`
	Remediation  string    `json:"remediation"`
	FirstSeen    time.Time `json:"first_seen"`
	LastSeen     time.Time `json:"last_seen"`
	Occurrences  int       `json:"occurrences"`
}

// VulnTypeCount 漏洞类型计数
type VulnTypeCount struct {
	Type  string `json:"type"`
	Count int    `json:"count"`
}

// HostCount 主机计数
type HostCount struct {
	Host  string `json:"host"`
	Count int    `json:"count"`
}

// GenerateOptions 生成选项
type GenerateOptions struct {
	SessionID string
	Title     string
	Severity  []string
	Status    []string
}

// Generate 生成报告数据
func (g *Generator) Generate(opts GenerateOptions) (*ReportData, error) {
	// 获取漏洞列表
	records, total, err := g.storage.ListVulnerabilities(opts.SessionID, "", "", 1000, 0)
	if err != nil {
		return nil, err
	}

	// 获取统计
	stats, err := g.storage.GetVulnerabilityStats(opts.SessionID)
	if err != nil {
		return nil, err
	}

	// 过滤漏洞
	filtered := make([]storage.VulnerabilityRecord, 0)
	for _, r := range records {
		if len(opts.Severity) > 0 {
			found := false
			for _, s := range opts.Severity {
				if r.Severity == s {
					found = true
					break
				}
			}
			if !found {
				continue
			}
		}
		if len(opts.Status) > 0 {
			found := false
			for _, s := range opts.Status {
				if r.Status == s {
					found = true
					break
				}
			}
			if !found {
				continue
			}
		}
		filtered = append(filtered, r)
	}

	// 构建报告数据
	report := &ReportData{
		Title:       opts.Title,
		GeneratedAt: time.Now(),
		SessionID:   opts.SessionID,
		Summary: &ReportSummary{
			Total:        total,
			BySeverity:   stats,
			ByStatus:     make(map[string]int),
			TopVulnTypes: make([]VulnTypeCount, 0),
			TopHosts:     make([]HostCount, 0),
		},
		Vulns:    make([]VulnerabilityReport, len(filtered)),
		Metadata: make(map[string]interface{}),
	}

	// 转换漏洞数据
	for i, r := range filtered {
		report.Vulns[i] = VulnerabilityReport{
			ID:          r.ID,
			Name:        r.Name,
			Severity:    r.Severity,
			Confidence:  r.Confidence,
			URL:         r.URL,
			Parameter:   r.Parameter,
			Status:      r.Status,
			Evidence:    r.Evidence,
			Remediation: r.Remediation,
			FirstSeen:   r.FirstSeen,
			LastSeen:    r.LastSeen,
			Occurrences: r.Occurrences,
		}
	}

	return report, nil
}

// ToJSON 生成 JSON 报告
func (g *Generator) ToJSON(opts GenerateOptions) ([]byte, error) {
	report, err := g.Generate(opts)
	if err != nil {
		return nil, err
	}
	return json.MarshalIndent(report, "", "  ")
}

// ToMarkdown 生成 Markdown 报告
func (g *Generator) ToMarkdown(opts GenerateOptions) ([]byte, error) {
	report, err := g.Generate(opts)
	if err != nil {
		return nil, err
	}

	tmpl := `# {{.Title}}

**Generated:** {{.GeneratedAt.Format "2006-01-02 15:04:05"}}
**Session:** {{.SessionID}}

## Summary

| Severity | Count |
|----------|-------|
{{range $k, $v := .Summary.BySeverity}}| {{$k}} | {{$v}} |
{{end}}

**Total Vulnerabilities:** {{.Summary.Total}}

## Vulnerabilities

{{range .Vulns}}
### {{.Name}}

| Field | Value |
|-------|-------|
| ID | {{.ID}} |
| Severity | **{{.Severity}}** |
| Confidence | {{printf "%.0f" .Confidence}}% |
| Status | {{.Status}} |
| URL | {{.URL}} |
| Parameter | {{.Parameter}} |
| Occurrences | {{.Occurrences}} |
| First Seen | {{.FirstSeen.Format "2006-01-02 15:04"}} |
| Last Seen | {{.LastSeen.Format "2006-01-02 15:04"}} |

**Evidence:**
    {{.Evidence}}

**Remediation:**
{{.Remediation}}

---
{{end}}
`

	t, err := template.New("report").Parse(tmpl)
	if err != nil {
		return nil, err
	}

	var buf bytes.Buffer
	if err := t.Execute(&buf, report); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

// ToHTML 生成 HTML 报告
func (g *Generator) ToHTML(opts GenerateOptions) ([]byte, error) {
	report, err := g.Generate(opts)
	if err != nil {
		return nil, err
	}

	tmpl := `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>{{.Title}}</title>
    <style>
        body { font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif; margin: 0; padding: 20px; background: #f5f5f5; }
        .container { max-width: 1200px; margin: 0 auto; background: white; padding: 30px; border-radius: 8px; box-shadow: 0 2px 4px rgba(0,0,0,0.1); }
        h1 { color: #333; border-bottom: 2px solid #4CAF50; padding-bottom: 10px; }
        h2 { color: #555; margin-top: 30px; }
        .meta { color: #666; font-size: 14px; margin-bottom: 20px; }
        .summary-grid { display: grid; grid-template-columns: repeat(auto-fit, minmax(150px, 1fr)); gap: 15px; margin: 20px 0; }
        .summary-card { background: #f9f9f9; padding: 15px; border-radius: 6px; text-align: center; }
        .summary-card .count { font-size: 24px; font-weight: bold; color: #333; }
        .summary-card .label { font-size: 12px; color: #666; text-transform: uppercase; }
        .severity-critical { color: #d32f2f; }
        .severity-high { color: #f57c00; }
        .severity-medium { color: #fbc02d; }
        .severity-low { color: #388e3c; }
        .severity-info { color: #1976d2; }
        .vuln-card { border: 1px solid #e0e0e0; border-radius: 6px; margin: 15px 0; padding: 20px; }
        .vuln-header { display: flex; justify-content: space-between; align-items: center; }
        .vuln-title { font-size: 18px; font-weight: 600; margin: 0; }
        .vuln-severity { padding: 4px 12px; border-radius: 4px; font-size: 12px; font-weight: 600; text-transform: uppercase; }
        .vuln-details { display: grid; grid-template-columns: repeat(2, 1fr); gap: 10px; margin: 15px 0; }
        .vuln-detail-item { font-size: 14px; }
        .vuln-detail-item strong { color: #666; }
        .evidence { background: #f5f5f5; padding: 15px; border-radius: 4px; font-family: monospace; font-size: 13px; overflow-x: auto; }
        .remediation { background: #e8f5e9; padding: 15px; border-radius: 4px; margin-top: 10px; }
    </style>
</head>
<body>
    <div class="container">
        <h1>{{.Title}}</h1>
        <div class="meta">
            <strong>Generated:</strong> {{.GeneratedAt.Format "2006-01-02 15:04:05"}} |
            <strong>Session:</strong> {{.SessionID}}
        </div>

        <h2>Summary</h2>
        <div class="summary-grid">
            <div class="summary-card">
                <div class="count">{{.Summary.Total}}</div>
                <div class="label">Total</div>
            </div>
            {{range $k, $v := .Summary.BySeverity}}
            {{if eq $k "critical"}}
            <div class="summary-card">
                <div class="count severity-critical">{{$v}}</div>
                <div class="label">Critical</div>
            </div>
            {{end}}
            {{if eq $k "high"}}
            <div class="summary-card">
                <div class="count severity-high">{{$v}}</div>
                <div class="label">High</div>
            </div>
            {{end}}
            {{if eq $k "medium"}}
            <div class="summary-card">
                <div class="count severity-medium">{{$v}}</div>
                <div class="label">Medium</div>
            </div>
            {{end}}
            {{if eq $k "low"}}
            <div class="summary-card">
                <div class="count severity-low">{{$v}}</div>
                <div class="label">Low</div>
            </div>
            {{end}}
            {{end}}
        </div>

        <h2>Vulnerabilities</h2>
        {{range .Vulns}}
        <div class="vuln-card">
            <div class="vuln-header">
                <h3 class="vuln-title">{{.Name}}</h3>
                <span class="vuln-severity severity-{{.Severity}}">{{.Severity}}</span>
            </div>
            <div class="vuln-details">
                <div class="vuln-detail-item"><strong>ID:</strong> {{.ID}}</div>
                <div class="vuln-detail-item"><strong>Status:</strong> {{.Status}}</div>
                <div class="vuln-detail-item"><strong>URL:</strong> {{.URL}}</div>
                <div class="vuln-detail-item"><strong>Parameter:</strong> {{.Parameter}}</div>
                <div class="vuln-detail-item"><strong>Confidence:</strong> {{printf "%.0f" .Confidence}}%</div>
                <div class="vuln-detail-item"><strong>Occurrences:</strong> {{.Occurrences}}</div>
            </div>
            {{if .Evidence}}
            <div class="evidence"><strong>Evidence:</strong><br>{{.Evidence}}</div>
            {{end}}
            {{if .Remediation}}
            <div class="remediation"><strong>Remediation:</strong> {{.Remediation}}</div>
            {{end}}
        </div>
        {{end}}
    </div>
</body>
</html>
`

	t, err := template.New("report").Parse(tmpl)
	if err != nil {
		return nil, err
	}

	var buf bytes.Buffer
	if err := t.Execute(&buf, report); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

// ReportAPI 报告 API
type ReportAPI struct {
	generator *Generator
	storage   *storage.SQLiteStorage
}

// NewReportAPI 创建报告 API
func NewReportAPI(storage *storage.SQLiteStorage) *ReportAPI {
	return &ReportAPI{
		generator: NewGenerator(storage),
		storage:   storage,
	}
}

// RegisterRoutes 注册路由
func (api *ReportAPI) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /api/reports/generate", api.GenerateReport)
	mux.HandleFunc("GET /api/reports", api.ListReports)
}

// GenerateReport 生成报告
func (api *ReportAPI) GenerateReport(w http.ResponseWriter, r *http.Request) {
	var req struct {
		SessionID string   `json:"session_id"`
		Title     string   `json:"title"`
		Format    string   `json:"format"` // json, markdown, html
		Severity  []string `json:"severity"`
		Status    []string `json:"status"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.Title == "" {
		req.Title = "Security Assessment Report"
	}
	if req.Format == "" {
		req.Format = "json"
	}

	opts := GenerateOptions{
		SessionID: req.SessionID,
		Title:     req.Title,
		Severity:  req.Severity,
		Status:    req.Status,
	}

	var data []byte
	var err error
	var contentType string

	switch req.Format {
	case "json":
		data, err = api.generator.ToJSON(opts)
		contentType = "application/json"
	case "markdown", "md":
		data, err = api.generator.ToMarkdown(opts)
		contentType = "text/markdown"
	case "html":
		data, err = api.generator.ToHTML(opts)
		contentType = "text/html"
	default:
		writeError(w, http.StatusBadRequest, "Invalid format. Supported: json, markdown, html")
		return
	}

	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to generate report: "+err.Error())
		return
	}

	w.Header().Set("Content-Type", contentType)
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=security-report.%s", req.Format))
	w.Write(data)
}

// ListReports 列出报告（占位）
func (api *ReportAPI) ListReports(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"data": []interface{}{},
		"message": "Reports are generated on-demand. Use POST /api/reports/generate to create a report.",
	})
}

// Helper functions

func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]interface{}{
		"error":  message,
		"status": status,
	})
}
