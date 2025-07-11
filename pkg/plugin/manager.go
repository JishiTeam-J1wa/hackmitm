// Package plugin 插件管理器实现
package plugin

import (
	"context"
	"fmt"
	"net/http"
	"path/filepath"
	"plugin"
	"sort"
	"strings"
	"sync"
	"time"

	"hackmitm/pkg/logger"
)

// Manager 插件管理器
type Manager struct {
	// plugins 已加载的插件
	plugins map[string]*PluginWrapper
	// pluginsByType 按类型分组的插件
	pluginsByType map[PluginType][]*PluginWrapper
	// config 插件配置
	config map[string]*PluginConfig
	// mutex 保护并发访问
	mutex sync.RWMutex
	// ctx 上下文
	ctx context.Context
	// cancel 取消函数
	cancel context.CancelFunc
	// basePath 插件基础路径
	basePath string
	// watchers 文件监视器
	watchers map[string]*FileWatcher
}

// PluginWrapper 插件包装器
type PluginWrapper struct {
	Plugin     Plugin
	Info       *PluginInfo
	Config     *PluginConfig
	Status     PluginStatus
	Error      error
	LoadTime   time.Time
	StartTime  time.Time
	StopTime   time.Time
	CallCount  int64
	ErrorCount int64
	mutex      sync.RWMutex
}

// FileWatcher 文件监视器
type FileWatcher struct {
	path     string
	lastMod  time.Time
	ticker   *time.Ticker
	onChange func(string)
	stop     chan bool
}

// NewManager 创建插件管理器
func NewManager(basePath string) *Manager {
	ctx, cancel := context.WithCancel(context.Background())

	return &Manager{
		plugins:       make(map[string]*PluginWrapper),
		pluginsByType: make(map[PluginType][]*PluginWrapper),
		config:        make(map[string]*PluginConfig),
		ctx:           ctx,
		cancel:        cancel,
		basePath:      basePath,
		watchers:      make(map[string]*FileWatcher),
	}
}

// LoadPlugin 加载插件
func (m *Manager) LoadPlugin(config *PluginConfig) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	if !config.Enabled {
		logger.Debugf("插件 %s 已禁用，跳过加载", config.Name)
		return nil
	}

	// 检查插件是否已加载
	if wrapper, exists := m.plugins[config.Name]; exists {
		if wrapper.Status != StatusError {
			return fmt.Errorf("插件 %s 已经加载", config.Name)
		}
		// 如果之前加载失败，先清理
		m.unloadPlugin(config.Name)
	}

	logger.Infof("开始加载插件: %s", config.Name)

	// 解析插件路径
	pluginPath := config.Path
	if !filepath.IsAbs(pluginPath) {
		pluginPath = filepath.Join(m.basePath, pluginPath)
	}

	// 加载插件文件
	p, err := plugin.Open(pluginPath)
	if err != nil {
		return fmt.Errorf("加载插件文件失败: %w", err)
	}

	// 查找插件工厂函数
	factorySymbol, err := p.Lookup("NewPlugin")
	if err != nil {
		// 尝试查找加载器函数
		loaderSymbol, err := p.Lookup("LoadPlugin")
		if err != nil {
			return fmt.Errorf("未找到插件入口函数 NewPlugin 或 LoadPlugin: %w", err)
		}

		// 使用加载器函数
		loader, ok := loaderSymbol.(func() Plugin)
		if !ok {
			return fmt.Errorf("LoadPlugin 函数签名不正确")
		}

		pluginInstance := loader()
		return m.initializePlugin(config, pluginInstance, pluginPath)
	}

	// 使用工厂函数
	factory, ok := factorySymbol.(func(map[string]interface{}) (Plugin, error))
	if !ok {
		return fmt.Errorf("NewPlugin 函数签名不正确")
	}

	pluginInstance, err := factory(config.Config)
	if err != nil {
		return fmt.Errorf("创建插件实例失败: %w", err)
	}

	return m.initializePlugin(config, pluginInstance, pluginPath)
}

// initializePlugin 初始化插件
func (m *Manager) initializePlugin(config *PluginConfig, pluginInstance Plugin, pluginPath string) error {
	// 创建插件包装器
	wrapper := &PluginWrapper{
		Plugin:   pluginInstance,
		Config:   config,
		Status:   StatusLoaded,
		LoadTime: time.Now(),
		Info: &PluginInfo{
			Name:        pluginInstance.Name(),
			Version:     pluginInstance.Version(),
			Description: pluginInstance.Description(),
			Path:        pluginPath,
			LoadTime:    time.Now(),
		},
	}

	// 初始化插件
	if err := pluginInstance.Initialize(config.Config); err != nil {
		wrapper.Status = StatusError
		wrapper.Error = err
		m.plugins[config.Name] = wrapper
		return fmt.Errorf("初始化插件失败: %w", err)
	}

	wrapper.Status = StatusInitialized
	m.plugins[config.Name] = wrapper

	// 按类型分组
	m.addToTypeGroup(wrapper)

	// 存储配置
	m.config[config.Name] = config

	// 启动插件
	if err := m.StartPlugin(config.Name); err != nil {
		logger.Errorf("启动插件失败: %v", err)
		return err
	}

	logger.Infof("插件加载并启动成功: %s v%s", wrapper.Info.Name, wrapper.Info.Version)
	return nil
}

// addToTypeGroup 将插件添加到类型分组中
func (m *Manager) addToTypeGroup(wrapper *PluginWrapper) {
	// 检测插件类型并添加到对应分组
	if _, ok := wrapper.Plugin.(RequestPlugin); ok {
		m.pluginsByType[TypeRequest] = append(m.pluginsByType[TypeRequest], wrapper)
	}
	if _, ok := wrapper.Plugin.(ResponsePlugin); ok {
		m.pluginsByType[TypeResponse] = append(m.pluginsByType[TypeResponse], wrapper)
	}
	if _, ok := wrapper.Plugin.(FilterPlugin); ok {
		m.pluginsByType[TypeFilter] = append(m.pluginsByType[TypeFilter], wrapper)
	}
	if _, ok := wrapper.Plugin.(LoggerPlugin); ok {
		m.pluginsByType[TypeLogger] = append(m.pluginsByType[TypeLogger], wrapper)
	}
	if _, ok := wrapper.Plugin.(ModifierPlugin); ok {
		m.pluginsByType[TypeModifier] = append(m.pluginsByType[TypeModifier], wrapper)
	}
	if _, ok := wrapper.Plugin.(AnalyticsPlugin); ok {
		m.pluginsByType[TypeAnalytics] = append(m.pluginsByType[TypeAnalytics], wrapper)
	}

	// 对每个类型的插件按优先级排序
	m.sortPluginsByPriority()
}

// sortPluginsByPriority 按优先级排序插件
func (m *Manager) sortPluginsByPriority() {
	for pluginType := range m.pluginsByType {
		plugins := m.pluginsByType[pluginType]
		sort.Slice(plugins, func(i, j int) bool {
			var priI, priJ int

			switch pluginType {
			case TypeRequest:
				if rp, ok := plugins[i].Plugin.(RequestPlugin); ok {
					priI = rp.Priority()
				}
				if rp, ok := plugins[j].Plugin.(RequestPlugin); ok {
					priJ = rp.Priority()
				}
			case TypeResponse:
				if rp, ok := plugins[i].Plugin.(ResponsePlugin); ok {
					priI = rp.Priority()
				}
				if rp, ok := plugins[j].Plugin.(ResponsePlugin); ok {
					priJ = rp.Priority()
				}
			case TypeFilter:
				if fp, ok := plugins[i].Plugin.(FilterPlugin); ok {
					priI = fp.Priority()
				}
				if fp, ok := plugins[j].Plugin.(FilterPlugin); ok {
					priJ = fp.Priority()
				}
			case TypeModifier:
				if mp, ok := plugins[i].Plugin.(ModifierPlugin); ok {
					priI = mp.Priority()
				}
				if mp, ok := plugins[j].Plugin.(ModifierPlugin); ok {
					priJ = mp.Priority()
				}
			}

			return priI < priJ // 数字越小优先级越高
		})
	}
}

// StartPlugin 启动插件
func (m *Manager) StartPlugin(name string) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	wrapper, exists := m.plugins[name]
	if !exists {
		return fmt.Errorf("插件 %s 未找到", name)
	}

	if wrapper.Status != StatusInitialized && wrapper.Status != StatusStopped {
		return fmt.Errorf("插件 %s 状态不允许启动: %s", name, wrapper.Status)
	}

	if err := wrapper.Plugin.Start(m.ctx); err != nil {
		wrapper.Status = StatusError
		wrapper.Error = err
		return fmt.Errorf("启动插件失败: %w", err)
	}

	wrapper.Status = StatusStarted
	wrapper.StartTime = time.Now()
	wrapper.Error = nil

	logger.Infof("插件启动成功: %s", name)
	return nil
}

// StopPlugin 停止插件
func (m *Manager) StopPlugin(name string) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	wrapper, exists := m.plugins[name]
	if !exists {
		return fmt.Errorf("插件 %s 未找到", name)
	}

	if wrapper.Status != StatusStarted {
		return fmt.Errorf("插件 %s 状态不允许停止: %s", name, wrapper.Status)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := wrapper.Plugin.Stop(ctx); err != nil {
		wrapper.Status = StatusError
		wrapper.Error = err
		return fmt.Errorf("停止插件失败: %w", err)
	}

	wrapper.Status = StatusStopped
	wrapper.StopTime = time.Now()

	logger.Infof("插件停止成功: %s", name)
	return nil
}

// UnloadPlugin 卸载插件
func (m *Manager) UnloadPlugin(name string) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	return m.unloadPlugin(name)
}

// unloadPlugin 内部卸载插件方法（不加锁）
func (m *Manager) unloadPlugin(name string) error {
	wrapper, exists := m.plugins[name]
	if !exists {
		return fmt.Errorf("插件 %s 未找到", name)
	}

	// 先停止插件
	if wrapper.Status == StatusStarted {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		wrapper.Plugin.Stop(ctx)
		cancel()
	}

	// 清理资源
	if err := wrapper.Plugin.Cleanup(); err != nil {
		logger.Errorf("清理插件资源失败: %v", err)
	}

	// 从类型分组中移除
	m.removeFromTypeGroup(wrapper)

	// 从插件列表中移除
	delete(m.plugins, name)
	delete(m.config, name)

	wrapper.Status = StatusUnloaded

	logger.Infof("插件卸载成功: %s", name)
	return nil
}

// removeFromTypeGroup 从类型分组中移除插件
func (m *Manager) removeFromTypeGroup(wrapper *PluginWrapper) {
	for pluginType, plugins := range m.pluginsByType {
		for i, p := range plugins {
			if p == wrapper {
				// 从切片中移除
				m.pluginsByType[pluginType] = append(plugins[:i], plugins[i+1:]...)
				break
			}
		}
	}
}

// ReloadPlugin 重新加载插件
func (m *Manager) ReloadPlugin(name string) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	config, exists := m.config[name]
	if !exists {
		return fmt.Errorf("插件配置 %s 未找到", name)
	}

	// 先卸载
	if err := m.unloadPlugin(name); err != nil {
		logger.Errorf("卸载插件失败: %v", err)
	}

	// 重新加载
	return m.LoadPlugin(config)
}

// ProcessRequest 处理请求插件链
func (m *Manager) ProcessRequest(req *http.Request, ctx *RequestContext) error {
	m.mutex.RLock()
	plugins := make([]*PluginWrapper, len(m.pluginsByType[TypeRequest]))
	copy(plugins, m.pluginsByType[TypeRequest])
	m.mutex.RUnlock()

	for _, wrapper := range plugins {
		if wrapper.Status != StatusStarted {
			continue
		}

		if requestPlugin, ok := wrapper.Plugin.(RequestPlugin); ok {
			wrapper.mutex.Lock()
			wrapper.CallCount++
			wrapper.mutex.Unlock()

			if err := requestPlugin.ProcessRequest(req, ctx); err != nil {
				wrapper.mutex.Lock()
				wrapper.ErrorCount++
				wrapper.mutex.Unlock()

				logger.Errorf("请求插件 %s 处理失败: %v", wrapper.Info.Name, err)
				return err
			}
		}
	}

	return nil
}

// ProcessResponse 处理响应插件链
func (m *Manager) ProcessResponse(resp *http.Response, req *http.Request, ctx *ResponseContext) error {
	m.mutex.RLock()
	plugins := make([]*PluginWrapper, len(m.pluginsByType[TypeResponse]))
	copy(plugins, m.pluginsByType[TypeResponse])
	m.mutex.RUnlock()

	for _, wrapper := range plugins {
		if wrapper.Status != StatusStarted {
			continue
		}

		if responsePlugin, ok := wrapper.Plugin.(ResponsePlugin); ok {
			wrapper.mutex.Lock()
			wrapper.CallCount++
			wrapper.mutex.Unlock()

			if err := responsePlugin.ProcessResponse(resp, req, ctx); err != nil {
				wrapper.mutex.Lock()
				wrapper.ErrorCount++
				wrapper.mutex.Unlock()

				logger.Errorf("响应插件 %s 处理失败: %v", wrapper.Info.Name, err)
				return err
			}
		}
	}

	return nil
}

// ShouldAllow 执行过滤插件链
func (m *Manager) ShouldAllow(req *http.Request, ctx *FilterContext) (bool, error) {
	m.mutex.RLock()
	plugins := make([]*PluginWrapper, len(m.pluginsByType[TypeFilter]))
	copy(plugins, m.pluginsByType[TypeFilter])
	m.mutex.RUnlock()

	for _, wrapper := range plugins {
		if wrapper.Status != StatusStarted {
			continue
		}

		if filterPlugin, ok := wrapper.Plugin.(FilterPlugin); ok {
			wrapper.mutex.Lock()
			wrapper.CallCount++
			wrapper.mutex.Unlock()

			allowed, err := filterPlugin.ShouldAllow(req, ctx)
			if err != nil {
				wrapper.mutex.Lock()
				wrapper.ErrorCount++
				wrapper.mutex.Unlock()

				logger.Errorf("过滤插件 %s 处理失败: %v", wrapper.Info.Name, err)
				return false, err
			}

			if !allowed {
				return false, nil
			}
		}
	}

	return true, nil
}

// GetPluginInfo 获取插件信息
func (m *Manager) GetPluginInfo(name string) (*PluginInfo, error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	wrapper, exists := m.plugins[name]
	if !exists {
		return nil, fmt.Errorf("插件 %s 未找到", name)
	}

	return wrapper.Info, nil
}

// ListPlugins 列出所有插件
func (m *Manager) ListPlugins() []*PluginInfo {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	var plugins []*PluginInfo
	for _, wrapper := range m.plugins {
		info := *wrapper.Info
		info.Status = string(wrapper.Status)
		plugins = append(plugins, &info)
	}

	return plugins
}

// GetStats 获取插件统计信息
func (m *Manager) GetStats() map[string]interface{} {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	stats := make(map[string]interface{})
	pluginStats := make(map[string]interface{})

	for name, wrapper := range m.plugins {
		wrapper.mutex.RLock()
		pluginStats[name] = map[string]interface{}{
			"status":      string(wrapper.Status),
			"call_count":  wrapper.CallCount,
			"error_count": wrapper.ErrorCount,
			"load_time":   wrapper.LoadTime.Format(time.RFC3339),
			"start_time":  wrapper.StartTime.Format(time.RFC3339),
		}
		wrapper.mutex.RUnlock()
	}

	stats["plugins"] = pluginStats
	stats["total_plugins"] = len(m.plugins)

	// 按类型统计
	typeStats := make(map[string]int)
	for pluginType, plugins := range m.pluginsByType {
		typeStats[string(pluginType)] = len(plugins)
	}
	stats["plugins_by_type"] = typeStats

	return stats
}

// StartAll 启动所有插件
func (m *Manager) StartAll() error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	logger.Info("正在启动所有插件...")
	var errors []string
	for name, wrapper := range m.plugins {
		if wrapper.Status != StatusInitialized && wrapper.Status != StatusStopped {
			logger.Debugf("跳过插件 %s，状态不允许启动: %s", name, wrapper.Status)
			continue
		}

		if err := wrapper.Plugin.Start(m.ctx); err != nil {
			errors = append(errors, fmt.Sprintf("%s: %v", name, err))
			wrapper.Status = StatusError
			wrapper.Error = err
		} else {
			wrapper.Status = StatusStarted
			wrapper.StartTime = time.Now()
			wrapper.Error = nil
			logger.Infof("插件 %s 启动成功", name)
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("启动插件失败: %s", strings.Join(errors, "; "))
	}

	logger.Info("所有插件启动完成")
	return nil
}

// StopAll 停止所有插件
func (m *Manager) StopAll() error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	var errors []string
	for name := range m.plugins {
		if err := m.StopPlugin(name); err != nil {
			errors = append(errors, fmt.Sprintf("%s: %v", name, err))
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("停止插件失败: %s", strings.Join(errors, "; "))
	}

	return nil
}

// Shutdown 关闭插件管理器
func (m *Manager) Shutdown() error {
	m.cancel()

	// 停止所有插件
	if err := m.StopAll(); err != nil {
		logger.Errorf("停止插件失败: %v", err)
	}

	// 卸载所有插件
	m.mutex.Lock()
	defer m.mutex.Unlock()

	for name := range m.plugins {
		if err := m.unloadPlugin(name); err != nil {
			logger.Errorf("卸载插件失败: %v", err)
		}
	}

	// 停止文件监视器
	for _, watcher := range m.watchers {
		close(watcher.stop)
	}

	logger.Info("插件管理器已关闭")
	return nil
}
