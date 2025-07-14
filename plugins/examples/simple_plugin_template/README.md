# 简化插件开发指南

## 概述

这是一个基于新框架的简化插件开发示例，展示了如何快速开发高质量的插件。

## 主要改进

### 1. 框架化开发
- 使用 `PluginFramework` 简化插件开发
- 内置钩子系统，支持请求/响应的各个阶段
- 内置中间件支持，方便功能扩展

### 2. 配置验证
- 自动配置验证和类型转换
- 支持默认值、范围检查、正则验证
- 详细的错误信息

### 3. 工具库
- 丰富的工具函数库
- 请求/响应处理工具
- 安全检查工具
- 类型转换工具
- 时间处理工具

### 4. 简化的开发模式
- 基于钩子的事件驱动模式
- 声明式配置
- 自动错误处理和恢复

## 快速开始

### 1. 创建插件

```go
package main

import (
    "fmt"
    "hackmitm/pkg/plugin"
)

type MyPlugin struct {
    *plugin.PluginFramework
    utils *plugin.PluginUtils
}

func NewPlugin(config map[string]interface{}) (plugin.Plugin, error) {
    // 创建框架配置
    frameworkConfig := &plugin.FrameworkConfig{
        Name:        "my-plugin",
        Version:     "1.0.0",
        Description: "我的插件",
        Priority:    100,
        Timeout:     30,
        Async:       false,
    }

    // 创建插件框架
    framework := plugin.NewPluginFramework(frameworkConfig)
    
    // 创建插件实例
    p := &MyPlugin{
        PluginFramework: framework,
        utils:          plugin.NewPluginUtils(),
    }

    // 配置验证
    validator := plugin.NewConfigValidator()
    validator.AddBoolRule("enable_debug", false, false)
    
    validatedConfig, err := validator.Validate(config)
    if err != nil {
        return nil, err
    }

    // 初始化
    if err := p.Initialize(validatedConfig); err != nil {
        return nil, err
    }

    // 注册钩子
    p.registerHooks()

    return p, nil
}

func (p *MyPlugin) registerHooks() {
    // 请求前处理
    p.AddHook(plugin.HookBeforeRequest, func(ctx *plugin.HookContext) error {
        // 你的逻辑
        return nil
    })
}
```

### 2. 配置验证

```go
// 创建验证器
validator := plugin.NewConfigValidator()

// 添加验证规则
validator.AddStringRule("log_level", false, "info", "", []string{"debug", "info", "warn", "error"})
validator.AddIntRule("max_requests", false, 1000, 1, 10000)
validator.AddBoolRule("enable_debug", false, false)
validator.AddArrayRule("allowed_ips", false, "string", 0, 100)

// 验证配置
validatedConfig, err := validator.Validate(config)
```

### 3. 使用工具库

```go
// 请求工具
clientIP := p.utils.Request.GetClientIP(ctx.Request)
isJSON := p.utils.Request.IsJSON(ctx.Request)
body, err := p.utils.Request.ReadBody(ctx.Request)

// 响应工具
p.utils.Response.SetHeader(ctx.Response, "X-Custom", "value")
p.utils.Response.SetJSONBody(ctx.Response, data)

// 安全工具
if p.utils.Security.IsSQLInjection(input) {
    return fmt.Errorf("检测到SQL注入")
}
if p.utils.Security.IsXSS(input) {
    return fmt.Errorf("检测到XSS攻击")
}

// 类型转换
str := p.utils.Conversion.ToString(value)
num, err := p.utils.Conversion.ToInt(value)
```

### 4. 钩子系统

```go
// 请求处理钩子
p.AddHook(plugin.HookBeforeRequest, func(ctx *plugin.HookContext) error {
    // 请求前处理
    return nil
})

p.AddHook(plugin.HookAfterRequest, func(ctx *plugin.HookContext) error {
    // 请求后处理
    return nil
})

// 响应处理钩子
p.AddHook(plugin.HookBeforeResponse, func(ctx *plugin.HookContext) error {
    // 响应前处理
    return nil
})

p.AddHook(plugin.HookAfterResponse, func(ctx *plugin.HookContext) error {
    // 响应后处理
    return nil
})

// 过滤钩子
p.AddHook(plugin.HookOnFilter, func(ctx *plugin.HookContext) error {
    // 设置是否允许请求通过
    ctx.Data["allow"] = true
    return nil
})
```

### 5. 中间件

```go
// 添加日志中间件
p.AddMiddleware(func(next plugin.HookFunc) plugin.HookFunc {
    return func(ctx *plugin.HookContext) error {
        start := time.Now()
        err := next(ctx)
        duration := time.Since(start)
        
        ctx.Logger.Info(fmt.Sprintf("钩子执行耗时: %v", duration))
        return err
    }
})

// 添加错误处理中间件
p.AddMiddleware(func(next plugin.HookFunc) plugin.HookFunc {
    return func(ctx *plugin.HookContext) error {
        defer func() {
            if r := recover(); r != nil {
                ctx.Logger.Error(fmt.Sprintf("钩子执行panic: %v", r))
            }
        }()
        return next(ctx)
    }
})
```

## 配置示例

```json
{
  "name": "my-plugin",
  "enabled": true,
  "path": "examples/my_plugin.so",
  "priority": 100,
  "config": {
    "enable_debug": true,
    "log_level": "debug",
    "max_requests": 1000,
    "custom_header": "MyPlugin/1.0",
    "allowed_ips": ["127.0.0.1", "192.168.1.0/24"]
  }
}
```

## 最佳实践

1. **使用配置验证器**：确保配置的正确性和类型安全
2. **合理使用钩子**：在适当的时机处理请求和响应
3. **利用工具库**：减少重复代码，提高开发效率
4. **添加中间件**：实现横切关注点，如日志、监控、错误处理
5. **异步处理**：对于耗时操作，考虑使用异步模式
6. **错误处理**：提供详细的错误信息，便于调试
7. **性能优化**：避免阻塞操作，合理使用缓存

## 调试和测试

1. **启用调试模式**：在配置中设置 `enable_debug: true`
2. **查看日志**：使用内置的日志系统记录关键信息
3. **监控指标**：通过 `/status` 端点查看插件运行状态
4. **单元测试**：为关键逻辑编写单元测试

## 常见问题

1. **Q: 如何处理异步操作？**
   A: 设置 `Async: true`，框架会自动处理异步执行

2. **Q: 如何实现插件间通信？**
   A: 使用 `HookContext.Data` 在钩子间传递数据

3. **Q: 如何处理超时？**
   A: 设置 `Timeout` 参数，框架会自动处理超时

4. **Q: 如何实现热重载？**
   A: 插件管理器支持动态重载，无需重启服务

通过这个简化的框架，插件开发变得更加简单、高效和可靠。 