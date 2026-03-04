// Package plugins provides vulnerability scanning plugins
package plugins

import (
	"bytes"
	"io"
	"net/http"
	"strings"
	"time"

	"hackmitm-desktop/internal/activescan"
)

// BasePlugin provides common functionality for plugins
type BasePlugin struct {
	id          string
	name        string
	description string
	severity    activescan.Severity
	enabled     bool
	payloads    []string
}

// NewBasePlugin creates a new base plugin
func NewBasePlugin(id, name, description string, severity activescan.Severity, payloads []string) *BasePlugin {
	return &BasePlugin{
		id:          id,
		name:        name,
		description: description,
		severity:    severity,
		enabled:     true,
		payloads:    payloads,
	}
}

func (p *BasePlugin) ID() string           { return p.id }
func (p *BasePlugin) Name() string         { return p.name }
func (p *BasePlugin) Description() string  { return p.description }
func (p *BasePlugin) Severity() activescan.Severity { return p.severity }
func (p *BasePlugin) Enabled() bool        { return p.enabled }
func (p *BasePlugin) SetEnabled(e bool)    { p.enabled = e }
func (p *BasePlugin) Payloads() []string   { return p.payloads }

// SendRequest sends an HTTP request and returns response
func SendRequest(client *http.Client, method, url string, headers map[string]string, body []byte) (*http.Response, string, error) {
	req, err := http.NewRequest(method, url, bytes.NewReader(body))
	if err != nil {
		return nil, "", err
	}

	for key, value := range headers {
		req.Header.Set(key, value)
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, "", err
	}

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		resp.Body.Close()
		return nil, "", err
	}
	resp.Body.Close()

	return resp, string(respBody), nil
}

// CheckResponseForPattern checks if response contains a pattern
func CheckResponseForPattern(response, pattern string) bool {
	return strings.Contains(strings.ToLower(response), strings.ToLower(pattern))
}

// CheckMultiplePatterns checks if response contains any of the patterns
func CheckMultiplePatterns(response string, patterns []string) bool {
	lowerResp := strings.ToLower(response)
	for _, pattern := range patterns {
		if strings.Contains(lowerResp, strings.ToLower(pattern)) {
			return true
		}
	}
	return false
}

// CreateFinding creates a finding from scan result
func CreateFinding(plugin *BasePlugin, target *activescan.Target, payload, evidence, request, response string) *activescan.Finding {
	return &activescan.Finding{
		PluginID:    plugin.ID(),
		PluginName:  plugin.Name(),
		Severity:    plugin.Severity(),
		Title:       plugin.Name(),
		Description: plugin.Description(),
		URL:         target.URL,
		Method:      target.Method,
		Payload:     payload,
		Evidence:    evidence,
		Request:     request,
		Response:    truncateString(response, 5000),
		Timestamp:   time.Now(),
		Confidence:  80,
		Headers:     target.Headers,
	}
}

func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}
