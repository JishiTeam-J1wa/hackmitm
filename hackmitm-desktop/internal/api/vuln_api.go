package api

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"
)

// VulnAPI handles vulnerability-related operations
type VulnAPI struct {
	ctx    context.Context
	db     *sql.DB
	dbPath string
}

// NewVulnAPI creates a new VulnAPI instance
func NewVulnAPI() *VulnAPI {
	return &VulnAPI{}
}

// SetContext sets the context for the API
func (a *VulnAPI) SetContext(ctx context.Context) {
	a.ctx = ctx
}

// SetDB sets the database connection
func (a *VulnAPI) SetDB(db *sql.DB) {
	a.db = db
}

// Vulnerability represents a stored vulnerability
type Vulnerability struct {
	ID           int64    `json:"id"`
	Title        string   `json:"title"`
	Severity     string   `json:"severity"`
	Type         string   `json:"type"`
	URL          string   `json:"url"`
	Method       string   `json:"method"`
	Request      string   `json:"request"`
	Response     string   `json:"response"`
	Description  string   `json:"description"`
	Remediation  string   `json:"remediation"`
	References   []string `json:"references"`
	Status       string   `json:"status"`
	CreatedAt    string   `json:"createdAt"`
	UpdatedAt    string   `json:"updatedAt"`
	Source       string   `json:"source"`
	CWE          string   `json:"cwe"`
	CVSS         float64  `json:"cvss"`
}

// ScanResult represents a scan result
type ScanResult struct {
	ID            int64    `json:"id"`
	PluginName    string   `json:"pluginName"`
	PluginID      string   `json:"pluginId"`
	Severity      string   `json:"severity"`
	Title         string   `json:"title"`
	Description   string   `json:"description"`
	URL           string   `json:"url"`
	Method        string   `json:"method"`
	Evidence      string   `json:"evidence"`
	Request       string   `json:"request"`
	Response      string   `json:"response"`
	Timestamp     string   `json:"timestamp"`
	FalsePositive bool     `json:"falsePositive"`
	Tags          []string `json:"tags"`
}

// WebSocketMessage represents a WebSocket message
type WebSocketMessage struct {
	ID           int64  `json:"id"`
	Timestamp    string `json:"timestamp"`
	Direction    string `json:"direction"`
	Type         string `json:"type"`
	URL          string `json:"url"`
	Host         string `json:"host"`
	Size         int64  `json:"size"`
	Content      string `json:"content"`
	ContentType  string `json:"contentType"`
	ConnectionID string `json:"connectionId"`
}

// ============ Vulnerability Operations ============

// GetVulnerabilities retrieves vulnerabilities with filters
func (a *VulnAPI) GetVulnerabilities(severity, status, vulnType string, limit int) ([]Vulnerability, error) {
	if a.db == nil {
		return nil, fmt.Errorf("database not initialized")
	}

	query := "SELECT id, title, severity, type, url, method, request, response, description, remediation, references, status, created_at, updated_at, source, cwe, cvss FROM vulnerabilities WHERE 1=1"
	args := []interface{}{}

	if severity != "" && severity != "all" {
		query += " AND severity = ?"
		args = append(args, severity)
	}
	if status != "" && status != "all" {
		query += " AND status = ?"
		args = append(args, status)
	}
	if vulnType != "" && vulnType != "all" {
		query += " AND type = ?"
		args = append(args, vulnType)
	}

	query += " ORDER BY created_at DESC LIMIT ?"
	args = append(args, limit)

	rows, err := a.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var vulns []Vulnerability
	for rows.Next() {
		var v Vulnerability
		var refsStr sql.NullString
		var createdAt, updatedAt sql.NullString

		err := rows.Scan(
			&v.ID, &v.Title, &v.Severity, &v.Type, &v.URL, &v.Method,
			&v.Request, &v.Response, &v.Description, &v.Remediation,
			&refsStr, &v.Status, &createdAt, &updatedAt, &v.Source, &v.CWE, &v.CVSS,
		)
		if err != nil {
			continue
		}

		if refsStr.Valid && refsStr.String != "" {
			json.Unmarshal([]byte(refsStr.String), &v.References)
		}
		v.CreatedAt = createdAt.String
		v.UpdatedAt = updatedAt.String

		vulns = append(vulns, v)
	}

	return vulns, nil
}

// AddVulnerability adds a new vulnerability
func (a *VulnAPI) AddVulnerability(v Vulnerability) (int64, error) {
	if a.db == nil {
		return 0, fmt.Errorf("database not initialized")
	}

	refsJSON, _ := json.Marshal(v.References)

	result, err := a.db.Exec(
		`INSERT INTO vulnerabilities (title, severity, type, url, method, request, response, description, remediation, references, status, source, cwe, cvss)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		v.Title, v.Severity, v.Type, v.URL, v.Method, v.Request, v.Response,
		v.Description, v.Remediation, string(refsJSON), v.Status, v.Source, v.CWE, v.CVSS,
	)
	if err != nil {
		return 0, err
	}

	return result.LastInsertId()
}

// UpdateVulnerabilityStatus updates the status of a vulnerability
func (a *VulnAPI) UpdateVulnerabilityStatus(id int64, status string) error {
	if a.db == nil {
		return fmt.Errorf("database not initialized")
	}

	_, err := a.db.Exec(
		"UPDATE vulnerabilities SET status = ?, updated_at = ? WHERE id = ?",
		status, time.Now().Format(time.RFC3339), id,
	)
	return err
}

// DeleteVulnerability deletes a vulnerability
func (a *VulnAPI) DeleteVulnerability(id int64) error {
	if a.db == nil {
		return fmt.Errorf("database not initialized")
	}

	_, err := a.db.Exec("DELETE FROM vulnerabilities WHERE id = ?", id)
	return err
}

// GetVulnStats returns vulnerability statistics
func (a *VulnAPI) GetVulnStats() (map[string]int64, error) {
	if a.db == nil {
		return nil, fmt.Errorf("database not initialized")
	}

	stats := map[string]int64{}

	// Total count
	var total int64
	a.db.QueryRow("SELECT COUNT(*) FROM vulnerabilities").Scan(&total)
	stats["total"] = total

	// Count by severity
	severities := []string{"critical", "high", "medium", "low"}
	for _, s := range severities {
		var count int64
		a.db.QueryRow("SELECT COUNT(*) FROM vulnerabilities WHERE severity = ?", s).Scan(&count)
		stats[s] = count
	}

	// Count by status
	statuses := []string{"open", "fixed", "ignored"}
	for _, s := range statuses {
		var count int64
		a.db.QueryRow("SELECT COUNT(*) FROM vulnerabilities WHERE status = ?", s).Scan(&count)
		stats[s] = count
	}

	return stats, nil
}

// ============ Scan Result Operations ============

// GetScanResults retrieves scan results with filters
func (a *VulnAPI) GetScanResults(severity, pluginID string, falsePositive string, limit int) ([]ScanResult, error) {
	if a.db == nil {
		return nil, fmt.Errorf("database not initialized")
	}

	query := "SELECT id, plugin_name, plugin_id, severity, title, description, url, method, evidence, request, response, timestamp, false_positive, tags FROM scan_results WHERE 1=1"
	args := []interface{}{}

	if severity != "" && severity != "all" {
		query += " AND severity = ?"
		args = append(args, severity)
	}
	if pluginID != "" && pluginID != "all" {
		query += " AND plugin_id = ?"
		args = append(args, pluginID)
	}
	if falsePositive == "true" {
		query += " AND false_positive = 1"
	} else if falsePositive == "false" {
		query += " AND false_positive = 0"
	}

	query += " ORDER BY timestamp DESC LIMIT ?"
	args = append(args, limit)

	rows, err := a.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []ScanResult
	for rows.Next() {
		var r ScanResult
		var tagsStr sql.NullString
		var timestamp sql.NullString
		var fp int

		err := rows.Scan(
			&r.ID, &r.PluginName, &r.PluginID, &r.Severity, &r.Title, &r.Description,
			&r.URL, &r.Method, &r.Evidence, &r.Request, &r.Response, &timestamp, &fp, &tagsStr,
		)
		if err != nil {
			continue
		}

		r.Timestamp = timestamp.String
		r.FalsePositive = fp == 1

		if tagsStr.Valid && tagsStr.String != "" {
			json.Unmarshal([]byte(tagsStr.String), &r.Tags)
		}

		results = append(results, r)
	}

	return results, nil
}

// AddScanResult adds a new scan result
func (a *VulnAPI) AddScanResult(r ScanResult) (int64, error) {
	if a.db == nil {
		return 0, fmt.Errorf("database not initialized")
	}

	tagsJSON, _ := json.Marshal(r.Tags)
	fp := 0
	if r.FalsePositive {
		fp = 1
	}

	result, err := a.db.Exec(
		`INSERT INTO scan_results (plugin_name, plugin_id, severity, title, description, url, method, evidence, request, response, false_positive, tags)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		r.PluginName, r.PluginID, r.Severity, r.Title, r.Description,
		r.URL, r.Method, r.Evidence, r.Request, r.Response, fp, string(tagsJSON),
	)
	if err != nil {
		return 0, err
	}

	return result.LastInsertId()
}

// MarkScanResultFalsePositive marks a scan result as false positive or not
func (a *VulnAPI) MarkScanResultFalsePositive(id int64, isFP bool) error {
	if a.db == nil {
		return fmt.Errorf("database not initialized")
	}

	fp := 0
	if isFP {
		fp = 1
	}

	_, err := a.db.Exec("UPDATE scan_results SET false_positive = ? WHERE id = ?", fp, id)
	return err
}

// DeleteScanResult deletes a scan result
func (a *VulnAPI) DeleteScanResult(id int64) error {
	if a.db == nil {
		return fmt.Errorf("database not initialized")
	}

	_, err := a.db.Exec("DELETE FROM scan_results WHERE id = ?", id)
	return err
}

// ClearScanResults clears all scan results
func (a *VulnAPI) ClearScanResults() error {
	if a.db == nil {
		return fmt.Errorf("database not initialized")
	}

	_, err := a.db.Exec("DELETE FROM scan_results")
	return err
}

// ExportScanResultToVuln exports a scan result to vulnerabilities table
func (a *VulnAPI) ExportScanResultToVuln(scanID int64) (int64, error) {
	if a.db == nil {
		return 0, fmt.Errorf("database not initialized")
	}

	// Get the scan result
	var r ScanResult
	var tagsStr sql.NullString
	err := a.db.QueryRow(
		"SELECT plugin_name, plugin_id, severity, title, description, url, method, evidence, request, response, tags FROM scan_results WHERE id = ?",
		scanID,
	).Scan(&r.PluginName, &r.PluginID, &r.Severity, &r.Title, &r.Description, &r.URL, &r.Method, &r.Evidence, &r.Request, &r.Response, &tagsStr)

	if err != nil {
		return 0, err
	}

	// Create vulnerability from scan result
	v := Vulnerability{
		Title:       r.Title,
		Severity:    r.Severity,
		Type:        r.PluginName,
		URL:         r.URL,
		Method:      r.Method,
		Request:     r.Request,
		Response:    r.Response,
		Description: r.Description,
		Remediation: "",
		Status:      "open",
		Source:      "passive",
	}

	if tagsStr.Valid && tagsStr.String != "" {
		json.Unmarshal([]byte(tagsStr.String), &v.References)
	}

	return a.AddVulnerability(v)
}

// ============ WebSocket Message Operations ============

// GetWebSocketMessages retrieves WebSocket messages
func (a *VulnAPI) GetWebSocketMessages(direction, msgType, connectionID string, limit int) ([]WebSocketMessage, error) {
	if a.db == nil {
		return nil, fmt.Errorf("database not initialized")
	}

	query := "SELECT id, timestamp, direction, type, url, host, size, content, content_type, connection_id FROM websocket_messages WHERE 1=1"
	args := []interface{}{}

	if direction != "" && direction != "all" {
		query += " AND direction = ?"
		args = append(args, direction)
	}
	if msgType != "" && msgType != "all" {
		query += " AND type = ?"
		args = append(args, msgType)
	}
	if connectionID != "" {
		query += " AND connection_id = ?"
		args = append(args, connectionID)
	}

	query += " ORDER BY timestamp DESC LIMIT ?"
	args = append(args, limit)

	rows, err := a.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var messages []WebSocketMessage
	for rows.Next() {
		var m WebSocketMessage
		var timestamp sql.NullString
		var host, content, contentType, connID sql.NullString

		err := rows.Scan(
			&m.ID, &timestamp, &m.Direction, &m.Type, &m.URL, &host, &m.Size, &content, &contentType, &connID,
		)
		if err != nil {
			continue
		}

		m.Timestamp = timestamp.String
		m.Host = host.String
		m.Content = content.String
		m.ContentType = contentType.String
		m.ConnectionID = connID.String

		messages = append(messages, m)
	}

	return messages, nil
}

// AddWebSocketMessage adds a new WebSocket message
func (a *VulnAPI) AddWebSocketMessage(m WebSocketMessage) (int64, error) {
	if a.db == nil {
		return 0, fmt.Errorf("database not initialized")
	}

	result, err := a.db.Exec(
		`INSERT INTO websocket_messages (direction, type, url, host, size, content, content_type, connection_id)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		m.Direction, m.Type, m.URL, m.Host, m.Size, m.Content, m.ContentType, m.ConnectionID,
	)
	if err != nil {
		return 0, err
	}

	return result.LastInsertId()
}

// ClearWebSocketMessages clears all WebSocket messages
func (a *VulnAPI) ClearWebSocketMessages() error {
	if a.db == nil {
		return fmt.Errorf("database not initialized")
	}

	_, err := a.db.Exec("DELETE FROM websocket_messages")
	return err
}
