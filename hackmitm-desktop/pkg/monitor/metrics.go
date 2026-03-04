// Package monitor 提供监控和指标收集功能
// Package monitor provides monitoring and metrics collection functionality
package monitor

import (
	"encoding/json"
	"fmt"
	"net/http"
	"runtime"
	"sync"
	"sync/atomic"
	"time"

	"hackmitm/pkg/logger"
	"hackmitm/pkg/pool"
)

// ProxyStatsProvider 代理服务器统计信息提供者接口
type ProxyStatsProvider interface {
	GetStats() map[string]interface{}
}

// Metrics 指标收集器
type Metrics struct {
	// 基础指标
	startTime     time.Time
	requestCount  int64
	responseCount int64
	errorCount    int64
	bytesIn       int64
	bytesOut      int64

	// 响应时间统计
	responseTimes []time.Duration
	responseTime  time.Duration
	maxResponse   time.Duration
	minResponse   time.Duration

	// 状态码统计
	statusCodes map[int]int64

	// 连接统计
	activeConns int64
	totalConns  int64

	// 内存池引用
	bufferPool *pool.BufferPool

	// 代理服务器统计提供者
	proxyStatsProvider ProxyStatsProvider

	// 锁保护
	mutex sync.RWMutex
}

// HealthChecker 健康检查器
type HealthChecker struct {
	checks map[string]HealthCheck
	mutex  sync.RWMutex
}

// HealthCheck 健康检查接口
type HealthCheck interface {
	Check() error
	Name() string
}

// HealthStatus 健康状态
type HealthStatus struct {
	Status    string            `json:"status"`
	Timestamp time.Time         `json:"timestamp"`
	Checks    map[string]string `json:"checks,omitempty"`
	Uptime    string            `json:"uptime"`
	Version   string            `json:"version,omitempty"`
}

// MonitorServer 监控服务器
type MonitorServer struct {
	metrics       *Metrics
	healthChecker *HealthChecker
	server        *http.Server
	port          int
}

// NewMetrics 创建指标收集器
func NewMetrics() *Metrics {
	return &Metrics{
		startTime:     time.Now(),
		statusCodes:   make(map[int]int64),
		responseTimes: make([]time.Duration, 0, 1000), // 保留最近1000次请求
		minResponse:   time.Hour,                      // 初始值设为很大
	}
}

// NewHealthChecker 创建健康检查器
func NewHealthChecker() *HealthChecker {
	return &HealthChecker{
		checks: make(map[string]HealthCheck),
	}
}

// NewMonitorServer 创建监控服务器
func NewMonitorServer(port int, metrics *Metrics, healthChecker *HealthChecker) *MonitorServer {
	return &MonitorServer{
		metrics:       metrics,
		healthChecker: healthChecker,
		port:          port,
	}
}

// RecordRequest 记录请求
func (m *Metrics) RecordRequest(size int64) {
	atomic.AddInt64(&m.requestCount, 1)
	atomic.AddInt64(&m.totalConns, 1)
	atomic.AddInt64(&m.bytesIn, size)
}

// RecordResponse 记录响应
func (m *Metrics) RecordResponse(statusCode int, size int64, duration time.Duration) {
	atomic.AddInt64(&m.responseCount, 1)
	atomic.AddInt64(&m.bytesOut, size)

	m.mutex.Lock()
	defer m.mutex.Unlock()

	// 更新状态码统计
	m.statusCodes[statusCode]++

	// 更新响应时间统计
	if len(m.responseTimes) >= 1000 {
		// 移除最老的记录
		m.responseTimes = m.responseTimes[1:]
	}
	m.responseTimes = append(m.responseTimes, duration)

	// 更新响应时间统计
	if duration > m.maxResponse {
		m.maxResponse = duration
	}
	if duration < m.minResponse {
		m.minResponse = duration
	}

	// 计算平均响应时间
	var total time.Duration
	for _, rt := range m.responseTimes {
		total += rt
	}
	m.responseTime = total / time.Duration(len(m.responseTimes))
}

// RecordError 记录错误
func (m *Metrics) RecordError() {
	atomic.AddInt64(&m.errorCount, 1)
}

// SetActiveConnections 设置活跃连接数
func (m *Metrics) SetActiveConnections(count int64) {
	atomic.StoreInt64(&m.activeConns, count)
}

// SetBufferPool 设置内存池引用
func (m *Metrics) SetBufferPool(bufferPool *pool.BufferPool) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.bufferPool = bufferPool
}

// SetProxyStatsProvider 设置代理服务器统计提供者
func (m *Metrics) SetProxyStatsProvider(provider ProxyStatsProvider) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.proxyStatsProvider = provider
}

// GetStats 获取统计信息（增强版，包含代理服务器统计）
func (m *Metrics) GetStats() map[string]interface{} {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	uptime := time.Since(m.startTime)
	requestCount := atomic.LoadInt64(&m.requestCount)
	rps := float64(requestCount) / uptime.Seconds()

	// 计算内存使用情况
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	stats := map[string]interface{}{
		"uptime":            uptime.String(),
		"requests":          requestCount,
		"responses":         atomic.LoadInt64(&m.responseCount),
		"errors":            atomic.LoadInt64(&m.errorCount),
		"active_conns":      atomic.LoadInt64(&m.activeConns),
		"total_conns":       atomic.LoadInt64(&m.totalConns),
		"bytes_in":          atomic.LoadInt64(&m.bytesIn),
		"bytes_out":         atomic.LoadInt64(&m.bytesOut),
		"requests_per_sec":  rps,
		"avg_response_time": m.responseTime.String(),
		"max_response_time": m.maxResponse.String(),
		"min_response_time": m.minResponse.String(),
		"status_codes":      m.statusCodes,
		"memory": map[string]interface{}{
			"alloc":       memStats.Alloc,
			"total_alloc": memStats.TotalAlloc,
			"sys":         memStats.Sys,
			"heap_alloc":  memStats.HeapAlloc,
			"heap_sys":    memStats.HeapSys,
			"gc_runs":     memStats.NumGC,
		},
		"goroutines": runtime.NumGoroutine(),
	}

	// 添加内存池统计信息
	if m.bufferPool != nil {
		poolStats := m.bufferPool.GetStats()
		stats["buffer_pool"] = map[string]interface{}{
			"total_allocated":        poolStats.TotalAllocated,
			"total_released":         poolStats.TotalReleased,
			"current_in_use":         poolStats.CurrentInUse,
			"hit_rate":               poolStats.HitRate,
			"total_hits":             poolStats.TotalHits,
			"total_miss":             poolStats.TotalMiss,
			"total_memory_allocated": poolStats.TotalMemoryAllocated,
			"total_memory_released":  poolStats.TotalMemoryReleased,
			"size_distribution":      poolStats.SizeDistribution,
			"last_gc_time":           poolStats.LastGCTime.Format(time.RFC3339),
			"gc_count":               poolStats.GCCount,
		}
	}

	// 添加代理服务器统计信息（包含访问控制器信息）
	if m.proxyStatsProvider != nil {
		proxyStats := m.proxyStatsProvider.GetStats()
		// 合并代理服务器的统计信息
		for key, value := range proxyStats {
			stats[key] = value
		}
	}

	return stats
}

// AddCheck 添加健康检查
func (hc *HealthChecker) AddCheck(check HealthCheck) {
	hc.mutex.Lock()
	defer hc.mutex.Unlock()
	hc.checks[check.Name()] = check
}

// RemoveCheck 移除健康检查
func (hc *HealthChecker) RemoveCheck(name string) {
	hc.mutex.Lock()
	defer hc.mutex.Unlock()
	delete(hc.checks, name)
}

// CheckHealth 执行健康检查
func (hc *HealthChecker) CheckHealth() HealthStatus {
	hc.mutex.RLock()
	defer hc.mutex.RUnlock()

	status := "healthy"
	checks := make(map[string]string)

	for name, check := range hc.checks {
		if err := check.Check(); err != nil {
			status = "unhealthy"
			checks[name] = fmt.Sprintf("FAIL: %v", err)
		} else {
			checks[name] = "OK"
		}
	}

	return HealthStatus{
		Status:    status,
		Timestamp: time.Now(),
		Checks:    checks,
		Uptime:    time.Since(time.Now()).String(), // 这里应该从外部传入startTime
	}
}

// Start 启动监控服务器
func (ms *MonitorServer) Start() error {
	mux := http.NewServeMux()

	// 注册路由
	mux.HandleFunc("/metrics", ms.handleMetrics)
	mux.HandleFunc("/health", ms.handleHealth)
	mux.HandleFunc("/status", ms.handleStatus)
	mux.HandleFunc("/patterns", ms.handlePatterns)
	mux.HandleFunc("/patterns/stats", ms.handlePatternStats)
	mux.HandleFunc("/fingerprint", ms.handleFingerprint)
	mux.HandleFunc("/fingerprint/stats", ms.handleFingerprintStats)
	mux.HandleFunc("/fingerprint/identify", ms.handleFingerprintIdentify)

	ms.server = &http.Server{
		Addr:    fmt.Sprintf(":%d", ms.port),
		Handler: mux,
	}

	logger.Infof("监控服务器启动在端口: %d", ms.port)
	return ms.server.ListenAndServe()
}

// Stop 停止监控服务器
func (ms *MonitorServer) Stop() error {
	if ms.server != nil {
		return ms.server.Close()
	}
	return nil
}

// handleMetrics 处理指标请求
func (ms *MonitorServer) handleMetrics(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	stats := ms.metrics.GetStats()
	json.NewEncoder(w).Encode(stats)
}

// handleHealth 处理健康检查请求
func (ms *MonitorServer) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	health := ms.healthChecker.CheckHealth()

	if health.Status != "healthy" {
		w.WriteHeader(http.StatusServiceUnavailable)
	}

	json.NewEncoder(w).Encode(health)
}

// handleStatus 处理状态请求
func (ms *MonitorServer) handleStatus(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	status := map[string]interface{}{
		"metrics": ms.metrics.GetStats(),
		"health":  ms.healthChecker.CheckHealth(),
	}

	json.NewEncoder(w).Encode(status)
}

// handlePatterns 处理流量模式请求
func (ms *MonitorServer) handlePatterns(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// 从代理服务器获取流量模式信息
	stats := ms.metrics.GetStats()
	if patternStats, exists := stats["pattern_recognition"]; exists {
		json.NewEncoder(w).Encode(patternStats)
	} else {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "流量模式识别功能未启用",
		})
	}
}

// handlePatternStats 处理流量模式识别统计信息请求
func (ms *MonitorServer) handlePatternStats(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if ms.metrics.proxyStatsProvider == nil {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error": "Pattern recognition not available",
		})
		return
	}

	// 获取代理服务器统计信息
	stats := ms.metrics.proxyStatsProvider.GetStats()

	// 提取流量模式识别统计信息
	var patternStats map[string]interface{}
	if patternData, exists := stats["pattern_recognition"]; exists {
		if patternMap, ok := patternData.(map[string]interface{}); ok {
			patternStats = patternMap
		}
	}

	if patternStats == nil {
		patternStats = map[string]interface{}{
			"enabled": false,
		}
	}

	json.NewEncoder(w).Encode(patternStats)
}

// handleFingerprint 处理指纹识别请求
func (ms *MonitorServer) handleFingerprint(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if ms.metrics.proxyStatsProvider == nil {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error": "Fingerprint recognition not available",
		})
		return
	}

	// 获取代理服务器统计信息
	stats := ms.metrics.proxyStatsProvider.GetStats()

	// 提取指纹识别信息
	var fingerprintData map[string]interface{}
	if fingerprint, exists := stats["fingerprint"]; exists {
		if fingerprintMap, ok := fingerprint.(map[string]interface{}); ok {
			fingerprintData = fingerprintMap
		}
	}

	if fingerprintData == nil {
		fingerprintData = map[string]interface{}{
			"enabled": false,
		}
	}

	json.NewEncoder(w).Encode(fingerprintData)
}

// handleFingerprintStats 处理指纹识别统计信息请求
func (ms *MonitorServer) handleFingerprintStats(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if ms.metrics.proxyStatsProvider == nil {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error": "Fingerprint recognition not available",
		})
		return
	}

	// 获取代理服务器统计信息
	stats := ms.metrics.proxyStatsProvider.GetStats()

	// 提取指纹识别统计信息
	var fingerprintStats map[string]interface{}
	if fingerprintData, exists := stats["fingerprint_stats"]; exists {
		if fingerprintMap, ok := fingerprintData.(map[string]interface{}); ok {
			fingerprintStats = fingerprintMap
		}
	}

	if fingerprintStats == nil {
		fingerprintStats = map[string]interface{}{
			"enabled": false,
		}
	}

	json.NewEncoder(w).Encode(fingerprintStats)
}

// handleFingerprintIdentify 处理指纹识别URL请求
func (ms *MonitorServer) handleFingerprintIdentify(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// 解析请求参数
	var request struct {
		URL string `json:"url"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if request.URL == "" {
		http.Error(w, "URL is required", http.StatusBadRequest)
		return
	}

	// 尝试从代理服务器获取指纹识别处理器
	if ms.metrics.proxyStatsProvider != nil {
		// 获取代理服务器统计信息，检查是否有指纹识别功能
		stats := ms.metrics.proxyStatsProvider.GetStats()
		if fingerprintData, exists := stats["fingerprint"]; exists {
			if fingerprintMap, ok := fingerprintData.(map[string]interface{}); ok {
				if status, exists := fingerprintMap["status"]; exists && status == "active" {
					// 指纹识别功能可用，但由于架构限制，我们无法直接调用
					// 建议用户通过代理方式进行指纹识别
					response := map[string]interface{}{
						"url":         request.URL,
						"fingerprint": []string{},
						"confidence":  0.0,
						"message":     "请通过代理服务器访问URL以进行指纹识别",
						"proxy_url":   "http://localhost:8082",
						"suggestion":  fmt.Sprintf("curl -x http://localhost:8082 %s", request.URL),
					}
					json.NewEncoder(w).Encode(response)
					return
				}
			}
		}
	}

	// 指纹识别功能不可用
	response := map[string]interface{}{
		"url":         request.URL,
		"fingerprint": []string{},
		"confidence":  0.0,
		"message":     "指纹识别功能不可用",
		"error":       "Fingerprint recognition service not available",
	}

	json.NewEncoder(w).Encode(response)
}

// 预定义的健康检查

// MemoryCheck 内存检查
type MemoryCheck struct {
	maxMemoryMB int
}

// NewMemoryCheck 创建内存检查
func NewMemoryCheck(maxMemoryMB int) *MemoryCheck {
	return &MemoryCheck{maxMemoryMB: maxMemoryMB}
}

func (mc *MemoryCheck) Name() string {
	return "memory"
}

func (mc *MemoryCheck) Check() error {
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	currentMB := int(memStats.Alloc / 1024 / 1024)
	if currentMB > mc.maxMemoryMB {
		return fmt.Errorf("内存使用过高: %dMB > %dMB", currentMB, mc.maxMemoryMB)
	}

	return nil
}

// GoroutineCheck Goroutine检查
type GoroutineCheck struct {
	maxGoroutines int
}

// NewGoroutineCheck 创建Goroutine检查
func NewGoroutineCheck(maxGoroutines int) *GoroutineCheck {
	return &GoroutineCheck{maxGoroutines: maxGoroutines}
}

func (gc *GoroutineCheck) Name() string {
	return "goroutines"
}

func (gc *GoroutineCheck) Check() error {
	current := runtime.NumGoroutine()
	if current > gc.maxGoroutines {
		return fmt.Errorf("Goroutine数量过多: %d > %d", current, gc.maxGoroutines)
	}

	return nil
}

// BufferPoolCheck 内存池健康检查
type BufferPoolCheck struct {
	bufferPool     *pool.BufferPool
	maxHitRate     float64
	maxMemoryUsage int64
}

// NewBufferPoolCheck 创建内存池健康检查
func NewBufferPoolCheck(bufferPool *pool.BufferPool, maxHitRate float64, maxMemoryUsage int64) *BufferPoolCheck {
	return &BufferPoolCheck{
		bufferPool:     bufferPool,
		maxHitRate:     maxHitRate,
		maxMemoryUsage: maxMemoryUsage,
	}
}

func (bpc *BufferPoolCheck) Name() string {
	return "buffer_pool"
}

func (bpc *BufferPoolCheck) Check() error {
	if bpc.bufferPool == nil {
		return fmt.Errorf("内存池未初始化")
	}

	stats := bpc.bufferPool.GetStats()

	// 检查命中率
	if stats.HitRate < bpc.maxHitRate {
		return fmt.Errorf("内存池命中率过低: %.2f%% < %.2f%%", stats.HitRate*100, bpc.maxHitRate*100)
	}

	// 检查内存使用
	memoryInUse := stats.TotalMemoryAllocated - stats.TotalMemoryReleased
	if memoryInUse > bpc.maxMemoryUsage {
		return fmt.Errorf("内存池使用过多: %d bytes > %d bytes", memoryInUse, bpc.maxMemoryUsage)
	}

	return nil
}
