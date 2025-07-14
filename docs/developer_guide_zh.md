# HackMITM 开发者指南

## 📋 目录
- [项目概述](#项目概述)
- [架构设计](#架构设计)
- [核心模块详解](#核心模块详解)
- [指纹识别系统](#指纹识别系统)
- [性能优化](#性能优化)
- [插件系统深入](#插件系统深入)
- [API参考](#api参考)
- [开发环境搭建](#开发环境搭建)
- [代码贡献指南](#代码贡献指南)
- [调试和测试](#调试和测试)
- [安全考虑](#安全考虑)
- [部署和运维](#部署和运维)

## 🎯 项目概述

HackMITM 是一个企业级高性能 HTTP/HTTPS 代理服务器，专为安全研究、流量分析和网络调试而设计。采用现代化的Go语言开发，具备以下核心特性：

### 🚀 核心特性
- **高性能代理**：支持 HTTP/HTTPS/WebSocket 代理，处理大规模并发流量
- **智能指纹识别**：内置24,981条指纹规则，支持Web应用识别
- **分层索引系统**：三层过滤架构，查询复杂度从O(n)降至O(log n)
- **LRU缓存机制**：智能缓存系统，支持TTL和自动清理
- **插件系统**：灵活的插件架构，支持动态加载和热更新
- **安全防护**：内置多种安全检测机制和访问控制
- **监控系统**：实时监控和指标收集，支持Prometheus
- **证书管理**：自动 HTTPS 证书生成和管理

### 🛠️ 技术栈
- **语言**：Go 1.21+
- **架构**：模块化设计 + 插件系统
- **并发**：Goroutine + Channel + 协程池
- **缓存**：LRU + TTL + 分层索引
- **监控**：Prometheus + 自定义指标
- **存储**：内存 + 文件系统
- **网络**：标准库 + 自定义优化

## 🏗️ 架构设计

### 整体架构

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                           HackMITM 架构图                                   │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                             │
│  ┌─────────────┐    ┌─────────────┐    ┌─────────────┐    ┌─────────────┐  │
│  │   Client    │    │   Proxy     │    │  Fingerprint│    │   Plugin    │  │
│  │  Requests   │───▶│   Server    │◄──▶│   Engine    │◄──▶│   Manager   │  │
│  └─────────────┘    └─────────────┘    └─────────────┘    └─────────────┘  │
│                            │                   │                   │        │
│                            ▼                   ▼                   ▼        │
│  ┌─────────────┐    ┌─────────────┐    ┌─────────────┐    ┌─────────────┐  │
│  │   Target    │    │   Security  │    │  Layered    │    │   Buffer    │  │
│  │   Servers   │◄───│   Manager   │    │   Index     │    │    Pool     │  │
│  └─────────────┘    └─────────────┘    └─────────────┘    └─────────────┘  │
│                            │                   │                   │        │
│                            ▼                   ▼                   ▼        │
│  ┌─────────────┐    ┌─────────────┐    ┌─────────────┐    ┌─────────────┐  │
│  │   Monitor   │    │   Config    │    │  LRU Cache  │    │   Logger    │  │
│  │   System    │    │   Manager   │    │   System    │    │   System    │  │
│  └─────────────┘    └─────────────┘    └─────────────┘    └─────────────┘  │
│                                                                             │
└─────────────────────────────────────────────────────────────────────────────┘
```

### 数据流向

```
Request Flow:
Client → Proxy Server → Security Check → Plugin Processing → Target Server
                ↓              ↓              ↓
         Fingerprint     Cache Check    Monitor Update
         Recognition    (LRU + TTL)    (Metrics + Logs)
                ↓              ↓              ↓
         Layered Index   Cache Update   Alert System
         (3-Layer)      (Statistics)   (Notifications)
```

## 🔧 核心模块详解

### 1. 代理服务器 (`pkg/proxy/server.go`)

代理服务器是系统的核心组件，负责处理所有的HTTP/HTTPS请求：

```go
type ProxyServer struct {
    config          *config.Config
    server          *http.Server
    certManager     *cert.Manager
    pluginManager   *plugin.Manager
    securityManager *security.AccessControl
    monitor         *monitor.Metrics
    fingerprintEngine *fingerprint.Engine
    bufferPool      *pool.BufferPool
}
```

**关键特性**：
- 支持HTTP/HTTPS/WebSocket代理
- 自动证书生成和管理
- 请求/响应拦截和修改
- 插件系统集成
- 性能监控和指标收集

### 2. 配置管理 (`pkg/config/config.go`)

统一的配置管理系统，支持动态配置更新：

```go
type Config struct {
    Server      ServerConfig      `json:"server"`
    TLS         TLSConfig         `json:"tls"`
    Proxy       ProxyConfig       `json:"proxy"`
    Security    SecurityConfig    `json:"security"`
    Monitoring  MonitoringConfig  `json:"monitoring"`
    Plugins     PluginsConfig     `json:"plugins"`
    Fingerprint FingerprintConfig `json:"fingerprint"`
}
```

### 3. 监控系统 (`pkg/monitor/metrics.go`)

实时监控和指标收集系统：

```go
type Metrics struct {
    requestCount     prometheus.Counter
    responseTime     prometheus.Histogram
    errorCount       prometheus.Counter
    cacheHitRate     prometheus.Gauge
    fingerprintStats prometheus.GaugeVec
    systemMetrics    prometheus.GaugeVec
}
```

## 🔍 指纹识别系统

### 系统概述

指纹识别系统是HackMITM的核心功能之一，能够识别Web应用的技术栈、框架和服务。系统包含24,981条指纹规则，支持高效的实时识别。

### 核心组件

#### 1. 指纹引擎 (`pkg/fingerprint/fingerprint.go`)

```go
type Engine struct {
    rules         []FingerprintRule
    cache         *LRUCache
    layeredIndex  *LayeredIndex
    compiledRegex map[string]*regexp.Regexp
    stats         *Statistics
    config        *FingerprintConfig
}
```

#### 2. LRU缓存系统 (`pkg/fingerprint/lru_cache.go`)

智能缓存系统，支持TTL和自动清理：

```go
type LRUCache struct {
    capacity    int
    ttl         time.Duration
    cache       map[string]*CacheItem
    lruList     *list.List
    mutex       sync.RWMutex
    cleanupTicker *time.Ticker
    stats       *CacheStats
}
```

**特性**：
- 容量限制和LRU淘汰策略
- TTL过期自动清理
- 线程安全的并发访问
- 详细的统计信息

#### 3. 分层索引系统 (`pkg/fingerprint/layered_index.go`)

三层过滤架构，大幅提升查询性能：

```go
type LayeredIndex struct {
    // Layer 1: 快速过滤
    headerIndex    map[string][]int
    statusIndex    map[int][]int
    pathIndex      map[string][]int
    
    // Layer 2: 内容特征
    keywordIndex   map[string][]int
    titleIndex     map[string][]int
    
    // Layer 3: 深度匹配
    faviconRules   []int
    regexRules     []int
    
    mutex          sync.RWMutex
    stats          *IndexStats
}
```

**层级设计**：
- **Layer 1（快速过滤）**：HTTP头、状态码、URL路径
- **Layer 2（内容特征）**：标题关键词、正文内容
- **Layer 3（深度匹配）**：正则表达式、favicon匹配

### 性能优化

#### 1. 查询复杂度优化

```
传统方式：O(n) - 遍历所有规则
分层索引：O(log n) - 基于索引快速定位
```

#### 2. 缓存策略

```go
// 缓存配置
type CacheConfig struct {
    Size        int           `json:"cache_size"`        // 缓存大小：2000
    TTL         time.Duration `json:"cache_ttl"`         // TTL：1800秒
    CleanupInterval time.Duration                        // 清理间隔：600秒
}
```

#### 3. 并发优化

- 读写锁分离
- 协程池管理
- 内存池复用

## ⚡ 性能优化

### 1. 内存优化

#### 缓冲池系统 (`pkg/pool/buffer_pool.go`)

```go
type BufferPool struct {
    pool sync.Pool
    size int
}

func (p *BufferPool) Get() *bytes.Buffer {
    return p.pool.Get().(*bytes.Buffer)
}

func (p *BufferPool) Put(buf *bytes.Buffer) {
    buf.Reset()
    p.pool.Put(buf)
}
```

#### 内存分配策略

- 预分配缓冲区
- 对象池复用
- 垃圾回收优化

### 2. 并发优化

#### 协程池管理

```go
type WorkerPool struct {
    workerCount int
    jobQueue    chan Job
    workers     []*Worker
    wg          sync.WaitGroup
}
```

#### 锁优化

- 读写锁分离
- 细粒度锁
- 无锁数据结构

### 3. 网络优化

#### 连接池

```go
type ConnectionPool struct {
    maxIdle     int
    maxActive   int
    idleTimeout time.Duration
    connections chan net.Conn
}
```

#### 传输优化

- HTTP/2 支持
- 连接复用
- 压缩传输

## 🔌 插件系统深入

### 插件架构

```go
type Plugin interface {
    Name() string
    Version() string
    Init(config map[string]interface{}) error
    ProcessRequest(req *http.Request) error
    ProcessResponse(resp *http.Response) error
    Cleanup() error
}
```

### 插件管理器

```go
type Manager struct {
    plugins     map[string]Plugin
    config      *PluginsConfig
    basePath    string
    mutex       sync.RWMutex
}
```

### 插件开发指南

1. **实现Plugin接口**
2. **配置插件元数据**
3. **处理请求/响应**
4. **错误处理和日志**
5. **性能考虑**

## 📚 API参考

### 核心API

#### 1. 指纹识别API

```go
// 识别单个请求
func (e *Engine) IdentifyFingerprint(req *http.Request, resp *http.Response) []FingerprintResult

// 批量识别
func (e *Engine) BatchIdentify(requests []RequestResponse) [][]FingerprintResult

// 获取统计信息
func (e *Engine) GetStatistics() *Statistics
```

#### 2. 缓存API

```go
// 获取缓存项
func (c *LRUCache) Get(key string) (interface{}, bool)

// 设置缓存项
func (c *LRUCache) Set(key string, value interface{})

// 清理缓存
func (c *LRUCache) Clear()

// 获取统计信息
func (c *LRUCache) GetStats() *CacheStats
```

#### 3. 监控API

```go
// 记录请求
func (m *Metrics) RecordRequest(method, path string, duration time.Duration)

// 记录错误
func (m *Metrics) RecordError(errorType string)

// 更新缓存指标
func (m *Metrics) UpdateCacheMetrics(hitRate float64, size int)
```

## 🛠️ 开发环境搭建

### 环境要求

- Go 1.21+
- Git
- Make
- Docker (可选)

### 快速开始

```bash
# 克隆项目
git clone https://github.com/JishiTeam-J1wa/hackmitm.git
cd hackmitm

# 安装依赖
go mod tidy

# 构建项目
make build

# 运行测试
make test

# 启动服务
./hackmitm
```

### 开发工具

推荐使用以下工具：

- **IDE**: GoLand, VS Code
- **调试**: Delve
- **性能分析**: pprof
- **代码质量**: golangci-lint

## 🤝 代码贡献指南

### 贡献流程

1. **Fork项目**
2. **创建特性分支**
3. **编写代码和测试**
4. **提交Pull Request**
5. **代码审查**
6. **合并代码**

### 代码规范

- 遵循Go官方代码规范
- 使用golangci-lint检查
- 编写单元测试
- 添加文档注释

### 提交规范

```
feat: 添加新功能
fix: 修复bug
docs: 更新文档
style: 代码格式化
refactor: 重构代码
test: 添加测试
chore: 构建相关
```

## 🧪 调试和测试

### 单元测试

```bash
# 运行所有测试
go test ./...

# 运行特定包测试
go test ./pkg/fingerprint

# 生成覆盖率报告
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

### 性能测试

```bash
# 基准测试
go test -bench=. ./pkg/fingerprint

# 内存分析
go test -memprofile=mem.prof ./pkg/fingerprint
go tool pprof mem.prof
```

### 调试技巧

- 使用pprof进行性能分析
- 添加详细的日志
- 使用断点调试
- 监控系统指标

## 🔒 安全考虑

### 安全特性

- 访问控制和认证
- 输入验证和过滤
- 安全头设置
- 证书管理
- 日志脱敏

### 最佳实践

1. **输入验证**：严格验证所有输入
2. **权限控制**：最小权限原则
3. **数据加密**：敏感数据加密存储
4. **日志安全**：避免记录敏感信息
5. **定期更新**：及时更新依赖

## 🚀 部署和运维

### 部署方式

#### 1. 二进制部署

```bash
# 构建
make build

# 配置
cp configs/config.json.example configs/config.json

# 启动
./hackmitm
```

#### 2. Docker部署

```bash
# 构建镜像
docker build -t hackmitm .

# 运行容器
docker run -p 8081:8081 -v ./configs:/app/configs hackmitm
```

#### 3. Docker Compose

```bash
docker-compose up -d
```

### 监控和运维

#### 1. 健康检查

```bash
# 健康检查端点
curl http://localhost:9090/health

# 指标端点
curl http://localhost:9090/metrics
```

#### 2. 日志管理

- 结构化日志
- 日志轮转
- 集中式日志收集

#### 3. 性能监控

- CPU和内存使用率
- 请求处理时间
- 缓存命中率
- 错误率统计

### 故障排除

#### 常见问题

1. **内存泄漏**：检查缓存配置和对象池
2. **性能问题**：分析pprof输出
3. **连接问题**：检查网络配置
4. **证书问题**：验证证书配置

#### 调试命令

```bash
# 查看运行状态
ps aux | grep hackmitm

# 查看端口占用
netstat -tlnp | grep :8081

# 查看日志
tail -f logs/hackmitm.log
```

## 📞 支持和联系

- **项目主页**：https://github.com/JishiTeam-J1wa/hackmitm
- **文档**：https://github.com/JishiTeam-J1wa/hackmitm/docs
- **Issue报告**：https://github.com/JishiTeam-J1wa/hackmitm/issues
- **邮箱**：admin@jishiteam.com

---

**注意**：本项目仅供合法的安全研究和教育目的使用，请遵守相关法律法规。 