# HackMITM 快速参考手册

## 常用命令

### 构建和运行
```bash
# 构建主程序
go build -o bin/hackmitm cmd/hackmitm/main.go

# 构建插件
cd plugins && make examples

# 运行程序
./bin/hackmitm -config configs/config.json

# 查看帮助
./bin/hackmitm -h
```

### 插件开发
```bash
# 创建插件目录
mkdir plugins/examples/my_plugin

# 构建插件
cd plugins/examples/my_plugin
go build -buildmode=plugin -o my_plugin.so main.go

# 测试插件
go test -v
```

## 核心接口

### 基础插件接口
```go
type Plugin interface {
    Name() string
    Version() string
    Description() string
    Initialize(config map[string]interface{}) error
    Start(ctx context.Context) error
    Stop(ctx context.Context) error
    Cleanup() error
}
```

### 请求处理插件
```go
type RequestPlugin interface {
    Plugin
    ProcessRequest(req *http.Request, ctx *RequestContext) error
    Priority() int
}
```

### 响应处理插件
```go
type ResponsePlugin interface {
    Plugin
    ProcessResponse(resp *http.Response, req *http.Request, ctx *ResponseContext) error
    Priority() int
}
```

## 新框架钩子

### 钩子类型
```go
const (
    HookBeforeRequest  = "before_request"   // 请求前
    HookAfterRequest   = "after_request"    // 请求后
    HookBeforeResponse = "before_response"  // 响应前
    HookAfterResponse  = "after_response"   // 响应后
    HookOnError        = "on_error"         // 错误时
    HookOnFilter       = "on_filter"        // 过滤时
)
```

### 钩子函数
```go
type HookFunc func(ctx *HookContext) error

type HookContext struct {
    Request     *http.Request
    Response    *http.Response
    RequestCtx  *RequestContext
    ResponseCtx *ResponseContext
    Data        map[string]interface{}
    Logger      *logger.Logger
    Cancel      context.CancelFunc
}
```

## 工具库

### 请求工具
```go
// 获取客户端IP
clientIP := plugin.Utils.Request.GetClientIP(req)

// 检查内容类型
isJSON := plugin.Utils.Request.IsJSON(req)

// 读取请求体
body, err := plugin.Utils.Request.ReadBody(req)

// 解析JSON
var data map[string]interface{}
err := plugin.Utils.Request.ParseJSON(req, &data)
```

### 响应工具
```go
// 设置响应头
plugin.Utils.Response.SetHeader(resp, "X-Custom", "value")

// 设置JSON响应
plugin.Utils.Response.JSON(w, data, http.StatusOK)

// 读取响应体
body, err := plugin.Utils.Response.ReadBody(resp)
```

### 安全工具
```go
// SQL注入检查
if plugin.Utils.Security.IsSQLInjection(input) {
    // 处理SQL注入
}

// XSS检查
if plugin.Utils.Security.IsXSS(input) {
    // 处理XSS攻击
}

// 路径遍历检查
if plugin.Utils.Security.IsPathTraversal(path) {
    // 处理路径遍历
}

// 输入清理
cleaned := plugin.Utils.Security.SanitizeInput(input)
```

### 类型转换
```go
// 字符串转换
str := plugin.Utils.Conversion.ToString(value)

// 整数转换
num := plugin.Utils.Conversion.ToInt(value)

// 布尔值转换
flag := plugin.Utils.Conversion.ToBool(value)
```

## 配置验证

### 验证规则
```go
validator := plugin.NewConfigValidator()
validator.AddRule("timeout", plugin.ValidationRule{
    Required:    true,
    Type:        "string",
    Default:     "30s",
    Pattern:     `^\d+[smh]$`,
    Description: "超时时间",
})
```

### 规则类型
- `string`: 字符串
- `int`: 整数
- `bool`: 布尔值
- `float`: 浮点数
- `array`: 数组

### 验证选项
- `Required`: 是否必需
- `Default`: 默认值
- `Min/Max`: 最小/最大值
- `Pattern`: 正则表达式
- `Options`: 可选值列表

## HTTP API

### 监控端点
```bash
# 健康检查
curl http://localhost:9090/health

# 获取指标
curl http://localhost:9090/metrics

# 获取状态
curl http://localhost:9090/status
```

### 响应格式
```json
{
  "status": "healthy",
  "timestamp": "2023-07-14T08:31:13Z",
  "uptime": "1h30m",
  "requests": 1234,
  "errors": 5
}
```

## 错误处理

### 错误类型
```go
var (
    ErrPluginNotFound = errors.New("插件未找到")
    ErrInvalidConfig  = errors.New("配置无效")
    ErrTimeout        = errors.New("操作超时")
)
```

### 错误包装
```go
if err != nil {
    return fmt.Errorf("处理请求失败: %w", err)
}
```

### 错误检查
```go
if errors.Is(err, ErrTimeout) {
    // 处理超时错误
}
```

## 并发模式

### Goroutine管理
```go
var wg sync.WaitGroup
wg.Add(1)
go func() {
    defer wg.Done()
    // 执行任务
}()
wg.Wait()
```

### Context使用
```go
ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()

select {
case <-ctx.Done():
    return ctx.Err()
case result := <-ch:
    return result
}
```

### 互斥锁
```go
var mu sync.RWMutex

// 读锁
mu.RLock()
defer mu.RUnlock()

// 写锁
mu.Lock()
defer mu.Unlock()
```

## 性能优化

### 对象池
```go
var pool = sync.Pool{
    New: func() interface{} {
        return &MyStruct{}
    },
}

obj := pool.Get().(*MyStruct)
defer pool.Put(obj)
```

### 缓存策略
```go
type LRUCache struct {
    capacity int
    cache    map[string]*list.Element
    list     *list.List
    mutex    sync.RWMutex
}
```

## 调试技巧

### 日志级别
```go
logger.Debug("调试信息")
logger.Info("一般信息")
logger.Warn("警告信息")
logger.Error("错误信息")
```

### pprof分析
```bash
# CPU分析
go tool pprof http://localhost:6060/debug/pprof/profile

# 内存分析
go tool pprof http://localhost:6060/debug/pprof/heap
```

### 基准测试
```go
func BenchmarkMyFunction(b *testing.B) {
    for i := 0; i < b.N; i++ {
        MyFunction()
    }
}
```

## 常见问题

### Q: 插件加载失败？
```bash
# 检查插件路径
ls -la plugins/examples/

# 检查构建
go build -buildmode=plugin -o plugin.so main.go

# 检查符号
nm plugin.so | grep NewPlugin
```

### Q: 配置验证失败？
```go
// 检查配置格式
validator.AddRule("key", plugin.ValidationRule{
    Type:     "string",
    Required: true,
})
```

### Q: 内存泄漏？
```bash
# 内存分析
go tool pprof http://localhost:6060/debug/pprof/heap

# 检查goroutine
go tool pprof http://localhost:6060/debug/pprof/goroutine
```

## 最佳实践

### 1. 插件设计
- 保持插件职责单一
- 使用配置验证器
- 实现优雅关闭
- 添加监控指标

### 2. 错误处理
- 包装错误信息
- 使用特定错误类型
- 记录详细日志
- 实现重试机制

### 3. 性能优化
- 使用对象池
- 避免不必要的内存分配
- 合理使用缓存
- 监控关键指标

### 4. 安全考虑
- 验证所有输入
- 使用安全工具库
- 实现访问控制
- 定期安全审计

## 示例代码

### 简单插件
```go
package main

import (
    "context"
    "net/http"
    "hackmitm/pkg/plugin"
)

type SimplePlugin struct {
    *plugin.PluginFramework
}

func NewPlugin(config map[string]interface{}) (plugin.Plugin, error) {
    framework := plugin.NewPluginFramework(&plugin.FrameworkConfig{
        Name:        "simple-plugin",
        Version:     "1.0.0",
        Description: "简单插件示例",
    })
    
    p := &SimplePlugin{
        PluginFramework: framework,
    }
    
    p.AddHook(plugin.HookBeforeRequest, func(ctx *plugin.HookContext) error {
        ctx.Logger.Infof("处理请求: %s", ctx.Request.URL.String())
        return nil
    })
    
    return p, nil
}
```

### 配置文件
```json
{
  "server": {
    "listen_addr": ":8081",
    "monitor_addr": ":9090"
  },
  "plugins": [
    {
      "name": "simple-plugin",
      "path": "simple_plugin.so",
      "enabled": true,
      "config": {
        "log_level": "info"
      }
    }
  ]
}
```

这份快速参考手册涵盖了开发和使用HackMITM时最常用的信息，可以作为日常开发的快速查询工具。 