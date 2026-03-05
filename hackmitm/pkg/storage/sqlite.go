package storage

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	_ "modernc.org/sqlite" // 纯Go SQLite驱动，无需CGO
)

// TrafficRecord 流量记录
type TrafficRecord struct {
	ID              int64
	Timestamp       time.Time
	Method          string
	URL             string
	Host            string
	Path            string
	StatusCode      int
	ContentType     string
	RequestSize     int64
	ResponseSize    int64
	Duration        int64
	RequestHeaders  string
	ResponseHeaders string
	RequestBody     string
	ResponseBody    string
	ClientIP        string
	Protocol        string
	Fingerprint     string // 识别到的指纹
}

// FingerprintRecord 指纹识别记录
type FingerprintRecord struct {
	ID          int64
	URL         string
	Timestamp   time.Time
	Fingerprint string
	Confidence  float64
	Title       string
	StatusCode  int
	ProcessTime int64
}

// SQLiteStorage SQLite存储
type SQLiteStorage struct {
	db    *sql.DB
	path  string
	mutex sync.RWMutex
}

// NewSQLiteStorage 创建SQLite存储
func NewSQLiteStorage(dataDir string) (*SQLiteStorage, error) {
	// 确保数据目录存在
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return nil, fmt.Errorf("创建数据目录失败: %w", err)
	}

	dbPath := filepath.Join(dataDir, "hackmitm.db")

	// 打开数据库
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("打开数据库失败: %w", err)
	}

	// 设置连接池
	db.SetMaxOpenConns(1) // SQLite建议单连接
	db.SetMaxIdleConns(1)

	storage := &SQLiteStorage{
		db:   db,
		path: dbPath,
	}

	// 初始化表结构
	if err := storage.initSchema(); err != nil {
		db.Close()
		return nil, fmt.Errorf("初始化数据库失败: %w", err)
	}

	return storage, nil
}

// initSchema 初始化数据库表结构
func (s *SQLiteStorage) initSchema() error {
	schema := `
	-- 流量表
	CREATE TABLE IF NOT EXISTS traffic (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		session_id TEXT DEFAULT 'default',
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
		tags TEXT,
		notes TEXT
	);

	-- 创建索引
	CREATE INDEX IF NOT EXISTS idx_traffic_timestamp ON traffic(timestamp);
	CREATE INDEX IF NOT EXISTS idx_traffic_host ON traffic(host);
	CREATE INDEX IF NOT EXISTS idx_traffic_method ON traffic(method);
	CREATE INDEX IF NOT EXISTS idx_traffic_status ON traffic(status_code);
	CREATE INDEX IF NOT EXISTS idx_traffic_session ON traffic(session_id);

	-- 指纹识别记录表
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

	-- 创建索引
	CREATE INDEX IF NOT EXISTS idx_fingerprints_url ON fingerprints(url);
	CREATE INDEX IF NOT EXISTS idx_fingerprints_timestamp ON fingerprints(timestamp);

	-- 漏洞表
	CREATE TABLE IF NOT EXISTS vulnerabilities (
		id TEXT PRIMARY KEY,
		session_id TEXT DEFAULT 'default',
		traffic_id INTEGER,
		rule_id TEXT NOT NULL,
		hash TEXT UNIQUE NOT NULL,
		name TEXT NOT NULL,
		description TEXT,
		severity TEXT DEFAULT 'medium',
		confidence REAL DEFAULT 0.0,
		url TEXT,
		parameter TEXT,
		evidence TEXT,
		remediation TEXT,
		status TEXT DEFAULT 'open',
		occurrences INTEGER DEFAULT 1,
		first_seen DATETIME DEFAULT CURRENT_TIMESTAMP,
		last_seen DATETIME DEFAULT CURRENT_TIMESTAMP,
		timestamp DATETIME DEFAULT CURRENT_TIMESTAMP,
		metadata TEXT,
		request_method TEXT,
		request_url TEXT,
		request_headers TEXT,
		request_body TEXT,
		response_status INTEGER,
		response_headers TEXT,
		response_body TEXT,
		tags TEXT,
		notes TEXT
	);

	-- 漏洞索引
	CREATE INDEX IF NOT EXISTS idx_vulns_session ON vulnerabilities(session_id);
	CREATE INDEX IF NOT EXISTS idx_vulns_severity ON vulnerabilities(severity);
	CREATE INDEX IF NOT EXISTS idx_vulns_status ON vulnerabilities(status);
	CREATE INDEX IF NOT EXISTS idx_vulns_rule ON vulnerabilities(rule_id);
	CREATE INDEX IF NOT EXISTS idx_vulns_hash ON vulnerabilities(hash);
	CREATE INDEX IF NOT EXISTS idx_vulns_timestamp ON vulnerabilities(timestamp);

	-- 扫描结果表
	CREATE TABLE IF NOT EXISTS scan_results (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		traffic_id INTEGER NOT NULL,
		rule_id TEXT NOT NULL,
		matched INTEGER DEFAULT 0,
		confidence REAL DEFAULT 0.0,
		evidence TEXT,
		metadata TEXT,
		timestamp DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (traffic_id) REFERENCES traffic(id)
	);

	-- 扫描结果索引
	CREATE INDEX IF NOT EXISTS idx_scan_traffic ON scan_results(traffic_id);
	CREATE INDEX IF NOT EXISTS idx_scan_rule ON scan_results(rule_id);

	-- 会话表
	CREATE TABLE IF NOT EXISTS sessions (
		id TEXT PRIMARY KEY,
		name TEXT NOT NULL,
		description TEXT,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		traffic_count INTEGER DEFAULT 0,
		vuln_count INTEGER DEFAULT 0,
		metadata TEXT
	);

	-- 扫描规则配置表
	CREATE TABLE IF NOT EXISTS scan_rules (
		id TEXT PRIMARY KEY,
		name TEXT NOT NULL,
		description TEXT,
		severity TEXT DEFAULT 'medium',
		enabled INTEGER DEFAULT 1,
		priority INTEGER DEFAULT 0,
		config TEXT,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);

	-- 配置表
	CREATE TABLE IF NOT EXISTS config (
		key TEXT PRIMARY KEY,
		value TEXT,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);
	`

	_, err := s.db.Exec(schema)
	return err
}

// Close 关闭数据库
func (s *SQLiteStorage) Close() error {
	return s.db.Close()
}

// ============ 流量操作 ============

// SaveTraffic 保存流量记录
func (s *SQLiteStorage) SaveTraffic(record *TrafficRecord) (int64, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	result, err := s.db.Exec(`
		INSERT INTO traffic (
			timestamp, method, url, host, path, status_code, content_type,
			request_size, response_size, duration, request_headers, response_headers,
			request_body, response_body, client_ip, protocol, fingerprint
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`,
		record.Timestamp, record.Method, record.URL, record.Host, record.Path,
		record.StatusCode, record.ContentType, record.RequestSize, record.ResponseSize,
		record.Duration, record.RequestHeaders, record.ResponseHeaders,
		record.RequestBody, record.ResponseBody, record.ClientIP, record.Protocol,
		record.Fingerprint,
	)
	if err != nil {
		return 0, err
	}

	return result.LastInsertId()
}

// GetTraffic 获取流量列表
func (s *SQLiteStorage) GetTraffic(limit int, offset int) ([]TrafficRecord, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	if limit <= 0 {
		limit = 1000
	}

	rows, err := s.db.Query(`
		SELECT id, timestamp, method, url, host, path, status_code, content_type,
			   request_size, response_size, duration, client_ip, protocol
		FROM traffic
		ORDER BY timestamp DESC
		LIMIT ? OFFSET ?
	`, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var records []TrafficRecord
	for rows.Next() {
		var r TrafficRecord
		var timestampStr string
		err := rows.Scan(
			&r.ID, &timestampStr, &r.Method, &r.URL, &r.Host, &r.Path,
			&r.StatusCode, &r.ContentType, &r.RequestSize, &r.ResponseSize,
			&r.Duration, &r.ClientIP, &r.Protocol,
		)
		if err != nil {
			continue
		}
		r.Timestamp, _ = time.Parse("2006-01-02 15:04:05", timestampStr)
		records = append(records, r)
	}

	return records, nil
}

// GetTrafficByID 根据ID获取流量详情
func (s *SQLiteStorage) GetTrafficByID(id int64) (*TrafficRecord, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	var r TrafficRecord
	var timestampStr string
	err := s.db.QueryRow(`
		SELECT id, timestamp, method, url, host, path, status_code, content_type,
			   request_size, response_size, duration, request_headers, response_headers,
			   request_body, response_body, client_ip, protocol, fingerprint
		FROM traffic
		WHERE id = ?
	`, id).Scan(
		&r.ID, &timestampStr, &r.Method, &r.URL, &r.Host, &r.Path,
		&r.StatusCode, &r.ContentType, &r.RequestSize, &r.ResponseSize,
		&r.Duration, &r.RequestHeaders, &r.ResponseHeaders,
		&r.RequestBody, &r.ResponseBody, &r.ClientIP, &r.Protocol, &r.Fingerprint,
	)
	if err != nil {
		return nil, err
	}
	r.Timestamp, _ = time.Parse("2006-01-02 15:04:05", timestampStr)
	return &r, nil
}

// ClearTraffic 清空流量记录
func (s *SQLiteStorage) ClearTraffic() error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	_, err := s.db.Exec("DELETE FROM traffic")
	return err
}

// GetTrafficCount 获取流量总数
func (s *SQLiteStorage) GetTrafficCount() (int64, error) {
	var count int64
	err := s.db.QueryRow("SELECT COUNT(*) FROM traffic").Scan(&count)
	return count, err
}

// ============ 指纹操作 ============

// SaveFingerprint 保存指纹识别记录
func (s *SQLiteStorage) SaveFingerprint(record *FingerprintRecord) (int64, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	result, err := s.db.Exec(`
		INSERT INTO fingerprints (url, timestamp, fingerprint, confidence, title, status_code, process_time)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`,
		record.URL, record.Timestamp, record.Fingerprint, record.Confidence,
		record.Title, record.StatusCode, record.ProcessTime,
	)
	if err != nil {
		return 0, err
	}

	return result.LastInsertId()
}

// GetFingerprintHistory 获取指纹识别历史
func (s *SQLiteStorage) GetFingerprintHistory(limit int) ([]FingerprintRecord, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	if limit <= 0 {
		limit = 100
	}

	rows, err := s.db.Query(`
		SELECT id, url, timestamp, fingerprint, confidence, title, status_code, process_time
		FROM fingerprints
		ORDER BY timestamp DESC
		LIMIT ?
	`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var records []FingerprintRecord
	for rows.Next() {
		var r FingerprintRecord
		var timestampStr string
		err := rows.Scan(
			&r.ID, &r.URL, &timestampStr, &r.Fingerprint, &r.Confidence,
			&r.Title, &r.StatusCode, &r.ProcessTime,
		)
		if err != nil {
			continue
		}
		r.Timestamp, _ = time.Parse("2006-01-02 15:04:05", timestampStr)
		records = append(records, r)
	}

	return records, nil
}

// ============ 配置操作 ============

// SetConfig 设置配置
func (s *SQLiteStorage) SetConfig(key, value string) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	_, err := s.db.Exec(`
		INSERT INTO config (key, value, updated_at)
		VALUES (?, ?, CURRENT_TIMESTAMP)
		ON CONFLICT(key) DO UPDATE SET value = ?, updated_at = CURRENT_TIMESTAMP
	`, key, value, value)
	return err
}

// GetConfig 获取配置
func (s *SQLiteStorage) GetConfig(key string) (string, error) {
	var value string
	err := s.db.QueryRow("SELECT value FROM config WHERE key = ?", key).Scan(&value)
	if err == sql.ErrNoRows {
		return "", nil
	}
	return value, err
}

// GetDBPath 获取数据库路径
func (s *SQLiteStorage) GetDBPath() string {
	return s.path
}

// GetDBSize 获取数据库大小
func (s *SQLiteStorage) GetDBSize() int64 {
	info, err := os.Stat(s.path)
	if err != nil {
		return 0
	}
	return info.Size()
}

// ============ 漏洞操作 ============

// VulnerabilityRecord 漏洞记录
type VulnerabilityRecord struct {
	ID              string
	SessionID       string
	TrafficID       int64
	RuleID          string
	Hash            string
	Name            string
	Description     string
	Severity        string
	Confidence      float64
	URL             string
	Parameter       string
	Evidence        string
	Remediation     string
	Status          string
	Occurrences     int
	FirstSeen       time.Time
	LastSeen        time.Time
	Timestamp       time.Time
	Metadata        string
	RequestMethod   string
	RequestURL      string
	RequestHeaders  string
	RequestBody     string
	ResponseStatus  int
	ResponseHeaders string
	ResponseBody    string
	Tags            string
	Notes           string
}

// SaveVulnerability 保存漏洞记录
func (s *SQLiteStorage) SaveVulnerability(record *VulnerabilityRecord) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	_, err := s.db.Exec(`
		INSERT INTO vulnerabilities (
			id, session_id, traffic_id, rule_id, hash, name, description, severity,
			confidence, url, parameter, evidence, remediation, status, occurrences,
			first_seen, last_seen, timestamp, metadata, request_method, request_url,
			request_headers, request_body, response_status, response_headers,
			response_body, tags, notes
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(hash) DO UPDATE SET
			occurrences = occurrences + 1,
			last_seen = CURRENT_TIMESTAMP
	`,
		record.ID, record.SessionID, record.TrafficID, record.RuleID, record.Hash,
		record.Name, record.Description, record.Severity, record.Confidence,
		record.URL, record.Parameter, record.Evidence, record.Remediation,
		record.Status, record.Occurrences, record.FirstSeen, record.LastSeen,
		record.Timestamp, record.Metadata, record.RequestMethod, record.RequestURL,
		record.RequestHeaders, record.RequestBody, record.ResponseStatus,
		record.ResponseHeaders, record.ResponseBody, record.Tags, record.Notes,
	)
	return err
}

// GetVulnerability 获取漏洞
func (s *SQLiteStorage) GetVulnerability(id string) (*VulnerabilityRecord, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	var r VulnerabilityRecord
	var firstSeen, lastSeen, timestamp sql.NullString
	err := s.db.QueryRow(`
		SELECT id, session_id, traffic_id, rule_id, hash, name, description, severity,
			   confidence, url, parameter, evidence, remediation, status, occurrences,
			   first_seen, last_seen, timestamp, metadata, request_method, request_url,
			   request_headers, request_body, response_status, response_headers,
			   response_body, tags, notes
		FROM vulnerabilities WHERE id = ?
	`, id).Scan(
		&r.ID, &r.SessionID, &r.TrafficID, &r.RuleID, &r.Hash, &r.Name,
		&r.Description, &r.Severity, &r.Confidence, &r.URL, &r.Parameter,
		&r.Evidence, &r.Remediation, &r.Status, &r.Occurrences,
		&firstSeen, &lastSeen, &timestamp, &r.Metadata, &r.RequestMethod,
		&r.RequestURL, &r.RequestHeaders, &r.RequestBody, &r.ResponseStatus,
		&r.ResponseHeaders, &r.ResponseBody, &r.Tags, &r.Notes,
	)
	if err != nil {
		return nil, err
	}
	r.FirstSeen, _ = time.Parse("2006-01-02 15:04:05", firstSeen.String)
	r.LastSeen, _ = time.Parse("2006-01-02 15:04:05", lastSeen.String)
	r.Timestamp, _ = time.Parse("2006-01-02 15:04:05", timestamp.String)
	return &r, nil
}

// ListVulnerabilities 列出漏洞
func (s *SQLiteStorage) ListVulnerabilities(sessionID, severity, status string, limit, offset int) ([]VulnerabilityRecord, int, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	// 构建查询条件
	where := "WHERE 1=1"
	args := []interface{}{}
	if sessionID != "" {
		where += " AND session_id = ?"
		args = append(args, sessionID)
	}
	if severity != "" {
		where += " AND severity = ?"
		args = append(args, severity)
	}
	if status != "" {
		where += " AND status = ?"
		args = append(args, status)
	}

	// 获取总数
	var total int
	countQuery := "SELECT COUNT(*) FROM vulnerabilities " + where
	err := s.db.QueryRow(countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	// 获取列表
	if limit <= 0 {
		limit = 100
	}
	query := `SELECT id, session_id, traffic_id, rule_id, name, severity, confidence,
			  url, parameter, status, occurrences, first_seen, last_seen
			  FROM vulnerabilities ` + where + ` ORDER BY timestamp DESC LIMIT ? OFFSET ?`
	args = append(args, limit, offset)

	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var records []VulnerabilityRecord
	for rows.Next() {
		var r VulnerabilityRecord
		var firstSeen, lastSeen sql.NullString
		err := rows.Scan(
			&r.ID, &r.SessionID, &r.TrafficID, &r.RuleID, &r.Name,
			&r.Severity, &r.Confidence, &r.URL, &r.Parameter,
			&r.Status, &r.Occurrences, &firstSeen, &lastSeen,
		)
		if err != nil {
			continue
		}
		r.FirstSeen, _ = time.Parse("2006-01-02 15:04:05", firstSeen.String)
		r.LastSeen, _ = time.Parse("2006-01-02 15:04:05", lastSeen.String)
		records = append(records, r)
	}

	return records, total, nil
}

// UpdateVulnerabilityStatus 更新漏洞状态
func (s *SQLiteStorage) UpdateVulnerabilityStatus(id, status, notes string) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	_, err := s.db.Exec("UPDATE vulnerabilities SET status = ?, notes = ? WHERE id = ?", status, notes, id)
	return err
}

// DeleteVulnerability 删除漏洞
func (s *SQLiteStorage) DeleteVulnerability(id string) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	_, err := s.db.Exec("DELETE FROM vulnerabilities WHERE id = ?", id)
	return err
}

// GetVulnerabilityStats 获取漏洞统计
func (s *SQLiteStorage) GetVulnerabilityStats(sessionID string) (map[string]int, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	stats := map[string]int{
		"total":    0,
		"critical": 0,
		"high":     0,
		"medium":   0,
		"low":      0,
		"info":     0,
		"open":     0,
		"confirmed": 0,
		"fixed":    0,
		"false_positive": 0,
	}

	where := ""
	args := []interface{}{}
	if sessionID != "" {
		where = "WHERE session_id = ?"
		args = append(args, sessionID)
	}

	// 按严重程度统计
	rows, err := s.db.Query("SELECT severity, COUNT(*) FROM vulnerabilities "+where+" GROUP BY severity", args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var severity string
		var count int
		rows.Scan(&severity, &count)
		stats[severity] = count
		stats["total"] += count
	}

	// 按状态统计
	rows2, err := s.db.Query("SELECT status, COUNT(*) FROM vulnerabilities "+where+" GROUP BY status", args...)
	if err != nil {
		return nil, err
	}
	defer rows2.Close()
	for rows2.Next() {
		var status string
		var count int
		rows2.Scan(&status, &count)
		stats[status] = count
	}

	return stats, nil
}

// ============ 扫描结果操作 ============

// ScanResultRecord 扫描结果记录
type ScanResultRecord struct {
	ID         int64
	TrafficID  int64
	RuleID     string
	Matched    bool
	Confidence float64
	Evidence   string
	Metadata   string
	Timestamp  time.Time
}

// SaveScanResult 保存扫描结果
func (s *SQLiteStorage) SaveScanResult(record *ScanResultRecord) (int64, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	result, err := s.db.Exec(`
		INSERT INTO scan_results (traffic_id, rule_id, matched, confidence, evidence, metadata, timestamp)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`, record.TrafficID, record.RuleID, record.Matched, record.Confidence,
		record.Evidence, record.Metadata, record.Timestamp)
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

// GetScanResultsByTraffic 获取流量的扫描结果
func (s *SQLiteStorage) GetScanResultsByTraffic(trafficID int64) ([]ScanResultRecord, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	rows, err := s.db.Query(`
		SELECT id, traffic_id, rule_id, matched, confidence, evidence, metadata, timestamp
		FROM scan_results WHERE traffic_id = ?
	`, trafficID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var records []ScanResultRecord
	for rows.Next() {
		var r ScanResultRecord
		var timestamp sql.NullString
		err := rows.Scan(
			&r.ID, &r.TrafficID, &r.RuleID, &r.Matched, &r.Confidence,
			&r.Evidence, &r.Metadata, &timestamp,
		)
		if err != nil {
			continue
		}
		r.Timestamp, _ = time.Parse("2006-01-02 15:04:05", timestamp.String)
		records = append(records, r)
	}

	return records, nil
}

// ============ 会话操作 ============

// SessionRecord 会话记录
type SessionRecord struct {
	ID           string
	Name         string
	Description  string
	CreatedAt    time.Time
	UpdatedAt    time.Time
	TrafficCount int
	VulnCount    int
	Metadata     string
}

// SaveSession 保存会话
func (s *SQLiteStorage) SaveSession(record *SessionRecord) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	_, err := s.db.Exec(`
		INSERT INTO sessions (id, name, description, created_at, updated_at, traffic_count, vuln_count, metadata)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(id) DO UPDATE SET
			name = ?, description = ?, updated_at = CURRENT_TIMESTAMP,
			traffic_count = ?, vuln_count = ?, metadata = ?
	`,
		record.ID, record.Name, record.Description, record.CreatedAt, record.UpdatedAt,
		record.TrafficCount, record.VulnCount, record.Metadata,
		record.Name, record.Description, record.TrafficCount, record.VulnCount, record.Metadata,
	)
	return err
}

// GetSession 获取会话
func (s *SQLiteStorage) GetSession(id string) (*SessionRecord, error) {
	var r SessionRecord
	var createdAt, updatedAt sql.NullString
	err := s.db.QueryRow(`
		SELECT id, name, description, created_at, updated_at, traffic_count, vuln_count, metadata
		FROM sessions WHERE id = ?
	`, id).Scan(
		&r.ID, &r.Name, &r.Description, &createdAt, &updatedAt,
		&r.TrafficCount, &r.VulnCount, &r.Metadata,
	)
	if err != nil {
		return nil, err
	}
	r.CreatedAt, _ = time.Parse("2006-01-02 15:04:05", createdAt.String)
	r.UpdatedAt, _ = time.Parse("2006-01-02 15:04:05", updatedAt.String)
	return &r, nil
}

// ListSessions 列出会话
func (s *SQLiteStorage) ListSessions(limit int) ([]SessionRecord, error) {
	if limit <= 0 {
		limit = 100
	}
	rows, err := s.db.Query(`
		SELECT id, name, description, created_at, updated_at, traffic_count, vuln_count
		FROM sessions ORDER BY updated_at DESC LIMIT ?
	`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var records []SessionRecord
	for rows.Next() {
		var r SessionRecord
		var createdAt, updatedAt sql.NullString
		err := rows.Scan(
			&r.ID, &r.Name, &r.Description, &createdAt, &updatedAt,
			&r.TrafficCount, &r.VulnCount,
		)
		if err != nil {
			continue
		}
		r.CreatedAt, _ = time.Parse("2006-01-02 15:04:05", createdAt.String)
		r.UpdatedAt, _ = time.Parse("2006-01-02 15:04:05", updatedAt.String)
		records = append(records, r)
	}

	return records, nil
}

// DeleteSession 删除会话
func (s *SQLiteStorage) DeleteSession(id string) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	// 删除关联的漏洞
	_, err := s.db.Exec("DELETE FROM vulnerabilities WHERE session_id = ?", id)
	if err != nil {
		return err
	}

	// 删除关联的流量
	_, err = s.db.Exec("DELETE FROM traffic WHERE session_id = ?", id)
	if err != nil {
		return err
	}

	// 删除会话
	_, err = s.db.Exec("DELETE FROM sessions WHERE id = ?", id)
	return err
}

// UpdateSessionStats 更新会话统计
func (s *SQLiteStorage) UpdateSessionStats(id string) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	_, err := s.db.Exec(`
		UPDATE sessions SET
			traffic_count = (SELECT COUNT(*) FROM traffic WHERE session_id = ?),
			vuln_count = (SELECT COUNT(*) FROM vulnerabilities WHERE session_id = ?),
			updated_at = CURRENT_TIMESTAMP
		WHERE id = ?
	`, id, id, id)
	return err
}
