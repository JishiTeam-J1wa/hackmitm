// Package config 提供配置管理功能
// Package config provides configuration management functionality
package config

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"sync"
	"time"

	"hackmitm/pkg/logger"
)

// Config 主配置结构体
// Config main configuration structure
type Config struct {
	// Server 服务器配置
	Server ServerConfig `json:"server"`
	// TLS TLS配置
	TLS TLSConfig `json:"tls"`
	// Proxy 代理配置
	Proxy ProxyConfig `json:"proxy"`
	// Logging 日志配置
	Logging LoggingConfig `json:"logging"`
	// Performance 性能配置
	Performance PerformanceConfig `json:"performance"`

	// 内部字段
	mu       sync.RWMutex
	filePath string
	lastMod  time.Time
}

// ServerConfig 服务器配置
// ServerConfig server configuration
type ServerConfig struct {
	// ListenPort 监听端口
	ListenPort int `json:"listen_port"`
	// ListenAddr 监听地址
	ListenAddr string `json:"listen_addr"`
	// ReadTimeout 读取超时时间
	ReadTimeout time.Duration `json:"read_timeout"`
	// WriteTimeout 写入超时时间
	WriteTimeout time.Duration `json:"write_timeout"`
}

// TLSConfig TLS配置
// TLSConfig TLS configuration
type TLSConfig struct {
	// CertDir 证书存储目录
	CertDir string `json:"cert_dir"`
	// CAKeyFile CA私钥文件路径
	CAKeyFile string `json:"ca_key_file"`
	// CACertFile CA证书文件路径
	CACertFile string `json:"ca_cert_file"`
	// EnableCertCache 启用证书缓存
	EnableCertCache bool `json:"enable_cert_cache"`
	// CertCacheTTL 证书缓存TTL
	CertCacheTTL time.Duration `json:"cert_cache_ttl"`
}

// ProxyConfig 代理配置
// ProxyConfig proxy configuration
type ProxyConfig struct {
	// EnableHTTP 启用HTTP代理
	EnableHTTP bool `json:"enable_http"`
	// EnableHTTPS 启用HTTPS代理
	EnableHTTPS bool `json:"enable_https"`
	// UpstreamTimeout 上游超时时间
	UpstreamTimeout time.Duration `json:"upstream_timeout"`
	// MaxIdleConns 最大空闲连接数
	MaxIdleConns int `json:"max_idle_conns"`
	// EnableCompression 启用压缩
	EnableCompression bool `json:"enable_compression"`
}

// LoggingConfig 日志配置
// LoggingConfig logging configuration
type LoggingConfig struct {
	// Level 日志级别
	Level string `json:"level"`
	// Output 输出位置 ("stdout", "stderr", 或文件路径)
	Output string `json:"output"`
	// Format 日志格式 ("text", "json")
	Format string `json:"format"`
	// EnableFileRotation 启用文件轮转
	EnableFileRotation bool `json:"enable_file_rotation"`
}

// PerformanceConfig 性能配置
// PerformanceConfig performance configuration
type PerformanceConfig struct {
	// MaxGoroutines 最大goroutine数量
	MaxGoroutines int `json:"max_goroutines"`
	// BufferSize 缓冲区大小
	BufferSize int `json:"buffer_size"`
	// EnablePProf 启用性能分析
	EnablePProf bool `json:"enable_pprof"`
	// PProfPort 性能分析端口
	PProfPort int `json:"pprof_port"`
}

var (
	// DefaultConfig 默认配置
	DefaultConfig *Config
	// configMutex 配置锁
	configMutex sync.RWMutex
)

func init() {
	DefaultConfig = getDefaultConfig()
}

// getDefaultConfig 获取默认配置
// getDefaultConfig returns the default configuration
func getDefaultConfig() *Config {
	return &Config{
		Server: ServerConfig{
			ListenPort:   8080,
			ListenAddr:   "0.0.0.0",
			ReadTimeout:  30 * time.Second,
			WriteTimeout: 30 * time.Second,
		},
		TLS: TLSConfig{
			CertDir:         "./certs",
			CAKeyFile:       "./certs/ca-key.pem",
			CACertFile:      "./certs/ca-cert.pem",
			EnableCertCache: true,
			CertCacheTTL:    24 * time.Hour,
		},
		Proxy: ProxyConfig{
			EnableHTTP:        true,
			EnableHTTPS:       true,
			UpstreamTimeout:   30 * time.Second,
			MaxIdleConns:      100,
			EnableCompression: true,
		},
		Logging: LoggingConfig{
			Level:              "info",
			Output:             "stdout",
			Format:             "text",
			EnableFileRotation: false,
		},
		Performance: PerformanceConfig{
			MaxGoroutines: 10000,
			BufferSize:    4096,
			EnablePProf:   false,
			PProfPort:     6060,
		},
	}
}

// LoadConfig 从文件加载配置
// LoadConfig loads configuration from file
func LoadConfig(filePath string) (*Config, error) {
	config := getDefaultConfig()
	config.filePath = filePath

	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		logger.Warnf("配置文件不存在，使用默认配置: %s", filePath)
		return config, nil
	}

	data, err := ioutil.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("读取配置文件失败: %w", err)
	}

	if err := json.Unmarshal(data, config); err != nil {
		return nil, fmt.Errorf("解析配置文件失败: %w", err)
	}

	// 获取文件修改时间
	if stat, err := os.Stat(filePath); err == nil {
		config.lastMod = stat.ModTime()
	}

	logger.Infof("配置文件加载成功: %s", filePath)
	return config, nil
}

// SaveConfig 保存配置到文件
// SaveConfig saves configuration to file
func (c *Config) SaveConfig() error {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.filePath == "" {
		return fmt.Errorf("配置文件路径为空")
	}

	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return fmt.Errorf("序列化配置失败: %w", err)
	}

	if err := ioutil.WriteFile(c.filePath, data, 0644); err != nil {
		return fmt.Errorf("写入配置文件失败: %w", err)
	}

	logger.Infof("配置文件保存成功: %s", c.filePath)
	return nil
}

// Reload 重新加载配置
// Reload reloads the configuration
func (c *Config) Reload() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.filePath == "" {
		return fmt.Errorf("配置文件路径为空")
	}

	stat, err := os.Stat(c.filePath)
	if err != nil {
		return fmt.Errorf("获取配置文件状态失败: %w", err)
	}

	// 检查文件是否已修改
	if !stat.ModTime().After(c.lastMod) {
		return nil // 文件未修改
	}

	newConfig, err := LoadConfig(c.filePath)
	if err != nil {
		return fmt.Errorf("重新加载配置失败: %w", err)
	}

	// 更新配置
	*c = *newConfig
	logger.Info("配置已重新加载")
	return nil
}

// GetServer 获取服务器配置
// GetServer returns server configuration
func (c *Config) GetServer() ServerConfig {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.Server
}

// GetTLS 获取TLS配置
// GetTLS returns TLS configuration
func (c *Config) GetTLS() TLSConfig {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.TLS
}

// GetProxy 获取代理配置
// GetProxy returns proxy configuration
func (c *Config) GetProxy() ProxyConfig {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.Proxy
}

// GetLogging 获取日志配置
// GetLogging returns logging configuration
func (c *Config) GetLogging() LoggingConfig {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.Logging
}

// GetPerformance 获取性能配置
// GetPerformance returns performance configuration
func (c *Config) GetPerformance() PerformanceConfig {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.Performance
}

// StartConfigWatcher 启动配置文件监控
// StartConfigWatcher starts configuration file watcher
func (c *Config) StartConfigWatcher(interval time.Duration) {
	if c.filePath == "" {
		return
	}

	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for range ticker.C {
			if err := c.Reload(); err != nil {
				logger.Errorf("配置重新加载失败: %v", err)
			}
		}
	}()

	logger.Info("配置文件监控已启动")
}
