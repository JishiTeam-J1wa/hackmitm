# HackMITM 代理稳定性分析报告

**分析日期**: 2025-02-24
**重点场景**: 大量 JS 文件加载、高并发请求

---

## 🔴 发现的潜在问题

### 问题 1: bodyBuffer 无实际大小限制 (严重)

**位置**: `pkg/proxy/server.go:572`

```go
bodyBuffer = make([]byte, 0, 1024*1024) // 1MB限制 - 这只是初始容量！
teeReader := io.TeeReader(resp.Body, &bodyWriter{&bodyBuffer})
```

**问题**:
- `make([]byte, 0, 1024*1024)` 只是设置初始容量为 1MB
- `bodyWriter.Write()` 使用 `append()` 会自动扩容
- **实际没有 1MB 限制**，响应体可以是任意大小

**影响**: 加载大文件或大型 JS bundle 时，内存会无限增长

**风险等级**: 🔴 严重

---

### 问题 2: 请求体读取无限制 (严重)

**位置**: `pkg/proxy/server.go:704-709`

```go
if r.Body != nil {
    if data, err := io.ReadAll(r.Body); err == nil {
        body = data  // 没有大小限制！
        r.Body = io.NopCloser(strings.NewReader(string(data)))
    }
}
```

**问题**:
- `io.ReadAll()` 会读取整个请求体到内存
- 没有大小限制，恶意客户端可发送超大请求体

**影响**: 可能导致 OOM (Out of Memory)

**风险等级**: 🔴 严重

---

### 问题 3: Goroutine 泄漏风险 (中等)

**位置**: `pkg/proxy/server.go:580, 682`

```go
// 每个请求都启动一个新的 goroutine
go s.fingerprintHandler.HandleRequest(r, resp, bodyBuffer)
```

**问题**:
- 每个请求都启动一个 goroutine 进行指纹识别
- 没有并发控制（semaphore 或 worker pool）
- 高并发时可能产生数万个 goroutine

**场景模拟**:
```
访问一个包含 100 个 JS/CSS 资源的页面
→ 产生 100 个并发 goroutine
→ 指纹识别每个资源（即使大多是静态资源）
→ 内存和 CPU 压力增大
```

**风险等级**: 🟡 中等

---

### 问题 4: 静态资源也进行指纹识别 (性能)

**位置**: `pkg/proxy/server.go:571-580`

**问题**:
- 所有请求（包括 JS/CSS/图片）都会进行指纹识别
- 静态资源指纹识别意义不大，浪费资源

**场景**: 现代网页通常包含 50-200 个 JS/CSS/图片资源

**风险等级**: 🟡 中等

---

### 问题 5: Processor 的 maxBodySize 未生效

**位置**: `pkg/traffic/processor.go:93-95`

```go
func (p *Processor) ProcessRequest(req *http.Request) error {
    if req.ContentLength > p.maxBodySize {
        return fmt.Errorf("请求体过大: %d bytes", req.ContentLength)
    }
    // ...
}
```

**问题**:
- 只检查了 `ContentLength` 头
- 客户端可以设置错误的 ContentLength 或不设置
- 实际读取时没有限制

**风险等级**: 🟡 中等

---

## 📊 大量 JS 加载场景分析

### 场景模拟

```
用户访问一个现代 SPA 应用（如 React/Vue 网站）

典型资源加载:
├── main.html (50KB)
├── main.js (500KB-2MB)
├── vendor.js (1-3MB)
├── chunk-1.js ~ chunk-50.js (每个 50-200KB)
├── main.css (100KB)
├── fonts (200KB)
└── images (500KB-2MB)

总计: 可能有 50-100+ 个请求
```

### 当前代码行为

| 资源类型 | 请求处理 | bodyBuffer | 指纹识别 | Goroutine |
|---------|---------|------------|---------|-----------|
| HTML | ✅ 正常 | 1MB cap | ✅ 有意义 | +1 |
| JS | ✅ 正常 | **无限制** | ⚠️ 意义不大 | +1 |
| CSS | ✅ 正常 | **无限制** | ⚠️ 意义不大 | +1 |
| 图片 | ✅ 正常 | **无限制** | ❌ 无意义 | +1 |
| 字体 | ✅ 正常 | **无限制** | ❌ 无意义 | +1 |

### 内存消耗估算

```
单页面访问（100个资源）:
- 每个 JS 请求: 平均 100KB bodyBuffer
- 100 个并发 goroutine
- 指纹识别缓存: 每个约 10KB
- 总计: ~15-20MB 额外内存

10 个并发用户:
- 1000 个并发 goroutine
- ~150-200MB 额外内存

100 个并发用户:
- 10000 个并发 goroutine ⚠️
- ~1.5-2GB 额外内存 ⚠️
```

---

## ✅ 代码已有的保护措施

### 1. 内存池 (BufferPool)
```go
// 支持多种大小: 1KB ~ 4MB
var defaultSizes = []int{1024, 4096, 8192, 16384, 32768, 65536, ...}
```
✅ 有效减少内存分配

### 2. 定时 GC
```go
bp.gcTicker = time.NewTicker(5 * time.Minute)
```
✅ 每 5 分钟执行一次 GC

### 3. 连接池
```go
MaxIdleConns:        100,
MaxIdleConnsPerHost: 20,
IdleConnTimeout:     90 * time.Second,
```
✅ 连接复用，减少开销

### 4. 请求超时
```go
UpstreamTimeout: 30 * time.Second
```
✅ 防止慢请求阻塞

---

## ✅ 已应用的修复方案

### 修复 1: 添加实际的 bodyBuffer 大小限制 ✅

```go
const maxBodyBufferSize = 1 << 20 // 1MB

type limitedBodyWriter struct {
    buffer *[]byte
    max    int
}

func (bw *limitedBodyWriter) Write(p []byte) (n int, err error) {
    if len(*bw.buffer)+len(p) > bw.max {
        return 0, fmt.Errorf("body exceeds maximum size")
    }
    *bw.buffer = append(*bw.buffer, p...)
    return len(p), nil
}
```

### 修复 2: 使用 io.LimitReader 限制请求体

```go
const maxRequestBodySize = 10 << 20 // 10MB

if r.Body != nil {
    limitedReader := io.LimitReader(r.Body, maxRequestBodySize)
    if data, err := io.ReadAll(limitedReader); err == nil {
        body = data
        r.Body = io.NopCloser(bytes.NewReader(data))
    }
}
```

### 修复 3: 跳过静态资源的指纹识别

```go
// 检查是否为静态资源
func isStaticResource(req *http.Request) bool {
    path := req.URL.Path
    staticExts := []string{".js", ".css", ".png", ".jpg", ".gif", ".ico", ".woff", ".woff2", ".ttf"}
    for _, ext := range staticExts {
        if strings.HasSuffix(path, ext) {
            return true
        }
    }
    return false
}

// 在处理响应时
if s.fingerprintHandler != nil && !isStaticResource(r) {
    bodyBuffer = make([]byte, 0, 1024*1024)
    // ... 指纹识别
}
```

### 修复 4: 使用 Goroutine 池控制并发

```go
// 使用 semaphore 限制并发
var fingerprintSemaphore = make(chan struct{}, 10) // 最多10个并发

func (s *Server) handleFingerprint(r *http.Request, resp *http.Response, body []byte) {
    select {
    case fingerprintSemaphore <- struct{}{}:
        defer func() { <-fingerprintSemaphore }()
        s.fingerprintHandler.HandleRequest(r, resp, body)
    default:
        // 跳过，太忙了
        logger.Debug("指纹识别队列已满，跳过")
    }
}
```

---

## 📋 总结

### 当前稳定性评估

| 指标 | 评分 | 说明 |
|------|------|------|
| **内存安全** | ⭐⭐⭐☆☆ | 存在无限制读取问题 |
| **并发安全** | ⭐⭐⭐⭐☆ | 基本安全，但 goroutine 可能过多 |
| **大文件处理** | ⭐⭐☆☆☆ | 缺少大小限制 |
| **静态资源** | ⭐⭐☆☆☆ | 无优化，浪费资源 |
| **整体稳定性** | ⭐⭐⭐☆☆ | 适合轻量使用 |

### 使用建议

1. **短期**:
   - 禁用指纹识别处理静态资源
   - 监控内存使用

2. **中期**:
   - 应用上述修复方案
   - 添加请求/响应体大小限制

3. **长期**:
   - 添加单元测试和压力测试
   - 实现更完善的资源管理

### 风险场景

| 场景 | 风险 | 可能后果 |
|------|------|---------|
| 普通网页浏览 | 🟢 低 | 正常使用 |
| 大型 SPA 应用 | 🟡 中 | 内存增长 |
| 10+ 并发用户 | 🟡 中 | 性能下降 |
| 100+ 并发用户 | 🔴 高 | 可能 OOM |
| 恶意大文件请求 | 🔴 严重 | 立即 OOM |

---

**结论**: HackMITM 当前版本 **基本可用**，但在高并发和加载大量资源的场景下可能存在稳定性问题。建议应用上述修复方案后再用于生产环境。
