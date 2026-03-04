package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/wailsapp/wails/v2/pkg/runtime"

	"hackmitm-desktop/internal/api"
	"hackmitm-desktop/internal/intruder"
	"hackmitm-desktop/internal/models"
	"hackmitm-desktop/internal/scanner"
	"hackmitm-desktop/internal/service"
)

// ConnectionMode represents the current connection mode
type ConnectionMode string

const (
	ModeLocal  ConnectionMode = "local"
	ModeRemote ConnectionMode = "remote"
)

// App struct holds all API instances
type App struct {
	ctx             context.Context
	trafficAPI      *api.TrafficAPI
	fingerprintAPI  *api.FingerprintAPI
	proxyAPI        *api.ProxyAPI
	dashboardAPI    *api.DashboardAPI
	repeaterAPI     *api.RepeaterAPI
	initAPI         *api.InitAPI
	proxyConfigAPI  *api.ProxyConfigAPI
	vulnAPI         *api.VulnAPI
	scannerAPI      *api.ScannerAPI
	intruderAPI     *api.IntruderAPI
	activeScanAPI   *api.ActiveScanAPI
	reportAPI       *api.ReportAPI
	localService    *service.LocalService
	configManager   *service.ConfigManager
	apiEndpoint     string
	connected       bool
	connectionMode  ConnectionMode
	pollingInterval time.Duration
	cancelPolling   context.CancelFunc
}

// NewApp creates a new App application struct
func NewApp() *App {
	return &App{
		trafficAPI:      api.NewTrafficAPI(),
		fingerprintAPI:  api.NewFingerprintAPI(),
		proxyAPI:        api.NewProxyAPI(),
		dashboardAPI:    api.NewDashboardAPI(),
		repeaterAPI:     api.NewRepeaterAPI(),
		initAPI:         api.NewInitAPI(),
		proxyConfigAPI:  api.NewProxyConfigAPI(),
		vulnAPI:         api.NewVulnAPI(),
		scannerAPI:      api.NewScannerAPI(),
		intruderAPI:     api.NewIntruderAPI(),
		activeScanAPI:   api.NewActiveScanAPI(),
		reportAPI:       api.NewReportAPI(),
		localService:    service.NewLocalService(),
		configManager:   service.NewConfigManager(),
		apiEndpoint:     "http://localhost:9090",
		pollingInterval: 2 * time.Second,
	}
}

// startup is called when the app starts
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx

	// Set context for all APIs
	a.trafficAPI.SetContext(ctx)
	a.fingerprintAPI.SetContext(ctx)
	a.proxyAPI.SetContext(ctx)
	a.dashboardAPI.SetContext(ctx)
	a.repeaterAPI.SetContext(ctx)
	a.initAPI.SetContext(ctx)
	a.vulnAPI.SetContext(ctx)
	a.intruderAPI.SetContext(ctx)
	a.activeScanAPI.SetContext(ctx)
	a.localService.SetContext(ctx)

	// Set default API endpoint
	a.SetAPIEndpoint(a.apiEndpoint)
}

// shutdown is called when the app is closing
func (a *App) shutdown(ctx context.Context) {
	// Stop polling
	if a.cancelPolling != nil {
		a.cancelPolling()
	}

	// Stop local service if running
	if a.localService != nil {
		a.localService.Stop()
	}

	// Close database connection
	if a.initAPI != nil {
		a.initAPI.Close()
	}
}

// ============ Connection Management ============

// SetAPIEndpoint sets the API endpoint for all APIs
func (a *App) SetAPIEndpoint(endpoint string) {
	a.apiEndpoint = endpoint
	a.trafficAPI.SetAPIEndpoint(endpoint)
	a.fingerprintAPI.SetAPIEndpoint(endpoint)
	a.proxyAPI.SetAPIEndpoint(endpoint)
	a.dashboardAPI.SetAPIEndpoint(endpoint)
	a.repeaterAPI.SetAPIEndpoint(endpoint)
	a.scannerAPI.SetAPIEndpoint(endpoint)
	a.reportAPI.SetAPIEndpoint(endpoint)
}

// GetAPIEndpoint returns the current API endpoint
func (a *App) GetAPIEndpoint() string {
	return a.apiEndpoint
}

// Connect attempts to connect to the HackMITM server
func (a *App) Connect() error {
	health, err := a.proxyAPI.HealthCheck()
	if err != nil {
		a.connected = false
		runtime.EventsEmit(a.ctx, "connection:status", map[string]interface{}{
			"connected": false,
			"error":     err.Error(),
		})
		return fmt.Errorf("failed to connect to server: %w", err)
	}

	a.connected = true
	runtime.EventsEmit(a.ctx, "connection:status", map[string]interface{}{
		"connected": true,
		"health":    health,
	})

	// Start polling for updates
	go a.startPolling()

	return nil
}

// Disconnect disconnects from the server
func (a *App) Disconnect() {
	a.connected = false
	if a.cancelPolling != nil {
		a.cancelPolling()
	}
	runtime.EventsEmit(a.ctx, "connection:status", map[string]interface{}{
		"connected": false,
	})
}

// IsConnected returns the connection status
func (a *App) IsConnected() bool {
	return a.connected
}

// ============ Proxy Operations ============

// GetProxyStatus fetches the proxy server status
func (a *App) GetProxyStatus() (*models.ProxyStatus, error) {
	return a.proxyAPI.GetStatus()
}

// SetInterceptMode enables or disables intercept mode
func (a *App) SetInterceptMode(enabled bool) error {
	return a.proxyAPI.SetInterceptMode(enabled)
}

// ForwardIntercepted forwards an intercepted request
func (a *App) ForwardIntercepted(requestID string) error {
	return a.proxyAPI.ForwardIntercepted(requestID)
}

// DropIntercepted drops an intercepted request
func (a *App) DropIntercepted(requestID string) error {
	return a.proxyAPI.DropIntercepted(requestID)
}

// ============ Traffic Operations ============

// GetTraffic fetches traffic history
func (a *App) GetTraffic(limit int) ([]models.TrafficItem, error) {
	return a.trafficAPI.GetTraffic(limit)
}

// ClearTraffic clears all traffic history
func (a *App) ClearTraffic() error {
	return a.trafficAPI.ClearTraffic()
}

// ============ Fingerprint Operations ============

// GetFingerprintStats fetches fingerprint statistics
func (a *App) GetFingerprintStats() (map[string]any, error) {
	return a.fingerprintAPI.GetFingerprintStats()
}

// IdentifyFingerprint manually triggers fingerprint identification
func (a *App) IdentifyFingerprint(url string) (*models.FingerprintResult, error) {
	return a.fingerprintAPI.IdentifyFingerprint(url)
}

// GetFingerprintHistory fetches fingerprint identification history
func (a *App) GetFingerprintHistory(limit int) ([]models.FingerprintResult, error) {
	return a.fingerprintAPI.GetFingerprintHistory(limit)
}

// ============ Dashboard Operations ============

// GetMetrics fetches dashboard metrics
func (a *App) GetMetrics() (*models.DashboardMetrics, error) {
	return a.dashboardAPI.GetMetrics()
}

// GetTrafficPatterns fetches traffic pattern statistics
func (a *App) GetTrafficPatterns() (map[string]any, error) {
	return a.dashboardAPI.GetTrafficPatterns()
}

// ============ Repeater Operations ============

// SendRequest sends an HTTP request and returns the response
func (a *App) SendRequest(req models.RepeaterRequest) (*models.RepeaterResponse, error) {
	return a.repeaterAPI.SendRequest(req)
}

// ============ Polling ============

// startPolling starts polling the server for updates
func (a *App) startPolling() {
	ctx, cancel := context.WithCancel(a.ctx)
	a.cancelPolling = cancel

	ticker := time.NewTicker(a.pollingInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if !a.connected {
				continue
			}

			// Fetch and emit metrics
			metrics, err := a.dashboardAPI.GetMetrics()
			if err == nil {
				runtime.EventsEmit(a.ctx, "dashboard:metrics", metrics)
			}

			// Fetch and emit status
			status, err := a.proxyAPI.GetStatus()
			if err == nil {
				runtime.EventsEmit(a.ctx, "proxy:status", status)
			}
		}
	}
}

// SetPollingInterval sets the polling interval
func (a *App) SetPollingInterval(interval int) {
	a.pollingInterval = time.Duration(interval) * time.Second
}

// ============ Initialization Operations ============

// InitDatabase initializes the local database
func (a *App) InitDatabase(config api.DatabaseConfig) (*api.InitResult, error) {
	result, err := a.initAPI.InitDatabase(config)
	if err != nil {
		return nil, err
	}

	// Emit event on successful initialization
	if result.Success {
		runtime.EventsEmit(a.ctx, "database:initialized", result)
	}

	return result, nil
}

// GetDatabaseInfo returns information about the current database
func (a *App) GetDatabaseInfo() (map[string]interface{}, error) {
	return a.initAPI.GetDatabaseInfo()
}

// SelectDatabaseFolder opens a folder selection dialog
func (a *App) SelectDatabaseFolder() (string, error) {
	// Use Wails runtime to open folder dialog
	folder, err := runtime.OpenDirectoryDialog(a.ctx, runtime.OpenDialogOptions{
		Title: "Select Database Folder",
	})
	if err != nil {
		return a.initAPI.SelectDatabaseFolder()
	}
	return folder, nil
}

// GetDictionaryEntries retrieves dictionary entries
func (a *App) GetDictionaryEntries(category, dictType string, limit int) ([]api.DictEntry, error) {
	return a.initAPI.GetDictionaryEntries(category, dictType, limit)
}

// AddDictionaryEntry adds a new dictionary entry
func (a *App) AddDictionaryEntry(entry api.DictEntry) error {
	return a.initAPI.AddDictionaryEntry(entry)
}

// ImportDictionary imports dictionary entries from a file
func (a *App) ImportDictionary() (int, error) {
	file, err := runtime.OpenFileDialog(a.ctx, runtime.OpenDialogOptions{
		Title: "Import Dictionary",
		Filters: []runtime.FileFilter{
			{DisplayName: "JSON Files (*.json)", Pattern: "*.json"},
			{DisplayName: "All Files (*.*)", Pattern: "*.*"},
		},
	})
	if err != nil || file == "" {
		return 0, fmt.Errorf("no file selected")
	}
	return a.initAPI.ImportDictionary(file)
}

// ExportDictionary exports dictionary entries to a file
func (a *App) ExportDictionary(category string) error {
	file, err := runtime.SaveFileDialog(a.ctx, runtime.SaveDialogOptions{
		Title: "Export Dictionary",
		DefaultFilename: "dictionary.json",
		Filters: []runtime.FileFilter{
			{DisplayName: "JSON Files (*.json)", Pattern: "*.json"},
		},
	})
	if err != nil || file == "" {
		return fmt.Errorf("no file selected")
	}
	return a.initAPI.ExportDictionary(file, category)
}

// ============ Proxy Configuration Operations ============

// GetNetworkInterfaces returns all network interfaces with their IP addresses
func (a *App) GetNetworkInterfaces() ([]api.NetworkInterface, error) {
	return a.proxyConfigAPI.GetNetworkInterfaces()
}

// GetBindAddressOptions returns available bind address options
func (a *App) GetBindAddressOptions() ([]map[string]string, error) {
	return a.proxyConfigAPI.GetBindAddressOptions()
}

// GetProxyConfig returns the current proxy configuration
func (a *App) GetProxyConfig() api.ProxyConfig {
	return a.proxyConfigAPI.GetConfig()
}

// SaveProxyConfig saves the proxy configuration
func (a *App) SaveProxyConfig(config api.ProxyConfig) map[string]string {
	errors := a.proxyConfigAPI.ValidateConfig(config)
	if len(errors) > 0 {
		return errors
	}

	if err := a.proxyConfigAPI.SaveConfig(config); err != nil {
		errors["general"] = err.Error()
		return errors
	}

	return nil
}

// ============ Local/Remote Mode Operations ============

// StartLocalMode starts the local embedded HackMITM service
func (a *App) StartLocalMode(config service.LocalConfig) error {
	a.connectionMode = ModeLocal

	// Start the local service
	if err := a.localService.Start(&config); err != nil {
		return err
	}

	// Set API endpoint to local
	endpoint := fmt.Sprintf("http://localhost:%d", config.ApiPort)
	a.SetAPIEndpoint(endpoint)

	// Connect to the local service
	if err := a.Connect(); err != nil {
		a.localService.Stop()
		return err
	}

	runtime.EventsEmit(a.ctx, "mode:changed", map[string]string{
		"mode": string(ModeLocal),
	})

	return nil
}

// ConnectRemoteMode connects to a remote HackMITM server
func (a *App) ConnectRemoteMode(config service.RemoteConfig) error {
	a.connectionMode = ModeRemote

	// Set API endpoint
	endpoint := fmt.Sprintf("http://%s:%d", config.Host, config.Port)
	a.SetAPIEndpoint(endpoint)

	// Connect to remote
	if err := a.Connect(); err != nil {
		return err
	}

	runtime.EventsEmit(a.ctx, "mode:changed", map[string]string{
		"mode": string(ModeRemote),
	})

	return nil
}

// GetConnectionMode returns the current connection mode
func (a *App) GetConnectionMode() string {
	return string(a.connectionMode)
}

// StopLocalService stops the local service if running
func (a *App) StopLocalService() error {
	if a.localService != nil && a.localService.IsRunning() {
		return a.localService.Stop()
	}
	return nil
}

// IsLocalServiceRunning returns whether the local service is running
func (a *App) IsLocalServiceRunning() bool {
	if a.localService == nil {
		return false
	}
	return a.localService.IsRunning()
}

// GetLocalServiceOutput returns the recent output from the local service
func (a *App) GetLocalServiceOutput() []string {
	if a.localService == nil {
		return []string{}
	}
	return a.localService.GetOutput()
}

// ============ Request Modification ============

// requestModifier handles request modification
var requestModifier = service.NewRequestModifier()

// ModifyAndForwardRequest modifies an intercepted request and forwards it
func (a *App) ModifyAndForwardRequest(requestID string, modifications map[string]interface{}) (*service.ModifiedRequestResult, error) {
	// Convert modifications to RequestModification struct
	mod := service.RequestModification{}

	if method, ok := modifications["method"].(string); ok {
		mod.Method = method
	}
	if urlStr, ok := modifications["url"].(string); ok {
		mod.URL = urlStr
	}
	if headers, ok := modifications["headers"].(map[string]interface{}); ok {
		mod.Headers = make(map[string]string)
		for k, v := range headers {
			if vs, ok := v.(string); ok {
				mod.Headers[k] = vs
			}
		}
	}
	if body, ok := modifications["body"].(string); ok {
		mod.Body = body
	}

	// Set API endpoint
	requestModifier.SetAPIEndpoint(a.apiEndpoint)

	// Try to use the proxy API first
	result, err := requestModifier.SendModifiedIntercepted(requestID, mod)
	if err != nil {
		// If proxy API fails, try direct forwarding
		originalURL, _ := modifications["originalUrl"].(string)
		if originalURL != "" {
			return requestModifier.ModifyAndForward(originalURL, mod)
		}
		return nil, err
	}

	return result, nil
}

// TestConnection tests the connection to the proxy server
func (a *App) TestConnection(endpoint string) map[string]interface{} {
	result := map[string]interface{}{
		"success": false,
		"message": "",
		"latency": 0,
	}

	// Set temporary endpoint
	originalEndpoint := a.apiEndpoint
	a.SetAPIEndpoint(endpoint)
	defer func() { a.SetAPIEndpoint(originalEndpoint) }()

	startTime := time.Now()

	// Try to connect
	err := a.Connect()
	if err != nil {
		result["message"] = err.Error()
		return result
	}

	result["latency"] = time.Since(startTime).Milliseconds()

	// Get proxy status to verify connection
	status, err := a.GetProxyStatus()
	if err != nil {
		result["message"] = "连接成功，但无法获取状态"
		result["success"] = true
		return result
	}

	// Disconnect test connection
	a.Disconnect()

	result["success"] = true
	result["message"] = fmt.Sprintf("连接成功 - 代理端口: %d, 活跃连接: %d",
		status.Port, status.ActiveConnections)

	return result
}

// ============ Configuration Management ============

// GetAppConfig returns the current application configuration
func (a *App) GetAppConfig() *service.AppConfig {
	return a.configManager.Get()
}

// SaveAppConfig saves the application configuration
func (a *App) SaveAppConfig(config *service.AppConfig) error {
	return a.configManager.Set(config)
}

// UpdateAppConfig updates specific fields of the configuration
func (a *App) UpdateAppConfig(updates map[string]any) error {
	return a.configManager.Update(updates)
}

// ResetAppConfig resets the configuration to defaults
func (a *App) ResetAppConfig() error {
	return a.configManager.Reset()
}

// GetConfigPath returns the path to the configuration file
func (a *App) GetConfigPath() string {
	return a.configManager.GetConfigPath()
}

// ============ Vulnerability Operations ============

// GetVulnerabilities retrieves vulnerabilities from database
func (a *App) GetVulnerabilities(severity, status, vulnType string, limit int) ([]api.Vulnerability, error) {
	return a.vulnAPI.GetVulnerabilities(severity, status, vulnType, limit)
}

// AddVulnerability adds a new vulnerability to database
func (a *App) AddVulnerability(v api.Vulnerability) (int64, error) {
	return a.vulnAPI.AddVulnerability(v)
}

// UpdateVulnerabilityStatus updates the status of a vulnerability
func (a *App) UpdateVulnerabilityStatus(id int64, status string) error {
	return a.vulnAPI.UpdateVulnerabilityStatus(id, status)
}

// DeleteVulnerability deletes a vulnerability
func (a *App) DeleteVulnerability(id int64) error {
	return a.vulnAPI.DeleteVulnerability(id)
}

// GetVulnStats returns vulnerability statistics
func (a *App) GetVulnStats() (map[string]int64, error) {
	return a.vulnAPI.GetVulnStats()
}

// ============ Scan Result Operations ============

// GetScanResults retrieves scan results from database
func (a *App) GetScanResults(severity, pluginID, falsePositive string, limit int) ([]api.ScanResult, error) {
	return a.vulnAPI.GetScanResults(severity, pluginID, falsePositive, limit)
}

// AddScanResult adds a new scan result to database
func (a *App) AddScanResult(r api.ScanResult) (int64, error) {
	return a.vulnAPI.AddScanResult(r)
}

// MarkScanResultFalsePositive marks a scan result as false positive
func (a *App) MarkScanResultFalsePositive(id int64, isFP bool) error {
	return a.vulnAPI.MarkScanResultFalsePositive(id, isFP)
}

// DeleteScanResult deletes a scan result
func (a *App) DeleteScanResult(id int64) error {
	return a.vulnAPI.DeleteScanResult(id)
}

// ClearScanResults clears all scan results
func (a *App) ClearScanResults() error {
	return a.vulnAPI.ClearScanResults()
}

// ExportScanResultToVuln exports a scan result to vulnerabilities table
func (a *App) ExportScanResultToVuln(scanID int64) (int64, error) {
	return a.vulnAPI.ExportScanResultToVuln(scanID)
}

// ============ WebSocket Operations ============

// GetWebSocketMessages retrieves WebSocket messages from database
func (a *App) GetWebSocketMessages(direction, msgType, connectionID string, limit int) ([]api.WebSocketMessage, error) {
	return a.vulnAPI.GetWebSocketMessages(direction, msgType, connectionID, limit)
}

// AddWebSocketMessage adds a WebSocket message to database
func (a *App) AddWebSocketMessage(m api.WebSocketMessage) (int64, error) {
	return a.vulnAPI.AddWebSocketMessage(m)
}

// ClearWebSocketMessages clears all WebSocket messages
func (a *App) ClearWebSocketMessages() error {
	return a.vulnAPI.ClearWebSocketMessages()
}

// ============ Database Connection ============

// SetDatabaseForAPIs sets the database connection for all APIs that need it
func (a *App) SetDatabaseForAPIs() {
	if a.initAPI != nil {
		// The initAPI has the DB connection, pass it to APIs that need it
		// This is called after InitDatabase succeeds
		a.vulnAPI.SetDB(a.initAPI.GetDB())
		a.intruderAPI.SetDB(a.initAPI.GetDB())
		a.activeScanAPI.SetDB(a.initAPI.GetDB())
	}
}

// ============ Scanner Operations ============

// GetScannerRules fetches all scanner rules
func (a *App) GetScannerRules() ([]api.Rule, error) {
	return a.scannerAPI.GetRules()
}

// GetScannerRule fetches a single rule by ID
func (a *App) GetScannerRule(id string) (*api.Rule, error) {
	return a.scannerAPI.GetRule(id)
}

// EnableScannerRule enables a scanner rule
func (a *App) EnableScannerRule(id string) error {
	return a.scannerAPI.EnableRule(id)
}

// DisableScannerRule disables a scanner rule
func (a *App) DisableScannerRule(id string) error {
	return a.scannerAPI.DisableRule(id)
}

// CreateScannerRule creates a new custom scanner rule
func (a *App) CreateScannerRule(rule api.Rule) error {
	return a.scannerAPI.CreateRule(rule)
}

// ReloadScannerRules reloads all rules from disk
func (a *App) ReloadScannerRules() (int, error) {
	return a.scannerAPI.ReloadRules()
}

// ============ Report Operations ============

// GenerateReport generates a report with the given options
func (a *App) GenerateReport(options api.ReportOptions) ([]byte, string, error) {
	return a.reportAPI.GenerateReport(options)
}

// GenerateJSONReport generates a JSON report
func (a *App) GenerateJSONReport(sessionID, title string, severity, status []string) (*api.ReportData, error) {
	return a.reportAPI.GenerateJSONReport(sessionID, title, severity, status)
}

// GenerateHTMLReport generates an HTML report
func (a *App) GenerateHTMLReport(sessionID, title string, severity, status []string) ([]byte, error) {
	return a.reportAPI.GenerateHTMLReport(sessionID, title, severity, status)
}

// GenerateMarkdownReport generates a Markdown report
func (a *App) GenerateMarkdownReport(sessionID, title string, severity, status []string) ([]byte, error) {
	return a.reportAPI.GenerateMarkdownReport(sessionID, title, severity, status)
}

// ListReports lists available reports
func (a *App) ListReports() ([]map[string]any, error) {
	return a.reportAPI.ListReports()
}

// SaveReportToFile saves report data to a file
func (a *App) SaveReportToFile(sessionID, title, format string, severity, status []string) (string, error) {
	var data []byte
	var defaultFilename string

	options := api.ReportOptions{
		SessionID: sessionID,
		Title:     title,
		Format:    format,
		Severity:  severity,
		Status:    status,
	}

	data, _, err := a.reportAPI.GenerateReport(options)
	if err != nil {
		return "", err
	}

	// Determine file extension
	switch format {
	case "html":
		defaultFilename = "report.html"
	case "markdown":
		defaultFilename = "report.md"
	case "json":
		defaultFilename = "report.json"
	default:
		defaultFilename = "report.txt"
	}

	// Open save dialog
	file, err := runtime.SaveFileDialog(a.ctx, runtime.SaveDialogOptions{
		Title:           "Save Report",
		DefaultFilename: defaultFilename,
		Filters: []runtime.FileFilter{
			{DisplayName: "Report Files", Pattern: "*.*"},
		},
	})
	if err != nil || file == "" {
		return "", fmt.Errorf("no file selected")
	}

	// Write to file
	if err := os.WriteFile(file, data, 0644); err != nil {
		return "", err
	}

	return file, nil
}

// ============ Intruder Operations ============

// CreateIntruderAttack creates a new intruder attack
func (a *App) CreateIntruderAttack(config intruder.AttackConfig) (string, error) {
	attack, err := a.intruderAPI.CreateAttack(config)
	if err != nil {
		return "", err
	}
	return attack.Config.ID, nil
}

// StartIntruderAttack starts an attack
func (a *App) StartIntruderAttack(attackID string) error {
	return a.intruderAPI.StartAttack(attackID)
}

// PauseIntruderAttack pauses a running attack
func (a *App) PauseIntruderAttack(attackID string) error {
	return a.intruderAPI.PauseAttack(attackID)
}

// ResumeIntruderAttack resumes a paused attack
func (a *App) ResumeIntruderAttack(attackID string) error {
	return a.intruderAPI.ResumeAttack(attackID)
}

// StopIntruderAttack stops an attack
func (a *App) StopIntruderAttack(attackID string) error {
	return a.intruderAPI.StopAttack(attackID)
}

// GetIntruderAttackProgress returns the progress of an attack
func (a *App) GetIntruderAttackProgress(attackID string) (map[string]interface{}, error) {
	progress, err := a.intruderAPI.GetAttackProgress(attackID)
	if err != nil {
		return nil, err
	}
	return map[string]interface{}{
		"total":      progress.Total,
		"completed":  progress.Completed,
		"errors":     progress.Errors,
		"status":     string(progress.Status),
		"currentRps": progress.CurrentRPS,
	}, nil
}

// GetIntruderAttackResults returns all results from an attack
func (a *App) GetIntruderAttackResults(attackID string) ([]map[string]interface{}, error) {
	results, err := a.intruderAPI.GetAttackResults(attackID)
	if err != nil {
		return nil, err
	}

	var mappedResults []map[string]interface{}
	for _, r := range results {
		mappedResults = append(mappedResults, map[string]interface{}{
			"id":           r.ID,
			"payload":      r.Payload,
			"statusCode":   r.StatusCode,
			"statusText":   r.StatusText,
			"responseTime": r.ResponseTime,
			"length":       r.Length,
			"error":        r.Error,
			"request":      r.Request,
			"response":     r.Response,
			"timestamp":    r.Timestamp,
		})
	}
	return mappedResults, nil
}

// ListIntruderAttacks returns all attack IDs
func (a *App) ListIntruderAttacks() []string {
	return a.intruderAPI.ListAttacks()
}

// RemoveIntruderAttack removes an attack
func (a *App) RemoveIntruderAttack(attackID string) error {
	return a.intruderAPI.RemoveAttack(attackID)
}

// DetectIntruderPositions finds payload positions in a request
func (a *App) DetectIntruderPositions(request string) []map[string]int {
	positions := a.intruderAPI.DetectPositions(request)
	var mapped []map[string]int
	for _, p := range positions {
		mapped = append(mapped, map[string]int{
			"start": p.Start,
			"end":   p.End,
		})
	}
	return mapped
}

// GetIntruderAttackTypes returns available attack types
func (a *App) GetIntruderAttackTypes() []api.AttackTypeInfo {
	return a.intruderAPI.GetAttackTypes()
}

// EstimateIntruderRequestCount estimates total requests for an attack
func (a *App) EstimateIntruderRequestCount(attackType string, positionCount int, payloadSetSizes []int) int {
	return a.intruderAPI.EstimateRequestCount(intruder.AttackType(attackType), positionCount, payloadSetSizes)
}

// ============ Embedded Scanner Operations ============

// ScanRequest scans an HTTP request for vulnerabilities
func (a *App) ScanRequest(url, method string, headers map[string]string, body string) ([]map[string]interface{}, error) {
	msg := &scanner.HTTPMessage{
		URL:       url,
		Method:    method,
		Headers:   headers,
		Body:      body,
		IsRequest: true,
	}
	findings := a.scannerAPI.ScanRequest(msg)
	return convertFindingsToMap(findings), nil
}

// ScanResponse scans an HTTP response for vulnerabilities
func (a *App) ScanResponse(url, method string, headers map[string]string, body string, statusCode int) ([]map[string]interface{}, error) {
	msg := &scanner.HTTPMessage{
		URL:        url,
		Method:     method,
		Headers:    headers,
		Body:       body,
		StatusCode: statusCode,
		IsRequest:  false,
	}
	findings := a.scannerAPI.ScanResponse(msg)
	return convertFindingsToMap(findings), nil
}

// GetEmbeddedScanFindings returns all findings from embedded scanner
func (a *App) GetEmbeddedScanFindings() []map[string]interface{} {
	findings := a.scannerAPI.GetEmbeddedFindings()
	return convertFindingsToMap(findings)
}

// ClearEmbeddedScanFindings clears all findings from embedded scanner
func (a *App) ClearEmbeddedScanFindings() {
	a.scannerAPI.ClearEmbeddedFindings()
}

// GetEmbeddedScanRules returns all rules from embedded scanner
func (a *App) GetEmbeddedScanRules() []map[string]interface{} {
	rules := a.scannerAPI.GetEmbeddedRules()
	var mapped []map[string]interface{}
	for _, r := range rules {
		mapped = append(mapped, map[string]interface{}{
			"id":          r.ID,
			"name":        r.Name,
			"severity":    string(r.Severity),
			"pattern":     r.Pattern,
			"location":    r.Location,
			"description": r.Description,
			"remediation": r.Remediation,
			"category":    r.Category,
			"enabled":     r.Enabled,
		})
	}
	return mapped
}

// EnableEmbeddedScanRule enables a rule in embedded scanner
func (a *App) EnableEmbeddedScanRule(ruleID string) {
	a.scannerAPI.EnableEmbeddedRule(ruleID)
}

// DisableEmbeddedScanRule disables a rule in embedded scanner
func (a *App) DisableEmbeddedScanRule(ruleID string) {
	a.scannerAPI.DisableEmbeddedRule(ruleID)
}

// SetScannerEnabled enables or disables the embedded scanner
func (a *App) SetScannerEnabled(enabled bool) {
	a.scannerAPI.SetScannerEnabled(enabled)
}

// IsScannerEnabled returns whether the embedded scanner is enabled
func (a *App) IsScannerEnabled() bool {
	return a.scannerAPI.IsScannerEnabled()
}

// ScanTrafficMessage scans both request and response for vulnerabilities
func (a *App) ScanTrafficMessage(url, method string, headers map[string]string, reqBody, respBody string, statusCode int) []map[string]interface{} {
	findings := a.scannerAPI.ScanTrafficMessage(url, method, headers, reqBody, respBody, statusCode)
	return convertFindingsToMap(findings)
}

// Helper function to convert scanner findings to map
func convertFindingsToMap(findings []*scanner.Vulnerability) []map[string]interface{} {
	var mapped []map[string]interface{}
	for _, f := range findings {
		mapped = append(mapped, map[string]interface{}{
			"id":            f.ID,
			"ruleId":        f.RuleID,
			"name":          f.Name,
			"severity":      string(f.Severity),
			"description":   f.Description,
			"remediation":   f.Remediation,
			"url":           f.URL,
			"method":        f.Method,
			"evidence":      f.Evidence,
			"location":      f.Location,
			"request":       f.Request,
			"response":      f.Response,
			"timestamp":     f.Timestamp,
			"falsePositive": f.FalsePositive,
		})
	}
	return mapped
}

// ============ Active Scan Operations ============

// CreateActiveScan creates a new active scan
func (a *App) CreateActiveScan(id, name string, concurrency, rateLimit, timeout int, followRedirects bool, enabledPlugins []string) error {
	config := a.activeScanAPI.CreateScanConfig(id, name, concurrency, rateLimit, timeout, followRedirects, enabledPlugins)
	_, err := a.activeScanAPI.CreateScan(*config)
	return err
}

// StartActiveScan starts an active scan
func (a *App) StartActiveScan(scanID string) error {
	return a.activeScanAPI.StartScan(scanID)
}

// PauseActiveScan pauses an active scan
func (a *App) PauseActiveScan(scanID string) error {
	return a.activeScanAPI.PauseScan(scanID)
}

// ResumeActiveScan resumes an active scan
func (a *App) ResumeActiveScan(scanID string) error {
	return a.activeScanAPI.ResumeScan(scanID)
}

// StopActiveScan stops an active scan
func (a *App) StopActiveScan(scanID string) error {
	return a.activeScanAPI.StopScan(scanID)
}

// GetActiveScanProgress returns the progress of an active scan
func (a *App) GetActiveScanProgress(scanID string) (map[string]interface{}, error) {
	progress, err := a.activeScanAPI.GetScanProgress(scanID)
	if err != nil {
		return nil, err
	}
	return map[string]interface{}{
		"totalTargets":   progress.TotalTargets,
		"scannedTargets": progress.ScannedTargets,
		"totalRequests":  progress.TotalRequests,
		"completedReqs":  progress.CompletedReqs,
		"findingsCount":  progress.FindingsCount,
		"errorCount":     progress.ErrorCount,
		"status":         string(progress.Status),
		"currentTarget":  progress.CurrentTarget,
		"currentPlugin":  progress.CurrentPlugin,
		"startTime":      progress.StartTime,
		"elapsedTime":    progress.ElapsedTime,
		"estimatedTime":  progress.EstimatedTime,
		"requestsPerSec": progress.RequestsPerSec,
	}, nil
}

// GetActiveScanFindings returns all findings from an active scan
func (a *App) GetActiveScanFindings(scanID string) ([]map[string]interface{}, error) {
	findings, err := a.activeScanAPI.GetScanFindings(scanID)
	if err != nil {
		return nil, err
	}

	var mapped []map[string]interface{}
	for _, f := range findings {
		mapped = append(mapped, map[string]interface{}{
			"id":          f.ID,
			"pluginId":    f.PluginID,
			"pluginName":  f.PluginName,
			"severity":    string(f.Severity),
			"title":       f.Title,
			"description": f.Description,
			"url":         f.URL,
			"method":      f.Method,
			"payload":     f.Payload,
			"evidence":    f.Evidence,
			"request":     f.Request,
			"response":    f.Response,
			"confidence":  f.Confidence,
			"timestamp":   f.Timestamp,
		})
	}
	return mapped, nil
}

// GetActiveScanStatus returns the status of an active scan
func (a *App) GetActiveScanStatus(scanID string) (string, error) {
	status, err := a.activeScanAPI.GetScanStatus(scanID)
	if err != nil {
		return "", err
	}
	return string(status), nil
}

// ListActiveScans returns all active scan IDs
func (a *App) ListActiveScans() []string {
	return a.activeScanAPI.ListScans()
}

// RemoveActiveScan removes an active scan
func (a *App) RemoveActiveScan(scanID string) error {
	return a.activeScanAPI.RemoveScan(scanID)
}

// AddActiveScanTarget adds a target to an active scan
func (a *App) AddActiveScanTarget(scanID, targetID, url, method string, headers map[string]string, body string) error {
	target := a.activeScanAPI.CreateTarget(targetID, url, method, headers, body)
	return a.activeScanAPI.AddTarget(scanID, *target)
}

// RemoveActiveScanTarget removes a target from an active scan
func (a *App) RemoveActiveScanTarget(scanID, targetID string) error {
	return a.activeScanAPI.RemoveTarget(scanID, targetID)
}

// GetActiveScanTargets returns all targets for an active scan
func (a *App) GetActiveScanTargets(scanID string) ([]map[string]interface{}, error) {
	targets, err := a.activeScanAPI.GetTargets(scanID)
	if err != nil {
		return nil, err
	}

	var mapped []map[string]interface{}
	for _, t := range targets {
		mapped = append(mapped, map[string]interface{}{
			"id":      t.ID,
			"url":     t.URL,
			"method":  t.Method,
			"headers": t.Headers,
			"body":    t.Body,
			"enabled": t.Enabled,
		})
	}
	return mapped, nil
}

// GetActiveScanPlugins returns all plugins for an active scan
func (a *App) GetActiveScanPlugins(scanID string) ([]map[string]interface{}, error) {
	return a.activeScanAPI.GetPlugins(scanID)
}

// GetDefaultActiveScanPlugins returns default plugin information
func (a *App) GetDefaultActiveScanPlugins() []map[string]interface{} {
	return a.activeScanAPI.GetDefaultPlugins()
}

// EnableActiveScanPlugin enables a plugin
func (a *App) EnableActiveScanPlugin(scanID, pluginID string) error {
	return a.activeScanAPI.EnablePlugin(scanID, pluginID)
}

// DisableActiveScanPlugin disables a plugin
func (a *App) DisableActiveScanPlugin(scanID, pluginID string) error {
	return a.activeScanAPI.DisablePlugin(scanID, pluginID)
}

// CreateActiveScanTarget creates a target struct
func (a *App) CreateActiveScanTarget(id, url, method string, headers map[string]string, body string) map[string]interface{} {
	target := a.activeScanAPI.CreateTarget(id, url, method, headers, body)
	return map[string]interface{}{
		"id":      target.ID,
		"url":     target.URL,
		"method":  target.Method,
		"headers": target.Headers,
		"body":    target.Body,
		"enabled": target.Enabled,
	}
}

// CreateActiveScanConfig creates a scan configuration
func (a *App) CreateActiveScanConfig(id, name string, concurrency, rateLimit, timeout int, followRedirects bool, enabledPlugins []string) map[string]interface{} {
	config := a.activeScanAPI.CreateScanConfig(id, name, concurrency, rateLimit, timeout, followRedirects, enabledPlugins)
	return map[string]interface{}{
		"id":              config.ID,
		"name":            config.Name,
		"concurrency":     config.Concurrency,
		"rateLimit":       config.RateLimit,
		"timeout":         config.Timeout,
		"followRedirects": config.FollowRedirects,
		"enabledPlugins":  config.EnabledPlugins,
	}
}
