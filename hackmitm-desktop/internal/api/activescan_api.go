package api

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"hackmitm-desktop/internal/activescan"
	"hackmitm-desktop/internal/activescan/plugins"
)

// ActiveScanAPI handles active scanning operations
type ActiveScanAPI struct {
	ctx     context.Context
	db      *sql.DB
	scans   map[string]*activescan.ActiveScan
	mu      sync.RWMutex
}

// NewActiveScanAPI creates a new ActiveScanAPI instance
func NewActiveScanAPI() *ActiveScanAPI {
	return &ActiveScanAPI{
		scans: make(map[string]*activescan.ActiveScan),
	}
}

// SetContext sets the application context
func (a *ActiveScanAPI) SetContext(ctx context.Context) {
	a.ctx = ctx
}

// SetDB sets the database connection
func (a *ActiveScanAPI) SetDB(db *sql.DB) {
	a.db = db
}

// ============ Target Management ============

// AddTarget adds a target to a scan
func (a *ActiveScanAPI) AddTarget(scanID string, target activescan.Target) error {
	a.mu.RLock()
	scan, exists := a.scans[scanID]
	a.mu.RUnlock()

	if !exists {
		return fmt.Errorf("scan not found: %s", scanID)
	}

	scan.AddTarget(&target)
	return nil
}

// RemoveTarget removes a target from a scan
func (a *ActiveScanAPI) RemoveTarget(scanID, targetID string) error {
	a.mu.RLock()
	scan, exists := a.scans[scanID]
	a.mu.RUnlock()

	if !exists {
		return fmt.Errorf("scan not found: %s", scanID)
	}

	scan.RemoveTarget(targetID)
	return nil
}

// GetTargets returns all targets for a scan
func (a *ActiveScanAPI) GetTargets(scanID string) ([]*activescan.Target, error) {
	a.mu.RLock()
	scan, exists := a.scans[scanID]
	a.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("scan not found: %s", scanID)
	}

	return scan.GetTargets(), nil
}

// ClearTargets clears all targets from a scan
func (a *ActiveScanAPI) ClearTargets(scanID string) error {
	a.mu.RLock()
	scan, exists := a.scans[scanID]
	a.mu.RUnlock()

	if !exists {
		return fmt.Errorf("scan not found: %s", scanID)
	}

	scan.ClearTargets()
	return nil
}

// ============ Scan Management ============

// CreateScan creates a new scan with the given configuration
func (a *ActiveScanAPI) CreateScan(config activescan.ScanConfig) (*activescan.ActiveScan, error) {
	a.mu.Lock()
	defer a.mu.Unlock()

	scan := activescan.NewActiveScan(&config)

	// Register default plugins
	scan.RegisterPlugin(plugins.NewSQLInjectionPlugin())
	scan.RegisterPlugin(plugins.NewXSSPlugin())
	scan.RegisterPlugin(plugins.NewTraversalPlugin())
	scan.RegisterPlugin(plugins.NewCommandInjectionPlugin())

	// Set up callbacks
	scan.OnProgress = func(progress *activescan.ScanProgress) {
		// Could emit event via context if available
	}

	scan.OnFinding = func(finding *activescan.Finding) {
		// Store finding in database
		if a.db != nil {
			go a.storeFinding(config.ID, finding)
		}
	}

	a.scans[config.ID] = scan
	return scan, nil
}

// StartScan starts a scan
func (a *ActiveScanAPI) StartScan(scanID string) error {
	a.mu.RLock()
	scan, exists := a.scans[scanID]
	a.mu.RUnlock()

	if !exists {
		return fmt.Errorf("scan not found: %s", scanID)
	}

	return scan.Start(a.ctx)
}

// PauseScan pauses a running scan
func (a *ActiveScanAPI) PauseScan(scanID string) error {
	a.mu.RLock()
	scan, exists := a.scans[scanID]
	a.mu.RUnlock()

	if !exists {
		return fmt.Errorf("scan not found: %s", scanID)
	}

	scan.Pause()
	return nil
}

// ResumeScan resumes a paused scan
func (a *ActiveScanAPI) ResumeScan(scanID string) error {
	a.mu.RLock()
	scan, exists := a.scans[scanID]
	a.mu.RUnlock()

	if !exists {
		return fmt.Errorf("scan not found: %s", scanID)
	}

	scan.Resume()
	return nil
}

// StopScan stops a scan
func (a *ActiveScanAPI) StopScan(scanID string) error {
	a.mu.RLock()
	scan, exists := a.scans[scanID]
	a.mu.RUnlock()

	if !exists {
		return fmt.Errorf("scan not found: %s", scanID)
	}

	scan.Stop()
	return nil
}

// GetScanProgress returns the progress of a scan
func (a *ActiveScanAPI) GetScanProgress(scanID string) (*activescan.ScanProgress, error) {
	a.mu.RLock()
	scan, exists := a.scans[scanID]
	a.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("scan not found: %s", scanID)
	}

	return scan.GetProgress(), nil
}

// GetScanFindings returns all findings from a scan
func (a *ActiveScanAPI) GetScanFindings(scanID string) ([]*activescan.Finding, error) {
	a.mu.RLock()
	scan, exists := a.scans[scanID]
	a.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("scan not found: %s", scanID)
	}

	return scan.GetFindings(), nil
}

// GetScanStatus returns the status of a scan
func (a *ActiveScanAPI) GetScanStatus(scanID string) (activescan.ScanStatus, error) {
	a.mu.RLock()
	scan, exists := a.scans[scanID]
	a.mu.RUnlock()

	if !exists {
		return "", fmt.Errorf("scan not found: %s", scanID)
	}

	return scan.GetStatus(), nil
}

// RemoveScan removes a scan
func (a *ActiveScanAPI) RemoveScan(scanID string) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	delete(a.scans, scanID)
	return nil
}

// ListScans returns all scan IDs
func (a *ActiveScanAPI) ListScans() []string {
	a.mu.RLock()
	defer a.mu.RUnlock()

	ids := make([]string, 0, len(a.scans))
	for id := range a.scans {
		ids = append(ids, id)
	}
	return ids
}

// ============ Plugin Management ============

// GetPlugins returns all available plugins
func (a *ActiveScanAPI) GetPlugins(scanID string) ([]map[string]interface{}, error) {
	a.mu.RLock()
	scan, exists := a.scans[scanID]
	a.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("scan not found: %s", scanID)
	}

	pluginList := scan.GetPlugins()
	var result []map[string]interface{}

	for _, p := range pluginList {
		result = append(result, map[string]interface{}{
			"id":          p.ID(),
			"name":        p.Name(),
			"description": p.Description(),
			"severity":    string(p.Severity()),
			"enabled":     p.Enabled(),
		})
	}

	return result, nil
}

// EnablePlugin enables a plugin
func (a *ActiveScanAPI) EnablePlugin(scanID, pluginID string) error {
	a.mu.RLock()
	scan, exists := a.scans[scanID]
	a.mu.RUnlock()

	if !exists {
		return fmt.Errorf("scan not found: %s", scanID)
	}

	scan.EnablePlugin(pluginID)
	return nil
}

// DisablePlugin disables a plugin
func (a *ActiveScanAPI) DisablePlugin(scanID, pluginID string) error {
	a.mu.RLock()
	scan, exists := a.scans[scanID]
	a.mu.RUnlock()

	if !exists {
		return fmt.Errorf("scan not found: %s", scanID)
	}

	scan.DisablePlugin(pluginID)
	return nil
}

// ============ Default Plugins ============

// GetDefaultPlugins returns information about default plugins
func (a *ActiveScanAPI) GetDefaultPlugins() []map[string]interface{} {
	return []map[string]interface{}{
		{
			"id":          "SQL-INJECTION",
			"name":        "SQL Injection",
			"description": "Detects SQL injection vulnerabilities by injecting various SQL payloads",
			"severity":    "high",
			"enabled":     true,
		},
		{
			"id":          "XSS",
			"name":        "Cross-Site Scripting",
			"description": "Detects reflected and stored XSS vulnerabilities",
			"severity":    "high",
			"enabled":     true,
		},
		{
			"id":          "PATH-TRAVERSAL",
			"name":        "Path Traversal",
			"description": "Detects path traversal vulnerabilities that allow reading files outside web root",
			"severity":    "high",
			"enabled":     true,
		},
		{
			"id":          "COMMAND-INJECTION",
			"name":        "Command Injection",
			"description": "Detects OS command injection vulnerabilities",
			"severity":    "critical",
			"enabled":     true,
		},
	}
}

// ============ Database Storage ============

// storeFinding stores a finding in the database
func (a *ActiveScanAPI) storeFinding(scanID string, finding *activescan.Finding) error {
	if a.db == nil {
		return nil
	}

	_, err := a.db.Exec(`
		INSERT INTO active_scan_findings (scan_id, plugin_id, plugin_name, severity, title, description, url, method, payload, evidence, request, response, confidence, timestamp)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		scanID, finding.PluginID, finding.PluginName, string(finding.Severity),
		finding.Title, finding.Description, finding.URL, finding.Method,
		finding.Payload, finding.Evidence, finding.Request, finding.Response,
		finding.Confidence, finding.Timestamp.Format(time.RFC3339),
	)

	return err
}

// GetStoredFindings retrieves findings from database
func (a *ActiveScanAPI) GetStoredFindings(scanID string, limit int) ([]ActiveScanFinding, error) {
	if a.db == nil {
		return nil, fmt.Errorf("database not initialized")
	}

	query := `SELECT id, scan_id, plugin_id, plugin_name, severity, title, description, url, method, payload, evidence, request, response, confidence, timestamp
			  FROM active_scan_findings WHERE scan_id = ? ORDER BY timestamp DESC LIMIT ?`

	rows, err := a.db.Query(query, scanID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var findings []ActiveScanFinding
	for rows.Next() {
		var f ActiveScanFinding
		var timestamp sql.NullString

		err := rows.Scan(
			&f.ID, &f.ScanID, &f.PluginID, &f.PluginName, &f.Severity,
			&f.Title, &f.Description, &f.URL, &f.Method, &f.Payload,
			&f.Evidence, &f.Request, &f.Response, &f.Confidence, &timestamp,
		)
		if err != nil {
			continue
		}

		f.Timestamp = timestamp.String
		findings = append(findings, f)
	}

	return findings, nil
}

// ActiveScanFinding represents a stored active scan finding
type ActiveScanFinding struct {
	ID          int64  `json:"id"`
	ScanID      string `json:"scanId"`
	PluginID    string `json:"pluginId"`
	PluginName  string `json:"pluginName"`
	Severity    string `json:"severity"`
	Title       string `json:"title"`
	Description string `json:"description"`
	URL         string `json:"url"`
	Method      string `json:"method"`
	Payload     string `json:"payload"`
	Evidence    string `json:"evidence"`
	Request     string `json:"request"`
	Response    string `json:"response"`
	Confidence  int    `json:"confidence"`
	Timestamp   string `json:"timestamp"`
}

// ============ Utility Functions ============

// CreateScanConfig creates a scan configuration
func (a *ActiveScanAPI) CreateScanConfig(id, name string, concurrency, rateLimit, timeout int, followRedirects bool, enabledPlugins []string) *activescan.ScanConfig {
	return &activescan.ScanConfig{
		ID:              id,
		Name:            name,
		Concurrency:     concurrency,
		RateLimit:       rateLimit,
		Timeout:         timeout,
		FollowRedirects: followRedirects,
		EnabledPlugins:  enabledPlugins,
	}
}

// CreateTarget creates a target for scanning
func (a *ActiveScanAPI) CreateTarget(id, url, method string, headers map[string]string, body string) *activescan.Target {
	return &activescan.Target{
		ID:      id,
		URL:     url,
		Method:  method,
		Headers: headers,
		Body:    body,
		Enabled: true,
	}
}

// ExportFindingsToJSON exports findings to JSON format
func (a *ActiveScanAPI) ExportFindingsToJSON(scanID string) ([]byte, error) {
	findings, err := a.GetScanFindings(scanID)
	if err != nil {
		return nil, err
	}

	return json.MarshalIndent(findings, "", "  ")
}
