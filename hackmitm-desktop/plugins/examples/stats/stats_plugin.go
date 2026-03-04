// Package main 请求统计插件
package main

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"sync/atomic"
	"time"

	"hackmitm/pkg/logger"
	"hackmitm/pkg/plugin"
)

// StatsPlugin 请求统计插件
type StatsPlugin struct {
	*plugin.BasePlugin
	// 配置项
	config struct {
		EnableDebug bool
		LogInterval int // 统计信息输出间隔（秒）
	}
	// 统计数据
	stats struct {
		sync.RWMutex
		totalRequests   int64            // 总请求数
		totalResponses  int64            // 总响应数
		statusCodes     map[int]int64    // HTTP状态码统计
		methods         map[string]int64 // HTTP方法统计
		paths           map[string]int64 // 请求路径统计
		avgResponseTime float64          // 平均响应时间
		lastUpdate      time.Time        // 最后更新时间
	}
}

// NewPlugin 创建插件实例
func NewPlugin(config map[string]interface{}) (plugin.Plugin, error) {
	p := &StatsPlugin{
		BasePlugin: plugin.NewBasePlugin("stats", "1.0.0", "统计分析插件"),
	}
	p.stats.statusCodes = make(map[int]int64)
	p.stats.methods = make(map[string]int64)
	p.stats.paths = make(map[string]int64)
	return p, nil
}

// Initialize 初始化插件
func (p *StatsPlugin) Initialize(config map[string]interface{}) error {
	if err := p.BasePlugin.Initialize(config); err != nil {
		return fmt.Errorf("初始化基础插件失败: %w", err)
	}

	// 加载配置
	p.config.EnableDebug = p.GetConfigBool("enable_debug", false)
	p.config.LogInterval = p.GetConfigInt("log_interval", 60)

	logger.Infof("统计插件初始化完成，调试模式: %v, 输出间隔: %d秒",
		p.config.EnableDebug, p.config.LogInterval)
	return nil
}

// Start 启动插件
func (p *StatsPlugin) Start(ctx context.Context) error {
	logger.Infof("统计插件启动")

	// 启动统计信息输出任务
	go p.statsReporter(ctx)

	return nil
}

// Stop 停止插件
func (p *StatsPlugin) Stop(ctx context.Context) error {
	logger.Info("统计插件停止")
	return nil
}

// ProcessRequest 处理请求
func (p *StatsPlugin) ProcessRequest(req *http.Request, ctx *plugin.RequestContext) error {
	atomic.AddInt64(&p.stats.totalRequests, 1)

	p.stats.Lock()
	p.stats.methods[req.Method]++
	p.stats.paths[req.URL.Path]++
	p.stats.lastUpdate = time.Now()
	p.stats.Unlock()

	if p.config.EnableDebug {
		logger.Debugf("[统计插件] 处理请求: %s %s", req.Method, req.URL.Path)
	}
	return nil
}

// ProcessResponse 处理响应
func (p *StatsPlugin) ProcessResponse(resp *http.Response, req *http.Request, ctx *plugin.ResponseContext) error {
	atomic.AddInt64(&p.stats.totalResponses, 1)

	p.stats.Lock()
	p.stats.statusCodes[resp.StatusCode]++
	// 更新平均响应时间
	p.stats.avgResponseTime = (p.stats.avgResponseTime*float64(p.stats.totalResponses-1) + ctx.Duration.Seconds()) / float64(p.stats.totalResponses)
	p.stats.Unlock()

	if p.config.EnableDebug {
		logger.Debugf("[统计插件] 处理响应: %s %s -> %d (%.2fms)",
			req.Method, req.URL.Path, resp.StatusCode, ctx.Duration.Seconds()*1000)
	}
	return nil
}

// Priority 返回插件优先级
func (p *StatsPlugin) Priority() int {
	return p.GetConfigInt("priority", 1000) // 统计插件优先级较低
}

// GetStatistics 获取统计信息
func (p *StatsPlugin) GetStatistics() map[string]interface{} {
	p.stats.RLock()
	defer p.stats.RUnlock()

	// 复制统计数据
	statusCodes := make(map[int]int64)
	methods := make(map[string]int64)
	paths := make(map[string]int64)
	for k, v := range p.stats.statusCodes {
		statusCodes[k] = v
	}
	for k, v := range p.stats.methods {
		methods[k] = v
	}
	for k, v := range p.stats.paths {
		paths[k] = v
	}

	return map[string]interface{}{
		"total_requests":    p.stats.totalRequests,
		"total_responses":   p.stats.totalResponses,
		"status_codes":      statusCodes,
		"methods":           methods,
		"paths":             paths,
		"avg_response_time": fmt.Sprintf("%.2fms", p.stats.avgResponseTime*1000),
		"last_update":       p.stats.lastUpdate,
		"plugin_version":    p.Version(),
		"plugin_enabled":    p.IsStarted(),
	}
}

// statsReporter 定期输出统计信息
func (p *StatsPlugin) statsReporter(ctx context.Context) {
	ticker := time.NewTicker(time.Duration(p.config.LogInterval) * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			stats := p.GetStatistics()
			logger.Infof("[统计插件] 统计信息:")
			logger.Infof("  总请求数: %d", stats["total_requests"])
			logger.Infof("  总响应数: %d", stats["total_responses"])
			logger.Infof("  平均响应时间: %s", stats["avg_response_time"])
			logger.Infof("  状态码统计: %v", stats["status_codes"])
			logger.Infof("  请求方法统计: %v", stats["methods"])
			logger.Infof("  热门路径 (前5个):")
			paths := stats["paths"].(map[string]int64)
			count := 0
			for path, hits := range paths {
				if count >= 5 {
					break
				}
				logger.Infof("    %s: %d次访问", path, hits)
				count++
			}
		}
	}
}

// 确保实现了所有必要的接口
var (
	_ plugin.Plugin         = (*StatsPlugin)(nil)
	_ plugin.RequestPlugin  = (*StatsPlugin)(nil)
	_ plugin.ResponsePlugin = (*StatsPlugin)(nil)
)
