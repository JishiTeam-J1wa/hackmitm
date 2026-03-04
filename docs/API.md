# HackMITM API Documentation

## Overview

HackMITM provides a RESTful API for managing the proxy server, traffic analysis, vulnerability scanning, and report generation.

**Base URL:** `http://localhost:9090`

---

## Table of Contents

1. [Traffic API](#traffic-api)
2. [Fingerprint API](#fingerprint-api)
3. [Vulnerability API](#vulnerability-api)
4. [Scanner API](#scanner-api)
5. [Session API](#session-api)
6. [Report API](#report-api)
7. [WebSocket API](#websocket-api)

---

## Traffic API

### List Traffic

```
GET /api/traffic
```

**Query Parameters:**
- `limit` (int): Maximum number of entries to return (default: 1000)

**Response:**
```json
[
  {
    "id": 1,
    "timestamp": "2024-01-15 10:30:45",
    "method": "GET",
    "url": "https://example.com/api/users",
    "host": "example.com",
    "path": "/api/users",
    "statusCode": 200,
    "contentType": "application/json",
    "requestSize": 256,
    "responseSize": 1024,
    "duration": 150,
    "clientIP": "192.168.1.100",
    "protocol": "HTTP/1.1",
    "fingerprint": "nginx"
  }
]
```

### Clear Traffic

```
DELETE /api/traffic/clear
```

**Response:**
```json
{
  "success": true,
  "message": "Traffic cleared"
}
```

### Get Traffic History

```
GET /api/traffic/history
```

**Query Parameters:**
- `limit` (int): Maximum entries (default: 100)
- `offset` (int): Offset for pagination (default: 0)

### Get Traffic Statistics

```
GET /api/traffic/stats
```

**Response:**
```json
{
  "memory_count": 150,
  "db_count": 5000,
  "db_size": 1048576,
  "db_path": "/path/to/hackmitm.db"
}
```

---

## Fingerprint API

### Get Fingerprint Statistics

```
GET /fingerprint/stats
```

**Response:**
```json
{
  "enabled": true,
  "total_identified": 150,
  "by_type": {
    "nginx": 50,
    "apache": 30,
    "express": 25,
    "django": 20
  }
}
```

### Get Fingerprint History

```
GET /api/fingerprint/history
```

**Query Parameters:**
- `limit` (int): Maximum entries (default: 100)

### Identify Fingerprint

```
POST /fingerprint/identify
```

**Request Body:**
```json
{
  "url": "https://example.com"
}
```

**Response:**
```json
{
  "url": "https://example.com",
  "fingerprint": ["nginx", "php"],
  "confidence": 0.85,
  "message": "Fingerprint identified successfully"
}
```

---

## Vulnerability API

### List Vulnerabilities

```
GET /api/vulns
```

**Query Parameters:**
- `session_id` (string): Filter by session ID
- `severity` (string): Filter by severity (critical, high, medium, low, info)
- `status` (string): Filter by status (open, confirmed, false_positive, fixed)
- `limit` (int): Maximum entries (default: 100)
- `offset` (int): Offset for pagination

**Response:**
```json
{
  "data": [
    {
      "id": "vuln-001",
      "session_id": "sess-001",
      "traffic_id": 123,
      "rule_id": "sqli-001",
      "name": "SQL Injection",
      "severity": "high",
      "confidence": 0.95,
      "url": "https://example.com/api/users?id=1",
      "parameter": "id",
      "status": "open",
      "occurrences": 3,
      "first_seen": "2024-01-15T10:30:00Z",
      "last_seen": "2024-01-15T11:00:00Z"
    }
  ],
  "total": 25,
  "limit": 100,
  "offset": 0
}
```

### Get Vulnerability

```
GET /api/vulns/{id}
```

**Response:**
```json
{
  "id": "vuln-001",
  "session_id": "sess-001",
  "name": "SQL Injection",
  "description": "Potential SQL injection vulnerability detected",
  "severity": "high",
  "confidence": 0.95,
  "url": "https://example.com/api/users",
  "parameter": "id",
  "evidence": "' OR '1'='1",
  "remediation": "Use parameterized queries or prepared statements",
  "status": "open",
  "occurrences": 3,
  "first_seen": "2024-01-15T10:30:00Z",
  "last_seen": "2024-01-15T11:00:00Z",
  "request_method": "GET",
  "request_url": "https://example.com/api/users?id=1'+OR+'1'='1",
  "request_headers": "{\"User-Agent\": \"Mozilla/5.0\"}",
  "response_status": 500,
  "traffic": {
    "id": 123,
    "method": "GET",
    "url": "https://example.com/api/users",
    "status_code": 500
  }
}
```

### Update Vulnerability Status

```
PATCH /api/vulns/{id}/status
```

**Request Body:**
```json
{
  "status": "confirmed",
  "notes": "Verified by security team"
}
```

**Valid Status Values:**
- `open` - New vulnerability
- `confirmed` - Confirmed as valid
- `false_positive` - Marked as false positive
- `fixed` - Issue has been fixed

### Delete Vulnerability

```
DELETE /api/vulns/{id}
```

**Response:**
```json
{
  "id": "vuln-001",
  "deleted": true
}
```

### Get Vulnerability Statistics

```
GET /api/vulns/stats
```

**Query Parameters:**
- `session_id` (string): Filter by session ID

**Response:**
```json
{
  "total": 50,
  "critical": 5,
  "high": 15,
  "medium": 20,
  "low": 8,
  "info": 2,
  "open": 35,
  "confirmed": 10,
  "fixed": 3,
  "false_positive": 2
}
```

---

## Scanner API

### List Scanner Rules

```
GET /api/scanner/rules
```

**Response:**
```json
{
  "data": [
    {
      "id": "sqli-001",
      "name": "SQL Injection Detector",
      "description": "Detects SQL injection patterns in request parameters",
      "severity": "high",
      "enabled": true,
      "priority": 1,
      "tags": ["injection", "sql", "database"]
    }
  ],
  "total": 25
}
```

### Get Scanner Rule

```
GET /api/scanner/rules/{id}
```

**Response:**
```json
{
  "id": "sqli-001",
  "name": "SQL Injection Detector",
  "description": "Detects SQL injection patterns",
  "severity": "high",
  "enabled": true,
  "priority": 1,
  "tags": ["injection", "sql"],
  "remediation": "Use parameterized queries"
}
```

### Update Rule (Enable/Disable)

```
PATCH /api/scanner/rules/{id}
```

**Request Body:**
```json
{
  "enabled": false
}
```

### Create Custom Rule

```
POST /api/scanner/rules
```

**Request Body:**
```json
{
  "id": "custom-001",
  "name": "Custom XSS Detector",
  "description": "Detects custom XSS patterns",
  "severity": "medium",
  "pattern": "<script[^>]*>",
  "tags": ["xss", "custom"]
}
```

### Reload Rules

```
POST /api/scanner/reload
```

**Response:**
```json
{
  "message": "Rules reloaded successfully",
  "count": 30
}
```

---

## Session API

### List Sessions

```
GET /api/sessions
```

**Query Parameters:**
- `limit` (int): Maximum entries (default: 100)

**Response:**
```json
{
  "data": [
    {
      "id": "sess-001",
      "name": "Bug Bounty Target A",
      "description": "Testing target A for bug bounty",
      "created_at": "2024-01-15T10:00:00Z",
      "updated_at": "2024-01-15T15:00:00Z",
      "traffic_count": 1500,
      "vuln_count": 25
    }
  ]
}
```

### Create Session

```
POST /api/sessions
```

**Request Body:**
```json
{
  "name": "New Session",
  "description": "Session description"
}
```

### Get Session

```
GET /api/sessions/{id}
```

**Response:**
```json
{
  "id": "sess-001",
  "name": "Bug Bounty Target A",
  "description": "Testing target A",
  "created_at": "2024-01-15T10:00:00Z",
  "updated_at": "2024-01-15T15:00:00Z",
  "traffic_count": 1500,
  "vuln_count": 25,
  "vuln_stats": {
    "critical": 2,
    "high": 8,
    "medium": 10,
    "low": 5
  }
}
```

### Update Session

```
PUT /api/sessions/{id}
```

**Request Body:**
```json
{
  "name": "Updated Name",
  "description": "Updated description"
}
```

### Delete Session

```
DELETE /api/sessions/{id}
```

---

## Report API

### Generate Report

```
POST /api/reports/generate
```

**Request Body:**
```json
{
  "session_id": "sess-001",
  "title": "Security Assessment Report",
  "format": "html",
  "severity": ["critical", "high"],
  "status": ["open", "confirmed"]
}
```

**Supported Formats:**
- `json` - JSON format
- `html` - HTML format (styled)
- `markdown` or `md` - Markdown format

**Response:**
Returns the report content with appropriate Content-Type header.

### List Reports

```
GET /api/reports
```

**Response:**
```json
{
  "data": [],
  "message": "Reports are generated on-demand. Use POST /api/reports/generate to create a report."
}
```

---

## WebSocket API

Connect to the WebSocket endpoint for real-time updates.

**Endpoint:** `ws://localhost:9090/ws`

### Message Format

All messages follow this structure:

```json
{
  "type": "data|ack|error",
  "action": "subscribe|unsubscribe|update|create",
  "channel": "traffic|vulns|intercept",
  "data": {},
  "timestamp": "2024-01-15T10:30:00Z",
  "client_id": "client-001"
}
```

### Subscribing to Channels

**Subscribe:**
```json
{
  "action": "subscribe",
  "channel": "traffic"
}
```

**Unsubscribe:**
```json
{
  "action": "unsubscribe",
  "channel": "traffic"
}
```

### Available Channels

| Channel | Description |
|---------|-------------|
| `traffic` | Real-time HTTP traffic updates |
| `vulns` | New vulnerability detections |
| `intercept` | Intercepted requests for approval |

### Example: Traffic Update

```json
{
  "type": "data",
  "action": "create",
  "channel": "traffic",
  "data": {
    "id": 1234,
    "method": "GET",
    "url": "https://example.com/api/data",
    "statusCode": 200,
    "duration": 150
  },
  "timestamp": "2024-01-15T10:30:00Z"
}
```

### Example: Vulnerability Detection

```json
{
  "type": "data",
  "action": "create",
  "channel": "vulns",
  "data": {
    "id": "vuln-001",
    "name": "SQL Injection",
    "severity": "high",
    "url": "https://example.com/api/users",
    "parameter": "id"
  },
  "timestamp": "2024-01-15T10:30:00Z"
}
```

---

## Error Responses

All API errors follow this format:

```json
{
  "error": "Error message description",
  "status": 400
}
```

### Common HTTP Status Codes

| Code | Description |
|------|-------------|
| 200 | Success |
| 201 | Created |
| 400 | Bad Request - Invalid input |
| 404 | Not Found |
| 500 | Internal Server Error |

---

## Rate Limiting

API endpoints are rate-limited to 100 requests per second per client. Exceeding this limit will result in HTTP 429 responses.

---

## Authentication

Currently, the API does not require authentication when running locally. For production deployments, consider adding authentication middleware.
