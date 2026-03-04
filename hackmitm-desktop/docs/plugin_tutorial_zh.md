# HackMITM 插件开发教程

## 目录
- [插件系统概述](#插件系统概述)
- [传统插件开发](#传统插件开发)
- [新框架插件开发](#新框架插件开发)
- [实战案例](#实战案例)
- [最佳实践](#最佳实践)
- [调试和测试](#调试和测试)
- [部署和发布](#部署和发布)

## 插件系统概述

HackMITM 的插件系统经过重构，现在提供两种开发方式：

### 传统方式
- 直接实现插件接口
- 手动处理所有生命周期
- 适合简单插件

### 新框架方式
- 使用钩子系统
- 内置配置验证
- 丰富的工具库
- 中间件支持

## 传统插件开发

### 基础插件结构

```go
// plugins/examples/basic_plugin/basic_plugin.go
package main

import (
    "context"
    "fmt"
    "net/http"
    "time"
)

// 插件必须实现的接口
type Plugin interface {
    Name() string
    Version() string
    Description() string
    Initialize(config map[string]interface{}) error
    Start(ctx context.Context) error
    Stop(ctx context.Context) error
    Cleanup() error
}

// 请求处理插件接口
type RequestPlugin interface {
    Plugin
    ProcessRequest(req *http.Request, ctx *RequestContext) error
    Priority() int
}

// 响应处理插件接口
type ResponsePlugin interface {
    Plugin
    ProcessResponse(resp *http.Response, req *http.Request, ctx *ResponseContext) error
    Priority() int
}

// 基础插件结构
type BasicPlugin struct {
    name        string
    version     string
    description string
    config      map[string]interface{}
    enabled     bool
    startTime   time.Time
}

// 工厂函数 - 插件入口点
func NewPlugin(config map[string]interface{}) (Plugin, error) {
    plugin := &BasicPlugin{
        name:        "basic-plugin",
        version:     "1.0.0",
        description: "基础插件示例",
        config:      config,
    }
    
    return plugin, nil
}

// 实现Plugin接口
func (p *BasicPlugin) Name() string        { return p.name }
func (p *BasicPlugin) Version() string     { return p.version }
func (p *BasicPlugin) Description() string { return p.description }

func (p *BasicPlugin) Initialize(config map[string]interface{}) error {
    p.config = config
    p.enabled = true
    
    // 验证配置
    if err := p.validateConfig(); err != nil {
        return fmt.Errorf("配置验证失败: %w", err)
    }
    
    return nil
}

func (p *BasicPlugin) Start(ctx context.Context) error {
    p.startTime = time.Now()
    fmt.Printf("插件 %s 启动成功\n", p.name)
    return nil
}

func (p *BasicPlugin) Stop(ctx context.Context) error {
    p.enabled = false
    fmt.Printf("插件 %s 停止运行\n", p.name)
    return nil
}

func (p *BasicPlugin) Cleanup() error {
    // 清理资源
    fmt.Printf("插件 %s 清理完成\n", p.name)
    return nil
}

// 实现RequestPlugin接口
func (p *BasicPlugin) ProcessRequest(req *http.Request, ctx *RequestContext) error {
    if !p.enabled {
        return nil
    }
    
    // 处理请求
    fmt.Printf("处理请求: %s %s\n", req.Method, req.URL.String())
    
    // 添加自定义头
    req.Header.Set("X-Processed-By", p.name)
    
    return nil
}

func (p *BasicPlugin) Priority() int {
    return 100 // 默认优先级
}

// 配置验证
func (p *BasicPlugin) validateConfig() error {
    // 检查必需的配置项
    if _, ok := p.config["enabled"]; !ok {
        p.config["enabled"] = true
    }
    
    return nil
}
```

### 构建传统插件

```bash
# 在plugins目录下
cd plugins/examples/basic_plugin
go build -buildmode=plugin -o basic_plugin.so basic_plugin.go
```

## 新框架插件开发

### 使用新框架的优势

1. **钩子系统**：无需实现复杂接口，只需注册钩子函数
2. **配置验证**：自动验证和类型转换
3. **工具库**：丰富的工具函数
4. **中间件支持**：类似Web框架的中间件机制

### 基础框架插件

```go
// plugins/examples/framework_plugin/framework_plugin.go
package main

import (
    "context"
    "fmt"
    "net/http"
    "time"
    
    "hackmitm/pkg/plugin"
)

type FrameworkPlugin struct {
    *plugin.PluginFramework
    requestCount int64
    startTime    time.Time
}

func NewPlugin(config map[string]interface{}) (plugin.Plugin, error) {
    // 1. 创建配置验证器
    validator := plugin.NewConfigValidator()
    validator.AddRule("log_level", plugin.ValidationRule{
        Type:        "string",
        Default:     "info",
        Options:     []interface{}{"debug", "info", "warn", "error"},
        Description: "日志级别",
    })
    validator.AddRule("max_requests", plugin.ValidationRule{
        Type:        "int",
        Default:     1000,
        Min:         1,
        Max:         10000,
        Description: "最大请求数",
    })
    validator.AddRule("timeout", plugin.ValidationRule{
        Type:        "string",
        Default:     "30s",
        Pattern:     `^\d+[smh]$`,
        Description: "超时时间",
    })
    
    // 2. 验证配置
    validatedConfig, err := validator.Validate(config)
    if err != nil {
        return nil, fmt.Errorf("配置验证失败: %w", err)
    }
    
    // 3. 解析超时时间
    timeout, err := time.ParseDuration(validatedConfig["timeout"].(string))
    if err != nil {
        return nil, fmt.Errorf("解析超时时间失败: %w", err)
    }
    
    // 4. 创建插件框架
    framework := plugin.NewPluginFramework(&plugin.FrameworkConfig{
        Name:        "framework-plugin",
        Version:     "1.0.0",
        Description: "新框架插件示例",
        Timeout:     timeout,
        Config:      validatedConfig,
    })
    
    // 5. 创建插件实例
    p := &FrameworkPlugin{
        PluginFramework: framework,
        startTime:       time.Now(),
    }
    
    // 6. 注册钩子
    p.registerHooks()
    
    // 7. 添加中间件
    p.addMiddleware()
    
    return p, nil
}

func (p *FrameworkPlugin) registerHooks() {
    // 请求前钩子
    p.AddHook(plugin.HookBeforeRequest, p.beforeRequestHook)
    
    // 请求后钩子
    p.AddHook(plugin.HookAfterRequest, p.afterRequestHook)
    
    // 响应前钩子
    p.AddHook(plugin.HookBeforeResponse, p.beforeResponseHook)
    
    // 响应后钩子
    p.AddHook(plugin.HookAfterResponse, p.afterResponseHook)
    
    // 错误处理钩子
    p.AddHook(plugin.HookOnError, p.errorHook)
}

func (p *FrameworkPlugin) addMiddleware() {
    // 日志中间件
    p.AddMiddleware(p.loggingMiddleware)
    
    // 计时中间件
    p.AddMiddleware(p.timingMiddleware)
    
    // 错误恢复中间件
    p.AddMiddleware(p.recoveryMiddleware)
}

// 钩子函数实现
func (p *FrameworkPlugin) beforeRequestHook(ctx *plugin.HookContext) error {
    p.requestCount++
    
    // 检查请求数量限制
    maxRequests := p.GetConfig("max_requests").(int)
    if p.requestCount > int64(maxRequests) {
        return fmt.Errorf("请求数量超过限制: %d", maxRequests)
    }
    
    // 使用工具库获取客户端IP
    clientIP := plugin.Utils.Request.GetClientIP(ctx.Request)
    ctx.Data["client_ip"] = clientIP
    
    // 安全检查
    if plugin.Utils.Security.IsXSS(ctx.Request.URL.String()) {
        return fmt.Errorf("检测到XSS攻击")
    }
    
    ctx.Logger.Infof("处理请求: %s %s，客户端IP: %s", 
        ctx.Request.Method, ctx.Request.URL.String(), clientIP)
    
    return nil
}

func (p *FrameworkPlugin) afterRequestHook(ctx *plugin.HookContext) error {
    // 记录请求处理时间
    if startTime, ok := ctx.Data["start_time"].(time.Time); ok {
        duration := time.Since(startTime)
        ctx.Logger.Infof("请求处理完成，耗时: %v", duration)
    }
    
    return nil
}

func (p *FrameworkPlugin) beforeResponseHook(ctx *plugin.HookContext) error {
    // 添加安全头
    if ctx.Response != nil {
        ctx.Response.Header.Set("X-Frame-Options", "DENY")
        ctx.Response.Header.Set("X-Content-Type-Options", "nosniff")
        ctx.Response.Header.Set("X-XSS-Protection", "1; mode=block")
    }
    
    return nil
}

func (p *FrameworkPlugin) afterResponseHook(ctx *plugin.HookContext) error {
    // 记录响应状态
    if ctx.Response != nil {
        ctx.Logger.Infof("响应状态码: %d", ctx.Response.StatusCode)
    }
    
    return nil
}

func (p *FrameworkPlugin) errorHook(ctx *plugin.HookContext) error {
    // 错误处理
    if err, ok := ctx.Data["error"].(error); ok {
        ctx.Logger.Errorf("处理请求时发生错误: %v", err)
        
        // 可以在这里实现错误报告、重试等逻辑
    }
    
    return nil
}

// 中间件实现
func (p *FrameworkPlugin) loggingMiddleware(next plugin.HookFunc) plugin.HookFunc {
    return func(ctx *plugin.HookContext) error {
        logLevel := p.GetConfig("log_level").(string)
        
        if logLevel == "debug" {
            ctx.Logger.Debugf("开始执行钩子")
        }
        
        err := next(ctx)
        
        if err != nil {
            ctx.Logger.Errorf("钩子执行失败: %v", err)
        } else if logLevel == "debug" {
            ctx.Logger.Debugf("钩子执行成功")
        }
        
        return err
    }
}

func (p *FrameworkPlugin) timingMiddleware(next plugin.HookFunc) plugin.HookFunc {
    return func(ctx *plugin.HookContext) error {
        start := time.Now()
        ctx.Data["start_time"] = start
        
        err := next(ctx)
        
        duration := time.Since(start)
        ctx.Data["duration"] = duration
        
        return err
    }
}

func (p *FrameworkPlugin) recoveryMiddleware(next plugin.HookFunc) plugin.HookFunc {
    return func(ctx *plugin.HookContext) error {
        defer func() {
            if r := recover(); r != nil {
                ctx.Logger.Errorf("钩子执行时发生panic: %v", r)
                ctx.Data["error"] = fmt.Errorf("panic: %v", r)
            }
        }()
        
        return next(ctx)
    }
}
```

### 构建新框架插件

```bash
# 在plugins目录下
cd plugins/examples/framework_plugin
go build -buildmode=plugin -o framework_plugin.so framework_plugin.go
```

## 实战案例

### 案例1：HTTP缓存插件

```go
package main

import (
    "crypto/md5"
    "fmt"
    "net/http"
    "sync"
    "time"
    
    "hackmitm/pkg/plugin"
)

type CachePlugin struct {
    *plugin.PluginFramework
    cache     map[string]*CacheEntry
    cacheMux  sync.RWMutex
    maxSize   int
    ttl       time.Duration
}

type CacheEntry struct {
    Data      []byte
    Headers   http.Header
    Status    int
    Timestamp time.Time
}

func NewPlugin(config map[string]interface{}) (plugin.Plugin, error) {
    // 配置验证
    validator := plugin.NewConfigValidator()
    validator.AddRule("max_size", plugin.ValidationRule{
        Type:        "int",
        Default:     1000,
        Min:         1,
        Description: "缓存最大条目数",
    })
    validator.AddRule("ttl", plugin.ValidationRule{
        Type:        "string",
        Default:     "5m",
        Pattern:     `^\d+[smh]$`,
        Description: "缓存生存时间",
    })
    
    validatedConfig, err := validator.Validate(config)
    if err != nil {
        return nil, err
    }
    
    ttl, err := time.ParseDuration(validatedConfig["ttl"].(string))
    if err != nil {
        return nil, err
    }
    
    framework := plugin.NewPluginFramework(&plugin.FrameworkConfig{
        Name:        "cache-plugin",
        Version:     "1.0.0",
        Description: "HTTP缓存插件",
        Config:      validatedConfig,
    })
    
    p := &CachePlugin{
        PluginFramework: framework,
        cache:           make(map[string]*CacheEntry),
        maxSize:         validatedConfig["max_size"].(int),
        ttl:             ttl,
    }
    
    // 注册钩子
    p.AddHook(plugin.HookBeforeRequest, p.checkCache)
    p.AddHook(plugin.HookAfterResponse, p.storeCache)
    
    // 启动清理goroutine
    go p.cleanupExpired()
    
    return p, nil
}

func (p *CachePlugin) checkCache(ctx *plugin.HookContext) error {
    // 只缓存GET请求
    if ctx.Request.Method != "GET" {
        return nil
    }
    
    key := p.getCacheKey(ctx.Request)
    
    p.cacheMux.RLock()
    entry, exists := p.cache[key]
    p.cacheMux.RUnlock()
    
    if !exists {
        return nil
    }
    
    // 检查是否过期
    if time.Since(entry.Timestamp) > p.ttl {
        p.cacheMux.Lock()
        delete(p.cache, key)
        p.cacheMux.Unlock()
        return nil
    }
    
    // 从缓存返回响应
    ctx.Logger.Infof("从缓存返回响应: %s", key)
    
    // 这里需要设置响应，具体实现取决于框架
    ctx.Data["cached_response"] = entry
    
    return nil
}

func (p *CachePlugin) storeCache(ctx *plugin.HookContext) error {
    // 检查是否有缓存的响应
    if _, exists := ctx.Data["cached_response"]; exists {
        return nil
    }
    
    if ctx.Request.Method != "GET" || ctx.Response == nil {
        return nil
    }
    
    // 只缓存成功的响应
    if ctx.Response.StatusCode != 200 {
        return nil
    }
    
    key := p.getCacheKey(ctx.Request)
    
    // 读取响应体
    body, err := plugin.Utils.Response.ReadBody(ctx.Response)
    if err != nil {
        return err
    }
    
    entry := &CacheEntry{
        Data:      body,
        Headers:   ctx.Response.Header,
        Status:    ctx.Response.StatusCode,
        Timestamp: time.Now(),
    }
    
    p.cacheMux.Lock()
    defer p.cacheMux.Unlock()
    
    // 检查缓存大小
    if len(p.cache) >= p.maxSize {
        p.evictOldest()
    }
    
    p.cache[key] = entry
    ctx.Logger.Infof("响应已缓存: %s", key)
    
    return nil
}

func (p *CachePlugin) getCacheKey(req *http.Request) string {
    data := fmt.Sprintf("%s:%s", req.Method, req.URL.String())
    return fmt.Sprintf("%x", md5.Sum([]byte(data)))
}

func (p *CachePlugin) evictOldest() {
    var oldestKey string
    var oldestTime time.Time
    
    for key, entry := range p.cache {
        if oldestKey == "" || entry.Timestamp.Before(oldestTime) {
            oldestKey = key
            oldestTime = entry.Timestamp
        }
    }
    
    if oldestKey != "" {
        delete(p.cache, oldestKey)
    }
}

func (p *CachePlugin) cleanupExpired() {
    ticker := time.NewTicker(time.Minute)
    defer ticker.Stop()
    
    for range ticker.C {
        p.cacheMux.Lock()
        for key, entry := range p.cache {
            if time.Since(entry.Timestamp) > p.ttl {
                delete(p.cache, key)
            }
        }
        p.cacheMux.Unlock()
    }
}
```

### 案例2：请求限流插件

```go
package main

import (
    "fmt"
    "net/http"
    "sync"
    "time"
    
    "hackmitm/pkg/plugin"
)

type RateLimitPlugin struct {
    *plugin.PluginFramework
    limiters map[string]*TokenBucket
    mutex    sync.RWMutex
    rate     int
    capacity int
}

type TokenBucket struct {
    tokens     int
    capacity   int
    rate       int
    lastUpdate time.Time
    mutex      sync.Mutex
}

func NewPlugin(config map[string]interface{}) (plugin.Plugin, error) {
    validator := plugin.NewConfigValidator()
    validator.AddRule("rate", plugin.ValidationRule{
        Type:        "int",
        Default:     100,
        Min:         1,
        Description: "每秒允许的请求数",
    })
    validator.AddRule("capacity", plugin.ValidationRule{
        Type:        "int",
        Default:     1000,
        Min:         1,
        Description: "令牌桶容量",
    })
    
    validatedConfig, err := validator.Validate(config)
    if err != nil {
        return nil, err
    }
    
    framework := plugin.NewPluginFramework(&plugin.FrameworkConfig{
        Name:        "rate-limit-plugin",
        Version:     "1.0.0",
        Description: "请求限流插件",
        Config:      validatedConfig,
    })
    
    p := &RateLimitPlugin{
        PluginFramework: framework,
        limiters:        make(map[string]*TokenBucket),
        rate:            validatedConfig["rate"].(int),
        capacity:        validatedConfig["capacity"].(int),
    }
    
    p.AddHook(plugin.HookBeforeRequest, p.checkRateLimit)
    
    return p, nil
}

func (p *RateLimitPlugin) checkRateLimit(ctx *plugin.HookContext) error {
    clientIP := plugin.Utils.Request.GetClientIP(ctx.Request)
    
    p.mutex.RLock()
    bucket, exists := p.limiters[clientIP]
    p.mutex.RUnlock()
    
    if !exists {
        bucket = &TokenBucket{
            tokens:     p.capacity,
            capacity:   p.capacity,
            rate:       p.rate,
            lastUpdate: time.Now(),
        }
        
        p.mutex.Lock()
        p.limiters[clientIP] = bucket
        p.mutex.Unlock()
    }
    
    if !bucket.consume() {
        return fmt.Errorf("请求频率超限，客户端IP: %s", clientIP)
    }
    
    return nil
}

func (tb *TokenBucket) consume() bool {
    tb.mutex.Lock()
    defer tb.mutex.Unlock()
    
    now := time.Now()
    elapsed := now.Sub(tb.lastUpdate)
    
    // 添加令牌
    tokensToAdd := int(elapsed.Seconds()) * tb.rate
    tb.tokens += tokensToAdd
    if tb.tokens > tb.capacity {
        tb.tokens = tb.capacity
    }
    
    tb.lastUpdate = now
    
    // 消费令牌
    if tb.tokens > 0 {
        tb.tokens--
        return true
    }
    
    return false
}
```

### 案例3：安全检测插件

```go
package main

import (
    "fmt"
    "net/http"
    "strings"
    
    "hackmitm/pkg/plugin"
)

type SecurityPlugin struct {
    *plugin.PluginFramework
    blockedIPs    map[string]bool
    blockedPaths  []string
    enableXSSCheck bool
    enableSQLCheck bool
}

func NewPlugin(config map[string]interface{}) (plugin.Plugin, error) {
    validator := plugin.NewConfigValidator()
    validator.AddRule("blocked_ips", plugin.ValidationRule{
        Type:        "array",
        Default:     []interface{}{},
        Description: "黑名单IP列表",
    })
    validator.AddRule("blocked_paths", plugin.ValidationRule{
        Type:        "array",
        Default:     []interface{}{},
        Description: "黑名单路径列表",
    })
    validator.AddRule("enable_xss_check", plugin.ValidationRule{
        Type:        "bool",
        Default:     true,
        Description: "启用XSS检查",
    })
    validator.AddRule("enable_sql_check", plugin.ValidationRule{
        Type:        "bool",
        Default:     true,
        Description: "启用SQL注入检查",
    })
    
    validatedConfig, err := validator.Validate(config)
    if err != nil {
        return nil, err
    }
    
    framework := plugin.NewPluginFramework(&plugin.FrameworkConfig{
        Name:        "security-plugin",
        Version:     "1.0.0",
        Description: "安全检测插件",
        Config:      validatedConfig,
    })
    
    // 转换配置
    blockedIPs := make(map[string]bool)
    for _, ip := range validatedConfig["blocked_ips"].([]interface{}) {
        blockedIPs[ip.(string)] = true
    }
    
    var blockedPaths []string
    for _, path := range validatedConfig["blocked_paths"].([]interface{}) {
        blockedPaths = append(blockedPaths, path.(string))
    }
    
    p := &SecurityPlugin{
        PluginFramework: framework,
        blockedIPs:      blockedIPs,
        blockedPaths:    blockedPaths,
        enableXSSCheck:  validatedConfig["enable_xss_check"].(bool),
        enableSQLCheck:  validatedConfig["enable_sql_check"].(bool),
    }
    
    p.AddHook(plugin.HookBeforeRequest, p.securityCheck)
    
    return p, nil
}

func (p *SecurityPlugin) securityCheck(ctx *plugin.HookContext) error {
    clientIP := plugin.Utils.Request.GetClientIP(ctx.Request)
    
    // IP黑名单检查
    if p.blockedIPs[clientIP] {
        ctx.Logger.Warnf("阻止黑名单IP访问: %s", clientIP)
        return fmt.Errorf("访问被拒绝")
    }
    
    // 路径黑名单检查
    for _, blockedPath := range p.blockedPaths {
        if strings.Contains(ctx.Request.URL.Path, blockedPath) {
            ctx.Logger.Warnf("阻止访问黑名单路径: %s", ctx.Request.URL.Path)
            return fmt.Errorf("路径访问被拒绝")
        }
    }
    
    // XSS检查
    if p.enableXSSCheck {
        if plugin.Utils.Security.IsXSS(ctx.Request.URL.String()) {
            ctx.Logger.Warnf("检测到XSS攻击: %s", ctx.Request.URL.String())
            return fmt.Errorf("检测到XSS攻击")
        }
        
        // 检查请求体
        if ctx.Request.Body != nil {
            body, err := plugin.Utils.Request.ReadBody(ctx.Request)
            if err == nil && plugin.Utils.Security.IsXSS(string(body)) {
                ctx.Logger.Warnf("请求体中检测到XSS攻击")
                return fmt.Errorf("请求体中检测到XSS攻击")
            }
        }
    }
    
    // SQL注入检查
    if p.enableSQLCheck {
        if plugin.Utils.Security.IsSQLInjection(ctx.Request.URL.String()) {
            ctx.Logger.Warnf("检测到SQL注入攻击: %s", ctx.Request.URL.String())
            return fmt.Errorf("检测到SQL注入攻击")
        }
    }
    
    return nil
}
```

## 最佳实践

### 1. 配置管理

```go
// 使用配置验证器
validator := plugin.NewConfigValidator()
validator.AddRule("timeout", plugin.ValidationRule{
    Type:        "string",
    Default:     "30s",
    Pattern:     `^\d+[smh]$`,
    Description: "操作超时时间",
})

// 配置热重载
func (p *MyPlugin) ReloadConfig(newConfig map[string]interface{}) error {
    validatedConfig, err := p.validator.Validate(newConfig)
    if err != nil {
        return err
    }
    
    p.configMux.Lock()
    p.config = validatedConfig
    p.configMux.Unlock()
    
    return nil
}
```

### 2. 错误处理

```go
// 自定义错误类型
type PluginError struct {
    Code    int
    Message string
    Cause   error
}

func (e *PluginError) Error() string {
    if e.Cause != nil {
        return fmt.Sprintf("%s: %v", e.Message, e.Cause)
    }
    return e.Message
}

// 错误处理钩子
func (p *MyPlugin) errorHook(ctx *plugin.HookContext) error {
    if err, ok := ctx.Data["error"].(error); ok {
        // 记录错误
        p.logger.Errorf("插件错误: %v", err)
        
        // 发送告警
        p.sendAlert(err)
        
        // 可选：尝试恢复
        return p.tryRecover(err)
    }
    
    return nil
}
```

### 3. 性能优化

```go
// 使用对象池
var contextPool = sync.Pool{
    New: func() interface{} {
        return &MyContext{
            Data: make(map[string]interface{}),
        }
    },
}

func (p *MyPlugin) getContext() *MyContext {
    ctx := contextPool.Get().(*MyContext)
    // 重置状态
    for k := range ctx.Data {
        delete(ctx.Data, k)
    }
    return ctx
}

func (p *MyPlugin) putContext(ctx *MyContext) {
    contextPool.Put(ctx)
}

// 异步处理
func (p *MyPlugin) asyncProcess(data interface{}) {
    select {
    case p.workQueue <- data:
        // 成功加入队列
    default:
        // 队列满，记录错误
        p.logger.Warnf("工作队列已满，丢弃数据")
    }
}
```

### 4. 监控和指标

```go
// 指标收集
type PluginMetrics struct {
    RequestCount    int64
    ErrorCount      int64
    ProcessingTime  time.Duration
    LastError       time.Time
}

func (p *MyPlugin) recordMetrics(duration time.Duration, err error) {
    atomic.AddInt64(&p.metrics.RequestCount, 1)
    
    if err != nil {
        atomic.AddInt64(&p.metrics.ErrorCount, 1)
        p.metrics.LastError = time.Now()
    }
    
    // 计算平均处理时间
    p.metrics.ProcessingTime = (p.metrics.ProcessingTime + duration) / 2
}

// 健康检查
func (p *MyPlugin) HealthCheck() error {
    // 检查关键组件状态
    if !p.isHealthy() {
        return fmt.Errorf("插件不健康")
    }
    
    return nil
}
```

## 调试和测试

### 1. 单元测试

```go
// plugin_test.go
package main

import (
    "context"
    "net/http"
    "testing"
    
    "hackmitm/pkg/plugin"
)

func TestMyPlugin(t *testing.T) {
    // 创建测试配置
    config := map[string]interface{}{
        "timeout": "10s",
        "enabled": true,
    }
    
    // 创建插件实例
    p, err := NewPlugin(config)
    if err != nil {
        t.Fatalf("创建插件失败: %v", err)
    }
    
    // 测试初始化
    if err := p.Initialize(config); err != nil {
        t.Fatalf("初始化失败: %v", err)
    }
    
    // 测试启动
    ctx := context.Background()
    if err := p.Start(ctx); err != nil {
        t.Fatalf("启动失败: %v", err)
    }
    
    // 测试请求处理
    req, _ := http.NewRequest("GET", "http://example.com", nil)
    reqCtx := &plugin.RequestContext{
        StartTime: time.Now(),
        ClientIP:  "127.0.0.1",
    }
    
    if err := p.(plugin.RequestPlugin).ProcessRequest(req, reqCtx); err != nil {
        t.Fatalf("处理请求失败: %v", err)
    }
    
    // 测试停止
    if err := p.Stop(ctx); err != nil {
        t.Fatalf("停止失败: %v", err)
    }
}

// 基准测试
func BenchmarkMyPlugin(b *testing.B) {
    config := map[string]interface{}{
        "timeout": "10s",
    }
    
    p, _ := NewPlugin(config)
    p.Initialize(config)
    p.Start(context.Background())
    
    req, _ := http.NewRequest("GET", "http://example.com", nil)
    reqCtx := &plugin.RequestContext{
        StartTime: time.Now(),
        ClientIP:  "127.0.0.1",
    }
    
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        p.(plugin.RequestPlugin).ProcessRequest(req, reqCtx)
    }
}
```

### 2. 集成测试

```go
// integration_test.go
func TestPluginIntegration(t *testing.T) {
    // 启动测试服务器
    server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        w.WriteHeader(http.StatusOK)
        w.Write([]byte("test response"))
    }))
    defer server.Close()
    
    // 配置代理服务器
    proxyConfig := &config.Config{
        Server: config.ServerConfig{
            ListenAddr: ":0", // 随机端口
        },
        Plugins: []config.PluginConfig{
            {
                Name:    "test-plugin",
                Path:    "test_plugin.so",
                Enabled: true,
                Config: map[string]interface{}{
                    "target_url": server.URL,
                },
            },
        },
    }
    
    // 创建代理服务器
    proxy, err := proxy.NewServer(proxyConfig)
    if err != nil {
        t.Fatalf("创建代理失败: %v", err)
    }
    
    // 启动代理
    if err := proxy.Start(); err != nil {
        t.Fatalf("启动代理失败: %v", err)
    }
    defer proxy.Stop()
    
    // 发送测试请求
    client := &http.Client{}
    resp, err := client.Get(fmt.Sprintf("http://localhost:%d", proxy.Port()))
    if err != nil {
        t.Fatalf("请求失败: %v", err)
    }
    defer resp.Body.Close()
    
    // 验证响应
    if resp.StatusCode != http.StatusOK {
        t.Errorf("期望状态码 200，实际 %d", resp.StatusCode)
    }
}
```

## 部署和发布

### 1. 构建脚本

```makefile
# Makefile
PLUGIN_DIR = plugins/examples
PLUGINS = $(wildcard $(PLUGIN_DIR)/*/main.go)
PLUGIN_NAMES = $(notdir $(patsubst %/main.go,%,$(PLUGINS)))

.PHONY: all clean $(PLUGIN_NAMES)

all: $(PLUGIN_NAMES)

$(PLUGIN_NAMES):
	@echo "构建插件: $@"
	cd $(PLUGIN_DIR)/$@ && go build -buildmode=plugin -o $@.so main.go

clean:
	find $(PLUGIN_DIR) -name "*.so" -delete

test:
	go test -v ./...

benchmark:
	go test -bench=. -benchmem ./...
```

### 2. 配置管理

```json
{
  "plugins": [
    {
      "name": "my-plugin",
      "path": "my_plugin.so",
      "enabled": true,
      "priority": 100,
      "config": {
        "timeout": "30s",
        "log_level": "info",
        "max_requests": 1000
      }
    }
  ]
}
```

### 3. 热重载

```go
// 支持热重载的插件管理器
func (m *Manager) ReloadPlugin(name string, config *PluginConfig) error {
    m.mutex.Lock()
    defer m.mutex.Unlock()
    
    // 停止旧插件
    if oldPlugin, exists := m.plugins[name]; exists {
        oldPlugin.Stop(context.Background())
        oldPlugin.Cleanup()
    }
    
    // 加载新插件
    return m.LoadPlugin(config)
}
```

这个教程涵盖了从基础插件开发到高级特性的完整流程，包括实际的代码示例和最佳实践。通过这个教程，开发者可以快速上手插件开发，并创建出功能强大、性能优秀的插件。 