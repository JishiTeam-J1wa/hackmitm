package main

// Config 请求日志插件配置
type Config struct {
	// EnableDebug 是否启用调试模式
	EnableDebug bool `json:"enable_debug"`
	// LogLevel 日志级别
	LogLevel string `json:"log_level"`
	// LogFormat 日志格式 (simple, json, detailed)
	LogFormat string `json:"log_format"`
	// LogFile 日志文件路径
	LogFile string `json:"log_file"`
	// MaxSize 单个日志文件最大大小（MB）
	MaxSize int `json:"max_size"`
	// MaxBackups 最大备份文件数
	MaxBackups int `json:"max_backups"`
	// MaxAge 日志文件最大保留天数
	MaxAge int `json:"max_age"`
	// Compress 是否压缩旧日志文件
	Compress bool `json:"compress"`
}

// DefaultConfig 返回默认配置
func DefaultConfig() *Config {
	return &Config{
		EnableDebug: false,
		LogLevel:    "info",
		LogFormat:   "detailed",
		LogFile:     "./logs/requests.log",
		MaxSize:     100,
		MaxBackups:  3,
		MaxAge:      7,
		Compress:    true,
	}
}

// Validate 验证配置
func (c *Config) Validate() error {
	// 验证日志级别
	validLevels := map[string]bool{
		"debug": true,
		"info":  true,
		"warn":  true,
		"error": true,
	}
	if !validLevels[c.LogLevel] {
		c.LogLevel = "info"
	}

	// 验证日志格式
	validFormats := map[string]bool{
		"simple":   true,
		"json":     true,
		"detailed": true,
	}
	if !validFormats[c.LogFormat] {
		c.LogFormat = "detailed"
	}

	// 验证文件大小限制
	if c.MaxSize <= 0 {
		c.MaxSize = 100
	}

	// 验证备份数量
	if c.MaxBackups < 0 {
		c.MaxBackups = 3
	}

	// 验证保留天数
	if c.MaxAge < 0 {
		c.MaxAge = 7
	}

	return nil
}
