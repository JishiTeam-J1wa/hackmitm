# HackMITM 插件开发指南

HackMITM 提供了强大的动态插件系统，允许开发者扩展代理服务器的功能。本文档详细介绍如何开发、构建和部署插件。

## 🏗️ 插件架构

### 插件类型

HackMITM 支持以下类型的插件：

1. **RequestPlugin** - 处理HTTP请求
2. **ResponsePlugin** - 处理HTTP响应
3. **FilterPlugin** - 过滤和阻止请求
4. **LoggerPlugin** - 自定义日志记录
5. **ModifierPlugin** - 修改请求/响应
6. **AnalyticsPlugin** - 流量分析和统计

### 插件接口

所有插件必须实现基础的 `Plugin` 接口：

```go
type Plugin interface {
    Name() string
    Version() string
    Description() string
    Initialize(config map[string]interface{}) error
    Start(ctx context.Context) error
    Stop(ctx context.Context) error
    Cleanup() error
}
```

## 📝 开发插件

### 1. 创建插件结构

```go
package main

import (
    "context"
    "net/http"
    "hackmitm/pkg/plugin"
    "hackmitm/pkg/logger"
)

// MyPlugin 自定义插件
type MyPlugin struct {
    *plugin.BasePlugin
    // 添加自定义字段
}

// NewPlugin 插件工厂函数（必需）
func NewPlugin(config map[string]interface{}) (plugin.Plugin, error) {
    base := plugin.NewBasePlugin("my-plugin", "1.0.0", "我的自定义插件")
    
    p := &MyPlugin{
        BasePlugin: base,
    }
    
    return p, nil
}
```

### 2. 实现特定插件类型

#### 请求处理插件

```go
// Priority 返回插件优先级
func (p *MyPlugin) Priority() int {
    return p.GetConfigInt("priority", 100)
}

// ProcessRequest 处理HTTP请求
func (p *MyPlugin) ProcessRequest(req *http.Request, ctx *plugin.RequestContext) error {
    logger.Infof("处理请求: %s %s", req.Method, req.URL.String())
    
    // 在这里添加请求处理逻辑
    
    return nil
}
```

#### 过滤插件

```go
// ShouldAllow 判断是否允许请求通过
func (p *MyPlugin) ShouldAllow(req *http.Request, ctx *plugin.FilterContext) (bool, error) {
    // 在这里添加过滤逻辑
    
    if req.URL.Path == "/blocked" {
        return false, nil // 阻止请求
    }
    
    return true, nil // 允许请求
}
```

#### 分析插件

```go
// AnalyzeRequest 分析请求
func (p *MyPlugin) AnalyzeRequest(req *http.Request, ctx *plugin.RequestContext) (*plugin.AnalysisResult, error) {
    result := &plugin.AnalysisResult{
        Threat:      false,
        ThreatLevel: "none",
        Description: "正常请求",
        Confidence:  1.0,
        Timestamp:   time.Now(),
        Metadata:    make(map[string]interface{}),
    }
    
    // 在这里添加分析逻辑
    
    return result, nil
}

// GetStatistics 获取统计信息
func (p *MyPlugin) GetStatistics() map[string]interface{} {
    return map[string]interface{}{
        "processed_requests": p.processedCount,
        "plugin_version":     p.Version(),
    }
}
```

### 3. 配置处理

```go
// Initialize 初始化插件
func (p *MyPlugin) Initialize(config map[string]interface{}) error {
    if err := p.BasePlugin.Initialize(config); err != nil {
        return err
    }
    
    // 读取配置
    p.enabled = p.GetConfigBool("enabled", true)
    p.threshold = p.GetConfigInt("threshold", 100)
    p.apiKey = p.GetConfigString("api_key", "")
    
    logger.Infof("插件 %s 初始化完成", p.Name())
    return nil
}
```

## 🔧 构建插件

### 1. 插件目录结构

```
plugins/
├── examples/
│   ├── my_plugin.go          # 插件源码
│   └── my_plugin.so          # 编译后的插件
├── Makefile                  # 构建脚本
└── build.sh                  # 构建助手脚本
```

### 2. 构建命令

#### 单个插件构建

```bash
go build -buildmode=plugin -o my_plugin.so my_plugin.go
```

#### 使用构建脚本

创建 `build.sh` 脚本：

```bash
#!/bin/bash

PLUGIN_NAME=$1
if [ -z "$PLUGIN_NAME" ]; then
    echo "使用方法: ./build.sh <plugin_name>"
    exit 1
fi

echo "构建插件: $PLUGIN_NAME"
go build -buildmode=plugin -o "examples/${PLUGIN_NAME}.so" "examples/${PLUGIN_NAME}.go"

if [ $? -eq 0 ]; then
    echo "插件构建成功: examples/${PLUGIN_NAME}.so"
else
    echo "插件构建失败"
    exit 1
fi
```

### 3. Makefile

```makefile
.PHONY: all clean request-logger traffic-stats sql-detector

# 构建所有插件
all: request-logger traffic-stats sql-detector

# 单个插件构建
request-logger:
	go build -buildmode=plugin -o examples/request_logger.so examples/request_logger.go

traffic-stats:
	go build -buildmode=plugin -o examples/traffic_stats.so examples/traffic_stats.go

sql-detector:
	go build -buildmode=plugin -o examples/sql_injection_detector.so examples/sql_injection_detector.go

# 清理构建文件
clean:
	rm -f examples/*.so

# 安装插件到指定目录
install: all
	mkdir -p $(DESTDIR)/plugins/examples
	cp examples/*.so $(DESTDIR)/plugins/examples/
```

## ⚙️ 配置插件

### 1. 在配置文件中添加插件

编辑 `configs/config.json`：

```json
{
  "plugins": {
    "enabled": true,
    "base_path": "./plugins",
    "auto_load": true,
    "plugins": [
      {
        "name": "my-plugin",
        "enabled": true,
        "path": "examples/my_plugin.so",
        "priority": 100,
        "config": {
          "api_key": "your-api-key",
          "threshold": 50,
          "enable_debug": false
        }
      }
    ]
  }
}
```

### 2. 配置项说明

- `enabled`: 是否启用插件系统
- `base_path`: 插件基础路径
- `auto_load`: 是否自动加载插件
- `plugins`: 插件列表
  - `name`: 插件名称
  - `enabled`: 是否启用该插件
  - `path`: 插件文件路径（相对于base_path）
  - `priority`: 优先级（数字越小优先级越高）
  - `config`: 插件特定配置

## 🚀 部署和测试

### 1. 部署插件

```bash
# 构建插件
make my-plugin

# 复制到插件目录
cp examples/my_plugin.so /path/to/hackmitm/plugins/examples/

# 更新配置文件
# 重启HackMITM
```

### 2. 测试插件

```bash
# 启动HackMITM，查看插件加载情况
./hackmitm --config=./configs/config.json --verbose

# 检查插件状态
curl http://localhost:9090/status

# 查看插件统计
curl http://localhost:9090/metrics
```

### 3. 调试插件

```bash
# 启用详细日志
./hackmitm --log-level=debug

# 禁用特定插件进行测试
# 在配置文件中设置 "enabled": false

# 完全禁用插件系统
./hackmitm --disable-plugins
```

## 📋 最佳实践

### 1. 错误处理

```go
func (p *MyPlugin) ProcessRequest(req *http.Request, ctx *plugin.RequestContext) error {
    defer func() {
        if r := recover(); r != nil {
            logger.Errorf("插件 %s 崩溃: %v", p.Name(), r)
        }
    }()
    
    // 插件逻辑
    return nil
}
```

### 2. 性能优化

- 使用原子操作进行计数器操作
- 避免在插件中执行耗时操作
- 合理使用缓存减少重复计算
- 使用连接池管理外部资源

### 3. 内存管理

```go
func (p *MyPlugin) Cleanup() error {
    // 清理资源
    if p.httpClient != nil {
        p.httpClient.CloseIdleConnections()
    }
    
    if p.file != nil {
        p.file.Close()
    }
    
    return p.BasePlugin.Cleanup()
}
```

### 4. 线程安全

```go
type MyPlugin struct {
    *plugin.BasePlugin
    mutex sync.RWMutex
    data  map[string]interface{}
}

func (p *MyPlugin) updateData(key string, value interface{}) {
    p.mutex.Lock()
    defer p.mutex.Unlock()
    p.data[key] = value
}
```

## 🔍 故障排除

### 常见问题

1. **插件加载失败**
   - 检查插件文件路径是否正确
   - 确认插件文件权限
   - 验证Go版本兼容性

2. **插件初始化失败**
   - 检查配置格式是否正确
   - 验证必需的配置项
   - 查看详细错误日志

3. **插件运行时错误**
   - 启用详细日志模式
   - 检查插件中的错误处理
   - 验证外部依赖可用性

### 调试技巧

```bash
# 检查插件符号
nm -D examples/my_plugin.so | grep -E "(NewPlugin|LoadPlugin)"

# 验证插件格式
file examples/my_plugin.so

# 查看插件依赖
ldd examples/my_plugin.so
```

## 📚 示例插件

参考 `plugins/examples/` 目录中的示例插件：

- `request_logger.go` - 请求日志插件
- `traffic_stats.go` - 流量统计插件
- `sql_injection_detector.go` - SQL注入检测插件

这些示例展示了不同类型插件的完整实现。

## 🤝 贡献

欢迎提交插件到HackMITM项目！请遵循以下步骤：

1. Fork项目仓库
2. 创建功能分支
3. 开发和测试插件
4. 提交Pull Request
5. 更新文档

---

更多信息请访问 [HackMITM GitHub 仓库](https://github.com/JishiTeam-J1wa/hackmitm) 