package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/wailsapp/wails/v2/pkg/runtime"

	"hackmitm-desktop/internal/models"
)

// ProxyAPI handles proxy-related operations
type ProxyAPI struct {
	client      *http.Client
	apiEndpoint string
	ctx         context.Context
}

// NewProxyAPI creates a new ProxyAPI instance
func NewProxyAPI() *ProxyAPI {
	return &ProxyAPI{
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// SetContext sets the application context
func (p *ProxyAPI) SetContext(ctx context.Context) {
	p.ctx = ctx
}

// SetAPIEndpoint sets the API endpoint
func (p *ProxyAPI) SetAPIEndpoint(endpoint string) {
	p.apiEndpoint = endpoint
}

// GetStatus fetches the proxy server status from /status endpoint
func (p *ProxyAPI) GetStatus() (*models.ProxyStatus, error) {
	if p.apiEndpoint == "" {
		return nil, fmt.Errorf("API endpoint not configured")
	}

	resp, err := p.client.Get(fmt.Sprintf("%s/status", p.apiEndpoint))
	if err != nil {
		return nil, fmt.Errorf("failed to fetch status: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API returned status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	// Parse the response - /status returns {metrics: {...}, health: {...}}
	var statusResp struct {
		Metrics map[string]interface{} `json:"metrics"`
		Health  struct {
			Status string `json:"status"`
		} `json:"health"`
	}
	if err := json.Unmarshal(body, &statusResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	// Extract relevant fields from metrics
	status := &models.ProxyStatus{
		Running: statusResp.Health.Status == "healthy",
	}

	if metrics := statusResp.Metrics; metrics != nil {
		if v, ok := metrics["active_conns"].(float64); ok {
			status.ActiveConnections = int64(v)
		}
		if v, ok := metrics["requests"].(float64); ok {
			status.TotalRequests = int64(v)
		}
		if v, ok := metrics["uptime"].(string); ok {
			// Parse uptime string (e.g., "1h30m15s")
			if duration, err := time.ParseDuration(v); err == nil {
				status.Uptime = int64(duration.Seconds())
			}
		}
	}

	return status, nil
}

// HealthCheck performs a health check on the server
func (p *ProxyAPI) HealthCheck() (map[string]interface{}, error) {
	if p.apiEndpoint == "" {
		return nil, fmt.Errorf("API endpoint not configured")
	}

	resp, err := p.client.Get(fmt.Sprintf("%s/health", p.apiEndpoint))
	if err != nil {
		return nil, fmt.Errorf("health check failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API returned status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var health map[string]interface{}
	if err := json.Unmarshal(body, &health); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return health, nil
}

// SetInterceptMode enables or disables intercept mode
// Note: This requires the HackMITM server to have an intercept endpoint
func (p *ProxyAPI) SetInterceptMode(enabled bool) error {
	if p.apiEndpoint == "" {
		return fmt.Errorf("API endpoint not configured")
	}

	reqBody := map[string]bool{"enabled": enabled}
	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequest(
		http.MethodPost,
		fmt.Sprintf("%s/proxy/intercept", p.apiEndpoint),
		bytes.NewReader(bodyBytes),
	)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := p.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to set intercept mode: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("API returned status %d: %s", resp.StatusCode, string(body))
	}

	return nil
}

// ForwardIntercepted forwards an intercepted request
func (p *ProxyAPI) ForwardIntercepted(requestID string) error {
	if p.apiEndpoint == "" {
		return fmt.Errorf("API endpoint not configured")
	}

	req, err := http.NewRequest(
		http.MethodPost,
		fmt.Sprintf("%s/proxy/intercept/%s/forward", p.apiEndpoint, requestID),
		nil,
	)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := p.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to forward request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("API returned status %d", resp.StatusCode)
	}

	return nil
}

// DropIntercepted drops an intercepted request
func (p *ProxyAPI) DropIntercepted(requestID string) error {
	if p.apiEndpoint == "" {
		return fmt.Errorf("API endpoint not configured")
	}

	req, err := http.NewRequest(
		http.MethodPost,
		fmt.Sprintf("%s/proxy/intercept/%s/drop", p.apiEndpoint, requestID),
		nil,
	)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := p.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to drop request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("API returned status %d", resp.StatusCode)
	}

	return nil
}

// OnStatusUpdate emits a status update event to the frontend
func (p *ProxyAPI) OnStatusUpdate(status models.ProxyStatus) {
	if p.ctx != nil {
		runtime.EventsEmit(p.ctx, "proxy:status", status)
	}
}
