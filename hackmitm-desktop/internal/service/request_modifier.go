package service

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// RequestModification defines the modifications to apply to a request
type RequestModification struct {
	Method  string            `json:"method,omitempty"`
	URL     string            `json:"url,omitempty"`
	Headers map[string]string `json:"headers,omitempty"`
	Body    string            `json:"body,omitempty"`
}

// ModifiedRequestResult contains the result of a modified request
type ModifiedRequestResult struct {
	StatusCode   int               `json:"statusCode"`
	StatusText   string            `json:"statusText"`
	Headers      map[string]string `json:"headers"`
	Body         string            `json:"body"`
	ResponseTime int64             `json:"responseTime"`
	ContentType  string            `json:"contentType"`
	Error        string            `json:"error,omitempty"`
}

// RequestModifier handles request modification and forwarding
type RequestModifier struct {
	apiEndpoint string
	client      *http.Client
}

// NewRequestModifier creates a new request modifier
func NewRequestModifier() *RequestModifier {
	return &RequestModifier{
		client: &http.Client{
			Timeout: 30 * time.Second,
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				return http.ErrUseLastResponse // Don't follow redirects
			},
		},
	}
}

// SetAPIEndpoint sets the API endpoint
func (m *RequestModifier) SetAPIEndpoint(endpoint string) {
	m.apiEndpoint = endpoint
}

// ModifyAndForward modifies a request and forwards it to the target
func (m *RequestModifier) ModifyAndForward(originalURL string, mod RequestModification) (*ModifiedRequestResult, error) {
	startTime := time.Now()
	result := &ModifiedRequestResult{}

	// Determine the target URL
	targetURL := originalURL
	if mod.URL != "" {
		// If it's a relative URL, combine with original host
		if strings.HasPrefix(mod.URL, "/") {
			parsedURL, err := url.Parse(originalURL)
			if err == nil {
				parsedURL.Path = mod.URL
				targetURL = parsedURL.String()
			}
		} else {
			targetURL = mod.URL
		}
	}

	// Determine the method
	method := "GET"
	if mod.Method != "" {
		method = strings.ToUpper(mod.Method)
	}

	// Create the request
	var bodyReader io.Reader
	if mod.Body != "" && (method == "POST" || method == "PUT" || method == "PATCH") {
		bodyReader = bytes.NewReader([]byte(mod.Body))
	}

	req, err := http.NewRequest(method, targetURL, bodyReader)
	if err != nil {
		result.Error = err.Error()
		return result, fmt.Errorf("failed to create request: %w", err)
	}

	// Apply headers
	for key, value := range mod.Headers {
		req.Header.Set(key, value)
	}

	// Set Content-Type if body exists and Content-Type not set
	if mod.Body != "" && req.Header.Get("Content-Type") == "" {
		// Try to detect content type
		if json.Valid([]byte(mod.Body)) {
			req.Header.Set("Content-Type", "application/json")
		} else if strings.Contains(mod.Body, "=") && !strings.Contains(mod.Body, "{") {
			req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		}
	}

	// Execute the request
	resp, err := m.client.Do(req)
	if err != nil {
		result.Error = err.Error()
		return result, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	// Read response body
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		result.Error = err.Error()
		return result, fmt.Errorf("failed to read response: %w", err)
	}

	// Build result
	result.StatusCode = resp.StatusCode
	result.StatusText = resp.Status
	result.ResponseTime = time.Since(startTime).Milliseconds()
	result.ContentType = resp.Header.Get("Content-Type")

	// Convert headers
	result.Headers = make(map[string]string)
	for key, values := range resp.Header {
		if len(values) > 0 {
			result.Headers[key] = values[0]
		}
	}

	// Convert body to string
	result.Body = string(bodyBytes)

	return result, nil
}

// SendModifiedIntercepted sends a modified intercepted request through the proxy
// This uses the proxy's API to forward the modified request
func (m *RequestModifier) SendModifiedIntercepted(requestID string, mod RequestModification) (*ModifiedRequestResult, error) {
	if m.apiEndpoint == "" {
		return nil, fmt.Errorf("API endpoint not set")
	}

	// Build the request to the proxy API
	apiURL := fmt.Sprintf("%s/api/v1/intercept/%s/modify", m.apiEndpoint, requestID)

	bodyBytes, err := json.Marshal(mod)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal modification: %w", err)
	}

	req, err := http.NewRequest("POST", apiURL, bytes.NewReader(bodyBytes))
	if err != nil {
		return nil, fmt.Errorf("failed to create API request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := m.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to call proxy API: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read API response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("proxy API error: %s - %s", resp.Status, string(respBody))
	}

	var result ModifiedRequestResult
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("failed to parse API response: %w", err)
	}

	return &result, nil
}
