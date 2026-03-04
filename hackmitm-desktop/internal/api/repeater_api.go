package api

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"hackmitm-desktop/internal/models"
)

// RepeaterAPI handles repeater-related operations
type RepeaterAPI struct {
	client      *http.Client
	apiEndpoint string
	ctx         context.Context
}

// NewRepeaterAPI creates a new RepeaterAPI instance
func NewRepeaterAPI() *RepeaterAPI {
	return &RepeaterAPI{
		client: &http.Client{
			Timeout: 60 * time.Second,
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				return http.ErrUseLastResponse // Don't follow redirects
			},
		},
	}
}

// SetContext sets the application context
func (r *RepeaterAPI) SetContext(ctx context.Context) {
	r.ctx = ctx
}

// SetAPIEndpoint sets the API endpoint
func (r *RepeaterAPI) SetAPIEndpoint(endpoint string) {
	r.apiEndpoint = endpoint
}

// SendRequest sends an HTTP request and returns the response
func (r *RepeaterAPI) SendRequest(req models.RepeaterRequest) (*models.RepeaterResponse, error) {
	// Create the HTTP request
	var bodyReader io.Reader
	if req.Body != "" {
		bodyReader = strings.NewReader(req.Body)
	}

	httpReq, err := http.NewRequest(req.Method, req.URL, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	for key, value := range req.Headers {
		httpReq.Header.Set(key, value)
	}

	// Set default User-Agent if not provided
	if httpReq.Header.Get("User-Agent") == "" {
		httpReq.Header.Set("User-Agent", "HackMITM-Desktop/1.0")
	}

	// Send request
	startTime := time.Now()
	httpResp, err := r.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer httpResp.Body.Close()
	duration := time.Since(startTime)

	// Read response body
	respBody, err := io.ReadAll(httpResp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	// Build response headers
	respHeaders := make(map[string]string)
	for key, values := range httpResp.Header {
		respHeaders[key] = strings.Join(values, ", ")
	}

	// Build response
	response := &models.RepeaterResponse{
		StatusCode:    httpResp.StatusCode,
		StatusText:    httpResp.Status,
		Headers:       respHeaders,
		Body:          string(respBody),
		ResponseTime:  duration.Milliseconds(),
		ContentLength: int64(len(respBody)),
	}

	return response, nil
}

// SendRequestViaProxy sends a request through the HackMITM proxy
func (r *RepeaterAPI) SendRequestViaProxy(req models.RepeaterRequest, proxyURL string) (*models.RepeaterResponse, error) {
	// This would use the proxy as an upstream
	// For now, just use direct request
	return r.SendRequest(req)
}
