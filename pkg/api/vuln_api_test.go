package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"hackmitm/pkg/storage"
)

// setupTestAPI creates a test API with an in-memory database
func setupTestAPI(t *testing.T) (*VulnAPI, *SessionAPI, *storage.SQLiteStorage, func()) {
	// Create a temporary directory for the test database
	tmpDir, err := os.MkdirTemp("", "hackmitm-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}

	// Create SQLite storage
	store, err := storage.NewSQLiteStorage(tmpDir)
	if err != nil {
		os.RemoveAll(tmpDir)
		t.Fatalf("Failed to create storage: %v", err)
	}

	cleanup := func() {
		store.Close()
		os.RemoveAll(tmpDir)
	}

	return NewVulnAPI(store), NewSessionAPI(store), store, cleanup
}

// TestVulnAPI_ListVulnerabilities tests listing vulnerabilities
func TestVulnAPI_ListVulnerabilities(t *testing.T) {
	vulnAPI, _, _, cleanup := setupTestAPI(t)
	defer cleanup()

	// Create a test vulnerability
	testVuln := &storage.VulnerabilityRecord{
		ID:          "vuln-test-1",
		SessionID:   "sess-test",
		Name:        "SQL Injection",
		Severity:    "high",
		URL:         "https://example.com/api/users",
		Parameter:   "id",
		Status:      "open",
		Confidence:  0.9,
		Timestamp:   time.Now(),
		FirstSeen:   time.Now(),
		LastSeen:    time.Now(),
	}
	err := vulnAPI.storage.SaveVulnerability(testVuln)
	if err != nil {
		t.Fatalf("Failed to save test vulnerability: %v", err)
	}

	// Create request
	req := httptest.NewRequest("GET", "/api/vulns?limit=10", nil)
	w := httptest.NewRecorder()

	// Call handler
	vulnAPI.ListVulnerabilities(w, req)

	// Check response
	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var response map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	data, ok := response["data"].([]interface{})
	if !ok {
		t.Fatal("Expected data to be an array")
	}

	if len(data) != 1 {
		t.Errorf("Expected 1 vulnerability, got %d", len(data))
	}
}

// TestVulnAPI_GetVulnerability tests getting a single vulnerability
func TestVulnAPI_GetVulnerability(t *testing.T) {
	vulnAPI, _, _, cleanup := setupTestAPI(t)
	defer cleanup()

	// Create a test vulnerability
	testVuln := &storage.VulnerabilityRecord{
		ID:          "vuln-test-2",
		SessionID:   "sess-test",
		Name:        "Cross-Site Scripting",
		Severity:    "medium",
		URL:         "https://example.com/search",
		Parameter:   "q",
		Status:      "open",
		Confidence:  0.85,
		Evidence:    "<script>alert(1)</script>",
		Timestamp:   time.Now(),
		FirstSeen:   time.Now(),
		LastSeen:    time.Now(),
	}
	err := vulnAPI.storage.SaveVulnerability(testVuln)
	if err != nil {
		t.Fatalf("Failed to save test vulnerability: %v", err)
	}

	// Create request with path value
	req := httptest.NewRequest("GET", "/api/vulns/vuln-test-2", nil)
	req.SetPathValue("id", "vuln-test-2")
	w := httptest.NewRecorder()

	// Call handler
	vulnAPI.GetVulnerability(w, req)

	// Check response
	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var response map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	if response["id"] != "vuln-test-2" {
		t.Errorf("Expected ID 'vuln-test-2', got '%v'", response["id"])
	}

	if response["name"] != "Cross-Site Scripting" {
		t.Errorf("Expected name 'Cross-Site Scripting', got '%v'", response["name"])
	}
}

// TestVulnAPI_GetVulnerability_NotFound tests getting a non-existent vulnerability
func TestVulnAPI_GetVulnerability_NotFound(t *testing.T) {
	vulnAPI, _, _, cleanup := setupTestAPI(t)
	defer cleanup()

	req := httptest.NewRequest("GET", "/api/vulns/non-existent", nil)
	req.SetPathValue("id", "non-existent")
	w := httptest.NewRecorder()

	vulnAPI.GetVulnerability(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("Expected status 404, got %d", w.Code)
	}
}

// TestVulnAPI_UpdateVulnerabilityStatus tests updating vulnerability status
func TestVulnAPI_UpdateVulnerabilityStatus(t *testing.T) {
	vulnAPI, _, _, cleanup := setupTestAPI(t)
	defer cleanup()

	// Create a test vulnerability
	testVuln := &storage.VulnerabilityRecord{
		ID:          "vuln-test-3",
		SessionID:   "sess-test",
		Name:        "Path Traversal",
		Severity:    "high",
		URL:         "https://example.com/files",
		Parameter:   "path",
		Status:      "open",
		Confidence:  0.95,
		Timestamp:   time.Now(),
		FirstSeen:   time.Now(),
		LastSeen:    time.Now(),
	}
	err := vulnAPI.storage.SaveVulnerability(testVuln)
	if err != nil {
		t.Fatalf("Failed to save test vulnerability: %v", err)
	}

	// Create update request
	updateBody := map[string]string{
		"status": "confirmed",
		"notes":  "Verified by security team",
	}
	bodyBytes, _ := json.Marshal(updateBody)

	req := httptest.NewRequest("PATCH", "/api/vulns/vuln-test-3/status", bytes.NewReader(bodyBytes))
	req.SetPathValue("id", "vuln-test-3")
	w := httptest.NewRecorder()

	vulnAPI.UpdateVulnerabilityStatus(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var response map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	if response["status"] != "confirmed" {
		t.Errorf("Expected status 'confirmed', got '%v'", response["status"])
	}
}

// TestVulnAPI_UpdateVulnerabilityStatus_InvalidStatus tests updating with invalid status
func TestVulnAPI_UpdateVulnerabilityStatus_InvalidStatus(t *testing.T) {
	vulnAPI, _, _, cleanup := setupTestAPI(t)
	defer cleanup()

	// Create a test vulnerability
	testVuln := &storage.VulnerabilityRecord{
		ID:        "vuln-test-4",
		SessionID: "sess-test",
		Name:      "Test Vuln",
		Severity:  "low",
		Status:    "open",
		Timestamp: time.Now(),
	}
	_ = vulnAPI.storage.SaveVulnerability(testVuln)

	// Create update request with invalid status
	updateBody := map[string]string{
		"status": "invalid_status",
	}
	bodyBytes, _ := json.Marshal(updateBody)

	req := httptest.NewRequest("PATCH", "/api/vulns/vuln-test-4/status", bytes.NewReader(bodyBytes))
	req.SetPathValue("id", "vuln-test-4")
	w := httptest.NewRecorder()

	vulnAPI.UpdateVulnerabilityStatus(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", w.Code)
	}
}

// TestVulnAPI_DeleteVulnerability tests deleting a vulnerability
func TestVulnAPI_DeleteVulnerability(t *testing.T) {
	vulnAPI, _, _, cleanup := setupTestAPI(t)
	defer cleanup()

	// Create a test vulnerability
	testVuln := &storage.VulnerabilityRecord{
		ID:        "vuln-test-5",
		SessionID: "sess-test",
		Name:      "Test Vuln to Delete",
		Severity:  "low",
		Status:    "open",
		Timestamp: time.Now(),
	}
	err := vulnAPI.storage.SaveVulnerability(testVuln)
	if err != nil {
		t.Fatalf("Failed to save test vulnerability: %v", err)
	}

	req := httptest.NewRequest("DELETE", "/api/vulns/vuln-test-5", nil)
	req.SetPathValue("id", "vuln-test-5")
	w := httptest.NewRecorder()

	vulnAPI.DeleteVulnerability(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	// Verify it's deleted
	req2 := httptest.NewRequest("GET", "/api/vulns/vuln-test-5", nil)
	req2.SetPathValue("id", "vuln-test-5")
	w2 := httptest.NewRecorder()
	vulnAPI.GetVulnerability(w2, req2)

	if w2.Code != http.StatusNotFound {
		t.Errorf("Expected 404 after deletion, got %d", w2.Code)
	}
}

// TestVulnAPI_GetVulnerabilityStats tests getting vulnerability statistics
func TestVulnAPI_GetVulnerabilityStats(t *testing.T) {
	vulnAPI, _, _, cleanup := setupTestAPI(t)
	defer cleanup()

	// Create test vulnerabilities with different severities and unique hashes
	vulns := []*storage.VulnerabilityRecord{
		{ID: "vuln-1", SessionID: "sess-1", Name: "Vuln 1", Severity: "critical", Status: "open", Timestamp: time.Now(), Hash: "hash-1"},
		{ID: "vuln-2", SessionID: "sess-1", Name: "Vuln 2", Severity: "high", Status: "open", Timestamp: time.Now(), Hash: "hash-2"},
		{ID: "vuln-3", SessionID: "sess-1", Name: "Vuln 3", Severity: "high", Status: "confirmed", Timestamp: time.Now(), Hash: "hash-3"},
		{ID: "vuln-4", SessionID: "sess-1", Name: "Vuln 4", Severity: "medium", Status: "open", Timestamp: time.Now(), Hash: "hash-4"},
		{ID: "vuln-5", SessionID: "sess-1", Name: "Vuln 5", Severity: "low", Status: "fixed", Timestamp: time.Now(), Hash: "hash-5"},
	}

	for _, v := range vulns {
		if err := vulnAPI.storage.SaveVulnerability(v); err != nil {
			t.Fatalf("Failed to save vulnerability %s: %v", v.ID, err)
		}
	}

	req := httptest.NewRequest("GET", "/api/vulns/stats?session_id=sess-1", nil)
	w := httptest.NewRecorder()

	vulnAPI.GetVulnerabilityStats(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var response map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	// Check total exists
	total := int(response["total"].(float64))
	if total != 5 {
		t.Errorf("Expected total 5, got %d", total)
	}

	// Verify stats
	critical := int(response["critical"].(float64))
	high := int(response["high"].(float64))

	if critical != 1 {
		t.Errorf("Expected 1 critical, got %d", critical)
	}
	if high != 2 {
		t.Errorf("Expected 2 high, got %d", high)
	}
}

// TestSessionAPI_ListSessions tests listing sessions
func TestSessionAPI_ListSessions(t *testing.T) {
	_, sessionAPI, _, cleanup := setupTestAPI(t)
	defer cleanup()

	// Create test sessions
	sessions := []*storage.SessionRecord{
		{ID: "sess-1", Name: "Session 1", CreatedAt: time.Now(), UpdatedAt: time.Now()},
		{ID: "sess-2", Name: "Session 2", CreatedAt: time.Now(), UpdatedAt: time.Now()},
	}

	for _, s := range sessions {
		_ = sessionAPI.storage.SaveSession(s)
	}

	req := httptest.NewRequest("GET", "/api/sessions?limit=10", nil)
	w := httptest.NewRecorder()

	sessionAPI.ListSessions(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var response map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	data, ok := response["data"].([]interface{})
	if !ok {
		t.Fatal("Expected data to be an array")
	}

	if len(data) != 2 {
		t.Errorf("Expected 2 sessions, got %d", len(data))
	}
}

// TestSessionAPI_CreateSession tests creating a session
func TestSessionAPI_CreateSession(t *testing.T) {
	_, sessionAPI, _, cleanup := setupTestAPI(t)
	defer cleanup()

	createBody := map[string]string{
		"name":        "New Test Session",
		"description": "Test session description",
	}
	bodyBytes, _ := json.Marshal(createBody)

	req := httptest.NewRequest("POST", "/api/sessions", bytes.NewReader(bodyBytes))
	w := httptest.NewRecorder()

	sessionAPI.CreateSession(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("Expected status 201, got %d", w.Code)
	}

	var response map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	if response["name"] != "New Test Session" {
		t.Errorf("Expected name 'New Test Session', got '%v'", response["name"])
	}
}

// TestSessionAPI_GetSession tests getting a session
func TestSessionAPI_GetSession(t *testing.T) {
	_, sessionAPI, _, cleanup := setupTestAPI(t)
	defer cleanup()

	// Create a test session
	testSession := &storage.SessionRecord{
		ID:          "sess-get-test",
		Name:        "Session to Get",
		Description: "Test description",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	_ = sessionAPI.storage.SaveSession(testSession)

	req := httptest.NewRequest("GET", "/api/sessions/sess-get-test", nil)
	req.SetPathValue("id", "sess-get-test")
	w := httptest.NewRecorder()

	sessionAPI.GetSession(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var response map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	if response["name"] != "Session to Get" {
		t.Errorf("Expected name 'Session to Get', got '%v'", response["name"])
	}
}

// TestSessionAPI_UpdateSession tests updating a session
func TestSessionAPI_UpdateSession(t *testing.T) {
	_, sessionAPI, _, cleanup := setupTestAPI(t)
	defer cleanup()

	// Create a test session
	testSession := &storage.SessionRecord{
		ID:        "sess-update-test",
		Name:      "Original Name",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	_ = sessionAPI.storage.SaveSession(testSession)

	updateBody := map[string]string{
		"name":        "Updated Name",
		"description": "Updated description",
	}
	bodyBytes, _ := json.Marshal(updateBody)

	req := httptest.NewRequest("PUT", "/api/sessions/sess-update-test", bytes.NewReader(bodyBytes))
	req.SetPathValue("id", "sess-update-test")
	w := httptest.NewRecorder()

	sessionAPI.UpdateSession(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
}

// TestSessionAPI_DeleteSession tests deleting a session
func TestSessionAPI_DeleteSession(t *testing.T) {
	_, sessionAPI, _, cleanup := setupTestAPI(t)
	defer cleanup()

	// Create a test session
	testSession := &storage.SessionRecord{
		ID:        "sess-delete-test",
		Name:      "Session to Delete",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	_ = sessionAPI.storage.SaveSession(testSession)

	req := httptest.NewRequest("DELETE", "/api/sessions/sess-delete-test", nil)
	req.SetPathValue("id", "sess-delete-test")
	w := httptest.NewRecorder()

	sessionAPI.DeleteSession(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	// Verify it's deleted
	req2 := httptest.NewRequest("GET", "/api/sessions/sess-delete-test", nil)
	req2.SetPathValue("id", "sess-delete-test")
	w2 := httptest.NewRecorder()
	sessionAPI.GetSession(w2, req2)

	if w2.Code != http.StatusNotFound {
		t.Errorf("Expected 404 after deletion, got %d", w2.Code)
	}
}
