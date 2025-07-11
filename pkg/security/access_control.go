// Package security 提供安全控制功能
// Package security provides security control functionality
package security

import (
	"crypto/sha256"
	"crypto/subtle"
	"fmt"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	"hackmitm/pkg/logger"
)

// AccessController 访问控制器
type AccessController struct {
	// whitelist IP白名单
	whitelist map[string]bool
	// blacklist IP黑名单
	blacklist map[string]bool
	// auth 认证配置
	auth *AuthConfig
	// rateLimiter 限流器
	rateLimiter *RateLimiter
	// mutex 保护并发访问
	mutex sync.RWMutex
}

// AuthConfig 认证配置
type AuthConfig struct {
	// Username 用户名
	Username string
	// PasswordHash 密码哈希
	PasswordHash []byte
	// Enabled 是否启用认证
	Enabled bool
}

// RateLimiter 限流器
type RateLimiter struct {
	// clients 客户端访问记录
	clients map[string]*ClientRecord
	// maxRequests 最大请求数
	maxRequests int
	// window 时间窗口
	window time.Duration
	// mutex 保护并发访问
	mutex sync.RWMutex
	// cleanup 清理定时器
	cleanup *time.Ticker
}

// ClientRecord 客户端记录
type ClientRecord struct {
	// requests 请求列表
	requests []time.Time
	// lastAccess 最后访问时间
	lastAccess time.Time
}

// NewAccessController 创建访问控制器
func NewAccessController() *AccessController {
	ac := &AccessController{
		whitelist: make(map[string]bool),
		blacklist: make(map[string]bool),
		rateLimiter: &RateLimiter{
			clients:     make(map[string]*ClientRecord),
			maxRequests: 100, // 默认每分钟100请求
			window:      time.Minute,
		},
	}

	// 启动限流器清理
	ac.rateLimiter.startCleanup()
	return ac
}

// SetAuth 设置认证
func (ac *AccessController) SetAuth(username, password string) {
	ac.mutex.Lock()
	defer ac.mutex.Unlock()

	hash := sha256.Sum256([]byte(password))
	ac.auth = &AuthConfig{
		Username:     username,
		PasswordHash: hash[:],
		Enabled:      true,
	}

	logger.Info("访问认证已启用")
}

// AddToWhitelist 添加到白名单
func (ac *AccessController) AddToWhitelist(ip string) {
	ac.mutex.Lock()
	defer ac.mutex.Unlock()

	ac.whitelist[ip] = true
	logger.Infof("IP已添加到白名单: %s", ip)
}

// AddToBlacklist 添加到黑名单
func (ac *AccessController) AddToBlacklist(ip string) {
	ac.mutex.Lock()
	defer ac.mutex.Unlock()

	ac.blacklist[ip] = true
	logger.Infof("IP已添加到黑名单: %s", ip)
}

// IsAllowed 检查是否允许访问
func (ac *AccessController) IsAllowed(r *http.Request) error {
	clientIP := ac.getClientIP(r)

	// 检查黑名单
	ac.mutex.RLock()
	if ac.blacklist[clientIP] {
		ac.mutex.RUnlock()
		return fmt.Errorf("IP在黑名单中: %s", clientIP)
	}

	// 检查白名单（如果有白名单则只允许白名单IP）
	if len(ac.whitelist) > 0 && !ac.whitelist[clientIP] {
		ac.mutex.RUnlock()
		return fmt.Errorf("IP不在白名单中: %s", clientIP)
	}
	ac.mutex.RUnlock()

	// 检查认证
	if err := ac.checkAuth(r); err != nil {
		return fmt.Errorf("认证失败: %w", err)
	}

	// 检查限流
	if err := ac.rateLimiter.checkRate(clientIP); err != nil {
		return fmt.Errorf("限流触发: %w", err)
	}

	return nil
}

// getClientIP 获取客户端IP
func (ac *AccessController) getClientIP(r *http.Request) string {
	// 检查X-Forwarded-For头
	xff := r.Header.Get("X-Forwarded-For")
	if xff != "" {
		ips := strings.Split(xff, ",")
		return strings.TrimSpace(ips[0])
	}

	// 检查X-Real-IP头
	xri := r.Header.Get("X-Real-IP")
	if xri != "" {
		return strings.TrimSpace(xri)
	}

	// 使用RemoteAddr
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}

// checkAuth 检查认证
func (ac *AccessController) checkAuth(r *http.Request) error {
	ac.mutex.RLock()
	auth := ac.auth
	ac.mutex.RUnlock()

	if auth == nil || !auth.Enabled {
		return nil
	}

	// 检查Proxy-Authorization头
	proxyAuth := r.Header.Get("Proxy-Authorization")
	if proxyAuth == "" {
		return fmt.Errorf("缺少代理认证")
	}

	// 解析Basic认证
	if !strings.HasPrefix(proxyAuth, "Basic ") {
		return fmt.Errorf("不支持的认证类型")
	}

	// 这里简化处理，实际应该解码Base64
	// 为演示目的，假设格式为 "Basic username:password"
	credentials := strings.TrimPrefix(proxyAuth, "Basic ")
	parts := strings.Split(credentials, ":")
	if len(parts) != 2 {
		return fmt.Errorf("认证格式错误")
	}

	username, password := parts[0], parts[1]
	if username != auth.Username {
		return fmt.Errorf("用户名错误")
	}

	// 验证密码
	hash := sha256.Sum256([]byte(password))
	if subtle.ConstantTimeCompare(hash[:], auth.PasswordHash) != 1 {
		return fmt.Errorf("密码错误")
	}

	return nil
}

// startCleanup 启动清理任务
func (rl *RateLimiter) startCleanup() {
	rl.cleanup = time.NewTicker(5 * time.Minute)
	go func() {
		for range rl.cleanup.C {
			rl.cleanupOldRecords()
		}
	}()
}

// checkRate 检查限流
func (rl *RateLimiter) checkRate(clientIP string) error {
	rl.mutex.Lock()
	defer rl.mutex.Unlock()

	now := time.Now()
	record, exists := rl.clients[clientIP]
	if !exists {
		record = &ClientRecord{
			requests:   make([]time.Time, 0),
			lastAccess: now,
		}
		rl.clients[clientIP] = record
	}

	// 清理过期请求
	cutoff := now.Add(-rl.window)
	validRequests := make([]time.Time, 0)
	for _, reqTime := range record.requests {
		if reqTime.After(cutoff) {
			validRequests = append(validRequests, reqTime)
		}
	}
	record.requests = validRequests

	// 检查是否超过限制
	if len(record.requests) >= rl.maxRequests {
		return fmt.Errorf("请求频率过高，限制: %d/%s", rl.maxRequests, rl.window)
	}

	// 记录当前请求
	record.requests = append(record.requests, now)
	record.lastAccess = now

	return nil
}

// cleanupOldRecords 清理旧记录
func (rl *RateLimiter) cleanupOldRecords() {
	rl.mutex.Lock()
	defer rl.mutex.Unlock()

	cutoff := time.Now().Add(-rl.window * 2) // 保留2个窗口的数据
	for ip, record := range rl.clients {
		if record.lastAccess.Before(cutoff) {
			delete(rl.clients, ip)
		}
	}
}

// Stop 停止访问控制器
func (ac *AccessController) Stop() {
	if ac.rateLimiter.cleanup != nil {
		ac.rateLimiter.cleanup.Stop()
	}
}

// GetStats 获取统计信息
func (ac *AccessController) GetStats() map[string]interface{} {
	ac.mutex.RLock()
	defer ac.mutex.RUnlock()

	ac.rateLimiter.mutex.RLock()
	defer ac.rateLimiter.mutex.RUnlock()

	return map[string]interface{}{
		"whitelist_size": len(ac.whitelist),
		"blacklist_size": len(ac.blacklist),
		"auth_enabled":   ac.auth != nil && ac.auth.Enabled,
		"active_clients": len(ac.rateLimiter.clients),
		"rate_limit":     fmt.Sprintf("%d/%s", ac.rateLimiter.maxRequests, ac.rateLimiter.window),
	}
}
