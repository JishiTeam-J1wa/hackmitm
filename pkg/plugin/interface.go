// Package plugin 提供动态插件加载功能
// Package plugin provides dynamic plugin loading functionality
package plugin

import (
	"context"
	"net/http"
	"time"
)

// Plugin 插件基础接口
// Plugin base interface for all plugins
type Plugin interface {
	// Name 返回插件名称
	Name() string
	// Version 返回插件版本
	Version() string
	// Description 返回插件描述
	Description() string
	// Initialize 初始化插件
	Initialize(config map[string]interface{}) error
	// Start 启动插件
	Start(ctx context.Context) error
	// Stop 停止插件
	Stop(ctx context.Context) error
	// Cleanup 清理插件资源
	Cleanup() error
}

// RequestPlugin 请求处理插件接口
// RequestPlugin interface for request processing plugins
type RequestPlugin interface {
	Plugin
	// ProcessRequest 处理HTTP请求
	ProcessRequest(req *http.Request, ctx *RequestContext) error
	// Priority 返回插件优先级（数字越小优先级越高）
	Priority() int
}

// ResponsePlugin 响应处理插件接口
// ResponsePlugin interface for response processing plugins
type ResponsePlugin interface {
	Plugin
	// ProcessResponse 处理HTTP响应
	ProcessResponse(resp *http.Response, req *http.Request, ctx *ResponseContext) error
	// Priority 返回插件优先级
	Priority() int
}

// FilterPlugin 过滤插件接口
// FilterPlugin interface for filtering plugins
type FilterPlugin interface {
	Plugin
	// ShouldAllow 判断是否允许请求通过
	ShouldAllow(req *http.Request, ctx *FilterContext) (bool, error)
	// Priority 返回插件优先级
	Priority() int
}

// LoggerPlugin 日志插件接口
// LoggerPlugin interface for logging plugins
type LoggerPlugin interface {
	Plugin
	// LogRequest 记录请求日志
	LogRequest(req *http.Request, ctx *RequestContext) error
	// LogResponse 记录响应日志
	LogResponse(resp *http.Response, req *http.Request, ctx *ResponseContext) error
	// LogError 记录错误日志
	LogError(err error, req *http.Request, ctx *ErrorContext) error
}

// ModifierPlugin 修改器插件接口
// ModifierPlugin interface for modifier plugins
type ModifierPlugin interface {
	Plugin
	// ModifyRequest 修改请求
	ModifyRequest(req *http.Request, ctx *RequestContext) error
	// ModifyResponse 修改响应
	ModifyResponse(resp *http.Response, req *http.Request, ctx *ResponseContext) error
	// Priority 返回插件优先级
	Priority() int
}

// AnalyticsPlugin 分析插件接口
// AnalyticsPlugin interface for analytics plugins
type AnalyticsPlugin interface {
	Plugin
	// AnalyzeRequest 分析请求
	AnalyzeRequest(req *http.Request, ctx *RequestContext) (*AnalysisResult, error)
	// AnalyzeResponse 分析响应
	AnalyzeResponse(resp *http.Response, req *http.Request, ctx *ResponseContext) (*AnalysisResult, error)
	// GetStatistics 获取统计信息
	GetStatistics() map[string]interface{}
}

// RequestContext 请求上下文
type RequestContext struct {
	StartTime time.Time
	ClientIP  string
	UserAgent string
	Method    string
	URL       string
	Headers   map[string]string
	Body      []byte
	Metadata  map[string]interface{}
}

// ResponseContext 响应上下文
type ResponseContext struct {
	StatusCode int
	Headers    map[string]string
	Body       []byte
	Size       int64
	Duration   time.Duration
	Metadata   map[string]interface{}
}

// FilterContext 过滤上下文
type FilterContext struct {
	ClientIP     string
	UserAgent    string
	RequestCount int64
	LastRequest  time.Time
	Metadata     map[string]interface{}
}

// ErrorContext 错误上下文
type ErrorContext struct {
	ErrorType    string
	ErrorMessage string
	StackTrace   string
	Timestamp    time.Time
	Metadata     map[string]interface{}
}

// AnalysisResult 分析结果
type AnalysisResult struct {
	Threat      bool                   `json:"threat"`
	ThreatLevel string                 `json:"threat_level"`
	Description string                 `json:"description"`
	Confidence  float64                `json:"confidence"`
	Metadata    map[string]interface{} `json:"metadata"`
	Timestamp   time.Time              `json:"timestamp"`
}

// PluginInfo 插件信息
type PluginInfo struct {
	Name        string                 `json:"name"`
	Version     string                 `json:"version"`
	Description string                 `json:"description"`
	Author      string                 `json:"author"`
	License     string                 `json:"license"`
	Type        string                 `json:"type"`
	Config      map[string]interface{} `json:"config"`
	Status      string                 `json:"status"`
	LoadTime    time.Time              `json:"load_time"`
	Path        string                 `json:"path"`
}

// PluginConfig 插件配置
type PluginConfig struct {
	Name     string                 `json:"name"`
	Enabled  bool                   `json:"enabled"`
	Path     string                 `json:"path"`
	Config   map[string]interface{} `json:"config"`
	Priority int                    `json:"priority"`
}

// PluginStatus 插件状态
type PluginStatus string

const (
	StatusLoaded      PluginStatus = "loaded"
	StatusInitialized PluginStatus = "initialized"
	StatusStarted     PluginStatus = "started"
	StatusStopped     PluginStatus = "stopped"
	StatusError       PluginStatus = "error"
	StatusUnloaded    PluginStatus = "unloaded"
)

// PluginType 插件类型
type PluginType string

const (
	TypeRequest   PluginType = "request"
	TypeResponse  PluginType = "response"
	TypeFilter    PluginType = "filter"
	TypeLogger    PluginType = "logger"
	TypeModifier  PluginType = "modifier"
	TypeAnalytics PluginType = "analytics"
)

// LoaderFunc 插件加载器函数类型
type LoaderFunc func() Plugin

// Factory 插件工厂函数类型
type Factory func(config map[string]interface{}) (Plugin, error)
