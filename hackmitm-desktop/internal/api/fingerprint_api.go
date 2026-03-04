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

// FingerprintAPI handles fingerprint-related operations
type FingerprintAPI struct {
	client      *http.Client
	apiEndpoint string
	ctx         context.Context
}

// NewFingerprintAPI creates a new FingerprintAPI instance
func NewFingerprintAPI() *FingerprintAPI {
	return &FingerprintAPI{
		client: &http.Client{
			Timeout: 60 * time.Second,
		},
	}
}

// SetContext sets the application context
func (f *FingerprintAPI) SetContext(ctx context.Context) {
	f.ctx = ctx
}

// SetAPIEndpoint sets the API endpoint
func (f *FingerprintAPI) SetAPIEndpoint(endpoint string) {
	f.apiEndpoint = endpoint
}

// GetFingerprintStats fetches fingerprint statistics
func (f *FingerprintAPI) GetFingerprintStats() (map[string]interface{}, error) {
	if f.apiEndpoint == "" {
		return nil, fmt.Errorf("API endpoint not configured")
	}

	resp, err := f.client.Get(fmt.Sprintf("%s/fingerprint/stats", f.apiEndpoint))
	if err != nil {
		return nil, fmt.Errorf("failed to fetch fingerprint stats: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API returned status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var stats map[string]interface{}
	if err := json.Unmarshal(body, &stats); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return stats, nil
}

// IdentifyFingerprint manually triggers fingerprint identification for a URL
func (f *FingerprintAPI) IdentifyFingerprint(targetURL string) (*models.FingerprintResult, error) {
	if f.apiEndpoint == "" {
		return nil, fmt.Errorf("API endpoint not configured")
	}

	// Prepare request body
	reqBody := map[string]string{"url": targetURL}
	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	resp, err := f.client.Post(
		fmt.Sprintf("%s/fingerprint/identify", f.apiEndpoint),
		"application/json",
		bytes.NewReader(bodyBytes),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to identify fingerprint: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API returned status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var result models.FingerprintResult
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &result, nil
}

// GetFingerprintHistory fetches fingerprint identification history
func (f *FingerprintAPI) GetFingerprintHistory(limit int) ([]models.FingerprintResult, error) {
	if f.apiEndpoint == "" {
		return nil, fmt.Errorf("API endpoint not configured")
	}

	resp, err := f.client.Get(fmt.Sprintf("%s/fingerprint/history?limit=%d", f.apiEndpoint, limit))
	if err != nil {
		return nil, fmt.Errorf("failed to fetch fingerprint history: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API returned status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var results []models.FingerprintResult
	if err := json.Unmarshal(body, &results); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return results, nil
}

// OnFingerprintEvent emits a fingerprint event to the frontend
func (f *FingerprintAPI) OnFingerprintEvent(result models.FingerprintResult) {
	if f.ctx != nil {
		runtime.EventsEmit(f.ctx, "fingerprint:new", result)
	}
}
