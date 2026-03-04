package api

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/wailsapp/wails/v2/pkg/runtime"

	"hackmitm-desktop/internal/models"
)

// TrafficAPI handles traffic-related operations
type TrafficAPI struct {
	client      *http.Client
	apiEndpoint string
	ctx         context.Context
}

// NewTrafficAPI creates a new TrafficAPI instance
func NewTrafficAPI() *TrafficAPI {
	return &TrafficAPI{
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// SetContext sets the application context
func (t *TrafficAPI) SetContext(ctx context.Context) {
	t.ctx = ctx
}

// SetAPIEndpoint sets the API endpoint
func (t *TrafficAPI) SetAPIEndpoint(endpoint string) {
	t.apiEndpoint = endpoint
}

// GetTraffic fetches traffic history
func (t *TrafficAPI) GetTraffic(limit int) ([]models.TrafficItem, error) {
	if t.apiEndpoint == "" {
		return nil, fmt.Errorf("API endpoint not configured")
	}

	url := fmt.Sprintf("%s/api/traffic?limit=%d", t.apiEndpoint, limit)
	resp, err := t.client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch traffic: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API returned status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var items []models.TrafficItem
	if err := json.Unmarshal(body, &items); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return items, nil
}

// ClearTraffic clears all traffic history
func (t *TrafficAPI) ClearTraffic() error {
	if t.apiEndpoint == "" {
		return fmt.Errorf("API endpoint not configured")
	}

	url := fmt.Sprintf("%s/api/traffic", t.apiEndpoint)
	req, err := http.NewRequest(http.MethodDelete, url, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := t.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to clear traffic: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("API returned status %d", resp.StatusCode)
	}

	return nil
}

// OnTrafficEvent emits a traffic event to the frontend
func (t *TrafficAPI) OnTrafficEvent(item models.TrafficItem) {
	if t.ctx != nil {
		runtime.EventsEmit(t.ctx, "traffic:new", item)
	}
}
