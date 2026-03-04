// Package main 请求日志插件示例
package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"sync"
	"time"

	"hackmitm/pkg/logger"
	"hackmitm/pkg/plugin"
)

// RequestLoggerPlugin 请求日志插件
type RequestLoggerPlugin struct {
	*plugin.BasePlugin
	Config  *Config
	handler *Handler
	logger  *logger.Logger
	logFile *os.File
	stats   struct {
		sync.RWMutex
		requestCount int64
		errorCount   int64
		bytesLogged  int64
		lastLogTime  time.Time
	}
}

// NewPlugin 创建插件实例
func NewPlugin(config map[string]interface{}) (plugin.Plugin, error) {
	base := plugin.NewBasePlugin("request-logger", "1.0.0", "记录HTTP请求详细信息的插件")

	p := &RequestLoggerPlugin{
		BasePlugin: base,
		Config:     DefaultConfig(),
	}

	// 创建处理器
	p.handler = NewHandler(p)

	return p, nil
}

// Initialize 初始化插件
func (p *RequestLoggerPlugin) Initialize(config map[string]interface{}) error {
	if err := p.BasePlugin.Initialize(config); err != nil {
		return fmt.Errorf("初始化基础插件失败: %w", err)
	}

	// 加载配置
	if err := p.loadConfig(config); err != nil {
		return fmt.Errorf("加载配置失败: %w", err)
	}

	// 初始化日志
	if err := p.initLogger(); err != nil {
		return fmt.Errorf("初始化日志失败: %w", err)
	}

	logger.Infof("请求日志插件初始化完成，调试模式: %v", p.Config.EnableDebug)
	return nil
}

// loadConfig 加载配置
func (p *RequestLoggerPlugin) loadConfig(config map[string]interface{}) error {
	// 从配置映射更新配置结构
	if enableDebug, ok := config["enable_debug"].(bool); ok {
		p.Config.EnableDebug = enableDebug
	}
	if logLevel, ok := config["log_level"].(string); ok {
		p.Config.LogLevel = logLevel
	}
	if logFormat, ok := config["log_format"].(string); ok {
		p.Config.LogFormat = logFormat
	}
	if logFile, ok := config["log_file"].(string); ok {
		p.Config.LogFile = logFile
	}
	if maxSize, ok := config["max_size"].(int); ok {
		p.Config.MaxSize = maxSize
	}
	if maxBackups, ok := config["max_backups"].(int); ok {
		p.Config.MaxBackups = maxBackups
	}
	if maxAge, ok := config["max_age"].(int); ok {
		p.Config.MaxAge = maxAge
	}
	if compress, ok := config["compress"].(bool); ok {
		p.Config.Compress = compress
	}

	// 验证配置
	return p.Config.Validate()
}

// initLogger 初始化日志器
func (p *RequestLoggerPlugin) initLogger() error {
	// 创建日志器
	p.logger = logger.NewLogger()

	// 设置日志级别
	switch p.Config.LogLevel {
	case "debug":
		p.logger.SetLevel(logger.DebugLevel)
	case "warn":
		p.logger.SetLevel(logger.WarnLevel)
	case "error":
		p.logger.SetLevel(logger.ErrorLevel)
	default:
		p.logger.SetLevel(logger.InfoLevel)
	}

	// 如果指定了日志文件，设置输出
	if p.Config.LogFile != "" {
		file, err := os.OpenFile(p.Config.LogFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
		if err != nil {
			return fmt.Errorf("打开日志文件失败: %w", err)
		}
		p.logFile = file
		p.logger.SetOutput(file)
	}

	return nil
}

// Start 启动插件
func (p *RequestLoggerPlugin) Start(ctx context.Context) error {
	logger.Info("请求日志插件启动")
	return nil
}

// Stop 停止插件
func (p *RequestLoggerPlugin) Stop(ctx context.Context) error {
	logger.Info("请求日志插件停止")
	return nil
}

// Cleanup 清理资源
func (p *RequestLoggerPlugin) Cleanup() error {
	if p.logFile != nil {
		p.logFile.Close()
	}
	return p.BasePlugin.Cleanup()
}

// ProcessRequest 处理HTTP请求
func (p *RequestLoggerPlugin) ProcessRequest(req *http.Request, ctx *plugin.RequestContext) error {
	if !p.IsStarted() {
		return nil
	}

	// 更新统计信息
	p.stats.Lock()
	p.stats.requestCount++
	p.stats.lastLogTime = time.Now()
	p.stats.Unlock()

	// 处理请求
	if err := p.handler.HandleRequest(req, ctx); err != nil {
		p.stats.Lock()
		p.stats.errorCount++
		p.stats.Unlock()
		return err
	}

	return nil
}

// Priority 返回插件优先级
func (p *RequestLoggerPlugin) Priority() int {
	return p.GetConfigInt("priority", 100)
}

// GetStatistics 获取统计信息
func (p *RequestLoggerPlugin) GetStatistics() map[string]interface{} {
	p.stats.RLock()
	defer p.stats.RUnlock()

	return map[string]interface{}{
		"request_count": p.stats.requestCount,
		"error_count":   p.stats.errorCount,
		"bytes_logged":  p.stats.bytesLogged,
		"last_log_time": p.stats.lastLogTime,
		"config": map[string]interface{}{
			"enable_debug": p.Config.EnableDebug,
			"log_level":    p.Config.LogLevel,
			"log_format":   p.Config.LogFormat,
			"log_file":     p.Config.LogFile,
		},
	}
}

// 确保实现了所有必要的接口
var (
	_ plugin.Plugin        = (*RequestLoggerPlugin)(nil)
	_ plugin.RequestPlugin = (*RequestLoggerPlugin)(nil)
)
