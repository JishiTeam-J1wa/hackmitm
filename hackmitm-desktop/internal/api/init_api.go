package api

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	_ "modernc.org/sqlite"
)

// InitAPI handles initialization and database operations
type InitAPI struct {
	ctx        context.Context
	db         *sql.DB
	dbPath     string
	dbInitOnce sync.Once
}

// NewInitAPI creates a new InitAPI instance
func NewInitAPI() *InitAPI {
	return &InitAPI{}
}

// SetContext sets the context for the API
func (a *InitAPI) SetContext(ctx context.Context) {
	a.ctx = ctx
}

// DatabaseConfig represents database configuration
type DatabaseConfig struct {
	Path   string `json:"path"`
	Name   string `json:"name"`
	IsNew  bool   `json:"isNew"`
}

// ServerConfig represents server configuration
type ServerConfig struct {
	Mode      string `json:"mode"`      // "local" or "remote"
	Host      string `json:"host"`      // For remote mode
	Port      int    `json:"port"`      // API port
	ProxyPort int    `json:"proxyPort"` // Proxy port (for local mode)
	APIKey    string `json:"apiKey"`    // Optional API key
}

// InitResult represents initialization result
type InitResult struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Error   string `json:"error,omitempty"`
}

// DictEntry represents a dictionary entry for paths/payloads
type DictEntry struct {
	ID          int64  `json:"id"`
	Category    string `json:"category"`    // e.g., "path", "payload", "header"
	Type        string `json:"type"`        // e.g., "admin", "sql", "xss"
	Content     string `json:"content"`     // The actual path/payload
	Description string `json:"description"` // Description
	Source      string `json:"source"`      // Where it came from
}

// InitDatabase initializes the database
func (a *InitAPI) InitDatabase(config DatabaseConfig) (*InitResult, error) {
	// Ensure directory exists
	dbDir := config.Path
	if err := os.MkdirAll(dbDir, 0755); err != nil {
		return &InitResult{
			Success: false,
			Message: "Failed to create database directory",
			Error:   err.Error(),
		}, err
	}

	// Build database path
	dbFileName := config.Name
	if dbFileName == "" {
		dbFileName = "hackmitm"
	}
	a.dbPath = filepath.Join(dbDir, dbFileName+".db")

	// Check if database exists
	dbExists := false
	if _, err := os.Stat(a.dbPath); err == nil {
		dbExists = true
	}

	// Open database
	db, err := sql.Open("sqlite", a.dbPath)
	if err != nil {
		return &InitResult{
			Success: false,
			Message: "Failed to open database",
			Error:   err.Error(),
		}, err
	}

	a.db = db

	// Initialize schema if new database
	if !dbExists || config.IsNew {
		if err := a.initSchema(); err != nil {
			return &InitResult{
				Success: false,
				Message: "Failed to initialize database schema",
				Error:   err.Error(),
			}, err
		}
	}

	return &InitResult{
		Success: true,
		Message: fmt.Sprintf("Database initialized at %s", a.dbPath),
	}, nil
}

// initSchema creates the database schema
func (a *InitAPI) initSchema() error {
	schema := `
	-- Traffic records
	CREATE TABLE IF NOT EXISTS traffic (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		timestamp DATETIME DEFAULT CURRENT_TIMESTAMP,
		method TEXT NOT NULL,
		url TEXT NOT NULL,
		host TEXT,
		path TEXT,
		status_code INTEGER,
		content_type TEXT,
		request_size INTEGER DEFAULT 0,
		response_size INTEGER DEFAULT 0,
		duration INTEGER DEFAULT 0,
		request_headers TEXT,
		response_headers TEXT,
		request_body TEXT,
		response_body TEXT,
		client_ip TEXT,
		protocol TEXT,
		fingerprint TEXT,
		intercepted INTEGER DEFAULT 0
	);

	CREATE INDEX IF NOT EXISTS idx_traffic_timestamp ON traffic(timestamp);
	CREATE INDEX IF NOT EXISTS idx_traffic_host ON traffic(host);
	CREATE INDEX IF NOT EXISTS idx_traffic_method ON traffic(method);
	CREATE INDEX IF NOT EXISTS idx_traffic_status ON traffic(status_code);

	-- Fingerprint records
	CREATE TABLE IF NOT EXISTS fingerprints (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		url TEXT NOT NULL,
		timestamp DATETIME DEFAULT CURRENT_TIMESTAMP,
		fingerprint TEXT,
		confidence REAL,
		title TEXT,
		status_code INTEGER,
		process_time INTEGER
	);

	CREATE INDEX IF NOT EXISTS idx_fingerprints_url ON fingerprints(url);
	CREATE INDEX IF NOT EXISTS idx_fingerprints_timestamp ON fingerprints(timestamp);

	-- Dictionary table for paths and payloads
	CREATE TABLE IF NOT EXISTS dictionary (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		category TEXT NOT NULL,
		type TEXT NOT NULL,
		content TEXT NOT NULL,
		description TEXT,
		source TEXT,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		UNIQUE(category, type, content)
	);

	CREATE INDEX IF NOT EXISTS idx_dictionary_category ON dictionary(category);
	CREATE INDEX IF NOT EXISTS idx_dictionary_type ON dictionary(type);

	-- Repeater history
	CREATE TABLE IF NOT EXISTS repeater_history (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		timestamp DATETIME DEFAULT CURRENT_TIMESTAMP,
		name TEXT,
		method TEXT,
		url TEXT,
		headers TEXT,
		body TEXT,
		response_status INTEGER,
		response_headers TEXT,
		response_body TEXT,
		response_time INTEGER
	);

	CREATE INDEX IF NOT EXISTS idx_repeater_timestamp ON repeater_history(timestamp);

	-- Scope/Target configuration
	CREATE TABLE IF NOT EXISTS scope (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		host TEXT NOT NULL,
		protocol TEXT DEFAULT 'https',
		include_subdomains INTEGER DEFAULT 0,
		enabled INTEGER DEFAULT 1,
		UNIQUE(host, protocol)
	);

	-- Vulnerabilities table
	CREATE TABLE IF NOT EXISTS vulnerabilities (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		title TEXT NOT NULL,
		severity TEXT NOT NULL,
		type TEXT NOT NULL,
		url TEXT NOT NULL,
		method TEXT,
		request TEXT,
		response TEXT,
		description TEXT,
		remediation TEXT,
		references TEXT,
		status TEXT DEFAULT 'open',
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		source TEXT,
		cwe TEXT,
		cvss REAL
	);

	CREATE INDEX IF NOT EXISTS idx_vuln_severity ON vulnerabilities(severity);
	CREATE INDEX IF NOT EXISTS idx_vuln_status ON vulnerabilities(status);
	CREATE INDEX IF NOT EXISTS idx_vuln_type ON vulnerabilities(type);
	CREATE INDEX IF NOT EXISTS idx_vuln_created ON vulnerabilities(created_at);

	-- Scan results table
	CREATE TABLE IF NOT EXISTS scan_results (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		plugin_name TEXT NOT NULL,
		plugin_id TEXT NOT NULL,
		severity TEXT NOT NULL,
		title TEXT NOT NULL,
		description TEXT,
		url TEXT NOT NULL,
		method TEXT,
		evidence TEXT,
		request TEXT,
		response TEXT,
		timestamp DATETIME DEFAULT CURRENT_TIMESTAMP,
		false_positive INTEGER DEFAULT 0,
		tags TEXT
	);

	CREATE INDEX IF NOT EXISTS idx_scan_severity ON scan_results(severity);
	CREATE INDEX IF NOT EXISTS idx_scan_plugin ON scan_results(plugin_id);
	CREATE INDEX IF NOT EXISTS idx_scan_timestamp ON scan_results(timestamp);
	CREATE INDEX IF NOT EXISTS idx_scan_fp ON scan_results(false_positive);

	-- WebSocket messages table
	CREATE TABLE IF NOT EXISTS websocket_messages (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		timestamp DATETIME DEFAULT CURRENT_TIMESTAMP,
		direction TEXT NOT NULL,
		type TEXT NOT NULL,
		url TEXT NOT NULL,
		host TEXT,
		size INTEGER DEFAULT 0,
		content TEXT,
		content_type TEXT,
		connection_id TEXT
	);

	CREATE INDEX IF NOT EXISTS idx_ws_timestamp ON websocket_messages(timestamp);
	CREATE INDEX IF NOT EXISTS idx_ws_connection ON websocket_messages(connection_id);
	CREATE INDEX IF NOT EXISTS idx_ws_direction ON websocket_messages(direction);

	-- Configuration
	CREATE TABLE IF NOT EXISTS config (
		key TEXT PRIMARY KEY,
		value TEXT,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);

	-- Application state
	CREATE TABLE IF NOT EXISTS app_state (
		key TEXT PRIMARY KEY,
		value TEXT,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);

	-- Intruder attack results
	CREATE TABLE IF NOT EXISTS intruder_results (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		attack_id TEXT NOT NULL,
		payload TEXT,
		status_code INTEGER,
		status_text TEXT,
		response_time INTEGER,
		length INTEGER,
		error TEXT,
		request TEXT,
		response TEXT,
		timestamp DATETIME DEFAULT CURRENT_TIMESTAMP
	);

	CREATE INDEX IF NOT EXISTS idx_intruder_attack_id ON intruder_results(attack_id);
	CREATE INDEX IF NOT EXISTS idx_intruder_timestamp ON intruder_results(timestamp);
	CREATE INDEX IF NOT EXISTS idx_intruder_status ON intruder_results(status_code);

	-- Active scan findings
	CREATE TABLE IF NOT EXISTS active_scan_findings (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		scan_id TEXT NOT NULL,
		plugin_id TEXT NOT NULL,
		plugin_name TEXT NOT NULL,
		severity TEXT NOT NULL,
		title TEXT NOT NULL,
		description TEXT,
		url TEXT NOT NULL,
		method TEXT,
		payload TEXT,
		evidence TEXT,
		request TEXT,
		response TEXT,
		confidence INTEGER DEFAULT 0,
		timestamp DATETIME DEFAULT CURRENT_TIMESTAMP
	);

	CREATE INDEX IF NOT EXISTS idx_active_scan_id ON active_scan_findings(scan_id);
	CREATE INDEX IF NOT EXISTS idx_active_severity ON active_scan_findings(severity);
	CREATE INDEX IF NOT EXISTS idx_active_plugin ON active_scan_findings(plugin_id);
	CREATE INDEX IF NOT EXISTS idx_active_timestamp ON active_scan_findings(timestamp);
	`

	_, err := a.db.Exec(schema)
	if err != nil {
		return err
	}

	// Insert default dictionary entries
	return a.insertDefaultDictionary()
}

// insertDefaultDictionary inserts common paths and payloads
func (a *InitAPI) insertDefaultDictionary() error {
	// Common admin paths
	adminPaths := []struct {
		Type    string
		Content string
		Desc    string
	}{
		{"admin", "/admin", "Admin panel"},
		{"admin", "/admin/login", "Admin login page"},
		{"admin", "/administrator", "Administrator panel"},
		{"admin", "/admin.php", "Admin PHP file"},
		{"admin", "/admin/index.php", "Admin index"},
		{"admin", "/wp-admin", "WordPress admin"},
		{"admin", "/wp-login.php", "WordPress login"},
		{"admin", "/phpmyadmin", "phpMyAdmin"},
		{"admin", "/pma", "phpMyAdmin short"},
		{"admin", "/mysql", "MySQL admin"},
		{"admin", "/manager", "Manager panel"},
		{"admin", "/console", "Console panel"},
		{"admin", "/control", "Control panel"},
		{"admin", "/cpanel", "cPanel"},
		{"admin", "/backend", "Backend panel"},
		{"api", "/api", "API endpoint"},
		{"api", "/api/v1", "API v1"},
		{"api", "/api/v2", "API v2"},
		{"api", "/graphql", "GraphQL endpoint"},
		{"config", "/config.php", "Config file"},
		{"config", "/configuration.php", "Configuration file"},
		{"config", "/settings.php", "Settings file"},
		{"config", "/.env", "Environment file"},
		{"config", "/.git/config", "Git config"},
		{"upload", "/upload", "Upload endpoint"},
		{"upload", "/upload.php", "Upload PHP"},
		{"upload", "/files", "Files directory"},
		{"backup", "/backup", "Backup directory"},
		{"backup", "/backup.sql", "SQL backup"},
		{"backup", "/dump.sql", "SQL dump"},
		{"debug", "/debug", "Debug endpoint"},
		{"debug", "/test", "Test endpoint"},
		{"debug", "/info.php", "PHP info"},
	}

	for _, entry := range adminPaths {
		a.db.Exec(
			`INSERT OR IGNORE INTO dictionary (category, type, content, description, source) VALUES (?, ?, ?, ?, ?)`,
			"path", entry.Type, entry.Content, entry.Desc, "default",
		)
	}

	// SQL Injection payloads
	sqlPayloads := []struct {
		Type    string
		Content string
		Desc    string
	}{
		{"basic", "' OR '1'='1", "Basic SQL injection"},
		{"basic", "' OR '1'='1'--", "Basic with comment"},
		{"basic", "' OR '1'='1'/*", "Basic with block comment"},
		{"basic", "1' OR '1'='1", "Numeric with injection"},
		{"union", "' UNION SELECT NULL--", "Union select null"},
		{"union", "' UNION SELECT NULL,NULL--", "Union select two columns"},
		{"union", "' UNION SELECT 1,2,3--", "Union select numbers"},
		{"error", "'", "Single quote"},
		{"error", "\"", "Double quote"},
		{"error", "\\'", "Escaped quote"},
		{"time", "'; WAITFOR DELAY '0:0:5'--", "Time-based MSSQL"},
		{"time", "' AND SLEEP(5)--", "Time-based MySQL"},
		{"boolean", "1 AND 1=1", "Boolean true"},
		{"boolean", "1 AND 1=2", "Boolean false"},
	}

	for _, entry := range sqlPayloads {
		a.db.Exec(
			`INSERT OR IGNORE INTO dictionary (category, type, content, description, source) VALUES (?, ?, ?, ?, ?)`,
			"payload", "sql_" + entry.Type, entry.Content, entry.Desc, "default",
		)
	}

	// XSS payloads
	xssPayloads := []struct {
		Type    string
		Content string
		Desc    string
	}{
		{"basic", "<script>alert('XSS')</script>", "Basic script tag"},
		{"basic", "<img src=x onerror=alert('XSS')>", "Image onerror"},
		{"basic", "<svg onload=alert('XSS')>", "SVG onload"},
		{"basic", "javascript:alert('XSS')", "JavaScript protocol"},
		{"encoded", "%3Cscript%3Ealert('XSS')%3C/script%3E", "URL encoded"},
		{"event", "\" onfocus=alert('XSS') autofocus \"", "Event handler"},
		{"event", "' onmouseover=alert('XSS') '", "Mouseover event"},
	}

	for _, entry := range xssPayloads {
		a.db.Exec(
			`INSERT OR IGNORE INTO dictionary (category, type, content, description, source) VALUES (?, ?, ?, ?, ?)`,
			"payload", "xss_" + entry.Type, entry.Content, entry.Desc, "default",
		)
	}

	return nil
}

// SelectDatabaseFolder opens a folder selection dialog
func (a *InitAPI) SelectDatabaseFolder() (string, error) {
	// This would use Wails runtime to open a folder dialog
	// For now, return a default path
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(homeDir, ".hackmitm", "data"), nil
}

// GetDatabaseInfo returns information about the current database
func (a *InitAPI) GetDatabaseInfo() (map[string]interface{}, error) {
	if a.db == nil {
		return map[string]interface{}{
			"connected": false,
		}, nil
	}

	// Get file info
	var size int64 = 0
	if info, err := os.Stat(a.dbPath); err == nil {
		size = info.Size()
	}

	// Get table counts
	var trafficCount, fingerprintCount, dictCount int64
	a.db.QueryRow("SELECT COUNT(*) FROM traffic").Scan(&trafficCount)
	a.db.QueryRow("SELECT COUNT(*) FROM fingerprints").Scan(&fingerprintCount)
	a.db.QueryRow("SELECT COUNT(*) FROM dictionary").Scan(&dictCount)

	return map[string]interface{}{
		"connected":         true,
		"path":              a.dbPath,
		"size":              size,
		"traffic_count":     trafficCount,
		"fingerprint_count": fingerprintCount,
		"dictionary_count":  dictCount,
	}, nil
}

// GetDictionaryEntries retrieves dictionary entries
func (a *InitAPI) GetDictionaryEntries(category, dictType string, limit int) ([]DictEntry, error) {
	if a.db == nil {
		return nil, fmt.Errorf("database not initialized")
	}

	query := "SELECT id, category, type, content, description, source FROM dictionary WHERE 1=1"
	args := []interface{}{}

	if category != "" {
		query += " AND category = ?"
		args = append(args, category)
	}
	if dictType != "" {
		query += " AND type = ?"
		args = append(args, dictType)
	}

	query += " ORDER BY id LIMIT ?"
	args = append(args, limit)

	rows, err := a.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var entries []DictEntry
	for rows.Next() {
		var e DictEntry
		if err := rows.Scan(&e.ID, &e.Category, &e.Type, &e.Content, &e.Description, &e.Source); err != nil {
			continue
		}
		entries = append(entries, e)
	}

	return entries, nil
}

// AddDictionaryEntry adds a new dictionary entry
func (a *InitAPI) AddDictionaryEntry(entry DictEntry) error {
	if a.db == nil {
		return fmt.Errorf("database not initialized")
	}

	_, err := a.db.Exec(
		`INSERT OR IGNORE INTO dictionary (category, type, content, description, source) VALUES (?, ?, ?, ?, ?)`,
		entry.Category, entry.Type, entry.Content, entry.Description, entry.Source,
	)
	return err
}

// ImportDictionary imports entries from a JSON file
func (a *InitAPI) ImportDictionary(filePath string) (int, error) {
	if a.db == nil {
		return 0, fmt.Errorf("database not initialized")
	}

	data, err := os.ReadFile(filePath)
	if err != nil {
		return 0, err
	}

	var entries []DictEntry
	if err := json.Unmarshal(data, &entries); err != nil {
		return 0, err
	}

	count := 0
	for _, entry := range entries {
		if err := a.AddDictionaryEntry(entry); err == nil {
			count++
		}
	}

	return count, nil
}

// ExportDictionary exports dictionary entries to a JSON file
func (a *InitAPI) ExportDictionary(filePath, category string) error {
	if a.db == nil {
		return fmt.Errorf("database not initialized")
	}

	entries, err := a.GetDictionaryEntries(category, "", 0)
	if err != nil {
		return err
	}

	data, err := json.MarshalIndent(entries, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(filePath, data, 0644)
}

// Close closes the database connection
func (a *InitAPI) Close() error {
	if a.db != nil {
		return a.db.Close()
	}
	return nil
}

// GetDB returns the database connection
func (a *InitAPI) GetDB() *sql.DB {
	return a.db
}
