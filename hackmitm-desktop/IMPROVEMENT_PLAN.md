# HackMITM-Desktop 改进计划

## 问题清单

### P0 - 关键问题（影响核心功能）

| # | 问题 | 位置 | 影响 |
|---|------|------|------|
| 1 | **修改请求后转发未实现** | `ProxyTab.tsx:242` | 用户无法修改拦截的请求 |
| 2 | **配置持久化未实现** | `proxy_config_api.go:138` | 重启后配置丢失 |
| 3 | **拦截队列状态不同步** | `interceptStore` vs `trafficStore` | 状态管理混乱 |

### P1 - 重要问题（影响用户体验）

| # | 问题 | 位置 | 影响 |
|---|------|------|------|
| 4 | **测试连接按钮空实现** | `ProxySettings.tsx:548` | 按钮无响应 |
| 5 | **字典导入/导出未集成** | 前端未调用 | 功能不可用 |
| 6 | **后端事件未监听** | `connection:status` 等 | 状态不同步 |
| 7 | **API Key 明文存储** | localStorage | 安全风险 |

### P2 - 次要问题（代码质量）

| # | 问题 | 位置 | 影响 |
|---|------|------|------|
| 8 | Console.log 调试代码 | Header.tsx, ProxyTab.tsx | 生产环境污染 |
| 9 | 部分更新未实现 | `proxy_config_api.go:144` | 功能不完整 |
| 10 | Favicon 匹配未实现 | `layered_index.go:362` | 指纹识别不完整 |

---

## 分段实施计划

### 第一阶段：核心功能完善（1-2周）

#### 1.1 请求修改转发功能
**目标**：实现拦截请求的修改和转发

```
实现步骤：
1. 创建 ModifyRequest API 方法 (Go)
2. 实现请求修改逻辑（headers, body, method, url）
3. 前端集成修改后的数据发送
4. 添加修改历史记录

技术要点：
- 使用 http.Request 的 Clone 方法
- 支持修改：Method, URL, Headers, Body
- 保留原始请求的 ID 关联
```

**Go 后端新增方法**：
```go
// ModifyAndForwardRequest 修改并转发请求
func (a *App) ModifyAndForwardRequest(requestID string, modifications map[string]interface{}) error {
    // 1. 获取原始请求
    // 2. 应用修改
    // 3. 重新发送
    // 4. 返回新响应
}
```

#### 1.2 配置持久化
**目标**：配置保存到本地文件/数据库

```
实现步骤：
1. 设计配置文件格式 (JSON/YAML)
2. 实现配置读写服务
3. 启动时加载配置
4. 修改时自动保存

技术要点：
- 配置文件位置：~/.hackmitm/config.json
- 使用 fsnotify 监听配置变化
- 支持配置导入/导出
```

#### 1.3 统一状态管理
**目标**：解决状态重复和不同步问题

```
实现步骤：
1. 合并 interceptStore 到 trafficStore
2. 添加后端事件监听机制
3. 实现状态自动同步
4. 添加状态持久化

技术要点：
- 使用 Zustand 的 subscribeWithSelector
- WebSocket 或 SSE 实现推送
- 事件驱动更新
```

---

### 第二阶段：安全增强（1周）

#### 2.1 敏感数据保护
```
实现步骤：
1. 使用系统密钥环存储 API Key
   - macOS: Keychain
   - Windows: Credential Manager
   - Linux: Secret Service API

2. 敏感字段加密存储
   - 使用 AES-GCM 加密
   - 密钥派生使用 Argon2

3. 内存安全
   - 敏感数据使用后立即清零
   - 避免日志记录敏感信息
```

**Go 密钥环集成**：
```go
import "github.com/zalando/go-keyring"

const service = "HackMITM"

func SaveAPIKey(key string) error {
    return keyring.Set(service, "api_key", key)
}

func GetAPIKey() (string, error) {
    return keyring.Get(service, "api_key")
}
```

#### 2.2 输入验证
```
实现步骤：
1. 前端验证
   - URL 格式验证
   - 端口范围验证
   - 输入长度限制

2. 后端验证
   - 参数类型检查
   - 业务逻辑验证
   - 防注入过滤
```

---

### 第三阶段：高级功能（2-3周）

#### 3.1 实时事件推送
**目标**：使用 WebSocket 替代轮询

```
架构设计：
┌─────────────┐     WebSocket      ┌─────────────┐
│   Desktop   │ ◄─────────────────► │  HackMITM   │
│    App      │                     │   Server    │
└─────────────┘                     └─────────────┘
      │                                   │
      ▼                                   ▼
  事件处理                            事件源
  - 流量更新                          - 代理事件
  - 拦截通知                          - 流量捕获
  - 状态变化                          - 系统状态
```

**事件类型定义**：
```typescript
interface WSEvent {
  type: 'traffic' | 'intercept' | 'status' | 'fingerprint'
  action: 'create' | 'update' | 'delete'
  data: any
  timestamp: number
}
```

#### 3.2 流量分析引擎
**目标**：智能流量分析和异常检测

```
功能设计：
1. 流量模式识别
   - 请求频率分析
   - 参数模式提取
   - 响应特征分析

2. 异常检测
   - 响应时间异常
   - 状态码异常
   - 大小异常

3. 智能分类
   - 自动标记 API 请求
   - 识别静态资源
   - 检测敏感数据
```

**算法实现**：
```go
// 流量分析器
type TrafficAnalyzer struct {
    patterns    []*Pattern
    anomalies   []*Anomaly
    classifier  *Classifier
}

// 异常检测算法
func (a *TrafficAnalyzer) DetectAnomaly(item *TrafficItem) *Anomaly {
    // 1. 统计分析（Z-Score）
    // 2. 频率分析（滑动窗口）
    // 3. 内容分析（正则/ML）
}
```

#### 3.3 请求编排（Macro）
**目标**：实现请求链和自动化测试

```
功能设计：
1. 请求链定义
   - 顺序执行
   - 条件分支
   - 循环控制

2. 数据提取
   - 正则提取
   - JSONPath
   - XPath

3. 变量传递
   - 请求间数据传递
   - 环境变量
   - 动态替换
```

**编排示例**：
```yaml
name: "登录流程测试"
steps:
  - name: "获取验证码"
    request:
      method: GET
      url: /api/captcha
    extract:
      captcha_id: "$.data.id"

  - name: "登录"
    request:
      method: POST
      url: /api/login
      body:
        username: "${env.USERNAME}"
        password: "${env.PASSWORD}"
        captcha_id: "${captcha_id}"
    assert:
      - "$.code == 0"
```

---

### 第四阶段：性能优化（1周）

#### 4.1 内存优化
```
优化策略：
1. 流量数据分页加载
2. LRU 缓存策略
3. 大响应体压缩存储
4. 定期清理过期数据

目标：
- 内存占用 < 500MB（10万条记录）
- 响应时间 < 100ms
```

#### 4.2 渲染优化
```
优化策略：
1. 虚拟列表（react-window）
2. 懒加载详情面板
3. 防抖/节流处理
4. Web Worker 处理大数据

目标：
- 首屏渲染 < 1s
- 滚动流畅度 60fps
```

---

### 第五阶段：用户体验（1周）

#### 5.1 快捷键系统
```
快捷键设计：
- Ctrl+R: 发送到 Repeater
- Ctrl+I: 开关拦截
- Ctrl+F: 快速搜索
- Ctrl+Shift+C: 复制为 cURL
- Space: 快速转发
- Delete: 丢弃请求
```

#### 5.2 主题系统
```
主题设计：
- 浅色主题（默认）
- 深色主题
- 高对比度主题
- 自定义主题（JSON 配置）
```

---

## 技术栈升级建议

### 前端
| 当前 | 建议 | 原因 |
|------|------|------|
| Zustand | 保持 | 轻量、足够用 |
| - | TanStack Query | 数据缓存和同步 |
| - | React Hook Form | 表单验证 |
| - | Zod | 运行时类型验证 |

### 后端
| 当前 | 建议 | 原因 |
|------|------|------|
| HTTP 轮询 | WebSocket | 实时性 |
| 内存存储 | SQLite/BoltDB | 持久化 |
| - | go-keyring | 安全存储 |

---

## 实施优先级

```
Week 1-2: 阶段一（核心功能）
├── 请求修改转发
├── 配置持久化
└── 状态管理重构

Week 3: 阶段二（安全增强）
├── API Key 安全存储
└── 输入验证

Week 4-6: 阶段三（高级功能）
├── WebSocket 推送
├── 流量分析引擎
└── 请求编排

Week 7: 阶段四（性能优化）
├── 内存优化
└── 渲染优化

Week 8: 阶段五（用户体验）
├── 快捷键
└── 主题系统
```

---

## 验收标准

### 功能验收
- [ ] 拦截请求可修改并正确转发
- [ ] 配置重启后保留
- [ ] 所有按钮有实际功能
- [ ] 无 console.log 残留

### 性能验收
- [ ] 10万条记录内存 < 500MB
- [ ] 搜索响应 < 100ms
- [ ] UI 无明显卡顿

### 安全验收
- [ ] API Key 不在日志中出现
- [ ] 敏感数据加密存储
- [ ] 输入验证覆盖所有用户输入

---

## 附录：未发现的潜在问题

### 竞态条件
- 并发访问 trafficStore
- 拦截队列操作

### 资源泄漏
- 长连接未正确关闭
- 定时器未清理

### 边界情况
- 空响应体处理
- 超大请求体处理
- 特殊字符编码
