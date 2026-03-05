// Package plugin 插件开发框架
package plugin

import (
	"context"
	"fmt"
	"net/http"
	"reflect"
	"sync"
	"time"

	"hackmitm/pkg/logger"
)

// PluginFramework 插件框架
type PluginFramework struct {
	*BasePlugin
	hooks      map[string][]HookFunc
	middleware []MiddlewareFunc
	config     *FrameworkConfig
	mutex      sync.RWMutex
}

// FrameworkConfig 框架配置
type FrameworkConfig struct {
	Name        string `json:"name"`
	Version     string `json:"version"`
	Description string `json:"description"`
	Author      string `json:"author"`
	License     string `json:"license"`
	Priority    int    `json:"priority"`
	Timeout     int    `json:"timeout"` // 超时时间（秒）
	Retry       int    `json:"retry"`   // 重试次数
	Async       bool   `json:"async"`   // 是否异步执行
}

// HookFunc 钩子函数类型
type HookFunc func(ctx *HookContext) error

// MiddlewareFunc 中间件函数类型
type MiddlewareFunc func(next HookFunc) HookFunc

// HookContext 钩子上下文
type HookContext struct {
	Request     *http.Request
	Response    *http.Response
	RequestCtx  *RequestContext
	ResponseCtx *ResponseContext
	FilterCtx   *FilterContext
	ErrorCtx    *ErrorContext
	Data        map[string]interface{}
	Logger      *logger.Logger
	Cancel      context.CancelFunc
}

// HookType 钩子类型
type HookType string

const (
	HookBeforeRequest  HookType = "before_request"
	HookAfterRequest   HookType = "after_request"
	HookBeforeResponse HookType = "before_response"
	HookAfterResponse  HookType = "after_response"
	HookOnError        HookType = "on_error"
	HookOnFilter       HookType = "on_filter"
)

// NewPluginFramework 创建插件框架
func NewPluginFramework(config *FrameworkConfig) *PluginFramework {
	base := NewBasePlugin(config.Name, config.Version, config.Description)

	return &PluginFramework{
		BasePlugin: base,
		hooks:      make(map[string][]HookFunc),
		middleware: make([]MiddlewareFunc, 0),
		config:     config,
	}
}

// AddHook 添加钩子
func (pf *PluginFramework) AddHook(hookType HookType, hook HookFunc) {
	pf.mutex.Lock()
	defer pf.mutex.Unlock()

	key := string(hookType)
	pf.hooks[key] = append(pf.hooks[key], hook)
}

// AddMiddleware 添加中间件
func (pf *PluginFramework) AddMiddleware(middleware MiddlewareFunc) {
	pf.mutex.Lock()
	defer pf.mutex.Unlock()

	pf.middleware = append(pf.middleware, middleware)
}

// executeHooks 执行钩子
func (pf *PluginFramework) executeHooks(hookType HookType, ctx *HookContext) error {
	pf.mutex.RLock()
	hooks := pf.hooks[string(hookType)]
	pf.mutex.RUnlock()

	if len(hooks) == 0 {
		return nil
	}

	// 应用中间件
	for _, hook := range hooks {
		finalHook := hook

		// 从后往前应用中间件
		for i := len(pf.middleware) - 1; i >= 0; i-- {
			finalHook = pf.middleware[i](finalHook)
		}

		// 执行钩子
		if pf.config.Async {
			go func(h HookFunc) {
				if err := pf.executeWithTimeout(h, ctx); err != nil {
					logger.Errorf("异步钩子执行失败: %v", err)
				}
			}(finalHook)
		} else {
			if err := pf.executeWithTimeout(finalHook, ctx); err != nil {
				return err
			}
		}
	}

	return nil
}

// executeWithTimeout 带超时的执行
func (pf *PluginFramework) executeWithTimeout(hook HookFunc, ctx *HookContext) error {
	timeout := time.Duration(pf.config.Timeout) * time.Second
	if timeout <= 0 {
		timeout = 30 * time.Second
	}

	done := make(chan error, 1)

	go func() {
		defer func() {
			if r := recover(); r != nil {
				done <- fmt.Errorf("钩子执行panic: %v", r)
			}
		}()
		done <- hook(ctx)
	}()

	select {
	case err := <-done:
		return err
	case <-time.After(timeout):
		if ctx.Cancel != nil {
			ctx.Cancel()
		}
		return fmt.Errorf("钩子执行超时")
	}
}

// ProcessRequest 处理请求
func (pf *PluginFramework) ProcessRequest(req *http.Request, ctx *RequestContext) error {
	hookCtx := &HookContext{
		Request:    req,
		RequestCtx: ctx,
		Data:       make(map[string]interface{}),
		Logger:     logger.NewLogger(),
	}

	// 执行前置钩子
	if err := pf.executeHooks(HookBeforeRequest, hookCtx); err != nil {
		return err
	}

	// 执行后置钩子
	if err := pf.executeHooks(HookAfterRequest, hookCtx); err != nil {
		return err
	}

	return nil
}

// ProcessResponse 处理响应
func (pf *PluginFramework) ProcessResponse(resp *http.Response, req *http.Request, ctx *ResponseContext) error {
	hookCtx := &HookContext{
		Request:     req,
		Response:    resp,
		ResponseCtx: ctx,
		Data:        make(map[string]interface{}),
		Logger:      logger.NewLogger(),
	}

	// 执行前置钩子
	if err := pf.executeHooks(HookBeforeResponse, hookCtx); err != nil {
		return err
	}

	// 执行后置钩子
	if err := pf.executeHooks(HookAfterResponse, hookCtx); err != nil {
		return err
	}

	return nil
}

// ShouldAllow 过滤请求
func (pf *PluginFramework) ShouldAllow(req *http.Request, ctx *FilterContext) (bool, error) {
	hookCtx := &HookContext{
		Request:   req,
		FilterCtx: ctx,
		Data:      make(map[string]interface{}),
		Logger:    logger.NewLogger(),
	}

	// 设置默认允许
	hookCtx.Data["allow"] = true

	// 执行过滤钩子
	if err := pf.executeHooks(HookOnFilter, hookCtx); err != nil {
		return false, err
	}

	// 检查结果
	if allow, ok := hookCtx.Data["allow"].(bool); ok {
		return allow, nil
	}

	return true, nil
}

// Priority 返回优先级
func (pf *PluginFramework) Priority() int {
	return pf.config.Priority
}

// GetConfig 获取配置
func (pf *PluginFramework) GetConfig() *FrameworkConfig {
	return pf.config
}

// SetConfig 设置配置
func (pf *PluginFramework) SetConfig(config *FrameworkConfig) {
	pf.mutex.Lock()
	defer pf.mutex.Unlock()
	pf.config = config
}

// GetHooks 获取所有钩子
func (pf *PluginFramework) GetHooks() map[string][]HookFunc {
	pf.mutex.RLock()
	defer pf.mutex.RUnlock()

	result := make(map[string][]HookFunc)
	for k, v := range pf.hooks {
		result[k] = make([]HookFunc, len(v))
		copy(result[k], v)
	}
	return result
}

// RemoveHook 移除钩子
func (pf *PluginFramework) RemoveHook(hookType HookType, hook HookFunc) {
	pf.mutex.Lock()
	defer pf.mutex.Unlock()

	key := string(hookType)
	hooks := pf.hooks[key]

	for i, h := range hooks {
		if reflect.ValueOf(h).Pointer() == reflect.ValueOf(hook).Pointer() {
			pf.hooks[key] = append(hooks[:i], hooks[i+1:]...)
			break
		}
	}
}

// ClearHooks 清除所有钩子
func (pf *PluginFramework) ClearHooks() {
	pf.mutex.Lock()
	defer pf.mutex.Unlock()
	pf.hooks = make(map[string][]HookFunc)
}

// GetStats 获取统计信息
func (pf *PluginFramework) GetStats() map[string]interface{} {
	pf.mutex.RLock()
	defer pf.mutex.RUnlock()

	stats := make(map[string]interface{})
	stats["name"] = pf.config.Name
	stats["version"] = pf.config.Version
	stats["priority"] = pf.config.Priority
	stats["async"] = pf.config.Async
	stats["timeout"] = pf.config.Timeout
	stats["retry"] = pf.config.Retry

	// 统计钩子数量
	hookStats := make(map[string]int)
	for hookType, hooks := range pf.hooks {
		hookStats[hookType] = len(hooks)
	}
	stats["hooks"] = hookStats
	stats["middleware_count"] = len(pf.middleware)

	return stats
}
