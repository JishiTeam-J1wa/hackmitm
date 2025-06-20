<div align="center">

# 🚀 HackMITM

<img src="https://img.shields.io/badge/Go-1.21+-00ADD8?style=for-the-badge&logo=go" alt="Go Version">
<img src="https://img.shields.io/badge/License-MIT-green.svg?style=for-the-badge" alt="License">
<img src="https://img.shields.io/badge/Platform-Linux%20%7C%20macOS%20%7C%20Windows-lightgrey?style=for-the-badge" alt="Platform">
<img src="https://img.shields.io/badge/Status-Production%20Ready-brightgreen?style=for-the-badge" alt="Status">

**🔥 高性能 MITM 代理工具 · 纯 Go 语言实现 · 企业级安全 🔥**

*一个革命性的中间人代理处理包，专为安全研究、渗透测试和流量分析而生*

[✨ 特性介绍](#-主要特性) • [🚀 快速开始](#-快速开始) • [📖 文档](#-文档) • [🛠️ 高级用法](#️-高级用法) • [🤝 贡献](#-贡献指南)

---

</div>

## 📋 目录

- [🌟 项目亮点](#-项目亮点)
- [✨ 主要特性](#-主要特性)
- [🏗️ 架构设计](#️-架构设计)
- [🚀 快速开始](#-快速开始)
- [📦 安装方式](#-安装方式)
- [⚙️ 配置说明](#️-配置说明)
- [🛠️ 高级用法](#️-高级用法)
- [📊 性能测试](#-性能测试)
- [🔒 安全设计](#-安全设计)
- [📖 文档](#-文档)
- [🤝 贡献指南](#-贡献指南)
- [📄 许可证](#-许可证)
- [🙏 致谢](#-致谢)

## 🌟 项目亮点

<table>
<tr>
<td width="50%" valign="top">

### 🎯 **核心优势**
- 🚀 **零依赖**: 纯 Go 语言实现，无外部依赖
- ⚡ **高性能**: 基于 goroutine 的高并发架构
- 🔐 **企业级安全**: 完整的 TLS 证书管理和加密
- 🧩 **模块化设计**: 可插拔的处理器架构
- 🔄 **热配置**: 无需重启的配置热加载
- 📈 **生产就绪**: 内置监控、日志和容器化支持

</td>
<td width="50%" valign="top">

### 🎨 **技术特色**
- 📡 **透明代理**: 对客户端完全透明
- 🔒 **动态证书**: 自动生成域名相关 TLS 证书
- 🎛️ **流量处理**: 可扩展的请求/响应处理链
- 📊 **实时监控**: 内置性能监控和统计
- 🐳 **容器化**: 完整的 Docker 和 Kubernetes 支持
- 🔧 **开发友好**: 丰富的开发工具和调试功能

</td>
</tr>
</table>

## ✨ 主要特性

### 🔥 **核心功能**

<details>
<summary><b>🛡️ 完整的 MITM 代理功能</b></summary>

- **HTTP/HTTPS 透明代理**: 完全透明的代理服务，对客户端应用无感知
- **动态证书生成**: 根据目标域名自动生成有效的 TLS 证书
- **CA 证书管理**: 完整的 CA 证书生成、存储和管理系统
- **TLS 流量解密**: 支持 HTTPS 流量的完整解密和重加密
- **证书缓存**: 智能缓存机制，避免重复生成相同域名证书

</details>

<details>
<summary><b>⚡ 高性能架构</b></summary>

- **并发处理**: 每个连接独立的 goroutine 处理，支持数万并发连接
- **连接池优化**: 智能连接复用和池化管理，减少资源消耗
- **零拷贝**: 优化的数据传输，减少内存拷贝开销
- **缓存机制**: 多级缓存系统（证书缓存、连接缓存、响应缓存）
- **内存优化**: 精心设计的内存管理，避免内存泄漏

</details>

<details>
<summary><b>🧩 可扩展架构</b></summary>

- **插件式处理器**: 支持自定义请求/响应处理器
- **中间件架构**: 类似于 Web 框架的中间件模式
- **事件驱动**: 基于事件的处理模型，易于扩展
- **接口化设计**: 清晰的接口定义，便于二次开发
- **热插拔**: 支持运行时动态添加/移除处理器

</details>

### 🎛️ **管理功能**

| 功能 | 描述 | 状态 |
|------|------|------|
| 📊 **实时监控** | 内置 pprof 性能分析，支持 Prometheus 指标 | ✅ |
| 📝 **分级日志** | 支持 Debug/Info/Warn/Error 四级日志 | ✅ |
| ⚙️ **配置管理** | JSON 配置文件 + 命令行参数 + 环境变量 | ✅ |
| 🔄 **热加载** | 配置文件变更自动检测和应用 | ✅ |
| 🐳 **容器化** | 完整的 Docker 和 docker-compose 支持 | ✅ |
| 📈 **负载均衡** | 支持上游代理链和负载均衡 | ✅ |

## 🏗️ 架构设计

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   📱 客户端应用   │────│  🚀 HackMITM     │────│  🌐 目标服务器    │
│   Client App    │    │  Proxy Server   │    │  Target Server  │
└─────────────────┘    └─────────────────┘    └─────────────────┘
                              │
                              ▼
                    ┌─────────────────────┐
                    │    🧩 模块化组件     │
                    │  ┌───────────────┐  │
                    │  │ 🔐 证书管理模块 │  │
                    │  │ Certificate   │  │
                    │  │   Manager     │  │
                    │  └───────────────┘  │
                    │  ┌───────────────┐  │
                    │  │ 🔄 流量处理模块 │  │
                    │  │   Traffic     │  │
                    │  │  Processor    │  │
                    │  └───────────────┘  │
                    │  ┌───────────────┐  │
                    │  │ ⚙️ 配置管理模块 │  │
                    │  │ Configuration │  │
                    │  │   Manager     │  │
                    │  └───────────────┘  │
                    │  ┌───────────────┐  │
                    │  │ 📝 日志系统     │  │
                    │  │    Logger     │  │
                    │  │    System     │  │
                    │  └───────────────┘  │
                    └─────────────────────┘
```

## 🚀 快速开始

### ⭐ **5分钟快速体验**

```bash
# 🎯 第一步：克隆项目
git clone https://github.com/your-org/hackmitm.git
cd hackmitm

# 🔨 第二步：构建项目
go build -o hackmitm ./cmd/hackmitm

# 🚀 第三步：启动代理服务器
./hackmitm

# 🎉 第四步：测试代理功能
curl --proxy http://localhost:8080 https://httpbin.org/get
```

### 🔒 **HTTPS 证书设置**

```bash
# 导出 CA 证书
./hackmitm -export-ca ./ca-cert.pem

# 安装 CA 证书（macOS）
sudo security add-trusted-cert -d -r trustRoot -k /Library/Keychains/System.keychain ca-cert.pem

# 安装 CA 证书（Linux）
sudo cp ca-cert.pem /usr/local/share/ca-certificates/hackmitm.crt
sudo update-ca-certificates

# 测试 HTTPS 代理
curl --proxy http://localhost:8080 https://httpbin.org/get
```

## 📦 安装方式

<details>
<summary><b>📥 方式一：从源码构建（推荐）</b></summary>

```bash
# 环境要求：Go 1.21+
git clone https://github.com/your-org/hackmitm.git
cd hackmitm

# 安装依赖
go mod tidy

# 构建可执行文件
make build

# 或者直接使用 go build
go build -ldflags "-X main.Version=1.0.0" -o hackmitm ./cmd/hackmitm
```

</details>

<details>
<summary><b>🐳 方式二：Docker 容器（最简单）</b></summary>

```bash
# 使用 docker-compose（推荐）
docker-compose up -d

# 或者直接运行 Docker
docker run -d \
  --name hackmitm \
  -p 8080:8080 \
  -v $(pwd)/certs:/app/certs \
  hackmitm:latest
```

</details>

<details>
<summary><b>📦 方式三：预编译二进制文件</b></summary>

```bash
# 下载最新版本
wget https://github.com/your-org/hackmitm/releases/latest/download/hackmitm-linux-amd64.tar.gz

# 解压安装
tar -xzf hackmitm-linux-amd64.tar.gz
sudo mv hackmitm /usr/local/bin/
```

</details>

## ⚙️ 配置说明

<details>
<summary><b>📝 基础配置示例</b></summary>

```json
{
  "server": {
    "listen_port": 8080,
    "listen_addr": "0.0.0.0"
  },
  "tls": {
    "cert_dir": "./certs",
    "enable_cert_cache": true,
    "cert_cache_ttl": "24h"
  },
  "proxy": {
    "enable_http": true,
    "enable_https": true,
    "enable_compression": true
  },
  "logging": {
    "level": "info",
    "output": "stdout",
    "format": "text"
  }
}
```

</details>

<details>
<summary><b>🔧 高级配置选项</b></summary>

| 配置项 | 描述 | 默认值 |
|--------|------|--------|
| `server.listen_port` | 代理服务器监听端口 | `8080` |
| `tls.enable_cert_cache` | 启用证书缓存 | `true` |
| `proxy.max_idle_conns` | 最大空闲连接数 | `100` |
| `logging.level` | 日志级别 | `info` |
| `performance.max_goroutines` | 最大协程数 | `10000` |

</details>

## 🛠️ 高级用法

### 🔌 **自定义处理器开发**

```go
// 创建自定义请求处理器
type CustomHandler struct {
    config map[string]interface{}
}

func (h *CustomHandler) HandleRequest(req *http.Request) error {
    // 添加自定义请求头
    req.Header.Set("X-Custom-Header", "MyValue")
    
    // 记录请求信息
    log.Printf("处理请求: %s %s", req.Method, req.URL.String())
    
    return nil
}

// 注册处理器
server.AddRequestHandler(&CustomHandler{})
```

### 📊 **性能监控设置**

```bash
# 启用 pprof 性能分析
./hackmitm -config config.json

# 访问性能分析页面
go tool pprof http://localhost:6060/debug/pprof/profile

# CPU 性能分析
make pprof-cpu

# 内存分析
make pprof-mem
```

### 🔗 **代理链配置**

```bash
# 设置上游代理
./hackmitm -upstream http://upstream-proxy:8080

# 代理链：Client → HackMITM → Upstream → Target
```

## 📊 性能测试

<details>
<summary><b>⚡ 性能基准测试结果</b></summary>

| 指标 | 数值 | 说明 |
|------|------|------|
| **并发连接** | 10,000+ | 同时处理的连接数 |
| **吞吐量** | 50,000 RPS | 每秒请求处理数 |
| **延迟** | < 1ms | 平均代理延迟 |
| **内存使用** | < 100MB | 稳定状态内存占用 |
| **CPU 使用** | < 10% | 高负载下 CPU 占用 |

</details>

<details>
<summary><b>🧪 运行性能测试</b></summary>

```bash
# 基准测试
make bench

# 压力测试
go test -bench=. -benchmem ./...

# 自定义测试
ab -n 10000 -c 100 -X localhost:8080 http://httpbin.org/get
```

</details>

## 🔒 安全设计

### 🛡️ **安全特性**

- **🔐 加密算法**: 使用 ECDSA P-256 高强度加密
- **🔒 证书安全**: CA 私钥权限控制 (0600)
- **🚫 访问控制**: 支持基于 IP 和用户的访问控制
- **📝 审计日志**: 完整的访问和操作日志记录
- **🛡️ 输入验证**: 严格的输入参数验证和清理
- **⚠️ 错误处理**: 安全的错误处理，防止信息泄露

### 🔍 **安全最佳实践**

<details>
<summary><b>查看安全配置建议</b></summary>

```bash
# 设置安全权限
chmod 700 ./certs/
chmod 600 ./certs/ca-key.pem

# 限制访问来源
./hackmitm -config secure-config.json

# 启用审计日志
./hackmitm -log-level info -verbose
```

</details>

## 📖 文档

| 文档类型 | 链接 | 描述 |
|----------|------|------|
| 🇨🇳 **中文文档** | [docs/README_zh.md](./docs/README_zh.md) | 完整的中文技术文档 |
| 📚 **用户手册** | [docs/user_manual_zh.md](./docs/user_manual_zh.md) | 详细的使用指南 |
| 🔧 **API 文档** | [pkg/](./pkg/) | Go 包文档和 API 参考 |
| 💡 **示例代码** | [examples/](./examples/) | 自定义处理器示例 |
| 🐳 **部署指南** | [Dockerfile](./Dockerfile) | 容器化部署说明 |

## 🚀 **项目路线图**

<details>
<summary><b>🗓️ 开发计划</b></summary>

### ✅ **已完成**
- [x] 核心代理功能
- [x] 证书管理系统
- [x] 流量处理框架
- [x] 配置热加载
- [x] 容器化支持
- [x] 完整文档

### 🔄 **进行中**
- [ ] WebUI 管理界面
- [ ] RESTful API
- [ ] 插件市场
- [ ] 集群模式

### 📋 **计划中**
- [ ] 图形界面客户端
- [ ] 云原生支持
- [ ] 机器学习流量分析
- [ ] 更多协议支持

</details>

## 🤝 贡献指南

我们欢迎所有形式的贡献！🎉

### 🌟 **如何贡献**

1. **🍴 Fork** 项目仓库
2. **🔀 创建** 特性分支: `git checkout -b feature/amazing-feature`
3. **💾 提交** 更改: `git commit -m 'Add amazing feature'`
4. **📤 推送** 分支: `git push origin feature/amazing-feature`
5. **🔗 提交** Pull Request

### 📋 **贡献类型**

- 🐛 **Bug 修复**: 发现并修复项目中的问题
- ✨ **新功能**: 添加新的功能特性
- 📝 **文档**: 改进项目文档和注释
- 🎨 **代码优化**: 提升代码质量和性能
- 🧪 **测试**: 增加测试覆盖率
- 🌐 **国际化**: 添加多语言支持

### 🏆 **贡献者**

感谢以下贡献者对项目的支持：

<a href="https://github.com/your-org/hackmitm/graphs/contributors">
  <img src="https://contrib.rocks/image?repo=your-org/hackmitm" />
</a>

## 📊 **项目统计**

<div align="center">

![GitHub stars](https://img.shields.io/github/stars/your-org/hackmitm?style=social)
![GitHub forks](https://img.shields.io/github/forks/your-org/hackmitm?style=social)
![GitHub issues](https://img.shields.io/github/issues/your-org/hackmitm)
![GitHub pull requests](https://img.shields.io/github/issues-pr/your-org/hackmitm)

![Code size](https://img.shields.io/github/languages/code-size/your-org/hackmitm)
![License](https://img.shields.io/github/license/your-org/hackmitm)
![Go Report Card](https://goreportcard.com/badge/github.com/your-org/hackmitm)

</div>

## 📞 **支持与联系**

<div align="center">

### 💬 **获取帮助**

| 方式 | 链接 | 说明 |
|------|------|------|
| 📖 **文档** | [用户手册](./docs/user_manual_zh.md) | 详细使用说明 |
| 🐛 **Bug 报告** | [GitHub Issues](https://github.com/your-org/hackmitm/issues) | 问题反馈 |
| 💬 **讨论** | [GitHub Discussions](https://github.com/your-org/hackmitm/discussions) | 社区讨论 |
| 📧 **邮件** | hackmitm@example.com | 技术支持 |

### 🌐 **社区**

[![Telegram](https://img.shields.io/badge/Telegram-2CA5E0?style=for-the-badge&logo=telegram&logoColor=white)](https://t.me/hackmitm)
[![Discord](https://img.shields.io/badge/Discord-7289DA?style=for-the-badge&logo=discord&logoColor=white)](https://discord.gg/hackmitm)
[![QQ群](https://img.shields.io/badge/QQ群-EB1923?style=for-the-badge&logo=tencent-qq&logoColor=white)](https://qm.qq.com/cgi-bin/qm/qr?k=xxx)

</div>

## 📄 许可证

<div align="center">

本项目采用 **MIT 许可证** 开源。

详情请参阅 [LICENSE](LICENSE) 文件。

```
MIT License - 自由使用、修改、分发
```

</div>

## 🙏 致谢

<div align="center">

### 💖 **特别感谢**

感谢以下项目和技术为 HackMITM 提供灵感和支持：

- [Go 语言](https://golang.org/) - 优雅的系统编程语言
- [Gin](https://github.com/gin-gonic/gin) - 架构设计参考
- [mitmproxy](https://mitmproxy.org/) - 功能设计灵感
- [所有贡献者](https://github.com/your-org/hackmitm/graphs/contributors) - 项目发展的推动者

### ⭐ **支持项目**

如果 HackMITM 对您有帮助，请给我们一个 ⭐ Star！

这是对我们最大的鼓励和支持 💪

</div>

---

<div align="center">

**🚀 HackMITM - 让网络流量分析变得简单而强大！**

*Made with ❤️ by HackMITM Team*

</div> 