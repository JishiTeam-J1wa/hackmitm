// Package plugin 插件工具库
package plugin

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// RequestUtils 请求工具
type RequestUtils struct{}

// NewRequestUtils 创建请求工具
func NewRequestUtils() *RequestUtils {
	return &RequestUtils{}
}

// GetClientIP 获取客户端IP
func (ru *RequestUtils) GetClientIP(req *http.Request) string {
	// 检查 X-Forwarded-For 头
	if xff := req.Header.Get("X-Forwarded-For"); xff != "" {
		ips := strings.Split(xff, ",")
		if len(ips) > 0 {
			return strings.TrimSpace(ips[0])
		}
	}

	// 检查 X-Real-IP 头
	if xri := req.Header.Get("X-Real-IP"); xri != "" {
		return strings.TrimSpace(xri)
	}

	// 使用 RemoteAddr
	if req.RemoteAddr != "" {
		if host, _, err := net.SplitHostPort(req.RemoteAddr); err == nil {
			return host
		}
		return req.RemoteAddr
	}

	return "unknown"
}

// GetUserAgent 获取用户代理
func (ru *RequestUtils) GetUserAgent(req *http.Request) string {
	return req.Header.Get("User-Agent")
}

// GetContentType 获取内容类型
func (ru *RequestUtils) GetContentType(req *http.Request) string {
	return req.Header.Get("Content-Type")
}

// IsJSON 检查是否为JSON请求
func (ru *RequestUtils) IsJSON(req *http.Request) bool {
	contentType := ru.GetContentType(req)
	return strings.Contains(strings.ToLower(contentType), "application/json")
}

// IsXML 检查是否为XML请求
func (ru *RequestUtils) IsXML(req *http.Request) bool {
	contentType := ru.GetContentType(req)
	return strings.Contains(strings.ToLower(contentType), "application/xml") ||
		strings.Contains(strings.ToLower(contentType), "text/xml")
}

// IsFormData 检查是否为表单数据
func (ru *RequestUtils) IsFormData(req *http.Request) bool {
	contentType := ru.GetContentType(req)
	return strings.Contains(strings.ToLower(contentType), "application/x-www-form-urlencoded") ||
		strings.Contains(strings.ToLower(contentType), "multipart/form-data")
}

// GetQueryParam 获取查询参数
func (ru *RequestUtils) GetQueryParam(req *http.Request, key string) string {
	return req.URL.Query().Get(key)
}

// GetQueryParams 获取所有查询参数
func (ru *RequestUtils) GetQueryParams(req *http.Request) map[string][]string {
	return req.URL.Query()
}

// GetHeader 获取请求头
func (ru *RequestUtils) GetHeader(req *http.Request, key string) string {
	return req.Header.Get(key)
}

// GetHeaders 获取所有请求头
func (ru *RequestUtils) GetHeaders(req *http.Request) map[string][]string {
	return req.Header
}

// ReadBody 读取请求体
func (ru *RequestUtils) ReadBody(req *http.Request) ([]byte, error) {
	if req.Body == nil {
		return nil, nil
	}

	body, err := io.ReadAll(req.Body)
	if err != nil {
		return nil, err
	}

	// 重置请求体以便后续读取
	req.Body = io.NopCloser(bytes.NewBuffer(body))

	return body, nil
}

// ParseJSONBody 解析JSON请求体
func (ru *RequestUtils) ParseJSONBody(req *http.Request, v interface{}) error {
	body, err := ru.ReadBody(req)
	if err != nil {
		return err
	}

	return json.Unmarshal(body, v)
}

// ResponseUtils 响应工具
type ResponseUtils struct{}

// NewResponseUtils 创建响应工具
func NewResponseUtils() *ResponseUtils {
	return &ResponseUtils{}
}

// SetHeader 设置响应头
func (ru *ResponseUtils) SetHeader(resp *http.Response, key, value string) {
	if resp.Header == nil {
		resp.Header = make(http.Header)
	}
	resp.Header.Set(key, value)
}

// AddHeader 添加响应头
func (ru *ResponseUtils) AddHeader(resp *http.Response, key, value string) {
	if resp.Header == nil {
		resp.Header = make(http.Header)
	}
	resp.Header.Add(key, value)
}

// GetHeader 获取响应头
func (ru *ResponseUtils) GetHeader(resp *http.Response, key string) string {
	return resp.Header.Get(key)
}

// SetContentType 设置内容类型
func (ru *ResponseUtils) SetContentType(resp *http.Response, contentType string) {
	ru.SetHeader(resp, "Content-Type", contentType)
}

// ReadBody 读取响应体
func (ru *ResponseUtils) ReadBody(resp *http.Response) ([]byte, error) {
	if resp.Body == nil {
		return nil, nil
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	// 重置响应体以便后续读取
	resp.Body = io.NopCloser(bytes.NewBuffer(body))

	return body, nil
}

// SetJSONBody 设置JSON响应体
func (ru *ResponseUtils) SetJSONBody(resp *http.Response, v interface{}) error {
	body, err := json.Marshal(v)
	if err != nil {
		return err
	}

	resp.Body = io.NopCloser(bytes.NewBuffer(body))
	resp.ContentLength = int64(len(body))
	ru.SetContentType(resp, "application/json")

	return nil
}

// SecurityUtils 安全工具
type SecurityUtils struct{}

// NewSecurityUtils 创建安全工具
func NewSecurityUtils() *SecurityUtils {
	return &SecurityUtils{}
}

// IsSQLInjection 检查SQL注入
func (su *SecurityUtils) IsSQLInjection(input string) bool {
	patterns := []string{
		`(?i)(union|select|insert|update|delete|drop|create|alter|exec|execute)\s+`,
		`(?i)(\s|^)(or|and)\s+\d+\s*=\s*\d+`,
		`(?i)'.*--`,
		`(?i)\/\*.*\*\/`,
		`(?i)(char|ascii|substring|length|user|database|version)\s*\(`,
	}

	for _, pattern := range patterns {
		if matched, _ := regexp.MatchString(pattern, input); matched {
			return true
		}
	}

	return false
}

// IsXSS 检查XSS攻击
func (su *SecurityUtils) IsXSS(input string) bool {
	patterns := []string{
		`(?i)<script[^>]*>.*?</script>`,
		`(?i)<iframe[^>]*>.*?</iframe>`,
		`(?i)<object[^>]*>.*?</object>`,
		`(?i)<embed[^>]*>.*?</embed>`,
		`(?i)<link[^>]*>`,
		`(?i)<meta[^>]*>`,
		`(?i)javascript:`,
		`(?i)vbscript:`,
		`(?i)on\w+\s*=`,
		`(?i)expression\s*\(`,
	}

	for _, pattern := range patterns {
		if matched, _ := regexp.MatchString(pattern, input); matched {
			return true
		}
	}

	return false
}

// IsPathTraversal 检查路径遍历
func (su *SecurityUtils) IsPathTraversal(input string) bool {
	patterns := []string{
		`\.\.\/`,
		`\.\.\\`,
		`%2e%2e%2f`,
		`%2e%2e%5c`,
		`%252e%252e%252f`,
		`%252e%252e%255c`,
	}

	for _, pattern := range patterns {
		if matched, _ := regexp.MatchString(pattern, input); matched {
			return true
		}
	}

	return false
}

// IsCommandInjection 检查命令注入
func (su *SecurityUtils) IsCommandInjection(input string) bool {
	patterns := []string{
		`(?i)[;&|` + "`" + `]\s*(cat|ls|pwd|whoami|id|uname|ps|netstat|ifconfig|ping|nslookup|dig|curl|wget|nc|telnet|ssh|ftp|sudo|su|chmod|chown|rm|mv|cp|mkdir|rmdir|kill|killall|pkill|service|systemctl|crontab|at|batch|nohup|screen|tmux|history|env|export|set|unset|alias|which|whereis|locate|find|grep|awk|sed|sort|uniq|head|tail|more|less|vi|vim|nano|emacs|tar|gzip|gunzip|zip|unzip|7z|rar|unrar)`,
		`(?i)\$\(.*\)`,
		"(?i)`.*`",
		`(?i)&&|\|\|`,
	}

	for _, pattern := range patterns {
		if matched, _ := regexp.MatchString(pattern, input); matched {
			return true
		}
	}

	return false
}

// SanitizeInput 清理输入
func (su *SecurityUtils) SanitizeInput(input string) string {
	// HTML转义
	replacements := map[string]string{
		"<":  "&lt;",
		">":  "&gt;",
		"&":  "&amp;",
		"\"": "&quot;",
		"'":  "&#39;",
	}

	result := input
	for old, new := range replacements {
		result = strings.ReplaceAll(result, old, new)
	}

	return result
}

// ValidateEmail 验证邮箱
func (su *SecurityUtils) ValidateEmail(email string) bool {
	pattern := `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`
	matched, _ := regexp.MatchString(pattern, email)
	return matched
}

// ValidateURL 验证URL
func (su *SecurityUtils) ValidateURL(urlStr string) bool {
	_, err := url.ParseRequestURI(urlStr)
	return err == nil
}

// ValidateIP 验证IP地址
func (su *SecurityUtils) ValidateIP(ip string) bool {
	pattern := `^((25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\.){3}(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)$`
	matched, _ := regexp.MatchString(pattern, ip)
	return matched
}

// ConversionUtils 转换工具
type ConversionUtils struct{}

// NewConversionUtils 创建转换工具
func NewConversionUtils() *ConversionUtils {
	return &ConversionUtils{}
}

// ToString 转换为字符串
func (cu *ConversionUtils) ToString(v interface{}) string {
	switch val := v.(type) {
	case string:
		return val
	case []byte:
		return string(val)
	case int:
		return strconv.Itoa(val)
	case int32:
		return strconv.FormatInt(int64(val), 10)
	case int64:
		return strconv.FormatInt(val, 10)
	case float32:
		return strconv.FormatFloat(float64(val), 'f', -1, 32)
	case float64:
		return strconv.FormatFloat(val, 'f', -1, 64)
	case bool:
		return strconv.FormatBool(val)
	default:
		return fmt.Sprintf("%v", v)
	}
}

// ToInt 转换为整数
func (cu *ConversionUtils) ToInt(v interface{}) (int, error) {
	switch val := v.(type) {
	case int:
		return val, nil
	case int32:
		return int(val), nil
	case int64:
		return int(val), nil
	case float32:
		return int(val), nil
	case float64:
		return int(val), nil
	case string:
		return strconv.Atoi(val)
	case []byte:
		return strconv.Atoi(string(val))
	default:
		return 0, fmt.Errorf("无法转换为整数: %T", v)
	}
}

// ToBool 转换为布尔值
func (cu *ConversionUtils) ToBool(v interface{}) (bool, error) {
	switch val := v.(type) {
	case bool:
		return val, nil
	case string:
		return strconv.ParseBool(val)
	case []byte:
		return strconv.ParseBool(string(val))
	case int:
		return val != 0, nil
	case int32:
		return val != 0, nil
	case int64:
		return val != 0, nil
	case float32:
		return val != 0, nil
	case float64:
		return val != 0, nil
	default:
		return false, fmt.Errorf("无法转换为布尔值: %T", v)
	}
}

// ToFloat 转换为浮点数
func (cu *ConversionUtils) ToFloat(v interface{}) (float64, error) {
	switch val := v.(type) {
	case float32:
		return float64(val), nil
	case float64:
		return val, nil
	case int:
		return float64(val), nil
	case int32:
		return float64(val), nil
	case int64:
		return float64(val), nil
	case string:
		return strconv.ParseFloat(val, 64)
	case []byte:
		return strconv.ParseFloat(string(val), 64)
	default:
		return 0, fmt.Errorf("无法转换为浮点数: %T", v)
	}
}

// TimeUtils 时间工具
type TimeUtils struct{}

// NewTimeUtils 创建时间工具
func NewTimeUtils() *TimeUtils {
	return &TimeUtils{}
}

// Now 获取当前时间
func (tu *TimeUtils) Now() time.Time {
	return time.Now()
}

// Format 格式化时间
func (tu *TimeUtils) Format(t time.Time, layout string) string {
	return t.Format(layout)
}

// Parse 解析时间
func (tu *TimeUtils) Parse(layout, value string) (time.Time, error) {
	return time.Parse(layout, value)
}

// Sleep 睡眠
func (tu *TimeUtils) Sleep(duration time.Duration) {
	time.Sleep(duration)
}

// Timeout 超时控制
func (tu *TimeUtils) Timeout(ctx context.Context, duration time.Duration, fn func() error) error {
	done := make(chan error, 1)

	go func() {
		done <- fn()
	}()

	select {
	case err := <-done:
		return err
	case <-time.After(duration):
		return fmt.Errorf("操作超时")
	case <-ctx.Done():
		return ctx.Err()
	}
}

// PluginUtils 插件工具集合
type PluginUtils struct {
	Request    *RequestUtils
	Response   *ResponseUtils
	Security   *SecurityUtils
	Conversion *ConversionUtils
	Time       *TimeUtils
}

// NewPluginUtils 创建插件工具集合
func NewPluginUtils() *PluginUtils {
	return &PluginUtils{
		Request:    NewRequestUtils(),
		Response:   NewResponseUtils(),
		Security:   NewSecurityUtils(),
		Conversion: NewConversionUtils(),
		Time:       NewTimeUtils(),
	}
}
