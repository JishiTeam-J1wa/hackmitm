// Package activescan implements active vulnerability scanning
package activescan

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"sync"
	"sync/atomic"
	"time"
)

// ScanStatus represents the status of a scan
type ScanStatus string

const (
	StatusIdle      ScanStatus = "idle"
	StatusRunning   ScanStatus = "running"
	StatusPaused    ScanStatus = "paused"
	StatusComplete  ScanStatus = "complete"
	StatusError     ScanStatus = "error"
	StatusCancelled ScanStatus = "cancelled"
)

// Severity levels for findings
type Severity string

const (
	SeverityCritical Severity = "critical"
	SeverityHigh     Severity = "high"
	SeverityMedium   Severity = "medium"
	SeverityLow      Severity = "low"
	SeverityInfo     Severity = "info"
)

// Target represents a scan target
type Target struct {
	ID      string `json:"id"`
	URL     string `json:"url"`
	Method  string `json:"method"`
	Headers map[string]string `json:"headers"`
	Body    string `json:"body"`
	Enabled bool   `json:"enabled"`
}

// Finding represents a vulnerability finding from active scanning
type Finding struct {
	ID          string            `json:"id"`
	PluginID    string            `json:"pluginId"`
	PluginName  string            `json:"pluginName"`
	Severity    Severity          `json:"severity"`
	Title       string            `json:"title"`
	Description string            `json:"description"`
	URL         string            `json:"url"`
	Method      string            `json:"method"`
	Payload     string            `json:"payload"`
	Evidence    string            `json:"evidence"`
	Request     string            `json:"request"`
	Response    string            `json:"response"`
	Timestamp   time.Time         `json:"timestamp"`
	Confidence  int               `json:"confidence"` // 0-100
	Headers     map[string]string `json:"headers"`
}

// ScanProgress tracks scan progress
type ScanProgress struct {
	TotalTargets    int32 `json:"totalTargets"`
	ScannedTargets  int32 `json:"scannedTargets"`
	TotalRequests   int32 `json:"totalRequests"`
	CompletedReqs   int32 `json:"completedRequests"`
	FindingsCount   int32 `json:"findingsCount"`
	ErrorCount      int32 `json:"errorCount"`
	Status          ScanStatus `json:"status"`
	CurrentTarget   string `json:"currentTarget"`
	CurrentPlugin   string `json:"currentPlugin"`
	StartTime       time.Time `json:"startTime"`
	ElapsedTime     int64 `json:"elapsedTime"` // seconds
	EstimatedTime   int64 `json:"estimatedTime"` // seconds
	RequestsPerSec  float64 `json:"requestsPerSec"`
}

// ScanConfig represents scan configuration
type ScanConfig struct {
	ID              string   `json:"id"`
	Name            string   `json:"name"`
	Concurrency     int      `json:"concurrency"`
	RateLimit       int      `json:"rateLimit"` // requests per second, 0 = unlimited
	Timeout         int      `json:"timeout"`   // request timeout in seconds
	FollowRedirects bool     `json:"followRedirects"`
	MaxDepth        int      `json:"maxDepth"`
	UserAgent       string   `json:"userAgent"`
	EnabledPlugins  []string `json:"enabledPlugins"`
}

// Plugin interface for scan plugins
type Plugin interface {
	ID() string
	Name() string
	Description() string
	Severity() Severity
	Scan(target *Target, client *http.Client, config *ScanConfig) ([]*Finding, error)
	Enabled() bool
	SetEnabled(bool)
}

// ActiveScan represents an active scan instance
type ActiveScan struct {
	config    *ScanConfig
	targets   []*Target
	plugins   []Plugin
	progress  *ScanProgress
	findings  []*Finding
	client    *http.Client
	mu        sync.RWMutex
	ctx       context.Context
	cancel    context.CancelFunc
	pauseChan chan struct{}
	resumeChan chan struct{}

	// Callbacks
	OnProgress func(*ScanProgress)
	OnFinding  func(*Finding)
	OnComplete func()
}

// NewActiveScan creates a new active scan instance
func NewActiveScan(config *ScanConfig) *ActiveScan {
	if config.Concurrency <= 0 {
		config.Concurrency = 5
	}
	if config.Timeout <= 0 {
		config.Timeout = 30
	}
	if config.UserAgent == "" {
		config.UserAgent = "HackMITM-ActiveScanner/1.0"
	}

	return &ActiveScan{
		config:    config,
		targets:   make([]*Target, 0),
		plugins:   make([]Plugin, 0),
		findings:  make([]*Finding, 0),
		pauseChan: make(chan struct{}),
		resumeChan: make(chan struct{}),
		progress: &ScanProgress{
			Status: StatusIdle,
		},
	}
}

// AddTarget adds a target to scan
func (s *ActiveScan) AddTarget(target *Target) {
	s.mu.Lock()
	defer s.mu.Unlock()
	target.Enabled = true
	s.targets = append(s.targets, target)
}

// RemoveTarget removes a target by ID
func (s *ActiveScan) RemoveTarget(id string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for i, t := range s.targets {
		if t.ID == id {
			s.targets = append(s.targets[:i], s.targets[i+1:]...)
			break
		}
	}
}

// GetTargets returns all targets
func (s *ActiveScan) GetTargets() []*Target {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.targets
}

// ClearTargets clears all targets
func (s *ActiveScan) ClearTargets() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.targets = make([]*Target, 0)
}

// RegisterPlugin registers a scan plugin
func (s *ActiveScan) RegisterPlugin(plugin Plugin) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.plugins = append(s.plugins, plugin)
}

// GetPlugins returns all registered plugins
func (s *ActiveScan) GetPlugins() []Plugin {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.plugins
}

// GetPluginByID returns a plugin by ID
func (s *ActiveScan) GetPluginByID(id string) Plugin {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, p := range s.plugins {
		if p.ID() == id {
			return p
		}
	}
	return nil
}

// EnablePlugin enables a plugin by ID
func (s *ActiveScan) EnablePlugin(id string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, p := range s.plugins {
		if p.ID() == id {
			p.SetEnabled(true)
			break
		}
	}
}

// DisablePlugin disables a plugin by ID
func (s *ActiveScan) DisablePlugin(id string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, p := range s.plugins {
		if p.ID() == id {
			p.SetEnabled(false)
			break
		}
	}
}

// Start begins the active scan
func (s *ActiveScan) Start(ctx context.Context) error {
	s.mu.Lock()
	if s.progress.Status == StatusRunning {
		s.mu.Unlock()
		return fmt.Errorf("scan already running")
	}

	s.ctx, s.cancel = context.WithCancel(ctx)
	s.progress.Status = StatusRunning
	s.progress.StartTime = time.Now()
	s.progress.TotalTargets = int32(len(s.targets))
	atomic.StoreInt32(&s.progress.ScannedTargets, 0)
	atomic.StoreInt32(&s.progress.CompletedReqs, 0)
	atomic.StoreInt32(&s.progress.FindingsCount, 0)
	atomic.StoreInt32(&s.progress.ErrorCount, 0)
	s.mu.Unlock()

	// Create HTTP client
	s.client = &http.Client{
		Timeout: time.Duration(s.config.Timeout) * time.Second,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if !s.config.FollowRedirects {
				return http.ErrUseLastResponse
			}
			if len(via) >= s.config.MaxDepth && s.config.MaxDepth > 0 {
				return fmt.Errorf("max redirect depth reached")
			}
			return nil
		},
	}

	// Start scan in background
	go s.runScan()

	return nil
}

// runScan executes the scan
func (s *ActiveScan) runScan() {
	var wg sync.WaitGroup
	targetChan := make(chan *Target, s.config.Concurrency)

	// Rate limiter
	var lastRequest time.Time
	rateLimitInterval := time.Duration(0)
	if s.config.RateLimit > 0 {
		rateLimitInterval = time.Second / time.Duration(s.config.RateLimit)
	}

	// Start workers
	for i := 0; i < s.config.Concurrency; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for target := range targetChan {
				select {
				case <-s.ctx.Done():
					return
				default:
				}

				// Handle pause
				select {
				case <-s.pauseChan:
					<-s.resumeChan // Wait for resume
				default:
				}

				s.scanTarget(target, rateLimitInterval, &lastRequest)
				atomic.AddInt32(&s.progress.ScannedTargets, 1)

				if s.OnProgress != nil {
					s.OnProgress(s.progress)
				}
			}
		}()
	}

	// Send targets to workers
	go func() {
		for _, target := range s.targets {
			if !target.Enabled {
				continue
			}
			select {
			case <-s.ctx.Done():
				break
			default:
				targetChan <- target
			}
		}
		close(targetChan)
	}()

	// Wait for completion
	wg.Wait()

	// Update final status
	s.mu.Lock()
	if s.progress.Status == StatusRunning {
		s.progress.Status = StatusComplete
	}
	s.progress.ElapsedTime = int64(time.Since(s.progress.StartTime).Seconds())
	if s.progress.ElapsedTime > 0 {
		s.progress.RequestsPerSec = float64(s.progress.CompletedReqs) / float64(s.progress.ElapsedTime)
	}
	s.mu.Unlock()

	if s.OnComplete != nil {
		s.OnComplete()
	}
}

// scanTarget scans a single target with all enabled plugins
func (s *ActiveScan) scanTarget(target *Target, rateLimitInterval time.Duration, lastRequest *time.Time) {
	s.mu.Lock()
	s.progress.CurrentTarget = target.URL
	s.mu.Unlock()

	for _, plugin := range s.plugins {
		select {
		case <-s.ctx.Done():
			return
		default:
		}

		if !plugin.Enabled() {
			continue
		}

		// Check if plugin is in enabled list (if specified)
		if len(s.config.EnabledPlugins) > 0 {
			found := false
			for _, id := range s.config.EnabledPlugins {
				if id == plugin.ID() {
					found = true
					break
				}
			}
			if !found {
				continue
			}
		}

		s.mu.Lock()
		s.progress.CurrentPlugin = plugin.Name()
		s.mu.Unlock()

		// Rate limiting
		if rateLimitInterval > 0 {
			elapsed := time.Since(*lastRequest)
			if elapsed < rateLimitInterval {
				time.Sleep(rateLimitInterval - elapsed)
			}
			*lastRequest = time.Now()
		}

		// Run plugin
		findings, err := plugin.Scan(target, s.client, s.config)
		atomic.AddInt32(&s.progress.CompletedReqs, 1)

		if err != nil {
			atomic.AddInt32(&s.progress.ErrorCount, 1)
			continue
		}

		// Add findings
		for _, f := range findings {
			s.addFinding(f)
		}
	}
}

// addFinding adds a finding to the scan results
func (s *ActiveScan) addFinding(finding *Finding) {
	s.mu.Lock()
	defer s.mu.Unlock()

	finding.ID = generateFindingID(finding.PluginID, finding.URL)
	finding.Timestamp = time.Now()
	s.findings = append(s.findings, finding)
	atomic.AddInt32(&s.progress.FindingsCount, 1)

	if s.OnFinding != nil {
		go s.OnFinding(finding)
	}
}

// Pause pauses the scan
func (s *ActiveScan) Pause() {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.progress.Status == StatusRunning {
		s.progress.Status = StatusPaused
		s.pauseChan <- struct{}{}
	}
}

// Resume resumes a paused scan
func (s *ActiveScan) Resume() {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.progress.Status == StatusPaused {
		s.progress.Status = StatusRunning
		s.resumeChan <- struct{}{}
	}
}

// Stop stops the scan
func (s *ActiveScan) Stop() {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.cancel != nil {
		s.cancel()
	}
	s.progress.Status = StatusCancelled
}

// GetProgress returns current scan progress
func (s *ActiveScan) GetProgress() *ScanProgress {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Update elapsed time
	if !s.progress.StartTime.IsZero() {
		s.progress.ElapsedTime = int64(time.Since(s.progress.StartTime).Seconds())
		if s.progress.ElapsedTime > 0 && s.progress.CompletedReqs > 0 {
			s.progress.RequestsPerSec = float64(s.progress.CompletedReqs) / float64(s.progress.ElapsedTime)
		}
	}

	return s.progress
}

// GetFindings returns all findings
func (s *ActiveScan) GetFindings() []*Finding {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.findings
}

// ClearFindings clears all findings
func (s *ActiveScan) ClearFindings() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.findings = make([]*Finding, 0)
	atomic.StoreInt32(&s.progress.FindingsCount, 0)
}

// GetStatus returns current scan status
func (s *ActiveScan) GetStatus() ScanStatus {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.progress.Status
}

// Helper functions

func generateFindingID(pluginID, url string) string {
	return fmt.Sprintf("%s-%s-%d", pluginID, url, time.Now().UnixNano())
}

// BuildRequest builds an HTTP request from a target
func BuildRequest(target *Target, body []byte) (*http.Request, error) {
	req, err := http.NewRequest(target.Method, target.URL, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}

	// Add headers
	for key, value := range target.Headers {
		req.Header.Set(key, value)
	}

	return req, nil
}

// ReadResponse reads an HTTP response and returns body as string
func ReadResponse(resp *http.Response) (string, error) {
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	return string(body), nil
}

// RequestToString converts an HTTP request to string for logging
func RequestToString(req *http.Request) string {
	var buf bytes.Buffer
	buf.WriteString(fmt.Sprintf("%s %s HTTP/1.1\r\n", req.Method, req.URL.Path))
	for key, values := range req.Header {
		for _, value := range values {
			buf.WriteString(fmt.Sprintf("%s: %s\r\n", key, value))
		}
	}
	buf.WriteString("\r\n")
	return buf.String()
}

// ResponseToString converts an HTTP response to string for logging
func ResponseToString(resp *http.Response, body string) string {
	var buf bytes.Buffer
	buf.WriteString(fmt.Sprintf("HTTP/1.1 %d %s\r\n", resp.StatusCode, resp.Status))
	for key, values := range resp.Header {
		for _, value := range values {
			buf.WriteString(fmt.Sprintf("%s: %s\r\n", key, value))
		}
	}
	buf.WriteString("\r\n")
	if len(body) > 1000 {
		buf.WriteString(body[:1000] + "...")
	} else {
		buf.WriteString(body)
	}
	return buf.String()
}
