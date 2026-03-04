package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"hackmitm/pkg/plugin"
)

// Handler 请求处理器
type Handler struct {
	plugin *RequestLoggerPlugin
}

// NewHandler 创建处理器
func NewHandler(p *RequestLoggerPlugin) *Handler {
	return &Handler{
		plugin: p,
	}
}

// HandleRequest 处理请求
func (h *Handler) HandleRequest(req *http.Request, ctx *plugin.RequestContext) error {
	// 构建日志消息
	logMsg := h.buildLogMessage(req, ctx)

	// 根据日志级别输出
	switch strings.ToLower(h.plugin.Config.LogLevel) {
	case "debug":
		h.plugin.logger.Debug(logMsg)
	case "warn":
		h.plugin.logger.Warn(logMsg)
	case "error":
		h.plugin.logger.Error(logMsg)
	default:
		h.plugin.logger.Info(logMsg)
	}

	return nil
}

// buildLogMessage 构建日志消息
func (h *Handler) buildLogMessage(req *http.Request, ctx *plugin.RequestContext) string {
	switch h.plugin.Config.LogFormat {
	case "simple":
		return h.buildSimpleLogMessage(req, ctx)
	case "json":
		return h.buildJSONLogMessage(req, ctx)
	default:
		return h.buildDetailedLogMessage(req, ctx)
	}
}

// buildSimpleLogMessage 构建简单格式日志
func (h *Handler) buildSimpleLogMessage(req *http.Request, ctx *plugin.RequestContext) string {
	return fmt.Sprintf("[%s] %s %s %s %s",
		time.Now().Format("2006-01-02 15:04:05"),
		ctx.ClientIP,
		req.Method,
		req.URL.Path,
		ctx.UserAgent)
}

// buildJSONLogMessage 构建JSON格式日志
func (h *Handler) buildJSONLogMessage(req *http.Request, ctx *plugin.RequestContext) string {
	logData := map[string]interface{}{
		"timestamp":    time.Now().Format(time.RFC3339),
		"client_ip":    ctx.ClientIP,
		"method":       req.Method,
		"url":          req.URL.String(),
		"user_agent":   ctx.UserAgent,
		"headers":      ctx.Headers,
		"body_size":    len(ctx.Body),
		"query_params": req.URL.Query(),
		"host":         req.Host,
		"proto":        req.Proto,
		"remote_addr":  req.RemoteAddr,
		"request_uri":  req.RequestURI,
	}

	jsonData, err := json.Marshal(logData)
	if err != nil {
		return fmt.Sprintf("Error marshaling log data: %v", err)
	}

	return string(jsonData)
}

// buildDetailedLogMessage 构建详细格式日志
func (h *Handler) buildDetailedLogMessage(req *http.Request, ctx *plugin.RequestContext) string {
	var headers []string
	for name, value := range ctx.Headers {
		headers = append(headers, fmt.Sprintf("%s=%s", name, value))
	}

	return fmt.Sprintf("[%s] %s %s %s %s Headers:[%s] BodySize:%d QueryParams:%v",
		time.Now().Format("2006-01-02 15:04:05"),
		ctx.ClientIP,
		req.Method,
		req.URL.String(),
		ctx.UserAgent,
		strings.Join(headers, ","),
		len(ctx.Body),
		req.URL.Query())
}
