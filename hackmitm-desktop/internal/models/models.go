package models

import "time"

// TrafficItem represents a single HTTP traffic entry
type TrafficItem struct {
	ID              string            `json:"id"`
	Timestamp       string            `json:"timestamp"`
	Method          string            `json:"method"`
	URL             string            `json:"url"`
	Host            string            `json:"host"`
	Path            string            `json:"path"`
	StatusCode      int               `json:"statusCode"`
	ContentType     string            `json:"contentType"`
	RequestSize     int64             `json:"requestSize"`
	ResponseSize    int64             `json:"responseSize"`
	Duration        int64             `json:"duration"` // milliseconds
	RequestHeaders  map[string]string `json:"requestHeaders"`
	ResponseHeaders map[string]string `json:"responseHeaders"`
	RequestBody     string            `json:"requestBody"`
	ResponseBody    string            `json:"responseBody"`
	ClientIP        string            `json:"clientIP"`
	Protocol        string            `json:"protocol"`
	Intercepted     bool              `json:"intercepted"`
}

// FingerprintResult represents fingerprint identification result
type FingerprintResult struct {
	URL          string   `json:"url"`
	Fingerprints []string `json:"fingerprints"`
	Confidence   float64  `json:"confidence"`
	ProcessTime  int64    `json:"processTime"` // milliseconds
	Title        string   `json:"title"`
	StatusCode   int      `json:"statusCode"`
	Timestamp    string   `json:"timestamp"`
}

// ProxyStatus represents the proxy server status
type ProxyStatus struct {
	Running           bool  `json:"running"`
	Port              int   `json:"port"`
	InterceptMode     bool  `json:"interceptMode"`
	ActiveConnections int64 `json:"activeConnections"`
	TotalRequests     int64 `json:"totalRequests"`
	Uptime            int64 `json:"uptime"` // seconds
}

// DashboardMetrics represents dashboard metrics
type DashboardMetrics struct {
	QPS              float64 `json:"qps"`
	AvgResponseTime  float64 `json:"avgResponseTime"`
	ActiveConns      int64   `json:"activeConnections"`
	TotalRequests    int64   `json:"totalRequests"`
	TotalBytesIn     int64   `json:"totalBytesIn"`
	TotalBytesOut    int64   `json:"totalBytesOut"`
	ErrorRate        float64 `json:"errorRate"`
	Uptime           int64   `json:"uptime"`
}

// Target represents a target host
type Target struct {
	ID           string   `json:"id"`
	Host         string   `json:"host"`
	Port         int      `json:"port"`
	Protocol     string   `json:"protocol"`
	Title        string   `json:"title,omitempty"`
	Technologies []string `json:"technologies,omitempty"`
	InScope      bool     `json:"inScope"`
	LastAccessed string   `json:"lastAccessed"`
	RequestCount int      `json:"requestCount"`
}

// RepeaterRequest represents a repeater request
type RepeaterRequest struct {
	ID        string            `json:"id"`
	Name      string            `json:"name"`
	Method    string            `json:"method"`
	URL       string            `json:"url"`
	Headers   map[string]string `json:"headers"`
	Body      string            `json:"body"`
	CreatedAt time.Time         `json:"createdAt"`
	UpdatedAt time.Time         `json:"updatedAt"`
}

// RepeaterResponse represents a repeater response
type RepeaterResponse struct {
	StatusCode    int               `json:"statusCode"`
	StatusText    string            `json:"statusText"`
	Headers       map[string]string `json:"headers"`
	Body          string            `json:"body"`
	ResponseTime  int64             `json:"responseTime"` // milliseconds
	ContentLength int64             `json:"contentLength"`
}

// ConnectionConfig represents connection configuration
type ConnectionConfig struct {
	Host        string `json:"host"`
	Port        int    `json:"port"`
	APIEndpoint string `json:"apiEndpoint"`
}

// InterceptAction represents intercept action type
type InterceptAction string

const (
	InterceptForward InterceptAction = "forward"
	InterceptDrop    InterceptAction = "drop"
	InterceptModify  InterceptAction = "modify"
)
