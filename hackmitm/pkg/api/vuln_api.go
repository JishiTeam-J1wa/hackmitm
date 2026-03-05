// Package api 提供 HTTP API 接口
// Package api provides HTTP API interfaces
package api

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"hackmitm/pkg/storage"
	"hackmitm/pkg/vuln"
)

// VulnAPI 漏洞 API 处理器
type VulnAPI struct {
	storage *storage.SQLiteStorage
}

// NewVulnAPI 创建漏洞 API
func NewVulnAPI(storage *storage.SQLiteStorage) *VulnAPI {
	return &VulnAPI{storage: storage}
}

// RegisterRoutes 注册路由
func (api *VulnAPI) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/vulns", api.ListVulnerabilities)
	mux.HandleFunc("GET /api/vulns/{id}", api.GetVulnerability)
	mux.HandleFunc("PATCH /api/vulns/{id}/status", api.UpdateVulnerabilityStatus)
	mux.HandleFunc("DELETE /api/vulns/{id}", api.DeleteVulnerability)
	mux.HandleFunc("GET /api/vulns/stats", api.GetVulnerabilityStats)
}

// ListVulnerabilities 列出漏洞
// GET /api/vulns?session_id=&severity=&status=&limit=100&offset=0
func (api *VulnAPI) ListVulnerabilities(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	sessionID := query.Get("session_id")
	severity := query.Get("severity")
	status := query.Get("status")
	limit, _ := strconv.Atoi(query.Get("limit"))
	offset, _ := strconv.Atoi(query.Get("offset"))

	if limit <= 0 {
		limit = 100
	}

	records, total, err := api.storage.ListVulnerabilities(sessionID, severity, status, limit, offset)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to list vulnerabilities: "+err.Error())
		return
	}

	// 转换为响应格式
	vulns := make([]map[string]interface{}, len(records))
	for i, r := range records {
		vulns[i] = map[string]interface{}{
			"id":          r.ID,
			"session_id":  r.SessionID,
			"traffic_id":  r.TrafficID,
			"rule_id":     r.RuleID,
			"name":        r.Name,
			"severity":    r.Severity,
			"confidence":  r.Confidence,
			"url":         r.URL,
			"parameter":   r.Parameter,
			"status":      r.Status,
			"occurrences": r.Occurrences,
			"first_seen":  r.FirstSeen,
			"last_seen":   r.LastSeen,
		}
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"data":   vulns,
		"total":  total,
		"limit":  limit,
		"offset": offset,
	})
}

// GetVulnerability 获取漏洞详情
// GET /api/vulns/{id}
func (api *VulnAPI) GetVulnerability(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		writeError(w, http.StatusBadRequest, "Missing vulnerability ID")
		return
	}

	record, err := api.storage.GetVulnerability(id)
	if err != nil {
		writeError(w, http.StatusNotFound, "Vulnerability not found")
		return
	}

	// 获取关联的流量
	var traffic map[string]interface{}
	if record.TrafficID > 0 {
		trafficRecord, err := api.storage.GetTrafficByID(record.TrafficID)
		if err == nil {
			traffic = map[string]interface{}{
				"id":           trafficRecord.ID,
				"method":       trafficRecord.Method,
				"url":          trafficRecord.URL,
				"status_code":  trafficRecord.StatusCode,
				"content_type": trafficRecord.ContentType,
				"timestamp":    trafficRecord.Timestamp,
			}
		}
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"id":              record.ID,
		"session_id":      record.SessionID,
		"traffic_id":      record.TrafficID,
		"rule_id":         record.RuleID,
		"hash":            record.Hash,
		"name":            record.Name,
		"description":     record.Description,
		"severity":        record.Severity,
		"confidence":      record.Confidence,
		"url":             record.URL,
		"parameter":       record.Parameter,
		"evidence":        record.Evidence,
		"remediation":     record.Remediation,
		"status":          record.Status,
		"occurrences":     record.Occurrences,
		"first_seen":      record.FirstSeen,
		"last_seen":       record.LastSeen,
		"timestamp":       record.Timestamp,
		"notes":           record.Notes,
		"request_method":  record.RequestMethod,
		"request_url":     record.RequestURL,
		"request_headers": record.RequestHeaders,
		"request_body":    record.RequestBody,
		"response_status": record.ResponseStatus,
		"response_headers": record.ResponseHeaders,
		"response_body":   record.ResponseBody,
		"traffic":         traffic,
	})
}

// UpdateVulnerabilityStatus 更新漏洞状态
// PATCH /api/vulns/{id}/status
func (api *VulnAPI) UpdateVulnerabilityStatus(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		writeError(w, http.StatusBadRequest, "Missing vulnerability ID")
		return
	}

	var req struct {
		Status string `json:"status"`
		Notes  string `json:"notes"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// 验证状态
	validStatuses := map[string]bool{
		"open":           true,
		"confirmed":      true,
		"false_positive": true,
		"fixed":          true,
	}
	if !validStatuses[req.Status] {
		writeError(w, http.StatusBadRequest, "Invalid status")
		return
	}

	if err := api.storage.UpdateVulnerabilityStatus(id, req.Status, req.Notes); err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to update vulnerability: "+err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"id":      id,
		"status":  req.Status,
		"notes":   req.Notes,
		"updated": time.Now(),
	})
}

// DeleteVulnerability 删除漏洞
// DELETE /api/vulns/{id}
func (api *VulnAPI) DeleteVulnerability(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		writeError(w, http.StatusBadRequest, "Missing vulnerability ID")
		return
	}

	if err := api.storage.DeleteVulnerability(id); err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to delete vulnerability: "+err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"id":      id,
		"deleted": true,
	})
}

// GetVulnerabilityStats 获取漏洞统计
// GET /api/vulns/stats?session_id=
func (api *VulnAPI) GetVulnerabilityStats(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	sessionID := query.Get("session_id")

	stats, err := api.storage.GetVulnerabilityStats(sessionID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to get stats: "+err.Error())
		return
	}

	writeJSON(w, http.StatusOK, stats)
}

// SessionAPI 会话 API 处理器
type SessionAPI struct {
	storage *storage.SQLiteStorage
}

// NewSessionAPI 创建会话 API
func NewSessionAPI(storage *storage.SQLiteStorage) *SessionAPI {
	return &SessionAPI{storage: storage}
}

// RegisterRoutes 注册路由
func (api *SessionAPI) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/sessions", api.ListSessions)
	mux.HandleFunc("POST /api/sessions", api.CreateSession)
	mux.HandleFunc("GET /api/sessions/{id}", api.GetSession)
	mux.HandleFunc("PUT /api/sessions/{id}", api.UpdateSession)
	mux.HandleFunc("DELETE /api/sessions/{id}", api.DeleteSession)
}

// ListSessions 列出会话
func (api *SessionAPI) ListSessions(w http.ResponseWriter, r *http.Request) {
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit <= 0 {
		limit = 100
	}

	sessions, err := api.storage.ListSessions(limit)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to list sessions: "+err.Error())
		return
	}

	// 转换为响应格式
	result := make([]map[string]interface{}, len(sessions))
	for i, s := range sessions {
		result[i] = map[string]interface{}{
			"id":            s.ID,
			"name":          s.Name,
			"description":   s.Description,
			"created_at":    s.CreatedAt,
			"updated_at":    s.UpdatedAt,
			"traffic_count": s.TrafficCount,
			"vuln_count":    s.VulnCount,
		}
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"data": result,
	})
}

// CreateSession 创建会话
func (api *SessionAPI) CreateSession(w http.ResponseWriter, r *http.Request) {
	var req struct {
		ID          string `json:"id"`
		Name        string `json:"name"`
		Description string `json:"description"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.ID == "" {
		req.ID = generateSessionID()
	}
	if req.Name == "" {
		req.Name = "Session " + req.ID[:8]
	}

	record := &storage.SessionRecord{
		ID:          req.ID,
		Name:        req.Name,
		Description: req.Description,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if err := api.storage.SaveSession(record); err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to create session: "+err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, map[string]interface{}{
		"id":          record.ID,
		"name":        record.Name,
		"description": record.Description,
		"created_at":  record.CreatedAt,
	})
}

// GetSession 获取会话
func (api *SessionAPI) GetSession(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		writeError(w, http.StatusBadRequest, "Missing session ID")
		return
	}

	session, err := api.storage.GetSession(id)
	if err != nil {
		writeError(w, http.StatusNotFound, "Session not found")
		return
	}

	// 获取统计数据
	stats, _ := api.storage.GetVulnerabilityStats(id)

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"id":            session.ID,
		"name":          session.Name,
		"description":   session.Description,
		"created_at":    session.CreatedAt,
		"updated_at":    session.UpdatedAt,
		"traffic_count": session.TrafficCount,
		"vuln_count":    session.VulnCount,
		"vuln_stats":    stats,
	})
}

// UpdateSession 更新会话
func (api *SessionAPI) UpdateSession(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		writeError(w, http.StatusBadRequest, "Missing session ID")
		return
	}

	var req struct {
		Name        string `json:"name"`
		Description string `json:"description"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	existing, err := api.storage.GetSession(id)
	if err != nil {
		writeError(w, http.StatusNotFound, "Session not found")
		return
	}

	if req.Name != "" {
		existing.Name = req.Name
	}
	if req.Description != "" {
		existing.Description = req.Description
	}
	existing.UpdatedAt = time.Now()

	if err := api.storage.SaveSession(existing); err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to update session: "+err.Error())
		return
	}

	writeJSON(w, http.StatusOK, existing)
}

// DeleteSession 删除会话
func (api *SessionAPI) DeleteSession(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		writeError(w, http.StatusBadRequest, "Missing session ID")
		return
	}

	if err := api.storage.DeleteSession(id); err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to delete session: "+err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"id":      id,
		"deleted": true,
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

func generateSessionID() string {
	return "sess_" + time.Now().Format("20060102150405")
}

// Convert functions for type conversion
func convertVulnToRecord(v *vuln.Vulnerability) *storage.VulnerabilityRecord {
	headersJSON := ""
	if v.Request != nil && v.Request.Headers != nil {
		if data, err := json.Marshal(v.Request.Headers); err == nil {
			headersJSON = string(data)
		}
	}

	return &storage.VulnerabilityRecord{
		ID:             v.ID,
		SessionID:      v.SessionID,
		TrafficID:      parseInt64(v.TrafficID),
		RuleID:         v.RuleID,
		Hash:           v.Hash,
		Name:           v.Name,
		Description:    v.Description,
		Severity:       string(v.Severity),
		Confidence:     v.Confidence,
		URL:            v.URL,
		Parameter:      v.Parameter,
		Evidence:       v.Evidence,
		Remediation:    v.Remediation,
		Status:         string(v.Status),
		Occurrences:    v.Occurrences,
		FirstSeen:      v.FirstSeen,
		LastSeen:       v.LastSeen,
		Timestamp:      v.Timestamp,
		RequestMethod:  func() string { if v.Request != nil { return v.Request.Method }; return "" }(),
		RequestURL:     func() string { if v.Request != nil { return v.Request.URL }; return "" }(),
		RequestHeaders: headersJSON,
		RequestBody:    func() string { if v.Request != nil { return v.Request.Body }; return "" }(),
	}
}

func parseInt64(s string) int64 {
	var result int64
	for _, c := range s {
		if c >= '0' && c <= '9' {
			result = result*10 + int64(c-'0')
		}
	}
	return result
}
