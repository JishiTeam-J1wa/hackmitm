package rules

import (
	"context"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
)

// HotReloader 规则热重载器
// HotReloader provides hot reload capability for rules
type HotReloader struct {
	loader     *RuleLoader
	watcher    *fsnotify.Watcher
	ruleDir    string
	mu         sync.RWMutex
	running    bool
	ctx        context.Context
	cancel     context.CancelFunc
	wg         sync.WaitGroup
	onReload   func()
	debounce   time.Duration
	lastReload time.Time
}

// HotReloaderConfig 热重载器配置
type HotReloaderConfig struct {
	RuleDir   string
	Debounce  time.Duration // 防抖时间
	OnReload  func()        // 重载回调
}

// NewHotReloader 创建热重载器
func NewHotReloader(loader *RuleLoader, config HotReloaderConfig) (*HotReloader, error) {
	if config.Debounce == 0 {
		config.Debounce = 500 * time.Millisecond
	}

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithCancel(context.Background())

	return &HotReloader{
		loader:   loader,
		watcher:  watcher,
		ruleDir:  config.RuleDir,
		onReload: config.OnReload,
		debounce: config.Debounce,
		ctx:      ctx,
		cancel:   cancel,
	}, nil
}

// Start 启动热重载监听
func (h *HotReloader) Start() error {
	h.mu.Lock()
	if h.running {
		h.mu.Unlock()
		return nil
	}
	h.running = true
	h.mu.Unlock()

	// 添加规则目录监听
	if err := h.watcher.Add(h.ruleDir); err != nil {
		return err
	}

	// 监听子目录
	filepath.Walk(h.ruleDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if info.IsDir() {
			h.watcher.Add(path)
		}
		return nil
	})

	h.wg.Add(1)
	go h.watchLoop()

	log.Printf("[HotReload] Started watching: %s", h.ruleDir)
	return nil
}

// Stop 停止热重载监听
func (h *HotReloader) Stop() {
	h.mu.Lock()
	if !h.running {
		h.mu.Unlock()
		return
	}
	h.running = false
	h.mu.Unlock()

	h.cancel()
	h.wg.Wait()
	h.watcher.Close()

	log.Printf("[HotReload] Stopped")
}

// watchLoop 监听循环
func (h *HotReloader) watchLoop() {
	defer h.wg.Done()

	debounceTimer := time.NewTimer(0)
	<-debounceTimer.C // drain initial tick
	pendingReload := false

	for {
		select {
		case <-h.ctx.Done():
			return

		case event, ok := <-h.watcher.Events:
			if !ok {
				return
			}

			// 只处理 yaml 文件
			if !isYAMLFile(event.Name) {
				continue
			}

			// 处理事件
			if event.Op&fsnotify.Write == fsnotify.Write ||
				event.Op&fsnotify.Create == fsnotify.Create ||
				event.Op&fsnotify.Remove == fsnotify.Remove {
				log.Printf("[HotReload] File changed: %s (%v)", event.Name, event.Op)
				pendingReload = true
				debounceTimer.Reset(h.debounce)
			}

		case <-debounceTimer.C:
			if pendingReload {
				h.reload()
				pendingReload = false
			}

		case err, ok := <-h.watcher.Errors:
			if !ok {
				return
			}
			log.Printf("[HotReload] Watcher error: %v", err)
		}
	}
}

// reload 执行重载
func (h *HotReloader) reload() {
	// 防止频繁重载
	if time.Since(h.lastReload) < h.debounce {
		return
	}
	h.lastReload = time.Now()

	log.Printf("[HotReload] Reloading rules...")

	// 重载规则
	if err := h.loader.Load(); err != nil {
		log.Printf("[HotReload] Failed to reload: %v", err)
		return
	}

	log.Printf("[HotReload] Reloaded %d rules", h.loader.Count())

	// 执行回调
	if h.onReload != nil {
		h.onReload()
	}
}

// ForceReload 强制重载
func (h *HotReloader) ForceReload() error {
	return h.loader.Load()
}

// isYAMLFile 检查是否为 YAML 文件
func isYAMLFile(path string) bool {
	ext := filepath.Ext(path)
	return ext == ".yaml" || ext == ".yml"
}

// RuleWatcher 规则监听器（简化版，不依赖 fsnotify）
type RuleWatcher struct {
	loader    *RuleLoader
	ruleDir   string
	interval  time.Duration
	running   bool
	stopChan  chan struct{}
	modTimes  map[string]time.Time
	mu        sync.RWMutex
	onChange  func()
}

// NewRuleWatcher 创建规则监听器
func NewRuleWatcher(loader *RuleLoader, ruleDir string, interval time.Duration) *RuleWatcher {
	if interval == 0 {
		interval = 5 * time.Second
	}
	return &RuleWatcher{
		loader:   loader,
		ruleDir:  ruleDir,
		interval: interval,
		stopChan: make(chan struct{}),
		modTimes: make(map[string]time.Time),
	}
}

// OnChange 设置变更回调
func (w *RuleWatcher) OnChange(fn func()) {
	w.onChange = fn
}

// Start 启动监听
func (w *RuleWatcher) Start() {
	w.mu.Lock()
	if w.running {
		w.mu.Unlock()
		return
	}
	w.running = true
	w.mu.Unlock()

	// 初始化文件修改时间
	w.scanModTimes()

	go w.watchLoop()
}

// Stop 停止监听
func (w *RuleWatcher) Stop() {
	w.mu.Lock()
	defer w.mu.Unlock()
	if !w.running {
		return
	}
	w.running = false
	close(w.stopChan)
}

// watchLoop 监听循环
func (w *RuleWatcher) watchLoop() {
	ticker := time.NewTicker(w.interval)
	defer ticker.Stop()

	for {
		select {
		case <-w.stopChan:
			return
		case <-ticker.C:
			if w.hasChanges() {
				log.Printf("[RuleWatcher] Detected changes, reloading...")
				if err := w.loader.Load(); err != nil {
					log.Printf("[RuleWatcher] Reload failed: %v", err)
				} else {
					w.scanModTimes()
					if w.onChange != nil {
						w.onChange()
					}
				}
			}
		}
	}
}

// scanModTimes 扫描文件修改时间
func (w *RuleWatcher) scanModTimes() {
	filepath.Walk(w.ruleDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if !info.IsDir() && isYAMLFile(path) {
			w.modTimes[path] = info.ModTime()
		}
		return nil
	})
}

// hasChanges 检查是否有变更
func (w *RuleWatcher) hasChanges() bool {
	changed := false

	filepath.Walk(w.ruleDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if !info.IsDir() && isYAMLFile(path) {
			oldTime, exists := w.modTimes[path]
			if !exists || info.ModTime().After(oldTime) {
				changed = true
				return filepath.SkipAll
			}
		}
		return nil
	})

	return changed
}
