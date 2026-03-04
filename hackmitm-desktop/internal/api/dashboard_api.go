package api

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"hackmitm-desktop/internal/models"
)

// DashboardAPI handles dashboard-related operations
type DashboardAPI struct {
	client      *http.Client
	apiEndpoint string
	ctx         context.Context
}

// NewDashboardAPI creates a new DashboardAPI instance
func NewDashboardAPI() *DashboardAPI {
	return &DashboardAPI{
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// SetContext sets the application context
func (d *DashboardAPI) SetContext(ctx context.Context) {
	d.ctx = ctx
}

// SetAPIEndpoint sets the API endpoint
func (d *DashboardAPI) SetAPIEndpoint(endpoint string) {
	d.apiEndpoint = endpoint
}

// GetMetrics fetches dashboard metrics from /metrics endpoint
func (d *DashboardAPI) GetMetrics() (*models.DashboardMetrics, error) {
	if d.apiEndpoint == "" {
		return nil, fmt.Errorf("API endpoint not configured")
	}

	resp, err := d.client.Get(fmt.Sprintf("%s/metrics", d.apiEndpoint))
	if err != nil {
		return nil, fmt.Errorf("failed to fetch metrics: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API returned status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	// Parse the raw metrics response
	var rawMetrics map[string]interface{}
	if err := json.Unmarshal(body, &rawMetrics); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	// Convert to DashboardMetrics
	metrics := &models.DashboardMetrics{}

	// Try different field names for QPS
	if v, ok := rawMetrics["requests_per_sec"].(float64); ok {
		metrics.QPS = v
	} else if v, ok := rawMetrics["requests_per_second"].(float64); ok {
		metrics.QPS = v
	}

	// Try different field names for total requests
	if v, ok := rawMetrics["requests"].(float64); ok {
		metrics.TotalRequests = int64(v)
	} else if v, ok := rawMetrics["total_requests"].(float64); ok {
		metrics.TotalRequests = int64(v)
	}

	// Try different field names for active connections
	if v, ok := rawMetrics["active_conns"].(float64); ok {
		metrics.ActiveConns = int64(v)
	} else if v, ok := rawMetrics["active_connections"].(float64); ok {
		metrics.ActiveConns = int64(v)
	}

	// Bytes
	if v, ok := rawMetrics["bytes_in"].(float64); ok {
		metrics.TotalBytesIn = int64(v)
	}
	if v, ok := rawMetrics["bytes_out"].(float64); ok {
		metrics.TotalBytesOut = int64(v)
	}

	// Errors
	if v, ok := rawMetrics["errors"].(float64); ok {
		totalReqs := metrics.TotalRequests
		if totalReqs > 0 {
			metrics.ErrorRate = v / float64(totalReqs)
		}
	}

	// Parse average response time from string (e.g., "15.234ms")
	if v, ok := rawMetrics["avg_response_time"].(string); ok {
		if duration, err := time.ParseDuration(v); err == nil {
			metrics.AvgResponseTime = float64(duration.Milliseconds())
		}
	}

	// Parse uptime from string
	if v, ok := rawMetrics["uptime"].(string); ok {
		if duration, err := time.ParseDuration(v); err == nil {
			metrics.Uptime = int64(duration.Seconds())
		}
	}

	return metrics, nil
}

// GetTrafficPatterns fetches traffic pattern statistics from /patterns/stats
func (d *DashboardAPI) GetTrafficPatterns() (map[string]any, error) {
	if d.apiEndpoint == "" {
		return nil, fmt.Errorf("API endpoint not configured")
	}

	resp, err := d.client.Get(fmt.Sprintf("%s/patterns/stats", d.apiEndpoint))
	if err != nil {
		return nil, fmt.Errorf("failed to fetch patterns: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API returned status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var patterns map[string]any
	if err := json.Unmarshal(body, &patterns); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return patterns, nil
}
