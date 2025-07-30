package main

import (
	"fmt"
	"hackmitm/pkg/plugin"
)

// SimplePlugin 简单插件示例
type SimplePlugin struct {
	*plugin.PluginFramework
	utils *plugin.PluginUtils
}

// NewPlugin 创建插件实例
func NewPlugin(config map[string]interface{}) (plugin.Plugin, error) {
	// 创建框架配置
	frameworkConfig := &plugin.FrameworkConfig{
		Name:        "simple-plugin",
		Version:     "1.0.0",
		Description: "简单插件示例",
		Author:      "HackMITM",
		License:     "MIT",
		Priority:    100,
		Timeout:     30,
		Retry:       3,
		Async:       false,
	}

	// 创建插件框架
	framework := plugin.NewPluginFramework(frameworkConfig)

	// 创建插件实例
	p := &SimplePlugin{
		PluginFramework: framework,
		utils:           plugin.NewPluginUtils(),
	}

	// 设置配置验证器
	validator := plugin.NewConfigValidator()
	validator.AddBoolRule("enable_debug", false, false)
	validator.AddStringRule("log_level", false, "info", "", []string{"debug", "info", "warn", "error"})
	validator.AddIntRule("max_requests", false, 1000, 1, 10000)
	validator.AddStringRule("custom_header", false, "", "", nil)

	// 验证配置
	validatedConfig, err := validator.Validate(config)
	if err != nil {
		return nil, fmt.Errorf("配置验证失败: %w", err)
	}

	// 初始化基础插件
	if err := p.Initialize(validatedConfig); err != nil {
		return nil, fmt.Errorf("初始化插件失败: %w", err)
	}

	// 注册钩子
	p.registerHooks()

	return p, nil
}

// registerHooks 注册钩子
func (p *SimplePlugin) registerHooks() {
	// 请求前钩子
	p.AddHook(plugin.HookBeforeRequest, func(ctx *plugin.HookContext) error {
		// 获取客户端IP
		clientIP := p.utils.Request.GetClientIP(ctx.Request)
		ctx.Logger.Info(fmt.Sprintf("处理来自 %s 的请求", clientIP))

		// 检查安全性
		if p.utils.Security.IsSQLInjection(ctx.Request.URL.RawQuery) {
			ctx.Logger.Warn("检测到SQL注入攻击")
			return fmt.Errorf("检测到SQL注入攻击")
		}

		// 添加自定义头
		if customHeader := p.GetConfigString("custom_header", ""); customHeader != "" {
			ctx.Request.Header.Set("X-Custom-Header", customHeader)
		}

		return nil
	})

	// 请求后钩子
	p.AddHook(plugin.HookAfterRequest, func(ctx *plugin.HookContext) error {
		// 记录请求处理完成
		ctx.Logger.Info("请求处理完成")
		return nil
	})

	// 响应前钩子
	p.AddHook(plugin.HookBeforeResponse, func(ctx *plugin.HookContext) error {
		// 添加响应头
		p.utils.Response.SetHeader(ctx.Response, "X-Processed-By", "simple-plugin")
		return nil
	})

	// 响应后钩子
	p.AddHook(plugin.HookAfterResponse, func(ctx *plugin.HookContext) error {
		// 记录响应处理完成
		ctx.Logger.Info(fmt.Sprintf("响应处理完成，状态码: %d", ctx.Response.StatusCode))
		return nil
	})

	// 过滤钩子
	p.AddHook(plugin.HookOnFilter, func(ctx *plugin.HookContext) error {
		// 检查请求数量限制
		maxRequests := p.GetConfigInt("max_requests", 1000)
		if ctx.FilterCtx.RequestCount > int64(maxRequests) {
			ctx.Data["allow"] = false
			ctx.Logger.Warn(fmt.Sprintf("请求数量超过限制: %d", maxRequests))
		}
		return nil
	})

	// 添加日志中间件
	p.AddMiddleware(func(next plugin.HookFunc) plugin.HookFunc {
		return func(ctx *plugin.HookContext) error {
			start := p.utils.Time.Now()
			err := next(ctx)
			duration := p.utils.Time.Now().Sub(start)

			if err != nil {
				ctx.Logger.Error(fmt.Sprintf("钩子执行失败，耗时: %v, 错误: %v", duration, err))
			} else {
				ctx.Logger.Debug(fmt.Sprintf("钩子执行成功，耗时: %v", duration))
			}

			return err
		}
	})
}

// 确保实现了所有必要的接口
var (
	_ plugin.Plugin         = (*SimplePlugin)(nil)
	_ plugin.RequestPlugin  = (*SimplePlugin)(nil)
	_ plugin.ResponsePlugin = (*SimplePlugin)(nil)
	_ plugin.FilterPlugin   = (*SimplePlugin)(nil)
)
