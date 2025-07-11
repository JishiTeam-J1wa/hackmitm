// Package main 请求日志插件示例
package main

import (
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"hackmitm/pkg/logger"
	"hackmitm/pkg/plugin"
)

// RequestLoggerPlugin 请求日志插件
type RequestLoggerPlugin struct {
	*plugin.BasePlugin
	logFile     *os.File
	enableDebug bool
	logFormat   string
}

// NewPlugin 创建插件实例（插件入口函数）
func NewPlugin(config map[string]interface{}) (plugin.Plugin, error) {
	base := plugin.NewBasePlugin("request-logger", "1.0.0", "记录HTTP请求详细信息的插件")

	p := &RequestLoggerPlugin{
		BasePlugin: base,
	}

	return p, nil
}

// Priority 返回插件优先级
func (p *RequestLoggerPlugin) Priority() int {
	return p.GetConfigInt("priority", 100)
}

// Initialize 初始化插件
func (p *RequestLoggerPlugin) Initialize(config map[string]interface{}) error {
	if err := p.BasePlugin.Initialize(config); err != nil {
		return err
	}

	// 配置项
	p.enableDebug = p.GetConfigBool("enable_debug", false)
	p.logFormat = p.GetConfigString("log_format", "detailed")

	// 初始化日志文件
	logPath := p.GetConfigString("log_file", "")
	if logPath != "" {
		file, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
		if err != nil {
			return fmt.Errorf("打开日志文件失败: %w", err)
		}
		p.logFile = file
	}

	logger.Infof("请求日志插件初始化完成，调试模式: %v", p.enableDebug)
	return nil
}

// ProcessRequest 处理HTTP请求
func (p *RequestLoggerPlugin) ProcessRequest(req *http.Request, ctx *plugin.RequestContext) error {
	if !p.IsStarted() {
		return nil
	}

	// 构建日志信息
	logMsg := p.buildLogMessage(req, ctx)

	// 输出日志
	if p.logFile != nil {
		p.logFile.WriteString(logMsg + "\n")
	}

	// 根据配置级别决定是否打印到控制台
	if p.enableDebug {
		logger.Debugf("[REQUEST LOGGER] %s", logMsg)
	} else {
		logger.Infof("[REQUEST LOGGER] %s %s from %s", req.Method, req.URL.Path, ctx.ClientIP)
	}

	return nil
}

// buildLogMessage 构建日志消息
func (p *RequestLoggerPlugin) buildLogMessage(req *http.Request, ctx *plugin.RequestContext) string {
	timestamp := time.Now().Format("2006-01-02 15:04:05")

	switch p.logFormat {
	case "simple":
		return fmt.Sprintf("[%s] %s %s %s %s",
			timestamp, ctx.ClientIP, req.Method, req.URL.Path, ctx.UserAgent)

	case "json":
		return fmt.Sprintf(`{"timestamp":"%s","client_ip":"%s","method":"%s","url":"%s","user_agent":"%s","headers_count":%d}`,
			timestamp, ctx.ClientIP, req.Method, req.URL.String(), ctx.UserAgent, len(ctx.Headers))

	case "detailed":
		fallthrough
	default:
		var headers []string
		for name, value := range ctx.Headers {
			headers = append(headers, fmt.Sprintf("%s=%s", name, value))
		}

		return fmt.Sprintf("[%s] %s %s %s %s Headers:[%s] BodySize:%d",
			timestamp, ctx.ClientIP, req.Method, req.URL.String(),
			ctx.UserAgent, strings.Join(headers, ","), len(ctx.Body))
	}
}

// Cleanup 清理资源
func (p *RequestLoggerPlugin) Cleanup() error {
	if p.logFile != nil {
		p.logFile.Close()
	}
	return p.BasePlugin.Cleanup()
}

// 确保实现了接口
var _ plugin.RequestPlugin = (*RequestLoggerPlugin)(nil)

// 导出函数供插件系统调用
func LoadPlugin() plugin.Plugin {
	p, _ := NewPlugin(nil)
	return p
}
