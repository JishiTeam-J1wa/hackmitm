# HackMITM 初学者学习手册

## 目录
- [前言](#前言)
- [基础知识](#基础知识)
- [项目结构解析](#项目结构解析)
- [核心概念理解](#核心概念理解)
- [逐步学习路径](#逐步学习路径)
- [实践练习](#实践练习)
- [常见问题解答](#常见问题解答)
- [进阶学习](#进阶学习)

## 前言

欢迎来到 HackMITM 的学习之旅！这个手册是专门为初学者设计的，无论你是Go语言新手还是网络编程初学者，都可以通过这个手册循序渐进地学习和理解这个项目。

### 学习目标

通过本手册，你将学会：
- 理解HTTP/HTTPS代理的工作原理
- 掌握Go语言的并发编程模式
- 学会插件系统的设计和实现
- 了解网络安全的基本概念
- 掌握项目的架构设计思路

### 学习前提

- 基础的Go语言知识（变量、函数、结构体、接口）
- 对HTTP协议有基本了解
- 熟悉命令行操作

## 基础知识

### 1. HTTP代理是什么？

HTTP代理是一个位于客户端和服务器之间的中间服务器，它接收客户端的请求，然后转发给目标服务器，再将服务器的响应返回给客户端。

```
客户端 → 代理服务器 → 目标服务器
客户端 ← 代理服务器 ← 目标服务器
```

### 2. HTTPS代理的特殊性

HTTPS代理需要处理TLS加密，通常使用CONNECT方法建立隧道：

```go
// CONNECT方法示例
func handleCONNECT(w http.ResponseWriter, r *http.Request) {
    // 1. 连接到目标服务器
    conn, err := net.Dial("tcp", r.Host)
    if err != nil {
        http.Error(w, err.Error(), http.StatusBadGateway)
        return
    }
    defer conn.Close()
    
    // 2. 告诉客户端连接已建立
    w.WriteHeader(http.StatusOK)
    
    // 3. 获取客户端连接
    hijacker, ok := w.(http.Hijacker)
    if !ok {
        http.Error(w, "不支持hijacking", http.StatusInternalServerError)
        return
    }
    
    clientConn, _, err := hijacker.Hijack()
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    defer clientConn.Close()
    
    // 4. 双向数据转发
    go io.Copy(conn, clientConn)
    io.Copy(clientConn, conn)
}
```

### 3. Go语言的并发模式

HackMITM大量使用了Go的并发特性：

#### Goroutine
```go
// 启动一个goroutine处理请求
go func() {
    defer wg.Done()
    handleRequest(req)
}()
```

#### Channel
```go
// 使用channel进行通信
type WorkerPool struct {
    tasks chan Task
    done  chan struct{}
}

func (wp *WorkerPool) worker() {
    for {
        select {
        case task := <-wp.tasks:
            task.Execute()
        case <-wp.done:
            return
        }
    }
}
```

#### 同步原语
```go
// 互斥锁保护共享资源
type SafeCounter struct {
    mu    sync.RWMutex
    count int64
}

func (c *SafeCounter) Inc() {
    c.mu.Lock()
    defer c.mu.Unlock()
    c.count++
}

func (c *SafeCounter) Get() int64 {
    c.mu.RLock()
    defer c.mu.RUnlock()
    return c.count
}
```

## 项目结构解析

让我们逐个目录了解项目结构：

### 根目录文件

```
HackMIT/
├── go.mod              # Go模块定义
├── go.sum              # 依赖校验和
├── Makefile            # 构建脚本
├── README.md           # 项目说明
├── LICENSE             # 许可证
├── SECURITY.md         # 安全政策
├── CHANGELOG.md        # 变更日志
├── CONTRIBUTING.md     # 贡献指南
└── docker-compose.yml  # Docker编排
```

### 核心目录

#### `cmd/` - 应用程序入口
```go
// cmd/hackmitm/main.go - 主程序入口
func main() {
    // 1. 解析命令行参数
    flag.Parse()
    
    // 2. 加载配置
    config, err := config.Load(*configFile)
    if err != nil {
        log.Fatal(err)
    }
    
    // 3. 创建服务器
    server, err := proxy.NewServer(config)
    if err != nil {
        log.Fatal(err)
    }
    
    // 4. 启动服务器
    if err := server.Start(); err != nil {
        log.Fatal(err)
    }
    
    // 5. 等待退出信号
    waitForSignal()
}
```

#### `pkg/` - 核心包
```
pkg/
├── config/         # 配置管理
├── proxy/          # 代理服务器
├── plugin/         # 插件系统
├── cert/           # 证书管理
├── security/       # 安全模块
├── monitor/        # 监控系统
├── logger/         # 日志系统
└── traffic/        # 流量处理
```

#### `plugins/` - 插件目录
```
plugins/
├── Makefile        # 插件构建脚本
└── examples/       # 示例插件
    ├── request_logger/     # 请求日志插件
    ├── security/           # 安全插件
    ├── stats/              # 统计插件
    └── simple_plugin_template/  # 简单插件模板
```

#### `configs/` - 配置文件
```json
{
  "server": {
    "listen_addr": ":8081",
    "monitor_addr": ":9090",
    "read_timeout": "30s",
    "write_timeout": "30s"
  },
  "plugins": [
    {
      "name": "request-logger",
      "path": "request_logger.so",
      "enabled": true,
      "config": {
        "log_file": "logs/requests.log"
      }
    }
  ]
}
```

## 核心概念理解

### 1. 配置系统

#### 配置结构定义
```go
// pkg/config/config.go
type Config struct {
    Server  ServerConfig  `json:"server"`   // 服务器配置
    Plugins []PluginConfig `json:"plugins"` // 插件配置
    Security SecurityConfig `json:"security"` // 安全配置
}

type ServerConfig struct {
    ListenAddr   string        `json:"listen_addr"`   // 监听地址
    MonitorAddr  string        `json:"monitor_addr"`  // 监控地址
    ReadTimeout  time.Duration `json:"read_timeout"`  // 读取超时
    WriteTimeout time.Duration `json:"write_timeout"` // 写入超时
}
```

#### 配置加载过程
```go
func Load(configFile string) (*Config, error) {
    // 1. 读取配置文件
    data, err := os.ReadFile(configFile)
    if err != nil {
        return nil, fmt.Errorf("读取配置文件失败: %w", err)
    }
    
    // 2. 解析JSON
    var config Config
    if err := json.Unmarshal(data, &config); err != nil {
        return nil, fmt.Errorf("解析配置失败: %w", err)
    }
    
    // 3. 验证配置
    if err := config.Validate(); err != nil {
        return nil, fmt.Errorf("配置验证失败: %w", err)
    }
    
    return &config, nil
}
```

### 2. 代理服务器核心

#### 服务器结构
```go
// pkg/proxy/server.go
type Server struct {
    config      *Config         // 配置
    listener    net.Listener    // 网络监听器
    server      *http.Server    // HTTP服务器
    pluginMgr   *plugin.Manager // 插件管理器
    certMgr     *cert.Manager   // 证书管理器
    securityMgr *security.AccessControl // 安全管理器
    monitor     *monitor.Metrics // 监控系统
    
    // 运行状态
    ctx    context.Context
    cancel context.CancelFunc
    wg     sync.WaitGroup
}
```

#### 请求处理流程
```go
func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
    // 1. 记录请求开始时间
    startTime := time.Now()
    
    // 2. 获取客户端信息
    clientIP := s.getClientIP(r)
    
    // 3. 安全检查
    if !s.securityMgr.CheckRequest(r) {
        http.Error(w, "请求被拒绝", http.StatusForbidden)
        return
    }
    
    // 4. 创建请求上下文
    ctx := &plugin.RequestContext{
        StartTime: startTime,
        ClientIP:  clientIP,
        Method:    r.Method,
        URL:       r.URL.String(),
    }
    
    // 5. 插件预处理
    if err := s.pluginMgr.ProcessRequest(r, ctx); err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }
    
    // 6. 根据方法分发处理
    if r.Method == "CONNECT" {
        s.handleHTTPS(w, r)
    } else {
        s.handleHTTP(w, r)
    }
}
```

### 3. 插件系统详解

#### 插件接口设计
```go
// pkg/plugin/interface.go
type Plugin interface {
    Name() string        // 插件名称
    Version() string     // 版本号
    Description() string // 描述
    
    // 生命周期方法
    Initialize(config map[string]interface{}) error
    Start(ctx context.Context) error
    Stop(ctx context.Context) error
    Cleanup() error
}

// 请求处理插件
type RequestPlugin interface {
    Plugin
    ProcessRequest(req *http.Request, ctx *RequestContext) error
    Priority() int // 优先级，数字越小优先级越高
}
```

#### 插件管理器
```go
// pkg/plugin/manager.go
type Manager struct {
    plugins       map[string]*PluginWrapper
    pluginsByType map[PluginType][]*PluginWrapper
    config        map[string]*PluginConfig
    mutex         sync.RWMutex
}

func (m *Manager) LoadPlugin(config *PluginConfig) error {
    // 1. 加载动态库
    p, err := plugin.Open(config.Path)
    if err != nil {
        return err
    }
    
    // 2. 查找工厂函数
    factory, err := p.Lookup("NewPlugin")
    if err != nil {
        return err
    }
    
    // 3. 创建插件实例
    pluginInstance, err := factory.(func(map[string]interface{}) (Plugin, error))(config.Config)
    if err != nil {
        return err
    }
    
    // 4. 注册插件
    m.registerPlugin(pluginInstance, config)
    
    return nil
}
```

### 4. 新插件框架

#### 钩子系统
```go
// pkg/plugin/framework.go
type PluginFramework struct {
    *BasePlugin
    hooks      map[string][]HookFunc
    middleware []MiddlewareFunc
    config     *FrameworkConfig
}

// 钩子函数类型
type HookFunc func(ctx *HookContext) error

// 钩子上下文
type HookContext struct {
    Request  *http.Request
    Response *http.Response
    Data     map[string]interface{}
    Logger   *logger.Logger
}

// 添加钩子
func (pf *PluginFramework) AddHook(hookType HookType, hook HookFunc) {
    pf.hooks[string(hookType)] = append(pf.hooks[string(hookType)], hook)
}
```

#### 配置验证器
```go
// pkg/plugin/validator.go
type ConfigValidator struct {
    rules map[string]ValidationRule
}

type ValidationRule struct {
    Required    bool
    Type        string
    Default     interface{}
    Min         interface{}
    Max         interface{}
    Pattern     string
    Options     []interface{}
    Description string
}

func (cv *ConfigValidator) Validate(config map[string]interface{}) (map[string]interface{}, error) {
    result := make(map[string]interface{})
    
    for key, rule := range cv.rules {
        value, exists := config[key]
        
        // 检查必需项
        if rule.Required && !exists {
            return nil, fmt.Errorf("必需配置项 %s 缺失", key)
        }
        
        // 使用默认值
        if !exists && rule.Default != nil {
            result[key] = rule.Default
            continue
        }
        
        // 验证值
        validatedValue, err := cv.validateValue(key, value, rule)
        if err != nil {
            return nil, err
        }
        
        result[key] = validatedValue
    }
    
    return result, nil
}
```

## 逐步学习路径

### 第一阶段：理解基础结构

#### 1. 从main函数开始
```go
// 学习目标：理解程序启动流程
func main() {
    // 解析命令行参数
    flag.Parse()
    
    // 加载配置
    config, err := config.Load(*configFile)
    if err != nil {
        log.Fatal(err)
    }
    
    // 创建服务器
    server, err := proxy.NewServer(config)
    if err != nil {
        log.Fatal(err)
    }
    
    // 启动服务器
    if err := server.Start(); err != nil {
        log.Fatal(err)
    }
    
    // 等待信号
    waitForSignal()
}
```

**学习要点：**
- 理解Go程序的入口点
- 学会使用flag包处理命令行参数
- 了解错误处理的基本模式
- 理解程序的生命周期管理

#### 2. 配置系统学习
```go
// 学习目标：理解配置的加载和验证
type Config struct {
    Server   ServerConfig   `json:"server"`
    Plugins  []PluginConfig `json:"plugins"`
    Security SecurityConfig `json:"security"`
}

func (c *Config) Validate() error {
    // 验证服务器配置
    if c.Server.ListenAddr == "" {
        return errors.New("监听地址不能为空")
    }
    
    // 验证插件配置
    for _, plugin := range c.Plugins {
        if plugin.Name == "" {
            return errors.New("插件名称不能为空")
        }
    }
    
    return nil
}
```

**学习要点：**
- 学会使用结构体标签进行JSON绑定
- 理解配置验证的重要性
- 掌握错误处理的最佳实践

### 第二阶段：深入代理服务器

#### 1. HTTP服务器创建
```go
// 学习目标：理解HTTP服务器的创建和配置
func NewServer(config *Config) (*Server, error) {
    s := &Server{
        config: config,
    }
    
    // 创建HTTP服务器
    s.server = &http.Server{
        Addr:         config.Server.ListenAddr,
        Handler:      s,  // 服务器本身实现了Handler接口
        ReadTimeout:  config.Server.ReadTimeout,
        WriteTimeout: config.Server.WriteTimeout,
    }
    
    return s, nil
}
```

**学习要点：**
- 理解http.Server的配置
- 学会实现http.Handler接口
- 了解超时设置的重要性

#### 2. 请求处理机制
```go
// 学习目标：理解HTTP请求的处理流程
func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
    // 这是所有HTTP请求的入口点
    
    // 1. 记录请求信息
    log.Printf("收到请求: %s %s", r.Method, r.URL)
    
    // 2. 根据请求方法分发
    switch r.Method {
    case "CONNECT":
        s.handleHTTPS(w, r)
    default:
        s.handleHTTP(w, r)
    }
}
```

**学习要点：**
- 理解HTTP请求的生命周期
- 学会区分不同的HTTP方法
- 掌握请求分发的模式

### 第三阶段：插件系统理解

#### 1. 插件接口设计
```go
// 学习目标：理解接口设计的思想
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

**学习要点：**
- 理解接口的抽象能力
- 学会设计生命周期方法
- 掌握上下文的使用

#### 2. 动态插件加载
```go
// 学习目标：理解Go的plugin机制
func (m *Manager) LoadPlugin(config *PluginConfig) error {
    // 1. 加载动态库
    p, err := plugin.Open(config.Path)
    if err != nil {
        return fmt.Errorf("加载插件失败: %w", err)
    }
    
    // 2. 查找符号
    symbol, err := p.Lookup("NewPlugin")
    if err != nil {
        return fmt.Errorf("未找到插件入口: %w", err)
    }
    
    // 3. 类型断言
    factory, ok := symbol.(func(map[string]interface{}) (Plugin, error))
    if !ok {
        return errors.New("插件入口类型不匹配")
    }
    
    // 4. 创建实例
    instance, err := factory(config.Config)
    if err != nil {
        return fmt.Errorf("创建插件实例失败: %w", err)
    }
    
    return nil
}
```

**学习要点：**
- 理解Go的plugin包
- 学会使用类型断言
- 掌握动态加载的机制

### 第四阶段：新框架学习

#### 1. 钩子系统
```go
// 学习目标：理解钩子模式的实现
func (pf *PluginFramework) AddHook(hookType HookType, hook HookFunc) {
    pf.mutex.Lock()
    defer pf.mutex.Unlock()
    
    key := string(hookType)
    pf.hooks[key] = append(pf.hooks[key], hook)
}

func (pf *PluginFramework) executeHooks(hookType HookType, ctx *HookContext) error {
    pf.mutex.RLock()
    hooks := pf.hooks[string(hookType)]
    pf.mutex.RUnlock()
    
    for _, hook := range hooks {
        if err := hook(ctx); err != nil {
            return err
        }
    }
    
    return nil
}
```

**学习要点：**
- 理解钩子模式的设计思想
- 学会使用读写锁保护并发访问
- 掌握函数作为一等公民的特性

#### 2. 中间件系统
```go
// 学习目标：理解中间件的链式调用
type MiddlewareFunc func(next HookFunc) HookFunc

func (pf *PluginFramework) AddMiddleware(middleware MiddlewareFunc) {
    pf.middleware = append(pf.middleware, middleware)
}

func (pf *PluginFramework) executeWithMiddleware(hook HookFunc, ctx *HookContext) error {
    // 从后往前应用中间件
    finalHook := hook
    for i := len(pf.middleware) - 1; i >= 0; i-- {
        finalHook = pf.middleware[i](finalHook)
    }
    
    return finalHook(ctx)
}
```

**学习要点：**
- 理解中间件的装饰器模式
- 学会函数式编程的思想
- 掌握链式调用的实现

## 实践练习

### 练习1：创建简单的HTTP代理

#### 目标
创建一个最基本的HTTP代理服务器

#### 代码框架
```go
package main

import (
    "io"
    "log"
    "net/http"
    "net/url"
)

func main() {
    // TODO: 实现一个简单的HTTP代理
    http.HandleFunc("/", proxyHandler)
    log.Println("代理服务器启动在 :8080")
    log.Fatal(http.ListenAndServe(":8080", nil))
}

func proxyHandler(w http.ResponseWriter, r *http.Request) {
    // TODO: 实现代理逻辑
    // 1. 解析目标URL
    // 2. 创建新的请求
    // 3. 发送请求到目标服务器
    // 4. 将响应返回给客户端
}
```

#### 参考实现
```go
func proxyHandler(w http.ResponseWriter, r *http.Request) {
    // 1. 解析目标URL
    targetURL, err := url.Parse(r.URL.String())
    if err != nil {
        http.Error(w, "无效的URL", http.StatusBadRequest)
        return
    }
    
    // 2. 创建新的请求
    proxyReq, err := http.NewRequest(r.Method, targetURL.String(), r.Body)
    if err != nil {
        http.Error(w, "创建请求失败", http.StatusInternalServerError)
        return
    }
    
    // 复制请求头
    for key, values := range r.Header {
        for _, value := range values {
            proxyReq.Header.Add(key, value)
        }
    }
    
    // 3. 发送请求
    client := &http.Client{}
    resp, err := client.Do(proxyReq)
    if err != nil {
        http.Error(w, "请求失败", http.StatusBadGateway)
        return
    }
    defer resp.Body.Close()
    
    // 4. 复制响应头
    for key, values := range resp.Header {
        for _, value := range values {
            w.Header().Add(key, value)
        }
    }
    
    // 5. 设置状态码
    w.WriteHeader(resp.StatusCode)
    
    // 6. 复制响应体
    io.Copy(w, resp.Body)
}
```

### 练习2：实现简单的插件

#### 目标
创建一个记录请求信息的插件

#### 插件结构
```go
// plugins/examples/my_logger/logger.go
package main

import (
    "context"
    "log"
    "net/http"
    "os"
)

type MyLogger struct {
    name        string
    version     string
    description string
    logFile     *os.File
}

func NewPlugin(config map[string]interface{}) (Plugin, error) {
    logger := &MyLogger{
        name:        "my-logger",
        version:     "1.0.0",
        description: "我的第一个插件",
    }
    
    // 打开日志文件
    filename := config["log_file"].(string)
    file, err := os.OpenFile(filename, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
    if err != nil {
        return nil, err
    }
    logger.logFile = file
    
    return logger, nil
}

func (l *MyLogger) Name() string        { return l.name }
func (l *MyLogger) Version() string     { return l.version }
func (l *MyLogger) Description() string { return l.description }

func (l *MyLogger) Initialize(config map[string]interface{}) error {
    log.Println("初始化MyLogger插件")
    return nil
}

func (l *MyLogger) Start(ctx context.Context) error {
    log.Println("启动MyLogger插件")
    return nil
}

func (l *MyLogger) Stop(ctx context.Context) error {
    log.Println("停止MyLogger插件")
    return nil
}

func (l *MyLogger) Cleanup() error {
    if l.logFile != nil {
        l.logFile.Close()
    }
    return nil
}

func (l *MyLogger) ProcessRequest(req *http.Request, ctx *RequestContext) error {
    // 记录请求信息
    logEntry := fmt.Sprintf("[%s] %s %s %s %s\n",
        ctx.StartTime.Format("2006-01-02 15:04:05"),
        ctx.ClientIP,
        req.Method,
        req.URL.String(),
        req.UserAgent())
    
    l.logFile.WriteString(logEntry)
    return nil
}

func (l *MyLogger) Priority() int {
    return 100 // 较低优先级
}
```

### 练习3：使用新框架创建插件

#### 目标
使用新的插件框架创建一个更复杂的插件

#### 代码实现
```go
package main

import (
    "context"
    "fmt"
    "net/http"
    "time"
    
    "hackmitm/pkg/plugin"
)

type AdvancedPlugin struct {
    *plugin.PluginFramework
    requestCount int64
    startTime    time.Time
}

func NewPlugin(config map[string]interface{}) (plugin.Plugin, error) {
    // 创建配置验证器
    validator := plugin.NewConfigValidator()
    validator.AddRule("max_requests", plugin.ValidationRule{
        Type:        "int",
        Default:     1000,
        Min:         1,
        Description: "最大请求数限制",
    })
    
    // 验证配置
    validatedConfig, err := validator.Validate(config)
    if err != nil {
        return nil, err
    }
    
    // 创建插件框架
    framework := plugin.NewPluginFramework(&plugin.FrameworkConfig{
        Name:        "advanced-plugin",
        Version:     "1.0.0",
        Description: "高级插件示例",
        Timeout:     30 * time.Second,
    })
    
    p := &AdvancedPlugin{
        PluginFramework: framework,
        startTime:       time.Now(),
    }
    
    // 注册钩子
    p.AddHook(plugin.HookBeforeRequest, p.beforeRequestHook)
    p.AddHook(plugin.HookAfterRequest, p.afterRequestHook)
    
    // 添加中间件
    p.AddMiddleware(p.loggingMiddleware)
    p.AddMiddleware(p.timingMiddleware)
    
    return p, nil
}

func (p *AdvancedPlugin) beforeRequestHook(ctx *plugin.HookContext) error {
    // 检查请求数量限制
    p.requestCount++
    if p.requestCount > 1000 {
        return fmt.Errorf("请求数量超过限制")
    }
    
    // 添加自定义头
    ctx.Request.Header.Set("X-Processed-By", "AdvancedPlugin")
    
    return nil
}

func (p *AdvancedPlugin) afterRequestHook(ctx *plugin.HookContext) error {
    // 记录处理时间
    duration := time.Since(ctx.RequestCtx.StartTime)
    ctx.Logger.Infof("请求处理完成，耗时: %v", duration)
    
    return nil
}

func (p *AdvancedPlugin) loggingMiddleware(next plugin.HookFunc) plugin.HookFunc {
    return func(ctx *plugin.HookContext) error {
        ctx.Logger.Infof("开始执行钩子")
        err := next(ctx)
        if err != nil {
            ctx.Logger.Errorf("钩子执行失败: %v", err)
        } else {
            ctx.Logger.Infof("钩子执行成功")
        }
        return err
    }
}

func (p *AdvancedPlugin) timingMiddleware(next plugin.HookFunc) plugin.HookFunc {
    return func(ctx *plugin.HookContext) error {
        start := time.Now()
        err := next(ctx)
        duration := time.Since(start)
        ctx.Data["hook_duration"] = duration
        return err
    }
}
```

## 常见问题解答

### Q1: 为什么使用接口而不是直接使用结构体？

**A:** 接口提供了抽象层，使得：
- 代码更容易测试（可以创建mock对象）
- 系统更容易扩展（可以有多种实现）
- 组件之间耦合度更低

```go
// 使用接口
type Logger interface {
    Log(message string)
}

// 可以有多种实现
type FileLogger struct{}
func (f *FileLogger) Log(message string) { /* 写入文件 */ }

type ConsoleLogger struct{}
func (c *ConsoleLogger) Log(message string) { /* 输出到控制台 */ }
```

### Q2: 为什么要使用context.Context？

**A:** Context提供了：
- 取消信号传递
- 超时控制
- 请求范围的值传递

```go
func processRequest(ctx context.Context) error {
    select {
    case <-ctx.Done():
        return ctx.Err() // 处理取消或超时
    default:
        // 正常处理
    }
}
```

### Q3: 什么时候使用sync.RWMutex而不是sync.Mutex？

**A:** 当读操作远多于写操作时：
```go
type Cache struct {
    mu    sync.RWMutex
    data  map[string]interface{}
}

func (c *Cache) Get(key string) interface{} {
    c.mu.RLock()         // 读锁，允许并发读
    defer c.mu.RUnlock()
    return c.data[key]
}

func (c *Cache) Set(key string, value interface{}) {
    c.mu.Lock()          // 写锁，独占访问
    defer c.mu.Unlock()
    c.data[key] = value
}
```

### Q4: 如何处理goroutine的生命周期？

**A:** 使用WaitGroup和Context：
```go
func (s *Server) Start() error {
    s.wg.Add(1)
    go func() {
        defer s.wg.Done()
        s.worker()
    }()
    return nil
}

func (s *Server) Stop() error {
    s.cancel()  // 发送取消信号
    s.wg.Wait() // 等待所有goroutine结束
    return nil
}

func (s *Server) worker() {
    for {
        select {
        case <-s.ctx.Done():
            return  // 收到取消信号
        default:
            // 执行工作
        }
    }
}
```

### Q5: 如何设计好的错误处理？

**A:** 遵循以下原则：
```go
// 1. 包装错误，提供上下文
func processFile(filename string) error {
    file, err := os.Open(filename)
    if err != nil {
        return fmt.Errorf("打开文件 %s 失败: %w", filename, err)
    }
    defer file.Close()
    
    // 处理文件...
    return nil
}

// 2. 定义特定的错误类型
var (
    ErrFileNotFound = errors.New("文件未找到")
    ErrInvalidFormat = errors.New("文件格式无效")
)

// 3. 使用errors.Is检查错误
if errors.Is(err, ErrFileNotFound) {
    // 处理文件未找到的情况
}
```

## 进阶学习

### 1. 深入学习Go并发编程

#### 推荐资源
- 《Go并发编程实战》
- Go官方博客关于并发的文章
- 实践项目：实现一个工作池

#### 学习重点
- Channel的各种用法
- Select语句的使用
- 并发安全的数据结构
- 性能优化技巧

### 2. 网络编程深入

#### 学习内容
- TCP/UDP编程
- HTTP/2和HTTP/3
- WebSocket编程
- TLS/SSL原理

#### 实践项目
- 实现一个简单的Web服务器
- 创建WebSocket聊天室
- 实现负载均衡器

### 3. 系统设计和架构

#### 学习重点
- 微服务架构
- 分布式系统设计
- 缓存策略
- 监控和日志

#### 实践建议
- 设计一个完整的代理系统
- 实现分布式插件管理
- 添加监控和告警功能

### 4. 安全编程

#### 学习内容
- 常见的安全漏洞
- 加密和解密
- 认证和授权
- 安全编码实践

#### 实践项目
- 实现JWT认证
- 添加SQL注入防护
- 实现访问控制列表

通过这个初学者手册，你应该能够逐步理解和掌握HackMITM项目的各个方面。记住，学习编程最重要的是实践，建议你跟着手册一步步实现这些功能，这样能够更好地理解代码的工作原理。 