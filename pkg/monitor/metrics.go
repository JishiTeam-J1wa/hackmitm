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
)

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

// GetStats 获取统计信息
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

	// 指标端点
	mux.HandleFunc("/metrics", ms.handleMetrics)
	// 健康检查端点
	mux.HandleFunc("/health", ms.handleHealth)
	// 详细状态端点
	mux.HandleFunc("/status", ms.handleStatus)

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
