// Package plugin 插件基础实现
package plugin

import (
	"context"
	"fmt"
	"time"

	"hackmitm/pkg/logger"
)

// BasePlugin 插件基础实现
type BasePlugin struct {
	name        string
	version     string
	description string
	author      string
	license     string
	config      map[string]interface{}
	started     bool
	startTime   time.Time
}

// NewBasePlugin 创建基础插件
func NewBasePlugin(name, version, description string) *BasePlugin {
	return &BasePlugin{
		name:        name,
		version:     version,
		description: description,
		config:      make(map[string]interface{}),
	}
}

// Name 返回插件名称
func (bp *BasePlugin) Name() string {
	return bp.name
}

// Version 返回插件版本
func (bp *BasePlugin) Version() string {
	return bp.version
}

// Description 返回插件描述
func (bp *BasePlugin) Description() string {
	return bp.description
}

// Initialize 初始化插件
func (bp *BasePlugin) Initialize(config map[string]interface{}) error {
	bp.config = config
	logger.Debugf("插件 %s 初始化完成", bp.name)
	return nil
}

// Start 启动插件
func (bp *BasePlugin) Start(ctx context.Context) error {
	bp.started = true
	bp.startTime = time.Now()
	logger.Debugf("插件 %s 启动完成", bp.name)
	return nil
}

// Stop 停止插件
func (bp *BasePlugin) Stop(ctx context.Context) error {
	bp.started = false
	logger.Debugf("插件 %s 停止完成", bp.name)
	return nil
}

// Cleanup 清理插件资源
func (bp *BasePlugin) Cleanup() error {
	logger.Debugf("插件 %s 清理完成", bp.name)
	return nil
}

// GetConfig 获取配置项
func (bp *BasePlugin) GetConfig(key string) (interface{}, bool) {
	value, exists := bp.config[key]
	return value, exists
}

// GetConfigString 获取字符串配置项
func (bp *BasePlugin) GetConfigString(key, defaultValue string) string {
	if value, exists := bp.config[key]; exists {
		if str, ok := value.(string); ok {
			return str
		}
	}
	return defaultValue
}

// GetConfigInt 获取整数配置项
func (bp *BasePlugin) GetConfigInt(key string, defaultValue int) int {
	if value, exists := bp.config[key]; exists {
		switch v := value.(type) {
		case int:
			return v
		case float64:
			return int(v)
		case string:
			// 尝试解析字符串
			if i, err := fmt.Sscanf(v, "%d", &defaultValue); err == nil && i == 1 {
				return defaultValue
			}
		}
	}
	return defaultValue
}

// GetConfigBool 获取布尔配置项
func (bp *BasePlugin) GetConfigBool(key string, defaultValue bool) bool {
	if value, exists := bp.config[key]; exists {
		if b, ok := value.(bool); ok {
			return b
		}
	}
	return defaultValue
}

// IsStarted 检查插件是否已启动
func (bp *BasePlugin) IsStarted() bool {
	return bp.started
}

// GetStartTime 获取启动时间
func (bp *BasePlugin) GetStartTime() time.Time {
	return bp.startTime
}
