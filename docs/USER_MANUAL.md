# HackMITM User Manual

## Table of Contents

1. [Introduction](#introduction)
2. [Getting Started](#getting-started)
3. [Dashboard](#dashboard)
4. [Traffic Analysis](#traffic-analysis)
5. [Vulnerability Management](#vulnerability-management)
6. [Passive Scanner](#passive-scanner)
7. [Sessions](#sessions)
8. [Reports](#reports)
9. [Intercept Mode](#intercept-mode)
10. [Repeater](#repeater)
11. [Keyboard Shortcuts](#keyboard-shortcuts)
12. [Troubleshooting](#troubleshooting)

---

## Introduction

HackMITM is a professional-grade HTTP/HTTPS proxy tool designed for security researchers, penetration testers, and bug bounty hunters. It provides:

- **Passive Vulnerability Scanning**: Automatic detection of security issues in HTTP traffic
- **Real-time Traffic Analysis**: Live monitoring and filtering of HTTP requests/responses
- **Intercept & Modify**: Intercept requests and modify them before forwarding
- **Session Management**: Organize testing by targets or engagements
- **Report Generation**: Export findings in multiple formats

---

## Getting Started

### System Requirements

- **Operating System**: Windows 10+, macOS 10.15+, or Linux
- **Memory**: 4GB RAM minimum, 8GB recommended
- **Disk Space**: 500MB for installation, additional space for traffic logs

### Installation

#### Desktop Application

1. Download the appropriate installer for your platform
2. Run the installer and follow the prompts
3. Launch HackMITM from your applications folder

#### Command Line

```bash
# Clone the repository
git clone https://github.com/yourorg/hackmitm.git
cd hackmitm

# Build
go build -o hackmitm ./cmd/hackmitm

# Run
./hackmitm -config configs/config.json
```

### First Run

1. **Configure Proxy Settings**: Set up your browser or application to use the HackMITM proxy
   - Default proxy address: `http://localhost:8082`
   - Default API address: `http://localhost:9090`

2. **Install CA Certificate**: For HTTPS interception, install the generated CA certificate:
   - **Windows**: Double-click the certificate and install to "Trusted Root Certification Authorities"
   - **macOS**: Add to Keychain and set to "Always Trust"
   - **Linux**: Add to `/usr/local/share/ca-certificates/` and run `update-ca-certificates`

---

## Dashboard

The dashboard provides an overview of your testing session.

### Metrics Displayed

- **Total Requests**: Number of HTTP requests processed
- **Active Connections**: Current proxy connections
- **Vulnerabilities Found**: Total and by severity
- **Traffic Patterns**: Request methods, status codes, content types
- **Response Times**: Average, min, max response times

### Real-time Updates

The dashboard updates in real-time via WebSocket connections, showing:
- New vulnerabilities as they are detected
- Traffic statistics changes
- Proxy health status

---

## Traffic Analysis

### Traffic List View

The main traffic view shows all captured HTTP requests/responses.

**Columns:**
- **ID**: Unique request identifier
- **Method**: HTTP method (GET, POST, etc.)
- **URL**: Full request URL
- **Status**: HTTP response status code
- **Size**: Request/Response size
- **Time**: Response time in milliseconds
- **Timestamp**: When the request was captured

### Filtering

Use the filter bar to narrow down results:

- **Method Filter**: `GET`, `POST`, `PUT`, `DELETE`, etc.
- **Status Filter**: `2xx`, `4xx`, `5xx`, or specific codes
- **Domain Filter**: Filter by host/domain
- **Regex Filter**: Use regular expressions for complex matching
- **Saved Filters**: Save frequently used filter combinations

### Traffic Details

Click on any request to view details:

- **Request Tab**: Headers, body, parameters
- **Response Tab**: Headers, body, status
- **Vulnerabilities Tab**: Any detected issues for this request
- **Timing Tab**: Request timing breakdown

### Request Comparison

Select multiple requests and click "Compare" to see differences side-by-side.

### Traffic Marking

Mark requests with colors or tags for later reference:
- Right-click → Mark → Select color/tag
- Use marked items for report generation

---

## Vulnerability Management

### Vulnerability List

View all detected vulnerabilities in a dedicated panel.

**Columns:**
- **Name**: Vulnerability type
- **Severity**: Critical, High, Medium, Low, Info
- **URL**: Affected URL
- **Parameter**: Vulnerable parameter (if applicable)
- **Status**: Open, Confirmed, False Positive, Fixed
- **Confidence**: Detection confidence (0-100%)
- **Occurrences**: Number of times detected

### Vulnerability Details

Click a vulnerability to see:
- **Evidence**: Proof of the vulnerability
- **Request/Response**: Full HTTP exchange
- **Remediation**: Suggested fixes
- **Notes**: Add your own notes

### Status Management

Change vulnerability status:
1. Right-click vulnerability → Set Status
2. Choose: Open, Confirmed, False Positive, Fixed
3. Add notes (optional)

### Filtering

- **By Severity**: Show only critical/high/medium/low/info
- **By Status**: Show only open/confirmed/false_positive/fixed
- **By Type**: Filter by vulnerability category

---

## Passive Scanner

The passive scanner automatically analyzes traffic for security issues.

### Built-in Detection Rules

| Category | Description |
|----------|-------------|
| **SQL Injection** | Detects SQL injection patterns |
| **XSS** | Cross-site scripting detection |
| **Information Disclosure** | Sensitive data in responses |
| **Path Traversal** | Directory traversal attempts |
| **SSRF** | Server-side request forgery patterns |
| **Authentication Issues** | Auth/authorization problems |

### Rule Management

Access scanner rules from Settings → Scanner Rules:

- **Enable/Disable**: Toggle individual rules
- **Create Custom Rules**: Add your own detection patterns
- **Import Rules**: Load rules from JSON files
- **Export Rules**: Save rules for sharing

### Custom Rule Format

```json
{
  "id": "custom-001",
  "name": "API Key Exposure",
  "description": "Detects exposed API keys in responses",
  "severity": "high",
  "pattern": "api_key\\s*[:=]\\s*['\"][a-zA-Z0-9]{32}['\"]",
  "enabled": true,
  "tags": ["secrets", "api"]
}
```

### Hot Reload

Rules can be reloaded without restarting:
- Click "Reload Rules" in Scanner settings
- Or use API: `POST /api/scanner/reload`

---

## Sessions

Sessions help organize your testing by target or engagement.

### Creating a Session

1. Go to Sessions → New Session
2. Enter name and description
3. Click Create

### Session Data Isolation

Each session maintains separate:
- Traffic history
- Vulnerabilities
- Scan results
- Notes

### Switching Sessions

Use the session dropdown to switch between sessions. All views update to show only data from the selected session.

### Exporting Sessions

Export session data for backup or sharing:
1. Select session → Export
2. Choose format (JSON, ZIP archive)
3. Select destination

---

## Reports

### Generating Reports

1. Go to Reports → Generate Report
2. Configure options:
   - **Title**: Report title
   - **Session**: Which session to report on
   - **Format**: HTML, JSON, or Markdown
   - **Severities**: Include all or select specific levels
   - **Statuses**: Include all or select specific statuses
3. Click Generate
4. Preview and download

### Report Contents

Reports include:
- Executive summary with statistics
- Vulnerability breakdown by severity
- Detailed vulnerability entries
- Evidence and remediation suggestions
- Request/response details

### Report Formats

| Format | Use Case |
|--------|----------|
| **HTML** | Professional presentation, easy sharing |
| **JSON** | Integration with other tools |
| **Markdown** | Documentation, GitHub issues |

---

## Intercept Mode

### Enabling Intercept

1. Toggle "Intercept" button in toolbar
2. Requests will be held before forwarding

### Intercepted Requests

When intercept is enabled:
1. Requests appear in the Intercept panel
2. Review request details
3. Choose action:
   - **Forward**: Send request as-is
   - **Drop**: Discard the request
   - **Modify**: Edit before forwarding

### Modifying Requests

In the modification dialog:
- Change HTTP method
- Edit URL
- Modify headers
- Edit request body
- Add/remove parameters

### Automatic Rules

Set up automatic interception rules:
1. Settings → Intercept Rules
2. Define conditions (URL patterns, methods, etc.)
3. Set actions (intercept, drop, modify)

---

## Repeater

The Repeater allows you to manually craft and send HTTP requests.

### Creating a Request

1. Go to Repeater tab
2. Enter:
   - **Method**: HTTP method
   - **URL**: Target URL
   - **Headers**: One per line (Name: Value)
   - **Body**: Request body (for POST/PUT/PATCH)

### Sending Requests

- **Send**: Send the request
- **Send & Keep Open**: Keep editing after sending

### Response View

View the response:
- **Headers**: Response headers
- **Body**: Response content
- **Status**: HTTP status code and message
- **Time**: Response time

### History

Access previous requests from the history panel. Re-send with modifications.

### Import from Traffic

Right-click any traffic item → Send to Repeater

---

## Keyboard Shortcuts

| Shortcut | Action |
|----------|--------|
| `Ctrl/Cmd + F` | Focus filter bar |
| `Ctrl/Cmd + I` | Toggle intercept mode |
| `Ctrl/Cmd + R` | Send to repeater |
| `Ctrl/Cmd + S` | Save current state |
| `Ctrl/Cmd + E` | Export |
| `Esc` | Clear selection |
| `Space` | Toggle request selection |
| `Enter` | Open request details |
| `Delete` | Delete selected items |
| `F5` | Refresh traffic list |

---

## Troubleshooting

### HTTPS Not Working

**Symptoms**: HTTPS sites show connection errors or blank pages

**Solutions**:
1. Ensure CA certificate is installed
2. Check certificate location in config
3. Restart browser after certificate installation

### Proxy Not Capturing Traffic

**Symptoms**: No traffic appears in HackMITM

**Solutions**:
1. Verify browser proxy settings point to HackMITM
2. Check firewall isn't blocking connections
3. Ensure HackMITM is running

### High Memory Usage

**Symptoms**: Application becomes slow

**Solutions**:
1. Clear old traffic: Traffic → Clear
2. Reduce history retention in settings
3. Archive old sessions

### WebSocket Connection Errors

**Symptoms**: Dashboard not updating

**Solutions**:
1. Check WebSocket port (default: 9090)
2. Verify no proxy blocking WebSocket
3. Restart application

### Scan Rules Not Loading

**Symptoms**: No vulnerabilities detected

**Solutions**:
1. Check rules directory exists
2. Verify rule file format (YAML/JSON)
3. Use "Reload Rules" button

---

## Getting Help

- **Documentation**: [https://docs.hackmitm.io](https://docs.hackmitm.io)
- **GitHub Issues**: [https://github.com/yourorg/hackmitm/issues](https://github.com/yourorg/hackmitm/issues)
- **Community Discord**: [https://discord.gg/hackmitm](https://discord.gg/hackmitm)

---

## License

HackMITM is released under the MIT License. See LICENSE file for details.
